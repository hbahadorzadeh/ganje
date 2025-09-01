package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/hbahadorzadeh/ganje/internal/database"
	"github.com/hbahadorzadeh/ganje/internal/messaging"
)

// DispatcherConfig controls retry/backoff behavior.
type DispatcherConfig struct {
	MaxRetries      int
	InitialBackoff  time.Duration
	MaxBackoff      time.Duration
	HTTPTimeout     time.Duration
}

// Dispatcher asynchronously delivers webhook events.
type Dispatcher struct {
	db     database.DatabaseInterface
	cfg    DispatcherConfig
	queue  chan messaging.Event
	wg     sync.WaitGroup
	quit   chan struct{}
	client *http.Client
}

func NewDispatcher(db database.DatabaseInterface, cfg DispatcherConfig) *Dispatcher {
	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = 5
	}
	if cfg.InitialBackoff == 0 {
		cfg.InitialBackoff = 500 * time.Millisecond
	}
	if cfg.MaxBackoff == 0 {
		cfg.MaxBackoff = 10 * time.Second
	}
	if cfg.HTTPTimeout == 0 {
		cfg.HTTPTimeout = 10 * time.Second
	}
	return &Dispatcher{
		db:    db,
		cfg:   cfg,
		queue: make(chan messaging.Event, 100),
		quit:  make(chan struct{}),
		client: &http.Client{Timeout: cfg.HTTPTimeout},
	}
}

func (d *Dispatcher) Start(workers int) {
	if workers <= 0 {
		workers = 2
	}
	for i := 0; i < workers; i++ {
		d.wg.Add(1)
		go d.worker()
	}
}

func (d *Dispatcher) Stop(ctx context.Context) error {
	close(d.quit)
	close(d.queue)
	d.wg.Wait()
	return nil
}

func (d *Dispatcher) Enqueue(evt messaging.Event) {
	select {
	case d.queue <- evt:
	default:
		// drop if full to avoid backpressure on HTTP handlers
	}
}

func (d *Dispatcher) worker() {
	defer d.wg.Done()
	for evt := range d.queue {
		select {
		case <-d.quit:
			return
		default:
		}
		_ = d.dispatchEvent(context.Background(), evt)
	}
}

func (d *Dispatcher) dispatchEvent(ctx context.Context, evt messaging.Event) error {
	// Load webhooks for repository
	hooks, err := d.db.ListWebhooksByRepository(ctx, evt.Repository)
	if err != nil {
		return err
	}
	for _, h := range hooks {
		if !h.Enabled {
			continue
		}
		if !eventEnabled(h.Events, string(evt.Type)) {
			continue
		}
		payload, err := renderPayload(h.PayloadTemplate, evt)
		if err != nil {
			payload = []byte(defaultPayload(evt))
		}
		headers := parseHeaders(h.HeadersJSON)
		if h.SigningSecret != "" {
			// HMAC SHA256 signature of payload
			sig := signPayload(h.SigningSecret, payload)
			headers["X-Ganje-Signature"] = sig
		}
		if h.BearerToken != "" {
			headers["Authorization"] = "Bearer " + h.BearerToken
		} else if h.BasicUsername != "" || h.BasicPassword != "" {
			headers["Authorization"] = "Basic " + basicAuth(h.BasicUsername, h.BasicPassword)
		}
		statusCode, respErr := d.tryDeliver(h.URL, headers, payload)
		_ = d.db.RecordWebhookDelivery(ctx, &database.WebhookDelivery{
			WebhookID:  h.ID,
			Event:      string(evt.Type),
			StatusCode: statusCode,
			Success:    respErr == nil && statusCode >= 200 && statusCode < 300,
			Error:      errString(respErr),
			Payload:    string(payload),
		})
	}
	return nil
}

func (d *Dispatcher) tryDeliver(url string, headers map[string]string, body []byte) (int, error) {
	backoff := d.cfg.InitialBackoff
	var lastErr error
	status := 0
	for attempt := 0; attempt <= d.cfg.MaxRetries; attempt++ {
		req, _ := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		for k, v := range headers {
			req.Header.Set(k, v)
		}
		resp, err := d.client.Do(req)
		if err == nil && resp != nil {
			status = resp.StatusCode
			_ = resp.Body.Close()
			if status >= 200 && status < 300 {
				return status, nil
			}
			lastErr = fmt.Errorf("non-2xx status: %d", status)
		} else if err != nil {
			lastErr = err
		}
		// Retry on transport errors or 5xx
		if status >= 500 || lastErr != nil {
			time.Sleep(backoff)
			backoff *= 2
			if backoff > d.cfg.MaxBackoff {
				backoff = d.cfg.MaxBackoff
			}
			continue
		}
		break
	}
	return status, lastErr
}

func eventEnabled(eventsCSV string, t string) bool {
	if strings.TrimSpace(eventsCSV) == "" {
		return true
	}
	parts := strings.Split(eventsCSV, ",")
	for _, p := range parts {
		if strings.EqualFold(strings.TrimSpace(p), t) ||
			(strings.EqualFold(strings.TrimSpace(p), "add") && t == string(messaging.EventAdd)) ||
			(strings.EqualFold(strings.TrimSpace(p), "remove") && t == string(messaging.EventRemove)) ||
			(strings.EqualFold(strings.TrimSpace(p), "change") && t == string(messaging.EventChange)) {
			return true
		}
	}
	return false
}

func renderPayload(tmpl string, evt messaging.Event) ([]byte, error) {
	if strings.TrimSpace(tmpl) == "" {
		return []byte(defaultPayload(evt)), nil
	}
	t, err := template.New("payload").Parse(tmpl)
	if err != nil {
		return nil, err
	}
	buf := &bytes.Buffer{}
	if err := t.Execute(buf, evt); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func defaultPayload(evt messaging.Event) string {
	b, _ := json.Marshal(evt)
	return string(b)
}

func parseHeaders(hdrJSON string) map[string]string {
	m := map[string]string{}
	if strings.TrimSpace(hdrJSON) == "" {
		return m
	}
	_ = json.Unmarshal([]byte(hdrJSON), &m)
	return m
}

func signPayload(secret string, payload []byte) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	return "sha256=" + hex.EncodeToString(h.Sum(nil))
}

func basicAuth(username, password string) string {
	return base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
}

func errString(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}

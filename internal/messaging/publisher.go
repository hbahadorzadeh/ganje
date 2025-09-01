package messaging

import "time"

// EventType represents the kind of artifact lifecycle event
const (
	EventAdd    = "artifact.add"
	EventRemove = "artifact.remove"
	EventChange = "artifact.change"
)

// Event describes an artifact lifecycle event to publish
type Event struct {
	Type       string    `json:"type"`
	Repository string    `json:"repository"`
	Path       string    `json:"path"`
	Name       string    `json:"name,omitempty"`
	Version    string    `json:"version,omitempty"`
	Group      string    `json:"group,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
}

// Publisher defines a minimal interface for event publishing
type Publisher interface {
	Publish(e Event) error
	Close() error
}

// NoopPublisher is used when messaging is disabled
type NoopPublisher struct{}

func (n *NoopPublisher) Publish(e Event) error { return nil }
func (n *NoopPublisher) Close() error         { return nil }

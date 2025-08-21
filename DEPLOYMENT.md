# Ganje Deployment Guide

This guide covers deploying Ganje artifact repository manager using Docker and Kubernetes with Helm.

## Prerequisites

- Docker 20.10+
- Kubernetes 1.19+
- Helm 3.2+
- PostgreSQL database

## Quick Start

### 1. Build Docker Image

```bash
# Build the image
docker build -t ganje:latest .

# Test locally
docker run -p 8080:8080 ganje:latest
```

### 2. Deploy to Kubernetes

```bash
# Install with default values
helm install ganje ./helm/ganje

# Install with custom configuration
helm install ganje ./helm/ganje -f production-values.yaml
```

## Production Deployment

### Database Setup

Deploy PostgreSQL first:

```bash
# Add Bitnami repo
helm repo add bitnami https://charts.bitnami.com/bitnami

# Install PostgreSQL
helm install postgresql bitnami/postgresql \
  --set auth.postgresPassword=secure-password \
  --set auth.database=ganje \
  --set auth.username=ganje \
  --set auth.password=ganje-password \
  --set primary.persistence.size=20Gi
```

### Storage Configuration

For production, use a dedicated storage class:

```yaml
# storage-class.yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: ganje-storage
provisioner: kubernetes.io/aws-ebs  # or your cloud provider
parameters:
  type: gp3
  fsType: ext4
allowVolumeExpansion: true
```

### SSL/TLS Setup

Create TLS certificate:

```bash
# Using cert-manager
kubectl apply -f - <<EOF
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: ganje-tls
spec:
  secretName: ganje-tls-secret
  issuerRef:
    name: letsencrypt-prod
    kind: ClusterIssuer
  dnsNames:
  - artifacts.yourdomain.com
EOF
```

## Configuration Examples

### Development Environment

```yaml
# dev-values.yaml
replicaCount: 1

resources:
  requests:
    cpu: 100m
    memory: 128Mi
  limits:
    cpu: 500m
    memory: 512Mi

persistence:
  size: 5Gi

database:
  host: postgresql
  password: dev-password

ingress:
  hosts:
    - host: ganje.local
      paths:
        - path: /
          pathType: Prefix
```

### Production Environment

```yaml
# production-values.yaml
replicaCount: 3

image:
  repository: your-registry.com/ganje
  tag: "v1.0.0"
  pullPolicy: Always

resources:
  requests:
    cpu: 500m
    memory: 512Mi
  limits:
    cpu: 2000m
    memory: 2Gi

autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 10
  targetCPUUtilizationPercentage: 70

persistence:
  enabled: true
  size: 100Gi
  storageClass: ganje-storage

database:
  host: postgres.production.com
  port: 5432
  username: ganje_prod
  password: secure-production-password
  database: ganje_prod
  sslMode: require

ingress:
  enabled: true
  className: nginx
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/proxy-body-size: "500m"
    nginx.ingress.kubernetes.io/rate-limit: "100"
  hosts:
    - host: artifacts.yourdomain.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: ganje-tls-secret
      hosts:
        - artifacts.yourdomain.com

nodeSelector:
  node-type: compute-optimized

tolerations:
  - key: "dedicated"
    operator: "Equal"
    value: "ganje"
    effect: "NoSchedule"

affinity:
  podAntiAffinity:
    preferredDuringSchedulingIgnoredDuringExecution:
    - weight: 100
      podAffinityTerm:
        labelSelector:
          matchExpressions:
          - key: app.kubernetes.io/name
            operator: In
            values:
            - ganje
        topologyKey: kubernetes.io/hostname
```

## Monitoring and Observability

### Prometheus Metrics

Add monitoring annotations:

```yaml
# monitoring-values.yaml
podAnnotations:
  prometheus.io/scrape: "true"
  prometheus.io/port: "8080"
  prometheus.io/path: "/metrics"
```

### Logging

Configure structured logging:

```yaml
logging:
  level: info
  format: json
  output: stdout
```

## Backup and Recovery

### Database Backup

```bash
# Create backup job
kubectl create job --from=cronjob/postgres-backup postgres-backup-manual
```

### Storage Backup

```bash
# Snapshot PVC (AWS EBS example)
kubectl apply -f - <<EOF
apiVersion: snapshot.storage.k8s.io/v1
kind: VolumeSnapshot
metadata:
  name: ganje-storage-snapshot
spec:
  source:
    persistentVolumeClaimName: ganje-storage
  volumeSnapshotClassName: ebs-csi-snapshot-class
EOF
```

## Security Considerations

### Network Policies

```yaml
# network-policy.yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: ganje-network-policy
spec:
  podSelector:
    matchLabels:
      app.kubernetes.io/name: ganje
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: ingress-nginx
    ports:
    - protocol: TCP
      port: 8080
  egress:
  - to:
    - namespaceSelector:
        matchLabels:
          name: database
    ports:
    - protocol: TCP
      port: 5432
```

### Pod Security Standards

```yaml
podSecurityContext:
  runAsNonRoot: true
  runAsUser: 1001
  runAsGroup: 1001
  fsGroup: 1001
  seccompProfile:
    type: RuntimeDefault

securityContext:
  allowPrivilegeEscalation: false
  capabilities:
    drop:
    - ALL
  readOnlyRootFilesystem: true
  runAsNonRoot: true
  runAsUser: 1001
```

## Troubleshooting

### Common Issues

1. **Pod not starting**
   ```bash
   kubectl describe pod <pod-name>
   kubectl logs <pod-name>
   ```

2. **Database connection issues**
   ```bash
   kubectl exec -it <pod-name> -- nc -zv postgresql 5432
   ```

3. **Storage issues**
   ```bash
   kubectl get pvc
   kubectl describe pvc ganje-storage
   ```

4. **Ingress not working**
   ```bash
   kubectl get ingress
   kubectl describe ingress ganje
   ```

### Health Checks

Test application health:

```bash
# Port forward to test
kubectl port-forward svc/ganje 8080:8080

# Test health endpoint
curl http://localhost:8080/health

# Test repository endpoints
curl http://localhost:8080/maven-local/
```

## Scaling and Performance

### Horizontal Scaling

```bash
# Manual scaling
kubectl scale deployment ganje --replicas=5

# Enable autoscaling
kubectl autoscale deployment ganje --cpu-percent=70 --min=3 --max=10
```

### Performance Tuning

```yaml
resources:
  requests:
    cpu: 1000m
    memory: 1Gi
  limits:
    cpu: 4000m
    memory: 4Gi

# JVM tuning (if applicable)
env:
  - name: JAVA_OPTS
    value: "-Xmx2g -Xms1g -XX:+UseG1GC"
```

## Maintenance

### Updates

```bash
# Update image
helm upgrade ganje ./helm/ganje --set image.tag=v1.1.0

# Rollback if needed
helm rollback ganje 1
```

### Cleanup

```bash
# Uninstall application
helm uninstall ganje

# Clean up PVC (if needed)
kubectl delete pvc ganje-storage
```

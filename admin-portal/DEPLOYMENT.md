# Ganje Admin Portal Deployment Guide

This guide covers deployment options for the Ganje Admin Portal using Docker and Kubernetes (Helm).

## Prerequisites

- Docker and Docker Compose
- Kubernetes cluster (for Helm deployment)
- Helm 3.x
- Ganje backend service running

## Docker Deployment

### Building the Image

```bash
# Build the Docker image
docker build -t ganje/admin-portal:latest .

# Or using docker-compose
docker-compose build
```

### Running with Docker Compose

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f ganje-admin-portal

# Stop services
docker-compose down
```

The admin portal will be available at `http://localhost:4200`.

### Environment Variables

- `NODE_ENV`: Environment mode (production/development)
- `GANJE_BACKEND_URL`: URL of the Ganje backend service

## Kubernetes Deployment with Helm

### Installing the Chart

```bash
# Add the Helm repository (if applicable)
helm repo add ganje https://charts.ganje.io
helm repo update

# Install the chart
helm install ganje-admin-portal ./helm \
  --namespace ganje-system \
  --create-namespace \
  --set image.repository=ganje/admin-portal \
  --set image.tag=latest \
  --set ganje.backend.url=http://ganje-backend:8080
```

### Configuration Options

Key configuration values in `values.yaml`:

```yaml
# Scaling
replicaCount: 2
autoscaling:
  enabled: true
  minReplicas: 2
  maxReplicas: 10

# Ingress
ingress:
  enabled: true
  className: nginx
  hosts:
    - host: ganje-admin.example.com
      paths:
        - path: /
          pathType: Prefix

# Resources
resources:
  limits:
    cpu: 500m
    memory: 512Mi
  requests:
    cpu: 100m
    memory: 128Mi
```

### Upgrading the Deployment

```bash
# Upgrade with new values
helm upgrade ganje-admin-portal ./helm \
  --namespace ganje-system \
  --set image.tag=v1.1.0

# Rollback if needed
helm rollback ganje-admin-portal 1 --namespace ganje-system
```

### Monitoring and Troubleshooting

```bash
# Check pod status
kubectl get pods -n ganje-system -l app.kubernetes.io/name=ganje-admin-portal

# View logs
kubectl logs -n ganje-system -l app.kubernetes.io/name=ganje-admin-portal -f

# Check service endpoints
kubectl get svc -n ganje-system

# Port forward for local access
kubectl port-forward -n ganje-system svc/ganje-admin-portal 8080:80
```

## Security Considerations

1. **Container Security**:
   - Runs as non-root user (UID 1001)
   - Read-only root filesystem
   - Minimal Alpine-based image

2. **Network Security**:
   - All traffic over HTTPS in production
   - Proper ingress configuration with TLS
   - Network policies for pod-to-pod communication

3. **Authentication**:
   - Integrates with Dex OAuth provider
   - JWT token-based authentication
   - Configurable session timeouts

## Production Checklist

- [ ] Configure proper resource limits and requests
- [ ] Set up ingress with TLS certificates
- [ ] Configure horizontal pod autoscaling
- [ ] Set up monitoring and alerting
- [ ] Configure backup strategies
- [ ] Review security policies
- [ ] Test disaster recovery procedures

## Health Checks

The application provides a health check endpoint at `/health` that returns:
- HTTP 200 with "healthy" response when the application is running properly
- Used by Kubernetes liveness and readiness probes

## Troubleshooting

### Common Issues

1. **Pod not starting**:
   - Check resource constraints
   - Verify image pull secrets
   - Review pod logs

2. **Cannot connect to backend**:
   - Verify `GANJE_BACKEND_URL` environment variable
   - Check network connectivity
   - Ensure backend service is running

3. **Authentication issues**:
   - Verify Dex configuration
   - Check JWT token validity
   - Review OAuth client settings

### Support

For additional support and documentation, visit:
- GitHub: https://github.com/hbahadorzadeh/ganje
- Documentation: https://docs.ganje.io

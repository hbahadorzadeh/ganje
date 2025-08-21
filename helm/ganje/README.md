# Ganje Helm Chart

This Helm chart deploys Ganje, a universal artifact repository manager, on a Kubernetes cluster.

## Prerequisites

- Kubernetes 1.16+
- Helm 3.0+
- PostgreSQL database (can be deployed separately)

## Installing the Chart

To install the chart with the release name `ganje`:

```bash
helm install ganje ./helm/ganje
```

## Uninstalling the Chart

To uninstall/delete the `ganje` deployment:

```bash
helm delete ganje
```

## Configuration

The following table lists the configurable parameters of the Ganje chart and their default values.

| Parameter | Description | Default |
|-----------|-------------|---------|
| `replicaCount` | Number of replicas | `1` |
| `image.repository` | Image repository | `ganje` |
| `image.tag` | Image tag | `latest` |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `service.type` | Kubernetes service type | `ClusterIP` |
| `service.port` | Service port | `8080` |
| `ingress.enabled` | Enable ingress | `true` |
| `ingress.className` | Ingress class name | `nginx` |
| `ingress.hosts[0].host` | Hostname | `ganje.local` |
| `persistence.enabled` | Enable persistent storage | `true` |
| `persistence.size` | Storage size | `10Gi` |
| `persistence.storageClass` | Storage class | `""` |
| `database.host` | Database host | `postgresql` |
| `database.port` | Database port | `5432` |
| `database.username` | Database username | `ganje` |
| `database.password` | Database password | `password` |
| `database.database` | Database name | `ganje` |

## Supported Artifact Types

Ganje supports the following artifact types:

- **Maven**: Java artifacts (JAR, WAR, POM)
- **NPM**: Node.js packages
- **Docker**: Container images
- **PyPI**: Python packages
- **Generic**: Any file type

## Repository Types

- **Local**: Store artifacts locally
- **Remote**: Proxy remote repositories
- **Virtual**: Aggregate multiple repositories

## Examples

### Custom Values

Create a `values.yaml` file:

```yaml
ingress:
  hosts:
    - host: artifacts.example.com
      paths:
        - path: /
          pathType: Prefix

persistence:
  size: 50Gi
  storageClass: fast-ssd

database:
  host: my-postgres.example.com
  password: secure-password
```

Install with custom values:

```bash
helm install ganje ./helm/ganje -f values.yaml
```

### Using with External Database

```yaml
database:
  host: postgres.example.com
  port: 5432
  username: ganje_user
  password: secure_password
  database: ganje_db
  sslMode: require
```

### High Availability Setup

```yaml
replicaCount: 3

autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 10
  targetCPUUtilizationPercentage: 70

resources:
  requests:
    cpu: 500m
    memory: 512Mi
  limits:
    cpu: 1000m
    memory: 1Gi
```

## Accessing the Application

After installation, you can access Ganje through:

1. **Ingress** (if enabled): `http://ganje.local`
2. **Port Forward**: `kubectl port-forward svc/ganje 8080:8080`
3. **LoadBalancer** (if configured): External IP on port 8080

## Repository Endpoints

Once deployed, you can access different artifact repositories:

- Maven: `http://ganje.local/maven-local/`
- NPM: `http://ganje.local/npm-local/`
- Docker: `http://ganje.local/docker-local/`
- PyPI: `http://ganje.local/pypi-local/`

## Troubleshooting

### Check Pod Status
```bash
kubectl get pods -l app.kubernetes.io/name=ganje
```

### View Logs
```bash
kubectl logs -f deployment/ganje
```

### Check Configuration
```bash
kubectl get configmap ganje-config -o yaml
```

### Verify Storage
```bash
kubectl get pvc ganje-storage
```

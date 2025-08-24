# Kubernetes Deployment

This directory contains Kubernetes manifests for deploying the URL shortener application with a complete monitoring stack.

## Components

The deployment includes:

- **Application**: Go URL shortener API
- **Database**: PostgreSQL 16
- **Monitoring Stack**:
  - **Prometheus**: Metrics collection and storage
  - **Loki**: Log aggregation
  - **Promtail**: Log collection agent
  - **Grafana**: Monitoring dashboard and visualization

## Prerequisites

1. Kubernetes cluster (local or remote)
2. `kubectl` configured to access your cluster
3. `kustomize` (optional, for using overlays)
4. `kind` (for local development)
5. `docker` (for building images)

## Quick Start

### Deploy to Development Environment

#### Option 1: Quick Setup (Recommended)
```bash
# Use the provided script to recreate cluster with proper DNS
./scripts/recreate-cluster.sh
```

#### Option 2: Manual Setup
```bash
# Create KIND cluster with proper configuration
kind create cluster --config kind-config.yaml

# Build and load the Docker image
docker buildx build -t myapi:latest -f .docker/Dockerfile .
kind load docker-image myapi:latest

# Deploy the base configuration
kubectl apply -k k8s/base/

# Or use kustomize for development overlay
kubectl apply -k k8s/overlays/dev/
```

#### Option 3: Fix DNS Issues in Existing Cluster
If you have DNS issues in an existing cluster:
```bash
./scripts/fix-dns.sh
```

### Access the Services

After deployment, you can access the services:

#### Using NodePort (if no ingress controller)
- **API**: `http://<node-ip>:32440`
- **Grafana**: `http://<node-ip>:30300`

#### Using Ingress (if ingress controller is installed)
Add these entries to your `/etc/hosts` file:
```
127.0.0.1 api.local
127.0.0.1 grafana.local
127.0.0.1 prometheus.local
```

Then access:
- **API**: `http://api.local`
- **Grafana**: `http://grafana.local` (admin/admin)
- **Prometheus**: `http://prometheus.local`

## Configuration

### Environment Variables

The application uses the following environment variables:
- `DB_HOST`: PostgreSQL host (default: postgres)
- `DB_PORT`: PostgreSQL port (default: 5432)
- `DB_USER`: Database user
- `DB_PASSWORD`: Database password
- `DB_NAME`: Database name (default: shortener)

### Secrets

Database credentials are stored in Kubernetes secrets:
- `postgres-secret`: Contains database credentials

### Monitoring Configuration

#### Prometheus
- Scrapes metrics from pods with `prometheus.io/scrape: "true"` annotation
- Metrics endpoint: `/metrics` on port 4000
- Accessible via service `prometheus:9090`

#### Loki
- Collects logs from all pods via Promtail
- Accessible via service `loki:3100`

#### Promtail
- Runs as a DaemonSet to collect logs from all nodes
- Sends logs to Loki
- Uses Kubernetes service discovery

#### Grafana
- Pre-configured with Prometheus and Loki data sources
- Default credentials: admin/admin
- Accessible via service `grafana:3000`

## Health Checks

The application includes:
- **Liveness Probe**: `/health` endpoint
- **Readiness Probe**: `/health` endpoint
- **Resource Limits**: CPU and memory limits configured

## Scaling

### Application Scaling
```bash
# Scale the API deployment
kubectl scale deployment myapi --replicas=3 -n myapp
```

### Database Scaling
For production, consider using a managed PostgreSQL service or StatefulSet with persistent volumes.

## Monitoring

### Metrics
The application exposes Prometheus metrics at `/metrics` endpoint. Key metrics include:
- HTTP request duration
- Request count by status code
- Database connection metrics

### Logs
All application logs are collected by Promtail and sent to Loki. You can query logs in Grafana using LogQL.

### Dashboards
Grafana comes pre-configured with data sources. You can create custom dashboards for:
- Application performance metrics
- Database performance
- System resource usage

## Troubleshooting

### DNS Issues
If you encounter DNS resolution problems:
```bash
# Test DNS resolution
kubectl run test-dns --image=busybox --rm -it --restart=Never -n myapp -- nslookup postgres

# If DNS fails, try fixing it
./scripts/fix-dns.sh

# Or recreate the cluster with proper DNS
./scripts/recreate-cluster.sh
```

### Check Pod Status
```bash
kubectl get pods -n myapp
```

### View Logs
```bash
# Application logs
kubectl logs -f deployment/myapi -n myapp

# Prometheus logs
kubectl logs -f deployment/prometheus -n myapp

# Grafana logs
kubectl logs -f deployment/grafana -n myapp
```

### Access Services
```bash
# Port forward to access services locally
kubectl port-forward service/myapi 4000:80 -n myapp
kubectl port-forward service/grafana 3000:3000 -n myapp
kubectl port-forward service/prometheus 9090:9090 -n myapp
```

### Check ConfigMaps
```bash
kubectl get configmaps -n myapp
kubectl describe configmap prometheus-config -n myapp
```

## Cleanup

To remove all resources:
```bash
kubectl delete -k k8s/base/
```

## Production Considerations

1. **Persistent Storage**: Use PersistentVolumes for database and monitoring data
2. **Security**: Use proper secrets management and RBAC
3. **High Availability**: Deploy across multiple nodes
4. **Backup**: Implement database backup strategies
5. **Monitoring**: Set up alerting rules in Prometheus/Grafana
6. **SSL/TLS**: Configure proper certificates for ingress
7. **Resource Limits**: Adjust resource requests/limits based on actual usage

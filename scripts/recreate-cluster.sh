#!/bin/bash

set -e

echo "🔄 Recreating KIND cluster with proper DNS configuration..."

# Delete existing cluster if it exists
if kind get clusters | grep -q "kind"; then
    echo "🗑️  Deleting existing KIND cluster..."
    kind delete cluster --name kind
fi

# Create new cluster with our configuration
echo "🏗️  Creating new KIND cluster..."
kind create cluster --config kind-config.yaml

# Wait for the cluster to be ready
echo "⏳ Waiting for cluster to be ready..."
kubectl wait --for=condition=Ready nodes --all --timeout=300s

# Check if CoreDNS is running
echo "🔍 Checking CoreDNS status..."
kubectl get pods -n kube-system | grep coredns || echo "⚠️  No CoreDNS pods found, this might be normal for KIND"

# Test DNS resolution
echo "🧪 Testing DNS resolution..."
kubectl run test-dns --image=busybox --rm -it --restart=Never -- nslookup kubernetes.default || echo "⚠️  DNS test failed, but this might be normal during startup"

# Build and load the Docker image
echo "🐳 Building and loading Docker image..."
docker buildx build -t myapi:latest -f .docker/Dockerfile .
kind load docker-image myapi:latest

# Deploy the application
echo "🚀 Deploying application..."
kubectl apply -k k8s/base/

# Wait for pods to be ready
echo "⏳ Waiting for pods to be ready..."
kubectl wait --for=condition=Ready pods --all -n myapp --timeout=300s

# Test the deployment
echo "🧪 Testing deployment..."
kubectl get pods -n myapp

echo "✅ Cluster recreation complete!"
echo ""
echo "📋 Next steps:"
echo "1. Test DNS resolution: kubectl run test-dns --image=busybox --rm -it --restart=Never -n myapp -- nslookup postgres"
echo "2. Access the API: kubectl port-forward service/myapi 4000:80 -n myapp"
echo "3. Access Grafana: kubectl port-forward service/grafana 3000:3000 -n myapp"
echo "4. Access Prometheus: kubectl port-forward service/prometheus 9090:9090 -n myapp"

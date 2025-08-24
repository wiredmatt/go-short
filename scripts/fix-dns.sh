#!/bin/bash

set -e

echo "🔧 Attempting to fix DNS issues in existing cluster..."

# Check if we're in a KIND cluster
if ! kubectl config current-context | grep -q "kind"; then
    echo "Not connected to a KIND cluster. Please ensure you're using a KIND cluster."
    exit 1
fi

echo "🔍 Checking current DNS status..."
kubectl get pods -n kube-system | grep -E "(coredns|kube-dns)" || echo "⚠️  No DNS pods found"

# Try to install CoreDNS if it's missing
echo "📦 Installing CoreDNS..."
kubectl apply -f k8s/coredns.yaml

# Wait for CoreDNS to be ready
echo "⏳ Waiting for CoreDNS to be ready..."
kubectl wait --for=condition=Ready pods -l k8s-app=kube-dns -n kube-system --timeout=120s || echo "⚠️  CoreDNS not ready yet"

echo "Testing DNS resolution..."
sleep 10 
kubectl run test-dns --image=busybox --rm -it --restart=Never -- nslookup kubernetes.default || echo "⚠️  DNS test failed"

echo "DNS fix completed."

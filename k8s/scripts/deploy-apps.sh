#!/bin/bash
# Deploy application services (Backend, Frontend)

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
K8S_DIR="$(dirname "$SCRIPT_DIR")"

echo "=============================================="
echo "  Deploying Florus Applications"
echo "=============================================="

# 1. Deploy Backend
echo "[1/3] Deploying Backend..."
kubectl apply -f "$K8S_DIR/apps/backend.yaml"
echo "✓ Backend deployed"

# 2. Deploy Frontend
echo "[2/3] Deploying Frontend..."
kubectl apply -f "$K8S_DIR/apps/frontend.yaml"
echo "✓ Frontend deployed"

# 3. Deploy Ingress
echo "[3/3] Deploying Ingress..."
kubectl apply -f "$K8S_DIR/apps/ingress.yaml"
echo "✓ Ingress deployed"

# Wait for pods to be ready
echo ""
echo "Waiting for applications to be ready..."
kubectl wait --namespace florus \
  --for=condition=Ready pod \
  --selector=app=backend \
  --timeout=120s || echo "Backend not ready yet"

kubectl wait --namespace florus \
  --for=condition=Ready pod \
  --selector=app=frontend \
  --timeout=120s || echo "Frontend not ready yet"

echo ""
echo "=============================================="
echo "  Application deployment complete!"
echo "=============================================="
echo ""
echo "Services status:"
kubectl get pods -n florus
echo ""
echo "Access points:"
echo "  - Frontend: http://localhost:30080"
echo "  - Backend API: http://localhost:30081/api/health"
echo ""
echo "With Ingress (add to /etc/hosts: 127.0.0.1 florus.local api.florus.local):"
echo "  - Frontend: http://florus.local"
echo "  - API: http://api.florus.local"
echo ""

#!/bin/bash
# Setup Kind cluster for Florus local development

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
K8S_DIR="$(dirname "$SCRIPT_DIR")"
PROJECT_DIR="$(dirname "$K8S_DIR")"

echo "=============================================="
echo "  Florus Local Kubernetes Setup"
echo "=============================================="

# 1. Check prerequisites
echo "[1/7] Checking prerequisites..."
command -v docker >/dev/null 2>&1 || { echo "Docker is required but not installed."; exit 1; }
command -v kind >/dev/null 2>&1 || { echo "Kind is required but not installed."; exit 1; }
command -v kubectl >/dev/null 2>&1 || { echo "kubectl is required but not installed."; exit 1; }
echo "✓ All prerequisites met"

# 2. Delete existing cluster if exists
echo "[2/7] Cleaning up existing cluster..."
kind delete cluster --name florus-local 2>/dev/null || true
echo "✓ Cleanup complete"

# 3. Create Kind cluster
echo "[3/7] Creating Kind cluster..."
kind create cluster --config "$K8S_DIR/local/kind-config.yaml"
echo "✓ Cluster created"

# 4. Wait for cluster to be ready
echo "[4/7] Waiting for cluster to be ready..."
kubectl wait --for=condition=Ready nodes --all --timeout=120s
echo "✓ Cluster is ready"

# 5. Install NGINX Ingress Controller
echo "[5/7] Installing NGINX Ingress Controller..."
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml
echo "Waiting for ingress controller..."
sleep 10
kubectl wait --namespace ingress-nginx \
  --for=condition=ready pod \
  --selector=app.kubernetes.io/component=controller \
  --timeout=120s || true
echo "✓ Ingress controller installed"

# 6. Create namespaces
echo "[6/7] Creating namespaces..."
kubectl apply -f "$K8S_DIR/base/namespace.yaml"
echo "✓ Namespaces created"

# 7. Apply secrets and configmaps
echo "[7/7] Applying base configurations..."
kubectl apply -f "$K8S_DIR/base/configmap.yaml"
kubectl apply -f "$K8S_DIR/base/secrets.yaml"
echo "✓ Base configurations applied"

echo ""
echo "=============================================="
echo "  Cluster setup complete!"
echo "=============================================="
echo ""
echo "Next steps:"
echo "  1. Build Docker images: ./build-images.sh"
echo "  2. Deploy data layer: ./deploy-data.sh"
echo "  3. Deploy applications: ./deploy-apps.sh"
echo ""
echo "Useful commands:"
echo "  kubectl get nodes"
echo "  kubectl get pods -A"
echo "  kind get clusters"
echo ""

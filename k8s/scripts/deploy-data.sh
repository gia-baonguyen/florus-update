#!/bin/bash
# Deploy data layer services (Redis, Neo4j, Kafka, MinIO)

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
K8S_DIR="$(dirname "$SCRIPT_DIR")"

echo "=============================================="
echo "  Deploying Florus Data Layer"
echo "=============================================="

# 1. Deploy Redis
echo "[1/4] Deploying Redis..."
kubectl apply -f "$K8S_DIR/data/redis.yaml"
echo "✓ Redis deployed"

# 2. Deploy Neo4j
echo "[2/4] Deploying Neo4j..."
kubectl apply -f "$K8S_DIR/data/neo4j.yaml"
echo "✓ Neo4j deployed"

# 3. Deploy MinIO
echo "[3/4] Deploying MinIO..."
kubectl apply -f "$K8S_DIR/data/minio.yaml"
echo "✓ MinIO deployed"

# 4. Deploy Kafka
echo "[4/4] Deploying Kafka..."
kubectl apply -f "$K8S_DIR/data/kafka.yaml"
echo "✓ Kafka deployed"

# Wait for pods to be ready
echo ""
echo "Waiting for data services to be ready..."
kubectl wait --namespace florus-data \
  --for=condition=Ready pod \
  --selector=app=redis \
  --timeout=120s || echo "Redis not ready yet"

kubectl wait --namespace florus-data \
  --for=condition=Ready pod \
  --selector=app=neo4j \
  --timeout=180s || echo "Neo4j not ready yet"

kubectl wait --namespace florus-data \
  --for=condition=Ready pod \
  --selector=app=minio \
  --timeout=120s || echo "MinIO not ready yet"

echo ""
echo "=============================================="
echo "  Data layer deployment complete!"
echo "=============================================="
echo ""
echo "Services status:"
kubectl get pods -n florus-data
echo ""
echo "Access points (via NodePort):"
echo "  - Neo4j Browser: http://localhost:30474"
echo "  - Neo4j Bolt: bolt://localhost:30687"
echo "  - MinIO Console: http://localhost:30901"
echo ""

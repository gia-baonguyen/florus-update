#!/bin/bash
# Build Docker images and load them into Kind cluster

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$(dirname "$SCRIPT_DIR")")"

echo "=============================================="
echo "  Building Docker Images for Florus"
echo "=============================================="

# 1. Build Backend image
echo "[1/3] Building Backend image..."
cd "$PROJECT_DIR/backend"

# Create Dockerfile if not exists
if [ ! -f Dockerfile ]; then
cat > Dockerfile << 'EOF'
# Build stage
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o api ./cmd/api

# Run stage
FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/api .
COPY --from=builder /app/migrations ./migrations
EXPOSE 8081
CMD ["./api"]
EOF
fi

docker build -t florus-backend:local .
echo "✓ Backend image built"

# 2. Build Frontend image
echo "[2/3] Building Frontend image..."
cd "$PROJECT_DIR/frontend"

# Create Dockerfile if not exists
if [ ! -f Dockerfile ]; then
cat > Dockerfile << 'EOF'
# Build stage
FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

# Run stage
FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
EOF
fi

docker build -t florus-frontend:local .
echo "✓ Frontend image built"

# 3. Load images into Kind cluster
echo "[3/3] Loading images into Kind cluster..."
kind load docker-image florus-backend:local --name florus-local
kind load docker-image florus-frontend:local --name florus-local
echo "✓ Images loaded into cluster"

echo ""
echo "=============================================="
echo "  Image build complete!"
echo "=============================================="
echo ""
echo "Images available in cluster:"
echo "  - florus-backend:local"
echo "  - florus-frontend:local"
echo ""

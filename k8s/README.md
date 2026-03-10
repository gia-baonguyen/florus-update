# Florus Kubernetes Deployment

## Prerequisites

- Docker Desktop (running)
- Kind (Kubernetes IN Docker)
- kubectl

## Quick Start (Windows)

```batch
# 1. Deploy everything (first time)
cd k8s\scripts
deploy-all.bat

# 2. Check status
status.bat

# 3. Cleanup when done
cleanup.bat
```

## Manual Deployment Steps

### Step 1: Create Kind Cluster

```batch
kind create cluster --config k8s\local\kind-config.yaml
```

### Step 2: Install Ingress Controller

```batch
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml
```

### Step 3: Create Namespaces & Configs

```batch
kubectl apply -f k8s\base\namespace.yaml
kubectl apply -f k8s\base\configmap.yaml
kubectl apply -f k8s\base\secrets.yaml
```

### Step 4: Deploy Data Layer

```batch
kubectl apply -f k8s\data\redis.yaml
kubectl apply -f k8s\data\neo4j.yaml
kubectl apply -f k8s\data\minio.yaml
kubectl apply -f k8s\data\kafka.yaml
```

### Step 5: Build & Load Images

```batch
# Build backend
cd backend
docker build -t florus-backend:local .
kind load docker-image florus-backend:local --name florus-local

# Build frontend
cd frontend
docker build -t florus-frontend:local .
kind load docker-image florus-frontend:local --name florus-local
```

### Step 6: Deploy Applications

```batch
kubectl apply -f k8s\apps\backend.yaml
kubectl apply -f k8s\apps\frontend.yaml
kubectl apply -f k8s\apps\ingress.yaml
```

## Access Points

| Service | URL | Credentials |
|---------|-----|-------------|
| Frontend | http://localhost:30080 | - |
| Backend API | http://localhost:30081/api/health | - |
| Neo4j Browser | http://localhost:30474 | neo4j / florusneo4j123 |
| MinIO Console | http://localhost:30901 | minioadmin / minioadmin123 |

## Using Ingress (Optional)

Add to `C:\Windows\System32\drivers\etc\hosts`:
```
127.0.0.1 florus.local api.florus.local
```

Then access:
- Frontend: http://florus.local
- API: http://api.florus.local

## Useful Commands

```batch
# View all pods
kubectl get pods -A

# View logs
kubectl logs -n florus deployment/backend -f

# Shell into pod
kubectl exec -it -n florus deployment/backend -- sh

# Port forward (alternative to NodePort)
kubectl port-forward -n florus svc/backend-service 8081:8081

# Scale deployment
kubectl scale -n florus deployment/backend --replicas=3

# Delete and recreate pod
kubectl delete pod -n florus -l app=backend
```

## Troubleshooting

### Pods not starting
```batch
kubectl describe pod -n florus <pod-name>
kubectl logs -n florus <pod-name>
```

### Image not found
```batch
# Rebuild and reload image
docker build -t florus-backend:local ./backend
kind load docker-image florus-backend:local --name florus-local
kubectl rollout restart -n florus deployment/backend
```

### Database connection issues
```batch
# Check if data services are running
kubectl get pods -n florus-data

# Check service endpoints
kubectl get endpoints -n florus-data
```

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                     Kind Cluster (florus-local)                  │
│                                                                  │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │  Namespace: florus                                       │   │
│  │  ┌─────────┐  ┌─────────┐  ┌─────────┐                 │   │
│  │  │Frontend │  │ Backend │  │ Backend │                 │   │
│  │  │  Pod    │  │  Pod 1  │  │  Pod 2  │                 │   │
│  │  └────┬────┘  └────┬────┘  └────┬────┘                 │   │
│  │       └────────────┴────────────┘                       │   │
│  │                    │                                     │   │
│  │              ┌─────┴─────┐                              │   │
│  │              │  Services │                              │   │
│  │              └───────────┘                              │   │
│  └─────────────────────────────────────────────────────────┘   │
│                           │                                      │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │  Namespace: florus-data                                  │   │
│  │  ┌───────┐ ┌───────┐ ┌───────┐ ┌───────┐ ┌───────┐    │   │
│  │  │ Redis │ │ Neo4j │ │ MinIO │ │ Kafka │ │  ZK   │    │   │
│  │  └───────┘ └───────┘ └───────┘ └───────┘ └───────┘    │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                  │
│  NodePorts: 30080 (Frontend), 30081 (Backend),                  │
│             30474 (Neo4j), 30901 (MinIO)                        │
└─────────────────────────────────────────────────────────────────┘
```

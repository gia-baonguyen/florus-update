# Hướng dẫn CI/CD cho Florus - Dành cho người mới bắt đầu

## Mục lục
1. [Tổng quan CI/CD](#1-tổng-quan-cicd)
2. [Công nghệ sử dụng](#2-công-nghệ-sử-dụng)
3. [Cấu trúc thư mục và giải thích files](#3-cấu-trúc-thư-mục-và-giải-thích-files)
4. [Luồng CI/CD chi tiết](#4-luồng-cicd-chi-tiết)
5. [Kiến trúc Kubernetes Cluster](#5-kiến-trúc-kubernetes-cluster)
6. [Phân bổ workload trên cluster thực tế](#6-phân-bổ-workload-trên-cluster-thực-tế)
7. [ArgoCD và GitOps](#7-argocd-và-gitops)
8. [Hướng dẫn sử dụng](#8-hướng-dẫn-sử-dụng)
9. [Troubleshooting](#9-troubleshooting)

---

## 1. Tổng quan CI/CD

### CI/CD là gì?

**CI (Continuous Integration - Tích hợp liên tục):**
- Là quá trình tự động kiểm tra và build code mỗi khi developer push code lên repository
- Mục đích: Phát hiện lỗi sớm, đảm bảo code mới không làm hỏng code cũ

**CD (Continuous Deployment/Delivery - Triển khai liên tục):**
- Là quá trình tự động deploy ứng dụng lên server sau khi CI thành công
- Mục đích: Giảm thời gian từ code đến production, giảm lỗi do deploy thủ công

### Tại sao cần CI/CD?

```
Không có CI/CD:                    Có CI/CD:

Developer viết code                Developer viết code
      ↓                                  ↓
Developer test thủ công            git push (tự động)
      ↓                                  ↓
Developer build thủ công           CI tự động test & build
      ↓                                  ↓
DevOps deploy thủ công             CD tự động deploy
      ↓                                  ↓
Dễ có lỗi, mất nhiều thời gian     Nhanh, ít lỗi, nhất quán
```

---

## 2. Công nghệ sử dụng

| Công nghệ | Vai trò | Giải thích đơn giản |
|-----------|---------|---------------------|
| **GitHub Actions** | CI Pipeline | "Robot" tự động chạy khi bạn push code |
| **Docker** | Container | "Hộp" chứa ứng dụng và mọi thứ nó cần |
| **ghcr.io** | Container Registry | "Kho" lưu trữ các Docker images |
| **Kubernetes (K8s)** | Orchestration | "Quản lý" chạy và scale các containers |
| **ArgoCD** | CD/GitOps | "Robot" theo dõi Git và tự động deploy |
| **Kustomize** | Config Management | "Công cụ" tùy chỉnh config cho từng môi trường |
| **Kind** | Local K8s | Chạy Kubernetes trên máy local để test |

---

## 3. Cấu trúc thư mục và giải thích files

```
florus_datn/
│
├── .github/
│   └── workflows/
│       ├── ci.yml              # [1] CI Pipeline - Build & Push images
│       └── cd.yml              # [2] CD Pipeline - Deploy (optional)
│
├── k8s/                        # Kubernetes configurations
│   │
│   ├── base/                   # [3] Base configs - dùng chung
│   │   ├── configmap.yaml      # Biến môi trường (PORT, DB_HOST,...)
│   │   ├── secret.yaml         # Thông tin nhạy cảm (passwords)
│   │   └── kustomization.yaml  # Khai báo resources
│   │
│   ├── apps/                   # [4] Application deployments
│   │   ├── backend/
│   │   │   ├── deployment.yaml # Cách chạy backend (replicas, image,...)
│   │   │   ├── service.yaml    # Network cho backend
│   │   │   └── hpa.yaml        # Auto-scaling rules
│   │   ├── frontend/
│   │   │   ├── deployment.yaml
│   │   │   └── service.yaml
│   │   └── kustomization.yaml
│   │
│   ├── overlays/               # [5] Environment-specific configs
│   │   ├── staging/
│   │   │   ├── kustomization.yaml  # Override cho staging
│   │   │   └── namespace.yaml      # Namespace riêng
│   │   └── production/
│   │       ├── kustomization.yaml  # Override cho production
│   │       └── namespace.yaml
│   │
│   ├── argocd/                 # [6] ArgoCD Application definitions
│   │   ├── app-staging.yaml    # ArgoCD app cho staging
│   │   └── app-production.yaml # ArgoCD app cho production
│   │
│   └── local/                  # [7] Local development
│       └── kind-config.yaml    # Kind cluster configuration
│
├── backend/
│   ├── Dockerfile              # [8] Build instructions cho backend
│   └── ...
│
├── frontend/
│   ├── Dockerfile              # [9] Build instructions cho frontend
│   └── ...
│
└── docs/
    └── README-CICD.md          # File này
```

### Chi tiết từng file:

#### [1] `.github/workflows/ci.yml` - CI Pipeline

```yaml
# Khi nào chạy?
on:
  push:
    branches: [main, develop]    # Push vào main hoặc develop
    paths:
      - 'backend/**'             # Chỉ khi thay đổi backend
      - 'frontend/**'            # Hoặc frontend

# Làm gì?
jobs:
  backend-build:
    steps:
      - Build Docker image       # Đóng gói ứng dụng
      - Push to ghcr.io          # Upload lên registry
      - Update kustomization     # Cập nhật image tag mới
      - Commit changes           # Commit để ArgoCD detect
```

**Nhiệm vụ:**
- Tự động build Docker image khi có code mới
- Push image lên GitHub Container Registry (ghcr.io)
- Update file kustomization.yaml với image tag mới
- Commit lại để ArgoCD biết có version mới

#### [2] `.github/workflows/cd.yml` - CD Pipeline (Optional)

**Nhiệm vụ:**
- Deploy trực tiếp nếu không dùng ArgoCD
- Có thể trigger manual cho production

#### [3] `k8s/base/` - Base Configurations

**Nhiệm vụ:**
- Chứa configs dùng chung cho mọi môi trường
- Staging và Production sẽ kế thừa và override

```yaml
# base/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: florus-config
data:
  SERVER_PORT: "8081"      # Port chung
  SERVER_MODE: "release"   # Mode mặc định
```

#### [4] `k8s/apps/` - Application Deployments

**deployment.yaml** - Định nghĩa cách chạy ứng dụng:
```yaml
spec:
  replicas: 2                    # Chạy 2 bản copy
  template:
    spec:
      containers:
      - name: backend
        image: florus-backend    # Image sẽ được override
        ports:
        - containerPort: 8081
        resources:
          limits:
            memory: "512Mi"      # Giới hạn RAM
            cpu: "500m"          # Giới hạn CPU
```

**service.yaml** - Network cho ứng dụng:
```yaml
spec:
  type: ClusterIP              # Internal network
  ports:
  - port: 8081                 # Port bên trong cluster
```

#### [5] `k8s/overlays/` - Environment Overlays

**staging/kustomization.yaml:**
```yaml
namespace: florus-staging      # Namespace riêng

images:
- name: florus-backend
  newName: ghcr.io/user/florus-backend
  newTag: v1752d3d            # Tag cụ thể (được CI update)

replicas:
- name: backend
  count: 1                    # Staging chỉ cần 1 replica

patches:
- target:
    kind: Deployment
  patch: |
    - op: replace
      path: /spec/.../memory
      value: 256Mi            # Ít RAM hơn production
```

**production/kustomization.yaml:**
```yaml
namespace: florus-production

replicas:
- name: backend
  count: 3                    # Production cần 3 replicas

# Resources cao hơn, configs khác,...
```

#### [6] `k8s/argocd/` - ArgoCD Applications

**app-staging.yaml:**
```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: florus-staging
  namespace: argocd
spec:
  source:
    repoURL: https://github.com/user/repo.git  # Repo để watch
    targetRevision: main                        # Branch
    path: k8s/overlays/staging                  # Folder chứa config

  destination:
    server: https://kubernetes.default.svc      # Deploy vào cluster nào
    namespace: florus-staging

  syncPolicy:
    automated:                 # TỰ ĐỘNG sync
      prune: true              # Xóa resources không còn trong Git
      selfHeal: true           # Tự fix nếu ai đó sửa manual
```

#### [7] `k8s/local/kind-config.yaml` - Local Cluster

```yaml
kind: Cluster
nodes:
- role: control-plane          # Node quản lý
  extraPortMappings:
  - containerPort: 30080       # Map port ra localhost
    hostPort: 30080
- role: worker                 # Node chạy workload
- role: worker                 # Thêm 1 worker nữa
```

#### [8] & [9] `Dockerfile` - Container Build

```dockerfile
# Stage 1: Build
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o main ./cmd/api

# Stage 2: Run (image nhỏ hơn)
FROM alpine:latest
COPY --from=builder /app/main .
CMD ["./main"]
```

---

## 4. Luồng CI/CD chi tiết

### Sơ đồ tổng quan:

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           DEVELOPER                                      │
│                                                                          │
│    1. Sửa code (ví dụ: thêm feature mới)                                │
│    2. git commit -m "feat: add new feature"                              │
│    3. git push origin main                                               │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                        GITHUB ACTIONS (CI)                               │
│                                                                          │
│    4. Detect push → Trigger CI workflow                                  │
│    5. Checkout code                                                      │
│    6. Build Docker image                                                 │
│    7. Push image to ghcr.io (tag: v{commit_sha})                        │
│    8. Update k8s/overlays/staging/kustomization.yaml với tag mới        │
│    9. git commit & push (by github-actions bot)                          │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                            GITHUB REPO                                   │
│                                                                          │
│    kustomization.yaml đã được update với image tag mới                   │
│    Ví dụ: newTag: v1752d3d → newTag: v8d6e058                           │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                         ARGOCD (CD/GitOps)                               │
│                                                                          │
│    10. Poll Git repo mỗi 3 phút (hoặc webhook)                          │
│    11. Detect: "kustomization.yaml changed!"                             │
│    12. Status: OutOfSync → Trigger auto-sync                             │
│    13. Render manifests với Kustomize                                    │
│    14. Apply to Kubernetes cluster                                       │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                       KUBERNETES CLUSTER                                 │
│                                                                          │
│    15. Receive new Deployment spec                                       │
│    16. Pull image từ ghcr.io                                             │
│    17. Create new pods với image mới                                     │
│    18. Rolling update: dừng pods cũ, chạy pods mới                       │
│    19. Health check → Ready!                                             │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                              USERS                                       │
│                                                                          │
│    20. Truy cập application → Thấy version mới!                          │
│        http://staging.florus.com/api/health                              │
│        → {"version": "1.4.0"}                                            │
└─────────────────────────────────────────────────────────────────────────┘
```

### Chi tiết từng bước:

#### Bước 1-3: Developer Push Code
```bash
# Developer sửa file
vim backend/internal/handler/health_handler.go
# Thay đổi version từ "1.3.0" → "1.4.0"

# Commit và push
git add .
git commit -m "feat: update version to 1.4.0"
git push origin main
```

#### Bước 4-9: GitHub Actions CI

CI workflow tự động chạy:

```
Run 1: Detect Changes
├── Check paths: backend/** changed? ✓
└── Output: backend=true

Run 2: Backend Build & Push
├── Setup Docker Buildx
├── Login to ghcr.io
├── Build image: docker build -t ghcr.io/user/florus-backend:v8d6e058 .
├── Push: docker push ghcr.io/user/florus-backend:v8d6e058
├── Update kustomization.yaml:
│   └── sed -i "s|newTag: v.*|newTag: v8d6e058|" kustomization.yaml
└── Commit: git commit -m "ci: update backend image to v8d6e058"
```

#### Bước 10-14: ArgoCD Sync

```
ArgoCD Controller (chạy trong cluster):
│
├── Poll https://github.com/user/repo.git
├── Compare:
│   ├── Git state:  newTag: v8d6e058
│   └── Live state: newTag: v1752d3d
│   └── Result: OutOfSync!
│
├── Auto-sync triggered
│   ├── kustomize build k8s/overlays/staging
│   └── kubectl apply -f -
│
└── Status: Synced ✓
```

#### Bước 15-19: Kubernetes Rolling Update

```
Deployment Controller:
│
├── Detect: Deployment spec changed (new image tag)
├── Create ReplicaSet: backend-v8d6e058
│   └── Create Pod: backend-v8d6e058-xyz
│       ├── Pull image: ghcr.io/user/florus-backend:v8d6e058
│       ├── Start container
│       └── Health check: /api/health → 200 OK
│
├── Pod ready → Scale down old ReplicaSet
│   └── Terminate Pod: backend-v1752d3d-abc
│
└── Rollout complete!
```

---

## 5. Kiến trúc Kubernetes Cluster

### Các thành phần chính:

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        KUBERNETES CLUSTER                                │
│                                                                          │
│  ┌────────────────────────────────────────────────────────────────────┐ │
│  │                      CONTROL PLANE (Master)                        │ │
│  │                                                                    │ │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────────┐ │ │
│  │  │  API Server  │  │  Scheduler   │  │  Controller Manager      │ │ │
│  │  │              │  │              │  │                          │ │ │
│  │  │ Nhận lệnh    │  │ Quyết định   │  │ Đảm bảo trạng thái      │ │ │
│  │  │ từ kubectl   │  │ pod chạy     │  │ mong muốn               │ │ │
│  │  │ và ArgoCD    │  │ ở node nào   │  │                          │ │ │
│  │  └──────────────┘  └──────────────┘  └──────────────────────────┘ │ │
│  │                                                                    │ │
│  │  ┌──────────────────────────────────────────────────────────────┐ │ │
│  │  │                          etcd                                 │ │ │
│  │  │            Database lưu trữ state của cluster                 │ │ │
│  │  └──────────────────────────────────────────────────────────────┘ │ │
│  └────────────────────────────────────────────────────────────────────┘ │
│                                                                          │
│  ┌─────────────────────────┐  ┌─────────────────────────┐               │
│  │      WORKER NODE 1      │  │      WORKER NODE 2      │               │
│  │                         │  │                         │               │
│  │  ┌───────────────────┐  │  │  ┌───────────────────┐  │               │
│  │  │      kubelet      │  │  │  │      kubelet      │  │               │
│  │  │  Quản lý pods     │  │  │  │  Quản lý pods     │  │               │
│  │  └───────────────────┘  │  │  └───────────────────┘  │               │
│  │                         │  │                         │               │
│  │  ┌─────────┐ ┌───────┐  │  │  ┌─────────┐ ┌───────┐  │               │
│  │  │ Pod 1   │ │ Pod 2 │  │  │  │ Pod 3   │ │ Pod 4 │  │               │
│  │  │ Backend │ │ Redis │  │  │  │ Backend │ │ Front │  │               │
│  │  └─────────┘ └───────┘  │  │  └─────────┘ └───────┘  │               │
│  │                         │  │                         │               │
│  └─────────────────────────┘  └─────────────────────────┘               │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### Giải thích các khái niệm:

#### Pod
- Đơn vị nhỏ nhất trong K8s
- Chứa 1 hoặc nhiều containers
- Ví dụ: 1 pod backend = 1 container chạy Go app

```yaml
# Pod đơn giản
apiVersion: v1
kind: Pod
metadata:
  name: backend-pod
spec:
  containers:
  - name: backend
    image: ghcr.io/user/florus-backend:v1.0.0
    ports:
    - containerPort: 8081
```

#### Deployment
- Quản lý nhiều pods giống nhau
- Đảm bảo luôn có đủ số replicas
- Hỗ trợ rolling update

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: backend
spec:
  replicas: 3          # Luôn chạy 3 pods
  selector:
    matchLabels:
      app: backend
  template:            # Template cho mỗi pod
    spec:
      containers:
      - name: backend
        image: florus-backend:latest
```

#### Service
- Network endpoint cho pods
- Load balance giữa các pods
- Các loại:
  - **ClusterIP**: Chỉ truy cập từ trong cluster
  - **NodePort**: Mở port trên mỗi node
  - **LoadBalancer**: Tạo external load balancer

```yaml
apiVersion: v1
kind: Service
metadata:
  name: backend-service
spec:
  type: ClusterIP
  selector:
    app: backend       # Route traffic đến pods có label này
  ports:
  - port: 8081         # Port của service
    targetPort: 8081   # Port của container
```

#### Namespace
- Phân chia cluster thành các "phòng" riêng
- Mỗi môi trường có namespace riêng

```
Cluster
├── namespace: florus-staging
│   ├── deployment/backend (1 replica)
│   └── deployment/frontend (1 replica)
│
├── namespace: florus-production
│   ├── deployment/backend (3 replicas)
│   └── deployment/frontend (2 replicas)
│
└── namespace: argocd
    └── ArgoCD components
```

---

## 6. Phân bổ workload trên cluster thực tế

### Kiến trúc cluster production đề xuất:

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    PRODUCTION CLUSTER ARCHITECTURE                       │
│                                                                          │
│  ┌─────────────────────────────────────────────────────────────────────┐│
│  │                     CONTROL PLANE (3 nodes)                         ││
│  │                                                                     ││
│  │  ┌───────────────┐ ┌───────────────┐ ┌───────────────┐             ││
│  │  │   Master 1    │ │   Master 2    │ │   Master 3    │             ││
│  │  │   (Active)    │ │   (Standby)   │ │   (Standby)   │             ││
│  │  │               │ │               │ │               │             ││
│  │  │ • API Server  │ │ • API Server  │ │ • API Server  │             ││
│  │  │ • Scheduler   │ │ • Scheduler   │ │ • Scheduler   │             ││
│  │  │ • Controller  │ │ • Controller  │ │ • Controller  │             ││
│  │  │ • etcd        │ │ • etcd        │ │ • etcd        │             ││
│  │  │               │ │               │ │               │             ││
│  │  │ 4 CPU         │ │ 4 CPU         │ │ 4 CPU         │             ││
│  │  │ 16GB RAM      │ │ 16GB RAM      │ │ 16GB RAM      │             ││
│  │  └───────────────┘ └───────────────┘ └───────────────┘             ││
│  └─────────────────────────────────────────────────────────────────────┘│
│                                                                          │
│  ┌─────────────────────────────────────────────────────────────────────┐│
│  │                     WORKER NODES - APPLICATION                      ││
│  │                                                                     ││
│  │  ┌───────────────┐ ┌───────────────┐ ┌───────────────┐             ││
│  │  │   Worker 1    │ │   Worker 2    │ │   Worker 3    │             ││
│  │  │  (App Node)   │ │  (App Node)   │ │  (App Node)   │             ││
│  │  │               │ │               │ │               │             ││
│  │  │ ┌───────────┐ │ │ ┌───────────┐ │ │ ┌───────────┐ │             ││
│  │  │ │  Backend  │ │ │ │  Backend  │ │ │ │  Backend  │ │             ││
│  │  │ │  Pod 1    │ │ │ │  Pod 2    │ │ │ │  Pod 3    │ │             ││
│  │  │ └───────────┘ │ │ └───────────┘ │ │ └───────────┘ │             ││
│  │  │ ┌───────────┐ │ │ ┌───────────┐ │ │               │             ││
│  │  │ │ Frontend  │ │ │ │ Frontend  │ │ │               │             ││
│  │  │ │  Pod 1    │ │ │ │  Pod 2    │ │ │               │             ││
│  │  │ └───────────┘ │ │ └───────────┘ │ │               │             ││
│  │  │               │ │               │ │               │             ││
│  │  │ 8 CPU         │ │ 8 CPU         │ │ 8 CPU         │             ││
│  │  │ 32GB RAM      │ │ 32GB RAM      │ │ 32GB RAM      │             ││
│  │  │ Labels:       │ │ Labels:       │ │ Labels:       │             ││
│  │  │ type=app      │ │ type=app      │ │ type=app      │             ││
│  │  └───────────────┘ └───────────────┘ └───────────────┘             ││
│  └─────────────────────────────────────────────────────────────────────┘│
│                                                                          │
│  ┌─────────────────────────────────────────────────────────────────────┐│
│  │                     WORKER NODES - DATA/INFRA                       ││
│  │                                                                     ││
│  │  ┌───────────────┐ ┌───────────────┐ ┌───────────────┐             ││
│  │  │   Worker 4    │ │   Worker 5    │ │   Worker 6    │             ││
│  │  │ (Data Node)   │ │ (Data Node)   │ │ (Infra Node)  │             ││
│  │  │               │ │               │ │               │             ││
│  │  │ ┌───────────┐ │ │ ┌───────────┐ │ │ ┌───────────┐ │             ││
│  │  │ │   Redis   │ │ │ │   Redis   │ │ │ │  ArgoCD   │ │             ││
│  │  │ │  Master   │ │ │ │  Replica  │ │ │ │           │ │             ││
│  │  │ └───────────┘ │ │ └───────────┘ │ │ └───────────┘ │             ││
│  │  │ ┌───────────┐ │ │ ┌───────────┐ │ │ ┌───────────┐ │             ││
│  │  │ │   Neo4j   │ │ │ │   Kafka   │ │ │ │Prometheus │ │             ││
│  │  │ │           │ │ │ │           │ │ │ │  Grafana  │ │             ││
│  │  │ └───────────┘ │ │ └───────────┘ │ │ └───────────┘ │             ││
│  │  │               │ │               │ │               │             ││
│  │  │ 8 CPU         │ │ 8 CPU         │ │ 4 CPU         │             ││
│  │  │ 64GB RAM      │ │ 64GB RAM      │ │ 16GB RAM      │             ││
│  │  │ SSD 500GB     │ │ SSD 500GB     │ │ SSD 100GB     │             ││
│  │  │ Labels:       │ │ Labels:       │ │ Labels:       │             ││
│  │  │ type=data     │ │ type=data     │ │ type=infra    │             ││
│  │  └───────────────┘ └───────────────┘ └───────────────┘             ││
│  └─────────────────────────────────────────────────────────────────────┘│
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### Node Types và vai trò:

| Node Type | Số lượng | CPU | RAM | Storage | Chạy gì |
|-----------|----------|-----|-----|---------|---------|
| **Master/Control Plane** | 3 (HA) | 4 | 16GB | SSD 100GB | API Server, Scheduler, etcd |
| **App Worker** | 3+ | 8 | 32GB | SSD 100GB | Backend, Frontend pods |
| **Data Worker** | 2+ | 8 | 64GB | SSD 500GB | Redis, Neo4j, Kafka, MinIO |
| **Infra Worker** | 1-2 | 4 | 16GB | SSD 100GB | ArgoCD, Prometheus, Grafana |

### Node Selectors và Affinity:

```yaml
# Backend deployment - chạy trên app nodes
apiVersion: apps/v1
kind: Deployment
metadata:
  name: backend
spec:
  template:
    spec:
      nodeSelector:
        type: app                    # Chỉ chạy trên nodes có label type=app

      affinity:
        podAntiAffinity:             # Không chạy 2 pods trên cùng 1 node
          preferredDuringScheduling:
            - podAffinityTerm:
                labelSelector:
                  matchLabels:
                    app: backend
                topologyKey: kubernetes.io/hostname
```

```yaml
# Redis - chạy trên data nodes
apiVersion: apps/v1
kind: Deployment
metadata:
  name: redis
spec:
  template:
    spec:
      nodeSelector:
        type: data                   # Chạy trên data nodes

      tolerations:                   # Chấp nhận taint của data nodes
      - key: "dedicated"
        operator: "Equal"
        value: "data"
        effect: "NoSchedule"
```

### Taints và Tolerations:

```bash
# Đánh dấu node chỉ dành cho data workloads
kubectl taint nodes worker-4 dedicated=data:NoSchedule
kubectl taint nodes worker-5 dedicated=data:NoSchedule

# Đánh dấu node cho infrastructure
kubectl taint nodes worker-6 dedicated=infra:NoSchedule
```

### Resource Quotas per Namespace:

```yaml
# Giới hạn resources cho staging
apiVersion: v1
kind: ResourceQuota
metadata:
  name: staging-quota
  namespace: florus-staging
spec:
  hard:
    requests.cpu: "4"           # Tổng CPU requests
    requests.memory: 8Gi        # Tổng memory requests
    limits.cpu: "8"             # Tổng CPU limits
    limits.memory: 16Gi         # Tổng memory limits
    pods: "20"                  # Tối đa 20 pods
```

```yaml
# Resources cho production cao hơn
apiVersion: v1
kind: ResourceQuota
metadata:
  name: production-quota
  namespace: florus-production
spec:
  hard:
    requests.cpu: "16"
    requests.memory: 32Gi
    limits.cpu: "32"
    limits.memory: 64Gi
    pods: "100"
```

---

## 7. ArgoCD và GitOps

### GitOps là gì?

**GitOps** = Git + Operations
- Git là "single source of truth" (nguồn sự thật duy nhất)
- Mọi thay đổi infrastructure phải qua Git
- Tự động sync từ Git → Cluster

```
Truyền thống:                      GitOps:

Developer → kubectl apply          Developer → git push
    ↓                                  ↓
Cluster thay đổi                   Git repo thay đổi
    ↓                                  ↓
Không ai biết ai đã thay đổi gì    ArgoCD detect & sync
                                       ↓
                                   Cluster thay đổi
                                       ↓
                                   Có audit trail đầy đủ
```

### ArgoCD Components:

```
┌─────────────────────────────────────────────────────────────────────────┐
│                            ARGOCD ARCHITECTURE                          │
│                                                                          │
│  ┌─────────────────────────────────────────────────────────────────────┐│
│  │                         argocd namespace                            ││
│  │                                                                     ││
│  │  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐     ││
│  │  │  argocd-server  │  │ argocd-repo-srv │  │argocd-app-ctrl │     ││
│  │  │                 │  │                 │  │                 │     ││
│  │  │ • Web UI        │  │ • Clone repos   │  │ • Watch apps    │     ││
│  │  │ • API server    │  │ • Cache repos   │  │ • Compare state │     ││
│  │  │ • Auth          │  │ • Generate      │  │ • Sync clusters │     ││
│  │  │                 │  │   manifests     │  │                 │     ││
│  │  └─────────────────┘  └─────────────────┘  └─────────────────┘     ││
│  │           │                    │                    │               ││
│  │           └────────────────────┼────────────────────┘               ││
│  │                                │                                    ││
│  │                    ┌───────────▼───────────┐                       ││
│  │                    │     argocd-redis      │                       ││
│  │                    │   (Cache & Queue)     │                       ││
│  │                    └───────────────────────┘                       ││
│  │                                                                     ││
│  └─────────────────────────────────────────────────────────────────────┘│
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                    ┌───────────────┼───────────────┐
                    │               │               │
                    ▼               ▼               ▼
              ┌──────────┐   ┌──────────┐   ┌──────────┐
              │  GitHub  │   │ Staging  │   │Production│
              │   Repo   │   │ Namespace│   │Namespace │
              └──────────┘   └──────────┘   └──────────┘
```

### ArgoCD Application Lifecycle:

```
1. CREATE APPLICATION
   └── kubectl apply -f app-staging.yaml
       │
       ▼
2. INITIAL SYNC
   ├── ArgoCD clones repo
   ├── Runs: kustomize build k8s/overlays/staging
   ├── Compares with live state (empty)
   └── Applies all resources
       │
       ▼
3. RUNNING STATE
   ├── Status: Synced, Healthy
   ├── ArgoCD polls Git every 3 minutes
   └── Watches for drift
       │
       ▼
4. CHANGE DETECTED
   ├── Git: newTag: v1.0.0 → v1.1.0
   ├── Status: OutOfSync
   └── If automated: trigger sync
       │
       ▼
5. SYNC IN PROGRESS
   ├── Apply new manifests
   ├── K8s creates new pods
   └── Rolling update
       │
       ▼
6. SYNC COMPLETE
   └── Status: Synced, Healthy
```

### Sync Policies:

```yaml
syncPolicy:
  automated:
    prune: true        # Xóa resources không còn trong Git
    selfHeal: true     # Tự động fix nếu ai đó sửa manual trong cluster
    allowEmpty: false  # Không cho phép delete tất cả resources

  syncOptions:
  - CreateNamespace=true              # Tự tạo namespace nếu chưa có
  - PrunePropagationPolicy=foreground # Đợi child resources xóa xong
  - PruneLast=true                    # Xóa resources cũ sau cùng

  retry:
    limit: 5           # Retry 5 lần nếu sync fail
    backoff:
      duration: 5s
      factor: 2
      maxDuration: 3m
```

### Manual vs Automated Sync:

| Aspect | Manual Sync | Automated Sync |
|--------|-------------|----------------|
| Trigger | Người dùng click "Sync" | Tự động khi Git thay đổi |
| Use case | Production (cần review) | Staging (nhanh gọn) |
| Control | Cao | Thấp |
| Speed | Chậm | Nhanh |
| Config | `automated: {}` bỏ trống | `automated: { prune: true }` |

---

## 8. Hướng dẫn sử dụng

### 8.1. Setup Local Development

#### Bước 1: Cài đặt tools

```bash
# Docker Desktop
# Download từ: https://www.docker.com/products/docker-desktop

# Kind (Kubernetes in Docker)
# Windows (PowerShell)
choco install kind

# Hoặc download binary từ: https://kind.sigs.k8s.io/

# kubectl
choco install kubernetes-cli

# Verify
docker --version
kind --version
kubectl version --client
```

#### Bước 2: Tạo Kind cluster

```bash
# Tạo cluster với config
kind create cluster --name florus-local --config k8s/local/kind-config.yaml

# Verify
kubectl cluster-info
kubectl get nodes
```

#### Bước 3: Build và load images

```bash
# Build backend
docker build -t florus-backend:local ./backend

# Build frontend
docker build -t florus-frontend:local ./frontend

# Load vào Kind
kind load docker-image florus-backend:local --name florus-local
kind load docker-image florus-frontend:local --name florus-local
```

#### Bước 4: Deploy ứng dụng

```bash
# Deploy staging
kubectl apply -k k8s/overlays/staging

# Check status
kubectl get pods -n florus-staging
kubectl get services -n florus-staging

# Test
curl http://localhost:30181/api/health
```

### 8.2. Install ArgoCD

```bash
# Create namespace
kubectl create namespace argocd

# Install ArgoCD
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml

# Wait for pods
kubectl wait --for=condition=Ready pods --all -n argocd --timeout=300s

# Expose UI (NodePort)
kubectl patch svc argocd-server -n argocd -p '{"spec": {"type": "NodePort", "ports": [{"port": 443, "nodePort": 30443}]}}'

# Get admin password
kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d

# Access UI
# https://localhost:30443
# Username: admin
# Password: (từ command trên)
```

### 8.3. Create ArgoCD Applications

```bash
# Apply staging app
kubectl apply -f k8s/argocd/app-staging.yaml

# Apply production app
kubectl apply -f k8s/argocd/app-production.yaml

# Check applications
kubectl get applications -n argocd
```

### 8.4. Test CI/CD Flow

```bash
# 1. Make a code change
vim backend/internal/handler/health_handler.go
# Change version: "1.3.0" → "1.4.0"

# 2. Commit and push
git add .
git commit -m "feat: update version to 1.4.0"
git push origin main

# 3. Watch CI on GitHub
# Go to: https://github.com/YOUR_REPO/actions

# 4. Wait for ArgoCD to sync (or trigger manually)
kubectl patch application florus-staging -n argocd \
  --type merge -p '{"metadata":{"annotations":{"argocd.argoproj.io/refresh":"hard"}}}'

# 5. Check sync status
kubectl get applications -n argocd

# 6. Verify deployment
curl http://localhost:30181/api/health
# Should show: {"version": "1.4.0"}
```

---

## 9. Troubleshooting

### Common Issues:

#### 1. Pod stuck in "Pending"

```bash
# Check events
kubectl describe pod <pod-name> -n <namespace>

# Common causes:
# - Insufficient resources
# - Node selector không match
# - PVC không bind được

# Solution:
kubectl get nodes -o wide
kubectl describe node <node-name>
```

#### 2. Pod stuck in "ImagePullBackOff"

```bash
# Check image name
kubectl describe pod <pod-name> -n <namespace> | grep Image

# Common causes:
# - Image không tồn tại
# - Registry cần authentication
# - imagePullPolicy: Never nhưng image không có local

# Solution cho Kind (load image local):
kind load docker-image <image:tag> --name florus-local

# Solution cho registry (tạo secret):
kubectl create secret docker-registry regcred \
  --docker-server=ghcr.io \
  --docker-username=<username> \
  --docker-password=<token>
```

#### 3. ArgoCD OutOfSync nhưng không auto-sync

```bash
# Check application
kubectl get application <app-name> -n argocd -o yaml

# Verify syncPolicy.automated exists
# If manual sync needed:
kubectl patch application <app-name> -n argocd \
  --type merge -p '{"operation":{"initiatedBy":{"username":"admin"},"sync":{}}}'
```

#### 4. Service không accessible

```bash
# Check service
kubectl get svc -n <namespace>

# Check endpoints
kubectl get endpoints -n <namespace>

# For NodePort, verify port mapping:
kubectl get svc <service-name> -n <namespace> -o jsonpath='{.spec.ports[0].nodePort}'

# Test từ trong cluster:
kubectl run test --rm -it --image=curlimages/curl -- curl http://<service-name>:8081/api/health
```

#### 5. CI không trigger

```bash
# Check GitHub Actions tab
# Verify:
# 1. Workflow file exists: .github/workflows/ci.yml
# 2. Branch matches trigger (main, develop)
# 3. Path filters match changed files

# Check workflow syntax:
# Go to: https://github.com/YOUR_REPO/actions
# Click on workflow → "..." → "View workflow file"
```

### Useful Commands:

```bash
# Logs
kubectl logs <pod-name> -n <namespace>
kubectl logs <pod-name> -n <namespace> --previous  # Previous container

# Exec into pod
kubectl exec -it <pod-name> -n <namespace> -- /bin/sh

# Port forward (debug locally)
kubectl port-forward svc/<service-name> 8080:8081 -n <namespace>

# Restart deployment
kubectl rollout restart deployment/<deployment-name> -n <namespace>

# Rollback deployment
kubectl rollout undo deployment/<deployment-name> -n <namespace>

# Scale deployment
kubectl scale deployment/<deployment-name> --replicas=3 -n <namespace>

# View all resources in namespace
kubectl get all -n <namespace>

# ArgoCD CLI
argocd app list
argocd app get <app-name>
argocd app sync <app-name>
argocd app history <app-name>
```

---

## Tổng kết

### Key Takeaways:

1. **CI/CD tự động hóa quy trình** từ code → production
2. **GitOps đảm bảo** mọi thay đổi đều qua Git và có audit trail
3. **Kustomize cho phép** tùy chỉnh config cho từng môi trường
4. **ArgoCD liên tục sync** Git → Cluster
5. **Kubernetes orchestrate** containers với high availability

### Flow tóm tắt:

```
Code → Git Push → GitHub Actions (build) → ghcr.io (image)
    → Git (update manifest) → ArgoCD (detect) → K8s (deploy) → Users
```

### Best Practices:

- Luôn test trên staging trước khi deploy production
- Sử dụng semantic versioning cho images
- Giữ secrets trong Kubernetes Secrets hoặc external vault
- Monitor cluster với Prometheus/Grafana
- Backup etcd thường xuyên
- Document mọi thay đổi infrastructure

---

**Author:** Claude AI
**Last Updated:** 2026-03-11
**Version:** 1.0.0

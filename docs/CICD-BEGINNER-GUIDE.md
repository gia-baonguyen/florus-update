# Hướng dẫn CI/CD cho người mới bắt đầu

## Mục lục
1. [CI/CD là gì?](#1-cicd-là-gì)
2. [Tại sao cần CI/CD?](#2-tại-sao-cần-cicd)
3. [Các thành phần trong CI/CD](#3-các-thành-phần-trong-cicd)
4. [GitHub Actions là gì?](#4-github-actions-là-gì)
5. [ArgoCD là gì?](#5-argocd-là-gì)
6. [Luồng hoạt động chi tiết](#6-luồng-hoạt-động-chi-tiết)
7. [Giải thích từng file cấu hình](#7-giải-thích-từng-file-cấu-hình)
8. [Thực hành từng bước](#8-thực-hành-từng-bước)

---

## 1. CI/CD là gì?

### CI - Continuous Integration (Tích hợp liên tục)

**Định nghĩa đơn giản**: CI là quá trình tự động kiểm tra code mỗi khi developer push code lên.

```
Trước khi có CI:
─────────────────────────────────────────────────────────────────

Developer A viết code ──▶ Push lên Git ──▶ Không ai kiểm tra
                                                    │
Developer B viết code ──▶ Push lên Git ──▶ Không ai kiểm tra
                                                    │
                                                    ▼
                                          Merge tất cả vào main
                                                    │
                                                    ▼
                                          BUG! CONFLICT! LỖI!
                                          (Phát hiện quá muộn)


Sau khi có CI:
─────────────────────────────────────────────────────────────────

Developer A push code ──▶ CI tự động chạy ──▶ ✅ Tests pass ──▶ OK để merge
                              │
                              ├── Chạy lint (kiểm tra code style)
                              ├── Chạy unit tests
                              ├── Chạy integration tests
                              └── Build thử xem có lỗi không

Developer B push code ──▶ CI tự động chạy ──▶ ❌ Tests fail ──▶ Báo lỗi ngay
                                                                     │
                                                          Developer B sửa lỗi
```

### CD - Continuous Deployment (Triển khai liên tục)

**Định nghĩa đơn giản**: CD là quá trình tự động deploy code lên server sau khi CI pass.

```
Trước khi có CD:
─────────────────────────────────────────────────────────────────

Code đã merge ──▶ DevOps phải SSH vào server
                         │
                         ├── git pull
                         ├── npm install
                         ├── npm run build
                         ├── Restart service
                         └── Cầu nguyện không có bug 🙏

Vấn đề:
- Tốn thời gian (30 phút - 1 tiếng mỗi lần deploy)
- Dễ sai sót (quên bước, gõ sai lệnh)
- Không nhất quán (mỗi người deploy một kiểu)
- Khó rollback khi có lỗi


Sau khi có CD:
─────────────────────────────────────────────────────────────────

Code merge vào main ──▶ CD tự động ──▶ Build Docker image
                                              │
                                              ▼
                                       Push image lên Registry
                                              │
                                              ▼
                                       Deploy lên Kubernetes
                                              │
                                              ▼
                                       Health check
                                              │
                              ┌───────────────┴───────────────┐
                              │                               │
                              ▼                               ▼
                        ✅ Healthy                      ❌ Unhealthy
                        (Xong!)                         (Auto rollback)
```

---

## 2. Tại sao cần CI/CD?

### So sánh thực tế

| Tiêu chí | Không có CI/CD | Có CI/CD |
|----------|----------------|----------|
| **Thời gian deploy** | 30-60 phút | 5-10 phút |
| **Tần suất deploy** | 1-2 lần/tuần | Nhiều lần/ngày |
| **Phát hiện bug** | Sau khi deploy | Ngay khi push code |
| **Rollback** | 30 phút+ | 1-2 phút |
| **Rủi ro lỗi người** | Cao | Thấp |
| **Tài liệu quy trình** | Trong đầu DevOps | Trong code (version control) |

### Ví dụ thực tế

```
Tình huống: Team có 5 developers, mỗi ngày có 10 pull requests

KHÔNG CÓ CI/CD:
──────────────────────────────────────────────────────────────
- Reviewer phải checkout code, chạy tests thủ công
- Mỗi PR mất 30 phút để review + test
- Tổng: 10 PR × 30 phút = 5 giờ/ngày CHỈ ĐỂ TEST

CÓ CI/CD:
──────────────────────────────────────────────────────────────
- CI tự động chạy tests trong 5 phút
- Reviewer chỉ cần đọc code
- CI báo ✅ hoặc ❌ ngay trên GitHub
- Tiết kiệm: 4+ giờ/ngày
```

---

## 3. Các thành phần trong CI/CD

### Sơ đồ tổng quan

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           CI/CD PIPELINE                                     │
│                                                                              │
│  ┌──────────────────────────────────────────────────────────────────────┐  │
│  │                        SOURCE CODE                                    │  │
│  │                                                                       │  │
│  │   Developer ──▶ Git Push ──▶ GitHub Repository                       │  │
│  │                                                                       │  │
│  └──────────────────────────────────────────────────────────────────────┘  │
│                                    │                                         │
│                                    ▼                                         │
│  ┌──────────────────────────────────────────────────────────────────────┐  │
│  │                    CI - GITHUB ACTIONS                                │  │
│  │                                                                       │  │
│  │   ┌─────────┐    ┌─────────┐    ┌─────────┐    ┌─────────────────┐  │  │
│  │   │  LINT   │───▶│  TEST   │───▶│  BUILD  │───▶│  PUSH TO        │  │  │
│  │   │         │    │         │    │  IMAGE  │    │  REGISTRY       │  │  │
│  │   └─────────┘    └─────────┘    └─────────┘    └─────────────────┘  │  │
│  │                                                                       │  │
│  │   Kiểm tra       Chạy unit     Build Docker   Lưu image lên         │  │
│  │   code style     tests         image          ghcr.io/DockerHub      │  │
│  │                                                                       │  │
│  └──────────────────────────────────────────────────────────────────────┘  │
│                                    │                                         │
│                                    ▼                                         │
│  ┌──────────────────────────────────────────────────────────────────────┐  │
│  │                    CD - ARGOCD (GitOps)                               │  │
│  │                                                                       │  │
│  │   ┌─────────────┐    ┌─────────────┐    ┌─────────────────────────┐ │  │
│  │   │  WATCH GIT  │───▶│  COMPARE    │───▶│  SYNC TO KUBERNETES     │ │  │
│  │   │  REPO       │    │  CHANGES    │    │                         │ │  │
│  │   └─────────────┘    └─────────────┘    └─────────────────────────┘ │  │
│  │                                                                       │  │
│  │   Theo dõi thay     So sánh với       Deploy pods mới               │  │
│  │   đổi trong Git     cluster hiện tại  lên K8s cluster               │  │
│  │                                                                       │  │
│  └──────────────────────────────────────────────────────────────────────┘  │
│                                    │                                         │
│                                    ▼                                         │
│  ┌──────────────────────────────────────────────────────────────────────┐  │
│  │                       KUBERNETES CLUSTER                              │  │
│  │                                                                       │  │
│  │   ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐  │  │
│  │   │ Backend │  │ Backend │  │Frontend │  │  Redis  │  │  Neo4j  │  │  │
│  │   │  Pod 1  │  │  Pod 2  │  │   Pod   │  │   Pod   │  │   Pod   │  │  │
│  │   └─────────┘  └─────────┘  └─────────┘  └─────────┘  └─────────┘  │  │
│  │                                                                       │  │
│  └──────────────────────────────────────────────────────────────────────┘  │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Giải thích từng thành phần

#### 1. GitHub Repository
- **Là gì?**: Nơi lưu trữ source code
- **Vai trò**: Là "nguồn sự thật" (source of truth) của dự án

#### 2. GitHub Actions (CI)
- **Là gì?**: Dịch vụ CI/CD miễn phí của GitHub
- **Vai trò**: Tự động chạy tests, build code khi có push/PR
- **File cấu hình**: `.github/workflows/*.yml`

#### 3. Container Registry (ghcr.io)
- **Là gì?**: Nơi lưu trữ Docker images
- **Vai trò**: Lưu các phiên bản đã build của ứng dụng
- **Ví dụ**: `ghcr.io/your-org/florus-backend:v1.0.0`

#### 4. ArgoCD
- **Là gì?**: Công cụ CD theo mô hình GitOps
- **Vai trò**: Tự động deploy lên Kubernetes khi có thay đổi trong Git
- **Đặc điểm**: "Git là nguồn sự thật"

#### 5. Kubernetes Cluster
- **Là gì?**: Nền tảng orchestration container
- **Vai trò**: Chạy và quản lý các containers

---

## 4. GitHub Actions là gì?

### Khái niệm cơ bản

```
GitHub Actions = Workflow + Jobs + Steps

┌─────────────────────────────────────────────────────────────────┐
│                        WORKFLOW (ci.yml)                         │
│                                                                  │
│   Khi nào chạy? (trigger)                                       │
│   - push vào main/develop                                        │
│   - tạo pull request                                             │
│                                                                  │
│   ┌─────────────────────────────────────────────────────────┐  │
│   │                    JOB: backend-test                     │  │
│   │                    (Chạy trên: ubuntu-latest)            │  │
│   │                                                          │  │
│   │   Step 1: Checkout code                                  │  │
│   │           └── Tải source code về                         │  │
│   │                                                          │  │
│   │   Step 2: Setup Go                                       │  │
│   │           └── Cài đặt Go 1.21                            │  │
│   │                                                          │  │
│   │   Step 3: Run tests                                      │  │
│   │           └── go test ./...                              │  │
│   │                                                          │  │
│   └─────────────────────────────────────────────────────────┘  │
│                              │                                   │
│                              ▼ (chạy song song)                  │
│   ┌─────────────────────────────────────────────────────────┐  │
│   │                    JOB: frontend-test                    │  │
│   │                    (Chạy trên: ubuntu-latest)            │  │
│   │                                                          │  │
│   │   Step 1: Checkout code                                  │  │
│   │   Step 2: Setup Node.js                                  │  │
│   │   Step 3: npm install                                    │  │
│   │   Step 4: npm test                                       │  │
│   │                                                          │  │
│   └─────────────────────────────────────────────────────────┘  │
│                              │                                   │
│                              ▼ (chờ 2 jobs trên xong)            │
│   ┌─────────────────────────────────────────────────────────┐  │
│   │                    JOB: build-and-push                   │  │
│   │                    (needs: [backend-test, frontend-test])│  │
│   │                                                          │  │
│   │   Step 1: Build Docker image                             │  │
│   │   Step 2: Push to Registry                               │  │
│   │                                                          │  │
│   └─────────────────────────────────────────────────────────┘  │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### Giải thích file ci.yml

```yaml
# .github/workflows/ci.yml

# ═══════════════════════════════════════════════════════════════
# PHẦN 1: TÊN WORKFLOW
# ═══════════════════════════════════════════════════════════════
name: CI Pipeline    # Tên hiển thị trên GitHub

# ═══════════════════════════════════════════════════════════════
# PHẦN 2: KHI NÀO CHẠY? (Triggers)
# ═══════════════════════════════════════════════════════════════
on:
  push:
    branches: [main, develop]    # Chạy khi push vào main hoặc develop
  pull_request:
    branches: [main, develop]    # Chạy khi tạo PR vào main hoặc develop

# ═══════════════════════════════════════════════════════════════
# PHẦN 3: BIẾN MÔI TRƯỜNG (dùng chung cho tất cả jobs)
# ═══════════════════════════════════════════════════════════════
env:
  REGISTRY: ghcr.io                              # GitHub Container Registry
  BACKEND_IMAGE: ${{ github.repository }}/florus-backend   # Tên image

# ═══════════════════════════════════════════════════════════════
# PHẦN 4: CÁC JOBS
# ═══════════════════════════════════════════════════════════════
jobs:
  # ─────────────────────────────────────────────────────────────
  # JOB 1: Test Backend
  # ─────────────────────────────────────────────────────────────
  backend-test:
    name: Backend Tests              # Tên hiển thị
    runs-on: ubuntu-latest           # Chạy trên máy Ubuntu

    steps:
      # Bước 1: Tải source code về máy chạy CI
      - name: Checkout code
        uses: actions/checkout@v4    # Dùng action có sẵn của GitHub

      # Bước 2: Cài đặt Go
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.21'         # Phiên bản Go cần cài

      # Bước 3: Chạy tests
      - name: Run tests
        run: go test -v ./...        # Lệnh shell để chạy tests
        working-directory: ./backend # Chạy trong thư mục backend

  # ─────────────────────────────────────────────────────────────
  # JOB 2: Build và Push Docker Image
  # ─────────────────────────────────────────────────────────────
  backend-build:
    name: Backend Build & Push
    runs-on: ubuntu-latest
    needs: backend-test              # ⚠️ CHỈ CHẠY SAU KHI backend-test PASS

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      # Đăng nhập vào GitHub Container Registry
      - name: Login to Registry
        uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}         # Tên user GitHub
          password: ${{ secrets.GITHUB_TOKEN }} # Token tự động

      # Build và push Docker image
      - name: Build and push
        uses: docker/build-push-action@v5
        with:
          context: ./backend
          push: true
          tags: ghcr.io/my-org/florus-backend:latest
```

### Các khái niệm quan trọng

| Khái niệm | Giải thích | Ví dụ |
|-----------|------------|-------|
| **Workflow** | File YAML định nghĩa pipeline | `ci.yml` |
| **Job** | Một nhóm các bước, chạy trên 1 máy | `backend-test` |
| **Step** | Một bước trong job | `Run tests` |
| **Action** | Code có sẵn để tái sử dụng | `actions/checkout@v4` |
| **Runner** | Máy chạy workflow | `ubuntu-latest` |
| **Trigger** | Sự kiện kích hoạt workflow | `push`, `pull_request` |
| **Secrets** | Biến bí mật (passwords, tokens) | `secrets.GITHUB_TOKEN` |

---

## 5. ArgoCD là gì?

### GitOps là gì?

```
TRUYỀN THỐNG:
─────────────────────────────────────────────────────────────────

Developer ──▶ CI build xong ──▶ CI chạy: kubectl apply -f ...
                                              │
                                              ▼
                                         Kubernetes

Vấn đề:
- CI cần quyền truy cập trực tiếp vào K8s cluster
- Không biết ai deploy cái gì khi nào
- Khó audit và rollback


GITOPS (với ArgoCD):
─────────────────────────────────────────────────────────────────

                    ┌──────────────────────────────┐
                    │      Git Repository          │
                    │   (K8s manifests - YAML)     │
                    │                              │
                    │   k8s/                       │
                    │   ├── deployment.yaml        │
                    │   ├── service.yaml           │
                    │   └── configmap.yaml         │
                    └──────────────┬───────────────┘
                                   │
                                   │ ArgoCD theo dõi
                                   │
                    ┌──────────────▼───────────────┐
                    │          ArgoCD              │
                    │                              │
                    │   "Git nói sao, tôi làm vậy" │
                    │                              │
                    └──────────────┬───────────────┘
                                   │
                                   │ Sync
                                   │
                    ┌──────────────▼───────────────┐
                    │      Kubernetes Cluster      │
                    └──────────────────────────────┘

Ưu điểm:
✅ Git là "source of truth" - mọi thay đổi đều được track
✅ Pull-based - ArgoCD kéo thay đổi, không cần cho CI quyền push
✅ Tự động rollback bằng git revert
✅ Dễ audit - xem git log là biết ai làm gì
```

### Cách ArgoCD hoạt động

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         ARGOCD WORKFLOW                                      │
│                                                                              │
│   ┌─────────────────┐                                                       │
│   │  Git Repository │                                                       │
│   │                 │                                                       │
│   │  k8s/apps/      │                                                       │
│   │  └── backend.yaml                                                       │
│   │      image: v1.0 │                                                       │
│   └────────┬────────┘                                                       │
│            │                                                                 │
│            │  ① ArgoCD poll mỗi 3 phút                                      │
│            │     (hoặc webhook khi có push)                                 │
│            ▼                                                                 │
│   ┌─────────────────┐        ┌─────────────────┐                           │
│   │     ArgoCD      │        │   K8s Cluster   │                           │
│   │                 │        │                 │                           │
│   │  Desired State: │   ②    │  Live State:    │                           │
│   │  image: v1.0    │◀──────▶│  image: v0.9    │                           │
│   │                 │Compare │                 │                           │
│   └────────┬────────┘        └─────────────────┘                           │
│            │                                                                 │
│            │  ③ Phát hiện KHÁC BIỆT!                                        │
│            │     (OutOfSync)                                                │
│            │                                                                 │
│            ▼                                                                 │
│   ┌─────────────────────────────────────────────────────────────────────┐  │
│   │                         SYNC OPTIONS                                 │  │
│   │                                                                      │  │
│   │   AUTO SYNC (Staging):           MANUAL SYNC (Production):          │  │
│   │   ┌─────────────────┐            ┌─────────────────┐                │  │
│   │   │ Tự động deploy │            │ Chờ approval    │                │  │
│   │   │ ngay lập tức   │            │ từ DevOps/Admin │                │  │
│   │   └─────────────────┘            └─────────────────┘                │  │
│   │                                                                      │  │
│   └─────────────────────────────────────────────────────────────────────┘  │
│            │                                                                 │
│            │  ④ SYNC: Apply changes                                         │
│            │                                                                 │
│            ▼                                                                 │
│   ┌─────────────────┐                                                       │
│   │   K8s Cluster   │                                                       │
│   │                 │                                                       │
│   │  image: v1.0 ✅ │  ← Đã cập nhật!                                       │
│   │                 │                                                       │
│   └─────────────────┘                                                       │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Trạng thái trong ArgoCD

```
┌─────────────────────────────────────────────────────────────────┐
│                    ARGOCD APPLICATION STATES                     │
│                                                                  │
│   SYNC STATUS (So sánh Git vs Cluster)                          │
│   ─────────────────────────────────────                         │
│                                                                  │
│   ┌─────────┐     Git và Cluster GIỐNG nhau                     │
│   │ Synced  │     → Không cần làm gì                            │
│   │   ✅    │                                                    │
│   └─────────┘                                                    │
│                                                                  │
│   ┌──────────────┐  Git và Cluster KHÁC nhau                    │
│   │ OutOfSync    │  → Cần sync để cập nhật                      │
│   │     ⚠️       │                                               │
│   └──────────────┘                                               │
│                                                                  │
│                                                                  │
│   HEALTH STATUS (Tình trạng pods)                               │
│   ─────────────────────────────────                             │
│                                                                  │
│   ┌─────────┐     Tất cả pods đang chạy tốt                     │
│   │ Healthy │                                                    │
│   │   💚    │                                                    │
│   └─────────┘                                                    │
│                                                                  │
│   ┌─────────────┐  Đang deploy/starting                         │
│   │ Progressing │                                                │
│   │     🔄      │                                                │
│   └─────────────┘                                                │
│                                                                  │
│   ┌──────────┐    Có pod bị lỗi                                 │
│   │ Degraded │                                                   │
│   │    🟡    │                                                   │
│   └──────────┘                                                   │
│                                                                  │
│   ┌──────────┐    Ứng dụng không hoạt động                      │
│   │ Unhealthy│                                                   │
│   │    ❌    │                                                   │
│   └──────────┘                                                   │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## 6. Luồng hoạt động chi tiết

### Ví dụ: Developer sửa bug và deploy

```
BƯỚC 1: Developer sửa code
═══════════════════════════════════════════════════════════════════

Developer Ngọc phát hiện bug trong file handler.go
                │
                ▼
        Sửa code locally
                │
                ▼
        git add .
        git commit -m "fix: sửa lỗi tính toán giá"
        git push origin develop


BƯỚC 2: GitHub Actions tự động chạy CI
═══════════════════════════════════════════════════════════════════

GitHub nhận được push event
                │
                ▼
        ┌───────────────────────────────────────────────────────┐
        │              CI PIPELINE RUNNING                       │
        │                                                        │
        │   ⏳ backend-test    ───▶  ✅ Passed (2m 30s)         │
        │   ⏳ frontend-test   ───▶  ✅ Passed (1m 45s)         │
        │   ⏳ backend-build   ───▶  ✅ Pushed image (3m 10s)   │
        │   ⏳ security-scan   ───▶  ✅ No vulnerabilities      │
        │   ⏳ update-manifests ───▶ ✅ Updated kustomization   │
        │                                                        │
        │   Total time: 5 minutes                                │
        └───────────────────────────────────────────────────────┘
                │
                ▼
        Image được push lên: ghcr.io/florus/backend:abc123


BƯỚC 3: CI tự động cập nhật K8s manifests
═══════════════════════════════════════════════════════════════════

CI workflow commit vào repo:

        k8s/overlays/staging/kustomization.yaml
        ────────────────────────────────────────
        images:
          - name: florus-backend
            newName: ghcr.io/florus/backend
            newTag: abc123    ← Tag mới!


BƯỚC 4: ArgoCD phát hiện thay đổi
═══════════════════════════════════════════════════════════════════

ArgoCD (chạy trong K8s cluster):
                │
                ├── Poll Git repo mỗi 3 phút
                │
                ▼
        "Ồ! kustomization.yaml đã thay đổi!"
        "Image tag mới: abc123"
                │
                ▼
        So sánh với cluster hiện tại:
        - Git:     image: ghcr.io/florus/backend:abc123
        - Cluster: image: ghcr.io/florus/backend:xyz789
                │
                ▼
        Status: OutOfSync ⚠️


BƯỚC 5: ArgoCD sync (auto hoặc manual)
═══════════════════════════════════════════════════════════════════

STAGING (Auto-sync enabled):
        │
        ▼
ArgoCD tự động apply changes:
        │
        ├── kubectl set image deployment/backend ...
        │
        ▼
Kubernetes rolling update:
        │
        ├── Tạo pod mới với image abc123
        ├── Health check pod mới
        ├── Terminate pod cũ
        │
        ▼
        ✅ Deployment complete!
        ✅ Health: Healthy
        ✅ Sync: Synced


BƯỚC 6: Thông báo
═══════════════════════════════════════════════════════════════════

Slack notification:
        ┌─────────────────────────────────────────────────┐
        │  ✅ Staging deployment successful               │
        │                                                  │
        │  App: florus-staging                            │
        │  Commit: abc123                                  │
        │  Author: Ngọc                                    │
        │  Message: fix: sửa lỗi tính toán giá            │
        └─────────────────────────────────────────────────┘
```

---

## 7. Giải thích từng file cấu hình

### Cấu trúc thư mục

```
florus/
├── .github/
│   └── workflows/
│       ├── ci.yml              # CI pipeline
│       └── cd.yml              # CD pipeline
│
├── k8s/
│   ├── base/                   # Configs dùng chung
│   │   ├── kustomization.yaml
│   │   ├── namespace.yaml
│   │   ├── configmap.yaml
│   │   └── secrets.yaml
│   │
│   ├── apps/                   # Ứng dụng
│   │   ├── kustomization.yaml
│   │   ├── backend.yaml
│   │   ├── frontend.yaml
│   │   └── ingress.yaml
│   │
│   ├── data/                   # Data services
│   │   ├── kustomization.yaml
│   │   ├── redis.yaml
│   │   ├── neo4j.yaml
│   │   └── kafka.yaml
│   │
│   ├── overlays/               # Configs theo môi trường
│   │   ├── staging/
│   │   │   └── kustomization.yaml
│   │   └── production/
│   │       ├── kustomization.yaml
│   │       ├── hpa.yaml        # Auto-scaling
│   │       └── pdb.yaml        # Pod disruption budget
│   │
│   └── argocd/                 # ArgoCD configs
│       ├── project.yaml
│       ├── application-staging.yaml
│       └── application-production.yaml
│
└── docs/
    └── CICD.md
```

### Kustomize là gì?

```
VẤN ĐỀ: Làm sao deploy cùng app lên nhiều môi trường khác nhau?
═══════════════════════════════════════════════════════════════════

Staging:                         Production:
- 2 replicas                     - 5 replicas
- 256MB RAM                      - 1GB RAM
- LOG_LEVEL=debug                - LOG_LEVEL=info
- image:latest                   - image:v1.2.3


GIẢI PHÁP 1 (Tệ): Copy paste
─────────────────────────────
k8s/
├── staging/
│   └── backend.yaml     # Copy từ production, sửa vài chỗ
└── production/
    └── backend.yaml     # 90% giống staging

Vấn đề: Sửa 1 chỗ phải sửa cả 2 file


GIẢI PHÁP 2 (Tốt): Kustomize
─────────────────────────────
k8s/
├── base/
│   └── backend.yaml     # Cấu hình CHUNG (replicas, ports, ...)
│
├── overlays/
│   ├── staging/
│   │   └── kustomization.yaml   # CHỈ GHI KHÁC BIỆT
│   │       replicas: 2
│   │       memory: 256Mi
│   │
│   └── production/
│       └── kustomization.yaml   # CHỈ GHI KHÁC BIỆT
│           replicas: 5
│           memory: 1Gi


Khi deploy staging:
$ kustomize build overlays/staging | kubectl apply -f -

Kustomize sẽ:
1. Đọc base/backend.yaml
2. Áp dụng các thay đổi từ overlays/staging/kustomization.yaml
3. Output ra YAML hoàn chỉnh
```

### Giải thích kustomization.yaml

```yaml
# k8s/overlays/staging/kustomization.yaml

apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

# ═══════════════════════════════════════════════════════════════
# NAMESPACE: Tất cả resources sẽ được deploy vào namespace này
# ═══════════════════════════════════════════════════════════════
namespace: florus

# ═══════════════════════════════════════════════════════════════
# RESOURCES: Các file YAML base cần include
# ═══════════════════════════════════════════════════════════════
resources:
  - ../../base          # Include tất cả từ k8s/base/
  - ../../data          # Include tất cả từ k8s/data/
  - ../../apps          # Include tất cả từ k8s/apps/

# ═══════════════════════════════════════════════════════════════
# LABELS: Thêm labels vào TẤT CẢ resources
# ═══════════════════════════════════════════════════════════════
commonLabels:
  environment: staging   # Để phân biệt staging/production

# ═══════════════════════════════════════════════════════════════
# IMAGES: Override image tags
# ═══════════════════════════════════════════════════════════════
images:
  - name: florus-backend                    # Tên image trong base YAML
    newName: ghcr.io/your-org/florus-backend  # Tên mới (full path)
    newTag: latest                          # Tag mới

# ═══════════════════════════════════════════════════════════════
# REPLICAS: Override số replicas
# ═══════════════════════════════════════════════════════════════
replicas:
  - name: backend        # Tên deployment
    count: 2             # Số replicas cho staging

# ═══════════════════════════════════════════════════════════════
# PATCHES: Thay đổi chi tiết hơn
# ═══════════════════════════════════════════════════════════════
patches:
  - target:
      kind: Deployment
      name: backend
    patch: |
      - op: replace
        path: /spec/template/spec/containers/0/resources/limits/memory
        value: 256Mi      # Staging dùng ít RAM hơn

# ═══════════════════════════════════════════════════════════════
# CONFIGMAP GENERATOR: Tạo/override ConfigMap
# ═══════════════════════════════════════════════════════════════
configMapGenerator:
  - name: florus-config
    behavior: merge       # Merge với ConfigMap từ base
    literals:
      - LOG_LEVEL=debug   # Staging cần debug log
      - GIN_MODE=debug
```

---

## 8. Thực hành từng bước

### Bước 1: Setup local Kubernetes

```batch
REM 1. Kiểm tra prerequisites
docker --version
kind --version
kubectl version --client

REM 2. Tạo cluster
cd k8s\scripts
deploy-all.bat

REM 3. Kiểm tra
kubectl get pods -A
```

### Bước 2: Cài ArgoCD

```batch
REM 1. Cài ArgoCD
setup-argocd.bat

REM 2. Truy cập UI
REM Mở browser: http://localhost:30443
REM Username: admin
REM Password: (xem output của script)
```

### Bước 3: Test CI locally

```batch
REM Giả lập CI bằng cách chạy các bước thủ công

REM Backend tests
cd backend
go test -v ./...

REM Frontend tests
cd frontend
npm test

REM Build Docker images
docker build -t florus-backend:local ./backend
docker build -t florus-frontend:local ./frontend
```

### Bước 4: Test CD locally

```batch
REM 1. Load images vào Kind
kind load docker-image florus-backend:local --name florus-local
kind load docker-image florus-frontend:local --name florus-local

REM 2. Deploy bằng Kustomize
cd k8s
kustomize build overlays/staging | kubectl apply -f -

REM 3. Kiểm tra
kubectl get pods -n florus
kubectl logs -n florus deployment/backend
```

### Bước 5: Setup GitHub Actions

```yaml
# 1. Push code lên GitHub

# 2. Tạo secrets trong GitHub repo:
#    Settings → Secrets → Actions → New repository secret

# Secrets cần tạo:
# - Không cần tạo GITHUB_TOKEN (tự động có)
# - SLACK_WEBHOOK_URL (optional, cho notifications)

# 3. Push code để trigger CI:
git push origin develop

# 4. Xem CI chạy:
# GitHub → Actions tab
```

---

## Thuật ngữ tham khảo

| Thuật ngữ | Tiếng Việt | Giải thích |
|-----------|------------|------------|
| CI | Tích hợp liên tục | Tự động test code khi push |
| CD | Triển khai liên tục | Tự động deploy code |
| Pipeline | Đường ống | Chuỗi các bước tự động |
| Workflow | Luồng công việc | File định nghĩa CI/CD |
| Job | Công việc | Nhóm các bước trong workflow |
| Step | Bước | Một hành động trong job |
| Trigger | Kích hoạt | Sự kiện khởi chạy pipeline |
| Runner | Máy chạy | Server chạy CI/CD |
| Artifact | Sản phẩm | File output (images, logs) |
| Registry | Kho lưu trữ | Nơi chứa Docker images |
| GitOps | Git Operations | Dùng Git quản lý infrastructure |
| Sync | Đồng bộ | Cập nhật cluster theo Git |
| Rollback | Hoàn tác | Quay về phiên bản cũ |
| Canary | Chim hoàng yến | Deploy thử 1 pod trước |
| Rolling Update | Cập nhật cuộn | Thay pods từng cái một |

---

## Tài liệu tham khảo

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [ArgoCD Documentation](https://argo-cd.readthedocs.io/)
- [Kustomize Documentation](https://kustomize.io/)
- [Kubernetes Documentation](https://kubernetes.io/docs/)

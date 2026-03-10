# Florus CI/CD Pipeline

## Overview

Florus sử dụng **GitHub Actions** cho CI (Continuous Integration) và **ArgoCD** cho CD (Continuous Deployment) theo mô hình **GitOps**.

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                            CI/CD Pipeline Flow                               │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│   Developer                                                                  │
│      │                                                                       │
│      ▼                                                                       │
│   ┌──────┐    ┌─────────────────────────────────────────────────────────┐  │
│   │ Push │───▶│                  GitHub Actions (CI)                     │  │
│   └──────┘    │  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌───────────┐  │  │
│               │  │  Lint   │─▶│  Test   │─▶│  Build  │─▶│   Push    │  │  │
│               │  └─────────┘  └─────────┘  └─────────┘  │  Images   │  │  │
│               │                                          └─────┬─────┘  │  │
│               │                                                │        │  │
│               │  ┌─────────────────────────────────────────────┼────┐   │  │
│               │  │  Update K8s Manifests (kustomize)           │    │   │  │
│               │  │  - Set new image tags                       ▼    │   │  │
│               │  │  - Commit to main branch            ┌───────────┐│   │  │
│               │  │                                     │  GHCR     ││   │  │
│               │  └─────────────────────────────────────└───────────┘│   │  │
│               └─────────────────────────────────────────────────────────┘  │
│                                        │                                     │
│                                        ▼                                     │
│               ┌─────────────────────────────────────────────────────────┐  │
│               │                     ArgoCD (CD)                          │  │
│               │                                                          │  │
│               │  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐ │  │
│               │  │   Watch     │───▶│    Sync     │───▶│   Deploy    │ │  │
│               │  │   Git Repo  │    │   Changes   │    │   to K8s    │ │  │
│               │  └─────────────┘    └─────────────┘    └─────────────┘ │  │
│               │                                                          │  │
│               │  ┌─────────────────────────────────────────────────────┐│  │
│               │  │  Environments:                                       ││  │
│               │  │  • Staging:    Auto-sync from develop branch        ││  │
│               │  │  • Production: Manual sync from main branch         ││  │
│               │  └─────────────────────────────────────────────────────┘│  │
│               └─────────────────────────────────────────────────────────┘  │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

## CI Pipeline (GitHub Actions)

### Triggers

| Event | Branch | Action |
|-------|--------|--------|
| Push | `main`, `develop` | Full CI + Deploy |
| Pull Request | `main`, `develop` | Tests only |

### Jobs

#### 1. Backend Tests
```yaml
- Checkout code
- Setup Go 1.21
- Run golangci-lint
- Run unit tests with coverage
- Upload coverage to Codecov
```

#### 2. Frontend Tests
```yaml
- Checkout code
- Setup Node.js 20
- npm ci
- Run ESLint
- Run type-check
- Run tests with coverage
```

#### 3. Build & Push Images
```yaml
- Build Docker images (multi-stage)
- Push to GitHub Container Registry (ghcr.io)
- Tag: branch name, commit SHA, 'latest' for main
```

#### 4. Security Scan
```yaml
- Run Trivy vulnerability scanner
- Upload results to GitHub Security
```

#### 5. Update Manifests
```yaml
- Update image tags in Kustomize overlays
- Commit changes (triggers ArgoCD sync)
```

### Secrets Required

| Secret | Description |
|--------|-------------|
| `GITHUB_TOKEN` | Auto-provided, for GHCR push |
| `CODECOV_TOKEN` | Coverage upload |
| `SLACK_WEBHOOK_URL` | Notifications |
| `KUBE_CONFIG_STAGING` | Staging cluster kubeconfig (base64) |
| `KUBE_CONFIG_PRODUCTION` | Production cluster kubeconfig (base64) |

## CD Pipeline (ArgoCD)

### GitOps Workflow

```
Git Repository (Source of Truth)
        │
        ▼
┌───────────────────┐
│     ArgoCD        │
│  ┌─────────────┐  │
│  │   Watcher   │◀─┼── Polls every 3 minutes
│  └──────┬──────┘  │
│         │         │
│  ┌──────▼──────┐  │
│  │  Comparator │──┼── Detects drift
│  └──────┬──────┘  │
│         │         │
│  ┌──────▼──────┐  │
│  │   Syncer    │──┼── Applies changes to cluster
│  └─────────────┘  │
└───────────────────┘
```

### Environments

#### Staging
- **Branch**: `develop`
- **Sync**: Automatic
- **Self-heal**: Enabled
- **Prune**: Enabled

#### Production
- **Branch**: `main`
- **Sync**: Manual approval required
- **Sync Window**: Mon-Fri 9AM-5PM
- **Rollback**: Automatic on failure

### ArgoCD Applications

```
k8s/argocd/
├── project.yaml              # AppProject definition
├── application-staging.yaml  # Staging application
├── application-production.yaml # Production application
└── install.yaml              # ArgoCD customization
```

## Deployment Strategies

### Staging
- **Strategy**: Rolling Update
- **Replicas**: 2
- **Auto-sync**: Yes

### Production
- **Strategy**: Canary Deployment
- **Initial**: 1 replica (canary)
- **Full rollout**: After smoke tests pass
- **Replicas**: 3-10 (HPA)

```
Canary Deployment Flow:

  ┌─────────────────────────────────────────────────────┐
  │  1. Deploy 1 canary pod                             │
  │                                                      │
  │     [Pod 1] [Pod 2] [Pod 3]  ──▶  [Pod 1*] [Pod 2] [Pod 3]
  │      old     old     old          canary    old     old
  │                                                      │
  │  2. Run smoke tests on canary                       │
  │                                                      │
  │  3. If success: Scale to full replicas              │
  │                                                      │
  │     [Pod 1*] [Pod 2*] [Pod 3*]                      │
  │      new      new      new                          │
  │                                                      │
  │  4. If failure: Rollback                            │
  │                                                      │
  │     [Pod 1] [Pod 2] [Pod 3]                         │
  │      old     old     old                            │
  └─────────────────────────────────────────────────────┘
```

## Local Setup

### 1. Install ArgoCD

```batch
cd k8s\scripts
setup-argocd.bat
```

### 2. Access ArgoCD UI

```
URL: http://localhost:30443
Username: admin
Password: (see setup script output)
```

### 3. Manual Sync

```bash
# Sync staging
argocd app sync florus-staging

# Sync production (requires approval)
argocd app sync florus-production
```

## Monitoring & Notifications

### Slack Notifications

| Event | Channel |
|-------|---------|
| Build Success | #florus-ci |
| Build Failure | #florus-ci |
| Deploy Success | #florus-deploys |
| Deploy Failure | #florus-oncall |

### ArgoCD Notifications

```yaml
# In argocd-notifications-cm
triggers:
  - on-sync-succeeded
  - on-sync-failed
  - on-health-degraded
```

## Rollback

### Via ArgoCD UI
1. Go to application
2. Click "History"
3. Select previous revision
4. Click "Rollback"

### Via CLI

```bash
# View history
argocd app history florus-production

# Rollback to specific revision
argocd app rollback florus-production <revision>

# Or using kubectl
kubectl rollout undo deployment/backend -n florus
```

## Troubleshooting

### CI Failures

```bash
# View workflow logs
gh run view <run-id> --log

# Re-run failed jobs
gh run rerun <run-id> --failed
```

### ArgoCD Sync Issues

```bash
# Check application status
argocd app get florus-production

# View sync details
argocd app diff florus-production

# Force sync
argocd app sync florus-production --force

# Check ArgoCD logs
kubectl logs -n argocd deployment/argocd-application-controller
```

### Image Pull Errors

```bash
# Check image exists
docker pull ghcr.io/your-org/florus-backend:latest

# Check pull secrets
kubectl get secret -n florus
kubectl describe secret ghcr-secret -n florus
```

## Best Practices

1. **Never commit secrets** - Use sealed-secrets or external-secrets
2. **Use semantic versioning** for releases
3. **Tag images with commit SHA** for traceability
4. **Review ArgoCD diff** before production sync
5. **Monitor deployment metrics** post-deploy
6. **Keep rollback history** (at least 10 revisions)
7. **Use sync windows** for production
8. **Implement proper health checks** in applications

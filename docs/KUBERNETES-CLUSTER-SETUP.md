# Hướng dẫn cài đặt Kubernetes Cluster 7 Nodes - Chi tiết cho người mới

## Mục lục
1. [Tổng quan kiến trúc cluster](#1-tổng-quan-kiến-trúc-cluster)
2. [Chuẩn bị hệ thống](#2-chuẩn-bị-hệ-thống)
3. [Cài đặt Kubernetes](#3-cài-đặt-kubernetes)
4. [Khởi tạo Cluster](#4-khởi-tạo-cluster)
5. [Phân bổ Nodes và Pods](#5-phân-bổ-nodes-và-pods)
6. [Giao tiếp giữa các Nodes](#6-giao-tiếp-giữa-các-nodes)
7. [Deploy ứng dụng Florus](#7-deploy-ứng-dụng-florus)
8. [Monitoring và Troubleshooting](#8-monitoring-và-troubleshooting)

---

## 1. Tổng quan kiến trúc cluster

### 1.1. Cluster 7 Nodes đề xuất

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        KUBERNETES CLUSTER - 7 NODES                         │
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────────┐│
│  │                    CONTROL PLANE (3 Master Nodes)                       ││
│  │                                                                         ││
│  │  ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐           ││
│  │  │    master-1     │ │    master-2     │ │    master-3     │           ││
│  │  │  192.168.1.10   │ │  192.168.1.11   │ │  192.168.1.12   │           ││
│  │  │                 │ │                 │ │                 │           ││
│  │  │ • kube-apiserver│ │ • kube-apiserver│ │ • kube-apiserver│           ││
│  │  │ • etcd          │ │ • etcd          │ │ • etcd          │           ││
│  │  │ • scheduler     │ │ • scheduler     │ │ • scheduler     │           ││
│  │  │ • controller    │ │ • controller    │ │ • controller    │           ││
│  │  │                 │ │                 │ │                 │           ││
│  │  │ 4 CPU, 8GB RAM  │ │ 4 CPU, 8GB RAM  │ │ 4 CPU, 8GB RAM  │           ││
│  │  └─────────────────┘ └─────────────────┘ └─────────────────┘           ││
│  │                              │                                          ││
│  │                    ┌─────────▼─────────┐                               ││
│  │                    │   Load Balancer   │                               ││
│  │                    │   192.168.1.100   │                               ││
│  │                    │    (HAProxy)      │                               ││
│  │                    └─────────┬─────────┘                               ││
│  └──────────────────────────────┼──────────────────────────────────────────┘│
│                                 │                                            │
│  ┌──────────────────────────────┼──────────────────────────────────────────┐│
│  │                    WORKER NODES (4 Nodes)                               ││
│  │                              │                                          ││
│  │  ┌─────────────────┐ ┌──────┴────────┐ ┌─────────────────┐             ││
│  │  │    worker-1     │ │    worker-2   │ │    worker-3     │             ││
│  │  │  192.168.1.20   │ │ 192.168.1.21  │ │  192.168.1.22   │             ││
│  │  │                 │ │               │ │                 │             ││
│  │  │ Label:          │ │ Label:        │ │ Label:          │             ││
│  │  │ type=app        │ │ type=app      │ │ type=data       │             ││
│  │  │                 │ │               │ │                 │             ││
│  │  │ • Backend pods  │ │ • Backend pods│ │ • Redis         │             ││
│  │  │ • Frontend pods │ │ • Frontend    │ │ • Neo4j         │             ││
│  │  │                 │ │               │ │ • Kafka         │             ││
│  │  │ 8 CPU, 16GB RAM │ │ 8 CPU, 16GB   │ │ 8 CPU, 32GB RAM │             ││
│  │  └─────────────────┘ └───────────────┘ └─────────────────┘             ││
│  │                                                                         ││
│  │  ┌─────────────────┐                                                   ││
│  │  │    worker-4     │                                                   ││
│  │  │  192.168.1.23   │                                                   ││
│  │  │                 │                                                   ││
│  │  │ Label:          │                                                   ││
│  │  │ type=infra      │                                                   ││
│  │  │                 │                                                   ││
│  │  │ • ArgoCD        │                                                   ││
│  │  │ • Prometheus    │                                                   ││
│  │  │ • Grafana       │                                                   ││
│  │  │                 │                                                   ││
│  │  │ 4 CPU, 8GB RAM  │                                                   ││
│  │  └─────────────────┘                                                   ││
│  └─────────────────────────────────────────────────────────────────────────┘│
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 1.2. Vai trò từng node

| Node | IP | Vai trò | Specs | Chạy gì |
|------|-----|---------|-------|---------|
| **master-1** | 192.168.1.10 | Control Plane (Leader) | 4 CPU, 8GB RAM | API Server, etcd, Scheduler |
| **master-2** | 192.168.1.11 | Control Plane (Standby) | 4 CPU, 8GB RAM | API Server, etcd, Scheduler |
| **master-3** | 192.168.1.12 | Control Plane (Standby) | 4 CPU, 8GB RAM | API Server, etcd, Scheduler |
| **worker-1** | 192.168.1.20 | App Worker | 8 CPU, 16GB RAM | Backend, Frontend |
| **worker-2** | 192.168.1.21 | App Worker | 8 CPU, 16GB RAM | Backend, Frontend |
| **worker-3** | 192.168.1.22 | Data Worker | 8 CPU, 32GB RAM | Redis, Neo4j, Kafka |
| **worker-4** | 192.168.1.23 | Infra Worker | 4 CPU, 8GB RAM | ArgoCD, Monitoring |

### 1.3. Tại sao cần 3 Master nodes?

```
Nếu chỉ có 1 Master:
┌──────────────┐
│   Master-1   │ ──── Nếu chết → Cluster không hoạt động!
│   (Single)   │
└──────────────┘

Nếu có 3 Masters (High Availability):
┌──────────────┐ ┌──────────────┐ ┌──────────────┐
│   Master-1   │ │   Master-2   │ │   Master-3   │
│   (Leader)   │ │  (Standby)   │ │  (Standby)   │
└──────────────┘ └──────────────┘ └──────────────┘
       │                │                │
       └────────────────┼────────────────┘
                        │
              Quorum (đa số) = 2

- Nếu Master-1 chết → Master-2 hoặc 3 lên thay
- Cần ít nhất 2/3 nodes sống để cluster hoạt động
- Gọi là "etcd quorum" - cơ chế đồng thuận
```

---

## 2. Chuẩn bị hệ thống

### 2.1. Yêu cầu phần cứng

| Component | Master Node | Worker Node (App) | Worker Node (Data) |
|-----------|-------------|-------------------|-------------------|
| **CPU** | 4 cores | 8 cores | 8 cores |
| **RAM** | 8 GB | 16 GB | 32 GB |
| **Disk** | 100 GB SSD | 100 GB SSD | 500 GB SSD |
| **Network** | 1 Gbps | 1 Gbps | 1 Gbps |

### 2.2. Yêu cầu mạng

```
┌─────────────────────────────────────────────────────────────────┐
│                     NETWORK REQUIREMENTS                         │
│                                                                  │
│  Tất cả nodes phải:                                              │
│  • Ping được lẫn nhau                                            │
│  • Có hostname duy nhất                                          │
│  • Có IP tĩnh                                                    │
│  • Đồng bộ thời gian (NTP)                                       │
│                                                                  │
│  Ports cần mở:                                                   │
│                                                                  │
│  Master Nodes:                                                   │
│  ├── 6443    : Kubernetes API Server                             │
│  ├── 2379-80 : etcd client/peer                                  │
│  ├── 10250   : Kubelet API                                       │
│  ├── 10251   : kube-scheduler                                    │
│  └── 10252   : kube-controller-manager                           │
│                                                                  │
│  Worker Nodes:                                                   │
│  ├── 10250   : Kubelet API                                       │
│  └── 30000-32767 : NodePort Services                             │
│                                                                  │
│  All Nodes:                                                      │
│  └── 6783    : Weave Net (CNI)                                   │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### 2.3. Chuẩn bị trên TẤT CẢ nodes (7 nodes)

```bash
# ============================================================
# BƯỚC 1: Cập nhật hệ thống
# ============================================================
sudo apt update && sudo apt upgrade -y

# ============================================================
# BƯỚC 2: Đặt hostname cho từng node
# ============================================================
# Trên master-1:
sudo hostnamectl set-hostname master-1

# Trên master-2:
sudo hostnamectl set-hostname master-2

# Trên master-3:
sudo hostnamectl set-hostname master-3

# Trên worker-1:
sudo hostnamectl set-hostname worker-1

# ... tương tự cho các node khác

# ============================================================
# BƯỚC 3: Cấu hình /etc/hosts (trên TẤT CẢ nodes)
# ============================================================
sudo tee -a /etc/hosts << EOF
192.168.1.10 master-1
192.168.1.11 master-2
192.168.1.12 master-3
192.168.1.20 worker-1
192.168.1.21 worker-2
192.168.1.22 worker-3
192.168.1.23 worker-4
192.168.1.100 k8s-api  # Virtual IP cho Load Balancer
EOF

# ============================================================
# BƯỚC 4: Tắt Swap (Kubernetes yêu cầu)
# ============================================================
# Tắt ngay lập tức
sudo swapoff -a

# Tắt vĩnh viễn (comment dòng swap trong fstab)
sudo sed -i '/ swap / s/^/#/' /etc/fstab

# Verify
free -h  # Swap phải là 0

# ============================================================
# BƯỚC 5: Cấu hình kernel modules
# ============================================================
cat << EOF | sudo tee /etc/modules-load.d/k8s.conf
overlay
br_netfilter
EOF

sudo modprobe overlay
sudo modprobe br_netfilter

# ============================================================
# BƯỚC 6: Cấu hình sysctl
# ============================================================
cat << EOF | sudo tee /etc/sysctl.d/k8s.conf
net.bridge.bridge-nf-call-iptables  = 1
net.bridge.bridge-nf-call-ip6tables = 1
net.ipv4.ip_forward                 = 1
EOF

sudo sysctl --system

# ============================================================
# BƯỚC 7: Cài đặt containerd (Container Runtime)
# ============================================================
# Cài đặt dependencies
sudo apt install -y apt-transport-https ca-certificates curl gnupg

# Thêm Docker repository (để lấy containerd)
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg

echo "deb [arch=amd64 signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list

sudo apt update
sudo apt install -y containerd.io

# Cấu hình containerd
sudo mkdir -p /etc/containerd
containerd config default | sudo tee /etc/containerd/config.toml

# Bật SystemdCgroup (quan trọng!)
sudo sed -i 's/SystemdCgroup = false/SystemdCgroup = true/' /etc/containerd/config.toml

# Khởi động containerd
sudo systemctl restart containerd
sudo systemctl enable containerd

# Verify
sudo systemctl status containerd
```

### 2.4. Cài đặt kubeadm, kubelet, kubectl (TẤT CẢ nodes)

```bash
# ============================================================
# BƯỚC 8: Thêm Kubernetes repository
# ============================================================
# Tạo thư mục cho keyrings
sudo mkdir -p /etc/apt/keyrings

# Thêm key
curl -fsSL https://pkgs.k8s.io/core:/stable:/v1.29/deb/Release.key | sudo gpg --dearmor -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg

# Thêm repository
echo 'deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v1.29/deb/ /' | sudo tee /etc/apt/sources.list.d/kubernetes.list

# ============================================================
# BƯỚC 9: Cài đặt Kubernetes components
# ============================================================
sudo apt update
sudo apt install -y kubelet kubeadm kubectl

# Giữ phiên bản (tránh auto update)
sudo apt-mark hold kubelet kubeadm kubectl

# Verify
kubeadm version
kubectl version --client
kubelet --version

# ============================================================
# BƯỚC 10: Bật kubelet
# ============================================================
sudo systemctl enable kubelet
```

---

## 3. Cài đặt Kubernetes

### 3.1. Cài đặt HAProxy Load Balancer

**Tại sao cần Load Balancer?**

```
Không có Load Balancer:           Có Load Balancer:

Worker → master-1:6443            Worker → k8s-api:6443 (VIP)
                                           ↓
Nếu master-1 chết?               HAProxy phân phối:
→ Worker mất kết nối!            → master-1 (nếu sống)
                                  → master-2 (nếu sống)
                                  → master-3 (nếu sống)
```

**Cài đặt HAProxy (trên một server riêng hoặc master-1):**

```bash
# Cài đặt HAProxy
sudo apt install -y haproxy

# Cấu hình HAProxy
sudo tee /etc/haproxy/haproxy.cfg << EOF
global
    log /dev/log local0
    log /dev/log local1 notice
    chroot /var/lib/haproxy
    stats socket /run/haproxy/admin.sock mode 660 level admin
    stats timeout 30s
    user haproxy
    group haproxy
    daemon

defaults
    log     global
    mode    tcp
    option  tcplog
    option  dontlognull
    timeout connect 5000
    timeout client  50000
    timeout server  50000

# Kubernetes API Server
frontend k8s-api
    bind *:6443
    mode tcp
    default_backend k8s-api-backend

backend k8s-api-backend
    mode tcp
    balance roundrobin
    option tcp-check
    server master-1 192.168.1.10:6443 check fall 3 rise 2
    server master-2 192.168.1.11:6443 check fall 3 rise 2
    server master-3 192.168.1.12:6443 check fall 3 rise 2

# HAProxy Stats (optional)
listen stats
    bind *:8404
    mode http
    stats enable
    stats uri /stats
    stats refresh 10s
EOF

# Khởi động HAProxy
sudo systemctl restart haproxy
sudo systemctl enable haproxy

# Verify
sudo systemctl status haproxy
curl -k https://192.168.1.100:6443  # Sẽ trả về lỗi (chưa có cluster)
```

### 3.2. Khởi tạo Master Node đầu tiên (master-1)

```bash
# ============================================================
# TRÊN MASTER-1: Khởi tạo cluster
# ============================================================

# Tạo file cấu hình kubeadm
cat << EOF > kubeadm-config.yaml
apiVersion: kubeadm.k8s.io/v1beta3
kind: ClusterConfiguration
kubernetesVersion: v1.29.0
controlPlaneEndpoint: "192.168.1.100:6443"  # Load Balancer IP
networking:
  podSubnet: "10.244.0.0/16"                # Pod network CIDR
  serviceSubnet: "10.96.0.0/12"             # Service network CIDR
apiServer:
  certSANs:
  - "192.168.1.100"
  - "k8s-api"
  - "master-1"
  - "master-2"
  - "master-3"
  - "192.168.1.10"
  - "192.168.1.11"
  - "192.168.1.12"
etcd:
  local:
    dataDir: /var/lib/etcd
---
apiVersion: kubeadm.k8s.io/v1beta3
kind: InitConfiguration
localAPIEndpoint:
  advertiseAddress: "192.168.1.10"
  bindPort: 6443
nodeRegistration:
  criSocket: unix:///var/run/containerd/containerd.sock
EOF

# Khởi tạo cluster
sudo kubeadm init --config=kubeadm-config.yaml --upload-certs

# OUTPUT SẼ GẦN NHƯ SAU:
# ================================================
# Your Kubernetes control-plane has initialized successfully!
#
# To start using your cluster, you need to run the following as a regular user:
#
#   mkdir -p $HOME/.kube
#   sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
#   sudo chown $(id -u):$(id -g) $HOME/.kube/config
#
# You can now join any number of control-plane nodes by running:
#
#   kubeadm join 192.168.1.100:6443 --token abcdef.1234567890abcdef \
#     --discovery-token-ca-cert-hash sha256:1234... \
#     --control-plane --certificate-key abc123...
#
# You can now join any number of worker nodes by running:
#
#   kubeadm join 192.168.1.100:6443 --token abcdef.1234567890abcdef \
#     --discovery-token-ca-cert-hash sha256:1234...
# ================================================

# LƯU LẠI CÁC LỆNH JOIN NÀY!
```

```bash
# ============================================================
# Cấu hình kubectl cho user hiện tại
# ============================================================
mkdir -p $HOME/.kube
sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
sudo chown $(id -u):$(id -g) $HOME/.kube/config

# Verify
kubectl get nodes
# NAME       STATUS     ROLES           AGE   VERSION
# master-1   NotReady   control-plane   1m    v1.29.0
```

### 3.3. Cài đặt CNI (Container Network Interface)

**CNI là gì?**

```
CNI = Plugin mạng cho Kubernetes

Không có CNI:
┌─────────┐     ┌─────────┐
│  Pod A  │ ✗✗✗ │  Pod B  │   Pods không thể nói chuyện!
└─────────┘     └─────────┘

Có CNI (ví dụ: Calico):
┌─────────┐     ┌─────────┐
│  Pod A  │ ←→  │  Pod B  │   Pods giao tiếp qua overlay network
│10.244.1.│     │10.244.2.│
└─────────┘     └─────────┘
       ↓              ↓
    ┌──────────────────────┐
    │   Calico/Weave/etc   │
    │   (Overlay Network)  │
    └──────────────────────┘
```

```bash
# ============================================================
# TRÊN MASTER-1: Cài đặt Calico CNI
# ============================================================

# Cài đặt Calico
kubectl apply -f https://raw.githubusercontent.com/projectcalico/calico/v3.27.0/manifests/calico.yaml

# Đợi Calico khởi động (2-3 phút)
kubectl get pods -n kube-system -w

# Khi tất cả pods Running, node sẽ Ready
kubectl get nodes
# NAME       STATUS   ROLES           AGE   VERSION
# master-1   Ready    control-plane   5m    v1.29.0
```

### 3.4. Join các Master nodes khác (master-2, master-3)

```bash
# ============================================================
# TRÊN MASTER-2 và MASTER-3: Join cluster làm control-plane
# ============================================================

# Sử dụng lệnh join từ output của kubeadm init (với --control-plane)
sudo kubeadm join 192.168.1.100:6443 \
    --token abcdef.1234567890abcdef \
    --discovery-token-ca-cert-hash sha256:1234567890abcdef... \
    --control-plane \
    --certificate-key abc123def456...

# Cấu hình kubectl (trên mỗi master)
mkdir -p $HOME/.kube
sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
sudo chown $(id -u):$(id -g) $HOME/.kube/config
```

### 3.5. Join các Worker nodes (worker-1 đến worker-4)

```bash
# ============================================================
# TRÊN TẤT CẢ WORKER NODES: Join cluster
# ============================================================

# Sử dụng lệnh join từ output của kubeadm init (KHÔNG có --control-plane)
sudo kubeadm join 192.168.1.100:6443 \
    --token abcdef.1234567890abcdef \
    --discovery-token-ca-cert-hash sha256:1234567890abcdef...

# Nếu token hết hạn, tạo token mới trên master:
# kubeadm token create --print-join-command
```

### 3.6. Verify Cluster

```bash
# ============================================================
# TRÊN MASTER-1: Kiểm tra cluster
# ============================================================

# Xem tất cả nodes
kubectl get nodes -o wide

# OUTPUT:
# NAME       STATUS   ROLES           AGE   VERSION   INTERNAL-IP     OS-IMAGE
# master-1   Ready    control-plane   10m   v1.29.0   192.168.1.10    Ubuntu 22.04
# master-2   Ready    control-plane   8m    v1.29.0   192.168.1.11    Ubuntu 22.04
# master-3   Ready    control-plane   7m    v1.29.0   192.168.1.12    Ubuntu 22.04
# worker-1   Ready    <none>          5m    v1.29.0   192.168.1.20    Ubuntu 22.04
# worker-2   Ready    <none>          4m    v1.29.0   192.168.1.21    Ubuntu 22.04
# worker-3   Ready    <none>          3m    v1.29.0   192.168.1.22    Ubuntu 22.04
# worker-4   Ready    <none>          2m    v1.29.0   192.168.1.23    Ubuntu 22.04

# Xem system pods
kubectl get pods -n kube-system

# Xem etcd members (3 masters)
kubectl get pods -n kube-system -l component=etcd
```

---

## 4. Khởi tạo Cluster

### 4.1. Sơ đồ khởi tạo

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        CLUSTER INITIALIZATION FLOW                           │
│                                                                              │
│  STEP 1: Prepare All Nodes                                                   │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐   │
│  │master-1 │ │master-2 │ │master-3 │ │worker-1 │ │worker-2 │ │worker-3/4│   │
│  │ Install │ │ Install │ │ Install │ │ Install │ │ Install │ │ Install │   │
│  │ kubeadm │ │ kubeadm │ │ kubeadm │ │ kubeadm │ │ kubeadm │ │ kubeadm │   │
│  └─────────┘ └─────────┘ └─────────┘ └─────────┘ └─────────┘ └─────────┘   │
│                                                                              │
│  STEP 2: kubeadm init (master-1)                                             │
│  ┌─────────────────────────────────────────────────────────────────────────┐│
│  │ master-1                                                                ││
│  │ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐        ││
│  │ │Generate     │ │ Start       │ │ Create      │ │ Generate    │        ││
│  │ │Certificates │→│ etcd        │→│ API Server  │→│ Join Tokens │        ││
│  │ └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘        ││
│  └─────────────────────────────────────────────────────────────────────────┘│
│                                          │                                   │
│  STEP 3: Install CNI                     │                                   │
│  ┌─────────────────────────────────────────────────────────────────────────┐│
│  │ kubectl apply -f calico.yaml                                            ││
│  │ → Creates overlay network                                                ││
│  │ → Enables pod-to-pod communication                                       ││
│  └─────────────────────────────────────────────────────────────────────────┘│
│                                          │                                   │
│  STEP 4: Join Other Masters              ▼                                   │
│  ┌─────────────────┐ ┌─────────────────┐                                    │
│  │ master-2        │ │ master-3        │                                    │
│  │ kubeadm join    │ │ kubeadm join    │                                    │
│  │ --control-plane │ │ --control-plane │                                    │
│  │                 │ │                 │                                    │
│  │ → Joins etcd    │ │ → Joins etcd    │                                    │
│  │ → Syncs certs   │ │ → Syncs certs   │                                    │
│  └─────────────────┘ └─────────────────┘                                    │
│                                          │                                   │
│  STEP 5: Join Workers                    ▼                                   │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐           │
│  │ worker-1    │ │ worker-2    │ │ worker-3    │ │ worker-4    │           │
│  │ kubeadm     │ │ kubeadm     │ │ kubeadm     │ │ kubeadm     │           │
│  │ join        │ │ join        │ │ join        │ │ join        │           │
│  │             │ │             │ │             │ │             │           │
│  │ → Registers │ │ → Registers │ │ → Registers │ │ → Registers │           │
│  │   with API  │ │   with API  │ │   with API  │ │   with API  │           │
│  └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘           │
│                                                                              │
│  STEP 6: Cluster Ready!                                                      │
│  ┌─────────────────────────────────────────────────────────────────────────┐│
│  │ kubectl get nodes → 7 nodes Ready                                       ││
│  └─────────────────────────────────────────────────────────────────────────┘│
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 5. Phân bổ Nodes và Pods

### 5.1. Labels - Đánh dấu nodes

**Labels là gì?**
- Giống như "nhãn dán" trên mỗi node
- Dùng để phân loại nodes
- Pods sẽ dựa vào labels để chọn node phù hợp

```bash
# ============================================================
# Gán labels cho từng node
# ============================================================

# Worker-1 và Worker-2: Chạy ứng dụng
kubectl label nodes worker-1 type=app
kubectl label nodes worker-2 type=app

# Worker-3: Chạy databases
kubectl label nodes worker-3 type=data

# Worker-4: Chạy infrastructure
kubectl label nodes worker-4 type=infra

# Verify labels
kubectl get nodes --show-labels

# Xem nodes theo label
kubectl get nodes -l type=app
kubectl get nodes -l type=data
kubectl get nodes -l type=infra
```

### 5.2. Node Selector - Chọn node cho Pod

```yaml
# backend-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: backend
spec:
  replicas: 3
  selector:
    matchLabels:
      app: backend
  template:
    metadata:
      labels:
        app: backend
    spec:
      # NODE SELECTOR: Chỉ chạy trên nodes có label type=app
      nodeSelector:
        type: app
      containers:
      - name: backend
        image: ghcr.io/user/florus-backend:latest
        ports:
        - containerPort: 8081
```

```
Kết quả:
                                    ┌─────────────────┐
                                    │    worker-1     │
                                    │    type=app     │
                                    │                 │
                                    │ ┌─────────────┐ │
                                    │ │ backend-1   │ │
                                    │ └─────────────┘ │
                                    └─────────────────┘

┌─────────────────┐                 ┌─────────────────┐
│    worker-3     │                 │    worker-2     │
│    type=data    │                 │    type=app     │
│                 │                 │                 │
│ ┌─────────────┐ │                 │ ┌─────────────┐ │
│ │   Redis     │ │                 │ │ backend-2   │ │
│ └─────────────┘ │                 │ └─────────────┘ │
│ ┌─────────────┐ │                 │ ┌─────────────┐ │
│ │   Neo4j     │ │                 │ │ backend-3   │ │
│ └─────────────┘ │                 │ └─────────────┘ │
└─────────────────┘                 └─────────────────┘

Backend pods CHỈ chạy trên worker-1 và worker-2 (type=app)
Redis, Neo4j CHỈ chạy trên worker-3 (type=data)
```

### 5.3. Taints và Tolerations - Ngăn pods không mong muốn

**Taints**: "Xịt thuốc" lên node, ngăn pods chạy
**Tolerations**: Pods có "khả năng chịu" được thuốc

```bash
# ============================================================
# Taint nodes để đảm bảo chỉ pods phù hợp mới chạy được
# ============================================================

# Worker-3 CHỈ dành cho data workloads
kubectl taint nodes worker-3 dedicated=data:NoSchedule

# Worker-4 CHỈ dành cho infrastructure
kubectl taint nodes worker-4 dedicated=infra:NoSchedule

# Xem taints
kubectl describe node worker-3 | grep Taints
# Taints: dedicated=data:NoSchedule
```

```yaml
# redis-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: redis
spec:
  replicas: 1
  selector:
    matchLabels:
      app: redis
  template:
    metadata:
      labels:
        app: redis
    spec:
      # NODE SELECTOR: Chạy trên data nodes
      nodeSelector:
        type: data

      # TOLERATIONS: Chấp nhận taint của data nodes
      tolerations:
      - key: "dedicated"
        operator: "Equal"
        value: "data"
        effect: "NoSchedule"

      containers:
      - name: redis
        image: redis:7
        ports:
        - containerPort: 6379
```

```
Giải thích Taints + Tolerations:

Node worker-3 có Taint: dedicated=data:NoSchedule
Nghĩa là: "Chỉ pods có Toleration phù hợp mới được chạy ở đây"

Pod redis có Toleration: dedicated=data
Nghĩa là: "Tôi chấp nhận được node có taint dedicated=data"

→ Redis được phép chạy trên worker-3
→ Các pods khác (không có toleration) KHÔNG được chạy trên worker-3
```

### 5.4. Pod Affinity và Anti-Affinity

**Affinity**: Muốn pods chạy GẦN nhau (cùng node/zone)
**Anti-Affinity**: Muốn pods chạy XA nhau (khác node)

```yaml
# backend-deployment-with-affinity.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: backend
spec:
  replicas: 3
  selector:
    matchLabels:
      app: backend
  template:
    metadata:
      labels:
        app: backend
    spec:
      affinity:
        # ANTI-AFFINITY: Không chạy 2 backend pods trên cùng 1 node
        podAntiAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
          - labelSelector:
              matchLabels:
                app: backend
            topologyKey: kubernetes.io/hostname

        # AFFINITY: Muốn chạy gần pods frontend (optional)
        podAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 100
            podAffinityTerm:
              labelSelector:
                matchLabels:
                  app: frontend
              topologyKey: kubernetes.io/hostname

      containers:
      - name: backend
        image: ghcr.io/user/florus-backend:latest
```

```
Kết quả Anti-Affinity:

TRƯỚC (không có anti-affinity):
┌─────────────────┐     ┌─────────────────┐
│    worker-1     │     │    worker-2     │
│                 │     │                 │
│ ┌─────────────┐ │     │                 │
│ │ backend-1   │ │     │  (trống)        │
│ └─────────────┘ │     │                 │
│ ┌─────────────┐ │     │                 │
│ │ backend-2   │ │     │                 │
│ └─────────────┘ │     │                 │
│ ┌─────────────┐ │     │                 │
│ │ backend-3   │ │     │                 │
│ └─────────────┘ │     │                 │
└─────────────────┘     └─────────────────┘
Nếu worker-1 chết → Mất hết 3 backend pods!

SAU (có anti-affinity):
┌─────────────────┐     ┌─────────────────┐
│    worker-1     │     │    worker-2     │
│                 │     │                 │
│ ┌─────────────┐ │     │ ┌─────────────┐ │
│ │ backend-1   │ │     │ │ backend-2   │ │
│ └─────────────┘ │     │ └─────────────┘ │
│                 │     │ ┌─────────────┐ │
│                 │     │ │ backend-3   │ │
│                 │     │ └─────────────┘ │
└─────────────────┘     └─────────────────┘
Nếu worker-1 chết → Vẫn còn 2 backend pods trên worker-2!
```

### 5.5. Tổng hợp phân bổ Florus

```yaml
# ============================================================
# File: k8s/overlays/production/node-distribution.yaml
# ============================================================

# Backend - chạy trên app nodes, spread across nodes
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: backend
spec:
  replicas: 3
  template:
    spec:
      nodeSelector:
        type: app
      affinity:
        podAntiAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
          - labelSelector:
              matchLabels:
                app: backend
            topologyKey: kubernetes.io/hostname
      containers:
      - name: backend
        resources:
          requests:
            cpu: "500m"
            memory: "512Mi"
          limits:
            cpu: "1000m"
            memory: "1Gi"

# Frontend - chạy trên app nodes
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: frontend
spec:
  replicas: 2
  template:
    spec:
      nodeSelector:
        type: app
      containers:
      - name: frontend
        resources:
          requests:
            cpu: "200m"
            memory: "256Mi"

# Redis - chạy trên data nodes
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: redis
spec:
  replicas: 1
  template:
    spec:
      nodeSelector:
        type: data
      tolerations:
      - key: "dedicated"
        operator: "Equal"
        value: "data"
        effect: "NoSchedule"
      containers:
      - name: redis
        resources:
          requests:
            memory: "1Gi"
          limits:
            memory: "2Gi"

# ArgoCD - chạy trên infra nodes
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: argocd-server
  namespace: argocd
spec:
  template:
    spec:
      nodeSelector:
        type: infra
      tolerations:
      - key: "dedicated"
        operator: "Equal"
        value: "infra"
        effect: "NoSchedule"
```

---

## 6. Giao tiếp giữa các Nodes

### 6.1. Kubernetes Networking Model

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                     KUBERNETES NETWORKING LAYERS                             │
│                                                                              │
│  LAYER 1: Node Network (Physical/Virtual Network)                            │
│  ┌─────────────────────────────────────────────────────────────────────────┐│
│  │  192.168.1.0/24                                                         ││
│  │  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐       ││
│  │  │  worker-1   │ │  worker-2   │ │  worker-3   │ │  worker-4   │       ││
│  │  │ .20         │ │ .21         │ │ .22         │ │ .23         │       ││
│  │  └──────┬──────┘ └──────┬──────┘ └──────┬──────┘ └──────┬──────┘       ││
│  │         └───────────────┴───────────────┴───────────────┘               ││
│  │                            Physical Switch                               ││
│  └─────────────────────────────────────────────────────────────────────────┘│
│                                                                              │
│  LAYER 2: Pod Network (Overlay Network - Calico/Flannel/Weave)               │
│  ┌─────────────────────────────────────────────────────────────────────────┐│
│  │  10.244.0.0/16 (Pod CIDR)                                               ││
│  │                                                                         ││
│  │  worker-1: 10.244.1.0/24    worker-2: 10.244.2.0/24                    ││
│  │  ┌─────────────────────┐    ┌─────────────────────┐                    ││
│  │  │ Pod: 10.244.1.5     │    │ Pod: 10.244.2.10    │                    ││
│  │  │ Pod: 10.244.1.6     │←──→│ Pod: 10.244.2.11    │                    ││
│  │  │ Pod: 10.244.1.7     │    │ Pod: 10.244.2.12    │                    ││
│  │  └─────────────────────┘    └─────────────────────┘                    ││
│  │                    ↑                    ↑                               ││
│  │                    └────────────────────┘                               ││
│  │                      Calico Tunnel (VXLAN/IPIP)                         ││
│  └─────────────────────────────────────────────────────────────────────────┘│
│                                                                              │
│  LAYER 3: Service Network (Virtual IPs - kube-proxy)                         │
│  ┌─────────────────────────────────────────────────────────────────────────┐│
│  │  10.96.0.0/12 (Service CIDR)                                            ││
│  │                                                                         ││
│  │  ┌──────────────────────────────────────┐                              ││
│  │  │  backend-service: 10.96.45.123:8081  │                              ││
│  │  │            ↓ (kube-proxy)            │                              ││
│  │  │  ┌─────────────────────────────────┐ │                              ││
│  │  │  │ 10.244.1.5:8081 (backend-1)     │ │                              ││
│  │  │  │ 10.244.2.10:8081 (backend-2)    │ │ ← Load Balance               ││
│  │  │  │ 10.244.1.7:8081 (backend-3)     │ │                              ││
│  │  │  └─────────────────────────────────┘ │                              ││
│  │  └──────────────────────────────────────┘                              ││
│  └─────────────────────────────────────────────────────────────────────────┘│
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 6.2. Pod-to-Pod Communication

```
Ví dụ: Frontend pod gọi Backend pod

┌─────────────────────────────────────────────────────────────────────────────┐
│                     POD-TO-POD COMMUNICATION                                 │
│                                                                              │
│  1. Frontend muốn gọi Backend                                                │
│     curl http://backend-service:8081/api/health                              │
│                                                                              │
│  2. DNS Resolution (CoreDNS)                                                 │
│     backend-service.default.svc.cluster.local → 10.96.45.123                │
│                                                                              │
│  3. Service → Pod (kube-proxy)                                               │
│     10.96.45.123:8081 → 10.244.2.10:8081 (một trong các backend pods)       │
│                                                                              │
│  4. Network Flow                                                             │
│                                                                              │
│  worker-1                                   worker-2                         │
│  ┌─────────────────────────┐               ┌─────────────────────────┐      │
│  │ frontend-pod            │               │ backend-pod             │      │
│  │ IP: 10.244.1.5          │               │ IP: 10.244.2.10         │      │
│  │                         │               │                         │      │
│  │ curl backend-service    │               │ Receives request        │      │
│  │         │               │               │         ↑               │      │
│  │         ▼               │               │         │               │      │
│  │ ┌─────────────────────┐ │               │ ┌─────────────────────┐ │      │
│  │ │    kube-proxy       │ │               │ │    kube-proxy       │ │      │
│  │ │ (iptables/ipvs)     │ │               │ │                     │ │      │
│  │ └──────────┬──────────┘ │               │ └──────────┬──────────┘ │      │
│  │            │            │               │            ↑            │      │
│  │ ┌──────────▼──────────┐ │               │ ┌──────────┴──────────┐ │      │
│  │ │    Calico Agent     │ │               │ │    Calico Agent     │ │      │
│  │ │ (routing + tunnel)  │ │               │ │                     │ │      │
│  │ └──────────┬──────────┘ │               │ └──────────┬──────────┘ │      │
│  └────────────┼────────────┘               └────────────┼────────────┘      │
│               │                                         │                    │
│               │        Physical Network                 │                    │
│               │       192.168.1.0/24                    │                    │
│               └────────────────────────────────────────→┘                    │
│                                                                              │
│  5. Packet Flow:                                                             │
│     SRC: 10.244.1.5 → DST: 10.244.2.10                                      │
│     Encapsulated in: SRC: 192.168.1.20 → DST: 192.168.1.21                  │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 6.3. Service Types và External Access

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          SERVICE TYPES                                       │
│                                                                              │
│  1. ClusterIP (Default) - Internal only                                      │
│  ┌─────────────────────────────────────────────────────────────────────────┐│
│  │                                                                         ││
│  │  ┌─────────────┐         ┌─────────────┐         ┌─────────────┐       ││
│  │  │   Pod A     │ ──────→ │  ClusterIP  │ ──────→ │   Pod B     │       ││
│  │  │ (frontend)  │         │ 10.96.1.100 │         │ (backend)   │       ││
│  │  └─────────────┘         └─────────────┘         └─────────────┘       ││
│  │                                                                         ││
│  │  Chỉ truy cập từ trong cluster                                          ││
│  └─────────────────────────────────────────────────────────────────────────┘│
│                                                                              │
│  2. NodePort - Expose qua port của node                                      │
│  ┌─────────────────────────────────────────────────────────────────────────┐│
│  │                                                                         ││
│  │  External Client                                                        ││
│  │       │                                                                 ││
│  │       ▼ http://192.168.1.20:30080                                       ││
│  │  ┌─────────────┐         ┌─────────────┐         ┌─────────────┐       ││
│  │  │  worker-1   │         │  NodePort   │         │   Pod       │       ││
│  │  │  :30080     │ ──────→ │  Service    │ ──────→ │  :8081      │       ││
│  │  └─────────────┘         └─────────────┘         └─────────────┘       ││
│  │                                                                         ││
│  │  Range: 30000-32767                                                     ││
│  └─────────────────────────────────────────────────────────────────────────┘│
│                                                                              │
│  3. LoadBalancer - Cloud provider LB                                         │
│  ┌─────────────────────────────────────────────────────────────────────────┐│
│  │                                                                         ││
│  │  External Client                                                        ││
│  │       │                                                                 ││
│  │       ▼ http://external-ip:80                                           ││
│  │  ┌─────────────┐                                                        ││
│  │  │ Cloud LB    │ (AWS ELB, GCP LB, Azure LB)                            ││
│  │  │ 203.0.113.5 │                                                        ││
│  │  └──────┬──────┘                                                        ││
│  │         │                                                               ││
│  │  ┌──────▼──────┐ ┌─────────────┐ ┌─────────────┐                       ││
│  │  │  worker-1   │ │  worker-2   │ │  worker-3   │                       ││
│  │  │  :30080     │ │  :30080     │ │  :30080     │                       ││
│  │  └─────────────┘ └─────────────┘ └─────────────┘                       ││
│  │                                                                         ││
│  │  Tự động tạo external IP (chỉ cloud)                                    ││
│  └─────────────────────────────────────────────────────────────────────────┘│
│                                                                              │
│  4. Ingress - HTTP/HTTPS routing                                             │
│  ┌─────────────────────────────────────────────────────────────────────────┐│
│  │                                                                         ││
│  │  External Client                                                        ││
│  │       │ https://api.florus.com/api/users                                ││
│  │       │ https://app.florus.com/dashboard                                ││
│  │       ▼                                                                 ││
│  │  ┌──────────────────────────────────────────────┐                      ││
│  │  │            Ingress Controller                │                      ││
│  │  │            (nginx, traefik)                  │                      ││
│  │  │                                              │                      ││
│  │  │  Rules:                                      │                      ││
│  │  │  - api.florus.com → backend-service:8081    │                      ││
│  │  │  - app.florus.com → frontend-service:80     │                      ││
│  │  └──────────────────────────────────────────────┘                      ││
│  │         │                           │                                   ││
│  │         ▼                           ▼                                   ││
│  │  ┌─────────────┐             ┌─────────────┐                           ││
│  │  │  Backend    │             │  Frontend   │                           ││
│  │  │  Service    │             │  Service    │                           ││
│  │  └─────────────┘             └─────────────┘                           ││
│  └─────────────────────────────────────────────────────────────────────────┘│
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 6.4. DNS trong Kubernetes

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                     KUBERNETES DNS (CoreDNS)                                 │
│                                                                              │
│  DNS Format:                                                                 │
│  <service-name>.<namespace>.svc.cluster.local                                │
│                                                                              │
│  Ví dụ:                                                                      │
│  ┌─────────────────────────────────────────────────────────────────────────┐│
│  │                                                                         ││
│  │  Service: backend-service                                               ││
│  │  Namespace: florus-production                                           ││
│  │                                                                         ││
│  │  DNS Names:                                                             ││
│  │  • backend-service                        (cùng namespace)              ││
│  │  • backend-service.florus-production      (cross namespace)             ││
│  │  • backend-service.florus-production.svc  (full)                        ││
│  │  • backend-service.florus-production.svc.cluster.local (FQDN)           ││
│  │                                                                         ││
│  │  IP: 10.96.45.123                                                       ││
│  │                                                                         ││
│  └─────────────────────────────────────────────────────────────────────────┘│
│                                                                              │
│  Cross-Namespace Communication:                                              │
│  ┌─────────────────────────────────────────────────────────────────────────┐│
│  │                                                                         ││
│  │  namespace: florus-staging              namespace: florus-data          ││
│  │  ┌─────────────────────┐               ┌─────────────────────┐          ││
│  │  │  backend-pod        │               │  redis-service      │          ││
│  │  │                     │ ─────────────→│                     │          ││
│  │  │  redis-service.     │               │  10.96.100.50       │          ││
│  │  │  florus-data        │               │                     │          ││
│  │  └─────────────────────┘               └─────────────────────┘          ││
│  │                                                                         ││
│  │  Backend trong florus-staging gọi Redis trong florus-data:              ││
│  │  → curl http://redis-service.florus-data:6379                           ││
│  │                                                                         ││
│  └─────────────────────────────────────────────────────────────────────────┘│
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 6.5. Network Policies - Firewall cho Pods

```yaml
# ============================================================
# Network Policy: Backend chỉ nhận traffic từ Frontend
# ============================================================
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: backend-policy
  namespace: florus-production
spec:
  # Áp dụng cho pods có label app=backend
  podSelector:
    matchLabels:
      app: backend

  policyTypes:
  - Ingress
  - Egress

  # INGRESS: Ai được gọi vào Backend?
  ingress:
  # Cho phép từ Frontend pods
  - from:
    - podSelector:
        matchLabels:
          app: frontend
    ports:
    - protocol: TCP
      port: 8081

  # Cho phép từ Ingress Controller
  - from:
    - namespaceSelector:
        matchLabels:
          name: ingress-nginx
    ports:
    - protocol: TCP
      port: 8081

  # EGRESS: Backend được gọi đi đâu?
  egress:
  # Cho phép gọi Redis
  - to:
    - podSelector:
        matchLabels:
          app: redis
    ports:
    - protocol: TCP
      port: 6379

  # Cho phép gọi DNS
  - to:
    - namespaceSelector: {}
      podSelector:
        matchLabels:
          k8s-app: kube-dns
    ports:
    - protocol: UDP
      port: 53
```

```
Kết quả Network Policy:

┌─────────────────────────────────────────────────────────────────────────────┐
│                                                                              │
│  ✓ Frontend → Backend : ALLOWED                                              │
│  ✓ Ingress  → Backend : ALLOWED                                              │
│  ✗ Other    → Backend : BLOCKED                                              │
│                                                                              │
│  ✓ Backend  → Redis   : ALLOWED                                              │
│  ✓ Backend  → DNS     : ALLOWED                                              │
│  ✗ Backend  → Internet: BLOCKED                                              │
│                                                                              │
│  ┌─────────────┐      ┌─────────────┐      ┌─────────────┐                  │
│  │  Frontend   │ ───→ │   Backend   │ ───→ │    Redis    │                  │
│  │             │  ✓   │             │  ✓   │             │                  │
│  └─────────────┘      └──────┬──────┘      └─────────────┘                  │
│                              │                                               │
│                              │ ✗ (blocked)                                   │
│                              ▼                                               │
│                        ┌───────────┐                                         │
│                        │  Internet │                                         │
│                        └───────────┘                                         │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 7. Deploy ứng dụng Florus

### 7.1. Áp dụng cấu hình production

```bash
# ============================================================
# Deploy Florus lên cluster production
# ============================================================

# 1. Tạo namespaces
kubectl create namespace florus-production
kubectl create namespace florus-data
kubectl create namespace argocd

# 2. Label nodes (đã làm ở section 5)
kubectl label nodes worker-1 type=app
kubectl label nodes worker-2 type=app
kubectl label nodes worker-3 type=data
kubectl label nodes worker-4 type=infra

# 3. Taint nodes
kubectl taint nodes worker-3 dedicated=data:NoSchedule
kubectl taint nodes worker-4 dedicated=infra:NoSchedule

# 4. Deploy data services (Redis, Neo4j, Kafka)
kubectl apply -k k8s/data/ -n florus-data

# 5. Deploy application
kubectl apply -k k8s/overlays/production

# 6. Deploy ArgoCD
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml

# 7. Verify
kubectl get pods -A -o wide
```

### 7.2. Verify Pod Distribution

```bash
# Xem pods chạy trên node nào
kubectl get pods -o wide -n florus-production

# Output:
# NAME                       READY   STATUS    NODE
# backend-xxx-1              1/1     Running   worker-1
# backend-xxx-2              1/1     Running   worker-2
# frontend-xxx-1             1/1     Running   worker-1
# frontend-xxx-2             1/1     Running   worker-2

kubectl get pods -o wide -n florus-data

# Output:
# NAME                       READY   STATUS    NODE
# redis-xxx                  1/1     Running   worker-3
# neo4j-xxx                  1/1     Running   worker-3
# kafka-xxx                  1/1     Running   worker-3

kubectl get pods -o wide -n argocd

# Output:
# NAME                       READY   STATUS    NODE
# argocd-server-xxx          1/1     Running   worker-4
# argocd-repo-server-xxx     1/1     Running   worker-4
```

---

## 8. Monitoring và Troubleshooting

### 8.1. Useful Commands

```bash
# ============================================================
# Cluster Health
# ============================================================
kubectl cluster-info
kubectl get componentstatuses
kubectl get nodes
kubectl top nodes

# ============================================================
# Pod Debugging
# ============================================================
# Xem logs
kubectl logs <pod-name> -n <namespace>
kubectl logs <pod-name> -n <namespace> --previous  # Logs trước khi crash

# Exec vào pod
kubectl exec -it <pod-name> -n <namespace> -- /bin/sh

# Xem events
kubectl get events -n <namespace> --sort-by='.lastTimestamp'

# Describe pod
kubectl describe pod <pod-name> -n <namespace>

# ============================================================
# Network Debugging
# ============================================================
# Test DNS
kubectl run test-dns --rm -it --image=busybox -- nslookup backend-service

# Test connectivity
kubectl run test-curl --rm -it --image=curlimages/curl -- curl http://backend-service:8081/api/health

# ============================================================
# Resource Usage
# ============================================================
kubectl top pods -n <namespace>
kubectl top nodes
```

### 8.2. Common Issues

| Issue | Nguyên nhân | Giải pháp |
|-------|------------|-----------|
| Pod Pending | Không đủ resources / Node selector không match | Check `kubectl describe pod` |
| Pod CrashLoopBackOff | App crash / Config sai | Check logs với `kubectl logs` |
| ImagePullBackOff | Image không tồn tại / Auth failed | Check image name, imagePullSecrets |
| Node NotReady | kubelet down / Network issue | Check `systemctl status kubelet` |
| Service không accessible | Selector sai / Port sai | Check `kubectl describe svc` |

### 8.3. Backup etcd (Quan trọng!)

```bash
# Backup etcd (chạy trên master node)
ETCDCTL_API=3 etcdctl snapshot save /backup/etcd-snapshot.db \
  --endpoints=https://127.0.0.1:2379 \
  --cacert=/etc/kubernetes/pki/etcd/ca.crt \
  --cert=/etc/kubernetes/pki/etcd/server.crt \
  --key=/etc/kubernetes/pki/etcd/server.key

# Restore etcd
ETCDCTL_API=3 etcdctl snapshot restore /backup/etcd-snapshot.db
```

---

## Tổng kết

### Checklist cài đặt:

- [ ] Chuẩn bị 7 servers với IP tĩnh
- [ ] Cài đặt containerd trên tất cả nodes
- [ ] Cài đặt kubeadm, kubelet, kubectl
- [ ] Cấu hình HAProxy load balancer
- [ ] Khởi tạo cluster với kubeadm init
- [ ] Cài đặt CNI (Calico)
- [ ] Join master nodes (HA)
- [ ] Join worker nodes
- [ ] Label và Taint nodes
- [ ] Deploy ứng dụng

### Key Concepts:

| Concept | Mục đích |
|---------|----------|
| **Labels** | Đánh dấu nodes để pods chọn |
| **Node Selector** | Pods chọn node dựa trên labels |
| **Taints** | Ngăn pods không mong muốn |
| **Tolerations** | Pods chấp nhận taints |
| **Affinity** | Pods muốn chạy gần nhau |
| **Anti-Affinity** | Pods muốn chạy xa nhau |
| **Services** | Expose pods với stable IP |
| **Network Policies** | Firewall cho pods |

---

**Author:** Claude AI
**Last Updated:** 2026-03-11
**Version:** 1.0.0

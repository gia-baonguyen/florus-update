@echo off
REM Complete deployment script for Florus on Kind (Windows)

echo ==============================================
echo   Florus Local Kubernetes Deployment
echo ==============================================
echo.

set SCRIPT_DIR=%~dp0
set K8S_DIR=%SCRIPT_DIR%..
set PROJECT_DIR=%K8S_DIR%\..

REM Check prerequisites
echo [1/8] Checking prerequisites...
where docker >nul 2>&1 || (echo Docker is required but not installed. && exit /b 1)
where kind >nul 2>&1 || (echo Kind is required but not installed. && exit /b 1)
where kubectl >nul 2>&1 || (echo kubectl is required but not installed. && exit /b 1)
echo √ All prerequisites met

REM Delete existing cluster
echo.
echo [2/8] Cleaning up existing cluster...
kind delete cluster --name florus-local 2>nul
echo √ Cleanup complete

REM Create Kind cluster
echo.
echo [3/8] Creating Kind cluster (this may take a few minutes)...
kind create cluster --config "%K8S_DIR%\local\kind-config.yaml"
if errorlevel 1 (
    echo Failed to create cluster
    exit /b 1
)
echo √ Cluster created

REM Wait for cluster
echo.
echo [4/8] Waiting for cluster to be ready...
kubectl wait --for=condition=Ready nodes --all --timeout=120s
echo √ Cluster is ready

REM Install Ingress Controller
echo.
echo [5/8] Installing NGINX Ingress Controller...
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml
echo Waiting for ingress controller (60s)...
timeout /t 60 /nobreak >nul
echo √ Ingress controller installed

REM Create namespaces and configs
echo.
echo [6/8] Creating namespaces and configurations...
kubectl apply -f "%K8S_DIR%\base\namespace.yaml"
kubectl apply -f "%K8S_DIR%\base\configmap.yaml"
kubectl apply -f "%K8S_DIR%\base\secrets.yaml"
echo √ Base configurations applied

REM Deploy data layer
echo.
echo [7/8] Deploying data layer (Redis, Neo4j, MinIO, Kafka)...
kubectl apply -f "%K8S_DIR%\data\redis.yaml"
kubectl apply -f "%K8S_DIR%\data\neo4j.yaml"
kubectl apply -f "%K8S_DIR%\data\minio.yaml"
kubectl apply -f "%K8S_DIR%\data\kafka.yaml"
echo Waiting for data services (120s)...
timeout /t 120 /nobreak >nul
echo √ Data layer deployed

REM Build and deploy applications
echo.
echo [8/8] Building and deploying applications...
call "%SCRIPT_DIR%\build-and-deploy.bat"

echo.
echo ==============================================
echo   Deployment Complete!
echo ==============================================
echo.
echo Checking pod status...
kubectl get pods -A
echo.
echo Access points:
echo   - Frontend: http://localhost:30080
echo   - Backend API: http://localhost:30081/api/health
echo   - Neo4j Browser: http://localhost:30474
echo   - MinIO Console: http://localhost:30901
echo.
echo Add to C:\Windows\System32\drivers\etc\hosts:
echo   127.0.0.1 florus.local api.florus.local
echo.
pause

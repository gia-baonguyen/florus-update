@echo off
REM Build Docker images and deploy to Kind cluster (Windows)

set SCRIPT_DIR=%~dp0
set K8S_DIR=%SCRIPT_DIR%..
set PROJECT_DIR=%K8S_DIR%\..

echo ==============================================
echo   Building and Deploying Florus Applications
echo ==============================================
echo.

REM Build Backend
echo [1/4] Building Backend image...
cd /d "%PROJECT_DIR%\backend"

REM Create Dockerfile if not exists
if not exist Dockerfile (
    echo Creating Backend Dockerfile...
    (
        echo # Build stage
        echo FROM golang:1.21-alpine AS builder
        echo WORKDIR /app
        echo COPY go.mod go.sum ./
        echo RUN go mod download
        echo COPY . .
        echo RUN CGO_ENABLED=0 GOOS=linux go build -o api ./cmd/api
        echo.
        echo # Run stage
        echo FROM alpine:3.19
        echo RUN apk --no-cache add ca-certificates tzdata
        echo WORKDIR /app
        echo COPY --from=builder /app/api .
        echo EXPOSE 8081
        echo CMD ["./api"]
    ) > Dockerfile
)

docker build -t florus-backend:local .
if errorlevel 1 (
    echo Failed to build backend image
    exit /b 1
)
echo √ Backend image built

REM Build Frontend
echo.
echo [2/4] Building Frontend image...
cd /d "%PROJECT_DIR%\frontend"

REM Create Dockerfile if not exists
if not exist Dockerfile (
    echo Creating Frontend Dockerfile...
    (
        echo # Build stage
        echo FROM node:20-alpine AS builder
        echo WORKDIR /app
        echo COPY package*.json ./
        echo RUN npm ci
        echo COPY . .
        echo RUN npm run build
        echo.
        echo # Run stage
        echo FROM nginx:alpine
        echo COPY --from=builder /app/dist /usr/share/nginx/html
        echo EXPOSE 80
        echo CMD ["nginx", "-g", "daemon off;"]
    ) > Dockerfile
)

docker build -t florus-frontend:local .
if errorlevel 1 (b 
    echo Failed to build frontend image
    exit /b 1
)
echo √ Frontend image built

REM Load images into Kind
echo.
echo [3/4] Loading images into Kind cluster...
kind load docker-image florus-backend:local --name florus-local
kind load docker-image florus-frontend:local --name florus-local
echo √ Images loaded

REM Deploy applications
echo.
echo [4/4] Deploying applications...
cd /d "%K8S_DIR%"
kubectl apply -f apps\backend.yaml
kubectl apply -f apps\frontend.yaml
kubectl apply -f apps\ingress.yaml

echo.
echo Waiting for applications to be ready (60s)...
timeout /t 60 /nobreak >nul

echo.
echo √ Applications deployed
echo.
kubectl get pods -n florus

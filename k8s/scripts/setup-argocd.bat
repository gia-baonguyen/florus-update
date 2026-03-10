@echo off
REM Install and configure ArgoCD for Florus

echo ==============================================
echo   Installing ArgoCD
echo ==============================================
echo.

set SCRIPT_DIR=%~dp0
set K8S_DIR=%SCRIPT_DIR%..

REM Create argocd namespace
echo [1/5] Creating ArgoCD namespace...
kubectl create namespace argocd 2>nul
echo √ Namespace created

REM Install ArgoCD
echo.
echo [2/5] Installing ArgoCD (this may take a few minutes)...
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
echo √ ArgoCD installed

REM Wait for ArgoCD to be ready
echo.
echo [3/5] Waiting for ArgoCD to be ready (120s)...
timeout /t 120 /nobreak >nul
kubectl wait --for=condition=Available deployment/argocd-server -n argocd --timeout=300s
echo √ ArgoCD is ready

REM Apply custom configuration
echo.
echo [4/5] Applying custom configuration...
kubectl apply -f "%K8S_DIR%\argocd\install.yaml"
echo √ Configuration applied

REM Apply Florus project and applications
echo.
echo [5/5] Creating Florus ArgoCD project and applications...
kubectl apply -f "%K8S_DIR%\argocd\project.yaml"
kubectl apply -f "%K8S_DIR%\argocd\application-staging.yaml"
kubectl apply -f "%K8S_DIR%\argocd\application-production.yaml"
echo √ Applications created

echo.
echo ==============================================
echo   ArgoCD Installation Complete!
echo ==============================================
echo.

REM Get initial admin password
echo Getting ArgoCD admin password...
for /f "tokens=*" %%i in ('kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath^="{.data.password}"') do set ENCODED_PASS=%%i

echo.
echo ArgoCD Access:
echo   URL: http://localhost:30443
echo   Username: admin
echo   Password: (run the following command to decode)
echo.
echo   PowerShell: [System.Text.Encoding]::UTF8.GetString([System.Convert]::FromBase64String("%ENCODED_PASS%"))
echo.
echo   Or use: kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" ^| base64 -d
echo.

pause

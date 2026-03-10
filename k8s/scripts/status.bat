@echo off
REM Check status of Florus deployment

echo ==============================================
echo   Florus Deployment Status
echo ==============================================
echo.

echo Cluster info:
kubectl cluster-info
echo.

echo --- Nodes ---
kubectl get nodes
echo.

echo --- All Pods ---
kubectl get pods -A
echo.

echo --- Services (florus namespace) ---
kubectl get svc -n florus
echo.

echo --- Services (florus-data namespace) ---
kubectl get svc -n florus-data
echo.

echo --- Ingress ---
kubectl get ingress -n florus
echo.

echo ==============================================
echo   Access Points
echo ==============================================
echo.
echo Frontend:     http://localhost:30080
echo Backend API:  http://localhost:30081/api/health
echo Neo4j:        http://localhost:30474 (user: neo4j, pass: florusneo4j123)
echo MinIO:        http://localhost:30901 (user: minioadmin, pass: minioadmin123)
echo.

pause

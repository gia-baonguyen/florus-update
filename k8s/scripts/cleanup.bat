@echo off
REM Cleanup Kind cluster

echo ==============================================
echo   Cleaning up Florus Kind Cluster
echo ==============================================
echo.

echo Deleting Kind cluster...
kind delete cluster --name florus-local

echo.
echo √ Cluster deleted
echo.

REM Optional: Remove Docker images
set /p REMOVE_IMAGES="Remove Docker images? (y/n): "
if /i "%REMOVE_IMAGES%"=="y" (
    echo Removing Docker images...
    docker rmi florus-backend:local 2>nul
    docker rmi florus-frontend:local 2>nul
    echo √ Images removed
)

echo.
echo Cleanup complete!
pause

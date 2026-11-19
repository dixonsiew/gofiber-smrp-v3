@echo off
setlocal EnableDelayedExpansion

REM ============================================
REM Configuration
REM ============================================
set IMAGE_NAME=docker.io/wfsiew888/smrp-gofiber
set IMAGE_TAG=1.0.0
set FULL_IMAGE=%IMAGE_NAME%:%IMAGE_TAG%
set SERVICE_NAME=smrp-gofiber
set DEPLOYMENT_NAME=smrp-gofiber
set DEPLOY_FILE=deployment.yaml

echo ============================================
echo  Starting deployment of %SERVICE_NAME%
echo ============================================
echo.

REM ============================================
REM Step 1: Delete Service
REM ============================================
echo [1/6] Deleting service %SERVICE_NAME%...
kubectl delete service %SERVICE_NAME% --ignore-not-found=true
if errorlevel 1 (
    echo [WARN] Failed to delete service. Continuing...
)

REM ============================================
REM Step 2: Delete Deployment
REM ============================================
echo.
echo [2/6] Deleting deployment %DEPLOYMENT_NAME%...
kubectl delete deployment %DEPLOYMENT_NAME% --ignore-not-found=true
if errorlevel 1 (
    echo [WARN] Failed to delete deployment. Continuing...
)

REM ============================================
REM Step 3: Remove Local Docker Image
REM ============================================
echo.
echo [3/6] Removing local Docker image %FULL_IMAGE%...
docker rmi -f %FULL_IMAGE% 2>nul
if errorlevel 1 (
    echo [WARN] Local image not found or already removed. Continuing...
) else (
    echo [OK] Local image removed.
)

REM ============================================
REM Step 4: Build Docker Image
REM ============================================
echo.
echo [4/6] Building Docker image %FULL_IMAGE%...
docker build -t %FULL_IMAGE% .
if errorlevel 1 (
    echo [ERROR] Docker build failed. Aborting.
    exit /b 1
)

REM ============================================
REM Step 5: Push Docker Image
REM ============================================
echo.
echo [5/6] Pushing Docker image to registry...
docker push %FULL_IMAGE%
if errorlevel 1 (
    echo [ERROR] Docker push failed. Aborting.
    exit /b 1
)

REM ============================================
REM Step 6: Apply Deployment
REM ============================================
echo.
echo [6/6] Applying %DEPLOY_FILE%...
kubectl apply -f %DEPLOY_FILE%
if errorlevel 1 (
    echo [ERROR] kubectl apply failed.
    exit /b 1
)

REM ============================================
REM Restart Deployment to Pick Up New Image
REM ============================================
echo.
echo Restarting deployment to use freshly pushed image...
kubectl rollout restart deployment/%DEPLOYMENT_NAME%
if errorlevel 1 (
    echo [WARN] Rollout restart failed. You may need to manually delete pods.
) else (
    echo [OK] Rollout restarted.
)

echo.
echo Waiting for rollout to complete...
kubectl rollout status deployment/%DEPLOYMENT_NAME% --timeout=180s
if errorlevel 1 (
    echo [WARN] Rollout did not complete within timeout. Check pod status.
)

echo.
echo ============================================
echo  Deployment completed!
echo ============================================
endlocal
exit /b 0
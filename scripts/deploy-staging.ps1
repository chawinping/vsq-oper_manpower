# Staging Deployment Script
# Usage: .\scripts\deploy-staging.ps1

param(
    [switch]$Build = $false,
    [switch]$Pull = $false
)

$ErrorActionPreference = "Stop"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "VSQ Operations Manpower - Staging Deployment" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Check if .env.staging exists
if (-not (Test-Path ".env.staging")) {
    Write-Host "ERROR: .env.staging file not found!" -ForegroundColor Red
    Write-Host "Please copy .env.staging.example to .env.staging and update with actual values." -ForegroundColor Yellow
    exit 1
}

# Load environment variables
Write-Host "Loading environment variables from .env.staging..." -ForegroundColor Green
Get-Content ".env.staging" | ForEach-Object {
    if ($_ -match '^\s*([^#][^=]+)=(.*)$') {
        $key = $matches[1].Trim()
        $value = $matches[2].Trim()
        if ($value -and $key) {
            [Environment]::SetEnvironmentVariable($key, $value, "Process")
        }
    }
}

# Pull latest images if requested
if ($Pull) {
    Write-Host "Pulling latest images..." -ForegroundColor Green
    docker-compose -f docker-compose.yml -f docker-compose.staging.yml pull
}

# Build images if requested
if ($Build) {
    Write-Host "Building images..." -ForegroundColor Green
    docker-compose -f docker-compose.yml -f docker-compose.staging.yml build --no-cache
}

# Stop existing containers
Write-Host "Stopping existing containers..." -ForegroundColor Yellow
docker-compose -f docker-compose.yml -f docker-compose.staging.yml down

# Start services
Write-Host "Starting staging services..." -ForegroundColor Green
docker-compose -f docker-compose.yml -f docker-compose.staging.yml up -d

# Wait for services to be healthy
Write-Host "Waiting for services to be healthy..." -ForegroundColor Yellow
Start-Sleep -Seconds 10

# Check health
Write-Host "Checking service health..." -ForegroundColor Green
$backendHealth = docker exec vsq-manpower-backend-staging wget -q -O- http://localhost:8080/health 2>$null
if ($backendHealth -match "ok") {
    Write-Host "✓ Backend is healthy" -ForegroundColor Green
} else {
    Write-Host "✗ Backend health check failed" -ForegroundColor Red
}

$frontendHealth = docker exec vsq-manpower-frontend-staging wget -q -O- http://localhost:3000/ 2>$null
if ($frontendHealth) {
    Write-Host "✓ Frontend is healthy" -ForegroundColor Green
} else {
    Write-Host "✗ Frontend health check failed" -ForegroundColor Red
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Deployment Complete!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Services:" -ForegroundColor Yellow
Write-Host "  - Frontend: http://localhost:$env:FRONTEND_PORT" -ForegroundColor White
Write-Host "  - Backend API: http://localhost:$env:BACKEND_PORT/api" -ForegroundColor White
Write-Host "  - Nginx: http://localhost:$env:NGINX_PORT" -ForegroundColor White
Write-Host ""
Write-Host "View logs: docker-compose -f docker-compose.yml -f docker-compose.staging.yml logs -f" -ForegroundColor Gray
Write-Host ""







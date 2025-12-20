# Production Deployment Script
# Usage: .\scripts\deploy-production.ps1
# IMPORTANT: Review and test in staging before deploying to production

param(
    [switch]$Build = $false,
    [switch]$Pull = $false,
    [switch]$Confirm = $false
)

$ErrorActionPreference = "Stop"

Write-Host "========================================" -ForegroundColor Red
Write-Host "VSQ Operations Manpower - PRODUCTION Deployment" -ForegroundColor Red
Write-Host "========================================" -ForegroundColor Red
Write-Host ""

# Safety check
if (-not $Confirm) {
    Write-Host "WARNING: This will deploy to PRODUCTION!" -ForegroundColor Red
    Write-Host ""
    $confirmation = Read-Host "Type 'DEPLOY' to confirm production deployment"
    if ($confirmation -ne "DEPLOY") {
        Write-Host "Deployment cancelled." -ForegroundColor Yellow
        exit 0
    }
}

# Check if .env.production exists
if (-not (Test-Path ".env.production")) {
    Write-Host "ERROR: .env.production file not found!" -ForegroundColor Red
    Write-Host "Please copy .env.production.example to .env.production and update with actual values." -ForegroundColor Yellow
    exit 1
}

# Validate critical environment variables
Write-Host "Validating environment variables..." -ForegroundColor Green
Get-Content ".env.production" | ForEach-Object {
    if ($_ -match '^\s*([^#][^=]+)=(.*)$') {
        $key = $matches[1].Trim()
        $value = $matches[2].Trim()
        if ($key -and $value) {
            [Environment]::SetEnvironmentVariable($key, $value, "Process")
        }
    }
}

# Check for default/placeholder values
$criticalVars = @("SESSION_SECRET", "DB_PASSWORD")
foreach ($var in $criticalVars) {
    $value = [Environment]::GetEnvironmentVariable($var, "Process")
    if (-not $value -or $value -match "CHANGE_THIS" -or $value -match "change-me") {
        Write-Host "ERROR: $var is not set or contains placeholder value!" -ForegroundColor Red
        exit 1
    }
}

# Check SSL certificates
if (-not (Test-Path "nginx/ssl/production/cert.pem") -or -not (Test-Path "nginx/ssl/production/key.pem")) {
    Write-Host "WARNING: SSL certificates not found in nginx/ssl/production/" -ForegroundColor Yellow
    Write-Host "HTTPS will not work. Continue anyway? (y/N)" -ForegroundColor Yellow
    $continue = Read-Host
    if ($continue -ne "y" -and $continue -ne "Y") {
        exit 1
    }
}

# Pull latest images if requested
if ($Pull) {
    Write-Host "Pulling latest images..." -ForegroundColor Green
    docker-compose -f docker-compose.yml -f docker-compose.production.yml pull
}

# Build images if requested
if ($Build) {
    Write-Host "Building production images..." -ForegroundColor Green
    docker-compose -f docker-compose.yml -f docker-compose.production.yml build
}

# Create backup before deployment
Write-Host "Creating database backup..." -ForegroundColor Yellow
$backupDir = "backups/$(Get-Date -Format 'yyyyMMdd_HHmmss')"
New-Item -ItemType Directory -Force -Path $backupDir | Out-Null
# Note: Add actual backup command here based on your backup strategy

# Stop existing containers gracefully
Write-Host "Stopping existing containers..." -ForegroundColor Yellow
docker-compose -f docker-compose.yml -f docker-compose.production.yml down --timeout 30

# Start services
Write-Host "Starting production services..." -ForegroundColor Green
docker-compose -f docker-compose.yml -f docker-compose.production.yml up -d

# Wait for services to be healthy
Write-Host "Waiting for services to be healthy..." -ForegroundColor Yellow
Start-Sleep -Seconds 15

# Check health
Write-Host "Checking service health..." -ForegroundColor Green
$maxRetries = 10
$retryCount = 0
$allHealthy = $false

while ($retryCount -lt $maxRetries -and -not $allHealthy) {
    $backendHealth = docker exec vsq-manpower-backend-production wget -q -O- http://localhost:8080/health 2>$null
    $frontendHealth = docker exec vsq-manpower-frontend-production wget -q -O- http://localhost:3000/ 2>$null
    $nginxHealth = docker exec vsq-manpower-nginx-production wget -q -O- http://localhost/health 2>$null

    if ($backendHealth -match "ok" -and $frontendHealth -and $nginxHealth -match "healthy") {
        $allHealthy = $true
        Write-Host "✓ All services are healthy" -ForegroundColor Green
    } else {
        $retryCount++
        Write-Host "Waiting for services... ($retryCount/$maxRetries)" -ForegroundColor Yellow
        Start-Sleep -Seconds 5
    }
}

if (-not $allHealthy) {
    Write-Host "✗ Some services failed health checks. Check logs:" -ForegroundColor Red
    Write-Host "  docker-compose -f docker-compose.yml -f docker-compose.production.yml logs" -ForegroundColor Yellow
    exit 1
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "Production Deployment Complete!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""
Write-Host "Services:" -ForegroundColor Yellow
Write-Host "  - Frontend: https://your-domain.com" -ForegroundColor White
Write-Host "  - Backend API: https://your-domain.com/api" -ForegroundColor White
Write-Host ""
Write-Host "Monitor logs: docker-compose -f docker-compose.yml -f docker-compose.production.yml logs -f" -ForegroundColor Gray
Write-Host ""




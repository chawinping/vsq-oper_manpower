# Recreate Container with Current postgres:18-alpine Image
# Ensures container uses the latest postgres:18-alpine image tag

param(
    [switch]$BackupFirst
)

$ErrorActionPreference = "Stop"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Recreate Container with Current Image" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Check current state
Write-Host "[1/6] Checking current state..." -ForegroundColor Yellow
$currentImage = docker images postgres:18-alpine --format "{{.ID}}"
$containerImage = docker inspect vsq-manpower-db --format "{{.Image}}" 2>&1 | Out-String

Write-Host "Current postgres:18-alpine tag: $currentImage" -ForegroundColor Gray
Write-Host "Container's image ID: $($containerImage.Trim())" -ForegroundColor Gray

if ($containerImage -match $currentImage) {
    Write-Host "[OK] Container is already using current image" -ForegroundColor Green
    exit 0
}

Write-Host "[INFO] Container is using different image ID" -ForegroundColor Yellow
Write-Host "       Will recreate container with current image tag" -ForegroundColor Yellow

# Backup if requested
if ($BackupFirst) {
    Write-Host ""
    Write-Host "[2/6] Creating backup..." -ForegroundColor Yellow
    docker exec vsq-manpower-db pg_dump -U vsq_user -d vsq_manpower -F c -f /tmp/backup.dump 2>&1 | Out-Null
    $backupFile = "recreate_backup_$(Get-Date -Format 'yyyyMMdd_HHmmss').dump"
    docker cp vsq-manpower-db:/tmp/backup.dump "./$backupFile"
    Write-Host "[OK] Backup created: $backupFile" -ForegroundColor Green
} else {
    Write-Host ""
    Write-Host "[2/6] Skipping backup (use -BackupFirst to backup)" -ForegroundColor Yellow
}

# Stop services
Write-Host ""
Write-Host "[3/6] Stopping services..." -ForegroundColor Yellow
docker-compose --profile dev stop backend-dev
docker-compose --profile dev stop postgres
Write-Host "[OK] Services stopped" -ForegroundColor Green

# Remove old container
Write-Host ""
Write-Host "[4/6] Removing old container..." -ForegroundColor Yellow
docker-compose --profile dev rm -f postgres
Write-Host "[OK] Container removed" -ForegroundColor Green

# Ensure we have the latest image
Write-Host ""
Write-Host "[5/6] Ensuring latest postgres:18-alpine image..." -ForegroundColor Yellow
docker pull postgres:18-alpine
Write-Host "[OK] Latest image ready" -ForegroundColor Green

# Recreate container with current image
Write-Host ""
Write-Host "[6/6] Recreating container with current image..." -ForegroundColor Yellow
docker-compose --profile dev up -d postgres

# Wait for database
Write-Host "Waiting for database to be ready..." -ForegroundColor Yellow
$maxAttempts = 30
$attempt = 0
$ready = $false

while ($attempt -lt $maxAttempts -and -not $ready) {
    Start-Sleep -Seconds 2
    $attempt++
    
    try {
        $result = docker exec vsq-manpower-db pg_isready -U vsq_user -d vsq_manpower 2>&1
        if ($LASTEXITCODE -eq 0) {
            $ready = $true
            Write-Host "[OK] Database is ready" -ForegroundColor Green
        }
    } catch {
        # Continue waiting
    }
}

# Verify
Write-Host ""
Write-Host "Verifying..." -ForegroundColor Cyan
$newImage = docker inspect vsq-manpower-db --format "{{.Image}}" 2>&1 | Out-String
$currentTag = docker images postgres:18-alpine --format "{{.ID}}"

Write-Host "Container image ID: $($newImage.Trim())" -ForegroundColor Gray
Write-Host "Current tag image ID: $currentTag" -ForegroundColor Gray

if ($newImage -match $currentTag) {
    Write-Host "[OK] Container now using current image!" -ForegroundColor Green
} else {
    Write-Host "[WARN] Image IDs don't match, but container is running" -ForegroundColor Yellow
}

# Check PostgreSQL version
$version = docker exec vsq-manpower-db psql -U vsq_user -d vsq_manpower -c "SELECT version();" 2>&1 | Out-String
Write-Host "PostgreSQL version: $($version.Trim())" -ForegroundColor Gray

# Restore backup if created
if ($BackupFirst -and (Test-Path "./$backupFile")) {
    Write-Host ""
    Write-Host "Restoring data from backup..." -ForegroundColor Yellow
    docker cp "./$backupFile" vsq-manpower-db:/tmp/backup.dump
    docker exec vsq-manpower-db pg_restore -U vsq_user -d vsq_manpower --data-only --disable-triggers /tmp/backup.dump 2>&1 | Out-Null
    Write-Host "[OK] Data restored" -ForegroundColor Green
}

# Start backend
Write-Host ""
Write-Host "Starting backend..." -ForegroundColor Yellow
docker-compose --profile dev up -d backend-dev
Start-Sleep -Seconds 5

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Recreation Complete!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Container is now using the current postgres:18-alpine image tag" -ForegroundColor Green
Write-Host ""
Write-Host "Verify in Docker Desktop:" -ForegroundColor Yellow
Write-Host "  - Container should show: postgres:18-alpine" -ForegroundColor Gray
Write-Host "  - Image ID should match current tag" -ForegroundColor Gray

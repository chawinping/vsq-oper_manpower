# Development Database Migration to PostgreSQL 18
# This script migrates your development database from PostgreSQL 15 to PostgreSQL 18

param(
    [switch]$PreserveData,
    [switch]$SkipBackup
)

$ErrorActionPreference = "Stop"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Development Database Migration" -ForegroundColor Cyan
Write-Host "PostgreSQL 15 -> PostgreSQL 18" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Check if Docker is running
Write-Host "[1/8] Checking Docker..." -ForegroundColor Yellow
try {
    docker ps | Out-Null
    Write-Host "[OK] Docker is running" -ForegroundColor Green
} catch {
    Write-Host "[FAIL] Docker is not running" -ForegroundColor Red
    exit 1
}

# Check current database state
Write-Host ""
Write-Host "[2/8] Checking current database state..." -ForegroundColor Yellow
docker-compose --profile dev up -d postgres
Start-Sleep -Seconds 5

try {
    $version = docker exec vsq-manpower-db psql -U vsq_user -d vsq_manpower -t -c "SELECT version();" 2>&1 | Out-String
    if ($version -match "PostgreSQL (\d+)") {
        $currentVersion = $matches[1]
        Write-Host "[OK] Current PostgreSQL version: $currentVersion" -ForegroundColor Green
        
        if ($currentVersion -eq "18") {
            Write-Host "[INFO] Already on PostgreSQL 18. No migration needed." -ForegroundColor Yellow
            exit 0
        }
    }
} catch {
    Write-Host "[WARN] Could not check version (database might be empty): $_" -ForegroundColor Yellow
}

# Backup database
if (-not $SkipBackup) {
    Write-Host ""
    Write-Host "[3/8] Creating backup..." -ForegroundColor Yellow
    try {
        $backupFile = "dev_backup_$(Get-Date -Format 'yyyyMMdd_HHmmss').dump"
        docker exec vsq-manpower-db pg_dump -U vsq_user -d vsq_manpower -F c -f /tmp/dev_backup.dump 2>&1 | Out-Null
        
        if ($LASTEXITCODE -eq 0) {
            docker cp vsq-manpower-db:/tmp/dev_backup.dump "./$backupFile"
            Write-Host "[OK] Backup created: $backupFile" -ForegroundColor Green
        } else {
            Write-Host "[WARN] Backup failed (database might be empty)" -ForegroundColor Yellow
        }
    } catch {
        Write-Host "[WARN] Backup failed: $_" -ForegroundColor Yellow
        Write-Host "       Continuing without backup..." -ForegroundColor Yellow
    }
} else {
    Write-Host ""
    Write-Host "[3/8] Skipping backup (--SkipBackup flag)" -ForegroundColor Yellow
}

# Stop services
Write-Host ""
Write-Host "[4/8] Stopping services..." -ForegroundColor Yellow
docker-compose --profile dev down
Write-Host "[OK] Services stopped" -ForegroundColor Green

# Check docker-compose.yml
Write-Host ""
Write-Host "[5/8] Checking docker-compose.yml..." -ForegroundColor Yellow
$composeContent = Get-Content docker-compose.yml -Raw
if ($composeContent -match "image: postgres:15-alpine") {
    Write-Host "[WARN] docker-compose.yml still uses PostgreSQL 15" -ForegroundColor Yellow
    Write-Host "       Please update docker-compose.yml:" -ForegroundColor Yellow
    Write-Host "       Change: image: postgres:15-alpine" -ForegroundColor Gray
    Write-Host "       To:     image: postgres:18-alpine" -ForegroundColor Gray
    Write-Host ""
    $confirm = Read-Host "Have you updated docker-compose.yml? (y/n)"
    if ($confirm -ne "y") {
        Write-Host "[FAIL] Please update docker-compose.yml first" -ForegroundColor Red
        exit 1
    }
} elseif ($composeContent -match "image: postgres:18-alpine") {
    Write-Host "[OK] docker-compose.yml already configured for PostgreSQL 18" -ForegroundColor Green
} else {
    Write-Host "[WARN] Could not detect PostgreSQL version in docker-compose.yml" -ForegroundColor Yellow
}

# Handle data preservation
if ($PreserveData) {
    Write-Host ""
    Write-Host "[6/8] Preserving data..." -ForegroundColor Yellow
    
    # Export data only
    docker-compose --profile dev up -d postgres
    Start-Sleep -Seconds 5
    
    $dataFile = "dev_data_$(Get-Date -Format 'yyyyMMdd_HHmmss').dump"
    docker exec vsq-manpower-db pg_dump -U vsq_user -d vsq_manpower --data-only -F c -f /tmp/dev_data.dump 2>&1 | Out-Null
    docker cp vsq-manpower-db:/tmp/dev_data.dump "./$dataFile"
    
    docker-compose --profile dev down
    Write-Host "[OK] Data exported: $dataFile" -ForegroundColor Green
} else {
    Write-Host ""
    Write-Host "[6/8] Fresh start (data will be lost)..." -ForegroundColor Yellow
    Write-Host "      To preserve data, use: -PreserveData flag" -ForegroundColor Gray
    
    # Remove old volume
    $confirm = Read-Host "Remove old database volume? (y/n)"
    if ($confirm -eq "y") {
        docker volume rm vsq-oper_manpower_postgres_data -f 2>&1 | Out-Null
        Write-Host "[OK] Old volume removed" -ForegroundColor Green
    }
}

# Start PostgreSQL 18
Write-Host ""
Write-Host "[7/8] Starting PostgreSQL 18..." -ForegroundColor Yellow
docker-compose --profile dev up -d postgres

# Wait for database to be ready
Write-Host "Waiting for PostgreSQL 18 to be ready..." -ForegroundColor Yellow
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
            Write-Host "[OK] PostgreSQL 18 is ready" -ForegroundColor Green
        }
    } catch {
        # Continue waiting
    }
}

if (-not $ready) {
    Write-Host "[FAIL] PostgreSQL 18 failed to start" -ForegroundColor Red
    exit 1
}

# Verify version
$version = docker exec vsq-manpower-db psql -U vsq_user -d vsq_manpower -t -c "SELECT version();" 2>&1 | Out-String
if ($version -match "PostgreSQL 18") {
    Write-Host "[OK] PostgreSQL 18 confirmed" -ForegroundColor Green
} else {
    Write-Host "[FAIL] Unexpected version: $version" -ForegroundColor Red
    exit 1
}

# Start backend (runs migrations)
Write-Host ""
Write-Host "[8/8] Starting backend (migrations will run)..." -ForegroundColor Yellow
docker-compose --profile dev up -d backend-dev

# Wait for backend
Start-Sleep -Seconds 10

# Import data if preserved
if ($PreserveData -and (Test-Path "./$dataFile")) {
    Write-Host ""
    Write-Host "Importing data..." -ForegroundColor Yellow
    docker cp "./$dataFile" vsq-manpower-db:/tmp/dev_data.dump
    docker exec vsq-manpower-db pg_restore -U vsq_user -d vsq_manpower --data-only /tmp/dev_data.dump 2>&1 | Out-Null
    Write-Host "[OK] Data imported" -ForegroundColor Green
}

# Verify migration
Write-Host ""
Write-Host "Verifying migration..." -ForegroundColor Cyan
.\scripts\verify-database-state.ps1 -ContainerName "vsq-manpower-db" -DatabaseName "vsq_manpower" -Detailed

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Migration Complete!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Yellow
Write-Host "  1. Test your application: http://localhost:4000" -ForegroundColor Gray
Write-Host "  2. Verify all features work" -ForegroundColor Gray
Write-Host "  3. Check backend logs: docker-compose --profile dev logs backend-dev" -ForegroundColor Gray
Write-Host ""
Write-Host "If you need to rollback:" -ForegroundColor Yellow
Write-Host "  - Restore from backup: dev_backup_*.dump" -ForegroundColor Gray
Write-Host "  - Or revert docker-compose.yml to postgres:15-alpine" -ForegroundColor Gray

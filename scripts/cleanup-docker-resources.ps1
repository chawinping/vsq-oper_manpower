# Docker Resources Cleanup Script
# Removes unused Docker images, containers, and volumes

param(
    [switch]$RemoveOldPostgres,
    [switch]$RemoveTestResources,
    [switch]$DryRun
)

$ErrorActionPreference = "Stop"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Docker Resources Cleanup" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

if ($DryRun) {
    Write-Host "[DRY RUN MODE] - No changes will be made" -ForegroundColor Yellow
    Write-Host ""
}

# Check Docker
Write-Host "[1/5] Checking Docker..." -ForegroundColor Yellow
try {
    docker ps | Out-Null
    Write-Host "[OK] Docker is running" -ForegroundColor Green
} catch {
    Write-Host "[FAIL] Docker is not running" -ForegroundColor Red
    exit 1
}

# Remove old PostgreSQL 15 image
if ($RemoveOldPostgres) {
    Write-Host ""
    Write-Host "[2/5] Removing old PostgreSQL 15 image..." -ForegroundColor Yellow
    
    $oldImage = docker images postgres:15-alpine --format "{{.ID}}"
    if ($oldImage) {
        if ($DryRun) {
            Write-Host "[DRY RUN] Would remove: postgres:15-alpine ($oldImage)" -ForegroundColor Gray
        } else {
            docker rmi postgres:15-alpine 2>&1 | Out-Null
            if ($LASTEXITCODE -eq 0) {
                Write-Host "[OK] Removed postgres:15-alpine" -ForegroundColor Green
            } else {
                Write-Host "[WARN] Could not remove postgres:15-alpine (might be in use)" -ForegroundColor Yellow
            }
        }
    } else {
        Write-Host "[INFO] postgres:15-alpine not found" -ForegroundColor Gray
    }
} else {
    Write-Host ""
    Write-Host "[2/5] Skipping PostgreSQL 15 removal (use -RemoveOldPostgres)" -ForegroundColor Yellow
}

# Remove test resources
if ($RemoveTestResources) {
    Write-Host ""
    Write-Host "[3/5] Removing test containers and volumes..." -ForegroundColor Yellow
    
    # Stop test containers
    if ($DryRun) {
        Write-Host "[DRY RUN] Would stop test containers" -ForegroundColor Gray
    } else {
        docker-compose -f docker-compose.yml -f docker-compose.pg18-test.yml down --remove-orphans 2>&1 | Out-Null
        Write-Host "[OK] Test containers stopped" -ForegroundColor Green
    }
    
    # Remove test volume
    $testVolume = docker volume ls --format "{{.Name}}" | Select-String "postgres_pg18_test"
    if ($testVolume) {
        if ($DryRun) {
            Write-Host "[DRY RUN] Would remove volume: $testVolume" -ForegroundColor Gray
        } else {
            docker volume rm $testVolume 2>&1 | Out-Null
            if ($LASTEXITCODE -eq 0) {
                Write-Host "[OK] Removed test volume: $testVolume" -ForegroundColor Green
            } else {
                Write-Host "[WARN] Could not remove test volume" -ForegroundColor Yellow
            }
        }
    }
    
    # Remove test backend image
    $testImage = docker images vsq-oper_manpower-backend-test --format "{{.ID}}"
    if ($testImage) {
        if ($DryRun) {
            Write-Host "[DRY RUN] Would remove: vsq-oper_manpower-backend-test" -ForegroundColor Gray
        } else {
            docker rmi vsq-oper_manpower-backend-test:latest 2>&1 | Out-Null
            if ($LASTEXITCODE -eq 0) {
                Write-Host "[OK] Removed test backend image" -ForegroundColor Green
            } else {
                Write-Host "[WARN] Could not remove test backend image (might be in use)" -ForegroundColor Yellow
            }
        }
    }
} else {
    Write-Host ""
    Write-Host "[3/5] Skipping test resources removal (use -RemoveTestResources)" -ForegroundColor Yellow
}

# Show current disk usage
Write-Host ""
Write-Host "[4/5] Current Docker disk usage..." -ForegroundColor Yellow
docker system df

# Summary
Write-Host ""
Write-Host "[5/5] Summary" -ForegroundColor Yellow
Write-Host "========================================" -ForegroundColor Cyan

if ($DryRun) {
    Write-Host "[DRY RUN] No changes were made" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "To actually clean up, run without -DryRun:" -ForegroundColor Gray
    Write-Host "  .\scripts\cleanup-docker-resources.ps1 -RemoveOldPostgres -RemoveTestResources" -ForegroundColor Gray
} else {
    Write-Host "[OK] Cleanup completed" -ForegroundColor Green
    Write-Host ""
    Write-Host "To see space saved, run:" -ForegroundColor Gray
    Write-Host "  docker system df" -ForegroundColor Gray
}

Write-Host ""
Write-Host "Active development resources (kept):" -ForegroundColor Green
Write-Host "  - postgres:18-alpine" -ForegroundColor Gray
Write-Host "  - vsq-oper_manpower-backend-dev" -ForegroundColor Gray
Write-Host "  - vsq-oper_manpower-frontend-dev" -ForegroundColor Gray
Write-Host "  - vsq-oper_manpower_postgres_data (volume)" -ForegroundColor Gray

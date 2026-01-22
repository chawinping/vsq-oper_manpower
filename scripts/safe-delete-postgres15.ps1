# Safe Delete PostgreSQL 15 Image
# Checks if postgres:15-alpine is safe to delete

$ErrorActionPreference = "Stop"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "PostgreSQL 15 Image Deletion Check" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Check if image exists
Write-Host "[1/5] Checking if postgres:15-alpine exists..." -ForegroundColor Yellow
$imageExists = docker images postgres:15-alpine --format "{{.ID}}" 2>&1
if (-not $imageExists -or $imageExists -match "error") {
    Write-Host "[INFO] postgres:15-alpine not found. Nothing to delete." -ForegroundColor Yellow
    exit 0
}
Write-Host "[OK] Image exists: $imageExists" -ForegroundColor Green

# Check if any containers are using it
Write-Host ""
Write-Host "[2/5] Checking for containers using postgres:15-alpine..." -ForegroundColor Yellow
$containers = docker ps -a --filter "ancestor=postgres:15-alpine" --format "{{.Names}}" 2>&1
if ($containers) {
    Write-Host "[WARN] Found containers using postgres:15-alpine:" -ForegroundColor Yellow
    $containers | ForEach-Object { Write-Host "  - $_" -ForegroundColor Gray }
    Write-Host ""
    Write-Host "[FAIL] Cannot delete - containers are using this image" -ForegroundColor Red
    Write-Host "       Stop and remove containers first, then try again" -ForegroundColor Yellow
    exit 1
} else {
    Write-Host "[OK] No containers using postgres:15-alpine" -ForegroundColor Green
}

# Check docker-compose files
Write-Host ""
Write-Host "[3/5] Checking docker-compose files..." -ForegroundColor Yellow
$composeFiles = @("docker-compose.yml", "docker-compose.staging.yml", "docker-compose.production.yml")
$foundInFiles = @()

foreach ($file in $composeFiles) {
    if (Test-Path $file) {
        $content = Get-Content $file -Raw
        if ($content -match "postgres:15-alpine") {
            $foundInFiles += $file
            Write-Host "[WARN] Found in: $file" -ForegroundColor Yellow
        }
    }
}

if ($foundInFiles.Count -gt 0) {
    Write-Host ""
    Write-Host "[WARN] postgres:15-alpine is referenced in:" -ForegroundColor Yellow
    $foundInFiles | ForEach-Object { Write-Host "  - $_" -ForegroundColor Gray }
    Write-Host ""
    Write-Host "[INFO] These files reference PostgreSQL 15" -ForegroundColor Yellow
    Write-Host "       If you're sure you won't use them, you can delete the image" -ForegroundColor Gray
    Write-Host "       But consider updating these files first" -ForegroundColor Gray
} else {
    Write-Host "[OK] No docker-compose files reference postgres:15-alpine" -ForegroundColor Green
}

# Check image size
Write-Host ""
Write-Host "[4/5] Checking image size..." -ForegroundColor Yellow
$imageSize = docker images postgres:15-alpine --format "{{.Size}}"
Write-Host "Image size: $imageSize" -ForegroundColor Gray

# Summary and recommendation
Write-Host ""
Write-Host "[5/5] Summary and Recommendation" -ForegroundColor Yellow
Write-Host "========================================" -ForegroundColor Cyan

if ($containers) {
    Write-Host "[FAIL] Cannot delete - containers are using it" -ForegroundColor Red
    Write-Host ""
    Write-Host "To delete:" -ForegroundColor Yellow
    Write-Host "  1. Stop containers: docker ps -a --filter 'ancestor=postgres:15-alpine' --format '{{.Names}}' | ForEach-Object { docker stop $_ }" -ForegroundColor Gray
    Write-Host "  2. Remove containers: docker ps -a --filter 'ancestor=postgres:15-alpine' --format '{{.Names}}' | ForEach-Object { docker rm $_ }" -ForegroundColor Gray
    Write-Host "  3. Then delete image: docker rmi postgres:15-alpine" -ForegroundColor Gray
} else {
    Write-Host "[OK] Safe to delete!" -ForegroundColor Green
    Write-Host ""
    Write-Host "You can delete postgres:15-alpine to free up $imageSize" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "To delete:" -ForegroundColor Yellow
    Write-Host "  docker rmi postgres:15-alpine" -ForegroundColor Gray
    Write-Host ""
    
    if ($foundInFiles.Count -gt 0) {
        Write-Host "[NOTE] Consider updating these files first:" -ForegroundColor Yellow
        $foundInFiles | ForEach-Object { Write-Host "  - $_" -ForegroundColor Gray }
    }
    
    $confirm = Read-Host "`nDelete postgres:15-alpine now? (y/n)"
    if ($confirm -eq "y") {
        Write-Host ""
        Write-Host "Deleting postgres:15-alpine..." -ForegroundColor Yellow
        docker rmi postgres:15-alpine
        if ($LASTEXITCODE -eq 0) {
            Write-Host "[OK] Deleted successfully! Freed up $imageSize" -ForegroundColor Green
        } else {
            Write-Host "[FAIL] Could not delete image" -ForegroundColor Red
        }
    } else {
        Write-Host "[INFO] Deletion cancelled" -ForegroundColor Yellow
    }
}

Write-Host ""
Write-Host "Current postgres images:" -ForegroundColor Cyan
docker images postgres --format "table {{.Repository}}\t{{.Tag}}\t{{.Size}}"

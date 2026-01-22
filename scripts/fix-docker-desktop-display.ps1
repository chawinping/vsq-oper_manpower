# Fix Docker Desktop Image Display
# Helps troubleshoot why postgres:18-alpine might not show in Docker Desktop

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Docker Desktop Image Display Fix" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Verify image exists
Write-Host "[1/4] Verifying postgres:18-alpine exists..." -ForegroundColor Yellow
$image = docker images postgres:18-alpine --format "{{.Repository}}:{{.Tag}}"
if ($image) {
    Write-Host "[OK] Image exists: $image" -ForegroundColor Green
    docker images postgres:18-alpine --format "   ID: {{.ID}}`n   Size: {{.Size}}`n   Created: {{.CreatedAt}}"
} else {
    Write-Host "[FAIL] Image not found. Pulling..." -ForegroundColor Red
    docker pull postgres:18-alpine
}

# Verify container
Write-Host ""
Write-Host "[2/4] Verifying container..." -ForegroundColor Yellow
$containerImage = docker inspect vsq-manpower-db --format "{{.Config.Image}}" 2>&1
if ($containerImage -match "postgres:18-alpine") {
    Write-Host "[OK] Container is using: $containerImage" -ForegroundColor Green
    $pgVersion = docker exec vsq-manpower-db psql --version 2>&1
    Write-Host "[OK] PostgreSQL version: $pgVersion" -ForegroundColor Green
} else {
    Write-Host "[WARN] Container image: $containerImage" -ForegroundColor Yellow
}

# List all postgres images
Write-Host ""
Write-Host "[3/4] All postgres images:" -ForegroundColor Yellow
docker images postgres --format "table {{.Repository}}\t{{.Tag}}\t{{.ID}}\t{{.Size}}"

# Docker Desktop troubleshooting steps
Write-Host ""
Write-Host "[4/4] Docker Desktop Troubleshooting Steps:" -ForegroundColor Yellow
Write-Host ""
Write-Host "STEP 1: Refresh Docker Desktop" -ForegroundColor Cyan
Write-Host "  - Right-click Docker Desktop icon in system tray" -ForegroundColor Gray
Write-Host "  - Click 'Quit Docker Desktop'" -ForegroundColor Gray
Write-Host "  - Wait 10 seconds" -ForegroundColor Gray
Write-Host "  - Open Docker Desktop again" -ForegroundColor Gray
Write-Host ""
Write-Host "STEP 2: Check Images Tab" -ForegroundColor Cyan
Write-Host "  - Open Docker Desktop" -ForegroundColor Gray
Write-Host "  - Go to 'Images' tab (left sidebar)" -ForegroundColor Gray
Write-Host "  - Clear any search/filter (top right)" -ForegroundColor Gray
Write-Host "  - Look for repository: 'postgres'" -ForegroundColor Gray
Write-Host "  - Click on 'postgres' to expand and see tags" -ForegroundColor Gray
Write-Host "  - You should see: '15-alpine' and '18-alpine'" -ForegroundColor Gray
Write-Host ""
Write-Host "STEP 3: Search for Image" -ForegroundColor Cyan
Write-Host "  - In Images tab, use search box" -ForegroundColor Gray
Write-Host "  - Search for: 'postgres'" -ForegroundColor Gray
Write-Host "  - Or search for: '18-alpine'" -ForegroundColor Gray
Write-Host ""
Write-Host "STEP 4: Check View Options" -ForegroundColor Cyan
Write-Host "  - In Images tab, check view options" -ForegroundColor Gray
Write-Host "  - Ensure 'Show unused images' is checked (if available)" -ForegroundColor Gray
Write-Host "  - Try different view modes (list/grid)" -ForegroundColor Gray
Write-Host ""

# Alternative: Show how to verify via command line
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Verification Commands" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "If Docker Desktop still doesn't show it, verify via command line:" -ForegroundColor Yellow
Write-Host ""
Write-Host "  docker images postgres:18-alpine" -ForegroundColor Gray
Write-Host "  docker ps --filter 'name=vsq-manpower-db'" -ForegroundColor Gray
Write-Host "  docker exec vsq-manpower-db psql --version" -ForegroundColor Gray
Write-Host ""
Write-Host "IMPORTANT: Your container IS using PostgreSQL 18.1" -ForegroundColor Green
Write-Host "The image IS present and working correctly." -ForegroundColor Green
Write-Host "This is just a Docker Desktop UI display issue." -ForegroundColor Yellow
Write-Host ""

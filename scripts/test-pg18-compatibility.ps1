# PostgreSQL 18 Compatibility Test Script
# This script tests the application with PostgreSQL 18

param(
    [switch]$Clean,
    [switch]$SkipTests,
    [switch]$Verbose
)

$ErrorActionPreference = "Stop"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "PostgreSQL 18 Compatibility Test" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Check if Docker is running
Write-Host "[1/7] Checking Docker..." -ForegroundColor Yellow
try {
    docker ps | Out-Null
    Write-Host "[OK] Docker is running" -ForegroundColor Green
} catch {
    Write-Host "[FAIL] Docker is not running. Please start Docker Desktop." -ForegroundColor Red
    exit 1
}

# Clean up if requested
if ($Clean) {
    Write-Host ""
    Write-Host "[Cleanup] Removing existing test containers and volumes..." -ForegroundColor Yellow
    docker-compose -f docker-compose.yml -f docker-compose.pg18-test.yml down -v 2>$null
    Write-Host "[OK] Cleanup complete" -ForegroundColor Green
}

# Start PostgreSQL 18
Write-Host ""
Write-Host "[2/7] Starting PostgreSQL 18..." -ForegroundColor Yellow
docker-compose -f docker-compose.yml -f docker-compose.pg18-test.yml up -d postgres

# Wait for PostgreSQL to be ready
Write-Host "Waiting for PostgreSQL 18 to be ready..." -ForegroundColor Yellow
$maxAttempts = 30
$attempt = 0
$ready = $false

while ($attempt -lt $maxAttempts -and -not $ready) {
    Start-Sleep -Seconds 2
    $attempt++
    
    try {
        $result = docker exec vsq-manpower-db-pg18-test pg_isready -U vsq_user -d vsq_manpower_pg18_test 2>&1
        if ($LASTEXITCODE -eq 0) {
            $ready = $true
            Write-Host "[OK] PostgreSQL 18 is ready" -ForegroundColor Green
        }
    } catch {
        # Continue waiting
    }
    
    if ($attempt -eq $maxAttempts) {
        Write-Host "[FAIL] PostgreSQL 18 failed to start within timeout" -ForegroundColor Red
        exit 1
    }
}

# Check PostgreSQL version
Write-Host ""
Write-Host "[3/7] Verifying PostgreSQL version..." -ForegroundColor Yellow
$version = docker exec vsq-manpower-db-pg18-test psql -U vsq_user -d vsq_manpower_pg18_test -t -c "SELECT version();" 2>&1
if ($version -match "PostgreSQL 18") {
    Write-Host "[OK] PostgreSQL 18 confirmed: $($version.Trim())" -ForegroundColor Green
} else {
    Write-Host "[FAIL] Unexpected PostgreSQL version: $version" -ForegroundColor Red
    exit 1
}

# Start backend with PostgreSQL 18
Write-Host ""
Write-Host "[4/7] Starting backend with PostgreSQL 18..." -ForegroundColor Yellow
# Start backend-test service - it will use the postgres service from the merged compose files
docker-compose -f docker-compose.yml -f docker-compose.pg18-test.yml --profile pg18-test up -d backend-test postgres

# Wait for backend to be ready
Write-Host "Waiting for backend to be ready..." -ForegroundColor Yellow
$maxAttempts = 30
$attempt = 0
$ready = $false

while ($attempt -lt $maxAttempts -and -not $ready) {
    Start-Sleep -Seconds 2
    $attempt++
    
    try {
        $response = Invoke-WebRequest -Uri "http://localhost:8082/health" -TimeoutSec 2 -ErrorAction SilentlyContinue
        if ($response.StatusCode -eq 200) {
            $ready = $true
            Write-Host "[OK] Backend is ready" -ForegroundColor Green
        }
    } catch {
        # Continue waiting
    }
    
    if ($attempt -eq $maxAttempts) {
        Write-Host "[FAIL] Backend failed to start within timeout" -ForegroundColor Red
        Write-Host "Checking backend logs..." -ForegroundColor Yellow
        docker-compose -f docker-compose.yml -f docker-compose.pg18-test.yml logs backend-test --tail=50
        exit 1
    }
}

# Check database migrations
Write-Host ""
Write-Host "[5/7] Checking database migrations..." -ForegroundColor Yellow
$migrationCheck = docker exec vsq-manpower-db-pg18-test psql -U vsq_user -d vsq_manpower_pg18_test -t -c "
    SELECT COUNT(*) FROM information_schema.tables 
    WHERE table_schema = 'public' AND table_type = 'BASE TABLE';
" 2>&1

if ($LASTEXITCODE -eq 0) {
    $tableCount = $migrationCheck.Trim()
    Write-Host "[OK] Migrations completed. Found $tableCount tables." -ForegroundColor Green
    
    # Check for key tables
    $keyTables = @("users", "roles", "branches", "staff", "positions")
    $missingTables = @()
    
    foreach ($table in $keyTables) {
        $exists = docker exec vsq-manpower-db-pg18-test psql -U vsq_user -d vsq_manpower_pg18_test -t -c "
            SELECT COUNT(*) FROM information_schema.tables 
            WHERE table_schema = 'public' AND table_name = '$table';
        " 2>&1 | Out-String
        
        if ($exists.Trim() -eq "0") {
            $missingTables += $table
        }
    }
    
    if ($missingTables.Count -gt 0) {
        Write-Host "[WARN] Missing tables: $($missingTables -join ', ')" -ForegroundColor Yellow
    } else {
        Write-Host "[OK] All key tables exist" -ForegroundColor Green
    }
} else {
    Write-Host "[FAIL] Failed to check migrations: $migrationCheck" -ForegroundColor Red
    exit 1
}

# Check custom functions
Write-Host ""
Write-Host "[6/7] Checking custom PL/pgSQL functions..." -ForegroundColor Yellow
$functions = @("get_revenue_level_tier", "scenario_matches")
$missingFunctions = @()

foreach ($func in $functions) {
    $exists = docker exec vsq-manpower-db-pg18-test psql -U vsq_user -d vsq_manpower_pg18_test -t -c "
        SELECT COUNT(*) FROM pg_proc p
        JOIN pg_namespace n ON p.pronamespace = n.oid
        WHERE n.nspname = 'public' AND p.proname = '$func';
    " 2>&1 | Out-String
    
    if ($exists.Trim() -eq "0") {
        $missingFunctions += $func
    }
}

if ($missingFunctions.Count -gt 0) {
    Write-Host "[WARN] Missing functions: $($missingFunctions -join ', ')" -ForegroundColor Yellow
} else {
    Write-Host "[OK] All custom functions exist" -ForegroundColor Green
}

# Run tests if not skipped
if (-not $SkipTests) {
    Write-Host ""
    Write-Host "[7/7] Running Go unit tests..." -ForegroundColor Yellow
    
    Push-Location backend
    
    try {
        # Set test database connection
        $env:DB_HOST = "localhost"
        $env:DB_PORT = "5435"
        $env:DB_USER = "vsq_user"
        $env:DB_PASSWORD = "vsq_password"
        $env:DB_NAME = "vsq_manpower_pg18_test"
        $env:DB_SSLMODE = "disable"
        
        Write-Host "Running tests against PostgreSQL 18..." -ForegroundColor Yellow
        go test ./tests/... -v
        
        if ($LASTEXITCODE -eq 0) {
            Write-Host "[OK] All tests passed" -ForegroundColor Green
        } else {
            Write-Host "[FAIL] Some tests failed" -ForegroundColor Red
            Pop-Location
            exit 1
        }
    } catch {
        Write-Host "[FAIL] Test execution failed: $_" -ForegroundColor Red
        Pop-Location
        exit 1
    } finally {
        Pop-Location
    }
} else {
    Write-Host ""
    Write-Host "[7/7] Skipping tests (--SkipTests flag)" -ForegroundColor Yellow
}

# Summary
Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Test Summary" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "PostgreSQL Version: PostgreSQL 18" -ForegroundColor Green
Write-Host "Database: vsq_manpower_pg18_test" -ForegroundColor Green
Write-Host "Backend Port: 8082" -ForegroundColor Green
Write-Host "Database Port: 5435" -ForegroundColor Green
Write-Host ""
Write-Host "[OK] Compatibility test completed successfully!" -ForegroundColor Green
Write-Host ""
Write-Host "To clean up test environment:" -ForegroundColor Yellow
Write-Host "  docker-compose -f docker-compose.yml -f docker-compose.pg18-test.yml down -v" -ForegroundColor Gray
Write-Host ""
Write-Host "To view logs:" -ForegroundColor Yellow
Write-Host "  docker-compose -f docker-compose.yml -f docker-compose.pg18-test.yml logs -f" -ForegroundColor Gray

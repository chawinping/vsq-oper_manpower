# Database State Verification Script
# Verifies current database state and migration readiness

param(
    [switch]$Detailed,
    [string]$ContainerName = "vsq-manpower-db",
    [string]$DatabaseName = "vsq_manpower"
)

$ErrorActionPreference = "Stop"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Database State Verification" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Check if container exists
Write-Host "[1/8] Checking database container..." -ForegroundColor Yellow

# Try to find the container
$containerExists = docker ps -a --filter "name=$ContainerName" --format "{{.Names}}" | Select-String -Pattern $ContainerName

# If not found, try to find any vsq postgres container
if (-not $containerExists) {
    $allContainers = docker ps -a --format "{{.Names}}"
    $vsqContainers = $allContainers | Select-String -Pattern "vsq.*db|postgres.*vsq"
    
    if ($vsqContainers) {
        Write-Host "[INFO] Container '$ContainerName' not found, but found VSQ containers:" -ForegroundColor Yellow
        $vsqContainers | ForEach-Object { Write-Host "  - $_" -ForegroundColor Gray }
        
        # Try the first one
        $firstContainer = ($vsqContainers | Select-Object -First 1).ToString().Trim()
        Write-Host "[INFO] Attempting to use: $firstContainer" -ForegroundColor Yellow
        $ContainerName = $firstContainer
    } else {
        Write-Host "[FAIL] Container '$ContainerName' not found" -ForegroundColor Red
        Write-Host "Available containers:" -ForegroundColor Yellow
        docker ps -a --format "{{.Names}}"
        Write-Host ""
        Write-Host "Usage: .\scripts\verify-database-state.ps1 -ContainerName <container-name>" -ForegroundColor Yellow
        exit 1
    }
}
Write-Host "[OK] Using container: $ContainerName" -ForegroundColor Green

# Check if container is running
$isRunning = docker ps --filter "name=$ContainerName" --format "{{.Names}}" | Select-String -Pattern $ContainerName
if (-not $isRunning) {
    Write-Host "[WARN] Container is not running. Starting..." -ForegroundColor Yellow
    docker start $ContainerName
    Start-Sleep -Seconds 5
}

# Get PostgreSQL version
Write-Host ""
Write-Host "[2/8] Checking PostgreSQL version..." -ForegroundColor Yellow
try {
    $versionOutput = docker exec $ContainerName psql -U vsq_user -d $DatabaseName -t -c "SELECT version();" 2>&1 | Out-String
    $versionOutput = $versionOutput.Trim()
    
    if ($versionOutput -match "PostgreSQL (\d+)") {
        $pgVersion = $matches[1]
        Write-Host "[OK] PostgreSQL version: $pgVersion" -ForegroundColor Green
        Write-Host "     Full version: $versionOutput" -ForegroundColor Gray
        
        if ($pgVersion -eq "15") {
            Write-Host "     Status: Current version (supported until Nov 2027)" -ForegroundColor Green
        } elseif ($pgVersion -eq "18") {
            Write-Host "     Status: PostgreSQL 18 (latest)" -ForegroundColor Green
        } else {
            Write-Host "     Status: Version $pgVersion" -ForegroundColor Yellow
        }
    } else {
        Write-Host "[FAIL] Could not parse version: $versionOutput" -ForegroundColor Red
        exit 1
    }
} catch {
    Write-Host "[FAIL] Could not connect to database: $_" -ForegroundColor Red
    exit 1
}

# Count tables
Write-Host ""
Write-Host "[3/8] Checking database tables..." -ForegroundColor Yellow
try {
    $tableCount = docker exec $ContainerName psql -U vsq_user -d $DatabaseName -t -c "
        SELECT COUNT(*) 
        FROM information_schema.tables 
        WHERE table_schema = 'public' AND table_type = 'BASE TABLE';
    " 2>&1 | Out-String
    
    $tableCount = $tableCount.Trim()
    Write-Host "[OK] Found $tableCount tables" -ForegroundColor Green
    
    if ($Detailed) {
        $tables = docker exec $ContainerName psql -U vsq_user -d $DatabaseName -t -c "
            SELECT table_name 
            FROM information_schema.tables 
            WHERE table_schema = 'public' AND table_type = 'BASE TABLE'
            ORDER BY table_name;
        " 2>&1 | Out-String
        
        Write-Host "Tables:" -ForegroundColor Gray
        $tables.Trim() -split "`n" | ForEach-Object { Write-Host "  - $_" -ForegroundColor Gray }
    }
    
    # Check key tables
    $keyTables = @("users", "roles", "branches", "staff", "positions")
    $missingTables = @()
    
    foreach ($table in $keyTables) {
        $exists = docker exec $ContainerName psql -U vsq_user -d $DatabaseName -t -c "
            SELECT COUNT(*) 
            FROM information_schema.tables 
            WHERE table_schema = 'public' AND table_name = '$table';
        " 2>&1 | Out-String
        
        if ($exists.Trim() -eq "0") {
            $missingTables += $table
        }
    }
    
    if ($missingTables.Count -gt 0) {
        Write-Host "[WARN] Missing key tables: $($missingTables -join ', ')" -ForegroundColor Yellow
    } else {
        Write-Host "[OK] All key tables exist" -ForegroundColor Green
    }
} catch {
    Write-Host "[FAIL] Error checking tables: $_" -ForegroundColor Red
}

# Check custom functions
Write-Host ""
Write-Host "[4/8] Checking custom PL/pgSQL functions..." -ForegroundColor Yellow
try {
    $functions = @("get_revenue_level_tier", "scenario_matches")
    $missingFunctions = @()
    
    foreach ($func in $functions) {
        $exists = docker exec $ContainerName psql -U vsq_user -d $DatabaseName -t -c "
            SELECT COUNT(*) 
            FROM pg_proc p
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
} catch {
    Write-Host "[FAIL] Error checking functions: $_" -ForegroundColor Red
}

# Check foreign key constraint
Write-Host ""
Write-Host "[5/8] Checking foreign key constraints..." -ForegroundColor Yellow
try {
    $constraintExists = docker exec $ContainerName psql -U vsq_user -d $DatabaseName -t -c "
        SELECT COUNT(*) 
        FROM information_schema.table_constraints 
        WHERE constraint_name = 'branches_area_manager_id_fkey' 
        AND table_name = 'branches';
    " 2>&1 | Out-String
    
    if ($constraintExists.Trim() -eq "1") {
        Write-Host "[OK] Foreign key constraint 'branches_area_manager_id_fkey' exists" -ForegroundColor Green
    } else {
        Write-Host "[WARN] Foreign key constraint 'branches_area_manager_id_fkey' is missing" -ForegroundColor Yellow
        Write-Host "       This will be added automatically on next migration run" -ForegroundColor Gray
    }
} catch {
    Write-Host "[FAIL] Error checking constraints: $_" -ForegroundColor Red
}

# Check database size
Write-Host ""
Write-Host "[6/8] Checking database size..." -ForegroundColor Yellow
try {
    $dbSize = docker exec $ContainerName psql -U vsq_user -d $DatabaseName -t -c "
        SELECT pg_size_pretty(pg_database_size('$DatabaseName'));
    " 2>&1 | Out-String
    
    Write-Host "[OK] Database size: $($dbSize.Trim())" -ForegroundColor Green
} catch {
    Write-Host "[WARN] Could not get database size: $_" -ForegroundColor Yellow
}

# Check row counts for key tables
Write-Host ""
Write-Host "[7/8] Checking data counts..." -ForegroundColor Yellow
try {
    $tablesToCheck = @("users", "roles", "branches", "staff", "positions")
    foreach ($table in $tablesToCheck) {
        $count = docker exec $ContainerName psql -U vsq_user -d $DatabaseName -t -c "
            SELECT COUNT(*) FROM $table;
        " 2>&1 | Out-String
        
        $count = $count.Trim()
        Write-Host "  $table`: $count rows" -ForegroundColor Gray
    }
} catch {
    Write-Host "[WARN] Could not get row counts: $_" -ForegroundColor Yellow
}

# Migration readiness check
Write-Host ""
Write-Host "[8/8] Migration readiness check..." -ForegroundColor Yellow
$readyForMigration = $true
$issues = @()

if ($pgVersion -eq "15") {
    Write-Host "[OK] Ready for PostgreSQL 18 migration" -ForegroundColor Green
} elseif ($pgVersion -eq "18") {
    Write-Host "[OK] Already on PostgreSQL 18" -ForegroundColor Green
} else {
    Write-Host "[WARN] Unusual PostgreSQL version: $pgVersion" -ForegroundColor Yellow
    $readyForMigration = $false
    $issues += "Unusual PostgreSQL version"
}

if ($tableCount -lt 30) {
    Write-Host "[WARN] Low table count - migrations may not have run" -ForegroundColor Yellow
    $issues += "Low table count"
}

if ($missingTables.Count -gt 0) {
    $readyForMigration = $false
    $issues += "Missing key tables"
}

# Summary
Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Verification Summary" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "PostgreSQL Version: $pgVersion" -ForegroundColor $(if ($pgVersion -eq "18") { "Green" } else { "Yellow" })
Write-Host "Tables: $tableCount" -ForegroundColor Green
Write-Host "Migration Ready: $(if ($readyForMigration) { '[OK] Yes' } else { '[WARN] Issues found' })" -ForegroundColor $(if ($readyForMigration) { "Green" } else { "Yellow" })

if ($issues.Count -gt 0) {
    Write-Host ""
    Write-Host "Issues found:" -ForegroundColor Yellow
    foreach ($issue in $issues) {
        Write-Host "  - $issue" -ForegroundColor Yellow
    }
}

Write-Host ""
Write-Host "Next steps:" -ForegroundColor Cyan
if ($pgVersion -eq "15") {
    Write-Host "  1. Review migration guide: docs/postgresql-18-migration-guide.md" -ForegroundColor Gray
    Write-Host "  2. Backup database: .\scripts\backup-database.ps1" -ForegroundColor Gray
    Write-Host "  3. Plan migration timeline" -ForegroundColor Gray
} elseif ($pgVersion -eq "18") {
    Write-Host "  - Database is already on PostgreSQL 18" -ForegroundColor Gray
    Write-Host "  - No migration needed" -ForegroundColor Gray
} else {
    Write-Host "  - Review database state" -ForegroundColor Gray
    Write-Host "  - Check for issues above" -ForegroundColor Gray
}

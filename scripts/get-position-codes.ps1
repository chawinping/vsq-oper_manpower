# Script to get all positions with their codes from the database
# Usage: .\scripts\get-position-codes.ps1

$DB_HOST = $env:DB_HOST
if (-not $DB_HOST) { $DB_HOST = "localhost" }

$DB_PORT = $env:DB_PORT
if (-not $DB_PORT) { $DB_PORT = "5434" }

$DB_USER = $env:DB_USER
if (-not $DB_USER) { $DB_USER = "vsq_user" }

$DB_PASSWORD = $env:DB_PASSWORD
if (-not $DB_PASSWORD) { $DB_PASSWORD = "vsq_password" }

$DB_NAME = $env:DB_NAME
if (-not $DB_NAME) { $DB_NAME = "vsq_manpower" }

Write-Host "Connecting to database: ${DB_NAME}@${DB_HOST}:${DB_PORT}" -ForegroundColor Cyan

# Check if psql is available
$psqlPath = Get-Command psql -ErrorAction SilentlyContinue
if (-not $psqlPath) {
    Write-Host "Error: psql command not found. Please install PostgreSQL client tools." -ForegroundColor Red
    Write-Host "Alternatively, you can use Docker to run the query:" -ForegroundColor Yellow
    Write-Host "docker exec -it vsq-manpower-db psql -U $DB_USER -d $DB_NAME -c `"SELECT id, name, position_code, position_type, display_order FROM positions ORDER BY display_order, name;`"" -ForegroundColor Yellow
    exit 1
}

# Set PGPASSWORD environment variable for psql
$env:PGPASSWORD = $DB_PASSWORD

# Query to get all positions with codes
$query = @"
SELECT 
    id,
    name,
    COALESCE(position_code, '') as position_code,
    position_type,
    display_order
FROM positions
ORDER BY display_order, name;
"@

Write-Host "`nQuerying positions..." -ForegroundColor Cyan
Write-Host "=" * 80 -ForegroundColor Gray

# Execute query
$result = psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -A -F "|" -c $query

if ($LASTEXITCODE -ne 0) {
    Write-Host "Error querying database. Trying Docker method..." -ForegroundColor Yellow
    
    # Try Docker method
    $dockerResult = docker exec -it vsq-manpower-db psql -U $DB_USER -d $DB_NAME -t -A -F "|" -c $query 2>&1
    if ($LASTEXITCODE -eq 0) {
        $result = $dockerResult
    } else {
        Write-Host "Error: Could not connect to database." -ForegroundColor Red
        Write-Host "Make sure the database is running and accessible." -ForegroundColor Yellow
        exit 1
    }
}

# Parse results
$positions = @()
$result | ForEach-Object {
    if ($_ -match '^(.+?)\|(.+?)\|(.+?)\|(.+?)\|(.+?)$') {
        $positions += [PSCustomObject]@{
            ID = $matches[1].Trim()
            Name = $matches[2].Trim()
            PositionCode = $matches[3].Trim()
            PositionType = $matches[4].Trim()
            DisplayOrder = [int]$matches[5].Trim()
        }
    }
}

# Display results
Write-Host "`nFound $($positions.Count) positions:`n" -ForegroundColor Green

# Group by position type
$branchPositions = $positions | Where-Object { $_.PositionType -eq 'branch' } | Sort-Object DisplayOrder, Name
$rotationPositions = $positions | Where-Object { $_.PositionType -eq 'rotation' } | Sort-Object DisplayOrder, Name

Write-Host "BRANCH-TYPE POSITIONS (for quota configuration):" -ForegroundColor Yellow
Write-Host "-" * 80 -ForegroundColor Gray
Write-Host ("{0,-20} {1,-50} {2,-10}" -f "Position Code", "Position Name", "Display Order") -ForegroundColor Cyan
Write-Host "-" * 80 -ForegroundColor Gray
foreach ($pos in $branchPositions) {
    $code = if ($pos.PositionCode) { $pos.PositionCode } else { "(not set)" }
    Write-Host ("{0,-20} {1,-50} {2,-10}" -f $code, $pos.Name, $pos.DisplayOrder)
}

Write-Host "`nROTATION-TYPE POSITIONS (not for quota configuration):" -ForegroundColor Yellow
Write-Host "-" * 80 -ForegroundColor Gray
Write-Host ("{0,-20} {1,-50} {2,-10}" -f "Position Code", "Position Name", "Display Order") -ForegroundColor Cyan
Write-Host "-" * 80 -ForegroundColor Gray
foreach ($pos in $rotationPositions) {
    $code = if ($pos.PositionCode) { $pos.PositionCode } else { "(not set)" }
    Write-Host ("{0,-20} {1,-50} {2,-10}" -f $code, $pos.Name, $pos.DisplayOrder)
}

# Generate CSV output
Write-Host "`n`nCSV Format (for easy copy-paste):" -ForegroundColor Green
Write-Host "=" * 80 -ForegroundColor Gray
Write-Host "Position Code,Position Name (Thai),Position Type,Display Order"
foreach ($pos in ($branchPositions + $rotationPositions)) {
    $code = if ($pos.PositionCode) { $pos.PositionCode } else { "" }
    Write-Host "$code,$($pos.Name),$($pos.PositionType),$($pos.DisplayOrder)"
}

# Generate markdown table
Write-Host "`n`nMarkdown Table Format:" -ForegroundColor Green
Write-Host "=" * 80 -ForegroundColor Gray
Write-Host "| Position Code | Position Name | Position Type | Display Order |"
Write-Host "|---------------|---------------|---------------|---------------|"
foreach ($pos in ($branchPositions + $rotationPositions)) {
    $code = if ($pos.PositionCode) { $pos.PositionCode } else { "" }
    Write-Host "| $code | $($pos.Name) | $($pos.PositionType) | $($pos.DisplayOrder) |"
}

# Count positions without codes
$noCodeCount = ($positions | Where-Object { -not $_.PositionCode }).Count
if ($noCodeCount -gt 0) {
    Write-Host "`n`n⚠️  Warning: $noCodeCount position(s) do not have position codes set." -ForegroundColor Yellow
    Write-Host "Please set position codes in the Position Management page (/positions)" -ForegroundColor Yellow
}

# Clean up
Remove-Item Env:\PGPASSWORD -ErrorAction SilentlyContinue

Write-Host "`nDone!" -ForegroundColor Green

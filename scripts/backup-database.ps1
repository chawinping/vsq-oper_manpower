# Database Backup Script
# Usage: .\scripts\backup-database.ps1 -Environment staging

param(
    [Parameter(Mandatory=$true)]
    [ValidateSet("staging", "production", "development")]
    [string]$Environment
)

$ErrorActionPreference = "Stop"

Write-Host "Creating database backup for $Environment environment..." -ForegroundColor Cyan

# Determine compose file
$composeFile = if ($Environment -eq "staging") {
    "docker-compose.staging.yml"
} elseif ($Environment -eq "production") {
    "docker-compose.production.yml"
} else {
    "docker-compose.yml"
}

# Create backup directory
$backupDir = "backups/$Environment"
$timestamp = Get-Date -Format "yyyyMMdd_HHmmss"
$backupFile = "$backupDir/backup_$timestamp.sql"

New-Item -ItemType Directory -Force -Path $backupDir | Out-Null

# Get database container name
$dbContainer = if ($Environment -eq "staging") {
    "vsq-manpower-db"
} elseif ($Environment -eq "production") {
    "vsq-manpower-db"
} else {
    "vsq-manpower-db"
}

# Load environment variables
$envFile = if ($Environment -eq "staging") {
    ".env.staging"
} elseif ($Environment -eq "production") {
    ".env.production"
} else {
    ".env"
}

if (Test-Path $envFile) {
    Get-Content $envFile | ForEach-Object {
        if ($_ -match '^\s*([^#][^=]+)=(.*)$') {
            $key = $matches[1].Trim()
            $value = $matches[2].Trim()
            if ($value -and $key) {
                [Environment]::SetEnvironmentVariable($key, $value, "Process")
            }
        }
    }
}

$dbUser = [Environment]::GetEnvironmentVariable("DB_USER", "Process") ?? "vsq_user"
$dbName = [Environment]::GetEnvironmentVariable("DB_NAME", "Process") ?? "vsq_manpower"

Write-Host "Backing up database: $dbName" -ForegroundColor Green

# Create backup
docker exec $dbContainer pg_dump -U $dbUser -d $dbName > $backupFile

if ($LASTEXITCODE -eq 0) {
    $fileSize = (Get-Item $backupFile).Length / 1MB
    Write-Host "✓ Backup created successfully: $backupFile ($([math]::Round($fileSize, 2)) MB)" -ForegroundColor Green
    
    # Compress backup
    Write-Host "Compressing backup..." -ForegroundColor Yellow
    Compress-Archive -Path $backupFile -DestinationPath "$backupFile.zip" -Force
    Remove-Item $backupFile
    Write-Host "✓ Compressed backup: $backupFile.zip" -ForegroundColor Green
} else {
    Write-Host "✗ Backup failed!" -ForegroundColor Red
    exit 1
}

# Keep only last 10 backups
Write-Host "Cleaning up old backups (keeping last 10)..." -ForegroundColor Yellow
Get-ChildItem -Path $backupDir -Filter "backup_*.zip" | 
    Sort-Object LastWriteTime -Descending | 
    Select-Object -Skip 10 | 
    Remove-Item -Force

Write-Host "Backup complete!" -ForegroundColor Green




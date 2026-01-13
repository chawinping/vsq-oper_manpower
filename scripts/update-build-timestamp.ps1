# VSQ Operations Manpower - Build Timestamp Update Script
# This script updates build timestamps without changing version numbers
# It's designed to be called automatically during the build process

param(
    [Parameter(Mandatory=$false)]
    [ValidateSet("frontend", "backend", "database", "all")]
    [string]$Component = "all"
)

# Get current date/time in Thailand timezone (UTC+7)
$utcNow = [DateTime]::UtcNow
$thailandTime = $utcNow.AddHours(7)
$buildDate = $thailandTime.ToString("yyyy-MM-dd HH:mm:ss")
$buildTime = $buildDate  # Use same simple format as buildDate

function Update-BuildTimestamp {
    param(
        [string]$FilePath
    )
    
    if (-not (Test-Path $FilePath)) {
        Write-Warning "Version file not found: $FilePath (skipping)"
        return $false
    }
    
    try {
        # Read current version file
        $versionContent = Get-Content $FilePath -Raw | ConvertFrom-Json
        
        # Update only timestamps, keep version number
        $versionContent.buildDate = $buildDate
        $versionContent.buildTime = $buildTime
        
        # Write back to file
        $jsonContent = $versionContent | ConvertTo-Json -Depth 10
        # Ensure proper JSON formatting (no trailing newline)
        $jsonContent = $jsonContent.Trim()
        [System.IO.File]::WriteAllText((Resolve-Path $FilePath), $jsonContent, [System.Text.Encoding]::UTF8)
        
        Write-Host "Updated build timestamp: $FilePath" -ForegroundColor Green
        Write-Host "  Version: $($versionContent.version)" -ForegroundColor Cyan
        Write-Host "  Build Date: $buildDate" -ForegroundColor Cyan
        
        return $true
    }
    catch {
        Write-Error "Failed to update $FilePath : $_"
        return $false
    }
}

# Get project root directory
$projectRoot = Split-Path -Parent $PSScriptRoot
Set-Location $projectRoot

Write-Host "`nUpdating build timestamps..." -ForegroundColor Yellow
Write-Host "Build Date: $buildDate" -ForegroundColor Yellow
Write-Host ""

$success = $true

switch ($Component) {
    "frontend" {
        Update-BuildTimestamp -FilePath "frontend\VERSION.json" | Out-Null
        Update-BuildTimestamp -FilePath "frontend\public\VERSION.json" | Out-Null
    }
    "backend" {
        Update-BuildTimestamp -FilePath "backend\VERSION.json" | Out-Null
    }
    "database" {
        Update-BuildTimestamp -FilePath "backend\DATABASE_VERSION.json" | Out-Null
    }
    "all" {
        Update-BuildTimestamp -FilePath "frontend\VERSION.json" | Out-Null
        Update-BuildTimestamp -FilePath "frontend\public\VERSION.json" | Out-Null
        Update-BuildTimestamp -FilePath "backend\VERSION.json" | Out-Null
        Update-BuildTimestamp -FilePath "backend\DATABASE_VERSION.json" | Out-Null
    }
}

Write-Host ""
Write-Host "Build timestamp update completed!" -ForegroundColor Green



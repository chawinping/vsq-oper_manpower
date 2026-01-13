# VSQ Operations Manpower - Version Update Script
# This script updates version numbers and timestamps for frontend, backend, and database components

param(
    [Parameter(Mandatory=$true)]
    [ValidateSet("frontend", "backend", "database", "all")]
    [string]$Component,
    
    [Parameter(Mandatory=$true)]
    [ValidateSet("major", "minor", "patch")]
    [string]$BumpType
)

# Get current date/time in Thailand timezone (UTC+7)
$thailandTime = [DateTime]::Now.AddHours(7)
$buildDate = $thailandTime.ToString("yyyy-MM-dd HH:mm:ss")
$buildTime = $buildDate  # Use same simple format as buildDate

function Update-VersionFile {
    param(
        [string]$FilePath,
        [string]$BumpType
    )
    
    if (-not (Test-Path $FilePath)) {
        Write-Error "Version file not found: $FilePath"
        return $false
    }
    
    try {
        # Read current version
        $versionContent = Get-Content $FilePath -Raw | ConvertFrom-Json
        $currentVersion = $versionContent.version
        
        # Parse version
        $versionParts = $currentVersion -split '\.'
        $major = [int]$versionParts[0]
        $minor = [int]$versionParts[1]
        $patch = [int]$versionParts[2]
        
        # Bump version
        switch ($BumpType) {
            "major" {
                $major++
                $minor = 0
                $patch = 0
            }
            "minor" {
                $minor++
                $patch = 0
            }
            "patch" {
                $patch++
            }
        }
        
        $newVersion = "$major.$minor.$patch"
        
        # Update version object
        $versionContent.version = $newVersion
        $versionContent.buildDate = $buildDate
        $versionContent.buildTime = $buildTime
        
        # Write back to file
        $versionContent | ConvertTo-Json | Set-Content $FilePath -NoNewline
        
        Write-Host "Updated $FilePath" -ForegroundColor Green
        Write-Host "  Version: $currentVersion -> $newVersion" -ForegroundColor Cyan
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

Write-Host "`nVSQ Operations Manpower - Version Update" -ForegroundColor Yellow
Write-Host "Component: $Component" -ForegroundColor Yellow
Write-Host "Bump Type: $BumpType" -ForegroundColor Yellow
Write-Host "Build Date: $buildDate" -ForegroundColor Yellow
Write-Host ""

$success = $true

switch ($Component) {
    "frontend" {
        Write-Host "Updating Frontend versions..." -ForegroundColor Cyan
        $success = Update-VersionFile -FilePath "frontend\VERSION.json" -BumpType $BumpType
        if ($success) {
            $success = Update-VersionFile -FilePath "frontend\public\VERSION.json" -BumpType $BumpType
        }
    }
    "backend" {
        Write-Host "Updating Backend version..." -ForegroundColor Cyan
        $success = Update-VersionFile -FilePath "backend\VERSION.json" -BumpType $BumpType
    }
    "database" {
        Write-Host "Updating Database version..." -ForegroundColor Cyan
        $success = Update-VersionFile -FilePath "backend\DATABASE_VERSION.json" -BumpType $BumpType
    }
    "all" {
        Write-Host "Updating all components..." -ForegroundColor Cyan
        Write-Host "`nFrontend:" -ForegroundColor Yellow
        $fe1 = Update-VersionFile -FilePath "frontend\VERSION.json" -BumpType $BumpType
        $fe2 = Update-VersionFile -FilePath "frontend\public\VERSION.json" -BumpType $BumpType
        
        Write-Host "`nBackend:" -ForegroundColor Yellow
        $be = Update-VersionFile -FilePath "backend\VERSION.json" -BumpType $BumpType
        
        Write-Host "`nDatabase:" -ForegroundColor Yellow
        $db = Update-VersionFile -FilePath "backend\DATABASE_VERSION.json" -BumpType $BumpType
        
        $success = $fe1 -and $fe2 -and $be -and $db
    }
}

Write-Host ""

if ($success) {
    Write-Host "Version update completed successfully!" -ForegroundColor Green
    exit 0
} else {
    Write-Host "Version update failed!" -ForegroundColor Red
    exit 1
}




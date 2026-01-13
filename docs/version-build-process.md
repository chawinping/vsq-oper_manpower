# Version Build Process

## Overview

This document explains how version information is automatically updated during the build process.

## Problem

Previously, `VERSION.json` files contained static timestamps that didn't update when you rebuilt the application, even though you were building new versions.

## Solution

The build process now automatically updates build timestamps (buildDate and buildTime) whenever you build the application, without requiring manual intervention.

## How It Works

### Frontend

1. **Prebuild Script**: The `prebuild` script in `package.json` automatically runs before `npm run build`
   - Updates `frontend/VERSION.json`
   - Updates `frontend/public/VERSION.json`
   - Sets buildDate and buildTime to current Thailand timezone (UTC+7)

2. **Docker Build**: The Dockerfile also includes a timestamp update step as a safety measure

### Backend

1. **Docker Build**: The Dockerfile updates timestamps before building:
   - Updates `backend/VERSION.json`
   - Updates `backend/DATABASE_VERSION.json`
   - Uses `jq` or `python3` (whichever is available) to update JSON files

## Manual Version Updates

If you need to bump version numbers (not just timestamps), use the existing script:

### PowerShell (Windows)
```powershell
.\scripts\update-version.ps1 -Component all -BumpType patch
```

### Bash (Linux/Mac)
```bash
# First, manually update version numbers if needed
# Then build timestamps will be updated automatically during build
```

## Build Timestamps Only

If you just want to update timestamps without changing version numbers:

### PowerShell (Windows)
```powershell
.\scripts\update-build-timestamp.ps1 -Component all
```

### Bash (Linux/Mac)
```bash
./scripts/update-build-timestamp.sh all
```

## What Gets Updated

- **Version Number**: Only updated manually using `update-version.ps1`
- **Build Date**: Automatically updated on every build (format: `YYYY-MM-DD HH:mm:ss`)
- **Build Time**: Automatically updated on every build (format: `YYYY-MM-DDTHH:mm:ss+07:00`)

## Timezone

All timestamps use **Thailand timezone (UTC+7)** to match the application's timezone settings.

## Files Updated

1. `frontend/VERSION.json`
2. `frontend/public/VERSION.json` (served to browser)
3. `backend/VERSION.json`
4. `backend/DATABASE_VERSION.json`

## Verification

After building, check the login page - it should display the current build timestamps for all components.




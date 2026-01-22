---
title: Docker Desktop Image Display Issue
description: Troubleshooting why postgres:18-alpine might not show in Docker Desktop
version: 1.0.0
lastUpdated: 2025-01-08
---

# Docker Desktop Image Display Issue

## Issue

PostgreSQL 18-alpine image exists and is working, but may not be visible in Docker Desktop UI.

## Verification

The image **definitely exists** and is **working correctly**:

```powershell
# Check images via command line
docker images postgres

# Output should show:
# REPOSITORY   TAG         IMAGE ID       CREATED
# postgres     18-alpine   154ea39af68f   ...
# postgres     15-alpine   b3968e348b48   ...
```

## Why Docker Desktop Might Not Show It

### 1. Docker Desktop UI Cache

**Solution:** Refresh Docker Desktop
- Close and reopen Docker Desktop
- Or click the refresh button in the Images tab

### 2. Filter/Search Settings

**Check:**
- Make sure no filters are applied in Docker Desktop
- Search for "postgres" (not just "postgres 18")
- Check "Show unused images" if that option exists

### 3. Image Tag Display

Docker Desktop might show images differently:
- Look for: `postgres` with tag `18-alpine`
- Or: `library/postgres:18-alpine`
- Check both "Images" and "Containers" tabs

### 4. Docker Desktop Version

Older versions might have display issues. Update Docker Desktop if needed.

## Verification Commands

Run these to confirm the image exists:

```powershell
# List all postgres images
docker images postgres

# Check what image the container is using
docker inspect vsq-manpower-db --format "{{.Config.Image}}"

# Verify PostgreSQL version in container
docker exec vsq-manpower-db psql --version
# Should show: psql (PostgreSQL) 18.1
```

## Current Status

✅ **Image exists:** `postgres:18-alpine` (ID: 154ea39af68f)  
✅ **Container using it:** `vsq-manpower-db`  
✅ **PostgreSQL version:** 18.1  
✅ **Working correctly:** Database is operational

## What to Do

### Option 1: Refresh Docker Desktop

1. Close Docker Desktop completely
2. Reopen Docker Desktop
3. Go to Images tab
4. Search for "postgres"

### Option 2: Check via Command Line

The image is definitely there and working. Docker Desktop UI might just need a refresh:

```powershell
# This confirms the image exists
docker images postgres:18-alpine

# This confirms the container is using it
docker inspect vsq-manpower-db --format "{{.Config.Image}}"
```

### Option 3: Force Pull (if needed)

If you want to ensure the image is fully downloaded:

```powershell
docker pull postgres:18-alpine
```

## Important Note

**The image IS there and IS working.** Your container is running PostgreSQL 18.1 successfully. This is just a Docker Desktop UI display issue, not a functional problem.

If Docker Desktop still doesn't show it after refreshing, you can safely ignore it - everything is working correctly via command line and your application.

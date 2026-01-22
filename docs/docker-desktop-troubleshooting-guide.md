---
title: Docker Desktop Troubleshooting Guide
description: Step-by-step guide to troubleshoot Docker Desktop image display issues
version: 1.0.0
lastUpdated: 2025-01-08
---

# Docker Desktop Troubleshooting Guide

## Issue: PostgreSQL 18-alpine Image Not Visible in Docker Desktop

### Current Status

✅ **Image EXISTS** - Verified via command line  
✅ **Container WORKING** - Running PostgreSQL 18.1  
⚠️ **Docker Desktop UI** - May not be displaying correctly

## Step-by-Step Troubleshooting

### Step 1: Verify Image Exists (Command Line)

```powershell
# Check if image exists
docker images postgres:18-alpine

# Expected output:
# REPOSITORY   TAG         IMAGE ID       CREATED        SIZE
# postgres     18-alpine   b40d931bd0e7  ...           402MB
```

**If this shows the image:** ✅ Image exists, issue is with Docker Desktop UI  
**If this doesn't show:** Image needs to be pulled (see Step 2)

### Step 2: Pull Image (If Missing)

```powershell
# Pull the latest PostgreSQL 18-alpine image
docker pull postgres:18-alpine

# Verify it was pulled
docker images postgres:18-alpine
```

### Step 3: Refresh Docker Desktop

**Method 1: Restart Docker Desktop**
1. Right-click Docker Desktop icon in system tray
2. Click "Quit Docker Desktop"
3. Wait 10 seconds
4. Open Docker Desktop again
5. Go to Images tab
6. Search for "postgres"

**Method 2: Refresh Images Tab**
1. Open Docker Desktop
2. Go to Images tab
3. Press `F5` or click refresh button (if available)
4. Search for "postgres"

**Method 3: Clear Docker Desktop Cache**
1. Close Docker Desktop
2. Delete cache (optional, advanced):
   ```powershell
   # Windows cache location (usually)
   Remove-Item "$env:LOCALAPPDATA\Docker\*.cache" -ErrorAction SilentlyContinue
   ```
3. Restart Docker Desktop

### Step 4: Check Docker Desktop Settings

**Image Display Settings:**
1. Open Docker Desktop
2. Go to Settings (gear icon)
3. Check "General" settings:
   - Ensure "Show system containers" is enabled (if you want to see all)
   - Check "Display options" for image filtering

**View Options:**
1. In Images tab, check:
   - No filters applied (top right)
   - Search box is empty or search for "postgres"
   - View mode is set to show all images

### Step 5: Check Image Tags

Docker Desktop might show images differently:

**Look for:**
- `postgres` (repository) with tag `18-alpine`
- `library/postgres:18-alpine` (full name)
- `postgres:18-alpine` (short name)

**In Docker Desktop:**
1. Go to Images tab
2. Look for repository: `postgres`
3. Click to expand and see tags
4. You should see: `15-alpine` and `18-alpine`

### Step 6: Verify Container is Using Correct Image

```powershell
# Check what image the container is using
docker inspect vsq-manpower-db --format "{{.Config.Image}}"

# Should show: postgres:18-alpine

# Verify PostgreSQL version
docker exec vsq-manpower-db psql --version

# Should show: psql (PostgreSQL) 18.1
```

### Step 7: Force Docker Desktop to Recognize Image

**Option A: Restart Container**
```powershell
docker-compose --profile dev restart postgres
```

**Option B: Recreate Container**
```powershell
docker-compose --profile dev up -d --force-recreate postgres
```

**Option C: Tag Image Explicitly**
```powershell
# This ensures Docker Desktop recognizes it
docker tag postgres:18-alpine postgres:18-alpine
```

## Common Docker Desktop Issues

### Issue 1: Images Tab Not Updating

**Symptoms:** New images don't appear in Docker Desktop

**Solutions:**
1. Restart Docker Desktop
2. Check if image exists via command line first
3. Try pulling image again: `docker pull postgres:18-alpine`
4. Check Docker Desktop logs (Help > Troubleshoot)

### Issue 2: Filter Applied

**Symptoms:** Only some images visible

**Solutions:**
1. Clear search box in Images tab
2. Remove any filters (check filter dropdown)
3. Check "Show unused images" option

### Issue 3: Docker Desktop Version

**Symptoms:** UI behaves unexpectedly

**Solutions:**
1. Check Docker Desktop version: Help > About Docker Desktop
2. Update to latest version if outdated
3. Check release notes for known issues

### Issue 4: Image Name Display

**Symptoms:** Image exists but name looks different

**Solutions:**
- Docker Desktop might show: `postgres` (click to see tags)
- Or: `library/postgres:18-alpine`
- Search for just "postgres" and expand to see all tags

## Diagnostic Commands

Run these to gather information:

```powershell
# 1. List all postgres images
docker images postgres

# 2. Check image details
docker image inspect postgres:18-alpine --format "{{.RepoTags}} | {{.Id}} | {{.Size}}"

# 3. Check container image
docker inspect vsq-manpower-db --format "Image: {{.Config.Image}} | ImageID: {{.Image}}"

# 4. List all images (to see if it's there)
docker images --format "{{.Repository}}:{{.Tag}}"

# 5. Check Docker system info
docker system df
```

## Expected Results

After troubleshooting, you should see:

**In Command Line:**
```
REPOSITORY   TAG         IMAGE ID       SIZE
postgres     18-alpine   b40d931bd0e7   402MB
```

**In Docker Desktop:**
- Images tab shows `postgres` repository
- Expanding shows tags: `15-alpine` and `18-alpine`
- `18-alpine` tag is present

**Container Status:**
- `vsq-manpower-db` container running
- Using image: `postgres:18-alpine`
- PostgreSQL version: 18.1

## If Still Not Visible

### Workaround: Use Command Line

Even if Docker Desktop doesn't show it, you can manage everything via command line:

```powershell
# View images
docker images

# View containers
docker ps

# View volumes
docker volume ls

# Manage containers
docker-compose --profile dev ps
docker-compose --profile dev logs
```

### Report Issue to Docker

If image definitely exists but Docker Desktop doesn't show it:

1. **Gather Information:**
   ```powershell
   docker version
   docker info
   docker images postgres:18-alpine
   ```

2. **Check Docker Desktop Version:**
   - Help > About Docker Desktop

3. **Report to Docker:**
   - Include Docker Desktop version
   - Include output of diagnostic commands
   - Describe what you see vs. what you expect

## Quick Fix Script

Run this to ensure everything is correct:

```powershell
# Verify image exists
Write-Host "Checking postgres:18-alpine image..." -ForegroundColor Yellow
docker images postgres:18-alpine

# Verify container is using it
Write-Host "`nChecking container image..." -ForegroundColor Yellow
docker inspect vsq-manpower-db --format "Container Image: {{.Config.Image}}"

# Verify PostgreSQL version
Write-Host "`nChecking PostgreSQL version..." -ForegroundColor Yellow
docker exec vsq-manpower-db psql --version

# Refresh Docker Desktop (suggestion)
Write-Host "`nTo refresh Docker Desktop:" -ForegroundColor Cyan
Write-Host "1. Close Docker Desktop" -ForegroundColor Gray
Write-Host "2. Reopen Docker Desktop" -ForegroundColor Gray
Write-Host "3. Go to Images tab" -ForegroundColor Gray
Write-Host "4. Search for 'postgres'" -ForegroundColor Gray
```

## Summary

**The image IS there and IS working.** Your container is successfully running PostgreSQL 18.1. If Docker Desktop doesn't show it:

1. ✅ **Try refreshing Docker Desktop** (restart it)
2. ✅ **Check image tags** (expand `postgres` repository)
3. ✅ **Clear filters** in Images tab
4. ✅ **Use command line** if UI still doesn't show it

The important thing: **Your application is working correctly** with PostgreSQL 18, regardless of what Docker Desktop shows.

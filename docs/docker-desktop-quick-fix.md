---
title: Docker Desktop Quick Fix - PostgreSQL 18 Image
description: Quick steps to make postgres:18-alpine visible in Docker Desktop
version: 1.0.0
lastUpdated: 2025-01-08
---

# Docker Desktop Quick Fix

## The Image EXISTS and is WORKING ✅

Your `postgres:18-alpine` image is present and your container is running PostgreSQL 18.1 successfully.

## Quick Fix Steps

### Method 1: Restart Docker Desktop (Most Common Fix)

1. **Quit Docker Desktop:**
   - Right-click Docker Desktop icon in system tray (bottom right)
   - Click "Quit Docker Desktop"
   - Wait 10-15 seconds

2. **Reopen Docker Desktop:**
   - Open Docker Desktop from Start menu
   - Wait for it to fully load

3. **Check Images Tab:**
   - Click "Images" in left sidebar
   - Look for repository: `postgres`
   - Click on `postgres` to expand
   - You should see tags: `15-alpine` and `18-alpine`

### Method 2: Refresh Images Tab

1. Open Docker Desktop
2. Go to **Images** tab
3. Press `F5` or click refresh (if available)
4. Clear search box (if anything is typed)
5. Look for `postgres` repository

### Method 3: Search for Image

1. In **Images** tab, use the search box
2. Type: `postgres`
3. You should see both `postgres:15-alpine` and `postgres:18-alpine`
4. Or search: `18-alpine` directly

### Method 4: Check Image Tags

Docker Desktop might group images by repository:

1. Find repository: `postgres` (or `library/postgres`)
2. Click to expand/collapse
3. You'll see tags: `15-alpine` and `18-alpine`
4. The `18-alpine` tag is your PostgreSQL 18 image

## Verification

Even if Docker Desktop doesn't show it, verify it exists:

```powershell
# Run this to confirm
docker images postgres:18-alpine

# Expected output:
# REPOSITORY   TAG         IMAGE ID       CREATED        SIZE
# postgres     18-alpine   b40d931bd0e7   ...           402MB
```

## Why This Happens

Docker Desktop UI sometimes:
- Needs a refresh after pulling new images
- Groups images by repository (need to expand to see tags)
- Has caching issues
- Shows images differently than command line

## Important Note

**Your application IS working correctly!**

- ✅ Container is running PostgreSQL 18.1
- ✅ Image exists and is being used
- ✅ Database is operational
- ✅ All your data is intact

The Docker Desktop UI display is just a cosmetic issue - everything is functioning properly.

## If Still Not Visible

**Option 1: Use Command Line**
- You can manage everything via `docker` commands
- Everything works the same way

**Option 2: Check Docker Desktop Version**
- Help > About Docker Desktop
- Update if outdated

**Option 3: Report to Docker**
- If image definitely exists but UI doesn't show it
- Include Docker Desktop version and diagnostic output

## Quick Verification Script

Run this to verify everything:

```powershell
.\scripts\fix-docker-desktop-display.ps1
```

This will show you:
- ✅ Image exists
- ✅ Container is using it
- ✅ PostgreSQL version is 18.1
- ✅ Troubleshooting steps

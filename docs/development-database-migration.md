---
title: Development Database Migration Guide
description: Guide for migrating development environment to PostgreSQL 18
version: 1.0.0
lastUpdated: 2025-01-08
---

# Development Database Migration Guide

## Overview

This guide covers migrating your **development environment** from PostgreSQL 15 to PostgreSQL 18. This should be done **before** staging and production migrations.

**Why migrate development first?**
- ✅ Test migration process in safe environment
- ✅ Verify all code works with PostgreSQL 18
- ✅ No impact on other environments
- ✅ Can test and fix issues without pressure

## Current Development Setup

Your development environment uses:
- **Backend:** `vsq-manpower-backend-dev` (running)
- **Frontend:** `vsq-manpower-frontend-dev` (running)
- **Database:** PostgreSQL 15 (configured in `docker-compose.yml`)

## Migration Steps for Development

### Step 1: Verify Current Development Database

```powershell
# Check if development database container exists
docker ps -a | Select-String "vsq-manpower-db"

# If container exists, verify its state
.\scripts\verify-database-state.ps1 -ContainerName "vsq-manpower-db" -DatabaseName "vsq_manpower"
```

**If you're using a local PostgreSQL (not Docker):**
```powershell
# Check local PostgreSQL version
psql --version

# Or connect and check
psql -U vsq_user -d vsq_manpower -c "SELECT version();"
```

### Step 2: Backup Development Database

**Option A: Docker Container**
```powershell
# Start database if not running
docker-compose --profile dev up -d postgres

# Create backup
docker exec vsq-manpower-db pg_dump -U vsq_user -d vsq_manpower -F c -f /tmp/dev_backup.dump
docker cp vsq-manpower-db:/tmp/dev_backup.dump ./dev_backup_$(Get-Date -Format "yyyyMMdd_HHmmss").dump
```

**Option B: Local PostgreSQL**
```powershell
# Create backup
pg_dump -U vsq_user -d vsq_manpower -F c -f dev_backup_$(Get-Date -Format "yyyyMMdd_HHmmss").dump
```

### Step 3: Update docker-compose.yml for Development

Update the PostgreSQL image in `docker-compose.yml`:

```yaml
services:
  postgres:
    profiles: ["fullstack", "dev", "fullstack-dev", "backend-only", "hybrid-dev"]
    image: postgres:18-alpine  # Changed from postgres:15-alpine
    container_name: vsq-manpower-db
    # ... rest of configuration stays the same
```

### Step 4: Migrate Development Database

**Option A: Fresh Start (Recommended for Development)**

Since development data is less critical, you can start fresh:

```powershell
# Stop all services
docker-compose --profile dev down

# Remove old database volume (WARNING: This deletes all data!)
docker volume rm vsq-oper_manpower_postgres_data

# Start PostgreSQL 18
docker-compose --profile dev up -d postgres

# Wait for database to be ready
docker-compose --profile dev ps postgres

# Start backend - migrations will run automatically
docker-compose --profile dev up -d backend-dev

# Verify migration
.\scripts\verify-database-state.ps1 -ContainerName "vsq-manpower-db" -DatabaseName "vsq_manpower" -Detailed
```

**Option B: Preserve Development Data**

If you want to keep your development data:

```powershell
# Stop services
docker-compose --profile dev down

# Export data from PostgreSQL 15
docker-compose --profile dev up -d postgres
docker exec vsq-manpower-db pg_dump -U vsq_user -d vsq_manpower -F c -f /tmp/dev_data.dump
docker cp vsq-manpower-db:/tmp/dev_data.dump ./dev_data.dump

# Stop and remove old container
docker-compose --profile dev down
docker volume rm vsq-oper_manpower_postgres_data  # Optional: remove old data

# Update docker-compose.yml (change to postgres:18-alpine)

# Start PostgreSQL 18
docker-compose --profile dev up -d postgres

# Wait for database
Start-Sleep -Seconds 10

# Let migrations create schema
docker-compose --profile dev up -d backend-dev

# Wait for migrations
Start-Sleep -Seconds 10

# Import data
docker cp ./dev_data.dump vsq-manpower-db:/tmp/dev_data.dump
docker exec vsq-manpower-db pg_restore -U vsq_user -d vsq_manpower --data-only /tmp/dev_data.dump
```

### Step 5: Verify Development Migration

```powershell
# Run verification script
.\scripts\verify-database-state.ps1 -ContainerName "vsq-manpower-db" -DatabaseName "vsq_manpower" -Detailed

# Check backend logs
docker-compose --profile dev logs backend-dev --tail=50

# Test application
# - Open http://localhost:4000
# - Login
# - Test key features
# - Verify data is accessible
```

### Step 6: Test Development Workflow

After migration, test your development workflow:

- [ ] Backend starts without errors
- [ ] Frontend connects to backend
- [ ] Login works
- [ ] Can create/read/update/delete data
- [ ] Migrations run correctly
- [ ] Hot reload works
- [ ] All features work as expected

## Development-Specific Considerations

### Hot Reload

After migration, verify hot reload still works:

```powershell
# Make a small code change
# Backend should auto-reload
docker-compose --profile dev logs backend-dev -f
```

### Test Data

If you lost test data, you can:

1. **Re-seed manually** through the application
2. **Use test fixtures** if available
3. **Import from backup** (if you preserved data)

### Local Development (Non-Docker)

If you're running PostgreSQL locally (not in Docker):

1. **Install PostgreSQL 18** locally
2. **Update connection string** in your `.env` or config
3. **Run migrations** manually or via backend startup
4. **Import data** if needed

## Troubleshooting Development Migration

### Issue: Backend Can't Connect

**Symptoms:** Backend logs show connection errors

**Solution:**
```powershell
# Check database is running
docker-compose --profile dev ps postgres

# Check connection from backend
docker exec vsq-manpower-backend-dev ping postgres

# Verify database name matches
docker exec vsq-manpower-db psql -U vsq_user -l
```

### Issue: Migrations Fail

**Symptoms:** Migration errors in backend logs

**Solution:**
```powershell
# Check migration logs
docker-compose --profile dev logs backend-dev | Select-String "migration"

# Verify database is empty/clean
docker exec vsq-manpower-db psql -U vsq_user -d vsq_manpower -c "\dt"

# Try running migrations manually
docker exec vsq-manpower-backend-dev go run cmd/migrate/main.go
```

### Issue: Data Missing After Migration

**Symptoms:** Tables exist but no data

**Solution:**
```powershell
# Check if you have a backup
ls dev_backup_*.dump

# Restore data
docker cp ./dev_backup_YYYYMMDD_HHMMSS.dump vsq-manpower-db:/tmp/backup.dump
docker exec vsq-manpower-db pg_restore -U vsq_user -d vsq_manpower --data-only /tmp/backup.dump
```

## Quick Migration Script for Development

Here's a quick script to migrate development:

```powershell
# Development Migration Script
Write-Host "Development Database Migration to PostgreSQL 18" -ForegroundColor Cyan

# 1. Backup (optional but recommended)
Write-Host "`n[1/5] Creating backup..." -ForegroundColor Yellow
docker-compose --profile dev up -d postgres
docker exec vsq-manpower-db pg_dump -U vsq_user -d vsq_manpower -F c -f /tmp/dev_backup.dump
docker cp vsq-manpower-db:/tmp/dev_backup.dump ./dev_backup_$(Get-Date -Format "yyyyMMdd_HHmmss").dump
Write-Host "[OK] Backup created" -ForegroundColor Green

# 2. Stop services
Write-Host "`n[2/5] Stopping services..." -ForegroundColor Yellow
docker-compose --profile dev down
Write-Host "[OK] Services stopped" -ForegroundColor Green

# 3. Update docker-compose.yml (manual step)
Write-Host "`n[3/5] Update docker-compose.yml:" -ForegroundColor Yellow
Write-Host "  Change: image: postgres:15-alpine" -ForegroundColor Gray
Write-Host "  To:     image: postgres:18-alpine" -ForegroundColor Gray
Read-Host "Press Enter after updating docker-compose.yml"

# 4. Start PostgreSQL 18
Write-Host "`n[4/5] Starting PostgreSQL 18..." -ForegroundColor Yellow
docker-compose --profile dev up -d postgres
Start-Sleep -Seconds 10
Write-Host "[OK] PostgreSQL 18 started" -ForegroundColor Green

# 5. Start backend (runs migrations)
Write-Host "`n[5/5] Starting backend (migrations will run)..." -ForegroundColor Yellow
docker-compose --profile dev up -d backend-dev
Start-Sleep -Seconds 10
Write-Host "[OK] Backend started" -ForegroundColor Green

# Verify
Write-Host "`nVerifying migration..." -ForegroundColor Cyan
.\scripts\verify-database-state.ps1 -ContainerName "vsq-manpower-db" -DatabaseName "vsq_manpower" -Detailed

Write-Host "`n[OK] Development migration complete!" -ForegroundColor Green
```

## After Development Migration

Once development is migrated:

1. ✅ **Test thoroughly** - Use development environment for all testing
2. ✅ **Document issues** - Note any problems or gotchas
3. ✅ **Update team** - Let team know development is on PostgreSQL 18
4. ✅ **Plan staging** - Use lessons learned for staging migration
5. ✅ **Update docs** - Document any process changes

## Next Steps

After successful development migration:

1. **Staging Migration** - Follow `docs/postgresql-18-migration-plan.md`
2. **Production Migration** - After staging is verified

## Related Documentation

- [Migration Plan](./postgresql-18-migration-plan.md) - Overall migration strategy
- [Migration Guide](./postgresql-18-migration-guide.md) - Detailed technical guide
- [Compatibility Report](./postgresql-18-compatibility-report.md) - Compatibility analysis

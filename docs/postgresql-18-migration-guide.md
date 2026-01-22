---
title: PostgreSQL 18 Migration Guide
description: Step-by-step guide for migrating from PostgreSQL 15 to PostgreSQL 18
version: 1.0.0
lastUpdated: 2025-01-08
---

# PostgreSQL 18 Migration Guide

## Overview

This guide covers migrating the VSQ Operations Manpower application from PostgreSQL 15 to PostgreSQL 18.

**Current Status:**
- ✅ Compatibility tested and verified
- ✅ Migration code fixes applied
- ✅ Ready for staging deployment

## Pre-Migration Checklist

### 1. Backup Current Database

**CRITICAL:** Always backup before migration!

```powershell
# Using the backup script
.\scripts\backup-database.ps1

# Or manually
docker exec vsq-manpower-db pg_dump -U vsq_user vsq_manpower > backup_$(Get-Date -Format "yyyyMMdd_HHmmss").sql
```

### 2. Verify Current Database State

```powershell
# Check current PostgreSQL version
docker exec vsq-manpower-db psql -U vsq_user -d vsq_manpower -c "SELECT version();"

# Verify all tables exist
docker exec vsq-manpower-db psql -U vsq_user -d vsq_manpower -c "\dt"

# Check foreign key constraints
docker exec vsq-manpower-db psql -U vsq_user -d vsq_manpower -c "
SELECT conname, conrelid::regclass, confrelid::regclass 
FROM pg_constraint 
WHERE contype = 'f' AND conrelid::regclass::text = 'branches';
"
```

### 3. Review Migration Changes

The following changes were made to support PostgreSQL 18:

1. **Migration Order Fix:**
   - `branches` table now created before `users` table
   - Resolves circular dependency issue

2. **Foreign Key Constraint:**
   - Foreign key from `branches.area_manager_id` to `users.id` is now added after both tables exist
   - Uses `IF NOT EXISTS` check for safety

## Migration Methods

### Method 1: Using pg_upgrade (Recommended for Production)

**Advantages:**
- Fastest method
- Minimal downtime
- Preserves all data and settings

**Steps:**

1. **Stop the application:**
```powershell
docker-compose down
```

2. **Backup the database:**
```powershell
.\scripts\backup-database.ps1
```

3. **Create new PostgreSQL 18 container:**
```powershell
# Update docker-compose.yml to use postgres:18-alpine
# Or use the test configuration as reference
```

4. **Use pg_upgrade:**
```powershell
# This requires both PostgreSQL 15 and 18 binaries
# For Docker, you may need to use a migration container
docker run --rm \
  -v vsq-oper_manpower_postgres_data:/var/lib/postgresql/15/data \
  -v vsq-oper_manpower_postgres_data_pg18:/var/lib/postgresql/18/data \
  postgres:18-alpine \
  pg_upgrade \
  --old-datadir=/var/lib/postgresql/15/data \
  --new-datadir=/var/lib/postgresql/18/data \
  --old-bindir=/usr/lib/postgresql/15/bin \
  --new-bindir=/usr/lib/postgresql/18/bin
```

**Note:** pg_upgrade in Docker can be complex. Consider Method 2 for simplicity.

### Method 2: Dump and Restore (Recommended for Staging/Development)

**Advantages:**
- Simple and reliable
- Works well with Docker
- Good for testing

**Steps:**

1. **Export data from PostgreSQL 15:**
```powershell
# Export schema and data
docker exec vsq-manpower-db pg_dump -U vsq_user -d vsq_manpower -F c -f /tmp/backup.dump

# Copy backup out of container
docker cp vsq-manpower-db:/tmp/backup.dump ./backup_$(Get-Date -Format "yyyyMMdd_HHmmss").dump
```

2. **Stop old database:**
```powershell
docker-compose down
```

3. **Update docker-compose.yml:**
```yaml
# Change postgres service image
postgres:
  image: postgres:18-alpine  # Changed from postgres:15-alpine
```

4. **Start new PostgreSQL 18 database:**
```powershell
docker-compose up -d postgres
```

5. **Wait for database to be ready:**
```powershell
# Wait for health check
docker-compose ps postgres
```

6. **Restore data:**
```powershell
# Copy backup into new container
docker cp ./backup_YYYYMMDD_HHMMSS.dump vsq-manpower-db:/tmp/backup.dump

# Restore
docker exec vsq-manpower-db pg_restore -U vsq_user -d vsq_manpower -c /tmp/backup.dump
```

7. **Run migrations (to apply any new changes):**
```powershell
# Start backend - migrations will run automatically
docker-compose up -d backend
```

### Method 3: Fresh Install with Data Migration (For Development)

**Steps:**

1. **Export data only (no schema):**
```powershell
docker exec vsq-manpower-db pg_dump -U vsq_user -d vsq_manpower --data-only -F c -f /tmp/data.dump
docker cp vsq-manpower-db:/tmp/data.dump ./data.dump
```

2. **Start fresh PostgreSQL 18:**
```powershell
docker-compose down -v  # Remove old volumes
# Update docker-compose.yml to use postgres:18-alpine
docker-compose up -d postgres
```

3. **Let migrations create schema:**
```powershell
docker-compose up -d backend  # Migrations run automatically
```

4. **Import data:**
```powershell
docker cp ./data.dump vsq-manpower-db:/tmp/data.dump
docker exec vsq-manpower-db pg_restore -U vsq_user -d vsq_manpower --data-only /tmp/data.dump
```

## Post-Migration Verification

### 1. Verify Database Version

```powershell
docker exec vsq-manpower-db psql -U vsq_user -d vsq_manpower -c "SELECT version();"
# Should show: PostgreSQL 18.x
```

### 2. Verify All Tables Exist

```powershell
docker exec vsq-manpower-db psql -U vsq_user -d vsq_manpower -c "
SELECT COUNT(*) as table_count 
FROM information_schema.tables 
WHERE table_schema = 'public' AND table_type = 'BASE TABLE';
"
# Should return: 32
```

### 3. Verify Custom Functions

```powershell
docker exec vsq-manpower-db psql -U vsq_user -d vsq_manpower -c "
SELECT proname 
FROM pg_proc p 
JOIN pg_namespace n ON p.pronamespace = n.oid 
WHERE n.nspname = 'public' 
AND proname IN ('get_revenue_level_tier', 'scenario_matches');
"
# Should return both functions
```

### 4. Verify Foreign Key Constraints

```powershell
docker exec vsq-manpower-db psql -U vsq_user -d vsq_manpower -c "
SELECT conname, conrelid::regclass, confrelid::regclass 
FROM pg_constraint 
WHERE contype = 'f' 
AND conrelid::regclass::text = 'branches'
AND conname = 'branches_area_manager_id_fkey';
"
# Should return the constraint
```

### 5. Test Application

```powershell
# Check health endpoint
curl http://localhost:8081/health

# Test API endpoints
# Login, create data, query data, etc.
```

### 6. Check Application Logs

```powershell
docker-compose logs backend --tail=100
# Look for any errors or warnings
```

## Rollback Plan

If migration fails, rollback steps:

1. **Stop new database:**
```powershell
docker-compose down
```

2. **Restore old docker-compose.yml:**
```yaml
postgres:
  image: postgres:15-alpine  # Revert to old version
```

3. **Restore from backup:**
```powershell
# If using volumes, restore the volume
# Or restore from SQL dump
docker exec vsq-manpower-db psql -U vsq_user -d vsq_manpower < backup.sql
```

4. **Start old database:**
```powershell
docker-compose up -d
```

## Migration for Existing Databases

### For Existing PostgreSQL 15 Databases

The migration code changes are **backward compatible**. Existing databases will:

1. ✅ Continue to work with existing schema
2. ✅ The foreign key constraint check uses `IF NOT EXISTS`, so it won't fail if already exists
3. ✅ No data migration needed

**However**, if you want to ensure the foreign key constraint exists (for consistency), you can run:

```sql
-- This is safe to run multiple times
DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'branches_area_manager_id_fkey' 
        AND table_name = 'branches'
    ) THEN
        ALTER TABLE branches 
        ADD CONSTRAINT branches_area_manager_id_fkey 
        FOREIGN KEY (area_manager_id) REFERENCES users(id);
    END IF;
END $$;
```

## Environment-Specific Migration

### Development Environment

**Recommended:** Method 3 (Fresh Install)
- Fastest for development
- Tests migration code
- No data loss concerns

### Staging Environment

**Recommended:** Method 2 (Dump and Restore)
- Tests real migration process
- Preserves test data
- Good practice run for production

### Production Environment

**Recommended:** Method 1 (pg_upgrade) or Method 2 (Dump and Restore)
- pg_upgrade: Faster, less downtime
- Dump/Restore: More reliable, easier to verify

## Troubleshooting

### Issue: Foreign Key Constraint Already Exists

**Error:** `constraint "branches_area_manager_id_fkey" already exists`

**Solution:** The migration code handles this automatically with `IF NOT EXISTS` check. If you see this error, it means the constraint was added manually. The migration will skip it safely.

### Issue: Tables Already Exist

**Error:** `relation "branches" already exists`

**Solution:** This is normal for existing databases. The `CREATE TABLE IF NOT EXISTS` statements will skip existing tables.

### Issue: Migration Fails on Circular Dependency

**Error:** `relation "branches" does not exist` or `relation "users" does not exist`

**Solution:** This was fixed in the migration code. Ensure you're using the latest `migrations.go` file with the correct table creation order.

## Timeline Recommendations

### Development
- **When:** Anytime
- **Method:** Fresh install (Method 3)
- **Risk:** Low

### Staging
- **When:** Before production migration
- **Method:** Dump and restore (Method 2)
- **Risk:** Low
- **Duration:** ~30 minutes

### Production
- **When:** During maintenance window
- **Method:** pg_upgrade or dump/restore
- **Risk:** Medium (mitigated by backup)
- **Duration:** 1-2 hours (depending on database size)
- **Downtime:** 30 minutes - 1 hour

## Additional Resources

- [PostgreSQL 18 Release Notes](https://www.postgresql.org/docs/18/release-18.html)
- [PostgreSQL Upgrade Documentation](https://www.postgresql.org/docs/current/pgupgrade.html)
- [Compatibility Report](./postgresql-18-compatibility-report.md)
- [Test Results](./postgresql-18-test-results.md)

## Support

If you encounter issues during migration:

1. Check the troubleshooting section above
2. Review application logs: `docker-compose logs backend`
3. Review database logs: `docker-compose logs postgres`
4. Verify backup is available before proceeding
5. Consider testing in staging environment first

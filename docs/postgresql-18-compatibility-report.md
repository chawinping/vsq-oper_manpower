---
title: PostgreSQL 18 Compatibility Report
description: Compatibility analysis for migrating from PostgreSQL 15 to PostgreSQL 18
version: 1.0.0
lastUpdated: 2025-01-08
---

# PostgreSQL 18 Compatibility Report

## Executive Summary

**Current Version:** PostgreSQL 15 (`postgres:15-alpine`)  
**Target Version:** PostgreSQL 18  
**Status:** ✅ **LOW RISK** - Most features are compatible, minor configuration adjustments needed

## Support Status

- **PostgreSQL 15:** Supported until **November 11, 2027** (still receiving security patches)
- **PostgreSQL 18:** Released September 25, 2025, supported until **November 14, 2030**

## Compatibility Analysis

### ✅ Compatible Features

The following PostgreSQL features used in this codebase are fully compatible with PostgreSQL 18:

1. **UUID Functions**
   - `gen_random_uuid()` - ✅ Standard function, fully supported
   - Used extensively in migrations and repository code

2. **PL/pgSQL Functions**
   - `CREATE OR REPLACE FUNCTION` - ✅ Fully supported
   - Custom functions: `get_revenue_level_tier()`, `scenario_matches()`
   - Anonymous blocks: `DO $$ ... END $$;` - ✅ Fully supported

3. **SQL Features**
   - `ON CONFLICT DO UPDATE` / `ON CONFLICT DO NOTHING` - ✅ Fully supported
   - `COALESCE()` - ✅ Fully supported
   - `ROW_NUMBER() OVER()` - ✅ Window functions fully supported
   - `ILIKE` - ✅ Case-insensitive pattern matching fully supported
   - `JSONB` data type - ✅ Fully supported
   - `CHECK` constraints - ✅ Fully supported
   - `UNIQUE` constraints - ✅ Fully supported
   - Foreign keys with `ON DELETE CASCADE` - ✅ Fully supported

4. **Data Types**
   - `UUID` - ✅ Fully supported
   - `VARCHAR`, `TEXT` - ✅ Fully supported
   - `DECIMAL` - ✅ Fully supported
   - `BOOLEAN` - ✅ Fully supported
   - `DATE`, `TIMESTAMP` - ✅ Fully supported
   - `INTEGER` - ✅ Fully supported
   - `JSONB` - ✅ Fully supported

5. **Go Driver Compatibility**
   - `github.com/lib/pq v1.10.9` - ✅ Compatible with PostgreSQL 18
   - Standard `database/sql` interface - ✅ Fully compatible

### ⚠️ Potential Issues & Mitigation

#### 1. Data Checksums (Low Impact)

**Issue:** PostgreSQL 18 enables data checksums by default for new clusters.

**Impact:** 
- If upgrading via `pg_upgrade`, checksum settings must match between source and target
- If creating new cluster, checksums will be enabled by default

**Mitigation:**
- Use `pg_upgrade` which preserves checksum settings from source cluster
- Or explicitly set `--no-data-checksums` if you want to disable (not recommended)
- **Action Required:** None - `pg_upgrade` handles this automatically

#### 2. Docker Volume Path Changes (Medium Impact)

**Issue:** PostgreSQL 18 Docker images may use different data directory paths.

**Impact:**
- Volume mounts in `docker-compose.yml` may need adjustment
- Current path: `/var/lib/postgresql/data`
- PostgreSQL 18 Alpine may use: `/var/lib/postgresql` with version-specific subfolders

**Mitigation:**
- Test Docker setup with PostgreSQL 18 image
- Update volume paths if needed
- **Action Required:** Test and update `docker-compose.yml` if necessary

**Current Configuration:**
```yaml
volumes:
  - postgres_data:/var/lib/postgresql/data
```

**Recommended Test:**
```yaml
# Test with PostgreSQL 18
image: postgres:18-alpine
volumes:
  - postgres_data:/var/lib/postgresql/data  # May need to change
```

#### 3. MD5 Authentication Deprecation (No Impact)

**Issue:** MD5 password authentication is deprecated in PostgreSQL 18.

**Impact:** None - This codebase uses standard password authentication, not MD5.

**Current Configuration:**
- Uses `POSTGRES_USER` and `POSTGRES_PASSWORD` environment variables
- Standard password authentication (SCRAM-SHA-256 compatible)

**Action Required:** None

#### 4. Time Zone Abbreviation Handling (Low Impact)

**Issue:** Time zone abbreviation handling has changed in PostgreSQL 18.

**Impact:** 
- Code uses `TIMESTAMP DEFAULT CURRENT_TIMESTAMP` which uses server timezone
- No explicit timezone abbreviations found in code

**Mitigation:**
- Monitor for any timezone-related issues during testing
- Use explicit timezone names if needed (e.g., `Asia/Bangkok`)

**Action Required:** Test timezone handling in staging environment

### ✅ Not Applicable (No Issues)

The following PostgreSQL 18 changes do not affect this codebase:

1. **Partitioned Tables** - Not used in this codebase
2. **Inheritance Tables** - Not used in this codebase
3. **Unlogged Tables** - Not used in this codebase
4. **Triggers** - Not used in this codebase
5. **CSV COPY FROM** - Not used in this codebase
6. **System Catalog Dependencies** - No direct system catalog queries found
7. **Extensions** - No PostgreSQL extensions used (only standard features)

## Code Analysis Summary

### Database Features Used

| Feature | Usage Count | Status |
|---------|------------|--------|
| `gen_random_uuid()` | 30+ | ✅ Compatible |
| `ON CONFLICT` | 10+ | ✅ Compatible |
| `COALESCE()` | 5+ | ✅ Compatible |
| `ROW_NUMBER() OVER()` | 1 | ✅ Compatible |
| `ILIKE` | 1 | ✅ Compatible |
| PL/pgSQL Functions | 2 | ✅ Compatible |
| `DO $$` blocks | 10+ | ✅ Compatible |
| JSONB columns | 2 | ✅ Compatible |

### Repository Files Analyzed

- ✅ `backend/internal/repositories/postgres/migrations.go` - All SQL compatible
- ✅ `backend/internal/repositories/postgres/repositories.go` - All queries compatible
- ✅ `backend/internal/repositories/postgres/doctor_schedule_repos.go` - Compatible
- ✅ `backend/internal/repositories/postgres/branch_constraints_repo.go` - Compatible
- ✅ `backend/internal/repositories/postgres/scenario_position_requirement_repo.go` - Compatible
- ✅ `backend/internal/repositories/postgres/branch_weekly_revenue_repo.go` - Compatible

## Migration Checklist

### Pre-Migration

- [ ] Review this compatibility report
- [ ] Backup current database (use `scripts/backup-database.ps1`)
- [ ] Test backup restoration process
- [ ] Review Docker volume configuration
- [ ] Check Go driver version compatibility (current: `lib/pq v1.10.9` ✅)

### Migration Steps

1. **Update Docker Configuration**
   ```yaml
   # docker-compose.yml
   postgres:
     image: postgres:18-alpine  # Change from postgres:15-alpine
   ```

2. **Test in Staging Environment**
   - Deploy PostgreSQL 18 to staging
   - Run all migrations
   - Execute full test suite
   - Verify all functionality

3. **Production Migration**
   - Schedule maintenance window
   - Backup production database
   - Use `pg_upgrade` or dump/restore method
   - Verify data integrity
   - Monitor application performance

### Post-Migration

- [ ] Verify all migrations run successfully
- [ ] Test all API endpoints
- [ ] Verify timezone handling
- [ ] Monitor query performance
- [ ] Check application logs for errors
- [ ] Update documentation

## Testing Recommendations

### 1. Unit Tests
- Run existing Go unit tests (`backend/tests/`)
- All tests should pass without modification

### 2. Integration Tests
- Test database migrations
- Test all repository methods
- Test custom PL/pgSQL functions

### 3. E2E Tests
- Run Playwright tests (`frontend/tests/e2e/`)
- Verify all user workflows

### 4. Performance Testing
- Compare query execution times
- Monitor for query plan changes
- Test with production-like data volumes

## Risk Assessment

| Risk Level | Item | Mitigation |
|------------|------|------------|
| **Low** | Data checksums | Handled by `pg_upgrade` |
| **Low** | Timezone handling | Test in staging |
| **Medium** | Docker volume paths | Test and update if needed |
| **None** | MD5 authentication | Not used |
| **None** | Partitioned tables | Not used |
| **None** | Triggers | Not used |

## Recommended Timeline

1. **Week 1:** Review and approve migration plan
2. **Week 2:** Test PostgreSQL 18 in development environment
3. **Week 3:** Deploy to staging and run full test suite
4. **Week 4:** Production migration (if staging tests pass)

## Conclusion

**Overall Compatibility:** ✅ **EXCELLENT**

The codebase uses standard PostgreSQL features that are fully compatible with PostgreSQL 18. The migration should be straightforward with minimal code changes required. The main considerations are:

1. Docker configuration updates (if volume paths change)
2. Testing in staging environment
3. Monitoring for any performance changes

**Recommendation:** Proceed with migration planning. The codebase is well-positioned for PostgreSQL 18 upgrade.

## References

- [PostgreSQL 18 Release Notes](https://www.postgresql.org/docs/18/release-18.html)
- [PostgreSQL Versioning Policy](https://www.postgresql.org/support/versioning/)
- [pg_upgrade Documentation](https://www.postgresql.org/docs/current/pgupgrade.html)

---
title: PostgreSQL 18 Migration Plan
description: Detailed migration plan with timeline and checklists
version: 1.0.0
lastUpdated: 2025-01-08
---

# PostgreSQL 18 Migration Plan

## Executive Summary

**Current Status:** PostgreSQL 15 (supported until Nov 2027)  
**Target:** PostgreSQL 18  
**Compatibility:** ✅ Verified and tested  
**Risk Level:** Low (backward compatible, tested)  
**Estimated Downtime:** 30-60 minutes

## Migration Timeline

### Phase 1: Preparation (Week 1)

**Duration:** 1-2 days  
**No Downtime Required**

#### Tasks

- [ ] **Day 1: Verification**
  - [ ] Run database verification script
  - [ ] Document current database state
  - [ ] Verify backup procedures
  - [ ] Review migration guide

- [ ] **Day 2: Preparation**
  - [ ] Create backup of production database
  - [ ] Test backup restoration process
  - [ ] Review rollback procedures
  - [ ] Notify stakeholders of migration plan

#### Deliverables

- Database state report
- Verified backup
- Migration checklist
- Rollback plan

### Phase 2: Staging Migration (Week 2)

**Duration:** 1 day  
**Environment:** Staging

#### Tasks

- [ ] **Pre-Migration**
  - [ ] Backup staging database
  - [ ] Update staging docker-compose.yml
  - [ ] Prepare migration scripts

- [ ] **Migration**
  - [ ] Stop staging services
  - [ ] Export staging database
  - [ ] Start PostgreSQL 18 container
  - [ ] Import database
  - [ ] Run migrations
  - [ ] Verify data integrity

- [ ] **Post-Migration**
  - [ ] Run verification script
  - [ ] Test all application features
  - [ ] Monitor for errors
  - [ ] Document any issues

#### Success Criteria

- ✅ All tables migrated successfully
- ✅ All data intact
- ✅ Application functions correctly
- ✅ No errors in logs
- ✅ Performance acceptable

### Phase 3: Production Migration (Week 3-4)

**Duration:** 2-4 hours (including testing)  
**Environment:** Production  
**Downtime:** 30-60 minutes

#### Pre-Migration Checklist

**1 Week Before:**
- [ ] Schedule maintenance window
- [ ] Notify all users
- [ ] Prepare migration team
- [ ] Review staging migration results
- [ ] Finalize migration method

**1 Day Before:**
- [ ] Create fresh backup
- [ ] Verify backup integrity
- [ ] Test rollback procedure
- [ ] Prepare migration scripts
- [ ] Review communication plan

**Day Of Migration:**
- [ ] Final backup (just before migration)
- [ ] Notify users of maintenance start
- [ ] Verify all services are ready

#### Migration Steps

**Step 1: Pre-Migration Backup (5 minutes)**
```powershell
# Create final backup
.\scripts\backup-database.ps1

# Verify backup
docker exec vsq-manpower-db pg_restore --list backup.dump | Select-String "TABLE DATA"
```

**Step 2: Stop Services (2 minutes)**
```powershell
# Stop all services
docker-compose down

# Verify containers stopped
docker ps -a | Select-String "vsq-manpower"
```

**Step 3: Export Database (10-30 minutes)**
```powershell
# Export database
docker run --rm \
  -v vsq-oper_manpower_postgres_data:/var/lib/postgresql/data \
  -v ${PWD}:/backup \
  postgres:15-alpine \
  pg_dump -U vsq_user -d vsq_manpower -F c -f /backup/production_backup_$(Get-Date -Format "yyyyMMdd_HHmmss").dump

# Verify export
ls -lh production_backup_*.dump
```

**Step 4: Update Configuration (2 minutes)**
```yaml
# Update docker-compose.yml
postgres:
  image: postgres:18-alpine  # Changed from postgres:15-alpine
```

**Step 5: Start PostgreSQL 18 (5 minutes)**
```powershell
# Start new PostgreSQL 18 container
docker-compose up -d postgres

# Wait for health check
docker-compose ps postgres
# Should show: Healthy

# Verify version
docker exec vsq-manpower-db psql -U vsq_user -d vsq_manpower -c "SELECT version();"
# Should show: PostgreSQL 18.x
```

**Step 6: Import Database (10-30 minutes)**
```powershell
# Copy backup to container
docker cp production_backup_YYYYMMDD_HHMMSS.dump vsq-manpower-db:/tmp/backup.dump

# Restore database
docker exec vsq-manpower-db pg_restore -U vsq_user -d vsq_manpower -c /tmp/backup.dump

# Verify import
docker exec vsq-manpower-db psql -U vsq_user -d vsq_manpower -c "\dt" | Measure-Object -Line
# Should show: 32 tables
```

**Step 7: Run Migrations (2 minutes)**
```powershell
# Start backend - migrations run automatically
docker-compose up -d backend

# Check migration logs
docker-compose logs backend | Select-String "migration"
```

**Step 8: Verify Migration (10 minutes)**
```powershell
# Run verification script
.\scripts\verify-database-state.ps1 -Detailed

# Check application health
curl http://localhost:8081/health

# Test key endpoints
# Login, query data, etc.
```

**Step 9: Start All Services (2 minutes)**
```powershell
# Start frontend
docker-compose up -d frontend

# Verify all services
docker-compose ps
```

**Step 10: Post-Migration Verification (15 minutes)**
```powershell
# Run comprehensive checks
.\scripts\verify-database-state.ps1 -Detailed

# Check logs for errors
docker-compose logs --tail=100 | Select-String -Pattern "error|fail" -CaseSensitive:$false

# Test application features
# - Login
# - Create/read/update data
# - Run reports
# - Test all major workflows
```

#### Post-Migration Checklist

**Immediate (First Hour):**
- [ ] All services running
- [ ] Database version confirmed: PostgreSQL 18
- [ ] All tables present (32 tables)
- [ ] Application responding
- [ ] No errors in logs
- [ ] Health endpoint working

**First 24 Hours:**
- [ ] Monitor application logs
- [ ] Monitor database performance
- [ ] Check for user-reported issues
- [ ] Verify scheduled jobs running
- [ ] Monitor error rates

**First Week:**
- [ ] Review performance metrics
- [ ] Compare query performance
- [ ] Monitor for any issues
- [ ] Document lessons learned

## Rollback Plan

### If Migration Fails

**Step 1: Stop New Services (2 minutes)**
```powershell
docker-compose down
```

**Step 2: Restore Old Configuration**
```yaml
# Revert docker-compose.yml
postgres:
  image: postgres:15-alpine  # Revert to old version
```

**Step 3: Restore Database (10-30 minutes)**
```powershell
# Option A: Restore from volume (if available)
# Restore the old volume

# Option B: Restore from backup
docker-compose up -d postgres
docker cp production_backup_YYYYMMDD_HHMMSS.dump vsq-manpower-db:/tmp/backup.dump
docker exec vsq-manpower-db pg_restore -U vsq_user -d vsq_manpower -c /tmp/backup.dump
```

**Step 4: Start Services (2 minutes)**
```powershell
docker-compose up -d
```

**Step 5: Verify Rollback (5 minutes)**
```powershell
.\scripts\verify-database-state.ps1
curl http://localhost:8081/health
```

**Total Rollback Time:** 20-40 minutes

## Risk Assessment

### Low Risk Items ✅

- Database compatibility (tested)
- Migration code (backward compatible)
- Application code (no changes needed)
- Data integrity (verified)

### Medium Risk Items ⚠️

- **Downtime:** 30-60 minutes required
  - **Mitigation:** Schedule during low-traffic period
  - **Mitigation:** Have rollback plan ready

- **Data Loss:** Minimal risk with proper backups
  - **Mitigation:** Multiple backups before migration
  - **Mitigation:** Test backup restoration

- **Performance:** Unknown until tested
  - **Mitigation:** Monitor after migration
  - **Mitigation:** Test in staging first

### High Risk Items ❌

None identified - migration is low risk overall

## Communication Plan

### Stakeholders to Notify

1. **Development Team**
   - Notification: 1 week before
   - Update: Day before, day of, after completion

2. **End Users**
   - Notification: 3 days before
   - Maintenance window notice
   - Completion notification

3. **Management**
   - Notification: 1 week before
   - Status updates during migration
   - Completion report

### Communication Templates

**Pre-Migration Notice:**
```
Subject: Scheduled Maintenance - Database Upgrade

We will be performing a database upgrade on [DATE] from [TIME] to [TIME].

During this time, the application will be unavailable.

Expected downtime: 30-60 minutes

We apologize for any inconvenience.
```

**Post-Migration Notice:**
```
Subject: Maintenance Complete - System Restored

The database upgrade has been completed successfully.

The application is now available and running on PostgreSQL 18.

If you experience any issues, please contact support.
```

## Success Metrics

### Technical Metrics

- ✅ Database version: PostgreSQL 18.x
- ✅ All tables migrated: 32/32
- ✅ Data integrity: 100%
- ✅ Application uptime: >99.9%
- ✅ Error rate: <0.1%

### Business Metrics

- ✅ Zero data loss
- ✅ Minimal downtime (<60 minutes)
- ✅ No user-reported issues
- ✅ Performance maintained or improved

## Post-Migration Tasks

### Immediate (Day 1)

- [ ] Update documentation
- [ ] Update CHANGELOG.md
- [ ] Notify stakeholders
- [ ] Monitor systems

### Short Term (Week 1)

- [ ] Performance analysis
- [ ] User feedback collection
- [ ] Issue resolution
- [ ] Documentation updates

### Long Term (Month 1)

- [ ] Performance comparison report
- [ ] Migration retrospective
- [ ] Process improvements
- [ ] Knowledge sharing

## Resources

### Scripts

- `scripts/verify-database-state.ps1` - Database verification
- `scripts/backup-database.ps1` - Database backup
- `scripts/test-pg18-compatibility.ps1` - Compatibility testing

### Documentation

- [Migration Guide](./postgresql-18-migration-guide.md)
- [Compatibility Report](./postgresql-18-compatibility-report.md)
- [Test Results](./postgresql-18-test-results.md)
- [Migration Fix Summary](./migration-fix-summary.md)

### External Resources

- [PostgreSQL 18 Release Notes](https://www.postgresql.org/docs/18/release-18.html)
- [PostgreSQL Upgrade Guide](https://www.postgresql.org/docs/current/pgupgrade.html)

## Approval

**Prepared by:** [Your Name]  
**Date:** 2025-01-08  
**Reviewed by:** [ ]  
**Approved by:** [ ]

---

## Quick Reference

### Migration Command Sequence

```powershell
# 1. Backup
.\scripts\backup-database.ps1

# 2. Stop services
docker-compose down

# 3. Update docker-compose.yml (change postgres:15-alpine to postgres:18-alpine)

# 4. Export database
docker exec vsq-manpower-db pg_dump -U vsq_user -d vsq_manpower -F c -f /tmp/backup.dump
docker cp vsq-manpower-db:/tmp/backup.dump ./backup.dump

# 5. Start PostgreSQL 18
docker-compose up -d postgres

# 6. Import database
docker cp ./backup.dump vsq-manpower-db:/tmp/backup.dump
docker exec vsq-manpower-db pg_restore -U vsq_user -d vsq_manpower -c /tmp/backup.dump

# 7. Start backend (runs migrations)
docker-compose up -d backend

# 8. Verify
.\scripts\verify-database-state.ps1 -Detailed

# 9. Start frontend
docker-compose up -d frontend
```

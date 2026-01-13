---
title: Cursor Rules Gap Analysis
description: Files and directories in the project not yet referenced in .cursorrules
version: 1.0.0
lastUpdated: 2026-01-08 15:35:47
---

# Cursor Rules Gap Analysis

This document identifies files and directories in the VSQ Operations Manpower project that exist but are not yet referenced in `.cursorrules`.

## Summary

**Total Categories:** 8  
**Total Files/Directories:** 30+

---

## 1. Root Level Documentation Files

### Missing References

- **`SOFTWARE_REQUIREMENTS.md`** ⚠️
  - **Status:** Exists but cursor rules reference `requirements.md` instead
  - **Issue:** Cursor rules mention `requirements.md` or `docs/requirements.md`, but actual file is `SOFTWARE_REQUIREMENTS.md`
  - **Recommendation:** Update cursor rules to reference `SOFTWARE_REQUIREMENTS.md` OR add alias reference

- **`SOFTWARE_ARCHITECTURE.md`** ❌
  - **Status:** Not mentioned in cursor rules
  - **Content:** Architecture documentation
  - **Recommendation:** Add reference in Rule #8 (Documentation Updates) or Rule #15 (Project Structure)

- **`PROJECT_SETUP_EXPORT.md`** ❌
  - **Status:** Not mentioned in cursor rules
  - **Recommendation:** Add to project structure awareness

---

## 2. Docker Compose Files

### Missing References

- **`docker-compose.production.yml`** ❌
  - **Status:** Not mentioned in cursor rules
  - **Purpose:** Production environment configuration
  - **Recommendation:** Add to Rule #8 (Documentation Updates) under Deployment section

- **`docker-compose.staging.yml`** ❌
  - **Status:** Not mentioned in cursor rules
  - **Purpose:** Staging environment configuration
  - **Recommendation:** Add to Rule #8 (Documentation Updates) under Deployment section

- **`docker-compose.yml`** ⚠️
  - **Status:** Implicitly referenced but not explicitly mentioned
  - **Recommendation:** Explicitly mention in Rule #15 (Project Structure)

---

## 3. Nginx Configuration Directory

### Missing References

- **`nginx/` directory** ❌
  - **Status:** Not mentioned in cursor rules
  - **Contents:**
    - `nginx/production.conf` - Production nginx configuration
    - `nginx/staging.conf` - Staging nginx configuration
    - `nginx/README.md` - Nginx documentation
    - `nginx/ssl/` - SSL certificates directory
    - `nginx/logs/` - Log files directory
  - **Recommendation:** Add to Rule #15 (Project Structure) and Rule #8 (Documentation Updates)

---

## 4. Scripts Directory

### Missing References

- **`scripts/backup-database.ps1`** ❌
  - **Status:** Not mentioned in cursor rules
  - **Purpose:** Database backup utility
  - **Recommendation:** Add to Rule #15 (Project Structure) or create new rule for utility scripts

- **`scripts/deploy-production.ps1`** ❌
  - **Status:** Not mentioned in cursor rules
  - **Purpose:** Production deployment script
  - **Recommendation:** Add to Rule #8 (Documentation Updates) under Deployment

- **`scripts/deploy-staging.ps1`** ❌
  - **Status:** Not mentioned in cursor rules
  - **Purpose:** Staging deployment script
  - **Recommendation:** Add to Rule #8 (Documentation Updates) under Deployment

- **`scripts/force-stop-containers.ps1`** ❌
  - **Status:** Not mentioned in cursor rules
  - **Purpose:** Force stop Docker containers utility
  - **Recommendation:** Add to utility scripts section

- **`scripts/README.md`** ❌
  - **Status:** Not mentioned in cursor rules
  - **Purpose:** Scripts documentation
  - **Recommendation:** Reference in Rule #8 (Documentation Updates)

- **`scripts/update-build-timestamp.sh`** ⚠️
  - **Status:** Bash version exists but only PowerShell version mentioned
  - **Recommendation:** Mention both versions or note cross-platform support

---

## 5. Backend Utilities and Tools

### Missing References

- **`backend/cmd/check-link-user/`** ❌
  - **Status:** Not mentioned in cursor rules
  - **Purpose:** Utility to check user linking
  - **Recommendation:** Add to Rule #15 (Project Structure) under backend utilities

- **`backend/cmd/create-branch-managers/`** ❌
  - **Status:** Not mentioned in cursor rules
  - **Purpose:** Utility to create branch managers
  - **Contents:** `main.go`, `README.md`
  - **Recommendation:** Add to Rule #15 (Project Structure) under backend utilities

- **`backend/cmd/update-admin-password/`** ❌
  - **Status:** Not mentioned in cursor rules
  - **Purpose:** Utility to update admin password
  - **Recommendation:** Add to Rule #15 (Project Structure) under backend utilities

- **`backend/tmp/` directory** ❌
  - **Status:** Not mentioned in cursor rules
  - **Purpose:** Temporary utility scripts
  - **Contents:** `hash_password.go`, `verify-branch-id.go`
  - **Recommendation:** Add note about temporary utilities directory

---

## 6. Documentation Files in `docs/`

### Missing Specific References

While `docs/` directory is mentioned generally, these specific files are not referenced:

- **`docs/STAGING_PRODUCTION_SETUP.md`** ❌
  - **Recommendation:** Reference in Rule #8 (Documentation Updates) under Deployment

- **`docs/deployment-guide.md`** ⚠️
  - **Status:** Mentioned in Rule #8 but not explicitly listed
  - **Recommendation:** Already referenced, but could be more explicit

- **`docs/version-build-process.md`** ❌
  - **Status:** Not mentioned despite being related to versioning
  - **Recommendation:** Add to Rule #4 (Version Management)

- **`docs/versioning-rules.md`** ❌
  - **Status:** Not mentioned despite being related to versioning
  - **Recommendation:** Add to Rule #4 (Version Management)

- **`docs/docker-compose-commands.md`** ❌
  - **Recommendation:** Reference in Rule #8 (Documentation Updates)

- **`docs/docker-dev-hot-reload.md`** ❌
  - **Recommendation:** Reference in Rule #8 (Documentation Updates)

- **`docs/docker-setup-analysis.md`** ❌
  - **Recommendation:** Reference in Rule #8 (Documentation Updates)

- **`docs/hybrid-development-setup.md`** ❌
  - **Recommendation:** Reference in Rule #8 (Documentation Updates)

- **`docs/requirements-analysis.md`** ❌
  - **Recommendation:** Reference in Rule #1 (Documentation-First Development)

- **`docs/requirements-gap-analysis.md`** ❌
  - **Recommendation:** Reference in Rule #1 (Documentation-First Development)

- **`docs/branch-*.md` files** ❌
  - Multiple branch-related documentation files:
    - `branch-code-population-functions.md`
    - `branch-code-usage-locations.md`
    - `branch-data-structure-development-plan.md`
    - `branch-database-analysis.md`
  - **Recommendation:** Reference in Rule #8 (Documentation Updates)

- **`docs/rotation-*.md` files** ❌
  - Rotation staff related documentation:
    - `rotation-staff-assignment-ui-alternatives.md`
    - `rotation-staff-implementation-alternatives.md`
  - **Recommendation:** Reference in Rule #8 (Documentation Updates)

- **`docs/conflict-resolution-rules.md`** ⚠️
  - **Status:** Referenced in SOFTWARE_REQUIREMENTS.md but not in cursor rules
  - **Recommendation:** Add to Rule #7 (Business Rule Enforcement)

- **`docs/database-current-state.md`** ❌
  - **Recommendation:** Reference in Rule #8 (Documentation Updates) under Database schema

- **`docs/README.md`** ❌
  - **Recommendation:** Reference in Rule #8 (Documentation Updates)

---

## 7. Version Files

### Current Status

- **`frontend/VERSION.json`** ✅ Mentioned in Rule #4
- **`frontend/public/VERSION.json`** ✅ Mentioned in Rule #4
- **`backend/VERSION.json`** ✅ Mentioned in Rule #4
- **`backend/DATABASE_VERSION.json`** ✅ Mentioned in Rule #4

**Status:** All version files are properly referenced.

---

## 8. Configuration Files

### Missing References

- **`frontend/next.config.js`** ❌
  - **Recommendation:** Add to Rule #15 (Project Structure)

- **`frontend/playwright.config.ts`** ❌
  - **Recommendation:** Add to Rule #17 (Testing Standards)

- **`frontend/tailwind.config.js`** ❌
  - **Recommendation:** Add to Rule #15 (Project Structure)

- **`frontend/postcss.config.js`** ❌
  - **Recommendation:** Add to Rule #15 (Project Structure)

- **`backend/go.mod`** ⚠️
  - **Status:** Implicitly referenced but not explicitly mentioned
  - **Recommendation:** Explicitly mention in Rule #15 (Project Structure)

- **`backend/go.sum`** ❌
  - **Recommendation:** Add to Rule #15 (Project Structure)

- **`backend/Dockerfile`** ⚠️
  - **Status:** Implicitly referenced but not explicitly mentioned
  - **Recommendation:** Explicitly mention in Rule #15 (Project Structure)

- **`backend/Dockerfile.dev`** ❌
  - **Recommendation:** Add to Rule #15 (Project Structure)

- **`frontend/Dockerfile`** ⚠️
  - **Status:** Implicitly referenced but not explicitly mentioned
  - **Recommendation:** Explicitly mention in Rule #15 (Project Structure)

- **`frontend/Dockerfile.dev`** ❌
  - **Recommendation:** Add to Rule #15 (Project Structure)

---

## Recommendations Summary

### High Priority

1. **Fix `requirements.md` reference** - Update Rule #1, #2, #14 to reference `SOFTWARE_REQUIREMENTS.md`
2. **Add nginx directory** - Add to Rule #15 (Project Structure)
3. **Add deployment scripts** - Add to Rule #8 (Documentation Updates)
4. **Add versioning documentation** - Add `docs/version-build-process.md` and `docs/versioning-rules.md` to Rule #4

### Medium Priority

5. **Add backend utilities** - Document `backend/cmd/` utilities in Rule #15
6. **Add configuration files** - Document config files in Rule #15
7. **Add docker-compose variants** - Document staging/production compose files

### Low Priority

8. **Add specific docs references** - Reference specific documentation files in Rule #8
9. **Add scripts README** - Reference `scripts/README.md` in Rule #8

---

## Next Steps

1. Review this analysis with the team
2. Prioritize which references to add to `.cursorrules`
3. Update `.cursorrules` with missing references
4. Test that cursor rules work correctly with new references
5. Update this document when changes are made

---

## Version History

- **v1.0.0** (2026-01-08): Initial gap analysis created

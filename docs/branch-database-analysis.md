---
title: Branch Database Analysis
description: Analysis of branch code data flow and current database state
version: 1.0.0
lastUpdated: 2025-12-21 09:21:44
---

# Branch Database Analysis

## Question 1: Are All Places Pulling Branch Codes from Database?

### ✅ YES - All Frontend Components Pull from Database

All frontend components fetch branch data from the database via API calls:

#### Frontend Data Flow:
```
Frontend Component
  ↓
branchApi.list() [frontend/src/lib/api/branch.ts]
  ↓
GET /api/branches [HTTP Request]
  ↓
BranchHandler.List() [backend/internal/handlers/branch_handler.go:31]
  ↓
repos.Branch.List() [backend/internal/repositories/postgres/repositories.go:418]
  ↓
SELECT ... FROM branches ORDER BY code [SQL Query]
  ↓
PostgreSQL Database
```

#### Verified Locations:

1. **Staff Scheduling Page** (`staff-scheduling/page.tsx:27`)
   ```typescript
   const branchesData = await branchApi.list();
   ```
   ✅ Pulls from database

2. **Staff Management Page** (`staff-management/page.tsx:42`)
   ```typescript
   branchApi.list()
   ```
   ✅ Pulls from database

3. **Branch Management Page** (`branch-management/page.tsx:60`)
   ```typescript
   const branchesData = await branchApi.list();
   ```
   ✅ Pulls from database

4. **Rotation Assignment View** (`RotationAssignmentView.tsx:46`)
   ```typescript
   branchApi.list()
   ```
   ✅ Pulls from database

5. **Dashboard Handler** (`dashboard_handler.go:21`)
   ```go
   branches, err := h.repos.Branch.List()
   ```
   ✅ Pulls from database

6. **Rotation Suggestions Handler** (`rotation_handler.go:187`)
   ```go
   allBranches, err := h.repos.Branch.List()
   ```
   ✅ Pulls from database

### ❌ NO Hardcoded Branches Found

**Search Results:** No hardcoded branch lists found in frontend or backend code.

**Only Constants:** The `StandardBranchCodes` constant in `backend/internal/constants/branches.go` is used for:
- Validation (preventing deletion/modification)
- Seeding (creating branches in database)
- **NOT** for displaying branches to users

---

## Question 2: What is the Current Content of Branch Database?

### Expected Database State After Seeding

When the system initializes, `SeedStandardBranches()` should create **35 branches** with these codes:

```
CPN, CPN-LS, CTR, PNK, CNK, BNA, CLP, SQR, BKP, CMC, 
CSA, EMQ, ESV, GTW, MGA, MTA, PRM, RCT, RST, TMA, 
MBA, SCN, CWG, CRM, CWT, PSO, RCP, CRA, CTW, ONE, 
DCP, MNG, TLR, TLR-LS, TLR-WN
```

### Database Schema

**Table:** `branches`

**Columns:**
- `id` (UUID) - Primary key, deterministic UUID generated from branch code
- `name` (VARCHAR(255)) - Defaults to branch code (e.g., "CPN")
- `code` (VARCHAR(50)) - UNIQUE, NOT NULL - The branch code (e.g., "CPN")
- `address` (TEXT) - Empty string by default
- `area_manager_id` (UUID) - NULL by default
- `expected_revenue` (DECIMAL) - 0 by default
- `priority` (INTEGER) - 0 by default
- `created_at` (TIMESTAMP) - Auto-generated
- `updated_at` (TIMESTAMP) - Auto-generated

**Constraints:**
- `code` is UNIQUE (prevents duplicate codes)
- `code` is NOT NULL (required)

### Seeding Process

**Function:** `SeedStandardBranches()` in `migrations.go:196`

**Called:** Automatically after migrations in `RunMigrations()` (line 23)

**Process:**
1. Gets standard codes from `constants.GetStandardBranchCodes()`
2. For each code:
   - Generates deterministic UUID using SHA1 hash
   - Inserts branch with default values
   - Uses `ON CONFLICT (code) DO NOTHING` to prevent duplicates

**Default Values:**
- `name` = code (e.g., "CPN")
- `code` = standard code (e.g., "CPN")
- `address` = "" (empty)
- `expected_revenue` = 0
- `priority` = 0

---

## How to Check Current Database State

### Method 1: Query Database Directly

**Using psql:**
```bash
# Connect to database
psql -h localhost -p 5434 -U vsq_user -d vsq_manpower

# Query branches
SELECT code, name, address, expected_revenue, priority, created_at 
FROM branches 
ORDER BY code;

# Count branches
SELECT COUNT(*) FROM branches;

# Check if all standard codes exist
SELECT code FROM branches ORDER BY code;
```

**Using Docker:**
```bash
# Connect to database container
docker exec -it vsq-manpower-db psql -U vsq_user -d vsq_manpower

# Then run SQL queries above
```

### Method 2: Use API Endpoint

**Call Branch List API:**
```bash
# Get all branches
curl http://localhost:8081/api/branches

# Or in browser
http://localhost:8081/api/branches
```

**Expected Response:**
```json
{
  "branches": [
    {
      "id": "uuid-here",
      "name": "CPN",
      "code": "CPN",
      "address": "",
      "expected_revenue": 0,
      "priority": 0,
      "created_at": "2025-12-21T...",
      "updated_at": "2025-12-21T..."
    },
    // ... 34 more branches
  ]
}
```

### Method 3: Check Frontend UI

**Branch Management Page:**
- Navigate to: `http://localhost:4000/branch-management`
- View the branches table
- Should show 35 branches with codes in first column

---

## Current Database State Scenarios

### Scenario 1: Fresh Database (No Branches)
**Status:** Empty `branches` table
**After Server Start:**
- Migrations run → Creates `branches` table
- Seeding runs → Creates 35 standard branches
- **Result:** 35 branches in database

### Scenario 2: Existing Database (Some Branches)
**Status:** Some branches already exist
**After Server Start:**
- Migrations run → Table already exists, no changes
- Seeding runs → Creates missing standard branches
- Uses `ON CONFLICT (code) DO NOTHING` → Skips existing branches
- **Result:** All 35 standard branches exist (some may have been updated)

### Scenario 3: Database with Custom Branches
**Status:** Standard branches + custom branches
**After Server Start:**
- Seeding ensures all 35 standard branches exist
- Custom branches remain untouched
- **Result:** 35+ branches (35 standard + any custom)

### Scenario 4: Database with Modified Standard Branches
**Status:** Standard branches exist but have been modified (name, address, etc.)
**After Server Start:**
- Seeding uses `ON CONFLICT (code) DO NOTHING`
- Existing branches are NOT overwritten
- **Result:** Modified branches remain as-is (good - preserves user changes)

---

## Verification Checklist

### ✅ Code Verification (Completed)

- [x] All frontend components use `branchApi.list()`
- [x] All backend handlers use `repos.Branch.List()`
- [x] Repository queries database with `SELECT ... FROM branches`
- [x] No hardcoded branch lists in UI code
- [x] Seeding function is called after migrations
- [x] Seeding uses constants from `branches.go`

### ⚠️ Runtime Verification (Needs Testing)

- [ ] Database contains `branches` table
- [ ] Seeding function executes successfully
- [ ] All 35 standard branches exist in database
- [ ] Branch codes match standard codes exactly
- [ ] API returns all branches correctly
- [ ] Frontend displays branches correctly

---

## Potential Issues

### Issue 1: Seeding Not Running
**Symptom:** Database empty or missing standard branches
**Cause:** 
- Migration error preventing seeding
- Seeding function error not caught
- Database connection issue

**Solution:**
- Check server logs for migration/seeding errors
- Verify `SeedStandardBranches()` is called in `RunMigrations()`
- Manually run seeding if needed

### Issue 2: Duplicate Branch Codes
**Symptom:** Error when creating branches
**Cause:** 
- Manual insertion of duplicate codes
- Seeding running multiple times (should be prevented by `ON CONFLICT`)

**Solution:**
- Check for duplicate codes: `SELECT code, COUNT(*) FROM branches GROUP BY code HAVING COUNT(*) > 1;`
- Remove duplicates manually
- Re-run seeding

### Issue 3: Missing Standard Branches
**Symptom:** Some standard codes missing from database
**Cause:**
- Seeding failed for specific branches
- Manual deletion (should be prevented by validation)

**Solution:**
- Check which codes are missing
- Verify seeding function completed successfully
- Manually insert missing branches or re-run seeding

---

## Recommendations

### 1. Add Database Verification Endpoint
Create an admin endpoint to verify standard branches:
```go
GET /api/admin/verify-standard-branches
```
Returns:
- List of missing standard codes
- List of extra branches
- Verification status

### 2. Add Logging to Seeding
Log when branches are created:
```go
log.Printf("Seeded branch: %s (ID: %s)", code, branchID)
```

### 3. Add Health Check
Include branch count in health check endpoint:
```go
GET /health
{
  "status": "ok",
  "branches": {
    "total": 35,
    "standard": 35,
    "custom": 0
  }
}
```

### 4. Add Migration Status
Track which migrations/seeds have run to prevent re-running.

---

## Summary

### Data Flow: ✅ All Database-Driven
- **Frontend:** All components fetch from API → Database
- **Backend:** All handlers query database
- **No Hardcoded Data:** Only constants for validation/seeding

### Database State: ⚠️ Depends on Runtime
- **Expected:** 35 standard branches after seeding
- **Actual:** Need to verify by querying database
- **Seeding:** Runs automatically on server start

### Next Steps:
1. ✅ Verify code is database-driven (COMPLETED)
2. ⚠️ Query database to check current state (NEEDS VERIFICATION)
3. ⚠️ Test seeding function (NEEDS TESTING)
4. ⚠️ Verify all 35 branches exist (NEEDS VERIFICATION)





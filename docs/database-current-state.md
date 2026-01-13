---
title: Current Database State
description: Complete inventory of all tables and their current content
version: 1.0.0
lastUpdated: 2025-12-21 09:21:44
---

# Current Database State

## Database Information

- **Database Name:** `vsq_manpower`
- **Database User:** `vsq_user`
- **Database Type:** PostgreSQL 15
- **Container:** `vsq-manpower-db`
- **Port:** 5434 (host) → 5432 (container)

---

## Tables Overview

The database contains **11 tables**:

| Table Name | Row Count | Status | Description |
|------------|-----------|--------|-------------|
| `roles` | 5 | ✅ Seeded | User roles (admin, area_manager, etc.) |
| `users` | 1 | ✅ Has Data | User accounts |
| `positions` | 7 | ✅ Seeded | Staff positions (Branch Manager, Nurse, etc.) |
| `branches` | **0** | ⚠️ **EMPTY** | Branch locations (should have 35 standard branches) |
| `staff` | 0 | Empty | Staff members |
| `effective_branches` | 0 | Empty | Rotation staff effective branch assignments |
| `revenue_data` | 0 | Empty | Branch revenue tracking |
| `staff_schedules` | 0 | Empty | Staff work schedules |
| `rotation_assignments` | 0 | Empty | Rotation staff assignments |
| `system_settings` | 0 | Empty | System configuration |
| `staff_allocation_rules` | 0 | Empty | Staff allocation business rules |

---

## Table Details

### 1. `roles` Table (5 rows)

**Schema:**
```sql
CREATE TABLE roles (
    id UUID PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**Current Content:**
| id | name | created_at |
|----|------|------------|
| 00000000-0000-0000-0000-000000000001 | admin | 2025-12-15 12:31:20 |
| 00000000-0000-0000-0000-000000000002 | area_manager | 2025-12-15 12:31:20 |
| 00000000-0000-0000-0000-000000000003 | district_manager | 2025-12-15 12:31:20 |
| 00000000-0000-0000-0000-000000000004 | branch_manager | 2025-12-15 12:31:20 |
| 00000000-0000-0000-0000-000000000005 | viewer | 2025-12-15 12:31:20 |

**Status:** ✅ Seeded correctly with default roles

---

### 2. `users` Table (1 row)

**Schema:**
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY,
    username VARCHAR(100) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role_id UUID NOT NULL REFERENCES roles(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**Current Content:**
| id | username | email | role_id | created_at |
|----|----------|-------|---------|------------|
| 3e33c827-4f66-4ea3-9d8f-59634d61ba77 | admin | admin@vsq.com | 00000000-0000-0000-0000-000000000001 | 2025-12-15 16:41:23 |

**Status:** ✅ Has 1 admin user

---

### 3. `positions` Table (7 rows)

**Schema:**
```sql
CREATE TABLE positions (
    id UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    min_staff_per_branch INTEGER DEFAULT 1,
    revenue_multiplier DECIMAL(10,4) DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**Current Content:**
| id | name | min_staff_per_branch | revenue_multiplier |
|----|------|---------------------|-------------------|
| 10000000-0000-0000-0000-000000000001 | Branch Manager | 1 | 0.0000 |
| 10000000-0000-0000-0000-000000000002 | Assistant Branch Manager | 0 | 0.5000 |
| 10000000-0000-0000-0000-000000000003 | Service Consultant | 1 | 1.0000 |
| 10000000-0000-0000-0000-000000000004 | Coordinator | 1 | 0.8000 |
| 10000000-0000-0000-0000-000000000005 | Doctor Assistant | 2 | 1.2000 |
| 10000000-0000-0000-0000-000000000006 | Physiotherapist | 1 | 1.0000 |
| 10000000-0000-0000-0000-000000000007 | Nurse | 2 | 1.0000 |

**Status:** ✅ Seeded correctly with 7 default positions

---

### 4. `branches` Table (0 rows) ⚠️

**Schema:**
```sql
CREATE TABLE branches (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50) UNIQUE NOT NULL,
    address TEXT,
    area_manager_id UUID REFERENCES users(id),
    expected_revenue DECIMAL(15,2) DEFAULT 0,
    priority INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**Current Content:**
```
(0 rows) - TABLE IS EMPTY
```

**Expected Content:**
Should contain **35 standard branch codes**:
- CPN, CPN-LS, CTR, PNK, CNK, BNA, CLP, SQR, BKP, CMC, CSA, EMQ, ESV, GTW, MGA, MTA, PRM, RCT, RST, TMA, MBA, SCN, CWG, CRM, CWT, PSO, RCP, CRA, CTW, ONE, DCP, MNG, TLR, TLR-LS, TLR-WN

**Status:** ⚠️ **EMPTY - Seeding function needs to run**

**Issue:** The `SeedStandardBranches()` function was added after the database was created. The seeding function runs automatically when the server starts, but if the server was started before the seeding function was added, the branches won't exist.

**Solution:** Restart the backend server to trigger the seeding function, or manually run the seeding.

---

### 5. `staff` Table (0 rows)

**Schema:**
```sql
CREATE TABLE staff (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    staff_type VARCHAR(20) NOT NULL CHECK (staff_type IN ('branch', 'rotation')),
    position_id UUID NOT NULL REFERENCES positions(id),
    branch_id UUID REFERENCES branches(id),
    coverage_area VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**Current Content:**
```
(0 rows) - Empty (no staff members added yet)
```

**Status:** ✅ Empty (expected - needs manual data entry)

---

### 6. `effective_branches` Table (0 rows)

**Schema:**
```sql
CREATE TABLE effective_branches (
    id UUID PRIMARY KEY,
    rotation_staff_id UUID NOT NULL REFERENCES staff(id),
    branch_id UUID NOT NULL REFERENCES branches(id),
    level INTEGER NOT NULL CHECK (level IN (1, 2)),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(rotation_staff_id, branch_id)
);
```

**Current Content:**
```
(0 rows) - Empty (no effective branch assignments yet)
```

**Status:** ✅ Empty (expected - requires staff and branches to exist first)

---

### 7. `revenue_data` Table (0 rows)

**Schema:**
```sql
CREATE TABLE revenue_data (
    id UUID PRIMARY KEY,
    branch_id UUID NOT NULL REFERENCES branches(id),
    date DATE NOT NULL,
    expected_revenue DECIMAL(15,2) NOT NULL,
    actual_revenue DECIMAL(15,2),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(branch_id, date)
);
```

**Current Content:**
```
(0 rows) - Empty (no revenue data entered yet)
```

**Status:** ✅ Empty (expected - requires branches to exist first)

---

### 8. `staff_schedules` Table (0 rows)

**Schema:**
```sql
CREATE TABLE staff_schedules (
    id UUID PRIMARY KEY,
    staff_id UUID NOT NULL REFERENCES staff(id),
    branch_id UUID NOT NULL REFERENCES branches(id),
    date DATE NOT NULL,
    is_working_day BOOLEAN NOT NULL DEFAULT true,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(staff_id, branch_id, date)
);
```

**Current Content:**
```
(0 rows) - Empty (no schedules created yet)
```

**Status:** ✅ Empty (expected - requires staff and branches to exist first)

---

### 9. `rotation_assignments` Table (0 rows)

**Schema:**
```sql
CREATE TABLE rotation_assignments (
    id UUID PRIMARY KEY,
    rotation_staff_id UUID NOT NULL REFERENCES staff(id),
    branch_id UUID NOT NULL REFERENCES branches(id),
    date DATE NOT NULL,
    assignment_level INTEGER NOT NULL CHECK (assignment_level IN (1, 2)),
    assigned_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(rotation_staff_id, branch_id, date)
);
```

**Current Content:**
```
(0 rows) - Empty (no rotation assignments made yet)
```

**Status:** ✅ Empty (expected - requires rotation staff and branches to exist first)

---

### 10. `system_settings` Table (0 rows)

**Schema:**
```sql
CREATE TABLE system_settings (
    id UUID PRIMARY KEY,
    key VARCHAR(100) UNIQUE NOT NULL,
    value TEXT NOT NULL,
    description TEXT,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**Current Content:**
```
(0 rows) - Empty (no system settings configured yet)
```

**Status:** ✅ Empty (expected - settings can be added as needed)

---

### 11. `staff_allocation_rules` Table (0 rows)

**Schema:**
```sql
CREATE TABLE staff_allocation_rules (
    id UUID PRIMARY KEY,
    position_id UUID NOT NULL REFERENCES positions(id),
    min_staff INTEGER NOT NULL DEFAULT 1,
    revenue_threshold DECIMAL(15,2) DEFAULT 0,
    staff_count_formula TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(position_id)
);
```

**Current Content:**
```
(0 rows) - Empty (no allocation rules configured yet)
```

**Status:** ✅ Empty (expected - rules can be added as needed)

---

## Summary

### ✅ Tables with Data:
1. **roles** - 5 rows (seeded)
2. **users** - 1 row (admin user)
3. **positions** - 7 rows (seeded)

### ⚠️ Critical Issue:
4. **branches** - **0 rows** (should have 35 standard branches)

### ✅ Empty Tables (Expected):
5. **staff** - 0 rows (needs manual entry)
6. **effective_branches** - 0 rows (requires staff + branches)
7. **revenue_data** - 0 rows (requires branches)
8. **staff_schedules** - 0 rows (requires staff + branches)
9. **rotation_assignments** - 0 rows (requires rotation staff + branches)
10. **system_settings** - 0 rows (optional configuration)
11. **staff_allocation_rules** - 0 rows (optional configuration)

---

## Action Required

### ⚠️ Fix Branches Table

The `branches` table is empty but should contain 35 standard branch codes. The seeding function `SeedStandardBranches()` needs to run.

**Option 1: Restart Backend Server**
```bash
# Restart the backend container
docker restart vsq-manpower-backend-dev
# or
docker-compose restart backend-dev
```

**Option 2: Manually Trigger Seeding**
The seeding function runs automatically when `RunMigrations()` is called. If you restart the backend server, it will run migrations (which includes seeding) on startup.

**Option 3: Manual SQL Insert (if needed)**
If seeding fails, you can manually insert branches using the standard codes from `backend/internal/constants/branches.go`.

**Verify After Fix:**
```bash
docker exec vsq-manpower-db psql -U vsq_user -d vsq_manpower -c "SELECT COUNT(*) FROM branches;"
# Should return: 35
```

---

## Database Health Check

**Overall Status:** ⚠️ **Partially Configured**

- ✅ Core tables created
- ✅ Default roles seeded
- ✅ Default positions seeded
- ✅ Admin user exists
- ⚠️ **Branches table empty** (critical issue)
- ✅ Other tables ready for data entry

**Next Steps:**
1. Fix branches table by restarting backend server
2. Verify 35 branches are created
3. Add staff members
4. Configure effective branches for rotation staff
5. Enter revenue data
6. Create staff schedules
7. Make rotation assignments

---

## Query Commands Reference

### List All Tables
```bash
docker exec vsq-manpower-db psql -U vsq_user -d vsq_manpower -c "\dt"
```

### Count Rows in All Tables
```bash
docker exec vsq-manpower-db psql -U vsq_user -d vsq_manpower -c "
SELECT 
    schemaname,
    tablename,
    (SELECT COUNT(*) FROM information_schema.columns WHERE table_name = tablename) as column_count
FROM pg_tables 
WHERE schemaname = 'public'
ORDER BY tablename;"
```

### View Branches (after seeding)
```bash
docker exec vsq-manpower-db psql -U vsq_user -d vsq_manpower -c "SELECT code, name FROM branches ORDER BY code;"
```

### Check Specific Table Structure
```bash
docker exec vsq-manpower-db psql -U vsq_user -d vsq_manpower -c "\d branches"
```





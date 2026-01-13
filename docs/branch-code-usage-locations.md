---
title: Branch Code Usage Locations
description: Comprehensive list of all places where branch codes are used in the application
version: 1.0.0
lastUpdated: 2025-12-21 00:35:04
---

# Branch Code Usage Locations

This document lists all locations in the application where branch codes are displayed, used, or referenced.

## Standard Branch Codes (35 total)
CPN, CPN-LS, CTR, PNK, CNK, BNA, CLP, SQR, BKP, CMC, CSA, EMQ, ESV, GTW, MGA, MTA, PRM, RCT, RST, TMA, MBA, SCN, CWG, CRM, CWT, PSO, RCP, CRA, CTW, ONE, DCP, MNG, TLR, TLR-LS, TLR-WN

---

## Frontend UI Components

### 1. Staff Scheduling Page
**File:** `frontend/src/app/(manager)/staff-scheduling/page.tsx`

**Location:** Lines 70-81

**Usage:**
- Branch selection dropdown for viewing staff schedules
- **Display Format:** `{branch.name} ({branch.code})`
- Example: "Central Park (CPN)" or "Central Park LS (CPN-LS)"

**User Role:** Admin, Area Manager, District Manager

**Purpose:** Allows managers to select a branch to view and manage staff schedules

---

### 2. Staff Management Page
**File:** `frontend/src/app/(manager)/staff-management/page.tsx`

**Locations:**
- **Line 253:** Branch name displayed in staff list table (shows branch name, not code directly)
- **Lines 350-361:** Branch selection dropdown in Add/Edit Staff modal
- **Display Format:** `{b.name} ({b.code})` in dropdown
- Example: "Central Park (CPN)"

**User Role:** Admin, Area Manager, District Manager

**Purpose:**
- Display which branch a branch staff member belongs to
- Allow selection of branch when creating/editing branch staff

---

### 3. Branch Management Page
**File:** `frontend/src/app/(manager)/branch-management/page.tsx`

**Locations:**
- **Line 186:** Branch code displayed as first column in branches table
- **Line 105:** Branch code used in edit form data
- **Display Format:** Branch code shown in dedicated column

**User Role:** Admin, Area Manager, District Manager

**Purpose:**
- Display all branches in a table with code as primary identifier
- Edit branch details (name, address, revenue, priority)
- View revenue history per branch

**Table Structure:**
| Code | Name | Address | Expected Revenue | Priority | Actions |
|------|------|---------|------------------|----------|---------|
| CPN  | Central Park | ... | ... | ... | Edit/Delete |

---

### 4. Rotation Assignment View
**File:** `frontend/src/components/rotation/RotationAssignmentView.tsx`

**Locations:**
- **Line 331-342:** Branch filter dropdown (when viewMode is 'by-branch')
- **Line 339:** Branch name displayed in filter (currently shows name only, not code)
- **Line 398:** Branch name displayed in calendar header
- **Lines 476, 486, 560:** Branch name displayed in assignment cells

**User Role:** Area Manager, District Manager

**Purpose:**
- Filter rotation assignments by branch
- Display branch names in calendar view
- Show which branch rotation staff are assigned to

**Note:** Currently displays branch name only. Could be enhanced to show `{branch.name} ({branch.code})` for better identification.

---

### 5. Rotation Suggestions Display
**File:** `frontend/src/components/rotation/RotationAssignmentView.tsx`

**Location:** Line 271

**Usage:**
- Displays AI suggestions for rotation staff assignments
- **Format:** `{staff.name} → {branch.name}`
- Example: "John Doe → Central Park"

**Note:** Could be enhanced to show branch code: `{staff.name} → {branch.name} ({branch.code})`

---

## Backend API & Data Layer

### 6. Branch API Endpoints
**File:** `frontend/src/lib/api/branch.ts`

**Branch Interface:**
```typescript
export interface Branch {
  id: string;
  name: string;
  code: string;  // ← Branch code field
  address: string;
  area_manager_id?: string;
  expected_revenue: number;
  priority: number;
  created_at: string;
  updated_at: string;
}
```

**Endpoints:**
- `GET /api/branches` - Returns list of all branches (includes codes)
- `POST /api/branches` - Creates new branch (requires code)
- `PUT /api/branches/:id` - Updates branch (code cannot be changed for standard branches)
- `DELETE /api/branches/:id` - Deletes branch (prevented for standard branches)
- `GET /api/branches/:id/revenue` - Gets revenue data for branch

---

### 7. Branch Handler (Backend)
**File:** `backend/internal/handlers/branch_handler.go`

**Functions:**
- `List()` - Returns all branches with codes
- `Create()` - Creates branch with code validation
- `Update()` - Updates branch (prevents code change for standard branches)
- `Delete()` - Deletes branch (prevents deletion of standard branches)
- `GetRevenue()` - Gets revenue data for branch

**Validation:**
- Uses `constants.IsStandardBranchCode()` to check if code is standard
- Prevents code modification for standard branches
- Prevents deletion of standard branches

---

### 8. Branch Repository (Backend)
**File:** `backend/internal/repositories/postgres/repositories.go`

**Functions:**
- `Create()` - Inserts branch with code (line 376)
- `GetByID()` - Retrieves branch by ID (includes code) (line 387)
- `Update()` - Updates branch including code (line 407)
- `Delete()` - Deletes branch by ID (line 413)
- `List()` - Returns all branches ordered by code (line 432)

**Database Queries:**
- All queries include `code` field in SELECT statements
- Code is stored in `branches.code` column (UNIQUE constraint)

---

### 9. Database Schema
**File:** `backend/internal/repositories/postgres/migrations.go`

**Table Definition:**
```sql
CREATE TABLE IF NOT EXISTS branches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50) UNIQUE NOT NULL,  -- ← Branch code column
    address TEXT,
    area_manager_id UUID REFERENCES users(id),
    expected_revenue DECIMAL(15,2) DEFAULT 0,
    priority INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**Constraints:**
- `code` is UNIQUE (prevents duplicate codes)
- `code` is NOT NULL (required field)

---

### 10. Database Seeding
**File:** `backend/internal/repositories/postgres/migrations.go`

**Function:** `SeedStandardBranches()` (lines 194-230)

**Purpose:**
- Automatically creates all 35 standard branch codes on system initialization
- Uses constants from `backend/internal/constants/branches.go`
- Generates deterministic UUIDs for each branch code

---

## Data Relationships

### 11. Staff-Branch Relationship
**File:** `backend/internal/repositories/postgres/repositories.go`

**Usage:**
- Staff table has `branch_id` foreign key to branches
- Branch staff are linked to branches via `branch_id`
- Branch code is accessible through JOIN queries

**Display:** Staff list shows branch name (derived from branch_id → branches.code)

---

### 12. Effective Branches
**File:** `backend/internal/repositories/postgres/repositories.go`

**Usage:**
- `effective_branches` table links rotation staff to branches
- Uses `branch_id` to reference branches (which have codes)
- Branch codes are accessible when querying effective branches

---

### 13. Staff Schedules
**File:** `backend/internal/repositories/postgres/repositories.go`

**Usage:**
- `staff_schedules` table has `branch_id` foreign key
- Schedules are branch-specific
- Branch code identifies which branch the schedule belongs to

---

### 14. Rotation Assignments
**File:** `backend/internal/repositories/postgres/repositories.go`

**Usage:**
- `rotation_assignments` table has `branch_id` foreign key
- Assignments link rotation staff to specific branches
- Branch code identifies assignment target branch

---

### 15. Revenue Data
**File:** `backend/internal/repositories/postgres/repositories.go`

**Usage:**
- `revenue_data` table has `branch_id` foreign key
- Revenue is tracked per branch
- Branch code identifies which branch the revenue belongs to

---

## Constants & Validation

### 16. Branch Constants
**File:** `backend/internal/constants/branches.go`

**Functions:**
- `StandardBranchCodes` - Array of 35 standard codes
- `IsStandardBranchCode(code)` - Checks if code is standard
- `GetStandardBranchCodes()` - Returns copy of standard codes list

**Usage:**
- Used in validation to prevent deletion/modification
- Used in seeding to create standard branches
- Referenced throughout application for standard branch checks

---

## Summary by Category

### Display Locations (User-Visible)
1. ✅ **Staff Scheduling** - Dropdown: `{name} ({code})`
2. ✅ **Staff Management** - Table column (name), Dropdown: `{name} ({code})`
3. ✅ **Branch Management** - Table column (code as primary identifier)
4. ⚠️ **Rotation Assignment** - Filter dropdown (name only - could show code)
5. ⚠️ **Rotation Suggestions** - Display (name only - could show code)

### Data Storage
6. ✅ **Database** - `branches.code` column (UNIQUE, NOT NULL)
7. ✅ **API Responses** - Branch objects include `code` field
8. ✅ **TypeScript Interfaces** - `Branch.code: string`

### Business Logic
9. ✅ **Validation** - Prevents standard branch code changes/deletion
10. ✅ **Seeding** - Auto-creates standard branches on startup
11. ✅ **Filtering** - Branch codes used in queries and filters

### Relationships
12. ✅ **Staff** - Linked via `branch_id` → `branches.code`
13. ✅ **Schedules** - Linked via `branch_id` → `branches.code`
14. ✅ **Assignments** - Linked via `branch_id` → `branches.code`
15. ✅ **Revenue** - Linked via `branch_id` → `branches.code`
16. ✅ **Effective Branches** - Linked via `branch_id` → `branches.code`

---

## Recommendations for Enhancement

### 1. Add Branch Code to Rotation Views
**Current:** Rotation Assignment View shows branch name only
**Recommendation:** Display as `{branch.name} ({branch.code})` for consistency

**Files to Update:**
- `frontend/src/components/rotation/RotationAssignmentView.tsx`
  - Line 339: Filter dropdown
  - Line 398: Calendar header
  - Lines 476, 486, 560: Assignment cells
  - Line 271: Suggestions display

### 2. Add Branch Code to Staff List
**Current:** Staff Management table shows branch name only
**Recommendation:** Display as `{branch.name} ({branch.code})` or add separate code column

**File to Update:**
- `frontend/src/app/(manager)/staff-management/page.tsx`
  - Line 253: Branch column in table

### 3. Export/Reporting
**Future:** When implementing reporting features, branch codes should be included in:
- Staff allocation reports
- Revenue reports
- Schedule exports
- CSV/Excel exports

---

## Related Requirements

- **FR-BM-01:** Branch Configuration
- **FR-BM-03:** Standard Branch Codes (must always be available)
- **FR-DV-02:** Data Integrity (unique constraint on branch code)





---
title: Branch Code Population Functions
description: List of functions and menus that populate with standard branch codes
version: 1.0.0
lastUpdated: 2025-12-21 00:28:16
---

# Branch Code Population Functions

This document lists all functions and menus in the system that should populate with the standard branch codes (FR-BM-03).

## Standard Branch Codes (35 total)
CPN, CPN-LS, CTR, PNK, CNK, BNA, CLP, SQR, BKP, CMC, CSA, EMQ, ESV, GTW, MGA, MTA, PRM, RCT, RST, TMA, MBA, SCN, CWG, CRM, CWT, PSO, RCP, CRA, CTW, ONE, DCP, MNG, TLR, TLR-LS, TLR-WN

---

## Frontend Functions/Menus

### 1. Staff Scheduling Page (`frontend/src/app/(manager)/staff-scheduling/page.tsx`)
- **Function:** `fetchData()` → `branchApi.list()`
- **Menu Component:** Branch selection dropdown
- **Location:** Lines 27-28, 70-81
- **Purpose:** Allows admin/area_manager/district_manager to select a branch for viewing staff schedules
- **Display Format:** `{branch.name} ({branch.code})`
- **Status:** ✅ Should populate with standard branch codes

### 2. Staff Management Page (`frontend/src/app/(manager)/staff-management/page.tsx`)
- **Function:** `fetchData()` → `branchApi.list()`
- **Menu Component:** Branch selection dropdown in Add/Edit Staff modal
- **Location:** Lines 42, 350-361
- **Purpose:** Allows selection of branch when creating/editing branch staff
- **Display Format:** `{branch.name} ({branch.code})`
- **Status:** ✅ Should populate with standard branch codes

### 3. Branch Management Page (`frontend/src/app/(manager)/branch-management/page.tsx`)
- **Function:** `loadBranches()` → `branchApi.list()`
- **Menu Component:** Branch list table
- **Location:** Lines 58-66, 184-201
- **Purpose:** Displays all branches in a table with edit/delete actions
- **Display Format:** Branch code column (line 186)
- **Status:** ✅ Should populate with standard branch codes

### 4. Rotation Assignment View (`frontend/src/components/rotation/RotationAssignmentView.tsx`)
- **Function:** `loadData()` → `branchApi.list()`
- **Menu Component:** Branch filter dropdown (when viewMode is 'by-branch')
- **Location:** Lines 46, 331-342
- **Purpose:** Filters rotation assignments by branch
- **Display Format:** `{branch.name}`
- **Status:** ✅ Should populate with standard branch codes

### 5. Effective Branch Management (Future Implementation)
- **Function:** TBD - Effective branch selection for rotation staff
- **Menu Component:** Multi-select or checkbox list for selecting effective branches
- **Location:** Not yet implemented
- **Purpose:** Allows managers to assign effective branches to rotation staff
- **Display Format:** TBD
- **Status:** ⚠️ Should populate with standard branch codes when implemented

---

## Backend Functions/Endpoints

### 1. Branch List API (`backend/internal/handlers/branch_handler.go`)
- **Function:** `BranchHandler.List()`
- **Endpoint:** `GET /api/branches`
- **Location:** Lines 31-38
- **Purpose:** Returns list of all branches
- **Used By:** All frontend components that need branch data
- **Status:** ✅ Should return standard branch codes

### 2. Dashboard Overview API (`backend/internal/handlers/dashboard_handler.go`)
- **Function:** `DashboardHandler.GetOverview()`
- **Endpoint:** `GET /api/dashboard/overview`
- **Location:** Lines 19-40
- **Purpose:** Returns summary statistics including total branches count
- **Uses:** `h.repos.Branch.List()` (line 21)
- **Status:** ✅ Should count standard branch codes

### 3. Rotation Suggestions API (`backend/internal/handlers/rotation_handler.go`)
- **Function:** `RotationHandler.GenerateSuggestions()`
- **Endpoint:** `POST /api/rotations/suggestions`
- **Location:** Lines 147-242
- **Purpose:** Generates rotation staff assignment suggestions for branches
- **Uses:** `h.repos.Branch.List()` or `h.repos.Branch.GetByID()` (lines 180, 187)
- **Status:** ✅ Should include standard branch codes in suggestions

### 4. Effective Branch Repository (`backend/internal/repositories/postgres/repositories.go`)
- **Function:** `effectiveBranchRepository.GetByBranchID()`
- **Location:** Lines 512-530
- **Purpose:** Gets rotation staff assigned to a specific branch
- **Uses:** Branch ID to query effective branches
- **Status:** ✅ Should work with standard branch codes

### 5. Staff Repository - Branch Filter (`backend/internal/repositories/postgres/repositories.go`)
- **Function:** `staffRepository.List()` with branch filter
- **Location:** Lines 248-249
- **Purpose:** Filters staff by branch ID
- **Uses:** Branch ID in WHERE clause
- **Status:** ✅ Should filter using standard branch codes

### 6. Schedule Repository - Branch Schedule (`backend/internal/repositories/postgres/repositories.go`)
- **Function:** `scheduleRepository.GetBranchSchedule()`
- **Location:** Lines 642-664
- **Purpose:** Gets staff schedules for a specific branch
- **Uses:** Branch ID parameter
- **Status:** ✅ Should retrieve schedules for standard branch codes

### 7. Revenue Repository - Branch Revenue (`backend/internal/repositories/postgres/repositories.go`)
- **Function:** `revenueRepository.GetByBranchID()`
- **Location:** Lines 562-589
- **Purpose:** Gets revenue data for a specific branch
- **Uses:** Branch ID parameter
- **Status:** ✅ Should retrieve revenue for standard branch codes

### 8. Rotation Assignment Repository - Branch Filter (`backend/internal/repositories/postgres/repositories.go`)
- **Function:** `rotationAssignmentRepository.List()` with branch filter
- **Location:** Lines 803-804
- **Purpose:** Filters rotation assignments by branch ID
- **Uses:** Branch ID in WHERE clause
- **Status:** ✅ Should filter using standard branch codes

---

## Database Functions

### 1. Branch List Query (`backend/internal/repositories/postgres/repositories.go`)
- **Function:** `branchRepository.List()`
- **SQL Query:** `SELECT ... FROM branches ORDER BY code`
- **Location:** Lines 418-442
- **Purpose:** Retrieves all branches ordered by code
- **Status:** ✅ Should return standard branch codes

### 2. Branch GetByID Query (`backend/internal/repositories/postgres/repositories.go`)
- **Function:** `branchRepository.GetByID()`
- **SQL Query:** `SELECT ... FROM branches WHERE id = $1`
- **Location:** Lines 381-402
- **Purpose:** Retrieves a specific branch by ID
- **Status:** ✅ Should work with standard branch codes

---

## Functions That Should Auto-Populate Standard Branches

### 1. Database Seeding Function (Implemented)
- **Function:** `SeedStandardBranches()` - Database seed function
- **Location:** `backend/internal/repositories/postgres/migrations.go` (lines 190-230)
- **Purpose:** Automatically create standard branch records on system initialization
- **Status:** ✅ Implemented - Creates all 35 standard branch codes using constants from `backend/internal/constants/branches.go`
- **Details:**
  - Called automatically after migrations run
  - Uses deterministic UUIDs generated from branch codes
  - Uses `ON CONFLICT (code) DO NOTHING` to prevent duplicates
  - Creates branches with default values (name=code, address="", revenue=0, priority=0)
  - Can be updated later through the Branch Management UI

### 2. Branch Validation Function (Already Implemented)
- **Function:** `constants.IsStandardBranchCode()` and validation in handlers
- **Location:** `backend/internal/constants/branches.go`, `backend/internal/handlers/branch_handler.go`
- **Purpose:** Prevents deletion and code modification of standard branches
- **Status:** ✅ Implemented

---

## Summary

### Functions Currently Using Branch Codes:
1. ✅ Staff Scheduling - Branch selection dropdown
2. ✅ Staff Management - Branch selection in staff form
3. ✅ Branch Management - Branch list table
4. ✅ Rotation Assignment - Branch filter dropdown
5. ✅ Dashboard - Branch count statistics
6. ✅ Rotation Suggestions - Branch-based suggestions
7. ✅ Effective Branch Management - Branch queries (backend)
8. ✅ Staff Filtering - Branch-based staff filtering
9. ✅ Schedule Retrieval - Branch-based schedule queries
10. ✅ Revenue Tracking - Branch-based revenue queries
11. ✅ Rotation Assignment Filtering - Branch-based assignment queries

### Functions That Need Implementation:
1. ✅ Database seeding function to auto-create standard branches - **IMPLEMENTED**
2. ⚠️ Effective Branch Management UI (frontend component for selecting effective branches)

---

## Recommendations

1. **Create Database Seeding Function:** Implement a function that automatically creates all 35 standard branch codes in the database when the system is initialized or when migrations run.

2. **Update Branch List API:** Ensure the branch list API always includes standard branches, even if they don't exist in the database yet (or create them on-the-fly).

3. **Add Branch Initialization:** Add a startup function that checks for standard branches and creates missing ones.

4. **Effective Branch UI:** When implementing the effective branch management UI, ensure it uses the standard branch codes list.

---

## Related Requirements

- **FR-BM-03:** Standard Branch Codes - These codes must always be available in the system
- **FR-BM-01:** Branch Configuration - System manages branches
- **FR-BL-02:** Effective Branch Management - Rotation staff effective branches


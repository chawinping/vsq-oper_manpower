---
title: Rotation Staff Management Implementation Alternatives
description: Two alternative approaches for implementing enhanced rotation staff management requirements
version: 1.0.0
lastUpdated: 2025-12-23 12:55:58
---

# Rotation Staff Management Implementation Alternatives

## Overview

This document outlines two alternative approaches for implementing the enhanced rotation staff management requirements (FR-RM-03, FR-SM-02, FR-SM-03). Both alternatives address the same functional requirements but differ in their architectural approach and data modeling strategy.

## Requirements Summary

1. **Permission Control:** Only Admin and Area Manager roles can view/add/edit/delete/import rotation staff
2. **Branch Assignment Workflow:** Rotation staff eligible for a branch are automatically populated in the branch table. Admin/Area Manager selects dates for assignment.
3. **Schedule Management:** Rotation staff schedule edit menu allows setting day off/leave/sick leave, or working day (with branch specification).

---

## Alternative 1: Unified Schedule Model (Recommended)

### Architecture Overview

This approach uses a **unified schedule model** where both branch staff and rotation staff schedules are managed through the same `StaffSchedule` table, with rotation staff assignments derived from schedule records where `schedule_status = 'working'` and `branch_id` is set.

### Key Design Decisions

1. **Single Source of Truth:** `StaffSchedule` table is the primary source for all staff schedules
2. **Rotation Assignments as Derived Data:** `RotationAssignment` table becomes a view/derived table or is populated automatically from `StaffSchedule` records
3. **Unified UI:** Same schedule editing interface for both branch and rotation staff
4. **Automatic Population:** Eligible rotation staff are determined by querying `EffectiveBranch` relationships

### Database Schema

```sql
-- Existing StaffSchedule table (enhanced)
CREATE TABLE staff_schedules (
    id UUID PRIMARY KEY,
    staff_id UUID NOT NULL REFERENCES staff(id),
    branch_id UUID NOT NULL REFERENCES branches(id), -- Required for rotation staff working days
    date DATE NOT NULL,
    schedule_status VARCHAR(20) NOT NULL, -- 'working', 'off', 'leave', 'sick_leave'
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(staff_id, branch_id, date) -- Prevents duplicate assignments
);

-- RotationAssignment becomes a materialized view or is auto-populated
CREATE TABLE rotation_assignments (
    id UUID PRIMARY KEY,
    rotation_staff_id UUID NOT NULL REFERENCES staff(id),
    branch_id UUID NOT NULL REFERENCES branches(id),
    date DATE NOT NULL,
    assignment_level INT NOT NULL, -- 1 or 2 (from EffectiveBranch)
    assigned_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(rotation_staff_id, branch_id, date)
);

-- Trigger or application logic to sync RotationAssignment from StaffSchedule
-- When schedule_status = 'working' for rotation staff, create/update RotationAssignment
-- When schedule_status != 'working', delete RotationAssignment if exists
```

### Implementation Flow

#### 1. Permission Enforcement

**Backend Middleware:**
```go
// middleware/rotation_permissions.go
func RequireRotationStaffPermission() gin.HandlerFunc {
    return func(c *gin.Context) {
        userRole := c.MustGet("user_role").(string)
        if userRole != "admin" && userRole != "area_manager" {
            c.JSON(http.StatusForbidden, gin.H{"error": "Only admin and area manager can manage rotation staff"})
            c.Abort()
            return
        }
        c.Next()
    }
}
```

**Handler Updates:**
- `staff_handler.go`: Add role check for rotation staff CRUD operations
- `rotation_handler.go`: Add role check for assignment operations

#### 2. Branch Assignment Workflow

**Frontend Component:**
```typescript
// components/scheduling/BranchStaffSchedule.tsx
// When branch is selected:
// 1. Fetch branch staff (existing)
// 2. Fetch eligible rotation staff:
//    - Query EffectiveBranch where branch_id = selectedBranch
//    - Get all rotation staff with Level 1 or Level 2 relationship
// 3. Display rotation staff as additional rows in the same table
// 4. For each rotation staff row, show date columns
// 5. When date is clicked, show branch selection dropdown (if working day)
```

**Backend API:**
```go
// GET /api/branches/:branchId/staff-schedule
// Returns: branch staff + eligible rotation staff
func (h *ScheduleHandler) GetBranchSchedule(c *gin.Context) {
    branchID := c.Param("branchId")
    
    // Get branch staff
    branchStaff := h.repos.Staff.GetByBranchID(branchID)
    
    // Get eligible rotation staff (from EffectiveBranch)
    eligibleRotationStaff := h.repos.Rotation.GetEligibleStaffForBranch(branchID)
    
    // Combine and return
    response := gin.H{
        "branch_staff": branchStaff,
        "eligible_rotation_staff": eligibleRotationStaff,
    }
    c.JSON(http.StatusOK, response)
}
```

#### 3. Schedule Management

**Unified Schedule Editor:**
- Same UI component for both branch and rotation staff
- For rotation staff: if `schedule_status = 'working'`, branch selection is required
- For branch staff: branch is pre-filled (their assigned branch)

**Backend Logic:**
```go
// POST /api/schedules
func (h *ScheduleHandler) CreateOrUpdateSchedule(c *gin.Context) {
    var req ScheduleRequest
    
    // Validate rotation staff working day requires branch
    if staff.StaffType == "rotation" && req.ScheduleStatus == "working" {
        if req.BranchID == nil {
            return error("Branch ID required for rotation staff working day")
        }
        // Validate effective branch relationship
        if !h.repos.Rotation.IsEffectiveBranch(req.StaffID, req.BranchID) {
            return error("Rotation staff not eligible for this branch")
        }
    }
    
    // Create/update StaffSchedule
    schedule := &models.StaffSchedule{
        StaffID: req.StaffID,
        BranchID: req.BranchID, // Required for rotation staff working days
        Date: req.Date,
        ScheduleStatus: req.ScheduleStatus,
    }
    
    h.repos.Schedule.CreateOrUpdate(schedule)
    
    // Auto-sync RotationAssignment if working day
    if schedule.ScheduleStatus == "working" && staff.StaffType == "rotation" {
        h.syncRotationAssignment(schedule)
    }
}
```

### Pros

✅ **Unified Data Model:** Single source of truth for all schedules  
✅ **Consistency:** Same logic for branch and rotation staff  
✅ **Simpler Queries:** One table to query for all schedule data  
✅ **Easier Reporting:** Unified schedule data simplifies reporting  
✅ **Less Code Duplication:** Shared components and logic  

### Cons

❌ **Migration Complexity:** Need to migrate existing `RotationAssignment` data  
❌ **Branch ID Required:** Must always specify branch for rotation staff working days  
❌ **Potential Confusion:** Branch ID field used differently for branch vs rotation staff  

---

## Alternative 2: Dual Model with Schedule Integration

### Architecture Overview

This approach maintains **separate models** for rotation assignments (`RotationAssignment`) and schedules (`StaffSchedule`), but integrates them so that rotation staff schedules can reference assignments, and assignments can be created from schedules.

### Key Design Decisions

1. **Dual Models:** `RotationAssignment` and `StaffSchedule` remain separate but linked
2. **Schedule-Driven Assignments:** Rotation staff schedules can create assignments when status is "working"
3. **Assignment-Driven Schedules:** Assignments can create schedule records
4. **Flexible Relationship:** Either model can be the source of truth depending on workflow

### Database Schema

```sql
-- StaffSchedule table (enhanced with rotation assignment link)
CREATE TABLE staff_schedules (
    id UUID PRIMARY KEY,
    staff_id UUID NOT NULL REFERENCES staff(id),
    branch_id UUID, -- NULL for rotation staff non-working days, required for branch staff
    date DATE NOT NULL,
    schedule_status VARCHAR(20) NOT NULL,
    rotation_assignment_id UUID REFERENCES rotation_assignments(id), -- Link to assignment if applicable
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(staff_id, date) -- One schedule per staff per day
);

-- RotationAssignment table (existing, enhanced)
CREATE TABLE rotation_assignments (
    id UUID PRIMARY KEY,
    rotation_staff_id UUID NOT NULL REFERENCES staff(id),
    branch_id UUID NOT NULL REFERENCES branches(id),
    date DATE NOT NULL,
    assignment_level INT NOT NULL,
    schedule_id UUID REFERENCES staff_schedules(id), -- Link to schedule if created from schedule
    assigned_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(rotation_staff_id, branch_id, date)
);

-- Check constraint: rotation staff working days must have branch_id
ALTER TABLE staff_schedules ADD CONSTRAINT check_rotation_working_branch 
    CHECK (
        (SELECT staff_type FROM staff WHERE id = staff_id) != 'rotation' 
        OR schedule_status != 'working' 
        OR branch_id IS NOT NULL
    );
```

### Implementation Flow

#### 1. Permission Enforcement

Same as Alternative 1 - middleware and handler checks.

#### 2. Branch Assignment Workflow

**Frontend Component:**
```typescript
// components/rotation/RotationStaffAssignment.tsx
// Separate component for rotation staff assignment
// When branch is selected:
// 1. Fetch branch staff (existing)
// 2. Fetch eligible rotation staff (from EffectiveBranch)
// 3. Display rotation staff in separate section or as expandable rows
// 4. For each rotation staff, show date picker/calendar
// 5. When date selected, create assignment via API
```

**Backend API:**
```go
// GET /api/branches/:branchId/eligible-rotation-staff
func (h *RotationHandler) GetEligibleStaff(c *gin.Context) {
    branchID := c.Param("branchId")
    eligibleStaff := h.repos.Rotation.GetEligibleStaffForBranch(branchID)
    c.JSON(http.StatusOK, gin.H{"eligible_staff": eligibleStaff})
}

// POST /api/rotation/assign
// Creates both RotationAssignment and StaffSchedule
func (h *RotationHandler) Assign(c *gin.Context) {
    var req AssignRotationRequest
    
    // Create RotationAssignment
    assignment := &models.RotationAssignment{
        RotationStaffID: req.RotationStaffID,
        BranchID: req.BranchID,
        Date: req.Date,
        AssignmentLevel: req.AssignmentLevel,
    }
    h.repos.Rotation.Create(assignment)
    
    // Create corresponding StaffSchedule
    schedule := &models.StaffSchedule{
        StaffID: req.RotationStaffID,
        BranchID: req.BranchID,
        Date: req.Date,
        ScheduleStatus: "working",
        RotationAssignmentID: &assignment.ID,
    }
    h.repos.Schedule.CreateOrUpdate(schedule)
    
    c.JSON(http.StatusCreated, gin.H{"assignment": assignment, "schedule": schedule})
}
```

#### 3. Schedule Management

**Dedicated Rotation Staff Schedule Editor:**
```typescript
// components/rotation/RotationStaffScheduleEditor.tsx
// Separate component for editing rotation staff schedules
// Shows calendar view for selected rotation staff
// For each day:
//   - If "working": Show branch dropdown (filtered by effective branches)
//   - If "off"/"leave"/"sick_leave": No branch selection needed
//   - Save creates/updates StaffSchedule and syncs RotationAssignment
```

**Backend Logic:**
```go
// POST /api/rotation/staff/:staffId/schedule
func (h *RotationHandler) UpdateSchedule(c *gin.Context) {
    staffID := c.Param("staffId")
    var req UpdateScheduleRequest
    
    // Validate working day requires branch
    if req.ScheduleStatus == "working" {
        if req.BranchID == nil {
            return error("Branch required for working day")
        }
        // Validate effective branch
        if !h.repos.Rotation.IsEffectiveBranch(staffID, req.BranchID) {
            return error("Not eligible for this branch")
        }
    }
    
    // Update or create schedule
    schedule := &models.StaffSchedule{
        StaffID: staffID,
        BranchID: req.BranchID, // NULL for non-working days
        Date: req.Date,
        ScheduleStatus: req.ScheduleStatus,
    }
    
    if req.ScheduleStatus == "working" {
        // Create or update assignment
        assignment := h.repos.Rotation.GetAssignment(staffID, req.BranchID, req.Date)
        if assignment == nil {
            assignment = &models.RotationAssignment{
                RotationStaffID: staffID,
                BranchID: req.BranchID,
                Date: req.Date,
                AssignmentLevel: h.repos.Rotation.GetAssignmentLevel(staffID, req.BranchID),
            }
            h.repos.Rotation.Create(assignment)
        }
        schedule.RotationAssignmentID = &assignment.ID
    } else {
        // Remove assignment if exists
        assignment := h.repos.Rotation.GetAssignment(staffID, nil, req.Date)
        if assignment != nil {
            h.repos.Rotation.Delete(assignment.ID)
        }
    }
    
    h.repos.Schedule.CreateOrUpdate(schedule)
}
```

### Pros

✅ **Clear Separation:** Distinct models for assignments vs schedules  
✅ **Flexible Workflow:** Can create assignments from schedules or vice versa  
✅ **Backward Compatible:** Existing `RotationAssignment` logic can remain  
✅ **Explicit Relationships:** Clear links between assignments and schedules  
✅ **Easier Migration:** Less disruptive to existing code  

### Cons

❌ **Data Synchronization:** Must keep two models in sync  
❌ **More Complex Queries:** Need to join tables for complete view  
❌ **Potential Inconsistencies:** Risk of assignment/schedule mismatch  
❌ **More Code:** Separate handlers and logic for each model  
❌ **Duplicate Logic:** Similar validation in multiple places  

---

## Comparison Matrix

| Aspect | Alternative 1: Unified Model | Alternative 2: Dual Model |
|--------|------------------------------|--------------------------|
| **Data Consistency** | ✅ Single source of truth | ⚠️ Requires synchronization |
| **Code Complexity** | ✅ Simpler, unified logic | ❌ More complex, dual logic |
| **Migration Effort** | ❌ Higher (need to migrate assignments) | ✅ Lower (keep existing) |
| **Query Performance** | ✅ Single table queries | ⚠️ Requires joins |
| **Flexibility** | ⚠️ Less flexible | ✅ More flexible workflows |
| **Maintainability** | ✅ Easier to maintain | ⚠️ More code to maintain |
| **Reporting** | ✅ Unified data model | ⚠️ Need to combine models |
| **UI Consistency** | ✅ Same interface for all staff | ⚠️ May need separate interfaces |

---

## Recommendation

**Alternative 1 (Unified Schedule Model)** is recommended because:

1. **Simpler Architecture:** Single source of truth reduces complexity
2. **Better Consistency:** Unified model prevents data inconsistencies
3. **Easier Maintenance:** Less code duplication and simpler queries
4. **Future-Proof:** Easier to extend with new schedule types or features
5. **Better UX:** Consistent interface for all staff types

**Migration Strategy for Alternative 1:**

1. Add `branch_id` to `StaffSchedule` (nullable)
2. Migrate existing `RotationAssignment` records to `StaffSchedule` with `schedule_status = 'working'`
3. Keep `RotationAssignment` as a materialized view or auto-populated table for backward compatibility
4. Update handlers to use unified model
5. Update frontend to use unified schedule editor

---

## Implementation Checklist

### Phase 1: Permission Enforcement
- [ ] Add role-based middleware for rotation staff operations
- [ ] Update staff handler to check roles for rotation staff CRUD
- [ ] Update rotation handler to check roles for assignment operations
- [ ] Add unit tests for permission checks

### Phase 2: Database Schema Updates
- [ ] Add `branch_id` to `StaffSchedule` (if Alternative 1)
- [ ] Add constraints for rotation staff working days
- [ ] Create migration scripts
- [ ] Migrate existing `RotationAssignment` data (if Alternative 1)

### Phase 3: Backend API Updates
- [ ] Create endpoint to get eligible rotation staff for a branch
- [ ] Update schedule creation/update logic
- [ ] Add validation for rotation staff working days
- [ ] Implement assignment sync logic (if Alternative 1)

### Phase 4: Frontend Updates
- [ ] Update branch schedule view to show eligible rotation staff
- [ ] Create rotation staff schedule editor component
- [ ] Add branch selection for rotation staff working days
- [ ] Update UI to reflect permission restrictions

### Phase 5: Testing
- [ ] Unit tests for permission checks
- [ ] Integration tests for assignment workflow
- [ ] E2E tests for schedule management
- [ ] Test data migration scripts

---

## Related Requirements

- **FR-RM-03:** Staff Data Management (rotation staff permissions)
- **FR-SM-02:** Rotation Staff Assignment (branch assignment workflow)
- **FR-SM-03:** Rotation Staff Schedule Management (schedule edit menu)
- **FR-AUZ-01:** Role-Based Access Control (permission enforcement)

---

## Notes

- Both alternatives require careful consideration of existing `RotationAssignment` data
- Permission checks must be enforced at both API and UI levels
- Effective branch relationships (`EffectiveBranch` table) are crucial for determining eligibility
- Schedule status validation is critical: working days must have branch, non-working days should not



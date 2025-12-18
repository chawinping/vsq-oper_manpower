---
title: Conflict Resolution Rules
description: Rules and procedures for handling scheduling and assignment conflicts
version: 1.0.0
lastUpdated: 2025-12-18 13:34:43
---

# Conflict Resolution Rules

## Document Information

- **Version:** 1.0.0
- **Last Updated:** 2025-12-18 13:34:43
- **Status:** Active
- **Related Documents:** SOFTWARE_REQUIREMENTS.md, docs/business-rules.md

---

## 1. Overview

This document defines the rules and procedures for detecting, preventing, and resolving conflicts in staff scheduling and rotation assignments within the VSQ Operations Manpower System.

---

## 2. Conflict Types

### 2.1 Rotation Staff Double-Booking Conflict

**Definition:** A rotation staff member is assigned to multiple branches on the same date.

**Conflict Detection:**
- System must check existing assignments before creating new assignment
- Conflict occurs when: `rotation_staff_id` + `date` combination already exists in `rotation_assignments` table

**Conflict Resolution Priority:**
1. **Level 1 (Priority) assignments take precedence** over Level 2 (Reserved) assignments
2. **Earlier assignment takes precedence** if both are same level
3. **Manual override** - Manager can explicitly replace existing assignment

### 2.2 Branch Staff Schedule Conflict

**Definition:** A branch staff member has conflicting schedule entries for the same date.

**Conflict Detection:**
- System must check existing schedules before creating new schedule
- Conflict occurs when: `staff_id` + `branch_id` + `date` combination already exists in `staff_schedules` table

**Conflict Resolution:**
- **Latest update takes precedence** - New schedule entry replaces existing entry
- System should log the change for audit purposes

### 2.3 Effective Branch Access Conflict

**Definition:** Attempt to assign rotation staff to a branch not in their effective branches list.

**Conflict Detection:**
- System must verify `branch_id` exists in `effective_branches` table for the rotation staff
- Conflict occurs when: `rotation_staff_id` + `branch_id` combination does not exist in `effective_branches`

**Conflict Resolution:**
- **Assignment is rejected** - Cannot assign rotation staff to non-effective branch
- System must return clear error message
- Manager must first add branch to effective branches list

### 2.4 Coverage Area Conflict

**Definition:** Attempt to assign rotation staff to branch outside their coverage area.

**Conflict Detection:**
- System must verify branch location matches rotation staff's coverage area
- Conflict occurs when: Branch location is outside rotation staff's `coverage_area`

**Conflict Resolution:**
- **Assignment is rejected** - Cannot assign rotation staff outside coverage area
- System must return clear error message
- Manager must update coverage area or effective branches list

### 2.5 Required Staff vs. Available Staff Conflict

**Definition:** Calculated required staff exceeds available staff (branch + rotation).

**Conflict Detection:**
- System calculates required staff based on revenue (FR-BL-01)
- System counts available staff (branch staff working + rotation staff assigned)
- Conflict occurs when: `required_staff > available_staff`

**Conflict Resolution:**
- **Warning is displayed** - System shows shortfall
- **Assignment is allowed** - Manager can proceed with understaffed branch
- **AI suggestions** - System suggests additional rotation staff assignments
- **Priority branches** - Level 1 branches get priority in suggestions

---

## 3. Business Rules for Conflict Resolution

### BR-CR-01: Rotation Staff Double-Booking Prevention
- **Description:** System shall prevent double-booking of rotation staff on the same date
- **Status:** ❌ Not Implemented (detection code commented out)
- **Rule:**
  - Before creating rotation assignment, system must check if rotation staff is already assigned on that date
  - If conflict exists, system must reject assignment with error message
  - Exception: Manual override by Area/District Manager (with confirmation)

**Implementation Requirements:**
- Check in `rotation_assignments` table: `WHERE rotation_staff_id = ? AND date = ?`
- Return error: "Rotation staff [name] is already assigned to [branch] on [date]"
- Allow override with explicit confirmation flag

### BR-CR-02: Level Priority Resolution
- **Description:** Level 1 assignments have priority over Level 2 assignments
- **Status:** ⚠️ Partially Implemented
- **Rule:**
  - When replacing Level 2 assignment with Level 1 assignment, system should allow
  - When replacing Level 1 assignment with Level 2 assignment, system should warn and require confirmation
  - When replacing Level 1 with Level 1, system should warn and require confirmation

**Implementation Requirements:**
- Compare assignment levels before replacement
- Show warning message for level downgrade
- Require explicit confirmation for level downgrade

### BR-CR-03: Effective Branch Validation
- **Description:** Rotation staff can only be assigned to effective branches
- **Status:** ✅ Implemented (in CheckAvailability)
- **Rule:**
  - System must verify branch is in effective branches list before assignment
  - Assignment to non-effective branch is rejected
  - Error message must indicate which branches are available

**Implementation Requirements:**
- Check in `effective_branches` table before assignment
- Return error: "Rotation staff [name] cannot be assigned to [branch]. Available branches: [list]"
- Provide link to manage effective branches

### BR-CR-04: Coverage Area Validation
- **Description:** Rotation staff assignments must respect coverage area constraints
- **Status:** ⚠️ Partially Implemented
- **Rule:**
  - System should validate coverage area when assigning rotation staff
  - Assignment outside coverage area should be rejected or require override
  - Coverage area is informational but effective branches list is authoritative

**Implementation Requirements:**
- Validate coverage area matches branch location (if location data available)
- Show warning if assignment is outside coverage area
- Allow override with manager confirmation

### BR-CR-05: Staff Shortfall Handling
- **Description:** System shall handle situations where required staff exceeds available staff
- **Status:** ⚠️ Partially Implemented
- **Rule:**
  - System calculates required staff based on revenue
  - System counts available staff (branch + rotation)
  - If shortfall exists, system shows warning
  - System suggests additional rotation staff assignments
  - Manager can proceed with understaffed branch (with acknowledgment)

**Implementation Requirements:**
- Calculate required staff per position per branch per date
- Count available staff (branch staff working + rotation staff assigned)
- Display shortfall warning: "Branch [name] requires [X] staff but only [Y] available on [date]"
- Provide suggestions for additional assignments
- Allow manager to acknowledge and proceed

### BR-CR-06: Schedule Update Precedence
- **Description:** Latest schedule update takes precedence over previous entries
- **Status:** ✅ Implemented (implicitly via database updates)
- **Rule:**
  - When branch manager updates schedule for same staff+date, new entry replaces old entry
  - System logs the change (created_by, updated_at)
  - No conflict error - update is allowed

**Implementation Requirements:**
- Use UPSERT (INSERT ... ON CONFLICT UPDATE) for schedules
- Log update with user ID and timestamp
- No error message needed - update is expected behavior

---

## 4. Conflict Detection Implementation

### 4.1 Rotation Assignment Conflict Detection

**Current Status:** ❌ Not Implemented (code commented out in `allocation.go`)

**Required Implementation:**

```go
// CheckRotationStaffAvailability checks if rotation staff is available
func (e *AllocationEngine) CheckRotationStaffAvailability(
    rotationStaffID uuid.UUID,
    date time.Time,
) (bool, *models.RotationAssignment, error) {
    // Check existing assignments for this rotation staff on this date
    assignments, err := e.repos.Rotation.GetByRotationStaffIDAndDate(
        rotationStaffID, 
        date,
    )
    if err != nil {
        return false, nil, err
    }
    
    if len(assignments) > 0 {
        return false, assignments[0], nil
    }
    
    return true, nil, nil
}
```

**Integration Points:**
- Call before creating rotation assignment in `RotationHandler.Assign()`
- Return conflict error if assignment exists
- Allow override with confirmation flag

### 4.2 Effective Branch Conflict Detection

**Current Status:** ✅ Implemented (in `CheckAvailability`)

**Required Enhancement:**
- Return list of available branches in error message
- Provide better error messaging

### 4.3 Staff Shortfall Detection

**Current Status:** ❌ Not Implemented

**Required Implementation:**

```go
// CheckStaffShortfall checks if branch has enough staff
func (e *AllocationEngine) CheckStaffShortfall(
    branchID uuid.UUID,
    date time.Time,
    positionID uuid.UUID,
) (*StaffShortfall, error) {
    // Calculate required staff
    expectedRevenue := // Get expected revenue for branch+date
    requiredStaff := e.CalculateRequiredStaff(
        branchID, 
        date, 
        positionID, 
        expectedRevenue,
    )
    
    // Count available staff
    branchStaffCount := // Count branch staff working on date
    rotationStaffCount := // Count rotation staff assigned on date
    availableStaff := branchStaffCount + rotationStaffCount
    
    shortfall := requiredStaff - availableStaff
    
    return &StaffShortfall{
        Required: requiredStaff,
        Available: availableStaff,
        Shortfall: shortfall,
        HasShortfall: shortfall > 0,
    }, nil
}
```

---

## 5. Conflict Resolution User Interface

### 5.1 Conflict Warning Messages

**Rotation Staff Double-Booking:**
```
⚠️ Conflict Detected

Rotation staff [Name] is already assigned to [Branch Name] on [Date].

Current Assignment:
- Branch: [Branch Name]
- Level: [Level 1/2]
- Assigned by: [Manager Name]

Options:
[ ] Replace existing assignment
[ ] Cancel assignment
```

**Effective Branch Conflict:**
```
❌ Assignment Not Allowed

Rotation staff [Name] cannot be assigned to [Branch Name].

Reason: Branch is not in rotation staff's effective branches list.

Available Branches:
- [Branch 1] (Level 1)
- [Branch 2] (Level 1)
- [Branch 3] (Level 2)

[Manage Effective Branches] [Cancel]
```

**Staff Shortfall Warning:**
```
⚠️ Staff Shortfall Detected

Branch [Branch Name] requires [X] [Position] staff on [Date], but only [Y] available.

Required: [X] staff
Available: [Y] staff (Branch: [A], Rotation: [B])
Shortfall: [X-Y] staff

[View Suggestions] [Acknowledge & Proceed] [Cancel]
```

### 5.2 Conflict Resolution Actions

**Allowed Actions:**
1. **Replace Assignment** - Replace existing assignment with new one
2. **Cancel Assignment** - Cancel the new assignment attempt
3. **Override** - Force assignment with manager confirmation
4. **View Suggestions** - See AI suggestions for resolving conflict
5. **Manage Effective Branches** - Add branch to effective branches list

---

## 6. Conflict Notification

### 6.1 Notification Requirements

#### BR-CR-07: Conflict Notification
- **Description:** System shall notify relevant managers of conflicts
- **Status:** ❌ Not Implemented
- **Rule:**
  - When conflict is detected, notify the manager who created the conflicting assignment
  - Notification includes: conflict type, affected staff, date, resolution action
  - Notification methods: In-app notification, email (optional)

**Implementation Requirements:**
- Create notification when assignment is replaced
- Create notification when assignment is overridden
- Notification to: original assigner, branch manager (if applicable)

---

## 7. Conflict Prevention

### 7.1 Proactive Conflict Prevention

#### BR-CR-08: Real-Time Conflict Checking
- **Description:** System shall check for conflicts in real-time during assignment
- **Status:** ❌ Not Implemented
- **Rule:**
  - Check conflicts before allowing assignment
  - Show conflicts immediately in UI
  - Prevent invalid assignments from being created

**Implementation Requirements:**
- Client-side conflict checking (for immediate feedback)
- Server-side conflict checking (for security)
- Real-time validation during assignment creation

#### BR-CR-09: Conflict Prevention in Bulk Operations
- **Description:** System shall prevent conflicts in bulk assignment operations
- **Status:** ❌ Not Implemented (bulk operations not implemented)
- **Rule:**
  - Validate all assignments in bulk operation
  - Report all conflicts before processing
  - Allow partial processing (process valid, skip invalid)
  - Provide conflict report after bulk operation

---

## 8. Conflict Reporting

### 8.1 Conflict Reports

#### BR-CR-10: Conflict History Report
- **Description:** System shall provide conflict history reports
- **Status:** ❌ Not Implemented
- **Rule:**
  - Report all conflicts detected in date range
  - Report all conflict resolutions
  - Report override actions
  - Report staff shortfalls

**Report Contents:**
- Conflict type
- Affected staff and branches
- Date and time
- Resolution action
- Resolved by (user)

---

## 9. Implementation Status

### Implemented ✅
- BR-CR-03: Effective Branch Validation (in CheckAvailability)
- BR-CR-06: Schedule Update Precedence (implicitly via database)

### Partially Implemented ⚠️
- BR-CR-02: Level Priority Resolution (logic exists, needs UI)
- BR-CR-04: Coverage Area Validation (validation exists, needs enhancement)
- BR-CR-05: Staff Shortfall Handling (calculation exists, needs UI)

### Not Implemented ❌
- BR-CR-01: Rotation Staff Double-Booking Prevention (code commented out)
- BR-CR-07: Conflict Notification
- BR-CR-08: Real-Time Conflict Checking
- BR-CR-09: Conflict Prevention in Bulk Operations
- BR-CR-10: Conflict History Report

---

## 10. Priority Implementation Plan

### Phase 1: Critical Conflicts (High Priority)
1. Implement BR-CR-01: Rotation Staff Double-Booking Prevention
2. Enhance BR-CR-03: Effective Branch Validation (better error messages)
3. Implement BR-CR-08: Real-Time Conflict Checking

### Phase 2: Conflict Resolution (Medium Priority)
4. Implement BR-CR-02: Level Priority Resolution (complete)
5. Implement BR-CR-05: Staff Shortfall Handling (complete)
6. Implement BR-CR-07: Conflict Notification

### Phase 3: Advanced Features (Low Priority)
7. Implement BR-CR-09: Conflict Prevention in Bulk Operations
8. Implement BR-CR-10: Conflict History Report

---

## 11. Related Requirements

- **SOFTWARE_REQUIREMENTS.md:**
  - FR-BL-03: Availability Checking
  - FR-SM-02: Rotation Staff Assignment
- **Security Requirements:** Conflict resolution security implications
- **Business Rules:** BR-BL-02, BR-BL-04

---

## 12. Change Log

| Date | Version | Changes | Author |
|------|---------|---------|--------|
| 2025-12-18 | 1.0.0 | Initial conflict resolution rules document created | System |


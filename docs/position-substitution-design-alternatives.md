# Rotation Staff to Branch Position Mapping - Design Alternatives

## Overview
This document presents two alternative designs for a submenu under "Staff Groups" that allows administrators to configure which branch positions each rotation staff member can work in. Since some rotation staff can work in place of more than one branch position, this mapping enables flexible allocation suggestions.

## Context
Currently, the allocation suggestion engine matches rotation staff to branch positions based on exact `PositionID` match (rotation staff's position must match the branch position requirement). This feature will enable flexible substitution rules, allowing individual rotation staff members to be assigned to multiple branch positions they're qualified to fill, even if their primary rotation position doesn't match exactly.

**Key Requirement:** A rotation staff member can be mapped to multiple branch positions (one-to-many relationship).

## Design Alternative 1: Direct Staff-to-Position Mapping

### Concept
A simple many-to-many mapping table where each entry specifies: "Rotation Staff Member X can work in Branch Position Y". Each rotation staff member can have multiple entries, allowing them to fill multiple branch positions.

### Data Model
```go
// RotationStaffBranchPosition represents a mapping allowing a rotation staff member to work in a branch position
type RotationStaffBranchPosition struct {
    ID                  uuid.UUID   `json:"id" db:"id"`
    RotationStaffID     uuid.UUID   `json:"rotation_staff_id" db:"rotation_staff_id"`
    BranchPositionID    uuid.UUID   `json:"branch_position_id" db:"branch_position_id"`
    RotationStaff       *Staff      `json:"rotation_staff,omitempty"`
    BranchPosition      *Position   `json:"branch_position,omitempty"`
    SubstitutionLevel   int         `json:"substitution_level" db:"substitution_level"` // 1 = preferred, 2 = acceptable, 3 = emergency only
    IsActive            bool        `json:"is_active" db:"is_active"`
    Notes               string      `json:"notes,omitempty" db:"notes"`
    CreatedAt           time.Time   `json:"created_at" db:"created_at"`
    UpdatedAt           time.Time   `json:"updated_at" db:"updated_at"`
}
```

### UI Design
**Menu Structure:**
- Allocation Logic
  - Staff Groups
    - **Rotation Staff Position Mapping** (new submenu)

**Page Layout - Staff-Centric View:**
```
┌─────────────────────────────────────────────────────────────┐
│ Rotation Staff Position Mapping                            │
│                                                             │
│ Filter: [All Rotation Staff ▼] [Search: ________]         │
│                                                             │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ Rotation Staff        │ Can Fill Positions             │ │
│ ├─────────────────────────────────────────────────────────┤ │
│ │ John Doe (Nurse)      │ • Nurse (Preferred)            │ │
│ │                       │ • Junior Nurse (Acceptable)    │ │
│ │                       │ • Receptionist (Emergency)     │ │
│ │                       │ [+ Add Position]               │ │
│ ├─────────────────────────────────────────────────────────┤ │
│ │ Jane Smith (Senior)   │ • Senior Nurse (Preferred)     │ │
│ │                       │ • Nurse (Preferred)            │ │
│ │                       │ [+ Add Position]               │ │
│ └─────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

**Alternative Page Layout - Position-Centric View (Toggle):**
```
┌─────────────────────────────────────────────────────────────┐
│ Rotation Staff Position Mapping                            │
│                                                             │
│ View: [Staff View] [Position View]                         │
│                                                             │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ Branch Position    │ Eligible Rotation Staff            │ │
│ ├─────────────────────────────────────────────────────────┤ │
│ │ Nurse             │ • John Doe (Preferred)              │ │
│ │                   │ • Jane Smith (Preferred)            │ │
│ │                   │ • Bob Wilson (Acceptable)          │ │
│ │                   │ [+ Add Staff]                      │ │
│ ├─────────────────────────────────────────────────────────┤ │
│ │ Junior Nurse      │ • John Doe (Acceptable)            │ │
│ │                   │ • Alice Brown (Preferred)          │ │
│ │                   │ [+ Add Staff]                      │ │
│ └─────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

**Form Fields (Modal/Dialog):**
When adding/editing a mapping:
- Rotation Staff (dropdown - filtered to rotation staff only)
- Branch Position (dropdown - filtered to branch positions only)
- Substitution Level (dropdown: Preferred, Acceptable, Emergency Only)
- Active (checkbox)
- Notes (textarea, optional)

### Advantages
- ✅ Simple and intuitive - direct staff-to-position mapping
- ✅ Flexible - each staff member can have unique position mappings
- ✅ Fast lookups in allocation logic (query by staff ID)
- ✅ Easy to understand and maintain
- ✅ Supports individual staff capabilities
- ✅ Can show both staff-centric and position-centric views

### Disadvantages
- ❌ Requires manual entry for each staff-position pair
- ❌ May become tedious with many staff members
- ❌ No bulk operations (can't apply to multiple staff at once)
- ❌ Doesn't leverage position-level defaults

### Implementation Complexity
- **Backend:** Medium (new model, repository, handler)
- **Frontend:** Medium (CRUD page with dual view modes, staff/position filtering)
- **Database:** Low (single table with foreign keys, unique constraint on staff+position)

---

## Design Alternative 2: Position-Based Defaults with Staff Overrides

### Concept
Define default mappings at the position level ("Rotation Position X can fill Branch Position Y"), then allow individual staff members to have additional position mappings beyond their primary position's defaults. This combines position-level rules with staff-specific overrides.

### Data Model
```go
// PositionSubstitution represents a default mapping at position level
type PositionSubstitution struct {
    ID                  uuid.UUID   `json:"id" db:"id"`
    RotationPositionID  uuid.UUID   `json:"rotation_position_id" db:"rotation_position_id"`
    BranchPositionID    uuid.UUID   `json:"branch_position_id" db:"branch_position_id"`
    RotationPosition    *Position   `json:"rotation_position,omitempty"`
    BranchPosition      *Position   `json:"branch_position,omitempty"`
    SubstitutionLevel   int         `json:"substitution_level" db:"substitution_level"` // 1-3
    IsActive            bool        `json:"is_active" db:"is_active"`
    CreatedAt           time.Time   `json:"created_at" db:"created_at"`
    UpdatedAt           time.Time   `json:"updated_at" db:"updated_at"`
}

// StaffPositionOverride represents staff-specific position mappings beyond their primary position
type StaffPositionOverride struct {
    ID                  uuid.UUID   `json:"id" db:"id"`
    RotationStaffID     uuid.UUID   `json:"rotation_staff_id" db:"rotation_staff_id"`
    BranchPositionID    uuid.UUID   `json:"branch_position_id" db:"branch_position_id"`
    RotationStaff       *Staff      `json:"rotation_staff,omitempty"`
    BranchPosition      *Position   `json:"branch_position,omitempty"`
    SubstitutionLevel   int         `json:"substitution_level" db:"substitution_level"` // 1-3
    IsActive            bool        `json:"is_active" db:"is_active"`
    Notes               string      `json:"notes,omitempty" db:"notes"`
    CreatedAt           time.Time   `json:"created_at" db:"created_at"`
    UpdatedAt           time.Time   `json:"updated_at" db:"updated_at"`
}
```

### UI Design
**Menu Structure:**
- Allocation Logic
  - Staff Groups
    - **Position Substitutions** (position-level defaults)
    - **Staff Position Overrides** (staff-specific mappings)

**Page Layout - Position Substitutions (Tab 1):**
```
┌─────────────────────────────────────────────────────────────┐
│ Position Substitutions (Defaults)                          │
│                                                             │
│ [+ Add Position Substitution]                              │
│                                                             │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ Rotation Position  │ Branch Position │ Level │ Status  │ │
│ ├─────────────────────────────────────────────────────────┤ │
│ │ Senior Nurse      │ Nurse          │ 1     │ Active  │ │
│ │ Nurse             │ Junior Nurse    │ 2     │ Active  │ │
│ └─────────────────────────────────────────────────────────┘ │
│                                                             │
│ Note: These defaults apply to all rotation staff with     │
│       the specified rotation position.                    │
└─────────────────────────────────────────────────────────────┘
```

**Page Layout - Staff Overrides (Tab 2):**
```
┌─────────────────────────────────────────────────────────────┐
│ Staff Position Overrides                                    │
│                                                             │
│ Filter: [All Rotation Staff ▼] [Search: ________]          │
│                                                             │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ Rotation Staff        │ Additional Positions             │ │
│ ├─────────────────────────────────────────────────────────┤ │
│ │ John Doe (Nurse)      │ • Receptionist (Preferred)     │ │
│ │                       │   (beyond Nurse defaults)      │ │
│ │                       │ [+ Add Override]               │ │
│ ├─────────────────────────────────────────────────────────┤ │
│ │ Jane Smith (Senior)   │ • Admin Assistant (Acceptable) │ │
│ │                       │   (beyond Senior Nurse defaults)│ │
│ │                       │ [+ Add Override]               │ │
│ └─────────────────────────────────────────────────────────┘ │
│                                                             │
│ Note: These are additional positions beyond what's         │
│       defined by their primary position's defaults.        │
└─────────────────────────────────────────────────────────────┘
```

**Combined View (Tab 3 - "Staff Capabilities"):**
```
┌─────────────────────────────────────────────────────────────┐
│ Staff Capabilities Overview                                │
│                                                             │
│ Filter: [All Rotation Staff ▼]                            │
│                                                             │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ Staff            │ Primary │ Can Fill (from defaults)   │ │
│ │                 │         │ + Overrides                 │ │
│ ├─────────────────────────────────────────────────────────┤ │
│ │ John Doe        │ Nurse   │ ✓ Nurse (from position)    │ │
│ │                 │         │ ✓ Junior Nurse (from pos)   │ │
│ │                 │         │ ✓ Receptionist (override)   │ │
│ ├─────────────────────────────────────────────────────────┤ │
│ │ Jane Smith      │ Senior  │ ✓ Senior Nurse (from pos)  │ │
│ │                 │ Nurse   │ ✓ Nurse (from position)    │ │
│ │                 │         │ ✓ Admin Assistant (override)│ │
│ └─────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

### Advantages
- ✅ Scalable - define defaults once per position, applies to all staff
- ✅ Flexible - allows individual staff to have additional capabilities
- ✅ Efficient - fewer records needed (position defaults + individual overrides)
- ✅ Better for bulk management - update position defaults affects all staff
- ✅ Clear separation between defaults and exceptions
- ✅ Can show combined view of all capabilities

### Disadvantages
- ❌ More complex data model (two tables)
- ❌ More complex UI (two tabs/pages to manage)
- ❌ Requires understanding of position defaults vs overrides
- ❌ Logic needs to combine defaults + overrides when querying
- ❌ Potential confusion about which mappings come from where

### Implementation Complexity
- **Backend:** High (two models, repository methods to combine defaults+overrides, validation logic)
- **Frontend:** Medium-High (multiple tabs, combined view logic, showing source of mappings)
- **Database:** Medium (two tables with foreign keys, need efficient queries combining both)

---

## Comparison Matrix

| Aspect | Alternative 1: Direct Staff-to-Position | Alternative 2: Position Defaults + Overrides |
|--------|-----------------------------------------|----------------------------------------------|
| **Simplicity** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ |
| **Scalability** | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| **Maintainability** | ⭐⭐⭐⭐ | ⭐⭐⭐ |
| **User Experience** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ |
| **Implementation Effort** | Low-Medium | Medium-High |
| **Flexibility** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| **Bulk Operations** | ⭐⭐ | ⭐⭐⭐⭐⭐ |
| **Query Performance** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ |

## Recommendation

**For MVP/Initial Implementation:** Choose **Alternative 1 (Direct Staff-to-Position Mapping)**

**Rationale:**
1. Simpler to understand - direct mapping is intuitive
2. Faster to implement and test
3. Easier for users to configure - no need to understand defaults vs overrides
4. More flexible for individual staff capabilities
5. Lower risk of bugs due to simpler model
6. Can be enhanced later to support position defaults if patterns emerge

**For Future Enhancement:** Consider **Alternative 2** if:
- The number of rotation staff grows significantly (>50 staff)
- Clear patterns emerge where most staff in a position can fill the same branch positions
- Users request bulk operations or position-level defaults
- Maintenance becomes tedious with many individual mappings

## Integration with Allocation Logic

### Current Flow (suggestion_engine.go:144)
```go
// Check if staff matches the position
if staff.PositionID != positionID {
    continue
}
```

### Proposed Enhancement (Alternative 1)
```go
// Check if staff matches the position directly OR via mapping
matchesPosition := staff.PositionID == positionID
if !matchesPosition {
    // Check for staff-to-position mapping
    mapping, err := e.repos.RotationStaffBranchPosition.GetByStaffAndPosition(
        staff.ID, 
        positionID,
    )
    if err == nil && mapping != nil && mapping.IsActive {
        matchesPosition = true
        // Adjust confidence/priority based on substitution level
        // Level 1 (preferred): no penalty
        // Level 2 (acceptable): -0.1 penalty
        // Level 3 (emergency): -0.3 penalty
        priorityAdjustment := 0.0
        switch mapping.SubstitutionLevel {
        case 2:
            priorityAdjustment = -0.1
        case 3:
            priorityAdjustment = -0.3
        }
        // Apply adjustment to priority score
    }
}
if !matchesPosition {
    continue
}
```

### Proposed Enhancement (Alternative 2)
```go
// Check if staff matches the position directly OR via mapping
matchesPosition := staff.PositionID == positionID
if !matchesPosition {
    // First check position-level defaults
    positionSub, err := e.repos.PositionSubstitution.GetByRotationAndBranchPosition(
        staff.PositionID,
        positionID,
    )
    if err == nil && positionSub != nil && positionSub.IsActive {
        matchesPosition = true
        // Use position substitution level
    } else {
        // Check staff-specific override
        override, err := e.repos.StaffPositionOverride.GetByStaffAndPosition(
            staff.ID,
            positionID,
        )
        if err == nil && override != nil && override.IsActive {
            matchesPosition = true
            // Use override substitution level
        }
    }
    
    if matchesPosition {
        // Apply priority adjustment based on substitution level
        // (same as Alternative 1)
    }
}
if !matchesPosition {
    continue
}
```

## Database Schema

### Alternative 1: Direct Staff-to-Position Mapping
```sql
CREATE TABLE rotation_staff_branch_positions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rotation_staff_id UUID NOT NULL REFERENCES staff(id) ON DELETE CASCADE,
    branch_position_id UUID NOT NULL REFERENCES positions(id),
    substitution_level INTEGER NOT NULL DEFAULT 2 CHECK (substitution_level BETWEEN 1 AND 3),
    is_active BOOLEAN NOT NULL DEFAULT true,
    notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(rotation_staff_id, branch_position_id)
);

CREATE INDEX idx_rotation_staff_branch_positions_staff ON rotation_staff_branch_positions(rotation_staff_id) WHERE is_active = true;
CREATE INDEX idx_rotation_staff_branch_positions_position ON rotation_staff_branch_positions(branch_position_id) WHERE is_active = true;
```

### Alternative 2: Position Defaults + Staff Overrides
```sql
-- Position-level defaults
CREATE TABLE position_substitutions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rotation_position_id UUID NOT NULL REFERENCES positions(id),
    branch_position_id UUID NOT NULL REFERENCES positions(id),
    substitution_level INTEGER NOT NULL DEFAULT 2 CHECK (substitution_level BETWEEN 1 AND 3),
    is_active BOOLEAN NOT NULL DEFAULT true,
    notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(rotation_position_id, branch_position_id)
);

-- Staff-specific overrides
CREATE TABLE staff_position_overrides (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rotation_staff_id UUID NOT NULL REFERENCES staff(id) ON DELETE CASCADE,
    branch_position_id UUID NOT NULL REFERENCES positions(id),
    substitution_level INTEGER NOT NULL DEFAULT 2 CHECK (substitution_level BETWEEN 1 AND 3),
    is_active BOOLEAN NOT NULL DEFAULT true,
    notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(rotation_staff_id, branch_position_id)
);

CREATE INDEX idx_position_substitutions_rotation ON position_substitutions(rotation_position_id) WHERE is_active = true;
CREATE INDEX idx_position_substitutions_branch ON position_substitutions(branch_position_id) WHERE is_active = true;
CREATE INDEX idx_staff_overrides_staff ON staff_position_overrides(rotation_staff_id) WHERE is_active = true;
CREATE INDEX idx_staff_overrides_position ON staff_position_overrides(branch_position_id) WHERE is_active = true;
```

## Next Steps (if Alternative 1 is chosen)

1. **Backend:**
   - Create `RotationStaffBranchPosition` model
   - Create repository interface and PostgreSQL implementation
   - Create handler with CRUD endpoints
   - Update `suggestion_engine.go` to use staff-position mappings
   - Add unit tests

2. **Frontend:**
   - Create `/staff-groups/rotation-staff-position-mapping` page
   - Add submenu item in `AppLayout.tsx`
   - Create API client functions
   - Implement CRUD UI with staff-centric and position-centric views
   - Add filtering and search functionality

3. **Database:**
   - Create migration for `rotation_staff_branch_positions` table
   - Add seed data if needed

4. **Testing:**
   - Test allocation suggestions with staff-position mappings
   - Verify substitution level affects priority scores correctly
   - Test edge cases (staff with no mappings, multiple mappings, etc.)

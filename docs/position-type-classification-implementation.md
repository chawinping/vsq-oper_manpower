---
title: Position Type Classification Implementation Plan
description: Implementation plan for classifying positions into "branch" and "rotation" types and updating branch configuration UI
version: 1.0.0
lastUpdated: 2026-01-13 19:30:00
---

# Position Type Classification Implementation Plan

## Overview

This document outlines the implementation plan for classifying positions into two types ("branch" and "rotation") and updating the branch configuration UI to only show minimum/preferred quota settings for non-rotation positions.

## Current State Analysis

### Current Implementation

1. **Position Model** (`backend/internal/domain/models/staff.go`):
   - Positions currently have: `id`, `name`, `min_staff_per_branch`, `display_order`
   - No `position_type` field exists

2. **Branch Configuration UI** (`frontend/src/components/branch/BranchPositionQuotaConfig.tsx`):
   - Shows all target positions with "Required Staff" (designated_quota) and "Minimum Required" (minimum_required)
   - Currently displays positions like "ฟร้อนท์วนสาขา" (Front Rotation) which should be excluded

3. **Database Schema**:
   - `positions` table does not have a `position_type` column
   - Need to add migration to add this field

4. **Position Quota Model** (`backend/internal/domain/models/position_quota.go`):
   - Has `DesignatedQuota` (preferred/target) and `MinimumRequired` (minimum)
   - Backend validation exists but doesn't check position type

## Requirements

1. **Position Classification**:
   - Each position must be classified as either "branch" or "rotation"
   - Classification should be stored in the database

2. **Branch Configuration UI Changes**:
   - Only show positions with `position_type = 'branch'` in the quota configuration
   - Hide positions with `position_type = 'rotation'`
   - Rename labels:
     - "Required Staff" → "Preferred"
     - "Minimum Required" → "Minimum"

3. **Backend Validation**:
   - Only allow quota updates for branch-type positions
   - Reject quota updates for rotation-type positions

## Implementation Plan

### Phase 1: Database Schema Changes

#### 1.1 Add `position_type` Column to Positions Table

**File:** `backend/internal/repositories/postgres/migrations.go`

**Action:** Add migration to:
- Add `position_type VARCHAR(20) NOT NULL DEFAULT 'branch'` column
- Add CHECK constraint: `position_type IN ('branch', 'rotation')`
- Update existing positions:
  - Positions with "rotation" or "วนสาขา" in name → set to 'rotation'
  - All others → set to 'branch'

**SQL Migration:**
```sql
ALTER TABLE positions 
ADD COLUMN position_type VARCHAR(20) NOT NULL DEFAULT 'branch' 
CHECK (position_type IN ('branch', 'rotation'));

-- Update existing rotation positions
UPDATE positions 
SET position_type = 'rotation' 
WHERE name LIKE '%วนสาขา%' OR name LIKE '%Rotation%';

-- Ensure default constraint
ALTER TABLE positions 
ALTER COLUMN position_type SET DEFAULT 'branch';
```

### Phase 2: Backend Changes

#### 2.1 Update Position Model

**File:** `backend/internal/domain/models/staff.go`

**Changes:**
- Add `PositionType` field to `Position` struct
- Add constants for position types

```go
type PositionType string

const (
    PositionTypeBranch   PositionType = "branch"
    PositionTypeRotation PositionType = "rotation"
)

type Position struct {
    ID                  uuid.UUID    `json:"id" db:"id"`
    Name                string       `json:"name" db:"name"`
    PositionType        PositionType `json:"position_type" db:"position_type"`
    MinStaffPerBranch   int          `json:"min_staff_per_branch,omitempty" db:"min_staff_per_branch"`
    DisplayOrder        int          `json:"display_order" db:"display_order"`
    BranchStaffCount    *int         `json:"branch_staff_count,omitempty" db:"-"`
    RotationStaffCount  *int         `json:"rotation_staff_count,omitempty" db:"-"`
    CreatedAt           time.Time    `json:"created_at" db:"created_at"`
}
```

#### 2.2 Update Position Repository

**File:** `backend/internal/repositories/postgres/repositories.go`

**Changes:**
- Update SELECT queries to include `position_type`
- Update INSERT queries to include `position_type`
- Update UPDATE queries to include `position_type`

#### 2.3 Update Position Handler

**File:** `backend/internal/handlers/position_handler.go`

**Changes:**
- Include `position_type` in API responses
- Add validation for `position_type` in update/create requests

#### 2.4 Update Branch Config Handler Validation

**File:** `backend/internal/handlers/branch_config_handler.go`

**Changes:**
- Add validation in `UpdateQuotas` to check position type
- Reject quota updates for rotation-type positions
- Return appropriate error message

**Validation Logic:**
```go
// Before updating quotas, check position types
positions, err := h.repos.Position.List()
if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
    return
}

positionMap := make(map[uuid.UUID]*models.Position)
for _, pos := range positions {
    positionMap[pos.ID] = pos
}

// Validate that all positions are branch-type
for _, quota := range req.Quotas {
    position := positionMap[quota.PositionID]
    if position == nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Position not found"})
        return
    }
    if position.PositionType == models.PositionTypeRotation {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": fmt.Sprintf("Cannot set quotas for rotation-type position: %s", position.Name),
        })
        return
    }
}
```

### Phase 3: Frontend Changes

#### 3.1 Update Position API Interface

**File:** `frontend/src/lib/api/position.ts`

**Changes:**
- Add `position_type` field to `Position` interface

```typescript
export interface Position {
  id: string;
  name: string;
  position_type: 'branch' | 'rotation';
  display_order: number;
  branch_staff_count?: number;
  rotation_staff_count?: number;
  created_at: string;
}
```

#### 3.2 Update BranchPositionQuotaConfig Component

**File:** `frontend/src/components/branch/BranchPositionQuotaConfig.tsx`

**Changes:**

1. **Filter out rotation positions:**
```typescript
// Filter positions to show only branch-type positions
const filteredPositions = positions.filter((pos) => {
  // First check if it matches target positions
  const matchesTarget = targetPositions.some((target) =>
    pos.name.toLowerCase().includes(target.toLowerCase()) ||
    target.toLowerCase().includes(pos.name.toLowerCase())
  );
  
  // Then check if it's a branch-type position
  return matchesTarget && pos.position_type === 'branch';
});
```

2. **Update UI labels:**
   - Change "Required Staff" → "Preferred"
   - Change "Minimum Required" → "Minimum"

3. **Update table headers:**
```typescript
<th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
  Preferred
</th>
<th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
  Minimum
</th>
```

### Phase 4: Data Migration

#### 4.1 Identify Rotation Positions

Based on `docs/positions-inventory.md` and business requirements, the following positions should be classified as "rotation":
- `10000000-0000-0000-0000-000000000019` - Front+ล่ามวนสาขา (Front + Interpreter Rotation)
- `10000000-0000-0000-0000-000000000022` - ผู้จัดการเขต (District Manager)
- `10000000-0000-0000-0000-000000000023` - ผู้จัดการแผนกและกำกับพัฒนาระเบียบสาขา (Department Manager & Branch Development Supervisor)
- `10000000-0000-0000-0000-000000000024` - หัวหน้าผู้ช่วยแพทย์ (Head Doctor Assistant)
- `10000000-0000-0000-0000-000000000025` - ผู้ช่วยพิเศษ (Special Assistant)
- `10000000-0000-0000-0000-000000000026` - ผู้ช่วยแพทย์วนสาขา (Doctor Assistant Rotation)
- `10000000-0000-0000-0000-000000000027` - ฟร้อนท์วนสาขา (Front Rotation)

**Note:** The migration should use pattern matching to identify rotation positions:
- Names containing "วนสาขา" (Thai for "rotation")
- Names containing "Rotation" (English)

#### 4.2 Migration Script

Create a migration script or include in the main migration:

```sql
-- Update rotation positions based on name patterns
UPDATE positions 
SET position_type = 'rotation' 
WHERE name LIKE '%วนสาขา%' 
   OR name LIKE '%Rotation%'
   OR name ILIKE '%rotation%';
```

### Phase 5: Testing

#### 5.1 Unit Tests

**Backend:**
- Test position type validation in branch config handler
- Test position repository queries include position_type
- Test position handler includes position_type in responses

**Frontend:**
- Test BranchPositionQuotaConfig filters out rotation positions
- Test UI labels are updated correctly

#### 5.2 Integration Tests

- Test branch configuration API rejects rotation position quotas
- Test branch configuration UI only shows branch positions
- Test position API returns position_type correctly

#### 5.3 Manual Testing

1. Verify rotation positions don't appear in branch configuration UI
2. Verify branch positions appear correctly with new labels
3. Verify API rejects quota updates for rotation positions
4. Verify existing quotas for branch positions still work

## Files to Modify

### Backend Files
1. `backend/internal/repositories/postgres/migrations.go` - Add migration
2. `backend/internal/domain/models/staff.go` - Add PositionType
3. `backend/internal/repositories/postgres/repositories.go` - Update queries
4. `backend/internal/handlers/position_handler.go` - Include position_type
5. `backend/internal/handlers/branch_config_handler.go` - Add validation

### Frontend Files
1. `frontend/src/lib/api/position.ts` - Add position_type to interface
2. `frontend/src/components/branch/BranchPositionQuotaConfig.tsx` - Filter and rename labels

## Migration Strategy

1. **Backward Compatibility:**
   - Default all existing positions to 'branch' type
   - Migration updates rotation positions based on name patterns
   - API continues to work with existing data

2. **Rollout:**
   - Deploy backend changes first (with default 'branch' for all)
   - Run migration to update rotation positions
   - Deploy frontend changes
   - Verify functionality

## Risk Assessment

### Low Risk
- Adding new column with default value
- Filtering in UI (doesn't affect existing data)
- Renaming labels (cosmetic change)

### Medium Risk
- Migration script needs to correctly identify rotation positions
- Need to verify no existing quotas are set for rotation positions

### Mitigation
- Test migration on staging first
- Backup database before migration
- Verify rotation positions don't have existing quotas before deployment

## Success Criteria

1. ✅ All positions have a `position_type` value ('branch' or 'rotation')
2. ✅ Branch configuration UI only shows branch-type positions
3. ✅ Branch configuration UI shows "Preferred" and "Minimum" labels
4. ✅ API rejects quota updates for rotation-type positions
5. ✅ Existing branch position quotas continue to work
6. ✅ No data loss during migration

## Related Requirements

- **FR-BM-XX:** Branch Configuration Management (to be added to SOFTWARE_REQUIREMENTS.md)
- **BR-BM-XX:** Only branch-type positions can have quota settings (to be added to docs/business-rules.md)

## Notes

- The term "Preferred" replaces "Required Staff" to better reflect that it's a target/ideal number
- The term "Minimum" replaces "Minimum Required" for brevity
- Rotation positions are managed through rotation staff assignments, not branch quotas
- This change aligns with the business logic that rotation staff are assigned dynamically, not through fixed quotas

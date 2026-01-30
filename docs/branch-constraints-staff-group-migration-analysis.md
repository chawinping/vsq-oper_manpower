# Analysis: Migrating Branch Constraints from Fixed Fields to Staff Groups

## Current State

### Current Implementation (Deprecated Fields)
The Daily Staff Constraints feature currently uses **four fixed minimum count fields**:
- `min_front_staff` (includes managers)
- `min_managers`
- `min_doctor_assistant`
- `min_total_staff`

These fields are:
- ✅ Already marked as DEPRECATED in the codebase
- ✅ Stored in `branch_constraints` table
- ✅ Used in the UI (`BranchPositionQuotaConfig.tsx`)
- ✅ Have a newer staff group-based system already implemented

### Existing Staff Group System
The system **already has** a staff group-based constraint system:
- ✅ **Branch Type Constraints** already use staff groups (`BranchTypeConstraints` with `StaffGroupRequirements`)
- ✅ **Branch Constraints** have infrastructure for staff groups (`BranchConstraintStaffGroup` model exists)
- ✅ Database tables exist: `branch_constraint_staff_groups` table
- ✅ Repository methods exist: `BulkUpsertWithStaffGroups()` in `branch_constraints_repo.go`
- ✅ UI exists for branch type constraints using staff groups (`branch-types/page.tsx`)

## What Needs to Change

### 1. Frontend Changes

#### File: `frontend/src/components/branch/BranchPositionQuotaConfig.tsx`

**Current UI Structure:**
- Table with fixed columns: Day | Status | Min Front Staff | Min Managers | Min Doctor Assistant | Min Total Staff
- Input fields for each of the 4 deprecated fields
- `handleConstraintChange()` updates the 4 fields
- `handleSaveConstraints()` sends the 4 fields to backend

**Required Changes:**
1. **Load Staff Groups**
   - Fetch active staff groups using `staffGroupApi.list()`
   - Filter to only active staff groups (`is_active: true`)

2. **Update UI Table Structure**
   - Replace fixed 4 columns with dynamic columns based on staff groups
   - Similar to how `branch-types/page.tsx` displays constraints (lines 461-512)
   - Each staff group becomes a column header
   - Each day becomes a row

3. **Update State Management**
   - Change from storing 4 fixed fields to storing a map of `staff_group_id -> minimum_count`
   - Update `constraints` Map to store `StaffGroupRequirements[]` instead of fixed fields

4. **Update Constraint Change Handler**
   ```typescript
   // OLD:
   handleConstraintChange(dayOfWeek, 'min_front_staff', value)
   
   // NEW:
   handleConstraintChange(dayOfWeek, staffGroupId, value)
   ```

5. **Update Save Handler**
   - Convert constraints to `StaffGroupRequirement[]` format
   - Send to backend using new API structure

6. **Update Reset to Defaults**
   - Load branch type constraints (already uses staff groups)
   - Copy staff group requirements from branch type

7. **Update API Interface**
   - Modify `ConstraintsUpdate` in `branch-config.ts` to use staff groups

#### File: `frontend/src/lib/api/branch-config.ts`

**Current Interface:**
```typescript
export interface ConstraintsUpdate {
  day_of_week: number;
  min_front_staff: number;
  min_managers: number;
  min_doctor_assistant: number;
  min_total_staff: number;
}
```

**Required Changes:**
```typescript
export interface StaffGroupRequirement {
  staff_group_id: string;
  minimum_count: number;
}

export interface ConstraintsUpdate {
  day_of_week: number;
  staff_group_requirements: StaffGroupRequirement[];
}
```

### 2. Backend Changes

#### File: `backend/internal/handlers/branch_config_handler.go`

**Current Handler:**
- `UpdateConstraints()` accepts `ConstraintUpdate` with 4 fixed fields
- Converts to `BranchConstraints` model with deprecated fields
- Uses `BulkUpsert()` method

**Required Changes:**
1. **Update Request Structure**
   ```go
   // OLD:
   type ConstraintUpdate struct {
       DayOfWeek          int
       MinFrontStaff      int
       MinManagers        int
       MinDoctorAssistant int
       MinTotalStaff      int
   }
   
   // NEW:
   type StaffGroupRequirement struct {
       StaffGroupID uuid.UUID `json:"staff_group_id" binding:"required"`
       MinimumCount int       `json:"minimum_count" binding:"required,min=0"`
   }
   
   type ConstraintUpdate struct {
       DayOfWeek              int                     `json:"day_of_week" binding:"required"`
       StaffGroupRequirements []StaffGroupRequirement `json:"staff_group_requirements" binding:"required"`
   }
   ```

2. **Update Handler Logic**
   - Convert request to `BranchConstraints` with `StaffGroupRequirements[]`
   - Use `BulkUpsertWithStaffGroups()` instead of `BulkUpsert()`
   - Set deprecated fields to 0 (already done in `BulkUpsertWithStaffGroups`)

3. **Update getResolvedConstraints()**
   - Already loads staff group requirements ✅
   - Already handles inheritance correctly ✅
   - No changes needed here

#### File: `backend/internal/domain/models/branch_constraints.go`

**Current State:**
- Fields marked as DEPRECATED ✅
- `StaffGroupRequirements` field exists ✅
- No changes needed to model

#### File: `backend/internal/repositories/postgres/branch_constraints_repo.go`

**Current State:**
- `BulkUpsertWithStaffGroups()` already exists ✅
- Already sets deprecated fields to 0 ✅
- No changes needed

### 3. Database Changes

**Current State:**
- `branch_constraints` table has deprecated columns (can be kept for backward compatibility)
- `branch_constraint_staff_groups` table exists ✅
- No schema changes required immediately

**Optional Future Migration:**
- Could add migration to drop deprecated columns after migration is complete
- Should keep columns during transition period for rollback capability

### 4. API Contract Changes

**Breaking Change:**
- `PUT /api/branches/:id/config/constraints` request body changes
- Old format will no longer work
- Need to update API documentation

**Backward Compatibility Options:**
1. **Support Both Formats** (temporary)
   - Check if request has `staff_group_requirements` field
   - If yes, use new format
   - If no, convert old format to new format (map old fields to staff groups)
   - This requires mapping logic: `min_front_staff` → "Front Staff" group, etc.

2. **Version API**
   - Add `/api/v2/branches/:id/config/constraints` endpoint
   - Keep old endpoint for transition period

3. **Clean Break**
   - Update all clients at once
   - No backward compatibility

## Migration Strategy

### Phase 1: Preparation
1. ✅ Verify staff groups are configured in the system
2. ✅ Ensure all branches have staff groups that map to the 4 old fields
3. Create mapping document: which staff groups correspond to which old fields

### Phase 2: Backend Changes
1. Update `UpdateConstraints` handler to accept staff group format
2. Add backward compatibility layer (convert old format to new format)
3. Test with both old and new API calls
4. Update `GetConstraints` to return staff groups (already does ✅)

### Phase 3: Frontend Changes
1. Update UI to use staff groups
2. Update API client interfaces
3. Test constraint editing, saving, resetting
4. Test inheritance from branch types

### Phase 4: Data Migration (Optional)
1. Migrate existing constraints to staff groups
   - Read existing `min_front_staff`, `min_managers`, etc.
   - Map to appropriate staff groups
   - Create `branch_constraint_staff_groups` records
2. Verify migrated data

### Phase 5: Cleanup
1. Remove backward compatibility code
2. Remove deprecated fields from database (optional)
3. Update documentation

## Key Considerations

### 1. Staff Group Mapping
**Question:** How do the 4 old fields map to staff groups?
- `min_front_staff` → Which staff group(s)?
- `min_managers` → Which staff group(s)?
- `min_doctor_assistant` → Which staff group(s)?
- `min_total_staff` → How is this calculated? Sum of all groups?

**Answer Needed:** 
- Review existing staff groups in the system
- Determine if there are staff groups like "Front Staff", "Managers", "Doctor Assistants"
- Or if these need to be created

### 2. Inheritance Behavior
**Current:** Branch constraints inherit from branch type constraints
**After Migration:** Should still work the same way
- Branch type constraints already use staff groups ✅
- `getResolvedConstraints()` already handles staff groups ✅

### 2a. Override Features (CRITICAL - FULLY PRESERVED)
**✅ Override functionality will be COMPLETELY PRESERVED:**

1. **Override Flag (`is_overridden`)**
   - Still stored in `branch_constraints` table ✅
   - Still used in `getResolvedConstraints()` priority logic ✅
   - Still copied when saving overridden constraints ✅

2. **Override Priority Logic** (lines 488-520 in `branch_config_handler.go`)
   - **Priority 1:** Overridden branch constraints (with staff groups) ✅
   - **Priority 2:** Branch type constraints (with staff groups) ✅
   - **Priority 3:** Defaults (empty staff groups) ✅
   - This logic already works with staff groups! ✅

3. **Override Visual Indicators**
   - Yellow row background for overridden constraints ✅
   - "Overridden" vs "Inherited" badges ✅
   - Status column showing override state ✅

4. **Reset to Defaults**
   - Still works by copying branch type staff group requirements ✅
   - Sets `is_overridden: false` ✅

5. **Database Storage**
   - `is_overridden` column preserved ✅
   - `inherited_from_branch_type_id` column preserved ✅
   - Staff group requirements stored separately in `branch_constraint_staff_groups` ✅

**Conclusion:** The override mechanism is **completely independent** of whether constraints use fixed fields or staff groups. The `is_overridden` flag and priority logic remain unchanged.

### 3. UI Consistency
**Current:** Branch type constraints UI uses staff groups
**After Migration:** Branch constraints UI should match
- Same table structure
- Same input fields
- Same visual indicators

### 4. Validation
- Need to validate that staff groups exist
- Need to validate minimum_count >= 0
- Need to validate day_of_week is 0-6

### 5. Empty Constraints
**Question:** What if a day has no staff group requirements?
- Currently defaults to all zeros
- Should continue to default to empty array `[]`

## Files to Modify

### Frontend
1. `frontend/src/components/branch/BranchPositionQuotaConfig.tsx` - Major changes
2. `frontend/src/lib/api/branch-config.ts` - Update interfaces

### Backend
1. `backend/internal/handlers/branch_config_handler.go` - Update handler
2. `backend/internal/domain/models/branch_constraints.go` - Already has staff groups ✅
3. `backend/internal/repositories/postgres/branch_constraints_repo.go` - Already supports staff groups ✅

### Documentation
1. API documentation
2. User guide (if exists)

## Testing Checklist

### Backend Tests
- [ ] Update constraints with staff groups
- [ ] Get constraints returns staff groups
- [ ] Inheritance from branch type works
- [ ] Override behavior works
- [ ] Reset to defaults works
- [ ] Validation errors work correctly

### Frontend Tests
- [ ] UI displays staff groups correctly
- [ ] Editing constraints works
- [ ] Saving constraints works
- [ ] Reset to defaults works
- [ ] Visual indicators (inherited/overridden) work
- [ ] Empty staff groups handled correctly

### Integration Tests
- [ ] End-to-end constraint configuration
- [ ] Constraint inheritance flow
- [ ] Constraint override flow
- [ ] Staff allocation uses constraints correctly

## Risks and Mitigation

### Risk 1: Breaking Existing Integrations
**Mitigation:** Implement backward compatibility layer

### Risk 2: Data Loss During Migration
**Mitigation:** 
- Keep deprecated fields during transition
- Create data migration script
- Test migration on staging first

### Risk 3: Staff Groups Not Configured
**Mitigation:**
- Check for staff groups before allowing constraint updates
- Show helpful error message
- Provide link to staff groups configuration

### Risk 4: UI Complexity
**Mitigation:**
- Reuse branch type constraints UI pattern
- Keep table structure similar
- Maintain visual consistency

## Benefits of Migration

1. **Flexibility:** Can add/remove staff groups without code changes
2. **Consistency:** Branch constraints match branch type constraints
3. **Maintainability:** Remove deprecated code
4. **Scalability:** Easy to add new staff group types
5. **Alignment:** Matches the direction the system is already moving

## Estimated Effort

- **Backend Changes:** 2-4 hours
- **Frontend Changes:** 4-6 hours
- **Testing:** 2-3 hours
- **Documentation:** 1 hour
- **Total:** ~9-14 hours

## Next Steps

1. **Review this analysis** with team
2. **Identify staff group mapping** (old fields → staff groups)
3. **Decide on backward compatibility** approach
4. **Create implementation plan** with specific tasks
5. **Set up test environment** for migration
6. **Implement changes** following migration strategy

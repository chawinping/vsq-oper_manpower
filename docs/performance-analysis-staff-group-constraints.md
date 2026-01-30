# Performance Impact Analysis: Staff Group-Based Constraints

## Current Implementation

### Database Schema
- **Table**: `branch_type_constraints`
- **Structure**: Fixed columns (4 integer fields)
  - `min_front_staff` (INTEGER)
  - `min_managers` (INTEGER)
  - `min_doctor_assistant` (INTEGER)
  - `min_total_staff` (INTEGER)
- **Indexes**: 
  - `idx_branch_type_constraints_type` on `branch_type_id`
  - `idx_branch_type_constraints_day` on `day_of_week`
- **Unique Constraint**: `(branch_type_id, day_of_week)`

### Query Patterns
1. **Get constraints for branch type**: Single table SELECT with WHERE clause
   ```sql
   SELECT * FROM branch_type_constraints 
   WHERE branch_type_id = $1 
   ORDER BY day_of_week
   ```
   - **Complexity**: O(1) with index, returns 7 rows max
   - **Joins**: None
   - **Estimated time**: <1ms

2. **Bulk upsert**: Single table INSERT/UPDATE with ON CONFLICT
   ```sql
   INSERT INTO branch_type_constraints (...) 
   VALUES (...) 
   ON CONFLICT (branch_type_id, day_of_week) DO UPDATE ...
   ```
   - **Complexity**: O(1) per row, typically 7 rows
   - **Estimated time**: <5ms for 7 rows

3. **Usage in branch config**: Simple field access
   - Direct field access: `constraint.MinFrontStaff`
   - No additional queries needed

## Proposed Implementation Options

### Option 1: Junction Table (Normalized)
**Schema**:
```sql
CREATE TABLE branch_type_constraint_staff_groups (
    id UUID PRIMARY KEY,
    branch_type_constraint_id UUID REFERENCES branch_type_constraints(id),
    staff_group_id UUID REFERENCES staff_groups(id),
    minimum_count INTEGER NOT NULL DEFAULT 0,
    UNIQUE(branch_type_constraint_id, staff_group_id)
);
```

**Query Changes**:
1. **Get constraints**: Requires JOIN
   ```sql
   SELECT c.*, sg.staff_group_id, sg.minimum_count
   FROM branch_type_constraints c
   LEFT JOIN branch_type_constraint_staff_groups sg ON c.id = sg.branch_type_constraint_id
   WHERE c.branch_type_id = $1
   ORDER BY c.day_of_week, sg.staff_group_id
   ```
   - **Complexity**: O(n) where n = number of staff groups
   - **Joins**: 1 LEFT JOIN
   - **Estimated time**: 2-5ms (depends on staff group count)

2. **Bulk upsert**: Requires transaction with multiple inserts
   - Insert/update constraints (7 rows)
   - Delete old staff group mappings
   - Insert new staff group mappings (7 × staff_group_count rows)
   - **Estimated time**: 10-20ms for 7 days × 5 staff groups = 35 rows

**Performance Impact**:
- ✅ **Read**: +2-4ms per query (acceptable)
- ⚠️ **Write**: +5-15ms per bulk upsert (acceptable)
- ✅ **Scalability**: Good (normalized, proper indexes)
- ✅ **Data integrity**: Excellent (foreign keys)

### Option 2: JSONB Column (Denormalized)
**Schema**:
```sql
ALTER TABLE branch_type_constraints 
ADD COLUMN staff_group_requirements JSONB DEFAULT '{}';
-- Example: {"staff_group_id_1": 2, "staff_group_id_2": 1, ...}
```

**Query Changes**:
1. **Get constraints**: Single table SELECT with JSONB parsing
   ```sql
   SELECT *, staff_group_requirements 
   FROM branch_type_constraints 
   WHERE branch_type_id = $1
   ```
   - **Complexity**: O(1) table access, O(n) JSON parsing
   - **Estimated time**: 1-3ms (JSON parsing overhead)

2. **Bulk upsert**: Single table INSERT/UPDATE with JSONB
   ```sql
   UPDATE branch_type_constraints 
   SET staff_group_requirements = $1 
   WHERE branch_type_id = $2 AND day_of_week = $3
   ```
   - **Estimated time**: 2-5ms per row

**Performance Impact**:
- ✅ **Read**: +1-2ms per query (minimal)
- ✅ **Write**: +1-3ms per row (minimal)
- ⚠️ **Scalability**: Moderate (JSONB can grow large)
- ⚠️ **Data integrity**: Moderate (no foreign key validation)

### Option 3: Hybrid (Recommended)
**Schema**: Keep both for backward compatibility during migration
- Keep old columns (deprecated, for migration)
- Add JSONB column for new staff group-based constraints
- Migrate gradually

## Performance Impact Summary

### Database Layer
| Operation | Current | Option 1 (Junction) | Option 2 (JSONB) | Impact |
|-----------|---------|---------------------|------------------|--------|
| **Read (single branch type)** | <1ms | 2-5ms | 1-3ms | **Low** ✅ |
| **Write (bulk upsert 7 days)** | <5ms | 10-20ms | 5-10ms | **Low-Medium** ⚠️ |
| **Memory per row** | ~100 bytes | ~150 bytes | ~200 bytes | **Low** ✅ |

### Application Layer
| Component | Current | Proposed | Impact |
|-----------|---------|----------|--------|
| **Frontend rendering** | Fixed 4 columns | Dynamic N columns | **Low** ✅ |
| **API response size** | ~500 bytes | ~500-2000 bytes | **Low** ✅ |
| **Data transformation** | None | JSON parsing/grouping | **Low** ✅ |

### Critical Path Analysis

#### 1. Branch Config Resolution (Most Critical)
**Current**: Direct field access
```go
constraint.MinFrontStaff = branchType.MinFrontStaff
```

**Proposed**: Map lookup or JSONB access
```go
constraint.StaffGroupRequirements[staffGroupID] = branchType.StaffGroupRequirements[staffGroupID]
```

**Impact**: 
- **Time**: +0.1-0.5ms per constraint resolution
- **Memory**: +50-200 bytes per constraint
- **Verdict**: **Negligible** ✅

#### 2. Allocation Criteria Engine
**Current**: Uses `MinFrontStaff`, `MinManagers`, etc. directly
**Proposed**: Need to map staff groups to positions, then aggregate

**Impact**:
- **Additional queries**: 1 query to get staff groups (cached)
- **Processing**: +1-2ms per evaluation
- **Verdict**: **Low** ✅ (evaluation happens async)

#### 3. Frontend Constraints Modal
**Current**: Fixed 4 columns, simple rendering
**Proposed**: Dynamic columns based on active staff groups

**Impact**:
- **Initial load**: +1 API call to fetch staff groups (can be cached)
- **Rendering**: O(n) where n = staff group count (typically 3-10)
- **Verdict**: **Low** ✅ (UI rendering is fast)

## Recommendations

### Performance Impact: **LOW** ✅

**Reasons**:
1. **Query frequency**: Constraints are read infrequently (only when viewing/editing branch types)
2. **Data volume**: Small dataset (7 days × branch types × staff groups)
3. **Caching opportunity**: Staff groups rarely change, can be cached
4. **Modern hardware**: PostgreSQL handles JOINs and JSONB efficiently

### Recommended Approach: **Option 1 (Junction Table)**

**Why**:
- ✅ Better data integrity (foreign keys)
- ✅ Easier to query and filter
- ✅ Better for reporting/analytics
- ✅ Proper normalization
- ⚠️ Slightly more complex queries (acceptable trade-off)

**Optimization Strategies**:
1. **Add composite index**: `(branch_type_constraint_id, staff_group_id)`
2. **Cache staff groups**: Load once, reuse (they rarely change)
3. **Batch operations**: Use transactions for bulk updates
4. **Consider materialized view**: If reporting becomes slow

### Migration Strategy
1. **Phase 1**: Add new table, keep old columns
2. **Phase 2**: Migrate data from old to new format
3. **Phase 3**: Update application code
4. **Phase 4**: Remove old columns (after verification)

## Conclusion

**Overall Performance Impact: LOW** ✅

The change from fixed columns to staff group-based constraints will have minimal performance impact:
- **Read operations**: +2-5ms (acceptable)
- **Write operations**: +5-15ms (acceptable)
- **Memory usage**: +50-200 bytes per constraint (negligible)
- **Frontend**: No significant impact

The benefits (flexibility, maintainability, alignment with business logic) far outweigh the minimal performance cost.

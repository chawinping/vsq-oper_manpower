---
title: Daily Expected Revenue Split into 4 Types - Analysis & Alternatives
description: Analysis of splitting daily expected revenue into 4 types (Skin, LS HM, Vitamin cases, Slim Pen cases) for staff allocation calculation
version: 1.0.0
lastUpdated: 2026-01-18 15:30:00
---

# Daily Expected Revenue Split into 4 Types - Analysis & Alternatives

## Overview

Currently, the system uses a single `expected_revenue` field for daily revenue tracking per branch. The requirement is to split this into 4 separate types:

1. **Skin - revenue** (THB) - The original revenue field
2. **LS HM - revenue** (THB) - New revenue type
3. **Vitamin - no. of cases** (Integer) - Case count instead of revenue
4. **Slim Pen - no. of cases** (Integer) - Case count instead of revenue

These variables will be used to calculate staff allocation for each branch for each day.

---

## Current System Architecture

### Database Schema

#### `revenue_data` Table
```sql
CREATE TABLE revenue_data (
    id UUID PRIMARY KEY,
    branch_id UUID NOT NULL REFERENCES branches(id),
    date DATE NOT NULL,
    expected_revenue DECIMAL(15,2) NOT NULL,  -- Single revenue field
    actual_revenue DECIMAL(15,2),
    revenue_source VARCHAR(20) DEFAULT 'branch',
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    UNIQUE(branch_id, date)
);
```

#### `branch_weekly_revenue` Table
```sql
CREATE TABLE branch_weekly_revenue (
    id UUID PRIMARY KEY,
    branch_id UUID NOT NULL REFERENCES branches(id),
    day_of_week INTEGER NOT NULL CHECK (day_of_week >= 0 AND day_of_week <= 6),
    expected_revenue DECIMAL(15,2) NOT NULL DEFAULT 0,  -- Single revenue field
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    UNIQUE(branch_id, day_of_week)
);
```

### Current Usage

1. **Staff Allocation Calculation:**
   - `scenario_calculator.go` uses `expected_revenue` to match scenarios
   - Revenue is used to determine staff requirement tiers
   - Formula: `staff_count = min_staff + (revenue_threshold * multiplier)`

2. **UI Components:**
   - `BranchWeeklyRevenueConfig.tsx` - Configures weekly revenue per day
   - `BranchPositionQuotaConfig.tsx` - Uses revenue for quota calculations
   - Excel import functionality for bulk revenue import

3. **API Endpoints:**
   - `GET/PUT /branches/:id/config/weekly-revenue`
   - `GET /branches/:id/revenue` (for specific dates)
   - Excel import endpoint for revenue data

---

## Requirements Analysis

### Data Types Needed

| Type | Name | Unit | Data Type | Notes |
|------|------|------|-----------|-------|
| 1 | Skin Revenue | THB | DECIMAL(15,2) | Original revenue field |
| 2 | LS HM Revenue | THB | DECIMAL(15,2) | New revenue type |
| 3 | Vitamin Cases | Count | INTEGER | Case count, not revenue |
| 4 | Slim Pen Cases | Count | INTEGER | Case count, not revenue |

### Key Considerations

1. **Backward Compatibility:** Existing `expected_revenue` data should be preserved (migrated to Skin Revenue)
2. **Staff Calculation:** All 4 types need to be used in staff allocation formulas
3. **Excel Import:** Import functionality needs to support all 4 types
4. **UI Updates:** Weekly revenue config UI needs to show all 4 types
5. **API Changes:** All revenue-related endpoints need to handle 4 types
6. **Scenario Matching:** Staff requirement scenarios may need to consider multiple revenue types

---

## Alternative 1: Add Columns to Existing Tables (Column-Based Approach)

### Design Overview

Add 4 new columns to both `revenue_data` and `branch_weekly_revenue` tables, keeping the original `expected_revenue` column for backward compatibility (or rename it to `skin_revenue`).

### Database Schema Changes

#### Updated `revenue_data` Table
```sql
ALTER TABLE revenue_data 
  ADD COLUMN skin_revenue DECIMAL(15,2) DEFAULT 0,
  ADD COLUMN ls_hm_revenue DECIMAL(15,2) DEFAULT 0,
  ADD COLUMN vitamin_cases INTEGER DEFAULT 0,
  ADD COLUMN slim_pen_cases INTEGER DEFAULT 0;

-- Migrate existing expected_revenue to skin_revenue
UPDATE revenue_data SET skin_revenue = expected_revenue WHERE expected_revenue > 0;

-- Optional: Keep expected_revenue as computed column or drop after migration
-- Option A: Keep as computed (total of all revenue types)
-- Option B: Drop after migration period
```

#### Updated `branch_weekly_revenue` Table
```sql
ALTER TABLE branch_weekly_revenue 
  ADD COLUMN skin_revenue DECIMAL(15,2) DEFAULT 0,
  ADD COLUMN ls_hm_revenue DECIMAL(15,2) DEFAULT 0,
  ADD COLUMN vitamin_cases INTEGER DEFAULT 0,
  ADD COLUMN slim_pen_cases INTEGER DEFAULT 0;

-- Migrate existing expected_revenue to skin_revenue
UPDATE branch_weekly_revenue SET skin_revenue = expected_revenue WHERE expected_revenue > 0;
```

### Domain Model Changes

#### Updated `RevenueData` Model
```go
type RevenueData struct {
    ID              uuid.UUID `json:"id" db:"id"`
    BranchID        uuid.UUID `json:"branch_id" db:"branch_id"`
    Date            time.Time `json:"date" db:"date"`
    
    // Original (kept for backward compatibility)
    ExpectedRevenue float64  `json:"expected_revenue,omitempty" db:"expected_revenue"` // Deprecated
    
    // New fields
    SkinRevenue     float64  `json:"skin_revenue" db:"skin_revenue"`
    LSHMRevenue     float64  `json:"ls_hm_revenue" db:"ls_hm_revenue"`
    VitaminCases    int      `json:"vitamin_cases" db:"vitamin_cases"`
    SlimPenCases    int      `json:"slim_pen_cases" db:"slim_pen_cases"`
    
    ActualRevenue   *float64 `json:"actual_revenue,omitempty" db:"actual_revenue"`
    RevenueSource   string   `json:"revenue_source" db:"revenue_source"`
    CreatedAt       time.Time `json:"created_at" db:"created_at"`
    UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}
```

#### Updated `BranchWeeklyRevenue` Model
```go
type BranchWeeklyRevenue struct {
    ID              uuid.UUID `json:"id" db:"id"`
    BranchID        uuid.UUID `json:"branch_id" db:"branch_id"`
    DayOfWeek       int       `json:"day_of_week" db:"day_of_week"`
    
    // Original (kept for backward compatibility)
    ExpectedRevenue float64   `json:"expected_revenue,omitempty" db:"expected_revenue"` // Deprecated
    
    // New fields
    SkinRevenue     float64   `json:"skin_revenue" db:"skin_revenue"`
    LSHMRevenue     float64   `json:"ls_hm_revenue" db:"ls_hm_revenue"`
    VitaminCases    int       `json:"vitamin_cases" db:"vitamin_cases"`
    SlimPenCases    int       `json:"slim_pen_cases" db:"slim_pen_cases"`
    
    CreatedAt       time.Time `json:"created_at" db:"created_at"`
    UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}
```

### Staff Calculation Changes

The staff allocation calculation would need to consider all 4 types. Options:

**Option A: Weighted Sum**
```go
totalRevenueValue := skinRevenue + lsHMRevenue + (vitaminCases * vitaminCaseValue) + (slimPenCases * slimPenCaseValue)
```

**Option B: Separate Calculations per Type**
```go
staffFromSkin := calculateStaff(skinRevenue, skinMultiplier)
staffFromLSHM := calculateStaff(lsHMRevenue, lsHMMultiplier)
staffFromVitamin := calculateStaff(vitaminCases, vitaminCaseMultiplier)
staffFromSlimPen := calculateStaff(slimPenCases, slimPenCaseMultiplier)
totalStaff := max(staffFromSkin, staffFromLSHM, staffFromVitamin, staffFromSlimPen) // or sum
```

### Pros

✅ **Simple Implementation:** Minimal schema changes, straightforward migration
✅ **Easy Queries:** All data in one row, no joins needed
✅ **Backward Compatible:** Can keep `expected_revenue` during transition
✅ **Performance:** Single table queries, fast lookups
✅ **Atomic Updates:** All 4 types updated in single transaction
✅ **UI Friendly:** Easy to display all types in single form/table

### Cons

❌ **Schema Rigidity:** Adding new revenue types requires schema changes
❌ **Wide Tables:** Tables become wider with more columns
❌ **Mixed Data Types:** Revenue (DECIMAL) and Cases (INTEGER) in same table
❌ **Migration Complexity:** Need to migrate existing data carefully
❌ **Null Handling:** Need to handle cases where some types don't apply

### Implementation Effort

- **Database Migration:** Medium (add columns, migrate data)
- **Backend Changes:** Medium (update models, repositories, handlers)
- **Frontend Changes:** Medium (update UI components, forms)
- **Excel Import:** Medium (update parser to handle 4 columns)
- **Testing:** Medium (test all 4 types, migration, backward compatibility)

---

## Alternative 2: Separate Tables per Revenue Type (Table-Based Approach)

### Design Overview

Create separate tables for each revenue type, maintaining a relationship through `branch_id` and `date`/`day_of_week`.

### Database Schema Changes

#### New Tables Structure

```sql
-- Skin Revenue (original revenue)
CREATE TABLE revenue_data_skin (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    branch_id UUID NOT NULL REFERENCES branches(id),
    date DATE NOT NULL,
    revenue DECIMAL(15,2) NOT NULL DEFAULT 0,
    actual_revenue DECIMAL(15,2),
    revenue_source VARCHAR(20) DEFAULT 'branch',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(branch_id, date)
);

-- LS HM Revenue
CREATE TABLE revenue_data_ls_hm (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    branch_id UUID NOT NULL REFERENCES branches(id),
    date DATE NOT NULL,
    revenue DECIMAL(15,2) NOT NULL DEFAULT 0,
    actual_revenue DECIMAL(15,2),
    revenue_source VARCHAR(20) DEFAULT 'branch',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(branch_id, date)
);

-- Vitamin Cases
CREATE TABLE revenue_data_vitamin_cases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    branch_id UUID NOT NULL REFERENCES branches(id),
    date DATE NOT NULL,
    case_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(branch_id, date)
);

-- Slim Pen Cases
CREATE TABLE revenue_data_slim_pen_cases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    branch_id UUID NOT NULL REFERENCES branches(id),
    date DATE NOT NULL,
    case_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(branch_id, date)
);

-- Similar structure for weekly revenue tables
CREATE TABLE branch_weekly_revenue_skin (...);
CREATE TABLE branch_weekly_revenue_ls_hm (...);
CREATE TABLE branch_weekly_revenue_vitamin_cases (...);
CREATE TABLE branch_weekly_revenue_slim_pen_cases (...);
```

#### Migration Strategy

```sql
-- Migrate existing revenue_data to revenue_data_skin
INSERT INTO revenue_data_skin (branch_id, date, revenue, actual_revenue, revenue_source, created_at, updated_at)
SELECT branch_id, date, expected_revenue, actual_revenue, revenue_source, created_at, updated_at
FROM revenue_data;

-- Similar for branch_weekly_revenue
INSERT INTO branch_weekly_revenue_skin (branch_id, day_of_week, revenue, created_at, updated_at)
SELECT branch_id, day_of_week, expected_revenue, created_at, updated_at
FROM branch_weekly_revenue;
```

### Domain Model Changes

#### New Models
```go
type RevenueDataSkin struct {
    ID            uuid.UUID `json:"id" db:"id"`
    BranchID      uuid.UUID `json:"branch_id" db:"branch_id"`
    Date          time.Time `json:"date" db:"date"`
    Revenue       float64   `json:"revenue" db:"revenue"`
    ActualRevenue *float64  `json:"actual_revenue,omitempty" db:"actual_revenue"`
    RevenueSource string    `json:"revenue_source" db:"revenue_source"`
    CreatedAt     time.Time `json:"created_at" db:"created_at"`
    UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

type RevenueDataLSHM struct {
    // Similar structure
}

type RevenueDataVitaminCases struct {
    ID         uuid.UUID `json:"id" db:"id"`
    BranchID   uuid.UUID `json:"branch_id" db:"branch_id"`
    Date       time.Time `json:"date" db:"date"`
    CaseCount  int       `json:"case_count" db:"case_count"`
    CreatedAt  time.Time `json:"created_at" db:"created_at"`
    UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

type RevenueDataSlimPenCases struct {
    // Similar structure
}

// Aggregated view model for API responses
type BranchRevenueData struct {
    BranchID      uuid.UUID `json:"branch_id"`
    Date          time.Time `json:"date"`
    SkinRevenue   *RevenueDataSkin        `json:"skin_revenue,omitempty"`
    LSHMRevenue   *RevenueDataLSHM        `json:"ls_hm_revenue,omitempty"`
    VitaminCases  *RevenueDataVitaminCases `json:"vitamin_cases,omitempty"`
    SlimPenCases  *RevenueDataSlimPenCases `json:"slim_pen_cases,omitempty"`
}
```

### Repository Changes

Would need separate repositories or a unified repository interface:

```go
type RevenueRepository interface {
    // Skin
    CreateSkin(revenue *models.RevenueDataSkin) error
    GetSkinByBranchID(branchID uuid.UUID, startDate, endDate time.Time) ([]*models.RevenueDataSkin, error)
    
    // LS HM
    CreateLSHM(revenue *models.RevenueDataLSHM) error
    GetLSHMByBranchID(branchID uuid.UUID, startDate, endDate time.Time) ([]*models.RevenueDataLSHM, error)
    
    // Vitamin Cases
    CreateVitaminCases(cases *models.RevenueDataVitaminCases) error
    GetVitaminCasesByBranchID(branchID uuid.UUID, startDate, endDate time.Time) ([]*models.RevenueDataVitaminCases, error)
    
    // Slim Pen Cases
    CreateSlimPenCases(cases *models.RevenueDataSlimPenCases) error
    GetSlimPenCasesByBranchID(branchID uuid.UUID, startDate, endDate time.Time) ([]*models.RevenueDataSlimPenCases, error)
    
    // Aggregated
    GetAggregatedByBranchID(branchID uuid.UUID, startDate, endDate time.Time) ([]*models.BranchRevenueData, error)
}
```

### Staff Calculation Changes

Would need to query all 4 tables and aggregate:

```go
skinRevenue := getSkinRevenue(branchID, date)
lsHMRevenue := getLSHMRevenue(branchID, date)
vitaminCases := getVitaminCases(branchID, date)
slimPenCases := getSlimPenCases(branchID, date)

// Then calculate staff allocation using all 4 values
```

### Pros

✅ **Flexibility:** Easy to add new revenue types without schema changes
✅ **Type Safety:** Each table has appropriate data types (DECIMAL for revenue, INTEGER for cases)
✅ **Separation of Concerns:** Each revenue type is independent
✅ **Scalability:** Can add indexes per table type
✅ **Optional Types:** Some branches may not have all types (no NULL handling needed)
✅ **Clear Semantics:** Table names clearly indicate what they store

### Cons

❌ **Complex Queries:** Need joins or multiple queries to get all types
❌ **More Tables:** 8 new tables (4 for daily, 4 for weekly)
❌ **Transaction Complexity:** Updates across multiple tables need careful transaction handling
❌ **Performance:** Multiple queries/joins may be slower
❌ **Code Complexity:** More repositories, handlers, and models to maintain
❌ **Migration Effort:** More complex migration script

### Implementation Effort

- **Database Migration:** High (create 8 new tables, migrate data, potentially keep old tables)
- **Backend Changes:** High (new models, repositories, handlers, aggregation logic)
- **Frontend Changes:** Medium-High (update UI to handle aggregated data)
- **Excel Import:** High (update parser to handle multiple tables)
- **Testing:** High (test all 4 types, aggregation, transaction handling)

---

## Comparison Matrix

| Aspect | Alternative 1: Columns | Alternative 2: Tables |
|--------|----------------------|----------------------|
| **Schema Complexity** | Low (add columns) | High (8 new tables) |
| **Query Complexity** | Low (single query) | High (joins/multiple queries) |
| **Performance** | High (single table) | Medium (joins needed) |
| **Flexibility** | Low (schema change for new types) | High (add new tables) |
| **Type Safety** | Medium (mixed types) | High (separate types) |
| **Migration Effort** | Medium | High |
| **Code Changes** | Medium | High |
| **Maintenance** | Low | Medium-High |
| **Backward Compatibility** | Easy (keep old column) | Medium (migrate to new tables) |
| **UI Updates** | Easy (single form) | Medium (aggregate display) |

---

## Recommendation

### Recommended: **Alternative 1 (Column-Based Approach)**

**Rationale:**
1. **Simpler Implementation:** Less code changes, easier migration
2. **Better Performance:** Single table queries are faster
3. **Easier UI:** All 4 types in one form/table
4. **Atomic Updates:** All types updated together
5. **Sufficient Flexibility:** 4 types should be stable, unlikely to need frequent additions

**When to Consider Alternative 2:**
- If you expect to add many more revenue types frequently
- If different revenue types have vastly different attributes
- If you need different access patterns per type
- If you want to scale different types independently

### Implementation Plan for Alternative 1

1. **Phase 1: Database Migration**
   - Add 4 new columns to `revenue_data` and `branch_weekly_revenue`
   - Migrate existing `expected_revenue` to `skin_revenue`
   - Keep `expected_revenue` as computed column (or deprecate gradually)

2. **Phase 2: Backend Updates**
   - Update domain models (`RevenueData`, `BranchWeeklyRevenue`)
   - Update repositories to handle new columns
   - Update handlers/API endpoints
   - Update staff calculation logic to use all 4 types

3. **Phase 3: Frontend Updates**
   - Update `BranchWeeklyRevenueConfig.tsx` to show 4 input fields
   - Update revenue display components
   - Update Excel import/export functionality

4. **Phase 4: Testing & Migration**
   - Test staff calculation with all 4 types
   - Test Excel import with new format
   - Migrate existing data
   - Update documentation

---

## Staff Allocation Calculation Strategy

### Proposed Formula Options

**Option 1: Weighted Sum**
```go
// Convert cases to revenue equivalent, then sum
totalRevenueValue := skinRevenue + 
                     lsHMRevenue + 
                     (vitaminCases * vitaminCaseToRevenueMultiplier) + 
                     (slimPenCases * slimPenCaseToRevenueMultiplier)

staffCount := calculateStaffFromRevenue(totalRevenueValue)
```

**Option 2: Maximum Requirement**
```go
staffFromSkin := calculateStaffFromRevenue(skinRevenue)
staffFromLSHM := calculateStaffFromRevenue(lsHMRevenue)
staffFromVitamin := calculateStaffFromCases(vitaminCases)
staffFromSlimPen := calculateStaffFromCases(slimPenCases)

staffCount := max(staffFromSkin, staffFromLSHM, staffFromVitamin, staffFromSlimPen)
```

**Option 3: Position-Specific Calculation**
```go
// Different positions may depend on different revenue types
// E.g., Front staff depends on Skin + LS HM revenue
// Doctor assistants depend on Vitamin + Slim Pen cases

frontStaff := calculateFromRevenue(skinRevenue + lsHMRevenue)
doctorAssistantStaff := calculateFromCases(vitaminCases + slimPenCases)
```

**Recommendation:** Start with **Option 1 (Weighted Sum)** as it's simplest and maintains current calculation pattern. Can evolve to Option 3 if business rules require position-specific calculations.

---

## Questions to Clarify

1. **Staff Calculation Formula:** How should the 4 types be combined? (Weighted sum? Maximum? Position-specific?)
2. **Case-to-Revenue Conversion:** What multiplier should be used for Vitamin and Slim Pen cases?
3. **Backward Compatibility:** Should `expected_revenue` be kept indefinitely or deprecated after migration?
4. **Excel Import Format:** What should the Excel column structure be? (4 separate columns?)
5. **UI Display:** Should all 4 types be shown together or in separate sections?
6. **Validation Rules:** Any constraints on the values? (e.g., cases must be >= 0, revenue >= 0)
7. **Scenario Matching:** Should staff requirement scenarios match against individual types or combined value?

---

## Next Steps

1. **Review this analysis** with stakeholders
2. **Clarify questions** above
3. **Choose alternative** (recommend Alternative 1)
4. **Define staff calculation formula** (how to combine 4 types)
5. **Create detailed implementation plan**
6. **Implement changes** following chosen alternative

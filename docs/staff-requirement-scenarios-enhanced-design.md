---
title: Enhanced Staff Requirement Scenarios Design - Revenue Levels & Day-of-Week Integration
description: Enhanced design for Alternative 2 with revenue level tiers and day-of-week revenue integration
version: 1.0.0
lastUpdated: 2026-01-13 20:00:00
---

# Enhanced Staff Requirement Scenarios Design

## Overview

This document presents enhanced solutions for **Alternative 2: Scenario-Based Configuration System** with the following enhancements:

1. **Revenue Level Tiers**: Configurable revenue levels (e.g., Level 5 = 500K-600K THB, Level 4 = 400K-500K THB)
2. **Day-of-Week Revenue Integration**: Link scenarios to expected revenue per day of week from `branch_weekly_revenue` table
3. **Dual Revenue Matching**: Support matching on both day-of-week baseline revenue and specific date revenue

---

## Solution 1: Revenue Level Tiers with Day-of-Week Baseline (Recommended)

### Concept

Scenarios match based on:
- **Day-of-week baseline revenue** (from `branch_weekly_revenue`) - used as the primary matching criteria
- **Revenue level tier** (e.g., Level 5, Level 4) - configurable ranges
- **Doctor count** - number of doctors scheduled
- **Specific date revenue** (optional override) - from `revenue_data` if available

### Database Design

#### 1.1 Enhanced: `revenue_level_tiers` Table

```sql
CREATE TABLE IF NOT EXISTS revenue_level_tiers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    level_number INTEGER NOT NULL UNIQUE CHECK (level_number >= 1 AND level_number <= 10),
    level_name VARCHAR(50) NOT NULL, -- e.g., 'Very High', 'High', 'Medium'
    min_revenue DECIMAL(15,2) NOT NULL CHECK (min_revenue >= 0),
    max_revenue DECIMAL(15,2), -- NULL means no upper limit
    display_order INTEGER NOT NULL DEFAULT 0,
    color_code VARCHAR(20), -- For UI display (e.g., '#FF0000' for red)
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_revenue_level_tiers_level ON revenue_level_tiers(level_number);
CREATE INDEX idx_revenue_level_tiers_range ON revenue_level_tiers(min_revenue, max_revenue);
```

**Example Data:**
```sql
INSERT INTO revenue_level_tiers (level_number, level_name, min_revenue, max_revenue, display_order, color_code, description) VALUES
    (1, 'Very Low', 0, 100000, 1, '#CCCCCC', 'Low revenue days'),
    (2, 'Low', 100000, 200000, 2, '#99CCFF', 'Below average revenue'),
    (3, 'Medium', 200000, 300000, 3, '#66FF99', 'Average revenue days'),
    (4, 'High', 300000, 400000, 4, '#FFCC66', 'Above average revenue'),
    (5, 'Very High', 400000, 500000, 5, '#FF9966', 'High revenue days'),
    (6, 'Extremely High', 500000, 600000, 6, '#FF6666', 'Very high revenue days'),
    (7, 'Peak', 600000, NULL, 7, '#FF0000', 'Peak revenue days');
```

#### 1.2 Enhanced: `staff_requirement_scenarios` Table

```sql
CREATE TABLE IF NOT EXISTS staff_requirement_scenarios (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    scenario_name VARCHAR(100) NOT NULL,
    description TEXT,
    
    -- Revenue matching options (at least one must be specified)
    revenue_level_tier_id UUID REFERENCES revenue_level_tiers(id), -- Match by tier level
    min_revenue DECIMAL(15,2), -- Direct revenue threshold (alternative to tier)
    max_revenue DECIMAL(15,2), -- Direct revenue threshold (alternative to tier)
    
    -- Day-of-week matching
    use_day_of_week_revenue BOOLEAN NOT NULL DEFAULT true, -- Use branch_weekly_revenue as primary source
    use_specific_date_revenue BOOLEAN NOT NULL DEFAULT false, -- Use revenue_data as override
    
    -- Doctor count matching
    doctor_count INTEGER, -- NULL means any count, specific number means exact match
    min_doctor_count INTEGER, -- Minimum doctor count (alternative to exact match)
    
    -- Day of week filter (optional)
    day_of_week INTEGER CHECK (day_of_week >= 0 AND day_of_week <= 6), -- NULL means any day
    
    -- Scenario settings
    is_default BOOLEAN NOT NULL DEFAULT false,
    is_active BOOLEAN NOT NULL DEFAULT true,
    priority INTEGER NOT NULL DEFAULT 0, -- Higher priority scenarios checked first
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Ensure at least one matching criteria is specified
    CONSTRAINT check_revenue_criteria CHECK (
        revenue_level_tier_id IS NOT NULL OR 
        min_revenue IS NOT NULL OR 
        is_default = true
    )
);

CREATE INDEX idx_staff_requirement_scenarios_priority ON staff_requirement_scenarios(priority DESC, is_active);
CREATE INDEX idx_staff_requirement_scenarios_tier ON staff_requirement_scenarios(revenue_level_tier_id);
CREATE INDEX idx_staff_requirement_scenarios_day ON staff_requirement_scenarios(day_of_week);
```

**Example Data:**
```sql
-- Default scenario (no conditions)
INSERT INTO staff_requirement_scenarios (scenario_name, description, is_default, priority)
VALUES ('Normal Day', 'Standard staffing for normal operations', true, 0);

-- High revenue day-of-week scenario (Level 5)
INSERT INTO staff_requirement_scenarios (
    scenario_name, 
    description, 
    revenue_level_tier_id,
    use_day_of_week_revenue,
    use_specific_date_revenue,
    priority
)
VALUES (
    'High Revenue Day (Level 5)',
    'Increased staffing when day-of-week revenue is Level 5 (400K-500K)',
    (SELECT id FROM revenue_level_tiers WHERE level_number = 5),
    true,
    false,
    10
);

-- Very High revenue day-of-week scenario (Level 6)
INSERT INTO staff_requirement_scenarios (
    scenario_name, 
    description, 
    revenue_level_tier_id,
    use_day_of_week_revenue,
    use_specific_date_revenue,
    priority
)
VALUES (
    'Very High Revenue Day (Level 6)',
    'Maximum staffing when day-of-week revenue is Level 6 (500K-600K)',
    (SELECT id FROM revenue_level_tiers WHERE level_number = 6),
    true,
    false,
    20
);

-- Two doctors scenario
INSERT INTO staff_requirement_scenarios (
    scenario_name,
    description,
    doctor_count,
    priority
)
VALUES (
    'Two Doctors Day',
    'Additional staff when 2 doctors are scheduled',
    2,
    15
);

-- Combined scenario: High revenue + Two doctors
INSERT INTO staff_requirement_scenarios (
    scenario_name,
    description,
    revenue_level_tier_id,
    use_day_of_week_revenue,
    doctor_count,
    priority
)
VALUES (
    'High Revenue + Two Doctors',
    'Combined scenario for high revenue days with 2 doctors',
    (SELECT id FROM revenue_level_tiers WHERE level_number = 5),
    true,
    2,
    25
);

-- Specific date override scenario (uses revenue_data instead of branch_weekly_revenue)
INSERT INTO staff_requirement_scenarios (
    scenario_name,
    description,
    revenue_level_tier_id,
    use_day_of_week_revenue,
    use_specific_date_revenue,
    priority
)
VALUES (
    'Special Event Day',
    'Uses specific date revenue override (e.g., promotions, events)',
    (SELECT id FROM revenue_level_tiers WHERE level_number = 6),
    false,
    true,
    30
);
```

#### 1.3 Revenue Matching Logic Function

```sql
-- Helper function to get revenue level tier for a given revenue amount
CREATE OR REPLACE FUNCTION get_revenue_level_tier(revenue_amount DECIMAL(15,2))
RETURNS UUID AS $$
DECLARE
    tier_id UUID;
BEGIN
    SELECT id INTO tier_id
    FROM revenue_level_tiers
    WHERE revenue_amount >= min_revenue
      AND (max_revenue IS NULL OR revenue_amount < max_revenue)
    ORDER BY level_number DESC
    LIMIT 1;
    
    RETURN tier_id;
END;
$$ LANGUAGE plpgsql;

-- Helper function to check if scenario matches
CREATE OR REPLACE FUNCTION scenario_matches(
    p_scenario_id UUID,
    p_day_of_week_revenue DECIMAL(15,2),
    p_specific_date_revenue DECIMAL(15,2),
    p_doctor_count INTEGER,
    p_day_of_week INTEGER
)
RETURNS BOOLEAN AS $$
DECLARE
    v_scenario RECORD;
    v_revenue_to_check DECIMAL(15,2);
    v_revenue_tier_id UUID;
BEGIN
    -- Get scenario
    SELECT * INTO v_scenario
    FROM staff_requirement_scenarios
    WHERE id = p_scenario_id AND is_active = true;
    
    IF NOT FOUND THEN
        RETURN false;
    END IF;
    
    -- Check day of week filter
    IF v_scenario.day_of_week IS NOT NULL AND v_scenario.day_of_week != p_day_of_week THEN
        RETURN false;
    END IF;
    
    -- Determine which revenue to use
    IF v_scenario.use_specific_date_revenue AND p_specific_date_revenue IS NOT NULL THEN
        v_revenue_to_check := p_specific_date_revenue;
    ELSIF v_scenario.use_day_of_week_revenue THEN
        v_revenue_to_check := p_day_of_week_revenue;
    ELSE
        v_revenue_to_check := COALESCE(p_specific_date_revenue, p_day_of_week_revenue);
    END IF;
    
    -- Check revenue tier match
    IF v_scenario.revenue_level_tier_id IS NOT NULL THEN
        v_revenue_tier_id := get_revenue_level_tier(v_revenue_to_check);
        IF v_revenue_tier_id != v_scenario.revenue_level_tier_id THEN
            RETURN false;
        END IF;
    END IF;
    
    -- Check direct revenue range
    IF v_scenario.min_revenue IS NOT NULL THEN
        IF v_revenue_to_check < v_scenario.min_revenue THEN
            RETURN false;
        END IF;
    END IF;
    IF v_scenario.max_revenue IS NOT NULL THEN
        IF v_revenue_to_check >= v_revenue_to_check THEN
            RETURN false;
        END IF;
    END IF;
    
    -- Check doctor count
    IF v_scenario.doctor_count IS NOT NULL THEN
        IF p_doctor_count != v_scenario.doctor_count THEN
            RETURN false;
        END IF;
    END IF;
    IF v_scenario.min_doctor_count IS NOT NULL THEN
        IF p_doctor_count < v_scenario.min_doctor_count THEN
            RETURN false;
        END IF;
    END IF;
    
    RETURN true;
END;
$$ LANGUAGE plpgsql;
```

### Frontend Design

#### 1.1 Revenue Level Tiers Management Page

**Location:** `/admin/revenue-level-tiers`

**Features:**
- List all revenue level tiers
- Create/edit/delete tiers
- Visual color coding
- Range validation (no overlaps)
- Preview revenue ranges

**UI Layout:**
```
┌─────────────────────────────────────────────────────────┐
│ Revenue Level Tiers Management                          │
├─────────────────────────────────────────────────────────┤
│ [Sort: Level Number ▼]                                  │
├─────────────────────────────────────────────────────────┤
│ Level │ Name        │ Range (THB)      │ Color │ Actions │
├─────────────────────────────────────────────────────────┤
│   1   │ Very Low    │ 0 - 100,000     │  ⬜   │ [Edit]  │
│   2   │ Low         │ 100,000-200,000 │  ⬜   │ [Edit]  │
│   3   │ Medium      │ 200,000-300,000 │  ⬜   │ [Edit]  │
│   4   │ High        │ 300,000-400,000 │  ⬜   │ [Edit]  │
│   5   │ Very High   │ 400,000-500,000 │  ⬜   │ [Edit]  │
│   6   │ Extremely   │ 500,000-600,000 │  ⬜   │ [Edit]  │
│       │ High        │                 │       │         │
│   7   │ Peak        │ 600,000+        │  ⬜   │ [Edit]  │
├─────────────────────────────────────────────────────────┤
│ [+ Add Tier]                                            │
└─────────────────────────────────────────────────────────┘
```

#### 1.2 Enhanced Scenario Management Page

**Location:** `/admin/staff-requirement-scenarios`

**New Features:**
- Revenue matching: Select tier level OR specify direct range
- Day-of-week revenue toggle (primary source)
- Specific date revenue toggle (override)
- Day-of-week filter (optional)
- Doctor count matching (exact or minimum)

**UI Layout:**
```
┌─────────────────────────────────────────────────────────┐
│ Staff Requirement Scenarios                              │
├─────────────────────────────────────────────────────────┤
│ [Filter: Active Only] [Sort: Priority ▼]                 │
├─────────────────────────────────────────────────────────┤
│ Scenario Name │ Conditions │ Priority │ Status │ Actions │
├─────────────────────────────────────────────────────────┤
│ Normal Day    │ Default    │    0     │ Active │ [Edit]  │
│ High Rev L5   │ Tier 5     │   10     │ Active │ [Edit]  │
│               │ DoW Rev    │          │        │         │
│ Very High L6  │ Tier 6     │   20     │ Active │ [Edit]  │
│               │ DoW Rev    │          │        │         │
│ Two Doctors   │ Doctors=2  │   15     │ Active │ [Edit]  │
│ High+2 Docs   │ Tier 5     │   25     │ Active │ [Edit]  │
│               │ DoW Rev    │          │        │         │
│               │ Doctors=2  │          │        │         │
├─────────────────────────────────────────────────────────┤
│ [+ Create Scenario]                                      │
└─────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────┐
│ Create/Edit Scenario                                    │
├─────────────────────────────────────────────────────────┤
│ Scenario Name: [High Revenue Day (Level 5)        ]     │
│ Description: [Increased staffing for Level 5 revenue]   │
│                                                          │
│ Revenue Matching:                                        │
│   ○ Use Revenue Level Tier                              │
│     Tier: [Level 5 - Very High (400K-500K) ▼]          │
│   ○ Use Direct Revenue Range                            │
│     Min: [400000] Max: [500000]                         │
│                                                          │
│ Revenue Source:                                         │
│   ☑ Use Day-of-Week Revenue (from branch_weekly_revenue) │
│   ☐ Use Specific Date Revenue (from revenue_data, overrides DoW) │
│                                                          │
│ Day of Week Filter (Optional):                          │
│   [Any Day ▼]                                           │
│                                                          │
│ Doctor Count Matching:                                  │
│   ○ Any doctor count                                    │
│   ○ Exact count: [2]                                    │
│   ○ Minimum count: [2]                                  │
│                                                          │
│ Priority: [10]                                          │
│ Status: [☑ Active] [☐ Default]                          │
│                                                          │
│ Position Requirements:                                  │
│   Position          │ Preferred │ Minimum │ Override │
│   Front Staff       │    +2     │   +1    │   No     │
│   Assistant Manager │    +1     │    0    │   No     │
│   Doctor Assistant  │     0     │    0    │   No     │
│                                                          │
│ [Save] [Cancel]                                         │
└─────────────────────────────────────────────────────────┘
```

#### 1.3 Enhanced Branch Configuration Preview

**Location:** `/admin/branches/[branchId]/config`

**New Section:** Shows day-of-week revenue, tier level, and scenario matching

**UI Layout:**
```
┌─────────────────────────────────────────────────────────┐
│ Dynamic Staff Requirements Preview                      │
├─────────────────────────────────────────────────────────┤
│ Date: 2026-01-15 (Monday)                               │
│                                                          │
│ Revenue Information:                                    │
│   Day-of-Week Baseline: 450,000 THB                     │
│   Revenue Level: Level 5 - Very High (400K-500K)        │
│   Specific Date Override: 480,000 THB (if available)    │
│   Revenue Used for Matching: 450,000 THB (DoW baseline)  │
│                                                          │
│ Doctors Scheduled: 2                                    │
│                                                          │
│ Matching Scenarios:                                     │
│   ✅ High Revenue Day (Level 5) (Priority: 10)         │
│      Reason: Day-of-week revenue matches Tier 5         │
│   ✅ Two Doctors Day (Priority: 15)                     │
│      Reason: Doctor count = 2                            │
│   ✅ High Revenue + Two Doctors (Priority: 25)          │
│      Reason: Tier 5 AND Doctors = 2                     │
│                                                          │
│ Applied Scenario: High Revenue + Two Doctors (Priority: 25) │
│                                                          │
│ Calculated Requirements:                                 │
│   Position          │ Base │ Scenario │ Final │         │
│                     │ Pref │ Min │ Pref │ Min │ Pref │ Min │
│   Front Staff       │  3   │  2  │ +2  │ +1  │  5   │  3  │
│   Doctor Assistant  │  2   │  2  │ =2  │ =2  │  2   │  2  │
│   (Override base)   │      │     │     │     │      │     │
└─────────────────────────────────────────────────────────┘
```

### Calculation Logic

```go
// Enhanced calculation with day-of-week revenue and tiers
func CalculateStaffRequirementsByScenario(
    branchID string,
    date time.Time,
    positionID string,
    basePreferred int,
    baseMinimum int,
) (calculatedPreferred int, calculatedMinimum int, matchedScenarios []ScenarioMatch) {
    
    dayOfWeek := int(date.Weekday())
    
    // Get day-of-week baseline revenue
    dayOfWeekRevenue := GetDayOfWeekRevenue(branchID, dayOfWeek) // from branch_weekly_revenue
    
    // Get specific date revenue (if available)
    specificDateRevenue := GetSpecificDateRevenue(branchID, date) // from revenue_data
    
    // Get doctor count for the date
    doctorCount := GetDoctorCountForDate(branchID, date)
    
    // Get revenue level tier for day-of-week revenue
    revenueTierID := GetRevenueLevelTier(dayOfWeekRevenue)
    revenueTier := GetRevenueLevelTierByID(revenueTierID)
    
    // Find matching scenarios (ordered by priority DESC)
    scenarios := GetActiveScenariosOrderedByPriority()
    
    var applicableScenarios []Scenario
    for _, scenario := range scenarios {
        if MatchesScenarioEnhanced(
            scenario,
            dayOfWeekRevenue,
            specificDateRevenue,
            revenueTierID,
            doctorCount,
            dayOfWeek,
        ) {
            applicableScenarios = append(applicableScenarios, scenario)
            matchedScenarios = append(matchedScenarios, ScenarioMatch{
                ScenarioID: scenario.ID,
                ScenarioName: scenario.ScenarioName,
                MatchReason: BuildMatchReason(scenario, dayOfWeekRevenue, specificDateRevenue, doctorCount),
            })
        }
    }
    
    // Apply highest priority scenario
    if len(applicableScenarios) > 0 {
        highestPriorityScenario := applicableScenarios[0]
        requirement := GetScenarioRequirement(highestPriorityScenario.ID, positionID)
        
        if requirement != nil {
            if requirement.OverrideBase {
                calculatedPreferred = requirement.PreferredStaff
                calculatedMinimum = requirement.MinimumStaff
            } else {
                calculatedPreferred = basePreferred + requirement.PreferredStaff
                calculatedMinimum = baseMinimum + requirement.MinimumStaff
            }
        } else {
            calculatedPreferred = basePreferred
            calculatedMinimum = baseMinimum
        }
    } else {
        // Fallback to default scenario
        calculatedPreferred = basePreferred
        calculatedMinimum = baseMinimum
    }
    
    return calculatedPreferred, calculatedMinimum, matchedScenarios
}

func MatchesScenarioEnhanced(
    scenario Scenario,
    dayOfWeekRevenue decimal.Decimal,
    specificDateRevenue *decimal.Decimal,
    revenueTierID *uuid.UUID,
    doctorCount int,
    dayOfWeek int,
) bool {
    // Check day of week filter
    if scenario.DayOfWeek != nil && *scenario.DayOfWeek != dayOfWeek {
        return false
    }
    
    // Determine which revenue to use
    var revenueToCheck decimal.Decimal
    if scenario.UseSpecificDateRevenue && specificDateRevenue != nil {
        revenueToCheck = *specificDateRevenue
    } else if scenario.UseDayOfWeekRevenue {
        revenueToCheck = dayOfWeekRevenue
    } else {
        // Fallback: use specific date if available, otherwise day-of-week
        if specificDateRevenue != nil {
            revenueToCheck = *specificDateRevenue
        } else {
            revenueToCheck = dayOfWeekRevenue
        }
    }
    
    // Check revenue tier match
    if scenario.RevenueLevelTierID != nil {
        currentTierID := GetRevenueLevelTier(revenueToCheck)
        if currentTierID == nil || *currentTierID != *scenario.RevenueLevelTierID {
            return false
        }
    }
    
    // Check direct revenue range
    if scenario.MinRevenue != nil && revenueToCheck.LessThan(*scenario.MinRevenue) {
        return false
    }
    if scenario.MaxRevenue != nil && revenueToCheck.GreaterThanOrEqual(*scenario.MaxRevenue) {
        return false
    }
    
    // Check doctor count
    if scenario.DoctorCount != nil && doctorCount != *scenario.DoctorCount {
        return false
    }
    if scenario.MinDoctorCount != nil && doctorCount < *scenario.MinDoctorCount {
        return false
    }
    
    return true
}

func GetRevenueLevelTier(revenue decimal.Decimal) *uuid.UUID {
    // Query revenue_level_tiers table to find matching tier
    // Returns tier ID where revenue falls within min_revenue and max_revenue range
}
```

### Pros
- ✅ **Day-of-week baseline**: Uses existing `branch_weekly_revenue` table
- ✅ **Flexible tiers**: Configurable revenue level ranges
- ✅ **Dual revenue support**: Can use day-of-week OR specific date revenue
- ✅ **Clear hierarchy**: Tier-based matching is intuitive
- ✅ **Backward compatible**: Works with existing revenue data

### Cons
- ❌ **Tier management**: Requires maintaining tier definitions
- ❌ **Range gaps**: Need to ensure no gaps between tier ranges
- ❌ **Complexity**: More matching logic than simple thresholds

---

## Solution 2: Dynamic Revenue Level Calculation with Day-of-Week Weighting

### Concept

Instead of fixed tiers, calculate revenue level dynamically based on:
- Day-of-week baseline revenue (from `branch_weekly_revenue`)
- Percentage deviation from baseline
- Configurable level thresholds as percentages

### Database Design

#### 2.1 Simplified: `revenue_level_config` Table

```sql
CREATE TABLE IF NOT EXISTS revenue_level_config (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    level_number INTEGER NOT NULL UNIQUE CHECK (level_number >= 1 AND level_number <= 10),
    level_name VARCHAR(50) NOT NULL,
    min_percentage DECIMAL(5,2) NOT NULL, -- Percentage above baseline (e.g., 150% = 1.5x)
    max_percentage DECIMAL(5,2), -- NULL means no upper limit
    display_order INTEGER NOT NULL DEFAULT 0,
    color_code VARCHAR(20),
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**Example Data:**
```sql
INSERT INTO revenue_level_config (level_number, level_name, min_percentage, max_percentage, display_order) VALUES
    (1, 'Very Low', 0, 80, 1),      -- 0-80% of baseline
    (2, 'Low', 80, 100, 2),          -- 80-100% of baseline
    (3, 'Normal', 100, 120, 3),      -- 100-120% of baseline
    (4, 'High', 120, 150, 4),        -- 120-150% of baseline
    (5, 'Very High', 150, 200, 5),   -- 150-200% of baseline
    (6, 'Extremely High', 200, NULL, 6); -- 200%+ of baseline
```

#### 2.2 Scenarios Reference Percentage-Based Levels

Scenarios would match based on percentage of day-of-week baseline, making it branch-agnostic.

### Pros
- ✅ **Branch-agnostic**: Same percentage thresholds work for all branches
- ✅ **Adaptive**: Automatically adjusts to each branch's baseline
- ✅ **Simpler**: No need to define absolute revenue ranges

### Cons
- ❌ **Less intuitive**: Percentage-based is harder to understand than absolute values
- ❌ **Baseline dependency**: Requires accurate day-of-week revenue data
- ❌ **Complexity**: Calculation requires baseline lookup

---

## Solution 3: Hybrid Approach (Recommended Enhancement)

### Concept

Combine Solutions 1 and 2:
- **Fixed revenue tiers** for absolute matching (e.g., Level 5 = 400K-500K)
- **Percentage-based levels** as an alternative matching method
- Scenarios can use either method

### Database Design

#### 3.1 Enhanced: `revenue_level_tiers` with Both Methods

```sql
CREATE TABLE IF NOT EXISTS revenue_level_tiers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    level_number INTEGER NOT NULL UNIQUE CHECK (level_number >= 1 AND level_number <= 10),
    level_name VARCHAR(50) NOT NULL,
    
    -- Absolute revenue matching
    min_revenue DECIMAL(15,2) CHECK (min_revenue >= 0),
    max_revenue DECIMAL(15,2),
    
    -- Percentage-based matching (alternative)
    min_percentage DECIMAL(5,2) CHECK (min_percentage >= 0), -- Percentage of day-of-week baseline
    max_percentage DECIMAL(5,2),
    
    -- Matching method preference
    use_absolute BOOLEAN NOT NULL DEFAULT true, -- If true, use absolute; if false, use percentage
    use_percentage BOOLEAN NOT NULL DEFAULT false,
    
    display_order INTEGER NOT NULL DEFAULT 0,
    color_code VARCHAR(20),
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Ensure at least one matching method is specified
    CONSTRAINT check_matching_method CHECK (
        (use_absolute AND (min_revenue IS NOT NULL)) OR
        (use_percentage AND (min_percentage IS NOT NULL))
    )
);
```

**Example Data:**
```sql
-- Absolute-based tiers
INSERT INTO revenue_level_tiers (level_number, level_name, min_revenue, max_revenue, use_absolute, display_order) VALUES
    (5, 'Very High', 400000, 500000, true, 5),
    (6, 'Extremely High', 500000, 600000, true, 6);

-- Percentage-based tiers (for branches with different baselines)
INSERT INTO revenue_level_tiers (level_number, level_name, min_percentage, max_percentage, use_percentage, display_order) VALUES
    (5, 'Very High', 150, 200, false, 5), -- 150-200% of baseline
    (6, 'Extremely High', 200, NULL, false, 6); -- 200%+ of baseline
```

### Pros
- ✅ **Best of both worlds**: Absolute for consistency, percentage for flexibility
- ✅ **Flexible**: Can choose matching method per tier
- ✅ **Comprehensive**: Handles both high-revenue and low-revenue branches

### Cons
- ❌ **Complexity**: More complex to implement and understand
- ❌ **Maintenance**: Need to maintain both absolute and percentage tiers

---

## Recommendation

**I recommend Solution 1 (Revenue Level Tiers with Day-of-Week Baseline)** because:

1. **Intuitive**: Absolute revenue ranges (400K-500K) are easy to understand
2. **Day-of-week integration**: Uses existing `branch_weekly_revenue` table naturally
3. **Flexible**: Supports both day-of-week baseline and specific date overrides
4. **Clear matching**: Tier-based matching is straightforward
5. **Business-friendly**: Managers can easily understand "Level 5 revenue = 400K-500K"

### Implementation Priority

1. **Phase 1**: Revenue level tiers table and management UI
2. **Phase 2**: Enhanced scenarios with tier matching and day-of-week revenue
3. **Phase 3**: Scenario matching logic with dual revenue source support
4. **Phase 4**: Frontend preview and visualization
5. **Phase 5**: Integration with staff allocation calculations

---

## Questions for Clarification

1. **Tier Range Overlap**: Should tier ranges be exclusive (400K-500K, 500K-600K) or inclusive (400K-500K, 500K-600K with 500K matching higher tier)?

2. **Default Tier**: Should there be a default tier for revenue that doesn't match any tier, or should we ensure all revenue ranges are covered?

3. **Tier Updates**: If tier ranges are updated, should existing scenarios be automatically updated or require manual review?

4. **Multiple Tier Matching**: If a scenario matches multiple tiers (e.g., both Level 5 and Level 6), should we use the highest matching tier or require explicit scenario definition?

5. **Day-of-Week Revenue Updates**: How frequently should day-of-week revenue be updated? Should it be manual or calculated from historical data?

6. **Specific Date Override Priority**: When both day-of-week and specific date revenue exist, should specific date always override, or only when `use_specific_date_revenue` is true?

---

## Related Files

- **Database Schema**: `backend/internal/repositories/postgres/migrations.go`
- **Branch Weekly Revenue**: `backend/internal/repositories/postgres/branch_weekly_revenue_repo.go`
- **Domain Models**: `backend/internal/domain/models/branch_weekly_revenue.go`
- **Branch Config**: `frontend/src/components/branch/BranchPositionQuotaConfig.tsx`

---
title: Staff Requirement Factors Design - Revenue Level & Doctor Count
description: Design alternatives for linking revenue level and doctor count to preferred/minimum staff requirements
version: 1.0.0
lastUpdated: 2026-01-13 19:30:00
---

# Staff Requirement Factors Design

## Overview

This document presents two alternative design approaches for linking **revenue level** and **number of doctors** to preferred and minimum staff requirements. These factors will dynamically adjust staff requirements based on operational conditions.

**Requirements:**
1. **Revenue Level**: Significantly higher revenue requires more front and assistant staff
2. **Doctor Count**: Number of doctors on a given day (e.g., 2 doctors require at least 2 doctor assistants)

---

## Current System Context

### Existing Tables
- `position_quotas`: Stores `designated_quota` (preferred) and `minimum_required` (minimum) per branch/position
- `branch_constraints`: Stores daily constraints (min_front_staff, min_managers, min_total_staff) per day of week
- `doctor_assignments`: Tracks doctors assigned to branches on specific dates
- `revenue_data`: Stores expected/actual revenue per branch per date
- `branch_weekly_revenue`: Stores expected revenue per day of week

### Current Staff Calculation
- Base requirements come from `position_quotas` table
- Daily constraints from `branch_constraints` table
- No dynamic adjustment based on revenue level or doctor count

---

## Alternative 1: Rule-Based Multiplier System

### Database Design

#### 1.1 New Table: `staff_requirement_rules`

```sql
CREATE TABLE IF NOT EXISTS staff_requirement_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    position_id UUID NOT NULL REFERENCES positions(id),
    factor_type VARCHAR(50) NOT NULL CHECK (factor_type IN ('revenue_level', 'doctor_count')),
    factor_value DECIMAL(15,2) NOT NULL, -- Revenue threshold or doctor count threshold
    preferred_multiplier DECIMAL(5,4) NOT NULL DEFAULT 1.0 CHECK (preferred_multiplier >= 0),
    minimum_multiplier DECIMAL(5,4) NOT NULL DEFAULT 1.0 CHECK (minimum_multiplier >= 0),
    comparison_operator VARCHAR(10) NOT NULL DEFAULT '>=' CHECK (comparison_operator IN ('>=', '>', '=', '<=', '<')),
    is_active BOOLEAN NOT NULL DEFAULT true,
    priority INTEGER NOT NULL DEFAULT 0, -- Higher priority rules applied first
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_staff_requirement_rules_position ON staff_requirement_rules(position_id);
CREATE INDEX idx_staff_requirement_rules_factor ON staff_requirement_rules(factor_type, is_active);
```

**Example Data:**
```sql
-- Revenue level rule: If revenue >= 100,000, increase Front staff preferred by 1.5x
INSERT INTO staff_requirement_rules (position_id, factor_type, factor_value, preferred_multiplier, minimum_multiplier, comparison_operator, priority)
VALUES (
    (SELECT id FROM positions WHERE name LIKE '%Front%' LIMIT 1),
    'revenue_level',
    100000.00,
    1.5,
    1.2,
    '>=',
    10
);

-- Doctor count rule: If doctors >= 2, minimum doctor assistants = doctor count
INSERT INTO staff_requirement_rules (position_id, factor_type, factor_value, preferred_multiplier, minimum_multiplier, comparison_operator, priority)
VALUES (
    (SELECT id FROM positions WHERE name LIKE '%Doctor Assistant%' LIMIT 1),
    'doctor_count',
    2.0,
    1.0,
    1.0, -- Will be overridden by absolute value
    '>=',
    20
);
```

#### 1.2 New Table: `staff_requirement_absolute_rules`

For cases where absolute values are needed (e.g., doctor assistants = doctor count):

```sql
CREATE TABLE IF NOT EXISTS staff_requirement_absolute_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    position_id UUID NOT NULL REFERENCES positions(id),
    factor_type VARCHAR(50) NOT NULL CHECK (factor_type IN ('doctor_count')),
    calculation_type VARCHAR(50) NOT NULL CHECK (calculation_type IN ('equals', 'multiply', 'add')),
    calculation_value DECIMAL(10,2) NOT NULL DEFAULT 1.0, -- Multiplier or addend
    applies_to VARCHAR(20) NOT NULL DEFAULT 'minimum' CHECK (applies_to IN ('preferred', 'minimum', 'both')),
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(position_id, factor_type, applies_to)
);

CREATE INDEX idx_staff_requirement_absolute_rules_position ON staff_requirement_absolute_rules(position_id);
```

**Example Data:**
```sql
-- Doctor assistants minimum = number of doctors
INSERT INTO staff_requirement_absolute_rules (position_id, factor_type, calculation_type, calculation_value, applies_to)
VALUES (
    (SELECT id FROM positions WHERE name LIKE '%Doctor Assistant%' LIMIT 1),
    'doctor_count',
    'equals',
    1.0,
    'minimum'
);
```

#### 1.3 Enhanced: `revenue_levels` Lookup Table (Optional)

For standardized revenue level categorization:

```sql
CREATE TABLE IF NOT EXISTS revenue_levels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    level_name VARCHAR(50) NOT NULL UNIQUE, -- e.g., 'Low', 'Medium', 'High', 'Very High'
    min_revenue DECIMAL(15,2) NOT NULL,
    max_revenue DECIMAL(15,2), -- NULL means no upper limit
    display_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Example data
INSERT INTO revenue_levels (level_name, min_revenue, max_revenue, display_order) VALUES
    ('Low', 0, 50000, 1),
    ('Medium', 50000, 100000, 2),
    ('High', 100000, 200000, 3),
    ('Very High', 200000, NULL, 4);
```

### Frontend Design

#### 1.1 Staff Requirement Rules Management Page

**Location:** `/admin/staff-requirement-rules`

**Features:**
- List all rules grouped by position
- Filter by position, factor type (revenue/doctor count)
- Create/edit/delete rules
- Enable/disable rules
- Set priority order
- Preview calculated requirements based on sample values

**UI Components:**

```typescript
// Rule Configuration Form
interface StaffRequirementRule {
  id?: string;
  position_id: string;
  position_name: string;
  factor_type: 'revenue_level' | 'doctor_count';
  factor_value: number;
  preferred_multiplier: number;
  minimum_multiplier: number;
  comparison_operator: '>=' | '>' | '=' | '<=' | '<';
  priority: number;
  is_active: boolean;
}

// Absolute Rule Configuration Form
interface StaffRequirementAbsoluteRule {
  id?: string;
  position_id: string;
  position_name: string;
  factor_type: 'doctor_count';
  calculation_type: 'equals' | 'multiply' | 'add';
  calculation_value: number;
  applies_to: 'preferred' | 'minimum' | 'both';
  is_active: boolean;
}
```

**UI Layout:**
```
┌─────────────────────────────────────────────────────────┐
│ Staff Requirement Rules Management                      │
├─────────────────────────────────────────────────────────┤
│ [Filter: Position ▼] [Filter: Factor Type ▼] [Active] │
├─────────────────────────────────────────────────────────┤
│ Position          │ Factor │ Condition │ Preferred │ Min │ Priority │
├─────────────────────────────────────────────────────────┤
│ Front Staff       │ Revenue│ >= 100K   │ 1.5x      │1.2x│    10    │ [Edit] [Disable] │
│ Doctor Assistant  │ Doctors│ >= 2      │ 1.0x      │1.0x│    20    │ [Edit] [Disable] │
│ Front Staff       │ Revenue│ >= 200K   │ 2.0x      │1.5x│    15    │ [Edit] [Disable] │
├─────────────────────────────────────────────────────────┤
│ [+ Add Rule]                                            │
└─────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────┐
│ Absolute Rules (Doctor Count)                           │
├─────────────────────────────────────────────────────────┤
│ Position          │ Calculation │ Applies To │ Status │
├─────────────────────────────────────────────────────────┤
│ Doctor Assistant  │ equals      │ minimum    │ Active │ [Edit] │
├─────────────────────────────────────────────────────────┤
│ [+ Add Absolute Rule]                                    │
└─────────────────────────────────────────────────────────┘
```

#### 1.2 Enhanced Branch Configuration Page

**Location:** `/admin/branches/[branchId]/config`

**New Section:** "Dynamic Staff Requirements Preview"

Shows calculated requirements based on:
- Current revenue level (from `revenue_data` or `branch_weekly_revenue`)
- Number of doctors scheduled (from `doctor_assignments`)

**UI Component:**

```typescript
// Preview Component
interface CalculatedRequirement {
  position_id: string;
  position_name: string;
  base_preferred: number;
  base_minimum: number;
  calculated_preferred: number;
  calculated_minimum: number;
  factors_applied: string[]; // e.g., ["Revenue >= 100K", "Doctors = 2"]
}
```

**UI Layout:**
```
┌─────────────────────────────────────────────────────────┐
│ Dynamic Staff Requirements Preview                      │
├─────────────────────────────────────────────────────────┤
│ Revenue Level: High (150,000 THB)                       │
│ Doctors Scheduled: 2                                     │
├─────────────────────────────────────────────────────────┤
│ Position          │ Base │ Calculated │ Factors Applied │
│                   │ Pref │ Min │ Pref │ Min │           │
├─────────────────────────────────────────────────────────┤
│ Front Staff       │  3   │  2  │  4.5 │  2.4 │ Revenue>=100K │
│ Doctor Assistant  │  2   │  2  │  2   │  2   │ Doctors=2     │
│ Assistant Manager │  1   │  0  │  1   │  0   │ -             │
└─────────────────────────────────────────────────────────┘
```

### Calculation Logic

```go
// Pseudo-code for calculation
func CalculateStaffRequirements(
    branchID string,
    date time.Time,
    positionID string,
    basePreferred int,
    baseMinimum int,
) (calculatedPreferred int, calculatedMinimum int, factors []string) {
    
    // Get revenue for the date
    revenue := GetRevenueForDate(branchID, date)
    
    // Get doctor count for the date
    doctorCount := GetDoctorCountForDate(branchID, date)
    
    calculatedPreferred = basePreferred
    calculatedMinimum = baseMinimum
    
    // Apply multiplier rules (ordered by priority)
    rules := GetActiveRulesForPosition(positionID, "revenue_level")
    for _, rule := range rules {
        if EvaluateCondition(revenue, rule.comparison_operator, rule.factor_value) {
            calculatedPreferred = int(float64(calculatedPreferred) * rule.preferred_multiplier)
            calculatedMinimum = int(float64(calculatedMinimum) * rule.minimum_multiplier)
            factors = append(factors, fmt.Sprintf("Revenue %s %.0f", rule.comparison_operator, rule.factor_value))
        }
    }
    
    // Apply absolute rules (e.g., doctor assistants = doctor count)
    absoluteRules := GetActiveAbsoluteRulesForPosition(positionID, "doctor_count")
    for _, rule := range absoluteRules {
        if rule.calculation_type == "equals" {
            if rule.applies_to == "minimum" || rule.applies_to == "both" {
                calculatedMinimum = int(doctorCount * rule.calculation_value)
            }
            if rule.applies_to == "preferred" || rule.applies_to == "both" {
                calculatedPreferred = int(doctorCount * rule.calculation_value)
            }
            factors = append(factors, fmt.Sprintf("Doctors=%d", doctorCount))
        }
    }
    
    return calculatedPreferred, calculatedMinimum, factors
}
```

### Pros
- ✅ Flexible: Can create multiple rules per position
- ✅ Priority-based: Rules can override each other
- ✅ Reusable: Same rules apply across all branches
- ✅ Scalable: Easy to add new factor types
- ✅ Transparent: Clear audit trail of which rules applied

### Cons
- ❌ More complex: Requires understanding of multiplier system
- ❌ Potential conflicts: Multiple rules might conflict
- ❌ Performance: Requires rule evaluation for each calculation

---

## Alternative 2: Scenario-Based Configuration System

### Database Design

#### 2.1 New Table: `staff_requirement_scenarios`

```sql
CREATE TABLE IF NOT EXISTS staff_requirement_scenarios (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    scenario_name VARCHAR(100) NOT NULL,
    description TEXT,
    revenue_level VARCHAR(50), -- NULL or specific level: 'low', 'medium', 'high', 'very_high'
    min_revenue DECIMAL(15,2), -- NULL or minimum revenue threshold
    max_revenue DECIMAL(15,2), -- NULL or maximum revenue threshold
    doctor_count INTEGER, -- NULL or specific doctor count
    is_default BOOLEAN NOT NULL DEFAULT false,
    is_active BOOLEAN NOT NULL DEFAULT true,
    priority INTEGER NOT NULL DEFAULT 0, -- Higher priority scenarios checked first
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_staff_requirement_scenarios_priority ON staff_requirement_scenarios(priority DESC, is_active);
```

**Example Data:**
```sql
INSERT INTO staff_requirement_scenarios (scenario_name, description, revenue_level, min_revenue, doctor_count, priority, is_default) VALUES
    ('Normal Day', 'Standard staffing for normal operations', NULL, NULL, NULL, 0, true),
    ('High Revenue Day', 'Increased staffing for high revenue days', 'high', 100000, NULL, 10, false),
    ('Very High Revenue Day', 'Maximum staffing for very high revenue', 'very_high', 200000, NULL, 20, false),
    ('Two Doctors Day', 'Additional staff when 2 doctors are scheduled', NULL, NULL, 2, 15, false),
    ('High Revenue + Two Doctors', 'Combined scenario', 'high', 100000, 2, 25, false);
```

#### 2.2 New Table: `scenario_position_requirements`

```sql
CREATE TABLE IF NOT EXISTS scenario_position_requirements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    scenario_id UUID NOT NULL REFERENCES staff_requirement_scenarios(id) ON DELETE CASCADE,
    position_id UUID NOT NULL REFERENCES positions(id),
    preferred_staff INTEGER NOT NULL DEFAULT 0 CHECK (preferred_staff >= 0),
    minimum_staff INTEGER NOT NULL DEFAULT 0 CHECK (minimum_staff >= 0),
    override_base BOOLEAN NOT NULL DEFAULT false, -- If true, replaces base quota; if false, adds to base
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(scenario_id, position_id)
);

CREATE INDEX idx_scenario_position_requirements_scenario ON scenario_position_requirements(scenario_id);
CREATE INDEX idx_scenario_position_requirements_position ON scenario_position_requirements(position_id);
```

**Example Data:**
```sql
-- High Revenue Day scenario
INSERT INTO scenario_position_requirements (scenario_id, position_id, preferred_staff, minimum_staff, override_base)
SELECT 
    (SELECT id FROM staff_requirement_scenarios WHERE scenario_name = 'High Revenue Day'),
    p.id,
    CASE 
        WHEN p.name LIKE '%Front%' THEN 2  -- Add 2 more front staff
        WHEN p.name LIKE '%Assistant%' THEN 1  -- Add 1 more assistant
        ELSE 0
    END,
    CASE 
        WHEN p.name LIKE '%Front%' THEN 1  -- Add 1 minimum front staff
        ELSE 0
    END,
    false  -- Add to base, don't override
FROM positions p
WHERE p.position_type = 'branch';

-- Two Doctors Day scenario
INSERT INTO scenario_position_requirements (scenario_id, position_id, preferred_staff, minimum_staff, override_base)
SELECT 
    (SELECT id FROM staff_requirement_scenarios WHERE scenario_name = 'Two Doctors Day'),
    p.id,
    CASE 
        WHEN p.name LIKE '%Doctor Assistant%' THEN 2  -- Set preferred to 2
        ELSE 0
    END,
    CASE 
        WHEN p.name LIKE '%Doctor Assistant%' THEN 2  -- Set minimum to 2 (equals doctor count)
        ELSE 0
    END,
    true  -- Override base for doctor assistants
FROM positions p
WHERE p.name LIKE '%Doctor Assistant%';
```

#### 2.3 Enhanced: `revenue_levels` Lookup Table

Same as Alternative 1, for standardized revenue categorization.

### Frontend Design

#### 2.1 Scenario Management Page

**Location:** `/admin/staff-requirement-scenarios`

**Features:**
- List all scenarios with conditions
- Create/edit/delete scenarios
- Set priority order
- Enable/disable scenarios
- Configure position requirements per scenario
- Preview which scenarios would match given conditions

**UI Components:**

```typescript
interface StaffRequirementScenario {
  id?: string;
  scenario_name: string;
  description: string;
  revenue_level: 'low' | 'medium' | 'high' | 'very_high' | null;
  min_revenue: number | null;
  max_revenue: number | null;
  doctor_count: number | null;
  is_default: boolean;
  is_active: boolean;
  priority: number;
  position_requirements: ScenarioPositionRequirement[];
}

interface ScenarioPositionRequirement {
  position_id: string;
  position_name: string;
  preferred_staff: number;
  minimum_staff: number;
  override_base: boolean;
}
```

**UI Layout:**
```
┌─────────────────────────────────────────────────────────┐
│ Staff Requirement Scenarios                             │
├─────────────────────────────────────────────────────────┤
│ [Filter: Active Only] [Sort: Priority ▼]                │
├─────────────────────────────────────────────────────────┤
│ Scenario Name │ Conditions │ Priority │ Status │ Actions │
├─────────────────────────────────────────────────────────┤
│ Normal Day    │ Default    │    0     │ Active │ [Edit] [View] │
│ High Revenue  │ Rev>=100K  │   10     │ Active │ [Edit] [View] │
│ Two Doctors   │ Doctors=2  │   15     │ Active │ [Edit] [View] │
│ High+2 Docs   │ Rev>=100K  │   25     │ Active │ [Edit] [View] │
│               │ Doctors=2  │          │        │                │
├─────────────────────────────────────────────────────────┤
│ [+ Create Scenario]                                      │
└─────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────┐
│ Scenario: High Revenue Day                              │
├─────────────────────────────────────────────────────────┤
│ Conditions:                                             │
│   Revenue Level: [High ▼]                               │
│   Min Revenue: [100000]                                 │
│   Doctor Count: [Any ▼]                                 │
│   Priority: [10]                                        │
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

#### 2.2 Scenario Matching Preview

**Location:** `/admin/branches/[branchId]/config` (new section)

Shows which scenarios match current conditions and resulting requirements.

**UI Component:**

```typescript
interface ScenarioMatch {
  scenario_id: string;
  scenario_name: string;
  matches: boolean;
  match_reason: string;
  position_requirements: ScenarioPositionRequirement[];
}
```

**UI Layout:**
```
┌─────────────────────────────────────────────────────────┐
│ Scenario Matching Preview                               │
├─────────────────────────────────────────────────────────┤
│ Current Conditions:                                     │
│   Date: 2026-01-15 (Monday)                             │
│   Expected Revenue: 150,000 THB                         │
│   Doctors Scheduled: 2                                  │
├─────────────────────────────────────────────────────────┤
│ Matching Scenarios:                                     │
│   ✅ High Revenue Day (Priority: 10)                    │
│      Reason: Revenue >= 100,000                         │
│   ✅ Two Doctors Day (Priority: 15)                     │
│      Reason: Doctor count = 2                           │
│   ✅ High Revenue + Two Doctors (Priority: 25)         │
│      Reason: Revenue >= 100,000 AND Doctors = 2         │
├─────────────────────────────────────────────────────────┤
│ Calculated Requirements (Highest Priority Applied):     │
│   Position          │ Base │ Scenario │ Final │         │
│                     │ Pref │ Min │ Pref │ Min │ Pref │ Min │
│   Front Staff       │  3   │  2  │ +2  │ +1  │  5   │  3  │
│   Doctor Assistant  │  2   │  2  │ =2  │ =2  │  2   │  2  │
│   (Override base)   │      │     │     │     │      │     │
└─────────────────────────────────────────────────────────┘
```

### Calculation Logic

```go
// Pseudo-code for scenario-based calculation
func CalculateStaffRequirementsByScenario(
    branchID string,
    date time.Time,
    positionID string,
    basePreferred int,
    baseMinimum int,
) (calculatedPreferred int, calculatedMinimum int, matchedScenarios []string) {
    
    // Get revenue for the date
    revenue := GetRevenueForDate(branchID, date)
    revenueLevel := GetRevenueLevel(revenue) // e.g., "high"
    
    // Get doctor count for the date
    doctorCount := GetDoctorCountForDate(branchID, date)
    
    // Find matching scenarios (ordered by priority DESC)
    scenarios := GetActiveScenariosOrderedByPriority()
    
    var applicableScenarios []Scenario
    for _, scenario := range scenarios {
        if MatchesScenario(scenario, revenue, revenueLevel, doctorCount) {
            applicableScenarios = append(applicableScenarios, scenario)
            matchedScenarios = append(matchedScenarios, scenario.scenario_name)
        }
    }
    
    // Apply highest priority scenario only (or combine if needed)
    if len(applicableScenarios) > 0 {
        highestPriorityScenario := applicableScenarios[0]
        requirement := GetScenarioRequirement(highestPriorityScenario.id, positionID)
        
        if requirement != nil {
            if requirement.override_base {
                calculatedPreferred = requirement.preferred_staff
                calculatedMinimum = requirement.minimum_staff
            } else {
                calculatedPreferred = basePreferred + requirement.preferred_staff
                calculatedMinimum = baseMinimum + requirement.minimum_staff
            }
        } else {
            calculatedPreferred = basePreferred
            calculatedMinimum = baseMinimum
        }
    } else {
        calculatedPreferred = basePreferred
        calculatedMinimum = baseMinimum
    }
    
    return calculatedPreferred, calculatedMinimum, matchedScenarios
}

func MatchesScenario(scenario Scenario, revenue decimal.Decimal, revenueLevel string, doctorCount int) bool {
    // Check revenue level
    if scenario.revenue_level != nil && scenario.revenue_level != revenueLevel {
        return false
    }
    
    // Check revenue range
    if scenario.min_revenue != nil && revenue < scenario.min_revenue {
        return false
    }
    if scenario.max_revenue != nil && revenue > scenario.max_revenue {
        return false
    }
    
    // Check doctor count
    if scenario.doctor_count != nil && doctorCount != scenario.doctor_count {
        return false
    }
    
    return true
}
```

### Pros
- ✅ Intuitive: Scenarios are easy to understand and configure
- ✅ Business-friendly: Matches how managers think about staffing
- ✅ Flexible combinations: Can handle complex conditions
- ✅ Clear precedence: Priority-based matching
- ✅ Override capability: Can replace base requirements or add to them

### Cons
- ❌ More data entry: Need to configure each scenario separately
- ❌ Potential gaps: Scenarios might not cover all cases
- ❌ Less granular: Harder to fine-tune individual multipliers

---

## Comparison Summary

| Aspect | Alternative 1: Rule-Based Multiplier | Alternative 2: Scenario-Based |
|--------|--------------------------------------|-------------------------------|
| **Complexity** | Medium-High | Low-Medium |
| **Flexibility** | Very High | High |
| **Ease of Use** | Requires technical understanding | Business-friendly |
| **Granularity** | Fine-grained control | Coarser-grained |
| **Performance** | Requires rule evaluation | Direct scenario lookup |
| **Maintainability** | Rules can conflict | Clear scenario definitions |
| **Best For** | Complex, dynamic requirements | Predefined operational scenarios |

---

## Recommendation

**For this use case, I recommend Alternative 2 (Scenario-Based)** because:

1. **Business Alignment**: Scenarios match how clinic managers think about staffing ("High revenue day", "Two doctors day")
2. **Simplicity**: Easier for non-technical users to configure and understand
3. **Clear Precedence**: Priority-based matching is straightforward
4. **Maintainability**: Less risk of conflicting rules
5. **Performance**: Direct scenario matching is faster than rule evaluation

However, **Alternative 1 (Rule-Based)** would be better if:
- You need very fine-grained control over multipliers
- Requirements change frequently and need programmatic updates
- You want to support complex mathematical formulas

---

## Implementation Steps

### Phase 1: Database Setup
1. Create new tables (`staff_requirement_scenarios`, `scenario_position_requirements`, `revenue_levels`)
2. Create database migrations
3. Seed initial scenarios (Normal Day, High Revenue Day, Two Doctors Day)

### Phase 2: Backend API
1. Create domain models for scenarios
2. Create repository layer for scenario CRUD
3. Create use case layer for scenario matching and requirement calculation
4. Create API handlers for scenario management
5. Integrate scenario calculation into existing staff requirement endpoints

### Phase 3: Frontend UI
1. Create scenario management page (`/admin/staff-requirement-scenarios`)
2. Enhance branch configuration page with scenario preview
3. Add scenario matching visualization
4. Add scenario testing/preview functionality

### Phase 4: Integration & Testing
1. Integrate scenario calculation into staff allocation logic
2. Update allocation suggestions to use scenario-based requirements
3. Add unit tests for scenario matching logic
4. Add E2E tests for scenario management UI

---

## Related Files

- **Database Schema**: `backend/internal/repositories/postgres/migrations.go`
- **Domain Models**: `backend/internal/domain/models/`
- **API Handlers**: `backend/internal/handlers/`
- **Frontend Components**: `frontend/src/components/`
- **Branch Config**: `frontend/src/components/branch/BranchPositionQuotaConfig.tsx`

---

## Questions for Clarification

1. **Revenue Level Calculation**: Should revenue level be calculated per day, or per day-of-week average?
2. **Scenario Combination**: Should multiple scenarios combine (additive) or should only the highest priority apply?
3. **Doctor Count Source**: Should doctor count come from `doctor_assignments` table or a separate scheduling system?
4. **Historical Data**: Should scenarios apply retroactively to historical dates, or only going forward?
5. **Notification**: Should the system alert when scenarios don't match any conditions (fallback to default)?

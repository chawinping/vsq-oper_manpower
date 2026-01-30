# Allocation Criteria Priority-Ranked System - Three Alternative Designs

## Overview

The current allocation criteria system uses a **weighted scoring approach** where all criteria are evaluated, scored (0.0-1.0), and combined using configurable weights to produce a final priority score.

The user wants to change this to a **priority-ranked system** where criteria are ordered by priority, and the system determines which branch-position combinations should get highest priority based on this ranking.

---

## Current System (Weighted)

**How it works:**
- Each of 5 criteria groups is evaluated and returns a score (0.0-1.0)
- Scores are combined using weights: `(score1*weight1 + score2*weight2 + ...) / totalWeight`
- Final priority score determines ranking

**Limitations:**
- Complex to understand weight relationships
- Hard to ensure critical criteria (e.g., minimum staff) always take precedence
- Weight adjustments can have unexpected effects

---

## Alternative 1: Strict Priority Ranking (Lexicographic Ordering)

### Design Philosophy
Criteria are evaluated in strict priority order. The first criterion that differentiates between candidates determines the ranking. Only if candidates are equal on the first criterion do we consider the second, and so on.

### How It Works

1. **Priority Order Configuration**: Admin sets the priority order (1st, 2nd, 3rd, 4th, 5th)
   - Priority 1 = Highest priority (evaluated first)
   - Priority 5 = Lowest priority (evaluated last)

2. **Evaluation Process**:
   ```
   For each branch-position combination:
     Evaluate Priority 1 criterion → Get score
     If scores differ → Rank by Priority 1 score
     Else:
       Evaluate Priority 2 criterion → Get score
       If scores differ → Rank by Priority 2 score
       Else:
         Evaluate Priority 3 criterion → ...
   ```

3. **Ranking Logic**:
   - Sort all suggestions by Priority 1 score (descending)
   - For ties, sort by Priority 2 score (descending)
   - Continue until all criteria are considered or no more ties

### Example Scenario

**Priority Order:**
1. Priority 1: Minimum Staff Shortage (Third Criteria)
2. Priority 2: Preferred Staff Shortage (Second Criteria)
3. Priority 3: Branch-Level Variables (First Criteria)
4. Priority 4: Branch Type Staff Groups (Fourth Criteria)
5. Priority 5: Doctor Preferences (Zeroth Criteria)

**Branch-Position Combinations:**
- Branch A, Nurse: Min shortage = 2, Pref shortage = 3, Revenue = 0.8
- Branch B, Nurse: Min shortage = 0, Pref shortage = 4, Revenue = 0.9
- Branch C, Nurse: Min shortage = 1, Pref shortage = 2, Revenue = 0.7

**Ranking Result:**
1. Branch A (Min shortage = 2, highest)
2. Branch C (Min shortage = 1, second)
3. Branch B (Min shortage = 0, but Pref shortage = 4, highest)

### Configuration UI

```
┌─────────────────────────────────────────────────────────────┐
│  Allocation Criteria Priority Configuration                 │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  Drag and drop to reorder priorities (highest to lowest):   │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │ [⋮⋮] 🚨 Minimum Staff Shortage        [Priority 1]  │  │
│  │ [⋮⋮] ⭐ Preferred Staff Shortage      [Priority 2]  │  │
│  │ [⋮⋮] 📊 Branch-Level Variables        [Priority 3]  │  │
│  │ [⋮⋮] 🏢 Branch Type Staff Groups      [Priority 4]  │  │
│  │ [⋮⋮] 👨‍⚕️ Doctor Preferences (Optional) [Priority 5]  │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                              │
│  ☑ Enable Doctor Preferences                                │
│                                                              │
│  [Reset to Defaults]              [Save Configuration]     │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### Advantages
- ✅ **Clear and intuitive**: Easy to understand "what matters most"
- ✅ **Guaranteed precedence**: Critical criteria (minimum staff) always evaluated first
- ✅ **Simple configuration**: Just reorder priorities
- ✅ **Predictable**: No complex weight interactions
- ✅ **Easy to explain**: "We prioritize minimum staff first, then preferred staff, then revenue..."

### Disadvantages
- ❌ **Less flexible**: Can't balance multiple factors simultaneously
- ❌ **May ignore important factors**: Lower priority criteria rarely matter if higher ones differ
- ❌ **Tie-breaking complexity**: Many ties may require all criteria to be evaluated

### Implementation Changes

**Backend:**
- Replace `CriteriaWeights` with `CriteriaPriorityOrder` (array of criterion IDs in priority order)
- Modify `GenerateRankedSuggestions()` to use lexicographic sorting instead of weighted average
- Update database schema: Store priority order instead of weights

**Frontend:**
- Replace weight sliders with drag-and-drop priority list
- Update UI to show priority numbers instead of weights
- Remove weight percentage calculations

---

## Alternative 2: Tiered Priority with Score Thresholds

### Design Philosophy
Criteria are grouped into priority tiers. Within each tier, scores are combined (equal weight). Higher tiers always take precedence over lower tiers, but within a tier, multiple factors can be balanced.

### How It Works

1. **Tier Configuration**: Admin assigns each criterion to a tier (1-5, where 1 = highest)
   - Multiple criteria can be in the same tier
   - Criteria in the same tier are combined with equal weight

2. **Evaluation Process**:
   ```
   For each branch-position combination:
     Calculate Tier 1 score (average of all Tier 1 criteria)
     Calculate Tier 2 score (average of all Tier 2 criteria)
     ...
     
   Ranking:
     Sort by Tier 1 score (descending)
     For ties, sort by Tier 2 score (descending)
     Continue...
   ```

3. **Score Calculation**:
   - Tier score = Average of all criteria scores in that tier
   - If a tier has no active criteria, it's skipped

### Example Scenario

**Tier Configuration:**
- **Tier 1 (Critical)**: Minimum Staff Shortage
- **Tier 2 (Important)**: Preferred Staff Shortage, Branch Type Staff Groups
- **Tier 3 (Consideration)**: Branch-Level Variables
- **Tier 4 (Optional)**: Doctor Preferences

**Branch-Position Combinations:**
- Branch A: Tier1=1.0, Tier2=0.75 (Pref=0.8, Type=0.7), Tier3=0.6
- Branch B: Tier1=0.0, Tier2=0.90 (Pref=0.9, Type=0.9), Tier3=0.9
- Branch C: Tier1=1.0, Tier2=0.65 (Pref=0.6, Type=0.7), Tier3=0.5

**Ranking Result:**
1. Branch A (Tier1=1.0, Tier2=0.75)
2. Branch C (Tier1=1.0, Tier2=0.65)
3. Branch B (Tier1=0.0, even though Tier2 and Tier3 are higher)

### Configuration UI

```
┌─────────────────────────────────────────────────────────────┐
│  Allocation Criteria Tier Configuration                    │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  Assign criteria to priority tiers:                        │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │ TIER 1 (Critical - Highest Priority)                 │  │
│  │ ┌──────────────────────────────────────────────────┐ │  │
│  │ │ ☑ 🚨 Minimum Staff Shortage                     │ │  │
│  │ └──────────────────────────────────────────────────┘ │  │
│  │                                                       │  │
│  │ TIER 2 (Important)                                    │  │
│  │ ┌──────────────────────────────────────────────────┐ │  │
│  │ │ ☑ ⭐ Preferred Staff Shortage                     │ │  │
│  │ │ ☑ 🏢 Branch Type Staff Groups                     │ │  │
│  │ └──────────────────────────────────────────────────┘ │  │
│  │                                                       │  │
│  │ TIER 3 (Consideration)                               │  │
│  │ ┌──────────────────────────────────────────────────┐ │  │
│  │ │ ☑ 📊 Branch-Level Variables                     │ │  │
│  │ └──────────────────────────────────────────────────┘ │  │
│  │                                                       │  │
│  │ TIER 4 (Optional)                                    │  │
│  │ ┌──────────────────────────────────────────────────┐ │  │
│  │ │ ☐ 👨‍⚕️ Doctor Preferences                         │ │  │
│  │ └──────────────────────────────────────────────────┘ │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                              │
│  [Add Tier]  [Remove Empty Tiers]                           │
│                                                              │
│  [Reset to Defaults]              [Save Configuration]     │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### Advantages
- ✅ **Balanced approach**: Can combine multiple factors within tiers
- ✅ **Clear hierarchy**: Tiers provide clear priority levels
- ✅ **Flexible**: Can have multiple criteria per tier
- ✅ **Intuitive**: "Critical tier first, then important tier..."

### Disadvantages
- ❌ **More complex**: Need to understand tier concept
- ❌ **Equal weighting within tier**: Can't prioritize one criterion over another in same tier
- ❌ **More configuration**: Need to assign criteria to tiers

### Implementation Changes

**Backend:**
- Replace `CriteriaWeights` with `CriteriaTiers` (map of criterion ID to tier number)
- Modify `GenerateRankedSuggestions()` to:
  1. Group criteria by tier
  2. Calculate tier scores (average of criteria in tier)
  3. Sort by tier scores in priority order
- Update database schema: Store tier assignments

**Frontend:**
- Replace weight sliders with tier assignment UI
- Show tier visualization (tier 1, tier 2, etc.)
- Allow drag-and-drop between tiers

---

## Alternative 3: Priority Points System

### Design Philosophy
Each criterion is assigned a priority level (1-5). When evaluating, criteria are scored and multiplied by their priority multiplier. Higher priority criteria get exponentially more influence, ensuring they dominate when present.

### How It Works

1. **Priority Level Assignment**: Admin assigns each criterion a priority level (1-5)
   - Priority 1 = Highest (multiplier: 1000)
   - Priority 2 = High (multiplier: 100)
   - Priority 3 = Medium (multiplier: 10)
   - Priority 4 = Low (multiplier: 1)
   - Priority 5 = Very Low (multiplier: 0.1)

2. **Score Calculation**:
   ```
   Priority Score = Σ (criterion_score × priority_multiplier)
   
   Example:
   - Min Staff (Priority 1): score=1.0 → 1.0 × 1000 = 1000 points
   - Pref Staff (Priority 2): score=0.8 → 0.8 × 100 = 80 points
   - Revenue (Priority 3): score=0.9 → 0.9 × 10 = 9 points
   Total = 1089 points
   ```

3. **Ranking**: Sort by total priority points (descending)

### Example Scenario

**Priority Levels:**
- Minimum Staff Shortage: Priority 1 (×1000)
- Preferred Staff Shortage: Priority 2 (×100)
- Branch-Level Variables: Priority 3 (×10)
- Branch Type Staff Groups: Priority 4 (×1)
- Doctor Preferences: Priority 5 (×0.1)

**Branch-Position Combinations:**
- Branch A: Min=1.0(1000), Pref=0.7(70), Revenue=0.6(6) → **1076 points**
- Branch B: Min=0.0(0), Pref=0.9(90), Revenue=0.9(9) → **99 points**
- Branch C: Min=0.5(500), Pref=0.5(50), Revenue=0.5(5) → **555 points**

**Ranking Result:**
1. Branch A (1076 points)
2. Branch C (555 points)
3. Branch B (99 points)

### Configuration UI

```
┌─────────────────────────────────────────────────────────────┐
│  Allocation Criteria Priority Levels                       │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  Assign priority level to each criterion:                   │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │ 🚨 Minimum Staff Shortage                           │  │
│  │ Priority: [Priority 1 ▼]  (×1000 multiplier)        │  │
│  │                                                       │  │
│  │ ⭐ Preferred Staff Shortage                          │  │
│  │ Priority: [Priority 2 ▼]  (×100 multiplier)         │  │
│  │                                                       │  │
│  │ 📊 Branch-Level Variables                            │  │
│  │ Priority: [Priority 3 ▼]  (×10 multiplier)           │  │
│  │                                                       │  │
│  │ 🏢 Branch Type Staff Groups                           │  │
│  │ Priority: [Priority 4 ▼]  (×1 multiplier)            │  │
│  │                                                       │  │
│  │ 👨‍⚕️ Doctor Preferences                               │  │
│  │ Priority: [Priority 5 ▼]  (×0.1 multiplier)         │  │
│  │ ☑ Enable                                               │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                              │
│  Priority Level Reference:                                  │
│  • Priority 1: Critical (×1000) - Always dominates        │
│  • Priority 2: High (×100) - Strong influence              │
│  • Priority 3: Medium (×10) - Moderate influence           │
│  • Priority 4: Low (×1) - Weak influence                    │
│  • Priority 5: Very Low (×0.1) - Minimal influence         │
│                                                              │
│  [Reset to Defaults]              [Save Configuration]     │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### Advantages
- ✅ **Flexible**: Can still combine all criteria
- ✅ **Guaranteed precedence**: High priority criteria dominate through multipliers
- ✅ **Numeric scoring**: Still produces a single score for ranking
- ✅ **Graduated influence**: Priority levels create clear hierarchy
- ✅ **Backward compatible**: Similar to weighted system but with exponential differences

### Disadvantages
- ❌ **Complex multipliers**: Need to understand exponential scaling
- ❌ **Less intuitive**: Priority points less clear than strict ordering
- ❌ **Potential overflow**: Very high multipliers could cause numeric issues

### Implementation Changes

**Backend:**
- Replace `CriteriaWeights` with `CriteriaPriorityLevels` (map of criterion ID to priority level 1-5)
- Define priority multipliers: `[1000, 100, 10, 1, 0.1]`
- Modify `GenerateRankedSuggestions()` to:
  1. Calculate each criterion score
  2. Multiply by priority multiplier
  3. Sum all priority points
  4. Sort by total points
- Update database schema: Store priority levels

**Frontend:**
- Replace weight sliders with priority level dropdowns
- Show multiplier values and explain impact
- Display priority points in suggestion results

---

## Comparison Matrix

| Aspect | Alternative 1: Strict Priority | Alternative 2: Tiered Priority | Alternative 3: Priority Points |
|--------|-------------------------------|--------------------------------|--------------------------------|
| **Complexity** | Low | Medium | Medium |
| **Flexibility** | Low (strict order) | Medium (tiers) | High (all factors) |
| **Intuitiveness** | Very High | High | Medium |
| **Guaranteed Precedence** | Yes (absolute) | Yes (tier-based) | Yes (exponential) |
| **Configuration Effort** | Low | Medium | Low |
| **Balancing Multiple Factors** | No | Yes (within tiers) | Yes (all factors) |
| **Tie Handling** | Sequential evaluation | Sequential by tier | Single score |
| **Best For** | Clear, simple priority rules | Balanced multi-factor needs | Flexible scoring with precedence |

---

## Recommendations

### For Maximum Clarity and Simplicity
**Recommended: Alternative 1 (Strict Priority Ranking)**
- Easiest to understand and explain
- Guarantees critical criteria always take precedence
- Simple configuration (just reorder)
- Best for operational clarity

### For Balanced Multi-Factor Consideration
**Recommended: Alternative 2 (Tiered Priority)**
- Allows combining factors within tiers
- Still maintains clear priority hierarchy
- Good balance between simplicity and flexibility
- Best when multiple factors should be considered together

### For Maximum Flexibility with Precedence
**Recommended: Alternative 3 (Priority Points)**
- Most similar to current weighted system
- All factors contribute to final score
- Exponential multipliers ensure precedence
- Best for complex scenarios where all factors matter

---

## Migration Considerations

### Database Schema Changes

**Current:**
```sql
-- Settings table stores weights as JSON
{
  "zeroth_criteria": 0.0,
  "first_criteria": 0.25,
  "second_criteria": 0.20,
  "third_criteria": 0.30,
  "fourth_criteria": 0.25
}
```

**Alternative 1:**
```sql
-- Settings table stores priority order as JSON array
{
  "priority_order": [
    "third_criteria",    // Priority 1
    "second_criteria",   // Priority 2
    "first_criteria",    // Priority 3
    "fourth_criteria",   // Priority 4
    "zeroth_criteria"    // Priority 5
  ],
  "enable_doctor_preferences": false
}
```

**Alternative 2:**
```sql
-- Settings table stores tier assignments
{
  "tiers": {
    "1": ["third_criteria"],
    "2": ["second_criteria", "fourth_criteria"],
    "3": ["first_criteria"],
    "4": ["zeroth_criteria"]
  },
  "enable_doctor_preferences": false
}
```

**Alternative 3:**
```sql
-- Settings table stores priority levels
{
  "priority_levels": {
    "third_criteria": 1,   // Priority 1 (×1000)
    "second_criteria": 2,   // Priority 2 (×100)
    "first_criteria": 3,    // Priority 3 (×10)
    "fourth_criteria": 4,   // Priority 4 (×1)
    "zeroth_criteria": 5    // Priority 5 (×0.1)
  },
  "enable_doctor_preferences": false
}
```

### Code Changes Required

1. **Backend (`multi_criteria_filter.go`)**:
   - Replace `CriteriaWeights` struct with priority configuration struct
   - Modify `GenerateRankedSuggestions()` sorting logic
   - Update evaluation to use priority-based ranking

2. **Backend (`allocation_criteria_handler.go`)**:
   - Update API endpoints to handle priority configuration
   - Modify validation logic
   - Update default values

3. **Frontend (`allocation-criteria/page.tsx`)**:
   - Replace weight slider UI with priority configuration UI
   - Update state management
   - Modify display logic

4. **API Client (`allocation-criteria.ts`)**:
   - Update TypeScript interfaces
   - Modify API request/response types

---

## Next Steps

1. ✅ Review these three alternatives with stakeholders
2. ⏳ Choose primary approach (recommend Alternative 1 for simplicity)
3. ⏳ Create detailed implementation plan
4. ⏳ Update database schema
5. ⏳ Implement backend changes
6. ⏳ Update frontend UI
7. ⏳ Migrate existing weight configurations
8. ⏳ Test and validate
9. ⏳ Update documentation

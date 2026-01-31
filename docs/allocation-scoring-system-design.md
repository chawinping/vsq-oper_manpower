# Allocation Scoring System Design

## Overview

This document describes the new scoring-based allocation system that replaces the priority-based ranking. The system uses three separate scoring groups with lexicographic ordering (Group 1 has highest priority, then Group 2, then Group 3).

---

## Scoring Groups

### Group 1: Position Quota - Minimum Shortage (Negative Points)
**Priority: Highest (checked first)**

- **Calculation**: For each position below minimum, `-1 point × shortage_amount`
- **Formula**: `Group1_Score = -1 × Σ(max(0, minimum_required - current_count))` for all positions
- **Example**: 
  - Position A: Needs 5 minimum, has 3 → `-2 points`
  - Position B: Needs 3 minimum, has 3 → `0 points`
  - **Total Group1: -2**

**Purpose**: Identifies critical shortages where branches are below minimum staffing requirements.

---

### Group 2: Daily Staff Constraints - Minimum Shortage (Negative Points)
**Priority: Second (checked after Group 1)**

- **Calculation**: For each staff group below minimum, `-1 point × shortage_amount`
- **Formula**: `Group2_Score = -1 × Σ(max(0, minimum_count - actual_count))` for all staff groups
- **Example**:
  - Staff Group "Nurses": Needs 3 minimum, has 1 → `-2 points`
  - Staff Group "Managers": Needs 2 minimum, has 2 → `0 points`
  - **Total Group2: -2**

**Purpose**: Identifies shortages in staff group-based constraints (Daily Staff Constraints).

**Note**: This may overlap with Group 1 (same staff counted in both), but scores are shown separately as requested.

---

### Group 3: Position Quota - Preferred Excess (Positive Points)
**Priority: Third (checked after Group 1 and Group 2)**

- **Calculation**: For each position above preferred quota, `+1 point × excess_amount`
- **Formula**: `Group3_Score = +1 × Σ(max(0, current_count - preferred_quota))` for positions where `current_count > preferred_quota`
- **Example**: 
  - Position A: Needs 5 preferred, has 7 → `+2 points` (2 above preferred)
  - Position B: Needs 4 preferred, has 4 → `0 points` (at preferred)
  - Position C: Needs 3 preferred, has 2 → `0 points` (below preferred, not counted)
  - **Total Group3: +2**

**Purpose**: Identifies positions that are overstaffed relative to preferred quotas (informational only - shows how much above the limit for each position).

---

## Ranking Logic

### Lexicographic Ordering (Priority-Based)

The system ranks branch-position combinations using strict lexicographic ordering:

1. **Primary Sort**: Group 1 Score (ascending - more negative = higher priority)
   - Branch with `-5` ranks higher than branch with `-2`
   - Branch with `-2` ranks higher than branch with `0`

2. **Secondary Sort**: Group 2 Score (ascending - more negative = higher priority)
   - Only applied when Group 1 scores are equal
   - Branch with `-3` ranks higher than branch with `-1`

3. **Tertiary Sort**: Group 3 Score (descending - more positive = lower priority)
   - Only applied when Group 1 and Group 2 scores are equal
   - Branch with `+5` ranks lower than branch with `+2` (less urgent)
   - Note: Positive points indicate less urgency, so higher positive = lower priority

4. **Deterministic Tie-Breaker**: Branch Code (alphabetical)
   - Only applied when all three groups are equal
   - Ensures consistent, deterministic ordering

### Ranking Algorithm

```go
// Pseudocode
sort.Slice(suggestions, func(i, j int) bool {
    // Primary: Group 1 (more negative = higher priority)
    if suggestions[i].Group1Score != suggestions[j].Group1Score {
        return suggestions[i].Group1Score < suggestions[j].Group1Score
    }
    
    // Secondary: Group 2 (more negative = higher priority)
    if suggestions[i].Group2Score != suggestions[j].Group2Score {
        return suggestions[i].Group2Score < suggestions[j].Group2Score
    }
    
    // Tertiary: Group 3 (more positive = lower priority)
    if suggestions[i].Group3Score != suggestions[j].Group3Score {
        return suggestions[i].Group3Score > suggestions[j].Group3Score
    }
    
    // Deterministic tie-breaker
    return suggestions[i].BranchCode < suggestions[j].BranchCode
})
```

---

## Example Scenarios

### Scenario 1: Critical Shortage (Group 1 Priority)

**Branch A:**
- Position 1: Needs 5 min, has 2 → Group1: `-3`
- Position 2: Needs 3 min, has 3 → Group1: `0`
- Group "Nurses": Needs 2 min, has 1 → Group2: `-1`
- Position 1: Needs 5 pref, has 2 → Group3: `0` (below preferred, not counted)
- **Scores: Group1=-3, Group2=-1, Group3=0**

**Branch B:**
- Position 1: Needs 5 min, has 4 → Group1: `-1`
- Group "Nurses": Needs 3 min, has 1 → Group2: `-2`
- Position 1: Needs 5 pref, has 4 → Group3: `0` (below preferred, not counted)
- **Scores: Group1=-1, Group2=-2, Group3=0**

**Ranking Result:**
1. **Branch A** (Group1=-3 is more negative than Branch B's -1)
2. **Branch B** (Group1=-1)

---

### Scenario 2: Group 1 Tie, Group 2 Decides

**Branch A:**
- Position 1: Needs 5 min, has 3 → Group1: `-2`
- Group "Nurses": Needs 3 min, has 2 → Group2: `-1`
- **Scores: Group1=-2, Group2=-1, Group3=0**

**Branch B:**
- Position 1: Needs 5 min, has 3 → Group1: `-2`
- Group "Nurses": Needs 3 min, has 1 → Group2: `-2`
- **Scores: Group1=-2, Group2=-2, Group3=0**

**Ranking Result:**
1. **Branch B** (Group1 tied at -2, but Group2=-2 is more negative than Branch A's -1)
2. **Branch A** (Group2=-1)

---

### Scenario 3: All Groups Tie, Branch Code Decides

**Branch A (Code: "A01"):**
- **Scores: Group1=-2, Group2=-1, Group3=+1**

**Branch B (Code: "B02"):**
- **Scores: Group1=-2, Group2=-1, Group3=+1**

**Ranking Result:**
1. **Branch A** (alphabetically "A01" < "B02")
2. **Branch B**

---

## Data Structure Changes

### AllocationSuggestion Structure

```go
type AllocationSuggestion struct {
    BranchID           uuid.UUID         `json:"branch_id"`
    BranchName         string            `json:"branch_name"`
    BranchCode         string            `json:"branch_code"`
    PositionID         uuid.UUID         `json:"position_id"`
    PositionName       string            `json:"position_name"`
    Date               time.Time         `json:"date"`
    
    // New scoring system
    Group1Score        int               `json:"group1_score"`        // Position Quota - Minimum (negative)
    Group2Score        int               `json:"group2_score"`        // Daily Staff Constraints - Minimum (negative)
    Group3Score        int               `json:"group3_score"`         // Position Quota - Preferred (positive)
    
    // Legacy fields (deprecated, kept for backward compatibility)
    PriorityScore      float64           `json:"priority_score,omitempty"`
    Reason             string            `json:"reason"`
    SuggestedStaffID   *uuid.UUID        `json:"suggested_staff_id,omitempty"`
    SuggestedStaffName string            `json:"suggested_staff_name,omitempty"`
    
    // Detailed breakdown for debugging/display
    ScoreBreakdown     ScoreBreakdown    `json:"score_breakdown"`
}

type ScoreBreakdown struct {
    // Group 1 breakdown
    PositionQuotaMinimum []PositionQuotaScore `json:"position_quota_minimum"`
    
    // Group 2 breakdown
    DailyConstraintsMinimum []StaffGroupScore  `json:"daily_constraints_minimum"`
    
    // Group 3 breakdown
    PositionQuotaPreferred []PositionQuotaScore `json:"position_quota_preferred"`
}

type PositionQuotaScore struct {
    PositionID    uuid.UUID `json:"position_id"`
    PositionName  string    `json:"position_name"`
    MinimumRequired int     `json:"minimum_required"`
    CurrentCount    int     `json:"current_count"`
    Shortage        int     `json:"shortage"`
    Points          int     `json:"points"`
}

type StaffGroupScore struct {
    StaffGroupID   uuid.UUID `json:"staff_group_id"`
    StaffGroupName string    `json:"staff_group_name"`
    MinimumCount   int       `json:"minimum_count"`
    ActualCount    int       `json:"actual_count"`
    Shortage       int       `json:"shortage"`
    Points         int       `json:"points"`
}
```

---

## Implementation Details

### Calculation Functions

#### Group 1: Position Quota - Minimum Shortage

```go
func calculateGroup1Score(branchID uuid.UUID, date time.Time) (int, []PositionQuotaScore, error) {
    quotas, err := repos.PositionQuota.GetByBranchID(branchID)
    if err != nil {
        return 0, nil, err
    }
    
    totalScore := 0
    breakdown := []PositionQuotaScore{}
    
    for _, quota := range quotas {
        if !quota.IsActive {
            continue
        }
        
        currentCount, err := calculateCurrentStaffCount(branchID, quota.PositionID, date)
        if err != nil {
            continue
        }
        
        shortage := quota.MinimumRequired - currentCount
        if shortage > 0 {
            points := -1 * shortage
            totalScore += points
            
            breakdown = append(breakdown, PositionQuotaScore{
                PositionID:      quota.PositionID,
                MinimumRequired: quota.MinimumRequired,
                CurrentCount:    currentCount,
                Shortage:        shortage,
                Points:          points,
            })
        }
    }
    
    return totalScore, breakdown, nil
}
```

#### Group 2: Daily Staff Constraints - Minimum Shortage

```go
func calculateGroup2Score(branchID uuid.UUID, date time.Time) (int, []StaffGroupScore, error) {
    dayOfWeek := int(date.Weekday())
    
    // Get branch constraints for this day
    constraints, err := repos.BranchConstraints.GetByBranchID(branchID)
    if err != nil {
        return 0, nil, err
    }
    
    // Find constraint for this day
    var constraint *models.BranchConstraints
    for _, c := range constraints {
        if c.DayOfWeek == dayOfWeek {
            constraint = c
            break
        }
    }
    
    if constraint == nil {
        return 0, nil, nil // No constraints for this day
    }
    
    // Load staff group requirements
    if err := repos.BranchConstraints.LoadStaffGroupRequirements([]*models.BranchConstraints{constraint}); err != nil {
        return 0, nil, err
    }
    
    totalScore := 0
    breakdown := []StaffGroupScore{}
    
    for _, req := range constraint.StaffGroupRequirements {
        actualCount, err := calculateStaffGroupCount(branchID, req.StaffGroupID, date)
        if err != nil {
            continue
        }
        
        shortage := req.MinimumCount - actualCount
        if shortage > 0 {
            points := -1 * shortage
            totalScore += points
            
            breakdown = append(breakdown, StaffGroupScore{
                StaffGroupID: req.StaffGroupID,
                MinimumCount: req.MinimumCount,
                ActualCount:  actualCount,
                Shortage:     shortage,
                Points:      points,
            })
        }
    }
    
    return totalScore, breakdown, nil
}
```

#### Group 3: Position Quota - Preferred Excess

```go
func calculateGroup3Score(branchID uuid.UUID, date time.Time) (int, []PositionQuotaScore, error) {
    quotas, err := repos.PositionQuota.GetByBranchID(branchID)
    if err != nil {
        return 0, nil, err
    }
    
    totalScore := 0
    breakdown := []PositionQuotaScore{}
    
    for _, quota := range quotas {
        if !quota.IsActive {
            continue
        }
        
        currentCount, err := calculateCurrentStaffCount(branchID, quota.PositionID, date)
        if err != nil {
            continue
        }
        
        // Only count positions with actual staff number greater than preferred number
        excess := currentCount - quota.DesignatedQuota
        if excess > 0 {
            points := +1 * excess
            totalScore += points
            
            breakdown = append(breakdown, PositionQuotaScore{
                PositionID:      quota.PositionID,
                MinimumRequired: quota.MinimumRequired,
                CurrentCount:    currentCount,
                Shortage:        excess, // Represents excess amount (how much above preferred)
                Points:          points,
            })
        }
    }
    
    return totalScore, breakdown, nil
}
```

---

## UI Display

### Suggestion Card Display

```
┌─────────────────────────────────────────────────────────┐
│ Branch A01 - Branch Name                                │
│ Position: Nurse                                         │
│ Date: 2026-01-30                                        │
├─────────────────────────────────────────────────────────┤
│ Priority Scores:                                        │
│                                                          │
│ Group 1 (Position Quota - Minimum):  -3 points  🔴     │
│   • Nurse: Needs 5, Has 2 → -3 points                  │
│                                                          │
│ Group 2 (Daily Constraints - Minimum):  -1 points  🟡  │
│   • Nurses Group: Needs 2, Has 1 → -1 points           │
│                                                          │
│ Group 3 (Position Quota - Preferred):  +2 points  🟢 │
│   • Nurse: Needs 5 preferred, Has 7 → +2 points (2 above preferred) │
│                                                          │
│ [Assign Rotation Staff →]                                │
└─────────────────────────────────────────────────────────┘
```

### Ranking Display

```
Rank | Branch | Position | Group 1 | Group 2 | Group 3 | Action
-----|--------|----------|---------|---------|---------|--------
1    | A01    | Nurse    | -5      | -2      | 0       | [Assign]
2    | B02    | Nurse    | -3      | -1      | +1      | [Assign]
3    | C03    | Nurse    | -3      | 0       | +2      | [Assign]
4    | D04    | Nurse    | -2      | -3      | +1      | [Assign]
```

---

## Migration Considerations

### Backward Compatibility

1. **Keep `PriorityScore` field** (deprecated but populated for compatibility)
2. **Keep `CriteriaBreakdown` field** (deprecated but populated for compatibility)
3. **Add new scoring fields** alongside existing ones
4. **Update frontend gradually** to use new scoring system

### Database Changes

No database schema changes required. All calculations are done in-memory based on existing data:
- `position_quotas` table (for Group 1 and Group 3)
- `branch_constraints` and `branch_constraint_staff_groups` tables (for Group 2)

### API Changes

**Request**: No changes needed (same endpoint)

**Response**: Add new fields to `AllocationSuggestion`:
```json
{
  "branch_id": "...",
  "position_id": "...",
  "group1_score": -3,
  "group2_score": -1,
  "group3_score": 2,
  "score_breakdown": { ... },
  // ... existing fields
}
```

---

## Future Enhancements

1. **Revenue Integration**: Link revenue directly to daily staff constraints (as mentioned)
2. **Weighted Scoring**: Allow configurable weights for groups (if needed)
3. **Normalization**: Add option to normalize scores by branch size (if needed)
4. **Visual Indicators**: Color-code scores (red for Group 1, yellow for Group 2, green for Group 3)

---

## Testing Scenarios

### Test Case 1: Group 1 Priority
- Branch A: Group1=-5, Group2=-1, Group3=0
- Branch B: Group1=-2, Group2=-5, Group3=0
- **Expected**: Branch A ranks higher (Group1=-5 < -2)

### Test Case 2: Group 1 Tie, Group 2 Decides
- Branch A: Group1=-3, Group2=-2, Group3=0
- Branch B: Group1=-3, Group2=-1, Group3=0
- **Expected**: Branch A ranks higher (Group2=-2 < -1)

### Test Case 3: Group 1 & 2 Tie, Group 3 Decides
- Branch A: Group1=-2, Group2=-1, Group3=+5
- Branch B: Group1=-2, Group2=-1, Group3=+2
- **Expected**: Branch B ranks higher (Group3=+2 < +5, less positive = higher priority)

### Test Case 4: All Groups Tie
- Branch A: Group1=-2, Group2=-1, Group3=+1, Code="A01"
- Branch B: Group1=-2, Group2=-1, Group3=+1, Code="B02"
- **Expected**: Branch A ranks higher (alphabetical)

---

## Summary

- **Three separate scoring groups** displayed independently
- **Lexicographic ordering**: Group 1 → Group 2 → Group 3 → Branch Code
- **Negative points** for minimum shortages (more negative = higher priority)
- **Positive points** for preferred excesses (informational only - shows how much above preferred limit)
- **Magnitude matters**: Shortage/excess amount directly affects points
- **No normalization**: Absolute points used for comparison
- **Overlap allowed**: Same staff can contribute to multiple groups

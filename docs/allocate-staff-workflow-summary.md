# Allocate Staff Workflow - Quick Summary

## Overview
Three alternative workflow designs for allocating rotation staff to branches, following a 4-step process:
1. Select date
2. Select branch(es) or all branches
3. View allocation overview with staff summaries
4. Add rotation staff to branches

---

## Quick Comparison

| Aspect | Alternative 1: Wizard | Alternative 2: Dashboard | Alternative 3: Sidebar |
|--------|----------------------|-------------------------|------------------------|
| **Pattern** | Step-by-step guide | Overview-first grid | Master-detail |
| **Best For** | New users, guided flow | Power users, quick scan | Focused analysis |
| **32 Branches** | Compact grid (Step 3) | Compact grid (default) | Sidebar list |
| **Speed** | Slower (more clicks) | Fastest | Medium |
| **Learning Curve** | Low | Medium | Low-Medium |

---

## Alternative 1: Wizard/Stepper
**Key Features:**
- Linear progression through 3 steps
- Clear guidance at each stage
- Prevents skipping steps
- Good onboarding experience

**Best When:**
- Training new users
- Need explicit confirmation
- Infrequent use

---

## Alternative 2: Dashboard/Grid ⭐ **RECOMMENDED**
**Key Features:**
- All information visible at once
- Compact grid handles 32 branches efficiently
- Powerful filtering and search
- Fast scanning and comparison

**Best When:**
- Daily operations
- Need to compare multiple branches
- Quick decision making
- Bulk operations

**Layout:**
```
[Date Picker] [Branch Selector] [Summary Stats]
[Filter Bar]
[Grid of Branch Cards - 4-6 columns]
[Action Buttons]
```

---

## Alternative 3: Sidebar/Detail
**Key Features:**
- Persistent branch list sidebar
- Detailed view for selected branch
- Familiar master-detail pattern
- Good for focused work

**Best When:**
- Working on specific branches
- Need detailed information
- Sequential branch review
- Less comparison needed

**Layout:**
```
[Date Picker] [Branch Selector]
[Sidebar: Branch List] | [Main: Branch Details]
```

---

## Recommended Approach

### Primary: Alternative 2 (Dashboard/Grid)
- Best for handling 32 branches
- Efficient workflow for daily use
- Flexible filtering and views

### Optional Enhancement: Add Wizard Mode
- Toggle "Guided Mode" for new users
- Uses Alternative 1 workflow
- Can be disabled for experienced users

### Optional Enhancement: Add Sidebar Navigation
- Collapsible sidebar in Dashboard view
- Quick branch navigation
- Best of both worlds

---

## Key UI Components Needed

### 1. Date Picker
- Compact dropdown/input
- Quick select: Today, Tomorrow, Next Week
- Calendar popup for date selection

### 2. Branch Selector
- "All Branches" checkbox
- Multi-select dropdown
- Selected branches as chips/tags
- Search/filter capability

### 3. Branch Card (Compact - 32 branches)
- Branch code (large, bold)
- Branch name (small, gray)
- Staff ratio (current/preferred)
- Priority badge (color-coded)
- "Add Staff" button

### 4. Branch Card (Detailed - single branch)
- Full branch information
- Current staff breakdown table
- Allocation suggestions list
- Action buttons

### 5. Filter Bar
- Status filter (All, Needs Attention, Critical, OK)
- Priority filter (High, Medium, Low)
- Search box
- View toggle (Grid/List/Dense)

### 6. Add Staff Modal/Drawer
- Branch information header
- Available rotation staff list
- Search and filter
- Assignment form (staff, position, level)
- Action buttons

---

## Data Requirements

### From API:
1. **Branch List**: All branches with codes and names
2. **Current Staff**: Staff count by position for selected date
3. **Allocation Suggestions**: Prioritized suggestions from allocation engine
4. **Available Rotation Staff**: List of rotation staff available for assignment
5. **Staff Requirements**: Preferred and minimum counts per position

### Display Metrics:
- Current staff count per position
- Preferred staff count per position
- Minimum staff count per position
- Priority score (0.0 - 1.0)
- Suggestion reason/explanation

---

## User Flow

### Alternative 2 (Dashboard) Flow:
```
1. User opens "Allocate Staff" page
2. Selects date (default: today)
3. Selects "All Branches" or specific branches
4. System loads and displays overview grid
5. User scans branches, sees priority badges
6. User clicks branch card → Opens detail drawer
7. User reviews suggestions and available staff
8. User selects staff and assigns
9. System updates and refreshes view
```

### Alternative 1 (Wizard) Flow:
```
1. User opens "Allocate Staff" page
2. Step 1: Selects date → Clicks "Continue"
3. Step 2: Selects branches → Clicks "Continue"
4. Step 3: Views overview → Clicks branch
5. Modal opens: User assigns staff
6. Returns to Step 3, can continue or go back
```

---

## Next Steps

1. ✅ Review mockups with stakeholders
2. ⏳ Choose primary approach (recommend Alternative 2)
3. ⏳ Create detailed component specifications
4. ⏳ Design API endpoints if needed
5. ⏳ Build interactive prototype
6. ⏳ User testing
7. ⏳ Iterate based on feedback
8. ⏳ Implement final design

---

## Files Created

1. **allocate-staff-workflow-alternatives.md** - Detailed design document with all three alternatives
2. **allocate-staff-workflow-mockups.md** - Visual mockups with ASCII diagrams
3. **allocate-staff-workflow-summary.md** - This quick reference document

---

## Questions to Consider

1. **Primary Use Case**: Will users mostly work with all 32 branches or specific ones?
2. **User Expertise**: Are users experienced or will they need guidance?
3. **Frequency**: Is this a daily task or occasional?
4. **Mobile Usage**: Will this be used on mobile devices?
5. **Integration**: Does this need to integrate with other workflows?

---

## Notes

- All alternatives support the same 4-step process
- Alternative 2 is recommended for primary implementation
- Consider hybrid approach combining best features
- Ensure compact view works well for 32 branches
- Prioritize quick scanning and action for power users

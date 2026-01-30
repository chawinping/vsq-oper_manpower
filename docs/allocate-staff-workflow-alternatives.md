# Allocate Staff Workflow - Three Alternative Designs

## Overview
This document presents three alternative workflow designs for the "Allocate Staff" feature. Each design follows the same 4-step process but with different UX approaches:

1. **Select a date for allocation**
2. **Select a branch or branches or all branches**
3. **Show overview of allocation suggestions for each selected branch, along with summary of staff for that day**
4. **User can click any branch to add a rotation staff**

---

## Alternative 1: Wizard/Stepper Workflow

### Design Philosophy
A guided, step-by-step process that walks users through each stage. Best for users who prefer clear progression and explicit confirmation at each step.

### Layout Structure
```
┌─────────────────────────────────────────────────────────────┐
│  Allocate Staff                                             │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  [Step 1: Date] ──── [Step 2: Branches] ──── [Step 3: Overview] │
│     ✓              →        [Current]        →     [Pending]    │
│                                                              │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  STEP 2: SELECT BRANCHES                                    │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐   │
│  │ ☑ Select All Branches (32)                         │   │
│  └──────────────────────────────────────────────────────┘   │
│                                                              │
│  OR Select Individual Branches:                             │
│  ┌──────────────────────────────────────────────────────┐   │
│  │ ☐ Branch A (Code: A01)                            │   │
│  │ ☐ Branch B (Code: B02)                            │   │
│  │ ☐ Branch C (Code: C03)                            │   │
│  │ ... (scrollable list)                              │   │
│  └──────────────────────────────────────────────────────┘   │
│                                                              │
│  Selected: 0 branches                                       │
│                                                              │
│  [← Back]                    [Continue →]                  │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### Step-by-Step Flow

#### Step 1: Date Selection
- **Layout**: Full-width form with prominent date picker
- **Components**:
  - Large calendar date picker (default: today)
  - Quick select buttons: "Today", "Tomorrow", "Next Week"
  - Date display: "Selected: January 25, 2026 (Saturday)"
- **Navigation**: "Next →" button (disabled until date selected)

#### Step 2: Branch Selection
- **Layout**: Two-column layout
  - Left: Selection options (radio/checkbox)
  - Right: Preview of selected branches
- **Components**:
  - Radio button: "All Branches (32)"
  - Checkbox list: Individual branch selection
  - Search/filter box for branch list
  - Selected count badge
- **Navigation**: "← Back" and "Continue →" buttons

#### Step 3: Overview Display
- **Layout**: Full-width grid/table view
- **For Single/Multiple Branches (< 10)**:
  - Detailed card view with:
    - Branch name and code
    - Current staff count by position
    - Allocation suggestions (with priority scores)
    - Status indicators (meets minimum, preferred, etc.)
    - "Add Rotation Staff" button per branch
- **For All Branches (32)**:
  - Compact grid (4 columns x 8 rows)
  - Each card shows:
    - Branch code (large)
    - Branch name (small)
    - Staff summary: "5/8" (current/preferred)
    - Priority badge: 🔴 High / 🟡 Medium / 🟢 Low
    - Click to expand details
- **Navigation**: "← Back" and "Generate Suggestions" button

#### Step 4: Add Rotation Staff (Modal/Drawer)
- **Trigger**: Click on any branch card
- **Layout**: Right-side drawer or centered modal
- **Components**:
  - Branch info header
  - Current staff list
  - Suggested staff list (from allocation engine)
  - Search/filter for rotation staff
  - Assignment form (staff, position, level)
  - "Assign" button

### Advantages
- ✅ Clear progression and user guidance
- ✅ Prevents skipping steps
- ✅ Good for first-time users
- ✅ Explicit confirmation at each stage

### Disadvantages
- ❌ More clicks to reach overview
- ❌ Less flexible for power users
- ❌ May feel slow for frequent use

---

## Alternative 2: Dashboard/Grid Workflow

### Design Philosophy
Overview-first approach. Users see all information at once and can quickly scan and act. Best for users who need to see the big picture and make quick decisions.

### Layout Structure
```
┌─────────────────────────────────────────────────────────────┐
│  Allocate Staff                                             │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌──────────────────┐  ┌──────────────────────────────────┐ │
│  │ 📅 Date:        │  │ 🌳 Branch Selection:            │ │
│  │ [01/25/2026 ▼] │  │ [☑ All Branches] [Select... ▼] │ │
│  └──────────────────┘  └──────────────────────────────────┘ │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │ Summary: 32 branches | 8 need attention | 5 critical│  │
│  └──────────────────────────────────────────────────────┘  │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │ Filter: [All ▼] [Priority: High ▼] [Search...]     │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                              │
│  ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐  │
│  │ A01  │ │ B02  │ │ C03  │ │ D04  │ │ E05  │ │ F06  │  │
│  │Branch│ │Branch│ │Branch│ │Branch│ │Branch│ │Branch│  │
│  │ 5/8  │ │ 3/6  │ │ 7/7  │ │ 2/5  │ │ 6/8  │ │ 4/6  │  │
│  │ 🔴   │ │ 🔴   │ │ 🟢   │ │ 🔴   │ │ 🟡   │ │ 🟡   │  │
│  └──────┘ └──────┘ └──────┘ └──────┘ └──────┘ └──────┘  │
│                                                              │
│  ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐  │
│  │ G07  │ │ H08  │ │ ...  │ │ ...  │ │ ...  │ │ ...  │  │
│  │Branch│ │Branch│ │      │ │      │ │      │ │      │  │
│  │ 4/7  │ │ 5/8  │ │      │ │      │ │      │ │      │  │
│  │ 🟡   │ │ 🟢   │ │      │ │      │ │      │ │      │  │
│  └──────┘ └──────┘ └──────┘ └──────┘ └──────┘ └──────┘  │
│                                                              │
│  [Scroll for more...]                                       │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### Component Details

#### Top Control Bar
- **Date Picker**: Compact dropdown/date input (always visible)
- **Branch Selector**: 
  - Toggle: "All Branches" checkbox
  - Multi-select dropdown: "Select Branches..." (shows count)
  - Selected branches displayed as chips/tags
- **Summary Stats**: Quick metrics bar
  - Total branches selected
  - Branches needing attention
  - Critical priority count

#### Filter Bar
- **Filters**: 
  - Status filter: All / Needs Attention / Critical / OK
  - Priority filter: High / Medium / Low
  - Search box: Filter by branch name/code
- **View Options**: 
  - Grid view (default for 32 branches)
  - List view (for < 10 branches)
  - Compact/Dense toggle

#### Branch Grid (32 Branches - Compact View)
- **Grid Layout**: 4-6 columns (responsive)
- **Card Size**: Minimal, clickable
- **Card Content**:
  ```
  ┌─────────────┐
  │ A01         │ ← Branch code (large, bold)
  │ Branch Name │ ← Name (small, gray)
  │             │
  │ 5/8 🔴      │ ← Current/Preferred + Priority badge
  │             │
  │ + Add Staff │ ← Action button
  └─────────────┘
  ```
- **Hover State**: Slight elevation, show quick stats
- **Click Action**: Opens detail panel/drawer

#### Branch List (Single/Multiple Branches - Detailed View)
- **Layout**: Single column, expanded cards
- **Card Content**:
  ```
  ┌─────────────────────────────────────────┐
  │ A01 - Branch Name              [🔴 High]│
  ├─────────────────────────────────────────┤
  │ Current Staff:                          │
  │ • Nurse: 3/5 (2 short)                  │
  │ • Doctor Assistant: 2/3 (1 short)      │
  │                                         │
  │ Suggestions:                            │
  │ • Assign Nurse (Priority: 0.85)        │
  │ • Assign Doctor Assistant (Priority: 0.72)│
  │                                         │
  │ [Add Rotation Staff →]                  │
  └─────────────────────────────────────────┘
  ```

#### Detail Panel/Drawer (On Branch Click)
- **Position**: Right-side drawer (slides in)
- **Content**:
  - Branch header with full details
  - Current staff breakdown (table)
  - Allocation suggestions (prioritized list)
  - Available rotation staff (filterable)
  - Assignment form
- **Actions**: "Assign", "Assign & Continue", "Close"

### Advantages
- ✅ See everything at once
- ✅ Fast scanning and comparison
- ✅ Efficient for power users
- ✅ Compact view handles 32 branches well
- ✅ Quick filtering and search

### Disadvantages
- ❌ Can be overwhelming for new users
- ❌ Less guidance/onboarding
- ❌ Requires good visual hierarchy

---

## Alternative 3: Sidebar/Detail Workflow

### Design Philosophy
Master-detail pattern with persistent navigation. Users select branches from a sidebar and see detailed information in the main area. Best for users who work with specific branches and need detailed information.

### Layout Structure
```
┌─────────────────────────────────────────────────────────────┐
│  Allocate Staff                                             │
├──────────┬──────────────────────────────────────────────────┤
│          │                                                   │
│ 📅 Date: │  ┌────────────────────────────────────────────┐  │
│ [01/25/  │  │ Branch Selection:                        │  │
│  2026 ▼] │  │ ☑ All Branches  [Filter: All ▼]          │  │
│          │  └────────────────────────────────────────────┘  │
│          │                                                   │
│ ┌────────┴──────────────────────────────────────────────┐  │
│ │ BRANCHES (32)                    │ SEARCH: [____]    │  │
│ ├───────────────────────────────────────────────────────┤  │
│ │ 🔴 A01 - Branch A                5/8  [High]       │  │
│ │ 🔴 B02 - Branch B                3/6  [High]       │  │
│ │ 🟡 C03 - Branch C                4/7  [Medium]    │  │
│ │ 🟢 D04 - Branch D                7/7  [Low]        │  │
│ │ 🟡 E05 - Branch E                6/8  [Medium]    │  │
│ │ ... (scrollable)                                      │  │
│ └───────────────────────────────────────────────────────┘  │
│          │                                                   │
│          │  ┌────────────────────────────────────────────┐  │
│          │  │ A01 - Branch A                    [🔴 High]│  │
│          │  ├────────────────────────────────────────────┤  │
│          │  │ Current Staff Summary:                     │  │
│          │  │ • Nurses: 3/5 (Preferred: 5, Min: 3)       │  │
│          │  │ • Doctor Assistants: 2/3 (Preferred: 3)   │  │
│          │  │ • Receptionists: 2/2 ✓                    │  │
│          │  │                                             │  │
│          │  │ Allocation Suggestions:                    │  │
│          │  │ 1. Assign Nurse (Priority: 0.85)          │  │
│          │  │    Reason: Below preferred, high revenue   │  │
│          │  │    Suggested: Staff A, Staff B            │  │
│          │  │                                             │  │
│          │  │ 2. Assign Doctor Assistant (Priority: 0.72)│  │
│          │  │    Reason: Minimum threshold met, but...  │  │
│          │  │                                             │  │
│          │  │ [Add Rotation Staff →]                     │  │
│          │  └────────────────────────────────────────────┘  │
│          │                                                   │
└──────────┴───────────────────────────────────────────────────┘
```

### Component Details

#### Top Bar
- **Date Picker**: Always visible, compact
- **Branch Selection Toggle**: "All Branches" checkbox
- **Filter Dropdown**: Filter sidebar list (All, High Priority, Needs Attention, etc.)

#### Left Sidebar (Branch List)
- **Width**: ~300px (collapsible)
- **Content**: Scrollable list of branches
- **Each Item Shows**:
  - Branch code and name
  - Staff ratio (current/preferred)
  - Priority badge (color-coded)
  - Visual indicator if selected
- **Features**:
  - Search box at top
  - Sort options (by priority, name, code)
  - Selected branch highlighted
- **Interaction**: Click to select, shows details in main area

#### Main Detail Area
- **Layout**: Full-width, scrollable
- **When Branch Selected**:
  - Branch header with full name and code
  - Current staff breakdown (detailed table)
  - Allocation suggestions (prioritized, expandable)
  - Action buttons
- **When "All Branches" Selected**:
  - Summary table/grid view
  - All branches in compact format
  - Click to select individual branch
- **Empty State**: "Select a branch to view details"

#### Add Staff Modal/Drawer
- **Trigger**: "Add Rotation Staff" button
- **Layout**: Centered modal or right drawer
- **Content**: Same as other alternatives (staff selection, assignment form)

### Advantages
- ✅ Familiar master-detail pattern
- ✅ Easy navigation between branches
- ✅ Detailed view always available
- ✅ Good for focused work on specific branches
- ✅ Sidebar provides quick overview

### Disadvantages
- ❌ Less efficient for comparing multiple branches
- ❌ Requires clicking through branches
- ❌ Sidebar takes horizontal space
- ❌ May feel slower for bulk operations

---

## Comparison Matrix

| Feature | Wizard/Stepper | Dashboard/Grid | Sidebar/Detail |
|---------|---------------|----------------|----------------|
| **Best For** | First-time users, guided workflows | Power users, quick scanning | Focused work, detailed analysis |
| **32 Branches View** | Compact grid (Step 3) | Compact grid (default) | Sidebar list + detail view |
| **Single Branch View** | Detailed card | Detailed card | Full detail panel |
| **Navigation** | Linear (step-by-step) | Direct (all at once) | Hierarchical (sidebar → detail) |
| **Learning Curve** | Low | Medium | Low-Medium |
| **Speed** | Slower (more clicks) | Fastest (overview first) | Medium (click to navigate) |
| **Flexibility** | Low (guided) | High (filters, views) | Medium (sidebar navigation) |
| **Mobile Friendly** | Good (steps stack) | Good (grid responsive) | Challenging (sidebar) |

---

## Recommendations

### For Primary Use Case (32 Branches Overview)
**Recommended: Alternative 2 (Dashboard/Grid)**
- Best compact view for 32 branches
- Fast scanning and filtering
- Efficient for bulk operations

### For Combination Approach
**Recommended: Hybrid of Alternative 2 + Alternative 3**
- Use Dashboard/Grid as default
- Add collapsible sidebar for quick branch navigation
- Allow switching between grid and detail views

### For Guided Experience
**Recommended: Alternative 1 (Wizard)**
- Add as an optional "Guided Mode" toggle
- Useful for onboarding new users
- Can be skipped by experienced users

---

## Next Steps

1. Review these mockups with stakeholders
2. Choose primary approach (or combination)
3. Create detailed component specifications
4. Build interactive prototype
5. User testing with target users
6. Iterate based on feedback

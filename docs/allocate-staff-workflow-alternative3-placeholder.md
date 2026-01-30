# Alternative 3: Sidebar/Detail Workflow - Implementation Guide

## Status: Placeholder for Future Implementation

This document serves as a reference guide for implementing Alternative 3 (Sidebar/Detail Workflow) in the future. **This workflow is NOT currently implemented** - only Alternative 2 (Dashboard/Grid) is active.

---

## Overview

Alternative 3 uses a master-detail pattern with a persistent sidebar for branch navigation and a main detail area for viewing branch-specific allocation information.

---

## Design Philosophy

- **Master-Detail Pattern**: Familiar navigation pattern with sidebar list and detail view
- **Persistent Navigation**: Sidebar always visible for quick branch switching
- **Detailed Focus**: Main area shows comprehensive information for selected branch
- **Best For**: Users who work on specific branches sequentially and need detailed information

---

## Layout Structure

```
┌─────────────────────────────────────────────────────────────────────┐
│  Allocate Staff                                    [User] [Logout]  │
├──────────┬───────────────────────────────────────────────────────────┤
│          │                                                           │
│ 📅 Date: │  ┌─────────────────────────────────────────────────────┐ │
│ [01/25/  │  │ Branch Selection:                                   │ │
│  2026 ▼] │  │ ☑ All Branches  [Filter: All ▼]  [Search: ___]   │ │
│          │  └─────────────────────────────────────────────────────┘ │
│          │                                                           │
│ ┌────────┴───────────────────────────────────────────────────────┐ │
│ │ BRANCHES (32)                    │ SEARCH: [____________]      │ │
│ ├─────────────────────────────────────────────────────────────────┤ │
│ │ 🔴 A01 - Branch A                5/8  [High]        [Selected]│ │
│ │ 🔴 B02 - Branch B                3/6  [High]                  │ │
│ │ 🟡 C03 - Branch C                4/7  [Medium]                │ │
│ │ 🟢 D04 - Branch D                7/7  [Low]                   │ │
│ │ ... (scrollable)                                              │ │
│ └─────────────────────────────────────────────────────────────────┘ │
│          │                                                           │
│          │  ┌───────────────────────────────────────────────────┐  │
│          │  │ A01 - Branch A                          [🔴 High]│  │
│          │  ├───────────────────────────────────────────────────┤  │
│          │  │ Current Staff Summary:                            │  │
│          │  │ [Detailed table with current/preferred/minimum]   │  │
│          │  │                                                   │  │
│          │  │ Allocation Suggestions:                          │  │
│          │  │ [Prioritized list with details]                  │  │
│          │  │                                                   │  │
│          │  │ [Add Rotation Staff →]                           │  │
│          │  └───────────────────────────────────────────────────┘  │
│          │                                                           │
└──────────┴───────────────────────────────────────────────────────────┘
```

---

## Component Structure

### 1. Top Control Bar
- **Date Picker**: Compact dropdown/date input (always visible)
- **Branch Selection Toggle**: "All Branches" checkbox
- **Filter Dropdown**: Filter sidebar list (All, High Priority, Needs Attention, etc.)
- **Search Box**: Filter branches in sidebar

### 2. Left Sidebar (Branch List)
- **Width**: ~300px (collapsible to ~60px)
- **Content**: Scrollable list of branches
- **Each Item Shows**:
  - Branch code and name
  - Staff ratio (current/preferred)
  - Priority badge (color-coded: 🔴 High, 🟡 Medium, 🟢 Low)
  - Visual indicator if selected
- **Features**:
  - Search box at top
  - Sort options (by priority, name, code)
  - Selected branch highlighted
  - Click to select, shows details in main area
- **Collapsible**: Can collapse to icon-only view

### 3. Main Detail Area
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

### 4. Add Staff Modal/Drawer
- **Trigger**: "Add Rotation Staff" button
- **Layout**: Centered modal or right drawer
- **Content**: Same as Alternative 2 (staff selection, assignment form)

---

## Implementation Steps (Future)

### Phase 1: Basic Layout
1. Create sidebar component with branch list
2. Create main detail area component
3. Implement branch selection state management
4. Add collapsible sidebar functionality

### Phase 2: Branch List Sidebar
1. Fetch and display all branches
2. Add search/filter functionality
3. Implement priority badge display
4. Add staff ratio calculation and display
5. Implement sorting options
6. Add selected state highlighting

### Phase 3: Detail View
1. Fetch branch details on selection
2. Display current staff breakdown table
3. Fetch and display allocation suggestions
4. Add expandable suggestion items
5. Implement "Add Rotation Staff" button

### Phase 4: Integration
1. Integrate with allocation API
2. Add staff assignment functionality
3. Implement refresh after assignment
4. Add loading states
5. Handle error states

### Phase 5: Polish
1. Add animations/transitions
2. Improve responsive design
3. Add keyboard navigation
4. Optimize performance
5. Add accessibility features

---

## API Requirements

### Endpoints Needed:
1. **GET /branches** - List all branches (already exists)
2. **POST /rotation/allocation-suggestions** - Get suggestions (already exists)
3. **GET /branches/:id/staff-summary?date=YYYY-MM-DD** - Get current staff count (may need to be created)
4. **POST /rotation/assign-from-suggestion** - Assign staff (already exists)

### Data Structures:

```typescript
interface BranchSummary {
  branch_id: string;
  branch_name: string;
  branch_code: string;
  date: string;
  staff_by_position: {
    position_id: string;
    position_name: string;
    current_count: number;
    preferred_count: number;
    minimum_count: number;
  }[];
  priority_score: number;
  priority_level: 'high' | 'medium' | 'low';
  needs_attention: boolean;
}
```

---

## Component Files to Create

### Main Components:
1. `frontend/src/app/(manager)/allocate-staff-alternative3/page.tsx` - Main page
2. `frontend/src/components/allocation/BranchSidebar.tsx` - Sidebar component
3. `frontend/src/components/allocation/BranchDetailView.tsx` - Detail view component
4. `frontend/src/components/allocation/BranchListItem.tsx` - Sidebar list item
5. `frontend/src/components/allocation/StaffSummaryTable.tsx` - Staff breakdown table

### Supporting Components:
1. `frontend/src/components/allocation/SuggestionList.tsx` - Suggestions display
2. `frontend/src/components/allocation/AddStaffDrawer.tsx` - Add staff drawer (reuse from Alternative 2)

---

## State Management

```typescript
interface AllocateStaffState {
  selectedDate: string;
  selectedBranchId: string | null;
  allBranchesSelected: boolean;
  branches: Branch[];
  branchSummaries: Map<string, BranchSummary>;
  suggestions: AllocationSuggestion[];
  loading: boolean;
  sidebarCollapsed: boolean;
  filter: {
    status: 'all' | 'needs_attention' | 'critical' | 'ok';
    priority: 'all' | 'high' | 'medium' | 'low';
    search: string;
  };
}
```

---

## Key Differences from Alternative 2

1. **Navigation**: Sidebar-based vs. grid-based
2. **Focus**: One branch at a time vs. overview of all
3. **Detail Level**: Always detailed vs. compact/detailed toggle
4. **Comparison**: Sequential vs. visual comparison
5. **Space Usage**: Sidebar takes horizontal space vs. full-width grid

---

## Advantages

- ✅ Familiar master-detail pattern
- ✅ Easy navigation between branches
- ✅ Detailed view always available
- ✅ Good for focused work on specific branches
- ✅ Sidebar provides quick overview

## Disadvantages

- ❌ Less efficient for comparing multiple branches
- ❌ Requires clicking through branches
- ❌ Sidebar takes horizontal space
- ❌ May feel slower for bulk operations

---

## When to Implement

Consider implementing Alternative 3 when:
- Users request a sidebar-based navigation
- Need for detailed branch-by-branch review workflow
- Users prefer sequential over comparative workflows
- Mobile/tablet usage requires different navigation pattern

---

## Notes

- This workflow complements Alternative 2, not replaces it
- Consider making it a toggle/view option rather than separate page
- Ensure consistent API usage with Alternative 2
- Reuse components where possible (AddStaffDrawer, etc.)
- Maintain same data structures and API contracts

---

## References

- See `docs/allocate-staff-workflow-alternatives.md` for full design details
- See `docs/allocate-staff-workflow-mockups.md` for visual mockups
- See `frontend/src/app/(manager)/allocate-staff/page.tsx` for Alternative 2 implementation

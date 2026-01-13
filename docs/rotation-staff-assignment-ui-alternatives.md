---
title: Rotation Staff Assignment UI Alternatives
description: Two alternative UI/UX approaches for manually assigning rotation staff to branches
version: 1.0.0
lastUpdated: 2025-12-23 13:14:23
---

# Rotation Staff Assignment UI Alternatives

## Overview

This document presents two alternative UI/UX approaches for implementing the manual rotation staff assignment feature, where Admin and Area Manager roles can assign rotation staff to branches by selecting dates.

**Requirements:**
- Only Admin and Area Manager can assign rotation staff
- Rotation staff eligible for a branch are automatically populated (based on effective branches)
- Admin/Area Manager selects which date(s) a rotation staff will operate for that branch
- Rotation staff appear as new rows in the branch table alongside branch's proprietary staff

---

## Alternative 1: Branch-Centric Table View with Inline Date Selection

### Concept

A table-based interface where:
1. User selects a branch first
2. Table shows branch staff (proprietary) + eligible rotation staff (as separate rows)
3. Each row has date columns (calendar days)
4. Click on date cells to assign/unassign rotation staff
5. Visual distinction between branch staff and rotation staff rows

### UI Structure

```
┌─────────────────────────────────────────────────────────────────────────┐
│ Rotation Staff Assignment - Branch: [Branch Selector ▼]                 │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  Month: [◀ Dec 2024 ▶]  [Today]                                       │
│                                                                         │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │ Staff Name          │ Position │ Mon 1 │ Tue 2 │ Wed 3 │ ... │  │  │
│  ├─────────────────────┼──────────┼───────┼───────┼───────┼─────┤  │
│  │ John Doe            │ Manager  │   ✓   │   ✓   │   ✓   │ ... │  │  │
│  │ Jane Smith          │ Nurse    │   ✓   │   ✓   │   -   │ ... │  │  │
│  │ [Branch Staff]      │          │       │       │       │     │  │  │
│  ├─────────────────────┼──────────┼───────┼───────┼───────┼─────┤  │
│  │ + Add Rotation Staff│          │       │       │       │     │  │  │
│  ├─────────────────────┼──────────┼───────┼───────┼───────┼─────┤  │
│  │ Alice Rotation      │ Nurse    │   ○   │   ○   │   ○   │ ... │  │  │
│  │ [Rotation Staff]   │          │       │       │       │     │  │  │
│  │ Bob Rotation        │ Doctor   │   ○   │   ○   │   ○   │ ... │  │  │
│  │ [Rotation Staff]   │          │       │       │       │     │  │  │
│  └─────────────────────┴──────────┴───────┴───────┴───────┴─────┘  │
│                                                                         │
│  Legend: ✓ = Assigned, ○ = Available, - = Not Available               │
│                                                                         │
│  [Save Changes] [Cancel]                                               │
└─────────────────────────────────────────────────────────────────────────┘
```

### Key Features

1. **Branch Selection Dropdown**
   - Top-level branch selector
   - Shows branch name and code
   - Filters all data to selected branch

2. **Unified Staff Table**
   - **Section 1:** Branch Staff (proprietary staff)
     - Read-only display (managed in staff management)
     - Shows existing schedules
   - **Section 2:** Eligible Rotation Staff
     - Automatically populated based on effective branches
     - Shows all rotation staff eligible for selected branch
     - Each rotation staff as a separate row
     - Visual indicator (badge/icon) showing "Rotation Staff"

3. **Date Columns**
   - Calendar days as columns (left to right: closer dates first)
   - Each cell shows assignment status:
     - **Empty/○**: Not assigned (clickable)
     - **✓/Filled**: Assigned (clickable to remove)
     - **-**: Not eligible (grayed out, not clickable)

4. **Interaction**
   - Click date cell to toggle assignment
   - Hover shows tooltip with date and staff info
   - Bulk selection: Select date range, assign to multiple staff
   - Quick actions: "Assign for week", "Assign for month"

5. **Visual Design**
   - Branch staff rows: Standard background
   - Rotation staff rows: Light blue/cyan background or border
   - Assigned dates: Colored background (green/blue)
   - Hover effects on clickable cells

### Component Structure

```typescript
// frontend/src/components/rotation/BranchRotationAssignmentTable.tsx

interface BranchRotationAssignmentTableProps {
  branchId: string;
  month: Date;
}

interface StaffRow {
  id: string;
  name: string;
  position: string;
  staffType: 'branch' | 'rotation';
  assignments: Set<string>; // Set of date strings (YYYY-MM-DD)
}

// Main component structure:
- BranchSelector (dropdown)
- MonthNavigator (prev/next month, today button)
- StaffTable
  - StaffRow (for each staff member)
    - StaffInfo (name, position, type badge)
    - DateCells (array of DateCell components)
      - DateCell (clickable, shows assignment status)
- ActionButtons (Save, Cancel, Bulk Actions)
```

### User Flow

1. **Select Branch**
   - User opens page
   - Selects branch from dropdown
   - System loads:
     - Branch staff (from staff table)
     - Eligible rotation staff (from effective_branches)
     - Existing assignments

2. **View Current Assignments**
   - Table displays all staff (branch + rotation)
   - Assigned dates show checkmark/filled
   - Unassigned dates show empty/available indicator

3. **Assign Rotation Staff**
   - User clicks on empty date cell for rotation staff row
   - Cell changes to assigned state (checkmark/filled)
   - Assignment saved (optimistic update or on Save button)

4. **Remove Assignment**
   - User clicks on assigned date cell
   - Confirmation dialog (optional)
   - Assignment removed

5. **Bulk Operations**
   - Select date range (drag or shift-click)
   - Right-click or use bulk action menu
   - Assign/remove for selected dates

### Pros

✅ **Intuitive:** Familiar table structure, easy to understand  
✅ **Efficient:** See all staff and dates in one view  
✅ **Clear Visual Distinction:** Branch vs rotation staff clearly separated  
✅ **Bulk Operations:** Easy to assign multiple dates at once  
✅ **Matches Requirements:** Rotation staff appear as rows in branch table  
✅ **Mobile-Friendly:** Can be made responsive with horizontal scroll  

### Cons

❌ **Horizontal Scrolling:** Many date columns require horizontal scroll  
❌ **Limited Date Range:** Hard to show full year (365 days)  
❌ **Small Cells:** Date cells might be too small for easy clicking  
❌ **Performance:** Rendering many cells can be slow  

### Implementation Details

**Backend API Requirements:**
```typescript
// GET /api/branches/:branchId/staff-with-eligible-rotation
// Returns: {
//   branch_staff: Staff[],
//   eligible_rotation_staff: Staff[],
//   existing_assignments: RotationAssignment[]
// }

// POST /api/rotation/assign-bulk
// Body: {
//   rotation_staff_id: string,
//   branch_id: string,
//   dates: string[] // ['2024-12-01', '2024-12-02', ...]
// }
```

**State Management:**
- Local state for assignments (optimistic updates)
- Save button commits all changes
- Or auto-save on each cell click

**Performance Optimizations:**
- Virtual scrolling for date columns
- Lazy load assignments for visible dates only
- Debounce save operations

---

## Alternative 2: Branch Schedule View with Rotation Staff Panel

### Concept

A calendar/schedule view similar to branch staff scheduling, where:
1. Main view shows branch staff schedule (existing functionality)
2. Side panel or expandable section shows eligible rotation staff
3. Drag-and-drop or click-to-assign rotation staff to dates
4. Rotation staff assignments appear inline with branch staff

### UI Structure

```
┌─────────────────────────────────────────────────────────────────────────┐
│ Rotation Staff Assignment - Branch: [Branch Selector ▼]                 │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  Month: [◀ Dec 2024 ▶]  [Today]  [View: Month | Week | Day]            │
│                                                                         │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │                    │ Mon 1 │ Tue 2 │ Wed 3 │ Thu 4 │ ... │      │  │
│  ├────────────────────┼───────┼───────┼───────┼───────┼─────┤      │  │
│  │ John Doe           │   ✓   │   ✓   │   ✓   │   ✓   │ ... │      │  │
│  │ [Branch Staff]     │       │       │       │       │     │      │  │
│  ├────────────────────┼───────┼───────┼───────┼───────┼─────┤      │  │
│  │ Jane Smith         │   ✓   │   ✓   │   -   │   ✓   │ ... │      │  │
│  │ [Branch Staff]     │       │       │       │       │     │      │  │
│  ├────────────────────┼───────┼───────┼───────┼───────┼─────┤      │  │
│  │ + Add Rotation     │       │       │       │       │     │      │  │
│  │   Staff            │       │       │       │       │     │      │  │
│  ├────────────────────┼───────┼───────┼───────┼───────┼─────┤      │  │
│  │ Alice Rotation     │   ○   │   ○   │   ○   │   ○   │ ... │      │  │
│  │ [Rotation Staff]   │       │       │       │       │     │      │  │
│  │                    │ [Drag]│ [Drag]│ [Drag]│ [Drag]│     │      │  │
│  ├────────────────────┼───────┼───────┼───────┼───────┼─────┤      │  │
│  │ Bob Rotation       │   ○   │   ○   │   ○   │   ○   │ ... │      │  │
│  │ [Rotation Staff]   │       │       │       │       │     │      │  │
│  └────────────────────┴───────┴───────┴───────┴───────┴─────┴──────┘  │
│                                                                         │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │ Eligible Rotation Staff Panel                                    │  │
│  ├──────────────────────────────────────────────────────────────────┤  │
│  │ 🔍 [Search rotation staff...]                                    │  │
│  │                                                                  │  │
│  │ ┌──────────────────────────────────────────────────────────┐  │  │
│  │ │ Alice Rotation - Nurse                                    │  │  │
│  │ │ Coverage: Area A | Level 1: CPN, CTR                      │  │  │
│  │ │ [Add to Schedule]                                         │  │  │
│  │ └──────────────────────────────────────────────────────────┘  │  │  │
│  │                                                                  │  │
│  │ ┌──────────────────────────────────────────────────────────┐  │  │
│  │ │ Bob Rotation - Doctor                                     │  │  │
│  │ │ Coverage: Area B | Level 1: PNK, CNK                     │  │  │
│  │ │ [Add to Schedule]                                         │  │  │
│  │ └──────────────────────────────────────────────────────────┘  │  │  │
│  │                                                                  │  │
│  │ ... (more rotation staff)                                       │  │  │
│  └──────────────────────────────────────────────────────────────────┘  │
│                                                                         │
│  [Save Changes] [Cancel] [Bulk Assign]                                  │
└─────────────────────────────────────────────────────────────────────────┘
```

### Key Features

1. **Main Schedule View**
   - Similar to existing branch staff schedule view
   - Shows branch staff rows (read-only)
   - Shows rotation staff rows (editable)
   - Date columns (calendar days)

2. **Eligible Rotation Staff Panel**
   - Side panel or collapsible section
   - Lists all rotation staff eligible for selected branch
   - Shows staff info: name, position, coverage area, effective branches
   - Search/filter functionality
   - "Add to Schedule" button for each staff

3. **Assignment Methods**
   - **Method 1: Click-to-Assign**
     - Click "Add to Schedule" → Staff row appears in schedule
     - Click date cells to assign/unassign
   - **Method 2: Drag-and-Drop**
     - Drag rotation staff card to date cell
     - Visual feedback during drag
   - **Method 3: Quick Assign Dialog**
     - Click "Add to Schedule" → Dialog opens
     - Select date range
     - Confirm assignment

4. **Rotation Staff Row Behavior**
   - When added, appears as new row in schedule
   - Shows all dates (available/assigned)
   - Can be removed from schedule (hides row but keeps assignments)
   - Visual distinction from branch staff

5. **Bulk Assignment Dialog**
   - Select multiple rotation staff
   - Select date range
   - Assign all at once
   - Preview before confirming

### Component Structure

```typescript
// frontend/src/components/rotation/BranchScheduleWithRotation.tsx

interface BranchScheduleWithRotationProps {
  branchId: string;
  month: Date;
}

// Main component structure:
- BranchSelector
- MonthNavigator
- ViewModeToggle (Month/Week/Day)
- ScheduleGrid
  - BranchStaffRows (read-only)
  - RotationStaffRows (editable, can be added/removed)
    - DateCells (clickable)
- EligibleRotationStaffPanel
  - SearchBar
  - RotationStaffCard[] (each card has "Add to Schedule" button)
- AssignmentDialog (for quick assign)
- BulkAssignmentDialog
```

### User Flow

1. **Select Branch**
   - User selects branch
   - Schedule loads branch staff
   - Panel loads eligible rotation staff

2. **Add Rotation Staff to Schedule**
   - User clicks "Add to Schedule" on rotation staff card
   - Option A: Staff row appears in schedule immediately
   - Option B: Dialog opens to select dates first, then row appears

3. **Assign Dates**
   - User clicks date cells in rotation staff row
   - Dates toggle between assigned/unassigned
   - Visual feedback (checkmark, color change)

4. **Remove Rotation Staff**
   - Option A: Remove row (hides from schedule, keeps assignments)
   - Option B: Remove all assignments (row disappears)

5. **Bulk Operations**
   - Select multiple staff from panel
   - Click "Bulk Assign"
   - Select date range
   - Confirm → All assignments created

### Pros

✅ **Familiar Interface:** Similar to existing branch schedule view  
✅ **Flexible:** Multiple assignment methods (click, drag, dialog)  
✅ **Clear Separation:** Panel clearly shows eligible staff  
✅ **Scalable:** Can handle many rotation staff without cluttering schedule  
✅ **Searchable:** Panel can be filtered/searched  
✅ **Progressive Disclosure:** Only show rotation staff when needed  

### Cons

❌ **More Clicks:** Requires adding staff to schedule before assigning  
❌ **Two-Step Process:** Add to schedule, then assign dates  
❌ **Panel Space:** Requires side panel or collapsible section  
❌ **Less Overview:** Harder to see all assignments at once  

### Implementation Details

**Backend API Requirements:**
```typescript
// GET /api/branches/:branchId/eligible-rotation-staff
// Returns: Staff[] (with effective branch info)

// POST /api/rotation/assign
// Body: {
//   rotation_staff_id: string,
//   branch_id: string,
//   date: string,
//   assignment_level: number
// }

// POST /api/rotation/assign-bulk
// Body: {
//   assignments: [{
//     rotation_staff_id: string,
//     branch_id: string,
//     dates: string[]
//   }]
// }
```

**State Management:**
- Separate state for:
  - Branch staff (read-only)
  - Rotation staff in schedule (editable)
  - Eligible rotation staff panel (filterable)
- Optimistic updates for assignments

**Drag-and-Drop:**
- Use react-dnd or similar library
- Visual feedback during drag
- Drop zones on date cells

---

## Comparison Matrix

| Aspect | Alternative 1: Table View | Alternative 2: Schedule + Panel |
|--------|---------------------------|-------------------------------|
| **Initial Setup** | Simple - just select branch | Two-step - select branch, add staff |
| **Visual Clarity** | ✅ All staff visible at once | ⚠️ Panel separate from schedule |
| **Ease of Use** | ✅ Direct click to assign | ⚠️ Requires adding staff first |
| **Bulk Operations** | ✅ Easy with date range selection | ✅ Good with bulk dialog |
| **Mobile Experience** | ⚠️ Horizontal scroll needed | ✅ Panel can be modal on mobile |
| **Performance** | ⚠️ Many cells to render | ✅ Only render visible staff |
| **Scalability** | ⚠️ Table grows with many staff | ✅ Panel handles many staff well |
| **Matches Requirements** | ✅ Rotation staff as rows | ✅ Rotation staff as rows |
| **Learning Curve** | ✅ Intuitive table | ⚠️ Requires understanding panel |
| **Date Range** | ⚠️ Limited by horizontal space | ✅ Can show full month easily |

---

## Recommendation

**Alternative 1 (Branch-Centric Table View)** is recommended because:

1. **Direct and Intuitive:** Matches the requirement exactly - rotation staff appear as rows in the branch table
2. **Efficient Workflow:** Single view shows everything, direct assignment
3. **Familiar Pattern:** Table-based interfaces are well-understood
4. **Better for Bulk Operations:** Easy to select date ranges and assign multiple staff
5. **Clearer Overview:** All assignments visible at once

**However, consider a hybrid approach:**
- Use Alternative 1 as the primary interface
- Add a "Quick Add" feature from Alternative 2:
  - Small button/icon to add rotation staff row
  - Opens dialog to select from eligible staff
  - Adds selected staff as new row in table

---

## Implementation Checklist

### Phase 1: Backend API
- [ ] GET endpoint for branch staff + eligible rotation staff
- [ ] GET endpoint for existing assignments
- [ ] POST endpoint for bulk assignment
- [ ] DELETE endpoint for assignment removal
- [ ] Validation: Check effective branches before assignment

### Phase 2: Core UI Components
- [ ] BranchSelector component
- [ ] MonthNavigator component
- [ ] StaffTable component
- [ ] StaffRow component (branch vs rotation)
- [ ] DateCell component (clickable, shows status)

### Phase 3: Assignment Logic
- [ ] Click handler for date cells
- [ ] Toggle assignment state
- [ ] Visual feedback (loading, success, error)
- [ ] Optimistic updates

### Phase 4: Bulk Operations
- [ ] Date range selection
- [ ] Multi-staff selection
- [ ] Bulk assign/remove functionality

### Phase 5: Polish
- [ ] Loading states
- [ ] Error handling
- [ ] Confirmation dialogs
- [ ] Tooltips and help text
- [ ] Responsive design

---

## Related Requirements

- **FR-SM-02:** Rotation Staff Assignment
- **FR-AUZ-01:** Role-Based Access Control (Admin and Area Manager only)
- **FR-BL-02:** Effective Branch Management (eligible staff determination)

---

## Notes

- Both alternatives can be made responsive for mobile devices
- Consider adding keyboard shortcuts for power users
- Add undo/redo functionality for assignment changes
- Consider adding conflict detection/warnings
- Add export/print functionality for schedule view



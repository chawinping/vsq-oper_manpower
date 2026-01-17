---
name: Enhanced Role-Based Allocation System
overview: "Implement major enhancements to the allocation system with role-specific features: Area Managers get criteria-based allocation suggestions and overview views, Branch Managers can allocate monthly and designate doctor-on/off days, and Administrators can configure allocation criteria. Includes new quota management, 3-pillar criteria system, and enhanced overview interfaces."
todos:
  - id: db-models
    content: Create database migrations and domain models for allocation criteria, position quotas, doctor assignments, and allocation suggestions
    status: completed
  - id: criteria-engine
    content: Implement criteria evaluation engine supporting clinic-wide, doctor-specific, and branch-specific pillars
    status: completed
    dependencies:
      - db-models
  - id: quota-system
    content: Implement position quota management system with designated quota per position per branch
    status: completed
    dependencies:
      - db-models
  - id: doctor-assignment
    content: Implement doctor assignment system with doctor-on/off day designation
    status: completed
    dependencies:
      - db-models
  - id: suggestion-engine
    content: Build automated suggestion generation engine using criteria and quota data
    status: completed
    dependencies:
      - criteria-engine
      - quota-system
      - doctor-assignment
  - id: overview-apis
    content: Create API endpoints for all-branches day-by-day overview and single-branch monthly overview
    status: completed
    dependencies:
      - quota-system
      - doctor-assignment
  - id: area-manager-ui
    content: "Build Area Manager UI: allocation overview tables, suggestion approval panel, adhoc allocation dialog"
    status: completed
    dependencies:
      - overview-apis
      - suggestion-engine
  - id: branch-manager-ui
    content: "Build Branch Manager UI: monthly allocation interface, doctor-on/off editor, enhanced staff management"
    status: completed
    dependencies:
      - doctor-assignment
  - id: admin-criteria-ui
    content: Build Admin UI for configuring allocation criteria (3 pillars with weights)
    status: completed
    dependencies:
      - criteria-engine
  - id: integration-testing
    content: Integrate all components, add tests, optimize performance, and update documentation
    status: completed
    dependencies:
      - area-manager-ui
      - branch-manager-ui
      - admin-criteria-ui
---

# Enhanced Role-Based Allocation System Implementation Plan

## Overview

This plan implements major enhancements to the VSQ Operations Manpower system, focusing on role-specific allocation workflows, criteria-based automation, and comprehensive overview views. The implementation builds upon the existing allocation and scheduling infrastructure.

## Architecture Changes

### Database Schema Additions

1. **Allocation Criteria Configuration** (`allocation_criteria`)

- Store configurable criteria weights for clinic-wide, doctor-specific, and branch-specific pillars
- Support multiple criteria types (bookings, revenue, staff requirements, doctor count)

2. **Position Quota Management** (`position_quotas`)

- Store designated quota per position per branch
- Track preferred vs minimum requirements

3. **Doctor Assignment** (`doctor_assignments`)

- Track doctor assignments per branch per day
- Support doctor-on/doctor-off day designation
- Link to expected revenue calculations

4. **Allocation Suggestions** (`allocation_suggestions`)

- Store system-generated allocation suggestions
- Track suggestion status (pending, approved, rejected, modified)
- Link to criteria used for generation

5. **Adhoc Allocation Tracking** (`adhoc_allocations`)

- Mark allocations as adhoc (unplanned leave scenarios)
- Track reason for adhoc allocation

### Backend Changes

#### New Models

- `backend/internal/domain/models/allocation_criteria.go` - Criteria configuration
- `backend/internal/domain/models/position_quota.go` - Position quota management
- `backend/internal/domain/models/doctor_assignment.go` - Doctor assignments
- `backend/internal/domain/models/allocation_suggestion.go` - Allocation suggestions

#### Enhanced Use Cases

- `backend/internal/usecases/allocation/criteria_engine.go` - Criteria evaluation engine
- `backend/internal/usecases/allocation/suggestion_engine.go` - Automated suggestion generation
- `backend/internal/usecases/allocation/quota_calculator.go` - Quota and requirement calculations
- `backend/internal/usecases/allocation/overview_generator.go` - Overview view data aggregation

#### New Handlers

- `backend/internal/handlers/allocation_criteria_handler.go` - Criteria configuration (Admin only)
- `backend/internal/handlers/quota_handler.go` - Quota management
- `backend/internal/handlers/doctor_handler.go` - Doctor assignment management
- `backend/internal/handlers/overview_handler.go` - Overview views for Area Managers
- Enhanced `backend/internal/handlers/rotation_handler.go` - Add suggestion approval/rejection workflow

#### Enhanced Repositories

- Extend `backend/internal/repositories/postgres/repositories.go` with new repository interfaces
- Add methods for quota, doctor assignments, criteria, and suggestions

### Frontend Changes

#### New Pages

- `frontend/src/app/(admin)/allocation-criteria/page.tsx` - Criteria configuration (Admin)
- `frontend/src/app/(manager)/allocation-overview/page.tsx` - All branches day-by-day overview (Area Manager)
- `frontend/src/app/(manager)/branch-overview/page.tsx` - Single branch monthly overview
- `frontend/src/app/(manager)/doctor-schedule/page.tsx` - Doctor-on/off designation (Branch Manager)
- Enhanced `frontend/src/app/(manager)/rotation-scheduling/page.tsx` - Add suggestion approval UI

#### New Components

- `frontend/src/components/allocation/AllocationOverviewTable.tsx` - Multi-branch overview table
- `frontend/src/components/allocation/BranchOverviewCalendar.tsx` - Single branch monthly view
- `frontend/src/components/allocation/SuggestionPanel.tsx` - Suggestion review and approval
- `frontend/src/components/allocation/QuotaDisplay.tsx` - Quota visualization
- `frontend/src/components/allocation/AdhocAllocationDialog.tsx` - Adhoc allocation interface
- `frontend/src/components/doctor/DoctorScheduleEditor.tsx` - Doctor-on/off day editor

#### Enhanced Components

- `frontend/src/components/rotation/RotationAssignmentView.tsx` - Add suggestion approval workflow
- `frontend/src/components/scheduling/MonthlyCalendar.tsx` - Add doctor-on/off indicators

## Implementation Phases

### Phase 1: Foundation - Database and Models

1. Create database migrations for new tables
2. Implement domain models for new entities
3. Create repository interfaces and implementations
4. Add validation and business rule enforcement

### Phase 2: Allocation Criteria System

1. Implement criteria configuration models and API (Admin only)
2. Build criteria evaluation engine (3 pillars: clinic-wide, doctor-specific, branch-specific)
3. Create criteria weight configuration UI
4. Implement criteria calculation logic

### Phase 3: Quota Management

1. Implement position quota models and API
2. Create quota management UI (Admin/Area Manager)
3. Build quota calculation logic
4. Integrate quota into allocation suggestions

### Phase 4: Doctor Assignment System

1. Implement doctor assignment models and API
2. Create doctor-on/off day designation UI (Branch Manager)
3. Build doctor assignment validation
4. Link doctor assignments to operational status

### Phase 5: Enhanced Allocation Engine

1. Enhance allocation engine with criteria-based logic
2. Implement suggestion generation algorithm
3. Add suggestion approval/rejection workflow
4. Build adhoc allocation tracking

### Phase 6: Overview Views

1. Implement all-branches day-by-day overview API
2. Implement single-branch monthly overview API
3. Create overview UI components with drill-down
4. Add quota/available/assigned/required calculations

### Phase 7: Branch Manager Enhancements

1. Enhance monthly allocation interface
2. Add doctor-on/off day designation
3. Improve staff profile management UI

### Phase 8: Integration and Testing

1. Integrate all components
2. Add comprehensive tests
3. Performance optimization for overview views
4. Documentation updates

## Key Features by Role

### Area Manager Features

- **Criteria-Based Suggestions**: System suggests allocations based on configurable criteria (90% automation)
- **Suggestion Review**: Review, approve, or modify suggestions
- **Adhoc Allocation**: Handle unplanned leave scenarios
- **All Branches Overview**: Day-by-day view of all 30+ branches with drill-down
- **Single Branch Overview**: Monthly view of a branch with detailed breakdown
- **Staff Management**: Add/edit/delete staff profiles for branches

### Branch Manager Features

- **Monthly Allocation**: Allocate branch staff for entire month
- **Branch Overview**: View own branch's allocation for entire month
- **Doctor-On/Off Designation**: Set which days are doctor-on or doctor-off
- **Staff Profile Management**: Manage branch staff profiles

### Administrator Features

- **All Area Manager Features**: Full access to all Area Manager capabilities
- **Criteria Configuration**: Set and configure allocation criteria (3 pillars)
- **System Configuration**: Configure all system settings

## Technical Considerations

### Performance

- Overview views need efficient aggregation queries
- Consider caching for frequently accessed overview data
- Optimize database queries with proper indexes

### Data Integrity

- Validate quota configurations
- Ensure doctor assignments don't conflict
- Enforce business rules in allocation logic

### User Experience

- Provide clear visual indicators for quota status
- Show missing staff requirements prominently
- Make suggestion approval workflow intuitive
- Support bulk operations where appropriate

## Files to Create/Modify

### Backend

- New: `backend/internal/domain/models/allocation_criteria.go`
- New: `backend/internal/domain/models/position_quota.go`
- New: `backend/internal/domain/models/doctor_assignment.go`
- New: `backend/internal/domain/models/allocation_suggestion.go`
- New: `backend/internal/usecases/allocation/criteria_engine.go`
- New: `backend/internal/usecases/allocation/suggestion_engine.go`
- New: `backend/internal/usecases/allocation/quota_calculator.go`
- New: `backend/internal/usecases/allocation/overview_generator.go`
- New: `backend/internal/handlers/allocation_criteria_handler.go`
- New: `backend/internal/handlers/quota_handler.go`
- New: `backend/internal/handlers/doctor_handler.go`
- New: `backend/internal/handlers/overview_handler.go`
- Modify: `backend/internal/handlers/rotation_handler.go`
- Modify: `backend/internal/repositories/postgres/repositories.go`
- Modify: `backend/internal/repositories/postgres/migrations.go`

### Frontend

- New: `frontend/src/app/(admin)/allocation-criteria/page.tsx`
- New: `frontend/src/app/(manager)/allocation-overview/page.tsx`
- New: `frontend/src/app/(manager)/branch-overview/page.tsx`
- New: `frontend/src/app/(manager)/doctor-schedule/page.tsx`
- New: `frontend/src/components/allocation/AllocationOverviewTable.tsx`
- New: `frontend/src/components/allocation/BranchOverviewCalendar.tsx`
- New: `frontend/src/components/allocation/SuggestionPanel.tsx`
- New: `frontend/src/components/allocation/QuotaDisplay.tsx`
- New: `frontend/src/components/allocation/AdhocAllocationDialog.tsx`
- New: `frontend/src/components/doctor/DoctorScheduleEditor.tsx`
- Modify: `frontend/src/components/rotation/RotationAssignmentView.tsx`
- Modify: `frontend/src/components/scheduling/MonthlyCalendar.tsx`
- New: `frontend/src/lib/api/allocation-criteria.ts`
- New: `frontend/src/lib/api/quota.ts`
- New: `frontend/src/lib/api/doctor.ts`
- New: `frontend/src/lib/api/overview.ts`

### Documentation

- Update: `SOFTWARE_REQUIREMENTS.md` - Add new requirements
- Update: `SOFTWARE_ARCHITECTURE.md` - Document new components
- Update: `CHANGELOG.md` - Document changes

## Success Criteria

1. Area Managers can review and approve automated allocation suggestions
2. Area Managers can view comprehensive overviews of all branches
3. Branch Managers can designate doctor-on/off days
4. Administrators can configure allocation criteria
5. System generates suggestions based on configurable 3-pillar criteria
6. Overview views show quota, available, assigned, and required staff counts
7. Adhoc allocation workflow handles unplanned leave scenarios
8. All role-based access controls are properly enforced
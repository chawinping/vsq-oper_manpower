---
title: Software Requirements Specification
description: Detailed requirements for VSQ Operations Manpower System
version: 1.16.0
lastUpdated: 2026-01-17 20:34:25
---

# VSQ Operations Manpower - Software Requirements Specification

## Document Information

- **Version:** 1.16.0
- **Last Updated:** 2026-01-17 20:34:25
- **Status:** Active

## Related Documents

- **Security Requirements:** See `docs/security-requirements.md` for detailed security requirements
- **Conflict Resolution Rules:** See `docs/conflict-resolution-rules.md` for conflict handling rules
- **Business Rules:** See `docs/business-rules.md` for detailed business rules
- **Requirements Analysis:** See `docs/requirements-analysis.md` for gap analysis

## 1. Introduction

### 1.1 Purpose
This document specifies the functional and non-functional requirements for the VSQ Operations Manpower System, a web application designed to maximize the efficiency of allocating staff across 32 branches of an aesthetic clinic.

### 1.2 Scope
The system manages staff allocation, scheduling, and rotation assignments based on expected revenue and business rules. It supports multiple user roles and provides AI-powered suggestions for optimal staff allocation.

### 1.3 Definitions and Acronyms
- **Branch Staff:** Staff members fixed to a specific branch
- **Rotation Staff:** Staff members who can be assigned to multiple branches
- **Area Manager:** Manager responsible for multiple branches in an area
- **District Manager:** Manager responsible for multiple areas
- **MCP:** Model Context Protocol (for AI suggestions)

## 2. System Overview

### 2.1 Main Objective
To maximize the efficiency of allocating staff for branches of an aesthetic clinic with 32 branches. The goal is to ensure branches have enough staff (by the use of rotation staff) according to expected revenue for each day.

### 2.2 System Architecture
- **Backend:** Go (Gin framework) with Clean Architecture
- **Frontend:** Next.js with TypeScript (SSR/SSG/Client)
- **Database:** PostgreSQL
- **Deployment:** Docker containers
- **Testing:** Unit tests (Go) and E2E tests (Playwright)

## 3. Functional Requirements

### 3.1 User Management and Authentication

#### FR-AU-01: User Authentication
- **Description:** System shall provide session-based authentication
- **Status:** ✅ Implemented
- **Details:**
  - Users can login with username and password
  - Sessions are managed using secure cookies
  - Password hashing using bcrypt
  - Session timeout after 7 days of inactivity

#### FR-AUZ-01: Role-Based Access Control
- **Description:** System shall enforce role-based access control
- **Status:** ✅ Implemented (enhanced with new allocation features)
- **Roles:**
  1. **Admin:** Can set system configurations, manage users and roles. Can view/add/edit/delete/import rotation staff. Can assign rotation staff to branches. Can edit rotation staff schedules. Can configure allocation criteria (3 pillars). Can manage position quotas. Can view all allocation overviews. Can review and approve allocation suggestions. Can create adhoc allocations. Can manage doctor profiles and schedules.
  2. **Area Manager:** Can assign rotation staff to branches. Can view/add/edit/delete/import rotation staff. Can edit rotation staff schedules. Can view allocation overviews (all branches day-by-day, single branch monthly). Can review and approve allocation suggestions. Can create adhoc allocations. Can manage position quotas for branches in their area. Can manage doctor profiles and schedules.
  3. **District Manager:** Can view rotation staff assignments (read-only). Cannot manage rotation staff or edit schedules. Can view allocation overviews (read-only).
  4. **Branch Manager:** Can allocate staff workday, off day, and leave day in their branch. Can add, edit, and manage staff in their branch. Can only access staff and schedules from their assigned branch (enforced by branch code). Can view rotation staff assigned to their branch (read-only). Can designate doctor-on/doctor-off days for their branch. Can view their branch's monthly allocation overview.
  5. **Viewer:** Can only view staff allocation and dashboard data
- **Branch Manager Restrictions:**
  - Branch managers are linked to a branch via `branch_id` field in users table
  - Staff listings automatically filtered to show only their branch staff
  - Schedule operations restricted to their assigned branch
  - Cannot access or modify staff from other branches
  - Access enforced at API level through middleware and handler validation

### 3.2 Staff Management

#### FR-RM-01: Staff Types
- **Description:** System shall support two types of staff
- **Status:** ✅ Implemented
- **Types:**
  1. **Branch Staff:** Fixed to a specific branch
  2. **Rotation Staff:** Not fixed to any branch, can be assigned to multiple branches

#### FR-RM-02: Staff Positions
- **Description:** System shall support various staff positions with position type classification
- **Status:** ✅ Implemented
- **Position Types:**
  - **Branch Positions:** Positions that are assigned to specific branches and can have quota configurations (minimum/preferred staff counts)
  - **Rotation Positions:** Positions that are used for rotation staff assignments and cannot have branch-specific quota configurations
- **Position Classification:**
  - Each position must be classified as either "branch" or "rotation" type
  - Position type determines whether the position can have branch quota configurations
  - Only branch-type positions appear in branch configuration quota settings
  - Rotation-type positions are managed through rotation staff assignments, not fixed quotas
- **Positions:**
  - Branch Manager (branch type)
  - Assistant Branch Manager (branch type)
  - Service Consultant (branch type)
  - Coordinator (branch type)
  - Doctor Assistant (branch type)
  - Physiotherapist (branch type)
  - Nurses (branch type)
  - Front positions (branch type)
  - Laser Assistant (branch type)
  - District Manager (rotation type)
  - Department Manager & Branch Development Supervisor (rotation type)
  - Head Doctor Assistant (rotation type)
  - Special Assistant (rotation type)
  - Rotation positions (e.g., Front Rotation, Doctor Assistant Rotation) (rotation type)
  - (Extensible for future positions)

#### FR-RM-03: Staff Data Management
- **Description:** System shall allow manual entry and bulk import of staff data
- **Status:** ⚠️ Partially Implemented
- **Details:**
  - Manual entry via UI (✅ Implemented)
  - Excel import for bulk data (❌ Not Implemented - placeholder exists)
  - Staff information includes: nickname, full name, type, position, branch (for branch staff), coverage area (for rotation staff)
  - Staff addition/deletion is done in a separate module from schedule arrangement (✅ Implemented - separate pages)
  - Branch manager can add additional staff to their branch
- **Branch Manager Restrictions:**
  - Manual entry is available for branch manager role but only for branch staff of that branch
  - Branch manager role cannot add entry for rotation staff
  - Branch manager can only create, edit, and delete branch staff assigned to their own branch
- **Rotation Staff Management Permissions:**
  - Only Admin and Area Manager roles can view/add/edit/delete/import rotation staff
  - District Manager role does NOT have rotation staff management permissions (only Admin and Area Manager)
  - Rotation staff CRUD operations are restricted at API level with role-based access control

#### FR-RM-04: Staff Skills and Qualifications
- **Description:** System shall track staff skills and qualifications
- **Status:** ❌ Not Implemented
- **Details:**
  - Staff can have multiple skills/qualifications
  - Skills include: "can manage a branch alone" and other relevant skills
  - Skills are used to determine staff eligibility for specific assignments
  - "Can manage a branch alone" skill is required for staff assigned to branches with no doctor (doctor-off days)
  - Skills can be assigned to both branch staff and rotation staff
  - Skills are used as constraints when allocating rotation staff
  - Skills can be manually assigned or imported from Excel

### 3.3 Branch Management

#### FR-BM-01: Branch Configuration
- **Description:** System shall manage 32 branches
- **Status:** ✅ Implemented
- **Details:**
  - Each branch has: name, code, address
  - Expected revenue can be set per day (see FR-BM-02)
  - Priority level can be assigned
  - Area Manager can be assigned to branches
  - Branch operational status can vary by day (see FR-BM-04)

#### FR-BM-04: Branch Operational Status
- **Description:** System shall track branch operational status on a daily basis
- **Status:** ⚠️ Partially Implemented (doctor-on/off day designation implemented, full operational status tracking pending)
- **Details:**
  - Branch operational status can vary by day
  - A branch may be non-operational on specific days due to:
    1. **No Doctor:** Branch has no doctor assigned on that day (doctor-off day)
    2. **Branch Closed:** Branch is closed on that day (holiday, maintenance, etc.)
  - **Operational Status Values:**
    - **Operational:** Branch is open and has at least one doctor assigned (doctor-on day)
    - **No Doctor:** Branch is open but has no doctor assigned (doctor-off day, non-operational)
    - **Closed:** Branch is closed on that day (non-operational)
  - **Doctor-On/Doctor-Off Days:**
    - Branches typically have 1-6 doctors assigned per day (doctor-on days)
    - When a branch has no doctor assigned (doctor-off day), only one staff member is required
    - The single staff member on doctor-off days must have the "can manage a branch alone" skill
    - Doctor assignments determine whether a day is doctor-on or doctor-off
    - Maximum 6 doctors can be assigned to a branch per day
  - **Business Rules:**
    - Non-operational branches should not require staff allocation (or require minimal staff)
    - System should warn when trying to assign staff to non-operational branches
    - Operational status affects staff allocation calculations
    - Operational status can be set manually or automatically determined based on doctor assignments
    - On doctor-off days, only staff with "can manage a branch alone" skill can be assigned
    - System should validate that assigned staff has required skills before allowing assignment
  - **Data Storage:**
    - Operational status stored per branch per date
    - Status can be set manually or auto-calculated from doctor assignments
    - Historical operational status is maintained
    - Doctor assignments stored per branch per date (see FR-BM-05)

#### FR-BM-05: Doctor Data Structure
- **Description:** System shall manage doctor information, preferences, schedule, and daily expected revenue
- **Status:** ⚠️ Partially Implemented (doctor assignment structure implemented, doctor profile and schedule pending)
- **Details:**
  - **Doctor Information:**
    - Doctor ID (unique identifier)
    - Doctor name (full name)
    - Doctor code/nickname (optional)
    - Doctor specialization/type (optional)
    - Contact information (optional)
  - **Doctor Profile:**
    - Each doctor has a profile containing preferences and configuration
    - Doctor preferences include:
      - Preferred branches (if applicable)
      - Preferred working days/patterns
      - Special requirements or constraints
    - Doctor profile can be created, edited, and viewed by Admin and Area Manager roles
  - **Doctor Schedule:**
    - Doctors have schedules showing which days they work at which branches
    - Schedule uses the same UI pattern as branch staff scheduling (monthly calendar view)
    - Schedule shows: doctor × dates matrix with branch assignments
    - Schedule can be edited by Admin and Area Manager roles
    - Schedule supports monthly view with navigation (previous/next month)
    - Schedule displays branch codes/names for each day assignment
    - Schedule can be set up to 30 days in advance (same as branch staff scheduling)
  - **Doctor-Branch Assignment:**
    - Doctors can be assigned to branches on specific dates
    - A doctor can be assigned to multiple branches on different dates
    - A branch can have multiple doctors on the same day (up to 6 doctors per branch per day)
    - Doctor assignments are date-specific (can vary day by day)
    - Doctor assignments are managed through Doctor Profile schedule interface
  - **Doctor Daily Expected Revenue:**
    - Each doctor has a daily expected revenue value per branch per date
    - Expected revenue can vary by:
      - Date (different days may have different expected revenue)
      - Branch (same doctor may have different expected revenue at different branches)
    - Expected revenue can be set in doctor profile configuration
    - Expected revenue is stored per doctor per branch per date
  - **Business Rules:**
    - Doctor assignments determine branch operational status (if no doctor, branch is non-operational)
    - Branch expected revenue can be calculated as sum of all assigned doctors' daily expected revenue
    - Branch revenue can use branch daily revenue from branch configuration OR doctor configuration (aggregate of all doctors working in that branch on that day)
    - Doctor assignments can be imported from Excel
    - Doctor assignments can be manually created/edited through Doctor Profile schedule interface
    - Doctor expected revenue can be imported from Excel
    - Doctor expected revenue can be manually set per doctor per branch per date in doctor profile
    - Maximum 6 doctors can be assigned to a branch per day
  - **Use Cases:**
    - Admin/Area Manager views and manages doctor profiles
    - Admin/Area Manager sets doctor schedules (which days doctor works at which branches)
    - Admin/Area Manager configures doctor preferences and expected revenue
    - System calculates branch expected revenue from doctor assignments
    - System determines branch operational status based on doctor assignments
    - Import doctor assignments and expected revenue from Excel

#### FR-BM-03: Standard Branch Codes
- **Description:** System shall maintain a hardcoded list of standard branch codes that must always be available
- **Status:** ✅ Implemented
- **Details:**
  - Standard branch codes are hardcoded in the system constants
  - Standard branch codes cannot be deleted
  - Standard branch codes cannot have their codes changed
  - Standard branch codes must always be available in the system
  - The following branch codes are standard and must always be available:
    - CPN, CPN-LS, CTR, PNK, CNK, BNA, CLP, SQR, BKP, CMC, CSA, EMQ, ESV, GTW, MGA, MTA, PRM, RCT, RST, TMA, MBA, SCN, CWG, CRM, CWT, PSO, RCP, CRA, CTW, ONE, DCP, MNG, TLR, TLR-LS, TLR-WN
  - Standard branch codes are defined in `backend/internal/constants/branches.go`
  - Validation is enforced at the API level to prevent deletion or code modification of standard branches

#### FR-BM-02: Revenue Tracking
- **Description:** System shall track expected and actual revenue on a daily basis
- **Status:** ⚠️ Partially Implemented (needs enhancement for 365-day support and multiple input sources)
- **Details:**
  - Expected revenue can vary day by day for all 365 days of the year
  - Expected revenue is stored per branch per date
  - Actual revenue can be recorded per day
  - Revenue history is maintained
  - **Input Sources for Expected Revenue:**
    1. **Excel Import:** Expected revenue can be imported from Excel files
       - Excel format should support bulk import of daily revenue data
       - Import can cover multiple branches and date ranges
       - Validation required: date format, branch codes, revenue values
    2. **Branch Configuration:** Expected revenue can be set directly in branch configuration
       - Branch daily revenue can be configured per branch per date
       - Manual entry through branch configuration UI
    3. **Doctor Configuration:** Expected revenue can be calculated from doctor daily expected revenue
       - System aggregates daily expected revenue from all doctors assigned to the branch on that day
       - Doctor data structure required (see FR-BM-05)
       - Calculation: Sum of all doctors' daily expected revenue for the branch on that date
       - Branch revenue for a particular day = aggregate of revenues of all doctors working in that branch on that day
  - **Revenue Source Selection:**
    - System allows selection between branch configuration revenue and doctor configuration revenue
    - When using doctor configuration: Branch revenue = sum of all assigned doctors' expected revenue for that branch on that date
    - When using branch configuration: Branch revenue = manually set value in branch configuration
    - System should indicate which source is being used for each branch/date
  - **Business Rules:**
    - If no expected revenue is set for a day, default to 0 or use previous day's value (configurable)
    - Expected revenue from Excel import can override both branch and doctor-calculated revenue
    - Doctor-calculated revenue is automatically updated when doctor assignments change
    - Branch configuration revenue takes precedence over doctor-calculated revenue when both are set (configurable priority)
    - Revenue data must support date ranges for bulk operations

### 3.4 Doctor Profile Management

#### FR-DP-01: Doctor Profile Menu
- **Description:** System shall provide a "Doctor Profile" menu for managing doctor information, preferences, and schedules
- **Status:** ❌ Not Implemented
- **Details:**
  - **Menu Access:**
    - Doctor Profile menu accessible from main navigation
    - Available to Admin and Area Manager roles
    - Branch Managers and District Managers have read-only access (if applicable)
  - **Doctor Profile Features:**
    - View list of all doctors
    - Create new doctor profiles
    - Edit existing doctor profiles
    - Delete doctor profiles (with validation)
    - View doctor details: name, code, specialization, contact information
    - Manage doctor preferences and special requirements
    - Configure doctor expected revenue per branch per date
  - **Doctor Schedule Management:**
    - Schedule interface uses same UI pattern as branch staff scheduling
    - Monthly calendar view showing doctor × dates matrix
    - Display branch assignments for each day
    - Toggle doctor assignments to branches for specific dates
    - Navigate between months (previous/next month buttons)
    - View up to 30 days in advance
    - Visual indicators for assigned branches
    - Branch codes displayed in schedule cells
  - **Use Cases:**
    - Admin/Area Manager creates doctor profile with basic information
    - Admin/Area Manager sets doctor schedule (which days doctor works at which branches)
    - Admin/Area Manager configures doctor preferences (e.g., required staff when doctor works at specific branch)
    - Admin/Area Manager sets doctor expected revenue per branch per date
    - System uses doctor schedule to calculate branch revenue and operational status

#### FR-DP-02: Doctor Preferences and Rules
- **Description:** System shall support doctor-specific preferences and rules for staff allocation
- **Status:** ❌ Not Implemented
- **Details:**
  - **Doctor Preferences:**
    - Preferred branches (if applicable)
    - Preferred working days/patterns
    - Special requirements or constraints
    - Contact information and notes
  - **Doctor-Specific Rules:**
    - Rules specify staff requirements when a doctor works at a specific branch
    - Rules can specify:
      - Required positions (e.g., Doctor Assistant, Nurse)
      - Minimum staff count per position
      - Branch-specific requirements
      - Day-of-week or date-specific requirements
    - Example: "Doctor A must have 3 doctor assistants on the day she works at branch X"
    - Rules are stored in doctor profile configuration
    - Rules are evaluated during allocation suggestion generation
    - Rules override general branch quota configurations when applicable
  - **Rule Configuration:**
    - Admin and Area Manager can create/edit/delete doctor-specific rules
    - Rules are configured per doctor per branch (or per doctor globally)
    - Rules can be date-specific or recurring (e.g., every Monday)
    - Rules can specify multiple position requirements
  - **Business Rules:**
    - Doctor-specific rules take precedence over general branch quota when both apply
    - System validates rule requirements before allowing doctor assignment
    - Rules are enforced during allocation suggestion generation
    - Violations of doctor-specific rules trigger warnings

#### FR-DP-03: Doctor Schedule UI
- **Description:** System shall provide doctor schedule interface using the same UI pattern as branch staff scheduling
- **Status:** ❌ Not Implemented
- **Details:**
  - **UI Pattern:**
    - Uses same monthly calendar view as branch staff scheduling (see FR-SM-01)
    - Table layout: doctors × dates matrix
    - Dates displayed from left (closer dates) to right (further dates)
    - Monthly view with month navigation
  - **Schedule Display:**
    - Each row represents one doctor
    - Columns show: Doctor Name, Doctor Code, Specialization (if applicable)
    - Date columns show branch assignments
    - Visual indicators for assigned branches (branch codes displayed in cells)
    - Color coding or indicators for different branch assignments
  - **Schedule Editing:**
    - Click on date cell to assign/unassign doctor to branch
    - Branch selection dialog when assigning doctor
    - Can assign doctor to multiple branches on different dates
    - Can assign multiple doctors to same branch on same date (up to 6 doctors per branch per day)
    - Validation: Maximum 6 doctors per branch per day
    - Save and edit schedules
  - **Access Control:**
    - Admin and Area Manager can edit doctor schedules
    - Other roles have read-only access (if applicable)
  - **Integration:**
    - Doctor schedule directly relates to branch revenue calculation
    - Doctor schedule affects branch operational status
    - Doctor schedule used in allocation criteria evaluation

#### FR-DP-04: Doctor Default Scheduling and Overrides
- **Description:** System shall support default weekly schedules and date-specific overrides for doctors
- **Status:** ✅ Implemented
- **Details:**
  - **Default Scheduling:**
    - Admin and Area Manager can set default branch for each day of the week (Sunday-Saturday)
    - Each doctor can have a default branch assignment for each day of the week
    - Default schedule represents the typical weekly rotation pattern (e.g., doctor works at 3-5 branches per week)
    - Default schedule is used as the base schedule unless overridden
  - **Default Weekly Off Days:**
    - Admin and Area Manager can set default weekly off day(s) for each doctor
    - Weekly off days are recurring (e.g., every Sunday is off)
    - Multiple off days can be set per doctor
    - Weekly off days take precedence over default branch assignments
  - **Schedule Overrides:**
    - Admin and Area Manager can create date-specific overrides for any doctor
    - Override types:
      1. **Working Day Override:** Doctor works at a specific branch on a specific date (overrides default schedule)
      2. **Off Day Override:** Doctor is off on a specific date (overrides default schedule and default branch assignment)
    - Overrides require:
      - Type of day (working or off)
      - Actual date for override
      - Branch name (required if type is "working", null if type is "off")
    - Overrides take precedence over default schedules
    - Multiple overrides can be set for different dates
  - **Business Rules:**
    - Default schedules provide the base weekly pattern
    - Weekly off days override default branch assignments
    - Date-specific overrides take highest precedence
    - Priority order: Override > Weekly Off Day > Default Branch Assignment
    - When no override exists, system uses default schedule
    - When override is removed, system falls back to default schedule
  - **UI Features:**
    - Default schedule manager: Set branch for each day of week (Sunday-Saturday)
    - Weekly off days manager: Toggle off days for each day of week
    - Override manager: Monthly calendar view to set/remove overrides for specific dates
    - Visual indicators for default schedules, weekly off days, and overrides
    - Easy navigation between months for override management
  - **Access Control:**
    - Admin and Area Manager can manage default schedules, weekly off days, and overrides
    - Other roles have read-only access (if applicable)
  - **Use Cases:**
    - Admin/Area Manager sets up doctor's typical weekly rotation (e.g., Monday=Branch A, Tuesday=Branch B, Wednesday=Off, etc.)
    - Admin/Area Manager sets doctor's regular off day (e.g., every Sunday)
    - Admin/Area Manager overrides schedule for special dates (e.g., doctor works at Branch C on a specific Monday instead of default Branch A)
    - Admin/Area Manager marks doctor as off on a specific date (e.g., personal leave on a normally working day)
    - System uses default schedule when no override exists for a date
    - System applies overrides when they exist for specific dates

### 3.5 Staff Scheduling

#### FR-SM-03: Rotation Staff Schedule Management
- **Description:** System shall provide a rotation staff schedule edit menu
- **Status:** ❌ Not Implemented
- **Details:**
  - Rotation staff schedule edit menu allows editing individual rotation staff schedules
  - Schedule status options:
    1. **Day Off:** Rotation staff is not working (regular day off)
    2. **Leave Day:** Rotation staff is on leave (vacation, personal leave)
    3. **Sick Leave:** Rotation staff is on sick leave
    4. **Working Day:** Rotation staff is working - **branch must be specified**
  - Business Rules:
    - If schedule status is "working", a branch assignment is required
    - If schedule status is "day off", "leave", or "sick leave", no branch assignment is needed
    - Working day assignments must respect effective branch relationships (Level 1 and Level 2)
    - Working day assignments create or update rotation assignment records
  - Access Control:
    - Only Admin and Area Manager roles can edit rotation staff schedules
    - Branch Managers can only view rotation staff assigned to their branch (read-only)

#### FR-SM-01: Branch Staff Scheduling
- **Description:** Branch managers can schedule their branch staff
- **Status:** ⚠️ Partially Implemented
- **Details:**
  - Monthly calendar view (staff × dates matrix)
  - Can schedule up to 30 days in advance
  - Mark working days, off days, and leave days (❌ Leave day not implemented - currently only working/off day boolean)
  - Usually branch staff works 6 days a week (configurable)
  - Schedule arrangement is done in a dedicated scheduling module (separate from staff management)
  - Table view displays dates from left (closer dates) to right (further dates)
  - Branch manager can toggle each staff's workday/offday/leave day for any particular day
  - Branch manager can save and edit schedules
  - Staff are listed with Nickname, Full Name, and Position (❌ Nickname not implemented - currently only Name field)
  - Branch manager can see staff schedule of her own branch only
  - Branch manager can see rotation staff assigned to her branch too (read-only view)
  - In the staff schedule menu, show the branch name and branch code of the branch
- **Use Case:**
  1. Branch manager (a user in this system) arranges her own branch staff using this system
  2. Branch manager arranges staff for each day by toggling each staff workday/offday/leave day on any particular day to form work schedule for the whole branch
  3. Branch staff including the manager are listed by Nickname, Full Name, and position
  4. Use a table to assist staff arrangements - table view is from left (closer dates) to right (further dates)
  5. Branch manager can save, edit, or add additional staff
  6. Schedule arrangement is done in one module, while staff addition/deletion is done in another module
- **Technical Notes:**
  - Current implementation uses `is_working_day` boolean field
  - Need to change to enum/string type: `schedule_status` with values: "working", "off", "leave"
  - Database migration required: `ALTER TABLE staff_schedules ADD COLUMN schedule_status VARCHAR(20) DEFAULT 'off'`
  - Staff model needs `nickname` field addition

#### FR-SM-02: Rotation Staff Assignment
- **Description:** Admin and Area Manager roles can assign rotation staff to branches
- **Status:** ✅ Implemented (includes adhoc allocation support)
- **Details:**
  - Only Admin and Area Manager roles can assign rotation staff to branches
  - Assignment workflow:
    1. Admin/Area Manager selects a branch to view/edit
    2. Rotation staff eligible for the selected branch are automatically populated in the branch table
    3. Eligible rotation staff are determined by effective branch relationships (Level 1 and Level 2)
    4. Rotation staff appear as new rows in the branch's staff table, alongside the branch's proprietary staff
    5. Admin/Area Manager selects which date(s) a rotation staff will operate for that branch
    6. Assignment is saved as a rotation assignment record
  - Summary view showing assignments (rotation staff → branches × dates)
  - View toggles:
    - By branch group
    - By rotation staff group
    - By coverage area
  - Level 1 (priority) and Level 2 (reserved) effective branches
  - Manual override capability

#### FR-SM-04: Rotation Staff Travel Parameters
- **Description:** System shall allow configuration of travel parameters for each rotation staff per branch for shortest path calculation
- **Status:** ✅ Implemented
- **Details:**
  - Each rotation staff member can have travel parameters configured for each effective branch
  - Travel parameters include:
    1. **Commute Duration:** Travel time in minutes (default: 300 minutes / 5 hours)
    2. **Transit Count:** Number of transits required (default: 10)
    3. **Travel Cost:** Cost of traveling in currency units (default: 1000)
  - Parameters are stored per rotation staff per branch in the effective_branches table
  - Parameters can be configured when creating or updating effective branch assignments
  - Parameters are used to calculate the shortest path for rotation staff allocation
  - Default values are applied if parameters are not specified
  - Admin and Area Manager roles can configure travel parameters
  - Parameters are displayed and editable in the staff management UI

### 3.6 Business Logic

#### FR-BL-06: Allocation Criteria System
- **Description:** System shall support configurable allocation criteria across three pillars
- **Status:** ✅ Implemented (enhanced with doctor-specific pillar rules)
- **Details:**
  - **Three Pillars of Criteria:**
    1. **Clinic-Wide:** Criteria that apply across all branches (e.g., overall bookings, clinic-wide revenue targets)
    2. **Doctor-Specific:** Criteria related to doctor assignments (e.g., doctor count, doctor expected revenue, doctor-specific rules)
    3. **Branch-Specific:** Criteria specific to individual branches (e.g., branch revenue, branch minimum staff requirements)
  - **Doctor-Specific Pillar Rules:**
    - System supports doctor-specific rules that apply when certain doctors are assigned to branches
    - Rules are configured in Doctor Profile preferences
    - Example rules:
      - Doctor A must have 3 doctor assistants on the day she works at branch X
      - Doctor B requires 2 nurses when working at branch Y
      - Doctor C needs special assistant when working at branch Z
    - Rules are evaluated during allocation suggestion generation
    - Rules can specify:
      - Required staff positions and minimum counts
      - Specific branch requirements
      - Day-of-week or date-specific requirements
    - Doctor-specific rules override general branch quota configurations when applicable
  - **Criteria Types Supported:**
    - Bookings count (placeholder for future booking system integration)
    - Revenue (expected revenue per branch per day)
    - Minimum staff per position
    - Minimum staff per branch
    - Doctor count
    - Doctor-specific staff requirements (new)
  - **Criteria Configuration:**
    - Admin can configure criteria weights (0.0 - 1.0) for each criterion
    - Criteria can be activated/deactivated
    - Criteria-specific configuration via JSON config field
    - Doctor-specific rules configured in Doctor Profile
    - Criteria are evaluated and scored to generate allocation suggestions
  - **Allocation Scoring:**
    - System evaluates criteria across all three pillars
    - Generates weighted scores for each pillar
    - Doctor-specific pillar considers doctor assignments and doctor-specific rules
    - Combines pillar scores to generate overall allocation score
    - Scores are used to prioritize allocation suggestions

#### FR-BL-07: Position Quota Management
- **Description:** System shall manage designated quotas and minimum requirements per position per branch (only for branch-type positions)
- **Status:** ✅ Implemented
- **Details:**
  - **Position Type Restriction:**
    - Only branch-type positions can have quota configurations
    - Rotation-type positions cannot have branch-specific quota settings
    - System validates position type before allowing quota updates
    - Rotation positions are excluded from branch configuration UI
  - **Quota Configuration:**
    - Designated quota: Preferred/target number of staff for each branch-type position at each branch
    - Minimum required: Minimum number of staff required for each branch-type position at each branch
    - Quotas are branch-specific and position-specific (branch-type positions only)
    - Quotas can be activated/deactivated
    - UI labels: "Preferred" (for designated quota) and "Minimum" (for minimum required)
  - **Quota Calculation:**
    - System calculates available local staff for each position
    - Tracks assigned rotation staff per position
    - Calculates total assigned staff (local + rotation)
    - Determines still required staff count (minimum required - total assigned)
  - **Quota Status:**
    - Real-time quota fulfillment status per branch per day
    - Position-level quota status breakdown (branch-type positions only)
    - Visual indicators for positions with shortages
  - **Access Control:**
    - Admin and Area Manager can create/edit/delete quotas (branch-type positions only)
    - All roles can view quota status

#### FR-BL-08: Automated Allocation Suggestions
- **Description:** System shall generate automated allocation suggestions based on criteria and quota data
- **Status:** ✅ Implemented
- **Details:**
  - **Suggestion Generation:**
    - System analyzes quota status and allocation criteria to generate suggestions
    - Suggests rotation staff assignments to fill missing positions
    - Considers effective branch relationships (Level 1 and Level 2)
    - Checks staff availability before suggesting
    - Calculates confidence scores for each suggestion
    - Provides reasoning for each suggestion
  - **Suggestion Workflow:**
    - Suggestions are generated with "pending" status
    - Area Managers can review, approve, reject, or modify suggestions
    - Approved suggestions automatically create rotation assignments
    - Rejected suggestions are tracked for analysis
    - Modified suggestions allow manual adjustments before approval
  - **Suggestion Criteria:**
    - Based on 3-pillar criteria evaluation
    - Considers quota fulfillment status
    - Prioritizes positions with shortages
    - Respects effective branch relationships
    - Checks staff availability and skills

#### FR-BL-09: Adhoc Allocation Support
- **Description:** System shall support adhoc allocations for unplanned leave scenarios
- **Status:** ✅ Implemented
- **Details:**
  - Rotation assignments can be marked as adhoc allocations
  - Adhoc allocations require a reason (e.g., unplanned sick leave, emergency coverage)
  - Adhoc allocations are tracked separately from regular allocations
  - Area Managers can create adhoc allocations through dedicated interface
  - Adhoc allocations follow same validation rules as regular allocations
  - System maintains audit trail of adhoc allocations with reasons

#### FR-BL-10: Allocation Override Support
- **Description:** System shall support manual overrides for automatic allocation suggestions
- **Status:** ❌ Not Implemented
- **Details:**
  - Overrides allow manual adjustment of automatic allocation suggestions one by one
  - Overrides are used when there are latent logic/rules not covered in the allocation logic
  - Override functionality available for each individual allocation suggestion
  - When overriding, user can:
    - Change the assigned rotation staff
    - Change the assigned branch
    - Change the assigned date
    - Change the assigned position
    - Add override reason/notes explaining why override was necessary
  - Overridden allocations are marked with override status
  - Override history is maintained for audit and learning purposes
  - Overrides can be applied before or after suggestion approval
  - System tracks which user performed the override and when
  - Override reasons help identify gaps in allocation logic for future improvements

#### FR-BL-01: Revenue-Based Staff Calculation
- **Description:** System shall calculate required staff based on expected revenue
- **Status:** ✅ Implemented (includes preferred staff per position via quota system)
- **Details:**
  - Formula: `staff_count = min_staff + (revenue_threshold * multiplier)`
  - Configurable via System Settings
  - Minimum staff per position regardless of revenue
  - Doctor count consideration
  - **Preferred Staff per Position:** Each branch has a preferred number of staff for each position (✅ Implemented via Position Quota system)
    - Preferred staff counts act as constraints when allocating rotation staff
    - Preferred counts are branch-specific and position-specific
    - System should prioritize meeting preferred staff counts when allocating rotation staff
    - Preferred counts can be configured per branch per position

#### FR-BL-02: Effective Branch Management
- **Description:** Rotation staff can only be assigned to designated branches
- **Status:** ✅ Implemented
- **Details:**
  - Level 1: Priority/preferred branches
  - Level 2: Reserved branches (used when Level 1 staff are insufficient)
  - Based on travel distance/coverage area

#### FR-BL-03: Availability Checking
- **Description:** System shall check rotation staff availability before assignment
- **Status:** ⚠️ Partially Implemented
- **Details:**
  - Check if rotation staff is already assigned on the date (❌ Not Implemented - code commented out)
  - Verify effective branch assignment (✅ Implemented)
  - Check coverage area constraints (⚠️ Partially Implemented)
- **Related:** See `docs/conflict-resolution-rules.md` for conflict handling

#### FR-BL-04: Conflict Detection and Resolution
- **Description:** System shall detect and resolve scheduling conflicts
- **Status:** ❌ Not Implemented
- **Details:**
  - Detect rotation staff double-booking conflicts
  - Detect effective branch access conflicts
  - Detect staff shortfall situations
  - Provide conflict resolution options
  - Notify managers of conflicts
- **Related:** See `docs/conflict-resolution-rules.md` for detailed rules

#### FR-BL-05: Minimum Daily Staff Constraints
- **Description:** System shall enforce minimum daily staff requirements per branch
- **Status:** ✅ Implemented (via Position Quota system)
- **Details:**
  - Each branch has minimum staff requirements for specific positions on each day
  - Minimum requirements are position-based (e.g., at least 1 front/counter staff and 1 doctor assistant = minimum 2 total)
  - Minimum daily staff counts act as constraints when allocating rotation staff
  - System must ensure minimum requirements are met before allocating additional staff
  - Minimum requirements can vary by day (e.g., different requirements for weekdays vs weekends)
  - Minimum requirements are configurable per branch per position per day (✅ Implemented)
  - Violations of minimum requirements should trigger warnings or prevent allocation

### 3.7 AI Suggestions

#### FR-AI-01: MCP Integration
- **Description:** System shall provide AI suggestions via MCP
- **Status:** ✅ Implemented (criteria-based suggestion engine)
- **Details:**
  - MCP client framework implemented
  - Suggestion generation engine implemented with 3-pillar criteria system
  - Regenerate suggestions capability implemented
  - Multiple suggestions can be generated and compared
  - Suggestion approval/rejection workflow implemented
  - Criteria-based allocation scoring implemented (clinic-wide, doctor-specific, branch-specific)

### 3.8 Internationalization

#### FR-I18N-01: Multi-Language Support
- **Description:** System shall support Thai and English languages
- **Status:** ⚠️ Partially Implemented
- **Details:**
  - Language toggle in UI (structure ready)
  - Translation files for Thai and English (created)
  - Date/time formatting in Thailand timezone (✅ Implemented)

### 3.9 Dashboard and Reporting

#### FR-RP-01: Dashboard View
- **Description:** System shall provide dashboard with summary statistics
- **Status:** ⚠️ Partially Implemented
- **Details:**
  - Summary statistics (basic implementation)
  - Staff allocation overview
  - Revenue vs. staff allocation charts (❌ Not Implemented)

#### FR-RP-03: Allocation Overview Views
- **Description:** System shall provide comprehensive overview views for Area Managers
- **Status:** ✅ Implemented
- **Details:**
  - **All Branches Day-by-Day Overview:** Area Managers can view all 30+ branches' staff allocation for a specific day
    - Shows quota, available local staff, assigned rotation staff, and required staff for each branch
    - Drill-down capability to view position-level details
    - Visual indicators for branches with shortages
  - **Single Branch Monthly Overview:** Area Managers and Branch Managers can view a branch's staff allocation for an entire month
    - Day-by-day breakdown with quota fulfillment status
    - Position-level details available via drill-down
    - Average fulfillment rate calculation
  - Overview views show designated quota, available local staff, assigned rotation staff, and still required staff counts
  - Supports filtering and date navigation

#### FR-RP-02: Reporting Capabilities
- **Description:** System shall provide reporting capabilities
- **Status:** ❌ Not Implemented
- **Details:**
  - Staff allocation reports
  - Revenue vs. staff efficiency reports
  - Conflict history reports
  - Export reports to PDF/Excel formats
  - Scheduled report generation

#### FR-RP-04: Automatic Allocation Report
- **Description:** System shall generate and maintain reports for each automatic allocation iteration
- **Status:** ❌ Not Implemented
- **Details:**
  - **Report Generation:**
    - System generates a report for each automatic allocation iteration/run
    - Each report is uniquely identified and timestamped
    - Reports are stored as records for historical tracking
    - Reports can be viewed, filtered, and exported
  - **Report Content:**
    - **Assignment Details:** For each rotation staff assignment made in the iteration:
      - Rotation staff name and ID
      - Assigned branch name and code
      - Assigned date
      - Assigned position
      - Detailed reason for assignment (explaining why this staff was assigned to this branch)
      - Criteria used in the decision (which allocation criteria influenced the assignment)
      - Confidence score
      - Whether the assignment was overridden (if applicable)
      - Override reason (if overridden)
    - **Gap Analysis:** For each branch and position combination:
      - Number of roles/positions that still lack sufficient staff
      - Number of staff still needed to satisfy criteria
      - Breakdown by position showing:
        - Required staff count (from quota/criteria)
        - Available local staff count
        - Assigned rotation staff count
        - Still required staff count
      - Visual indicators for positions with shortages
    - **Summary Statistics:**
      - Total assignments made in this iteration
      - Total branches covered
      - Total positions filled
      - Total positions still requiring staff
      - Overall fulfillment rate
      - Average confidence score
  - **Report Features:**
    - Reports are searchable and filterable by:
      - Date range
      - Branch
      - Position
      - Rotation staff
      - Assignment status (approved, rejected, overridden)
    - Reports can be exported to PDF/Excel formats
    - Reports can be viewed in the UI with drill-down capabilities
    - Historical reports are maintained for analysis and learning
  - **Use Cases:**
    - Area Managers review allocation decisions and understand reasoning
    - System administrators identify gaps in allocation logic
    - Management track allocation effectiveness over time
    - Identify patterns in override usage to improve allocation logic
    - Audit trail for allocation decisions

### 3.10 Data Validation

#### FR-DV-01: Input Validation
- **Description:** System shall validate all user inputs
- **Status:** ⚠️ Partially Implemented
- **Details:**
  - Server-side validation for all inputs (✅ Implemented)
  - Field length limits (⚠️ Needs verification)
  - Data format validation (email, date, UUID, etc.) (✅ Implemented)
  - Business rule validation (⚠️ Needs enhancement)
  - File upload validation (❌ Not Implemented - Excel import)

#### FR-DV-02: Data Integrity
- **Description:** System shall maintain data integrity
- **Status:** ✅ Implemented
- **Details:**
  - Database constraints and foreign keys
  - Unique constraints (username, email, branch code)
  - Referential integrity
  - Transaction support

### 3.11 Notifications

#### FR-NT-01: Conflict Notifications
- **Description:** System shall notify managers of conflicts
- **Status:** ❌ Not Implemented
- **Details:**
  - Notify when rotation staff assignment conflicts occur
  - Notify when assignments are replaced or overridden
  - In-app notifications
  - Email notifications (optional)

#### FR-NT-02: System Notifications
- **Description:** System shall provide system-wide notifications
- **Status:** ❌ Not Implemented
- **Details:**
  - Staff shortfall warnings
  - Schedule change notifications
  - Assignment reminders
  - Notification preferences per user

## 4. Non-Functional Requirements

### 4.1 Performance

#### NFR-PF-01: Response Time
- **Description:** System shall respond to user requests within acceptable time
- **Status:** ❌ Not Implemented
- **Target:** API responses < 500ms, page loads < 2s

#### NFR-PF-02: Scalability
- **Description:** System shall support growth to additional branches
- **Status:** ✅ Implemented (architecture supports scalability)

### 4.2 Security

#### NFR-SC-01: Data Security
- **Description:** System shall protect sensitive data
- **Status:** ✅ Implemented
- **Details:**
  - Password hashing
  - Session management
  - Input validation
  - SQL injection prevention (parameterized queries)
- **Related:** See `docs/security-requirements.md` for comprehensive security requirements

**Note:** Additional security requirements are documented in `docs/security-requirements.md`, including:
- Password policy (NFR-SC-02)
- Password reset (NFR-SC-05)
- Session security (NFR-SC-06)
- API rate limiting (NFR-SC-11)
- Security logging (NFR-SC-17)
- Intrusion detection (NFR-SC-19)
- And more...

### 4.3 Usability

#### NFR-UI-01: Responsive Design
- **Description:** System shall work on laptops and mobile phones
- **Status:** ⚠️ Partially Implemented (structure ready, styling needed)

#### NFR-UI-02: User Interface
- **Description:** System shall provide intuitive user interface
- **Status:** ⚠️ Partially Implemented
- **Details:**
  - Summary view for rotation assignments (structure ready)
  - Monthly calendar view for branch scheduling (structure ready)
  - Easy-to-use navigation

### 4.4 Reliability

#### NFR-RL-01: Data Persistence
- **Description:** System shall maintain records of staff arrangements
- **Status:** ✅ Implemented
- **Details:**
  - All assignments are stored in database
  - Historical data can be reviewed
  - Audit trail for changes

## 5. Business Rules

### BR-BL-01: Branch Staff Work Schedule
- **Description:** Branch staff usually works 6 days a week
- **Status:** ✅ Implemented (configurable per schedule)

### BR-BL-02: Advance Scheduling Limit
- **Description:** Branch managers can schedule up to 30 days in advance
- **Status:** ✅ Implemented

### BR-BL-03: Minimum Staff Requirements
- **Description:** Each position has minimum staff requirements regardless of revenue
- **Status:** ✅ Implemented

### BR-BL-04: Rotation Staff Assignment Levels
- **Description:** Rotation staff assignments have two levels (priority and reserved)
- **Status:** ✅ Implemented

### BR-BL-05: Schedule Status Types
- **Description:** Staff schedules support three status types: working day, off day, and leave day
- **Status:** ✅ Implemented
- **Details:**
  - Working day: Staff is scheduled to work
  - Off day: Staff is not scheduled (regular day off)
  - Leave day: Staff is on leave (vacation, sick leave, etc.)
  - Schedule status is stored per staff per date
  - Branch manager can toggle between these three states

### BR-BL-06: Branch Manager Access Control
- **Description:** Branch managers can only add/arrange staff of their own branch
- **Status:** ✅ Implemented
- **Details:**
  - Branch managers are linked to a branch via `branch_id` in the users table
  - Branch managers can only view and manage staff from their assigned branch
  - Branch managers can only create schedules for their assigned branch
  - Staff listings are automatically filtered to show only staff from the branch manager's branch
  - Access is enforced at the API level through middleware and handler validation
  - Branch code is used to link branch managers to branches (username pattern: `{branchcode}mgr` or `{branchcode}amgr`)

### BR-BL-07: Branch Expected Revenue Calculation
- **Description:** Branch expected revenue can be calculated from multiple sources
- **Status:** ⚠️ Partially Implemented (doctor assignment structure implemented, calculation logic pending)
- **Details:**
  - Expected revenue can be set manually, imported from Excel, or calculated from doctor assignments
  - When calculated from doctors: Sum of all assigned doctors' daily expected revenue for that branch on that date (✅ Doctor assignment structure implemented, calculation logic pending)
  - Excel import can override doctor-calculated revenue
  - Manual entry can override both Excel and doctor-calculated revenue
  - Priority: Manual entry > Excel import > Doctor-calculated

### BR-BL-08: Branch Operational Status
- **Description:** Branch operational status is determined by doctor assignments and manual settings
- **Status:** ⚠️ Partially Implemented (doctor-on/off day designation implemented, full operational status tracking pending)
- **Details:**
  - Branch is operational if: Has at least one doctor assigned AND not manually marked as closed
  - Branch is non-operational if: No doctor assigned OR manually marked as closed
  - Operational status affects staff allocation requirements
  - Non-operational branches require minimal or no staff allocation
  - System should warn when assigning staff to non-operational branches
  - Doctor-on/doctor-off day designation implemented (✅ Branch Managers can set doctor-on/off days)

### BR-BL-09: Preferred Staff per Position per Branch
- **Description:** Each branch has preferred number of staff for each position
- **Status:** ✅ Implemented (via Position Quota system)
- **Details:**
  - Preferred staff counts are branch-specific and position-specific
  - Preferred counts act as constraints when allocating rotation staff
  - System should prioritize meeting preferred staff counts when allocating rotation staff
  - Preferred counts are configurable per branch per position (✅ Implemented)
  - Preferred counts are used as targets for staff allocation optimization

### BR-BL-10: Minimum Daily Staff Constraints
- **Description:** Each branch has minimum staff requirements for specific positions on each day
- **Status:** ✅ Implemented (via Position Quota system)
- **Details:**
  - Minimum requirements are position-based (e.g., at least 1 front/counter staff and 1 doctor assistant = minimum 2 total)
  - Minimum daily staff counts act as constraints when allocating rotation staff
  - System must ensure minimum requirements are met before allocating additional staff
  - Minimum requirements can vary by day (e.g., different requirements for weekdays vs weekends)
  - Minimum requirements are configurable per branch per position per day (✅ Implemented)
  - Violations of minimum requirements should trigger warnings or prevent allocation

### BR-BL-11: Doctor-Off Day Staff Requirement
- **Description:** When a branch has no doctor assigned (doctor-off day), only one staff with "can manage a branch alone" skill is required
- **Status:** ❌ Not Implemented
- **Details:**
  - Branches typically have 1-6 doctors assigned per day (doctor-on days)
  - When a branch has no doctor assigned (doctor-off day), only one staff member is required
  - The single staff member on doctor-off days must have the "can manage a branch alone" skill
  - System should validate that assigned staff has the required skill before allowing assignment
  - Doctor assignments determine whether a day is doctor-on or doctor-off
  - Maximum 6 doctors can be assigned to a branch per day

### BR-BL-12: Doctor Schedule Management
- **Description:** Doctor schedules use the same UI pattern as branch staff scheduling
- **Status:** ❌ Not Implemented
- **Details:**
  - Doctor schedule interface follows the same monthly calendar view pattern as branch staff scheduling
  - Schedule shows doctor × dates matrix with branch assignments
  - Schedule can be edited up to 30 days in advance
  - Schedule displays branch codes/names for each day assignment
  - Maximum 6 doctors can be assigned to a branch per day
  - Doctor schedule directly relates to branch revenue calculation
  - Doctor schedule affects branch operational status

### BR-BL-13: Doctor-Specific Pillar Rules
- **Description:** Doctor-specific rules override general branch quota configurations when applicable
- **Status:** ❌ Not Implemented
- **Details:**
  - Doctor-specific rules specify staff requirements when a doctor works at a specific branch
  - Rules are configured in Doctor Profile preferences
  - Rules can specify required positions and minimum counts (e.g., "Doctor A must have 3 doctor assistants on the day she works at branch X")
  - Rules are evaluated during allocation suggestion generation
  - Doctor-specific rules take precedence over general branch quota when both apply
  - System validates rule requirements before allowing doctor assignment
  - Violations of doctor-specific rules trigger warnings

### BR-BL-14: Branch Revenue Calculation Sources
- **Description:** Branch revenue can be calculated from branch configuration or doctor configuration
- **Status:** ⚠️ Partially Implemented (doctor assignment structure implemented, calculation logic pending)
- **Details:**
  - Branch revenue for a particular day can use one of two sources:
    1. **Branch Configuration:** Manually set daily revenue in branch configuration
    2. **Doctor Configuration:** Aggregate of revenues of all doctors working in that branch on that day
  - When using doctor configuration: Branch revenue = sum of all assigned doctors' expected revenue for that branch on that date
  - System should allow selection between branch configuration revenue and doctor configuration revenue
  - Branch configuration revenue takes precedence over doctor-calculated revenue when both are set (configurable priority)
  - System should indicate which source is being used for each branch/date
  - Doctor-calculated revenue is automatically updated when doctor assignments change

### BR-BL-15: Branch Working Day Determination and Staff Conversion
- **Description:** Branch working days are determined by doctor schedule, and branch staff temporarily become rotation staff on branch off days
- **Status:** ❌ Not Implemented
- **Details:**
  - **Working Day Definition:**
    - A branch's working day must have at least one doctor working on that day
    - Doctor schedule is used to verify if a branch has at least one doctor assigned for a specific date
    - System checks doctor assignments for each branch on each date to determine working status
  - **Branch Off Day Definition:**
    - If a branch has no doctor working on a specific day, that day is considered a branch off day
    - Branch off days are automatically determined based on doctor schedule (no manual override needed)
    - System should clearly indicate branch off days in UI and reports
  - **Staff Conversion Rule:**
    - For a branch's off day, any branch staff from that branch temporarily becomes rotation staff only for that day
    - Branch staff converted to rotation staff become available for rotation staff selection on that specific day
    - All rules for rotation staff apply to converted branch staff (e.g., rotation assignment levels, effective branch relationships, schedule constraints)
    - Conversion is automatic and temporary - branch staff revert to branch staff status on working days
    - System should track which branch staff are available as rotation staff due to branch off days
  - **Implementation Requirements:**
    - System must verify doctor schedule before determining branch working status
    - System must identify branch off days based on doctor schedule
    - System must make branch staff available as rotation staff on their branch's off days
    - Rotation staff selection logic must include converted branch staff from off-day branches
    - UI should indicate when branch staff are available as rotation staff due to branch off days
  - **Related Rules:**
    - This rule interacts with BR-BL-04 (Rotation Staff Assignment Levels)
    - This rule interacts with BR-BL-08 (Branch Operational Status)
    - This rule interacts with BR-BL-11 (Doctor-Off Day Staff Requirement)
    - This rule interacts with BR-BL-12 (Doctor Schedule Management)

### BR-BL-16: Doctor Temporary Branch Assignment
- **Description:** Doctors can temporarily work at branches that are not in their default schedule
- **Status:** ⚠️ Partially Implemented (override functionality exists, but explicit non-default branch assignment rule not documented)
- **Details:**
  - **Temporary Assignment:**
    - Doctors can be assigned to work at any branch on any date, regardless of whether that branch is in their default schedule
    - Temporary assignments are made through schedule overrides (see FR-DP-04)
    - A doctor's default schedule defines their typical weekly rotation pattern (usually 3-5 branches per week)
    - Temporary assignments allow flexibility for special circumstances (e.g., covering for another doctor, special events, branch needs)
  - **Override Mechanism:**
    - Temporary branch assignments are created using working day overrides
    - Overrides specify: date, branch assignment, and expected revenue for that assignment
    - Overrides take precedence over default schedule (priority: Override > Weekly Off Day > Default Branch Assignment)
    - Multiple temporary assignments can be created for different dates
    - Temporary assignments do not affect the default schedule
  - **Validation:**
    - System should allow assignment to any valid branch (not restricted to default branches only)
    - System should validate that the branch exists and is operational
    - System should validate that maximum doctor limit per branch per day is not exceeded (6 doctors maximum)
    - System should require expected revenue to be specified for temporary assignments
  - **Use Cases:**
    - Doctor temporarily covers at a branch where another doctor is unavailable
    - Doctor is assigned to a branch for a special event or promotion
    - Doctor fills in at a branch that needs additional support
    - Doctor is reassigned due to operational needs
  - **Related Rules:**
    - This rule relates to FR-DP-04 (Doctor Default Scheduling and Overrides)
    - This rule relates to BR-BL-12 (Doctor Schedule Management)
    - This rule relates to BR-BL-17 (Default Schedule Expected Revenue)

### BR-BL-17: Default Schedule Expected Revenue Requirement
- **Description:** Default schedules must have expected revenue associated with each day
- **Status:** ❌ Not Implemented
- **Details:**
  - **Expected Revenue Requirement:**
    - Each day in a doctor's default schedule must have an associated expected revenue value
    - Expected revenue is specified per doctor per branch per day of week
    - Expected revenue represents the doctor's typical daily revenue contribution at that branch
    - Expected revenue values are used to calculate branch revenue when doctor is assigned via default schedule
  - **Default Schedule Structure:**
    - Default schedule consists of: day of week (Sunday-Saturday), branch assignment, and expected revenue
    - All three components (day, branch, revenue) must be specified for each working day in default schedule
    - Days marked as weekly off days do not require expected revenue (doctor is not working)
    - Expected revenue can vary by day of week and by branch
  - **Revenue Configuration:**
    - Admin and Area Manager configure expected revenue when setting up default schedules
    - Expected revenue can be set individually for each day-branch combination
    - Expected revenue can be imported from Excel or manually entered
    - Expected revenue values should reflect realistic daily contribution expectations
  - **Override Behavior:**
    - When a temporary override is created, it can specify a different expected revenue for that specific date
    - Override expected revenue takes precedence over default schedule expected revenue
    - If override does not specify expected revenue, system should use default schedule expected revenue for that branch
  - **Calculation Impact:**
    - Branch revenue calculation uses expected revenue from default schedule when no override exists
    - Branch revenue calculation uses expected revenue from override when override exists
    - System aggregates expected revenue from all doctors assigned to a branch on a given day
    - Expected revenue from default schedules contributes to branch revenue calculation
  - **Validation:**
    - System must require expected revenue to be specified when creating or updating default schedule entries
    - System should validate that expected revenue is a positive number
    - System should warn if expected revenue seems unusually high or low compared to historical data
  - **Related Rules:**
    - This rule relates to FR-DP-04 (Doctor Default Scheduling and Overrides)
    - This rule relates to BR-BL-14 (Branch Revenue Calculation Sources)
    - This rule relates to BR-BL-16 (Doctor Temporary Branch Assignment)

### BR-BL-18: Multiple Doctors per Branch Simultaneously
- **Description:** A branch can have more than one doctor working simultaneously on the same day
- **Status:** ✅ Implemented (structure supports multiple doctors, maximum limit enforced)
- **Details:**
  - **Multiple Doctor Assignment:**
    - A branch can have multiple doctors assigned to work on the same day
    - Multiple doctors can work simultaneously at the same branch
    - Each doctor maintains their own schedule and expected revenue
    - All assigned doctors contribute to branch revenue calculation
  - **Maximum Limit:**
    - Maximum 6 doctors can be assigned to a branch per day
    - System enforces this limit when assigning doctors
    - System should prevent assignment if limit would be exceeded
    - System should display current doctor count and remaining capacity
  - **Assignment Sources:**
    - Doctors can be assigned via default schedules
    - Doctors can be assigned via temporary overrides
    - Multiple doctors can be assigned from different sources (mix of default and override)
    - System aggregates all assignments regardless of source
  - **Revenue Calculation:**
    - Branch revenue for a day with multiple doctors = sum of all assigned doctors' expected revenue
    - Each doctor's expected revenue is added to the total
    - Expected revenue can vary per doctor per branch
    - System calculates total branch revenue from all assigned doctors
  - **Operational Status:**
    - Branch is operational if at least one doctor is assigned (see BR-BL-15)
    - Multiple doctors increase branch operational capacity
    - Multiple doctors may require additional staff allocation (see doctor-specific pillar rules)
  - **Use Cases:**
    - High-revenue days require multiple doctors to handle patient load
    - Special events or promotions require additional doctor coverage
    - Branch has multiple doctors as part of regular operations
    - Temporary coverage when multiple doctors are needed
  - **Validation:**
    - System must enforce maximum 6 doctors per branch per day limit
    - System should validate that all assigned doctors have expected revenue specified
    - System should check for scheduling conflicts (e.g., doctor assigned to multiple branches on same day)
    - System should warn if doctor count seems unusually high or low
  - **Related Rules:**
    - This rule relates to BR-BL-12 (Doctor Schedule Management)
    - This rule relates to BR-BL-14 (Branch Revenue Calculation Sources)
    - This rule relates to BR-BL-15 (Branch Working Day Determination)
    - This rule relates to BR-BL-13 (Doctor-Specific Pillar Rules)

### BR-BM-01: Position Type Classification
- **Description:** Positions must be classified into two types: "branch" and "rotation"
- **Status:** ✅ Implemented
- **Details:**
  - **Branch Positions:**
    - Positions that are assigned to specific branches
    - Can have branch-specific quota configurations (minimum/preferred staff counts)
    - Appear in branch configuration quota settings UI
    - Examples: Branch Manager, Assistant Branch Manager, Doctor Assistant, Nurse, Front positions, Coordinator
  - **Rotation Positions:**
    - Positions used for rotation staff assignments
    - Cannot have branch-specific quota configurations
    - Excluded from branch configuration quota settings UI
    - Managed through rotation staff assignments, not fixed quotas
    - Examples: District Manager, Department Manager, Head Doctor Assistant, Special Assistant, Front Rotation, Doctor Assistant Rotation
  - **Classification Rules:**
    - Each position must have a position_type field with value "branch" or "rotation"
    - Position type is set when creating positions and can be updated by admins
    - System validates position type before allowing quota updates
    - API rejects quota update requests for rotation-type positions
  - **UI Behavior:**
    - Branch configuration UI only displays branch-type positions
    - Rotation-type positions are automatically filtered out
    - Admin position management UI displays position type with visual indicators
    - Position type can be edited in admin interface

**Note:** Additional business rules, including conflict resolution rules, are documented in:
- `docs/business-rules.md` - General business rules
- `docs/conflict-resolution-rules.md` - Conflict resolution specific rules

## 6. System Constraints

- **Timezone:** All dates and times must use Thailand Time (Asia/Bangkok)
- **Languages:** Thai and English
- **Browser Support:** Modern browsers (Chrome, Firefox, Safari, Edge)
- **Database:** PostgreSQL 15+

## 7. Future Enhancements

- Advanced reporting and analytics
- Mobile app (iOS/Android)
- Real-time notifications
- Integration with HR systems
- Advanced AI/ML optimization algorithms

---

## Change Log

| Date | Version | Changes | Author |
|------|---------|---------|--------|
| 2026-01-17 | 1.15.0 | Added BR-BL-15: Branch Working Day Determination and Staff Conversion - Branch working days are determined by doctor schedule (at least one doctor required). Branch off days occur when no doctor is scheduled. On branch off days, branch staff temporarily become rotation staff for that day and become available for rotation staff selection, with all rotation staff rules applying. | System |
| 2026-01-17 | 1.16.0 | Added three new business rules: BR-BL-16: Doctor Temporary Branch Assignment - Doctors can temporarily work at branches not in their default schedule via overrides. BR-BL-17: Default Schedule Expected Revenue Requirement - Default schedules must have expected revenue associated with each day. BR-BL-18: Multiple Doctors per Branch Simultaneously - A branch can have multiple doctors working simultaneously (up to 6 doctors per branch per day). | System |
| 2024-12-19 | 1.0.0 | Initial requirements document created | System |
| 2025-12-18 | 1.1.0 | Added conflict resolution requirements, data validation, notifications, reporting. Added references to security and conflict resolution documents. Updated status of FR-BL-03. | System |
| 2025-12-21 | 1.2.0 | Added FR-BM-03: Standard Branch Codes requirement. Documented hardcoded list of 35 standard branch codes that must always be available in the system. | System |
| 2025-12-21 | 1.3.0 | Enhanced FR-SM-01: Added leave day support requirement, table layout specification (left to right dates), and staff display requirements (nickname, full name, position). Enhanced FR-RM-03: Added nickname field requirement and clarified module separation. Updated FR-AUZ-01: Clarified Branch Manager capabilities. | System |
| 2025-12-21 | 1.4.0 | Added BR-BL-06: Branch Manager Access Control. Branch managers can only add/arrange staff of their own branch, enforced by branch code. Updated FR-AUZ-01 to include branch manager restrictions. Implemented branch_id field in users table and middleware to enforce branch access. | System |
| 2025-12-22 | 1.5.0 | Enhanced FR-RM-03: Added branch manager restrictions - manual entry only for branch staff of that branch, cannot add rotation staff. Enhanced FR-SM-01: Added requirement that branch manager can see rotation staff assigned to her branch (read-only), and requirement to display branch name and branch code in staff schedule menu. | System |
| 2025-12-23 | 1.6.0 | Enhanced FR-RM-03: Clarified rotation staff management permissions - only Admin and Area Manager can view/add/edit/delete/import rotation staff. Enhanced FR-SM-02: Updated assignment workflow - rotation staff appear in branch table, admin/area manager selects dates. Added FR-SM-03: Rotation Staff Schedule Management - schedule edit menu for day off/leave/sick leave/working days. Updated FR-AUZ-01: Clarified role permissions for rotation staff management. | System |
| 2025-12-23 | 1.7.0 | Enhanced FR-BM-02: Updated revenue tracking to support 365 days per year with multiple input sources (Excel import and doctor-calculated). Added FR-BM-04: Branch Operational Status - tracks daily operational status (operational/no doctor/closed). Added FR-BM-05: Doctor Data Structure - doctor information, branch assignments, and daily expected revenue. Added BR-BL-07: Branch Expected Revenue Calculation rules. Added BR-BL-08: Branch Operational Status rules. Updated FR-BM-01: Added reference to operational status. | System |
| 2026-01-09 | 1.8.0 | Enhanced FR-BL-01: Added preferred staff per position per branch requirement. Added FR-BL-05: Minimum Daily Staff Constraints - position-based minimum requirements per branch per day. Enhanced FR-BM-04: Added doctor-on/doctor-off days logic with "can manage a branch alone" skill requirement. Added FR-RM-04: Staff Skills and Qualifications - tracks staff skills including "can manage a branch alone". Added BR-BL-09: Preferred Staff per Position per Branch rules. Added BR-BL-10: Minimum Daily Staff Constraints rules. Added BR-BL-11: Doctor-Off Day Staff Requirement rules. | System |
| 2026-01-09 | 1.9.0 | Implemented enhanced role-based allocation system: Added FR-BL-06: Allocation Criteria System (3 pillars - clinic-wide, doctor-specific, branch-specific) with configurable weights. Added FR-BL-07: Position Quota Management system for designated quotas and minimum requirements. Added FR-BL-08: Automated Allocation Suggestions with approval workflow. Added FR-BL-09: Adhoc Allocation Support for unplanned leave scenarios. Added FR-RP-03: Allocation Overview Views (all branches day-by-day, single branch monthly). Updated FR-BM-05: Doctor Data Structure - implemented doctor assignment system with expected revenue tracking. Updated FR-BM-04: Branch Operational Status - implemented doctor-on/off day designation. Updated FR-SM-02: Rotation Staff Assignment - added adhoc allocation support. Updated FR-AI-01: MCP Integration - implemented criteria-based suggestion engine. Updated FR-AUZ-01: Enhanced role permissions for new allocation features. Updated BR-BL-09 and BR-BL-10: Marked as implemented via Position Quota system. Updated BR-BL-07 and BR-BL-08: Marked as partially implemented. | System |
| 2026-01-13 | 1.10.0 | Added position type classification: Updated FR-RM-02: Staff Positions - added position type classification (branch/rotation). Updated FR-BL-07: Position Quota Management - restricted quota configuration to branch-type positions only. Added BR-BM-01: Position Type Classification - business rule for position types. Rotation-type positions (District Manager, Department Manager, Head Doctor Assistant, Special Assistant, and rotation positions) are excluded from branch quota configuration. Branch configuration UI only displays branch-type positions. Updated UI labels: "Required Staff" → "Preferred", "Minimum Required" → "Minimum". | System |
| 2026-01-17 | 1.11.0 | Added Doctor Profile feature: Added new section 3.4 Doctor Profile Management with FR-DP-01 (Doctor Profile Menu), FR-DP-02 (Doctor Preferences and Rules), and FR-DP-03 (Doctor Schedule UI). Enhanced FR-BM-05: Doctor Data Structure - added doctor profile, preferences, and schedule management. Enhanced FR-BM-02: Revenue Tracking - clarified branch revenue can use branch configuration OR doctor configuration (aggregate of all doctors working in that branch on that day). Enhanced FR-BL-06: Allocation Criteria System - added doctor-specific pillar rules that override general branch quota configurations. Updated BR-BL-11: Doctor-Off Day Staff Requirement - updated maximum doctors per branch from 4 to 6. Added BR-BL-12: Doctor Schedule Management - doctor schedules use same UI pattern as branch staff scheduling. Added BR-BL-13: Doctor-Specific Pillar Rules - rules configured in Doctor Profile preferences, example: "Doctor A must have 3 doctor assistants on the day she works at branch X". Added BR-BL-14: Branch Revenue Calculation Sources - branch revenue can be calculated from branch configuration or doctor configuration. Maximum 6 doctors can be assigned to a branch per day. | System |
| 2026-01-17 | 1.12.0 | Added Allocation Report and Override features: Added FR-RP-04: Automatic Allocation Report - system generates detailed reports for each allocation iteration tracking assignment reasons, criteria used, and gap analysis. Added FR-BL-10: Allocation Override Support - manual override functionality for automatic allocation suggestions (one by one) to handle latent logic not covered in allocation rules. Reports include detailed assignment reasons, criteria breakdown, gap analysis showing roles/staff still needed, and override tracking. Created placeholder Report modules in backend and frontend. | System |
| 2026-01-17 | 1.13.0 | Added Rotation Staff Travel Parameters: Added FR-SM-04: Rotation Staff Travel Parameters - system allows configuration of travel parameters (commute duration in minutes, number of transits, travel cost) for each rotation staff per branch. These parameters are used to calculate the shortest path for rotation staff allocation. Default values: commute duration 300 minutes, transit count 10, travel cost 1000. Parameters are stored in effective_branches table and can be configured by Admin and Area Manager roles through the staff management UI. | System |
| 2026-01-17 | 1.14.0 | Added Doctor Default Scheduling and Overrides: Added FR-DP-04: Doctor Default Scheduling and Overrides - system supports default weekly schedules (default branch for each day of week), default weekly off days, and date-specific overrides (working day with branch or off day). Admin and Area Manager can set default schedules representing typical weekly rotation patterns (3-5 branches per week), set recurring weekly off days, and create overrides for specific dates. Overrides take precedence over default schedules. Priority order: Override > Weekly Off Day > Default Branch Assignment. Created database tables (doctor_default_schedules, doctor_weekly_off_days, doctor_schedule_overrides), API endpoints, and UI components for managing default schedules and overrides. | System |

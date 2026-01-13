---
title: Software Requirements Specification
description: Detailed requirements for VSQ Operations Manpower System
version: 1.8.0
lastUpdated: 2026-01-09 10:46:30
---

# VSQ Operations Manpower - Software Requirements Specification

## Document Information

- **Version:** 1.8.0
- **Last Updated:** 2026-01-09 10:46:30
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
- **Status:** ⚠️ Partially Implemented (needs enhancement for rotation staff permissions)
- **Roles:**
  1. **Admin:** Can set system configurations, manage users and roles. Can view/add/edit/delete/import rotation staff. Can assign rotation staff to branches. Can edit rotation staff schedules.
  2. **Area Manager:** Can assign rotation staff to branches. Can view/add/edit/delete/import rotation staff. Can edit rotation staff schedules.
  3. **District Manager:** Can view rotation staff assignments (read-only). Cannot manage rotation staff or edit schedules.
  4. **Branch Manager:** Can allocate staff workday, off day, and leave day in their branch. Can add, edit, and manage staff in their branch. Can only access staff and schedules from their assigned branch (enforced by branch code). Can view rotation staff assigned to their branch (read-only).
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
- **Description:** System shall support various staff positions
- **Status:** ✅ Implemented
- **Positions:**
  - Branch Manager
  - Assistant Branch Manager
  - Service Consultant
  - Coordinator
  - Doctor Assistant
  - Physiotherapist
  - Nurses
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
- **Status:** ❌ Not Implemented
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
    - Branches typically have 1-4 doctors assigned per day (doctor-on days)
    - When a branch has no doctor assigned (doctor-off day), only one staff member is required
    - The single staff member on doctor-off days must have the "can manage a branch alone" skill
    - Doctor assignments determine whether a day is doctor-on or doctor-off
    - Maximum 4 doctors can be assigned to a branch per day
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
- **Description:** System shall manage doctor information and daily expected revenue
- **Status:** ❌ Not Implemented
- **Details:**
  - **Doctor Information:**
    - Doctor ID (unique identifier)
    - Doctor name (full name)
    - Doctor code/nickname (optional)
    - Doctor specialization/type (optional)
    - Contact information (optional)
  - **Doctor-Branch Assignment:**
    - Doctors can be assigned to branches on specific dates
    - A doctor can be assigned to multiple branches on different dates
    - A branch can have multiple doctors on the same day
    - Doctor assignments are date-specific (can vary day by day)
  - **Doctor Daily Expected Revenue:**
    - Each doctor has a daily expected revenue value per branch per date
    - Expected revenue can vary by:
      - Date (different days may have different expected revenue)
      - Branch (same doctor may have different expected revenue at different branches)
    - Expected revenue is stored per doctor per branch per date
  - **Business Rules:**
    - Doctor assignments determine branch operational status (if no doctor, branch is non-operational)
    - Branch expected revenue can be calculated as sum of all assigned doctors' daily expected revenue
    - Doctor assignments can be imported from Excel
    - Doctor assignments can be manually created/edited
    - Doctor expected revenue can be imported from Excel
    - Doctor expected revenue can be manually set per doctor per branch per date
  - **Use Cases:**
    - Admin/Area Manager assigns doctors to branches for specific dates
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
    2. **Doctor Expected Revenue:** Expected revenue can be calculated from doctor daily expected revenue
       - System aggregates daily expected revenue from all doctors assigned to the branch on that day
       - Doctor data structure required (see FR-BM-05)
       - Calculation: Sum of all doctors' daily expected revenue for the branch on that date
  - **Business Rules:**
    - If no expected revenue is set for a day, default to 0 or use previous day's value (configurable)
    - Expected revenue from Excel import can override doctor-calculated revenue
    - Doctor-calculated revenue is automatically updated when doctor assignments change
    - Revenue data must support date ranges for bulk operations

### 3.4 Staff Scheduling

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
- **Status:** ⚠️ Partially Implemented (needs enhancement)
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

### 3.5 Business Logic

#### FR-BL-01: Revenue-Based Staff Calculation
- **Description:** System shall calculate required staff based on expected revenue
- **Status:** ⚠️ Partially Implemented (needs enhancement for preferred staff per position)
- **Details:**
  - Formula: `staff_count = min_staff + (revenue_threshold * multiplier)`
  - Configurable via System Settings
  - Minimum staff per position regardless of revenue
  - Doctor count consideration
  - **Preferred Staff per Position:** Each branch has a preferred number of staff for each position
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
- **Status:** ❌ Not Implemented
- **Details:**
  - Each branch has minimum staff requirements for specific positions on each day
  - Minimum requirements are position-based (e.g., at least 1 front/counter staff and 1 doctor assistant = minimum 2 total)
  - Minimum daily staff counts act as constraints when allocating rotation staff
  - System must ensure minimum requirements are met before allocating additional staff
  - Minimum requirements can vary by day (e.g., different requirements for weekdays vs weekends)
  - Minimum requirements are configurable per branch per position per day
  - Violations of minimum requirements should trigger warnings or prevent allocation

### 3.6 AI Suggestions

#### FR-AI-01: MCP Integration
- **Description:** System shall provide AI suggestions via MCP
- **Status:** ⚠️ Partially Implemented
- **Details:**
  - MCP client framework implemented
  - Suggestion generation endpoint exists (placeholder)
  - Regenerate suggestions capability (placeholder)
  - Multiple suggestions can be generated and compared

### 3.7 Internationalization

#### FR-I18N-01: Multi-Language Support
- **Description:** System shall support Thai and English languages
- **Status:** ⚠️ Partially Implemented
- **Details:**
  - Language toggle in UI (structure ready)
  - Translation files for Thai and English (created)
  - Date/time formatting in Thailand timezone (✅ Implemented)

### 3.8 Dashboard and Reporting

#### FR-RP-01: Dashboard View
- **Description:** System shall provide dashboard with summary statistics
- **Status:** ⚠️ Partially Implemented
- **Details:**
  - Summary statistics (basic implementation)
  - Staff allocation overview
  - Revenue vs. staff allocation charts (❌ Not Implemented)

#### FR-RP-02: Reporting Capabilities
- **Description:** System shall provide reporting capabilities
- **Status:** ❌ Not Implemented
- **Details:**
  - Staff allocation reports
  - Revenue vs. staff efficiency reports
  - Conflict history reports
  - Export reports to PDF/Excel formats
  - Scheduled report generation

### 3.9 Data Validation

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

### 3.10 Notifications

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
- **Status:** ❌ Not Implemented
- **Details:**
  - Expected revenue can be set manually, imported from Excel, or calculated from doctor assignments
  - When calculated from doctors: Sum of all assigned doctors' daily expected revenue for that branch on that date
  - Excel import can override doctor-calculated revenue
  - Manual entry can override both Excel and doctor-calculated revenue
  - Priority: Manual entry > Excel import > Doctor-calculated

### BR-BL-08: Branch Operational Status
- **Description:** Branch operational status is determined by doctor assignments and manual settings
- **Status:** ❌ Not Implemented
- **Details:**
  - Branch is operational if: Has at least one doctor assigned AND not manually marked as closed
  - Branch is non-operational if: No doctor assigned OR manually marked as closed
  - Operational status affects staff allocation requirements
  - Non-operational branches require minimal or no staff allocation
  - System should warn when assigning staff to non-operational branches

### BR-BL-09: Preferred Staff per Position per Branch
- **Description:** Each branch has preferred number of staff for each position
- **Status:** ❌ Not Implemented
- **Details:**
  - Preferred staff counts are branch-specific and position-specific
  - Preferred counts act as constraints when allocating rotation staff
  - System should prioritize meeting preferred staff counts when allocating rotation staff
  - Preferred counts are configurable per branch per position
  - Preferred counts are used as targets for staff allocation optimization

### BR-BL-10: Minimum Daily Staff Constraints
- **Description:** Each branch has minimum staff requirements for specific positions on each day
- **Status:** ❌ Not Implemented
- **Details:**
  - Minimum requirements are position-based (e.g., at least 1 front/counter staff and 1 doctor assistant = minimum 2 total)
  - Minimum daily staff counts act as constraints when allocating rotation staff
  - System must ensure minimum requirements are met before allocating additional staff
  - Minimum requirements can vary by day (e.g., different requirements for weekdays vs weekends)
  - Minimum requirements are configurable per branch per position per day
  - Violations of minimum requirements should trigger warnings or prevent allocation

### BR-BL-11: Doctor-Off Day Staff Requirement
- **Description:** When a branch has no doctor assigned (doctor-off day), only one staff with "can manage a branch alone" skill is required
- **Status:** ❌ Not Implemented
- **Details:**
  - Branches typically have 1-4 doctors assigned per day (doctor-on days)
  - When a branch has no doctor assigned (doctor-off day), only one staff member is required
  - The single staff member on doctor-off days must have the "can manage a branch alone" skill
  - System should validate that assigned staff has the required skill before allowing assignment
  - Doctor assignments determine whether a day is doctor-on or doctor-off
  - Maximum 4 doctors can be assigned to a branch per day

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
| 2024-12-19 | 1.0.0 | Initial requirements document created | System |
| 2025-12-18 | 1.1.0 | Added conflict resolution requirements, data validation, notifications, reporting. Added references to security and conflict resolution documents. Updated status of FR-BL-03. | System |
| 2025-12-21 | 1.2.0 | Added FR-BM-03: Standard Branch Codes requirement. Documented hardcoded list of 35 standard branch codes that must always be available in the system. | System |
| 2025-12-21 | 1.3.0 | Enhanced FR-SM-01: Added leave day support requirement, table layout specification (left to right dates), and staff display requirements (nickname, full name, position). Enhanced FR-RM-03: Added nickname field requirement and clarified module separation. Updated FR-AUZ-01: Clarified Branch Manager capabilities. | System |
| 2025-12-21 | 1.4.0 | Added BR-BL-06: Branch Manager Access Control. Branch managers can only add/arrange staff of their own branch, enforced by branch code. Updated FR-AUZ-01 to include branch manager restrictions. Implemented branch_id field in users table and middleware to enforce branch access. | System |
| 2025-12-22 | 1.5.0 | Enhanced FR-RM-03: Added branch manager restrictions - manual entry only for branch staff of that branch, cannot add rotation staff. Enhanced FR-SM-01: Added requirement that branch manager can see rotation staff assigned to her branch (read-only), and requirement to display branch name and branch code in staff schedule menu. | System |
| 2025-12-23 | 1.6.0 | Enhanced FR-RM-03: Clarified rotation staff management permissions - only Admin and Area Manager can view/add/edit/delete/import rotation staff. Enhanced FR-SM-02: Updated assignment workflow - rotation staff appear in branch table, admin/area manager selects dates. Added FR-SM-03: Rotation Staff Schedule Management - schedule edit menu for day off/leave/sick leave/working days. Updated FR-AUZ-01: Clarified role permissions for rotation staff management. | System |
| 2025-12-23 | 1.7.0 | Enhanced FR-BM-02: Updated revenue tracking to support 365 days per year with multiple input sources (Excel import and doctor-calculated). Added FR-BM-04: Branch Operational Status - tracks daily operational status (operational/no doctor/closed). Added FR-BM-05: Doctor Data Structure - doctor information, branch assignments, and daily expected revenue. Added BR-BL-07: Branch Expected Revenue Calculation rules. Added BR-BL-08: Branch Operational Status rules. Updated FR-BM-01: Added reference to operational status. | System |
| 2026-01-09 | 1.8.0 | Enhanced FR-BL-01: Added preferred staff per position per branch requirement. Added FR-BL-05: Minimum Daily Staff Constraints - position-based minimum requirements per branch per day. Enhanced FR-BM-04: Added doctor-on/doctor-off days logic with "can manage a branch alone" skill requirement. Added FR-RM-04: Staff Skills and Qualifications - tracks staff skills including "can manage a branch alone". Added BR-BL-09: Preferred Staff per Position per Branch rules. Added BR-BL-10: Minimum Daily Staff Constraints rules. Added BR-BL-11: Doctor-Off Day Staff Requirement rules. | System |



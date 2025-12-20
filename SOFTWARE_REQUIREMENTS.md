---
title: Software Requirements Specification
description: Detailed requirements for VSQ Operations Manpower System
version: 1.2.0
lastUpdated: 2025-12-21 00:28:16
---

# VSQ Operations Manpower - Software Requirements Specification

## Document Information

- **Version:** 1.2.0
- **Last Updated:** 2025-12-21 00:28:16
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
- **Status:** ✅ Implemented
- **Roles:**
  1. **Admin:** Can set system configurations, manage users and roles
  2. **Area Manager/District Manager:** Can assign rotation staff to branches
  3. **Branch Manager:** Can allocate staff workday and off day in their branch
  4. **Viewer:** Can only view staff allocation and dashboard data

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
  - Staff information includes: name, type, position, branch (for branch staff), coverage area (for rotation staff)

### 3.3 Branch Management

#### FR-BM-01: Branch Configuration
- **Description:** System shall manage 32 branches
- **Status:** ✅ Implemented
- **Details:**
  - Each branch has: name, code, address
  - Expected revenue can be set per branch
  - Priority level can be assigned
  - Area Manager can be assigned to branches

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
- **Description:** System shall track expected and actual revenue
- **Status:** ✅ Implemented
- **Details:**
  - Expected revenue can be set per day
  - Actual revenue can be recorded
  - Revenue history is maintained

### 3.4 Staff Scheduling

#### FR-SM-01: Branch Staff Scheduling
- **Description:** Branch managers can schedule their branch staff
- **Status:** ✅ Implemented
- **Details:**
  - Monthly calendar view (staff × dates matrix)
  - Can schedule up to 30 days in advance
  - Mark working days and off days
  - Usually branch staff works 6 days a week (configurable)

#### FR-SM-02: Rotation Staff Assignment
- **Description:** Area/District managers can assign rotation staff to branches
- **Status:** ✅ Implemented
- **Details:**
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
- **Status:** ✅ Implemented
- **Details:**
  - Formula: `staff_count = min_staff + (revenue_threshold * multiplier)`
  - Configurable via System Settings
  - Minimum staff per position regardless of revenue
  - Doctor count consideration

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



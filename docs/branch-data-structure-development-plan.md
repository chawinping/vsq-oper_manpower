---
title: Branch Data Structure Development Plan
description: Phased development plan for branch revenue tracking, operational status, and doctor data structure
version: 1.0.0
lastUpdated: 2025-12-23 13:14:23
---

# Branch Data Structure Development Plan

## Overview

This document outlines the phased development plan for implementing the enhanced branch data structure requirements:
- **FR-BM-02:** Enhanced Revenue Tracking (365 days, multiple input sources)
- **FR-BM-04:** Branch Operational Status
- **FR-BM-05:** Doctor Data Structure
- **BR-BL-07:** Branch Expected Revenue Calculation Rules
- **BR-BL-08:** Branch Operational Status Rules

## Development Phases

---

## Phase 1: Doctor Data Structure Foundation

### Objective
Establish the core doctor data model and basic CRUD operations.

### Tasks

#### 1.1 Database Schema
- [ ] Create `doctors` table
  - `id` (UUID, primary key)
  - `name` (VARCHAR(255), required)
  - `code` (VARCHAR(50), unique, optional)
  - `nickname` (VARCHAR(100), optional)
  - `specialization` (VARCHAR(255), optional)
  - `contact_info` (TEXT, optional)
  - `created_at` (TIMESTAMP)
  - `updated_at` (TIMESTAMP)

- [ ] Create `doctor_branch_assignments` table
  - `id` (UUID, primary key)
  - `doctor_id` (UUID, foreign key to doctors)
  - `branch_id` (UUID, foreign key to branches)
  - `date` (DATE, required)
  - `created_by` (UUID, foreign key to users)
  - `created_at` (TIMESTAMP)
  - `updated_at` (TIMESTAMP)
  - Unique constraint: `(doctor_id, branch_id, date)`
  - Index on `(branch_id, date)` for performance

- [ ] Create `doctor_expected_revenue` table
  - `id` (UUID, primary key)
  - `doctor_id` (UUID, foreign key to doctors)
  - `branch_id` (UUID, foreign key to branches)
  - `date` (DATE, required)
  - `expected_revenue` (DECIMAL(15,2), required)
  - `created_at` (TIMESTAMP)
  - `updated_at` (TIMESTAMP)
  - Unique constraint: `(doctor_id, branch_id, date)`
  - Index on `(branch_id, date)` for performance

- [ ] Create migration script
- [ ] Test migration on development database

#### 1.2 Domain Models
- [ ] Create `Doctor` model (`backend/internal/domain/models/doctor.go`)
- [ ] Create `DoctorBranchAssignment` model
- [ ] Create `DoctorExpectedRevenue` model
- [ ] Add model relationships and JSON tags

#### 1.3 Repository Layer
- [ ] Create `DoctorRepository` interface (`backend/internal/domain/interfaces/repositories.go`)
  - `Create(doctor *Doctor) error`
  - `GetByID(id uuid.UUID) (*Doctor, error)`
  - `List() ([]*Doctor, error)`
  - `Update(doctor *Doctor) error`
  - `Delete(id uuid.UUID) error`

- [ ] Create `DoctorBranchAssignmentRepository` interface
  - `Create(assignment *DoctorBranchAssignment) error`
  - `GetByDoctorID(doctorID uuid.UUID, startDate, endDate time.Time) ([]*DoctorBranchAssignment, error)`
  - `GetByBranchID(branchID uuid.UUID, date time.Time) ([]*DoctorBranchAssignment, error)`
  - `GetByBranchAndDateRange(branchID uuid.UUID, startDate, endDate time.Time) ([]*DoctorBranchAssignment, error)`
  - `Delete(id uuid.UUID) error`
  - `DeleteByDoctorBranchDate(doctorID, branchID uuid.UUID, date time.Time) error`

- [ ] Create `DoctorExpectedRevenueRepository` interface
  - `Create(revenue *DoctorExpectedRevenue) error`
  - `GetByDoctorBranchDate(doctorID, branchID uuid.UUID, date time.Time) (*DoctorExpectedRevenue, error)`
  - `GetByBranchDateRange(branchID uuid.UUID, startDate, endDate time.Time) ([]*DoctorExpectedRevenue, error)`
  - `Update(revenue *DoctorExpectedRevenue) error`
  - `Delete(id uuid.UUID) error`

- [ ] Implement PostgreSQL repositories
- [ ] Add repository methods to `postgres.Repositories` struct

#### 1.4 Handler Layer
- [ ] Create `DoctorHandler` (`backend/internal/handlers/doctor_handler.go`)
  - `Create(c *gin.Context)` - POST /api/doctors
  - `List(c *gin.Context)` - GET /api/doctors
  - `GetByID(c *gin.Context)` - GET /api/doctors/:id
  - `Update(c *gin.Context)` - PUT /api/doctors/:id
  - `Delete(c *gin.Context)` - DELETE /api/doctors/:id

- [ ] Create `DoctorBranchAssignmentHandler`
  - `Assign(c *gin.Context)` - POST /api/doctors/:doctorId/assignments
  - `GetAssignments(c *gin.Context)` - GET /api/doctors/:doctorId/assignments
  - `GetBranchAssignments(c *gin.Context)` - GET /api/branches/:branchId/doctor-assignments
  - `RemoveAssignment(c *gin.Context)` - DELETE /api/doctors/assignments/:id

- [ ] Create `DoctorExpectedRevenueHandler`
  - `SetRevenue(c *gin.Context)` - POST /api/doctors/:doctorId/revenue
  - `GetRevenue(c *gin.Context)` - GET /api/doctors/:doctorId/revenue
  - `GetBranchRevenue(c *gin.Context)` - GET /api/branches/:branchId/doctor-revenue

- [ ] Add routes to main router
- [ ] Add permission checks (Admin and Area Manager only)

#### 1.5 Testing
- [ ] Unit tests for repository methods
- [ ] Integration tests for handlers
- [ ] Test doctor CRUD operations
- [ ] Test assignment operations
- [ ] Test revenue operations

### Deliverables
- Database schema with doctors, assignments, and revenue tables
- Complete CRUD API for doctors
- API for doctor-branch assignments
- API for doctor expected revenue

### Estimated Duration
**2-3 weeks**

---

## Phase 2: Branch Operational Status

### Objective
Implement branch operational status tracking with automatic calculation from doctor assignments.

### Tasks

#### 2.1 Database Schema
- [ ] Create `branch_operational_status` table
  - `id` (UUID, primary key)
  - `branch_id` (UUID, foreign key to branches)
  - `date` (DATE, required)
  - `status` (VARCHAR(20), required) - 'operational', 'no_doctor', 'closed'
  - `reason` (TEXT, optional) - Additional context
  - `is_auto_calculated` (BOOLEAN, default false) - Whether status was auto-calculated
  - `created_by` (UUID, foreign key to users, nullable)
  - `created_at` (TIMESTAMP)
  - `updated_at` (TIMESTAMP)
  - Unique constraint: `(branch_id, date)`
  - Index on `(branch_id, date)` for performance
  - Check constraint: `status IN ('operational', 'no_doctor', 'closed')`

- [ ] Create migration script
- [ ] Test migration

#### 2.2 Domain Models
- [ ] Create `BranchOperationalStatus` model
- [ ] Add status constants/enum
- [ ] Add model relationships

#### 2.3 Business Logic
- [ ] Create `OperationalStatusService` (`backend/internal/usecases/operational_status/`)
  - `CalculateStatus(branchID uuid.UUID, date time.Time) (BranchOperationalStatus, error)`
    - Check if branch has doctor assignments for the date
    - Check if manually marked as closed
    - Return appropriate status
  - `GetStatus(branchID uuid.UUID, date time.Time) (*BranchOperationalStatus, error)`
  - `SetStatus(status *BranchOperationalStatus, isManual bool) error`
  - `GetStatusRange(branchID uuid.UUID, startDate, endDate time.Time) ([]*BranchOperationalStatus, error)`
  - `AutoCalculateStatusForDateRange(branchID uuid.UUID, startDate, endDate time.Time) error`

- [ ] Create background job/trigger to auto-calculate status when:
  - Doctor assignments change
  - Date range is updated

#### 2.4 Repository Layer
- [ ] Add `BranchOperationalStatusRepository` interface
  - `Create(status *BranchOperationalStatus) error`
  - `GetByBranchDate(branchID uuid.UUID, date time.Time) (*BranchOperationalStatus, error)`
  - `GetByBranchDateRange(branchID uuid.UUID, startDate, endDate time.Time) ([]*BranchOperationalStatus, error)`
  - `Update(status *BranchOperationalStatus) error`
  - `Delete(id uuid.UUID) error`

- [ ] Implement PostgreSQL repository

#### 2.5 Handler Layer
- [ ] Create `BranchOperationalStatusHandler`
  - `GetStatus(c *gin.Context)` - GET /api/branches/:branchId/operational-status?date=YYYY-MM-DD
  - `GetStatusRange(c *gin.Context)` - GET /api/branches/:branchId/operational-status?start_date=...&end_date=...
  - `SetStatus(c *gin.Context)` - POST /api/branches/:branchId/operational-status
  - `UpdateStatus(c *gin.Context)` - PUT /api/branches/:branchId/operational-status/:id
  - `AutoCalculateStatus(c *gin.Context)` - POST /api/branches/:branchId/operational-status/auto-calculate

- [ ] Add routes
- [ ] Add permission checks

#### 2.6 Integration
- [ ] Hook into doctor assignment handlers to auto-calculate status
- [ ] Add validation to prevent staff assignment to non-operational branches
- [ ] Add warnings in UI when assigning staff to non-operational branches

#### 2.7 Testing
- [ ] Unit tests for operational status calculation logic
- [ ] Integration tests for handlers
- [ ] Test auto-calculation triggers
- [ ] Test manual status override
- [ ] Test status affects staff allocation

### Deliverables
- Branch operational status tracking
- Automatic status calculation from doctor assignments
- Manual status override capability
- API endpoints for status management

### Estimated Duration
**1-2 weeks**

---

## Phase 3: Enhanced Revenue Tracking - Doctor-Calculated Revenue

### Objective
Implement automatic branch expected revenue calculation from doctor assignments.

### Tasks

#### 3.1 Business Logic
- [ ] Create `RevenueCalculationService` (`backend/internal/usecases/revenue/`)
  - `CalculateBranchRevenueFromDoctors(branchID uuid.UUID, date time.Time) (float64, error)`
    - Get all doctor assignments for branch on date
    - Sum all doctor expected revenue for those assignments
    - Return total expected revenue
  - `CalculateBranchRevenueRange(branchID uuid.UUID, startDate, endDate time.Time) (map[string]float64, error)`
  - `RecalculateRevenueForDateRange(branchID uuid.UUID, startDate, endDate time.Time) error`
    - Calculate revenue for each date in range
    - Update revenue_data table (only if source is 'doctor_calculated')

- [ ] Add revenue source tracking to `revenue_data` table
  - Add `source` column (VARCHAR(50)) - 'manual', 'excel_import', 'doctor_calculated'
  - Add `source_id` column (UUID, nullable) - Reference to source (doctor_id, import_id, etc.)
  - Migration script

#### 3.2 Integration Points
- [ ] Hook into doctor assignment handlers
  - When doctor assigned: Recalculate branch revenue for that date
  - When doctor unassigned: Recalculate branch revenue for that date
  - When doctor expected revenue updated: Recalculate branch revenue

- [ ] Hook into doctor expected revenue handlers
  - When revenue set: Recalculate branch revenue for affected dates

- [ ] Create background job for bulk recalculation
  - Recalculate all branches for date range
  - Can be triggered manually or scheduled

#### 3.3 Repository Updates
- [ ] Update `RevenueRepository` interface
  - `GetByBranchDateWithSource(branchID uuid.UUID, date time.Time) (*RevenueData, error)`
  - `CreateOrUpdateWithSource(revenue *RevenueData) error`
  - `GetBySource(source string, startDate, endDate time.Time) ([]*RevenueData, error)`

- [ ] Update PostgreSQL implementation

#### 3.4 Handler Updates
- [ ] Update `BranchHandler.GetRevenue` to show source
- [ ] Add endpoint: `POST /api/branches/:branchId/revenue/recalculate`
- [ ] Add endpoint: `POST /api/branches/:branchId/revenue/recalculate-range`

#### 3.5 Business Rules Implementation
- [ ] Implement BR-BL-07: Revenue calculation priority
  - Manual entry > Excel import > Doctor-calculated
  - Only update if current source has lower priority
  - Preserve manual/excel entries when recalculating

#### 3.6 Testing
- [ ] Unit tests for revenue calculation logic
- [ ] Test calculation from multiple doctors
- [ ] Test priority rules
- [ ] Test recalculation triggers
- [ ] Integration tests

### Deliverables
- Automatic branch revenue calculation from doctors
- Revenue source tracking
- Priority-based revenue updates
- Recalculation API endpoints

### Estimated Duration
**1-2 weeks**

---

## Phase 4: Excel Import for Revenue and Doctor Data

### Objective
Implement Excel import functionality for expected revenue and doctor assignments/revenue.

### Tasks

#### 4.1 Excel Import Infrastructure
- [ ] Enhance existing Excel importer (`backend/pkg/excel/importer.go`)
  - Add support for revenue import format
  - Add support for doctor assignment import format
  - Add support for doctor revenue import format
  - Add validation functions

- [ ] Create Excel template files
  - `revenue_import_template.xlsx`
  - `doctor_assignment_import_template.xlsx`
  - `doctor_revenue_import_template.xlsx`

#### 4.2 Revenue Import
- [ ] Create `RevenueImportService`
  - `ParseRevenueExcel(file io.Reader) ([]RevenueImportRow, error)`
  - `ValidateRevenueData(rows []RevenueImportRow) error`
  - `ImportRevenueData(rows []RevenueImportRow, source string) error`
    - Validate branch codes
    - Validate dates
    - Validate revenue values
    - Create/update revenue_data records with source='excel_import'

- [ ] Create `RevenueImportHandler`
  - `ImportRevenue(c *gin.Context)` - POST /api/revenue/import
  - `DownloadTemplate(c *gin.Context)` - GET /api/revenue/import/template
  - `ValidateImport(c *gin.Context)` - POST /api/revenue/import/validate

- [ ] Add file upload middleware
- [ ] Add routes

#### 4.3 Doctor Assignment Import
- [ ] Create `DoctorAssignmentImportService`
  - `ParseAssignmentExcel(file io.Reader) ([]DoctorAssignmentImportRow, error)`
  - `ValidateAssignmentData(rows []DoctorAssignmentImportRow) error`
  - `ImportAssignmentData(rows []DoctorAssignmentImportRow) error`
    - Validate doctor codes/names
    - Validate branch codes
    - Validate dates
    - Create doctor_branch_assignments records
    - Trigger operational status recalculation
    - Trigger revenue recalculation

- [ ] Create `DoctorAssignmentImportHandler`
  - `ImportAssignments(c *gin.Context)` - POST /api/doctors/assignments/import
  - `DownloadTemplate(c *gin.Context)` - GET /api/doctors/assignments/import/template
  - `ValidateImport(c *gin.Context)` - POST /api/doctors/assignments/import/validate

#### 4.4 Doctor Revenue Import
- [ ] Create `DoctorRevenueImportService`
  - `ParseRevenueExcel(file io.Reader) ([]DoctorRevenueImportRow, error)`
  - `ValidateRevenueData(rows []DoctorRevenueImportRow) error`
  - `ImportRevenueData(rows []DoctorRevenueImportRow) error`
    - Validate doctor codes/names
    - Validate branch codes
    - Validate dates and revenue values
    - Create/update doctor_expected_revenue records
    - Trigger branch revenue recalculation

- [ ] Create `DoctorRevenueImportHandler`
  - `ImportRevenue(c *gin.Context)` - POST /api/doctors/revenue/import
  - `DownloadTemplate(c *gin.Context)` - GET /api/doctors/revenue/import/template
  - `ValidateImport(c *gin.Context)` - POST /api/doctors/revenue/import/validate

#### 4.5 Error Handling and Reporting
- [ ] Create import result structure
  - Success count
  - Error count
  - Error details (row number, field, error message)
  - Warnings

- [ ] Return detailed import results
- [ ] Log import operations

#### 4.6 Testing
- [ ] Unit tests for Excel parsing
- [ ] Unit tests for validation
- [ ] Integration tests for import handlers
- [ ] Test with sample Excel files
- [ ] Test error handling
- [ ] Test bulk imports (365 days)

### Deliverables
- Excel import for branch expected revenue
- Excel import for doctor assignments
- Excel import for doctor expected revenue
- Import templates
- Validation and error reporting

### Estimated Duration
**2-3 weeks**

---

## Phase 5: Frontend Implementation

### Objective
Create user interfaces for managing doctors, operational status, and revenue.

### Tasks

#### 5.1 Doctor Management UI
- [ ] Create doctor list page (`frontend/src/app/(admin)/doctors/page.tsx`)
  - List all doctors
  - Add/edit/delete doctors
  - Search and filter

- [ ] Create doctor form component
  - Name, code, nickname, specialization, contact fields
  - Validation

- [ ] Create doctor detail page
  - Show doctor information
  - Show assignments calendar
  - Show revenue calendar

#### 5.2 Doctor Assignment UI
- [ ] Create doctor assignment calendar (`frontend/src/components/doctors/DoctorAssignmentCalendar.tsx`)
  - Monthly calendar view
  - Assign doctors to branches by date
  - Show existing assignments
  - Bulk assignment capability

- [ ] Create branch doctor assignment view
  - Show all doctors assigned to branch
  - Calendar view of assignments
  - Add/remove assignments

#### 5.3 Doctor Revenue UI
- [ ] Create doctor revenue management page
  - Set expected revenue per doctor per branch per date
  - Calendar view
  - Bulk update capability

#### 5.4 Branch Operational Status UI
- [ ] Create operational status calendar (`frontend/src/components/branches/OperationalStatusCalendar.tsx`)
  - Monthly calendar view per branch
  - Show status for each day (operational/no doctor/closed)
  - Manual status override
  - Visual indicators (colors/icons)

- [ ] Integrate into branch management page
  - Show operational status in branch list
  - Link to status calendar

#### 5.5 Enhanced Revenue UI
- [ ] Update branch revenue page
  - Show revenue source (manual/excel/doctor-calculated)
  - Show 365-day calendar view
  - Manual entry
  - Recalculate from doctors button
  - Visual indicators for source

- [ ] Create revenue import page
  - Upload Excel file
  - Preview import data
  - Show validation errors
  - Confirm and import
  - Download template

#### 5.6 Excel Import UI
- [ ] Create import page component (`frontend/src/components/import/ExcelImport.tsx`)
  - File upload
  - Template download
  - Preview/validation
  - Import confirmation
  - Progress indicator
  - Results display

- [ ] Add import buttons to relevant pages
  - Revenue import
  - Doctor assignment import
  - Doctor revenue import

#### 5.7 API Integration
- [ ] Create API client functions (`frontend/src/lib/api/`)
  - `doctor.ts` - Doctor CRUD operations
  - `doctorAssignment.ts` - Assignment operations
  - `doctorRevenue.ts` - Revenue operations
  - `operationalStatus.ts` - Status operations
  - `revenueImport.ts` - Import operations

- [ ] Update existing API clients
  - Update `branch.ts` for operational status
  - Update `revenue.ts` for enhanced revenue tracking

#### 5.8 Validation and Error Handling
- [ ] Add form validation
- [ ] Add error messages
- [ ] Add success notifications
- [ ] Handle API errors gracefully

#### 5.9 Testing
- [ ] Component unit tests
- [ ] E2E tests for doctor management
- [ ] E2E tests for assignments
- [ ] E2E tests for revenue import
- [ ] E2E tests for operational status

### Deliverables
- Complete UI for doctor management
- UI for doctor assignments
- UI for operational status
- UI for revenue management and import
- Excel import interfaces

### Estimated Duration
**3-4 weeks**

---

## Phase 6: Integration and Staff Allocation Updates

### Objective
Integrate new data structures into staff allocation calculations and update business logic.

### Tasks

#### 6.1 Update Staff Allocation Logic
- [ ] Update `AllocationService` (`backend/internal/usecases/allocation/`)
  - Check branch operational status before allocation
  - Skip or reduce allocation for non-operational branches
  - Use enhanced revenue data (with source tracking)

- [ ] Update revenue-based staff calculation
  - Use doctor-calculated revenue when available
  - Fall back to manual/excel revenue
  - Consider operational status

#### 6.2 Update Business Rules
- [ ] Implement BR-BL-08: Operational status affects allocation
  - Non-operational branches: minimal or no staff
  - Operational branches: normal allocation
  - Warning when assigning to non-operational branches

- [ ] Update conflict detection
  - Check operational status in conflict resolution
  - Warn about non-operational branch assignments

#### 6.3 Update Dashboard and Reports
- [ ] Update dashboard to show:
  - Operational status summary
  - Revenue sources breakdown
  - Doctor assignment statistics

- [ ] Update reports to include:
  - Operational status history
  - Revenue source analysis
  - Doctor assignment reports

#### 6.4 Performance Optimization
- [ ] Add database indexes for new queries
- [ ] Optimize revenue calculation queries
- [ ] Cache operational status calculations
- [ ] Optimize bulk operations (365-day calculations)

#### 6.5 Documentation
- [ ] Update API documentation
- [ ] Update user guide
- [ ] Document Excel import formats
- [ ] Document operational status rules

#### 6.6 Testing
- [ ] End-to-end testing of allocation with new data
- [ ] Performance testing for bulk operations
- [ ] Test integration points
- [ ] Regression testing

### Deliverables
- Updated staff allocation logic
- Integration with operational status
- Updated reports and dashboard
- Performance optimizations
- Complete documentation

### Estimated Duration
**2 weeks**

---

## Phase 7: Testing and Quality Assurance

### Objective
Comprehensive testing, bug fixes, and performance optimization.

### Tasks

#### 7.1 Unit Testing
- [ ] Achieve >80% code coverage
- [ ] Test all business logic
- [ ] Test edge cases
- [ ] Test error handling

#### 7.2 Integration Testing
- [ ] Test all API endpoints
- [ ] Test database operations
- [ ] Test service integrations
- [ ] Test Excel import workflows

#### 7.3 E2E Testing
- [ ] Test complete doctor management workflow
- [ ] Test assignment workflow
- [ ] Test revenue calculation workflow
- [ ] Test operational status workflow
- [ ] Test Excel import workflows
- [ ] Test staff allocation with new data

#### 7.4 Performance Testing
- [ ] Test bulk operations (365-day calculations)
- [ ] Test Excel import with large files
- [ ] Test concurrent operations
- [ ] Optimize slow queries

#### 7.5 Security Testing
- [ ] Test permission checks
- [ ] Test input validation
- [ ] Test file upload security
- [ ] Test SQL injection prevention

#### 7.6 User Acceptance Testing
- [ ] Prepare test scenarios
- [ ] Conduct UAT sessions
- [ ] Collect feedback
- [ ] Fix issues

### Deliverables
- Comprehensive test suite
- Performance optimizations
- Bug fixes
- UAT completion

### Estimated Duration
**2-3 weeks**

---

## Phase 8: Deployment and Migration

### Objective
Deploy to staging/production and migrate existing data.

### Tasks

#### 8.1 Data Migration
- [ ] Create migration scripts for existing revenue data
  - Set source='manual' for existing records
  - Preserve all existing data

- [ ] Create migration scripts for operational status
  - Calculate initial status from existing data (if possible)
  - Set default status for historical data

- [ ] Test migrations on staging database

#### 8.2 Deployment Preparation
- [ ] Update deployment scripts
- [ ] Update environment variables
- [ ] Prepare rollback plan
- [ ] Create deployment checklist

#### 8.3 Staging Deployment
- [ ] Deploy to staging environment
- [ ] Run migrations
- [ ] Verify functionality
- [ ] Performance testing
- [ ] User acceptance testing

#### 8.4 Production Deployment
- [ ] Schedule deployment window
- [ ] Backup production database
- [ ] Deploy application
- [ ] Run migrations
- [ ] Verify functionality
- [ ] Monitor for issues

#### 8.5 Post-Deployment
- [ ] Monitor logs and errors
- [ ] Monitor performance
- [ ] Collect user feedback
- [ ] Address issues quickly
- [ ] Update documentation

### Deliverables
- Migrated production data
- Deployed application
- Monitoring and support

### Estimated Duration
**1-2 weeks**

---

## Overall Timeline

| Phase | Duration | Dependencies |
|-------|----------|--------------|
| Phase 1: Doctor Data Structure | 2-3 weeks | None |
| Phase 2: Operational Status | 1-2 weeks | Phase 1 |
| Phase 3: Doctor-Calculated Revenue | 1-2 weeks | Phase 1, Phase 2 |
| Phase 4: Excel Import | 2-3 weeks | Phase 1, Phase 3 |
| Phase 5: Frontend Implementation | 3-4 weeks | Phase 1, Phase 2, Phase 3 |
| Phase 6: Integration | 2 weeks | Phase 1-5 |
| Phase 7: Testing & QA | 2-3 weeks | Phase 1-6 |
| Phase 8: Deployment | 1-2 weeks | Phase 7 |
| **Total** | **14-21 weeks** | |

**Estimated Total Duration: 3.5 - 5 months**

---

## Risk Mitigation

### Technical Risks
1. **Performance Issues with 365-Day Calculations**
   - Mitigation: Implement caching, batch processing, background jobs
   - Early performance testing

2. **Excel Import Complexity**
   - Mitigation: Use proven Excel libraries, extensive validation, clear error messages
   - Create comprehensive templates

3. **Data Consistency**
   - Mitigation: Use database transactions, implement proper validation, add constraints
   - Regular data integrity checks

### Business Risks
1. **User Adoption**
   - Mitigation: Comprehensive training, clear documentation, intuitive UI
   - Gradual rollout

2. **Data Migration Issues**
   - Mitigation: Thorough testing on staging, backup plans, rollback procedures
   - Phased migration

---

## Success Criteria

- [ ] All requirements (FR-BM-02, FR-BM-04, FR-BM-05) fully implemented
- [ ] All business rules (BR-BL-07, BR-BL-08) implemented and tested
- [ ] Excel import working for all data types
- [ ] Operational status automatically calculated
- [ ] Revenue calculation working with priority rules
- [ ] Staff allocation integrated with new data
- [ ] Performance acceptable for 365-day operations
- [ ] >80% test coverage
- [ ] User acceptance testing passed
- [ ] Production deployment successful

---

## Notes

- Phases can be partially parallelized (e.g., Frontend can start after Phase 1 backend is complete)
- Consider agile sprints (2-week sprints) for better progress tracking
- Regular demos after each phase completion
- Continuous integration and testing throughout development
- Regular stakeholder reviews



---
title: Requirements Gap Analysis
description: Comprehensive analysis of missing features compared to SOFTWARE_REQUIREMENTS.md
version: 1.0.0
lastUpdated: 2025-12-21 00:30:00
---

# Requirements Gap Analysis

## Document Information

- **Version:** 1.0.0
- **Last Updated:** 2025-12-21 00:30:00
- **Status:** Active
- **Related Documents:** SOFTWARE_REQUIREMENTS.md, docs/security-requirements.md, docs/conflict-resolution-rules.md

---

## Executive Summary

This document provides a comprehensive analysis of features that are still missing or partially implemented compared to the requirements specified in `SOFTWARE_REQUIREMENTS.md`. The analysis is based on code review conducted on 2025-12-21.

**Overall Status:**
- ✅ **Fully Implemented:** 15 requirements
- ⚠️ **Partially Implemented:** 8 requirements
- ❌ **Not Implemented:** 12 requirements

---

## 1. Functional Requirements - Missing Features

### 1.1 Staff Management

#### FR-RM-03: Staff Data Management - Excel Import
- **Status:** ⚠️ Partially Implemented
- **Backend:** ✅ Implemented (`backend/pkg/excel/importer.go`, `backend/internal/handlers/staff_handler.go`)
- **Frontend:** ❌ Not Implemented
- **Missing:**
  - UI for file upload
  - File validation UI (file type, size)
  - Import progress indicator
  - Error display for failed rows
  - Success summary with imported count
- **Location:** Backend endpoint exists at `POST /api/staff/import`
- **Priority:** Medium

### 1.2 Business Logic

#### FR-BL-03: Availability Checking
- **Status:** ⚠️ Partially Implemented
- **Issue:** Critical availability check code is commented out
- **Missing:**
  - Check if rotation staff is already assigned on the date (code commented out in `backend/internal/usecases/allocation/allocation.go:75-82`)
  - Integration of availability check in `RotationHandler.Assign()`
- **Location:** `backend/internal/usecases/allocation/allocation.go`
- **Priority:** High

#### FR-BL-04: Conflict Detection and Resolution
- **Status:** ❌ Not Implemented
- **Missing:**
  - Rotation staff double-booking conflict detection
  - Effective branch access conflict detection (partially exists but needs enhancement)
  - Staff shortfall detection
  - Conflict resolution UI
  - Conflict notification system
- **Related Document:** `docs/conflict-resolution-rules.md`
- **Priority:** High

### 1.3 AI Suggestions

#### FR-AI-01: MCP Integration
- **Status:** ⚠️ Partially Implemented
- **Issue:** MCP client exists but not integrated properly
- **Missing:**
  - Actual MCP server integration in `RotationHandler.GetSuggestions()`
  - Currently returns placeholder suggestions (hardcoded logic)
  - Regenerate suggestions doesn't call MCP
- **Location:** `backend/internal/handlers/rotation_handler.go:146-234`
- **Priority:** Medium

### 1.4 Internationalization

#### FR-I18N-01: Multi-Language Support
- **Status:** ⚠️ Partially Implemented
- **Missing:**
  - Language toggle UI component
  - Integration of translation files (`frontend/public/locales/`)
  - i18next configuration in Next.js
  - Translation usage in components
- **Location:** Translation files exist but not integrated
- **Priority:** Low

### 1.5 Dashboard and Reporting

#### FR-RP-01: Dashboard View
- **Status:** ⚠️ Partially Implemented
- **Missing:**
  - Revenue vs. staff allocation charts
  - Actual data display (currently shows "-" placeholders)
  - Staff allocation overview visualization
  - Summary statistics implementation
- **Location:** `frontend/src/app/dashboard/page.tsx`
- **Priority:** Medium

#### FR-RP-02: Reporting Capabilities
- **Status:** ❌ Not Implemented
- **Missing:**
  - Staff allocation reports
  - Revenue vs. staff efficiency reports
  - Conflict history reports
  - Export to PDF functionality
  - Export to Excel functionality
  - Scheduled report generation
  - Report generation endpoints
- **Priority:** Medium

### 1.6 Data Validation

#### FR-DV-01: Input Validation
- **Status:** ⚠️ Partially Implemented
- **Missing:**
  - File upload validation (Excel import)
  - Field length limits verification needed
  - Business rule validation enhancement
- **Priority:** Medium

### 1.7 Notifications

#### FR-NT-01: Conflict Notifications
- **Status:** ❌ Not Implemented
- **Missing:**
  - Notification system infrastructure
  - Conflict notification creation
  - Assignment replacement notifications
  - In-app notification UI
  - Email notification system (optional)
- **Priority:** Medium

#### FR-NT-02: System Notifications
- **Status:** ❌ Not Implemented
- **Missing:**
  - Staff shortfall warnings
  - Schedule change notifications
  - Assignment reminders
  - Notification preferences per user
  - Notification center UI
- **Priority:** Low

---

## 2. Non-Functional Requirements - Missing Features

### 2.1 Performance

#### NFR-PF-01: Response Time
- **Status:** ❌ Not Implemented
- **Missing:**
  - Performance monitoring
  - Response time measurement
  - Performance optimization
  - Target: API responses < 500ms, page loads < 2s
- **Priority:** Medium

### 2.2 Security (from `docs/security-requirements.md`)

#### NFR-SC-02: Password Policy
- **Status:** ❌ Not Implemented
- **Missing:**
  - Password complexity validation
  - Minimum 8 characters, max 128
  - Uppercase, lowercase, number, special character requirements
  - Password strength indicator
- **Priority:** High

#### NFR-SC-04: Password Expiration
- **Status:** ❌ Not Implemented
- **Missing:**
  - Password expiration policy (90 days default)
  - Warning notification 7 days before expiration
  - Force password change on expiration
  - Password history (cannot reuse last 5)
- **Priority:** Medium

#### NFR-SC-05: Password Reset
- **Status:** ❌ Not Implemented
- **Missing:**
  - Password reset via email
  - Secure reset token generation
  - Token expiration (1 hour)
  - Single-use tokens
  - Rate limiting (3 requests per hour)
- **Priority:** High

#### NFR-SC-06: Session Security (Partial)
- **Status:** ⚠️ Partially Implemented
- **Missing:**
  - Session invalidation on password change
  - Concurrent session limit (max 3 active sessions)
- **Priority:** Medium

#### NFR-SC-07: Session Fixation Prevention
- **Status:** ❌ Not Implemented
- **Missing:**
  - Regenerate session ID on login
  - Regenerate session ID on privilege escalation
- **Priority:** Medium

#### NFR-SC-11: API Rate Limiting
- **Status:** ❌ Not Implemented
- **Missing:**
  - Rate limiting middleware
  - 100 requests/minute per user
  - 10 requests/minute for auth endpoints
  - 1000 requests/hour per IP
  - Rate limit headers in response
- **Priority:** High

#### NFR-SC-16: File Upload Security
- **Status:** ❌ Not Implemented (Excel import feature)
- **Missing:**
  - File type validation (only Excel files)
  - File size limits (max 10MB)
  - File content validation
  - Secure file storage
- **Priority:** Medium

#### NFR-SC-17: Security Event Logging
- **Status:** ❌ Not Implemented
- **Missing:**
  - Authentication attempt logging
  - Authorization failure logging
  - Password change logging
  - Role/permission change logging
  - Data modification logging
  - Failed API request logging
  - Structured logging (JSON)
  - Log retention (90 days)
- **Priority:** High

#### NFR-SC-18: Audit Trail (Partial)
- **Status:** ⚠️ Partially Implemented
- **Missing:**
  - Change history tracking (what was changed)
  - Audit trail for staff assignments
  - Audit trail for schedule changes
  - Audit trail query and reporting
- **Priority:** Medium

#### NFR-SC-19: Intrusion Detection
- **Status:** ❌ Not Implemented
- **Missing:**
  - Brute force detection (5+ failed attempts)
  - Automatic account lockout (30 minutes)
  - Alert administrators
  - Unusual access pattern monitoring
- **Priority:** High

#### NFR-SC-20: Security Alerts
- **Status:** ❌ Not Implemented
- **Missing:**
  - Email alerts for failed logins
  - Email alerts for account lockouts
  - Email alerts for role/permission changes
  - Alert configuration management
- **Priority:** Medium

### 2.3 Usability

#### NFR-UI-01: Responsive Design
- **Status:** ⚠️ Partially Implemented
- **Missing:**
  - Mobile-responsive styling
  - Mobile-optimized layouts
  - Touch-friendly UI elements
- **Priority:** Medium

#### NFR-UI-02: User Interface
- **Status:** ⚠️ Partially Implemented
- **Missing:**
  - Complete styling for rotation assignment views
  - Complete styling for monthly calendar view
  - Enhanced navigation
- **Priority:** Low

---

## 3. Implementation Priority Summary

### High Priority (Critical - Must Have)

1. **FR-BL-03:** Availability Checking (uncomment and integrate)
2. **FR-BL-04:** Conflict Detection and Resolution
3. **NFR-SC-02:** Password Policy
4. **NFR-SC-05:** Password Reset
5. **NFR-SC-11:** API Rate Limiting
6. **NFR-SC-17:** Security Event Logging
7. **NFR-SC-19:** Intrusion Detection

### Medium Priority (Important - Should Have)

1. **FR-RM-03:** Excel Import Frontend UI
2. **FR-AI-01:** MCP Integration (complete)
3. **FR-RP-01:** Dashboard Charts
4. **FR-RP-02:** Reporting Capabilities
5. **FR-NT-01:** Conflict Notifications
6. **FR-DV-01:** File Upload Validation
7. **NFR-SC-04:** Password Expiration
8. **NFR-SC-06:** Session Security (complete)
9. **NFR-SC-16:** File Upload Security
10. **NFR-SC-18:** Audit Trail (complete)
11. **NFR-SC-20:** Security Alerts
12. **NFR-PF-01:** Response Time Monitoring
13. **NFR-UI-01:** Responsive Design

### Low Priority (Nice to Have)

1. **FR-I18N-01:** Multi-Language Support (complete integration)
2. **FR-NT-02:** System Notifications
3. **NFR-UI-02:** User Interface (enhancements)

---

## 4. Detailed Code Locations

### Backend Files Needing Updates

1. **Conflict Detection:**
   - `backend/internal/usecases/allocation/allocation.go` - Uncomment availability check (lines 75-82)
   - `backend/internal/handlers/rotation_handler.go` - Add conflict checking in `Assign()` method

2. **Excel Import:**
   - `backend/internal/handlers/staff_handler.go` - Add file validation (lines 139-210)

3. **MCP Integration:**
   - `backend/internal/handlers/rotation_handler.go` - Replace placeholder logic with actual MCP calls (lines 202-231)

4. **Security:**
   - Create `backend/internal/middleware/rate_limit.go` - Rate limiting middleware
   - Create `backend/internal/middleware/password_policy.go` - Password validation
   - Create `backend/internal/services/notification.go` - Notification service
   - Create `backend/internal/services/audit.go` - Audit logging service

5. **Reporting:**
   - Create `backend/internal/handlers/report_handler.go` - Report generation endpoints
   - Create `backend/internal/services/report.go` - Report generation logic

### Frontend Files Needing Updates

1. **Excel Import:**
   - Create `frontend/src/components/staff/StaffImport.tsx` - Import UI component
   - Update `frontend/src/app/(manager)/staff-management/page.tsx` - Add import button

2. **Dashboard:**
   - Update `frontend/src/app/dashboard/page.tsx` - Add charts and real data (lines 15-28)

3. **Conflict Resolution:**
   - Create `frontend/src/components/rotation/ConflictDialog.tsx` - Conflict resolution UI
   - Update `frontend/src/components/rotation/RotationAssignmentView.tsx` - Add conflict handling

4. **Notifications:**
   - Create `frontend/src/components/notifications/NotificationCenter.tsx` - Notification UI
   - Create `frontend/src/contexts/NotificationContext.tsx` - Notification context

5. **Internationalization:**
   - Update `frontend/src/app/layout.tsx` - Add i18next provider
   - Create `frontend/src/components/common/LanguageToggle.tsx` - Language switcher

---

## 5. Database Schema Considerations

### Potential New Tables Needed

1. **Notifications Table:**
   ```sql
   CREATE TABLE notifications (
     id UUID PRIMARY KEY,
     user_id UUID REFERENCES users(id),
     type VARCHAR(50),
     title VARCHAR(255),
     message TEXT,
     read BOOLEAN DEFAULT FALSE,
     created_at TIMESTAMP DEFAULT NOW()
   );
   ```

2. **Audit Log Table:**
   ```sql
   CREATE TABLE audit_logs (
     id UUID PRIMARY KEY,
     user_id UUID REFERENCES users(id),
     action VARCHAR(50),
     resource_type VARCHAR(50),
     resource_id UUID,
     changes JSONB,
     ip_address VARCHAR(45),
     created_at TIMESTAMP DEFAULT NOW()
   );
   ```

3. **Conflict History Table:**
   ```sql
   CREATE TABLE conflict_history (
     id UUID PRIMARY KEY,
     conflict_type VARCHAR(50),
     rotation_staff_id UUID REFERENCES staff(id),
     branch_id UUID REFERENCES branches(id),
     date DATE,
     resolution_action VARCHAR(50),
     resolved_by UUID REFERENCES users(id),
     created_at TIMESTAMP DEFAULT NOW()
   );
   ```

4. **Password History Table:**
   ```sql
   CREATE TABLE password_history (
     id UUID PRIMARY KEY,
     user_id UUID REFERENCES users(id),
     password_hash VARCHAR(255),
     created_at TIMESTAMP DEFAULT NOW()
   );
   ```

---

## 6. Testing Requirements

### Missing Test Coverage

1. **Conflict Detection Tests:**
   - Test rotation staff double-booking
   - Test effective branch validation
   - Test staff shortfall detection

2. **Security Tests:**
   - Password policy validation tests
   - Rate limiting tests
   - Session security tests
   - Intrusion detection tests

3. **Integration Tests:**
   - Excel import end-to-end
   - Conflict resolution flow
   - Notification delivery

---

## 7. Documentation Updates Needed

1. **API Documentation:**
   - Document Excel import endpoint
   - Document conflict resolution endpoints
   - Document notification endpoints
   - Document reporting endpoints

2. **User Documentation:**
   - Excel import guide
   - Conflict resolution guide
   - Notification management guide

---

## 8. Next Steps

### Phase 1: Critical Features (Week 1-2)
1. Implement conflict detection and resolution
2. Complete availability checking
3. Implement password policy
4. Add API rate limiting

### Phase 2: Important Features (Week 3-4)
1. Excel import frontend UI
2. Dashboard charts and data
3. Security logging
4. Intrusion detection

### Phase 3: Enhancement Features (Week 5-6)
1. Reporting capabilities
2. Notification system
3. Password reset
4. Audit trail completion

### Phase 4: Polish Features (Week 7+)
1. Internationalization integration
2. Responsive design
3. UI enhancements
4. Performance optimization

---

## 9. Related Documents

- **SOFTWARE_REQUIREMENTS.md** - Main requirements document
- **docs/security-requirements.md** - Detailed security requirements
- **docs/conflict-resolution-rules.md** - Conflict resolution specifications
- **docs/business-rules.md** - Business rules documentation

---

## 10. Change Log

| Date | Version | Changes | Author |
|------|---------|---------|--------|
| 2025-12-21 | 1.0.0 | Initial gap analysis document created | System |

---





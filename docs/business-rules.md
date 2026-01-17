---
title: Business Rules
description: Business rules and logic constraints for VSQ Operations Manpower
version: 1.2.0
lastUpdated: 2026-01-17 20:34:25
---

# Business Rules

## Overview

This document defines all business rules and logic constraints that govern the behavior of the VSQ Operations Manpower system.

## Status Legend

- ✅ **Implemented** - Rule implemented and tested
- ⚠️ **Partially Implemented** - Basic implementation exists, needs enhancement
- ❌ **Not Implemented** - Planned but not yet started

---

## Business Rules by Module

### Service Management (SM)

| ID | Rule Description | Status | Priority | Implementation Notes |
|----|------------------|--------|----------|----------------------|
| BR-SM-01 | *To be defined* | ❌ | - | - |

### Booking Management (BM)

| ID | Rule Description | Status | Priority | Implementation Notes |
|----|------------------|--------|----------|----------------------|
| BR-BM-01 | *To be defined* | ❌ | - | - |

### User Management (UM)

| ID | Rule Description | Status | Priority | Implementation Notes |
|----|------------------|--------|----------|----------------------|
| BR-UM-01 | *To be defined* | ❌ | - | - |

### Resource Management (RM)

| ID | Rule Description | Status | Priority | Implementation Notes |
|----|------------------|--------|----------|----------------------|
| BR-RM-01 | *To be defined* | ❌ | - | - |

### Branch Logic (BL)

| ID | Rule Description | Status | Priority | Implementation Notes |
|----|------------------|--------|----------|----------------------|
| BR-BL-15 | Branch Working Day Determination and Staff Conversion | ❌ | High | Branch working days require at least one doctor. Branch off days occur when no doctor is scheduled. Branch staff temporarily become rotation staff on branch off days. See SOFTWARE_REQUIREMENTS.md section 5 for details. |
| BR-BL-16 | Doctor Temporary Branch Assignment | ⚠️ | Medium | Doctors can temporarily work at branches not in their default schedule via overrides. Partially implemented (override functionality exists). See SOFTWARE_REQUIREMENTS.md section 5 for details. |
| BR-BL-17 | Default Schedule Expected Revenue Requirement | ❌ | High | Default schedules must have expected revenue associated with each day. See SOFTWARE_REQUIREMENTS.md section 5 for details. |
| BR-BL-18 | Multiple Doctors per Branch Simultaneously | ✅ | High | A branch can have multiple doctors working simultaneously (up to 6 doctors per branch per day). See SOFTWARE_REQUIREMENTS.md section 5 for details. |

### Reporting (RP)

| ID | Rule Description | Status | Priority | Implementation Notes |
|----|------------------|--------|----------|----------------------|
| BR-RP-01 | *To be defined* | ❌ | - | - |

### Authentication & Authorization (AU/AUZ)

| ID | Rule Description | Status | Priority | Implementation Notes |
|----|------------------|--------|----------|----------------------|
| BR-AU-01 | *To be defined* | ❌ | - | - |

---

## Business Rule Template

When documenting a new business rule, use this format:

```markdown
### BR-[Module]-[Number]: [Rule Name]

**Description:** 
Clear description of the business rule.

**Rationale:**
Why this rule exists and what problem it solves.

**Implementation:**
- How the rule is enforced
- Where in the codebase it's implemented
- Related tests

**Examples:**
- Example scenario 1
- Example scenario 2

**Status:** ❌ Not Implemented
**Priority:** High/Medium/Low
**Related Requirements:** FR-XX-XX, BR-XX-XX
```

---

## Cross-Module Rules

Rules that apply across multiple modules:

| ID | Rule Description | Affected Modules | Status |
|----|------------------|------------------|--------|
| BR-CROSS-01 | *To be defined* | - | ❌ |

---

## Change Log

| Date | Version | Changes | Author |
|------|---------|---------|--------|
| 2024-12-19 | 1.0.0 | Initial business rules document created | - |
| 2026-01-17 | 1.1.0 | Added BR-BL-15: Branch Working Day Determination and Staff Conversion rule. Added Branch Logic (BL) section to business rules table. | System |
| 2026-01-17 | 1.2.0 | Added three new business rules: BR-BL-16 (Doctor Temporary Branch Assignment), BR-BL-17 (Default Schedule Expected Revenue Requirement), and BR-BL-18 (Multiple Doctors per Branch Simultaneously). | System |

---

## Notes

- Business rules should be testable and verifiable
- Each rule should have corresponding unit tests
- Rules should be referenced in code comments using the BR-ID format
- Update status as rules are implemented
- Document any exceptions or edge cases







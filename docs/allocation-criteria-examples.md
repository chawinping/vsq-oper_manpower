# Allocation Criteria

This document defines the allocation criteria system for staff allocation across branches.

**Last Updated:** 2026-01-17 12:19:34

---

## Overview

The allocation criteria system governs how staff are allocated to branches based on three types of criteria:

1. **Doctor-Specific Criteria** - Rules that apply when specific doctors work at specific branches on specific days
2. **Clinic-Wide Criteria** - Rules that govern the entire allocation system including scoring mechanisms
3. **Branch-Specific Criteria** - Rules that prioritize certain branches that must be satisfied first

---

## Implementation Status

⚠️ **IMPORTANT:** The current implementation logic has been removed. This document describes the **intended design** for future implementation.

**Current Status:** Design phase - awaiting implementation

---

## 1. Doctor-Specific Criteria

### Purpose

Doctor-specific criteria define staff requirements that must be met when a specific doctor works at a specific branch on specific days.

### Structure

Doctor-specific criteria involve setting requirements as follows:

**If Doctor A works at Branch X on days (Monday, Wednesday, Friday, etc.), then the doctor must have at least:**
- **x** front staff
- **y** doctor assistants  
- **n** nurses
- (and other position requirements as needed)

### Example Scenarios

#### Example 1: Doctor A at Branch Central Plaza
```
Doctor: Doctor A
Branch: Central Plaza
Days: Monday, Wednesday, Friday
Requirements:
  - Front Staff: 3 minimum
  - Doctor Assistants: 2 minimum
  - Nurses: 2 minimum
```

#### Example 2: Doctor B at Branch Siam Paragon
```
Doctor: Doctor B
Branch: Siam Paragon
Days: Tuesday, Thursday, Saturday
Requirements:
  - Front Staff: 2 minimum
  - Doctor Assistants: 3 minimum
  - Nurses: 1 minimum
  - Laser Assistant: 1 minimum
```

#### Example 3: Doctor C at Multiple Branches
```
Doctor: Doctor C
Branch: Central Plaza
Days: Monday, Tuesday
Requirements:
  - Front Staff: 2 minimum
  - Doctor Assistants: 2 minimum

Doctor: Doctor C
Branch: Siam Paragon
Days: Wednesday, Thursday
Requirements:
  - Front Staff: 3 minimum
  - Doctor Assistants: 3 minimum
  - Nurses: 2 minimum
```

### Key Characteristics

- **Doctor-Branch-Day Specific:** Requirements are tied to a specific doctor working at a specific branch on specific days
- **Position-Based:** Requirements specify minimum counts for specific positions (front staff, doctor assistants, nurses, etc.)
- **Override General Rules:** Doctor-specific requirements take precedence over general branch quota configurations
- **Day-Specific:** Can specify different requirements for different days of the week

### Data Model (Proposed)

```json
{
  "doctor_id": "uuid",
  "branch_id": "uuid",
  "days": ["monday", "wednesday", "friday"],
  "requirements": [
    {
      "position_id": "uuid",
      "position_name": "Front Staff",
      "minimum_count": 3
    },
    {
      "position_id": "uuid",
      "position_name": "Doctor Assistant",
      "minimum_count": 2
    },
    {
      "position_id": "uuid",
      "position_name": "Nurse",
      "minimum_count": 2
    }
  ]
}
```

---

## 2. Clinic-Wide Criteria

### Purpose

Clinic-wide criteria define the rules that govern the entire allocation system, including the scoring system for each criterion type.

### Structure

Clinic-wide criteria establish:
- **Scoring mechanisms** for different types of criteria
- **Weighting systems** for combining multiple criteria
- **Overall allocation priorities** that apply across all branches
- **System-wide standards** and thresholds

### Example Criteria Types

#### Revenue-Based Scoring
- **Criterion:** Expected revenue per branch per day
- **Scoring Method:** Normalize revenue to 0-1 scale based on maximum expected revenue
- **Weight:** Configurable (e.g., 0.4)
- **Example:** Branch with 75,000 expected revenue out of 100,000 max = Score 0.75

#### Minimum Staff Standards
- **Criterion:** Minimum staff requirements per branch
- **Scoring Method:** Calculate fulfillment rate (actual staff / required staff)
- **Weight:** Configurable (e.g., 0.3)
- **Example:** Branch needs 10 staff, has 8 staff = Score 0.80

#### Position-Specific Requirements
- **Criterion:** Minimum staff per position
- **Scoring Method:** Average fulfillment rate across all positions
- **Weight:** Configurable (e.g., 0.3)
- **Example:** 
  - Nurse position: Need 2, Have 2 → Fulfillment 1.0
  - Doctor Assistant: Need 2, Have 1 → Fulfillment 0.5
  - Average: (1.0 + 0.5) / 2 = 0.75

### Scoring System

The clinic-wide scoring system combines multiple criteria:

```
Overall Score = Σ (Criterion Score × Criterion Weight) / Σ (Criterion Weights)
```

**Example:**
```
Revenue Score: 0.75, Weight: 0.4
Min Staff Score: 0.80, Weight: 0.3
Position Score: 0.75, Weight: 0.3

Weighted Sum = (0.75 × 0.4) + (0.80 × 0.3) + (0.75 × 0.3) = 0.30 + 0.24 + 0.225 = 0.765
Total Weight = 0.4 + 0.3 + 0.3 = 1.0
Clinic-Wide Score = 0.765 / 1.0 = 0.765
```

### Key Characteristics

- **System-Wide Application:** Rules apply uniformly across all branches
- **Configurable Weights:** Each criterion can have a configurable weight
- **Normalized Scoring:** All scores normalized to 0-1 scale for consistency
- **Combined Evaluation:** Multiple criteria combined into overall clinic-wide score

---

## 3. Branch-Specific Criteria

### Purpose

Branch-specific criteria define rules that prioritize certain branches, establishing an order in which branches must satisfy their requirements first.

### Structure

Branch-specific criteria involve:
- **Branch Priority Order:** Some branches must satisfy their requirements before others
- **Branch-Specific Rules:** Rules that apply only to specific branches
- **Priority-Based Allocation:** Staff allocation follows branch priority order

### Example Scenarios

#### Scenario 1: Priority Order
```
Priority 1: Central Plaza
  - Must satisfy all requirements first
  - Minimum staff: 15
  - Critical positions: Doctor Assistant (3), Nurse (2)

Priority 2: Siam Paragon
  - Satisfy after Priority 1 branches are met
  - Minimum staff: 12
  - Critical positions: Doctor Assistant (2), Nurse (2)

Priority 3: Emporium
  - Satisfy after Priority 1 and 2 branches are met
  - Minimum staff: 10
  - Critical positions: Doctor Assistant (2), Nurse (1)
```

#### Scenario 2: Branch-Specific Rules
```
Branch: Central Plaza
Rules:
  - Must have at least 2 nurses on weekends
  - Must have laser assistant when laser services are scheduled
  - Minimum front staff: 3 on weekdays, 4 on weekends

Branch: Siam Paragon
Rules:
  - Must have interpreter available on weekdays
  - Minimum doctor assistants: 3 when more than 2 doctors present
  - Special requirements for VIP patients
```

### Key Characteristics

- **Priority-Based:** Branches have an order that determines allocation priority
- **Branch-Specific:** Rules can be unique to individual branches
- **Sequential Satisfaction:** Higher priority branches must be satisfied before lower priority branches
- **Conditional Rules:** Rules can be conditional (e.g., "if X then Y")

### Data Model (Proposed)

```json
{
  "branch_id": "uuid",
  "priority_order": 1,
  "rules": [
    {
      "rule_type": "minimum_staff",
      "position_id": "uuid",
      "minimum_count": 3,
      "conditions": {
        "days": ["monday", "tuesday", "wednesday", "thursday", "friday"]
      }
    },
    {
      "rule_type": "conditional",
      "condition": "if_doctor_count > 2",
      "then": {
        "position_id": "uuid",
        "minimum_count": 3
      }
    }
  ]
}
```

---

## Combined Allocation Process (Proposed)

### Step 1: Evaluate Branch Priority Order
1. Identify all branches that need staff allocation
2. Sort branches by priority order (branch-specific criteria)
3. Process branches in priority order

### Step 2: Evaluate Doctor-Specific Requirements
For each branch in priority order:
1. Identify which doctors are working at the branch on the target date
2. For each doctor, check doctor-specific criteria
3. Aggregate requirements (take maximum if multiple doctors have requirements)
4. Calculate staff needed to satisfy doctor-specific requirements

### Step 3: Evaluate Clinic-Wide Criteria
1. Calculate clinic-wide scores for each branch
2. Use scores to determine overall allocation priority
3. Consider clinic-wide standards and thresholds

### Step 4: Generate Allocation Suggestions
1. Start with highest priority branch
2. Satisfy doctor-specific requirements first
3. Then satisfy branch-specific requirements
4. Use clinic-wide scores to optimize remaining allocations
5. Move to next priority branch
6. Repeat until all branches processed or staff exhausted

---

## Future Implementation Notes

### Database Schema (Proposed)

```sql
-- Doctor-Specific Criteria
CREATE TABLE doctor_allocation_criteria (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    doctor_id UUID NOT NULL REFERENCES doctors(id),
    branch_id UUID NOT NULL REFERENCES branches(id),
    days_of_week INTEGER[] NOT NULL, -- Array of day numbers (0=Sunday, 1=Monday, etc.)
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE doctor_allocation_requirements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    doctor_criteria_id UUID NOT NULL REFERENCES doctor_allocation_criteria(id),
    position_id UUID NOT NULL REFERENCES positions(id),
    minimum_count INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Branch-Specific Criteria
CREATE TABLE branch_allocation_criteria (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    branch_id UUID NOT NULL REFERENCES branches(id),
    priority_order INTEGER NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE branch_allocation_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    branch_criteria_id UUID NOT NULL REFERENCES branch_allocation_criteria(id),
    rule_type VARCHAR(50) NOT NULL,
    position_id UUID REFERENCES positions(id),
    minimum_count INTEGER,
    conditions JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Clinic-Wide Criteria (existing table can be enhanced)
CREATE TABLE clinic_wide_criteria (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    criterion_type VARCHAR(50) NOT NULL,
    weight DECIMAL(5,2) NOT NULL,
    scoring_config JSONB,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### API Endpoints (Proposed)

```
POST   /api/allocation-criteria/doctor-specific
GET    /api/allocation-criteria/doctor-specific
GET    /api/allocation-criteria/doctor-specific/:id
PUT    /api/allocation-criteria/doctor-specific/:id
DELETE /api/allocation-criteria/doctor-specific/:id

POST   /api/allocation-criteria/branch-specific
GET    /api/allocation-criteria/branch-specific
GET    /api/allocation-criteria/branch-specific/:id
PUT    /api/allocation-criteria/branch-specific/:id
DELETE /api/allocation-criteria/branch-specific/:id

GET    /api/allocation-criteria/clinic-wide
PUT    /api/allocation-criteria/clinic-wide/:id
```

---

## Related Files

- **Domain Model:** `backend/internal/domain/models/allocation_criteria.go` (to be updated)
- **Evaluation Engine:** `backend/internal/usecases/allocation/criteria_engine.go` (to be updated)
- **Handler:** `backend/internal/handlers/allocation_criteria_handler.go` (to be updated)
- **Frontend UI:** `frontend/src/app/(admin)/allocation-criteria/page.tsx` (to be updated)
- **API Client:** `frontend/src/lib/api/allocation-criteria.ts` (to be updated)

---

## Notes

- Current implementation logic has been removed
- This document describes the intended design for future implementation
- Database schema and API endpoints are proposed and subject to change
- Implementation should follow the three-pillar structure: Doctor-Specific, Clinic-Wide, and Branch-Specific

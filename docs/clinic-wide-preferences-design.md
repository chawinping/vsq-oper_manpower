---
title: Clinic-Wide Preferences Design - Staff Requirements by Criteria Ranges
description: Design alternatives for configuring clinic-wide staff requirements based on revenue ranges, case counts, and doctor counts
version: 1.0.0
lastUpdated: 2026-01-24
---

# Clinic-Wide Preferences Design

## Overview

The **Clinic-Wide Preferences** module allows administrators to configure staff requirements for each position based on different criteria ranges. This provides a centralized, clinic-wide configuration that applies to all branches.

### Requirements

The module must support configuring staff requirements (per position) based on:

1. **Skin Revenue Range** (THB) - e.g., "200K-400K THB requires 2 nurses"
2. **Laser YAG Revenue Range** (THB) - e.g., "100K-200K THB requires 1 laser technician"
3. **IV Cases Range** (Count) - e.g., "5-10 IV cases requires 1 IV nurse"
4. **Slim Pen Cases Range** (Count) - e.g., "3-5 slim pen cases requires 1 assistant"
5. **Doctor Count Range** (Count) - e.g., "3-4 doctors requires 2 doctor assistants"

Each range configuration should specify:
- Minimum and maximum values for the criteria
- Required staff count per position (minimum and preferred)
- Whether this is an active configuration

---

## Design Alternatives

### Alternative 1: Unified Criteria Table (Recommended)

**Concept:** Single table with a `criteria_type` enum field to distinguish between different criteria types.

#### Advantages
- ✅ Simple schema - one table to manage
- ✅ Easy to query all preferences together
- ✅ Consistent structure across all criteria types
- ✅ Easy to add new criteria types in the future
- ✅ Single API endpoint for all criteria types

#### Disadvantages
- ⚠️ Type-specific validation logic needed (revenue vs count)
- ⚠️ Some fields may be NULL depending on criteria type

#### Database Schema

```sql
-- Criteria type enumeration
CREATE TYPE clinic_preference_criteria_type AS ENUM (
    'skin_revenue',
    'laser_yag_revenue',
    'iv_cases',
    'slim_pen_cases',
    'doctor_count'
);

-- Main preferences table
CREATE TABLE IF NOT EXISTS clinic_wide_preferences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    criteria_type clinic_preference_criteria_type NOT NULL,
    criteria_name VARCHAR(100) NOT NULL, -- Display name, e.g., "Skin Revenue", "IV Cases"
    min_value DECIMAL(15,2) NOT NULL CHECK (min_value >= 0),
    max_value DECIMAL(15,2), -- NULL means no upper limit
    is_active BOOLEAN NOT NULL DEFAULT true,
    display_order INTEGER NOT NULL DEFAULT 0,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Ensure max_value > min_value if both are set
    CONSTRAINT check_value_range CHECK (max_value IS NULL OR max_value > min_value),
    
    -- Ensure no overlapping ranges for same criteria type (enforced at application level)
    CONSTRAINT check_unique_range UNIQUE NULLS NOT DISTINCT (criteria_type, min_value, max_value)
);

-- Position requirements for each preference
CREATE TABLE IF NOT EXISTS clinic_preference_position_requirements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    preference_id UUID NOT NULL REFERENCES clinic_wide_preferences(id) ON DELETE CASCADE,
    position_id UUID NOT NULL REFERENCES positions(id) ON DELETE CASCADE,
    minimum_staff INTEGER NOT NULL CHECK (minimum_staff >= 0),
    preferred_staff INTEGER NOT NULL CHECK (preferred_staff >= minimum_staff),
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- One requirement per position per preference
    CONSTRAINT unique_position_per_preference UNIQUE (preference_id, position_id)
);

-- Indexes
CREATE INDEX idx_clinic_preferences_type ON clinic_wide_preferences(criteria_type, is_active);
CREATE INDEX idx_clinic_preferences_range ON clinic_wide_preferences(criteria_type, min_value, max_value);
CREATE INDEX idx_preference_position_req ON clinic_preference_position_requirements(preference_id, position_id);
CREATE INDEX idx_preference_position_req_position ON clinic_preference_position_requirements(position_id);
```

#### Example Data

```sql
-- Skin Revenue Ranges
INSERT INTO clinic_wide_preferences (criteria_type, criteria_name, min_value, max_value, display_order, description) VALUES
    ('skin_revenue', 'Low Skin Revenue', 0, 200000, 1, 'Skin revenue 0-200K THB'),
    ('skin_revenue', 'Medium Skin Revenue', 200000, 400000, 2, 'Skin revenue 200K-400K THB'),
    ('skin_revenue', 'High Skin Revenue', 400000, 600000, 3, 'Skin revenue 400K-600K THB'),
    ('skin_revenue', 'Very High Skin Revenue', 600000, NULL, 4, 'Skin revenue 600K+ THB');

-- IV Cases Ranges
INSERT INTO clinic_wide_preferences (criteria_type, criteria_name, min_value, max_value, display_order, description) VALUES
    ('iv_cases', 'Low IV Cases', 0, 5, 1, '0-5 IV cases'),
    ('iv_cases', 'Medium IV Cases', 5, 10, 2, '5-10 IV cases'),
    ('iv_cases', 'High IV Cases', 10, NULL, 3, '10+ IV cases');

-- Position Requirements for Medium Skin Revenue (200K-400K)
INSERT INTO clinic_preference_position_requirements (preference_id, position_id, minimum_staff, preferred_staff) VALUES
    ((SELECT id FROM clinic_wide_preferences WHERE criteria_name = 'Medium Skin Revenue'), 
     (SELECT id FROM positions WHERE name = 'Nurse'), 2, 3),
    ((SELECT id FROM clinic_wide_preferences WHERE criteria_name = 'Medium Skin Revenue'), 
     (SELECT id FROM positions WHERE name = 'Front Staff'), 1, 2);
```

#### API Structure

```go
// Models
type ClinicWidePreference struct {
    ID           uuid.UUID                    `json:"id"`
    CriteriaType string                       `json:"criteria_type"` // "skin_revenue", "iv_cases", etc.
    CriteriaName string                       `json:"criteria_name"`
    MinValue     float64                      `json:"min_value"`
    MaxValue     *float64                     `json:"max_value,omitempty"`
    IsActive     bool                         `json:"is_active"`
    DisplayOrder int                          `json:"display_order"`
    Description  *string                      `json:"description,omitempty"`
    PositionRequirements []PreferencePositionRequirement `json:"position_requirements,omitempty"`
    CreatedAt    time.Time                    `json:"created_at"`
    UpdatedAt    time.Time                    `json:"updated_at"`
}

type PreferencePositionRequirement struct {
    ID            uuid.UUID `json:"id"`
    PreferenceID  uuid.UUID `json:"preference_id"`
    PositionID    uuid.UUID `json:"position_id"`
    Position      *Position `json:"position,omitempty"`
    MinimumStaff  int       `json:"minimum_staff"`
    PreferredStaff int      `json:"preferred_staff"`
    IsActive      bool      `json:"is_active"`
    CreatedAt     time.Time `json:"created_at"`
    UpdatedAt     time.Time `json:"updated_at"`
}

// Endpoints
GET    /api/clinic-preferences                    // List all preferences
GET    /api/clinic-preferences?criteria_type=skin_revenue  // Filter by type
GET    /api/clinic-preferences/:id                // Get single preference with requirements
POST   /api/clinic-preferences                    // Create preference
PUT    /api/clinic-preferences/:id                // Update preference
DELETE /api/clinic-preferences/:id                // Delete preference

POST   /api/clinic-preferences/:id/positions      // Add position requirement
PUT    /api/clinic-preferences/:id/positions/:positionId  // Update position requirement
DELETE /api/clinic-preferences/:id/positions/:positionId  // Remove position requirement
```

#### UI Structure

```
Clinic-Wide Preferences
├── Tabs/Filter by Criteria Type:
│   ├── Skin Revenue
│   ├── Laser YAG Revenue
│   ├── IV Cases
│   ├── Slim Pen Cases
│   └── Doctor Count
│
└── For each Criteria Type:
    ├── List of Ranges (table)
    │   ├── Range Name
    │   ├── Min Value
    │   ├── Max Value
    │   ├── Active Status
    │   └── Actions (Edit, Delete, View Requirements)
    │
    └── Range Detail View:
        ├── Range Information
        └── Position Requirements Table
            ├── Position Name
            ├── Minimum Staff
            ├── Preferred Staff
            └── Actions (Edit, Remove)
```

---

### Alternative 2: Separate Tables per Criteria Type

**Concept:** Each criteria type has its own dedicated table with type-specific fields.

#### Advantages
- ✅ Type-specific fields and validation
- ✅ Clear separation of concerns
- ✅ No NULL fields for type-specific data
- ✅ Easier to understand schema

#### Disadvantages
- ⚠️ More tables to manage
- ⚠️ Code duplication across similar tables
- ⚠️ Multiple API endpoints needed
- ⚠️ Harder to query across all criteria types

#### Database Schema

```sql
-- Base position requirements table (shared)
CREATE TABLE IF NOT EXISTS clinic_preference_position_requirements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    position_id UUID NOT NULL REFERENCES positions(id) ON DELETE CASCADE,
    minimum_staff INTEGER NOT NULL CHECK (minimum_staff >= 0),
    preferred_staff INTEGER NOT NULL CHECK (preferred_staff >= minimum_staff),
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Skin Revenue Preferences
CREATE TABLE IF NOT EXISTS skin_revenue_preferences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    preference_name VARCHAR(100) NOT NULL,
    min_revenue DECIMAL(15,2) NOT NULL CHECK (min_revenue >= 0),
    max_revenue DECIMAL(15,2), -- NULL means no upper limit
    is_active BOOLEAN NOT NULL DEFAULT true,
    display_order INTEGER NOT NULL DEFAULT 0,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT check_skin_revenue_range CHECK (max_revenue IS NULL OR max_revenue > min_revenue)
);

-- Link skin revenue preferences to position requirements
CREATE TABLE IF NOT EXISTS skin_revenue_position_requirements (
    preference_id UUID NOT NULL REFERENCES skin_revenue_preferences(id) ON DELETE CASCADE,
    requirement_id UUID NOT NULL REFERENCES clinic_preference_position_requirements(id) ON DELETE CASCADE,
    PRIMARY KEY (preference_id, requirement_id)
);

-- Laser YAG Revenue Preferences
CREATE TABLE IF NOT EXISTS laser_yag_revenue_preferences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    preference_name VARCHAR(100) NOT NULL,
    min_revenue DECIMAL(15,2) NOT NULL CHECK (min_revenue >= 0),
    max_revenue DECIMAL(15,2),
    is_active BOOLEAN NOT NULL DEFAULT true,
    display_order INTEGER NOT NULL DEFAULT 0,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT check_laser_yag_revenue_range CHECK (max_revenue IS NULL OR max_revenue > min_revenue)
);

CREATE TABLE IF NOT EXISTS laser_yag_revenue_position_requirements (
    preference_id UUID NOT NULL REFERENCES laser_yag_revenue_preferences(id) ON DELETE CASCADE,
    requirement_id UUID NOT NULL REFERENCES clinic_preference_position_requirements(id) ON DELETE CASCADE,
    PRIMARY KEY (preference_id, requirement_id)
);

-- IV Cases Preferences
CREATE TABLE IF NOT EXISTS iv_cases_preferences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    preference_name VARCHAR(100) NOT NULL,
    min_cases INTEGER NOT NULL CHECK (min_cases >= 0),
    max_cases INTEGER, -- NULL means no upper limit
    is_active BOOLEAN NOT NULL DEFAULT true,
    display_order INTEGER NOT NULL DEFAULT 0,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT check_iv_cases_range CHECK (max_cases IS NULL OR max_cases > min_cases)
);

CREATE TABLE IF NOT EXISTS iv_cases_position_requirements (
    preference_id UUID NOT NULL REFERENCES iv_cases_preferences(id) ON DELETE CASCADE,
    requirement_id UUID NOT NULL REFERENCES clinic_preference_position_requirements(id) ON DELETE CASCADE,
    PRIMARY KEY (preference_id, requirement_id)
);

-- Slim Pen Cases Preferences
CREATE TABLE IF NOT EXISTS slim_pen_cases_preferences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    preference_name VARCHAR(100) NOT NULL,
    min_cases INTEGER NOT NULL CHECK (min_cases >= 0),
    max_cases INTEGER,
    is_active BOOLEAN NOT NULL DEFAULT true,
    display_order INTEGER NOT NULL DEFAULT 0,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT check_slim_pen_cases_range CHECK (max_cases IS NULL OR max_cases > min_cases)
);

CREATE TABLE IF NOT EXISTS slim_pen_cases_position_requirements (
    preference_id UUID NOT NULL REFERENCES slim_pen_cases_preferences(id) ON DELETE CASCADE,
    requirement_id UUID NOT NULL REFERENCES clinic_preference_position_requirements(id) ON DELETE CASCADE,
    PRIMARY KEY (preference_id, requirement_id)
);

-- Doctor Count Preferences
CREATE TABLE IF NOT EXISTS doctor_count_preferences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    preference_name VARCHAR(100) NOT NULL,
    min_doctors INTEGER NOT NULL CHECK (min_doctors >= 0),
    max_doctors INTEGER,
    is_active BOOLEAN NOT NULL DEFAULT true,
    display_order INTEGER NOT NULL DEFAULT 0,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT check_doctor_count_range CHECK (max_doctors IS NULL OR max_doctors > min_doctors)
);

CREATE TABLE IF NOT EXISTS doctor_count_position_requirements (
    preference_id UUID NOT NULL REFERENCES doctor_count_preferences(id) ON DELETE CASCADE,
    requirement_id UUID NOT NULL REFERENCES clinic_preference_position_requirements(id) ON DELETE CASCADE,
    PRIMARY KEY (preference_id, requirement_id)
);
```

#### API Structure

```go
// Separate endpoints for each type
GET    /api/clinic-preferences/skin-revenue
POST   /api/clinic-preferences/skin-revenue
PUT    /api/clinic-preferences/skin-revenue/:id
DELETE /api/clinic-preferences/skin-revenue/:id

GET    /api/clinic-preferences/laser-yag-revenue
POST   /api/clinic-preferences/laser-yag-revenue
...

GET    /api/clinic-preferences/iv-cases
POST   /api/clinic-preferences/iv-cases
...

GET    /api/clinic-preferences/slim-pen-cases
POST   /api/clinic-preferences/slim-pen-cases
...

GET    /api/clinic-preferences/doctor-count
POST   /api/clinic-preferences/doctor-count
...
```

---

### Alternative 3: Hybrid Approach - Base Table + Criteria-Specific Extensions

**Concept:** Base table for common fields, separate tables for criteria-specific data.

#### Advantages
- ✅ Balance between unified and separated
- ✅ Type-specific fields where needed
- ✅ Can query all preferences together
- ✅ Extensible for future criteria types

#### Disadvantages
- ⚠️ More complex joins required
- ⚠️ Slightly more complex schema

#### Database Schema

```sql
-- Base preferences table
CREATE TABLE IF NOT EXISTS clinic_wide_preferences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    criteria_type VARCHAR(50) NOT NULL, -- 'skin_revenue', 'iv_cases', etc.
    preference_name VARCHAR(100) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    display_order INTEGER NOT NULL DEFAULT 0,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Revenue-based preferences (for skin_revenue and laser_yag_revenue)
CREATE TABLE IF NOT EXISTS revenue_preference_ranges (
    preference_id UUID PRIMARY KEY REFERENCES clinic_wide_preferences(id) ON DELETE CASCADE,
    min_revenue DECIMAL(15,2) NOT NULL CHECK (min_revenue >= 0),
    max_revenue DECIMAL(15,2),
    CONSTRAINT check_revenue_range CHECK (max_revenue IS NULL OR max_revenue > min_revenue)
);

-- Count-based preferences (for iv_cases, slim_pen_cases, doctor_count)
CREATE TABLE IF NOT EXISTS count_preference_ranges (
    preference_id UUID PRIMARY KEY REFERENCES clinic_wide_preferences(id) ON DELETE CASCADE,
    min_count INTEGER NOT NULL CHECK (min_count >= 0),
    max_count INTEGER,
    CONSTRAINT check_count_range CHECK (max_count IS NULL OR max_count > min_count)
);

-- Position requirements (same as Alternative 1)
CREATE TABLE IF NOT EXISTS clinic_preference_position_requirements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    preference_id UUID NOT NULL REFERENCES clinic_wide_preferences(id) ON DELETE CASCADE,
    position_id UUID NOT NULL REFERENCES positions(id) ON DELETE CASCADE,
    minimum_staff INTEGER NOT NULL CHECK (minimum_staff >= 0),
    preferred_staff INTEGER NOT NULL CHECK (preferred_staff >= minimum_staff),
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_position_per_preference UNIQUE (preference_id, position_id)
);
```

---

## Comparison Matrix

| Aspect | Alternative 1: Unified | Alternative 2: Separate | Alternative 3: Hybrid |
|--------|------------------------|-------------------------|----------------------|
| **Schema Complexity** | Low | Medium | Medium |
| **Code Duplication** | None | High | Low |
| **Query Simplicity** | High | Low | Medium |
| **Type Safety** | Medium | High | High |
| **Extensibility** | High | Medium | High |
| **API Endpoints** | Single set | Multiple sets | Single set |
| **Maintenance** | Easy | Moderate | Moderate |
| **Performance** | Good | Good | Good (with indexes) |

---

## Recommendation: Alternative 1 (Unified Criteria Table)

### Rationale

1. **Simplicity**: Single table structure is easier to understand and maintain
2. **Consistency**: All criteria types follow the same pattern
3. **Flexibility**: Easy to add new criteria types without schema changes
4. **Query Efficiency**: Can query all preferences with a single query
5. **API Simplicity**: Single set of endpoints with filtering by criteria type
6. **UI Consistency**: Same UI pattern for all criteria types

### Implementation Considerations

1. **Validation**: Implement application-level validation to ensure:
   - Revenue types use DECIMAL values
   - Count types use INTEGER values (can store as DECIMAL and cast)
   - No overlapping ranges for the same criteria type

2. **Display**: In the UI, format values appropriately:
   - Revenue types: Show as "XXX,XXX THB"
   - Count types: Show as "X cases" or "X doctors"

3. **Range Matching**: When calculating staff requirements:
   - Find matching preference where `value >= min_value AND (max_value IS NULL OR value <= max_value)`
   - If multiple matches, use the most specific range (smallest range)

4. **Integration with Allocation Engine**: 
   - Clinic-wide preferences serve as base requirements
   - Specific preferences (from Staff Requirement Scenarios) can override these
   - Priority: Specific Preferences > Clinic-Wide Preferences > Default

---

## Integration with Existing System

### Relationship to Staff Requirement Scenarios

- **Clinic-Wide Preferences**: Base/default requirements for all branches
- **Staff Requirement Scenarios**: Specific overrides for particular scenarios
- **Priority**: Scenarios override clinic-wide preferences when they match

### Calculation Flow

```
1. Get branch data (revenue, cases, doctor count)
2. Check Staff Requirement Scenarios (specific preferences)
   - If match found → use scenario requirements
3. If no scenario match → Check Clinic-Wide Preferences
   - Match by criteria type and value range
   - Use matching preference's position requirements
4. If no preference match → Use default/base requirements
```

---

## Migration Path

1. **Phase 1**: Create new tables and API endpoints
2. **Phase 2**: Build UI for managing clinic-wide preferences
3. **Phase 3**: Integrate with allocation engine
4. **Phase 4**: Migrate existing revenue level tier logic (if applicable)
5. **Phase 5**: Deprecate old revenue level tiers (if replaced)

---

## Future Enhancements

1. **Branch-Specific Overrides**: Allow branches to override clinic-wide preferences
2. **Time-Based Preferences**: Different preferences for different times/seasons
3. **Combined Criteria**: Support multiple criteria matching (e.g., skin revenue AND doctor count)
4. **Historical Tracking**: Track changes to preferences over time
5. **A/B Testing**: Test different preference configurations

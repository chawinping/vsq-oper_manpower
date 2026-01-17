---
title: Manpower Type and Branch Constraints Implementation
description: Implementation of manpower type classification and branch constraint configurations
version: 1.0.0
lastUpdated: 2026-01-13 21:00:00
---

# Manpower Type and Branch Constraints Implementation

## Overview

This document describes the implementation of:
1. **Manpower Type Classification** - Each position is classified into one of four manpower types
2. **Branch Constraints Configuration** - Minimum staff requirements per day for different staff categories

## Implementation Status

### ✅ Backend - Completed
- [x] Added `manpower_type` field to positions table
- [x] Created `branch_constraints` table
- [x] Updated Position model with ManpowerType
- [x] Created BranchConstraints model
- [x] Updated Position repository to handle manpower_type
- [x] Created BranchConstraintsRepository
- [x] Updated Position handler to include manpower_type
- [x] Added branch constraints handlers (GetConstraints, UpdateConstraints)
- [x] Added API routes for constraints endpoints
- [x] Database migrations for manpower_type and constraints

### ⚠️ Frontend - In Progress
- [ ] Update Position interface to include manpower_type
- [ ] Update positions admin page to allow editing manpower_type
- [ ] Update branch configuration UI to show constraint settings
- [ ] Add constraint configuration form (min_front_staff, min_managers, min_total_staff)

## Manpower Types

Each position must be classified into one of four manpower types:

1. **พนักงานฟร้อนท์** (Front/Counter staff) - Reception, front desk positions
2. **ผู้ช่วยแพทย์** (Doctor Assistant) - Medical support staff (nurses, doctor assistants, physiotherapists)
3. **อื่นๆ** (Others) - Management and coordination positions
4. **ทำความสะอาด** (Cleaning/Housekeeping) - Cleaning staff

## Branch Constraints

Each branch can configure minimum staff requirements per day of the week:

1. **Minimum Front Staff** (`min_front_staff`) - Combined minimum number of front/counter staff
2. **Minimum Managers** (`min_managers`) - Combined minimum number of Branch Managers and Assistant Branch Managers
3. **Minimum Total Staff** (`min_total_staff`) - Minimum total staff in the branch

## Database Schema

### Positions Table
```sql
ALTER TABLE positions 
ADD COLUMN manpower_type VARCHAR(50) NOT NULL DEFAULT 'อื่นๆ' 
CHECK (manpower_type IN ('พนักงานฟร้อนท์', 'ผู้ช่วยแพทย์', 'อื่นๆ', 'ทำความสะอาด'));
```

### Branch Constraints Table
```sql
CREATE TABLE branch_constraints (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    branch_id UUID NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
    day_of_week INTEGER NOT NULL CHECK (day_of_week >= 0 AND day_of_week <= 6),
    min_front_staff INTEGER NOT NULL DEFAULT 0 CHECK (min_front_staff >= 0),
    min_managers INTEGER NOT NULL DEFAULT 0 CHECK (min_managers >= 0),
    min_total_staff INTEGER NOT NULL DEFAULT 0 CHECK (min_total_staff >= 0),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(branch_id, day_of_week)
);
```

## API Endpoints

### Get Constraints
```
GET /api/branches/:id/config/constraints
```

### Update Constraints
```
PUT /api/branches/:id/config/constraints
Body: {
  "constraints": [
    {
      "day_of_week": 0,
      "min_front_staff": 2,
      "min_managers": 1,
      "min_total_staff": 5
    },
    ...
  ]
}
```

## Files Modified

### Backend
- `backend/internal/domain/models/staff.go` - Added ManpowerType
- `backend/internal/domain/models/branch_constraints.go` - New file
- `backend/internal/repositories/postgres/migrations.go` - Added migrations
- `backend/internal/repositories/postgres/repositories.go` - Updated Position repository
- `backend/internal/repositories/postgres/branch_constraints_repo.go` - New file
- `backend/internal/handlers/position_handler.go` - Updated to handle manpower_type
- `backend/internal/handlers/branch_config_handler.go` - Added constraints handlers
- `backend/cmd/server/main.go` - Added routes

### Frontend (To Be Updated)
- `frontend/src/lib/api/position.ts` - Add manpower_type
- `frontend/src/lib/api/branch-config.ts` - Add constraints API
- `frontend/src/app/(admin)/positions/page.tsx` - Add manpower_type editing
- `frontend/src/components/branch/BranchPositionQuotaConfig.tsx` - Add constraints UI

## Next Steps

1. Update frontend Position interface
2. Add manpower_type dropdown in positions admin page
3. Create constraints configuration UI component
4. Integrate constraints UI into branch configuration page
5. Test end-to-end functionality

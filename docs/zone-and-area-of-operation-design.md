---
title: Zone and Area of Operation Design
description: Design document for Zone-based Area of Operation system
version: 1.0.0
lastUpdated: 2026-01-17 16:00:00
---

# Zone and Area of Operation Design

## Overview

This document describes the design for implementing a Zone-based Area of Operation system where:
- **Area of Operation** is mainly a **Zone** in Bangkok
- A **Zone** consists of multiple branches
- Additionally, apart from zones, individual branches can be added to an Area of Operation
- There should be a menu to configure zones and their associated branches

## Business Logic

### Zone
- A Zone is a geographical grouping in Bangkok
- A Zone consists of multiple branches
- Zones can be created, edited, and deleted
- Zones have: ID, Name, Code, Description, IsActive status

### Area of Operation
- An Area of Operation can include:
  1. **Zones** - One or more zones (which include all their branches)
  2. **Individual Branches** - Branches added directly (apart from zones)
- When a rotation staff is assigned to an Area of Operation, they can work at:
  - All branches in the zones assigned to that area
  - All individual branches assigned to that area

### Relationships

```
Area of Operation
├── Zones (via area_of_operation_zones)
│   └── Branches (via zone_branches)
└── Individual Branches (via area_of_operation_branches)
```

## Database Schema

### Tables

1. **zones** - Zone master data
   - id (UUID, PK)
   - name (VARCHAR)
   - code (VARCHAR, UNIQUE)
   - description (TEXT)
   - is_active (BOOLEAN)
   - created_at, updated_at

2. **zone_branches** - Zone-Branch relationships
   - id (UUID, PK)
   - zone_id (UUID, FK → zones)
   - branch_id (UUID, FK → branches)
   - created_at
   - UNIQUE(zone_id, branch_id)

3. **area_of_operation_zones** - Area-Zone relationships
   - id (UUID, PK)
   - area_of_operation_id (UUID, FK → areas_of_operation)
   - zone_id (UUID, FK → zones)
   - created_at
   - UNIQUE(area_of_operation_id, zone_id)

4. **area_of_operation_branches** - Area-Branch relationships (individual branches)
   - id (UUID, PK)
   - area_of_operation_id (UUID, FK → areas_of_operation)
   - branch_id (UUID, FK → branches)
   - created_at
   - UNIQUE(area_of_operation_id, branch_id)

## API Endpoints

### Zone Management

- `GET /zones` - List all zones
- `GET /zones/:id` - Get zone by ID
- `POST /zones` - Create zone
- `PUT /zones/:id` - Update zone
- `DELETE /zones/:id` - Delete zone
- `GET /zones/:id/branches` - Get branches in zone
- `POST /zones/:id/branches` - Add branch to zone (bulk update)
- `DELETE /zones/:id/branches/:branchId` - Remove branch from zone

### Area of Operation Management (Enhanced)

- `GET /areas-of-operation/:id` - Get area with zones and branches
- `POST /areas-of-operation/:id/zones` - Add zone to area
- `DELETE /areas-of-operation/:id/zones/:zoneId` - Remove zone from area
- `POST /areas-of-operation/:id/branches` - Add individual branch to area
- `DELETE /areas-of-operation/:id/branches/:branchId` - Remove branch from area
- `GET /areas-of-operation/:id/branches` - Get all branches (from zones + individual)

## Frontend UI

### Zone Configuration Menu

**Location:** Admin menu → "Zone Configuration"

**Features:**
1. List all zones with their branches
2. Create/Edit/Delete zones
3. Manage branches for each zone:
   - Add branches to zone
   - Remove branches from zone
   - Bulk update branches
4. View zone details

### Area of Operation Configuration (Enhanced)

**Location:** Admin menu → "Areas of Operation"

**Features:**
1. List all areas of operation
2. Create/Edit/Delete areas
3. Configure zones for area:
   - Add zones to area
   - Remove zones from area
   - View branches included via zones
4. Configure individual branches for area:
   - Add individual branches
   - Remove individual branches
5. View all branches (zones + individual) for an area

## Implementation Plan

### Phase 1: Backend - Zone Repository and Handler
- [x] Create Zone model
- [x] Create database tables (zones, zone_branches, area_of_operation_zones, area_of_operation_branches)
- [ ] Implement Zone repository
- [ ] Implement Zone handler
- [ ] Update Area of Operation repository with zone/branch methods
- [ ] Update Area of Operation handler

### Phase 2: Frontend - Zone Configuration
- [ ] Create Zone API client
- [ ] Create Zone Configuration page
- [ ] Implement zone CRUD UI
- [ ] Implement branch management UI for zones

### Phase 3: Frontend - Area of Operation Enhancement
- [ ] Update Area of Operation API client
- [ ] Enhance Area of Operation configuration page
- [ ] Add zone selection UI
- [ ] Add individual branch selection UI
- [ ] Display all branches (from zones + individual)

### Phase 4: Integration
- [ ] Update rotation staff assignment logic to consider zones
- [ ] Update filtering logic
- [ ] Testing

## Access Control

- **Admin**: Full access to zones and areas of operation
- **Area Manager**: Can view zones and areas (read-only or limited edit)
- **District Manager**: Read-only access
- **Branch Manager**: No access

## Notes

- A branch can belong to multiple zones (if needed)
- A zone can belong to multiple areas of operation
- An area of operation can have both zones and individual branches
- When calculating branches for an area: Union of (all branches from zones) + (individual branches)

---

**Last Updated:** 2026-01-17 16:00:00

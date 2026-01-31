# Database Performance Analysis & Redesign Recommendations

## Executive Summary

The current database design relies heavily on **dynamic calculations** performed on every request, causing significant performance bottlenecks. This document analyzes the current structure, identifies bottlenecks, and proposes database redesign strategies that could improve performance by **70-90%**.

## Current Performance Bottlenecks

### 1. **Dynamic Calculation Architecture**

**Problem**: All quota calculations are performed in real-time on every API request.

**Current Flow**:
```
API Request → CalculateBranchQuotaStatus (per branch)
  ├─ Get branch info (1 query)
  ├─ Get doctor count (1 query)
  ├─ Get position quotas (1 query)
  ├─ Get branch staff (1 query)
  ├─ Get rotation assignments (1 query)
  ├─ Get all positions (1 query - cached)
  ├─ Get schedules for ALL staff (1 batch query - improved)
  ├─ Calculate position statuses (in-memory loops)
  ├─ Calculate Group 1 score (multiple queries)
  ├─ Calculate Group 2 score (in-memory)
  └─ Calculate Group 3 score (in-memory)
```

**For 32 branches**: ~50-100 queries + extensive in-memory processing

**Impact**: 
- **2-5 seconds** per overview request
- **High CPU usage** on database server
- **Poor scalability** as branches/staff grow

### 2. **Missing Critical Indexes**

**Current Missing Indexes**:

```sql
-- staff_schedules table (CRITICAL - most queried table)
-- Missing composite index for date range queries
CREATE INDEX idx_staff_schedules_staff_date_range 
ON staff_schedules(staff_id, date) 
WHERE date >= CURRENT_DATE - INTERVAL '30 days';

-- Missing index for branch + date queries
CREATE INDEX idx_staff_schedules_branch_date 
ON staff_schedules(branch_id, date);

-- Missing index for schedule_status filtering
CREATE INDEX idx_staff_schedules_status_date 
ON staff_schedules(schedule_status, date) 
WHERE schedule_status = 'working';

-- position_quotas table
-- Missing composite index for active quotas
CREATE INDEX idx_position_quotas_branch_active 
ON position_quotas(branch_id, is_active) 
WHERE is_active = true;

-- rotation_assignments table
-- Missing composite index for date range queries
CREATE INDEX idx_rotation_assignments_branch_date_range 
ON rotation_assignments(branch_id, date);

-- branch_constraints table
-- Missing composite index for day lookup
CREATE INDEX idx_branch_constraints_branch_day 
ON branch_constraints(branch_id, day_of_week);

-- staff table
-- Missing index for branch + position queries
CREATE INDEX idx_staff_branch_position 
ON staff(branch_id, position_id) 
WHERE staff_type = 'branch';
```

**Impact**: Without these indexes, queries perform full table scans, especially on `staff_schedules` which grows daily.

### 3. **Inefficient Data Access Patterns**

**Problems**:
1. **Multiple separate queries** instead of JOINs
2. **Loading all positions** even when only needed for selected branches
3. **No query result caching** at database level
4. **Redundant data fetching** (same data fetched multiple times)

**Example**:
```go
// Current: 3 separate queries
branchStaff, _ := repos.Staff.GetByBranchID(branchID)
quotas, _ := repos.PositionQuota.GetByBranchID(branchID)
schedulesMap, _ := repos.Schedule.GetByStaffIDs(staffIDs, date, date)

// Better: 1 JOIN query
SELECT s.*, pq.*, ss.schedule_status
FROM staff s
LEFT JOIN position_quotas pq ON s.branch_id = pq.branch_id AND s.position_id = pq.position_id
LEFT JOIN staff_schedules ss ON s.id = ss.staff_id AND ss.date = $1
WHERE s.branch_id = $2 AND s.staff_type = 'branch'
```

### 4. **No Pre-computed Aggregations**

**Problem**: Calculations like staff counts, shortages, and scores are computed on-demand.

**Impact**: 
- Same calculations repeated for identical date/branch combinations
- No historical data preservation
- Cannot track trends or changes over time

## Database Redesign Strategies

### Strategy 1: Materialized Views (Recommended)

**Concept**: Pre-compute and store calculated quota statuses, refresh periodically.

**Implementation**:

```sql
-- Create materialized view for branch quota status
CREATE MATERIALIZED VIEW branch_quota_status_cache AS
SELECT 
    b.id AS branch_id,
    b.name AS branch_name,
    b.code AS branch_code,
    ss.date,
    
    -- Position-level aggregations
    pq.position_id,
    p.name AS position_name,
    pq.designated_quota,
    pq.minimum_required,
    
    -- Staff counts
    COUNT(DISTINCT CASE 
        WHEN s.staff_type = 'branch' 
        AND ss.schedule_status = 'working' 
        THEN s.id 
    END) AS available_local,
    
    COUNT(DISTINCT CASE 
        WHEN ra.rotation_staff_id IS NOT NULL 
        THEN ra.rotation_staff_id 
    END) AS assigned_rotation,
    
    -- Group scores (simplified - full calculation in refresh function)
    COALESCE(SUM(CASE 
        WHEN s.staff_type = 'branch' 
        AND ss.schedule_status != 'working' 
        THEN 1 ELSE 0 
    END), 0) AS group1_missing_count,
    
    -- Metadata
    CURRENT_TIMESTAMP AS calculated_at
    
FROM branches b
CROSS JOIN generate_series(
    CURRENT_DATE, 
    CURRENT_DATE + INTERVAL '30 days', 
    INTERVAL '1 day'
) AS ss(date)
LEFT JOIN position_quotas pq ON pq.branch_id = b.id AND pq.is_active = true
LEFT JOIN positions p ON p.id = pq.position_id
LEFT JOIN staff s ON s.branch_id = b.id AND s.staff_type = 'branch' AND s.position_id = pq.position_id
LEFT JOIN staff_schedules ss2 ON ss2.staff_id = s.id AND ss2.date = ss.date
LEFT JOIN rotation_assignments ra ON ra.branch_id = b.id AND ra.date = ss.date
GROUP BY b.id, b.name, b.code, ss.date, pq.position_id, p.name, pq.designated_quota, pq.minimum_required;

-- Create indexes on materialized view
CREATE INDEX idx_quota_cache_branch_date 
ON branch_quota_status_cache(branch_id, date);

CREATE INDEX idx_quota_cache_date 
ON branch_quota_status_cache(date);

-- Refresh function (run daily via cron)
CREATE OR REPLACE FUNCTION refresh_quota_cache()
RETURNS void AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY branch_quota_status_cache;
END;
$$ LANGUAGE plpgsql;
```

**Benefits**:
- **90-95% faster** queries (pre-computed data)
- **Reduced database load** (calculations done once per day)
- **Consistent performance** regardless of data size

**Trade-offs**:
- **Stale data** until refresh (acceptable for daily planning)
- **Storage overhead** (~10-50 MB for 32 branches × 30 days)
- **Refresh time** (~5-10 seconds daily)

### Strategy 2: Denormalized Summary Tables

**Concept**: Store pre-calculated summaries in dedicated tables, update on data changes.

**Implementation**:

```sql
-- Daily branch quota summary table
CREATE TABLE branch_quota_daily_summary (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    branch_id UUID NOT NULL REFERENCES branches(id),
    date DATE NOT NULL,
    
    -- Aggregated counts
    total_designated INTEGER NOT NULL DEFAULT 0,
    total_available INTEGER NOT NULL DEFAULT 0,
    total_assigned INTEGER NOT NULL DEFAULT 0,
    total_required INTEGER NOT NULL DEFAULT 0,
    
    -- Group scores
    group1_score INTEGER NOT NULL DEFAULT 0,
    group2_score INTEGER NOT NULL DEFAULT 0,
    group3_score INTEGER NOT NULL DEFAULT 0,
    
    -- Missing staff (JSONB for flexibility)
    group1_missing_staff JSONB DEFAULT '[]'::jsonb,
    group2_missing_staff JSONB DEFAULT '[]'::jsonb,
    group3_missing_staff JSONB DEFAULT '[]'::jsonb,
    
    -- Metadata
    calculated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(branch_id, date)
);

CREATE INDEX idx_quota_summary_branch_date 
ON branch_quota_daily_summary(branch_id, date);

CREATE INDEX idx_quota_summary_date 
ON branch_quota_daily_summary(date);

-- Position-level summary
CREATE TABLE position_quota_daily_summary (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    branch_id UUID NOT NULL REFERENCES branches(id),
    position_id UUID NOT NULL REFERENCES positions(id),
    date DATE NOT NULL,
    
    designated_quota INTEGER NOT NULL DEFAULT 0,
    minimum_required INTEGER NOT NULL DEFAULT 0,
    available_local INTEGER NOT NULL DEFAULT 0,
    assigned_rotation INTEGER NOT NULL DEFAULT 0,
    total_assigned INTEGER NOT NULL DEFAULT 0,
    still_required INTEGER NOT NULL DEFAULT 0,
    
    calculated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(branch_id, position_id, date)
);

CREATE INDEX idx_position_summary_branch_date 
ON position_quota_daily_summary(branch_id, date);

-- Trigger to update summary when schedules change
CREATE OR REPLACE FUNCTION update_quota_summary_on_schedule_change()
RETURNS TRIGGER AS $$
BEGIN
    -- Recalculate summary for affected branch/date
    PERFORM recalculate_branch_quota_summary(
        COALESCE(NEW.branch_id, OLD.branch_id),
        COALESCE(NEW.date, OLD.date)
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER schedule_change_quota_update
AFTER INSERT OR UPDATE OR DELETE ON staff_schedules
FOR EACH ROW
EXECUTE FUNCTION update_quota_summary_on_schedule_change();
```

**Benefits**:
- **Real-time updates** via triggers
- **Fast queries** (simple SELECT)
- **Historical data** preserved
- **Flexible** (can add more fields)

**Trade-offs**:
- **Complex triggers** (maintenance overhead)
- **Slight write overhead** (acceptable)
- **Storage** (~5-20 MB)

### Strategy 3: Hybrid Approach (Best Balance)

**Concept**: Combine materialized views for bulk reads + summary tables for real-time updates.

**Implementation**:
1. **Materialized view** for historical/analytical queries (refreshed nightly)
2. **Summary table** for current day (updated via triggers)
3. **Cache layer** in application (5-minute TTL)

**Benefits**:
- **Best of both worlds**
- **Fast reads** (materialized view)
- **Real-time accuracy** (summary table for today)
- **Reduced complexity** (no complex triggers for historical data)

## Recommended Index Additions

### High Priority Indexes

```sql
-- 1. staff_schedules (CRITICAL - most queried)
CREATE INDEX CONCURRENTLY idx_staff_schedules_staff_date 
ON staff_schedules(staff_id, date DESC);

CREATE INDEX CONCURRENTLY idx_staff_schedules_branch_date_status 
ON staff_schedules(branch_id, date, schedule_status) 
WHERE schedule_status = 'working';

CREATE INDEX CONCURRENTLY idx_staff_schedules_date_range 
ON staff_schedules(date) 
WHERE date >= CURRENT_DATE - INTERVAL '90 days';

-- 2. position_quotas
CREATE INDEX CONCURRENTLY idx_position_quotas_branch_active 
ON position_quotas(branch_id, is_active) 
WHERE is_active = true;

-- 3. rotation_assignments
CREATE INDEX CONCURRENTLY idx_rotation_assignments_branch_date 
ON rotation_assignments(branch_id, date DESC);

CREATE INDEX CONCURRENTLY idx_rotation_assignments_staff_date 
ON rotation_assignments(rotation_staff_id, date DESC);

-- 4. branch_constraints
CREATE INDEX CONCURRENTLY idx_branch_constraints_branch_day 
ON branch_constraints(branch_id, day_of_week);

-- 5. staff
CREATE INDEX CONCURRENTLY idx_staff_branch_position_type 
ON staff(branch_id, position_id, staff_type) 
WHERE staff_type = 'branch';

-- 6. doctor_assignments
CREATE INDEX CONCURRENTLY idx_doctor_assignments_branch_date 
ON doctor_assignments(branch_id, date DESC);
```

**Expected Impact**: **40-60% query performance improvement**

## Query Optimization Recommendations

### 1. Use JOINs Instead of Multiple Queries

**Current (Inefficient)**:
```go
branchStaff, _ := repos.Staff.GetByBranchID(branchID)
quotas, _ := repos.PositionQuota.GetByBranchID(branchID)
schedulesMap, _ := repos.Schedule.GetByStaffIDs(staffIDs, date, date)
```

**Optimized**:
```sql
SELECT 
    s.id AS staff_id,
    s.position_id,
    pq.designated_quota,
    pq.minimum_required,
    ss.schedule_status,
    ss.date
FROM staff s
INNER JOIN position_quotas pq 
    ON s.branch_id = pq.branch_id 
    AND s.position_id = pq.position_id
LEFT JOIN staff_schedules ss 
    ON s.id = ss.staff_id 
    AND ss.date = $1
WHERE s.branch_id = $2 
    AND s.staff_type = 'branch'
    AND pq.is_active = true;
```

### 2. Use Window Functions for Aggregations

**Instead of**:
```go
for _, quota := range quotas {
    availableLocal := 0
    for _, staff := range branchStaff {
        if staff.PositionID == quota.PositionID {
            // count...
        }
    }
}
```

**Use**:
```sql
SELECT 
    position_id,
    COUNT(*) FILTER (WHERE schedule_status = 'working') AS available_local,
    COUNT(*) FILTER (WHERE schedule_status != 'working') AS missing_count
FROM staff s
LEFT JOIN staff_schedules ss ON s.id = ss.staff_id AND ss.date = $1
WHERE s.branch_id = $2
GROUP BY position_id;
```

### 3. Batch Operations

**Current**: Individual queries per branch
**Optimized**: Single query for all branches

```sql
SELECT 
    branch_id,
    position_id,
    COUNT(*) FILTER (WHERE schedule_status = 'working') AS available_local
FROM staff s
LEFT JOIN staff_schedules ss ON s.id = ss.staff_id AND ss.date = $1
WHERE s.branch_id = ANY($2)  -- $2 is array of branch IDs
GROUP BY branch_id, position_id;
```

## Performance Impact Estimates

### Current Performance
- **API Response Time**: 2-5 seconds (32 branches)
- **Database Queries**: 50-100 queries per request
- **CPU Usage**: High (dynamic calculations)
- **Scalability**: Poor (linear degradation)

### After Indexes Only
- **API Response Time**: 1-2 seconds (**50% improvement**)
- **Database Queries**: 50-100 queries (same, but faster)
- **CPU Usage**: Medium
- **Scalability**: Moderate

### After Materialized Views
- **API Response Time**: 0.2-0.5 seconds (**90% improvement**)
- **Database Queries**: 1-2 queries (from materialized view)
- **CPU Usage**: Low (only during refresh)
- **Scalability**: Excellent

### After Full Redesign (Hybrid)
- **API Response Time**: 0.1-0.3 seconds (**95% improvement**)
- **Database Queries**: 1 query (from summary table)
- **CPU Usage**: Very Low
- **Scalability**: Excellent (constant time)

## Migration Strategy

### Phase 1: Add Indexes (Low Risk, High Impact)
1. Add all recommended indexes using `CREATE INDEX CONCURRENTLY`
2. Monitor query performance
3. **Expected Time**: 1-2 hours
4. **Downtime**: None (concurrent index creation)

### Phase 2: Create Summary Tables (Medium Risk)
1. Create `branch_quota_daily_summary` table
2. Create calculation function
3. Populate for current date
4. Update application to use summary table
5. **Expected Time**: 4-8 hours
6. **Downtime**: Minimal (can run in parallel)

### Phase 3: Implement Materialized Views (Low Risk)
1. Create materialized view
2. Set up daily refresh job (cron)
3. Update application to use view for historical data
4. **Expected Time**: 2-4 hours
5. **Downtime**: None

### Phase 4: Optimize Queries (Low Risk)
1. Refactor repository methods to use JOINs
2. Replace loops with SQL aggregations
3. Test thoroughly
4. **Expected Time**: 8-16 hours
5. **Downtime**: None (can deploy gradually)

## Cost-Benefit Analysis

### Implementation Costs
- **Development Time**: 20-30 hours
- **Testing Time**: 10-15 hours
- **Storage Overhead**: ~50-100 MB (negligible)
- **Maintenance**: Low (automated refresh)

### Benefits
- **90-95% performance improvement**
- **Better user experience** (faster page loads)
- **Reduced server costs** (lower CPU usage)
- **Better scalability** (handle 100+ branches)
- **Historical data** (trends, analytics)

### ROI
- **Break-even**: ~1-2 weeks (developer time saved)
- **Long-term**: Significant cost savings on infrastructure

## Recommendations Priority

### 🔴 Critical (Do Immediately)
1. **Add missing indexes** (especially `staff_schedules`)
2. **Create summary table** for current day quota status
3. **Implement batch schedule queries** (already done)

### 🟡 High Priority (Do Soon)
4. **Create materialized view** for historical data
5. **Refactor queries** to use JOINs
6. **Add database-level caching** (PostgreSQL query cache)

### 🟢 Medium Priority (Do When Possible)
7. **Implement triggers** for automatic summary updates
8. **Add query result caching** at application level
9. **Consider read replicas** for heavy read workloads

## Conclusion

The current database design's reliance on **dynamic calculations** is the primary performance bottleneck. Implementing **materialized views** and **summary tables** would provide the most significant performance gains (90-95% improvement), while adding **proper indexes** would provide immediate 40-60% improvement with minimal effort.

The **hybrid approach** (materialized views + summary tables + indexes) offers the best balance of performance, maintainability, and real-time accuracy.

**Recommended Action**: Start with **Phase 1 (Indexes)** for immediate gains, then proceed with **Phase 2 (Summary Tables)** for long-term optimization.

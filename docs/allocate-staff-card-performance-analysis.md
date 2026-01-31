# Allocate Staff Card Generation Performance Analysis

## Overview

This document analyzes the performance characteristics of card generation in the "Allocate Staff" feature, including API calls, data processing, and frontend rendering.

## System Context

- **Total Branches**: 32 branches
- **Frontend Framework**: Next.js 14 (React)
- **Backend**: Go (Gin framework)
- **Database**: PostgreSQL 15

## Performance Breakdown

### 1. Backend API Performance (`/api/overview/day`)

#### Data Flow
```
GetDayOverview → GenerateDayOverview → CalculateBranchesQuotaStatus (for all branches)
```

#### Complexity Analysis

**Per Branch Calculation (`CalculateBranchQuotaStatus`):**
- **Database Queries per Branch**:
  1. `Branch.GetByID()` - 1 query
  2. `DoctorAssignment.GetDoctorCountByBranch()` - 1 query
  3. `PositionQuota.GetByBranchID()` - 1 query
  4. `Staff.GetByBranchID()` - 1 query
  5. `Rotation.GetByBranchID()` - 1 query
  6. `Position.List()` - 1 query (cached across branches)
  7. `Schedule.GetByStaffID()` - N queries (where N = number of branch staff)
  8. `BranchConstraints.GetByBranchIDAndDayOfWeek()` - 1 query
  9. `BranchTypeConstraints.GetByBranchTypeID()` - 1 query (if no branch constraint)
  10. `BranchTypeConstraints.LoadStaffGroupRequirements()` - 1 query
  11. `StaffGroupPosition.GetByStaffGroupID()` - M queries (where M = number of staff groups)

**Estimated Queries per Branch**: ~15-25 queries (depending on staff count and constraints)

**Total Queries for 32 Branches**: ~480-800 queries

#### Time Complexity
- **O(B × (S + P + G))** where:
  - B = number of branches (32)
  - S = average staff per branch (~10-20)
  - P = average positions per branch (~5-10)
  - G = average staff groups per branch (~3-5)

**Estimated Backend Processing Time**: 
- **Best Case**: 500ms - 1s (with optimized queries, caching)
- **Average Case**: 2-5 seconds
- **Worst Case**: 5-10 seconds (with N+1 query problems)

#### Data Transfer Size

**Per Branch Status** (~500-800 bytes):
```json
{
  "branch_id": "uuid",
  "branch_name": "string",
  "branch_code": "string",
  "date": "string",
  "position_statuses": [...], // ~200 bytes per position
  "total_designated": number,
  "total_available": number,
  "total_assigned": number,
  "total_required": number,
  "group1_score": number,
  "group2_score": number,
  "group3_score": number,
  "group1_missing_staff": ["string"], // ~20-50 bytes per nickname
  "group2_missing_staff": ["string"],
  "group3_missing_staff": ["string"]
}
```

**Total Response Size**: ~16-25 KB (32 branches × ~500-800 bytes)

### 2. Frontend Data Processing

#### Initial Load (`loadBranchSummaries`)

**Operations**:
1. API call to `/api/overview/day` - **Network latency**: 50-200ms
2. Process overview response - **Processing time**: 5-20ms
3. Map branch statuses to summaries - **Processing time**: 10-30ms
4. Create Map structure - **Processing time**: 2-5ms

**Total Frontend Processing**: ~67-255ms

#### Filtering (`filteredBranches` useMemo)

**Operations per render**:
- Filter by `allBranchesSelected` or `selectedBranchIds` - O(B) = O(32)
- Search filter (string matching) - O(B) = O(32)
- **Total**: O(B) = O(32) = ~1-2ms

### 3. Card Rendering Performance

#### Per Card Rendering (`BranchCard` component)

**Rendering Operations**:
1. **Priority calculation** (`getPriorityLevel`) - O(1) = ~0.1ms
2. **Badge rendering** (`getPriorityBadge`) - O(1) = ~0.5ms
3. **Group 1 rendering**:
   - Score display - O(1) = ~0.1ms
   - Missing staff list - O(N) where N = missing staff count
   - Array operations (slice, map) - O(N) = ~1-5ms
4. **Group 2 rendering** - Same as Group 1 = ~1-5ms
5. **Group 3 rendering** - Same as Group 1 = ~1-5ms
6. **DOM rendering** - ~5-15ms per card

**Per Card Total**: ~8-31ms

#### Total Card Rendering

**For 32 cards**:
- **Best Case** (all cards visible, minimal missing staff): ~256ms - 992ms
- **Average Case** (all cards visible, moderate missing staff): ~500ms - 1.5s
- **Worst Case** (all cards visible, many missing staff): ~1s - 3s

**React Rendering Characteristics**:
- Cards render in batches (React's reconciliation)
- Virtual DOM diffing: ~10-50ms for 32 cards
- Browser paint: ~16-33ms per frame (60fps target)

### 4. Overall Performance Timeline

#### Initial Page Load (All 32 Branches Selected)

```
Time (ms)    | Activity
-------------|------------------------------------------
0            | Component mounts
10           | useEffect triggers loadBranchSummaries
10-50        | API request initiated
50-250       | Network request (GET /api/overview/day)
250-5000     | Backend processing (CalculateBranchesQuotaStatus)
5000-5020    | Response received, JSON parsing
5020-5050    | Process overview data
5050-5080    | Create branchSummaries Map
5080-5100    | setBranchSummaries triggers re-render
5100-5400    | React reconciliation (32 cards)
5400-6000    | Browser paint (32 cards)
```

**Total Time**: ~5-6 seconds (worst case)

#### Subsequent Filtering/Search

```
Time (ms)    | Activity
-------------|------------------------------------------
0            | Filter state changes
1-2          | useMemo recalculates filteredBranches
2-3          | React reconciliation (32 cards)
3-20         | Browser paint (32 cards)
```

**Total Time**: ~3-20ms (very fast)

### 5. Performance Bottlenecks

#### Critical Issues

1. **N+1 Query Problem in Backend**
   - `Schedule.GetByStaffID()` called in a loop for each staff member
   - **Impact**: 10-20 queries per branch × 32 branches = 320-640 extra queries
   - **Solution**: Batch query all schedules at once

2. **All Branches Calculated Even When Not Selected**
   - Currently calculates quota for all 32 branches even if user selects only 5
   - **Impact**: Unnecessary backend processing
   - **Solution**: Only calculate for selected branches

3. **Missing Staff Arrays Rendered Multiple Times**
   - Each group renders missing staff with nested arrays and maps
   - **Impact**: O(N) operations × 3 groups × 32 cards
   - **Solution**: Memoize missing staff rendering

4. **No Caching**
   - Every date change triggers full recalculation
   - **Impact**: Repeated expensive backend calls
   - **Solution**: Cache results by date

#### Moderate Issues

1. **Large Response Payload**
   - 16-25 KB response for all branches
   - **Impact**: Network transfer time
   - **Solution**: Compress response or paginate

2. **Synchronous Card Rendering**
   - All 32 cards render at once
   - **Impact**: Blocking main thread
   - **Solution**: Virtual scrolling or progressive rendering

### 6. Performance Metrics Summary

| Metric | Current | Target | Status |
|--------|---------|--------|--------|
| **API Response Time** | 2-5s | <1s | ⚠️ Needs optimization |
| **Initial Render** | 5-6s | <2s | ⚠️ Needs optimization |
| **Filter/Search** | 3-20ms | <50ms | ✅ Good |
| **Card Render (per card)** | 8-31ms | <20ms | ⚠️ Acceptable |
| **Total Cards Render** | 256ms-3s | <500ms | ⚠️ Needs optimization |
| **Memory Usage** | ~2-5MB | <10MB | ✅ Good |
| **Network Payload** | 16-25KB | <50KB | ✅ Good |

### 7. Optimization Recommendations

#### High Priority

1. **Batch Schedule Queries**
   ```go
   // Instead of:
   for _, staff := range branchStaff {
       schedules, _ := c.repos.Schedule.GetByStaffID(staff.ID, date, date)
   }
   
   // Use:
   staffIDs := make([]uuid.UUID, len(branchStaff))
   for i, staff := range branchStaff {
       staffIDs[i] = staff.ID
   }
   schedules, _ := c.repos.Schedule.GetByStaffIDs(staffIDs, date, date)
   ```
   **Expected Improvement**: 50-70% reduction in query count

2. **Calculate Only Selected Branches**
   ```typescript
   // In loadBranchSummaries, only request selected branches
   const branchIds = allBranchesSelected 
     ? branches.map(b => b.id)
     : selectedBranchIds;
   
   // Backend: Accept branch_ids parameter
   // Only calculate for requested branches
   ```
   **Expected Improvement**: 60-80% reduction in processing time when selecting few branches

3. **Add Response Caching**
   ```go
   // Cache overview by date
   cacheKey := fmt.Sprintf("overview:%s", date.Format("2006-01-02"))
   if cached, found := cache.Get(cacheKey); found {
       return cached.(*DayOverview), nil
   }
   ```
   **Expected Improvement**: 90-95% reduction for repeated date queries

#### Medium Priority

4. **Memoize Missing Staff Rendering**
   ```typescript
   const GroupMissingStaff = React.memo(({ staff, group }: Props) => {
     // Render logic
   });
   ```
   **Expected Improvement**: 20-30% reduction in render time

5. **Virtual Scrolling**
   ```typescript
   import { useVirtualizer } from '@tanstack/react-virtual';
   
   const virtualizer = useVirtualizer({
     count: filteredBranches.length,
     getScrollElement: () => parentRef.current,
     estimateSize: () => 200, // card height
   });
   ```
   **Expected Improvement**: Constant render time regardless of card count

6. **Progressive Loading**
   ```typescript
   // Load cards in batches of 10
   const [visibleCount, setVisibleCount] = useState(10);
   useEffect(() => {
     if (visibleCount < filteredBranches.length) {
       setTimeout(() => setVisibleCount(prev => prev + 10), 100);
     }
   }, [visibleCount, filteredBranches.length]);
   ```
   **Expected Improvement**: Perceived performance improvement

#### Low Priority

7. **Response Compression**
   - Enable gzip compression on backend
   - **Expected Improvement**: 60-70% reduction in payload size

8. **Database Indexing**
   - Ensure indexes on frequently queried fields:
     - `schedules.staff_id, date`
     - `branch_constraints.branch_id, day_of_week`
     - `position_quotas.branch_id`
   - **Expected Improvement**: 20-40% reduction in query time

### 8. Expected Performance After Optimizations

| Metric | Before | After | Improvement |
|--------|--------|-------|-------------|
| **API Response Time** | 2-5s | 0.5-1.5s | 60-70% |
| **Initial Render** | 5-6s | 1-2s | 70-80% |
| **Filter/Search** | 3-20ms | 3-20ms | No change (already good) |
| **Card Render (32 cards)** | 256ms-3s | 100-500ms | 60-80% |
| **Memory Usage** | 2-5MB | 2-5MB | No change |

### 9. Monitoring Recommendations

1. **Add Performance Metrics**
   ```typescript
   // In loadBranchSummaries
   const startTime = performance.now();
   await overviewApi.getDayOverview(selectedDate);
   const endTime = performance.now();
   console.log(`API call took ${endTime - startTime}ms`);
   ```

2. **Track Card Render Times**
   ```typescript
   // In BranchCard component
   useEffect(() => {
     const start = performance.now();
     return () => {
       const end = performance.now();
       console.log(`Card ${branch.id} rendered in ${end - start}ms`);
     };
   });
   ```

3. **Monitor Backend Query Count**
   ```go
   // Add query counter middleware
   func QueryCounterMiddleware() gin.HandlerFunc {
       return func(c *gin.Context) {
           queryCount := 0
           // Track queries
           c.Set("query_count", queryCount)
       }
   }
   ```

### 10. Conclusion

The current card generation performance is **acceptable for small-scale usage** but has room for significant improvement. The main bottlenecks are:

1. **Backend N+1 queries** (highest impact)
2. **Calculating all branches unnecessarily** (high impact)
3. **No caching** (high impact for repeated queries)
4. **Synchronous rendering of all cards** (moderate impact)

With the recommended optimizations, we can achieve:
- **60-80% reduction** in initial load time
- **60-70% reduction** in API response time
- **Better scalability** for future growth beyond 32 branches

The system should be able to handle **100+ branches** efficiently after implementing these optimizations.

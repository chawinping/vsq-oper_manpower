# Test Data Generation: Staff Schedule Alternatives

This document outlines 3 different approaches for generating staff working days/leave days for all branches with rules-based randomization for manual testing in the development environment.

## Overview

The goal is to create a function that can:
1. Set staff working days/leave days for all branches
2. Apply randomization with configurable rules
3. Support bulk operations for efficient testing

## Alternative 1: Admin API Endpoint (Recommended for Interactive Testing)

### Description
Create a dedicated admin-only API endpoint that generates schedules for all staff across all branches based on configurable rules.

### Pros
- ✅ Easy to call from frontend or Postman/curl
- ✅ Can be triggered on-demand without server restart
- ✅ Can return progress/statistics
- ✅ Can be integrated into admin UI for easy testing
- ✅ Supports different rule configurations via request body
- ✅ Can validate rules before execution

### Cons
- ❌ Requires server to be running
- ❌ Needs authentication/authorization setup
- ❌ May timeout for very large datasets

### Implementation Structure

**Handler:** `backend/internal/handlers/test_data_handler.go`
```go
type TestDataHandler struct {
    repos *postgres.Repositories
}

type GenerateScheduleRequest struct {
    StartDate      string  `json:"start_date" binding:"required"` // YYYY-MM-DD
    EndDate        string  `json:"end_date" binding:"required"`   // YYYY-MM-DD
    Rules          ScheduleRules `json:"rules"`
    OverwriteExisting bool  `json:"overwrite_existing"` // Whether to overwrite existing schedules
}

type ScheduleRules struct {
    WorkingDaysPerWeek    int     `json:"working_days_per_week"`    // e.g., 5-6
    MinWorkingDaysPerWeek  int     `json:"min_working_days_per_week"` // e.g., 4
    MaxWorkingDaysPerWeek  int     `json:"max_working_days_per_week"` // e.g., 6
    LeaveProbability       float64 `json:"leave_probability"`        // 0.0-1.0, e.g., 0.1 = 10% chance
    ConsecutiveLeaveMax    int     `json:"consecutive_leave_max"`   // Max consecutive leave days
    WeekendWorkingRatio    float64 `json:"weekend_working_ratio"`   // 0.0-1.0, e.g., 0.3 = 30% work weekends
    ExcludeHolidays        bool    `json:"exclude_holidays"`        // Skip public holidays
    BranchSpecificRules    map[uuid.UUID]BranchScheduleRules `json:"branch_specific_rules,omitempty"`
}

type BranchScheduleRules struct {
    WorkingDaysPerWeek    *int     `json:"working_days_per_week,omitempty"`
    LeaveProbability      *float64 `json:"leave_probability,omitempty"`
}
```

**Use Case:** `backend/internal/usecases/test_data/schedule_generator.go`
- Contains the core logic for generating schedules
- Applies rules and randomization
- Handles bulk database operations

**Route:** Add to `backend/cmd/server/main.go`
```go
// Admin-only test data generation
admin := protected.Group("/admin")
admin.Use(middleware.RequireRole("admin"))
{
    admin.POST("/test-data/generate-schedules", h.TestData.GenerateSchedules)
    admin.GET("/test-data/generate-schedules/status", h.TestData.GetGenerationStatus)
}
```

### Usage Example
```bash
# Via curl
curl -X POST http://localhost:8081/api/admin/test-data/generate-schedules \
  -H "Content-Type: application/json" \
  -H "Cookie: vsq_session=..." \
  -d '{
    "start_date": "2026-01-01",
    "end_date": "2026-01-31",
    "overwrite_existing": false,
    "rules": {
      "min_working_days_per_week": 4,
      "max_working_days_per_week": 6,
      "leave_probability": 0.15,
      "consecutive_leave_max": 3,
      "weekend_working_ratio": 0.3,
      "exclude_holidays": true
    }
  }'
```

---

## Alternative 2: Standalone Go CLI Tool (Recommended for Scripting)

### Description
Create a standalone command-line tool that can be run independently to generate test data. This tool connects directly to the database and generates schedules.

### Pros
- ✅ Can be run independently without server
- ✅ Can be scheduled (cron, task scheduler)
- ✅ Can be version-controlled and shared easily
- ✅ Fast execution (no HTTP overhead)
- ✅ Can be integrated into CI/CD pipelines
- ✅ Supports command-line flags for flexibility

### Cons
- ❌ Requires Go runtime environment
- ❌ Less interactive than API endpoint
- ❌ Need to handle database connection directly

### Implementation Structure

**Main File:** `backend/cmd/test-data-generator/main.go`
```go
package main

import (
    "flag"
    "log"
    "time"
    "vsq-oper-manpower/backend/internal/config"
    "vsq-oper-manpower/backend/internal/repositories/postgres"
    "vsq-oper-manpower/backend/internal/usecases/test_data"
)

func main() {
    var (
        startDate      = flag.String("start-date", "", "Start date (YYYY-MM-DD)")
        endDate        = flag.String("end-date", "", "End date (YYYY-MM-DD)")
        workingDays    = flag.Int("working-days", 5, "Working days per week")
        leaveProb      = flag.Float64("leave-probability", 0.1, "Leave probability (0.0-1.0)")
        overwrite      = flag.Bool("overwrite", false, "Overwrite existing schedules")
        configFile     = flag.String("config", "", "Path to config file (optional)")
    )
    flag.Parse()
    
    // Load config and connect to DB
    cfg := config.Load()
    db, err := postgres.NewConnection(cfg.Database)
    if err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    defer db.Close()
    
    repos := postgres.NewRepositories(db)
    generator := test_data.NewScheduleGenerator(repos)
    
    // Generate schedules
    // ...
}
```

**Config File Support:** `backend/cmd/test-data-generator/config.json`
```json
{
  "rules": {
    "min_working_days_per_week": 4,
    "max_working_days_per_week": 6,
    "leave_probability": 0.15,
    "consecutive_leave_max": 3,
    "weekend_working_ratio": 0.3,
    "exclude_holidays": true
  },
  "branch_overrides": {
    "branch-uuid-here": {
      "working_days_per_week": 6,
      "leave_probability": 0.05
    }
  }
}
```

### Usage Example
```powershell
# Via PowerShell
cd backend
go run cmd/test-data-generator/main.go `
  -start-date "2026-01-01" `
  -end-date "2026-01-31" `
  -working-days 5 `
  -leave-probability 0.15 `
  -overwrite false

# Or build and run
go build -o bin/test-data-generator.exe cmd/test-data-generator/main.go
.\bin\test-data-generator.exe -start-date "2026-01-01" -end-date "2026-01-31"
```

---

## Alternative 3: Database Seed Script (Recommended for Initial Setup)

### Description
Create a Go script that can be run as a one-time seed operation or migration. This approach is best for initial test data setup or resetting test environments.

### Pros
- ✅ Can be run as part of database setup
- ✅ Ensures consistent test data across environments
- ✅ Can be versioned with migrations
- ✅ Fast bulk operations
- ✅ Can be run before server starts

### Cons
- ❌ Less flexible for iterative testing
- ❌ Typically run once, not on-demand
- ❌ May need to clear existing data first

### Implementation Structure

**Seed File:** `backend/internal/repositories/postgres/seeds/schedule_seed.go`
```go
package seeds

import (
    "log"
    "time"
    "vsq-oper-manpower/backend/internal/domain/models"
    "vsq-oper-manpower/backend/internal/repositories/postgres"
)

type ScheduleSeedConfig struct {
    StartDate           time.Time
    EndDate             time.Time
    WorkingDaysPerWeek  int
    LeaveProbability    float64
    WeekendWorkingRatio float64
    // ... other rules
}

func SeedSchedules(repos *postgres.Repositories, config ScheduleSeedConfig) error {
    // Get all branches
    branches, err := repos.Branch.List()
    if err != nil {
        return err
    }
    
    // Get all branch staff
    for _, branch := range branches {
        staff, err := repos.Staff.GetByBranchID(branch.ID)
        if err != nil {
            log.Printf("Error getting staff for branch %s: %v", branch.ID, err)
            continue
        }
        
        // Generate schedules for each staff member
        for _, s := range staff {
            schedules := generateSchedulesForStaff(s, branch, config)
            for _, schedule := range schedules {
                if err := repos.Schedule.Create(schedule); err != nil {
                    log.Printf("Error creating schedule: %v", err)
                }
            }
        }
    }
    
    return nil
}
```

**Migration Integration:** `backend/internal/repositories/postgres/migrations.go`
```go
// Add a function to run seeds
func RunSeeds(db *sql.DB) error {
    repos := NewRepositories(db)
    
    // Only run in development
    if os.Getenv("ENVIRONMENT") == "development" {
        config := seeds.ScheduleSeedConfig{
            StartDate:           time.Now(),
            EndDate:             time.Now().AddDate(0, 1, 0), // Next month
            WorkingDaysPerWeek:  5,
            LeaveProbability:    0.1,
            WeekendWorkingRatio: 0.3,
        }
        return seeds.SeedSchedules(repos, config)
    }
    
    return nil
}
```

**PowerShell Script:** `scripts/seed-test-schedules.ps1`
```powershell
# Script to seed test schedules
param(
    [string]$StartDate = (Get-Date).ToString("yyyy-MM-dd"),
    [string]$EndDate = (Get-Date).AddMonths(1).ToString("yyyy-MM-dd"),
    [int]$WorkingDays = 5,
    [double]$LeaveProbability = 0.15
)

Write-Host "Seeding test schedules..."
Write-Host "Start Date: $StartDate"
Write-Host "End Date: $EndDate"
Write-Host "Working Days Per Week: $WorkingDays"
Write-Host "Leave Probability: $LeaveProbability"

cd backend
$env:ENVIRONMENT = "development"
go run cmd/seed/main.go -start-date $StartDate -end-date $EndDate -working-days $WorkingDays -leave-probability $LeaveProbability
```

### Usage Example
```powershell
# Run seed script
.\scripts\seed-test-schedules.ps1 -StartDate "2026-01-01" -EndDate "2026-01-31"

# Or run directly
cd backend
go run cmd/seed/main.go
```

---

## Recommended Rules Implementation

All three alternatives should support these common rules:

### Core Rules
1. **Working Days Per Week**: Randomize between min/max (e.g., 4-6 days)
2. **Leave Probability**: Percentage chance of taking leave on any given day
3. **Consecutive Leave Limit**: Maximum consecutive leave days (e.g., max 3 days)
4. **Weekend Working Ratio**: Percentage of weekends that staff work
5. **Holiday Exclusion**: Skip public holidays (can use a holidays list)
6. **Branch-Specific Overrides**: Different rules per branch if needed

### Example Rule Logic
```go
// Pseudo-code for schedule generation
for each staff member:
    for each date in date range:
        if isHoliday(date) && excludeHolidays:
            schedule = "off"
        else if random() < leaveProbability:
            if consecutiveLeaveDays < maxConsecutiveLeave:
                schedule = "leave"
                consecutiveLeaveDays++
            else:
                schedule = "working"
                consecutiveLeaveDays = 0
        else if isWeekend(date):
            if random() < weekendWorkingRatio:
                schedule = "working"
            else:
                schedule = "off"
        else:
            if workingDaysThisWeek < minWorkingDays:
                schedule = "working"
            else if workingDaysThisWeek >= maxWorkingDays:
                schedule = "off"
            else:
                schedule = random() < 0.7 ? "working" : "off"
```

---

## Comparison Matrix

| Feature | API Endpoint | CLI Tool | Seed Script |
|---------|-------------|----------|-------------|
| **Ease of Use** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐ |
| **Flexibility** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐ |
| **Performance** | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| **Integration** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐ |
| **On-Demand** | ✅ | ✅ | ❌ |
| **Scheduled** | ⚠️ | ✅ | ❌ |
| **Initial Setup** | ⚠️ | ✅ | ✅ |

---

## Recommendation

**For your use case (manual testing in development):**

1. **Primary Choice: Alternative 1 (API Endpoint)** - Best for interactive testing, can be called from Postman or a simple admin UI
2. **Secondary Choice: Alternative 2 (CLI Tool)** - Good for automation and scripting scenarios
3. **Tertiary Choice: Alternative 3 (Seed Script)** - Useful for initial database setup or resetting test data

**Hybrid Approach:** Implement Alternative 1 (API) for flexibility, and Alternative 2 (CLI) for automation/CI scenarios. Both can share the same core use case logic.

---

## Next Steps

1. Choose which alternative(s) to implement
2. Create the core schedule generation logic in `backend/internal/usecases/test_data/`
3. Implement the chosen interface(s)
4. Add appropriate tests
5. Document usage in development guide

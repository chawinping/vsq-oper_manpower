package test_data

import (
	"fmt"
	"math/rand"
	"time"

	"vsq-oper-manpower/backend/internal/domain/interfaces"
	"vsq-oper-manpower/backend/internal/domain/models"
	"vsq-oper-manpower/backend/internal/repositories/postgres"

	"github.com/google/uuid"
)

// ScheduleRules defines the rules for generating staff schedules
type ScheduleRules struct {
	MinWorkingDaysPerWeek   int                    `json:"min_working_days_per_week"`
	MaxWorkingDaysPerWeek   int                    `json:"max_working_days_per_week"`
	LeaveProbability        float64                `json:"leave_probability"` // 0.0-1.0
	ConsecutiveLeaveMax     int                    `json:"consecutive_leave_max"`
	WeekendWorkingRatio     float64                `json:"weekend_working_ratio"` // 0.0-1.0
	ExcludeHolidays         bool                   `json:"exclude_holidays"`
	MinOffDaysPerMonth      int                    `json:"min_off_days_per_month"`      // Minimum off days per month per staff
	MaxOffDaysPerMonth      int                    `json:"max_off_days_per_month"`      // Maximum off days per month per staff
	EnforceMinStaffPerGroup bool                   `json:"enforce_min_staff_per_group"` // Enforce minimum staff per group per day
	BranchSpecificRules     map[string]BranchRules `json:"branch_specific_rules,omitempty"`
}

// BranchRules allows branch-specific overrides
type BranchRules struct {
	MinWorkingDaysPerWeek *int     `json:"min_working_days_per_week,omitempty"`
	MaxWorkingDaysPerWeek *int     `json:"max_working_days_per_week,omitempty"`
	LeaveProbability      *float64 `json:"leave_probability,omitempty"`
	WeekendWorkingRatio   *float64 `json:"weekend_working_ratio,omitempty"`
	MinOffDaysPerMonth    *int     `json:"min_off_days_per_month,omitempty"`
	MaxOffDaysPerMonth    *int     `json:"max_off_days_per_month,omitempty"`
}

// ScheduleGenerator handles generation of test schedules
type ScheduleGenerator struct {
	repos *postgres.Repositories
}

// NewScheduleGenerator creates a new schedule generator
func NewScheduleGenerator(repos *postgres.Repositories) *ScheduleGenerator {
	return &ScheduleGenerator{
		repos: repos,
	}
}

// GenerateSchedulesRequest contains parameters for schedule generation
type GenerateSchedulesRequest struct {
	StartDate         time.Time
	EndDate           time.Time
	Rules             ScheduleRules
	OverwriteExisting bool
	CreatedBy         uuid.UUID
	BranchIDs         []uuid.UUID // Optional: if provided, only generate for these branches. If empty/nil, generates for all branches.
}

// GenerateSchedulesResult contains statistics about the generation
type GenerateSchedulesResult struct {
	TotalStaff     int      `json:"total_staff"`
	TotalSchedules int      `json:"total_schedules"`
	WorkingDays    int      `json:"working_days"`
	LeaveDays      int      `json:"leave_days"`
	OffDays        int      `json:"off_days"`
	Errors         []string `json:"errors,omitempty"`
}

// scheduleEntry stores schedule with related staff and branch info
type scheduleEntry struct {
	schedule *models.StaffSchedule
	staff    *models.Staff
	branch   *models.Branch
}

// GenerateSchedules generates schedules for all branch staff across all branches
func (g *ScheduleGenerator) GenerateSchedules(req GenerateSchedulesRequest) (*GenerateSchedulesResult, error) {
	result := &GenerateSchedulesResult{
		Errors: []string{},
	}

	// Get all branches
	allBranches, err := g.repos.Branch.List()
	if err != nil {
		return nil, fmt.Errorf("failed to get branches: %w", err)
	}

	// Filter branches if BranchIDs are provided
	var branches []*models.Branch
	if len(req.BranchIDs) > 0 {
		// Create a map for quick lookup
		branchIDMap := make(map[uuid.UUID]bool)
		for _, id := range req.BranchIDs {
			branchIDMap[id] = true
		}
		// Filter branches
		for _, branch := range allBranches {
			if branchIDMap[branch.ID] {
				branches = append(branches, branch)
			}
		}
		if len(branches) == 0 {
			return nil, fmt.Errorf("no branches found matching the provided branch_ids")
		}
	} else {
		// Use all branches if no filter specified
		branches = allBranches
	}

	// Get all branch staff
	allStaff, err := g.repos.Staff.List(interfaces.StaffFilters{
		StaffType: &[]models.StaffType{models.StaffTypeBranch}[0],
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get staff: %w", err)
	}

	result.TotalStaff = len(allStaff)

	// Group staff by branch
	staffByBranch := make(map[uuid.UUID][]*models.Staff)
	for _, staff := range allStaff {
		if staff.BranchID != nil {
			staffByBranch[*staff.BranchID] = append(staffByBranch[*staff.BranchID], staff)
		}
	}

	// Store all schedules before committing to database (for minimum staff enforcement)
	allSchedules := make(map[string][]*scheduleEntry) // key: date (YYYY-MM-DD), value: schedules for that date

	// Generate schedules for each branch
	for _, branch := range branches {
		staffList := staffByBranch[branch.ID]
		if len(staffList) == 0 {
			continue
		}

		// Get branch-specific rules if available
		branchRules := req.Rules
		if branchSpecific, ok := req.Rules.BranchSpecificRules[branch.ID.String()]; ok {
			if branchSpecific.MinWorkingDaysPerWeek != nil {
				branchRules.MinWorkingDaysPerWeek = *branchSpecific.MinWorkingDaysPerWeek
			}
			if branchSpecific.MaxWorkingDaysPerWeek != nil {
				branchRules.MaxWorkingDaysPerWeek = *branchSpecific.MaxWorkingDaysPerWeek
			}
			if branchSpecific.LeaveProbability != nil {
				branchRules.LeaveProbability = *branchSpecific.LeaveProbability
			}
			if branchSpecific.WeekendWorkingRatio != nil {
				branchRules.WeekendWorkingRatio = *branchSpecific.WeekendWorkingRatio
			}
			if branchSpecific.MinOffDaysPerMonth != nil {
				branchRules.MinOffDaysPerMonth = *branchSpecific.MinOffDaysPerMonth
			}
			if branchSpecific.MaxOffDaysPerMonth != nil {
				branchRules.MaxOffDaysPerMonth = *branchSpecific.MaxOffDaysPerMonth
			}
		}

		// Generate schedules for each staff member in this branch
		for _, staff := range staffList {
			schedules, err := g.generateSchedulesForStaff(staff, branch, req.StartDate, req.EndDate, branchRules, req.OverwriteExisting, req.CreatedBy)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Error generating schedules for staff %s (%s): %v", staff.ID, staff.Name, err))
				continue
			}

			// Store schedules by date
			for _, schedule := range schedules {
				dateKey := schedule.Date.Format("2006-01-02")
				allSchedules[dateKey] = append(allSchedules[dateKey], &scheduleEntry{
					schedule: schedule,
					staff:    staff,
					branch:   branch,
				})
			}
		}
	}

	// Enforce minimum staff per group if enabled
	if req.Rules.EnforceMinStaffPerGroup {
		if err := g.enforceMinimumStaffPerGroup(allSchedules, branches, req.StartDate, req.EndDate); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Error enforcing minimum staff per group: %v", err))
		}
	}

	// Commit all schedules to database
	for _, dateSchedules := range allSchedules {
		for _, entry := range dateSchedules {
			if err := g.repos.Schedule.Create(entry.schedule); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Error creating schedule for staff %s on %s: %v", entry.staff.ID, entry.schedule.Date.Format("2006-01-02"), err))
				continue
			}

			result.TotalSchedules++
			switch entry.schedule.ScheduleStatus {
			case models.ScheduleStatusWorking:
				result.WorkingDays++
			case models.ScheduleStatusLeave, models.ScheduleStatusSickLeave:
				result.LeaveDays++
			case models.ScheduleStatusOff:
				result.OffDays++
			}
		}
	}

	return result, nil
}

// generateSchedulesForStaff generates schedules for a single staff member
func (g *ScheduleGenerator) generateSchedulesForStaff(
	staff *models.Staff,
	branch *models.Branch,
	startDate, endDate time.Time,
	rules ScheduleRules,
	overwriteExisting bool,
	createdBy uuid.UUID,
) ([]*models.StaffSchedule, error) {
	var schedules []*models.StaffSchedule

	// Get existing schedules if not overwriting
	existingSchedules := make(map[string]*models.StaffSchedule)
	if !overwriteExisting {
		existing, err := g.repos.Schedule.GetByStaffID(staff.ID, startDate, endDate)
		if err != nil {
			return nil, fmt.Errorf("failed to get existing schedules: %w", err)
		}
		for _, s := range existing {
			if s.BranchID == branch.ID {
				key := s.Date.Format("2006-01-02")
				existingSchedules[key] = s
			}
		}
	}

	// Track state per week
	type weekState struct {
		workingDays int
		targetDays  int // Random target for this week
	}
	weekStates := make(map[string]*weekState)

	// Track off days per month
	type monthState struct {
		offDays int
	}
	monthStates := make(map[string]*monthState) // key: YYYY-MM

	// Track consecutive leave days
	consecutiveLeaveDays := 0

	// Thai public holidays (2026) - can be expanded
	holidays := map[string]bool{
		"2026-01-01": true, // New Year
		"2026-02-10": true, // Chinese New Year
		"2026-04-06": true, // Chakri Day
		"2026-04-13": true, // Songkran
		"2026-04-14": true, // Songkran
		"2026-04-15": true, // Songkran
		"2026-05-01": true, // Labor Day
		"2026-05-05": true, // Coronation Day
		"2026-06-03": true, // Queen's Birthday
		"2026-07-28": true, // King's Birthday
		"2026-08-12": true, // Queen's Birthday
		"2026-10-13": true, // King's Memorial Day
		"2026-10-23": true, // Chulalongkorn Day
		"2026-12-05": true, // King's Birthday
		"2026-12-10": true, // Constitution Day
		"2026-12-31": true, // New Year's Eve
	}

	currentDate := startDate
	rng := rand.New(rand.NewSource(time.Now().UnixNano() + int64(staff.ID[0])*1000)) // Seed based on staff ID for consistency

	for currentDate.Before(endDate) || currentDate.Equal(endDate) {
		dateKey := currentDate.Format("2006-01-02")
		dayOfWeek := int(currentDate.Weekday())
		isWeekend := dayOfWeek == 0 || dayOfWeek == 6 // Sunday or Saturday
		isHoliday := holidays[dateKey]

		// Skip if already exists and not overwriting
		if _, exists := existingSchedules[dateKey]; exists && !overwriteExisting {
			currentDate = currentDate.AddDate(0, 0, 1)
			continue
		}

		// Get or create week state
		year, week := currentDate.ISOWeek()
		weekKey := fmt.Sprintf("%d-W%02d", year, week)
		ws, exists := weekStates[weekKey]
		if !exists {
			// Randomize target working days for this week
			targetDays := rules.MinWorkingDaysPerWeek
			if rules.MaxWorkingDaysPerWeek > rules.MinWorkingDaysPerWeek {
				targetDays = rules.MinWorkingDaysPerWeek + rng.Intn(rules.MaxWorkingDaysPerWeek-rules.MinWorkingDaysPerWeek+1)
			}
			ws = &weekState{
				targetDays: targetDays,
			}
			weekStates[weekKey] = ws
		}

		// Get or create month state for off days tracking
		monthKey := currentDate.Format("2006-01")
		ms, exists := monthStates[monthKey]
		if !exists {
			ms = &monthState{offDays: 0}
			monthStates[monthKey] = ms
		}

		var scheduleStatus models.ScheduleStatus

		// Rule 1: Holidays
		if isHoliday && rules.ExcludeHolidays {
			scheduleStatus = models.ScheduleStatusOff
			consecutiveLeaveDays = 0
		} else if isWeekend {
			// Rule 2: Weekends
			if rng.Float64() < rules.WeekendWorkingRatio {
				scheduleStatus = models.ScheduleStatusWorking
				ws.workingDays++
				consecutiveLeaveDays = 0
			} else {
				scheduleStatus = models.ScheduleStatusOff
				consecutiveLeaveDays = 0
			}
		} else {
			// Rule 3: Leave probability
			if rng.Float64() < rules.LeaveProbability && consecutiveLeaveDays < rules.ConsecutiveLeaveMax {
				scheduleStatus = models.ScheduleStatusLeave
				consecutiveLeaveDays++
			} else {
				consecutiveLeaveDays = 0

				// Rule 4: Working days per week constraint
				if ws.workingDays < ws.targetDays {
					// Need more working days this week
					scheduleStatus = models.ScheduleStatusWorking
					ws.workingDays++
				} else if ws.workingDays >= rules.MaxWorkingDaysPerWeek {
					// Already at max working days
					scheduleStatus = models.ScheduleStatusOff
					ms.offDays++
				} else {
					// Check off days per month constraint
					canTakeOff := true
					if rules.MaxOffDaysPerMonth > 0 && ms.offDays >= rules.MaxOffDaysPerMonth {
						// Already at max off days this month, must work
						canTakeOff = false
					}
					if rules.MinOffDaysPerMonth > 0 && ms.offDays < rules.MinOffDaysPerMonth {
						// Need more off days this month
						// Calculate remaining days in month to ensure we can meet minimum
						year, month, _ := currentDate.Date()
						lastDayOfMonth := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
						daysRemaining := lastDayOfMonth - currentDate.Day() + 1
						if ms.offDays+daysRemaining < rules.MinOffDaysPerMonth {
							// Not enough days left to meet minimum, take off now
							canTakeOff = true
						}
					}

					if canTakeOff && rng.Float64() < 0.7 {
						// Random decision with bias towards working
						scheduleStatus = models.ScheduleStatusWorking
						ws.workingDays++
					} else {
						scheduleStatus = models.ScheduleStatusOff
						ms.offDays++
					}
				}
			}
		}

		schedule := &models.StaffSchedule{
			ID:             uuid.New(),
			StaffID:        staff.ID,
			BranchID:       branch.ID,
			Date:           currentDate,
			ScheduleStatus: scheduleStatus,
			IsWorkingDay:   scheduleStatus == models.ScheduleStatusWorking,
			CreatedBy:      createdBy,
			CreatedAt:      time.Now(),
		}

		schedules = append(schedules, schedule)
		currentDate = currentDate.AddDate(0, 0, 1)
	}

	return schedules, nil
}

// enforceMinimumStaffPerGroup ensures minimum staff per group requirements are met
func (g *ScheduleGenerator) enforceMinimumStaffPerGroup(
	allSchedules map[string][]*scheduleEntry,
	branches []*models.Branch,
	startDate, endDate time.Time,
) error {
	// Get all staff groups and their positions
	staffGroups, err := g.repos.StaffGroup.List()
	if err != nil {
		return fmt.Errorf("failed to get staff groups: %w", err)
	}

	// Build position to staff group mapping
	positionToGroups := make(map[uuid.UUID][]uuid.UUID) // positionID -> []staffGroupIDs
	for _, sg := range staffGroups {
		if !sg.IsActive {
			continue
		}
		positions, err := g.repos.StaffGroupPosition.GetByStaffGroupID(sg.ID)
		if err != nil {
			continue
		}
		for _, sgp := range positions {
			positionToGroups[sgp.PositionID] = append(positionToGroups[sgp.PositionID], sg.ID)
		}
	}

	// Process each date
	currentDate := startDate
	for currentDate.Before(endDate) || currentDate.Equal(endDate) {
		dateKey := currentDate.Format("2006-01-02")
		dayOfWeek := int(currentDate.Weekday())
		dateSchedules := allSchedules[dateKey]

		// Group schedules by branch
		schedulesByBranch := make(map[uuid.UUID][]*scheduleEntry)
		for _, entry := range dateSchedules {
			schedulesByBranch[entry.branch.ID] = append(schedulesByBranch[entry.branch.ID], entry)
		}

		// Process each branch
		for _, branch := range branches {
			branchSchedules := schedulesByBranch[branch.ID]
			if len(branchSchedules) == 0 {
				continue
			}

			// Get branch constraints for this day of week
			constraints, err := g.repos.BranchConstraints.GetByBranchID(branch.ID)
			if err != nil {
				continue
			}

			// Load staff group requirements
			if err := g.repos.BranchConstraints.LoadStaffGroupRequirements(constraints); err != nil {
				continue
			}

			var dayConstraint *models.BranchConstraints
			for _, c := range constraints {
				if c.DayOfWeek == dayOfWeek {
					dayConstraint = c
					break
				}
			}

			if dayConstraint == nil || len(dayConstraint.StaffGroupRequirements) == 0 {
				continue
			}

			// Check each staff group requirement
			for _, req := range dayConstraint.StaffGroupRequirements {
				// Count working staff in this group for this branch on this date
				workingCount := 0
				var nonWorkingStaff []*scheduleEntry

				for _, entry := range branchSchedules {
					// Check if staff belongs to this group
					staffGroupsForPosition := positionToGroups[entry.staff.PositionID]
					isInGroup := false
					for _, sgID := range staffGroupsForPosition {
						if sgID == req.StaffGroupID {
							isInGroup = true
							break
						}
					}

					if isInGroup {
						if entry.schedule.ScheduleStatus == models.ScheduleStatusWorking {
							workingCount++
						} else {
							nonWorkingStaff = append(nonWorkingStaff, entry)
						}
					}
				}

				// If below minimum, change some non-working staff to working
				if workingCount < req.MinimumCount {
					needed := req.MinimumCount - workingCount
					if needed > len(nonWorkingStaff) {
						needed = len(nonWorkingStaff)
					}

					// Randomly select staff to make working
					rng := rand.New(rand.NewSource(time.Now().UnixNano() + int64(currentDate.Unix())))
					for i := 0; i < needed; i++ {
						if len(nonWorkingStaff) == 0 {
							break
						}
						idx := rng.Intn(len(nonWorkingStaff))
						entry := nonWorkingStaff[idx]
						entry.schedule.ScheduleStatus = models.ScheduleStatusWorking
						entry.schedule.IsWorkingDay = true
						// Remove from list
						nonWorkingStaff = append(nonWorkingStaff[:idx], nonWorkingStaff[idx+1:]...)
					}
				}
			}
		}

		currentDate = currentDate.AddDate(0, 0, 1)
	}

	return nil
}

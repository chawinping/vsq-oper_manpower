package allocation

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"vsq-oper-manpower/backend/internal/domain/models"
)

// QuotaCalculator calculates quota fulfillment and requirements
type QuotaCalculator struct {
	repos *RepositoriesWrapper
}

// NewQuotaCalculator creates a new quota calculator
func NewQuotaCalculator(repos *RepositoriesWrapper) *QuotaCalculator {
	return &QuotaCalculator{repos: repos}
}

// PositionQuotaStatus represents the quota status for a position
type PositionQuotaStatus struct {
	PositionID      uuid.UUID `json:"position_id"`
	PositionName    string    `json:"position_name"`
	DesignatedQuota int       `json:"designated_quota"`
	MinimumRequired int       `json:"minimum_required"`
	AvailableLocal  int       `json:"available_local"`  // Local branch staff available
	AssignedRotation int       `json:"assigned_rotation"` // Rotation staff assigned
	TotalAssigned    int       `json:"total_assigned"`   // Total staff (local + rotation)
	StillRequired    int       `json:"still_required"`   // Staff still needed
}

// DoctorInfo represents basic doctor information
type DoctorInfo struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Code string    `json:"code"`
}

// BranchQuotaStatus represents the quota status for a branch on a specific date
type BranchQuotaStatus struct {
	BranchID        uuid.UUID           `json:"branch_id"`
	BranchName      string              `json:"branch_name"`
	BranchCode      string              `json:"branch_code"`
	Date            time.Time           `json:"date"`
	PositionStatuses []PositionQuotaStatus `json:"position_statuses"`
	TotalDesignated int                 `json:"total_designated"`
	TotalAvailable  int                 `json:"total_available"`
	TotalAssigned   int                 `json:"total_assigned"`
	TotalRequired   int                 `json:"total_required"`
	// Doctors assigned to this branch on this date
	Doctors         []DoctorInfo        `json:"doctors"`
	// Scoring group points and missing staff
	Group1Score     int                 `json:"group1_score"`     // Daily Staff Constraints - Minimum Shortage
	Group2Score     int                 `json:"group2_score"`     // Position Quota - Minimum Shortage
	Group3Score     int                 `json:"group3_score"`     // Position Quota - Preferred Excess
	Group1MissingStaff []string          `json:"group1_missing_staff"` // Staff nicknames who don't work (Group 1)
	Group2MissingStaff []string          `json:"group2_missing_staff"` // Staff nicknames who don't work (Group 2)
	Group3MissingStaff []string          `json:"group3_missing_staff"` // Staff nicknames who don't work (Group 3)
}

// CalculateBranchQuotaStatus calculates quota status for a branch on a specific date
// First tries to use pre-computed summary table, falls back to calculation if not available
func (c *QuotaCalculator) CalculateBranchQuotaStatus(branchID uuid.UUID, date time.Time) (*BranchQuotaStatus, error) {
	// Try to get from summary table first (fast path)
	// If BranchQuotaSummary repository is not available, skip cache and calculate directly
	if c.repos.BranchQuotaSummary == nil {
		return c.calculateBranchQuotaStatus(branchID, date)
	}
	
	summary, err := c.repos.BranchQuotaSummary.GetByBranchIDAndDate(branchID, date)
	if err == nil && summary != nil {
		// Get branch info for name/code
		branch, err := c.repos.Branch.GetByID(branchID)
		if err != nil {
			return nil, fmt.Errorf("failed to get branch: %w", err)
		}
		if branch == nil {
			return nil, fmt.Errorf("branch not found")
		}
		
		// Get position summaries
		positionStatuses, err := c.getPositionStatusesFromSummary(branchID, date)
		if err != nil {
			// Fallback to calculation if position summaries fail
			return c.calculateBranchQuotaStatus(branchID, date)
		}
		
		// Get doctors for this branch and date
		doctors, err := c.getDoctorsForBranch(branchID, date)
		if err != nil {
			// Log error but don't fail - doctors are optional for display
			// Note: This could hide real issues, but we want to continue even if doctor fetch fails
			doctors = []DoctorInfo{}
		}
		
		// Recalculate Group 1 score using correct logic (summary table calculation is incorrect)
		// We need branch staff and schedules for missing staff tracking
		branchStaff, err := c.repos.Staff.GetByBranchID(branchID)
		if err != nil {
			// Fallback to calculation if we can't get branch staff
			return c.calculateBranchQuotaStatus(branchID, date)
		}
		
		staffIDs := make([]uuid.UUID, 0, len(branchStaff))
		for _, staff := range branchStaff {
			staffIDs = append(staffIDs, staff.ID)
		}
		schedulesMap, err := c.repos.Schedule.GetByStaffIDs(staffIDs, date, date)
		if err != nil {
			// Fallback to calculation if we can't get schedules
			return c.calculateBranchQuotaStatus(branchID, date)
		}
		
		rotationAssignments, err := c.repos.Rotation.GetByBranchID(branchID, date, date)
		if err != nil {
			// Use empty list if we can't get rotation assignments
			rotationAssignments = []*models.RotationAssignment{}
		}
		
		// Recalculate Group 1 and Group 2 with correct logic (summary table calculations are incorrect)
		group1Score, group1Missing := c.calculateGroup1ScoreAndMissingStaff(branchID, date, branchStaff, rotationAssignments, schedulesMap, positionStatuses)
		
		// Get quotas and position map for Group 2 calculation
		quotas, err := c.repos.PositionQuota.GetByBranchID(branchID)
		if err != nil {
			// Fallback to calculation if we can't get quotas
			return c.calculateBranchQuotaStatus(branchID, date)
		}
		
		positions, err := c.repos.Position.List()
		if err != nil {
			// Fallback to calculation if we can't get positions
			return c.calculateBranchQuotaStatus(branchID, date)
		}
		
		positionMap := make(map[uuid.UUID]*models.Position)
		for _, pos := range positions {
			positionMap[pos.ID] = pos
		}
		
		group2Score, group2Missing := c.calculateGroup2ScoreAndMissingStaff(branchID, date, quotas, branchStaff, rotationAssignments, positionMap, schedulesMap)
		
		return &BranchQuotaStatus{
			BranchID:         branchID,
			BranchName:       branch.Name,
			BranchCode:       branch.Code,
			Date:             date,
			PositionStatuses: positionStatuses,
			TotalDesignated:  summary.TotalDesignated,
			TotalAvailable:   summary.TotalAvailable,
			TotalAssigned:    summary.TotalAssigned,
			TotalRequired:    summary.TotalRequired,
			Doctors:          doctors,
			Group1Score:      group1Score, // Use recalculated value
			Group2Score:      group2Score, // Use recalculated value
			Group3Score:      summary.Group3Score,
			Group1MissingStaff: group1Missing, // Use recalculated value
			Group2MissingStaff: group2Missing, // Use recalculated value
			Group3MissingStaff: summary.Group3MissingStaff,
		}, nil
	}
	
	// Fallback to calculation if summary not available
	return c.calculateBranchQuotaStatus(branchID, date)
}

// calculateBranchQuotaStatus performs the actual calculation (original implementation)
func (c *QuotaCalculator) calculateBranchQuotaStatus(branchID uuid.UUID, date time.Time) (*BranchQuotaStatus, error) {
	// Get branch info
	branch, err := c.repos.Branch.GetByID(branchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get branch: %w", err)
	}
	if branch == nil {
		return nil, fmt.Errorf("branch not found")
	}

	// Check if branch is operational (has at least one doctor)
	// Branch operational status is determined by doctor assignments
	doctorCount, err := c.repos.DoctorAssignment.GetDoctorCountByBranch(branchID, date)
	if err != nil {
		return nil, fmt.Errorf("failed to get doctor count: %w", err)
	}

	// Get doctors for this branch and date (always fetch, even if count is 0)
	doctors, err := c.getDoctorsForBranch(branchID, date)
	if err != nil {
		// Log error but don't fail - doctors are optional for display
		// Note: This could hide real issues, but we want to continue even if doctor fetch fails
		doctors = []DoctorInfo{}
	}

	// If branch has no doctors, it's closed - return empty status (no rotation staff needed)
	if doctorCount == 0 {
		return &BranchQuotaStatus{
			BranchID:         branchID,
			BranchName:       branch.Name,
			BranchCode:       branch.Code,
			Date:             date,
			PositionStatuses: []PositionQuotaStatus{},
			TotalDesignated:  0,
			TotalAvailable:   0,
			TotalAssigned:    0,
			TotalRequired:    0,
			Doctors:          doctors, // Empty list when no doctors
			Group1Score:      0,
			Group2Score:      0,
			Group3Score:      0,
			Group1MissingStaff: []string{},
			Group2MissingStaff: []string{},
			Group3MissingStaff: []string{},
		}, nil
	}

	// Get position quotas for the branch
	quotas, err := c.repos.PositionQuota.GetByBranchID(branchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get position quotas: %w", err)
	}

	// Get branch staff
	branchStaff, err := c.repos.Staff.GetByBranchID(branchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get branch staff: %w", err)
	}

	// Get rotation assignments for the date
	rotationAssignments, err := c.repos.Rotation.GetByBranchID(branchID, date, date)
	if err != nil {
		return nil, fmt.Errorf("failed to get rotation assignments: %w", err)
	}

	// Get all positions
	positions, err := c.repos.Position.List()
	if err != nil {
		return nil, fmt.Errorf("failed to get positions: %w", err)
	}

	positionMap := make(map[uuid.UUID]*models.Position)
	for _, pos := range positions {
		positionMap[pos.ID] = pos
	}

	// Batch fetch all schedules for branch staff at once
	staffIDs := make([]uuid.UUID, 0, len(branchStaff))
	for _, staff := range branchStaff {
		staffIDs = append(staffIDs, staff.ID)
	}
	schedulesMap, err := c.repos.Schedule.GetByStaffIDs(staffIDs, date, date)
	if err != nil {
		return nil, fmt.Errorf("failed to get schedules: %w", err)
	}

	// Calculate status for each position
	positionStatuses := []PositionQuotaStatus{}
	totalDesignated := 0
	totalAvailable := 0
	totalAssigned := 0
	totalRequired := 0

	for _, quota := range quotas {
		if !quota.IsActive {
			continue
		}

		position := positionMap[quota.PositionID]
		if position == nil {
			continue
		}

		// Count available local staff for this position
		availableLocal := 0
		for _, staff := range branchStaff {
			if staff.PositionID == quota.PositionID {
				// Check if staff is working on this date using batch-fetched schedules
				schedules := schedulesMap[staff.ID]
				if len(schedules) > 0 {
					if schedules[0].ScheduleStatus == models.ScheduleStatusWorking {
						availableLocal++
					}
				}
			}
		}

		// Count rotation staff assigned for this position
		assignedRotation := 0
		for _, assignment := range rotationAssignments {
			staff, err := c.repos.Staff.GetByID(assignment.RotationStaffID)
			if err == nil && staff != nil && staff.PositionID == quota.PositionID {
				assignedRotation++
			}
		}

		totalAssignedForPosition := availableLocal + assignedRotation
		stillRequired := quota.MinimumRequired - totalAssignedForPosition
		if stillRequired < 0 {
			stillRequired = 0
		}

		positionStatuses = append(positionStatuses, PositionQuotaStatus{
			PositionID:      quota.PositionID,
			PositionName:    position.Name,
			DesignatedQuota: quota.DesignatedQuota,
			MinimumRequired: quota.MinimumRequired,
			AvailableLocal:  availableLocal,
			AssignedRotation: assignedRotation,
			TotalAssigned:    totalAssignedForPosition,
			StillRequired:    stillRequired,
		})

		totalDesignated += quota.DesignatedQuota
		totalAvailable += availableLocal
		totalAssigned += totalAssignedForPosition
		totalRequired += stillRequired
	}

	// Calculate scoring groups and missing staff
	// Group 1: Daily Staff Constraints - Minimum Shortage
	// Use positionStatuses to ensure consistency with what's displayed in the UI
	group1Score, group1Missing := c.calculateGroup1ScoreAndMissingStaff(branchID, date, branchStaff, rotationAssignments, schedulesMap, positionStatuses)
	// Group 2: Position Quota - Minimum Shortage
	group2Score, group2Missing := c.calculateGroup2ScoreAndMissingStaff(branchID, date, quotas, branchStaff, rotationAssignments, positionMap, schedulesMap)
	// Group 3: Position Quota - Preferred Excess
	group3Score, group3Missing := c.calculateGroup3ScoreAndMissingStaff(branchID, date, quotas, branchStaff, rotationAssignments, positionMap, schedulesMap)

	result := &BranchQuotaStatus{
		BranchID:         branchID,
		BranchName:       branch.Name,
		BranchCode:       branch.Code,
		Date:             date,
		PositionStatuses: positionStatuses,
		TotalDesignated:  totalDesignated,
		TotalAvailable:   totalAvailable,
		TotalAssigned:    totalAssigned,
		TotalRequired:    totalRequired,
		Doctors:          doctors,
		Group1Score:      group1Score,
		Group2Score:      group2Score,
		Group3Score:      group3Score,
		Group1MissingStaff: group1Missing,
		Group2MissingStaff: group2Missing,
		Group3MissingStaff: group3Missing,
	}
	
	// Save to summary table for future use (async, don't block on error)
	if c.repos.BranchQuotaSummary != nil {
		go func() {
			_ = c.repos.BranchQuotaSummary.Recalculate(branchID, date)
		}()
	}
	
	return result, nil
}

// getPositionStatusesFromSummary is a helper to get position statuses (simplified - still calculates)
// In future, this could use position_quota_daily_summary table
func (c *QuotaCalculator) getPositionStatusesFromSummary(branchID uuid.UUID, date time.Time) ([]PositionQuotaStatus, error) {
	// For now, we still calculate position statuses
	// In Phase 3, we can optimize this to use position_quota_daily_summary table
	quotas, err := c.repos.PositionQuota.GetByBranchID(branchID)
	if err != nil {
		return nil, err
	}
	
	positions, err := c.repos.Position.List()
	if err != nil {
		return nil, err
	}
	
	positionMap := make(map[uuid.UUID]*models.Position)
	for _, pos := range positions {
		positionMap[pos.ID] = pos
	}
	
	branchStaff, err := c.repos.Staff.GetByBranchID(branchID)
	if err != nil {
		return nil, err
	}
	
	staffIDs := make([]uuid.UUID, 0, len(branchStaff))
	for _, staff := range branchStaff {
		staffIDs = append(staffIDs, staff.ID)
	}
	schedulesMap, err := c.repos.Schedule.GetByStaffIDs(staffIDs, date, date)
	if err != nil {
		return nil, err
	}
	
	rotationAssignments, err := c.repos.Rotation.GetByBranchID(branchID, date, date)
	if err != nil {
		return nil, err
	}
	
	positionStatuses := []PositionQuotaStatus{}
	for _, quota := range quotas {
		if !quota.IsActive {
			continue
		}
		
		position := positionMap[quota.PositionID]
		if position == nil {
			continue
		}
		
		availableLocal := 0
		for _, staff := range branchStaff {
			if staff.PositionID == quota.PositionID {
				schedules := schedulesMap[staff.ID]
				if len(schedules) > 0 && schedules[0].ScheduleStatus == models.ScheduleStatusWorking {
					availableLocal++
				}
			}
		}
		
		assignedRotation := 0
		for _, assignment := range rotationAssignments {
			staff, err := c.repos.Staff.GetByID(assignment.RotationStaffID)
			if err == nil && staff != nil && staff.PositionID == quota.PositionID {
				assignedRotation++
			}
		}
		
		totalAssignedForPosition := availableLocal + assignedRotation
		stillRequired := quota.MinimumRequired - totalAssignedForPosition
		if stillRequired < 0 {
			stillRequired = 0
		}
		
		positionStatuses = append(positionStatuses, PositionQuotaStatus{
			PositionID:      quota.PositionID,
			PositionName:    position.Name,
			DesignatedQuota: quota.DesignatedQuota,
			MinimumRequired: quota.MinimumRequired,
			AvailableLocal:  availableLocal,
			AssignedRotation: assignedRotation,
			TotalAssigned:    totalAssignedForPosition,
			StillRequired:    stillRequired,
		})
	}
	
	return positionStatuses, nil
}

// CalculateBranchesQuotaStatus calculates quota status for multiple branches on a specific date
func (c *QuotaCalculator) CalculateBranchesQuotaStatus(branchIDs []uuid.UUID, date time.Time) ([]*BranchQuotaStatus, error) {
	statuses := []*BranchQuotaStatus{}

	for _, branchID := range branchIDs {
		status, err := c.CalculateBranchQuotaStatus(branchID, date)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate quota status for branch %s: %w", branchID, err)
		}
		statuses = append(statuses, status)
	}

	return statuses, nil
}

// CalculateMonthlyQuotaStatus calculates quota status for a branch across a month
func (c *QuotaCalculator) CalculateMonthlyQuotaStatus(branchID uuid.UUID, year int, month int) ([]*BranchQuotaStatus, error) {
	// Get all dates in the month
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, -1)

	statuses := []*BranchQuotaStatus{}
	currentDate := startDate

	for !currentDate.After(endDate) {
		status, err := c.CalculateBranchQuotaStatus(branchID, currentDate)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate quota status for date %s: %w", currentDate.Format("2006-01-02"), err)
		}
		statuses = append(statuses, status)
		currentDate = currentDate.AddDate(0, 0, 1)
	}

	return statuses, nil
}

// calculateGroup1ScoreAndMissingStaff calculates Group 1 score (Daily Staff Constraints - Minimum Shortage) and returns missing staff nicknames
// Uses positionStatuses to ensure consistency with UI display
func (c *QuotaCalculator) calculateGroup1ScoreAndMissingStaff(
	branchID uuid.UUID,
	date time.Time,
	branchStaff []*models.Staff,
	rotationAssignments []*models.RotationAssignment,
	schedulesMap map[uuid.UUID][]*models.StaffSchedule,
	positionStatuses []PositionQuotaStatus,
) (int, []string) {
	dayOfWeek := int(date.Weekday())

	// Get branch to find branch type
	branch, err := c.repos.Branch.GetByID(branchID)
	if err != nil || branch == nil {
		return 0, []string{}
	}

	// Get branch constraints for this day
	constraint, err := c.repos.BranchConstraints.GetByBranchIDAndDayOfWeek(branchID, dayOfWeek)
	if err != nil {
		return 0, []string{}
	}

	var staffGroupRequirements []*models.BranchConstraintStaffGroup

	// If branch-specific constraint exists, use it
	if constraint != nil {
		// Load staff group requirements from branch constraint
		if err := c.repos.BranchConstraints.LoadStaffGroupRequirements([]*models.BranchConstraints{constraint}); err != nil {
			return 0, []string{}
		}
		staffGroupRequirements = constraint.StaffGroupRequirements
	}
	
	// If no branch-specific constraints or they're empty, fallback to branch type constraints
	if len(staffGroupRequirements) == 0 && branch.BranchTypeID != nil {
		branchTypeConstraints, err := c.repos.BranchTypeConstraints.GetByBranchTypeID(*branch.BranchTypeID)
		if err != nil {
			return 0, []string{}
		}

		// Find constraint for this day
		var branchTypeConstraint *models.BranchTypeConstraints
		for _, bt := range branchTypeConstraints {
			if bt.DayOfWeek == dayOfWeek {
				branchTypeConstraint = bt
				break
			}
		}

		if branchTypeConstraint != nil {
			// Load staff group requirements from branch type constraint
			if err := c.repos.BranchTypeConstraints.LoadStaffGroupRequirements([]*models.BranchTypeConstraints{branchTypeConstraint}); err != nil {
				return 0, []string{}
			}

			// Convert BranchTypeConstraintStaffGroup to BranchConstraintStaffGroup
			staffGroupRequirements = make([]*models.BranchConstraintStaffGroup, 0, len(branchTypeConstraint.StaffGroupRequirements))
			for _, btReq := range branchTypeConstraint.StaffGroupRequirements {
				staffGroupRequirements = append(staffGroupRequirements, &models.BranchConstraintStaffGroup{
					StaffGroupID:  btReq.StaffGroupID,
					MinimumCount:  btReq.MinimumCount,
				})
			}
		}
	}

	// If no constraints found (neither branch-specific nor branch type), return 0
	if len(staffGroupRequirements) == 0 {
		return 0, []string{}
	}

	// Get all staff groups and their positions
	staffGroupPositionsMap := make(map[uuid.UUID][]uuid.UUID) // staffGroupID -> []positionID
	for _, req := range staffGroupRequirements {
		positions, err := c.repos.StaffGroupPosition.GetByStaffGroupID(req.StaffGroupID)
		if err != nil {
			continue
		}
		positionIDs := make([]uuid.UUID, 0, len(positions))
		for _, sgp := range positions {
			positionIDs = append(positionIDs, sgp.PositionID)
		}
		staffGroupPositionsMap[req.StaffGroupID] = positionIDs
	}

	totalScore := 0
	missingStaffSet := make(map[string]bool)

	for _, req := range staffGroupRequirements {
		positionIDs := staffGroupPositionsMap[req.StaffGroupID]
		
		// Calculate actual count by summing total_assigned from positionStatuses for all positions in this staff group
		// This ensures consistency with what's displayed in the UI modal
		actualCount := 0
		for _, posStatus := range positionStatuses {
			for _, positionID := range positionIDs {
				if posStatus.PositionID == positionID {
					actualCount += posStatus.TotalAssigned
					break
				}
			}
		}
		
		// Also track missing staff (branch staff who don't work)
		for _, staff := range branchStaff {
			for _, positionID := range positionIDs {
				if staff.PositionID == positionID {
					schedules := schedulesMap[staff.ID]
					if len(schedules) > 0 {
						if schedules[0].ScheduleStatus != models.ScheduleStatusWorking {
							// Staff doesn't work - add to missing staff
							if staff.Nickname != "" {
								missingStaffSet[staff.Nickname] = true
							}
						}
					} else {
						// No schedule found - assume not working
						if staff.Nickname != "" {
							missingStaffSet[staff.Nickname] = true
						}
					}
					break
				}
			}
		}

		shortage := req.MinimumCount - actualCount
		if shortage > 0 {
			points := -1 * shortage
			totalScore += points
		}
	}

	missingStaff := make([]string, 0, len(missingStaffSet))
	for nickname := range missingStaffSet {
		missingStaff = append(missingStaff, nickname)
	}

	return totalScore, missingStaff
}

// calculateGroup2ScoreAndMissingStaff calculates Group 2 score (Position Quota - Minimum Shortage) and returns missing staff nicknames
func (c *QuotaCalculator) calculateGroup2ScoreAndMissingStaff(
	branchID uuid.UUID,
	date time.Time,
	quotas []*models.PositionQuota,
	branchStaff []*models.Staff,
	rotationAssignments []*models.RotationAssignment,
	positionMap map[uuid.UUID]*models.Position,
	schedulesMap map[uuid.UUID][]*models.StaffSchedule,
) (int, []string) {
	totalScore := 0
	missingStaffSet := make(map[string]bool)

	for _, quota := range quotas {
		if !quota.IsActive {
			continue
		}

		// Count branch staff for this position
		branchCount := 0
		for _, staff := range branchStaff {
			if staff.PositionID == quota.PositionID {
				schedules := schedulesMap[staff.ID]
				if len(schedules) > 0 {
					if schedules[0].ScheduleStatus == models.ScheduleStatusWorking {
						branchCount++
					} else {
						// Staff doesn't work (off, leave, sick_leave) - add to missing staff
						if staff.Nickname != "" {
							missingStaffSet[staff.Nickname] = true
						}
					}
				} else {
					// No schedule found - assume not working
					if staff.Nickname != "" {
						missingStaffSet[staff.Nickname] = true
					}
				}
			}
		}

		// Count rotation staff assigned for this position
		rotationCount := 0
		for _, assignment := range rotationAssignments {
			staff, err := c.repos.Staff.GetByID(assignment.RotationStaffID)
			if err == nil && staff != nil && staff.PositionID == quota.PositionID {
				rotationCount++
			}
		}

		// Total current count includes both branch and rotation staff
		currentCount := branchCount + rotationCount

		shortage := quota.MinimumRequired - currentCount
		if shortage > 0 {
			points := -1 * shortage
			totalScore += points
		}
	}

	missingStaff := make([]string, 0, len(missingStaffSet))
	for nickname := range missingStaffSet {
		missingStaff = append(missingStaff, nickname)
	}

	return totalScore, missingStaff
}

// calculateGroup3ScoreAndMissingStaff calculates Group 3 score (Position Quota - Preferred Excess) and returns missing staff nicknames
// Note: For Group 3, missing staff would be those who ARE working but exceed preferred quota, 
// but since we're looking for staff who DON'T work, this will typically be empty
func (c *QuotaCalculator) calculateGroup3ScoreAndMissingStaff(
	branchID uuid.UUID,
	date time.Time,
	quotas []*models.PositionQuota,
	branchStaff []*models.Staff,
	rotationAssignments []*models.RotationAssignment,
	positionMap map[uuid.UUID]*models.Position,
	schedulesMap map[uuid.UUID][]*models.StaffSchedule,
) (int, []string) {
	totalScore := 0
	missingStaffSet := make(map[string]bool)

	for _, quota := range quotas {
		if !quota.IsActive {
			continue
		}

		// Count branch staff for this position
		branchCount := 0
		for _, staff := range branchStaff {
			if staff.PositionID == quota.PositionID {
				schedules := schedulesMap[staff.ID]
				if len(schedules) > 0 {
					if schedules[0].ScheduleStatus == models.ScheduleStatusWorking {
						branchCount++
					} else {
						// Staff doesn't work - add to missing staff
						if staff.Nickname != "" {
							missingStaffSet[staff.Nickname] = true
						}
					}
				} else {
					// No schedule found - assume not working
					if staff.Nickname != "" {
						missingStaffSet[staff.Nickname] = true
					}
				}
			}
		}

		// Count rotation staff assigned for this position
		rotationCount := 0
		for _, assignment := range rotationAssignments {
			staff, err := c.repos.Staff.GetByID(assignment.RotationStaffID)
			if err == nil && staff != nil && staff.PositionID == quota.PositionID {
				rotationCount++
			}
		}

		// Total current count includes both branch and rotation staff
		currentCount := branchCount + rotationCount

		// Only count positions with actual staff number greater than preferred number
		excess := currentCount - quota.DesignatedQuota
		if excess > 0 {
			points := +1 * excess
			totalScore += points
		}
	}

	missingStaff := make([]string, 0, len(missingStaffSet))
	for nickname := range missingStaffSet {
		missingStaff = append(missingStaff, nickname)
	}

	return totalScore, missingStaff
}

// getDoctorsForBranch retrieves doctors assigned to a branch on a specific date
func (c *QuotaCalculator) getDoctorsForBranch(branchID uuid.UUID, date time.Time) ([]DoctorInfo, error) {
	// Check if DoctorAssignment repository is available
	if c.repos.DoctorAssignment == nil {
		return []DoctorInfo{}, nil // Return empty list if repository not available
	}

	doctorAssignments, err := c.repos.DoctorAssignment.GetDoctorsByBranchAndDate(branchID, date)
	if err != nil {
		return nil, fmt.Errorf("failed to get doctors: %w", err)
	}

	doctors := make([]DoctorInfo, 0, len(doctorAssignments))
	for _, assignment := range doctorAssignments {
		// Use DoctorName and DoctorCode from assignment, which should be populated
		doctorName := assignment.DoctorName
		doctorCode := assignment.DoctorCode
		
		// Fallback: if name/code are empty, try to get from doctor repo
		if doctorName == "" || doctorCode == "" {
			if assignment.Doctor != nil {
				if doctorName == "" {
					doctorName = assignment.Doctor.Name
				}
				if doctorCode == "" {
					doctorCode = assignment.Doctor.Code
				}
			}
		}
		
		doctors = append(doctors, DoctorInfo{
			ID:   assignment.DoctorID,
			Name: doctorName,
			Code: doctorCode,
		})
	}

	return doctors, nil
}

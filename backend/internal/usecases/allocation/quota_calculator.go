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
}

// CalculateBranchQuotaStatus calculates quota status for a branch on a specific date
func (c *QuotaCalculator) CalculateBranchQuotaStatus(branchID uuid.UUID, date time.Time) (*BranchQuotaStatus, error) {
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
				// Check if staff is working on this date
				schedules, err := c.repos.Schedule.GetByStaffID(staff.ID, date, date)
				if err == nil && len(schedules) > 0 {
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

	return &BranchQuotaStatus{
		BranchID:         branchID,
		BranchName:       branch.Name,
		BranchCode:       branch.Code,
		Date:             date,
		PositionStatuses: positionStatuses,
		TotalDesignated:  totalDesignated,
		TotalAvailable:   totalAvailable,
		TotalAssigned:    totalAssigned,
		TotalRequired:    totalRequired,
	}, nil
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

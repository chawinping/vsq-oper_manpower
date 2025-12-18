package allocation

import (
	"fmt"
	"vsq-oper-manpower/backend/internal/domain/models"
	"vsq-oper-manpower/backend/internal/domain/interfaces"
)

type AllocationEngine struct {
	repos *interfaces.Repositories
}

func NewAllocationEngine(repos *postgres.Repositories) *AllocationEngine {
	return &AllocationEngine{repos: repos}
}

// CalculateRequiredStaff calculates required staff count based on revenue and rules
func (e *AllocationEngine) CalculateRequiredStaff(
	branchID uuid.UUID,
	date string,
	positionID uuid.UUID,
	expectedRevenue float64,
) (int, error) {
	// Get allocation rule for position
	rule, err := e.repos.AllocationRule.GetByPositionID(positionID)
	if err != nil {
		return 0, err
	}

	if rule == nil {
		// Use default minimum
		position, err := e.repos.Position.GetByID(positionID)
		if err != nil {
			return 0, err
		}
		return position.MinStaffPerBranch, nil
	}

	// Calculate based on formula: min_staff + (revenue / revenue_threshold) * multiplier
	minStaff := rule.MinStaff
	if rule.RevenueThreshold > 0 {
		additionalStaff := int((expectedRevenue / rule.RevenueThreshold) * float64(minStaff))
		return minStaff + additionalStaff, nil
	}

	return minStaff, nil
}

// CheckAvailability checks if rotation staff is available for assignment
func (e *AllocationEngine) CheckAvailability(
	rotationStaffID uuid.UUID,
	branchID uuid.UUID,
	date string,
) (bool, error) {
	// Check if rotation staff has effective branch assignment
	effectiveBranches, err := e.repos.EffectiveBranch.GetByRotationStaffID(rotationStaffID)
	if err != nil {
		return false, err
	}

	hasAccess := false
	for _, eb := range effectiveBranches {
		if eb.BranchID == branchID {
			hasAccess = true
			break
		}
	}

	if !hasAccess {
		return false, fmt.Errorf("rotation staff does not have access to this branch")
	}

	// Check if rotation staff is already assigned on this date
	// Note: This would need date parsing - simplified for now
	// assignments, err := e.repos.Rotation.GetByRotationStaffID(rotationStaffID, date, date)
	// if err != nil {
	// 	return false, err
	// }
	//
	// if len(assignments) > 0 {
	// 	return false, fmt.Errorf("rotation staff is already assigned on this date")
	// }

	return true, nil
}


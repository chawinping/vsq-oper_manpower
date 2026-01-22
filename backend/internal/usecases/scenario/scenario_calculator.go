package scenario

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"vsq-oper-manpower/backend/internal/domain/interfaces"
	"vsq-oper-manpower/backend/internal/domain/models"
)

// ScenarioCalculator calculates staff requirements based on scenarios
type ScenarioCalculator struct {
	repos *RepositoriesWrapper
}

// RepositoriesWrapper wraps all repositories needed for scenario calculation
type RepositoriesWrapper struct {
	BranchWeeklyRevenue          interfaces.BranchWeeklyRevenueRepository
	Revenue                      interfaces.RevenueRepository
	DoctorAssignment             interfaces.DoctorAssignmentRepository
	PositionQuota                interfaces.PositionQuotaRepository
	RevenueLevelTier             interfaces.RevenueLevelTierRepository
	StaffRequirementScenario     interfaces.StaffRequirementScenarioRepository
	ScenarioPositionRequirement  interfaces.ScenarioPositionRequirementRepository
	Position                     interfaces.PositionRepository
}

// NewScenarioCalculator creates a new scenario calculator
func NewScenarioCalculator(repos *RepositoriesWrapper) *ScenarioCalculator {
	return &ScenarioCalculator{repos: repos}
}

// CalculateStaffRequirements calculates staff requirements for a position based on scenarios
func (c *ScenarioCalculator) CalculateStaffRequirements(
	branchID uuid.UUID,
	date time.Time,
	positionID uuid.UUID,
	basePreferred int,
	baseMinimum int,
) (*models.CalculatedRequirement, error) {
	dayOfWeek := int(date.Weekday())

	// Get day-of-week baseline revenue
	dayOfWeekRevenue, err := c.repos.BranchWeeklyRevenue.GetByBranchIDAndDayOfWeek(branchID, dayOfWeek)
	if err != nil {
		return nil, fmt.Errorf("failed to get day-of-week revenue: %w", err)
	}
	var dayOfWeekRevenueValue float64
	if dayOfWeekRevenue != nil {
		// Calculate total revenue from all 4 types (weighted sum)
		dayOfWeekRevenueValue = c.calculateTotalRevenueValue(
			dayOfWeekRevenue.SkinRevenue,
			dayOfWeekRevenue.LSHMRevenue,
			dayOfWeekRevenue.VitaminCases,
			dayOfWeekRevenue.SlimPenCases,
		)
		// Fallback to ExpectedRevenue for backward compatibility if new fields are zero
		if dayOfWeekRevenueValue == 0 && dayOfWeekRevenue.ExpectedRevenue > 0 {
			dayOfWeekRevenueValue = dayOfWeekRevenue.ExpectedRevenue
		}
	}

	// Get specific date revenue (if available)
	specificDateRevenue, err := c.repos.Revenue.GetByBranchID(branchID, date, date)
	if err != nil {
		return nil, fmt.Errorf("failed to get specific date revenue: %w", err)
	}
	var specificDateRevenueValue *float64
	if len(specificDateRevenue) > 0 {
		revenue := specificDateRevenue[0]
		// Prefer actual revenue if available, otherwise use calculated total from all types
		if revenue.ActualRevenue != nil && *revenue.ActualRevenue > 0 {
			value := *revenue.ActualRevenue
			specificDateRevenueValue = &value
		} else {
			// Calculate total revenue from all 4 types
			totalValue := c.calculateTotalRevenueValue(
				revenue.SkinRevenue,
				revenue.LSHMRevenue,
				revenue.VitaminCases,
				revenue.SlimPenCases,
			)
			// Fallback to ExpectedRevenue for backward compatibility if new fields are zero
			if totalValue == 0 && revenue.ExpectedRevenue > 0 {
				totalValue = revenue.ExpectedRevenue
			}
			if totalValue > 0 {
				specificDateRevenueValue = &totalValue
			}
		}
	}

	// Get doctor count for the date
	doctorCount, err := c.repos.DoctorAssignment.GetDoctorCountByBranch(branchID, date)
	if err != nil {
		return nil, fmt.Errorf("failed to get doctor count: %w", err)
	}

	// Get position name
	position, err := c.repos.Position.GetByID(positionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get position: %w", err)
	}
	if position == nil {
		return nil, fmt.Errorf("position not found")
	}

	// Find matching scenarios (ordered by priority DESC)
	scenarios, err := c.repos.StaffRequirementScenario.GetActiveOrderedByPriority()
	if err != nil {
		return nil, fmt.Errorf("failed to get scenarios: %w", err)
	}

	var matchedScenario *models.StaffRequirementScenario
	var factorsApplied []string

	for _, scenario := range scenarios {
		if c.matchesScenario(scenario, dayOfWeekRevenueValue, specificDateRevenueValue, doctorCount, dayOfWeek) {
			matchedScenario = scenario
			factorsApplied = c.buildMatchReason(scenario, dayOfWeekRevenueValue, specificDateRevenueValue, doctorCount)
			break // Use highest priority matching scenario
		}
	}

	// Calculate requirements
	calculatedPreferred := basePreferred
	calculatedMinimum := baseMinimum

	if matchedScenario != nil {
		// Get position requirements for this scenario
		requirement, err := c.repos.ScenarioPositionRequirement.GetByScenarioAndPosition(matchedScenario.ID, positionID)
		if err != nil {
			return nil, fmt.Errorf("failed to get scenario requirement: %w", err)
		}

		if requirement != nil {
			if requirement.OverrideBase {
				calculatedPreferred = requirement.PreferredStaff
				calculatedMinimum = requirement.MinimumStaff
			} else {
				calculatedPreferred = basePreferred + requirement.PreferredStaff
				calculatedMinimum = baseMinimum + requirement.MinimumStaff
			}
		}
	}

	var matchedScenarioID *uuid.UUID
	var matchedScenarioName *string
	if matchedScenario != nil {
		matchedScenarioID = &matchedScenario.ID
		matchedScenarioName = &matchedScenario.ScenarioName
	}

	return &models.CalculatedRequirement{
		PositionID:          positionID,
		PositionName:        position.Name,
		BasePreferred:        basePreferred,
		BaseMinimum:          baseMinimum,
		CalculatedPreferred:  calculatedPreferred,
		CalculatedMinimum:    calculatedMinimum,
		MatchedScenarioID:    matchedScenarioID,
		MatchedScenarioName:  matchedScenarioName,
		FactorsApplied:       factorsApplied,
	}, nil
}

// matchesScenario checks if a scenario matches the given conditions
func (c *ScenarioCalculator) matchesScenario(
	scenario *models.StaffRequirementScenario,
	dayOfWeekRevenue float64,
	specificDateRevenue *float64,
	doctorCount int,
	dayOfWeek int,
) bool {
	// Check day of week filter
	if scenario.DayOfWeek != nil && *scenario.DayOfWeek != dayOfWeek {
		return false
	}

	// Determine which revenue to use
	var revenueToCheck float64
	if scenario.UseSpecificDateRevenue && specificDateRevenue != nil {
		revenueToCheck = *specificDateRevenue
	} else if scenario.UseDayOfWeekRevenue {
		revenueToCheck = dayOfWeekRevenue
	} else {
		// Fallback: use specific date if available, otherwise day-of-week
		if specificDateRevenue != nil {
			revenueToCheck = *specificDateRevenue
		} else {
			revenueToCheck = dayOfWeekRevenue
		}
	}

	// Check revenue tier match
	if scenario.RevenueLevelTierID != nil {
		tier, err := c.repos.RevenueLevelTier.GetTierForRevenue(revenueToCheck)
		if err != nil || tier == nil || tier.ID != *scenario.RevenueLevelTierID {
			return false
		}
	}

	// Check direct revenue range
	if scenario.MinRevenue != nil && revenueToCheck < *scenario.MinRevenue {
		return false
	}
	if scenario.MaxRevenue != nil && revenueToCheck >= *scenario.MaxRevenue {
		return false
	}

	// Check doctor count
	if scenario.DoctorCount != nil && doctorCount != *scenario.DoctorCount {
		return false
	}
	if scenario.MinDoctorCount != nil && doctorCount < *scenario.MinDoctorCount {
		return false
	}

	return true
}

// buildMatchReason builds a human-readable reason for why a scenario matched
func (c *ScenarioCalculator) buildMatchReason(
	scenario *models.StaffRequirementScenario,
	dayOfWeekRevenue float64,
	specificDateRevenue *float64,
	doctorCount int,
) []string {
	reasons := []string{}

	// Determine which revenue was used
	var revenueSource string
	if scenario.UseSpecificDateRevenue && specificDateRevenue != nil {
		revenueSource = "specific date"
	} else if scenario.UseDayOfWeekRevenue {
		revenueSource = "day-of-week"
	} else {
		if specificDateRevenue != nil {
			revenueSource = "specific date"
		} else {
			revenueSource = "day-of-week"
		}
	}

	// Add revenue tier reason
	if scenario.RevenueLevelTierID != nil {
		tier, err := c.repos.RevenueLevelTier.GetByID(*scenario.RevenueLevelTierID)
		if err == nil && tier != nil {
			reasons = append(reasons, fmt.Sprintf("Revenue Level %d (%s) from %s revenue", tier.LevelNumber, tier.LevelName, revenueSource))
		}
	}

	// Add direct revenue range reason
	if scenario.MinRevenue != nil || scenario.MaxRevenue != nil {
		if scenario.MinRevenue != nil && scenario.MaxRevenue != nil {
			reasons = append(reasons, fmt.Sprintf("Revenue %.0f-%.0f THB (%s)", *scenario.MinRevenue, *scenario.MaxRevenue, revenueSource))
		} else if scenario.MinRevenue != nil {
			reasons = append(reasons, fmt.Sprintf("Revenue >= %.0f THB (%s)", *scenario.MinRevenue, revenueSource))
		} else if scenario.MaxRevenue != nil {
			reasons = append(reasons, fmt.Sprintf("Revenue < %.0f THB (%s)", *scenario.MaxRevenue, revenueSource))
		}
	}

	// Add doctor count reason
	if scenario.DoctorCount != nil {
		reasons = append(reasons, fmt.Sprintf("Doctors = %d", *scenario.DoctorCount))
	}
	if scenario.MinDoctorCount != nil {
		reasons = append(reasons, fmt.Sprintf("Doctors >= %d", *scenario.MinDoctorCount))
	}

	// Add day of week reason
	if scenario.DayOfWeek != nil {
		dayNames := []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
		reasons = append(reasons, fmt.Sprintf("Day: %s", dayNames[*scenario.DayOfWeek]))
	}

	return reasons
}

// calculateTotalRevenueValue calculates total revenue value from all 4 revenue types
// Uses weighted sum: Skin + LS HM + (Vitamin Cases * multiplier) + (Slim Pen Cases * multiplier)
// TODO: Make multipliers configurable via system settings
func (c *ScenarioCalculator) calculateTotalRevenueValue(
	skinRevenue float64,
	lsHMRevenue float64,
	vitaminCases int,
	slimPenCases int,
) float64 {
	// Default multipliers for converting cases to revenue equivalent
	// These should be configurable via system settings in the future
	const vitaminCaseMultiplier = 1000.0 // Each vitamin case = 1000 THB equivalent
	const slimPenCaseMultiplier = 1500.0 // Each slim pen case = 1500 THB equivalent

	total := skinRevenue + lsHMRevenue
	total += float64(vitaminCases) * vitaminCaseMultiplier
	total += float64(slimPenCases) * slimPenCaseMultiplier

	return total
}

// GetMatchingScenarios returns all scenarios that match the given conditions
func (c *ScenarioCalculator) GetMatchingScenarios(
	branchID uuid.UUID,
	date time.Time,
) ([]*models.ScenarioMatch, error) {
	dayOfWeek := int(date.Weekday())

	// Get day-of-week baseline revenue
	dayOfWeekRevenue, err := c.repos.BranchWeeklyRevenue.GetByBranchIDAndDayOfWeek(branchID, dayOfWeek)
	if err != nil {
		return nil, fmt.Errorf("failed to get day-of-week revenue: %w", err)
	}
	var dayOfWeekRevenueValue float64
	if dayOfWeekRevenue != nil {
		// Calculate total revenue from all 4 types (weighted sum)
		dayOfWeekRevenueValue = c.calculateTotalRevenueValue(
			dayOfWeekRevenue.SkinRevenue,
			dayOfWeekRevenue.LSHMRevenue,
			dayOfWeekRevenue.VitaminCases,
			dayOfWeekRevenue.SlimPenCases,
		)
		// Fallback to ExpectedRevenue for backward compatibility if new fields are zero
		if dayOfWeekRevenueValue == 0 && dayOfWeekRevenue.ExpectedRevenue > 0 {
			dayOfWeekRevenueValue = dayOfWeekRevenue.ExpectedRevenue
		}
	}

	// Get specific date revenue (if available)
	specificDateRevenue, err := c.repos.Revenue.GetByBranchID(branchID, date, date)
	if err != nil {
		return nil, fmt.Errorf("failed to get specific date revenue: %w", err)
	}
	var specificDateRevenueValue *float64
	if len(specificDateRevenue) > 0 {
		revenue := specificDateRevenue[0]
		// Prefer actual revenue if available, otherwise use calculated total from all types
		if revenue.ActualRevenue != nil && *revenue.ActualRevenue > 0 {
			value := *revenue.ActualRevenue
			specificDateRevenueValue = &value
		} else {
			// Calculate total revenue from all 4 types
			totalValue := c.calculateTotalRevenueValue(
				revenue.SkinRevenue,
				revenue.LSHMRevenue,
				revenue.VitaminCases,
				revenue.SlimPenCases,
			)
			// Fallback to ExpectedRevenue for backward compatibility if new fields are zero
			if totalValue == 0 && revenue.ExpectedRevenue > 0 {
				totalValue = revenue.ExpectedRevenue
			}
			if totalValue > 0 {
				specificDateRevenueValue = &totalValue
			}
		}
	}

	// Get doctor count for the date
	doctorCount, err := c.repos.DoctorAssignment.GetDoctorCountByBranch(branchID, date)
	if err != nil {
		return nil, fmt.Errorf("failed to get doctor count: %w", err)
	}

	// Get all active scenarios
	scenarios, err := c.repos.StaffRequirementScenario.GetActiveOrderedByPriority()
	if err != nil {
		return nil, fmt.Errorf("failed to get scenarios: %w", err)
	}

	matches := []*models.ScenarioMatch{}
	for _, scenario := range scenarios {
		matchesScenario := c.matchesScenario(scenario, dayOfWeekRevenueValue, specificDateRevenueValue, doctorCount, dayOfWeek)
		matchReason := ""
		if matchesScenario {
			reasons := c.buildMatchReason(scenario, dayOfWeekRevenueValue, specificDateRevenueValue, doctorCount)
			matchReason = fmt.Sprintf("%v", reasons)
		}

		matches = append(matches, &models.ScenarioMatch{
			ScenarioID:   scenario.ID,
			ScenarioName: scenario.ScenarioName,
			Matches:      matchesScenario,
			MatchReason:  matchReason,
			Priority:     scenario.Priority,
		})
	}

	return matches, nil
}

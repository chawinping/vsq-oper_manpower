package allocation

import (
	"fmt"
	"time"

	"vsq-oper-manpower/backend/internal/domain/interfaces"
	"vsq-oper-manpower/backend/internal/domain/models"

	"github.com/google/uuid"
)

// RepositoriesWrapper wraps postgres repositories to provide a unified interface
// RepositoriesWrapper wraps repository interfaces for use cases
type RepositoriesWrapper struct {
	User                        interfaces.UserRepository
	Role                        interfaces.RoleRepository
	Staff                       interfaces.StaffRepository
	Position                    interfaces.PositionRepository
	Branch                      interfaces.BranchRepository
	EffectiveBranch             interfaces.EffectiveBranchRepository
	Revenue                     interfaces.RevenueRepository
	Schedule                    interfaces.ScheduleRepository
	Rotation                    interfaces.RotationRepository
	Settings                    interfaces.SettingsRepository
	AllocationRule              interfaces.AllocationRuleRepository
	AreaOfOperation             interfaces.AreaOfOperationRepository
	AllocationCriteria          interfaces.AllocationCriteriaRepository
	PositionQuota               interfaces.PositionQuotaRepository
	Doctor                      interfaces.DoctorRepository
	DoctorPreference            interfaces.DoctorPreferenceRepository
	DoctorAssignment            interfaces.DoctorAssignmentRepository
	DoctorOnOffDay              interfaces.DoctorOnOffDayRepository
	BranchType                  interfaces.BranchTypeRepository
	StaffGroup                  interfaces.StaffGroupRepository
	StaffGroupPosition          interfaces.StaffGroupPositionRepository
	BranchTypeRequirement       interfaces.BranchTypeStaffGroupRequirementRepository
	BranchTypeConstraints       interfaces.BranchTypeConstraintsRepository
	BranchConstraints           interfaces.BranchConstraintsRepository
	RotationStaffBranchPosition interfaces.RotationStaffBranchPositionRepository
	AllocationSuggestion        interfaces.AllocationSuggestionRepository
	BranchQuotaSummary          interfaces.BranchQuotaSummaryRepository
}

// CriteriaEngine evaluates allocation criteria across the three pillars
type CriteriaEngine struct {
	repos *RepositoriesWrapper
}

// NewCriteriaEngine creates a new criteria engine
func NewCriteriaEngine(repos *RepositoriesWrapper) *CriteriaEngine {
	return &CriteriaEngine{repos: repos}
}

// CriteriaScore represents a score for a specific criteria
type CriteriaScore struct {
	CriteriaID uuid.UUID
	Pillar     models.CriteriaPillar
	Type       models.CriteriaType
	Weight     float64
	Score      float64 // Normalized score (0.0 - 1.0)
	RawValue   float64 // Raw value before normalization
}

// PillarScore represents aggregated scores for a pillar
type PillarScore struct {
	Pillar models.CriteriaPillar
	Score  float64 // Weighted average score (0.0 - 1.0)
	Scores []CriteriaScore
}

// AllocationScore represents the overall allocation score
type AllocationScore struct {
	ClinicWideScore     float64
	DoctorSpecificScore float64
	BranchSpecificScore float64
	OverallScore        float64 // Weighted combination of all pillars
	PillarScores        []PillarScore
}

// EvaluateCriteria evaluates allocation criteria for a branch on a specific date
func (e *CriteriaEngine) EvaluateCriteria(branchID uuid.UUID, date time.Time) (*AllocationScore, error) {
	// Get all active criteria
	filters := interfaces.AllocationCriteriaFilters{IsActive: &[]bool{true}[0]}
	allCriteria, err := e.repos.AllocationCriteria.List(filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get criteria: %w", err)
	}

	// Group criteria by pillar
	clinicWideCriteria := []*models.AllocationCriteria{}
	doctorSpecificCriteria := []*models.AllocationCriteria{}
	branchSpecificCriteria := []*models.AllocationCriteria{}

	for _, criteria := range allCriteria {
		switch criteria.Pillar {
		case models.PillarClinicWide:
			clinicWideCriteria = append(clinicWideCriteria, criteria)
		case models.PillarDoctorSpecific:
			doctorSpecificCriteria = append(doctorSpecificCriteria, criteria)
		case models.PillarBranchSpecific:
			branchSpecificCriteria = append(branchSpecificCriteria, criteria)
		}
	}

	// Evaluate each pillar
	clinicWideScore, err := e.evaluatePillar(models.PillarClinicWide, clinicWideCriteria, branchID, date)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate clinic-wide criteria: %w", err)
	}

	doctorSpecificScore, err := e.evaluatePillar(models.PillarDoctorSpecific, doctorSpecificCriteria, branchID, date)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate doctor-specific criteria: %w", err)
	}

	branchSpecificScore, err := e.evaluatePillar(models.PillarBranchSpecific, branchSpecificCriteria, branchID, date)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate branch-specific criteria: %w", err)
	}

	// Calculate overall score (equal weight for now, can be made configurable)
	overallScore := (clinicWideScore.Score + doctorSpecificScore.Score + branchSpecificScore.Score) / 3.0

	return &AllocationScore{
		ClinicWideScore:     clinicWideScore.Score,
		DoctorSpecificScore: doctorSpecificScore.Score,
		BranchSpecificScore: branchSpecificScore.Score,
		OverallScore:        overallScore,
		PillarScores:        []PillarScore{clinicWideScore, doctorSpecificScore, branchSpecificScore},
	}, nil
}

// evaluatePillar evaluates criteria for a specific pillar
func (e *CriteriaEngine) evaluatePillar(pillar models.CriteriaPillar, criteriaList []*models.AllocationCriteria, branchID uuid.UUID, date time.Time) (PillarScore, error) {
	scores := []CriteriaScore{}
	totalWeight := 0.0
	weightedSum := 0.0

	for _, criteria := range criteriaList {
		score, err := e.evaluateCriterion(criteria, branchID, date)
		if err != nil {
			return PillarScore{}, fmt.Errorf("failed to evaluate criterion %s: %w", criteria.ID, err)
		}

		criteriaScore := CriteriaScore{
			CriteriaID: criteria.ID,
			Pillar:     criteria.Pillar,
			Type:       criteria.Type,
			Weight:     criteria.Weight,
			Score:      score,
		}

		scores = append(scores, criteriaScore)
		totalWeight += criteria.Weight
		weightedSum += score * criteria.Weight
	}

	// Calculate weighted average
	var pillarScore float64
	if totalWeight > 0 {
		pillarScore = weightedSum / totalWeight
	} else {
		pillarScore = 0.0
	}

	return PillarScore{
		Pillar: pillar,
		Score:  pillarScore,
		Scores: scores,
	}, nil
}

// evaluateCriterion evaluates a single criterion
func (e *CriteriaEngine) evaluateCriterion(criteria *models.AllocationCriteria, branchID uuid.UUID, date time.Time) (float64, error) {
	switch criteria.Type {
	case models.CriteriaTypeBookings:
		return e.evaluateBookings(criteria, branchID, date)
	case models.CriteriaTypeRevenue:
		return e.evaluateRevenue(criteria, branchID, date)
	case models.CriteriaTypeMinStaffPosition:
		return e.evaluateMinStaffPosition(criteria, branchID, date)
	case models.CriteriaTypeMinStaffBranch:
		return e.evaluateMinStaffBranch(criteria, branchID, date)
	case models.CriteriaTypeDoctorCount:
		return e.evaluateDoctorCount(criteria, branchID, date)
	default:
		return 0.0, fmt.Errorf("unknown criteria type: %s", criteria.Type)
	}
}

// evaluateBookings evaluates bookings criterion (placeholder - will connect to booking system later)
func (e *CriteriaEngine) evaluateBookings(criteria *models.AllocationCriteria, branchID uuid.UUID, date time.Time) (float64, error) {
	// Placeholder: return 0.5 as default score
	// TODO: Connect to booking system API to get actual booking count
	return 0.5, nil
}

// evaluateRevenue evaluates revenue criterion
func (e *CriteriaEngine) evaluateRevenue(criteria *models.AllocationCriteria, branchID uuid.UUID, date time.Time) (float64, error) {
	var revenue float64

	// Get revenue data for the branch on the date
	revenueData, err := e.repos.Revenue.GetByBranchID(branchID, date, date)
	if err != nil {
		return 0.0, fmt.Errorf("failed to get revenue data: %w", err)
	}

	if len(revenueData) > 0 {
		revenueRecord := revenueData[0]

		// Check revenue source
		if revenueRecord.RevenueSource == "doctor" {
			// Calculate revenue from doctor assignments
			doctorAssignments, err := e.repos.DoctorAssignment.GetDoctorsByBranchAndDate(branchID, date)
			if err != nil {
				return 0.0, fmt.Errorf("failed to get doctor assignments: %w", err)
			}

			revenue = 0.0
			for _, assignment := range doctorAssignments {
				revenue += assignment.ExpectedRevenue
			}
		} else {
			// Use branch configuration revenue
			revenue = revenueRecord.ExpectedRevenue
		}
	} else {
		// No revenue data - try to calculate from doctors if any are assigned
		doctorAssignments, err := e.repos.DoctorAssignment.GetDoctorsByBranchAndDate(branchID, date)
		if err == nil && len(doctorAssignments) > 0 {
			revenue = 0.0
			for _, assignment := range doctorAssignments {
				revenue += assignment.ExpectedRevenue
			}
		} else {
			return 0.0, nil // No revenue data = 0 score
		}
	}

	// Normalize revenue (0-1 scale)
	// TODO: Make normalization configurable via criteria config
	// For now, use a simple threshold-based normalization
	// Assuming max expected revenue is 100000, normalize to 0-1
	maxRevenue := 100000.0
	if revenue > maxRevenue {
		return 1.0, nil
	}
	return revenue / maxRevenue, nil
}

// evaluateMinStaffPosition evaluates minimum staff per position criterion
func (e *CriteriaEngine) evaluateMinStaffPosition(criteria *models.AllocationCriteria, branchID uuid.UUID, date time.Time) (float64, error) {
	// Get position quota for the branch
	quotas, err := e.repos.PositionQuota.GetByBranchID(branchID)
	if err != nil {
		return 0.0, fmt.Errorf("failed to get position quotas: %w", err)
	}

	if len(quotas) == 0 {
		return 0.0, nil // No quotas = 0 score
	}

	// Calculate average fulfillment rate across all positions
	totalFulfillment := 0.0
	count := 0

	for _, quota := range quotas {
		if !quota.IsActive {
			continue
		}

		// Get actual staff count for this position on this date
		// Count branch staff + rotation staff assigned
		branchStaff, err := e.repos.Staff.GetByBranchID(branchID)
		if err != nil {
			continue
		}

		actualCount := 0
		for _, staff := range branchStaff {
			if staff.PositionID == quota.PositionID {
				// Check if staff is working on this date
				schedules, err := e.repos.Schedule.GetByStaffID(staff.ID, date, date)
				if err == nil && len(schedules) > 0 {
					if schedules[0].ScheduleStatus == models.ScheduleStatusWorking {
						actualCount++
					}
				}
			}
		}

		// Count rotation staff assigned
		rotationAssignments, err := e.repos.Rotation.GetByBranchID(branchID, date, date)
		if err == nil {
			for _, assignment := range rotationAssignments {
				staff, err := e.repos.Staff.GetByID(assignment.RotationStaffID)
				if err == nil && staff != nil && staff.PositionID == quota.PositionID {
					actualCount++
				}
			}
		}

		// Calculate fulfillment rate (0-1)
		minRequired := quota.MinimumRequired
		if minRequired == 0 {
			minRequired = 1 // Avoid division by zero
		}
		fulfillment := float64(actualCount) / float64(minRequired)
		if fulfillment > 1.0 {
			fulfillment = 1.0
		}

		totalFulfillment += fulfillment
		count++
	}

	if count == 0 {
		return 0.0, nil
	}

	return totalFulfillment / float64(count), nil
}

// evaluateMinStaffBranch evaluates minimum staff per branch criterion
func (e *CriteriaEngine) evaluateMinStaffBranch(criteria *models.AllocationCriteria, branchID uuid.UUID, date time.Time) (float64, error) {
	// Get all quotas for the branch
	quotas, err := e.repos.PositionQuota.GetByBranchID(branchID)
	if err != nil {
		return 0.0, fmt.Errorf("failed to get position quotas: %w", err)
	}

	if len(quotas) == 0 {
		return 0.0, nil
	}

	// Calculate total minimum required
	totalMinRequired := 0
	for _, quota := range quotas {
		if quota.IsActive {
			totalMinRequired += quota.MinimumRequired
		}
	}

	if totalMinRequired == 0 {
		return 0.0, nil
	}

	// Get actual staff count (branch + rotation)
	branchStaff, err := e.repos.Staff.GetByBranchID(branchID)
	if err != nil {
		return 0.0, fmt.Errorf("failed to get branch staff: %w", err)
	}

	actualCount := 0
	for _, staff := range branchStaff {
		schedules, err := e.repos.Schedule.GetByStaffID(staff.ID, date, date)
		if err == nil && len(schedules) > 0 {
			if schedules[0].ScheduleStatus == models.ScheduleStatusWorking {
				actualCount++
			}
		}
	}

	// Count rotation staff
	rotationAssignments, err := e.repos.Rotation.GetByBranchID(branchID, date, date)
	if err == nil {
		actualCount += len(rotationAssignments)
	}

	// Calculate fulfillment rate
	fulfillment := float64(actualCount) / float64(totalMinRequired)
	if fulfillment > 1.0 {
		fulfillment = 1.0
	}

	return fulfillment, nil
}

// evaluateDoctorCount evaluates doctor count criterion
func (e *CriteriaEngine) evaluateDoctorCount(criteria *models.AllocationCriteria, branchID uuid.UUID, date time.Time) (float64, error) {
	// Get doctor count for the branch on the date
	doctorCount, err := e.repos.DoctorAssignment.GetDoctorCountByBranch(branchID, date)
	if err != nil {
		return 0.0, fmt.Errorf("failed to get doctor count: %w", err)
	}

	// Normalize doctor count (0-1 scale)
	// Maximum doctors per branch per day is 6
	maxDoctors := 6.0
	if doctorCount > int(maxDoctors) {
		return 1.0, nil
	}
	return float64(doctorCount) / maxDoctors, nil
}

// evaluateDoctorSpecificStaff evaluates doctor-specific staff requirement criterion
func (e *CriteriaEngine) evaluateDoctorSpecificStaff(criteria *models.AllocationCriteria, branchID uuid.UUID, date time.Time) (float64, error) {
	// Get doctors assigned to this branch on this date
	doctorAssignments, err := e.repos.DoctorAssignment.GetDoctorsByBranchAndDate(branchID, date)
	if err != nil {
		return 0.0, fmt.Errorf("failed to get doctor assignments: %w", err)
	}

	if len(doctorAssignments) == 0 {
		return 0.0, nil // No doctors = no doctor-specific requirements
	}

	// Get all active doctor preferences for these doctors
	totalFulfillment := 0.0
	requirementCount := 0

	for _, assignment := range doctorAssignments {
		// Get doctor preferences (branch-specific or global)
		preferences, err := e.repos.DoctorPreference.GetActiveByDoctorID(assignment.DoctorID)
		if err != nil {
			continue
		}

		for _, preference := range preferences {
			// Check if preference applies to this branch
			if preference.BranchID != nil && *preference.BranchID != branchID {
				continue // Skip if preference is for a different branch
			}

			// Only process staff_requirement rules
			if preference.RuleType != "staff_requirement" {
				continue
			}

			// Extract requirements from rule_config
			requirements, ok := preference.RuleConfig["requirements"].([]interface{})
			if !ok {
				continue
			}

			for _, reqInterface := range requirements {
				req, ok := reqInterface.(map[string]interface{})
				if !ok {
					continue
				}

				positionIDStr, ok := req["position_id"].(string)
				if !ok {
					continue
				}
				positionID, err := uuid.Parse(positionIDStr)
				if err != nil {
					continue
				}

				minCount, ok := req["min_count"].(float64)
				if !ok {
					// Try int conversion
					minCountInt, ok := req["min_count"].(int)
					if !ok {
						continue
					}
					minCount = float64(minCountInt)
				}

				// Count actual staff for this position
				actualCount := 0
				branchStaff, err := e.repos.Staff.GetByBranchID(branchID)
				if err == nil {
					for _, staff := range branchStaff {
						if staff.PositionID == positionID {
							schedules, err := e.repos.Schedule.GetByStaffID(staff.ID, date, date)
							if err == nil && len(schedules) > 0 {
								if schedules[0].ScheduleStatus == models.ScheduleStatusWorking {
									actualCount++
								}
							}
						}
					}
				}

				// Count rotation staff
				rotationAssignments, err := e.repos.Rotation.GetByBranchID(branchID, date, date)
				if err == nil {
					for _, rotAssignment := range rotationAssignments {
						staff, err := e.repos.Staff.GetByID(rotAssignment.RotationStaffID)
						if err == nil && staff != nil && staff.PositionID == positionID {
							actualCount++
						}
					}
				}

				// Calculate fulfillment rate
				if minCount > 0 {
					fulfillment := float64(actualCount) / minCount
					if fulfillment > 1.0 {
						fulfillment = 1.0
					}
					totalFulfillment += fulfillment
					requirementCount++
				}
			}
		}
	}

	if requirementCount == 0 {
		return 1.0, nil // No doctor-specific requirements = perfect score
	}

	return totalFulfillment / float64(requirementCount), nil
}

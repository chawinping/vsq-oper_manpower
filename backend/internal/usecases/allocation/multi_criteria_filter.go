package allocation

import (
	"fmt"
	"math"
	"sort"
	"time"

	"vsq-oper-manpower/backend/internal/domain/models"

	"github.com/google/uuid"
)

// MultiCriteriaFilter implements the 5-criteria-group filtering system for rotation staff allocation
type MultiCriteriaFilter struct {
	repos *RepositoriesWrapper
}

// NewMultiCriteriaFilter creates a new multi-criteria filter
func NewMultiCriteriaFilter(repos *RepositoriesWrapper) *MultiCriteriaFilter {
	return &MultiCriteriaFilter{repos: repos}
}

// CriteriaPriorityOrder represents the priority order of criteria (strict lexicographic ordering)
// Criteria are evaluated in order: PriorityOrder[0] is highest priority, PriorityOrder[len-1] is lowest
type CriteriaPriorityOrder struct {
	PriorityOrder []string `json:"priority_order"` // Array of criterion IDs in priority order (highest to lowest)
}

// Criterion ID constants
const (
	CriterionZeroth = "zeroth_criteria" // Doctor preferences
	CriterionFirst  = "first_criteria"  // Branch-level variables
	CriterionSecond = "second_criteria" // Preferred staff shortage
	CriterionThird  = "third_criteria"  // Minimum staff shortage
	CriterionFourth = "fourth_criteria" // Branch type staff groups
)

// DefaultCriteriaPriorityOrder returns default priority order for criteria groups
// Priority 1 (highest): Minimum staff shortage (third_criteria)
// Priority 2: Preferred staff shortage (second_criteria)
// Priority 3: Branch-level variables (first_criteria)
// Priority 4: Branch type staff groups (fourth_criteria)
// Priority 5 (lowest): Doctor preferences (zeroth_criteria)
func DefaultCriteriaPriorityOrder() CriteriaPriorityOrder {
	return CriteriaPriorityOrder{
		PriorityOrder: []string{
			CriterionThird,  // Priority 1: Minimum staff shortage (highest)
			CriterionSecond, // Priority 2: Preferred staff shortage
			CriterionFirst,  // Priority 3: Branch-level variables
			CriterionFourth, // Priority 4: Branch type staff groups
			CriterionZeroth, // Priority 5: Doctor preferences (lowest)
		},
	}
}

// AllocationSuggestion represents a ranked suggestion for rotation staff allocation
type AllocationSuggestion struct {
	BranchID           uuid.UUID         `json:"branch_id"`
	BranchName         string            `json:"branch_name"`
	BranchCode         string            `json:"branch_code"`
	PositionID         uuid.UUID         `json:"position_id"`
	PositionName       string            `json:"position_name"`
	Date               time.Time         `json:"date"`
	PriorityScore      float64           `json:"priority_score"`
	Reason             string            `json:"reason"`
	SuggestedStaffID   *uuid.UUID        `json:"suggested_staff_id,omitempty"`
	SuggestedStaffName string            `json:"suggested_staff_name,omitempty"`
	CriteriaBreakdown  CriteriaBreakdown `json:"criteria_breakdown"`
}

// CriteriaBreakdown shows the contribution of each criteria group to the final score
type CriteriaBreakdown struct {
	ZerothCriteriaScore float64 `json:"zeroth_criteria_score,omitempty"`
	FirstCriteriaScore  float64 `json:"first_criteria_score"`
	SecondCriteriaScore float64 `json:"second_criteria_score"`
	ThirdCriteriaScore  float64 `json:"third_criteria_score"`
	FourthCriteriaScore float64 `json:"fourth_criteria_score"`
}

// GenerateRankedSuggestions generates ranked suggestions for rotation staff allocation
// based on strict priority ordering (lexicographic sorting)
func (f *MultiCriteriaFilter) GenerateRankedSuggestions(
	branchIDs []uuid.UUID,
	date time.Time,
	priorityOrder CriteriaPriorityOrder,
	enableDoctorPreferences bool,
) ([]*AllocationSuggestion, error) {

	// Step 1: Apply zeroth criteria (doctor preferences) - optional filter
	var filteredBranchIDs []uuid.UUID
	if enableDoctorPreferences {
		// Check if zeroth criteria is in priority order
		hasZeroth := false
		for _, criterionID := range priorityOrder.PriorityOrder {
			if criterionID == CriterionZeroth {
				hasZeroth = true
				break
			}
		}
		if hasZeroth {
			var err error
			filteredBranchIDs, err = f.applyZerothCriteria(branchIDs, date)
			if err != nil {
				return nil, fmt.Errorf("failed to apply zeroth criteria: %w", err)
			}
		} else {
			filteredBranchIDs = branchIDs
		}
	} else {
		filteredBranchIDs = branchIDs
	}

	// Step 2: Evaluate all criteria groups for each branch-position combination
	suggestions := []*AllocationSuggestion{}

	for _, branchID := range filteredBranchIDs {
		// Get branch info
		branch, err := f.repos.Branch.GetByID(branchID)
		if err != nil {
			continue
		}
		if branch == nil {
			continue
		}

		// Check if branch is operational (has at least one doctor)
		doctorCount, err := f.repos.DoctorAssignment.GetDoctorCountByBranch(branchID, date)
		if err != nil {
			continue
		}
		if doctorCount == 0 {
			continue // Skip closed branches
		}

		// Get position quotas for this branch
		quotas, err := f.repos.PositionQuota.GetByBranchID(branchID)
		if err != nil {
			continue
		}

		// Get all positions (only branch-type positions can have quotas)
		positions, err := f.repos.Position.List()
		if err != nil {
			continue
		}

		// Evaluate for each position that has a quota
		for _, quota := range quotas {
			if !quota.IsActive {
				continue
			}

			// Find position details
			var position *models.Position
			for _, p := range positions {
				if p.ID == quota.PositionID {
					position = p
					break
				}
			}
			if position == nil {
				continue
			}

			// Only process branch-type positions
			if position.PositionType != models.PositionTypeBranch {
				continue
			}

			// Calculate current staff count for this position
			currentStaffCount, err := f.calculateCurrentStaffCount(branchID, quota.PositionID, date)
			if err != nil {
				continue
			}

			// Calculate shortage
			preferredShortage := quota.DesignatedQuota - currentStaffCount
			minimumShortage := quota.MinimumRequired - currentStaffCount

			// Skip if no shortage
			if preferredShortage <= 0 && minimumShortage <= 0 {
				continue
			}

			// Evaluate all criteria groups
			criteriaBreakdown := CriteriaBreakdown{}

			// First criteria: Branch-level variables
			firstScore, err := f.evaluateFirstCriteria(branchID, date)
			if err != nil {
				continue
			}
			criteriaBreakdown.FirstCriteriaScore = firstScore

			// Second criteria: Preferred staff shortage
			secondScore := f.evaluateSecondCriteria(preferredShortage, quota.DesignatedQuota)
			criteriaBreakdown.SecondCriteriaScore = secondScore

			// Third criteria: Minimum staff shortage (critical priority)
			thirdScore := f.evaluateThirdCriteria(minimumShortage, quota.MinimumRequired)
			criteriaBreakdown.ThirdCriteriaScore = thirdScore

			// Fourth criteria: Branch type staff groups
			fourthScore, err := f.evaluateFourthCriteria(branchID, quota.PositionID, date)
			if err != nil {
				continue
			}
			criteriaBreakdown.FourthCriteriaScore = fourthScore

			// Zeroth criteria: Doctor preferences (if enabled)
			var zerothScore float64
			if enableDoctorPreferences {
				// Check if zeroth criteria is in priority order
				hasZeroth := false
				for _, criterionID := range priorityOrder.PriorityOrder {
					if criterionID == CriterionZeroth {
						hasZeroth = true
						break
					}
				}
				if hasZeroth {
					zerothScore, err = f.evaluateZerothCriteria(branchID, quota.PositionID, date)
					if err != nil {
						zerothScore = 0.0 // Default to 0 if evaluation fails
					}
					criteriaBreakdown.ZerothCriteriaScore = zerothScore
				}
			}

			// Generate reason
			reason := f.generateReason(criteriaBreakdown, preferredShortage, minimumShortage, quota, position)

			// Create a map of criterion scores for easy lookup
			criteriaScores := map[string]float64{
				CriterionZeroth: zerothScore,
				CriterionFirst:  firstScore,
				CriterionSecond: secondScore,
				CriterionThird:  thirdScore,
				CriterionFourth: fourthScore,
			}

			// Calculate priority score based on lexicographic ordering
			// Priority score is a composite value that ensures strict ordering
			// We use a weighted sum where higher priority criteria have exponentially more weight
			priorityScore := 0.0
			multiplier := 10000.0 // Start with high multiplier for first priority
			for _, criterionID := range priorityOrder.PriorityOrder {
				if score, exists := criteriaScores[criterionID]; exists {
					priorityScore += score * multiplier
					multiplier /= 10.0 // Each lower priority gets 10x less weight
				}
			}

			suggestion := &AllocationSuggestion{
				BranchID:          branchID,
				BranchName:        branch.Name,
				BranchCode:        branch.Code,
				PositionID:        quota.PositionID,
				PositionName:      position.Name,
				Date:              date,
				PriorityScore:     priorityScore,
				Reason:            reason,
				CriteriaBreakdown: criteriaBreakdown,
			}

			suggestions = append(suggestions, suggestion)
		}
	}

	// Sort by priority score using lexicographic ordering (highest first)
	// The priority score is constructed so that higher priority criteria dominate
	sort.Slice(suggestions, func(i, j int) bool {
		return suggestions[i].PriorityScore > suggestions[j].PriorityScore
	})

	return suggestions, nil
}

// applyZerothCriteria filters branches based on doctor preferences
// Returns only branches that meet doctor preference requirements
func (f *MultiCriteriaFilter) applyZerothCriteria(branchIDs []uuid.UUID, date time.Time) ([]uuid.UUID, error) {
	dayOfWeek := int(date.Weekday())
	filteredBranches := []uuid.UUID{}

	for _, branchID := range branchIDs {
		// Get doctors assigned to this branch on this date
		doctorAssignments, err := f.repos.DoctorAssignment.GetDoctorsByBranchAndDate(branchID, date)
		if err != nil {
			continue
		}

		meetsRequirements := true

		// Check each doctor's preferences
		for _, assignment := range doctorAssignments {
			// Get rotation staff requirement preferences for this doctor
			preferences, err := f.repos.DoctorPreference.GetRotationStaffRequirements(
				assignment.DoctorID,
				&branchID,
				&dayOfWeek,
			)
			if err != nil {
				continue
			}

			// Check if preferences are met
			for _, preference := range preferences {
				if !f.checkDoctorPreferenceMet(preference, branchID, date) {
					meetsRequirements = false
					break
				}
			}

			if !meetsRequirements {
				break
			}
		}

		if meetsRequirements {
			filteredBranches = append(filteredBranches, branchID)
		}
	}

	return filteredBranches, nil
}

// checkDoctorPreferenceMet checks if a doctor preference requirement is met
func (f *MultiCriteriaFilter) checkDoctorPreferenceMet(
	preference *models.DoctorPreference,
	branchID uuid.UUID,
	date time.Time,
) bool {
	// Check position requirements
	if positionReqs, ok := preference.RuleConfig["position_requirements"].([]interface{}); ok {
		for _, req := range positionReqs {
			if reqMap, ok := req.(map[string]interface{}); ok {
				positionIDStr, ok := reqMap["position_id"].(string)
				if !ok {
					continue
				}
				positionID, err := uuid.Parse(positionIDStr)
				if err != nil {
					continue
				}
				minCount, ok := reqMap["min_count"].(float64)
				if !ok {
					continue
				}

				// Check current staff count
				currentCount, err := f.calculateCurrentStaffCount(branchID, positionID, date)
				if err != nil {
					return false
				}

				if currentCount < int(minCount) {
					return false // Requirement not met
				}
			}
		}
	}

	// Check specific staff requirements for this date
	if specificReqs, ok := preference.RuleConfig["specific_staff_requirements"].([]interface{}); ok {
		dateStr := date.Format("2006-01-02")
		for _, req := range specificReqs {
			if reqMap, ok := req.(map[string]interface{}); ok {
				reqDate, ok := reqMap["date"].(string)
				if !ok || reqDate != dateStr {
					continue
				}

				// Check if specific staff is assigned
				staffIDStr, ok := reqMap["staff_id"].(string)
				if !ok {
					continue
				}
				staffID, err := uuid.Parse(staffIDStr)
				if err != nil {
					continue
				}

				// Check if this staff is assigned to the branch on this date
				assignments, err := f.repos.Rotation.GetByBranchID(branchID, date, date)
				if err != nil {
					return false
				}

				found := false
				for _, assignment := range assignments {
					if assignment.RotationStaffID == staffID {
						found = true
						break
					}
				}

				if !found {
					return false // Specific staff requirement not met
				}
			}
		}
	}

	return true
}

// evaluateZerothCriteria evaluates doctor preferences criteria
func (f *MultiCriteriaFilter) evaluateZerothCriteria(
	branchID uuid.UUID,
	positionID uuid.UUID,
	date time.Time,
) (float64, error) {
	dayOfWeek := int(date.Weekday())

	// Get doctors assigned to this branch on this date
	doctorAssignments, err := f.repos.DoctorAssignment.GetDoctorsByBranchAndDate(branchID, date)
	if err != nil {
		return 0.0, err
	}

	if len(doctorAssignments) == 0 {
		return 0.0, nil // No doctors, no preference score
	}

	totalScore := 0.0
	count := 0

	// Check each doctor's preferences
	for _, assignment := range doctorAssignments {
		preferences, err := f.repos.DoctorPreference.GetRotationStaffRequirements(
			assignment.DoctorID,
			&branchID,
			&dayOfWeek,
		)
		if err != nil {
			continue
		}

		for _, preference := range preferences {
			// Check if this position is required by doctor preference
			if positionReqs, ok := preference.RuleConfig["position_requirements"].([]interface{}); ok {
				for _, req := range positionReqs {
					if reqMap, ok := req.(map[string]interface{}); ok {
						reqPositionIDStr, ok := reqMap["position_id"].(string)
						if !ok {
							continue
						}
						reqPositionID, err := uuid.Parse(reqPositionIDStr)
						if err != nil {
							continue
						}

						if reqPositionID == positionID {
							// This position is required by doctor preference
							// Score based on how well it's met
							minCount, ok := reqMap["min_count"].(float64)
							if !ok {
								continue
							}

							currentCount, err := f.calculateCurrentStaffCount(branchID, positionID, date)
							if err != nil {
								continue
							}

							// Score: 1.0 if requirement met, decreasing if not met
							if currentCount >= int(minCount) {
								totalScore += 1.0
							} else {
								// Partial score based on fulfillment ratio
								totalScore += float64(currentCount) / minCount
							}
							count++
						}
					}
				}
			}
		}
	}

	if count == 0 {
		return 0.5, nil // Neutral score if no doctor preferences for this position
	}

	return totalScore / float64(count), nil
}

// evaluateFirstCriteria evaluates branch-level variables (universal across branches)
func (f *MultiCriteriaFilter) evaluateFirstCriteria(branchID uuid.UUID, date time.Time) (float64, error) {
	// Get revenue data for this branch and date
	revenueData, err := f.repos.Revenue.GetByBranchID(branchID, date, date)
	if err != nil {
		return 0.0, err
	}

	var skinRevenue, laserYagRevenue float64
	var vitaminCases, slimPenCases int

	if len(revenueData) > 0 {
		rd := revenueData[0]
		skinRevenue = rd.SkinRevenue
		laserYagRevenue = rd.LSHMRevenue
		vitaminCases = rd.VitaminCases
		slimPenCases = rd.SlimPenCases
	}

	// Get doctor count
	doctorCount, err := f.repos.DoctorAssignment.GetDoctorCountByBranch(branchID, date)
	if err != nil {
		return 0.0, err
	}

	// Normalize each variable to 0-1 scale
	// We need to get max values for normalization - for now, use reasonable defaults
	maxSkinRevenue := 1000000.0    // 1M THB
	maxLaserYagRevenue := 500000.0 // 500K THB
	maxVitaminCases := 50
	maxSlimPenCases := 30
	maxDoctorCount := 6

	skinScore := math.Min(skinRevenue/maxSkinRevenue, 1.0)
	laserScore := math.Min(laserYagRevenue/maxLaserYagRevenue, 1.0)
	vitaminScore := math.Min(float64(vitaminCases)/float64(maxVitaminCases), 1.0)
	slimPenScore := math.Min(float64(slimPenCases)/float64(maxSlimPenCases), 1.0)
	doctorScore := math.Min(float64(doctorCount)/float64(maxDoctorCount), 1.0)

	// Combine scores with equal weights
	combinedScore := (skinScore + laserScore + vitaminScore + slimPenScore + doctorScore) / 5.0

	return combinedScore, nil
}

// evaluateSecondCriteria evaluates preferred staff shortage
func (f *MultiCriteriaFilter) evaluateSecondCriteria(shortage int, preferred int) float64 {
	if preferred == 0 {
		return 0.0
	}

	if shortage <= 0 {
		return 0.0 // No shortage, no priority
	}

	// Score increases with shortage ratio
	shortageRatio := float64(shortage) / float64(preferred)
	return math.Min(shortageRatio, 1.0)
}

// evaluateThirdCriteria evaluates minimum staff shortage (critical priority)
func (f *MultiCriteriaFilter) evaluateThirdCriteria(shortage int, minimum int) float64 {
	if minimum == 0 {
		return 0.0
	}

	if shortage <= 0 {
		return 0.0 // No shortage
	}

	// Critical priority: score is 1.0 if below minimum, decreasing as we approach minimum
	// This ensures positions below minimum get highest priority
	if shortage > 0 {
		return 1.0 // Maximum priority for positions below minimum
	}

	return 0.0
}

// evaluateFourthCriteria evaluates branch type staff group requirements
func (f *MultiCriteriaFilter) evaluateFourthCriteria(
	branchID uuid.UUID,
	positionID uuid.UUID,
	date time.Time,
) (float64, error) {
	// Get branch
	branch, err := f.repos.Branch.GetByID(branchID)
	if err != nil {
		return 0.0, err
	}
	if branch == nil || branch.BranchTypeID == nil {
		return 0.5, nil // Neutral score if no branch type assigned
	}

	// Get branch type requirements
	requirements, err := f.repos.BranchTypeRequirement.GetByBranchTypeID(*branch.BranchTypeID)
	if err != nil {
		return 0.5, nil
	}

	// Find which staff group this position belongs to
	staffGroupPositions, err := f.repos.StaffGroupPosition.GetByPositionID(positionID)
	if err != nil {
		return 0.5, nil
	}

	if len(staffGroupPositions) == 0 {
		return 0.5, nil // Position not in any staff group
	}

	// Check each staff group this position belongs to
	maxScore := 0.0
	for _, sgp := range staffGroupPositions {
		// Find requirement for this staff group
		for _, req := range requirements {
			if req.StaffGroupID == sgp.StaffGroupID && req.IsActive {
				// Calculate current staff count for this staff group
				currentCount, err := f.calculateStaffGroupCount(branchID, sgp.StaffGroupID, date)
				if err != nil {
					continue
				}

				// Score based on shortage
				if currentCount < req.MinimumStaffCount {
					shortage := req.MinimumStaffCount - currentCount
					// Maximum priority if below minimum
					score := 1.0
					if req.MinimumStaffCount > 0 {
						// Normalize by minimum requirement
						score = math.Min(float64(shortage)/float64(req.MinimumStaffCount), 1.0)
					}
					if score > maxScore {
						maxScore = score
					}
				}
			}
		}
	}

	if maxScore == 0.0 {
		return 0.5, nil // Neutral score if requirements are met
	}

	return maxScore, nil
}

// calculateCurrentStaffCount calculates current staff count (branch + rotation) for a position
func (f *MultiCriteriaFilter) calculateCurrentStaffCount(
	branchID uuid.UUID,
	positionID uuid.UUID,
	date time.Time,
) (int, error) {
	// Get branch staff for this position
	branchStaff, err := f.repos.Staff.GetByBranchID(branchID)
	if err != nil {
		return 0, err
	}

	branchCount := 0
	for _, staff := range branchStaff {
		if staff.PositionID == positionID {
			// Check if staff is working on this date
			schedules, err := f.repos.Schedule.GetByStaffID(staff.ID, date, date)
			if err != nil {
				continue
			}
			for _, schedule := range schedules {
				if schedule.ScheduleStatus == models.ScheduleStatusWorking {
					branchCount++
					break
				}
			}
		}
	}

	// Get rotation assignments for this position
	rotationAssignments, err := f.repos.Rotation.GetByBranchID(branchID, date, date)
	if err != nil {
		return branchCount, nil
	}

	rotationCount := 0
	for _, assignment := range rotationAssignments {
		// Get rotation staff
		staff, err := f.repos.Staff.GetByID(assignment.RotationStaffID)
		if err != nil {
			continue
		}
		if staff != nil && staff.PositionID == positionID {
			rotationCount++
		}
	}

	return branchCount + rotationCount, nil
}

// calculateStaffGroupCount calculates current staff count for a staff group
func (f *MultiCriteriaFilter) calculateStaffGroupCount(
	branchID uuid.UUID,
	staffGroupID uuid.UUID,
	date time.Time,
) (int, error) {
	// Get all positions in this staff group
	positions, err := f.repos.StaffGroupPosition.GetByStaffGroupID(staffGroupID)
	if err != nil {
		return 0, err
	}

	totalCount := 0
	for _, sgp := range positions {
		count, err := f.calculateCurrentStaffCount(branchID, sgp.PositionID, date)
		if err != nil {
			continue
		}
		totalCount += count
	}

	return totalCount, nil
}

// generateReason generates a human-readable reason for the suggestion
func (f *MultiCriteriaFilter) generateReason(
	breakdown CriteriaBreakdown,
	preferredShortage int,
	minimumShortage int,
	quota *models.PositionQuota,
	position *models.Position,
) string {
	reasons := []string{}

	if minimumShortage > 0 {
		reasons = append(reasons, fmt.Sprintf("Below minimum requirement (%d staff needed)", minimumShortage))
	}

	if preferredShortage > 0 && minimumShortage <= 0 {
		reasons = append(reasons, fmt.Sprintf("Below preferred quota (%d staff needed)", preferredShortage))
	}

	if breakdown.FirstCriteriaScore > 0.7 {
		reasons = append(reasons, "High branch activity/revenue")
	}

	if breakdown.FourthCriteriaScore > 0.7 {
		reasons = append(reasons, "Staff group requirement not met")
	}

	if breakdown.ZerothCriteriaScore > 0 {
		reasons = append(reasons, "Doctor preference requirement")
	}

	if len(reasons) == 0 {
		return "General allocation need"
	}

	result := reasons[0]
	for i := 1; i < len(reasons); i++ {
		result += "; " + reasons[i]
	}

	return result
}

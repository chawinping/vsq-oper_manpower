package allocation

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"vsq-oper-manpower/backend/internal/domain/models"
)

// SuggestionEngine generates allocation suggestions based on criteria and quota
type SuggestionEngine struct {
	repos            *RepositoriesWrapper
	criteriaEngine  *CriteriaEngine
	quotaCalculator *QuotaCalculator
}

// NewSuggestionEngine creates a new suggestion engine
func NewSuggestionEngine(repos *RepositoriesWrapper, criteriaEngine *CriteriaEngine, quotaCalculator *QuotaCalculator) *SuggestionEngine {
	return &SuggestionEngine{
		repos:           repos,
		criteriaEngine:  criteriaEngine,
		quotaCalculator: quotaCalculator,
	}
}

// GenerateSuggestions generates allocation suggestions for branches in a date range
func (e *SuggestionEngine) GenerateSuggestions(branchIDs []uuid.UUID, startDate, endDate time.Time) ([]*models.AllocationSuggestion, error) {
	suggestions := []*models.AllocationSuggestion{}
	currentDate := startDate

	for !currentDate.After(endDate) {
		// Generate suggestions for each branch on this date
		for _, branchID := range branchIDs {
			branchSuggestions, err := e.generateSuggestionsForBranch(branchID, currentDate)
			if err != nil {
				return nil, fmt.Errorf("failed to generate suggestions for branch %s on %s: %w", branchID, currentDate.Format("2006-01-02"), err)
			}
			suggestions = append(suggestions, branchSuggestions...)
		}
		currentDate = currentDate.AddDate(0, 0, 1)
	}

	return suggestions, nil
}

// generateSuggestionsForBranch generates suggestions for a specific branch on a specific date
func (e *SuggestionEngine) generateSuggestionsForBranch(branchID uuid.UUID, date time.Time) ([]*models.AllocationSuggestion, error) {
	// Check if branch is operational (has at least one doctor)
	// Branch operational status is determined by doctor assignments
	doctorCount, err := e.repos.DoctorAssignment.GetDoctorCountByBranch(branchID, date)
	if err != nil {
		return nil, fmt.Errorf("failed to get doctor count: %w", err)
	}

	// Skip branches with no doctors (closed branches don't need rotation staff)
	if doctorCount == 0 {
		return []*models.AllocationSuggestion{}, nil
	}

	// Get quota status for the branch
	quotaStatus, err := e.quotaCalculator.CalculateBranchQuotaStatus(branchID, date)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate quota status: %w", err)
	}

	// Evaluate criteria for the branch
	allocationScore, err := e.criteriaEngine.EvaluateCriteria(branchID, date)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate criteria: %w", err)
	}

	suggestions := []*models.AllocationSuggestion{}

	// Generate suggestions for positions that need staff
	for _, positionStatus := range quotaStatus.PositionStatuses {
		if positionStatus.StillRequired <= 0 {
			continue // No need for additional staff
		}

		// Find eligible rotation staff for this position
		eligibleStaff, err := e.findEligibleRotationStaff(branchID, positionStatus.PositionID, date)
		if err != nil {
			continue // Skip if we can't find eligible staff
		}

		// Generate suggestions based on how many staff are needed
		for i := 0; i < positionStatus.StillRequired && i < len(eligibleStaff); i++ {
			staff := eligibleStaff[i]

			// Calculate confidence based on criteria score and staff availability
			confidence := e.calculateConfidence(allocationScore, staff, branchID, date)

			// Generate reason
			reason := e.generateReason(allocationScore, positionStatus, staff)

			suggestion := &models.AllocationSuggestion{
				ID:              uuid.New(),
				RotationStaffID: staff.ID,
				BranchID:        branchID,
				Date:            date,
				PositionID:      positionStatus.PositionID,
				Status:          models.SuggestionStatusPending,
				Confidence:      confidence,
				Reason:          reason,
				CriteriaUsed:    e.getCriteriaUsedJSON(allocationScore),
			}

			suggestions = append(suggestions, suggestion)
		}
	}

	return suggestions, nil
}

// findEligibleRotationStaff finds rotation staff eligible for assignment to a branch for a specific position
func (e *SuggestionEngine) findEligibleRotationStaff(branchID uuid.UUID, positionID uuid.UUID, date time.Time) ([]*models.Staff, error) {
	// Get effective branches for this branch (rotation staff eligible for this branch)
	effectiveBranches, err := e.repos.EffectiveBranch.GetByBranchID(branchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get effective branches: %w", err)
	}

	// Get all rotation staff
	allRotationStaff, err := e.repos.Staff.GetRotationStaff()
	if err != nil {
		return nil, fmt.Errorf("failed to get rotation staff: %w", err)
	}

	// Filter to eligible staff with matching position
	eligibleStaff := []*models.Staff{}
	rotationStaffMap := make(map[uuid.UUID]bool)

	for _, eb := range effectiveBranches {
		rotationStaffMap[eb.RotationStaffID] = true
	}

	for _, staff := range allRotationStaff {
		// Check if staff is eligible for this branch
		if !rotationStaffMap[staff.ID] {
			continue
		}

		// Check if staff matches the position
		if staff.PositionID != positionID {
			continue
		}

		// Check if staff is available (not already assigned on this date)
		isAvailable, err := e.isStaffAvailable(staff.ID, date)
		if err != nil || !isAvailable {
			continue
		}

		eligibleStaff = append(eligibleStaff, staff)
	}

	// Sort by priority (Level 1 first, then Level 2)
	// TODO: Add more sophisticated sorting based on criteria scores

	return eligibleStaff, nil
}

// isStaffAvailable checks if rotation staff is available on a specific date
func (e *SuggestionEngine) isStaffAvailable(staffID uuid.UUID, date time.Time) (bool, error) {
	// Check if staff is already assigned on this date
	assignments, err := e.repos.Rotation.GetByRotationStaffID(staffID, date, date)
	if err != nil {
		return false, err
	}

	if len(assignments) > 0 {
		return false, nil // Already assigned
	}

	// Check rotation staff schedule (if they have a schedule entry)
	// For now, we assume rotation staff are available unless assigned
	// TODO: Add rotation staff schedule checking

	return true, nil
}

// calculateConfidence calculates confidence score for a suggestion
func (e *SuggestionEngine) calculateConfidence(score *AllocationScore, staff *models.Staff, branchID uuid.UUID, date time.Time) float64 {
	// Base confidence from overall allocation score
	confidence := score.OverallScore

	// Adjust based on staff skill level (if available)
	if staff.SkillLevel > 0 {
		skillBonus := float64(staff.SkillLevel) / 10.0 * 0.1 // Up to 10% bonus
		confidence += skillBonus
	}

	// Ensure confidence is between 0 and 1
	if confidence > 1.0 {
		confidence = 1.0
	}
	if confidence < 0.0 {
		confidence = 0.0
	}

	return confidence
}

// generateReason generates a human-readable reason for the suggestion
func (e *SuggestionEngine) generateReason(score *AllocationScore, positionStatus PositionQuotaStatus, staff *models.Staff) string {
	reason := fmt.Sprintf("Position '%s' requires %d more staff. ", positionStatus.PositionName, positionStatus.StillRequired)
	
	if score.OverallScore > 0.7 {
		reason += "High allocation score indicates strong need. "
	} else if score.OverallScore > 0.4 {
		reason += "Moderate allocation score. "
	} else {
		reason += "Low allocation score but minimum requirements not met. "
	}

	reason += fmt.Sprintf("Staff '%s' is eligible and available.", staff.Name)
	return reason
}

// getCriteriaUsedJSON returns JSON string of criteria IDs used
func (e *SuggestionEngine) getCriteriaUsedJSON(score *AllocationScore) string {
	// For now, return empty string
	// TODO: Track which specific criteria were used and return their IDs
	return "[]"
}

// ApproveSuggestion approves a suggestion and creates a rotation assignment
func (e *SuggestionEngine) ApproveSuggestion(suggestionID uuid.UUID, userID uuid.UUID) error {
	suggestion, err := e.repos.AllocationSuggestion.GetByID(suggestionID)
	if err != nil {
		return fmt.Errorf("failed to get suggestion: %w", err)
	}
	if suggestion == nil {
		return fmt.Errorf("suggestion not found")
	}

	if suggestion.Status != models.SuggestionStatusPending {
		return fmt.Errorf("suggestion is not pending")
	}

	// Create rotation assignment
	assignment := &models.RotationAssignment{
		ID:              uuid.New(),
		RotationStaffID: suggestion.RotationStaffID,
		BranchID:        suggestion.BranchID,
		Date:            suggestion.Date,
		AssignmentLevel: 1, // Default to Level 1, can be made configurable
		IsAdhoc:         false,
		AssignedBy:      userID,
	}

	if err := e.repos.Rotation.Create(assignment); err != nil {
		return fmt.Errorf("failed to create rotation assignment: %w", err)
	}

	// Update suggestion status
	suggestion.Status = models.SuggestionStatusApproved
	suggestion.ReviewedBy = &userID
	now := time.Now()
	suggestion.ReviewedAt = &now

	if err := e.repos.AllocationSuggestion.Update(suggestion); err != nil {
		return fmt.Errorf("failed to update suggestion: %w", err)
	}

	return nil
}

// RejectSuggestion rejects a suggestion
func (e *SuggestionEngine) RejectSuggestion(suggestionID uuid.UUID, userID uuid.UUID) error {
	suggestion, err := e.repos.AllocationSuggestion.GetByID(suggestionID)
	if err != nil {
		return fmt.Errorf("failed to get suggestion: %w", err)
	}
	if suggestion == nil {
		return fmt.Errorf("suggestion not found")
	}

	if suggestion.Status != models.SuggestionStatusPending {
		return fmt.Errorf("suggestion is not pending")
	}

	suggestion.Status = models.SuggestionStatusRejected
	suggestion.ReviewedBy = &userID
	now := time.Now()
	suggestion.ReviewedAt = &now

	if err := e.repos.AllocationSuggestion.Update(suggestion); err != nil {
		return fmt.Errorf("failed to update suggestion: %w", err)
	}

	return nil
}

package allocation

import (
	"encoding/json"
	"fmt"
	"time"

	"vsq-oper-manpower/backend/internal/domain/models"

	"github.com/google/uuid"
)

// SuggestionEngine generates allocation suggestions based on criteria and quota
type SuggestionEngine struct {
	repos               *RepositoriesWrapper
	multiCriteriaFilter *MultiCriteriaFilter
	quotaCalculator     *QuotaCalculator
}

// NewSuggestionEngine creates a new suggestion engine
func NewSuggestionEngine(repos *RepositoriesWrapper, multiCriteriaFilter *MultiCriteriaFilter, quotaCalculator *QuotaCalculator) *SuggestionEngine {
	return &SuggestionEngine{
		repos:               repos,
		multiCriteriaFilter: multiCriteriaFilter,
		quotaCalculator:     quotaCalculator,
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

	// Get criteria priority order from settings
	priorityOrder, enableDoctorPrefs, err := e.getCriteriaPriorityOrder()
	if err != nil {
		// Use defaults if settings not found
		priorityOrder = DefaultCriteriaPriorityOrder()
		enableDoctorPrefs = false
	}

	// Use MultiCriteriaFilter to generate ranked suggestions
	multiSuggestions, err := e.multiCriteriaFilter.GenerateRankedSuggestions(
		[]uuid.UUID{branchID},
		date,
		priorityOrder,
		enableDoctorPrefs,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ranked suggestions: %w", err)
	}

	// Convert MultiCriteriaFilter suggestions to models.AllocationSuggestion
	suggestions := []*models.AllocationSuggestion{}
	for _, ms := range multiSuggestions {
		// Find eligible rotation staff for this position
		eligibleStaff, err := e.findEligibleRotationStaff(ms.BranchID, ms.PositionID, date)
		if err != nil || len(eligibleStaff) == 0 {
			continue // Skip if no eligible staff
		}

		// Use the first eligible staff member
		staff := eligibleStaff[0]

		// Convert criteria breakdown to JSON
		criteriaJSON, _ := json.Marshal(ms.CriteriaBreakdown)

		suggestion := &models.AllocationSuggestion{
			ID:              uuid.New(),
			RotationStaffID: staff.ID,
			BranchID:        ms.BranchID,
			Date:            ms.Date,
			PositionID:      ms.PositionID,
			Status:          models.SuggestionStatusPending,
			Confidence:      ms.PriorityScore,
			Reason:          ms.Reason,
			CriteriaUsed:    string(criteriaJSON),
		}

		suggestions = append(suggestions, suggestion)
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

		// Check if staff matches the position directly OR via mapping
		matchesPosition := staff.PositionID == positionID
		if !matchesPosition {
			// Check for staff-to-position mapping
			mapping, err := e.repos.RotationStaffBranchPosition.GetByStaffAndPosition(staff.ID, positionID)
			if err == nil && mapping != nil && mapping.IsActive {
				matchesPosition = true
				// Note: substitution level can be used for priority adjustment later
			}
		}
		if !matchesPosition {
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

// getCriteriaPriorityOrder retrieves criteria priority order from settings
func (e *SuggestionEngine) getCriteriaPriorityOrder() (CriteriaPriorityOrder, bool, error) {
	// Get priority order setting
	priorityOrderSetting, err := e.repos.Settings.GetByKey("allocation_criteria_priority_order")
	if err != nil || priorityOrderSetting == nil {
		return CriteriaPriorityOrder{}, false, fmt.Errorf("priority order setting not found")
	}

	var priorityOrder CriteriaPriorityOrder
	if err := json.Unmarshal([]byte(priorityOrderSetting.Value), &priorityOrder); err != nil {
		return CriteriaPriorityOrder{}, false, fmt.Errorf("failed to parse priority order: %w", err)
	}

	// Validate priority order
	if len(priorityOrder.PriorityOrder) == 0 {
		return CriteriaPriorityOrder{}, false, fmt.Errorf("priority order is empty")
	}

	// Get doctor preferences setting
	doctorPrefSetting, _ := e.repos.Settings.GetByKey("allocation_enable_doctor_preferences")
	enableDoctorPrefs := false
	if doctorPrefSetting != nil && doctorPrefSetting.Value == "true" {
		enableDoctorPrefs = true
	}

	return priorityOrder, enableDoctorPrefs, nil
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

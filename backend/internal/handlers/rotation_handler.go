package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"vsq-oper-manpower/backend/internal/config"
	"vsq-oper-manpower/backend/internal/domain/interfaces"
	"vsq-oper-manpower/backend/internal/domain/models"
	"vsq-oper-manpower/backend/internal/repositories/postgres"
	"vsq-oper-manpower/backend/internal/usecases/allocation"
)

type RotationHandler struct {
	repos            *postgres.Repositories
	cfg              *config.Config
	suggestionEngine *allocation.SuggestionEngine
}

func NewRotationHandler(repos *postgres.Repositories, cfg *config.Config, suggestionEngine *allocation.SuggestionEngine) *RotationHandler {
	return &RotationHandler{
		repos:            repos,
		cfg:              cfg,
		suggestionEngine: suggestionEngine,
	}
}

type AssignRotationRequest struct {
	RotationStaffID uuid.UUID `json:"rotation_staff_id" binding:"required"`
	BranchID        uuid.UUID `json:"branch_id" binding:"required"`
	Date            string    `json:"date" binding:"required"`
	AssignmentLevel int       `json:"assignment_level" binding:"required"`
	IsAdhoc         bool      `json:"is_adhoc"`
	AdhocReason     string    `json:"adhoc_reason"`
}

func (h *RotationHandler) GetAssignments(c *gin.Context) {
	branchIDStr := c.Query("branch_id")
	rotationStaffIDStr := c.Query("rotation_staff_id")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	coverageArea := c.Query("coverage_area")

	filters := interfaces.RotationFilters{}

	if branchIDStr != "" {
		branchID, err := uuid.Parse(branchIDStr)
		if err == nil {
			filters.BranchID = &branchID
		}
	}

	if rotationStaffIDStr != "" {
		rotationStaffID, err := uuid.Parse(rotationStaffIDStr)
		if err == nil {
			filters.RotationStaffID = &rotationStaffID
		}
	}

	if startDateStr != "" {
		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err == nil {
			filters.StartDate = &startDate
		}
	}

	if endDateStr != "" {
		endDate, err := time.Parse("2006-01-02", endDateStr)
		if err == nil {
			filters.EndDate = &endDate
		}
	}

	if coverageArea != "" {
		filters.CoverageArea = &coverageArea
	}

	assignments, err := h.repos.Rotation.GetAssignments(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"assignments": assignments})
}

func (h *RotationHandler) Assign(c *gin.Context) {
	var req AssignRotationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format"})
		return
	}

	if req.AssignmentLevel != 1 && req.AssignmentLevel != 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "assignment_level must be 1 or 2"})
		return
	}

	userIDStr := c.MustGet("user_id").(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	assignment := &models.RotationAssignment{
		ID:              uuid.New(),
		RotationStaffID: req.RotationStaffID,
		BranchID:        req.BranchID,
		Date:            date,
		AssignmentLevel: req.AssignmentLevel,
		IsAdhoc:         req.IsAdhoc,
		AdhocReason:     req.AdhocReason,
		AssignedBy:      userID,
	}

	if err := h.repos.Rotation.Create(assignment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"assignment": assignment})
}

func (h *RotationHandler) RemoveAssignment(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := h.repos.Rotation.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Assignment removed successfully"})
}

type SuggestionsRequest struct {
	BranchID   string `json:"branch_id"`
	StartDate  string `json:"start_date"`
	EndDate    string `json:"end_date"`
}

func (h *RotationHandler) GetSuggestions(c *gin.Context) {
	branchIDStr := c.Query("branch_id")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	// Parse dates
	var startDate, endDate time.Time
	var err error
	if startDateStr != "" && endDateStr != "" {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format"})
			return
		}
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format"})
			return
		}
	} else {
		// Default to current month
		now := time.Now()
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		endDate = startDate.AddDate(0, 1, -1)
	}

	// Get branches to suggest for
	var branchIDs []uuid.UUID
	if branchIDStr != "" {
		branchID, err := uuid.Parse(branchIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid branch_id"})
			return
		}
		branch, err := h.repos.Branch.GetByID(branchID)
		if err != nil || branch == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Branch not found"})
			return
		}
		branchIDs = []uuid.UUID{branchID}
	} else {
		allBranches, err := h.repos.Branch.List()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		branchIDs = make([]uuid.UUID, len(allBranches))
		for i, branch := range allBranches {
			branchIDs[i] = branch.ID
		}
	}

	// Use SuggestionEngine to generate suggestions based on three pillars criteria
	suggestions, err := h.suggestionEngine.GenerateSuggestions(branchIDs, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to generate suggestions: %v", err)})
		return
	}

	// Convert suggestions to response format
	response := make([]gin.H, 0)
	for _, suggestion := range suggestions {
		// Get assignment level from effective branch
		assignmentLevel := 2 // Default to Level 2
		effectiveBranches, err := h.repos.EffectiveBranch.GetByBranchID(suggestion.BranchID)
		if err == nil {
			for _, eb := range effectiveBranches {
				if eb.RotationStaffID == suggestion.RotationStaffID {
					assignmentLevel = eb.Level
					break
				}
			}
		}

		response = append(response, gin.H{
			"id":                suggestion.ID,
			"rotation_staff_id": suggestion.RotationStaffID,
			"branch_id":         suggestion.BranchID,
			"date":              suggestion.Date.Format("2006-01-02"),
			"position_id":       suggestion.PositionID,
			"assignment_level":  assignmentLevel,
			"confidence":        suggestion.Confidence,
			"reason":            suggestion.Reason,
			"status":            suggestion.Status,
		})
	}

	c.JSON(http.StatusOK, gin.H{"suggestions": response})
}

func (h *RotationHandler) RegenerateSuggestions(c *gin.Context) {
	// Regenerate is the same as GetSuggestions for now
	// In production, this would request new suggestions from MCP
	h.GetSuggestions(c)
}

// GetEligibleStaff returns rotation staff eligible for a specific branch
func (h *RotationHandler) GetEligibleStaff(c *gin.Context) {
	branchIDStr := c.Param("branchId")
	if branchIDStr == "" {
		branchIDStr = c.Query("branch_id")
	}

	if branchIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "branch_id is required"})
		return
	}

	branchID, err := uuid.Parse(branchIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid branch_id"})
		return
	}

	// Get effective branches for this branch (rotation staff eligible for this branch)
	effectiveBranches, err := h.repos.EffectiveBranch.GetByBranchID(branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get all rotation staff IDs from effective branches
	rotationStaffIDs := make(map[uuid.UUID]int) // Map of staff ID to level
	for _, eb := range effectiveBranches {
		rotationStaffIDs[eb.RotationStaffID] = eb.Level
	}

	// Get all rotation staff
	allRotationStaff, err := h.repos.Staff.GetRotationStaff()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Filter to only eligible staff and add level information
	type EligibleStaff struct {
		*models.Staff
		AssignmentLevel int `json:"assignment_level"`
	}

	eligibleStaff := make([]EligibleStaff, 0)
	for _, staff := range allRotationStaff {
		if level, exists := rotationStaffIDs[staff.ID]; exists {
			eligibleStaff = append(eligibleStaff, EligibleStaff{
				Staff:           staff,
				AssignmentLevel: level,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{"eligible_staff": eligibleStaff})
}

type BulkAssignRequest struct {
	Assignments []struct {
		RotationStaffID uuid.UUID `json:"rotation_staff_id" binding:"required"`
		Dates            []string  `json:"dates" binding:"required"`
		AssignmentLevel int       `json:"assignment_level" binding:"required"`
	} `json:"assignments" binding:"required"`
	BranchID uuid.UUID `json:"branch_id" binding:"required"`
}

// BulkAssign allows assigning multiple rotation staff to multiple dates at once
func (h *RotationHandler) BulkAssign(c *gin.Context) {
	var req BulkAssignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDStr := c.MustGet("user_id").(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	createdAssignments := make([]*models.RotationAssignment, 0)
	errors := make([]string, 0)

	for _, assignment := range req.Assignments {
		if assignment.AssignmentLevel != 1 && assignment.AssignmentLevel != 2 {
			errors = append(errors, fmt.Sprintf("Invalid assignment_level for staff %s", assignment.RotationStaffID))
			continue
		}

		for _, dateStr := range assignment.Dates {
			date, err := time.Parse("2006-01-02", dateStr)
			if err != nil {
				errors = append(errors, fmt.Sprintf("Invalid date format: %s", dateStr))
				continue
			}

			assignmentModel := &models.RotationAssignment{
				ID:              uuid.New(),
				RotationStaffID: assignment.RotationStaffID,
				BranchID:        req.BranchID,
				Date:            date,
				AssignmentLevel: assignment.AssignmentLevel,
				AssignedBy:      userID,
			}

			if err := h.repos.Rotation.Create(assignmentModel); err != nil {
				// Check if it's a duplicate key error (already assigned) - ignore it
				errStr := err.Error()
				if !strings.Contains(errStr, "duplicate key") && !strings.Contains(errStr, "UNIQUE constraint") {
					errors = append(errors, fmt.Sprintf("Failed to assign %s on %s: %v", assignment.RotationStaffID, dateStr, err))
				}
			} else {
				createdAssignments = append(createdAssignments, assignmentModel)
			}
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"created": len(createdAssignments),
		"assignments": createdAssignments,
		"errors": errors,
	})
}


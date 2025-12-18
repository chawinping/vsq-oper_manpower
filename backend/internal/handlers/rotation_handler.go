package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"vsq-oper-manpower/backend/internal/config"
	"vsq-oper-manpower/backend/internal/domain/interfaces"
	"vsq-oper-manpower/backend/internal/domain/models"
	"vsq-oper-manpower/backend/internal/repositories/postgres"
)

type RotationHandler struct {
	repos *postgres.Repositories
	cfg   *config.Config
}

func NewRotationHandler(repos *postgres.Repositories, cfg *config.Config) *RotationHandler {
	return &RotationHandler{repos: repos, cfg: cfg}
}

type AssignRotationRequest struct {
	RotationStaffID uuid.UUID `json:"rotation_staff_id" binding:"required"`
	BranchID        uuid.UUID `json:"branch_id" binding:"required"`
	Date            string    `json:"date" binding:"required"`
	AssignmentLevel int       `json:"assignment_level" binding:"required"`
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
	var branches []*models.Branch
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
		branches = []*models.Branch{branch}
	} else {
		allBranches, err := h.repos.Branch.List()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		branches = allBranches
	}

	// Get rotation staff
	rotationStaff, err := h.repos.Staff.GetRotationStaff()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Generate suggestions based on business logic
	// This is a placeholder - in production, this would call MCP client
	suggestions := make([]gin.H, 0)
	
	for _, branch := range branches {
		// Generate suggestions for each day in the date range
		for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
			// Simple logic: suggest rotation staff based on branch priority and expected revenue
			// In production, this would use the allocation engine and MCP client
			if len(rotationStaff) > 0 {
				// Select staff based on simple criteria (coverage area, availability, etc.)
				selectedStaff := rotationStaff[0] // Simplified selection
				
				// Determine assignment level based on branch priority
				assignmentLevel := 2
				if branch.Priority == 1 {
					assignmentLevel = 1
				}

				suggestions = append(suggestions, gin.H{
					"rotation_staff_id": selectedStaff.ID,
					"branch_id":         branch.ID,
					"date":              d.Format("2006-01-02"),
					"assignment_level":  assignmentLevel,
					"confidence":        0.75, // Placeholder confidence
					"reason":            fmt.Sprintf("Suggested based on branch priority (%d) and expected revenue", branch.Priority),
				})
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"suggestions": suggestions})
}

func (h *RotationHandler) RegenerateSuggestions(c *gin.Context) {
	// Regenerate is the same as GetSuggestions for now
	// In production, this would request new suggestions from MCP
	h.GetSuggestions(c)
}


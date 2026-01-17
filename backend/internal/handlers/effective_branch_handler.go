package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"vsq-oper-manpower/backend/internal/domain/models"
	"vsq-oper-manpower/backend/internal/repositories/postgres"
)

type EffectiveBranchHandler struct {
	repos *postgres.Repositories
}

func NewEffectiveBranchHandler(repos *postgres.Repositories) *EffectiveBranchHandler {
	return &EffectiveBranchHandler{repos: repos}
}

type CreateEffectiveBranchRequest struct {
	RotationStaffID        uuid.UUID `json:"rotation_staff_id" binding:"required"`
	BranchID               uuid.UUID `json:"branch_id" binding:"required"`
	Level                  int       `json:"level" binding:"required,oneof=1 2"`
	CommuteDurationMinutes *int      `json:"commute_duration_minutes,omitempty"`
	TransitCount           *int      `json:"transit_count,omitempty"`
	TravelCost             *float64  `json:"travel_cost,omitempty"`
}

// GetByRotationStaffID returns all effective branches for a rotation staff member
func (h *EffectiveBranchHandler) GetByRotationStaffID(c *gin.Context) {
	rotationStaffIDStr := c.Param("rotationStaffId")
	if rotationStaffIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "rotation_staff_id is required"})
		return
	}

	rotationStaffID, err := uuid.Parse(rotationStaffIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rotation_staff_id"})
		return
	}

	// Verify the staff member exists and is rotation staff
	staff, err := h.repos.Staff.GetByID(rotationStaffID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Rotation staff not found"})
		return
	}

	if staff.StaffType != models.StaffTypeRotation {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Staff member is not a rotation staff"})
		return
	}

	effectiveBranches, err := h.repos.EffectiveBranch.GetByRotationStaffID(rotationStaffID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Load branch details for each effective branch
	type EffectiveBranchResponse struct {
		ID                    uuid.UUID     `json:"id"`
		RotationStaffID       uuid.UUID     `json:"rotation_staff_id"`
		BranchID              uuid.UUID     `json:"branch_id"`
		Branch                *models.Branch `json:"branch"`
		Level                 int            `json:"level"`
		CommuteDurationMinutes *int          `json:"commute_duration_minutes,omitempty"`
		TransitCount          *int          `json:"transit_count,omitempty"`
		TravelCost            *float64      `json:"travel_cost,omitempty"`
		CreatedAt             string         `json:"created_at"`
	}

	response := make([]EffectiveBranchResponse, 0)
	for _, eb := range effectiveBranches {
		branch, err := h.repos.Branch.GetByID(eb.BranchID)
		if err != nil {
			continue // Skip if branch not found
		}

		response = append(response, EffectiveBranchResponse{
			ID:                    eb.ID,
			RotationStaffID:       eb.RotationStaffID,
			BranchID:              eb.BranchID,
			Branch:                branch,
			Level:                 eb.Level,
			CommuteDurationMinutes: eb.CommuteDurationMinutes,
			TransitCount:          eb.TransitCount,
			TravelCost:            eb.TravelCost,
			CreatedAt:             eb.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	c.JSON(http.StatusOK, gin.H{"effective_branches": response})
}

// Create creates a new effective branch assignment
func (h *EffectiveBranchHandler) Create(c *gin.Context) {
	var req CreateEffectiveBranchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify the staff member exists and is rotation staff
	staff, err := h.repos.Staff.GetByID(req.RotationStaffID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Rotation staff not found"})
		return
	}

	if staff.StaffType != models.StaffTypeRotation {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Staff member is not a rotation staff"})
		return
	}

	// Verify branch exists
	_, err = h.repos.Branch.GetByID(req.BranchID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Branch not found"})
		return
	}

	// Check if this effective branch already exists
	existingEBs, err := h.repos.EffectiveBranch.GetByRotationStaffID(req.RotationStaffID)
	if err == nil {
		for _, eb := range existingEBs {
			if eb.BranchID == req.BranchID {
				c.JSON(http.StatusConflict, gin.H{"error": "Effective branch assignment already exists"})
				return
			}
		}
	}

	effectiveBranch := &models.EffectiveBranch{
		ID:                    uuid.New(),
		RotationStaffID:       req.RotationStaffID,
		BranchID:              req.BranchID,
		Level:                 req.Level,
		CommuteDurationMinutes: req.CommuteDurationMinutes,
		TransitCount:          req.TransitCount,
		TravelCost:            req.TravelCost,
	}

	if err := h.repos.EffectiveBranch.Create(effectiveBranch); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Load branch details
	branch, _ := h.repos.Branch.GetByID(req.BranchID)
	effectiveBranch.Branch = branch

	c.JSON(http.StatusCreated, gin.H{"effective_branch": effectiveBranch})
}

type UpdateEffectiveBranchRequest struct {
	BranchID               uuid.UUID `json:"branch_id" binding:"required"`
	Level                  int       `json:"level" binding:"required,oneof=1 2"`
	CommuteDurationMinutes *int     `json:"commute_duration_minutes,omitempty"`
	TransitCount           *int     `json:"transit_count,omitempty"`
	TravelCost             *float64  `json:"travel_cost,omitempty"`
}

// Update updates an existing effective branch assignment
func (h *EffectiveBranchHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var req UpdateEffectiveBranchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get existing effective branch
	existingEB, err := h.repos.EffectiveBranch.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Effective branch not found"})
		return
	}

	// Verify branch exists
	_, err = h.repos.Branch.GetByID(req.BranchID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Branch not found"})
		return
	}

	// Update the effective branch
	existingEB.BranchID = req.BranchID
	existingEB.Level = req.Level
	existingEB.CommuteDurationMinutes = req.CommuteDurationMinutes
	existingEB.TransitCount = req.TransitCount
	existingEB.TravelCost = req.TravelCost

	if err := h.repos.EffectiveBranch.Update(existingEB); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Load branch details
	branch, _ := h.repos.Branch.GetByID(req.BranchID)
	existingEB.Branch = branch

	c.JSON(http.StatusOK, gin.H{"effective_branch": existingEB})
}

// Delete removes an effective branch assignment
func (h *EffectiveBranchHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := h.repos.EffectiveBranch.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Effective branch deleted successfully"})
}

// BulkUpdate updates effective branches for a rotation staff member
// This replaces all existing effective branches with the new list
type BulkUpdateEffectiveBranchesRequest struct {
	RotationStaffID uuid.UUID `json:"rotation_staff_id" binding:"required"`
	EffectiveBranches []struct {
		BranchID              uuid.UUID `json:"branch_id" binding:"required"`
		Level                 int       `json:"level" binding:"required,oneof=1 2"`
		CommuteDurationMinutes *int     `json:"commute_duration_minutes,omitempty"`
		TransitCount          *int      `json:"transit_count,omitempty"`
		TravelCost            *float64  `json:"travel_cost,omitempty"`
	} `json:"effective_branches" binding:"required"`
}

func (h *EffectiveBranchHandler) BulkUpdate(c *gin.Context) {
	var req BulkUpdateEffectiveBranchesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify the staff member exists and is rotation staff
	staff, err := h.repos.Staff.GetByID(req.RotationStaffID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Rotation staff not found"})
		return
	}

	if staff.StaffType != models.StaffTypeRotation {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Staff member is not a rotation staff"})
		return
	}

	// Verify all branches exist
	for _, eb := range req.EffectiveBranches {
		_, err := h.repos.Branch.GetByID(eb.BranchID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Branch not found: " + eb.BranchID.String()})
			return
		}
	}

	// Delete all existing effective branches for this rotation staff
	if err := h.repos.EffectiveBranch.DeleteByRotationStaffID(req.RotationStaffID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete existing effective branches: " + err.Error()})
		return
	}

	// Create new effective branches
	created := make([]*models.EffectiveBranch, 0)
	for _, eb := range req.EffectiveBranches {
		effectiveBranch := &models.EffectiveBranch{
			ID:                    uuid.New(),
			RotationStaffID:       req.RotationStaffID,
			BranchID:              eb.BranchID,
			Level:                 eb.Level,
			CommuteDurationMinutes: eb.CommuteDurationMinutes,
			TransitCount:          eb.TransitCount,
			TravelCost:            eb.TravelCost,
		}

		if err := h.repos.EffectiveBranch.Create(effectiveBranch); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create effective branch: " + err.Error()})
			return
		}

		// Load branch details
		branch, _ := h.repos.Branch.GetByID(eb.BranchID)
		effectiveBranch.Branch = branch
		created = append(created, effectiveBranch)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":           "Effective branches updated successfully",
		"effective_branches": created,
	})
}



package handlers

import (
	"database/sql"
	"net/http"

	"vsq-oper-manpower/backend/internal/domain/models"
	"vsq-oper-manpower/backend/internal/repositories/postgres"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type BranchTypeConstraintsHandler struct {
	repos *postgres.Repositories
	db    *sql.DB
}

func NewBranchTypeConstraintsHandler(repos *postgres.Repositories, db *sql.DB) *BranchTypeConstraintsHandler {
	return &BranchTypeConstraintsHandler{repos: repos, db: db}
}

func (h *BranchTypeConstraintsHandler) GetByBranchTypeID(c *gin.Context) {
	branchTypeIDStr := c.Param("id")
	branchTypeID, err := uuid.Parse(branchTypeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid branch type ID"})
		return
	}

	constraints, err := h.repos.BranchTypeConstraints.GetByBranchTypeID(branchTypeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Load staff group requirements
	if err := h.repos.BranchTypeConstraints.LoadStaffGroupRequirements(constraints); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load staff group requirements: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"constraints": constraints})
}

type ConstraintsUpdate struct {
	DayOfWeek              int                     `json:"day_of_week" binding:"required"`
	StaffGroupRequirements []StaffGroupRequirement `json:"staff_group_requirements" binding:"required"`
}

type UpdateBranchTypeConstraintsRequest struct {
	Constraints []ConstraintsUpdate `json:"constraints" binding:"required"`
}

func (h *BranchTypeConstraintsHandler) UpdateConstraints(c *gin.Context) {
	branchTypeIDStr := c.Param("id")
	branchTypeID, err := uuid.Parse(branchTypeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid branch type ID"})
		return
	}

	var req UpdateBranchTypeConstraintsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data: " + err.Error()})
		return
	}

	if len(req.Constraints) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No constraints provided"})
		return
	}

	// Convert to models with staff group requirements
	constraints := make([]*models.BranchTypeConstraints, len(req.Constraints))
	for i, update := range req.Constraints {
		constraint := &models.BranchTypeConstraints{
			BranchTypeID: branchTypeID,
			DayOfWeek:    update.DayOfWeek,
		}

		// Convert staff group requirements
		if len(update.StaffGroupRequirements) > 0 {
			constraint.StaffGroupRequirements = make([]*models.BranchTypeConstraintStaffGroup, len(update.StaffGroupRequirements))
			for j, sgReq := range update.StaffGroupRequirements {
				constraint.StaffGroupRequirements[j] = &models.BranchTypeConstraintStaffGroup{
					StaffGroupID: sgReq.StaffGroupID,
					MinimumCount: sgReq.MinimumCount,
				}
			}
		}

		constraints[i] = constraint
	}

	if err := h.repos.BranchTypeConstraints.BulkUpsertWithStaffGroups(constraints); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save constraints: " + err.Error()})
		return
	}

	// Return updated constraints with staff group requirements loaded
	updatedConstraints, err := h.repos.BranchTypeConstraints.GetByBranchTypeID(branchTypeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve updated constraints: " + err.Error()})
		return
	}

	// Load staff group requirements
	if err := h.repos.BranchTypeConstraints.LoadStaffGroupRequirements(updatedConstraints); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load staff group requirements: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"constraints": updatedConstraints})
}

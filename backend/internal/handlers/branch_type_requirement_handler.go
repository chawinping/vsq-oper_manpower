package handlers

import (
	"database/sql"
	"net/http"

	"vsq-oper-manpower/backend/internal/domain/models"
	"vsq-oper-manpower/backend/internal/repositories/postgres"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type BranchTypeRequirementHandler struct {
	repos *postgres.Repositories
	db    *sql.DB
}

func NewBranchTypeRequirementHandler(repos *postgres.Repositories, db *sql.DB) *BranchTypeRequirementHandler {
	return &BranchTypeRequirementHandler{repos: repos, db: db}
}

func (h *BranchTypeRequirementHandler) GetByBranchTypeID(c *gin.Context) {
	branchTypeIDStr := c.Param("id")
	branchTypeID, err := uuid.Parse(branchTypeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid branch type ID"})
		return
	}

	requirements, err := h.repos.BranchTypeRequirement.GetByBranchTypeID(branchTypeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"requirements": requirements})
}

type CreateRequirementRequest struct {
	StaffGroupID      uuid.UUID `json:"staff_group_id" binding:"required"`
	DayOfWeek         int       `json:"day_of_week" binding:"required,min=0,max=6"`
	MinimumStaffCount int       `json:"minimum_staff_count" binding:"required,min=0"`
	IsActive          bool      `json:"is_active"`
}

func (h *BranchTypeRequirementHandler) Create(c *gin.Context) {
	branchTypeIDStr := c.Param("id")
	branchTypeID, err := uuid.Parse(branchTypeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid branch type ID"})
		return
	}

	var req CreateRequirementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	requirement := &models.BranchTypeStaffGroupRequirement{
		BranchTypeID:      branchTypeID,
		StaffGroupID:      req.StaffGroupID,
		DayOfWeek:         req.DayOfWeek,
		MinimumStaffCount: req.MinimumStaffCount,
		IsActive:          req.IsActive,
	}

	if err := h.repos.BranchTypeRequirement.Create(requirement); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"requirement": requirement})
}

type UpdateRequirementRequest struct {
	MinimumStaffCount int  `json:"minimum_staff_count" binding:"required,min=0"`
	IsActive          bool `json:"is_active"`
}

func (h *BranchTypeRequirementHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid requirement ID"})
		return
	}

	var req UpdateRequirementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	requirement, err := h.repos.BranchTypeRequirement.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if requirement == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Requirement not found"})
		return
	}

	requirement.MinimumStaffCount = req.MinimumStaffCount
	requirement.IsActive = req.IsActive

	if err := h.repos.BranchTypeRequirement.Update(requirement); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"requirement": requirement})
}

func (h *BranchTypeRequirementHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid requirement ID"})
		return
	}

	if err := h.repos.BranchTypeRequirement.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Requirement deleted successfully"})
}

type BulkUpsertRequirementsRequest struct {
	Requirements []RequirementUpdate `json:"requirements" binding:"required"`
}

type RequirementUpdate struct {
	StaffGroupID      uuid.UUID `json:"staff_group_id" binding:"required"`
	DayOfWeek         int       `json:"day_of_week" binding:"required,min=0,max=6"`
	MinimumStaffCount int       `json:"minimum_staff_count" binding:"required,min=0"`
	IsActive          bool      `json:"is_active"`
}

func (h *BranchTypeRequirementHandler) BulkUpsert(c *gin.Context) {
	branchTypeIDStr := c.Param("id")
	branchTypeID, err := uuid.Parse(branchTypeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid branch type ID"})
		return
	}

	var req BulkUpsertRequirementsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert to models
	requirements := make([]*models.BranchTypeStaffGroupRequirement, len(req.Requirements))
	for i, update := range req.Requirements {
		requirements[i] = &models.BranchTypeStaffGroupRequirement{
			BranchTypeID:      branchTypeID,
			StaffGroupID:      update.StaffGroupID,
			DayOfWeek:         update.DayOfWeek,
			MinimumStaffCount: update.MinimumStaffCount,
			IsActive:          update.IsActive,
		}
	}

	if err := h.repos.BranchTypeRequirement.BulkUpsert(requirements); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return updated requirements
	updatedRequirements, err := h.repos.BranchTypeRequirement.GetByBranchTypeID(branchTypeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"requirements": updatedRequirements})
}

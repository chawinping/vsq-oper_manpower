package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"vsq-oper-manpower/backend/internal/domain/interfaces"
	"vsq-oper-manpower/backend/internal/domain/models"
	"vsq-oper-manpower/backend/internal/repositories/postgres"
)

type AllocationCriteriaHandler struct {
	repos *postgres.Repositories
}

func NewAllocationCriteriaHandler(repos *postgres.Repositories) *AllocationCriteriaHandler {
	return &AllocationCriteriaHandler{repos: repos}
}

type CreateAllocationCriteriaRequest struct {
	Pillar      models.CriteriaPillar `json:"pillar" binding:"required"`
	Type        models.CriteriaType   `json:"type" binding:"required"`
	Weight      float64               `json:"weight" binding:"required"`
	IsActive    bool                  `json:"is_active"`
	Description string                `json:"description"`
	Config      string                `json:"config"`
}

func (h *AllocationCriteriaHandler) CreateCriteria(c *gin.Context) {
	var req CreateAllocationCriteriaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate weight range
	if req.Weight < 0.0 || req.Weight > 1.0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Weight must be between 0.0 and 1.0"})
		return
	}

	userIDStr := c.MustGet("user_id").(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	criteria := &models.AllocationCriteria{
		ID:          uuid.New(),
		Pillar:      req.Pillar,
		Type:        req.Type,
		Weight:      req.Weight,
		IsActive:    req.IsActive,
		Description: req.Description,
		Config:      req.Config,
		CreatedBy:   userID,
	}

	if err := h.repos.AllocationCriteria.Create(criteria); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"criteria": criteria})
}

func (h *AllocationCriteriaHandler) GetCriteria(c *gin.Context) {
	pillarStr := c.Query("pillar")
	typeStr := c.Query("type")
	isActiveStr := c.Query("is_active")

	filters := interfaces.AllocationCriteriaFilters{}

	if pillarStr != "" {
		pillar := models.CriteriaPillar(pillarStr)
		filters.Pillar = &pillar
	}

	if typeStr != "" {
		criteriaType := models.CriteriaType(typeStr)
		filters.Type = &criteriaType
	}

	if isActiveStr != "" {
		isActive := isActiveStr == "true"
		filters.IsActive = &isActive
	} else {
		// Default to active only
		isActive := true
		filters.IsActive = &isActive
	}

	criteriaList, err := h.repos.AllocationCriteria.List(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"criteria": criteriaList})
}

func (h *AllocationCriteriaHandler) GetCriteriaByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	criteria, err := h.repos.AllocationCriteria.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if criteria == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Criteria not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"criteria": criteria})
}

func (h *AllocationCriteriaHandler) UpdateCriteria(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var req struct {
		Pillar      *models.CriteriaPillar `json:"pillar"`
		Type        *models.CriteriaType    `json:"type"`
		Weight      *float64                `json:"weight"`
		IsActive    *bool                   `json:"is_active"`
		Description *string                 `json:"description"`
		Config      *string                 `json:"config"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	criteria, err := h.repos.AllocationCriteria.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if criteria == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Criteria not found"})
		return
	}

	if req.Pillar != nil {
		criteria.Pillar = *req.Pillar
	}
	if req.Type != nil {
		criteria.Type = *req.Type
	}
	if req.Weight != nil {
		if *req.Weight < 0.0 || *req.Weight > 1.0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Weight must be between 0.0 and 1.0"})
			return
		}
		criteria.Weight = *req.Weight
	}
	if req.IsActive != nil {
		criteria.IsActive = *req.IsActive
	}
	if req.Description != nil {
		criteria.Description = *req.Description
	}
	if req.Config != nil {
		criteria.Config = *req.Config
	}

	if err := h.repos.AllocationCriteria.Update(criteria); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"criteria": criteria})
}

func (h *AllocationCriteriaHandler) DeleteCriteria(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := h.repos.AllocationCriteria.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Criteria deleted successfully"})
}

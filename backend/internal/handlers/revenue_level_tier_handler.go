package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"vsq-oper-manpower/backend/internal/domain/models"
	"vsq-oper-manpower/backend/internal/repositories/postgres"
)

type RevenueLevelTierHandler struct {
	repos *postgres.Repositories
}

func NewRevenueLevelTierHandler(repos *postgres.Repositories) *RevenueLevelTierHandler {
	return &RevenueLevelTierHandler{repos: repos}
}

// List returns all revenue level tiers
func (h *RevenueLevelTierHandler) List(c *gin.Context) {
	tiers, err := h.repos.RevenueLevelTier.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tiers)
}

// GetByID returns a revenue level tier by ID
func (h *RevenueLevelTierHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tier ID"})
		return
	}

	tier, err := h.repos.RevenueLevelTier.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if tier == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tier not found"})
		return
	}
	c.JSON(http.StatusOK, tier)
}

// Create creates a new revenue level tier
func (h *RevenueLevelTierHandler) Create(c *gin.Context) {
	var req models.RevenueLevelTierCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tier := &models.RevenueLevelTier{
		ID:          uuid.New(),
		LevelNumber: req.LevelNumber,
		LevelName:   req.LevelName,
		MinRevenue:  req.MinRevenue,
		MaxRevenue:  req.MaxRevenue,
		DisplayOrder: req.DisplayOrder,
		ColorCode:   req.ColorCode,
		Description: req.Description,
	}

	if err := h.repos.RevenueLevelTier.Create(tier); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, tier)
}

// Update updates a revenue level tier
func (h *RevenueLevelTierHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tier ID"})
		return
	}

	tier, err := h.repos.RevenueLevelTier.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if tier == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tier not found"})
		return
	}

	var req models.RevenueLevelTierUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.LevelName != nil {
		tier.LevelName = *req.LevelName
	}
	if req.MinRevenue != nil {
		tier.MinRevenue = *req.MinRevenue
	}
	if req.MaxRevenue != nil {
		tier.MaxRevenue = req.MaxRevenue
	}
	if req.DisplayOrder != nil {
		tier.DisplayOrder = *req.DisplayOrder
	}
	if req.ColorCode != nil {
		tier.ColorCode = req.ColorCode
	}
	if req.Description != nil {
		tier.Description = req.Description
	}

	if err := h.repos.RevenueLevelTier.Update(tier); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tier)
}

// Delete deletes a revenue level tier
func (h *RevenueLevelTierHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tier ID"})
		return
	}

	if err := h.repos.RevenueLevelTier.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tier deleted successfully"})
}

// GetTierForRevenue returns the tier that matches a given revenue amount
func (h *RevenueLevelTierHandler) GetTierForRevenue(c *gin.Context) {
	var req struct {
		Revenue float64 `json:"revenue" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tier, err := h.repos.RevenueLevelTier.GetTierForRevenue(req.Revenue)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if tier == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No tier found for this revenue amount"})
		return
	}
	c.JSON(http.StatusOK, tier)
}

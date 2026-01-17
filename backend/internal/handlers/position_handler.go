package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"vsq-oper-manpower/backend/internal/domain/models"
	"vsq-oper-manpower/backend/internal/repositories/postgres"
)

type PositionHandler struct {
	repos *postgres.Repositories
}

func NewPositionHandler(repos *postgres.Repositories) *PositionHandler {
	return &PositionHandler{repos: repos}
}

func (h *PositionHandler) List(c *gin.Context) {
	positions, err := h.repos.Position.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"positions": positions})
}

func (h *PositionHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	position, err := h.repos.Position.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if position == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Position not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"position": position})
}

type UpdatePositionRequest struct {
	Name         string                `json:"name" binding:"required"`
	DisplayOrder int                   `json:"display_order"`
	PositionType models.PositionType   `json:"position_type" binding:"required"`
	ManpowerType models.ManpowerType   `json:"manpower_type" binding:"required"`
}

func (h *PositionHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	// Check if position exists
	existingPosition, err := h.repos.Position.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existingPosition == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Position not found"})
		return
	}

	var req UpdatePositionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	position := &models.Position{
		ID:                id,
		Name:              req.Name,
		PositionType:      req.PositionType,
		ManpowerType:      req.ManpowerType,
		MinStaffPerBranch: existingPosition.MinStaffPerBranch, // Keep existing value, don't update
		DisplayOrder:      req.DisplayOrder,
		CreatedAt:         existingPosition.CreatedAt,
	}

	if err := h.repos.Position.Update(position); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"position": position})
}






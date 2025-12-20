package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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




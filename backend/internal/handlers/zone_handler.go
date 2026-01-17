package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"vsq-oper-manpower/backend/internal/domain/models"
	"vsq-oper-manpower/backend/internal/repositories/postgres"
)

type ZoneHandler struct {
	repos *postgres.Repositories
}

func NewZoneHandler(repos *postgres.Repositories) *ZoneHandler {
	return &ZoneHandler{repos: repos}
}

type CreateZoneRequest struct {
	Name        string `json:"name" binding:"required"`
	Code        string `json:"code" binding:"required"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
}

type UpdateZoneBranchesRequest struct {
	BranchIDs []uuid.UUID `json:"branch_ids" binding:"required"`
}

func (h *ZoneHandler) List(c *gin.Context) {
	includeInactive := c.Query("include_inactive") == "true"
	
	zones, err := h.repos.Zone.List(includeInactive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"zones": zones})
}

func (h *ZoneHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	zone, err := h.repos.Zone.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Zone not found"})
		return
	}

	// Load branches for this zone
	branches, err := h.repos.Zone.GetBranches(id)
	if err == nil {
		zone.Branches = branches
	}

	c.JSON(http.StatusOK, gin.H{"zone": zone})
}

func (h *ZoneHandler) Create(c *gin.Context) {
	var req CreateZoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if code already exists
	existing, _ := h.repos.Zone.GetByCode(req.Code)
	if existing != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Zone with this code already exists"})
		return
	}

	zone := &models.Zone{
		ID:          uuid.New(),
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		IsActive:    req.IsActive,
	}

	if err := h.repos.Zone.Create(zone); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"zone": zone})
}

func (h *ZoneHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var req CreateZoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if code already exists (excluding current record)
	existing, _ := h.repos.Zone.GetByCode(req.Code)
	if existing != nil && existing.ID != id {
		c.JSON(http.StatusConflict, gin.H{"error": "Zone with this code already exists"})
		return
	}

	zone := &models.Zone{
		ID:          id,
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		IsActive:    req.IsActive,
	}

	if err := h.repos.Zone.Update(zone); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"zone": zone})
}

func (h *ZoneHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := h.repos.Zone.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Zone deleted successfully"})
}

func (h *ZoneHandler) GetBranches(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	branches, err := h.repos.Zone.GetBranches(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"branches": branches})
}

func (h *ZoneHandler) UpdateBranches(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var req UpdateZoneBranchesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repos.Zone.BulkUpdateBranches(id, req.BranchIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Zone branches updated successfully"})
}

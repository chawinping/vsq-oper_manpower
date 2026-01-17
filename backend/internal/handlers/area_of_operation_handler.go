package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"vsq-oper-manpower/backend/internal/domain/models"
	"vsq-oper-manpower/backend/internal/repositories/postgres"
)

type AreaOfOperationHandler struct {
	repos *postgres.Repositories
}

func NewAreaOfOperationHandler(repos *postgres.Repositories) *AreaOfOperationHandler {
	return &AreaOfOperationHandler{repos: repos}
}

type CreateAreaOfOperationRequest struct {
	Name        string `json:"name" binding:"required"`
	Code        string `json:"code" binding:"required"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
}

func (h *AreaOfOperationHandler) List(c *gin.Context) {
	includeInactive := c.Query("include_inactive") == "true"
	
	areas, err := h.repos.AreaOfOperation.List(includeInactive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"areas_of_operation": areas})
}

func (h *AreaOfOperationHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	area, err := h.repos.AreaOfOperation.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Area of operation not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"area_of_operation": area})
}

func (h *AreaOfOperationHandler) Create(c *gin.Context) {
	var req CreateAreaOfOperationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if code already exists
	existing, _ := h.repos.AreaOfOperation.GetByCode(req.Code)
	if existing != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Area of operation with this code already exists"})
		return
	}

	area := &models.AreaOfOperation{
		ID:          uuid.New(),
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		IsActive:    req.IsActive,
	}

	if err := h.repos.AreaOfOperation.Create(area); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"area_of_operation": area})
}

func (h *AreaOfOperationHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var req CreateAreaOfOperationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if code already exists (excluding current record)
	existing, _ := h.repos.AreaOfOperation.GetByCode(req.Code)
	if existing != nil && existing.ID != id {
		c.JSON(http.StatusConflict, gin.H{"error": "Area of operation with this code already exists"})
		return
	}

	area := &models.AreaOfOperation{
		ID:          id,
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		IsActive:    req.IsActive,
	}

	if err := h.repos.AreaOfOperation.Update(area); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"area_of_operation": area})
}

func (h *AreaOfOperationHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := h.repos.AreaOfOperation.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Area of operation deleted successfully"})
}

func (h *AreaOfOperationHandler) AddZone(c *gin.Context) {
	idStr := c.Param("id")
	areaID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid area ID"})
		return
	}

	zoneIDStr := c.Query("zone_id")
	if zoneIDStr == "" {
		var req struct {
			ZoneID uuid.UUID `json:"zone_id" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		zoneIDStr = req.ZoneID.String()
	}

	zoneID, err := uuid.Parse(zoneIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid zone ID"})
		return
	}

	if err := h.repos.AreaOfOperation.AddZone(areaID, zoneID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Zone added to area of operation successfully"})
}

func (h *AreaOfOperationHandler) RemoveZone(c *gin.Context) {
	idStr := c.Param("id")
	areaID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid area ID"})
		return
	}

	zoneIDStr := c.Param("zoneId")
	zoneID, err := uuid.Parse(zoneIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid zone ID"})
		return
	}

	if err := h.repos.AreaOfOperation.RemoveZone(areaID, zoneID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Zone removed from area of operation successfully"})
}

func (h *AreaOfOperationHandler) GetZones(c *gin.Context) {
	idStr := c.Param("id")
	areaID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid area ID"})
		return
	}

	zones, err := h.repos.AreaOfOperation.GetZones(areaID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"zones": zones})
}

func (h *AreaOfOperationHandler) AddBranch(c *gin.Context) {
	idStr := c.Param("id")
	areaID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid area ID"})
		return
	}

	var req struct {
		BranchID uuid.UUID `json:"branch_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repos.AreaOfOperation.AddBranch(areaID, req.BranchID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Branch added to area of operation successfully"})
}

func (h *AreaOfOperationHandler) RemoveBranch(c *gin.Context) {
	idStr := c.Param("id")
	areaID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid area ID"})
		return
	}

	branchIDStr := c.Param("branchId")
	branchID, err := uuid.Parse(branchIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid branch ID"})
		return
	}

	if err := h.repos.AreaOfOperation.RemoveBranch(areaID, branchID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Branch removed from area of operation successfully"})
}

func (h *AreaOfOperationHandler) GetBranches(c *gin.Context) {
	idStr := c.Param("id")
	areaID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid area ID"})
		return
	}

	branches, err := h.repos.AreaOfOperation.GetBranches(areaID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"branches": branches})
}

func (h *AreaOfOperationHandler) GetAllBranches(c *gin.Context) {
	idStr := c.Param("id")
	areaID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid area ID"})
		return
	}

	branches, err := h.repos.AreaOfOperation.GetAllBranches(areaID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"branches": branches})
}



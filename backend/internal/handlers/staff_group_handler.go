package handlers

import (
	"database/sql"
	"net/http"

	"vsq-oper-manpower/backend/internal/domain/models"
	"vsq-oper-manpower/backend/internal/repositories/postgres"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type StaffGroupHandler struct {
	repos *postgres.Repositories
	db    *sql.DB
}

func NewStaffGroupHandler(repos *postgres.Repositories, db *sql.DB) *StaffGroupHandler {
	return &StaffGroupHandler{repos: repos, db: db}
}

func (h *StaffGroupHandler) List(c *gin.Context) {
	staffGroups, err := h.repos.StaffGroup.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Load positions for each staff group
	for _, group := range staffGroups {
		positions, err := h.repos.StaffGroupPosition.GetByStaffGroupID(group.ID)
		if err == nil {
			group.Positions = positions
		}
	}

	c.JSON(http.StatusOK, gin.H{"staff_groups": staffGroups})
}

func (h *StaffGroupHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	staffGroup, err := h.repos.StaffGroup.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if staffGroup == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Staff group not found"})
		return
	}

	// Load positions
	positions, err := h.repos.StaffGroupPosition.GetByStaffGroupID(id)
	if err == nil {
		staffGroup.Positions = positions
	}

	c.JSON(http.StatusOK, gin.H{"staff_group": staffGroup})
}

type CreateStaffGroupRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
}

func (h *StaffGroupHandler) Create(c *gin.Context) {
	var req CreateStaffGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	staffGroup := &models.StaffGroup{
		Name:        req.Name,
		Description: req.Description,
		IsActive:    req.IsActive,
	}

	if err := h.repos.StaffGroup.Create(staffGroup); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"staff_group": staffGroup})
}

type UpdateStaffGroupRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
}

func (h *StaffGroupHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var req UpdateStaffGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	staffGroup, err := h.repos.StaffGroup.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if staffGroup == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Staff group not found"})
		return
	}

	staffGroup.Name = req.Name
	staffGroup.Description = req.Description
	staffGroup.IsActive = req.IsActive

	if err := h.repos.StaffGroup.Update(staffGroup); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"staff_group": staffGroup})
}

func (h *StaffGroupHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := h.repos.StaffGroup.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Staff group deleted successfully"})
}

type AddPositionRequest struct {
	PositionID uuid.UUID `json:"position_id" binding:"required"`
}

func (h *StaffGroupHandler) AddPosition(c *gin.Context) {
	staffGroupIDStr := c.Param("id")
	staffGroupID, err := uuid.Parse(staffGroupIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid staff group ID"})
		return
	}

	var req AddPositionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sgp := &models.StaffGroupPosition{
		StaffGroupID: staffGroupID,
		PositionID:   req.PositionID,
	}

	if err := h.repos.StaffGroupPosition.Create(sgp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"staff_group_position": sgp})
}

func (h *StaffGroupHandler) RemovePosition(c *gin.Context) {
	staffGroupIDStr := c.Param("id")
	positionIDStr := c.Param("positionId")

	staffGroupID, err := uuid.Parse(staffGroupIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid staff group ID"})
		return
	}

	positionID, err := uuid.Parse(positionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid position ID"})
		return
	}

	if err := h.repos.StaffGroupPosition.DeleteByStaffGroupAndPosition(staffGroupID, positionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Position removed from staff group successfully"})
}

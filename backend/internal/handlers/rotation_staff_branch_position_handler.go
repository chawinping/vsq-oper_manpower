package handlers

import (
	"database/sql"
	"net/http"

	"vsq-oper-manpower/backend/internal/domain/models"
	"vsq-oper-manpower/backend/internal/repositories/postgres"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RotationStaffBranchPositionHandler struct {
	repos *postgres.Repositories
	db    *sql.DB
}

func NewRotationStaffBranchPositionHandler(repos *postgres.Repositories, db *sql.DB) *RotationStaffBranchPositionHandler {
	return &RotationStaffBranchPositionHandler{repos: repos, db: db}
}

func (h *RotationStaffBranchPositionHandler) List(c *gin.Context) {
	// Get query parameters for filtering
	staffIDStr := c.Query("staff_id")
	positionIDStr := c.Query("position_id")

	var mappings []*models.RotationStaffBranchPosition
	var err error

	if staffIDStr != "" {
		staffID, parseErr := uuid.Parse(staffIDStr)
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid staff_id"})
			return
		}
		mappings, err = h.repos.RotationStaffBranchPosition.GetByStaffID(staffID)
	} else if positionIDStr != "" {
		positionID, parseErr := uuid.Parse(positionIDStr)
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid position_id"})
			return
		}
		mappings, err = h.repos.RotationStaffBranchPosition.GetByPositionID(positionID)
	} else {
		// If no filter, return all mappings
		mappings, err = h.repos.RotationStaffBranchPosition.List()
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Load related data
	for _, mapping := range mappings {
		// Load staff
		if staff, err := h.repos.Staff.GetByID(mapping.RotationStaffID); err == nil && staff != nil {
			mapping.RotationStaff = staff
			// Load staff position
			if staff.PositionID != uuid.Nil {
				if position, err := h.repos.Position.GetByID(staff.PositionID); err == nil && position != nil {
					staff.Position = position
				}
			}
		}
		// Load branch position
		if position, err := h.repos.Position.GetByID(mapping.BranchPositionID); err == nil && position != nil {
			mapping.BranchPosition = position
		}
	}

	c.JSON(http.StatusOK, gin.H{"mappings": mappings})
}

func (h *RotationStaffBranchPositionHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	mapping, err := h.repos.RotationStaffBranchPosition.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if mapping == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Mapping not found"})
		return
	}

	// Load related data
	if staff, err := h.repos.Staff.GetByID(mapping.RotationStaffID); err == nil && staff != nil {
		mapping.RotationStaff = staff
		if staff.PositionID != uuid.Nil {
			if position, err := h.repos.Position.GetByID(staff.PositionID); err == nil && position != nil {
				staff.Position = position
			}
		}
	}
	if position, err := h.repos.Position.GetByID(mapping.BranchPositionID); err == nil && position != nil {
		mapping.BranchPosition = position
	}

	c.JSON(http.StatusOK, gin.H{"mapping": mapping})
}

type CreateRotationStaffBranchPositionRequest struct {
	RotationStaffID   uuid.UUID `json:"rotation_staff_id" binding:"required"`
	BranchPositionID  uuid.UUID `json:"branch_position_id" binding:"required"`
	SubstitutionLevel int       `json:"substitution_level" binding:"required"`
	IsActive          bool      `json:"is_active"`
	Notes             string    `json:"notes"`
}

func (h *RotationStaffBranchPositionHandler) Create(c *gin.Context) {
	var req CreateRotationStaffBranchPositionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate substitution level
	if req.SubstitutionLevel < 1 || req.SubstitutionLevel > 3 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "substitution_level must be between 1 and 3"})
		return
	}

	// Verify staff is rotation staff
	staff, err := h.repos.Staff.GetByID(req.RotationStaffID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if staff == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Rotation staff not found"})
		return
	}
	if staff.StaffType != models.StaffTypeRotation {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Staff must be rotation staff"})
		return
	}

	// Verify position is branch position
	position, err := h.repos.Position.GetByID(req.BranchPositionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if position == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Branch position not found"})
		return
	}
	if position.PositionType != models.PositionTypeBranch {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Position must be a branch position"})
		return
	}

	// Check if mapping already exists
	existing, err := h.repos.RotationStaffBranchPosition.GetByStaffAndPosition(req.RotationStaffID, req.BranchPositionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existing != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Mapping already exists"})
		return
	}

	mapping := &models.RotationStaffBranchPosition{
		RotationStaffID:   req.RotationStaffID,
		BranchPositionID:  req.BranchPositionID,
		SubstitutionLevel: req.SubstitutionLevel,
		IsActive:          req.IsActive,
		Notes:             req.Notes,
	}

	if err := h.repos.RotationStaffBranchPosition.Create(mapping); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"mapping": mapping})
}

type UpdateRotationStaffBranchPositionRequest struct {
	SubstitutionLevel int    `json:"substitution_level" binding:"required"`
	IsActive          bool   `json:"is_active"`
	Notes             string `json:"notes"`
}

func (h *RotationStaffBranchPositionHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var req UpdateRotationStaffBranchPositionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate substitution level
	if req.SubstitutionLevel < 1 || req.SubstitutionLevel > 3 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "substitution_level must be between 1 and 3"})
		return
	}

	mapping, err := h.repos.RotationStaffBranchPosition.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if mapping == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Mapping not found"})
		return
	}

	mapping.SubstitutionLevel = req.SubstitutionLevel
	mapping.IsActive = req.IsActive
	mapping.Notes = req.Notes

	if err := h.repos.RotationStaffBranchPosition.Update(mapping); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"mapping": mapping})
}

func (h *RotationStaffBranchPositionHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := h.repos.RotationStaffBranchPosition.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Mapping deleted successfully"})
}

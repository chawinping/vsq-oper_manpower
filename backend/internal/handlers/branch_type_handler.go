package handlers

import (
	"database/sql"
	"net/http"

	"vsq-oper-manpower/backend/internal/domain/models"
	"vsq-oper-manpower/backend/internal/repositories/postgres"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type BranchTypeHandler struct {
	repos *postgres.Repositories
	db    *sql.DB
}

func NewBranchTypeHandler(repos *postgres.Repositories, db *sql.DB) *BranchTypeHandler {
	return &BranchTypeHandler{repos: repos, db: db}
}

func (h *BranchTypeHandler) List(c *gin.Context) {
	branchTypes, err := h.repos.BranchType.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"branch_types": branchTypes})
}

func (h *BranchTypeHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	branchType, err := h.repos.BranchType.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if branchType == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Branch type not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"branch_type": branchType})
}

type CreateBranchTypeRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
}

func (h *BranchTypeHandler) Create(c *gin.Context) {
	var req CreateBranchTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	branchType := &models.BranchType{
		Name:        req.Name,
		Description: req.Description,
		IsActive:    req.IsActive,
	}

	if err := h.repos.BranchType.Create(branchType); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"branch_type": branchType})
}

type UpdateBranchTypeRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
}

func (h *BranchTypeHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var req UpdateBranchTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	branchType, err := h.repos.BranchType.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if branchType == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Branch type not found"})
		return
	}

	branchType.Name = req.Name
	branchType.Description = req.Description
	branchType.IsActive = req.IsActive

	if err := h.repos.BranchType.Update(branchType); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"branch_type": branchType})
}

func (h *BranchTypeHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := h.repos.BranchType.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Branch type deleted successfully"})
}

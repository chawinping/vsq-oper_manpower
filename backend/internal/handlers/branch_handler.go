package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"vsq-oper-manpower/backend/internal/domain/models"
	"vsq-oper-manpower/backend/internal/repositories/postgres"
)

type BranchHandler struct {
	repos *postgres.Repositories
}

func NewBranchHandler(repos *postgres.Repositories) *BranchHandler {
	return &BranchHandler{repos: repos}
}

type CreateBranchRequest struct {
	Name            string     `json:"name" binding:"required"`
	Code            string     `json:"code" binding:"required"`
	Address         string     `json:"address"`
	AreaManagerID   *uuid.UUID `json:"area_manager_id,omitempty"`
	ExpectedRevenue float64    `json:"expected_revenue"`
	Priority        int        `json:"priority"`
}

func (h *BranchHandler) List(c *gin.Context) {
	branches, err := h.repos.Branch.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"branches": branches})
}

func (h *BranchHandler) Create(c *gin.Context) {
	var req CreateBranchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	branch := &models.Branch{
		ID:              uuid.New(),
		Name:            req.Name,
		Code:            req.Code,
		Address:         req.Address,
		AreaManagerID:   req.AreaManagerID,
		ExpectedRevenue: req.ExpectedRevenue,
		Priority:        req.Priority,
	}

	if err := h.repos.Branch.Create(branch); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"branch": branch})
}

func (h *BranchHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var req CreateBranchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	branch := &models.Branch{
		ID:              id,
		Name:            req.Name,
		Code:            req.Code,
		Address:         req.Address,
		AreaManagerID:   req.AreaManagerID,
		ExpectedRevenue: req.ExpectedRevenue,
		Priority:        req.Priority,
	}

	if err := h.repos.Branch.Update(branch); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"branch": branch})
}

func (h *BranchHandler) GetRevenue(c *gin.Context) {
	idStr := c.Param("id")
	branchID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid branch ID"})
		return
	}

	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	var startDate, endDate time.Time
	if startDateStr != "" && endDateStr != "" {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format"})
			return
		}
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format"})
			return
		}
	} else {
		// Default to last 30 days
		endDate = time.Now()
		startDate = endDate.AddDate(0, 0, -30)
	}

	revenues, err := h.repos.Revenue.GetByBranchID(branchID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"revenue_data": revenues})
}



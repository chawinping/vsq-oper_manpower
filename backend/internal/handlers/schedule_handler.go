package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"vsq-oper-manpower/backend/internal/domain/models"
	"vsq-oper-manpower/backend/internal/repositories/postgres"
)

type ScheduleHandler struct {
	repos *postgres.Repositories
}

func NewScheduleHandler(repos *postgres.Repositories) *ScheduleHandler {
	return &ScheduleHandler{repos: repos}
}

type CreateScheduleRequest struct {
	StaffID      uuid.UUID `json:"staff_id" binding:"required"`
	BranchID     uuid.UUID `json:"branch_id" binding:"required"`
	Date         string    `json:"date" binding:"required"`
	IsWorkingDay bool      `json:"is_working_day"`
}

func (h *ScheduleHandler) GetBranchSchedule(c *gin.Context) {
	branchIDStr := c.Param("branchId")
	branchID, err := uuid.Parse(branchIDStr)
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
		// Default to next 30 days
		startDate = time.Now()
		endDate = startDate.AddDate(0, 0, 30)
	}

	schedules, err := h.repos.Schedule.GetByBranchID(branchID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"schedules": schedules})
}

func (h *ScheduleHandler) Create(c *gin.Context) {
	var req CreateScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format"})
		return
	}

	userIDStr := c.MustGet("user_id").(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	schedule := &models.StaffSchedule{
		ID:           uuid.New(),
		StaffID:      req.StaffID,
		BranchID:     req.BranchID,
		Date:         date,
		IsWorkingDay: req.IsWorkingDay,
		CreatedBy:    userID,
	}

	if err := h.repos.Schedule.Create(schedule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"schedule": schedule})
}

func (h *ScheduleHandler) GetMonthlyView(c *gin.Context) {
	branchIDStr := c.Query("branch_id")
	yearStr := c.Query("year")
	monthStr := c.Query("month")

	if branchIDStr == "" || yearStr == "" || monthStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "branch_id, year, and month are required"})
		return
	}

	branchID, err := uuid.Parse(branchIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid branch ID"})
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid year"})
		return
	}

	month, err := strconv.Atoi(monthStr)
	if err != nil || month < 1 || month > 12 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid month"})
		return
	}

	schedules, err := h.repos.Schedule.GetMonthlyView(branchID, year, month)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"schedules": schedules})
}


package handlers

import (
	"fmt"
	"net/http"
	"time"

	"vsq-oper-manpower/backend/internal/repositories/postgres"
	"vsq-oper-manpower/backend/internal/usecases/test_data"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TestDataHandler struct {
	repos     *postgres.Repositories
	generator *test_data.ScheduleGenerator
}

func NewTestDataHandler(repos *postgres.Repositories) *TestDataHandler {
	return &TestDataHandler{
		repos:     repos,
		generator: test_data.NewScheduleGenerator(repos),
	}
}

type GenerateScheduleRequest struct {
	StartDate         string                  `json:"start_date" binding:"required"` // YYYY-MM-DD
	EndDate           string                  `json:"end_date" binding:"required"`   // YYYY-MM-DD
	Rules             test_data.ScheduleRules `json:"rules" binding:"required"`
	OverwriteExisting bool                    `json:"overwrite_existing"` // Whether to overwrite existing schedules
	BranchIDs         []string                `json:"branch_ids,omitempty"` // Optional: filter by specific branch IDs. If empty, generates for all branches.
}

// GenerateSchedules generates schedules for all branch staff
func (h *TestDataHandler) GenerateSchedules(c *gin.Context) {
	var req GenerateScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse dates
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format. Use YYYY-MM-DD"})
		return
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format. Use YYYY-MM-DD"})
		return
	}

	if endDate.Before(startDate) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "end_date must be after start_date"})
		return
	}

	// Validate rules
	if req.Rules.MinWorkingDaysPerWeek < 0 || req.Rules.MinWorkingDaysPerWeek > 7 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "min_working_days_per_week must be between 0 and 7"})
		return
	}
	if req.Rules.MaxWorkingDaysPerWeek < 0 || req.Rules.MaxWorkingDaysPerWeek > 7 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "max_working_days_per_week must be between 0 and 7"})
		return
	}
	if req.Rules.MinWorkingDaysPerWeek > req.Rules.MaxWorkingDaysPerWeek {
		c.JSON(http.StatusBadRequest, gin.H{"error": "min_working_days_per_week must be <= max_working_days_per_week"})
		return
	}
	if req.Rules.LeaveProbability < 0 || req.Rules.LeaveProbability > 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "leave_probability must be between 0.0 and 1.0"})
		return
	}
	if req.Rules.WeekendWorkingRatio < 0 || req.Rules.WeekendWorkingRatio > 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "weekend_working_ratio must be between 0.0 and 1.0"})
		return
	}
	if req.Rules.ConsecutiveLeaveMax < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "consecutive_leave_max must be >= 0"})
		return
	}
	if req.Rules.MinOffDaysPerMonth < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "min_off_days_per_month must be >= 0"})
		return
	}
	if req.Rules.MaxOffDaysPerMonth > 0 && req.Rules.MaxOffDaysPerMonth < req.Rules.MinOffDaysPerMonth {
		c.JSON(http.StatusBadRequest, gin.H{"error": "max_off_days_per_month must be >= min_off_days_per_month"})
		return
	}

	// Get user ID from context
	userIDStr := c.MustGet("user_id").(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Parse branch IDs if provided
	var branchIDs []uuid.UUID
	if len(req.BranchIDs) > 0 {
		branchIDs = make([]uuid.UUID, 0, len(req.BranchIDs))
		for _, branchIDStr := range req.BranchIDs {
			branchID, err := uuid.Parse(branchIDStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid branch_id format: %s", branchIDStr)})
				return
			}
			branchIDs = append(branchIDs, branchID)
		}
	}

	// Generate schedules
	generateReq := test_data.GenerateSchedulesRequest{
		StartDate:         startDate,
		EndDate:           endDate,
		Rules:             req.Rules,
		OverwriteExisting: req.OverwriteExisting,
		CreatedBy:         userID,
		BranchIDs:         branchIDs, // nil or empty means all branches
	}

	result, err := h.generator.GenerateSchedules(generateReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Schedules generated successfully",
		"result":  result,
	})
}

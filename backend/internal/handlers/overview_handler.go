package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"vsq-oper-manpower/backend/internal/repositories/postgres"
	"vsq-oper-manpower/backend/internal/usecases/allocation"
)

type OverviewHandler struct {
	repos            *postgres.Repositories
	overviewGenerator *allocation.OverviewGenerator
}

func NewOverviewHandler(repos *postgres.Repositories, overviewGenerator *allocation.OverviewGenerator) *OverviewHandler {
	return &OverviewHandler{
		repos:             repos,
		overviewGenerator: overviewGenerator,
	}
}

// GetDayOverview returns overview for specified branches on a specific day
// Query params:
//   - date: date in YYYY-MM-DD format (defaults to today)
//   - branch_ids: comma-separated list of branch UUIDs (optional, defaults to all branches)
func (h *OverviewHandler) GetDayOverview(c *gin.Context) {
	dateStr := c.Query("date")
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format"})
		return
	}

	// Parse branch IDs if provided
	var branchIDs []uuid.UUID
	branchIDsStr := c.Query("branch_ids")
	if branchIDsStr != "" {
		// Split by comma and parse each UUID
		parts := []string{}
		current := ""
		for _, char := range branchIDsStr {
			if char == ',' {
				if current != "" {
					parts = append(parts, current)
					current = ""
				}
			} else if char != ' ' {
				current += string(char)
			}
		}
		if current != "" {
			parts = append(parts, current)
		}
		
		branchIDs = make([]uuid.UUID, 0, len(parts))
		for _, part := range parts {
			id, err := uuid.Parse(part)
			if err == nil {
				branchIDs = append(branchIDs, id)
			}
		}
	}

	overview, err := h.overviewGenerator.GenerateDayOverview(date, branchIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"overview": overview})
}

// GetMonthlyOverview returns overview for a single branch across a month
func (h *OverviewHandler) GetMonthlyOverview(c *gin.Context) {
	branchIDStr := c.Query("branch_id")
	if branchIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "branch_id is required"})
		return
	}

	branchID, err := uuid.Parse(branchIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid branch_id"})
		return
	}

	yearStr := c.Query("year")
	monthStr := c.Query("month")

	var year, month int
	if yearStr == "" || monthStr == "" {
		now := time.Now()
		year = now.Year()
		month = int(now.Month())
	} else {
		year, err = strconv.Atoi(yearStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid year"})
			return
		}
		month, err = strconv.Atoi(monthStr)
		if err != nil || month < 1 || month > 12 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid month"})
			return
		}
	}

	// For branch managers, enforce their branch
	role := c.GetString("role")
	if role == "branch_manager" {
		userBranchID, exists := c.Get("user_branch_id")
		if exists {
			if userBranchUUID, ok := userBranchID.(uuid.UUID); ok {
				if userBranchUUID != branchID {
					c.JSON(http.StatusForbidden, gin.H{"error": "You can only access overview for your own branch"})
					return
				}
			}
		}
	}

	overview, err := h.overviewGenerator.GenerateMonthlyOverview(branchID, year, month)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"overview": overview})
}

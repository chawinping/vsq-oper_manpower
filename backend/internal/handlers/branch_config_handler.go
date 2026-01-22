package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"vsq-oper-manpower/backend/internal/domain/models"
	"vsq-oper-manpower/backend/internal/repositories/postgres"
)

type BranchConfigHandler struct {
	repos *postgres.Repositories
}

func NewBranchConfigHandler(repos *postgres.Repositories) *BranchConfigHandler {
	return &BranchConfigHandler{repos: repos}
}

// GetBranchConfig returns full branch configuration (quotas + weekly revenue)
func (h *BranchConfigHandler) GetBranchConfig(c *gin.Context) {
	branchIDStr := c.Param("id")
	branchID, err := uuid.Parse(branchIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid branch ID"})
		return
	}

	// Get quotas
	quotas, err := h.repos.PositionQuota.GetByBranchID(branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get weekly revenue
	weeklyRevenue, err := h.repos.BranchWeeklyRevenue.GetByBranchID(branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get positions for quota details
	positions, err := h.repos.Position.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	positionMap := make(map[uuid.UUID]*models.Position)
	for _, pos := range positions {
		positionMap[pos.ID] = pos
	}

	// Build quota response with position names
	quotaResponse := []map[string]interface{}{}
	for _, quota := range quotas {
		position := positionMap[quota.PositionID]
		quotaResponse = append(quotaResponse, map[string]interface{}{
			"id":                quota.ID,
			"position_id":       quota.PositionID,
			"position_name":     getPositionName(position),
			"designated_quota":  quota.DesignatedQuota,
			"minimum_required":  quota.MinimumRequired,
			"is_active":         quota.IsActive,
		})
	}

	// Get constraints
	constraints, err := h.repos.BranchConstraints.GetByBranchID(branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"quotas":         quotaResponse,
		"weekly_revenue": weeklyRevenue,
		"constraints":    constraints,
	})
}

// UpdateQuotas handles bulk update of position quotas
type UpdateQuotasRequest struct {
	Quotas []PositionQuotaUpdate `json:"quotas" binding:"required"`
}

type PositionQuotaUpdate struct {
	PositionID      uuid.UUID `json:"position_id" binding:"required"`
	DesignatedQuota int       `json:"designated_quota" binding:"required"`
	MinimumRequired int       `json:"minimum_required" binding:"required"`
}

func (h *BranchConfigHandler) UpdateQuotas(c *gin.Context) {
	branchIDStr := c.Param("id")
	branchID, err := uuid.Parse(branchIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid branch ID"})
		return
	}

	var req UpdateQuotasRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDStr := c.MustGet("user_id").(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get positions to validate position types
	positions, err := h.repos.Position.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	positionMap := make(map[uuid.UUID]*models.Position)
	for _, pos := range positions {
		positionMap[pos.ID] = pos
	}

	// Validate minimum_required <= designated_quota and position types
	for _, quota := range req.Quotas {
		// Check if position exists
		position := positionMap[quota.PositionID]
		if position == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Position not found: %s", quota.PositionID)})
			return
		}

		// Reject quota updates for rotation-type positions
		if position.PositionType == models.PositionTypeRotation {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("Cannot set quotas for rotation-type position: %s", position.Name),
			})
			return
		}

		// Validate quota values
		if quota.MinimumRequired > quota.DesignatedQuota {
			c.JSON(http.StatusBadRequest, gin.H{"error": "minimum_required cannot be greater than designated_quota"})
			return
		}
		if quota.DesignatedQuota < 0 || quota.MinimumRequired < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "quota values cannot be negative"})
			return
		}
	}

	// Upsert quotas
	for _, quotaReq := range req.Quotas {
		// Check if quota exists
		existingQuota, err := h.repos.PositionQuota.GetByBranchAndPosition(branchID, quotaReq.PositionID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if existingQuota != nil {
			// Update existing
			existingQuota.DesignatedQuota = quotaReq.DesignatedQuota
			existingQuota.MinimumRequired = quotaReq.MinimumRequired
			existingQuota.IsActive = true
			if err := h.repos.PositionQuota.Update(existingQuota); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		} else {
			// Create new
			quota := &models.PositionQuota{
				ID:              uuid.New(),
				BranchID:        branchID,
				PositionID:      quotaReq.PositionID,
				DesignatedQuota: quotaReq.DesignatedQuota,
				MinimumRequired: quotaReq.MinimumRequired,
				IsActive:        true,
				CreatedBy:       userID,
			}
			if err := h.repos.PositionQuota.Create(quota); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Quotas updated successfully"})
}

// UpdateWeeklyRevenue handles bulk update of weekly revenue
type UpdateWeeklyRevenueRequest struct {
	WeeklyRevenue []WeeklyRevenueUpdate `json:"weekly_revenue" binding:"required"`
}

type WeeklyRevenueUpdate struct {
	DayOfWeek       int     `json:"day_of_week" binding:"required"` // 0=Sunday, 6=Saturday
	ExpectedRevenue float64 `json:"expected_revenue,omitempty"`     // Deprecated: Use SkinRevenue instead
	SkinRevenue     float64 `json:"skin_revenue"`                   // Skin revenue (THB)
	LSHMRevenue     float64 `json:"ls_hm_revenue"`                  // LS HM revenue (THB)
	VitaminCases    int     `json:"vitamin_cases"`                   // Vitamin cases (count)
	SlimPenCases    int     `json:"slim_pen_cases"`                 // Slim Pen cases (count)
}

func (h *BranchConfigHandler) UpdateWeeklyRevenue(c *gin.Context) {
	branchIDStr := c.Param("id")
	branchID, err := uuid.Parse(branchIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid branch ID"})
		return
	}

	var req UpdateWeeklyRevenueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate day_of_week and revenue values
	for _, revenue := range req.WeeklyRevenue {
		if revenue.DayOfWeek < 0 || revenue.DayOfWeek > 6 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "day_of_week must be between 0 and 6"})
			return
		}
		if revenue.ExpectedRevenue < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "expected_revenue cannot be negative"})
			return
		}
		if revenue.SkinRevenue < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "skin_revenue cannot be negative"})
			return
		}
		if revenue.LSHMRevenue < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ls_hm_revenue cannot be negative"})
			return
		}
		if revenue.VitaminCases < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "vitamin_cases cannot be negative"})
			return
		}
		if revenue.SlimPenCases < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "slim_pen_cases cannot be negative"})
			return
		}
	}

	// Convert to models
	revenues := []*models.BranchWeeklyRevenue{}
	for _, rev := range req.WeeklyRevenue {
		revenues = append(revenues, &models.BranchWeeklyRevenue{
			ID:              uuid.New(),
			BranchID:        branchID,
			DayOfWeek:       rev.DayOfWeek,
			ExpectedRevenue: rev.ExpectedRevenue, // Keep for backward compatibility
			SkinRevenue:     rev.SkinRevenue,
			LSHMRevenue:     rev.LSHMRevenue,
			VitaminCases:    rev.VitaminCases,
			SlimPenCases:    rev.SlimPenCases,
		})
	}

	// Bulk upsert
	if err := h.repos.BranchWeeklyRevenue.BulkUpsert(revenues); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Weekly revenue updated successfully"})
}

// GetQuotas returns position quotas for a branch
func (h *BranchConfigHandler) GetQuotas(c *gin.Context) {
	branchIDStr := c.Param("id")
	branchID, err := uuid.Parse(branchIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid branch ID"})
		return
	}

	quotas, err := h.repos.PositionQuota.GetByBranchID(branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get positions for names
	positions, err := h.repos.Position.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	positionMap := make(map[uuid.UUID]*models.Position)
	for _, pos := range positions {
		positionMap[pos.ID] = pos
	}

	quotaResponse := []map[string]interface{}{}
	for _, quota := range quotas {
		position := positionMap[quota.PositionID]
		quotaResponse = append(quotaResponse, map[string]interface{}{
			"id":                quota.ID,
			"position_id":       quota.PositionID,
			"position_name":     getPositionName(position),
			"designated_quota":  quota.DesignatedQuota,
			"minimum_required":  quota.MinimumRequired,
			"is_active":         quota.IsActive,
		})
	}

	c.JSON(http.StatusOK, gin.H{"quotas": quotaResponse})
}

// GetWeeklyRevenue returns weekly revenue for a branch
func (h *BranchConfigHandler) GetWeeklyRevenue(c *gin.Context) {
	branchIDStr := c.Param("id")
	branchID, err := uuid.Parse(branchIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid branch ID"})
		return
	}

	weeklyRevenue, err := h.repos.BranchWeeklyRevenue.GetByBranchID(branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"weekly_revenue": weeklyRevenue})
}

// UpdateConstraints handles bulk update of branch constraints
type UpdateConstraintsRequest struct {
	Constraints []ConstraintUpdate `json:"constraints" binding:"required"`
}

type ConstraintUpdate struct {
	DayOfWeek         int `json:"day_of_week" binding:"required"` // 0=Sunday, 6=Saturday
	MinFrontStaff     int `json:"min_front_staff" binding:"required"`
	MinManagers       int `json:"min_managers" binding:"required"`
	MinDoctorAssistant int `json:"min_doctor_assistant" binding:"required"`
	MinTotalStaff     int `json:"min_total_staff" binding:"required"`
}

func (h *BranchConfigHandler) UpdateConstraints(c *gin.Context) {
	branchIDStr := c.Param("id")
	branchID, err := uuid.Parse(branchIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid branch ID"})
		return
	}

	var req UpdateConstraintsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate day_of_week and constraint values
	for _, constraint := range req.Constraints {
		if constraint.DayOfWeek < 0 || constraint.DayOfWeek > 6 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "day_of_week must be between 0 and 6"})
			return
		}
		if constraint.MinFrontStaff < 0 || constraint.MinManagers < 0 || constraint.MinDoctorAssistant < 0 || constraint.MinTotalStaff < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "constraint values cannot be negative"})
			return
		}
	}

	// Convert to models
	constraints := []*models.BranchConstraints{}
	for _, cons := range req.Constraints {
		constraints = append(constraints, &models.BranchConstraints{
			ID:                uuid.New(),
			BranchID:          branchID,
			DayOfWeek:         cons.DayOfWeek,
			MinFrontStaff:     cons.MinFrontStaff,
			MinManagers:       cons.MinManagers,
			MinDoctorAssistant: cons.MinDoctorAssistant,
			MinTotalStaff:     cons.MinTotalStaff,
		})
	}

	// Bulk upsert
	if err := h.repos.BranchConstraints.BulkUpsert(constraints); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Constraints updated successfully"})
}

// GetConstraints returns branch constraints
func (h *BranchConfigHandler) GetConstraints(c *gin.Context) {
	branchIDStr := c.Param("id")
	branchID, err := uuid.Parse(branchIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid branch ID"})
		return
	}

	constraints, err := h.repos.BranchConstraints.GetByBranchID(branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"constraints": constraints})
}

// Helper function to get position name
func getPositionName(position *models.Position) string {
	if position == nil {
		return "Unknown"
	}
	return position.Name
}

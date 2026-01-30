package handlers

import (
	"fmt"
	"net/http"

	"vsq-oper-manpower/backend/internal/domain/models"
	"vsq-oper-manpower/backend/internal/repositories/postgres"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
			"id":               quota.ID,
			"position_id":      quota.PositionID,
			"position_name":    getPositionName(position),
			"designated_quota": quota.DesignatedQuota,
			"minimum_required": quota.MinimumRequired,
			"is_active":        quota.IsActive,
		})
	}

	// Get constraints with inheritance resolution
	constraints, err := h.getResolvedConstraints(branchID)
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
	VitaminCases    int     `json:"vitamin_cases"`                  // Vitamin cases (count)
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
			"id":               quota.ID,
			"position_id":      quota.PositionID,
			"position_name":    getPositionName(position),
			"designated_quota": quota.DesignatedQuota,
			"minimum_required": quota.MinimumRequired,
			"is_active":        quota.IsActive,
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

type StaffGroupRequirement struct {
	StaffGroupID uuid.UUID `json:"staff_group_id" binding:"required"`
	MinimumCount int       `json:"minimum_count" binding:"required,min=0"`
}

type ConstraintUpdate struct {
	DayOfWeek              int                     `json:"day_of_week" binding:"required"` // 0=Sunday, 6=Saturday
	StaffGroupRequirements []StaffGroupRequirement `json:"staff_group_requirements" binding:"required"`
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

	// Get branch to check for branch type
	branch, err := h.repos.Branch.GetByID(branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if branch == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Branch not found"})
		return
	}

	// Validate day_of_week and constraint values
	for _, constraint := range req.Constraints {
		if constraint.DayOfWeek < 0 || constraint.DayOfWeek > 6 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "day_of_week must be between 0 and 6"})
			return
		}
		// Validate staff group requirements
		for _, sgReq := range constraint.StaffGroupRequirements {
			if sgReq.MinimumCount < 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "minimum_count cannot be negative"})
				return
			}
			// Verify staff group exists
			staffGroup, err := h.repos.StaffGroup.GetByID(sgReq.StaffGroupID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify staff group: " + err.Error()})
				return
			}
			if staffGroup == nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Staff group not found: " + sgReq.StaffGroupID.String()})
				return
			}
			if !staffGroup.IsActive {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Staff group is not active: " + staffGroup.Name})
				return
			}
		}
	}

	// Convert to models and mark as overridden
	constraints := []*models.BranchConstraints{}
	for _, cons := range req.Constraints {
		constraint := &models.BranchConstraints{
			ID:                        uuid.New(),
			BranchID:                  branchID,
			DayOfWeek:                 cons.DayOfWeek,
			IsOverridden:              true,                // Mark as overridden when explicitly set
			InheritedFromBranchTypeID: branch.BranchTypeID, // Track which branch type this overrides
		}

		// Convert staff group requirements
		if len(cons.StaffGroupRequirements) > 0 {
			constraint.StaffGroupRequirements = make([]*models.BranchConstraintStaffGroup, len(cons.StaffGroupRequirements))
			for i, sgReq := range cons.StaffGroupRequirements {
				constraint.StaffGroupRequirements[i] = &models.BranchConstraintStaffGroup{
					StaffGroupID: sgReq.StaffGroupID,
					MinimumCount: sgReq.MinimumCount,
				}
			}
		}

		constraints = append(constraints, constraint)
	}

	// Bulk upsert with staff groups
	if err := h.repos.BranchConstraints.BulkUpsertWithStaffGroups(constraints); err != nil {
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

	constraints, err := h.getResolvedConstraints(branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"constraints": constraints})
}

// getResolvedConstraints returns constraints for a branch with inheritance resolution
// Priority: Overridden branch constraints > Branch type constraints > Defaults (all zeros)
func (h *BranchConfigHandler) getResolvedConstraints(branchID uuid.UUID) ([]*models.BranchConstraints, error) {
	// Get branch to check for branch type
	branch, err := h.repos.Branch.GetByID(branchID)
	if err != nil {
		return nil, err
	}
	if branch == nil {
		return nil, fmt.Errorf("branch not found")
	}

	// Get branch-specific constraints (overridden ones)
	branchConstraints, err := h.repos.BranchConstraints.GetByBranchID(branchID)
	if err != nil {
		return nil, err
	}
	// Load staff group requirements for branch constraints
	if err := h.repos.BranchConstraints.LoadStaffGroupRequirements(branchConstraints); err != nil {
		return nil, fmt.Errorf("failed to load staff group requirements: %w", err)
	}

	// Create a map of overridden constraints by day_of_week
	overriddenMap := make(map[int]*models.BranchConstraints)
	for _, constraint := range branchConstraints {
		if constraint.IsOverridden {
			overriddenMap[constraint.DayOfWeek] = constraint
		}
	}

	// Get branch type constraints if branch has a branch type
	var branchTypeConstraints []*models.BranchTypeConstraints
	if branch.BranchTypeID != nil {
		branchTypeConstraints, err = h.repos.BranchTypeConstraints.GetByBranchTypeID(*branch.BranchTypeID)
		if err != nil {
			return nil, err
		}
		// Load staff group requirements for branch type constraints
		if err := h.repos.BranchTypeConstraints.LoadStaffGroupRequirements(branchTypeConstraints); err != nil {
			return nil, fmt.Errorf("failed to load staff group requirements: %w", err)
		}
	}

	// Create a map of branch type constraints by day_of_week
	branchTypeMap := make(map[int]*models.BranchTypeConstraints)
	for _, constraint := range branchTypeConstraints {
		branchTypeMap[constraint.DayOfWeek] = constraint
	}

	// Build resolved constraints for all 7 days
	resolvedConstraints := make([]*models.BranchConstraints, 7)
	for dayOfWeek := 0; dayOfWeek < 7; dayOfWeek++ {
		constraint := &models.BranchConstraints{
			ID:        uuid.New(),
			BranchID:  branchID,
			DayOfWeek: dayOfWeek,
		}

		// Priority 1: Overridden branch constraint
		if overridden, exists := overriddenMap[dayOfWeek]; exists {
			// Copy staff group requirements from overridden constraint
			if overridden.StaffGroupRequirements != nil && len(overridden.StaffGroupRequirements) > 0 {
				constraint.StaffGroupRequirements = make([]*models.BranchConstraintStaffGroup, len(overridden.StaffGroupRequirements))
				for i, sg := range overridden.StaffGroupRequirements {
					constraint.StaffGroupRequirements[i] = &models.BranchConstraintStaffGroup{
						StaffGroupID: sg.StaffGroupID,
						MinimumCount: sg.MinimumCount,
					}
				}
			}
			constraint.IsOverridden = true
			constraint.InheritedFromBranchTypeID = overridden.InheritedFromBranchTypeID
		} else if branchType, exists := branchTypeMap[dayOfWeek]; exists {
			// Priority 2: Branch type constraint - copy staff group requirements
			if branchType.StaffGroupRequirements != nil && len(branchType.StaffGroupRequirements) > 0 {
				constraint.StaffGroupRequirements = make([]*models.BranchConstraintStaffGroup, len(branchType.StaffGroupRequirements))
				for i, sg := range branchType.StaffGroupRequirements {
					constraint.StaffGroupRequirements[i] = &models.BranchConstraintStaffGroup{
						StaffGroupID: sg.StaffGroupID,
						MinimumCount: sg.MinimumCount,
					}
				}
			}
			constraint.IsOverridden = false
			constraint.InheritedFromBranchTypeID = branch.BranchTypeID
		} else {
			// Priority 3: Defaults (empty staff group requirements)
			constraint.StaffGroupRequirements = []*models.BranchConstraintStaffGroup{}
			constraint.IsOverridden = false
			constraint.InheritedFromBranchTypeID = branch.BranchTypeID
		}

		resolvedConstraints[dayOfWeek] = constraint
	}

	return resolvedConstraints, nil
}

// Helper function to get position name
func getPositionName(position *models.Position) string {
	if position == nil {
		return "Unknown"
	}
	return position.Name
}

package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

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
	IsOverridden           *bool                   `json:"is_overridden,omitempty"` // If false, delete constraint to inherit from branch type
}

func (h *BranchConfigHandler) UpdateConstraints(c *gin.Context) {
	branchIDStr := c.Param("id")
	branchID, err := uuid.Parse(branchIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid branch ID"})
		return
	}

	// #region agent log - Read raw request body before parsing
	var rawBody []byte
	if c.Request.Body != nil {
		rawBody, _ = io.ReadAll(c.Request.Body)
		if logFile, err2 := os.OpenFile("c:\\Users\\User\\dev_projects\\vsq-oper_manpower\\.cursor\\debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err2 == nil {
			logData, _ := json.Marshal(map[string]interface{}{"location": "branch_config_handler.go:363", "message": "UpdateConstraints: raw request body", "data": map[string]interface{}{"branchId": branchID.String(), "rawBody": string(rawBody)}, "timestamp": fmt.Sprintf("%d", 0), "sessionId": "debug-session", "runId": "run1", "hypothesisId": "B"})
			logFile.WriteString(string(logData) + "\n")
			logFile.Close()
		}
		c.Request.Body = io.NopCloser(bytes.NewBuffer(rawBody))
	}
	// #endregion

	var req UpdateConstraintsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// #region agent log
	if logFile, err := os.OpenFile("c:\\Users\\User\\dev_projects\\vsq-oper_manpower\\.cursor\\debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
		logData, _ := json.Marshal(map[string]interface{}{"location": "branch_config_handler.go:375", "message": "UpdateConstraints: parsed request", "data": map[string]interface{}{"branchId": branchID.String(), "constraintsCount": len(req.Constraints), "constraints": func() []map[string]interface{} {
			var result []map[string]interface{}
			for _, c := range req.Constraints {
				isOverridden := "nil"
				if c.IsOverridden != nil {
					isOverridden = fmt.Sprintf("%v", *c.IsOverridden)
				}
				result = append(result, map[string]interface{}{"day": c.DayOfWeek, "is_overridden": isOverridden, "isOverriddenPtr": c.IsOverridden != nil, "reqCount": len(c.StaffGroupRequirements)})
			}
			return result
		}()}, "timestamp": fmt.Sprintf("%d", 0), "sessionId": "debug-session", "runId": "run1", "hypothesisId": "B"})
		logFile.WriteString(string(logData) + "\n")
		logFile.Close()
	}
	// #endregion

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
	// Skip validation for constraints that will be deleted (is_overridden = false)
	for _, constraint := range req.Constraints {
		if constraint.DayOfWeek < 0 || constraint.DayOfWeek > 6 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "day_of_week must be between 0 and 6"})
			return
		}
		// Skip staff group validation if this constraint will be deleted
		if constraint.IsOverridden != nil && !*constraint.IsOverridden {
			continue
		}
		// Validate staff group requirements only for constraints that will be updated
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

	// Separate constraints to update vs delete
	constraintsToUpdate := []*models.BranchConstraints{}
	daysToDelete := []int{}

	for _, cons := range req.Constraints {
		// If is_overridden is explicitly false, delete the constraint to inherit from branch type
		if cons.IsOverridden != nil && !*cons.IsOverridden {
			daysToDelete = append(daysToDelete, cons.DayOfWeek)
			// #region agent log
			if logFile, err := os.OpenFile("c:\\Users\\User\\dev_projects\\vsq-oper_manpower\\.cursor\\debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
				logData, _ := json.Marshal(map[string]interface{}{"location": "branch_config_handler.go:414", "message": "UpdateConstraints: marking constraint for deletion", "data": map[string]interface{}{"branchId": branchID.String(), "dayOfWeek": cons.DayOfWeek}, "timestamp": fmt.Sprintf("%d", 0), "sessionId": "debug-session", "runId": "run1", "hypothesisId": "C"})
				logFile.WriteString(string(logData) + "\n")
				logFile.Close()
			}
			// #endregion
			continue
		}

		// Otherwise, mark as overridden (default behavior)
		// BUT: Only create overridden constraint if it has non-empty staff group requirements
		// If staff_group_requirements is empty, skip it (don't create an overridden constraint with zeros)
		if len(cons.StaffGroupRequirements) == 0 {
			// #region agent log
			if logFile, err := os.OpenFile("c:\\Users\\User\\dev_projects\\vsq-oper_manpower\\.cursor\\debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
				logData, _ := json.Marshal(map[string]interface{}{"location": "branch_config_handler.go:449", "message": "UpdateConstraints: skipping constraint with empty requirements", "data": map[string]interface{}{"branchId": branchID.String(), "dayOfWeek": cons.DayOfWeek}, "timestamp": fmt.Sprintf("%d", 0), "sessionId": "debug-session", "runId": "run1", "hypothesisId": "D"})
				logFile.WriteString(string(logData) + "\n")
				logFile.Close()
			}
			// #endregion
			// Skip creating overridden constraint with empty requirements
			// This constraint should inherit from branch type (or be deleted if it exists)
			// Check if there's an existing overridden constraint and delete it
			existingConstraint, err := h.repos.BranchConstraints.GetByBranchIDAndDayOfWeek(branchID, cons.DayOfWeek)
			if err == nil && existingConstraint != nil && existingConstraint.IsOverridden {
				// Delete existing overridden constraint so it can inherit from branch type
				if err := h.repos.BranchConstraints.Delete(existingConstraint.ID); err != nil {
					// #region agent log
					if logFile, err2 := os.OpenFile("c:\\Users\\User\\dev_projects\\vsq-oper_manpower\\.cursor\\debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err2 == nil {
						logData, _ := json.Marshal(map[string]interface{}{"location": "branch_config_handler.go:456", "message": "UpdateConstraints: failed to delete empty constraint", "data": map[string]interface{}{"branchId": branchID.String(), "dayOfWeek": cons.DayOfWeek, "error": err.Error()}, "timestamp": fmt.Sprintf("%d", 0), "sessionId": "debug-session", "runId": "run1", "hypothesisId": "D"})
						logFile.WriteString(string(logData) + "\n")
						logFile.Close()
					}
					// #endregion
				}
			}
			continue
		}

		constraint := &models.BranchConstraints{
			ID:                        uuid.New(),
			BranchID:                  branchID,
			DayOfWeek:                 cons.DayOfWeek,
			IsOverridden:              true,                // Mark as overridden when explicitly set
			InheritedFromBranchTypeID: branch.BranchTypeID, // Track which branch type this overrides
		}

		// Convert staff group requirements
		constraint.StaffGroupRequirements = make([]*models.BranchConstraintStaffGroup, len(cons.StaffGroupRequirements))
		for i, sgReq := range cons.StaffGroupRequirements {
			constraint.StaffGroupRequirements[i] = &models.BranchConstraintStaffGroup{
				StaffGroupID: sgReq.StaffGroupID,
				MinimumCount: sgReq.MinimumCount,
			}
		}

		constraintsToUpdate = append(constraintsToUpdate, constraint)
	}

	// Delete constraints that should inherit from branch type
	// IMPORTANT: Delete BEFORE upserting to avoid conflicts
	if len(daysToDelete) > 0 {
		// #region agent log
		if logFile, err := os.OpenFile("c:\\Users\\User\\dev_projects\\vsq-oper_manpower\\.cursor\\debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
			logData, _ := json.Marshal(map[string]interface{}{"location": "branch_config_handler.go:454", "message": "UpdateConstraints: starting deletion", "data": map[string]interface{}{"branchId": branchID.String(), "daysToDelete": daysToDelete, "daysToDeleteCount": len(daysToDelete)}, "timestamp": fmt.Sprintf("%d", 0), "sessionId": "debug-session", "runId": "run1", "hypothesisId": "C"})
			logFile.WriteString(string(logData) + "\n")
			logFile.Close()
		}
		// #endregion
		for _, dayOfWeek := range daysToDelete {
			existingConstraint, err := h.repos.BranchConstraints.GetByBranchIDAndDayOfWeek(branchID, dayOfWeek)
			if err != nil {
				// #region agent log
				if logFile, err2 := os.OpenFile("c:\\Users\\User\\dev_projects\\vsq-oper_manpower\\.cursor\\debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err2 == nil {
					logData, _ := json.Marshal(map[string]interface{}{"location": "branch_config_handler.go:462", "message": "UpdateConstraints: error checking constraint", "data": map[string]interface{}{"branchId": branchID.String(), "dayOfWeek": dayOfWeek, "error": err.Error()}, "timestamp": fmt.Sprintf("%d", 0), "sessionId": "debug-session", "runId": "run1", "hypothesisId": "C"})
					logFile.WriteString(string(logData) + "\n")
					logFile.Close()
				}
				// #endregion
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check existing constraint: " + err.Error()})
				return
			}
			if existingConstraint != nil {
				// #region agent log
				if logFile, err2 := os.OpenFile("c:\\Users\\User\\dev_projects\\vsq-oper_manpower\\.cursor\\debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err2 == nil {
					logData, _ := json.Marshal(map[string]interface{}{"location": "branch_config_handler.go:470", "message": "UpdateConstraints: deleting constraint", "data": map[string]interface{}{"branchId": branchID.String(), "dayOfWeek": dayOfWeek, "constraintId": existingConstraint.ID.String(), "isOverridden": existingConstraint.IsOverridden}, "timestamp": fmt.Sprintf("%d", 0), "sessionId": "debug-session", "runId": "run1", "hypothesisId": "C"})
					logFile.WriteString(string(logData) + "\n")
					logFile.Close()
				}
				// #endregion
				if err := h.repos.BranchConstraints.Delete(existingConstraint.ID); err != nil {
					// #region agent log
					if logFile, err2 := os.OpenFile("c:\\Users\\User\\dev_projects\\vsq-oper_manpower\\.cursor\\debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err2 == nil {
						logData, _ := json.Marshal(map[string]interface{}{"location": "branch_config_handler.go:473", "message": "UpdateConstraints: deletion failed", "data": map[string]interface{}{"branchId": branchID.String(), "dayOfWeek": dayOfWeek, "error": err.Error()}, "timestamp": fmt.Sprintf("%d", 0), "sessionId": "debug-session", "runId": "run1", "hypothesisId": "C"})
						logFile.WriteString(string(logData) + "\n")
						logFile.Close()
					}
					// #endregion
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete constraint: " + err.Error()})
					return
				}
				// #region agent log
				if logFile, err2 := os.OpenFile("c:\\Users\\User\\dev_projects\\vsq-oper_manpower\\.cursor\\debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err2 == nil {
					logData, _ := json.Marshal(map[string]interface{}{"location": "branch_config_handler.go:477", "message": "UpdateConstraints: constraint deleted successfully", "data": map[string]interface{}{"branchId": branchID.String(), "dayOfWeek": dayOfWeek}, "timestamp": fmt.Sprintf("%d", 0), "sessionId": "debug-session", "runId": "run1", "hypothesisId": "C"})
					logFile.WriteString(string(logData) + "\n")
					logFile.Close()
				}
				// #endregion
			} else {
				// #region agent log
				if logFile, err2 := os.OpenFile("c:\\Users\\User\\dev_projects\\vsq-oper_manpower\\.cursor\\debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err2 == nil {
					logData, _ := json.Marshal(map[string]interface{}{"location": "branch_config_handler.go:481", "message": "UpdateConstraints: constraint not found for deletion (already deleted or never existed)", "data": map[string]interface{}{"branchId": branchID.String(), "dayOfWeek": dayOfWeek}, "timestamp": fmt.Sprintf("%d", 0), "sessionId": "debug-session", "runId": "run1", "hypothesisId": "C"})
					logFile.WriteString(string(logData) + "\n")
					logFile.Close()
				}
				// #endregion
			}
		}
		// #region agent log
		if logFile, err := os.OpenFile("c:\\Users\\User\\dev_projects\\vsq-oper_manpower\\.cursor\\debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
			logData, _ := json.Marshal(map[string]interface{}{"location": "branch_config_handler.go:487", "message": "UpdateConstraints: deletion phase completed", "data": map[string]interface{}{"branchId": branchID.String(), "deletedCount": len(daysToDelete), "remainingToUpdate": len(constraintsToUpdate)}, "timestamp": fmt.Sprintf("%d", 0), "sessionId": "debug-session", "runId": "run1", "hypothesisId": "C"})
			logFile.WriteString(string(logData) + "\n")
			logFile.Close()
		}
		// #endregion
	}

	// Bulk upsert remaining constraints with staff groups
	// Only update constraints that are NOT being deleted
	if len(constraintsToUpdate) > 0 {
		// #region agent log
		if logFile, err := os.OpenFile("c:\\Users\\User\\dev_projects\\vsq-oper_manpower\\.cursor\\debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
			logData, _ := json.Marshal(map[string]interface{}{"location": "branch_config_handler.go:495", "message": "UpdateConstraints: upserting constraints", "data": map[string]interface{}{"branchId": branchID.String(), "constraintsCount": len(constraintsToUpdate), "constraints": func() []map[string]interface{} {
				var result []map[string]interface{}
				for _, c := range constraintsToUpdate {
					result = append(result, map[string]interface{}{"day": c.DayOfWeek, "isOverridden": c.IsOverridden, "reqCount": len(c.StaffGroupRequirements)})
				}
				return result
			}()}, "timestamp": fmt.Sprintf("%d", 0), "sessionId": "debug-session", "runId": "run1", "hypothesisId": "D"})
			logFile.WriteString(string(logData) + "\n")
			logFile.Close()
		}
		// #endregion
		if err := h.repos.BranchConstraints.BulkUpsertWithStaffGroups(constraintsToUpdate); err != nil {
			// #region agent log
			if logFile, err2 := os.OpenFile("c:\\Users\\User\\dev_projects\\vsq-oper_manpower\\.cursor\\debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err2 == nil {
				logData, _ := json.Marshal(map[string]interface{}{"location": "branch_config_handler.go:498", "message": "UpdateConstraints: upsert failed", "data": map[string]interface{}{"branchId": branchID.String(), "error": err.Error()}, "timestamp": fmt.Sprintf("%d", 0), "sessionId": "debug-session", "runId": "run1", "hypothesisId": "D"})
				logFile.WriteString(string(logData) + "\n")
				logFile.Close()
			}
			// #endregion
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		// #region agent log
		if logFile, err := os.OpenFile("c:\\Users\\User\\dev_projects\\vsq-oper_manpower\\.cursor\\debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
			logData, _ := json.Marshal(map[string]interface{}{"location": "branch_config_handler.go:505", "message": "UpdateConstraints: upsert completed successfully", "data": map[string]interface{}{"branchId": branchID.String(), "constraintsCount": len(constraintsToUpdate)}, "timestamp": fmt.Sprintf("%d", 0), "sessionId": "debug-session", "runId": "run1", "hypothesisId": "D"})
			logFile.WriteString(string(logData) + "\n")
			logFile.Close()
		}
		// #endregion
	} else {
		// #region agent log
		if logFile, err := os.OpenFile("c:\\Users\\User\\dev_projects\\vsq-oper_manpower\\.cursor\\debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
			logData, _ := json.Marshal(map[string]interface{}{"location": "branch_config_handler.go:509", "message": "UpdateConstraints: no constraints to upsert (all deleted)", "data": map[string]interface{}{"branchId": branchID.String()}, "timestamp": fmt.Sprintf("%d", 0), "sessionId": "debug-session", "runId": "run1", "hypothesisId": "D"})
			logFile.WriteString(string(logData) + "\n")
			logFile.Close()
		}
		// #endregion
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

	// #region agent log
	if logFile, err2 := os.OpenFile("c:\\Users\\User\\dev_projects\\vsq-oper_manpower\\.cursor\\debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err2 == nil {
		logData, _ := json.Marshal(map[string]interface{}{"location": "branch_config_handler.go:478", "message": "GetConstraints: returning resolved constraints", "data": map[string]interface{}{"branchId": branchID.String(), "constraintsCount": len(constraints), "constraints": func() []map[string]interface{} {
			var result []map[string]interface{}
			for _, c := range constraints {
				result = append(result, map[string]interface{}{"day": c.DayOfWeek, "is_overridden": c.IsOverridden, "reqCount": len(c.StaffGroupRequirements)})
			}
			return result
		}()}, "timestamp": fmt.Sprintf("%d", 0), "sessionId": "debug-session", "runId": "run1", "hypothesisId": "F"})
		logFile.WriteString(string(logData) + "\n")
		logFile.Close()
	}
	// #endregion

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

	// #region agent log
	if logFile, err2 := os.OpenFile("c:\\Users\\User\\dev_projects\\vsq-oper_manpower\\.cursor\\debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err2 == nil {
		logData, _ := json.Marshal(map[string]interface{}{"location": "branch_config_handler.go:500", "message": "getResolvedConstraints: loaded branch constraints from DB", "data": map[string]interface{}{"branchId": branchID.String(), "constraintsCount": len(branchConstraints), "constraints": func() []map[string]interface{} {
			var result []map[string]interface{}
			for _, c := range branchConstraints {
				result = append(result, map[string]interface{}{"day": c.DayOfWeek, "is_overridden": c.IsOverridden, "reqCount": len(c.StaffGroupRequirements)})
			}
			return result
		}()}, "timestamp": fmt.Sprintf("%d", 0), "sessionId": "debug-session", "runId": "run1", "hypothesisId": "D"})
		logFile.WriteString(string(logData) + "\n")
		logFile.Close()
	}
	// #endregion

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

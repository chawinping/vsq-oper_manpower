package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"vsq-oper-manpower/backend/internal/domain/models"
	"vsq-oper-manpower/backend/internal/repositories/postgres"
	"vsq-oper-manpower/backend/internal/usecases/allocation"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AllocationCriteriaHandler struct {
	repos *postgres.Repositories
}

func NewAllocationCriteriaHandler(repos *postgres.Repositories) *AllocationCriteriaHandler {
	return &AllocationCriteriaHandler{repos: repos}
}

// GetCriteriaPriorityOrder returns the current criteria priority order configuration
func (h *AllocationCriteriaHandler) GetCriteriaPriorityOrder(c *gin.Context) {
	// Try to get from settings, otherwise return defaults
	setting, err := h.repos.Settings.GetByKey("allocation_criteria_priority_order")
	if err != nil || setting == nil {
		// Return default priority order
		defaultPriorityOrder := allocation.DefaultCriteriaPriorityOrder()
		c.JSON(http.StatusOK, gin.H{
			"priority_order":            defaultPriorityOrder.PriorityOrder,
			"enable_doctor_preferences": false,
		})
		return
	}

	var priorityOrder allocation.CriteriaPriorityOrder
	if err := json.Unmarshal([]byte(setting.Value), &priorityOrder); err != nil {
		// If parsing fails, return defaults
		defaultPriorityOrder := allocation.DefaultCriteriaPriorityOrder()
		c.JSON(http.StatusOK, gin.H{
			"priority_order":            defaultPriorityOrder.PriorityOrder,
			"enable_doctor_preferences": false,
		})
		return
	}

	// Validate priority order
	if len(priorityOrder.PriorityOrder) == 0 {
		defaultPriorityOrder := allocation.DefaultCriteriaPriorityOrder()
		c.JSON(http.StatusOK, gin.H{
			"priority_order":            defaultPriorityOrder.PriorityOrder,
			"enable_doctor_preferences": false,
		})
		return
	}

	// Get doctor preferences setting
	doctorPrefSetting, _ := h.repos.Settings.GetByKey("allocation_enable_doctor_preferences")
	enableDoctorPrefs := false
	if doctorPrefSetting != nil && doctorPrefSetting.Value == "true" {
		enableDoctorPrefs = true
	}

	c.JSON(http.StatusOK, gin.H{
		"priority_order":            priorityOrder.PriorityOrder,
		"enable_doctor_preferences": enableDoctorPrefs,
	})
}

// UpdateCriteriaPriorityOrder updates the criteria priority order configuration
type UpdateCriteriaPriorityOrderRequest struct {
	PriorityOrder           []string `json:"priority_order" binding:"required"`
	EnableDoctorPreferences bool     `json:"enable_doctor_preferences"`
}

func (h *AllocationCriteriaHandler) UpdateCriteriaPriorityOrder(c *gin.Context) {
	var req UpdateCriteriaPriorityOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate priority order
	if len(req.PriorityOrder) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Priority order must not be empty"})
		return
	}

	// Validate that all required criteria are present
	validCriteria := map[string]bool{
		allocation.CriterionZeroth: false,
		allocation.CriterionFirst:  false,
		allocation.CriterionSecond: false,
		allocation.CriterionThird:  false,
		allocation.CriterionFourth: false,
	}

	for _, criterionID := range req.PriorityOrder {
		if _, exists := validCriteria[criterionID]; exists {
			validCriteria[criterionID] = true
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid criterion ID: %s", criterionID)})
			return
		}
	}

	// Check if all required criteria are present (except zeroth which is optional)
	requiredCriteria := []string{
		allocation.CriterionFirst,
		allocation.CriterionSecond,
		allocation.CriterionThird,
		allocation.CriterionFourth,
	}

	for _, criterionID := range requiredCriteria {
		if !validCriteria[criterionID] {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Missing required criterion: %s", criterionID)})
			return
		}
	}

	// Create priority order struct
	priorityOrder := allocation.CriteriaPriorityOrder{
		PriorityOrder: req.PriorityOrder,
	}

	// Serialize priority order to JSON
	priorityOrderJSON, err := json.Marshal(priorityOrder)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to serialize priority order"})
		return
	}

	// Save priority order to settings (upsert logic)
	existingSetting, _ := h.repos.Settings.GetByKey("allocation_criteria_priority_order")
	if existingSetting != nil {
		existingSetting.Value = string(priorityOrderJSON)
		existingSetting.Description = "Allocation criteria priority order for strict lexicographic ranking"
		err = h.repos.Settings.Update(existingSetting)
	} else {
		setting := &models.SystemSetting{
			ID:          uuid.New(),
			Key:         "allocation_criteria_priority_order",
			Value:       string(priorityOrderJSON),
			Description: "Allocation criteria priority order for strict lexicographic ranking",
		}
		err = h.repos.Settings.Create(setting)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Save doctor preferences setting (upsert logic)
	doctorPrefValue := "false"
	if req.EnableDoctorPreferences {
		doctorPrefValue = "true"
	}
	doctorPrefSetting, _ := h.repos.Settings.GetByKey("allocation_enable_doctor_preferences")
	if doctorPrefSetting != nil {
		doctorPrefSetting.Value = doctorPrefValue
		doctorPrefSetting.Description = "Enable doctor preferences in allocation criteria"
		err = h.repos.Settings.Update(doctorPrefSetting)
	} else {
		setting := &models.SystemSetting{
			ID:          uuid.New(),
			Key:         "allocation_enable_doctor_preferences",
			Value:       doctorPrefValue,
			Description: "Enable doctor preferences in allocation criteria",
		}
		err = h.repos.Settings.Create(setting)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":                   "Criteria priority order updated successfully",
		"priority_order":            req.PriorityOrder,
		"enable_doctor_preferences": req.EnableDoctorPreferences,
	})
}

// ResetCriteriaPriorityOrder resets to default priority order
func (h *AllocationCriteriaHandler) ResetCriteriaPriorityOrder(c *gin.Context) {
	defaultPriorityOrder := allocation.DefaultCriteriaPriorityOrder()

	priorityOrderJSON, err := json.Marshal(defaultPriorityOrder)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to serialize default priority order"})
		return
	}

	// Reset priority order (upsert logic)
	existingSetting, _ := h.repos.Settings.GetByKey("allocation_criteria_priority_order")
	if existingSetting != nil {
		existingSetting.Value = string(priorityOrderJSON)
		existingSetting.Description = "Allocation criteria priority order for strict lexicographic ranking"
		err = h.repos.Settings.Update(existingSetting)
	} else {
		setting := &models.SystemSetting{
			ID:          uuid.New(),
			Key:         "allocation_criteria_priority_order",
			Value:       string(priorityOrderJSON),
			Description: "Allocation criteria priority order for strict lexicographic ranking",
		}
		err = h.repos.Settings.Create(setting)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Reset doctor preferences setting
	doctorPrefSetting, _ := h.repos.Settings.GetByKey("allocation_enable_doctor_preferences")
	if doctorPrefSetting != nil {
		doctorPrefSetting.Value = "false"
		doctorPrefSetting.Description = "Enable doctor preferences in allocation criteria"
		err = h.repos.Settings.Update(doctorPrefSetting)
	} else {
		setting := &models.SystemSetting{
			ID:          uuid.New(),
			Key:         "allocation_enable_doctor_preferences",
			Value:       "false",
			Description: "Enable doctor preferences in allocation criteria",
		}
		err = h.repos.Settings.Create(setting)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":                   "Criteria priority order reset to defaults",
		"priority_order":            defaultPriorityOrder.PriorityOrder,
		"enable_doctor_preferences": false,
	})
}

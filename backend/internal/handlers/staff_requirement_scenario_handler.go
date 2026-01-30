package handlers

import (
	"net/http"
	"time"

	"vsq-oper-manpower/backend/internal/domain/models"
	"vsq-oper-manpower/backend/internal/repositories/postgres"
	"vsq-oper-manpower/backend/internal/usecases/scenario"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type StaffRequirementScenarioHandler struct {
	repos *postgres.Repositories
}

func NewStaffRequirementScenarioHandler(repos *postgres.Repositories) *StaffRequirementScenarioHandler {
	return &StaffRequirementScenarioHandler{repos: repos}
}

// List returns all staff requirement scenarios
func (h *StaffRequirementScenarioHandler) List(c *gin.Context) {
	includeInactive := c.Query("include_inactive") == "true"
	scenarios, err := h.repos.StaffRequirementScenario.List(includeInactive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Load position requirements and specific staff requirements for each scenario
	for _, scenario := range scenarios {
		requirements, err := h.repos.ScenarioPositionRequirement.GetByScenarioID(scenario.ID)
		if err == nil {
			// Convert []*models.ScenarioPositionRequirement to []models.ScenarioPositionRequirement
			scenario.PositionRequirements = make([]models.ScenarioPositionRequirement, len(requirements))
			for i, req := range requirements {
				scenario.PositionRequirements[i] = *req
			}
		}
		specificStaffReqs, err := h.repos.ScenarioSpecificStaffRequirement.GetByScenarioID(scenario.ID)
		if err == nil {
			scenario.SpecificStaffRequirements = make([]models.ScenarioSpecificStaffRequirement, len(specificStaffReqs))
			for i, req := range specificStaffReqs {
				scenario.SpecificStaffRequirements[i] = *req
			}
		}
	}

	c.JSON(http.StatusOK, scenarios)
}

// GetByID returns a scenario by ID with position requirements
func (h *StaffRequirementScenarioHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid scenario ID"})
		return
	}

	scenario, err := h.repos.StaffRequirementScenario.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if scenario == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Scenario not found"})
		return
	}

	// Load position requirements
	requirements, err := h.repos.ScenarioPositionRequirement.GetByScenarioID(scenario.ID)
	if err == nil {
		// Convert []*models.ScenarioPositionRequirement to []models.ScenarioPositionRequirement
		scenario.PositionRequirements = make([]models.ScenarioPositionRequirement, len(requirements))
		for i, req := range requirements {
			scenario.PositionRequirements[i] = *req
		}
	}

	// Load specific staff requirements
	specificStaffReqs, err := h.repos.ScenarioSpecificStaffRequirement.GetByScenarioID(scenario.ID)
	if err == nil {
		scenario.SpecificStaffRequirements = make([]models.ScenarioSpecificStaffRequirement, len(specificStaffReqs))
		for i, req := range specificStaffReqs {
			scenario.SpecificStaffRequirements[i] = *req
		}
	}

	// Load revenue tier if present
	if scenario.RevenueLevelTierID != nil {
		tier, err := h.repos.RevenueLevelTier.GetByID(*scenario.RevenueLevelTierID)
		if err == nil && tier != nil {
			scenario.RevenueLevelTier = tier
		}
	}

	// Load doctor if present
	if scenario.DoctorID != nil {
		doctor, err := h.repos.Doctor.GetByID(*scenario.DoctorID)
		if err == nil && doctor != nil {
			scenario.Doctor = doctor
		}
	}

	// Load branch if present
	if scenario.BranchID != nil {
		branch, err := h.repos.Branch.GetByID(*scenario.BranchID)
		if err == nil && branch != nil {
			scenario.Branch = branch
		}
	}

	c.JSON(http.StatusOK, scenario)
}

// Create creates a new scenario with position requirements
func (h *StaffRequirementScenarioHandler) Create(c *gin.Context) {
	var req models.StaffRequirementScenarioCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	scenario := &models.StaffRequirementScenario{
		ID:                     uuid.New(),
		ScenarioName:           req.ScenarioName,
		Description:            req.Description,
		DoctorID:               req.DoctorID,
		BranchID:               req.BranchID,
		RevenueLevelTierID:     req.RevenueLevelTierID,
		MinRevenue:             req.MinRevenue,
		MaxRevenue:             req.MaxRevenue,
		UseDayOfWeekRevenue:    req.UseDayOfWeekRevenue,
		UseSpecificDateRevenue: req.UseSpecificDateRevenue,
		DoctorCount:            req.DoctorCount,
		MinDoctorCount:         req.MinDoctorCount,
		DayOfWeek:              req.DayOfWeek,
		IsDefault:              false, // Default scenarios are not allowed
		IsActive:               req.IsActive,
		Priority:               req.Priority,
	}

	if err := h.repos.StaffRequirementScenario.Create(scenario); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Create position requirements
	if len(req.PositionRequirements) > 0 {
		requirements := make([]*models.ScenarioPositionRequirement, len(req.PositionRequirements))
		for i, reqReq := range req.PositionRequirements {
			requirements[i] = &models.ScenarioPositionRequirement{
				ID:             uuid.New(),
				ScenarioID:     scenario.ID,
				PositionID:     reqReq.PositionID,
				PreferredStaff: reqReq.PreferredStaff,
				MinimumStaff:   reqReq.MinimumStaff,
				OverrideBase:   reqReq.OverrideBase,
			}
		}
		if err := h.repos.ScenarioPositionRequirement.BulkUpsert(requirements); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create position requirements: " + err.Error()})
			return
		}
	}

	// Create specific staff requirements
	if len(req.SpecificStaffRequirements) > 0 {
		specificStaffReqs := make([]*models.ScenarioSpecificStaffRequirement, len(req.SpecificStaffRequirements))
		for i, reqReq := range req.SpecificStaffRequirements {
			specificStaffReqs[i] = &models.ScenarioSpecificStaffRequirement{
				ID:         uuid.New(),
				ScenarioID: scenario.ID,
				StaffID:    reqReq.StaffID,
			}
		}
		if err := h.repos.ScenarioSpecificStaffRequirement.BulkUpsert(specificStaffReqs); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create specific staff requirements: " + err.Error()})
			return
		}
	}

	// Reload scenario with requirements
	requirements, _ := h.repos.ScenarioPositionRequirement.GetByScenarioID(scenario.ID)
	if requirements != nil {
		// Convert []*models.ScenarioPositionRequirement to []models.ScenarioPositionRequirement
		scenario.PositionRequirements = make([]models.ScenarioPositionRequirement, len(requirements))
		for i, req := range requirements {
			scenario.PositionRequirements[i] = *req
		}
	}
	specificStaffReqs, _ := h.repos.ScenarioSpecificStaffRequirement.GetByScenarioID(scenario.ID)
	if specificStaffReqs != nil {
		scenario.SpecificStaffRequirements = make([]models.ScenarioSpecificStaffRequirement, len(specificStaffReqs))
		for i, req := range specificStaffReqs {
			scenario.SpecificStaffRequirements[i] = *req
		}
	}

	c.JSON(http.StatusCreated, scenario)
}

// Update updates a scenario
func (h *StaffRequirementScenarioHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid scenario ID"})
		return
	}

	scenario, err := h.repos.StaffRequirementScenario.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if scenario == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Scenario not found"})
		return
	}

	var req models.StaffRequirementScenarioUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.ScenarioName != nil {
		scenario.ScenarioName = *req.ScenarioName
	}
	if req.Description != nil {
		scenario.Description = req.Description
	}
	if req.DoctorID != nil {
		scenario.DoctorID = req.DoctorID
	}
	if req.BranchID != nil {
		scenario.BranchID = req.BranchID
	}
	if req.RevenueLevelTierID != nil {
		scenario.RevenueLevelTierID = req.RevenueLevelTierID
	}
	if req.MinRevenue != nil {
		scenario.MinRevenue = req.MinRevenue
	}
	if req.MaxRevenue != nil {
		scenario.MaxRevenue = req.MaxRevenue
	}
	if req.UseDayOfWeekRevenue != nil {
		scenario.UseDayOfWeekRevenue = *req.UseDayOfWeekRevenue
	}
	if req.UseSpecificDateRevenue != nil {
		scenario.UseSpecificDateRevenue = *req.UseSpecificDateRevenue
	}
	if req.DoctorCount != nil {
		scenario.DoctorCount = req.DoctorCount
	}
	if req.MinDoctorCount != nil {
		scenario.MinDoctorCount = req.MinDoctorCount
	}
	if req.DayOfWeek != nil {
		scenario.DayOfWeek = req.DayOfWeek
	}
	// Always set is_default to false - default scenarios are not allowed
	scenario.IsDefault = false
	if req.IsActive != nil {
		scenario.IsActive = *req.IsActive
	}
	if req.Priority != nil {
		scenario.Priority = *req.Priority
	}

	if err := h.repos.StaffRequirementScenario.Update(scenario); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, scenario)
}

// Delete deletes a scenario
func (h *StaffRequirementScenarioHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid scenario ID"})
		return
	}

	if err := h.repos.StaffRequirementScenario.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Scenario deleted successfully"})
}

// UpdatePositionRequirements updates position requirements for a scenario
func (h *StaffRequirementScenarioHandler) UpdatePositionRequirements(c *gin.Context) {
	idStr := c.Param("id")
	scenarioID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid scenario ID"})
		return
	}

	var req struct {
		Requirements []models.ScenarioPositionRequirementCreate `json:"requirements" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Delete existing requirements
	if err := h.repos.ScenarioPositionRequirement.DeleteByScenarioID(scenarioID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete existing requirements: " + err.Error()})
		return
	}

	// Create new requirements
	if len(req.Requirements) > 0 {
		requirements := make([]*models.ScenarioPositionRequirement, len(req.Requirements))
		for i, reqReq := range req.Requirements {
			requirements[i] = &models.ScenarioPositionRequirement{
				ID:             uuid.New(),
				ScenarioID:     scenarioID,
				PositionID:     reqReq.PositionID,
				PreferredStaff: reqReq.PreferredStaff,
				MinimumStaff:   reqReq.MinimumStaff,
				OverrideBase:   reqReq.OverrideBase,
			}
		}
		if err := h.repos.ScenarioPositionRequirement.BulkUpsert(requirements); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create requirements: " + err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Position requirements updated successfully"})
}

// UpdateSpecificStaffRequirements updates specific staff requirements for a scenario
func (h *StaffRequirementScenarioHandler) UpdateSpecificStaffRequirements(c *gin.Context) {
	idStr := c.Param("id")
	scenarioID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid scenario ID"})
		return
	}

	var req struct {
		Requirements []models.ScenarioSpecificStaffRequirementCreate `json:"requirements" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Delete existing requirements
	if err := h.repos.ScenarioSpecificStaffRequirement.DeleteByScenarioID(scenarioID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete existing requirements: " + err.Error()})
		return
	}

	// Create new requirements
	if len(req.Requirements) > 0 {
		requirements := make([]*models.ScenarioSpecificStaffRequirement, len(req.Requirements))
		for i, reqReq := range req.Requirements {
			requirements[i] = &models.ScenarioSpecificStaffRequirement{
				ID:         uuid.New(),
				ScenarioID: scenarioID,
				StaffID:    reqReq.StaffID,
			}
		}
		if err := h.repos.ScenarioSpecificStaffRequirement.BulkUpsert(requirements); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create requirements: " + err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Specific staff requirements updated successfully"})
}

// CalculateRequirements calculates staff requirements for a branch/date/position
func (h *StaffRequirementScenarioHandler) CalculateRequirements(c *gin.Context) {
	var req struct {
		BranchID      string `json:"branch_id" binding:"required"`
		Date          string `json:"date" binding:"required"`
		PositionID    string `json:"position_id" binding:"required"`
		BasePreferred int    `json:"base_preferred"`
		BaseMinimum   int    `json:"base_minimum"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	branchID, err := uuid.Parse(req.BranchID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid branch ID"})
		return
	}

	positionID, err := uuid.Parse(req.PositionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid position ID"})
		return
	}

	// Parse date
	date, err := parseDate(req.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format"})
		return
	}

	// Get base quotas if not provided
	if req.BasePreferred == 0 && req.BaseMinimum == 0 {
		quota, err := h.repos.PositionQuota.GetByBranchAndPosition(branchID, positionID)
		if err == nil && quota != nil {
			req.BasePreferred = quota.DesignatedQuota
			req.BaseMinimum = quota.MinimumRequired
		}
	}

	// Use scenario calculator
	reposWrapper := &scenario.RepositoriesWrapper{
		BranchWeeklyRevenue:         h.repos.BranchWeeklyRevenue,
		Revenue:                     h.repos.Revenue,
		DoctorAssignment:            h.repos.DoctorAssignment,
		PositionQuota:               h.repos.PositionQuota,
		RevenueLevelTier:            h.repos.RevenueLevelTier,
		StaffRequirementScenario:    h.repos.StaffRequirementScenario,
		ScenarioPositionRequirement: h.repos.ScenarioPositionRequirement,
		Position:                    h.repos.Position,
	}

	calculator := scenario.NewScenarioCalculator(reposWrapper)
	result, err := calculator.CalculateStaffRequirements(branchID, date, positionID, req.BasePreferred, req.BaseMinimum)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetMatchingScenarios returns all scenarios that match given conditions
func (h *StaffRequirementScenarioHandler) GetMatchingScenarios(c *gin.Context) {
	var req struct {
		BranchID string `json:"branch_id" binding:"required"`
		Date     string `json:"date" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	branchID, err := uuid.Parse(req.BranchID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid branch ID"})
		return
	}

	date, err := parseDate(req.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format"})
		return
	}

	reposWrapper := &scenario.RepositoriesWrapper{
		BranchWeeklyRevenue:         h.repos.BranchWeeklyRevenue,
		Revenue:                     h.repos.Revenue,
		DoctorAssignment:            h.repos.DoctorAssignment,
		PositionQuota:               h.repos.PositionQuota,
		RevenueLevelTier:            h.repos.RevenueLevelTier,
		StaffRequirementScenario:    h.repos.StaffRequirementScenario,
		ScenarioPositionRequirement: h.repos.ScenarioPositionRequirement,
		Position:                    h.repos.Position,
	}

	calculator := scenario.NewScenarioCalculator(reposWrapper)
	matches, err := calculator.GetMatchingScenarios(branchID, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, matches)
}

func parseDate(dateStr string) (time.Time, error) {
	return time.Parse("2006-01-02", dateStr)
}

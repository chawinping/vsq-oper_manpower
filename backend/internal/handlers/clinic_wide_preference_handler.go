package handlers

import (
	"net/http"
	"strconv"

	"vsq-oper-manpower/backend/internal/domain/models"
	"vsq-oper-manpower/backend/internal/repositories/postgres"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ClinicWidePreferenceHandler struct {
	repos *postgres.Repositories
}

func NewClinicWidePreferenceHandler(repos *postgres.Repositories) *ClinicWidePreferenceHandler {
	return &ClinicWidePreferenceHandler{repos: repos}
}

// List returns all clinic-wide preferences, optionally filtered by criteria type
func (h *ClinicWidePreferenceHandler) List(c *gin.Context) {
	filters := models.ClinicPreferenceFilters{}

	// Filter by criteria type if provided
	if criteriaTypeStr := c.Query("criteria_type"); criteriaTypeStr != "" {
		criteriaType := models.ClinicPreferenceCriteriaType(criteriaTypeStr)
		filters.CriteriaType = &criteriaType
	}

	// Filter by active status if provided
	if isActiveStr := c.Query("is_active"); isActiveStr != "" {
		isActive := isActiveStr == "true"
		filters.IsActive = &isActive
	}

	preferences, err := h.repos.ClinicWidePreference.List(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Load position requirements for each preference
	for _, preference := range preferences {
		requirements, err := h.repos.PreferencePositionRequirement.GetByPreferenceID(preference.ID)
		if err == nil {
			preference.PositionRequirements = make([]models.PreferencePositionRequirement, len(requirements))
			for i, req := range requirements {
				preference.PositionRequirements[i] = *req
				// Load position details
				if req.PositionID != uuid.Nil {
					position, err := h.repos.Position.GetByID(req.PositionID)
					if err == nil && position != nil {
						preference.PositionRequirements[i].Position = position
					}
				}
			}
		}
	}

	c.JSON(http.StatusOK, preferences)
}

// GetByID returns a preference by ID with position requirements
func (h *ClinicWidePreferenceHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid preference ID"})
		return
	}

	preference, err := h.repos.ClinicWidePreference.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if preference == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Preference not found"})
		return
	}

	// Load position requirements
	requirements, err := h.repos.PreferencePositionRequirement.GetByPreferenceID(preference.ID)
	if err == nil {
		preference.PositionRequirements = make([]models.PreferencePositionRequirement, len(requirements))
		for i, req := range requirements {
			preference.PositionRequirements[i] = *req
			// Load position details
			if req.PositionID != uuid.Nil {
				position, err := h.repos.Position.GetByID(req.PositionID)
				if err == nil && position != nil {
					preference.PositionRequirements[i].Position = position
				}
			}
		}
	}

	c.JSON(http.StatusOK, preference)
}

// Create creates a new clinic-wide preference with position requirements
func (h *ClinicWidePreferenceHandler) Create(c *gin.Context) {
	var req models.ClinicWidePreferenceCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Note: MinValue validation is handled by binding:"min=0"
	// We allow 0 as a valid value, so we don't need to check for zero value here

	// Validate max_value >= min_value if max_value is set
	// For doctor_count, allow equality (max_value == min_value)
	// For other criteria types, require max_value > min_value
	if req.MaxValue != nil {
		if req.CriteriaType == models.ClinicCriteriaTypeDoctorCount {
			// Allow equality for doctor_count
			if *req.MaxValue < req.MinValue {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Max value must be greater than or equal to min value"})
				return
			}
		} else {
			// Require strict inequality for other criteria types
			if *req.MaxValue <= req.MinValue {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Max value must be greater than min value"})
				return
			}
		}
	}

	// Validate preferred_staff >= minimum_staff for all requirements
	for _, reqReq := range req.PositionRequirements {
		if reqReq.PreferredStaff < reqReq.MinimumStaff {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Preferred staff must be >= minimum staff for all positions"})
			return
		}
	}

	preference := &models.ClinicWidePreference{
		ID:           uuid.New(),
		CriteriaType: req.CriteriaType,
		CriteriaName: req.CriteriaName,
		MinValue:     req.MinValue,
		MaxValue:     req.MaxValue,
		IsActive:     req.IsActive,
		DisplayOrder: req.DisplayOrder,
		Description:  req.Description,
	}

	if err := h.repos.ClinicWidePreference.Create(preference); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Create position requirements
	if len(req.PositionRequirements) > 0 {
		requirements := make([]*models.PreferencePositionRequirement, len(req.PositionRequirements))
		for i, reqReq := range req.PositionRequirements {
			requirements[i] = &models.PreferencePositionRequirement{
				ID:             uuid.New(),
				PreferenceID:   preference.ID,
				PositionID:     reqReq.PositionID,
				MinimumStaff:   reqReq.MinimumStaff,
				PreferredStaff: reqReq.PreferredStaff,
				IsActive:       reqReq.IsActive,
			}
		}
		if err := h.repos.PreferencePositionRequirement.BulkUpsert(requirements); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create position requirements: " + err.Error()})
			return
		}
	}

	// Reload preference with requirements
	requirements, _ := h.repos.PreferencePositionRequirement.GetByPreferenceID(preference.ID)
	if requirements != nil {
		preference.PositionRequirements = make([]models.PreferencePositionRequirement, len(requirements))
		for i, req := range requirements {
			preference.PositionRequirements[i] = *req
			// Load position details
			if req.PositionID != uuid.Nil {
				position, err := h.repos.Position.GetByID(req.PositionID)
				if err == nil && position != nil {
					preference.PositionRequirements[i].Position = position
				}
			}
		}
	}

	c.JSON(http.StatusCreated, preference)
}

// Update updates a clinic-wide preference
func (h *ClinicWidePreferenceHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid preference ID"})
		return
	}

	var req models.ClinicWidePreferenceUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	preference, err := h.repos.ClinicWidePreference.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if preference == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Preference not found"})
		return
	}

	// Validate max_value >= min_value if max_value is being updated
	// For doctor_count, allow equality (max_value == min_value)
	// For other criteria types, require max_value > min_value
	if req.MaxValue != nil {
		minValue := preference.MinValue
		if req.MinValue != nil {
			minValue = *req.MinValue
		}
		if preference.CriteriaType == models.ClinicCriteriaTypeDoctorCount {
			// Allow equality for doctor_count
			if *req.MaxValue < minValue {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Max value must be greater than or equal to min value"})
				return
			}
		} else {
			// Require strict inequality for other criteria types
			if *req.MaxValue <= minValue {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Max value must be greater than min value"})
				return
			}
		}
	} else if req.MinValue != nil {
		// If min_value is being updated but max_value exists, validate
		if preference.MaxValue != nil {
			if preference.CriteriaType == models.ClinicCriteriaTypeDoctorCount {
				// Allow equality for doctor_count
				if *preference.MaxValue < *req.MinValue {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Max value must be greater than or equal to min value"})
					return
				}
			} else {
				// Require strict inequality for other criteria types
				if *preference.MaxValue <= *req.MinValue {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Max value must be greater than min value"})
					return
				}
			}
		}
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if preference == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Preference not found"})
		return
	}

	// Update fields
	if req.CriteriaName != nil {
		preference.CriteriaName = *req.CriteriaName
	}
	if req.MinValue != nil {
		preference.MinValue = *req.MinValue
	}
	if req.MaxValue != nil {
		preference.MaxValue = req.MaxValue
	}
	if req.IsActive != nil {
		preference.IsActive = *req.IsActive
	}
	if req.DisplayOrder != nil {
		preference.DisplayOrder = *req.DisplayOrder
	}
	if req.Description != nil {
		preference.Description = req.Description
	}

	if err := h.repos.ClinicWidePreference.Update(preference); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Reload with requirements
	requirements, _ := h.repos.PreferencePositionRequirement.GetByPreferenceID(preference.ID)
	if requirements != nil {
		preference.PositionRequirements = make([]models.PreferencePositionRequirement, len(requirements))
		for i, req := range requirements {
			preference.PositionRequirements[i] = *req
			if req.PositionID != uuid.Nil {
				position, err := h.repos.Position.GetByID(req.PositionID)
				if err == nil && position != nil {
					preference.PositionRequirements[i].Position = position
				}
			}
		}
	}

	c.JSON(http.StatusOK, preference)
}

// Delete deletes a clinic-wide preference (cascade deletes position requirements)
func (h *ClinicWidePreferenceHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid preference ID"})
		return
	}

	if err := h.repos.ClinicWidePreference.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Preference deleted successfully"})
}

// AddPositionRequirement adds a position requirement to a preference
func (h *ClinicWidePreferenceHandler) AddPositionRequirement(c *gin.Context) {
	idStr := c.Param("id")
	preferenceID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid preference ID"})
		return
	}

	var req models.PreferencePositionRequirementCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.PreferredStaff < req.MinimumStaff {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Preferred staff must be >= minimum staff"})
		return
	}

	requirement := &models.PreferencePositionRequirement{
		ID:             uuid.New(),
		PreferenceID:   preferenceID,
		PositionID:     req.PositionID,
		MinimumStaff:   req.MinimumStaff,
		PreferredStaff: req.PreferredStaff,
		IsActive:       req.IsActive,
	}

	if err := h.repos.PreferencePositionRequirement.Create(requirement); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Load position details
	if requirement.PositionID != uuid.Nil {
		position, err := h.repos.Position.GetByID(requirement.PositionID)
		if err == nil && position != nil {
			requirement.Position = position
		}
	}

	c.JSON(http.StatusCreated, requirement)
}

// UpdatePositionRequirement updates a position requirement
func (h *ClinicWidePreferenceHandler) UpdatePositionRequirement(c *gin.Context) {
	idStr := c.Param("id")
	preferenceID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid preference ID"})
		return
	}

	positionIDStr := c.Param("positionId")
	positionID, err := uuid.Parse(positionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid position ID"})
		return
	}

	var req models.PreferencePositionRequirementUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	requirement, err := h.repos.PreferencePositionRequirement.GetByPreferenceAndPosition(preferenceID, positionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if requirement == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Position requirement not found"})
		return
	}

	// Update fields
	if req.MinimumStaff != nil {
		requirement.MinimumStaff = *req.MinimumStaff
	}
	if req.PreferredStaff != nil {
		requirement.PreferredStaff = *req.PreferredStaff
	}
	if req.IsActive != nil {
		requirement.IsActive = *req.IsActive
	}

	// Validate preferred_staff >= minimum_staff
	if requirement.PreferredStaff < requirement.MinimumStaff {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Preferred staff must be >= minimum staff"})
		return
	}

	if err := h.repos.PreferencePositionRequirement.Update(requirement); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Load position details
	if requirement.PositionID != uuid.Nil {
		position, err := h.repos.Position.GetByID(requirement.PositionID)
		if err == nil && position != nil {
			requirement.Position = position
		}
	}

	c.JSON(http.StatusOK, requirement)
}

// DeletePositionRequirement deletes a position requirement
func (h *ClinicWidePreferenceHandler) DeletePositionRequirement(c *gin.Context) {
	idStr := c.Param("id")
	preferenceID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid preference ID"})
		return
	}

	positionIDStr := c.Param("positionId")
	positionID, err := uuid.Parse(positionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid position ID"})
		return
	}

	requirement, err := h.repos.PreferencePositionRequirement.GetByPreferenceAndPosition(preferenceID, positionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if requirement == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Position requirement not found"})
		return
	}

	if err := h.repos.PreferencePositionRequirement.Delete(requirement.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Position requirement deleted successfully"})
}

// GetByCriteriaAndValue returns preferences matching a criteria type and value
func (h *ClinicWidePreferenceHandler) GetByCriteriaAndValue(c *gin.Context) {
	criteriaTypeStr := c.Param("criteriaType")
	criteriaType := models.ClinicPreferenceCriteriaType(criteriaTypeStr)

	valueStr := c.Query("value")
	if valueStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "value query parameter is required"})
		return
	}

	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid value parameter"})
		return
	}

	preferences, err := h.repos.ClinicWidePreference.GetByCriteriaTypeAndValue(criteriaType, value)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Load position requirements for each preference
	for _, preference := range preferences {
		requirements, err := h.repos.PreferencePositionRequirement.GetByPreferenceID(preference.ID)
		if err == nil {
			preference.PositionRequirements = make([]models.PreferencePositionRequirement, len(requirements))
			for i, req := range requirements {
				preference.PositionRequirements[i] = *req
				if req.PositionID != uuid.Nil {
					position, err := h.repos.Position.GetByID(req.PositionID)
					if err == nil && position != nil {
						preference.PositionRequirements[i].Position = position
					}
				}
			}
		}
	}

	c.JSON(http.StatusOK, preferences)
}

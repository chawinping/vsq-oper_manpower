package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"vsq-oper-manpower/backend/internal/domain/interfaces"
	"vsq-oper-manpower/backend/internal/domain/models"
	"vsq-oper-manpower/backend/internal/repositories/postgres"
)

type SpecificPreferenceHandler struct {
	repos *postgres.Repositories
}

func NewSpecificPreferenceHandler(repos *postgres.Repositories) *SpecificPreferenceHandler {
	return &SpecificPreferenceHandler{
		repos: repos,
	}
}

// Request/Response types
type CreateSpecificPreferenceRequest struct {
	BranchID        *string `json:"branch_id,omitempty"`        // UUID string or null for "any branch"
	DoctorID        *string `json:"doctor_id,omitempty"`        // UUID string or null for "any doctor"
	DayOfWeek       *int    `json:"day_of_week,omitempty"`     // 0-6 or null for "any day"
	PreferenceType  string  `json:"preference_type" binding:"required"` // "position_count" or "staff_name"
	PositionID      *string `json:"position_id,omitempty"`     // Required for position_count
	StaffCount      *int    `json:"staff_count,omitempty"`     // Required for position_count
	StaffID         *string `json:"staff_id,omitempty"`        // Required for staff_name
	IsActive        bool    `json:"is_active"`
}

type UpdateSpecificPreferenceRequest struct {
	BranchID        *string `json:"branch_id,omitempty"`
	DoctorID        *string `json:"doctor_id,omitempty"`
	DayOfWeek       *int    `json:"day_of_week,omitempty"`
	PreferenceType  string  `json:"preference_type"`
	PositionID      *string `json:"position_id,omitempty"`
	StaffCount      *int    `json:"staff_count,omitempty"`
	StaffID         *string `json:"staff_id,omitempty"`
	IsActive        *bool   `json:"is_active,omitempty"`
}

func (h *SpecificPreferenceHandler) List(c *gin.Context) {
	filters := interfaces.SpecificPreferenceFilters{}

	// Parse query parameters
	if branchIDStr := c.Query("branch_id"); branchIDStr != "" {
		branchID, err := uuid.Parse(branchIDStr)
		if err == nil {
			filters.BranchID = &branchID
		}
	}
	if doctorIDStr := c.Query("doctor_id"); doctorIDStr != "" {
		doctorID, err := uuid.Parse(doctorIDStr)
		if err == nil {
			filters.DoctorID = &doctorID
		}
	}
	if dayOfWeekStr := c.Query("day_of_week"); dayOfWeekStr != "" {
		dayOfWeek, err := strconv.Atoi(dayOfWeekStr)
		if err == nil && dayOfWeek >= 0 && dayOfWeek <= 6 {
			filters.DayOfWeek = &dayOfWeek
		}
	}
	if isActiveStr := c.Query("is_active"); isActiveStr != "" {
		isActive := isActiveStr == "true"
		filters.IsActive = &isActive
	}

	preferences, err := h.repos.SpecificPreference.List(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Load related entities
	for _, pref := range preferences {
		if pref.BranchID != nil {
			branch, _ := h.repos.Branch.GetByID(*pref.BranchID)
			pref.Branch = branch
		}
		if pref.DoctorID != nil {
			doctor, _ := h.repos.Doctor.GetByID(*pref.DoctorID)
			pref.Doctor = doctor
		}
		if pref.PositionID != nil {
			position, _ := h.repos.Position.GetByID(*pref.PositionID)
			pref.Position = position
		}
		if pref.StaffID != nil {
			staff, _ := h.repos.Staff.GetByID(*pref.StaffID)
			pref.Staff = staff
		}
	}

	c.JSON(http.StatusOK, gin.H{"preferences": preferences})
}

func (h *SpecificPreferenceHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	preference, err := h.repos.SpecificPreference.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if preference == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Preference not found"})
		return
	}

	// Load related entities
	if preference.BranchID != nil {
		branch, _ := h.repos.Branch.GetByID(*preference.BranchID)
		preference.Branch = branch
	}
	if preference.DoctorID != nil {
		doctor, _ := h.repos.Doctor.GetByID(*preference.DoctorID)
		preference.Doctor = doctor
	}
	if preference.PositionID != nil {
		position, _ := h.repos.Position.GetByID(*preference.PositionID)
		preference.Position = position
	}
	if preference.StaffID != nil {
		staff, _ := h.repos.Staff.GetByID(*preference.StaffID)
		preference.Staff = staff
	}

	c.JSON(http.StatusOK, gin.H{"preference": preference})
}

func (h *SpecificPreferenceHandler) Create(c *gin.Context) {
	var req CreateSpecificPreferenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	preference := &models.SpecificPreference{
		PreferenceType: models.SpecificPreferenceType(req.PreferenceType),
		IsActive:        req.IsActive,
	}

	// Parse branch_id
	if req.BranchID != nil && *req.BranchID != "" {
		branchID, err := uuid.Parse(*req.BranchID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid branch_id"})
			return
		}
		preference.BranchID = &branchID
	}

	// Parse doctor_id
	if req.DoctorID != nil && *req.DoctorID != "" {
		doctorID, err := uuid.Parse(*req.DoctorID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid doctor_id"})
			return
		}
		preference.DoctorID = &doctorID
	}

	// Parse day_of_week
	if req.DayOfWeek != nil {
		if *req.DayOfWeek < 0 || *req.DayOfWeek > 6 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "day_of_week must be between 0 and 6"})
			return
		}
		preference.DayOfWeek = req.DayOfWeek
	}

	// Parse position_id and staff_count for position_count type
	if req.PreferenceType == string(models.SpecificPreferenceTypePositionCount) {
		if req.PositionID == nil || *req.PositionID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "position_id is required for position_count type"})
			return
		}
		positionID, err := uuid.Parse(*req.PositionID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid position_id"})
			return
		}
		preference.PositionID = &positionID

		if req.StaffCount == nil || *req.StaffCount < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "staff_count must be at least 1 for position_count type"})
			return
		}
		preference.StaffCount = req.StaffCount
	}

	// Parse staff_id for staff_name type
	if req.PreferenceType == string(models.SpecificPreferenceTypeStaffName) {
		if req.StaffID == nil || *req.StaffID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "staff_id is required for staff_name type"})
			return
		}
		staffID, err := uuid.Parse(*req.StaffID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid staff_id"})
			return
		}
		preference.StaffID = &staffID
	}

	if err := h.repos.SpecificPreference.Create(preference); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"preference": preference})
}

func (h *SpecificPreferenceHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	// Get existing preference
	preference, err := h.repos.SpecificPreference.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if preference == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Preference not found"})
		return
	}

	var req UpdateSpecificPreferenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update fields
	if req.BranchID != nil {
		if *req.BranchID == "" {
			preference.BranchID = nil
		} else {
			branchID, err := uuid.Parse(*req.BranchID)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid branch_id"})
				return
			}
			preference.BranchID = &branchID
		}
	}

	if req.DoctorID != nil {
		if *req.DoctorID == "" {
			preference.DoctorID = nil
		} else {
			doctorID, err := uuid.Parse(*req.DoctorID)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid doctor_id"})
				return
			}
			preference.DoctorID = &doctorID
		}
	}

	if req.DayOfWeek != nil {
		if *req.DayOfWeek < 0 || *req.DayOfWeek > 6 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "day_of_week must be between 0 and 6"})
			return
		}
		preference.DayOfWeek = req.DayOfWeek
	}

	if req.PreferenceType != "" {
		preference.PreferenceType = models.SpecificPreferenceType(req.PreferenceType)
	}

	if req.PositionID != nil {
		if *req.PositionID == "" {
			preference.PositionID = nil
		} else {
			positionID, err := uuid.Parse(*req.PositionID)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid position_id"})
				return
			}
			preference.PositionID = &positionID
		}
	}

	if req.StaffCount != nil {
		preference.StaffCount = req.StaffCount
	}

	if req.StaffID != nil {
		if *req.StaffID == "" {
			preference.StaffID = nil
		} else {
			staffID, err := uuid.Parse(*req.StaffID)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid staff_id"})
				return
			}
			preference.StaffID = &staffID
		}
	}

	if req.IsActive != nil {
		preference.IsActive = *req.IsActive
	}

	if err := h.repos.SpecificPreference.Update(preference); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"preference": preference})
}

func (h *SpecificPreferenceHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := h.repos.SpecificPreference.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Preference deleted successfully"})
}

package handlers

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"vsq-oper-manpower/backend/internal/domain/models"
	"vsq-oper-manpower/backend/internal/repositories/postgres"
	"vsq-oper-manpower/backend/pkg/excel"
)

type DoctorHandler struct {
	repos         *postgres.Repositories
	excelImporter *excel.ExcelImporter
}

func NewDoctorHandler(repos *postgres.Repositories) *DoctorHandler {
	return &DoctorHandler{
		repos: repos,
		excelImporter: excel.NewExcelImporter(
			repos.Position,
			repos.Branch,
			repos.Doctor,
			repos.PositionQuota,
		),
	}
}

// Doctor CRUD operations
type CreateDoctorRequest struct {
	Name        string `json:"name" binding:"required"`
	Code        string `json:"code"`
	Preferences string `json:"preferences"` // Noted remark/preferences
}

type UpdateDoctorRequest struct {
	Name        string `json:"name"`
	Code        string `json:"code"`
	Preferences string `json:"preferences"` // Noted remark/preferences
}

func (h *DoctorHandler) List(c *gin.Context) {
	doctors, err := h.repos.Doctor.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"doctors": doctors})
}

func (h *DoctorHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	doctor, err := h.repos.Doctor.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if doctor == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Doctor not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"doctor": doctor})
}

func (h *DoctorHandler) Create(c *gin.Context) {
	var req CreateDoctorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	doctor := &models.Doctor{
		Name:        req.Name,
		Code:        req.Code,
		Preferences: req.Preferences,
	}

	if err := h.repos.Doctor.Create(doctor); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"doctor": doctor})
}

func (h *DoctorHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var req UpdateDoctorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	doctor, err := h.repos.Doctor.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if doctor == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Doctor not found"})
		return
	}

	if req.Name != "" {
		doctor.Name = req.Name
	}
	if req.Code != "" {
		doctor.Code = req.Code
	}
	doctor.Preferences = req.Preferences

	if err := h.repos.Doctor.Update(doctor); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"doctor": doctor})
}

func (h *DoctorHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := h.repos.Doctor.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Doctor deleted successfully"})
}

// Doctor Schedule operations
type CreateDoctorAssignmentRequest struct {
	DoctorID        uuid.UUID `json:"doctor_id" binding:"required"`
	BranchID        uuid.UUID `json:"branch_id" binding:"required"`
	Date            string    `json:"date" binding:"required"`
	ExpectedRevenue float64   `json:"expected_revenue"`
}

func (h *DoctorHandler) CreateAssignment(c *gin.Context) {
	var req CreateDoctorAssignmentRequest
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

	// Validate doctor exists
	doctor, err := h.repos.Doctor.GetByID(req.DoctorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if doctor == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Doctor not found"})
		return
	}

	// Check maximum doctors per branch (6)
	count, err := h.repos.DoctorAssignment.GetDoctorCountByBranch(req.BranchID, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if count >= 6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Maximum 6 doctors can be assigned to a branch per day"})
		return
	}

	assignment := &models.DoctorAssignment{
		ID:              uuid.New(),
		DoctorID:        req.DoctorID,
		BranchID:        req.BranchID,
		Date:            date,
		ExpectedRevenue: req.ExpectedRevenue,
		CreatedBy:       userID,
	}

	if err := h.repos.DoctorAssignment.Create(assignment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"assignment": assignment})
}

func (h *DoctorHandler) GetAssignments(c *gin.Context) {
	branchIDStr := c.Query("branch_id")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	doctorIDStr := c.Query("doctor_id")

	var assignments []*models.DoctorAssignment
	var err error

	if branchIDStr != "" {
		branchID, err := uuid.Parse(branchIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid branch_id"})
			return
		}

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
			// Default to current month
			now := time.Now()
			startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
			endDate = startDate.AddDate(0, 1, -1)
		}

		assignments, err = h.repos.DoctorAssignment.GetByBranchID(branchID, startDate, endDate)
	} else if doctorIDStr != "" {
		doctorID, err := uuid.Parse(doctorIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid doctor_id"})
			return
		}

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
			// Default to current month
			now := time.Now()
			startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
			endDate = startDate.AddDate(0, 1, -1)
		}

		assignments, err = h.repos.DoctorAssignment.GetByDoctorID(doctorID, startDate, endDate)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "branch_id or doctor_id is required"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"assignments": assignments})
}

func (h *DoctorHandler) DeleteAssignment(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := h.repos.DoctorAssignment.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Assignment deleted successfully"})
}

// DoctorOnOffDay handlers
type CreateDoctorOnOffDayRequest struct {
	BranchID   uuid.UUID `json:"branch_id" binding:"required"`
	Date       string    `json:"date" binding:"required"`
	IsDoctorOn bool      `json:"is_doctor_on"`
}

func (h *DoctorHandler) CreateDoctorOnOffDay(c *gin.Context) {
	var req CreateDoctorOnOffDayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format"})
		return
	}

	// For branch managers, enforce their branch
	role := c.GetString("role")
	if role == "branch_manager" {
		userBranchID, exists := c.Get("user_branch_id")
		if exists {
			if userBranchUUID, ok := userBranchID.(uuid.UUID); ok {
				req.BranchID = userBranchUUID
			}
		} else {
			c.JSON(http.StatusForbidden, gin.H{"error": "Branch manager must be assigned to a branch"})
			return
		}
	}

	userIDStr := c.MustGet("user_id").(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Check if already exists, update if so
	existing, err := h.repos.DoctorOnOffDay.GetByBranchAndDate(req.BranchID, date)
	if err == nil && existing != nil {
		existing.IsDoctorOn = req.IsDoctorOn
		if err := h.repos.DoctorOnOffDay.Update(existing); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"day": existing})
		return
	}

	day := &models.DoctorOnOffDay{
		ID:          uuid.New(),
		BranchID:    req.BranchID,
		Date:        date,
		IsDoctorOn:  req.IsDoctorOn,
		CreatedBy:   userID,
	}

	if err := h.repos.DoctorOnOffDay.Create(day); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"day": day})
}

func (h *DoctorHandler) GetDoctorOnOffDays(c *gin.Context) {
	branchIDStr := c.Query("branch_id")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if branchIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "branch_id is required"})
		return
	}

	branchID, err := uuid.Parse(branchIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid branch_id"})
		return
	}

	// For branch managers, enforce their branch
	role := c.GetString("role")
	if role == "branch_manager" {
		userBranchID, exists := c.Get("user_branch_id")
		if exists {
			if userBranchUUID, ok := userBranchID.(uuid.UUID); ok {
				if userBranchUUID != branchID {
					c.JSON(http.StatusForbidden, gin.H{"error": "You can only access doctor schedules for your own branch"})
					return
				}
			}
		}
	}

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
		// Default to current month
		now := time.Now()
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		endDate = startDate.AddDate(0, 1, -1)
	}

	days, err := h.repos.DoctorOnOffDay.GetByBranchID(branchID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"days": days})
}

func (h *DoctorHandler) DeleteDoctorOnOffDay(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := h.repos.DoctorOnOffDay.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Doctor on/off day deleted successfully"})
}

// Doctor Schedule - Monthly view
func (h *DoctorHandler) GetMonthlySchedule(c *gin.Context) {
	doctorIDStr := c.Param("id")
	yearStr := c.Query("year")
	monthStr := c.Query("month")

	doctorID, err := uuid.Parse(doctorIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid doctor_id"})
		return
	}

	var year, month int
	if yearStr != "" && monthStr != "" {
		_, err = fmt.Sscanf(yearStr, "%d", &year)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid year"})
			return
		}
		_, err = fmt.Sscanf(monthStr, "%d", &month)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid month"})
			return
		}
	} else {
		now := time.Now()
		year = now.Year()
		month = int(now.Month())
	}

	assignments, err := h.repos.DoctorAssignment.GetMonthlySchedule(doctorID, year, month)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"assignments": assignments, "year": year, "month": month})
}

// Doctor Preferences operations
type CreateDoctorPreferenceRequest struct {
	DoctorID   uuid.UUID              `json:"doctor_id" binding:"required"`
	BranchID   *uuid.UUID             `json:"branch_id"`
	RuleType   string                 `json:"rule_type" binding:"required"`
	RuleConfig map[string]interface{} `json:"rule_config" binding:"required"`
	IsActive   bool                   `json:"is_active"`
}

func (h *DoctorHandler) ListPreferences(c *gin.Context) {
	doctorIDStr := c.Query("doctor_id")
	branchIDStr := c.Query("branch_id")

	if doctorIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "doctor_id is required"})
		return
	}

	doctorID, err := uuid.Parse(doctorIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid doctor_id"})
		return
	}

	var preferences []*models.DoctorPreference
	if branchIDStr != "" {
		branchID, err := uuid.Parse(branchIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid branch_id"})
			return
		}
		preferences, err = h.repos.DoctorPreference.GetByDoctorAndBranch(doctorID, branchID)
	} else {
		preferences, err = h.repos.DoctorPreference.GetByDoctorID(doctorID)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"preferences": preferences})
}

func (h *DoctorHandler) CreatePreference(c *gin.Context) {
	var req CreateDoctorPreferenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	preference := &models.DoctorPreference{
		DoctorID:   req.DoctorID,
		BranchID:   req.BranchID,
		RuleType:   req.RuleType,
		RuleConfig: req.RuleConfig,
		IsActive:   req.IsActive,
	}

	if err := h.repos.DoctorPreference.Create(preference); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"preference": preference})
}

func (h *DoctorHandler) UpdatePreference(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var req CreateDoctorPreferenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	preference, err := h.repos.DoctorPreference.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if preference == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Preference not found"})
		return
	}

	preference.DoctorID = req.DoctorID
	preference.BranchID = req.BranchID
	preference.RuleType = req.RuleType
	preference.RuleConfig = req.RuleConfig
	preference.IsActive = req.IsActive

	if err := h.repos.DoctorPreference.Update(preference); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"preference": preference})
}

func (h *DoctorHandler) DeletePreference(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := h.repos.DoctorPreference.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Preference deleted successfully"})
}

// Doctor Default Schedule operations
type CreateDoctorDefaultScheduleRequest struct {
	DoctorID  uuid.UUID `json:"doctor_id" binding:"required"`
	DayOfWeek int       `json:"day_of_week" binding:"required,min=0,max=6"` // 0=Sunday, 6=Saturday
	BranchID  uuid.UUID `json:"branch_id" binding:"required"`
}

type UpdateDoctorDefaultScheduleRequest struct {
	BranchID uuid.UUID `json:"branch_id" binding:"required"`
}

func (h *DoctorHandler) CreateDefaultSchedule(c *gin.Context) {
	var req CreateDoctorDefaultScheduleRequest
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

	// Validate doctor exists
	doctor, err := h.repos.Doctor.GetByID(req.DoctorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if doctor == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Doctor not found"})
		return
	}

	schedule := &models.DoctorDefaultSchedule{
		DoctorID:  req.DoctorID,
		DayOfWeek: req.DayOfWeek,
		BranchID:  req.BranchID,
		CreatedBy: userID,
	}

	if err := h.repos.DoctorDefaultSchedule.Upsert(schedule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"schedule": schedule})
}

func (h *DoctorHandler) GetDefaultSchedules(c *gin.Context) {
	doctorIDStr := c.Query("doctor_id")
	if doctorIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "doctor_id is required"})
		return
	}

	doctorID, err := uuid.Parse(doctorIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid doctor_id"})
		return
	}

	schedules, err := h.repos.DoctorDefaultSchedule.GetByDoctorID(doctorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"schedules": schedules})
}

func (h *DoctorHandler) UpdateDefaultSchedule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var req UpdateDoctorDefaultScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	schedule, err := h.repos.DoctorDefaultSchedule.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if schedule == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Schedule not found"})
		return
	}

	schedule.BranchID = req.BranchID
	if err := h.repos.DoctorDefaultSchedule.Update(schedule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"schedule": schedule})
}

func (h *DoctorHandler) DeleteDefaultSchedule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := h.repos.DoctorDefaultSchedule.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Default schedule deleted successfully"})
}

// Doctor Weekly Off Day operations
type CreateDoctorWeeklyOffDayRequest struct {
	DoctorID  uuid.UUID `json:"doctor_id" binding:"required"`
	DayOfWeek int       `json:"day_of_week" binding:"required,min=0,max=6"` // 0=Sunday, 6=Saturday
}

func (h *DoctorHandler) CreateWeeklyOffDay(c *gin.Context) {
	var req CreateDoctorWeeklyOffDayRequest
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

	// Validate doctor exists
	doctor, err := h.repos.Doctor.GetByID(req.DoctorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if doctor == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Doctor not found"})
		return
	}

	// Check if already exists
	existing, err := h.repos.DoctorWeeklyOffDay.GetByDoctorAndDayOfWeek(req.DoctorID, req.DayOfWeek)
	if err != nil && err.Error() != "sql: no rows in result set" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existing != nil {
		c.JSON(http.StatusOK, gin.H{"off_day": existing})
		return
	}

	offDay := &models.DoctorWeeklyOffDay{
		DoctorID:  req.DoctorID,
		DayOfWeek: req.DayOfWeek,
		CreatedBy: userID,
	}

	if err := h.repos.DoctorWeeklyOffDay.Create(offDay); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"off_day": offDay})
}

func (h *DoctorHandler) GetWeeklyOffDays(c *gin.Context) {
	doctorIDStr := c.Query("doctor_id")
	if doctorIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "doctor_id is required"})
		return
	}

	doctorID, err := uuid.Parse(doctorIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid doctor_id"})
		return
	}

	offDays, err := h.repos.DoctorWeeklyOffDay.GetByDoctorID(doctorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"off_days": offDays})
}

func (h *DoctorHandler) DeleteWeeklyOffDay(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := h.repos.DoctorWeeklyOffDay.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Weekly off day deleted successfully"})
}

// Doctor Schedule Override operations
type CreateDoctorScheduleOverrideRequest struct {
	DoctorID uuid.UUID  `json:"doctor_id" binding:"required"`
	Date     string     `json:"date" binding:"required"`
	Type     string     `json:"type" binding:"required,oneof=working off"`
	BranchID *uuid.UUID `json:"branch_id"` // Required if type is "working"
}

func (h *DoctorHandler) CreateScheduleOverride(c *gin.Context) {
	var req CreateDoctorScheduleOverrideRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format"})
		return
	}

	// Validate: if type is "working", branch_id is required
	if req.Type == "working" && req.BranchID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "branch_id is required when type is 'working'"})
		return
	}

	// Validate: if type is "off", branch_id should be nil
	if req.Type == "off" && req.BranchID != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "branch_id must be null when type is 'off'"})
		return
	}

	userIDStr := c.MustGet("user_id").(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Validate doctor exists
	doctor, err := h.repos.Doctor.GetByID(req.DoctorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if doctor == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Doctor not found"})
		return
	}

	// Check if override already exists for this date
	existing, err := h.repos.DoctorScheduleOverride.GetByDoctorAndDate(req.DoctorID, date)
	if err != nil && err.Error() != "sql: no rows in result set" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existing != nil {
		// Update existing override
		existing.Type = req.Type
		existing.BranchID = req.BranchID
		if err := h.repos.DoctorScheduleOverride.Update(existing); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"override": existing})
		return
	}

	override := &models.DoctorScheduleOverride{
		DoctorID: req.DoctorID,
		Date:     date,
		Type:     req.Type,
		BranchID: req.BranchID,
		CreatedBy: userID,
	}

	if err := h.repos.DoctorScheduleOverride.Create(override); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"override": override})
}

func (h *DoctorHandler) GetScheduleOverrides(c *gin.Context) {
	doctorIDStr := c.Query("doctor_id")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if doctorIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "doctor_id is required"})
		return
	}

	doctorID, err := uuid.Parse(doctorIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid doctor_id"})
		return
	}

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
		// Default to current month
		now := time.Now()
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		endDate = startDate.AddDate(0, 1, -1)
	}

	overrides, err := h.repos.DoctorScheduleOverride.GetByDoctorID(doctorID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"overrides": overrides})
}

func (h *DoctorHandler) UpdateScheduleOverride(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var req CreateDoctorScheduleOverrideRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	override, err := h.repos.DoctorScheduleOverride.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if override == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Override not found"})
		return
	}

	// Validate: if type is "working", branch_id is required
	if req.Type == "working" && req.BranchID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "branch_id is required when type is 'working'"})
		return
	}

	// Validate: if type is "off", branch_id should be nil
	if req.Type == "off" && req.BranchID != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "branch_id must be null when type is 'off'"})
		return
	}

	override.Type = req.Type
	override.BranchID = req.BranchID
	if err := h.repos.DoctorScheduleOverride.Update(override); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"override": override})
}

func (h *DoctorHandler) DeleteScheduleOverride(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := h.repos.DoctorScheduleOverride.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Schedule override deleted successfully"})
}

// Import doctors from Excel
func (h *DoctorHandler) Import(c *gin.Context) {
	// Get the uploaded file
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open uploaded file"})
		return
	}
	defer src.Close()

	// Read file data
	fileData := make([]byte, file.Size)
	_, err = src.Read(fileData)
	if err != nil && err != io.EOF {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file data"})
		return
	}

	// Import doctors from Excel
	doctorList, parseErr := h.excelImporter.ImportDoctors(fileData)
	if parseErr != nil {
		// If no valid records were parsed, return error immediately
		if len(doctorList) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": parseErr.Error()})
			return
		}
		// If there are valid records but also parsing errors, continue to save the valid ones
		// The parsing errors will be included as warnings in the response
	}

	// Save imported doctors to database
	var savedDoctors []*models.Doctor
	var saveErrors []string
	for _, doctor := range doctorList {
		// Check if doctor with same code already exists (if code is provided)
		if doctor.Code != "" {
			existing, err := h.repos.Doctor.GetByCode(doctor.Code)
			if err == nil && existing != nil {
				saveErrors = append(saveErrors, fmt.Sprintf("Doctor code '%s' already exists: %s", doctor.Code, doctor.Name))
				continue
			}
		}

		if err := h.repos.Doctor.Create(doctor); err != nil {
			saveErrors = append(saveErrors, fmt.Sprintf("Failed to save %s: %v", doctor.Name, err))
			continue
		}
		savedDoctors = append(savedDoctors, doctor)
	}

	// If no records were saved at all, return error
	if len(savedDoctors) == 0 {
		errorMsg := "Failed to save any doctor records"
		if parseErr != nil {
			errorMsg = fmt.Sprintf("%s. Parse errors: %v", errorMsg, parseErr)
		}
		if len(saveErrors) > 0 {
			errorMsg = fmt.Sprintf("%s. Save errors: %s", errorMsg, saveErrors[0])
			if len(saveErrors) > 1 {
				errorMsg = fmt.Sprintf("%s (and %d more)", errorMsg, len(saveErrors)-1)
			}
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   errorMsg,
			"details": saveErrors,
		})
		return
	}

	// Build response with actual saved count
	response := gin.H{
		"message":     fmt.Sprintf("Successfully imported %d doctor(s)", len(savedDoctors)),
		"imported":    len(savedDoctors),
		"total_rows": len(doctorList),
		"doctors":    savedDoctors,
	}

	// Add parsing warnings if any
	if parseErr != nil {
		response["parse_warnings"] = parseErr.Error()
	}

	// Add save errors as warnings if any records were saved
	if len(saveErrors) > 0 {
		response["save_warnings"] = saveErrors
		response["message"] = fmt.Sprintf("Import completed with %d warnings", len(saveErrors))
	}

	// Return partial content if there were any errors (parsing or saving)
	if parseErr != nil || len(saveErrors) > 0 {
		c.JSON(http.StatusPartialContent, response)
		return
	}

	c.JSON(http.StatusOK, response)
}

// ImportDefaultSchedules imports doctor default schedules from Excel
func (h *DoctorHandler) ImportDefaultSchedules(c *gin.Context) {
	// Get the uploaded file
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open uploaded file"})
		return
	}
	defer src.Close()

	// Read file data
	fileData := make([]byte, file.Size)
	_, err = src.Read(fileData)
	if err != nil && err != io.EOF {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file data"})
		return
	}

	// Import default schedules from Excel
	result, parseErr := h.excelImporter.ImportDefaultSchedules(fileData)
	if parseErr != nil {
		// If no valid records were parsed, return error immediately
		if result == nil || (len(result.Schedules) == 0 && len(result.OffDays) == 0) {
			c.JSON(http.StatusBadRequest, gin.H{"error": parseErr.Error()})
			return
		}
		// If there are valid records but also parsing errors, continue to save the valid ones
		// The parsing errors will be included as warnings in the response
	}

	// Get user ID for CreatedBy field
	userIDStr := c.MustGet("user_id").(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// First, delete schedules for off days
	var deletedOffDays int
	var deleteErrors []string
	for _, offDay := range result.OffDays {
		// Find existing schedule for this doctor and day
		existing, err := h.repos.DoctorDefaultSchedule.GetByDoctorAndDayOfWeek(offDay.DoctorID, offDay.DayOfWeek)
		if err == nil && existing != nil {
			if err := h.repos.DoctorDefaultSchedule.Delete(existing.ID); err != nil {
				deleteErrors = append(deleteErrors, fmt.Sprintf("Failed to delete schedule for doctor %s, day %d: %v", offDay.DoctorID, offDay.DayOfWeek, err))
				continue
			}
			deletedOffDays++
		}
	}

	// Then, save imported schedules to database
	var savedSchedules []*models.DoctorDefaultSchedule
	var saveErrors []string
	for _, schedule := range result.Schedules {
		// Set CreatedBy
		schedule.CreatedBy = userID

		// Use Upsert to handle duplicates (last imported entry wins)
		// Upsert will update existing schedule if doctor_id + day_of_week combination exists
		if err := h.repos.DoctorDefaultSchedule.Upsert(schedule); err != nil {
			saveErrors = append(saveErrors, fmt.Sprintf("Failed to save schedule for doctor %s, day %d: %v", schedule.DoctorID, schedule.DayOfWeek, err))
			continue
		}
		savedSchedules = append(savedSchedules, schedule)
	}

	// If no records were saved and no off days were deleted, return error
	if len(savedSchedules) == 0 && deletedOffDays == 0 {
		errorMsg := "Failed to save any schedule records"
		if parseErr != nil {
			errorMsg = fmt.Sprintf("%s. Parse errors: %v", errorMsg, parseErr)
		}
		if len(saveErrors) > 0 {
			errorMsg = fmt.Sprintf("%s. Save errors: %s", errorMsg, saveErrors[0])
			if len(saveErrors) > 1 {
				errorMsg = fmt.Sprintf("%s (and %d more)", errorMsg, len(saveErrors)-1)
			}
		}
		if len(deleteErrors) > 0 {
			errorMsg = fmt.Sprintf("%s. Delete errors: %s", errorMsg, deleteErrors[0])
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   errorMsg,
			"details": saveErrors,
		})
		return
	}

	// Build response with actual saved count
	totalProcessed := len(savedSchedules) + deletedOffDays
	response := gin.H{
		"message":        fmt.Sprintf("Successfully imported %d default schedule(s) and set %d off day(s)", len(savedSchedules), deletedOffDays),
		"imported":       len(savedSchedules),
		"off_days_set":   deletedOffDays,
		"total_processed": totalProcessed,
		"schedules":      savedSchedules,
	}

	// Add parsing warnings if any
	if parseErr != nil {
		response["parse_warnings"] = parseErr.Error()
	}

	// Add save errors as warnings if any records were saved
	if len(saveErrors) > 0 {
		response["save_warnings"] = saveErrors
		response["message"] = fmt.Sprintf("Import completed with %d warnings", len(saveErrors))
	}

	// Return partial content if there were any errors (parsing or saving)
	if parseErr != nil || len(saveErrors) > 0 {
		c.JSON(http.StatusPartialContent, response)
		return
	}

	c.JSON(http.StatusOK, response)
}
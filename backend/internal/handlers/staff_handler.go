package handlers

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"vsq-oper-manpower/backend/internal/domain/interfaces"
	"vsq-oper-manpower/backend/internal/domain/models"
	"vsq-oper-manpower/backend/internal/repositories/postgres"
	"vsq-oper-manpower/backend/pkg/excel"
)

type StaffHandler struct {
	repos         *postgres.Repositories
	excelImporter *excel.ExcelImporter
}

func NewStaffHandler(repos *postgres.Repositories) *StaffHandler {
	return &StaffHandler{
		repos: repos,
		excelImporter: excel.NewExcelImporter(
			repos.Position,
			repos.Branch,
			repos.Doctor,
		),
	}
}

type CreateStaffRequest struct {
	Nickname          string      `json:"nickname"`
	Name              string      `json:"name" binding:"required"`
	StaffType         string      `json:"staff_type" binding:"required"`
	PositionID        uuid.UUID   `json:"position_id" binding:"required"`
	BranchID          *uuid.UUID  `json:"branch_id,omitempty"`
	CoverageArea      string      `json:"coverage_area"`
	AreaOfOperationID *uuid.UUID  `json:"area_of_operation_id,omitempty"` // Legacy field
	ZoneID            *uuid.UUID  `json:"zone_id,omitempty"`              // Zone assignment for rotation staff
	BranchIDs         []uuid.UUID `json:"branch_ids,omitempty"`           // Individual branches for rotation staff
	SkillLevel        int         `json:"skill_level" binding:"min=0,max=10"`
}

type UpdateStaffRequest struct {
	Nickname          *string     `json:"nickname,omitempty"`
	Name              *string     `json:"name,omitempty"`
	StaffType         *string     `json:"staff_type,omitempty"`
	PositionID        *uuid.UUID  `json:"position_id,omitempty"`
	BranchID          *uuid.UUID  `json:"branch_id,omitempty"`
	CoverageArea      *string     `json:"coverage_area,omitempty"`
	AreaOfOperationID *uuid.UUID  `json:"area_of_operation_id,omitempty"` // Legacy field
	ZoneID            *uuid.UUID  `json:"zone_id,omitempty"`              // Zone assignment for rotation staff
	BranchIDs         []uuid.UUID `json:"branch_ids,omitempty"`           // Individual branches for rotation staff
	SkillLevel        *int        `json:"skill_level,omitempty" binding:"omitempty,min=0,max=10"`
}

func (h *StaffHandler) List(c *gin.Context) {
	staffType := c.Query("staff_type")
	branchIDStr := c.Query("branch_id")
	positionIDStr := c.Query("position_id")

	filters := interfaces.StaffFilters{}
	if staffType != "" {
		st := models.StaffType(staffType)
		filters.StaffType = &st
	}
	
	// For branch managers, enforce their branch
	role := c.GetString("role")
	if role == "branch_manager" {
		userBranchID, exists := c.Get("user_branch_id")
		if exists {
			if branchUUID, ok := userBranchID.(uuid.UUID); ok {
				filters.BranchID = &branchUUID
			}
		}
	} else if branchIDStr != "" {
		// Other roles can specify branch_id
		branchID, err := uuid.Parse(branchIDStr)
		if err == nil {
			filters.BranchID = &branchID
		}
	}
	
	if positionIDStr != "" {
		positionID, err := uuid.Parse(positionIDStr)
		if err == nil {
			filters.PositionID = &positionID
		}
	}

	staffList, err := h.repos.Staff.List(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"staff": staffList})
}

func (h *StaffHandler) Create(c *gin.Context) {
	var req CreateStaffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// For branch managers, enforce their branch and prevent rotation staff creation
	role := c.GetString("role")
	if role == "branch_manager" {
		// Branch managers cannot create rotation staff
		if models.StaffType(req.StaffType) == models.StaffTypeRotation {
			c.JSON(http.StatusForbidden, gin.H{"error": "Branch managers cannot add rotation staff"})
			return
		}
		
		// Branch managers can only create branch staff for their branch
		if models.StaffType(req.StaffType) != models.StaffTypeBranch {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Branch managers can only add branch staff"})
			return
		}
		
		userBranchID, exists := c.Get("user_branch_id")
		if exists {
			if branchUUID, ok := userBranchID.(uuid.UUID); ok {
				req.BranchID = &branchUUID
			}
		} else {
			c.JSON(http.StatusForbidden, gin.H{"error": "Branch manager must be assigned to a branch"})
			return
		}
	}

	// Set default skill level if not provided (default to 5)
	skillLevel := req.SkillLevel
	if skillLevel == 0 {
		skillLevel = 5 // Default to 5 if not specified
	}

	staff := &models.Staff{
		ID:                uuid.New(),
		Nickname:          req.Nickname,
		Name:              req.Name,
		StaffType:         models.StaffType(req.StaffType),
		PositionID:        req.PositionID,
		BranchID:          req.BranchID,
		CoverageArea:      req.CoverageArea,
		AreaOfOperationID: req.AreaOfOperationID,
		ZoneID:            req.ZoneID,
		SkillLevel:        skillLevel,
	}

	if err := h.repos.Staff.Create(staff); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Save individual branches if this is rotation staff
	if models.StaffType(req.StaffType) == models.StaffTypeRotation && len(req.BranchIDs) > 0 {
		if err := h.repos.Staff.BulkUpdateBranches(staff.ID, req.BranchIDs); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to save branches: %v", err)})
			return
		}
	}

	c.JSON(http.StatusCreated, gin.H{"staff": staff})
}

func (h *StaffHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	// Check if staff exists and get current data
	existingStaff, err := h.repos.Staff.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Staff not found"})
		return
	}

	var req UpdateStaffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// For branch managers, enforce restrictions
	role := c.GetString("role")
	if role == "branch_manager" {
		// Branch managers cannot edit rotation staff
		if existingStaff.StaffType == models.StaffTypeRotation {
			c.JSON(http.StatusForbidden, gin.H{"error": "Branch managers cannot edit rotation staff"})
			return
		}
		
		// Branch managers cannot change staff type to rotation
		if req.StaffType != nil && models.StaffType(*req.StaffType) == models.StaffTypeRotation {
			c.JSON(http.StatusForbidden, gin.H{"error": "Branch managers cannot change staff type to rotation"})
			return
		}
		
		// Ensure branch manager can only edit staff from their branch
		userBranchID, exists := c.Get("user_branch_id")
		if exists {
			if branchUUID, ok := userBranchID.(uuid.UUID); ok {
				if existingStaff.BranchID == nil || *existingStaff.BranchID != branchUUID {
					c.JSON(http.StatusForbidden, gin.H{"error": "Branch managers can only edit staff from their own branch"})
					return
				}
				// Force branch ID to their branch
				req.BranchID = &branchUUID
			}
		}
	}

	// Merge request data with existing staff data (only update fields that are provided)
	staff := &models.Staff{
		ID:                id,
		Nickname:          existingStaff.Nickname,
		Name:              existingStaff.Name,
		StaffType:         existingStaff.StaffType,
		PositionID:        existingStaff.PositionID,
		BranchID:          existingStaff.BranchID,
		CoverageArea:      existingStaff.CoverageArea,
		AreaOfOperationID: existingStaff.AreaOfOperationID,
		ZoneID:            existingStaff.ZoneID,
		SkillLevel:        existingStaff.SkillLevel,
	}

	// Update fields that are provided in the request
	if req.Nickname != nil {
		staff.Nickname = *req.Nickname
	}
	if req.Name != nil {
		staff.Name = *req.Name
	}
	if req.StaffType != nil {
		staff.StaffType = models.StaffType(*req.StaffType)
	}
	if req.PositionID != nil {
		staff.PositionID = *req.PositionID
	}
	if req.BranchID != nil {
		staff.BranchID = req.BranchID
	}
	if req.CoverageArea != nil {
		staff.CoverageArea = *req.CoverageArea
	}
	if req.AreaOfOperationID != nil {
		staff.AreaOfOperationID = req.AreaOfOperationID
	}
	if req.ZoneID != nil {
		staff.ZoneID = req.ZoneID
	}
	if req.SkillLevel != nil {
		staff.SkillLevel = *req.SkillLevel
	} else if staff.SkillLevel == 0 {
		staff.SkillLevel = 5 // Default to 5 if existing is also 0
	}

	if err := h.repos.Staff.Update(staff); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Update individual branches if this is rotation staff and branch_ids is provided
	// If branch_ids is provided (even if empty array), update branches
	// If branch_ids is nil/not provided, don't update branches
	if staff.StaffType == models.StaffTypeRotation && req.BranchIDs != nil {
		if err := h.repos.Staff.BulkUpdateBranches(id, req.BranchIDs); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to update branches: %v", err)})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"staff": staff})
}

func (h *StaffHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	// Check if staff exists and get current data
	existingStaff, err := h.repos.Staff.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Staff not found"})
		return
	}

	// For branch managers, enforce restrictions
	role := c.GetString("role")
	if role == "branch_manager" {
		// Branch managers cannot delete rotation staff
		if existingStaff.StaffType == models.StaffTypeRotation {
			c.JSON(http.StatusForbidden, gin.H{"error": "Branch managers cannot delete rotation staff"})
			return
		}
		
		// Ensure branch manager can only delete staff from their branch
		userBranchID, exists := c.Get("user_branch_id")
		if exists {
			if branchUUID, ok := userBranchID.(uuid.UUID); ok {
				if existingStaff.BranchID == nil || *existingStaff.BranchID != branchUUID {
					c.JSON(http.StatusForbidden, gin.H{"error": "Branch managers can only delete staff from their own branch"})
					return
				}
			}
		}
	}

	// Delete related records before deleting staff
	// 1. Delete all schedules for this staff
	if err := h.repos.Schedule.DeleteByStaffID(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to delete staff schedules: %v", err)})
		return
	}

	// 2. If rotation staff, delete rotation assignments and effective branches
	if existingStaff.StaffType == models.StaffTypeRotation {
		// Delete rotation assignments
		if err := h.repos.Rotation.DeleteByRotationStaffID(id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to delete rotation assignments: %v", err)})
			return
		}
		// Delete effective branches
		if err := h.repos.EffectiveBranch.DeleteByRotationStaffID(id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to delete effective branches: %v", err)})
			return
		}
	}

	// 3. Finally, delete the staff member
	if err := h.repos.Staff.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to delete staff: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Staff deleted successfully"})
}

func (h *StaffHandler) Import(c *gin.Context) {
	// Get the uploaded file
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
		return
	}
	defer src.Close()

	// Read file content
	fileData, err := io.ReadAll(src)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}

	// Import staff from Excel
	staffList, parseErr := h.excelImporter.ImportStaff(fileData)
	if parseErr != nil {
		// If no valid records were parsed, return error immediately
		if len(staffList) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": parseErr.Error()})
			return
		}
		// If there are valid records but also parsing errors, continue to save the valid ones
		// The parsing errors will be included as warnings in the response
	}

	// Save imported staff to database
	var savedStaff []*models.Staff
	var saveErrors []string
	for _, staff := range staffList {
		if err := h.repos.Staff.Create(staff); err != nil {
			saveErrors = append(saveErrors, fmt.Sprintf("Failed to save %s: %v", staff.Name, err))
			continue
		}
		savedStaff = append(savedStaff, staff)
	}

	// If no records were saved at all, return error
	if len(savedStaff) == 0 {
		errorMsg := "Failed to save any staff records"
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
		"message":    "Import completed",
		"imported":    len(savedStaff),
		"total_rows": len(staffList),
		"staff":      savedStaff,
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


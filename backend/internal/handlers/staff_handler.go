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
		repos:         repos,
		excelImporter: excel.NewExcelImporter(),
	}
}

type CreateStaffRequest struct {
	Name         string    `json:"name" binding:"required"`
	StaffType    string    `json:"staff_type" binding:"required"`
	PositionID   uuid.UUID `json:"position_id" binding:"required"`
	BranchID     *uuid.UUID `json:"branch_id,omitempty"`
	CoverageArea string    `json:"coverage_area"`
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
	if branchIDStr != "" {
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

	staff := &models.Staff{
		ID:           uuid.New(),
		Name:         req.Name,
		StaffType:    models.StaffType(req.StaffType),
		PositionID:   req.PositionID,
		BranchID:     req.BranchID,
		CoverageArea: req.CoverageArea,
	}

	if err := h.repos.Staff.Create(staff); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
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

	var req CreateStaffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	staff := &models.Staff{
		ID:           id,
		Name:         req.Name,
		StaffType:    models.StaffType(req.StaffType),
		PositionID:   req.PositionID,
		BranchID:     req.BranchID,
		CoverageArea: req.CoverageArea,
	}

	if err := h.repos.Staff.Update(staff); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
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

	if err := h.repos.Staff.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
	staffList, err := h.excelImporter.ImportStaff(fileData)
	if err != nil {
		// Check if it's a partial success (some records imported, some failed)
		if len(staffList) > 0 {
			c.JSON(http.StatusPartialContent, gin.H{
				"error":      err.Error(),
				"imported":   len(staffList),
				"staff":      staffList,
			})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Save imported staff to database
	var savedStaff []*models.Staff
	var errors []string
	for _, staff := range staffList {
		if err := h.repos.Staff.Create(staff); err != nil {
			errors = append(errors, fmt.Sprintf("Failed to save %s: %v", staff.Name, err))
			continue
		}
		savedStaff = append(savedStaff, staff)
	}

	if len(errors) > 0 && len(savedStaff) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to save any staff records",
			"details": errors,
		})
		return
	}

	response := gin.H{
		"message":    "Import completed",
		"imported":    len(savedStaff),
		"total_rows": len(staffList),
		"staff":      savedStaff,
	}

	if len(errors) > 0 {
		response["warnings"] = errors
		response["message"] = fmt.Sprintf("Import completed with %d warnings", len(errors))
	}

	c.JSON(http.StatusOK, response)
}


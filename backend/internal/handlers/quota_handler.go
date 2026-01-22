package handlers

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"vsq-oper-manpower/backend/internal/domain/interfaces"
	"vsq-oper-manpower/backend/internal/domain/models"
	"vsq-oper-manpower/backend/internal/repositories/postgres"
	"vsq-oper-manpower/backend/internal/usecases/allocation"
	"vsq-oper-manpower/backend/pkg/excel"
)

type QuotaHandler struct {
	repos           *postgres.Repositories
	quotaCalculator *allocation.QuotaCalculator
	excelImporter   *excel.ExcelImporter
}

func NewQuotaHandler(repos *postgres.Repositories, quotaCalculator *allocation.QuotaCalculator) *QuotaHandler {
	return &QuotaHandler{
		repos:           repos,
		quotaCalculator: quotaCalculator,
		excelImporter: excel.NewExcelImporter(
			repos.Position,
			repos.Branch,
			repos.Doctor,
			repos.PositionQuota,
		),
	}
}

type CreatePositionQuotaRequest struct {
	BranchID        uuid.UUID `json:"branch_id" binding:"required"`
	PositionID      uuid.UUID `json:"position_id" binding:"required"`
	DesignatedQuota int       `json:"designated_quota" binding:"required"`
	MinimumRequired int       `json:"minimum_required" binding:"required"`
}

func (h *QuotaHandler) CreateQuota(c *gin.Context) {
	var req CreatePositionQuotaRequest
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

	quota := &models.PositionQuota{
		ID:              uuid.New(),
		BranchID:        req.BranchID,
		PositionID:      req.PositionID,
		DesignatedQuota: req.DesignatedQuota,
		MinimumRequired: req.MinimumRequired,
		IsActive:        true,
		CreatedBy:       userID,
	}

	if err := h.repos.PositionQuota.Create(quota); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"quota": quota})
}

func (h *QuotaHandler) GetQuotas(c *gin.Context) {
	branchIDStr := c.Query("branch_id")
	positionIDStr := c.Query("position_id")

	var quotas []*models.PositionQuota
	var err error

	if branchIDStr != "" {
		branchID, err := uuid.Parse(branchIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid branch_id"})
			return
		}
		quotas, err = h.repos.PositionQuota.GetByBranchID(branchID)
	} else {
		filters := interfaces.PositionQuotaFilters{}
		if positionIDStr != "" {
			positionID, err := uuid.Parse(positionIDStr)
			if err == nil {
				filters.PositionID = &positionID
			}
		}
		isActive := true
		filters.IsActive = &isActive
		quotas, err = h.repos.PositionQuota.List(filters)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"quotas": quotas})
}

func (h *QuotaHandler) UpdateQuota(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var req struct {
		DesignatedQuota *int  `json:"designated_quota"`
		MinimumRequired *int  `json:"minimum_required"`
		IsActive        *bool `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	quota, err := h.repos.PositionQuota.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if quota == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Quota not found"})
		return
	}

	if req.DesignatedQuota != nil {
		quota.DesignatedQuota = *req.DesignatedQuota
	}
	if req.MinimumRequired != nil {
		quota.MinimumRequired = *req.MinimumRequired
	}
	if req.IsActive != nil {
		quota.IsActive = *req.IsActive
	}

	if err := h.repos.PositionQuota.Update(quota); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"quota": quota})
}

func (h *QuotaHandler) DeleteQuota(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := h.repos.PositionQuota.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Quota deleted successfully"})
}

func (h *QuotaHandler) GetBranchQuotaStatus(c *gin.Context) {
	branchIDStr := c.Param("branchId")
	if branchIDStr == "" {
		branchIDStr = c.Query("branch_id")
	}

	if branchIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "branch_id is required"})
		return
	}

	branchID, err := uuid.Parse(branchIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid branch_id"})
		return
	}

	dateStr := c.Query("date")
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format"})
		return
	}

	status, err := h.quotaCalculator.CalculateBranchQuotaStatus(branchID, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": status})
}

func (h *QuotaHandler) Import(c *gin.Context) {
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

	// Get user ID for created_by field
	userIDStr := c.MustGet("user_id").(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Import position quotas from Excel
	result, importErr := h.excelImporter.ImportPositionQuotas(fileData, userID)
	if importErr != nil {
		// If no records were imported, return error immediately
		if result == nil || (result.Created == 0 && result.Updated == 0) {
			c.JSON(http.StatusBadRequest, gin.H{"error": importErr.Error()})
			return
		}
		// If there are valid records but also import errors, continue with partial success
	}

	// Prepare response
	response := gin.H{
		"message": fmt.Sprintf("Import completed: %d created, %d updated", result.Created, result.Updated),
		"created": result.Created,
		"updated": result.Updated,
	}

	// Include errors if any
	if len(result.Errors) > 0 {
		response["errors"] = result.Errors
		response["warning"] = fmt.Sprintf("Import completed with %d error(s)", len(result.Errors))
	}

	// If there were import errors but we still have a result, return partial success
	if importErr != nil && (result.Created > 0 || result.Updated > 0) {
		c.JSON(http.StatusOK, response)
		return
	}

	// Success
	c.JSON(http.StatusOK, response)
}

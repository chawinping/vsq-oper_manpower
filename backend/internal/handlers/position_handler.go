package handlers

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"vsq-oper-manpower/backend/internal/domain/interfaces"
	"vsq-oper-manpower/backend/internal/domain/models"
	"vsq-oper-manpower/backend/internal/repositories/postgres"
)

type PositionHandler struct {
	repos *postgres.Repositories
	db    *sql.DB
}

func NewPositionHandler(repos *postgres.Repositories, db *sql.DB) *PositionHandler {
	return &PositionHandler{repos: repos, db: db}
}

func (h *PositionHandler) List(c *gin.Context) {
	positions, err := h.repos.Position.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"positions": positions})
}

func (h *PositionHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	position, err := h.repos.Position.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if position == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Position not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"position": position})
}

type UpdatePositionRequest struct {
	Name         string                `json:"name" binding:"required"`
	PositionCode *string               `json:"position_code,omitempty"`
	DisplayOrder int                   `json:"display_order"`
	PositionType models.PositionType   `json:"position_type" binding:"required"`
	ManpowerType models.ManpowerType   `json:"manpower_type" binding:"required"`
}

func (h *PositionHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	// Check if position exists
	existingPosition, err := h.repos.Position.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existingPosition == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Position not found"})
		return
	}

	var req UpdatePositionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	position := &models.Position{
		ID:                id,
		Name:              req.Name,
		PositionCode:      req.PositionCode,
		PositionType:      req.PositionType,
		ManpowerType:      req.ManpowerType,
		MinStaffPerBranch: existingPosition.MinStaffPerBranch, // Keep existing value, don't update
		DisplayOrder:      req.DisplayOrder,
		CreatedAt:         existingPosition.CreatedAt,
	}

	if err := h.repos.Position.Update(position); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"position": position})
}

func (h *PositionHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	// Check if position exists
	existingPosition, err := h.repos.Position.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existingPosition == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Position not found"})
		return
	}

	// Automatically delete all associated position quotas
	quotas, err := h.repos.PositionQuota.List(interfaces.PositionQuotaFilters{
		PositionID: &id,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get position quotas: %v", err)})
		return
	}
	for _, quota := range quotas {
		if err := h.repos.PositionQuota.Delete(quota.ID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to delete position quota: %v", err)})
			return
		}
	}

	// Check if position has any other associations (staff, rules, suggestions, or scenario requirements)
	// Note: We exclude quotas from this check since we've already deleted them
	staffCountQuery := `SELECT COUNT(*) FROM staff WHERE position_id = $1`
	var staffCount int
	err = h.db.QueryRow(staffCountQuery, id).Scan(&staffCount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ruleCountQuery := `SELECT COUNT(*) FROM staff_allocation_rules WHERE position_id = $1`
	var ruleCount int
	err = h.db.QueryRow(ruleCountQuery, id).Scan(&ruleCount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	suggestionCountQuery := `SELECT COUNT(*) FROM allocation_suggestions WHERE position_id = $1`
	var suggestionCount int
	err = h.db.QueryRow(suggestionCountQuery, id).Scan(&suggestionCount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	scenarioCountQuery := `SELECT COUNT(*) FROM scenario_position_requirements WHERE position_id = $1`
	var scenarioCount int
	err = h.db.QueryRow(scenarioCountQuery, id).Scan(&scenarioCount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	totalOtherAssociations := staffCount + ruleCount + suggestionCount + scenarioCount
	if totalOtherAssociations > 0 {
		c.JSON(http.StatusConflict, gin.H{
			"error": fmt.Sprintf("Cannot delete position: position is associated with %d staff member(s), %d allocation rule(s), %d suggestion(s), and %d scenario requirement(s). Please remove all associations before deleting.", staffCount, ruleCount, suggestionCount, scenarioCount),
		})
		return
	}

	// Delete the position
	if err := h.repos.Position.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Position deleted successfully"})
}

// PositionAssociations represents all associations for a position
type PositionAssociations struct {
	StaffCount              int                      `json:"staff_count"`
	QuotaCount              int                      `json:"quota_count"`
	Quotas                  []PositionQuotaAssociation `json:"quotas"`
	AllocationRuleCount     int                       `json:"allocation_rule_count"`
	SuggestionCount         int                       `json:"suggestion_count"`
	ScenarioRequirementCount int                      `json:"scenario_requirement_count"`
	TotalCount              int                       `json:"total_count"`
}

type PositionQuotaAssociation struct {
	QuotaID         string `json:"quota_id"`
	BranchID        string `json:"branch_id"`
	BranchName      string `json:"branch_name"`
	DesignatedQuota int    `json:"designated_quota"`
	MinimumRequired int    `json:"minimum_required"`
}

// GetAssociations returns detailed information about position associations
func (h *PositionHandler) GetAssociations(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	// Check if position exists
	existingPosition, err := h.repos.Position.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existingPosition == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Position not found"})
		return
	}

	associations := &PositionAssociations{}

	// Get all quotas for this position
	allQuotas, err := h.repos.PositionQuota.List(interfaces.PositionQuotaFilters{
		PositionID: &id,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	associations.QuotaCount = len(allQuotas)

	// Get quota details with branch names
	if associations.QuotaCount > 0 {
		for _, quota := range allQuotas {
			branch, err := h.repos.Branch.GetByID(quota.BranchID)
			if err != nil {
				// Continue even if branch lookup fails
				continue
			}
			branchName := "Unknown Branch"
			if branch != nil {
				branchName = branch.Name
			}
			associations.Quotas = append(associations.Quotas, PositionQuotaAssociation{
				QuotaID:         quota.ID.String(),
				BranchID:        quota.BranchID.String(),
				BranchName:      branchName,
				DesignatedQuota: quota.DesignatedQuota,
				MinimumRequired: quota.MinimumRequired,
			})
		}
	}

	// Get staff count
	staffCountQuery := `SELECT COUNT(*) FROM staff WHERE position_id = $1`
	err = h.db.QueryRow(staffCountQuery, id).Scan(&associations.StaffCount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get allocation rule count
	ruleCountQuery := `SELECT COUNT(*) FROM staff_allocation_rules WHERE position_id = $1`
	err = h.db.QueryRow(ruleCountQuery, id).Scan(&associations.AllocationRuleCount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get suggestion count
	suggestionCountQuery := `SELECT COUNT(*) FROM allocation_suggestions WHERE position_id = $1`
	err = h.db.QueryRow(suggestionCountQuery, id).Scan(&associations.SuggestionCount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get scenario requirement count
	scenarioCountQuery := `SELECT COUNT(*) FROM scenario_position_requirements WHERE position_id = $1`
	err = h.db.QueryRow(scenarioCountQuery, id).Scan(&associations.ScenarioRequirementCount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	associations.TotalCount = associations.StaffCount + associations.QuotaCount + 
		associations.AllocationRuleCount + associations.SuggestionCount + associations.ScenarioRequirementCount

	c.JSON(http.StatusOK, gin.H{"associations": associations})
}


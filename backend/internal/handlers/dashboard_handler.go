package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"vsq-oper-manpower/backend/internal/domain/interfaces"
	"vsq-oper-manpower/backend/internal/repositories/postgres"
)

type DashboardHandler struct {
	repos *postgres.Repositories
}

func NewDashboardHandler(repos *postgres.Repositories) *DashboardHandler {
	return &DashboardHandler{repos: repos}
}

func (h *DashboardHandler) GetOverview(c *gin.Context) {
	// Get summary statistics
	branches, err := h.repos.Branch.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	filters := interfaces.StaffFilters{}
	staffList, err := h.repos.Staff.List(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"summary": gin.H{
			"total_branches": len(branches),
			"total_staff":    len(staffList),
		},
	})
}


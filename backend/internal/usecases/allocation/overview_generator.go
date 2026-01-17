package allocation

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// OverviewGenerator generates overview data for Area Managers
type OverviewGenerator struct {
	repos            *RepositoriesWrapper
	quotaCalculator *QuotaCalculator
}

// NewOverviewGenerator creates a new overview generator
func NewOverviewGenerator(repos *RepositoriesWrapper, quotaCalculator *QuotaCalculator) *OverviewGenerator {
	return &OverviewGenerator{
		repos:           repos,
		quotaCalculator: quotaCalculator,
	}
}

// DayOverview represents overview data for all branches on a specific day
type DayOverview struct {
	Date            time.Time `json:"date"`
	BranchStatuses  []*BranchQuotaStatus `json:"branch_statuses"`
	TotalBranches   int       `json:"total_branches"`
	BranchesWithShortage int  `json:"branches_with_shortage"`
}

// MonthlyOverview represents overview data for a single branch across a month
type MonthlyOverview struct {
	BranchID        uuid.UUID `json:"branch_id"`
	BranchName      string    `json:"branch_name"`
	BranchCode      string    `json:"branch_code"`
	Year            int       `json:"year"`
	Month           int       `json:"month"`
	DayStatuses     []*BranchQuotaStatus `json:"day_statuses"`
	AverageFulfillment float64 `json:"average_fulfillment"`
}

// GenerateDayOverview generates overview for all branches on a specific day
func (g *OverviewGenerator) GenerateDayOverview(date time.Time) (*DayOverview, error) {
	// Get all branches
	branches, err := g.repos.Branch.List()
	if err != nil {
		return nil, fmt.Errorf("failed to get branches: %w", err)
	}

	branchIDs := []uuid.UUID{}
	for _, branch := range branches {
		branchIDs = append(branchIDs, branch.ID)
	}

	// Calculate quota status for all branches
	branchStatuses, err := g.quotaCalculator.CalculateBranchesQuotaStatus(branchIDs, date)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate quota statuses: %w", err)
	}

	// Count branches with shortage
	branchesWithShortage := 0
	for _, status := range branchStatuses {
		if status.TotalRequired > 0 {
			branchesWithShortage++
		}
	}

	return &DayOverview{
		Date:              date,
		BranchStatuses:    branchStatuses,
		TotalBranches:     len(branchStatuses),
		BranchesWithShortage: branchesWithShortage,
	}, nil
}

// GenerateMonthlyOverview generates overview for a single branch across a month
func (g *OverviewGenerator) GenerateMonthlyOverview(branchID uuid.UUID, year int, month int) (*MonthlyOverview, error) {
	// Get branch info
	branch, err := g.repos.Branch.GetByID(branchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get branch: %w", err)
	}
	if branch == nil {
		return nil, fmt.Errorf("branch not found")
	}

	// Calculate quota status for the month
	dayStatuses, err := g.quotaCalculator.CalculateMonthlyQuotaStatus(branchID, year, month)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate monthly quota status: %w", err)
	}

	// Calculate average fulfillment
	totalFulfillment := 0.0
	daysWithQuota := 0
	for _, status := range dayStatuses {
		if status.TotalDesignated > 0 {
			fulfillment := float64(status.TotalAssigned) / float64(status.TotalDesignated)
			if fulfillment > 1.0 {
				fulfillment = 1.0
			}
			totalFulfillment += fulfillment
			daysWithQuota++
		}
	}

	averageFulfillment := 0.0
	if daysWithQuota > 0 {
		averageFulfillment = totalFulfillment / float64(daysWithQuota)
	}

	return &MonthlyOverview{
		BranchID:          branchID,
		BranchName:        branch.Name,
		BranchCode:        branch.Code,
		Year:              year,
		Month:             month,
		DayStatuses:       dayStatuses,
		AverageFulfillment: averageFulfillment,
	}, nil
}

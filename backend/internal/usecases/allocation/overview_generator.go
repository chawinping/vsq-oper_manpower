package allocation

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// cacheEntry represents a cached overview entry
type cacheEntry struct {
	data      *DayOverview
	expiresAt time.Time
}

// OverviewGenerator generates overview data for Area Managers
type OverviewGenerator struct {
	repos            *RepositoriesWrapper
	quotaCalculator *QuotaCalculator
	cache            sync.Map // map[string]*cacheEntry, key format: "date:branchIDs"
	cacheTTL         time.Duration
}

// NewOverviewGenerator creates a new overview generator
func NewOverviewGenerator(repos *RepositoriesWrapper, quotaCalculator *QuotaCalculator) *OverviewGenerator {
	return &OverviewGenerator{
		repos:           repos,
		quotaCalculator: quotaCalculator,
		cacheTTL:        5 * time.Minute, // Cache for 5 minutes
	}
}

// generateCacheKey generates a cache key from date and branch IDs
func (g *OverviewGenerator) generateCacheKey(date time.Time, branchIDs []uuid.UUID) string {
	dateStr := date.Format("2006-01-02")
	if len(branchIDs) == 0 {
		return fmt.Sprintf("overview:%s:all", dateStr)
	}
	// Create a deterministic key from sorted branch IDs
	idsStr := ""
	for _, id := range branchIDs {
		idsStr += id.String() + ","
	}
	return fmt.Sprintf("overview:%s:%s", dateStr, idsStr)
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

// GenerateDayOverview generates overview for specified branches on a specific day
// If branchIDs is nil or empty, calculates for all branches
func (g *OverviewGenerator) GenerateDayOverview(date time.Time, branchIDs []uuid.UUID) (*DayOverview, error) {
	// Check cache first
	cacheKey := g.generateCacheKey(date, branchIDs)
	if cached, found := g.cache.Load(cacheKey); found {
		entry := cached.(*cacheEntry)
		if time.Now().Before(entry.expiresAt) {
			return entry.data, nil
		}
		// Cache expired, remove it
		g.cache.Delete(cacheKey)
	}

	// If no branch IDs provided, get all branches
	if len(branchIDs) == 0 {
		branches, err := g.repos.Branch.List()
		if err != nil {
			return nil, fmt.Errorf("failed to get branches: %w", err)
		}
		branchIDs = make([]uuid.UUID, 0, len(branches))
		for _, branch := range branches {
			branchIDs = append(branchIDs, branch.ID)
		}
	}

	// Calculate quota status for specified branches
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

	overview := &DayOverview{
		Date:              date,
		BranchStatuses:    branchStatuses,
		TotalBranches:     len(branchStatuses),
		BranchesWithShortage: branchesWithShortage,
	}

	// Cache the result
	g.cache.Store(cacheKey, &cacheEntry{
		data:      overview,
		expiresAt: time.Now().Add(g.cacheTTL),
	})

	return overview, nil
}

// InvalidateCache invalidates the cache for a specific date
func (g *OverviewGenerator) InvalidateCache(date time.Time) {
	dateStr := date.Format("2006-01-02")
	// Remove all cache entries for this date
	g.cache.Range(func(key, value interface{}) bool {
		keyStr := key.(string)
		if len(keyStr) > len(dateStr) && keyStr[:len(dateStr)+9] == fmt.Sprintf("overview:%s:", dateStr) {
			g.cache.Delete(key)
		}
		return true
	})
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

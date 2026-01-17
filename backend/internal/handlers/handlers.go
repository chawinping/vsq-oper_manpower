package handlers

import (
	"vsq-oper-manpower/backend/internal/config"
	"vsq-oper-manpower/backend/internal/repositories/postgres"
	"vsq-oper-manpower/backend/internal/usecases/allocation"
)

type Handlers struct {
	Auth                *AuthHandler
	User                *UserHandler
	Staff               *StaffHandler
	Position            *PositionHandler
	Branch              *BranchHandler
	Schedule            *ScheduleHandler
	Rotation            *RotationHandler
	EffectiveBranch     *EffectiveBranchHandler
	AreaOfOperation     *AreaOfOperationHandler
	Zone                *ZoneHandler
	Settings            *SettingsHandler
	Dashboard           *DashboardHandler
	Version             *VersionHandler
	Doctor              *DoctorHandler
	Quota               *QuotaHandler
	Overview            *OverviewHandler
	AllocationCriteria      *AllocationCriteriaHandler
	BranchConfig            *BranchConfigHandler
	RevenueLevelTier         *RevenueLevelTierHandler
	StaffRequirementScenario *StaffRequirementScenarioHandler
	Report                  *ReportHandler
}

func NewHandlers(repos *postgres.Repositories, cfg *config.Config) *Handlers {
	// Initialize use cases
	// Create a wrapper that implements the interfaces needed by use cases
	reposWrapper := &allocation.RepositoriesWrapper{
		User:                repos.User,
		Role:                repos.Role,
		Staff:               repos.Staff,
		Position:            repos.Position,
		Branch:              repos.Branch,
		EffectiveBranch:     repos.EffectiveBranch,
		Revenue:             repos.Revenue,
		Schedule:            repos.Schedule,
		Rotation:            repos.Rotation,
		Settings:            repos.Settings,
		AllocationRule:      repos.AllocationRule,
		AreaOfOperation:     repos.AreaOfOperation,
		AllocationCriteria:  repos.AllocationCriteria,
		PositionQuota:       repos.PositionQuota,
		Doctor:              repos.Doctor,
		DoctorPreference:    repos.DoctorPreference,
		DoctorAssignment:    repos.DoctorAssignment,
		DoctorOnOffDay:      repos.DoctorOnOffDay,
		AllocationSuggestion: repos.AllocationSuggestion,
	}

	criteriaEngine := allocation.NewCriteriaEngine(reposWrapper)
	quotaCalculator := allocation.NewQuotaCalculator(reposWrapper)
	overviewGenerator := allocation.NewOverviewGenerator(reposWrapper, quotaCalculator)
	suggestionEngine := allocation.NewSuggestionEngine(reposWrapper, criteriaEngine, quotaCalculator)

	return &Handlers{
		Auth:               NewAuthHandler(repos, cfg),
		User:               NewUserHandler(repos),
		Staff:              NewStaffHandler(repos),
		Position:           NewPositionHandler(repos),
		Branch:             NewBranchHandler(repos),
		Schedule:           NewScheduleHandler(repos),
		Rotation:           NewRotationHandler(repos, cfg, suggestionEngine),
		EffectiveBranch:    NewEffectiveBranchHandler(repos),
		AreaOfOperation:    NewAreaOfOperationHandler(repos),
		Zone:               NewZoneHandler(repos),
		Settings:           NewSettingsHandler(repos),
		Dashboard:          NewDashboardHandler(repos),
		Version:           NewVersionHandler(),
		Doctor:             NewDoctorHandler(repos),
		Quota:              NewQuotaHandler(repos, quotaCalculator),
		Overview:           NewOverviewHandler(repos, overviewGenerator),
		AllocationCriteria:      NewAllocationCriteriaHandler(repos),
		BranchConfig:            NewBranchConfigHandler(repos),
		RevenueLevelTier:        NewRevenueLevelTierHandler(repos),
		StaffRequirementScenario: NewStaffRequirementScenarioHandler(repos),
		Report:                  NewReportHandler(repos),
	}
}


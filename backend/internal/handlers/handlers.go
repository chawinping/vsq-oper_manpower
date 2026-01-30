package handlers

import (
	"database/sql"
	"vsq-oper-manpower/backend/internal/config"
	"vsq-oper-manpower/backend/internal/repositories/postgres"
	"vsq-oper-manpower/backend/internal/usecases/allocation"
)

type Handlers struct {
	Auth                     *AuthHandler
	User                     *UserHandler
	Staff                    *StaffHandler
	Position                 *PositionHandler
	Branch                   *BranchHandler
	Schedule                 *ScheduleHandler
	Rotation                 *RotationHandler
	EffectiveBranch          *EffectiveBranchHandler
	AreaOfOperation          *AreaOfOperationHandler
	Zone                     *ZoneHandler
	Settings                 *SettingsHandler
	Dashboard                *DashboardHandler
	Version                  *VersionHandler
	Doctor                   *DoctorHandler
	Quota                    *QuotaHandler
	Overview                 *OverviewHandler
	AllocationCriteria       *AllocationCriteriaHandler
	BranchConfig             *BranchConfigHandler
	RevenueLevelTier         *RevenueLevelTierHandler
	StaffRequirementScenario *StaffRequirementScenarioHandler
	Report                   *ReportHandler
	BranchType               *BranchTypeHandler
	StaffGroup               *StaffGroupHandler
	BranchTypeRequirement    *BranchTypeRequirementHandler
	BranchTypeConstraints    *BranchTypeConstraintsHandler
	SpecificPreference       *SpecificPreferenceHandler
	ClinicWidePreference     *ClinicWidePreferenceHandler
	TestData                 *TestDataHandler
	RotationStaffBranchPosition *RotationStaffBranchPositionHandler
}

func NewHandlers(repos *postgres.Repositories, cfg *config.Config, db *sql.DB) *Handlers {
	// Initialize use cases
	// Create a wrapper that implements the interfaces needed by use cases
	reposWrapper := &allocation.RepositoriesWrapper{
		User:                  repos.User,
		Role:                  repos.Role,
		Staff:                 repos.Staff,
		Position:              repos.Position,
		Branch:                repos.Branch,
		EffectiveBranch:       repos.EffectiveBranch,
		Revenue:               repos.Revenue,
		Schedule:              repos.Schedule,
		Rotation:              repos.Rotation,
		Settings:              repos.Settings,
		AllocationRule:        repos.AllocationRule,
		AreaOfOperation:       repos.AreaOfOperation,
		AllocationCriteria:    repos.AllocationCriteria,
		PositionQuota:         repos.PositionQuota,
		Doctor:                repos.Doctor,
		DoctorPreference:      repos.DoctorPreference,
		DoctorAssignment:      repos.DoctorAssignment,
		DoctorOnOffDay:        repos.DoctorOnOffDay,
		BranchType:            repos.BranchType,
		StaffGroup:            repos.StaffGroup,
		StaffGroupPosition:    repos.StaffGroupPosition,
		BranchTypeRequirement: repos.BranchTypeRequirement,
		RotationStaffBranchPosition: repos.RotationStaffBranchPosition,
	}

	quotaCalculator := allocation.NewQuotaCalculator(reposWrapper)
	overviewGenerator := allocation.NewOverviewGenerator(reposWrapper, quotaCalculator)
	multiCriteriaFilter := allocation.NewMultiCriteriaFilter(reposWrapper)

	return &Handlers{
		Auth:                     NewAuthHandler(repos, cfg),
		User:                     NewUserHandler(repos),
		Staff:                    NewStaffHandler(repos),
		Position:                 NewPositionHandler(repos, db),
		Branch:                   NewBranchHandler(repos),
		Schedule:                 NewScheduleHandler(repos),
		Rotation:                 NewRotationHandler(repos, cfg, multiCriteriaFilter),
		EffectiveBranch:          NewEffectiveBranchHandler(repos),
		AreaOfOperation:          NewAreaOfOperationHandler(repos),
		Zone:                     NewZoneHandler(repos),
		Settings:                 NewSettingsHandler(repos),
		Dashboard:                NewDashboardHandler(repos),
		Version:                  NewVersionHandler(),
		Doctor:                   NewDoctorHandler(repos),
		Quota:                    NewQuotaHandler(repos, quotaCalculator),
		Overview:                 NewOverviewHandler(repos, overviewGenerator),
		AllocationCriteria:       NewAllocationCriteriaHandler(repos),
		BranchConfig:             NewBranchConfigHandler(repos),
		RevenueLevelTier:         NewRevenueLevelTierHandler(repos),
		StaffRequirementScenario: NewStaffRequirementScenarioHandler(repos),
		Report:                   NewReportHandler(repos),
		BranchType:               NewBranchTypeHandler(repos, db),
		StaffGroup:               NewStaffGroupHandler(repos, db),
		BranchTypeRequirement:    NewBranchTypeRequirementHandler(repos, db),
		BranchTypeConstraints:    NewBranchTypeConstraintsHandler(repos, db),
		SpecificPreference:       NewSpecificPreferenceHandler(repos),
		ClinicWidePreference:     NewClinicWidePreferenceHandler(repos),
		TestData:                 NewTestDataHandler(repos),
		RotationStaffBranchPosition: NewRotationStaffBranchPositionHandler(repos, db),
	}
}

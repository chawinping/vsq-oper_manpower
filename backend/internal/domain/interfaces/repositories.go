package interfaces

import (
	"time"

	"github.com/google/uuid"
	"vsq-oper-manpower/backend/internal/domain/models"
)

type UserRepository interface {
	Create(user *models.User) error
	GetByID(id uuid.UUID) (*models.User, error)
	GetByUsername(username string) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	Update(user *models.User) error
	Delete(id uuid.UUID) error
	List() ([]*models.User, error)
}

type RoleRepository interface {
	GetByID(id uuid.UUID) (*models.Role, error)
	GetByName(name string) (*models.Role, error)
	List() ([]*models.Role, error)
}

type StaffRepository interface {
	Create(staff *models.Staff) error
	GetByID(id uuid.UUID) (*models.Staff, error)
	Update(staff *models.Staff) error
	Delete(id uuid.UUID) error
	List(filters StaffFilters) ([]*models.Staff, error)
	GetByBranchID(branchID uuid.UUID) ([]*models.Staff, error)
	GetRotationStaff() ([]*models.Staff, error)
	// Zone and Branch management for rotation staff
	GetBranches(staffID uuid.UUID) ([]*models.Branch, error)
	BulkUpdateBranches(staffID uuid.UUID, branchIDs []uuid.UUID) error
}

type StaffFilters struct {
	StaffType          *models.StaffType
	BranchID           *uuid.UUID
	PositionID         *uuid.UUID
	AreaOfOperationID  *uuid.UUID
}

type PositionRepository interface {
	Create(position *models.Position) error
	GetByID(id uuid.UUID) (*models.Position, error)
	Update(position *models.Position) error
	Delete(id uuid.UUID) error
	List() ([]*models.Position, error)
	HasAssociatedStaff(id uuid.UUID) (bool, error)
}

type BranchRepository interface {
	Create(branch *models.Branch) error
	GetByID(id uuid.UUID) (*models.Branch, error)
	Update(branch *models.Branch) error
	Delete(id uuid.UUID) error
	List() ([]*models.Branch, error)
	GetByAreaManagerID(areaManagerID uuid.UUID) ([]*models.Branch, error)
}

type EffectiveBranchRepository interface {
	Create(eb *models.EffectiveBranch) error
	Update(eb *models.EffectiveBranch) error
	GetByID(id uuid.UUID) (*models.EffectiveBranch, error)
	GetByRotationStaffID(rotationStaffID uuid.UUID) ([]*models.EffectiveBranch, error)
	GetByBranchID(branchID uuid.UUID) ([]*models.EffectiveBranch, error)
	Delete(id uuid.UUID) error
	DeleteByRotationStaffID(rotationStaffID uuid.UUID) error
}

type RevenueRepository interface {
	Create(revenue *models.RevenueData) error
	GetByBranchID(branchID uuid.UUID, startDate, endDate time.Time) ([]*models.RevenueData, error)
	GetByDate(date time.Time) ([]*models.RevenueData, error)
	Update(revenue *models.RevenueData) error
	BulkCreateOrUpdate(revenues []*models.RevenueData) error
}

type ScheduleRepository interface {
	Create(schedule *models.StaffSchedule) error
	GetByBranchID(branchID uuid.UUID, startDate, endDate time.Time) ([]*models.StaffSchedule, error)
	GetByStaffID(staffID uuid.UUID, startDate, endDate time.Time) ([]*models.StaffSchedule, error)
	Update(schedule *models.StaffSchedule) error
	Delete(id uuid.UUID) error
	DeleteByStaffID(staffID uuid.UUID) error
	GetMonthlyView(branchID uuid.UUID, year int, month int) ([]*models.StaffSchedule, error)
}

type RotationRepository interface {
	Create(assignment *models.RotationAssignment) error
	GetByDate(date time.Time) ([]*models.RotationAssignment, error)
	GetByBranchID(branchID uuid.UUID, startDate, endDate time.Time) ([]*models.RotationAssignment, error)
	GetByRotationStaffID(rotationStaffID uuid.UUID, startDate, endDate time.Time) ([]*models.RotationAssignment, error)
	Delete(id uuid.UUID) error
	DeleteByRotationStaffID(rotationStaffID uuid.UUID) error
	GetAssignments(filters RotationFilters) ([]*models.RotationAssignment, error)
}

type RotationStaffScheduleRepository interface {
	Create(schedule *models.RotationStaffSchedule) error
	GetByID(id uuid.UUID) (*models.RotationStaffSchedule, error)
	GetByRotationStaffID(rotationStaffID uuid.UUID, startDate, endDate time.Time) ([]*models.RotationStaffSchedule, error)
	GetByDate(date time.Time) ([]*models.RotationStaffSchedule, error)
	GetByDateRange(startDate, endDate time.Time) ([]*models.RotationStaffSchedule, error)
	GetByRotationStaffIDAndDate(rotationStaffID uuid.UUID, date time.Time) (*models.RotationStaffSchedule, error)
	Update(schedule *models.RotationStaffSchedule) error
	Delete(id uuid.UUID) error
	DeleteByRotationStaffID(rotationStaffID uuid.UUID) error
}

type RotationFilters struct {
	BranchID        *uuid.UUID
	RotationStaffID *uuid.UUID
	StartDate       *time.Time
	EndDate         *time.Time
	CoverageArea    *string
}

type SettingsRepository interface {
	GetAll() ([]*models.SystemSetting, error)
	GetByKey(key string) (*models.SystemSetting, error)
	Update(setting *models.SystemSetting) error
	Create(setting *models.SystemSetting) error
}

type AllocationRuleRepository interface {
	Create(rule *models.StaffAllocationRule) error
	GetByPositionID(positionID uuid.UUID) (*models.StaffAllocationRule, error)
	Update(rule *models.StaffAllocationRule) error
	List() ([]*models.StaffAllocationRule, error)
}

type AreaOfOperationRepository interface {
	Create(aoo *models.AreaOfOperation) error
	GetByID(id uuid.UUID) (*models.AreaOfOperation, error)
	GetByCode(code string) (*models.AreaOfOperation, error)
	Update(aoo *models.AreaOfOperation) error
	Delete(id uuid.UUID) error
	List(includeInactive bool) ([]*models.AreaOfOperation, error)
	// Zone and Branch management
	AddZone(areaOfOperationID, zoneID uuid.UUID) error
	RemoveZone(areaOfOperationID, zoneID uuid.UUID) error
	GetZones(areaOfOperationID uuid.UUID) ([]*models.Zone, error)
	AddBranch(areaOfOperationID, branchID uuid.UUID) error
	RemoveBranch(areaOfOperationID, branchID uuid.UUID) error
	GetBranches(areaOfOperationID uuid.UUID) ([]*models.Branch, error)
	GetAllBranches(areaOfOperationID uuid.UUID) ([]*models.Branch, error) // Includes branches from zones + individual branches
}

type ZoneRepository interface {
	Create(zone *models.Zone) error
	GetByID(id uuid.UUID) (*models.Zone, error)
	GetByCode(code string) (*models.Zone, error)
	Update(zone *models.Zone) error
	Delete(id uuid.UUID) error
	List(includeInactive bool) ([]*models.Zone, error)
	// Branch management
	AddBranch(zoneID, branchID uuid.UUID) error
	RemoveBranch(zoneID, branchID uuid.UUID) error
	GetBranches(zoneID uuid.UUID) ([]*models.Branch, error)
	BulkUpdateBranches(zoneID uuid.UUID, branchIDs []uuid.UUID) error
}

type AllocationCriteriaRepository interface {
	Create(criteria *models.AllocationCriteria) error
	GetByID(id uuid.UUID) (*models.AllocationCriteria, error)
	Update(criteria *models.AllocationCriteria) error
	Delete(id uuid.UUID) error
	List(filters AllocationCriteriaFilters) ([]*models.AllocationCriteria, error)
	GetByPillar(pillar models.CriteriaPillar) ([]*models.AllocationCriteria, error)
}

type AllocationCriteriaFilters struct {
	Pillar   *models.CriteriaPillar
	Type     *models.CriteriaType
	IsActive *bool
}

type PositionQuotaRepository interface {
	Create(quota *models.PositionQuota) error
	GetByID(id uuid.UUID) (*models.PositionQuota, error)
	GetByBranchID(branchID uuid.UUID) ([]*models.PositionQuota, error)
	GetByBranchAndPosition(branchID, positionID uuid.UUID) (*models.PositionQuota, error)
	Update(quota *models.PositionQuota) error
	Delete(id uuid.UUID) error
	List(filters PositionQuotaFilters) ([]*models.PositionQuota, error)
}

type PositionQuotaFilters struct {
	BranchID   *uuid.UUID
	PositionID *uuid.UUID
	IsActive   *bool
}

type DoctorRepository interface {
	Create(doctor *models.Doctor) error
	GetByID(id uuid.UUID) (*models.Doctor, error)
	GetByCode(code string) (*models.Doctor, error)
	Update(doctor *models.Doctor) error
	Delete(id uuid.UUID) error
	List() ([]*models.Doctor, error)
}

type DoctorPreferenceRepository interface {
	Create(preference *models.DoctorPreference) error
	GetByID(id uuid.UUID) (*models.DoctorPreference, error)
	GetByDoctorID(doctorID uuid.UUID) ([]*models.DoctorPreference, error)
	GetByDoctorAndBranch(doctorID uuid.UUID, branchID uuid.UUID) ([]*models.DoctorPreference, error)
	GetActiveByDoctorID(doctorID uuid.UUID) ([]*models.DoctorPreference, error)
	Update(preference *models.DoctorPreference) error
	Delete(id uuid.UUID) error
	DeleteByDoctorID(doctorID uuid.UUID) error
}

type DoctorAssignmentRepository interface {
	Create(assignment *models.DoctorAssignment) error
	GetByID(id uuid.UUID) (*models.DoctorAssignment, error)
	GetByBranchID(branchID uuid.UUID, startDate, endDate time.Time) ([]*models.DoctorAssignment, error)
	GetByDate(date time.Time) ([]*models.DoctorAssignment, error)
	GetByDoctorID(doctorID uuid.UUID, startDate, endDate time.Time) ([]*models.DoctorAssignment, error)
	GetMonthlySchedule(doctorID uuid.UUID, year int, month int) ([]*models.DoctorAssignment, error)
	Update(assignment *models.DoctorAssignment) error
	Delete(id uuid.UUID) error
	DeleteByDoctorBranchDate(doctorID uuid.UUID, branchID uuid.UUID, date time.Time) error
	GetDoctorCountByBranch(branchID uuid.UUID, date time.Time) (int, error)
	GetDoctorsByBranchAndDate(branchID uuid.UUID, date time.Time) ([]*models.DoctorAssignment, error)
}

type DoctorOnOffDayRepository interface {
	Create(day *models.DoctorOnOffDay) error
	GetByID(id uuid.UUID) (*models.DoctorOnOffDay, error)
	GetByBranchID(branchID uuid.UUID, startDate, endDate time.Time) ([]*models.DoctorOnOffDay, error)
	GetByDate(date time.Time) ([]*models.DoctorOnOffDay, error)
	Update(day *models.DoctorOnOffDay) error
	Delete(id uuid.UUID) error
	GetByBranchAndDate(branchID uuid.UUID, date time.Time) (*models.DoctorOnOffDay, error)
}

type DoctorDefaultScheduleRepository interface {
	Create(schedule *models.DoctorDefaultSchedule) error
	GetByID(id uuid.UUID) (*models.DoctorDefaultSchedule, error)
	GetByDoctorID(doctorID uuid.UUID) ([]*models.DoctorDefaultSchedule, error)
	GetByDoctorAndDayOfWeek(doctorID uuid.UUID, dayOfWeek int) (*models.DoctorDefaultSchedule, error)
	Update(schedule *models.DoctorDefaultSchedule) error
	Delete(id uuid.UUID) error
	DeleteByDoctorID(doctorID uuid.UUID) error
	Upsert(schedule *models.DoctorDefaultSchedule) error
}

type DoctorWeeklyOffDayRepository interface {
	Create(offDay *models.DoctorWeeklyOffDay) error
	GetByID(id uuid.UUID) (*models.DoctorWeeklyOffDay, error)
	GetByDoctorID(doctorID uuid.UUID) ([]*models.DoctorWeeklyOffDay, error)
	GetByDoctorAndDayOfWeek(doctorID uuid.UUID, dayOfWeek int) (*models.DoctorWeeklyOffDay, error)
	Delete(id uuid.UUID) error
	DeleteByDoctorID(doctorID uuid.UUID) error
	DeleteByDoctorAndDayOfWeek(doctorID uuid.UUID, dayOfWeek int) error
}

type DoctorScheduleOverrideRepository interface {
	Create(override *models.DoctorScheduleOverride) error
	GetByID(id uuid.UUID) (*models.DoctorScheduleOverride, error)
	GetByDoctorID(doctorID uuid.UUID, startDate, endDate time.Time) ([]*models.DoctorScheduleOverride, error)
	GetByDoctorAndDate(doctorID uuid.UUID, date time.Time) (*models.DoctorScheduleOverride, error)
	Update(override *models.DoctorScheduleOverride) error
	Delete(id uuid.UUID) error
	DeleteByDoctorID(doctorID uuid.UUID) error
}

type AllocationSuggestionRepository interface {
	Create(suggestion *models.AllocationSuggestion) error
	GetByID(id uuid.UUID) (*models.AllocationSuggestion, error)
	GetByBranchID(branchID uuid.UUID, startDate, endDate time.Time) ([]*models.AllocationSuggestion, error)
	GetByStatus(status models.SuggestionStatus) ([]*models.AllocationSuggestion, error)
	Update(suggestion *models.AllocationSuggestion) error
	Delete(id uuid.UUID) error
	BulkCreate(suggestions []*models.AllocationSuggestion) error
	GetPendingByBranchID(branchID uuid.UUID, startDate, endDate time.Time) ([]*models.AllocationSuggestion, error)
}

type BranchWeeklyRevenueRepository interface {
	Create(revenue *models.BranchWeeklyRevenue) error
	Update(revenue *models.BranchWeeklyRevenue) error
	GetByID(id uuid.UUID) (*models.BranchWeeklyRevenue, error)
	GetByBranchID(branchID uuid.UUID) ([]*models.BranchWeeklyRevenue, error)
	GetByBranchIDAndDayOfWeek(branchID uuid.UUID, dayOfWeek int) (*models.BranchWeeklyRevenue, error)
	Delete(id uuid.UUID) error
	BulkUpsert(revenues []*models.BranchWeeklyRevenue) error
}

type BranchConstraintsRepository interface {
	Create(constraint *models.BranchConstraints) error
	Update(constraint *models.BranchConstraints) error
	GetByID(id uuid.UUID) (*models.BranchConstraints, error)
	GetByBranchID(branchID uuid.UUID) ([]*models.BranchConstraints, error)
	GetByBranchIDAndDayOfWeek(branchID uuid.UUID, dayOfWeek int) (*models.BranchConstraints, error)
	Delete(id uuid.UUID) error
	BulkUpsert(constraints []*models.BranchConstraints) error
}

type RevenueLevelTierRepository interface {
	Create(tier *models.RevenueLevelTier) error
	GetByID(id uuid.UUID) (*models.RevenueLevelTier, error)
	GetByLevelNumber(levelNumber int) (*models.RevenueLevelTier, error)
	Update(tier *models.RevenueLevelTier) error
	Delete(id uuid.UUID) error
	List() ([]*models.RevenueLevelTier, error)
	GetTierForRevenue(revenue float64) (*models.RevenueLevelTier, error)
}

type StaffRequirementScenarioRepository interface {
	Create(scenario *models.StaffRequirementScenario) error
	GetByID(id uuid.UUID) (*models.StaffRequirementScenario, error)
	Update(scenario *models.StaffRequirementScenario) error
	Delete(id uuid.UUID) error
	List(includeInactive bool) ([]*models.StaffRequirementScenario, error)
	GetActiveOrderedByPriority() ([]*models.StaffRequirementScenario, error)
	GetDefault() (*models.StaffRequirementScenario, error)
}

type ScenarioPositionRequirementRepository interface {
	Create(requirement *models.ScenarioPositionRequirement) error
	GetByID(id uuid.UUID) (*models.ScenarioPositionRequirement, error)
	GetByScenarioID(scenarioID uuid.UUID) ([]*models.ScenarioPositionRequirement, error)
	GetByScenarioAndPosition(scenarioID, positionID uuid.UUID) (*models.ScenarioPositionRequirement, error)
	Update(requirement *models.ScenarioPositionRequirement) error
	Delete(id uuid.UUID) error
	DeleteByScenarioID(scenarioID uuid.UUID) error
	BulkUpsert(requirements []*models.ScenarioPositionRequirement) error
}


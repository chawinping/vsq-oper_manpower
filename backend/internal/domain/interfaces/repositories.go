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
}


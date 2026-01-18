package doctor

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"vsq-oper-manpower/backend/internal/domain/interfaces"
	"vsq-oper-manpower/backend/internal/domain/models"
)

// DoctorScheduleCalculator calculates doctor assignments from default schedules and overrides
type DoctorScheduleCalculator struct {
	defaultScheduleRepo interfaces.DoctorDefaultScheduleRepository
	weeklyOffDayRepo    interfaces.DoctorWeeklyOffDayRepository
	overrideRepo        interfaces.DoctorScheduleOverrideRepository
	doctorRepo          interfaces.DoctorRepository
}

// NewDoctorScheduleCalculator creates a new schedule calculator
func NewDoctorScheduleCalculator(
	defaultScheduleRepo interfaces.DoctorDefaultScheduleRepository,
	weeklyOffDayRepo interfaces.DoctorWeeklyOffDayRepository,
	overrideRepo interfaces.DoctorScheduleOverrideRepository,
	doctorRepo interfaces.DoctorRepository,
) *DoctorScheduleCalculator {
	return &DoctorScheduleCalculator{
		defaultScheduleRepo: defaultScheduleRepo,
		weeklyOffDayRepo:    weeklyOffDayRepo,
		overrideRepo:        overrideRepo,
		doctorRepo:          doctorRepo,
	}
}

// CalculatedAssignment represents a calculated doctor assignment
type CalculatedAssignment struct {
	DoctorID        uuid.UUID
	BranchID        uuid.UUID
	Date            time.Time
	Source          string // "override" or "default"
	ExpectedRevenue float64 // Will be 0 for calculated assignments (can be set separately)
}

// CalculateAssignmentForDate calculates doctor assignment for a specific doctor and date
// Priority: Override > Weekly Off Day > Default Schedule
func (c *DoctorScheduleCalculator) CalculateAssignmentForDate(doctorID uuid.UUID, date time.Time) (*CalculatedAssignment, error) {
	// Get day of week (0=Sunday, 1=Monday, ..., 6=Saturday)
	dayOfWeek := int(date.Weekday())

	// Priority 1: Check for override
	override, err := c.overrideRepo.GetByDoctorAndDate(doctorID, date)
	if err != nil {
		return nil, fmt.Errorf("failed to get override: %w", err)
	}

	if override != nil {
		if override.Type == "off" {
			// Doctor is off - no assignment
			return nil, nil
		}
		// Working override
		if override.BranchID == nil {
			return nil, fmt.Errorf("override type is 'working' but branch_id is null")
		}
		return &CalculatedAssignment{
			DoctorID:        doctorID,
			BranchID:        *override.BranchID,
			Date:            date,
			Source:          "override",
			ExpectedRevenue: 0, // Can be set separately if needed
		}, nil
	}

	// Priority 2: Check for weekly off day
	weeklyOffDay, err := c.weeklyOffDayRepo.GetByDoctorAndDayOfWeek(doctorID, dayOfWeek)
	if err != nil {
		return nil, fmt.Errorf("failed to get weekly off day: %w", err)
	}

	if weeklyOffDay != nil {
		// Doctor has weekly off day - no assignment
		return nil, nil
	}

	// Priority 3: Check default schedule
	defaultSchedule, err := c.defaultScheduleRepo.GetByDoctorAndDayOfWeek(doctorID, dayOfWeek)
	if err != nil {
		return nil, fmt.Errorf("failed to get default schedule: %w", err)
	}

	if defaultSchedule != nil {
		return &CalculatedAssignment{
			DoctorID:        doctorID,
			BranchID:        defaultSchedule.BranchID,
			Date:            date,
			Source:          "default",
			ExpectedRevenue: 0, // Can be set separately if needed
		}, nil
	}

	// No schedule - doctor is off
	return nil, nil
}

// CalculateAssignmentsForDateRange calculates assignments for all doctors in a date range
// Returns a map: branchID -> date -> []doctorID
func (c *DoctorScheduleCalculator) CalculateAssignmentsForDateRange(startDate, endDate time.Time) (map[uuid.UUID]map[time.Time][]uuid.UUID, error) {
	// Get all doctors
	doctors, err := c.doctorRepo.List()
	if err != nil {
		return nil, fmt.Errorf("failed to get doctors: %w", err)
	}

	// Load all default schedules (cacheable)
	allDefaultSchedules := make(map[uuid.UUID][]*models.DoctorDefaultSchedule)
	for _, doctor := range doctors {
		schedules, err := c.defaultScheduleRepo.GetByDoctorID(doctor.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get default schedules for doctor %s: %w", doctor.ID, err)
		}
		allDefaultSchedules[doctor.ID] = schedules
	}

	// Load all weekly off days (cacheable)
	allWeeklyOffDays := make(map[uuid.UUID]map[int]bool) // doctorID -> dayOfWeek -> true
	for _, doctor := range doctors {
		offDays, err := c.weeklyOffDayRepo.GetByDoctorID(doctor.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get weekly off days for doctor %s: %w", doctor.ID, err)
		}
		offDayMap := make(map[int]bool)
		for _, offDay := range offDays {
			offDayMap[offDay.DayOfWeek] = true
		}
		allWeeklyOffDays[doctor.ID] = offDayMap
	}

	// Load all overrides for date range
	allOverrides := make(map[uuid.UUID]map[time.Time]*models.DoctorScheduleOverride) // doctorID -> date -> override
	for _, doctor := range doctors {
		overrides, err := c.overrideRepo.GetByDoctorID(doctor.ID, startDate, endDate)
		if err != nil {
			return nil, fmt.Errorf("failed to get overrides for doctor %s: %w", doctor.ID, err)
		}
		overrideMap := make(map[time.Time]*models.DoctorScheduleOverride)
		for _, override := range overrides {
			overrideMap[override.Date] = override
		}
		allOverrides[doctor.ID] = overrideMap
	}

	// Calculate assignments: branchID -> date -> []doctorID
	result := make(map[uuid.UUID]map[time.Time][]uuid.UUID)

	// Iterate through all dates
	currentDate := startDate
	for !currentDate.After(endDate) {
		dayOfWeek := int(currentDate.Weekday())

		// Process each doctor
		for _, doctor := range doctors {
			var branchID *uuid.UUID

			// Priority 1: Check override
			if overrideMap, ok := allOverrides[doctor.ID]; ok {
				if override, ok := overrideMap[currentDate]; ok {
					if override.Type == "off" {
						continue // Doctor is off
					}
					if override.BranchID != nil {
						branchID = override.BranchID
					}
				}
			}

			// Priority 2: Check weekly off day
			if branchID == nil {
				if offDayMap, ok := allWeeklyOffDays[doctor.ID]; ok {
					if offDayMap[dayOfWeek] {
						continue // Doctor has weekly off day
					}
				}
			}

			// Priority 3: Check default schedule
			if branchID == nil {
				if schedules, ok := allDefaultSchedules[doctor.ID]; ok {
					for _, schedule := range schedules {
						if schedule.DayOfWeek == dayOfWeek {
							branchID = &schedule.BranchID
							break
						}
					}
				}
			}

			// Add assignment if branch is determined
			if branchID != nil {
				if result[*branchID] == nil {
					result[*branchID] = make(map[time.Time][]uuid.UUID)
				}
				if result[*branchID][currentDate] == nil {
					result[*branchID][currentDate] = []uuid.UUID{}
				}
				result[*branchID][currentDate] = append(result[*branchID][currentDate], doctor.ID)
			}
		}

		// Move to next day
		currentDate = currentDate.AddDate(0, 0, 1)
	}

	return result, nil
}

// GetDoctorsByBranchAndDate returns doctor IDs assigned to a branch on a specific date
func (c *DoctorScheduleCalculator) GetDoctorsByBranchAndDate(branchID uuid.UUID, date time.Time) ([]uuid.UUID, error) {
	// Get all doctors
	doctors, err := c.doctorRepo.List()
	if err != nil {
		return nil, fmt.Errorf("failed to get doctors: %w", err)
	}

	dayOfWeek := int(date.Weekday())
	var doctorIDs []uuid.UUID

	for _, doctor := range doctors {
		var assignedBranchID *uuid.UUID

		// Priority 1: Check override
		override, err := c.overrideRepo.GetByDoctorAndDate(doctor.ID, date)
		if err != nil {
			return nil, fmt.Errorf("failed to get override: %w", err)
		}
		if override != nil {
			if override.Type == "off" {
				continue // Doctor is off
			}
			assignedBranchID = override.BranchID
		}

		// Priority 2: Check weekly off day
		if assignedBranchID == nil {
			weeklyOffDay, err := c.weeklyOffDayRepo.GetByDoctorAndDayOfWeek(doctor.ID, dayOfWeek)
			if err != nil {
				return nil, fmt.Errorf("failed to get weekly off day: %w", err)
			}
			if weeklyOffDay != nil {
				continue // Doctor has weekly off day
			}
		}

		// Priority 3: Check default schedule
		if assignedBranchID == nil {
			defaultSchedule, err := c.defaultScheduleRepo.GetByDoctorAndDayOfWeek(doctor.ID, dayOfWeek)
			if err != nil {
				return nil, fmt.Errorf("failed to get default schedule: %w", err)
			}
			if defaultSchedule != nil {
				assignedBranchID = &defaultSchedule.BranchID
			}
		}

		// Check if assigned to target branch
		if assignedBranchID != nil && *assignedBranchID == branchID {
			doctorIDs = append(doctorIDs, doctor.ID)
		}
	}

	return doctorIDs, nil
}

// GetDoctorCountByBranch returns the count of doctors assigned to a branch on a specific date
func (c *DoctorScheduleCalculator) GetDoctorCountByBranch(branchID uuid.UUID, date time.Time) (int, error) {
	doctorIDs, err := c.GetDoctorsByBranchAndDate(branchID, date)
	if err != nil {
		return 0, err
	}
	return len(doctorIDs), nil
}

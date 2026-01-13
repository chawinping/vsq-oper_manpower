package models

import (
	"time"

	"github.com/google/uuid"
)

type ScheduleStatus string

const (
	ScheduleStatusWorking  ScheduleStatus = "working"
	ScheduleStatusOff      ScheduleStatus = "off"
	ScheduleStatusLeave    ScheduleStatus = "leave"
	ScheduleStatusSickLeave ScheduleStatus = "sick_leave"
)

type StaffSchedule struct {
	ID            uuid.UUID     `json:"id" db:"id"`
	StaffID       uuid.UUID     `json:"staff_id" db:"staff_id"`
	Staff         *Staff        `json:"staff,omitempty"`
	BranchID      uuid.UUID     `json:"branch_id" db:"branch_id"`
	Branch        *Branch       `json:"branch,omitempty"`
	Date          time.Time     `json:"date" db:"date"`
	ScheduleStatus ScheduleStatus `json:"schedule_status" db:"schedule_status"`
	IsWorkingDay  bool          `json:"is_working_day" db:"is_working_day"` // Deprecated: kept for backward compatibility
	CreatedBy     uuid.UUID     `json:"created_by" db:"created_by"`
	CreatedAt     time.Time     `json:"created_at" db:"created_at"`
}

type RotationAssignment struct {
	ID              uuid.UUID `json:"id" db:"id"`
	RotationStaffID uuid.UUID `json:"rotation_staff_id" db:"rotation_staff_id"`
	RotationStaff   *Staff    `json:"rotation_staff,omitempty"`
	BranchID        uuid.UUID `json:"branch_id" db:"branch_id"`
	Branch          *Branch   `json:"branch,omitempty"`
	Date            time.Time `json:"date" db:"date"`
	AssignmentLevel int      `json:"assignment_level" db:"assignment_level"` // 1 or 2
	AssignedBy      uuid.UUID `json:"assigned_by" db:"assigned_by"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}




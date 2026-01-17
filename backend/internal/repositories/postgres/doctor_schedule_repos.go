package postgres

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"vsq-oper-manpower/backend/internal/domain/interfaces"
	"vsq-oper-manpower/backend/internal/domain/models"
)

// DoctorDefaultScheduleRepository implementation
type doctorDefaultScheduleRepository struct {
	db *sql.DB
}

func NewDoctorDefaultScheduleRepository(db *sql.DB) interfaces.DoctorDefaultScheduleRepository {
	return &doctorDefaultScheduleRepository{db: db}
}

func (r *doctorDefaultScheduleRepository) Create(schedule *models.DoctorDefaultSchedule) error {
	if schedule.ID == uuid.Nil {
		schedule.ID = uuid.New()
	}
	schedule.CreatedAt = time.Now()
	schedule.UpdatedAt = time.Now()
	query := `INSERT INTO doctor_default_schedules (id, doctor_id, day_of_week, branch_id, created_by, created_at, updated_at) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING created_at, updated_at`
	return r.db.QueryRow(query, schedule.ID, schedule.DoctorID, schedule.DayOfWeek, schedule.BranchID, schedule.CreatedBy, schedule.CreatedAt, schedule.UpdatedAt).
		Scan(&schedule.CreatedAt, &schedule.UpdatedAt)
}

func (r *doctorDefaultScheduleRepository) GetByID(id uuid.UUID) (*models.DoctorDefaultSchedule, error) {
	query := `SELECT id, doctor_id, day_of_week, branch_id, created_by, created_at, updated_at 
	          FROM doctor_default_schedules WHERE id = $1`
	schedule := &models.DoctorDefaultSchedule{}
	err := r.db.QueryRow(query, id).Scan(
		&schedule.ID, &schedule.DoctorID, &schedule.DayOfWeek, &schedule.BranchID, &schedule.CreatedBy, &schedule.CreatedAt, &schedule.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return schedule, err
}

func (r *doctorDefaultScheduleRepository) GetByDoctorID(doctorID uuid.UUID) ([]*models.DoctorDefaultSchedule, error) {
	query := `SELECT id, doctor_id, day_of_week, branch_id, created_by, created_at, updated_at 
	          FROM doctor_default_schedules WHERE doctor_id = $1 ORDER BY day_of_week`
	rows, err := r.db.Query(query, doctorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []*models.DoctorDefaultSchedule
	for rows.Next() {
		schedule := &models.DoctorDefaultSchedule{}
		if err := rows.Scan(
			&schedule.ID, &schedule.DoctorID, &schedule.DayOfWeek, &schedule.BranchID, &schedule.CreatedBy, &schedule.CreatedAt, &schedule.UpdatedAt,
		); err != nil {
			return nil, err
		}
		schedules = append(schedules, schedule)
	}
	return schedules, rows.Err()
}

func (r *doctorDefaultScheduleRepository) GetByDoctorAndDayOfWeek(doctorID uuid.UUID, dayOfWeek int) (*models.DoctorDefaultSchedule, error) {
	query := `SELECT id, doctor_id, day_of_week, branch_id, created_by, created_at, updated_at 
	          FROM doctor_default_schedules WHERE doctor_id = $1 AND day_of_week = $2`
	schedule := &models.DoctorDefaultSchedule{}
	err := r.db.QueryRow(query, doctorID, dayOfWeek).Scan(
		&schedule.ID, &schedule.DoctorID, &schedule.DayOfWeek, &schedule.BranchID, &schedule.CreatedBy, &schedule.CreatedAt, &schedule.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return schedule, err
}

func (r *doctorDefaultScheduleRepository) Update(schedule *models.DoctorDefaultSchedule) error {
	schedule.UpdatedAt = time.Now()
	query := `UPDATE doctor_default_schedules 
	          SET branch_id = $1, updated_at = $2 
	          WHERE id = $3 RETURNING updated_at`
	return r.db.QueryRow(query, schedule.BranchID, schedule.UpdatedAt, schedule.ID).Scan(&schedule.UpdatedAt)
}

func (r *doctorDefaultScheduleRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM doctor_default_schedules WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *doctorDefaultScheduleRepository) DeleteByDoctorID(doctorID uuid.UUID) error {
	query := `DELETE FROM doctor_default_schedules WHERE doctor_id = $1`
	_, err := r.db.Exec(query, doctorID)
	return err
}

func (r *doctorDefaultScheduleRepository) Upsert(schedule *models.DoctorDefaultSchedule) error {
	if schedule.ID == uuid.Nil {
		schedule.ID = uuid.New()
	}
	schedule.CreatedAt = time.Now()
	schedule.UpdatedAt = time.Now()
	query := `INSERT INTO doctor_default_schedules (id, doctor_id, day_of_week, branch_id, created_by, created_at, updated_at) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7)
	          ON CONFLICT (doctor_id, day_of_week) 
	          DO UPDATE SET branch_id = EXCLUDED.branch_id, updated_at = EXCLUDED.updated_at
	          RETURNING created_at, updated_at`
	return r.db.QueryRow(query, schedule.ID, schedule.DoctorID, schedule.DayOfWeek, schedule.BranchID, schedule.CreatedBy, schedule.CreatedAt, schedule.UpdatedAt).
		Scan(&schedule.CreatedAt, &schedule.UpdatedAt)
}

// DoctorWeeklyOffDayRepository implementation
type doctorWeeklyOffDayRepository struct {
	db *sql.DB
}

func NewDoctorWeeklyOffDayRepository(db *sql.DB) interfaces.DoctorWeeklyOffDayRepository {
	return &doctorWeeklyOffDayRepository{db: db}
}

func (r *doctorWeeklyOffDayRepository) Create(offDay *models.DoctorWeeklyOffDay) error {
	if offDay.ID == uuid.Nil {
		offDay.ID = uuid.New()
	}
	offDay.CreatedAt = time.Now()
	offDay.UpdatedAt = time.Now()
	query := `INSERT INTO doctor_weekly_off_days (id, doctor_id, day_of_week, created_by, created_at, updated_at) 
	          VALUES ($1, $2, $3, $4, $5, $6) RETURNING created_at, updated_at`
	return r.db.QueryRow(query, offDay.ID, offDay.DoctorID, offDay.DayOfWeek, offDay.CreatedBy, offDay.CreatedAt, offDay.UpdatedAt).
		Scan(&offDay.CreatedAt, &offDay.UpdatedAt)
}

func (r *doctorWeeklyOffDayRepository) GetByID(id uuid.UUID) (*models.DoctorWeeklyOffDay, error) {
	query := `SELECT id, doctor_id, day_of_week, created_by, created_at, updated_at 
	          FROM doctor_weekly_off_days WHERE id = $1`
	offDay := &models.DoctorWeeklyOffDay{}
	err := r.db.QueryRow(query, id).Scan(
		&offDay.ID, &offDay.DoctorID, &offDay.DayOfWeek, &offDay.CreatedBy, &offDay.CreatedAt, &offDay.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return offDay, err
}

func (r *doctorWeeklyOffDayRepository) GetByDoctorID(doctorID uuid.UUID) ([]*models.DoctorWeeklyOffDay, error) {
	query := `SELECT id, doctor_id, day_of_week, created_by, created_at, updated_at 
	          FROM doctor_weekly_off_days WHERE doctor_id = $1 ORDER BY day_of_week`
	rows, err := r.db.Query(query, doctorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var offDays []*models.DoctorWeeklyOffDay
	for rows.Next() {
		offDay := &models.DoctorWeeklyOffDay{}
		if err := rows.Scan(
			&offDay.ID, &offDay.DoctorID, &offDay.DayOfWeek, &offDay.CreatedBy, &offDay.CreatedAt, &offDay.UpdatedAt,
		); err != nil {
			return nil, err
		}
		offDays = append(offDays, offDay)
	}
	return offDays, rows.Err()
}

func (r *doctorWeeklyOffDayRepository) GetByDoctorAndDayOfWeek(doctorID uuid.UUID, dayOfWeek int) (*models.DoctorWeeklyOffDay, error) {
	query := `SELECT id, doctor_id, day_of_week, created_by, created_at, updated_at 
	          FROM doctor_weekly_off_days WHERE doctor_id = $1 AND day_of_week = $2`
	offDay := &models.DoctorWeeklyOffDay{}
	err := r.db.QueryRow(query, doctorID, dayOfWeek).Scan(
		&offDay.ID, &offDay.DoctorID, &offDay.DayOfWeek, &offDay.CreatedBy, &offDay.CreatedAt, &offDay.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return offDay, err
}

func (r *doctorWeeklyOffDayRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM doctor_weekly_off_days WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *doctorWeeklyOffDayRepository) DeleteByDoctorID(doctorID uuid.UUID) error {
	query := `DELETE FROM doctor_weekly_off_days WHERE doctor_id = $1`
	_, err := r.db.Exec(query, doctorID)
	return err
}

func (r *doctorWeeklyOffDayRepository) DeleteByDoctorAndDayOfWeek(doctorID uuid.UUID, dayOfWeek int) error {
	query := `DELETE FROM doctor_weekly_off_days WHERE doctor_id = $1 AND day_of_week = $2`
	_, err := r.db.Exec(query, doctorID, dayOfWeek)
	return err
}

// DoctorScheduleOverrideRepository implementation
type doctorScheduleOverrideRepository struct {
	db *sql.DB
}

func NewDoctorScheduleOverrideRepository(db *sql.DB) interfaces.DoctorScheduleOverrideRepository {
	return &doctorScheduleOverrideRepository{db: db}
}

func (r *doctorScheduleOverrideRepository) Create(override *models.DoctorScheduleOverride) error {
	if override.ID == uuid.Nil {
		override.ID = uuid.New()
	}
	override.CreatedAt = time.Now()
	override.UpdatedAt = time.Now()
	query := `INSERT INTO doctor_schedule_overrides (id, doctor_id, date, type, branch_id, created_by, created_at, updated_at) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING created_at, updated_at`
	return r.db.QueryRow(query, override.ID, override.DoctorID, override.Date, override.Type, override.BranchID, override.CreatedBy, override.CreatedAt, override.UpdatedAt).
		Scan(&override.CreatedAt, &override.UpdatedAt)
}

func (r *doctorScheduleOverrideRepository) GetByID(id uuid.UUID) (*models.DoctorScheduleOverride, error) {
	query := `SELECT id, doctor_id, date, type, branch_id, created_by, created_at, updated_at 
	          FROM doctor_schedule_overrides WHERE id = $1`
	override := &models.DoctorScheduleOverride{}
	var branchID sql.NullString
	err := r.db.QueryRow(query, id).Scan(
		&override.ID, &override.DoctorID, &override.Date, &override.Type, &branchID, &override.CreatedBy, &override.CreatedAt, &override.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if branchID.Valid {
		bID, _ := uuid.Parse(branchID.String)
		override.BranchID = &bID
	}
	return override, nil
}

func (r *doctorScheduleOverrideRepository) GetByDoctorID(doctorID uuid.UUID, startDate, endDate time.Time) ([]*models.DoctorScheduleOverride, error) {
	query := `SELECT id, doctor_id, date, type, branch_id, created_by, created_at, updated_at 
	          FROM doctor_schedule_overrides WHERE doctor_id = $1 AND date >= $2 AND date <= $3 ORDER BY date`
	rows, err := r.db.Query(query, doctorID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var overrides []*models.DoctorScheduleOverride
	for rows.Next() {
		override := &models.DoctorScheduleOverride{}
		var branchID sql.NullString
		if err := rows.Scan(
			&override.ID, &override.DoctorID, &override.Date, &override.Type, &branchID, &override.CreatedBy, &override.CreatedAt, &override.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if branchID.Valid {
			bID, _ := uuid.Parse(branchID.String)
			override.BranchID = &bID
		}
		overrides = append(overrides, override)
	}
	return overrides, rows.Err()
}

func (r *doctorScheduleOverrideRepository) GetByDoctorAndDate(doctorID uuid.UUID, date time.Time) (*models.DoctorScheduleOverride, error) {
	query := `SELECT id, doctor_id, date, type, branch_id, created_by, created_at, updated_at 
	          FROM doctor_schedule_overrides WHERE doctor_id = $1 AND date = $2`
	override := &models.DoctorScheduleOverride{}
	var branchID sql.NullString
	err := r.db.QueryRow(query, doctorID, date).Scan(
		&override.ID, &override.DoctorID, &override.Date, &override.Type, &branchID, &override.CreatedBy, &override.CreatedAt, &override.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if branchID.Valid {
		bID, _ := uuid.Parse(branchID.String)
		override.BranchID = &bID
	}
	return override, nil
}

func (r *doctorScheduleOverrideRepository) Update(override *models.DoctorScheduleOverride) error {
	override.UpdatedAt = time.Now()
	query := `UPDATE doctor_schedule_overrides 
	          SET type = $1, branch_id = $2, updated_at = $3 
	          WHERE id = $4 RETURNING updated_at`
	return r.db.QueryRow(query, override.Type, override.BranchID, override.UpdatedAt, override.ID).Scan(&override.UpdatedAt)
}

func (r *doctorScheduleOverrideRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM doctor_schedule_overrides WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *doctorScheduleOverrideRepository) DeleteByDoctorID(doctorID uuid.UUID) error {
	query := `DELETE FROM doctor_schedule_overrides WHERE doctor_id = $1`
	_, err := r.db.Exec(query, doctorID)
	return err
}

package postgres

import (
	"database/sql"

	"github.com/google/uuid"
	"vsq-oper-manpower/backend/internal/domain/interfaces"
	"vsq-oper-manpower/backend/internal/domain/models"
)

type staffRequirementScenarioRepository struct {
	db *sql.DB
}

func NewStaffRequirementScenarioRepository(db *sql.DB) interfaces.StaffRequirementScenarioRepository {
	return &staffRequirementScenarioRepository{db: db}
}

func (r *staffRequirementScenarioRepository) Create(scenario *models.StaffRequirementScenario) error {
	query := `INSERT INTO staff_requirement_scenarios 
	          (id, scenario_name, description, revenue_level_tier_id, min_revenue, max_revenue, 
	           use_day_of_week_revenue, use_specific_date_revenue, doctor_count, min_doctor_count, 
	           day_of_week, is_default, is_active, priority)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	          RETURNING created_at, updated_at`
	return r.db.QueryRow(
		query,
		scenario.ID, scenario.ScenarioName, scenario.Description, scenario.RevenueLevelTierID,
		scenario.MinRevenue, scenario.MaxRevenue, scenario.UseDayOfWeekRevenue,
		scenario.UseSpecificDateRevenue, scenario.DoctorCount, scenario.MinDoctorCount,
		scenario.DayOfWeek, scenario.IsDefault, scenario.IsActive, scenario.Priority,
	).Scan(&scenario.CreatedAt, &scenario.UpdatedAt)
}

func (r *staffRequirementScenarioRepository) GetByID(id uuid.UUID) (*models.StaffRequirementScenario, error) {
	scenario := &models.StaffRequirementScenario{}
	query := `SELECT id, scenario_name, description, revenue_level_tier_id, min_revenue, max_revenue,
	          use_day_of_week_revenue, use_specific_date_revenue, doctor_count, min_doctor_count,
	          day_of_week, is_default, is_active, priority, created_at, updated_at
	          FROM staff_requirement_scenarios WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(
		&scenario.ID, &scenario.ScenarioName, &scenario.Description, &scenario.RevenueLevelTierID,
		&scenario.MinRevenue, &scenario.MaxRevenue, &scenario.UseDayOfWeekRevenue,
		&scenario.UseSpecificDateRevenue, &scenario.DoctorCount, &scenario.MinDoctorCount,
		&scenario.DayOfWeek, &scenario.IsDefault, &scenario.IsActive, &scenario.Priority,
		&scenario.CreatedAt, &scenario.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return scenario, nil
}

func (r *staffRequirementScenarioRepository) Update(scenario *models.StaffRequirementScenario) error {
	query := `UPDATE staff_requirement_scenarios
	          SET scenario_name = $1, description = $2, revenue_level_tier_id = $3, min_revenue = $4, max_revenue = $5,
	              use_day_of_week_revenue = $6, use_specific_date_revenue = $7, doctor_count = $8, min_doctor_count = $9,
	              day_of_week = $10, is_default = $11, is_active = $12, priority = $13, updated_at = CURRENT_TIMESTAMP
	          WHERE id = $14 RETURNING updated_at`
	return r.db.QueryRow(
		query,
		scenario.ScenarioName, scenario.Description, scenario.RevenueLevelTierID,
		scenario.MinRevenue, scenario.MaxRevenue, scenario.UseDayOfWeekRevenue,
		scenario.UseSpecificDateRevenue, scenario.DoctorCount, scenario.MinDoctorCount,
		scenario.DayOfWeek, scenario.IsDefault, scenario.IsActive, scenario.Priority, scenario.ID,
	).Scan(&scenario.UpdatedAt)
}

func (r *staffRequirementScenarioRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM staff_requirement_scenarios WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *staffRequirementScenarioRepository) List(includeInactive bool) ([]*models.StaffRequirementScenario, error) {
	query := `SELECT id, scenario_name, description, revenue_level_tier_id, min_revenue, max_revenue,
	          use_day_of_week_revenue, use_specific_date_revenue, doctor_count, min_doctor_count,
	          day_of_week, is_default, is_active, priority, created_at, updated_at
	          FROM staff_requirement_scenarios`
	if !includeInactive {
		query += " WHERE is_active = true"
	}
	query += " ORDER BY priority DESC, scenario_name"
	
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	scenarios := []*models.StaffRequirementScenario{}
	for rows.Next() {
		scenario := &models.StaffRequirementScenario{}
		if err := rows.Scan(
			&scenario.ID, &scenario.ScenarioName, &scenario.Description, &scenario.RevenueLevelTierID,
			&scenario.MinRevenue, &scenario.MaxRevenue, &scenario.UseDayOfWeekRevenue,
			&scenario.UseSpecificDateRevenue, &scenario.DoctorCount, &scenario.MinDoctorCount,
			&scenario.DayOfWeek, &scenario.IsDefault, &scenario.IsActive, &scenario.Priority,
			&scenario.CreatedAt, &scenario.UpdatedAt,
		); err != nil {
			return nil, err
		}
		scenarios = append(scenarios, scenario)
	}
	return scenarios, rows.Err()
}

func (r *staffRequirementScenarioRepository) GetActiveOrderedByPriority() ([]*models.StaffRequirementScenario, error) {
	query := `SELECT id, scenario_name, description, revenue_level_tier_id, min_revenue, max_revenue,
	          use_day_of_week_revenue, use_specific_date_revenue, doctor_count, min_doctor_count,
	          day_of_week, is_default, is_active, priority, created_at, updated_at
	          FROM staff_requirement_scenarios
	          WHERE is_active = true
	          ORDER BY priority DESC, scenario_name`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	scenarios := []*models.StaffRequirementScenario{}
	for rows.Next() {
		scenario := &models.StaffRequirementScenario{}
		if err := rows.Scan(
			&scenario.ID, &scenario.ScenarioName, &scenario.Description, &scenario.RevenueLevelTierID,
			&scenario.MinRevenue, &scenario.MaxRevenue, &scenario.UseDayOfWeekRevenue,
			&scenario.UseSpecificDateRevenue, &scenario.DoctorCount, &scenario.MinDoctorCount,
			&scenario.DayOfWeek, &scenario.IsDefault, &scenario.IsActive, &scenario.Priority,
			&scenario.CreatedAt, &scenario.UpdatedAt,
		); err != nil {
			return nil, err
		}
		scenarios = append(scenarios, scenario)
	}
	return scenarios, rows.Err()
}

func (r *staffRequirementScenarioRepository) GetDefault() (*models.StaffRequirementScenario, error) {
	scenario := &models.StaffRequirementScenario{}
	query := `SELECT id, scenario_name, description, revenue_level_tier_id, min_revenue, max_revenue,
	          use_day_of_week_revenue, use_specific_date_revenue, doctor_count, min_doctor_count,
	          day_of_week, is_default, is_active, priority, created_at, updated_at
	          FROM staff_requirement_scenarios WHERE is_default = true LIMIT 1`
	err := r.db.QueryRow(query).Scan(
		&scenario.ID, &scenario.ScenarioName, &scenario.Description, &scenario.RevenueLevelTierID,
		&scenario.MinRevenue, &scenario.MaxRevenue, &scenario.UseDayOfWeekRevenue,
		&scenario.UseSpecificDateRevenue, &scenario.DoctorCount, &scenario.MinDoctorCount,
		&scenario.DayOfWeek, &scenario.IsDefault, &scenario.IsActive, &scenario.Priority,
		&scenario.CreatedAt, &scenario.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return scenario, nil
}

package postgres

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"vsq-oper-manpower/backend/internal/domain/interfaces"
	"vsq-oper-manpower/backend/internal/domain/models"
)

type clinicWidePreferenceRepository struct {
	db *sql.DB
}

func NewClinicWidePreferenceRepository(db *sql.DB) interfaces.ClinicWidePreferenceRepository {
	return &clinicWidePreferenceRepository{db: db}
}

func (r *clinicWidePreferenceRepository) Create(preference *models.ClinicWidePreference) error {
	query := `INSERT INTO clinic_wide_preferences 
	          (id, criteria_type, criteria_name, min_value, max_value, is_active, display_order, description)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	          RETURNING created_at, updated_at`
	return r.db.QueryRow(
		query,
		preference.ID, preference.CriteriaType, preference.CriteriaName,
		preference.MinValue, preference.MaxValue, preference.IsActive,
		preference.DisplayOrder, preference.Description,
	).Scan(&preference.CreatedAt, &preference.UpdatedAt)
}

func (r *clinicWidePreferenceRepository) GetByID(id uuid.UUID) (*models.ClinicWidePreference, error) {
	preference := &models.ClinicWidePreference{}
	query := `SELECT id, criteria_type, criteria_name, min_value, max_value, is_active, 
	          display_order, description, created_at, updated_at
	          FROM clinic_wide_preferences WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(
		&preference.ID, &preference.CriteriaType, &preference.CriteriaName,
		&preference.MinValue, &preference.MaxValue, &preference.IsActive,
		&preference.DisplayOrder, &preference.Description,
		&preference.CreatedAt, &preference.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return preference, nil
}

func (r *clinicWidePreferenceRepository) Update(preference *models.ClinicWidePreference) error {
	query := `UPDATE clinic_wide_preferences
	          SET criteria_name = $1, min_value = $2, max_value = $3, is_active = $4, 
	              display_order = $5, description = $6, updated_at = CURRENT_TIMESTAMP
	          WHERE id = $7 RETURNING updated_at`
	return r.db.QueryRow(
		query,
		preference.CriteriaName, preference.MinValue, preference.MaxValue,
		preference.IsActive, preference.DisplayOrder, preference.Description,
		preference.ID,
	).Scan(&preference.UpdatedAt)
}

func (r *clinicWidePreferenceRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM clinic_wide_preferences WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *clinicWidePreferenceRepository) List(filters models.ClinicPreferenceFilters) ([]*models.ClinicWidePreference, error) {
	query := `SELECT id, criteria_type, criteria_name, min_value, max_value, is_active, 
	          display_order, description, created_at, updated_at
	          FROM clinic_wide_preferences WHERE 1=1`
	args := []interface{}{}
	argPos := 1

	if filters.CriteriaType != nil {
		query += fmt.Sprintf(" AND criteria_type = $%d", argPos)
		args = append(args, *filters.CriteriaType)
		argPos++
	}

	if filters.IsActive != nil {
		query += fmt.Sprintf(" AND is_active = $%d", argPos)
		args = append(args, *filters.IsActive)
		argPos++
	}

	query += " ORDER BY criteria_type, display_order, min_value"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	preferences := []*models.ClinicWidePreference{}
	for rows.Next() {
		preference := &models.ClinicWidePreference{}
		if err := rows.Scan(
			&preference.ID, &preference.CriteriaType, &preference.CriteriaName,
			&preference.MinValue, &preference.MaxValue, &preference.IsActive,
			&preference.DisplayOrder, &preference.Description,
			&preference.CreatedAt, &preference.UpdatedAt,
		); err != nil {
			return nil, err
		}
		preferences = append(preferences, preference)
	}

	return preferences, rows.Err()
}

func (r *clinicWidePreferenceRepository) GetByCriteriaTypeAndValue(criteriaType models.ClinicPreferenceCriteriaType, value float64) ([]*models.ClinicWidePreference, error) {
	query := `SELECT id, criteria_type, criteria_name, min_value, max_value, is_active, 
	          display_order, description, created_at, updated_at
	          FROM clinic_wide_preferences 
	          WHERE criteria_type = $1 AND is_active = true
	          AND min_value <= $2 AND (max_value IS NULL OR max_value >= $2)
	          ORDER BY display_order, min_value DESC`
	
	rows, err := r.db.Query(query, criteriaType, value)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	preferences := []*models.ClinicWidePreference{}
	for rows.Next() {
		preference := &models.ClinicWidePreference{}
		if err := rows.Scan(
			&preference.ID, &preference.CriteriaType, &preference.CriteriaName,
			&preference.MinValue, &preference.MaxValue, &preference.IsActive,
			&preference.DisplayOrder, &preference.Description,
			&preference.CreatedAt, &preference.UpdatedAt,
		); err != nil {
			return nil, err
		}
		preferences = append(preferences, preference)
	}

	return preferences, rows.Err()
}

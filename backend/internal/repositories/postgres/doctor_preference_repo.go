package postgres

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"vsq-oper-manpower/backend/internal/domain/interfaces"
	"vsq-oper-manpower/backend/internal/domain/models"
)

type doctorPreferenceRepository struct {
	db *sql.DB
}

func NewDoctorPreferenceRepository(db *sql.DB) interfaces.DoctorPreferenceRepository {
	return &doctorPreferenceRepository{db: db}
}

func (r *doctorPreferenceRepository) Create(preference *models.DoctorPreference) error {
	preference.ID = uuid.New()
	preference.CreatedAt = time.Now()
	preference.UpdatedAt = time.Now()

	ruleConfigJSON, err := json.Marshal(preference.RuleConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal rule_config: %w", err)
	}

	query := `INSERT INTO doctor_preferences (id, doctor_id, branch_id, rule_type, rule_config, is_active, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING created_at, updated_at`
	
	var branchID interface{}
	if preference.BranchID != nil {
		branchID = *preference.BranchID
	}

	return r.db.QueryRow(query, preference.ID, preference.DoctorID, branchID, preference.RuleType, ruleConfigJSON, preference.IsActive, preference.CreatedAt, preference.UpdatedAt).
		Scan(&preference.CreatedAt, &preference.UpdatedAt)
}

func (r *doctorPreferenceRepository) GetByID(id uuid.UUID) (*models.DoctorPreference, error) {
	preference := &models.DoctorPreference{}
	query := `SELECT id, doctor_id, branch_id, rule_type, rule_config, is_active, created_at, updated_at
	          FROM doctor_preferences WHERE id = $1`
	
	var branchID sql.NullString
	var ruleConfigJSON []byte
	
	err := r.db.QueryRow(query, id).Scan(
		&preference.ID, &preference.DoctorID, &branchID, &preference.RuleType, &ruleConfigJSON, &preference.IsActive, &preference.CreatedAt, &preference.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if branchID.Valid {
		bID, _ := uuid.Parse(branchID.String)
		preference.BranchID = &bID
	}

	if len(ruleConfigJSON) > 0 {
		if err := json.Unmarshal(ruleConfigJSON, &preference.RuleConfig); err != nil {
			return nil, fmt.Errorf("failed to unmarshal rule_config: %w", err)
		}
	}

	return preference, nil
}

func (r *doctorPreferenceRepository) GetByDoctorID(doctorID uuid.UUID) ([]*models.DoctorPreference, error) {
	query := `SELECT id, doctor_id, branch_id, rule_type, rule_config, is_active, created_at, updated_at
	          FROM doctor_preferences WHERE doctor_id = $1 ORDER BY created_at DESC`
	
	rows, err := r.db.Query(query, doctorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var preferences []*models.DoctorPreference
	for rows.Next() {
		preference := &models.DoctorPreference{}
		var branchID sql.NullString
		var ruleConfigJSON []byte

		err := rows.Scan(
			&preference.ID, &preference.DoctorID, &branchID, &preference.RuleType, &ruleConfigJSON, &preference.IsActive, &preference.CreatedAt, &preference.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if branchID.Valid {
			bID, _ := uuid.Parse(branchID.String)
			preference.BranchID = &bID
		}

		if len(ruleConfigJSON) > 0 {
			if err := json.Unmarshal(ruleConfigJSON, &preference.RuleConfig); err != nil {
				return nil, fmt.Errorf("failed to unmarshal rule_config: %w", err)
			}
		}

		preferences = append(preferences, preference)
	}

	return preferences, rows.Err()
}

func (r *doctorPreferenceRepository) GetByDoctorAndBranch(doctorID uuid.UUID, branchID uuid.UUID) ([]*models.DoctorPreference, error) {
	query := `SELECT id, doctor_id, branch_id, rule_type, rule_config, is_active, created_at, updated_at
	          FROM doctor_preferences WHERE doctor_id = $1 AND (branch_id = $2 OR branch_id IS NULL) ORDER BY created_at DESC`
	
	rows, err := r.db.Query(query, doctorID, branchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var preferences []*models.DoctorPreference
	for rows.Next() {
		preference := &models.DoctorPreference{}
		var branchIDVal sql.NullString
		var ruleConfigJSON []byte

		err := rows.Scan(
			&preference.ID, &preference.DoctorID, &branchIDVal, &preference.RuleType, &ruleConfigJSON, &preference.IsActive, &preference.CreatedAt, &preference.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if branchIDVal.Valid {
			bID, _ := uuid.Parse(branchIDVal.String)
			preference.BranchID = &bID
		}

		if len(ruleConfigJSON) > 0 {
			if err := json.Unmarshal(ruleConfigJSON, &preference.RuleConfig); err != nil {
				return nil, fmt.Errorf("failed to unmarshal rule_config: %w", err)
			}
		}

		preferences = append(preferences, preference)
	}

	return preferences, rows.Err()
}

func (r *doctorPreferenceRepository) GetActiveByDoctorID(doctorID uuid.UUID) ([]*models.DoctorPreference, error) {
	query := `SELECT id, doctor_id, branch_id, rule_type, rule_config, is_active, created_at, updated_at
	          FROM doctor_preferences WHERE doctor_id = $1 AND is_active = true ORDER BY created_at DESC`
	
	rows, err := r.db.Query(query, doctorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var preferences []*models.DoctorPreference
	for rows.Next() {
		preference := &models.DoctorPreference{}
		var branchID sql.NullString
		var ruleConfigJSON []byte

		err := rows.Scan(
			&preference.ID, &preference.DoctorID, &branchID, &preference.RuleType, &ruleConfigJSON, &preference.IsActive, &preference.CreatedAt, &preference.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if branchID.Valid {
			bID, _ := uuid.Parse(branchID.String)
			preference.BranchID = &bID
		}

		if len(ruleConfigJSON) > 0 {
			if err := json.Unmarshal(ruleConfigJSON, &preference.RuleConfig); err != nil {
				return nil, fmt.Errorf("failed to unmarshal rule_config: %w", err)
			}
		}

		preferences = append(preferences, preference)
	}

	return preferences, rows.Err()
}

func (r *doctorPreferenceRepository) Update(preference *models.DoctorPreference) error {
	preference.UpdatedAt = time.Now()

	ruleConfigJSON, err := json.Marshal(preference.RuleConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal rule_config: %w", err)
	}

	query := `UPDATE doctor_preferences SET doctor_id = $1, branch_id = $2, rule_type = $3, 
	          rule_config = $4, is_active = $5, updated_at = $6 WHERE id = $7`
	
	var branchID interface{}
	if preference.BranchID != nil {
		branchID = *preference.BranchID
	}

	_, err = r.db.Exec(query, preference.DoctorID, branchID, preference.RuleType, ruleConfigJSON, preference.IsActive, preference.UpdatedAt, preference.ID)
	return err
}

func (r *doctorPreferenceRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM doctor_preferences WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *doctorPreferenceRepository) DeleteByDoctorID(doctorID uuid.UUID) error {
	query := `DELETE FROM doctor_preferences WHERE doctor_id = $1`
	_, err := r.db.Exec(query, doctorID)
	return err
}

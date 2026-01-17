package postgres

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"vsq-oper-manpower/backend/internal/domain/interfaces"
	"vsq-oper-manpower/backend/internal/domain/models"
)

type doctorRepository struct {
	db *sql.DB
}

func NewDoctorRepository(db *sql.DB) interfaces.DoctorRepository {
	return &doctorRepository{db: db}
}

func (r *doctorRepository) Create(doctor *models.Doctor) error {
	doctor.ID = uuid.New()
	doctor.CreatedAt = time.Now()
	doctor.UpdatedAt = time.Now()

	query := `INSERT INTO doctors (id, name, code, preferences, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5, $6) RETURNING created_at, updated_at`
	
	var code sql.NullString
	if doctor.Code != "" {
		code = sql.NullString{String: doctor.Code, Valid: true}
	}
	var preferences sql.NullString
	if doctor.Preferences != "" {
		preferences = sql.NullString{String: doctor.Preferences, Valid: true}
	}

	return r.db.QueryRow(query, doctor.ID, doctor.Name, code, preferences, doctor.CreatedAt, doctor.UpdatedAt).
		Scan(&doctor.CreatedAt, &doctor.UpdatedAt)
}

func (r *doctorRepository) GetByID(id uuid.UUID) (*models.Doctor, error) {
	doctor := &models.Doctor{}
	query := `SELECT id, name, code, preferences, created_at, updated_at
	          FROM doctors WHERE id = $1`
	
	var code, preferences sql.NullString
	
	err := r.db.QueryRow(query, id).Scan(
		&doctor.ID, &doctor.Name, &code, &preferences, &doctor.CreatedAt, &doctor.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if code.Valid {
		doctor.Code = code.String
	}
	if preferences.Valid {
		doctor.Preferences = preferences.String
	}

	return doctor, nil
}

func (r *doctorRepository) GetByCode(code string) (*models.Doctor, error) {
	doctor := &models.Doctor{}
	query := `SELECT id, name, code, preferences, created_at, updated_at
	          FROM doctors WHERE code = $1`
	
	var codeVal, preferences sql.NullString
	
	err := r.db.QueryRow(query, code).Scan(
		&doctor.ID, &doctor.Name, &codeVal, &preferences, &doctor.CreatedAt, &doctor.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if codeVal.Valid {
		doctor.Code = codeVal.String
	}
	if preferences.Valid {
		doctor.Preferences = preferences.String
	}

	return doctor, nil
}

func (r *doctorRepository) Update(doctor *models.Doctor) error {
	doctor.UpdatedAt = time.Now()

	query := `UPDATE doctors SET name = $1, code = $2, preferences = $3, updated_at = $4 WHERE id = $5`
	
	var code sql.NullString
	if doctor.Code != "" {
		code = sql.NullString{String: doctor.Code, Valid: true}
	}
	var preferences sql.NullString
	if doctor.Preferences != "" {
		preferences = sql.NullString{String: doctor.Preferences, Valid: true}
	}

	_, err := r.db.Exec(query, doctor.Name, code, preferences, doctor.UpdatedAt, doctor.ID)
	return err
}

func (r *doctorRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM doctors WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *doctorRepository) List() ([]*models.Doctor, error) {
	query := `SELECT id, name, code, preferences, created_at, updated_at
	          FROM doctors ORDER BY name`
	
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var doctors []*models.Doctor
	for rows.Next() {
		doctor := &models.Doctor{}
		var code, preferences sql.NullString

		err := rows.Scan(
			&doctor.ID, &doctor.Name, &code, &preferences, &doctor.CreatedAt, &doctor.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if code.Valid {
			doctor.Code = code.String
		}
		if preferences.Valid {
			doctor.Preferences = preferences.String
		}

		doctors = append(doctors, doctor)
	}

	return doctors, rows.Err()
}

package postgres

import (
	"database/sql"
	"fmt"
	"time"

	"vsq-oper-manpower/backend/internal/domain/interfaces"
	"vsq-oper-manpower/backend/internal/domain/models"

	"github.com/google/uuid"
)

type allocationSuggestionRepository struct {
	db *sql.DB
}

func NewAllocationSuggestionRepository(db *sql.DB) interfaces.AllocationSuggestionRepository {
	return &allocationSuggestionRepository{db: db}
}

func (r *allocationSuggestionRepository) Create(suggestion *models.AllocationSuggestion) error {
	suggestion.ID = uuid.New()
	now := time.Now()
	suggestion.CreatedAt = now
	suggestion.UpdatedAt = now

	query := `INSERT INTO allocation_suggestions 
	          (id, rotation_staff_id, branch_id, date, position_id, status, confidence, reason, criteria_used, reviewed_by, reviewed_at, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) RETURNING created_at, updated_at`

	var reviewedBy sql.NullString
	if suggestion.ReviewedBy != nil {
		reviewedBy = sql.NullString{String: suggestion.ReviewedBy.String(), Valid: true}
	}

	var reviewedAt sql.NullTime
	if suggestion.ReviewedAt != nil {
		reviewedAt = sql.NullTime{Time: *suggestion.ReviewedAt, Valid: true}
	}

	return r.db.QueryRow(query,
		suggestion.ID,
		suggestion.RotationStaffID,
		suggestion.BranchID,
		suggestion.Date,
		suggestion.PositionID,
		suggestion.Status,
		suggestion.Confidence,
		suggestion.Reason,
		suggestion.CriteriaUsed,
		reviewedBy,
		reviewedAt,
		suggestion.CreatedAt,
		suggestion.UpdatedAt,
	).Scan(&suggestion.CreatedAt, &suggestion.UpdatedAt)
}

func (r *allocationSuggestionRepository) GetByID(id uuid.UUID) (*models.AllocationSuggestion, error) {
	suggestion := &models.AllocationSuggestion{}
	query := `SELECT id, rotation_staff_id, branch_id, date, position_id, status, confidence, reason, criteria_used, reviewed_by, reviewed_at, created_at, updated_at
	          FROM allocation_suggestions WHERE id = $1`

	var reviewedBy sql.NullString
	var reviewedAt sql.NullTime

	err := r.db.QueryRow(query, id).Scan(
		&suggestion.ID,
		&suggestion.RotationStaffID,
		&suggestion.BranchID,
		&suggestion.Date,
		&suggestion.PositionID,
		&suggestion.Status,
		&suggestion.Confidence,
		&suggestion.Reason,
		&suggestion.CriteriaUsed,
		&reviewedBy,
		&reviewedAt,
		&suggestion.CreatedAt,
		&suggestion.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if reviewedBy.Valid {
		reviewedByUUID, err := uuid.Parse(reviewedBy.String)
		if err == nil {
			suggestion.ReviewedBy = &reviewedByUUID
		}
	}

	if reviewedAt.Valid {
		suggestion.ReviewedAt = &reviewedAt.Time
	}

	return suggestion, nil
}

func (r *allocationSuggestionRepository) Update(suggestion *models.AllocationSuggestion) error {
	suggestion.UpdatedAt = time.Now()

	query := `UPDATE allocation_suggestions 
	          SET status = $1, confidence = $2, reason = $3, criteria_used = $4, reviewed_by = $5, reviewed_at = $6, updated_at = $7 
	          WHERE id = $8`

	var reviewedBy sql.NullString
	if suggestion.ReviewedBy != nil {
		reviewedBy = sql.NullString{String: suggestion.ReviewedBy.String(), Valid: true}
	}

	var reviewedAt sql.NullTime
	if suggestion.ReviewedAt != nil {
		reviewedAt = sql.NullTime{Time: *suggestion.ReviewedAt, Valid: true}
	}

	_, err := r.db.Exec(query,
		suggestion.Status,
		suggestion.Confidence,
		suggestion.Reason,
		suggestion.CriteriaUsed,
		reviewedBy,
		reviewedAt,
		suggestion.UpdatedAt,
		suggestion.ID,
	)
	return err
}

func (r *allocationSuggestionRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM allocation_suggestions WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *allocationSuggestionRepository) List(filters interfaces.AllocationSuggestionFilters) ([]*models.AllocationSuggestion, error) {
	query := `SELECT id, rotation_staff_id, branch_id, date, position_id, status, confidence, reason, criteria_used, reviewed_by, reviewed_at, created_at, updated_at
	          FROM allocation_suggestions WHERE 1=1`
	args := []interface{}{}
	argIndex := 1

	if filters.BranchID != nil {
		query += fmt.Sprintf(" AND branch_id = $%d", argIndex)
		args = append(args, *filters.BranchID)
		argIndex++
	}

	if filters.RotationStaffID != nil {
		query += fmt.Sprintf(" AND rotation_staff_id = $%d", argIndex)
		args = append(args, *filters.RotationStaffID)
		argIndex++
	}

	if filters.PositionID != nil {
		query += fmt.Sprintf(" AND position_id = $%d", argIndex)
		args = append(args, *filters.PositionID)
		argIndex++
	}

	if filters.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, *filters.Status)
		argIndex++
	}

	if filters.StartDate != nil {
		query += fmt.Sprintf(" AND date >= $%d", argIndex)
		args = append(args, *filters.StartDate)
		argIndex++
	}

	if filters.EndDate != nil {
		query += fmt.Sprintf(" AND date <= $%d", argIndex)
		args = append(args, *filters.EndDate)
		argIndex++
	}

	query += " ORDER BY date DESC, confidence DESC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	suggestions := []*models.AllocationSuggestion{}
	for rows.Next() {
		suggestion := &models.AllocationSuggestion{}
		var reviewedBy sql.NullString
		var reviewedAt sql.NullTime

		if err := rows.Scan(
			&suggestion.ID,
			&suggestion.RotationStaffID,
			&suggestion.BranchID,
			&suggestion.Date,
			&suggestion.PositionID,
			&suggestion.Status,
			&suggestion.Confidence,
			&suggestion.Reason,
			&suggestion.CriteriaUsed,
			&reviewedBy,
			&reviewedAt,
			&suggestion.CreatedAt,
			&suggestion.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if reviewedBy.Valid {
			reviewedByUUID, err := uuid.Parse(reviewedBy.String)
			if err == nil {
				suggestion.ReviewedBy = &reviewedByUUID
			}
		}

		if reviewedAt.Valid {
			suggestion.ReviewedAt = &reviewedAt.Time
		}

		suggestions = append(suggestions, suggestion)
	}

	return suggestions, rows.Err()
}

func (r *allocationSuggestionRepository) GetByBranchID(branchID uuid.UUID, startDate, endDate time.Time) ([]*models.AllocationSuggestion, error) {
	filters := interfaces.AllocationSuggestionFilters{
		BranchID:  &branchID,
		StartDate: &startDate,
		EndDate:   &endDate,
	}
	return r.List(filters)
}

func (r *allocationSuggestionRepository) GetByDateRange(startDate, endDate time.Time) ([]*models.AllocationSuggestion, error) {
	filters := interfaces.AllocationSuggestionFilters{
		StartDate: &startDate,
		EndDate:   &endDate,
	}
	return r.List(filters)
}

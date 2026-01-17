package postgres

import (
	"database/sql"

	"github.com/google/uuid"
	"vsq-oper-manpower/backend/internal/domain/interfaces"
	"vsq-oper-manpower/backend/internal/domain/models"
)

// BranchWeeklyRevenueRepository implementation
type branchWeeklyRevenueRepository struct {
	db *sql.DB
}

func NewBranchWeeklyRevenueRepository(db *sql.DB) interfaces.BranchWeeklyRevenueRepository {
	return &branchWeeklyRevenueRepository{db: db}
}

func (r *branchWeeklyRevenueRepository) Create(revenue *models.BranchWeeklyRevenue) error {
	query := `INSERT INTO branch_weekly_revenue (id, branch_id, day_of_week, expected_revenue) 
	          VALUES ($1, $2, $3, $4) RETURNING created_at, updated_at`
	return r.db.QueryRow(query, revenue.ID, revenue.BranchID, revenue.DayOfWeek, revenue.ExpectedRevenue).
		Scan(&revenue.CreatedAt, &revenue.UpdatedAt)
}

func (r *branchWeeklyRevenueRepository) Update(revenue *models.BranchWeeklyRevenue) error {
	query := `UPDATE branch_weekly_revenue 
	          SET day_of_week = $2, expected_revenue = $3, updated_at = CURRENT_TIMESTAMP 
	          WHERE id = $1 RETURNING updated_at`
	return r.db.QueryRow(query, revenue.ID, revenue.DayOfWeek, revenue.ExpectedRevenue).
		Scan(&revenue.UpdatedAt)
}

func (r *branchWeeklyRevenueRepository) GetByID(id uuid.UUID) (*models.BranchWeeklyRevenue, error) {
	revenue := &models.BranchWeeklyRevenue{}
	query := `SELECT id, branch_id, day_of_week, expected_revenue, created_at, updated_at 
	          FROM branch_weekly_revenue WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(
		&revenue.ID, &revenue.BranchID, &revenue.DayOfWeek, &revenue.ExpectedRevenue,
		&revenue.CreatedAt, &revenue.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return revenue, nil
}

func (r *branchWeeklyRevenueRepository) GetByBranchID(branchID uuid.UUID) ([]*models.BranchWeeklyRevenue, error) {
	query := `SELECT id, branch_id, day_of_week, expected_revenue, created_at, updated_at 
	          FROM branch_weekly_revenue WHERE branch_id = $1 ORDER BY day_of_week`
	rows, err := r.db.Query(query, branchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	revenues := []*models.BranchWeeklyRevenue{}
	for rows.Next() {
		revenue := &models.BranchWeeklyRevenue{}
		if err := rows.Scan(
			&revenue.ID, &revenue.BranchID, &revenue.DayOfWeek, &revenue.ExpectedRevenue,
			&revenue.CreatedAt, &revenue.UpdatedAt,
		); err != nil {
			return nil, err
		}
		revenues = append(revenues, revenue)
	}
	return revenues, rows.Err()
}

func (r *branchWeeklyRevenueRepository) GetByBranchIDAndDayOfWeek(branchID uuid.UUID, dayOfWeek int) (*models.BranchWeeklyRevenue, error) {
	revenue := &models.BranchWeeklyRevenue{}
	query := `SELECT id, branch_id, day_of_week, expected_revenue, created_at, updated_at 
	          FROM branch_weekly_revenue WHERE branch_id = $1 AND day_of_week = $2`
	err := r.db.QueryRow(query, branchID, dayOfWeek).Scan(
		&revenue.ID, &revenue.BranchID, &revenue.DayOfWeek, &revenue.ExpectedRevenue,
		&revenue.CreatedAt, &revenue.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return revenue, nil
}

func (r *branchWeeklyRevenueRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM branch_weekly_revenue WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *branchWeeklyRevenueRepository) BulkUpsert(revenues []*models.BranchWeeklyRevenue) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `INSERT INTO branch_weekly_revenue (id, branch_id, day_of_week, expected_revenue, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	          ON CONFLICT (branch_id, day_of_week) 
	          DO UPDATE SET expected_revenue = EXCLUDED.expected_revenue, updated_at = CURRENT_TIMESTAMP
	          RETURNING created_at, updated_at`

	for _, revenue := range revenues {
		if revenue.ID == uuid.Nil {
			revenue.ID = uuid.New()
		}
		err := tx.QueryRow(query, revenue.ID, revenue.BranchID, revenue.DayOfWeek, revenue.ExpectedRevenue).
			Scan(&revenue.CreatedAt, &revenue.UpdatedAt)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

package postgres

import (
	"database/sql"

	"github.com/google/uuid"
	"vsq-oper-manpower/backend/internal/domain/interfaces"
	"vsq-oper-manpower/backend/internal/domain/models"
)

type revenueLevelTierRepository struct {
	db *sql.DB
}

func NewRevenueLevelTierRepository(db *sql.DB) interfaces.RevenueLevelTierRepository {
	return &revenueLevelTierRepository{db: db}
}

func (r *revenueLevelTierRepository) Create(tier *models.RevenueLevelTier) error {
	query := `INSERT INTO revenue_level_tiers (id, level_number, level_name, min_revenue, max_revenue, display_order, color_code, description)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING created_at, updated_at`
	return r.db.QueryRow(
		query,
		tier.ID, tier.LevelNumber, tier.LevelName, tier.MinRevenue, tier.MaxRevenue,
		tier.DisplayOrder, tier.ColorCode, tier.Description,
	).Scan(&tier.CreatedAt, &tier.UpdatedAt)
}

func (r *revenueLevelTierRepository) GetByID(id uuid.UUID) (*models.RevenueLevelTier, error) {
	tier := &models.RevenueLevelTier{}
	query := `SELECT id, level_number, level_name, min_revenue, max_revenue, display_order, color_code, description, created_at, updated_at
	          FROM revenue_level_tiers WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(
		&tier.ID, &tier.LevelNumber, &tier.LevelName, &tier.MinRevenue, &tier.MaxRevenue,
		&tier.DisplayOrder, &tier.ColorCode, &tier.Description, &tier.CreatedAt, &tier.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return tier, nil
}

func (r *revenueLevelTierRepository) GetByLevelNumber(levelNumber int) (*models.RevenueLevelTier, error) {
	tier := &models.RevenueLevelTier{}
	query := `SELECT id, level_number, level_name, min_revenue, max_revenue, display_order, color_code, description, created_at, updated_at
	          FROM revenue_level_tiers WHERE level_number = $1`
	err := r.db.QueryRow(query, levelNumber).Scan(
		&tier.ID, &tier.LevelNumber, &tier.LevelName, &tier.MinRevenue, &tier.MaxRevenue,
		&tier.DisplayOrder, &tier.ColorCode, &tier.Description, &tier.CreatedAt, &tier.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return tier, nil
}

func (r *revenueLevelTierRepository) Update(tier *models.RevenueLevelTier) error {
	query := `UPDATE revenue_level_tiers
	          SET level_name = $1, min_revenue = $2, max_revenue = $3, display_order = $4, color_code = $5, description = $6, updated_at = CURRENT_TIMESTAMP
	          WHERE id = $7 RETURNING updated_at`
	return r.db.QueryRow(
		query,
		tier.LevelName, tier.MinRevenue, tier.MaxRevenue, tier.DisplayOrder, tier.ColorCode, tier.Description, tier.ID,
	).Scan(&tier.UpdatedAt)
}

func (r *revenueLevelTierRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM revenue_level_tiers WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *revenueLevelTierRepository) List() ([]*models.RevenueLevelTier, error) {
	query := `SELECT id, level_number, level_name, min_revenue, max_revenue, display_order, color_code, description, created_at, updated_at
	          FROM revenue_level_tiers ORDER BY display_order, level_number`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tiers := []*models.RevenueLevelTier{}
	for rows.Next() {
		tier := &models.RevenueLevelTier{}
		if err := rows.Scan(
			&tier.ID, &tier.LevelNumber, &tier.LevelName, &tier.MinRevenue, &tier.MaxRevenue,
			&tier.DisplayOrder, &tier.ColorCode, &tier.Description, &tier.CreatedAt, &tier.UpdatedAt,
		); err != nil {
			return nil, err
		}
		tiers = append(tiers, tier)
	}
	return tiers, rows.Err()
}

func (r *revenueLevelTierRepository) GetTierForRevenue(revenue float64) (*models.RevenueLevelTier, error) {
	tier := &models.RevenueLevelTier{}
	query := `SELECT id, level_number, level_name, min_revenue, max_revenue, display_order, color_code, description, created_at, updated_at
	          FROM revenue_level_tiers
	          WHERE revenue >= min_revenue
	            AND (max_revenue IS NULL OR revenue < max_revenue)
	          ORDER BY level_number DESC
	          LIMIT 1`
	err := r.db.QueryRow(query, revenue).Scan(
		&tier.ID, &tier.LevelNumber, &tier.LevelName, &tier.MinRevenue, &tier.MaxRevenue,
		&tier.DisplayOrder, &tier.ColorCode, &tier.Description, &tier.CreatedAt, &tier.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return tier, nil
}

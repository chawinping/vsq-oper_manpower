package postgres

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"vsq-oper-manpower/backend/internal/domain/interfaces"
	"vsq-oper-manpower/backend/internal/domain/models"
)

type Repositories struct {
	User                interfaces.UserRepository
	Role                interfaces.RoleRepository
	Staff               interfaces.StaffRepository
	Position            interfaces.PositionRepository
	Branch              interfaces.BranchRepository
	EffectiveBranch     interfaces.EffectiveBranchRepository
	Revenue             interfaces.RevenueRepository
	Schedule            interfaces.ScheduleRepository
	Rotation            interfaces.RotationRepository
	Settings            interfaces.SettingsRepository
	AllocationRule      interfaces.AllocationRuleRepository
		AreaOfOperation     interfaces.AreaOfOperationRepository
		Zone                interfaces.ZoneRepository
	AllocationCriteria  interfaces.AllocationCriteriaRepository
	PositionQuota       interfaces.PositionQuotaRepository
	Doctor              interfaces.DoctorRepository
	DoctorPreference    interfaces.DoctorPreferenceRepository
	DoctorAssignment    interfaces.DoctorAssignmentRepository
	DoctorOnOffDay      interfaces.DoctorOnOffDayRepository
	DoctorDefaultSchedule interfaces.DoctorDefaultScheduleRepository
	DoctorWeeklyOffDay   interfaces.DoctorWeeklyOffDayRepository
	DoctorScheduleOverride interfaces.DoctorScheduleOverrideRepository
	AllocationSuggestion interfaces.AllocationSuggestionRepository
	BranchWeeklyRevenue  interfaces.BranchWeeklyRevenueRepository
	BranchConstraints    interfaces.BranchConstraintsRepository
	RevenueLevelTier     interfaces.RevenueLevelTierRepository
	StaffRequirementScenario interfaces.StaffRequirementScenarioRepository
	ScenarioPositionRequirement interfaces.ScenarioPositionRequirementRepository
}

func NewRepositories(db *sql.DB) *Repositories {
	return &Repositories{
		User:                 NewUserRepository(db),
		Role:                 NewRoleRepository(db),
		Staff:                NewStaffRepository(db),
		Position:             NewPositionRepository(db),
		Branch:               NewBranchRepository(db),
		EffectiveBranch:      NewEffectiveBranchRepository(db),
		Revenue:              NewRevenueRepository(db),
		Schedule:             NewScheduleRepository(db),
		Rotation:             NewRotationRepository(db),
		Settings:             NewSettingsRepository(db),
		AllocationRule:       NewAllocationRuleRepository(db),
		AreaOfOperation:      NewAreaOfOperationRepository(db),
		Zone:                  NewZoneRepository(db),
		AllocationCriteria:   NewAllocationCriteriaRepository(db),
		PositionQuota:        NewPositionQuotaRepository(db),
		Doctor:               NewDoctorRepository(db),
		DoctorPreference:     NewDoctorPreferenceRepository(db),
		DoctorAssignment:    NewDoctorAssignmentRepository(db),
		DoctorOnOffDay:       NewDoctorOnOffDayRepository(db),
		DoctorDefaultSchedule: NewDoctorDefaultScheduleRepository(db),
		DoctorWeeklyOffDay:   NewDoctorWeeklyOffDayRepository(db),
		DoctorScheduleOverride: NewDoctorScheduleOverrideRepository(db),
		AllocationSuggestion: NewAllocationSuggestionRepository(db),
		BranchWeeklyRevenue:  NewBranchWeeklyRevenueRepository(db),
		BranchConstraints:    NewBranchConstraintsRepository(db),
		RevenueLevelTier:     NewRevenueLevelTierRepository(db),
		StaffRequirementScenario: NewStaffRequirementScenarioRepository(db),
		ScenarioPositionRequirement: NewScenarioPositionRequirementRepository(db),
	}
}

// UserRepository implementation
type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) interfaces.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *models.User) error {
	query := `INSERT INTO users (id, username, email, password_hash, role_id, branch_id) 
	          VALUES ($1, $2, $3, $4, $5, $6) RETURNING created_at, updated_at`
	return r.db.QueryRow(query, user.ID, user.Username, user.Email, user.PasswordHash, user.RoleID, user.BranchID).
		Scan(&user.CreatedAt, &user.UpdatedAt)
}

func (r *userRepository) GetByID(id uuid.UUID) (*models.User, error) {
	user := &models.User{}
	query := `SELECT id, username, email, password_hash, role_id, branch_id, created_at, updated_at 
	          FROM users WHERE id = $1`
	var branchID sql.NullString
	err := r.db.QueryRow(query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.RoleID, &branchID, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if branchID.Valid {
		bID, _ := uuid.Parse(branchID.String)
		user.BranchID = &bID
	}
	return user, nil
}

func (r *userRepository) GetByUsername(username string) (*models.User, error) {
	user := &models.User{}
	query := `SELECT id, username, email, password_hash, role_id, branch_id, created_at, updated_at 
	          FROM users WHERE username = $1`
	var branchID sql.NullString
	err := r.db.QueryRow(query, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.RoleID, &branchID, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if branchID.Valid {
		bID, _ := uuid.Parse(branchID.String)
		user.BranchID = &bID
	}
	return user, nil
}

func (r *userRepository) GetByEmail(email string) (*models.User, error) {
	user := &models.User{}
	query := `SELECT id, username, email, password_hash, role_id, created_at, updated_at 
	          FROM users WHERE email = $1`
	err := r.db.QueryRow(query, email).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.RoleID, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return user, err
}

func (r *userRepository) Update(user *models.User) error {
	query := `UPDATE users SET username = $1, email = $2, password_hash = $3, 
	          role_id = $4, branch_id = $5, updated_at = CURRENT_TIMESTAMP WHERE id = $6`
	_, err := r.db.Exec(query, user.Username, user.Email, user.PasswordHash, user.RoleID, user.BranchID, user.ID)
	return err
}

func (r *userRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *userRepository) List() ([]*models.User, error) {
	query := `SELECT id, username, email, password_hash, role_id, branch_id, created_at, updated_at 
	          FROM users ORDER BY created_at DESC`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		var branchID sql.NullString
		if err := rows.Scan(
			&user.ID, &user.Username, &user.Email, &user.PasswordHash,
			&user.RoleID, &branchID, &user.CreatedAt, &user.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if branchID.Valid {
			bID, _ := uuid.Parse(branchID.String)
			user.BranchID = &bID
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

// RoleRepository implementation
type roleRepository struct {
	db *sql.DB
}

func NewRoleRepository(db *sql.DB) interfaces.RoleRepository {
	return &roleRepository{db: db}
}

func (r *roleRepository) GetByID(id uuid.UUID) (*models.Role, error) {
	role := &models.Role{}
	query := `SELECT id, name, created_at FROM roles WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(&role.ID, &role.Name, &role.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return role, err
}

func (r *roleRepository) GetByName(name string) (*models.Role, error) {
	role := &models.Role{}
	query := `SELECT id, name, created_at FROM roles WHERE name = $1`
	err := r.db.QueryRow(query, name).Scan(&role.ID, &role.Name, &role.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return role, err
}

func (r *roleRepository) List() ([]*models.Role, error) {
	query := `SELECT id, name, created_at FROM roles ORDER BY name`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []*models.Role
	for rows.Next() {
		role := &models.Role{}
		if err := rows.Scan(&role.ID, &role.Name, &role.CreatedAt); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, rows.Err()
}

// StaffRepository implementation
type staffRepository struct {
	db *sql.DB
}

func NewStaffRepository(db *sql.DB) interfaces.StaffRepository {
	return &staffRepository{db: db}
}

func (r *staffRepository) Create(staff *models.Staff) error {
	query := `INSERT INTO staff (id, nickname, name, staff_type, position_id, branch_id, coverage_area, area_of_operation_id, zone_id, skill_level) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING created_at, updated_at`
	return r.db.QueryRow(query, staff.ID, staff.Nickname, staff.Name, staff.StaffType, staff.PositionID,
		staff.BranchID, staff.CoverageArea, staff.AreaOfOperationID, staff.ZoneID, staff.SkillLevel).Scan(&staff.CreatedAt, &staff.UpdatedAt)
}

func (r *staffRepository) GetByID(id uuid.UUID) (*models.Staff, error) {
	staff := &models.Staff{}
	query := `SELECT id, nickname, name, staff_type, position_id, branch_id, coverage_area, area_of_operation_id, zone_id, skill_level, created_at, updated_at 
	          FROM staff WHERE id = $1`
	var branchID sql.NullString
	var areaOfOpID sql.NullString
	var zoneID sql.NullString
	var nickname sql.NullString
	err := r.db.QueryRow(query, id).Scan(
		&staff.ID, &nickname, &staff.Name, &staff.StaffType, &staff.PositionID,
		&branchID, &staff.CoverageArea, &areaOfOpID, &zoneID, &staff.SkillLevel, &staff.CreatedAt, &staff.UpdatedAt,
	)
	if areaOfOpID.Valid {
		aooID, _ := uuid.Parse(areaOfOpID.String)
		staff.AreaOfOperationID = &aooID
	}
	if zoneID.Valid {
		zID, _ := uuid.Parse(zoneID.String)
		staff.ZoneID = &zID
	}
	if nickname.Valid {
		staff.Nickname = nickname.String
	}
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if branchID.Valid {
		bID, _ := uuid.Parse(branchID.String)
		staff.BranchID = &bID
	}
	
	// Load branches if this is rotation staff
	if staff.StaffType == models.StaffTypeRotation {
		branches, err := r.GetBranches(id)
		if err == nil {
			staff.Branches = branches
		}
	}
	
	return staff, nil
}

func (r *staffRepository) Update(staff *models.Staff) error {
	query := `UPDATE staff SET nickname = $1, name = $2, staff_type = $3, position_id = $4, 
	          branch_id = $5, coverage_area = $6, area_of_operation_id = $7, zone_id = $8, skill_level = $9, updated_at = CURRENT_TIMESTAMP WHERE id = $10`
	_, err := r.db.Exec(query, staff.Nickname, staff.Name, staff.StaffType, staff.PositionID,
		staff.BranchID, staff.CoverageArea, staff.AreaOfOperationID, staff.ZoneID, staff.SkillLevel, staff.ID)
	return err
}

func (r *staffRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM staff WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *staffRepository) GetBranches(staffID uuid.UUID) ([]*models.Branch, error) {
	query := `SELECT b.id, b.name, b.code, b.area_manager_id, b.priority, b.created_at, b.updated_at
	          FROM branches b
	          INNER JOIN staff_branches sb ON b.id = sb.branch_id
	          WHERE sb.staff_id = $1
	          ORDER BY b.name`
	
	rows, err := r.db.Query(query, staffID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var branches []*models.Branch
	for rows.Next() {
		branch := &models.Branch{}
		var areaManagerID sql.NullString
		err := rows.Scan(
			&branch.ID, &branch.Name, &branch.Code, &areaManagerID, &branch.Priority,
			&branch.CreatedAt, &branch.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		if areaManagerID.Valid {
			if id, err := uuid.Parse(areaManagerID.String); err == nil {
				branch.AreaManagerID = &id
			}
		}
		branches = append(branches, branch)
	}
	return branches, rows.Err()
}

func (r *staffRepository) BulkUpdateBranches(staffID uuid.UUID, branchIDs []uuid.UUID) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete all existing branches for this staff
	deleteQuery := `DELETE FROM staff_branches WHERE staff_id = $1`
	if _, err := tx.Exec(deleteQuery, staffID); err != nil {
		return err
	}

	// Insert new branches
	insertQuery := `INSERT INTO staff_branches (id, staff_id, branch_id) 
	                VALUES (gen_random_uuid(), $1, $2)`
	for _, branchID := range branchIDs {
		if _, err := tx.Exec(insertQuery, staffID, branchID); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *staffRepository) List(filters interfaces.StaffFilters) ([]*models.Staff, error) {
	query := `SELECT s.id, s.nickname, s.name, s.staff_type, s.position_id, s.branch_id, s.coverage_area, s.area_of_operation_id, s.zone_id, s.skill_level, s.created_at, s.updated_at 
	          FROM staff s
	          LEFT JOIN positions p ON s.position_id = p.id
	          WHERE 1=1`
	args := []interface{}{}
	argPos := 1

	if filters.StaffType != nil {
		query += ` AND s.staff_type = $` + strconv.Itoa(argPos)
		args = append(args, *filters.StaffType)
		argPos++
	}
	if filters.BranchID != nil {
		query += ` AND s.branch_id = $` + strconv.Itoa(argPos)
		args = append(args, *filters.BranchID)
		argPos++
	}
	if filters.PositionID != nil {
		query += ` AND s.position_id = $` + strconv.Itoa(argPos)
		args = append(args, *filters.PositionID)
		argPos++
	}
	if filters.AreaOfOperationID != nil {
		query += ` AND s.area_of_operation_id = $` + strconv.Itoa(argPos)
		args = append(args, *filters.AreaOfOperationID)
		argPos++
	}

	query += ` ORDER BY COALESCE(p.display_order, 999), s.name`

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var staffList []*models.Staff
	for rows.Next() {
		staff := &models.Staff{}
		var branchID sql.NullString
		var areaOfOpID sql.NullString
		var zoneID sql.NullString
		var nickname sql.NullString
		if err := rows.Scan(
			&staff.ID, &nickname, &staff.Name, &staff.StaffType, &staff.PositionID,
			&branchID, &staff.CoverageArea, &areaOfOpID, &zoneID, &staff.SkillLevel, &staff.CreatedAt, &staff.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if nickname.Valid {
			staff.Nickname = nickname.String
		}
		if branchID.Valid {
			bID, _ := uuid.Parse(branchID.String)
			staff.BranchID = &bID
		}
		if areaOfOpID.Valid {
			aooID, _ := uuid.Parse(areaOfOpID.String)
			staff.AreaOfOperationID = &aooID
		}
		if zoneID.Valid {
			zID, _ := uuid.Parse(zoneID.String)
			staff.ZoneID = &zID
		}
		staffList = append(staffList, staff)
	}
	return staffList, rows.Err()
}

func (r *staffRepository) GetByBranchID(branchID uuid.UUID) ([]*models.Staff, error) {
	filters := interfaces.StaffFilters{BranchID: &branchID}
	return r.List(filters)
}

func (r *staffRepository) GetRotationStaff() ([]*models.Staff, error) {
	rotationType := models.StaffTypeRotation
	filters := interfaces.StaffFilters{StaffType: &rotationType}
	return r.List(filters)
}

// PositionRepository implementation
type positionRepository struct {
	db *sql.DB
}

func NewPositionRepository(db *sql.DB) interfaces.PositionRepository {
	return &positionRepository{db: db}
}

func (r *positionRepository) Create(position *models.Position) error {
	query := `INSERT INTO positions (id, name, min_staff_per_branch, display_order, position_type, manpower_type) 
	          VALUES ($1, $2, $3, $4, $5, $6) RETURNING created_at`
	return r.db.QueryRow(query, position.ID, position.Name, position.MinStaffPerBranch,
		position.DisplayOrder, position.PositionType, position.ManpowerType).Scan(&position.CreatedAt)
}

func (r *positionRepository) GetByID(id uuid.UUID) (*models.Position, error) {
	position := &models.Position{}
	query := `SELECT id, name, min_staff_per_branch, display_order, position_type, manpower_type, created_at 
	          FROM positions WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(
		&position.ID, &position.Name, &position.MinStaffPerBranch,
		&position.DisplayOrder, &position.PositionType, &position.ManpowerType, &position.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return position, err
}

func (r *positionRepository) Update(position *models.Position) error {
	query := `UPDATE positions SET name = $1, display_order = $2, position_type = $3, manpower_type = $4 WHERE id = $5`
	_, err := r.db.Exec(query, position.Name, position.DisplayOrder, position.PositionType, position.ManpowerType, position.ID)
	return err
}

func (r *positionRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM positions WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *positionRepository) List() ([]*models.Position, error) {
	query := `SELECT p.id, p.name, p.min_staff_per_branch, p.display_order, p.position_type, p.manpower_type, p.created_at,
	          COALESCE(COUNT(CASE WHEN s.staff_type = 'branch' THEN 1 END), 0) as branch_staff_count,
	          COALESCE(COUNT(CASE WHEN s.staff_type = 'rotation' THEN 1 END), 0) as rotation_staff_count
	          FROM positions p
	          LEFT JOIN staff s ON p.id = s.position_id
	          GROUP BY p.id, p.name, p.min_staff_per_branch, p.display_order, p.position_type, p.manpower_type, p.created_at
	          ORDER BY p.display_order, p.name`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var positions []*models.Position
	for rows.Next() {
		position := &models.Position{}
		var branchStaffCount int
		var rotationStaffCount int
		if err := rows.Scan(
			&position.ID, &position.Name, &position.MinStaffPerBranch,
			&position.DisplayOrder, &position.PositionType, &position.ManpowerType, &position.CreatedAt, &branchStaffCount, &rotationStaffCount,
		); err != nil {
			return nil, err
		}
		position.BranchStaffCount = &branchStaffCount
		position.RotationStaffCount = &rotationStaffCount
		positions = append(positions, position)
	}
	return positions, rows.Err()
}

// BranchRepository implementation
type branchRepository struct {
	db *sql.DB
}

func NewBranchRepository(db *sql.DB) interfaces.BranchRepository {
	return &branchRepository{db: db}
}

func (r *branchRepository) Create(branch *models.Branch) error {
	query := `INSERT INTO branches (id, name, code, area_manager_id, priority) 
	          VALUES ($1, $2, $3, $4, $5) RETURNING created_at, updated_at`
	return r.db.QueryRow(query, branch.ID, branch.Name, branch.Code,
		branch.AreaManagerID, branch.Priority).
		Scan(&branch.CreatedAt, &branch.UpdatedAt)
}

func (r *branchRepository) GetByID(id uuid.UUID) (*models.Branch, error) {
	branch := &models.Branch{}
	query := `SELECT id, name, code, area_manager_id, priority, created_at, updated_at 
	          FROM branches WHERE id = $1`
	var areaManagerID sql.NullString
	err := r.db.QueryRow(query, id).Scan(
		&branch.ID, &branch.Name, &branch.Code,
		&areaManagerID, &branch.Priority,
		&branch.CreatedAt, &branch.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if areaManagerID.Valid {
		amID, _ := uuid.Parse(areaManagerID.String)
		branch.AreaManagerID = &amID
	}
	return branch, nil
}

func (r *branchRepository) Update(branch *models.Branch) error {
	query := `UPDATE branches SET name = $1, code = $2, area_manager_id = $3, 
	          priority = $4, updated_at = CURRENT_TIMESTAMP WHERE id = $5`
	_, err := r.db.Exec(query, branch.Name, branch.Code,
		branch.AreaManagerID, branch.Priority, branch.ID)
	return err
}

func (r *branchRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM branches WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *branchRepository) List() ([]*models.Branch, error) {
	query := `SELECT id, name, code, area_manager_id, priority, created_at, updated_at 
	          FROM branches ORDER BY code`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var branches []*models.Branch
	for rows.Next() {
		branch := &models.Branch{}
		var areaManagerID sql.NullString
		if err := rows.Scan(
			&branch.ID, &branch.Name, &branch.Code,
			&areaManagerID, &branch.Priority,
			&branch.CreatedAt, &branch.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if areaManagerID.Valid {
			amID, _ := uuid.Parse(areaManagerID.String)
			branch.AreaManagerID = &amID
		}
		branches = append(branches, branch)
	}
	return branches, rows.Err()
}

func (r *branchRepository) GetByAreaManagerID(areaManagerID uuid.UUID) ([]*models.Branch, error) {
	query := `SELECT id, name, code, area_manager_id, priority, created_at, updated_at 
	          FROM branches WHERE area_manager_id = $1 ORDER BY code`
	rows, err := r.db.Query(query, areaManagerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var branches []*models.Branch
	for rows.Next() {
		branch := &models.Branch{}
		var areaManagerID sql.NullString
		if err := rows.Scan(
			&branch.ID, &branch.Name, &branch.Code,
			&areaManagerID, &branch.Priority,
			&branch.CreatedAt, &branch.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if areaManagerID.Valid {
			amID, _ := uuid.Parse(areaManagerID.String)
			branch.AreaManagerID = &amID
		}
		branches = append(branches, branch)
	}
	return branches, rows.Err()
}

// EffectiveBranchRepository implementation
type effectiveBranchRepository struct {
	db *sql.DB
}

func NewEffectiveBranchRepository(db *sql.DB) interfaces.EffectiveBranchRepository {
	return &effectiveBranchRepository{db: db}
}

func (r *effectiveBranchRepository) Create(eb *models.EffectiveBranch) error {
	// Set defaults if not provided
	commuteDuration := 300
	if eb.CommuteDurationMinutes != nil {
		commuteDuration = *eb.CommuteDurationMinutes
	}
	transitCount := 10
	if eb.TransitCount != nil {
		transitCount = *eb.TransitCount
	}
	travelCost := 1000.0
	if eb.TravelCost != nil {
		travelCost = *eb.TravelCost
	}

	query := `INSERT INTO effective_branches (id, rotation_staff_id, branch_id, level, commute_duration_minutes, transit_count, travel_cost) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING created_at`
	return r.db.QueryRow(query, eb.ID, eb.RotationStaffID, eb.BranchID, eb.Level, commuteDuration, transitCount, travelCost).
		Scan(&eb.CreatedAt)
}

func (r *effectiveBranchRepository) GetByID(id uuid.UUID) (*models.EffectiveBranch, error) {
	query := `SELECT id, rotation_staff_id, branch_id, level, commute_duration_minutes, transit_count, travel_cost, created_at 
	          FROM effective_branches WHERE id = $1`
	eb := &models.EffectiveBranch{}
	var commuteDuration, transitCount sql.NullInt64
	var travelCost sql.NullFloat64
	err := r.db.QueryRow(query, id).Scan(&eb.ID, &eb.RotationStaffID, &eb.BranchID, &eb.Level, &commuteDuration, &transitCount, &travelCost, &eb.CreatedAt)
	if err != nil {
		return nil, err
	}
	if commuteDuration.Valid {
		val := int(commuteDuration.Int64)
		eb.CommuteDurationMinutes = &val
	}
	if transitCount.Valid {
		val := int(transitCount.Int64)
		eb.TransitCount = &val
	}
	if travelCost.Valid {
		val := travelCost.Float64
		eb.TravelCost = &val
	}
	return eb, nil
}

func (r *effectiveBranchRepository) GetByRotationStaffID(rotationStaffID uuid.UUID) ([]*models.EffectiveBranch, error) {
	query := `SELECT id, rotation_staff_id, branch_id, level, commute_duration_minutes, transit_count, travel_cost, created_at 
	          FROM effective_branches WHERE rotation_staff_id = $1 ORDER BY level, created_at`
	rows, err := r.db.Query(query, rotationStaffID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ebs []*models.EffectiveBranch
	for rows.Next() {
		eb := &models.EffectiveBranch{}
		var commuteDuration, transitCount sql.NullInt64
		var travelCost sql.NullFloat64
		if err := rows.Scan(&eb.ID, &eb.RotationStaffID, &eb.BranchID, &eb.Level, &commuteDuration, &transitCount, &travelCost, &eb.CreatedAt); err != nil {
			return nil, err
		}
		if commuteDuration.Valid {
			val := int(commuteDuration.Int64)
			eb.CommuteDurationMinutes = &val
		}
		if transitCount.Valid {
			val := int(transitCount.Int64)
			eb.TransitCount = &val
		}
		if travelCost.Valid {
			val := travelCost.Float64
			eb.TravelCost = &val
		}
		ebs = append(ebs, eb)
	}
	return ebs, rows.Err()
}

func (r *effectiveBranchRepository) GetByBranchID(branchID uuid.UUID) ([]*models.EffectiveBranch, error) {
	query := `SELECT id, rotation_staff_id, branch_id, level, commute_duration_minutes, transit_count, travel_cost, created_at 
	          FROM effective_branches WHERE branch_id = $1 ORDER BY level, created_at`
	rows, err := r.db.Query(query, branchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ebs []*models.EffectiveBranch
	for rows.Next() {
		eb := &models.EffectiveBranch{}
		var commuteDuration, transitCount sql.NullInt64
		var travelCost sql.NullFloat64
		if err := rows.Scan(&eb.ID, &eb.RotationStaffID, &eb.BranchID, &eb.Level, &commuteDuration, &transitCount, &travelCost, &eb.CreatedAt); err != nil {
			return nil, err
		}
		if commuteDuration.Valid {
			val := int(commuteDuration.Int64)
			eb.CommuteDurationMinutes = &val
		}
		if transitCount.Valid {
			val := int(transitCount.Int64)
			eb.TransitCount = &val
		}
		if travelCost.Valid {
			val := travelCost.Float64
			eb.TravelCost = &val
		}
		ebs = append(ebs, eb)
	}
	return ebs, rows.Err()
}

func (r *effectiveBranchRepository) Update(eb *models.EffectiveBranch) error {
	// Set defaults if not provided
	commuteDuration := 300
	if eb.CommuteDurationMinutes != nil {
		commuteDuration = *eb.CommuteDurationMinutes
	}
	transitCount := 10
	if eb.TransitCount != nil {
		transitCount = *eb.TransitCount
	}
	travelCost := 1000.0
	if eb.TravelCost != nil {
		travelCost = *eb.TravelCost
	}

	query := `UPDATE effective_branches 
	          SET rotation_staff_id = $2, branch_id = $3, level = $4, 
	              commute_duration_minutes = $5, transit_count = $6, travel_cost = $7
	          WHERE id = $1`
	_, err := r.db.Exec(query, eb.ID, eb.RotationStaffID, eb.BranchID, eb.Level, commuteDuration, transitCount, travelCost)
	return err
}

func (r *effectiveBranchRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM effective_branches WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *effectiveBranchRepository) DeleteByRotationStaffID(rotationStaffID uuid.UUID) error {
	query := `DELETE FROM effective_branches WHERE rotation_staff_id = $1`
	_, err := r.db.Exec(query, rotationStaffID)
	return err
}

// RevenueRepository implementation
type revenueRepository struct {
	db *sql.DB
}

func NewRevenueRepository(db *sql.DB) interfaces.RevenueRepository {
	return &revenueRepository{db: db}
}

func (r *revenueRepository) Create(revenue *models.RevenueData) error {
	revenueSource := revenue.RevenueSource
	if revenueSource == "" {
		revenueSource = "branch" // Default
	}
	query := `INSERT INTO revenue_data (id, branch_id, date, expected_revenue, actual_revenue, revenue_source) 
	          VALUES ($1, $2, $3, $4, $5, $6) RETURNING created_at, updated_at`
	return r.db.QueryRow(query, revenue.ID, revenue.BranchID, revenue.Date,
		revenue.ExpectedRevenue, revenue.ActualRevenue, revenueSource).
		Scan(&revenue.CreatedAt, &revenue.UpdatedAt)
}

func (r *revenueRepository) GetByBranchID(branchID uuid.UUID, startDate, endDate time.Time) ([]*models.RevenueData, error) {
	query := `SELECT id, branch_id, date, expected_revenue, actual_revenue, COALESCE(revenue_source, 'branch') as revenue_source, created_at, updated_at 
	          FROM revenue_data WHERE branch_id = $1 AND date >= $2 AND date <= $3 ORDER BY date`
	rows, err := r.db.Query(query, branchID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var revenues []*models.RevenueData
	for rows.Next() {
		revenue := &models.RevenueData{}
		var actualRevenue sql.NullFloat64
		var revenueSource sql.NullString
		if err := rows.Scan(
			&revenue.ID, &revenue.BranchID, &revenue.Date,
			&revenue.ExpectedRevenue, &actualRevenue, &revenueSource,
			&revenue.CreatedAt, &revenue.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if actualRevenue.Valid {
			revenue.ActualRevenue = &actualRevenue.Float64
		}
		if revenueSource.Valid {
			revenue.RevenueSource = revenueSource.String
		} else {
			revenue.RevenueSource = "branch" // Default
		}
		revenues = append(revenues, revenue)
	}
	return revenues, rows.Err()
}

func (r *revenueRepository) GetByDate(date time.Time) ([]*models.RevenueData, error) {
	query := `SELECT id, branch_id, date, expected_revenue, actual_revenue, COALESCE(revenue_source, 'branch') as revenue_source, created_at, updated_at 
	          FROM revenue_data WHERE date = $1 ORDER BY branch_id`
	rows, err := r.db.Query(query, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var revenues []*models.RevenueData
	for rows.Next() {
		revenue := &models.RevenueData{}
		var actualRevenue sql.NullFloat64
		var revenueSource sql.NullString
		if err := rows.Scan(
			&revenue.ID, &revenue.BranchID, &revenue.Date,
			&revenue.ExpectedRevenue, &actualRevenue, &revenueSource,
			&revenue.CreatedAt, &revenue.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if actualRevenue.Valid {
			revenue.ActualRevenue = &actualRevenue.Float64
		}
		if revenueSource.Valid {
			revenue.RevenueSource = revenueSource.String
		} else {
			revenue.RevenueSource = "branch" // Default
		}
		revenues = append(revenues, revenue)
	}
	return revenues, rows.Err()
}

func (r *revenueRepository) Update(revenue *models.RevenueData) error {
	revenueSource := revenue.RevenueSource
	if revenueSource == "" {
		revenueSource = "branch" // Default
	}
	query := `UPDATE revenue_data SET expected_revenue = $1, actual_revenue = $2, revenue_source = $3,
	          updated_at = CURRENT_TIMESTAMP WHERE id = $4`
	_, err := r.db.Exec(query, revenue.ExpectedRevenue, revenue.ActualRevenue, revenueSource, revenue.ID)
	return err
}

// ScheduleRepository implementation
type scheduleRepository struct {
	db *sql.DB
}

func NewScheduleRepository(db *sql.DB) interfaces.ScheduleRepository {
	return &scheduleRepository{db: db}
}

func (r *scheduleRepository) Create(schedule *models.StaffSchedule) error {
	// Set default schedule_status if not provided
	if schedule.ScheduleStatus == "" {
		if schedule.IsWorkingDay {
			schedule.ScheduleStatus = models.ScheduleStatusWorking
		} else {
			schedule.ScheduleStatus = models.ScheduleStatusOff
		}
	}
	// Update is_working_day for backward compatibility
	schedule.IsWorkingDay = (schedule.ScheduleStatus == models.ScheduleStatusWorking)
	
	// Use UPSERT (ON CONFLICT DO UPDATE) to handle both create and update
	// If schedule already exists (staff_id, branch_id, date), update it instead of failing
	query := `INSERT INTO staff_schedules (id, staff_id, branch_id, date, schedule_status, is_working_day, created_by) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7)
	          ON CONFLICT (staff_id, branch_id, date) 
	          DO UPDATE SET 
	            schedule_status = EXCLUDED.schedule_status,
	            is_working_day = EXCLUDED.is_working_day
	          RETURNING id, created_at`
	var returnedID uuid.UUID
	err := r.db.QueryRow(query, schedule.ID, schedule.StaffID, schedule.BranchID,
		schedule.Date, schedule.ScheduleStatus, schedule.IsWorkingDay, schedule.CreatedBy).
		Scan(&returnedID, &schedule.CreatedAt)
	if err != nil {
		return err
	}
	// Update the schedule ID to the returned ID (existing ID if update, new ID if insert)
	schedule.ID = returnedID
	return nil
}

func (r *scheduleRepository) GetByBranchID(branchID uuid.UUID, startDate, endDate time.Time) ([]*models.StaffSchedule, error) {
	query := `SELECT id, staff_id, branch_id, date, schedule_status, is_working_day, created_by, created_at 
	          FROM staff_schedules WHERE branch_id = $1 AND date >= $2 AND date <= $3 ORDER BY date, staff_id`
	rows, err := r.db.Query(query, branchID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []*models.StaffSchedule
	for rows.Next() {
		schedule := &models.StaffSchedule{}
		var scheduleStatus sql.NullString
		if err := rows.Scan(
			&schedule.ID, &schedule.StaffID, &schedule.BranchID, &schedule.Date,
			&scheduleStatus, &schedule.IsWorkingDay, &schedule.CreatedBy, &schedule.CreatedAt,
		); err != nil {
			return nil, err
		}
		if scheduleStatus.Valid {
			schedule.ScheduleStatus = models.ScheduleStatus(scheduleStatus.String)
		} else {
			// Fallback for old data
			if schedule.IsWorkingDay {
				schedule.ScheduleStatus = models.ScheduleStatusWorking
			} else {
				schedule.ScheduleStatus = models.ScheduleStatusOff
			}
		}
		schedules = append(schedules, schedule)
	}
	return schedules, rows.Err()
}

func (r *scheduleRepository) GetByStaffID(staffID uuid.UUID, startDate, endDate time.Time) ([]*models.StaffSchedule, error) {
	query := `SELECT id, staff_id, branch_id, date, schedule_status, is_working_day, created_by, created_at 
	          FROM staff_schedules WHERE staff_id = $1 AND date >= $2 AND date <= $3 ORDER BY date`
	rows, err := r.db.Query(query, staffID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []*models.StaffSchedule
	for rows.Next() {
		schedule := &models.StaffSchedule{}
		var scheduleStatus sql.NullString
		if err := rows.Scan(
			&schedule.ID, &schedule.StaffID, &schedule.BranchID, &schedule.Date,
			&scheduleStatus, &schedule.IsWorkingDay, &schedule.CreatedBy, &schedule.CreatedAt,
		); err != nil {
			return nil, err
		}
		if scheduleStatus.Valid {
			schedule.ScheduleStatus = models.ScheduleStatus(scheduleStatus.String)
		} else {
			// Fallback for old data
			if schedule.IsWorkingDay {
				schedule.ScheduleStatus = models.ScheduleStatusWorking
			} else {
				schedule.ScheduleStatus = models.ScheduleStatusOff
			}
		}
		schedules = append(schedules, schedule)
	}
	return schedules, rows.Err()
}

func (r *scheduleRepository) Update(schedule *models.StaffSchedule) error {
	// Set default schedule_status if not provided
	if schedule.ScheduleStatus == "" {
		if schedule.IsWorkingDay {
			schedule.ScheduleStatus = models.ScheduleStatusWorking
		} else {
			schedule.ScheduleStatus = models.ScheduleStatusOff
		}
	}
	// Update is_working_day for backward compatibility
	schedule.IsWorkingDay = (schedule.ScheduleStatus == models.ScheduleStatusWorking)
	
	query := `UPDATE staff_schedules SET schedule_status = $1, is_working_day = $2 WHERE id = $3`
	_, err := r.db.Exec(query, schedule.ScheduleStatus, schedule.IsWorkingDay, schedule.ID)
	return err
}

func (r *scheduleRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM staff_schedules WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *scheduleRepository) DeleteByStaffID(staffID uuid.UUID) error {
	query := `DELETE FROM staff_schedules WHERE staff_id = $1`
	_, err := r.db.Exec(query, staffID)
	return err
}

func (r *scheduleRepository) GetMonthlyView(branchID uuid.UUID, year int, month int) ([]*models.StaffSchedule, error) {
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, -1)
	return r.GetByBranchID(branchID, startDate, endDate)
}

// RotationRepository implementation
type rotationRepository struct {
	db *sql.DB
}

func NewRotationRepository(db *sql.DB) interfaces.RotationRepository {
	return &rotationRepository{db: db}
}

func (r *rotationRepository) Create(assignment *models.RotationAssignment) error {
	query := `INSERT INTO rotation_assignments (id, rotation_staff_id, branch_id, date, assignment_level, assigned_by) 
	          VALUES ($1, $2, $3, $4, $5, $6) RETURNING created_at`
	return r.db.QueryRow(query, assignment.ID, assignment.RotationStaffID, assignment.BranchID,
		assignment.Date, assignment.AssignmentLevel, assignment.AssignedBy).
		Scan(&assignment.CreatedAt)
}

func (r *rotationRepository) GetByDate(date time.Time) ([]*models.RotationAssignment, error) {
	query := `SELECT id, rotation_staff_id, branch_id, date, assignment_level, assigned_by, created_at 
	          FROM rotation_assignments WHERE date = $1 ORDER BY branch_id, rotation_staff_id`
	rows, err := r.db.Query(query, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assignments []*models.RotationAssignment
	for rows.Next() {
		assignment := &models.RotationAssignment{}
		if err := rows.Scan(
			&assignment.ID, &assignment.RotationStaffID, &assignment.BranchID, &assignment.Date,
			&assignment.AssignmentLevel, &assignment.AssignedBy, &assignment.CreatedAt,
		); err != nil {
			return nil, err
		}
		assignments = append(assignments, assignment)
	}
	return assignments, rows.Err()
}

func (r *rotationRepository) GetByBranchID(branchID uuid.UUID, startDate, endDate time.Time) ([]*models.RotationAssignment, error) {
	query := `SELECT id, rotation_staff_id, branch_id, date, assignment_level, assigned_by, created_at 
	          FROM rotation_assignments WHERE branch_id = $1 AND date >= $2 AND date <= $3 ORDER BY date`
	rows, err := r.db.Query(query, branchID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assignments []*models.RotationAssignment
	for rows.Next() {
		assignment := &models.RotationAssignment{}
		if err := rows.Scan(
			&assignment.ID, &assignment.RotationStaffID, &assignment.BranchID, &assignment.Date,
			&assignment.AssignmentLevel, &assignment.AssignedBy, &assignment.CreatedAt,
		); err != nil {
			return nil, err
		}
		assignments = append(assignments, assignment)
	}
	return assignments, rows.Err()
}

func (r *rotationRepository) GetByRotationStaffID(rotationStaffID uuid.UUID, startDate, endDate time.Time) ([]*models.RotationAssignment, error) {
	query := `SELECT id, rotation_staff_id, branch_id, date, assignment_level, assigned_by, created_at 
	          FROM rotation_assignments WHERE rotation_staff_id = $1 AND date >= $2 AND date <= $3 ORDER BY date`
	rows, err := r.db.Query(query, rotationStaffID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assignments []*models.RotationAssignment
	for rows.Next() {
		assignment := &models.RotationAssignment{}
		if err := rows.Scan(
			&assignment.ID, &assignment.RotationStaffID, &assignment.BranchID, &assignment.Date,
			&assignment.AssignmentLevel, &assignment.AssignedBy, &assignment.CreatedAt,
		); err != nil {
			return nil, err
		}
		assignments = append(assignments, assignment)
	}
	return assignments, rows.Err()
}

func (r *rotationRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM rotation_assignments WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *rotationRepository) DeleteByRotationStaffID(rotationStaffID uuid.UUID) error {
	query := `DELETE FROM rotation_assignments WHERE rotation_staff_id = $1`
	_, err := r.db.Exec(query, rotationStaffID)
	return err
}

func (r *rotationRepository) GetAssignments(filters interfaces.RotationFilters) ([]*models.RotationAssignment, error) {
	query := `SELECT id, rotation_staff_id, branch_id, date, assignment_level, assigned_by, created_at 
	          FROM rotation_assignments WHERE 1=1`
	args := []interface{}{}
	argPos := 1

	if filters.BranchID != nil {
		query += ` AND branch_id = $` + strconv.Itoa(argPos)
		args = append(args, *filters.BranchID)
		argPos++
	}
	if filters.RotationStaffID != nil {
		query += ` AND rotation_staff_id = $` + strconv.Itoa(argPos)
		args = append(args, *filters.RotationStaffID)
		argPos++
	}
	if filters.StartDate != nil {
		query += ` AND date >= $` + strconv.Itoa(argPos)
		args = append(args, *filters.StartDate)
		argPos++
	}
	if filters.EndDate != nil {
		query += ` AND date <= $` + strconv.Itoa(argPos)
		args = append(args, *filters.EndDate)
		argPos++
	}

	query += ` ORDER BY date, branch_id`

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assignments []*models.RotationAssignment
	for rows.Next() {
		assignment := &models.RotationAssignment{}
		if err := rows.Scan(
			&assignment.ID, &assignment.RotationStaffID, &assignment.BranchID, &assignment.Date,
			&assignment.AssignmentLevel, &assignment.AssignedBy, &assignment.CreatedAt,
		); err != nil {
			return nil, err
		}
		assignments = append(assignments, assignment)
	}
	return assignments, rows.Err()
}

// SettingsRepository implementation
type settingsRepository struct {
	db *sql.DB
}

func NewSettingsRepository(db *sql.DB) interfaces.SettingsRepository {
	return &settingsRepository{db: db}
}

func (r *settingsRepository) GetAll() ([]*models.SystemSetting, error) {
	query := `SELECT id, key, value, description, updated_at FROM system_settings ORDER BY key`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var settings []*models.SystemSetting
	for rows.Next() {
		setting := &models.SystemSetting{}
		if err := rows.Scan(&setting.ID, &setting.Key, &setting.Value, &setting.Description, &setting.UpdatedAt); err != nil {
			return nil, err
		}
		settings = append(settings, setting)
	}
	return settings, rows.Err()
}

func (r *settingsRepository) GetByKey(key string) (*models.SystemSetting, error) {
	setting := &models.SystemSetting{}
	query := `SELECT id, key, value, description, updated_at FROM system_settings WHERE key = $1`
	err := r.db.QueryRow(query, key).Scan(&setting.ID, &setting.Key, &setting.Value, &setting.Description, &setting.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return setting, err
}

func (r *settingsRepository) Update(setting *models.SystemSetting) error {
	query := `UPDATE system_settings SET value = $1, description = $2, updated_at = CURRENT_TIMESTAMP WHERE key = $3`
	_, err := r.db.Exec(query, setting.Value, setting.Description, setting.Key)
	return err
}

func (r *settingsRepository) Create(setting *models.SystemSetting) error {
	query := `INSERT INTO system_settings (id, key, value, description) VALUES ($1, $2, $3, $4) RETURNING updated_at`
	return r.db.QueryRow(query, setting.ID, setting.Key, setting.Value, setting.Description).Scan(&setting.UpdatedAt)
}

// AllocationRuleRepository implementation
type allocationRuleRepository struct {
	db *sql.DB
}

func NewAllocationRuleRepository(db *sql.DB) interfaces.AllocationRuleRepository {
	return &allocationRuleRepository{db: db}
}

func (r *allocationRuleRepository) Create(rule *models.StaffAllocationRule) error {
	query := `INSERT INTO staff_allocation_rules (id, position_id, min_staff, revenue_threshold, staff_count_formula) 
	          VALUES ($1, $2, $3, $4, $5) RETURNING created_at, updated_at`
	return r.db.QueryRow(query, rule.ID, rule.PositionID, rule.MinStaff,
		rule.RevenueThreshold, rule.StaffCountFormula).
		Scan(&rule.CreatedAt, &rule.UpdatedAt)
}

func (r *allocationRuleRepository) GetByPositionID(positionID uuid.UUID) (*models.StaffAllocationRule, error) {
	rule := &models.StaffAllocationRule{}
	query := `SELECT id, position_id, min_staff, revenue_threshold, staff_count_formula, created_at, updated_at 
	          FROM staff_allocation_rules WHERE position_id = $1`
	err := r.db.QueryRow(query, positionID).Scan(
		&rule.ID, &rule.PositionID, &rule.MinStaff, &rule.RevenueThreshold,
		&rule.StaffCountFormula, &rule.CreatedAt, &rule.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return rule, err
}

func (r *allocationRuleRepository) Update(rule *models.StaffAllocationRule) error {
	query := `UPDATE staff_allocation_rules SET min_staff = $1, revenue_threshold = $2, 
	          staff_count_formula = $3, updated_at = CURRENT_TIMESTAMP WHERE id = $4`
	_, err := r.db.Exec(query, rule.MinStaff, rule.RevenueThreshold, rule.StaffCountFormula, rule.ID)
	return err
}

func (r *allocationRuleRepository) List() ([]*models.StaffAllocationRule, error) {
	query := `SELECT id, position_id, min_staff, revenue_threshold, staff_count_formula, created_at, updated_at 
	          FROM staff_allocation_rules ORDER BY position_id`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []*models.StaffAllocationRule
	for rows.Next() {
		rule := &models.StaffAllocationRule{}
		if err := rows.Scan(
			&rule.ID, &rule.PositionID, &rule.MinStaff, &rule.RevenueThreshold,
			&rule.StaffCountFormula, &rule.CreatedAt, &rule.UpdatedAt,
		); err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}
	return rules, rows.Err()
}

// AreaOfOperationRepository implementation
type areaOfOperationRepository struct {
	db *sql.DB
}

func NewAreaOfOperationRepository(db *sql.DB) interfaces.AreaOfOperationRepository {
	return &areaOfOperationRepository{db: db}
}

func (r *areaOfOperationRepository) Create(aoo *models.AreaOfOperation) error {
	query := `INSERT INTO areas_of_operation (id, name, code, description, is_active) 
	          VALUES ($1, $2, $3, $4, $5) RETURNING created_at, updated_at`
	return r.db.QueryRow(query, aoo.ID, aoo.Name, aoo.Code, aoo.Description, aoo.IsActive).
		Scan(&aoo.CreatedAt, &aoo.UpdatedAt)
}

func (r *areaOfOperationRepository) GetByID(id uuid.UUID) (*models.AreaOfOperation, error) {
	query := `SELECT id, name, code, description, is_active, created_at, updated_at 
	          FROM areas_of_operation WHERE id = $1`
	aoo := &models.AreaOfOperation{}
	err := r.db.QueryRow(query, id).Scan(
		&aoo.ID, &aoo.Name, &aoo.Code, &aoo.Description, &aoo.IsActive,
		&aoo.CreatedAt, &aoo.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return aoo, nil
}

func (r *areaOfOperationRepository) GetByCode(code string) (*models.AreaOfOperation, error) {
	query := `SELECT id, name, code, description, is_active, created_at, updated_at 
	          FROM areas_of_operation WHERE code = $1`
	aoo := &models.AreaOfOperation{}
	err := r.db.QueryRow(query, code).Scan(
		&aoo.ID, &aoo.Name, &aoo.Code, &aoo.Description, &aoo.IsActive,
		&aoo.CreatedAt, &aoo.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return aoo, nil
}

func (r *areaOfOperationRepository) Update(aoo *models.AreaOfOperation) error {
	query := `UPDATE areas_of_operation 
	          SET name = $1, code = $2, description = $3, is_active = $4, updated_at = CURRENT_TIMESTAMP 
	          WHERE id = $5 RETURNING updated_at`
	return r.db.QueryRow(query, aoo.Name, aoo.Code, aoo.Description, aoo.IsActive, aoo.ID).
		Scan(&aoo.UpdatedAt)
}

func (r *areaOfOperationRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM areas_of_operation WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *areaOfOperationRepository) List(includeInactive bool) ([]*models.AreaOfOperation, error) {
	var query string
	if includeInactive {
		query = `SELECT id, name, code, description, is_active, created_at, updated_at 
		         FROM areas_of_operation ORDER BY name`
	} else {
		query = `SELECT id, name, code, description, is_active, created_at, updated_at 
		         FROM areas_of_operation WHERE is_active = true ORDER BY name`
	}
	
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var areas []*models.AreaOfOperation
	for rows.Next() {
		aoo := &models.AreaOfOperation{}
		if err := rows.Scan(
			&aoo.ID, &aoo.Name, &aoo.Code, &aoo.Description, &aoo.IsActive,
			&aoo.CreatedAt, &aoo.UpdatedAt,
		); err != nil {
			return nil, err
		}
		areas = append(areas, aoo)
	}
	return areas, rows.Err()
}

func (r *areaOfOperationRepository) AddZone(areaOfOperationID, zoneID uuid.UUID) error {
	query := `INSERT INTO area_of_operation_zones (id, area_of_operation_id, zone_id) 
	          VALUES (gen_random_uuid(), $1, $2) 
	          ON CONFLICT (area_of_operation_id, zone_id) DO NOTHING`
	_, err := r.db.Exec(query, areaOfOperationID, zoneID)
	return err
}

func (r *areaOfOperationRepository) RemoveZone(areaOfOperationID, zoneID uuid.UUID) error {
	query := `DELETE FROM area_of_operation_zones WHERE area_of_operation_id = $1 AND zone_id = $2`
	_, err := r.db.Exec(query, areaOfOperationID, zoneID)
	return err
}

func (r *areaOfOperationRepository) GetZones(areaOfOperationID uuid.UUID) ([]*models.Zone, error) {
	query := `SELECT z.id, z.name, z.code, z.description, z.is_active, z.created_at, z.updated_at
	          FROM zones z
	          INNER JOIN area_of_operation_zones aoz ON z.id = aoz.zone_id
	          WHERE aoz.area_of_operation_id = $1
	          ORDER BY z.name`
	
	rows, err := r.db.Query(query, areaOfOperationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var zones []*models.Zone
	for rows.Next() {
		zone := &models.Zone{}
		if err := rows.Scan(
			&zone.ID, &zone.Name, &zone.Code, &zone.Description, &zone.IsActive,
			&zone.CreatedAt, &zone.UpdatedAt,
		); err != nil {
			return nil, err
		}
		zones = append(zones, zone)
	}
	return zones, rows.Err()
}

func (r *areaOfOperationRepository) AddBranch(areaOfOperationID, branchID uuid.UUID) error {
	query := `INSERT INTO area_of_operation_branches (id, area_of_operation_id, branch_id) 
	          VALUES (gen_random_uuid(), $1, $2) 
	          ON CONFLICT (area_of_operation_id, branch_id) DO NOTHING`
	_, err := r.db.Exec(query, areaOfOperationID, branchID)
	return err
}

func (r *areaOfOperationRepository) RemoveBranch(areaOfOperationID, branchID uuid.UUID) error {
	query := `DELETE FROM area_of_operation_branches WHERE area_of_operation_id = $1 AND branch_id = $2`
	_, err := r.db.Exec(query, areaOfOperationID, branchID)
	return err
}

func (r *areaOfOperationRepository) GetBranches(areaOfOperationID uuid.UUID) ([]*models.Branch, error) {
	// Get individual branches (not from zones)
	query := `SELECT b.id, b.name, b.code, b.area_manager_id, b.priority, b.created_at, b.updated_at
	          FROM branches b
	          INNER JOIN area_of_operation_branches aob ON b.id = aob.branch_id
	          WHERE aob.area_of_operation_id = $1
	          ORDER BY b.name`
	
	rows, err := r.db.Query(query, areaOfOperationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var branches []*models.Branch
	for rows.Next() {
		branch := &models.Branch{}
		var areaManagerID sql.NullString
		err := rows.Scan(
			&branch.ID, &branch.Name, &branch.Code, &areaManagerID, &branch.Priority,
			&branch.CreatedAt, &branch.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		if areaManagerID.Valid {
			if id, err := uuid.Parse(areaManagerID.String); err == nil {
				branch.AreaManagerID = &id
			}
		}
		branches = append(branches, branch)
	}
	return branches, rows.Err()
}

func (r *areaOfOperationRepository) GetAllBranches(areaOfOperationID uuid.UUID) ([]*models.Branch, error) {
	// Get all branches: from zones + individual branches
	// Using UNION to combine branches from zones and individual branches
	query := `
		SELECT DISTINCT b.id, b.name, b.code, b.area_manager_id, b.priority, b.created_at, b.updated_at
		FROM branches b
		WHERE b.id IN (
			-- Branches from zones
			SELECT zb.branch_id
			FROM zone_branches zb
			INNER JOIN area_of_operation_zones aoz ON zb.zone_id = aoz.zone_id
			WHERE aoz.area_of_operation_id = $1
			UNION
			-- Individual branches
			SELECT aob.branch_id
			FROM area_of_operation_branches aob
			WHERE aob.area_of_operation_id = $1
		)
		ORDER BY b.name`
	
	rows, err := r.db.Query(query, areaOfOperationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var branches []*models.Branch
	for rows.Next() {
		branch := &models.Branch{}
		var areaManagerID sql.NullString
		err := rows.Scan(
			&branch.ID, &branch.Name, &branch.Code, &areaManagerID, &branch.Priority,
			&branch.CreatedAt, &branch.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		if areaManagerID.Valid {
			if id, err := uuid.Parse(areaManagerID.String); err == nil {
				branch.AreaManagerID = &id
			}
		}
		branches = append(branches, branch)
	}
	return branches, rows.Err()
}

// ZoneRepository implementation
type zoneRepository struct {
	db *sql.DB
}

func NewZoneRepository(db *sql.DB) interfaces.ZoneRepository {
	return &zoneRepository{db: db}
}

func (r *zoneRepository) Create(zone *models.Zone) error {
	query := `INSERT INTO zones (id, name, code, description, is_active) 
	          VALUES ($1, $2, $3, $4, $5) RETURNING created_at, updated_at`
	return r.db.QueryRow(query, zone.ID, zone.Name, zone.Code, zone.Description, zone.IsActive).
		Scan(&zone.CreatedAt, &zone.UpdatedAt)
}

func (r *zoneRepository) GetByID(id uuid.UUID) (*models.Zone, error) {
	query := `SELECT id, name, code, description, is_active, created_at, updated_at 
	          FROM zones WHERE id = $1`
	zone := &models.Zone{}
	err := r.db.QueryRow(query, id).Scan(
		&zone.ID, &zone.Name, &zone.Code, &zone.Description, &zone.IsActive,
		&zone.CreatedAt, &zone.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return zone, nil
}

func (r *zoneRepository) GetByCode(code string) (*models.Zone, error) {
	query := `SELECT id, name, code, description, is_active, created_at, updated_at 
	          FROM zones WHERE code = $1`
	zone := &models.Zone{}
	err := r.db.QueryRow(query, code).Scan(
		&zone.ID, &zone.Name, &zone.Code, &zone.Description, &zone.IsActive,
		&zone.CreatedAt, &zone.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return zone, nil
}

func (r *zoneRepository) Update(zone *models.Zone) error {
	query := `UPDATE zones 
	          SET name = $1, code = $2, description = $3, is_active = $4, updated_at = CURRENT_TIMESTAMP 
	          WHERE id = $5 RETURNING updated_at`
	return r.db.QueryRow(query, zone.Name, zone.Code, zone.Description, zone.IsActive, zone.ID).
		Scan(&zone.UpdatedAt)
}

func (r *zoneRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM zones WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *zoneRepository) List(includeInactive bool) ([]*models.Zone, error) {
	var query string
	if includeInactive {
		query = `SELECT id, name, code, description, is_active, created_at, updated_at 
		         FROM zones ORDER BY name`
	} else {
		query = `SELECT id, name, code, description, is_active, created_at, updated_at 
		         FROM zones WHERE is_active = true ORDER BY name`
	}
	
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var zones []*models.Zone
	for rows.Next() {
		zone := &models.Zone{}
		if err := rows.Scan(
			&zone.ID, &zone.Name, &zone.Code, &zone.Description, &zone.IsActive,
			&zone.CreatedAt, &zone.UpdatedAt,
		); err != nil {
			return nil, err
		}
		zones = append(zones, zone)
	}
	return zones, rows.Err()
}

func (r *zoneRepository) AddBranch(zoneID, branchID uuid.UUID) error {
	query := `INSERT INTO zone_branches (id, zone_id, branch_id) 
	          VALUES (gen_random_uuid(), $1, $2) 
	          ON CONFLICT (zone_id, branch_id) DO NOTHING`
	_, err := r.db.Exec(query, zoneID, branchID)
	return err
}

func (r *zoneRepository) RemoveBranch(zoneID, branchID uuid.UUID) error {
	query := `DELETE FROM zone_branches WHERE zone_id = $1 AND branch_id = $2`
	_, err := r.db.Exec(query, zoneID, branchID)
	return err
}

func (r *zoneRepository) GetBranches(zoneID uuid.UUID) ([]*models.Branch, error) {
	query := `SELECT b.id, b.name, b.code, b.area_manager_id, b.priority, b.created_at, b.updated_at
	          FROM branches b
	          INNER JOIN zone_branches zb ON b.id = zb.branch_id
	          WHERE zb.zone_id = $1
	          ORDER BY b.name`
	
	rows, err := r.db.Query(query, zoneID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var branches []*models.Branch
	for rows.Next() {
		branch := &models.Branch{}
		var areaManagerID sql.NullString
		err := rows.Scan(
			&branch.ID, &branch.Name, &branch.Code, &areaManagerID, &branch.Priority,
			&branch.CreatedAt, &branch.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		if areaManagerID.Valid {
			if id, err := uuid.Parse(areaManagerID.String); err == nil {
				branch.AreaManagerID = &id
			}
		}
		branches = append(branches, branch)
	}
	return branches, rows.Err()
}

func (r *zoneRepository) BulkUpdateBranches(zoneID uuid.UUID, branchIDs []uuid.UUID) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete all existing branches for this zone
	deleteQuery := `DELETE FROM zone_branches WHERE zone_id = $1`
	if _, err := tx.Exec(deleteQuery, zoneID); err != nil {
		return err
	}

	// Insert new branches
	insertQuery := `INSERT INTO zone_branches (id, zone_id, branch_id) 
	                VALUES (gen_random_uuid(), $1, $2)`
	for _, branchID := range branchIDs {
		if _, err := tx.Exec(insertQuery, zoneID, branchID); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// AllocationCriteriaRepository implementation
type allocationCriteriaRepository struct {
	db *sql.DB
}

func NewAllocationCriteriaRepository(db *sql.DB) interfaces.AllocationCriteriaRepository {
	return &allocationCriteriaRepository{db: db}
}

func (r *allocationCriteriaRepository) Create(criteria *models.AllocationCriteria) error {
	query := `INSERT INTO allocation_criteria (id, pillar, type, weight, is_active, description, config, created_by) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING created_at, updated_at`
	return r.db.QueryRow(query, criteria.ID, criteria.Pillar, criteria.Type, criteria.Weight, criteria.IsActive,
		criteria.Description, criteria.Config, criteria.CreatedBy).Scan(&criteria.CreatedAt, &criteria.UpdatedAt)
}

func (r *allocationCriteriaRepository) GetByID(id uuid.UUID) (*models.AllocationCriteria, error) {
	query := `SELECT id, pillar, type, weight, is_active, description, config, created_by, created_at, updated_at 
	          FROM allocation_criteria WHERE id = $1`
	criteria := &models.AllocationCriteria{}
	err := r.db.QueryRow(query, id).Scan(
		&criteria.ID, &criteria.Pillar, &criteria.Type, &criteria.Weight, &criteria.IsActive,
		&criteria.Description, &criteria.Config, &criteria.CreatedBy, &criteria.CreatedAt, &criteria.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return criteria, err
}

func (r *allocationCriteriaRepository) Update(criteria *models.AllocationCriteria) error {
	query := `UPDATE allocation_criteria 
	          SET pillar = $1, type = $2, weight = $3, is_active = $4, description = $5, config = $6, updated_at = CURRENT_TIMESTAMP 
	          WHERE id = $7 RETURNING updated_at`
	return r.db.QueryRow(query, criteria.Pillar, criteria.Type, criteria.Weight, criteria.IsActive,
		criteria.Description, criteria.Config, criteria.ID).Scan(&criteria.UpdatedAt)
}

func (r *allocationCriteriaRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM allocation_criteria WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *allocationCriteriaRepository) List(filters interfaces.AllocationCriteriaFilters) ([]*models.AllocationCriteria, error) {
	query := `SELECT id, pillar, type, weight, is_active, description, config, created_by, created_at, updated_at 
	          FROM allocation_criteria WHERE 1=1`
	args := []interface{}{}
	argPos := 1

	if filters.Pillar != nil {
		query += fmt.Sprintf(" AND pillar = $%d", argPos)
		args = append(args, *filters.Pillar)
		argPos++
	}
	if filters.Type != nil {
		query += fmt.Sprintf(" AND type = $%d", argPos)
		args = append(args, *filters.Type)
		argPos++
	}
	if filters.IsActive != nil {
		query += fmt.Sprintf(" AND is_active = $%d", argPos)
		args = append(args, *filters.IsActive)
		argPos++
	}

	query += " ORDER BY pillar, type"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var criteriaList []*models.AllocationCriteria
	for rows.Next() {
		criteria := &models.AllocationCriteria{}
		if err := rows.Scan(
			&criteria.ID, &criteria.Pillar, &criteria.Type, &criteria.Weight, &criteria.IsActive,
			&criteria.Description, &criteria.Config, &criteria.CreatedBy, &criteria.CreatedAt, &criteria.UpdatedAt,
		); err != nil {
			return nil, err
		}
		criteriaList = append(criteriaList, criteria)
	}
	return criteriaList, rows.Err()
}

func (r *allocationCriteriaRepository) GetByPillar(pillar models.CriteriaPillar) ([]*models.AllocationCriteria, error) {
	filters := interfaces.AllocationCriteriaFilters{Pillar: &pillar, IsActive: &[]bool{true}[0]}
	return r.List(filters)
}

// PositionQuotaRepository implementation
type positionQuotaRepository struct {
	db *sql.DB
}

func NewPositionQuotaRepository(db *sql.DB) interfaces.PositionQuotaRepository {
	return &positionQuotaRepository{db: db}
}

func (r *positionQuotaRepository) Create(quota *models.PositionQuota) error {
	query := `INSERT INTO position_quotas (id, branch_id, position_id, designated_quota, minimum_required, is_active, created_by) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING created_at, updated_at`
	return r.db.QueryRow(query, quota.ID, quota.BranchID, quota.PositionID, quota.DesignatedQuota,
		quota.MinimumRequired, quota.IsActive, quota.CreatedBy).Scan(&quota.CreatedAt, &quota.UpdatedAt)
}

func (r *positionQuotaRepository) GetByID(id uuid.UUID) (*models.PositionQuota, error) {
	query := `SELECT id, branch_id, position_id, designated_quota, minimum_required, is_active, created_by, created_at, updated_at 
	          FROM position_quotas WHERE id = $1`
	quota := &models.PositionQuota{}
	err := r.db.QueryRow(query, id).Scan(
		&quota.ID, &quota.BranchID, &quota.PositionID, &quota.DesignatedQuota,
		&quota.MinimumRequired, &quota.IsActive, &quota.CreatedBy, &quota.CreatedAt, &quota.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return quota, err
}

func (r *positionQuotaRepository) GetByBranchID(branchID uuid.UUID) ([]*models.PositionQuota, error) {
	query := `SELECT id, branch_id, position_id, designated_quota, minimum_required, is_active, created_by, created_at, updated_at 
	          FROM position_quotas WHERE branch_id = $1 ORDER BY position_id`
	rows, err := r.db.Query(query, branchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var quotas []*models.PositionQuota
	for rows.Next() {
		quota := &models.PositionQuota{}
		if err := rows.Scan(
			&quota.ID, &quota.BranchID, &quota.PositionID, &quota.DesignatedQuota,
			&quota.MinimumRequired, &quota.IsActive, &quota.CreatedBy, &quota.CreatedAt, &quota.UpdatedAt,
		); err != nil {
			return nil, err
		}
		quotas = append(quotas, quota)
	}
	return quotas, rows.Err()
}

func (r *positionQuotaRepository) GetByBranchAndPosition(branchID, positionID uuid.UUID) (*models.PositionQuota, error) {
	query := `SELECT id, branch_id, position_id, designated_quota, minimum_required, is_active, created_by, created_at, updated_at 
	          FROM position_quotas WHERE branch_id = $1 AND position_id = $2`
	quota := &models.PositionQuota{}
	err := r.db.QueryRow(query, branchID, positionID).Scan(
		&quota.ID, &quota.BranchID, &quota.PositionID, &quota.DesignatedQuota,
		&quota.MinimumRequired, &quota.IsActive, &quota.CreatedBy, &quota.CreatedAt, &quota.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return quota, err
}

func (r *positionQuotaRepository) Update(quota *models.PositionQuota) error {
	query := `UPDATE position_quotas 
	          SET designated_quota = $1, minimum_required = $2, is_active = $3, updated_at = CURRENT_TIMESTAMP 
	          WHERE id = $4 RETURNING updated_at`
	return r.db.QueryRow(query, quota.DesignatedQuota, quota.MinimumRequired, quota.IsActive, quota.ID).Scan(&quota.UpdatedAt)
}

func (r *positionQuotaRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM position_quotas WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *positionQuotaRepository) List(filters interfaces.PositionQuotaFilters) ([]*models.PositionQuota, error) {
	query := `SELECT id, branch_id, position_id, designated_quota, minimum_required, is_active, created_by, created_at, updated_at 
	          FROM position_quotas WHERE 1=1`
	args := []interface{}{}
	argPos := 1

	if filters.BranchID != nil {
		query += fmt.Sprintf(" AND branch_id = $%d", argPos)
		args = append(args, *filters.BranchID)
		argPos++
	}
	if filters.PositionID != nil {
		query += fmt.Sprintf(" AND position_id = $%d", argPos)
		args = append(args, *filters.PositionID)
		argPos++
	}
	if filters.IsActive != nil {
		query += fmt.Sprintf(" AND is_active = $%d", argPos)
		args = append(args, *filters.IsActive)
		argPos++
	}

	query += " ORDER BY branch_id, position_id"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var quotas []*models.PositionQuota
	for rows.Next() {
		quota := &models.PositionQuota{}
		if err := rows.Scan(
			&quota.ID, &quota.BranchID, &quota.PositionID, &quota.DesignatedQuota,
			&quota.MinimumRequired, &quota.IsActive, &quota.CreatedBy, &quota.CreatedAt, &quota.UpdatedAt,
		); err != nil {
			return nil, err
		}
		quotas = append(quotas, quota)
	}
	return quotas, rows.Err()
}

// DoctorAssignmentRepository implementation
type doctorAssignmentRepository struct {
	db *sql.DB
}

func NewDoctorAssignmentRepository(db *sql.DB) interfaces.DoctorAssignmentRepository {
	return &doctorAssignmentRepository{db: db}
}

func (r *doctorAssignmentRepository) Create(assignment *models.DoctorAssignment) error {
	if assignment.ID == uuid.Nil {
		assignment.ID = uuid.New()
	}
	query := `INSERT INTO doctor_assignments (id, doctor_id, branch_id, date, expected_revenue, created_by) 
	          VALUES ($1, $2, $3, $4, $5, $6) RETURNING created_at, updated_at`
	return r.db.QueryRow(query, assignment.ID, assignment.DoctorID,
		assignment.BranchID, assignment.Date, assignment.ExpectedRevenue, assignment.CreatedBy).
		Scan(&assignment.CreatedAt, &assignment.UpdatedAt)
}

func (r *doctorAssignmentRepository) GetByID(id uuid.UUID) (*models.DoctorAssignment, error) {
	query := `SELECT da.id, da.doctor_id, COALESCE(d.name, '') as doctor_name, COALESCE(d.code, '') as doctor_code, 
	          da.branch_id, da.date, da.expected_revenue, da.created_by, da.created_at, da.updated_at 
	          FROM doctor_assignments da
	          LEFT JOIN doctors d ON da.doctor_id = d.id
	          WHERE da.id = $1`
	assignment := &models.DoctorAssignment{}
	var doctorName, doctorCode sql.NullString
	err := r.db.QueryRow(query, id).Scan(
		&assignment.ID, &assignment.DoctorID, &doctorName, &doctorCode,
		&assignment.BranchID, &assignment.Date, &assignment.ExpectedRevenue, &assignment.CreatedBy,
		&assignment.CreatedAt, &assignment.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if doctorName.Valid {
		assignment.DoctorName = doctorName.String
	}
	if doctorCode.Valid {
		assignment.DoctorCode = doctorCode.String
	}
	return assignment, nil
}

func (r *doctorAssignmentRepository) GetByBranchID(branchID uuid.UUID, startDate, endDate time.Time) ([]*models.DoctorAssignment, error) {
	query := `SELECT id, doctor_id, doctor_name, doctor_code, branch_id, date, expected_revenue, created_by, created_at, updated_at 
	          FROM doctor_assignments WHERE branch_id = $1 AND date >= $2 AND date <= $3 ORDER BY date, doctor_name`
	rows, err := r.db.Query(query, branchID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assignments []*models.DoctorAssignment
	for rows.Next() {
		assignment := &models.DoctorAssignment{}
		if err := rows.Scan(
			&assignment.ID, &assignment.DoctorID, &assignment.DoctorName, &assignment.DoctorCode,
			&assignment.BranchID, &assignment.Date, &assignment.ExpectedRevenue, &assignment.CreatedBy,
			&assignment.CreatedAt, &assignment.UpdatedAt,
		); err != nil {
			return nil, err
		}
		assignments = append(assignments, assignment)
	}
	return assignments, rows.Err()
}

func (r *doctorAssignmentRepository) GetByDate(date time.Time) ([]*models.DoctorAssignment, error) {
	query := `SELECT da.id, da.doctor_id, COALESCE(d.name, '') as doctor_name, COALESCE(d.code, '') as doctor_code, 
	          da.branch_id, da.date, da.expected_revenue, da.created_by, da.created_at, da.updated_at 
	          FROM doctor_assignments da
	          LEFT JOIN doctors d ON da.doctor_id = d.id
	          WHERE da.date = $1 ORDER BY da.branch_id, d.name`
	rows, err := r.db.Query(query, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assignments []*models.DoctorAssignment
	for rows.Next() {
		assignment := &models.DoctorAssignment{}
		var doctorName, doctorCode sql.NullString
		if err := rows.Scan(
			&assignment.ID, &assignment.DoctorID, &doctorName, &doctorCode,
			&assignment.BranchID, &assignment.Date, &assignment.ExpectedRevenue, &assignment.CreatedBy,
			&assignment.CreatedAt, &assignment.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if doctorName.Valid {
			assignment.DoctorName = doctorName.String
		}
		if doctorCode.Valid {
			assignment.DoctorCode = doctorCode.String
		}
		assignments = append(assignments, assignment)
	}
	return assignments, rows.Err()
}

func (r *doctorAssignmentRepository) GetByDoctorID(doctorID uuid.UUID, startDate, endDate time.Time) ([]*models.DoctorAssignment, error) {
	query := `SELECT da.id, da.doctor_id, COALESCE(d.name, '') as doctor_name, COALESCE(d.code, '') as doctor_code, 
	          da.branch_id, da.date, da.expected_revenue, da.created_by, da.created_at, da.updated_at 
	          FROM doctor_assignments da
	          LEFT JOIN doctors d ON da.doctor_id = d.id
	          WHERE da.doctor_id = $1 AND da.date >= $2 AND da.date <= $3 ORDER BY da.date, da.branch_id`
	rows, err := r.db.Query(query, doctorID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assignments []*models.DoctorAssignment
	for rows.Next() {
		assignment := &models.DoctorAssignment{}
		var doctorName, doctorCode sql.NullString
		if err := rows.Scan(
			&assignment.ID, &assignment.DoctorID, &doctorName, &doctorCode,
			&assignment.BranchID, &assignment.Date, &assignment.ExpectedRevenue, &assignment.CreatedBy,
			&assignment.CreatedAt, &assignment.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if doctorName.Valid {
			assignment.DoctorName = doctorName.String
		}
		if doctorCode.Valid {
			assignment.DoctorCode = doctorCode.String
		}
		assignments = append(assignments, assignment)
	}
	return assignments, rows.Err()
}

func (r *doctorAssignmentRepository) Update(assignment *models.DoctorAssignment) error {
	query := `UPDATE doctor_assignments 
	          SET expected_revenue = $1, updated_at = CURRENT_TIMESTAMP 
	          WHERE id = $2 RETURNING updated_at`
	return r.db.QueryRow(query, assignment.ExpectedRevenue, assignment.ID).
		Scan(&assignment.UpdatedAt)
}

func (r *doctorAssignmentRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM doctor_assignments WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *doctorAssignmentRepository) GetDoctorCountByBranch(branchID uuid.UUID, date time.Time) (int, error) {
	query := `SELECT COUNT(DISTINCT doctor_id) FROM doctor_assignments WHERE branch_id = $1 AND date = $2`
	var count int
	err := r.db.QueryRow(query, branchID, date).Scan(&count)
	return count, err
}

func (r *doctorAssignmentRepository) GetMonthlySchedule(doctorID uuid.UUID, year int, month int) ([]*models.DoctorAssignment, error) {
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, -1)
	return r.GetByDoctorID(doctorID, startDate, endDate)
}

func (r *doctorAssignmentRepository) DeleteByDoctorBranchDate(doctorID uuid.UUID, branchID uuid.UUID, date time.Time) error {
	query := `DELETE FROM doctor_assignments WHERE doctor_id = $1 AND branch_id = $2 AND date = $3`
	_, err := r.db.Exec(query, doctorID, branchID, date)
	return err
}

func (r *doctorAssignmentRepository) GetDoctorsByBranchAndDate(branchID uuid.UUID, date time.Time) ([]*models.DoctorAssignment, error) {
	query := `SELECT da.id, da.doctor_id, COALESCE(d.name, '') as doctor_name, COALESCE(d.code, '') as doctor_code, 
	          da.branch_id, da.date, da.expected_revenue, da.created_by, da.created_at, da.updated_at
	          FROM doctor_assignments da
	          LEFT JOIN doctors d ON da.doctor_id = d.id
	          WHERE da.branch_id = $1 AND da.date = $2 ORDER BY d.name`
	rows, err := r.db.Query(query, branchID, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assignments []*models.DoctorAssignment
	for rows.Next() {
		assignment := &models.DoctorAssignment{}
		var doctorName, doctorCode sql.NullString
		if err := rows.Scan(
			&assignment.ID, &assignment.DoctorID, &doctorName, &doctorCode,
			&assignment.BranchID, &assignment.Date, &assignment.ExpectedRevenue, &assignment.CreatedBy,
			&assignment.CreatedAt, &assignment.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if doctorName.Valid {
			assignment.DoctorName = doctorName.String
		}
		if doctorCode.Valid {
			assignment.DoctorCode = doctorCode.String
		}
		assignments = append(assignments, assignment)
	}
	return assignments, rows.Err()
}

// DoctorOnOffDayRepository implementation
type doctorOnOffDayRepository struct {
	db *sql.DB
}

func NewDoctorOnOffDayRepository(db *sql.DB) interfaces.DoctorOnOffDayRepository {
	return &doctorOnOffDayRepository{db: db}
}

func (r *doctorOnOffDayRepository) Create(day *models.DoctorOnOffDay) error {
	query := `INSERT INTO doctor_on_off_days (id, branch_id, date, is_doctor_on, created_by) 
	          VALUES ($1, $2, $3, $4, $5) RETURNING created_at, updated_at`
	return r.db.QueryRow(query, day.ID, day.BranchID, day.Date, day.IsDoctorOn, day.CreatedBy).
		Scan(&day.CreatedAt, &day.UpdatedAt)
}

func (r *doctorOnOffDayRepository) GetByID(id uuid.UUID) (*models.DoctorOnOffDay, error) {
	query := `SELECT id, branch_id, date, is_doctor_on, created_by, created_at, updated_at 
	          FROM doctor_on_off_days WHERE id = $1`
	day := &models.DoctorOnOffDay{}
	err := r.db.QueryRow(query, id).Scan(
		&day.ID, &day.BranchID, &day.Date, &day.IsDoctorOn, &day.CreatedBy, &day.CreatedAt, &day.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return day, err
}

func (r *doctorOnOffDayRepository) GetByBranchID(branchID uuid.UUID, startDate, endDate time.Time) ([]*models.DoctorOnOffDay, error) {
	query := `SELECT id, branch_id, date, is_doctor_on, created_by, created_at, updated_at 
	          FROM doctor_on_off_days WHERE branch_id = $1 AND date >= $2 AND date <= $3 ORDER BY date`
	rows, err := r.db.Query(query, branchID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var days []*models.DoctorOnOffDay
	for rows.Next() {
		day := &models.DoctorOnOffDay{}
		if err := rows.Scan(
			&day.ID, &day.BranchID, &day.Date, &day.IsDoctorOn, &day.CreatedBy, &day.CreatedAt, &day.UpdatedAt,
		); err != nil {
			return nil, err
		}
		days = append(days, day)
	}
	return days, rows.Err()
}

func (r *doctorOnOffDayRepository) GetByDate(date time.Time) ([]*models.DoctorOnOffDay, error) {
	query := `SELECT id, branch_id, date, is_doctor_on, created_by, created_at, updated_at 
	          FROM doctor_on_off_days WHERE date = $1 ORDER BY branch_id`
	rows, err := r.db.Query(query, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var days []*models.DoctorOnOffDay
	for rows.Next() {
		day := &models.DoctorOnOffDay{}
		if err := rows.Scan(
			&day.ID, &day.BranchID, &day.Date, &day.IsDoctorOn, &day.CreatedBy, &day.CreatedAt, &day.UpdatedAt,
		); err != nil {
			return nil, err
		}
		days = append(days, day)
	}
	return days, rows.Err()
}

func (r *doctorOnOffDayRepository) Update(day *models.DoctorOnOffDay) error {
	query := `UPDATE doctor_on_off_days 
	          SET is_doctor_on = $1, updated_at = CURRENT_TIMESTAMP 
	          WHERE id = $2 RETURNING updated_at`
	return r.db.QueryRow(query, day.IsDoctorOn, day.ID).Scan(&day.UpdatedAt)
}

func (r *doctorOnOffDayRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM doctor_on_off_days WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *doctorOnOffDayRepository) GetByBranchAndDate(branchID uuid.UUID, date time.Time) (*models.DoctorOnOffDay, error) {
	query := `SELECT id, branch_id, date, is_doctor_on, created_by, created_at, updated_at 
	          FROM doctor_on_off_days WHERE branch_id = $1 AND date = $2`
	day := &models.DoctorOnOffDay{}
	err := r.db.QueryRow(query, branchID, date).Scan(
		&day.ID, &day.BranchID, &day.Date, &day.IsDoctorOn, &day.CreatedBy, &day.CreatedAt, &day.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return day, err
}

// AllocationSuggestionRepository implementation
type allocationSuggestionRepository struct {
	db *sql.DB
}

func NewAllocationSuggestionRepository(db *sql.DB) interfaces.AllocationSuggestionRepository {
	return &allocationSuggestionRepository{db: db}
}

func (r *allocationSuggestionRepository) Create(suggestion *models.AllocationSuggestion) error {
	query := `INSERT INTO allocation_suggestions (id, rotation_staff_id, branch_id, date, position_id, status, confidence, reason, criteria_used) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING created_at`
	return r.db.QueryRow(query, suggestion.ID, suggestion.RotationStaffID, suggestion.BranchID, suggestion.Date,
		suggestion.PositionID, suggestion.Status, suggestion.Confidence, suggestion.Reason, suggestion.CriteriaUsed).
		Scan(&suggestion.CreatedAt)
}

func (r *allocationSuggestionRepository) GetByID(id uuid.UUID) (*models.AllocationSuggestion, error) {
	query := `SELECT id, rotation_staff_id, branch_id, date, position_id, status, confidence, reason, criteria_used, created_at, reviewed_by, reviewed_at 
	          FROM allocation_suggestions WHERE id = $1`
	suggestion := &models.AllocationSuggestion{}
	var reviewedBy sql.NullString
	var reviewedAt sql.NullTime
	err := r.db.QueryRow(query, id).Scan(
		&suggestion.ID, &suggestion.RotationStaffID, &suggestion.BranchID, &suggestion.Date,
		&suggestion.PositionID, &suggestion.Status, &suggestion.Confidence, &suggestion.Reason,
		&suggestion.CriteriaUsed, &suggestion.CreatedAt, &reviewedBy, &reviewedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if reviewedBy.Valid {
		rb, _ := uuid.Parse(reviewedBy.String)
		suggestion.ReviewedBy = &rb
	}
	if reviewedAt.Valid {
		suggestion.ReviewedAt = &reviewedAt.Time
	}
	return suggestion, nil
}

func (r *allocationSuggestionRepository) GetByBranchID(branchID uuid.UUID, startDate, endDate time.Time) ([]*models.AllocationSuggestion, error) {
	query := `SELECT id, rotation_staff_id, branch_id, date, position_id, status, confidence, reason, criteria_used, created_at, reviewed_by, reviewed_at 
	          FROM allocation_suggestions WHERE branch_id = $1 AND date >= $2 AND date <= $3 ORDER BY date, position_id`
	rows, err := r.db.Query(query, branchID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var suggestions []*models.AllocationSuggestion
	for rows.Next() {
		suggestion := &models.AllocationSuggestion{}
		var reviewedBy sql.NullString
		var reviewedAt sql.NullTime
		if err := rows.Scan(
			&suggestion.ID, &suggestion.RotationStaffID, &suggestion.BranchID, &suggestion.Date,
			&suggestion.PositionID, &suggestion.Status, &suggestion.Confidence, &suggestion.Reason,
			&suggestion.CriteriaUsed, &suggestion.CreatedAt, &reviewedBy, &reviewedAt,
		); err != nil {
			return nil, err
		}
		if reviewedBy.Valid {
			rb, _ := uuid.Parse(reviewedBy.String)
			suggestion.ReviewedBy = &rb
		}
		if reviewedAt.Valid {
			suggestion.ReviewedAt = &reviewedAt.Time
		}
		suggestions = append(suggestions, suggestion)
	}
	return suggestions, rows.Err()
}

func (r *allocationSuggestionRepository) GetByStatus(status models.SuggestionStatus) ([]*models.AllocationSuggestion, error) {
	query := `SELECT id, rotation_staff_id, branch_id, date, position_id, status, confidence, reason, criteria_used, created_at, reviewed_by, reviewed_at 
	          FROM allocation_suggestions WHERE status = $1 ORDER BY created_at DESC`
	rows, err := r.db.Query(query, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var suggestions []*models.AllocationSuggestion
	for rows.Next() {
		suggestion := &models.AllocationSuggestion{}
		var reviewedBy sql.NullString
		var reviewedAt sql.NullTime
		if err := rows.Scan(
			&suggestion.ID, &suggestion.RotationStaffID, &suggestion.BranchID, &suggestion.Date,
			&suggestion.PositionID, &suggestion.Status, &suggestion.Confidence, &suggestion.Reason,
			&suggestion.CriteriaUsed, &suggestion.CreatedAt, &reviewedBy, &reviewedAt,
		); err != nil {
			return nil, err
		}
		if reviewedBy.Valid {
			rb, _ := uuid.Parse(reviewedBy.String)
			suggestion.ReviewedBy = &rb
		}
		if reviewedAt.Valid {
			suggestion.ReviewedAt = &reviewedAt.Time
		}
		suggestions = append(suggestions, suggestion)
	}
	return suggestions, rows.Err()
}

func (r *allocationSuggestionRepository) Update(suggestion *models.AllocationSuggestion) error {
	query := `UPDATE allocation_suggestions 
	          SET status = $1, reviewed_by = $2, reviewed_at = CURRENT_TIMESTAMP 
	          WHERE id = $3 RETURNING reviewed_at`
	return r.db.QueryRow(query, suggestion.Status, suggestion.ReviewedBy, suggestion.ID).Scan(&suggestion.ReviewedAt)
}

func (r *allocationSuggestionRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM allocation_suggestions WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *allocationSuggestionRepository) BulkCreate(suggestions []*models.AllocationSuggestion) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`INSERT INTO allocation_suggestions (id, rotation_staff_id, branch_id, date, position_id, status, confidence, reason, criteria_used) 
	                         VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, suggestion := range suggestions {
		_, err := stmt.Exec(suggestion.ID, suggestion.RotationStaffID, suggestion.BranchID, suggestion.Date,
			suggestion.PositionID, suggestion.Status, suggestion.Confidence, suggestion.Reason, suggestion.CriteriaUsed)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *allocationSuggestionRepository) GetPendingByBranchID(branchID uuid.UUID, startDate, endDate time.Time) ([]*models.AllocationSuggestion, error) {
	query := `SELECT id, rotation_staff_id, branch_id, date, position_id, status, confidence, reason, criteria_used, created_at, reviewed_by, reviewed_at 
	          FROM allocation_suggestions WHERE branch_id = $1 AND date >= $2 AND date <= $3 AND status = $4 ORDER BY date, position_id`
	rows, err := r.db.Query(query, branchID, startDate, endDate, models.SuggestionStatusPending)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var suggestions []*models.AllocationSuggestion
	for rows.Next() {
		suggestion := &models.AllocationSuggestion{}
		var reviewedBy sql.NullString
		var reviewedAt sql.NullTime
		if err := rows.Scan(
			&suggestion.ID, &suggestion.RotationStaffID, &suggestion.BranchID, &suggestion.Date,
			&suggestion.PositionID, &suggestion.Status, &suggestion.Confidence, &suggestion.Reason,
			&suggestion.CriteriaUsed, &suggestion.CreatedAt, &reviewedBy, &reviewedAt,
		); err != nil {
			return nil, err
		}
		if reviewedBy.Valid {
			rb, _ := uuid.Parse(reviewedBy.String)
			suggestion.ReviewedBy = &rb
		}
		if reviewedAt.Valid {
			suggestion.ReviewedAt = &reviewedAt.Time
		}
		suggestions = append(suggestions, suggestion)
	}
	return suggestions, rows.Err()
}


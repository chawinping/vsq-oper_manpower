package postgres

import (
	"database/sql"
	"strconv"
	"time"

	"github.com/google/uuid"
	"vsq-oper-manpower/backend/internal/domain/interfaces"
	"vsq-oper-manpower/backend/internal/domain/models"
)

type Repositories struct {
	User            interfaces.UserRepository
	Role            interfaces.RoleRepository
	Staff           interfaces.StaffRepository
	Position        interfaces.PositionRepository
	Branch          interfaces.BranchRepository
	EffectiveBranch interfaces.EffectiveBranchRepository
	Revenue         interfaces.RevenueRepository
	Schedule        interfaces.ScheduleRepository
	Rotation        interfaces.RotationRepository
	Settings        interfaces.SettingsRepository
	AllocationRule  interfaces.AllocationRuleRepository
}

func NewRepositories(db *sql.DB) *Repositories {
	return &Repositories{
		User:            NewUserRepository(db),
		Role:            NewRoleRepository(db),
		Staff:           NewStaffRepository(db),
		Position:        NewPositionRepository(db),
		Branch:          NewBranchRepository(db),
		EffectiveBranch: NewEffectiveBranchRepository(db),
		Revenue:         NewRevenueRepository(db),
		Schedule:        NewScheduleRepository(db),
		Rotation:        NewRotationRepository(db),
		Settings:        NewSettingsRepository(db),
		AllocationRule:  NewAllocationRuleRepository(db),
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
	query := `INSERT INTO users (id, username, email, password_hash, role_id) 
	          VALUES ($1, $2, $3, $4, $5) RETURNING created_at, updated_at`
	return r.db.QueryRow(query, user.ID, user.Username, user.Email, user.PasswordHash, user.RoleID).
		Scan(&user.CreatedAt, &user.UpdatedAt)
}

func (r *userRepository) GetByID(id uuid.UUID) (*models.User, error) {
	user := &models.User{}
	query := `SELECT id, username, email, password_hash, role_id, created_at, updated_at 
	          FROM users WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.RoleID, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return user, err
}

func (r *userRepository) GetByUsername(username string) (*models.User, error) {
	user := &models.User{}
	query := `SELECT id, username, email, password_hash, role_id, created_at, updated_at 
	          FROM users WHERE username = $1`
	err := r.db.QueryRow(query, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.RoleID, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return user, err
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
	          role_id = $4, updated_at = CURRENT_TIMESTAMP WHERE id = $5`
	_, err := r.db.Exec(query, user.Username, user.Email, user.PasswordHash, user.RoleID, user.ID)
	return err
}

func (r *userRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *userRepository) List() ([]*models.User, error) {
	query := `SELECT id, username, email, password_hash, role_id, created_at, updated_at 
	          FROM users ORDER BY created_at DESC`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		if err := rows.Scan(
			&user.ID, &user.Username, &user.Email, &user.PasswordHash,
			&user.RoleID, &user.CreatedAt, &user.UpdatedAt,
		); err != nil {
			return nil, err
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
	query := `INSERT INTO staff (id, name, staff_type, position_id, branch_id, coverage_area) 
	          VALUES ($1, $2, $3, $4, $5, $6) RETURNING created_at, updated_at`
	return r.db.QueryRow(query, staff.ID, staff.Name, staff.StaffType, staff.PositionID,
		staff.BranchID, staff.CoverageArea).Scan(&staff.CreatedAt, &staff.UpdatedAt)
}

func (r *staffRepository) GetByID(id uuid.UUID) (*models.Staff, error) {
	staff := &models.Staff{}
	query := `SELECT id, name, staff_type, position_id, branch_id, coverage_area, created_at, updated_at 
	          FROM staff WHERE id = $1`
	var branchID sql.NullString
	err := r.db.QueryRow(query, id).Scan(
		&staff.ID, &staff.Name, &staff.StaffType, &staff.PositionID,
		&branchID, &staff.CoverageArea, &staff.CreatedAt, &staff.UpdatedAt,
	)
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
	return staff, nil
}

func (r *staffRepository) Update(staff *models.Staff) error {
	query := `UPDATE staff SET name = $1, staff_type = $2, position_id = $3, 
	          branch_id = $4, coverage_area = $5, updated_at = CURRENT_TIMESTAMP WHERE id = $6`
	_, err := r.db.Exec(query, staff.Name, staff.StaffType, staff.PositionID,
		staff.BranchID, staff.CoverageArea, staff.ID)
	return err
}

func (r *staffRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM staff WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *staffRepository) List(filters interfaces.StaffFilters) ([]*models.Staff, error) {
	query := `SELECT id, name, staff_type, position_id, branch_id, coverage_area, created_at, updated_at 
	          FROM staff WHERE 1=1`
	args := []interface{}{}
	argPos := 1

	if filters.StaffType != nil {
		query += ` AND staff_type = $` + strconv.Itoa(argPos)
		args = append(args, *filters.StaffType)
		argPos++
	}
	if filters.BranchID != nil {
		query += ` AND branch_id = $` + strconv.Itoa(argPos)
		args = append(args, *filters.BranchID)
		argPos++
	}
	if filters.PositionID != nil {
		query += ` AND position_id = $` + strconv.Itoa(argPos)
		args = append(args, *filters.PositionID)
		argPos++
	}

	query += ` ORDER BY name`

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var staffList []*models.Staff
	for rows.Next() {
		staff := &models.Staff{}
		var branchID sql.NullString
		if err := rows.Scan(
			&staff.ID, &staff.Name, &staff.StaffType, &staff.PositionID,
			&branchID, &staff.CoverageArea, &staff.CreatedAt, &staff.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if branchID.Valid {
			bID, _ := uuid.Parse(branchID.String)
			staff.BranchID = &bID
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
	query := `INSERT INTO positions (id, name, min_staff_per_branch, revenue_multiplier) 
	          VALUES ($1, $2, $3, $4) RETURNING created_at`
	return r.db.QueryRow(query, position.ID, position.Name, position.MinStaffPerBranch,
		position.RevenueMultiplier).Scan(&position.CreatedAt)
}

func (r *positionRepository) GetByID(id uuid.UUID) (*models.Position, error) {
	position := &models.Position{}
	query := `SELECT id, name, min_staff_per_branch, revenue_multiplier, created_at 
	          FROM positions WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(
		&position.ID, &position.Name, &position.MinStaffPerBranch,
		&position.RevenueMultiplier, &position.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return position, err
}

func (r *positionRepository) Update(position *models.Position) error {
	query := `UPDATE positions SET name = $1, min_staff_per_branch = $2, 
	          revenue_multiplier = $3 WHERE id = $4`
	_, err := r.db.Exec(query, position.Name, position.MinStaffPerBranch,
		position.RevenueMultiplier, position.ID)
	return err
}

func (r *positionRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM positions WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *positionRepository) List() ([]*models.Position, error) {
	query := `SELECT id, name, min_staff_per_branch, revenue_multiplier, created_at 
	          FROM positions ORDER BY name`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var positions []*models.Position
	for rows.Next() {
		position := &models.Position{}
		if err := rows.Scan(
			&position.ID, &position.Name, &position.MinStaffPerBranch,
			&position.RevenueMultiplier, &position.CreatedAt,
		); err != nil {
			return nil, err
		}
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
	query := `INSERT INTO branches (id, name, code, address, area_manager_id, expected_revenue, priority) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING created_at, updated_at`
	return r.db.QueryRow(query, branch.ID, branch.Name, branch.Code, branch.Address,
		branch.AreaManagerID, branch.ExpectedRevenue, branch.Priority).
		Scan(&branch.CreatedAt, &branch.UpdatedAt)
}

func (r *branchRepository) GetByID(id uuid.UUID) (*models.Branch, error) {
	branch := &models.Branch{}
	query := `SELECT id, name, code, address, area_manager_id, expected_revenue, priority, created_at, updated_at 
	          FROM branches WHERE id = $1`
	var areaManagerID sql.NullString
	err := r.db.QueryRow(query, id).Scan(
		&branch.ID, &branch.Name, &branch.Code, &branch.Address,
		&areaManagerID, &branch.ExpectedRevenue, &branch.Priority,
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
	query := `UPDATE branches SET name = $1, code = $2, address = $3, area_manager_id = $4, 
	          expected_revenue = $5, priority = $6, updated_at = CURRENT_TIMESTAMP WHERE id = $7`
	_, err := r.db.Exec(query, branch.Name, branch.Code, branch.Address,
		branch.AreaManagerID, branch.ExpectedRevenue, branch.Priority, branch.ID)
	return err
}

func (r *branchRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM branches WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *branchRepository) List() ([]*models.Branch, error) {
	query := `SELECT id, name, code, address, area_manager_id, expected_revenue, priority, created_at, updated_at 
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
			&branch.ID, &branch.Name, &branch.Code, &branch.Address,
			&areaManagerID, &branch.ExpectedRevenue, &branch.Priority,
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
	query := `SELECT id, name, code, address, area_manager_id, expected_revenue, priority, created_at, updated_at 
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
			&branch.ID, &branch.Name, &branch.Code, &branch.Address,
			&areaManagerID, &branch.ExpectedRevenue, &branch.Priority,
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
	query := `INSERT INTO effective_branches (id, rotation_staff_id, branch_id, level) 
	          VALUES ($1, $2, $3, $4) RETURNING created_at`
	return r.db.QueryRow(query, eb.ID, eb.RotationStaffID, eb.BranchID, eb.Level).
		Scan(&eb.CreatedAt)
}

func (r *effectiveBranchRepository) GetByRotationStaffID(rotationStaffID uuid.UUID) ([]*models.EffectiveBranch, error) {
	query := `SELECT id, rotation_staff_id, branch_id, level, created_at 
	          FROM effective_branches WHERE rotation_staff_id = $1 ORDER BY level, created_at`
	rows, err := r.db.Query(query, rotationStaffID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ebs []*models.EffectiveBranch
	for rows.Next() {
		eb := &models.EffectiveBranch{}
		if err := rows.Scan(&eb.ID, &eb.RotationStaffID, &eb.BranchID, &eb.Level, &eb.CreatedAt); err != nil {
			return nil, err
		}
		ebs = append(ebs, eb)
	}
	return ebs, rows.Err()
}

func (r *effectiveBranchRepository) GetByBranchID(branchID uuid.UUID) ([]*models.EffectiveBranch, error) {
	query := `SELECT id, rotation_staff_id, branch_id, level, created_at 
	          FROM effective_branches WHERE branch_id = $1 ORDER BY level, created_at`
	rows, err := r.db.Query(query, branchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ebs []*models.EffectiveBranch
	for rows.Next() {
		eb := &models.EffectiveBranch{}
		if err := rows.Scan(&eb.ID, &eb.RotationStaffID, &eb.BranchID, &eb.Level, &eb.CreatedAt); err != nil {
			return nil, err
		}
		ebs = append(ebs, eb)
	}
	return ebs, rows.Err()
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
	query := `INSERT INTO revenue_data (id, branch_id, date, expected_revenue, actual_revenue) 
	          VALUES ($1, $2, $3, $4, $5) RETURNING created_at, updated_at`
	return r.db.QueryRow(query, revenue.ID, revenue.BranchID, revenue.Date,
		revenue.ExpectedRevenue, revenue.ActualRevenue).
		Scan(&revenue.CreatedAt, &revenue.UpdatedAt)
}

func (r *revenueRepository) GetByBranchID(branchID uuid.UUID, startDate, endDate time.Time) ([]*models.RevenueData, error) {
	query := `SELECT id, branch_id, date, expected_revenue, actual_revenue, created_at, updated_at 
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
		if err := rows.Scan(
			&revenue.ID, &revenue.BranchID, &revenue.Date,
			&revenue.ExpectedRevenue, &actualRevenue,
			&revenue.CreatedAt, &revenue.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if actualRevenue.Valid {
			revenue.ActualRevenue = &actualRevenue.Float64
		}
		revenues = append(revenues, revenue)
	}
	return revenues, rows.Err()
}

func (r *revenueRepository) GetByDate(date time.Time) ([]*models.RevenueData, error) {
	query := `SELECT id, branch_id, date, expected_revenue, actual_revenue, created_at, updated_at 
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
		if err := rows.Scan(
			&revenue.ID, &revenue.BranchID, &revenue.Date,
			&revenue.ExpectedRevenue, &actualRevenue,
			&revenue.CreatedAt, &revenue.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if actualRevenue.Valid {
			revenue.ActualRevenue = &actualRevenue.Float64
		}
		revenues = append(revenues, revenue)
	}
	return revenues, rows.Err()
}

func (r *revenueRepository) Update(revenue *models.RevenueData) error {
	query := `UPDATE revenue_data SET expected_revenue = $1, actual_revenue = $2, 
	          updated_at = CURRENT_TIMESTAMP WHERE id = $3`
	_, err := r.db.Exec(query, revenue.ExpectedRevenue, revenue.ActualRevenue, revenue.ID)
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
	query := `INSERT INTO staff_schedules (id, staff_id, branch_id, date, is_working_day, created_by) 
	          VALUES ($1, $2, $3, $4, $5, $6) RETURNING created_at`
	return r.db.QueryRow(query, schedule.ID, schedule.StaffID, schedule.BranchID,
		schedule.Date, schedule.IsWorkingDay, schedule.CreatedBy).
		Scan(&schedule.CreatedAt)
}

func (r *scheduleRepository) GetByBranchID(branchID uuid.UUID, startDate, endDate time.Time) ([]*models.StaffSchedule, error) {
	query := `SELECT id, staff_id, branch_id, date, is_working_day, created_by, created_at 
	          FROM staff_schedules WHERE branch_id = $1 AND date >= $2 AND date <= $3 ORDER BY date, staff_id`
	rows, err := r.db.Query(query, branchID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []*models.StaffSchedule
	for rows.Next() {
		schedule := &models.StaffSchedule{}
		if err := rows.Scan(
			&schedule.ID, &schedule.StaffID, &schedule.BranchID, &schedule.Date,
			&schedule.IsWorkingDay, &schedule.CreatedBy, &schedule.CreatedAt,
		); err != nil {
			return nil, err
		}
		schedules = append(schedules, schedule)
	}
	return schedules, rows.Err()
}

func (r *scheduleRepository) GetByStaffID(staffID uuid.UUID, startDate, endDate time.Time) ([]*models.StaffSchedule, error) {
	query := `SELECT id, staff_id, branch_id, date, is_working_day, created_by, created_at 
	          FROM staff_schedules WHERE staff_id = $1 AND date >= $2 AND date <= $3 ORDER BY date`
	rows, err := r.db.Query(query, staffID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []*models.StaffSchedule
	for rows.Next() {
		schedule := &models.StaffSchedule{}
		if err := rows.Scan(
			&schedule.ID, &schedule.StaffID, &schedule.BranchID, &schedule.Date,
			&schedule.IsWorkingDay, &schedule.CreatedBy, &schedule.CreatedAt,
		); err != nil {
			return nil, err
		}
		schedules = append(schedules, schedule)
	}
	return schedules, rows.Err()
}

func (r *scheduleRepository) Update(schedule *models.StaffSchedule) error {
	query := `UPDATE staff_schedules SET is_working_day = $1 WHERE id = $2`
	_, err := r.db.Exec(query, schedule.IsWorkingDay, schedule.ID)
	return err
}

func (r *scheduleRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM staff_schedules WHERE id = $1`
	_, err := r.db.Exec(query, id)
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


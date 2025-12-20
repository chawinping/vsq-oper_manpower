package postgres

import (
	"database/sql"
	"fmt"
	"vsq-oper-manpower/backend/internal/constants"

	"github.com/google/uuid"
)

func RunMigrations(db *sql.DB) error {
	migrations := []string{
		createRolesTable,
		createUsersTable,
		createBranchesTable,
		createPositionsTable,
		createStaffTable,
		createEffectiveBranchesTable,
		createRevenueDataTable,
		createStaffSchedulesTable,
		createRotationAssignmentsTable,
		createSystemSettingsTable,
		createStaffAllocationRulesTable,
		insertDefaultRoles,
		insertDefaultPositions,
	}

	for _, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	// Seed standard branch codes
	if err := SeedStandardBranches(db); err != nil {
		return fmt.Errorf("failed to seed standard branches: %w", err)
	}

	return nil
}

const createRolesTable = `
CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
`

const createUsersTable = `
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(100) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role_id UUID NOT NULL REFERENCES roles(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
`

const createBranchesTable = `
CREATE TABLE IF NOT EXISTS branches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50) UNIQUE NOT NULL,
    address TEXT,
    area_manager_id UUID REFERENCES users(id),
    expected_revenue DECIMAL(15,2) DEFAULT 0,
    priority INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
`

const createPositionsTable = `
CREATE TABLE IF NOT EXISTS positions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    min_staff_per_branch INTEGER DEFAULT 1,
    revenue_multiplier DECIMAL(10,4) DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
`

const createStaffTable = `
CREATE TABLE IF NOT EXISTS staff (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    staff_type VARCHAR(20) NOT NULL CHECK (staff_type IN ('branch', 'rotation')),
    position_id UUID NOT NULL REFERENCES positions(id),
    branch_id UUID REFERENCES branches(id),
    coverage_area VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
`

const createEffectiveBranchesTable = `
CREATE TABLE IF NOT EXISTS effective_branches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rotation_staff_id UUID NOT NULL REFERENCES staff(id),
    branch_id UUID NOT NULL REFERENCES branches(id),
    level INTEGER NOT NULL CHECK (level IN (1, 2)),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(rotation_staff_id, branch_id)
);
`

const createRevenueDataTable = `
CREATE TABLE IF NOT EXISTS revenue_data (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    branch_id UUID NOT NULL REFERENCES branches(id),
    date DATE NOT NULL,
    expected_revenue DECIMAL(15,2) NOT NULL,
    actual_revenue DECIMAL(15,2),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(branch_id, date)
);
`

const createStaffSchedulesTable = `
CREATE TABLE IF NOT EXISTS staff_schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    staff_id UUID NOT NULL REFERENCES staff(id),
    branch_id UUID NOT NULL REFERENCES branches(id),
    date DATE NOT NULL,
    is_working_day BOOLEAN NOT NULL DEFAULT true,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(staff_id, branch_id, date)
);
`

const createRotationAssignmentsTable = `
CREATE TABLE IF NOT EXISTS rotation_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rotation_staff_id UUID NOT NULL REFERENCES staff(id),
    branch_id UUID NOT NULL REFERENCES branches(id),
    date DATE NOT NULL,
    assignment_level INTEGER NOT NULL CHECK (assignment_level IN (1, 2)),
    assigned_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(rotation_staff_id, branch_id, date)
);
`

const createSystemSettingsTable = `
CREATE TABLE IF NOT EXISTS system_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key VARCHAR(100) UNIQUE NOT NULL,
    value TEXT NOT NULL,
    description TEXT,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
`

const createStaffAllocationRulesTable = `
CREATE TABLE IF NOT EXISTS staff_allocation_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    position_id UUID NOT NULL REFERENCES positions(id),
    min_staff INTEGER NOT NULL DEFAULT 1,
    revenue_threshold DECIMAL(15,2) DEFAULT 0,
    staff_count_formula TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(position_id)
);
`

const insertDefaultRoles = `
INSERT INTO roles (id, name) VALUES
    ('00000000-0000-0000-0000-000000000001', 'admin'),
    ('00000000-0000-0000-0000-000000000002', 'area_manager'),
    ('00000000-0000-0000-0000-000000000003', 'district_manager'),
    ('00000000-0000-0000-0000-000000000004', 'branch_manager'),
    ('00000000-0000-0000-0000-000000000005', 'viewer')
ON CONFLICT (id) DO NOTHING;
`

const insertDefaultPositions = `
INSERT INTO positions (id, name, min_staff_per_branch, revenue_multiplier) VALUES
    ('10000000-0000-0000-0000-000000000001', 'Branch Manager', 1, 0),
    ('10000000-0000-0000-0000-000000000002', 'Assistant Branch Manager', 0, 0.5),
    ('10000000-0000-0000-0000-000000000003', 'Service Consultant', 1, 1.0),
    ('10000000-0000-0000-0000-000000000004', 'Coordinator', 1, 0.8),
    ('10000000-0000-0000-0000-000000000005', 'Doctor Assistant', 2, 1.2),
    ('10000000-0000-0000-0000-000000000006', 'Physiotherapist', 1, 1.0),
    ('10000000-0000-0000-0000-000000000007', 'Nurse', 2, 1.0)
ON CONFLICT (id) DO NOTHING;
`

// SeedStandardBranches seeds the database with standard branch codes.
// This ensures all standard branch codes (FR-BM-03) are always available in the system.
func SeedStandardBranches(db *sql.DB) error {
	// Generate deterministic UUIDs for each branch code
	// Using a base UUID pattern: 20000000-0000-0000-0000-XXXXXXXXXXXX
	// where X is a sequential hex number
	baseUUID := uuid.MustParse("20000000-0000-0000-0000-000000000000")
	
	standardCodes := constants.GetStandardBranchCodes()
	
	// Prepare the insert statement
	stmt := `INSERT INTO branches (id, name, code, address, expected_revenue, priority) 
	         VALUES ($1, $2, $3, $4, $5, $6)
	         ON CONFLICT (code) DO NOTHING`
	
	for _, code := range standardCodes {
		// Generate deterministic UUID for this branch code using SHA1 hash
		// This ensures the same branch code always gets the same UUID
		branchID := uuid.NewSHA1(baseUUID, []byte(code))
		
		// Branch name defaults to code if not specified
		branchName := code
		
		// Insert branch with default values
		_, err := db.Exec(stmt,
			branchID,
			branchName,
			code,
			"", // address - can be updated later
			0,  // expected_revenue - default 0
			0,  // priority - default 0
		)
		if err != nil {
			return fmt.Errorf("failed to seed branch %s: %w", code, err)
		}
	}
	
	return nil
}


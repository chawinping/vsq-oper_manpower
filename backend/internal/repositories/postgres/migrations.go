package postgres

import (
	"database/sql"
	"fmt"
	"strings"
	"vsq-oper-manpower/backend/internal/constants"

	"github.com/google/uuid"
)

func RunMigrations(db *sql.DB) error {
	migrations := []string{
		createRolesTable,
		createUsersTable,
		createBranchesTable,
		createPositionsTable,
		createAreasOfOperationTable,
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

	// Run data migrations for existing tables
	if err := runDataMigrations(db); err != nil {
		return fmt.Errorf("data migration failed: %w", err)
	}

	// Link branch managers to their branches based on username
	if err := linkBranchManagersToBranches(db); err != nil {
		return fmt.Errorf("failed to link branch managers to branches: %w", err)
	}

	// Seed standard branch codes
	if err := SeedStandardBranches(db); err != nil {
		return fmt.Errorf("failed to seed standard branches: %w", err)
	}

	return nil
}

// runDataMigrations handles migrations for existing data
func runDataMigrations(db *sql.DB) error {
	// Add nickname column to staff table if it doesn't exist
	_, err := db.Exec(`
		DO $$ 
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'staff' AND column_name = 'nickname'
			) THEN
				ALTER TABLE staff ADD COLUMN nickname VARCHAR(100);
			END IF;
		END $$;
	`)
	if err != nil {
		return fmt.Errorf("failed to add nickname column: %w", err)
	}

	// Add skill_level column to staff table if it doesn't exist
	_, err = db.Exec(`
		DO $$ 
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'staff' AND column_name = 'skill_level'
			) THEN
				ALTER TABLE staff ADD COLUMN skill_level INTEGER DEFAULT 5 CHECK (skill_level >= 0 AND skill_level <= 10);
			END IF;
		END $$;
	`)
	if err != nil {
		return fmt.Errorf("failed to add skill_level column: %w", err)
	}

	// Add area_of_operation_id column to staff table if it doesn't exist
	_, err = db.Exec(`
		DO $$ 
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'staff' AND column_name = 'area_of_operation_id'
			) THEN
				ALTER TABLE staff ADD COLUMN area_of_operation_id UUID REFERENCES areas_of_operation(id);
			END IF;
		END $$;
	`)
	if err != nil {
		return fmt.Errorf("failed to add area_of_operation_id column: %w", err)
	}

	// Add schedule_status column to staff_schedules table if it doesn't exist
	_, err = db.Exec(`
		DO $$ 
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'staff_schedules' AND column_name = 'schedule_status'
			) THEN
				ALTER TABLE staff_schedules ADD COLUMN schedule_status VARCHAR(20) DEFAULT 'off' CHECK (schedule_status IN ('working', 'off', 'leave', 'sick_leave'));
				-- Migrate existing data: convert is_working_day to schedule_status
				UPDATE staff_schedules SET schedule_status = CASE 
					WHEN is_working_day = true THEN 'working' 
					ELSE 'off' 
				END;
			END IF;
		END $$;
	`)
	if err != nil {
		return fmt.Errorf("failed to add schedule_status column: %w", err)
	}

	// Update CHECK constraint to include 'sick_leave' if it doesn't already include it
	_, err = db.Exec(`
		DO $$ 
		BEGIN
			-- Drop existing constraint if it exists and doesn't include 'sick_leave'
			IF EXISTS (
				SELECT 1 FROM information_schema.check_constraints 
				WHERE constraint_name LIKE 'staff_schedules_schedule_status_check%'
				AND constraint_schema = 'public'
			) THEN
				-- Check if constraint needs updating (doesn't include 'sick_leave')
				IF NOT EXISTS (
					SELECT 1 FROM information_schema.check_constraints 
					WHERE constraint_name LIKE 'staff_schedules_schedule_status_check%'
					AND constraint_schema = 'public'
					AND check_clause LIKE '%sick_leave%'
				) THEN
					ALTER TABLE staff_schedules DROP CONSTRAINT IF EXISTS staff_schedules_schedule_status_check;
					ALTER TABLE staff_schedules ADD CONSTRAINT staff_schedules_schedule_status_check 
						CHECK (schedule_status IN ('working', 'off', 'leave', 'sick_leave'));
				END IF;
			END IF;
		END $$;
	`)
	if err != nil {
		return fmt.Errorf("failed to update schedule_status constraint: %w", err)
	}

	// Add display_order column to positions table if it doesn't exist
	_, err = db.Exec(`
		DO $$ 
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'positions' AND column_name = 'display_order'
			) THEN
				ALTER TABLE positions ADD COLUMN display_order INTEGER DEFAULT 999;
				-- Set default display_order values for existing positions
				-- Branch Manager gets 1 (highest priority)
				UPDATE positions SET display_order = 1 WHERE name = 'Branch Manager' OR name = 'ผู้จัดการสาขา';
				-- Assistant Branch Manager gets 2
				UPDATE positions SET display_order = 2 WHERE name = 'Assistant Branch Manager' OR name = 'รองผู้จัดการสาขา' OR name = 'ผู้ช่วยผู้จัดการสาขา';
				-- Other positions get incremental values starting from 10
				UPDATE positions SET display_order = 10 WHERE display_order = 999 AND (name LIKE '%Manager%' OR name LIKE '%ผู้จัดการ%');
				UPDATE positions SET display_order = 20 WHERE display_order = 999 AND (name LIKE '%Doctor%' OR name LIKE '%แพทย์%');
				UPDATE positions SET display_order = 30 WHERE display_order = 999 AND (name LIKE '%Nurse%' OR name LIKE '%พยาบาล%');
				UPDATE positions SET display_order = 40 WHERE display_order = 999 AND (name LIKE '%Receptionist%' OR name LIKE '%ต้อนรับ%');
				UPDATE positions SET display_order = 50 WHERE display_order = 999 AND (name LIKE '%Coordinator%' OR name LIKE '%ประสานงาน%');
				-- Set remaining positions to 100+
				UPDATE positions SET display_order = 100 + ROW_NUMBER() OVER (ORDER BY name) WHERE display_order = 999;
			END IF;
		END $$;
	`)
	if err != nil {
		return fmt.Errorf("failed to add display_order column: %w", err)
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
    branch_id UUID REFERENCES branches(id),
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
    display_order INTEGER DEFAULT 999,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
`

const createAreasOfOperationTable = `
CREATE TABLE IF NOT EXISTS areas_of_operation (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50) NOT NULL UNIQUE,
    description TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
`

const createStaffTable = `
CREATE TABLE IF NOT EXISTS staff (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    nickname VARCHAR(100),
    name VARCHAR(255) NOT NULL,
    staff_type VARCHAR(20) NOT NULL CHECK (staff_type IN ('branch', 'rotation')),
    position_id UUID NOT NULL REFERENCES positions(id),
    branch_id UUID REFERENCES branches(id),
    coverage_area VARCHAR(255),
    area_of_operation_id UUID REFERENCES areas_of_operation(id),
    skill_level INTEGER DEFAULT 5 CHECK (skill_level >= 0 AND skill_level <= 10),
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
    schedule_status VARCHAR(20) NOT NULL DEFAULT 'off' CHECK (schedule_status IN ('working', 'off', 'leave', 'sick_leave')),
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
INSERT INTO positions (id, name, min_staff_per_branch, revenue_multiplier, display_order) VALUES
    ('10000000-0000-0000-0000-000000000001', 'Branch Manager', 1, 0, 1),
    ('10000000-0000-0000-0000-000000000002', 'Assistant Branch Manager', 0, 0.5, 2),
    ('10000000-0000-0000-0000-000000000003', 'Service Consultant', 1, 1.0, 50),
    ('10000000-0000-0000-0000-000000000004', 'Coordinator', 1, 0.8, 50),
    ('10000000-0000-0000-0000-000000000005', 'Doctor Assistant', 2, 1.2, 20),
    ('10000000-0000-0000-0000-000000000006', 'Physiotherapist', 1, 1.0, 30),
    ('10000000-0000-0000-0000-000000000007', 'Nurse', 2, 1.0, 30),
    ('10000000-0000-0000-0000-000000000008', 'ผู้จัดการสาขา', 1, 0, 1),
    ('10000000-0000-0000-0000-000000000009', 'รองผู้จัดการสาขา', 0, 0.5, 2),
    ('10000000-0000-0000-0000-000000000010', 'ผู้ช่วยแพทย์', 2, 1.2, 20),
    ('10000000-0000-0000-0000-000000000011', 'ผู้ประสานงานคลินิก (Clinic Coordination Officer)', 1, 0.8, 50),
    ('10000000-0000-0000-0000-000000000012', 'ผู้ช่วย Laser Specialist', 1, 1.0, 20),
    ('10000000-0000-0000-0000-000000000013', 'พนักงานต้อนรับ (Laser Receptionist)', 1, 0.5, 40),
    ('10000000-0000-0000-0000-000000000014', 'แม่บ้านประจำสาขา', 1, 0.3, 60),
    ('10000000-0000-0000-0000-000000000015', 'พยาบาล', 2, 1.0, 30),
    ('10000000-0000-0000-0000-000000000016', 'ผู้ช่วยผู้จัดการสาขา', 0, 0.5, 2),
    ('10000000-0000-0000-0000-000000000017', 'ผู้ช่วยแพทย์ Pico Laser', 1, 1.2, 20),
    ('10000000-0000-0000-0000-000000000018', 'รองผู้จัดการสาขาและล่าม', 0, 0.6, 2),
    ('10000000-0000-0000-0000-000000000019', 'Front+ล่ามวนสาขา', 1, 0.7, 40),
    ('10000000-0000-0000-0000-000000000020', 'ผู้ช่วยแพทย์ Pico', 1, 1.2, 20),
    ('10000000-0000-0000-0000-000000000021', 'พนักงานต้อนรับ (Pico Laser Receptionist)', 1, 0.5, 40),
    ('10000000-0000-0000-0000-000000000022', 'ผู้จัดการเขต', 0, 0, 10),
    ('10000000-0000-0000-0000-000000000023', 'ผู้จัดการแผนกและกำกับพัฒนาระเบียบสาขา', 0, 0, 10),
    ('10000000-0000-0000-0000-000000000024', 'หัวหน้าผู้ช่วยแพทย์', 1, 1.5, 15),
    ('10000000-0000-0000-0000-000000000025', 'ผู้ช่วยพิเศษ', 0, 0.8, 20),
    ('10000000-0000-0000-0000-000000000026', 'ผู้ช่วยแพทย์วนสาขา', 1, 1.2, 20),
    ('10000000-0000-0000-0000-000000000027', 'ฟร้อนท์วนสาขา', 1, 0.5, 40)
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

// linkBranchManagersToBranches links existing branch managers to their branches
// based on username pattern (e.g., "bkk01mgr" -> branch code "BKK01")
func linkBranchManagersToBranches(db *sql.DB) error {
	// Get branch manager role ID
	var branchManagerRoleID string
	err := db.QueryRow("SELECT id FROM roles WHERE name = 'branch_manager'").Scan(&branchManagerRoleID)
	if err != nil {
		// If role doesn't exist yet, skip linking (roles are created in migrations)
		return nil
	}

	// Get all branch managers without branch_id set
	rows, err := db.Query(`
		SELECT id, username 
		FROM users 
		WHERE role_id = $1 AND (branch_id IS NULL OR branch_id = '00000000-0000-0000-0000-000000000000'::uuid)
	`, branchManagerRoleID)
	if err != nil {
		return fmt.Errorf("failed to query branch managers: %w", err)
	}
	defer rows.Close()

	linked := 0
	for rows.Next() {
		var userID string
		var username string
		if err := rows.Scan(&userID, &username); err != nil {
			continue
		}

		// Extract branch code from username
		// Pattern: {branchcode}mgr or {branchcode}amgr
		branchCode := ""
		if strings.HasSuffix(strings.ToLower(username), "amgr") {
			branchCode = strings.ToUpper(username[:len(username)-4])
		} else if strings.HasSuffix(strings.ToLower(username), "mgr") {
			branchCode = strings.ToUpper(username[:len(username)-3])
		} else {
			// Skip if username doesn't match expected pattern
			continue
		}

		// Find branch by code
		var branchID string
		err := db.QueryRow("SELECT id FROM branches WHERE code = $1", branchCode).Scan(&branchID)
		if err != nil {
			// Branch not found, skip this user
			continue
		}

		// Update user's branch_id
		_, err = db.Exec("UPDATE users SET branch_id = $1 WHERE id = $2", branchID, userID)
		if err != nil {
			// Log error but continue with other users
			continue
		}
		linked++
	}

	return rows.Err()
}


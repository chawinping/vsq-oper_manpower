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
		createZonesTable,
		createZoneBranchesTable,
		createAreaOfOperationZonesTable,
		createAreaOfOperationBranchesTable,
		createStaffTable,
		createStaffBranchesTable,
		createEffectiveBranchesTable,
		createRevenueDataTable,
		createStaffSchedulesTable,
		createRotationAssignmentsTable,
		createSystemSettingsTable,
		createStaffAllocationRulesTable,
		createAllocationCriteriaTable,
		createPositionQuotasTable,
		createDoctorsTable,
		createDoctorPreferencesTable,
		createDoctorAssignmentsTable,
		createDoctorOnOffDaysTable,
		createDoctorDefaultSchedulesTable,
		createDoctorWeeklyOffDaysTable,
		createDoctorScheduleOverridesTable,
		createAllocationSuggestionsTable,
		createBranchWeeklyRevenueTable,
		createBranchConstraintsTable,
		createRevenueLevelTiersTable,
		createStaffRequirementScenariosTable,
		createScenarioPositionRequirementsTable,
		insertDefaultRoles,
		insertDefaultPositions,
		insertDefaultRevenueLevelTiers,
		insertDefaultStaffRequirementScenarios,
		createGetRevenueLevelTierFunction,
		createScenarioMatchesFunction,
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

	// Migrate English positions to Thai positions and remove English positions
	if err := MigrateRemoveEnglishPositions(db); err != nil {
		return fmt.Errorf("failed to migrate English positions: %w", err)
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

	// Add zone_id column to staff table if it doesn't exist
	_, err = db.Exec(`
		DO $$ 
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'staff' AND column_name = 'zone_id'
			) THEN
				ALTER TABLE staff ADD COLUMN zone_id UUID REFERENCES zones(id);
			END IF;
		END $$;
	`)
	if err != nil {
		return fmt.Errorf("failed to add zone_id column: %w", err)
	}

	// Add position_type column to positions table if it doesn't exist
	_, err = db.Exec(`
		DO $$ 
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'positions' AND column_name = 'position_type'
			) THEN
				ALTER TABLE positions ADD COLUMN position_type VARCHAR(20) NOT NULL DEFAULT 'branch' CHECK (position_type IN ('branch', 'rotation'));
			END IF;
		END $$;
	`)
	if err != nil {
		return fmt.Errorf("failed to add position_type column: %w", err)
	}

	// Update existing rotation positions based on name patterns and specific position IDs
	_, err = db.Exec(`
		UPDATE positions 
		SET position_type = 'rotation' 
		WHERE (
			name LIKE '%วนสาขา%' 
			OR name LIKE '%Rotation%' 
			OR name ILIKE '%rotation%'
			OR id IN (
				'10000000-0000-0000-0000-000000000019', -- Front+ล่ามวนสาขา
				'10000000-0000-0000-0000-000000000022', -- ผู้จัดการเขต
				'10000000-0000-0000-0000-000000000023', -- ผู้จัดการแผนกและกำกับพัฒนาระเบียบสาขา
				'10000000-0000-0000-0000-000000000024', -- หัวหน้าผู้ช่วยแพทย์
				'10000000-0000-0000-0000-000000000025', -- ผู้ช่วยพิเศษ
				'10000000-0000-0000-0000-000000000026', -- ผู้ช่วยแพทย์วนสาขา
				'10000000-0000-0000-0000-000000000027'  -- ฟร้อนท์วนสาขา
			)
		)
		AND position_type = 'branch';
	`)
	if err != nil {
		return fmt.Errorf("failed to update rotation positions: %w", err)
	}

	// Add manpower_type column to positions table if it doesn't exist
	_, err = db.Exec(`
		DO $$ 
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'positions' AND column_name = 'manpower_type'
			) THEN
				ALTER TABLE positions ADD COLUMN manpower_type VARCHAR(50) NOT NULL DEFAULT 'อื่นๆ' CHECK (manpower_type IN ('พนักงานฟร้อนท์', 'ผู้ช่วยแพทย์', 'อื่นๆ', 'ทำความสะอาด'));
			END IF;
		END $$;
	`)
	if err != nil {
		return fmt.Errorf("failed to add manpower_type column: %w", err)
	}

	// Update existing positions with appropriate manpower_type based on name patterns
	_, err = db.Exec(`
		UPDATE positions 
		SET manpower_type = CASE
			WHEN name LIKE '%ฟร้อนท์%' OR name LIKE '%Front%' OR name LIKE '%ต้อนรับ%' OR name LIKE '%Receptionist%' THEN 'พนักงานฟร้อนท์'
			WHEN name LIKE '%ผู้ช่วยแพทย์%' OR name LIKE '%Doctor Assistant%' OR name LIKE '%พยาบาล%' OR name LIKE '%Nurse%' OR name LIKE '%Physiotherapist%' THEN 'ผู้ช่วยแพทย์'
			WHEN name LIKE '%แม่บ้าน%' OR name LIKE '%Housekeeper%' OR name LIKE '%ทำความสะอาด%' THEN 'ทำความสะอาด'
			ELSE 'อื่นๆ'
		END
		WHERE manpower_type = 'อื่นๆ';
	`)
	if err != nil {
		return fmt.Errorf("failed to update manpower types: %w", err)
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

	// Add is_adhoc and adhoc_reason columns to rotation_assignments table if they don't exist
	_, err = db.Exec(`
		DO $$ 
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'rotation_assignments' AND column_name = 'is_adhoc'
			) THEN
				ALTER TABLE rotation_assignments ADD COLUMN is_adhoc BOOLEAN DEFAULT false;
			END IF;
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'rotation_assignments' AND column_name = 'adhoc_reason'
			) THEN
				ALTER TABLE rotation_assignments ADD COLUMN adhoc_reason TEXT;
			END IF;
		END $$;
	`)
	if err != nil {
		return fmt.Errorf("failed to add adhoc columns to rotation_assignments: %w", err)
	}

	// Add travel parameters columns to effective_branches table if they don't exist
	_, err = db.Exec(`
		DO $$ 
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'effective_branches' AND column_name = 'commute_duration_minutes'
			) THEN
				ALTER TABLE effective_branches ADD COLUMN commute_duration_minutes INTEGER DEFAULT 300;
			END IF;
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'effective_branches' AND column_name = 'transit_count'
			) THEN
				ALTER TABLE effective_branches ADD COLUMN transit_count INTEGER DEFAULT 10;
			END IF;
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'effective_branches' AND column_name = 'travel_cost'
			) THEN
				ALTER TABLE effective_branches ADD COLUMN travel_cost DECIMAL(10,2) DEFAULT 1000.00;
			END IF;
		END $$;
	`)
	if err != nil {
		return fmt.Errorf("failed to add travel parameters columns to effective_branches: %w", err)
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

	// Drop address and expected_revenue columns from branches table if they exist
	_, err = db.Exec(`
		DO $$ 
		BEGIN
			IF EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'branches' AND column_name = 'address'
			) THEN
				ALTER TABLE branches DROP COLUMN address;
			END IF;
			
			IF EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'branches' AND column_name = 'expected_revenue'
			) THEN
				ALTER TABLE branches DROP COLUMN expected_revenue;
			END IF;
		END $$;
	`)
	if err != nil {
		return fmt.Errorf("failed to drop address and expected_revenue columns: %w", err)
	}

	// Add min_doctor_assistant column to branch_constraints table if it doesn't exist
	_, err = db.Exec(`
		DO $$ 
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'branch_constraints' AND column_name = 'min_doctor_assistant'
			) THEN
				ALTER TABLE branch_constraints ADD COLUMN min_doctor_assistant INTEGER NOT NULL DEFAULT 0 CHECK (min_doctor_assistant >= 0);
			END IF;
		END $$;
	`)
	if err != nil {
		return fmt.Errorf("failed to add min_doctor_assistant column: %w", err)
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
    position_type VARCHAR(20) NOT NULL DEFAULT 'branch' CHECK (position_type IN ('branch', 'rotation')),
    manpower_type VARCHAR(50) NOT NULL DEFAULT 'อื่นๆ' CHECK (manpower_type IN ('พนักงานฟร้อนท์', 'ผู้ช่วยแพทย์', 'อื่นๆ', 'ทำความสะอาด')),
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

const createZonesTable = `
CREATE TABLE IF NOT EXISTS zones (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50) NOT NULL UNIQUE,
    description TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
`

const createZoneBranchesTable = `
CREATE TABLE IF NOT EXISTS zone_branches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    zone_id UUID NOT NULL REFERENCES zones(id) ON DELETE CASCADE,
    branch_id UUID NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(zone_id, branch_id)
);
`

const createAreaOfOperationZonesTable = `
CREATE TABLE IF NOT EXISTS area_of_operation_zones (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    area_of_operation_id UUID NOT NULL REFERENCES areas_of_operation(id) ON DELETE CASCADE,
    zone_id UUID NOT NULL REFERENCES zones(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(area_of_operation_id, zone_id)
);
`

const createAreaOfOperationBranchesTable = `
CREATE TABLE IF NOT EXISTS area_of_operation_branches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    area_of_operation_id UUID NOT NULL REFERENCES areas_of_operation(id) ON DELETE CASCADE,
    branch_id UUID NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(area_of_operation_id, branch_id)
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
    zone_id UUID REFERENCES zones(id),
    skill_level INTEGER DEFAULT 5 CHECK (skill_level >= 0 AND skill_level <= 10),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
`

const createStaffBranchesTable = `
CREATE TABLE IF NOT EXISTS staff_branches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    staff_id UUID NOT NULL REFERENCES staff(id) ON DELETE CASCADE,
    branch_id UUID NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(staff_id, branch_id)
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
    revenue_source VARCHAR(20) DEFAULT 'branch' CHECK (revenue_source IN ('branch', 'doctor', 'excel')),
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

const createAllocationCriteriaTable = `
CREATE TABLE IF NOT EXISTS allocation_criteria (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pillar VARCHAR(50) NOT NULL CHECK (pillar IN ('clinic_wide', 'doctor_specific', 'branch_specific')),
    type VARCHAR(50) NOT NULL CHECK (type IN ('bookings', 'revenue', 'min_staff_position', 'min_staff_branch', 'doctor_count')),
    weight DECIMAL(5,4) NOT NULL DEFAULT 0.0 CHECK (weight >= 0.0 AND weight <= 1.0),
    is_active BOOLEAN NOT NULL DEFAULT true,
    description TEXT,
    config TEXT, -- JSON config for criteria-specific settings
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
`

const createPositionQuotasTable = `
CREATE TABLE IF NOT EXISTS position_quotas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    branch_id UUID NOT NULL REFERENCES branches(id),
    position_id UUID NOT NULL REFERENCES positions(id),
    designated_quota INTEGER NOT NULL DEFAULT 0 CHECK (designated_quota >= 0),
    minimum_required INTEGER NOT NULL DEFAULT 0 CHECK (minimum_required >= 0),
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(branch_id, position_id)
);
`

const createDoctorAssignmentsTable = `
CREATE TABLE IF NOT EXISTS doctor_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    doctor_id UUID NOT NULL REFERENCES doctors(id) ON DELETE CASCADE,
    branch_id UUID NOT NULL REFERENCES branches(id),
    date DATE NOT NULL,
    expected_revenue DECIMAL(15,2) NOT NULL DEFAULT 0,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(doctor_id, branch_id, date)
);
CREATE INDEX IF NOT EXISTS idx_doctor_assignments_doctor_id ON doctor_assignments(doctor_id);
CREATE INDEX IF NOT EXISTS idx_doctor_assignments_branch_id ON doctor_assignments(branch_id);
CREATE INDEX IF NOT EXISTS idx_doctor_assignments_date ON doctor_assignments(date);
`

const createDoctorsTable = `
CREATE TABLE IF NOT EXISTS doctors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    code VARCHAR(100) UNIQUE,
    specialization VARCHAR(255),
    contact_info TEXT,
    preferences JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
`

const createDoctorPreferencesTable = `
CREATE TABLE IF NOT EXISTS doctor_preferences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    doctor_id UUID NOT NULL REFERENCES doctors(id) ON DELETE CASCADE,
    branch_id UUID REFERENCES branches(id) ON DELETE CASCADE,
    rule_type VARCHAR(100) NOT NULL,
    rule_config JSONB NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_doctor_preferences_doctor_id ON doctor_preferences(doctor_id);
CREATE INDEX IF NOT EXISTS idx_doctor_preferences_branch_id ON doctor_preferences(branch_id);
`

const createDoctorOnOffDaysTable = `
CREATE TABLE IF NOT EXISTS doctor_on_off_days (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    branch_id UUID NOT NULL REFERENCES branches(id),
    date DATE NOT NULL,
    is_doctor_on BOOLEAN NOT NULL DEFAULT true,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(branch_id, date)
);
`

const createDoctorDefaultSchedulesTable = `
CREATE TABLE IF NOT EXISTS doctor_default_schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    doctor_id UUID NOT NULL REFERENCES doctors(id) ON DELETE CASCADE,
    day_of_week INTEGER NOT NULL CHECK (day_of_week >= 0 AND day_of_week <= 6),
    branch_id UUID NOT NULL REFERENCES branches(id),
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(doctor_id, day_of_week)
);
CREATE INDEX IF NOT EXISTS idx_doctor_default_schedules_doctor_id ON doctor_default_schedules(doctor_id);
CREATE INDEX IF NOT EXISTS idx_doctor_default_schedules_branch_id ON doctor_default_schedules(branch_id);
CREATE INDEX IF NOT EXISTS idx_doctor_default_schedules_day_of_week ON doctor_default_schedules(day_of_week);
`

const createDoctorWeeklyOffDaysTable = `
CREATE TABLE IF NOT EXISTS doctor_weekly_off_days (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    doctor_id UUID NOT NULL REFERENCES doctors(id) ON DELETE CASCADE,
    day_of_week INTEGER NOT NULL CHECK (day_of_week >= 0 AND day_of_week <= 6),
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(doctor_id, day_of_week)
);
CREATE INDEX IF NOT EXISTS idx_doctor_weekly_off_days_doctor_id ON doctor_weekly_off_days(doctor_id);
CREATE INDEX IF NOT EXISTS idx_doctor_weekly_off_days_day_of_week ON doctor_weekly_off_days(day_of_week);
`

const createDoctorScheduleOverridesTable = `
CREATE TABLE IF NOT EXISTS doctor_schedule_overrides (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    doctor_id UUID NOT NULL REFERENCES doctors(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('working', 'off')),
    branch_id UUID REFERENCES branches(id),
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(doctor_id, date),
    CONSTRAINT check_working_branch CHECK (
        (type = 'working' AND branch_id IS NOT NULL) OR
        (type = 'off' AND branch_id IS NULL)
    )
);
CREATE INDEX IF NOT EXISTS idx_doctor_schedule_overrides_doctor_id ON doctor_schedule_overrides(doctor_id);
CREATE INDEX IF NOT EXISTS idx_doctor_schedule_overrides_date ON doctor_schedule_overrides(date);
CREATE INDEX IF NOT EXISTS idx_doctor_schedule_overrides_branch_id ON doctor_schedule_overrides(branch_id);
`

const createAllocationSuggestionsTable = `
CREATE TABLE IF NOT EXISTS allocation_suggestions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rotation_staff_id UUID NOT NULL REFERENCES staff(id),
    branch_id UUID NOT NULL REFERENCES branches(id),
    date DATE NOT NULL,
    position_id UUID NOT NULL REFERENCES positions(id),
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected', 'modified')),
    confidence DECIMAL(5,4) NOT NULL DEFAULT 0.0 CHECK (confidence >= 0.0 AND confidence <= 1.0),
    reason TEXT,
    criteria_used TEXT, -- JSON array of criteria IDs used
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    reviewed_by UUID REFERENCES users(id),
    reviewed_at TIMESTAMP,
    UNIQUE(rotation_staff_id, branch_id, date, position_id)
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
INSERT INTO positions (id, name, min_staff_per_branch, revenue_multiplier, display_order, position_type, manpower_type) VALUES
    -- Thai positions only (English positions removed)
    ('10000000-0000-0000-0000-000000000008', 'ผู้จัดการสาขา', 1, 0, 1, 'branch', 'อื่นๆ'),
    ('10000000-0000-0000-0000-000000000009', 'รองผู้จัดการสาขา', 0, 0.5, 2, 'branch', 'อื่นๆ'),
    ('10000000-0000-0000-0000-000000000010', 'ผู้ช่วยแพทย์', 2, 1.2, 20, 'branch', 'ผู้ช่วยแพทย์'),
    ('10000000-0000-0000-0000-000000000011', 'ผู้ประสานงานคลินิก (Clinic Coordination Officer)', 1, 0.8, 50, 'branch', 'อื่นๆ'),
    ('10000000-0000-0000-0000-000000000012', 'ผู้ช่วย Laser Specialist', 1, 1.0, 20, 'branch', 'ผู้ช่วยแพทย์'),
    ('10000000-0000-0000-0000-000000000013', 'พนักงานต้อนรับ (Laser Receptionist)', 1, 0.5, 40, 'branch', 'พนักงานฟร้อนท์'),
    ('10000000-0000-0000-0000-000000000014', 'แม่บ้านประจำสาขา', 1, 0.3, 60, 'branch', 'ทำความสะอาด'),
    ('10000000-0000-0000-0000-000000000015', 'พยาบาล', 2, 1.0, 30, 'branch', 'ผู้ช่วยแพทย์'),
    ('10000000-0000-0000-0000-000000000016', 'ผู้ช่วยผู้จัดการสาขา', 0, 0.5, 2, 'branch', 'อื่นๆ'),
    ('10000000-0000-0000-0000-000000000017', 'ผู้ช่วยแพทย์ Pico Laser', 1, 1.2, 20, 'branch', 'ผู้ช่วยแพทย์'),
    ('10000000-0000-0000-0000-000000000018', 'รองผู้จัดการสาขาและล่าม', 0, 0.6, 2, 'branch', 'อื่นๆ'),
    ('10000000-0000-0000-0000-000000000019', 'Front+ล่ามวนสาขา', 1, 0.7, 40, 'rotation', 'พนักงานฟร้อนท์'),
    ('10000000-0000-0000-0000-000000000020', 'ผู้ช่วยแพทย์ Pico', 1, 1.2, 20, 'branch', 'ผู้ช่วยแพทย์'),
    ('10000000-0000-0000-0000-000000000021', 'พนักงานต้อนรับ (Pico Laser Receptionist)', 1, 0.5, 40, 'branch', 'พนักงานฟร้อนท์'),
    ('10000000-0000-0000-0000-000000000022', 'ผู้จัดการเขต', 0, 0, 10, 'rotation', 'อื่นๆ'),
    ('10000000-0000-0000-0000-000000000023', 'ผู้จัดการแผนกและกำกับพัฒนาระเบียบสาขา', 0, 0, 10, 'rotation', 'อื่นๆ'),
    ('10000000-0000-0000-0000-000000000024', 'หัวหน้าผู้ช่วยแพทย์', 1, 1.5, 15, 'rotation', 'ผู้ช่วยแพทย์'),
    ('10000000-0000-0000-0000-000000000025', 'ผู้ช่วยพิเศษ', 0, 0.8, 20, 'rotation', 'อื่นๆ'),
    ('10000000-0000-0000-0000-000000000026', 'ผู้ช่วยแพทย์วนสาขา', 1, 1.2, 20, 'rotation', 'ผู้ช่วยแพทย์'),
    ('10000000-0000-0000-0000-000000000027', 'ฟร้อนท์วนสาขา', 1, 0.5, 40, 'rotation', 'พนักงานฟร้อนท์')
ON CONFLICT (id) DO NOTHING;
`

const createBranchWeeklyRevenueTable = `
CREATE TABLE IF NOT EXISTS branch_weekly_revenue (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    branch_id UUID NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
    day_of_week INTEGER NOT NULL CHECK (day_of_week >= 0 AND day_of_week <= 6),
    expected_revenue DECIMAL(15,2) NOT NULL DEFAULT 0 CHECK (expected_revenue >= 0),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(branch_id, day_of_week)
);
`

const createBranchConstraintsTable = `
CREATE TABLE IF NOT EXISTS branch_constraints (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    branch_id UUID NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
    day_of_week INTEGER NOT NULL CHECK (day_of_week >= 0 AND day_of_week <= 6),
    min_front_staff INTEGER NOT NULL DEFAULT 0 CHECK (min_front_staff >= 0),
    min_managers INTEGER NOT NULL DEFAULT 0 CHECK (min_managers >= 0),
    min_doctor_assistant INTEGER NOT NULL DEFAULT 0 CHECK (min_doctor_assistant >= 0),
    min_total_staff INTEGER NOT NULL DEFAULT 0 CHECK (min_total_staff >= 0),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(branch_id, day_of_week)
);
`

const createRevenueLevelTiersTable = `
CREATE TABLE IF NOT EXISTS revenue_level_tiers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    level_number INTEGER NOT NULL UNIQUE CHECK (level_number >= 1 AND level_number <= 10),
    level_name VARCHAR(50) NOT NULL,
    min_revenue DECIMAL(15,2) NOT NULL CHECK (min_revenue >= 0),
    max_revenue DECIMAL(15,2),
    display_order INTEGER NOT NULL DEFAULT 0,
    color_code VARCHAR(20),
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_revenue_level_tiers_level ON revenue_level_tiers(level_number);
CREATE INDEX IF NOT EXISTS idx_revenue_level_tiers_range ON revenue_level_tiers(min_revenue, max_revenue);
`

const createStaffRequirementScenariosTable = `
CREATE TABLE IF NOT EXISTS staff_requirement_scenarios (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    scenario_name VARCHAR(100) NOT NULL,
    description TEXT,
    revenue_level_tier_id UUID REFERENCES revenue_level_tiers(id),
    min_revenue DECIMAL(15,2),
    max_revenue DECIMAL(15,2),
    use_day_of_week_revenue BOOLEAN NOT NULL DEFAULT true,
    use_specific_date_revenue BOOLEAN NOT NULL DEFAULT false,
    doctor_count INTEGER,
    min_doctor_count INTEGER,
    day_of_week INTEGER CHECK (day_of_week >= 0 AND day_of_week <= 6),
    is_default BOOLEAN NOT NULL DEFAULT false,
    is_active BOOLEAN NOT NULL DEFAULT true,
    priority INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT check_revenue_criteria CHECK (
        revenue_level_tier_id IS NOT NULL OR 
        min_revenue IS NOT NULL OR 
        is_default = true
    )
);
CREATE INDEX IF NOT EXISTS idx_staff_requirement_scenarios_priority ON staff_requirement_scenarios(priority DESC, is_active);
CREATE INDEX IF NOT EXISTS idx_staff_requirement_scenarios_tier ON staff_requirement_scenarios(revenue_level_tier_id);
CREATE INDEX IF NOT EXISTS idx_staff_requirement_scenarios_day ON staff_requirement_scenarios(day_of_week);
`

const createScenarioPositionRequirementsTable = `
CREATE TABLE IF NOT EXISTS scenario_position_requirements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    scenario_id UUID NOT NULL REFERENCES staff_requirement_scenarios(id) ON DELETE CASCADE,
    position_id UUID NOT NULL REFERENCES positions(id),
    preferred_staff INTEGER NOT NULL DEFAULT 0 CHECK (preferred_staff >= 0),
    minimum_staff INTEGER NOT NULL DEFAULT 0 CHECK (minimum_staff >= 0),
    override_base BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(scenario_id, position_id)
);
CREATE INDEX IF NOT EXISTS idx_scenario_position_requirements_scenario ON scenario_position_requirements(scenario_id);
CREATE INDEX IF NOT EXISTS idx_scenario_position_requirements_position ON scenario_position_requirements(position_id);
`

const insertDefaultRevenueLevelTiers = `
INSERT INTO revenue_level_tiers (id, level_number, level_name, min_revenue, max_revenue, display_order, color_code, description) VALUES
    ('30000000-0000-0000-0000-000000000001', 1, 'Very Low', 0, 100000, 1, '#CCCCCC', 'Low revenue days'),
    ('30000000-0000-0000-0000-000000000002', 2, 'Low', 100000, 200000, 2, '#99CCFF', 'Below average revenue'),
    ('30000000-0000-0000-0000-000000000003', 3, 'Medium', 200000, 300000, 3, '#66FF99', 'Average revenue days'),
    ('30000000-0000-0000-0000-000000000004', 4, 'High', 300000, 400000, 4, '#FFCC66', 'Above average revenue'),
    ('30000000-0000-0000-0000-000000000005', 5, 'Very High', 400000, 500000, 5, '#FF9966', 'High revenue days'),
    ('30000000-0000-0000-0000-000000000006', 6, 'Extremely High', 500000, 600000, 6, '#FF6666', 'Very high revenue days'),
    ('30000000-0000-0000-0000-000000000007', 7, 'Peak', 600000, NULL, 7, '#FF0000', 'Peak revenue days')
ON CONFLICT (level_number) DO NOTHING;
`

const insertDefaultStaffRequirementScenarios = `
INSERT INTO staff_requirement_scenarios (id, scenario_name, description, is_default, priority) VALUES
    ('40000000-0000-0000-0000-000000000001', 'Normal Day', 'Standard staffing for normal operations', true, 0)
ON CONFLICT DO NOTHING;
`

// Helper function to get revenue level tier for a given revenue amount
const createGetRevenueLevelTierFunction = `
CREATE OR REPLACE FUNCTION get_revenue_level_tier(revenue_amount DECIMAL(15,2))
RETURNS UUID AS $$
DECLARE
    tier_id UUID;
BEGIN
    SELECT id INTO tier_id
    FROM revenue_level_tiers
    WHERE revenue_amount >= min_revenue
      AND (max_revenue IS NULL OR revenue_amount < max_revenue)
    ORDER BY level_number DESC
    LIMIT 1;
    
    RETURN tier_id;
END;
$$ LANGUAGE plpgsql;
`

// Helper function to check if scenario matches given conditions
const createScenarioMatchesFunction = `
CREATE OR REPLACE FUNCTION scenario_matches(
    p_scenario_id UUID,
    p_day_of_week_revenue DECIMAL(15,2),
    p_specific_date_revenue DECIMAL(15,2),
    p_doctor_count INTEGER,
    p_day_of_week INTEGER
)
RETURNS BOOLEAN AS $$
DECLARE
    v_scenario RECORD;
    v_revenue_to_check DECIMAL(15,2);
    v_revenue_tier_id UUID;
BEGIN
    -- Get scenario
    SELECT * INTO v_scenario
    FROM staff_requirement_scenarios
    WHERE id = p_scenario_id AND is_active = true;
    
    IF NOT FOUND THEN
        RETURN false;
    END IF;
    
    -- Check day of week filter
    IF v_scenario.day_of_week IS NOT NULL AND v_scenario.day_of_week != p_day_of_week THEN
        RETURN false;
    END IF;
    
    -- Determine which revenue to use
    IF v_scenario.use_specific_date_revenue AND p_specific_date_revenue IS NOT NULL THEN
        v_revenue_to_check := p_specific_date_revenue;
    ELSIF v_scenario.use_day_of_week_revenue THEN
        v_revenue_to_check := p_day_of_week_revenue;
    ELSE
        v_revenue_to_check := COALESCE(p_specific_date_revenue, p_day_of_week_revenue);
    END IF;
    
    -- Check revenue tier match
    IF v_scenario.revenue_level_tier_id IS NOT NULL THEN
        v_revenue_tier_id := get_revenue_level_tier(v_revenue_to_check);
        IF v_revenue_tier_id IS NULL OR v_revenue_tier_id != v_scenario.revenue_level_tier_id THEN
            RETURN false;
        END IF;
    END IF;
    
    -- Check direct revenue range
    IF v_scenario.min_revenue IS NOT NULL THEN
        IF v_revenue_to_check < v_scenario.min_revenue THEN
            RETURN false;
        END IF;
    END IF;
    IF v_scenario.max_revenue IS NOT NULL THEN
        IF v_revenue_to_check >= v_scenario.max_revenue THEN
            RETURN false;
        END IF;
    END IF;
    
    -- Check doctor count
    IF v_scenario.doctor_count IS NOT NULL THEN
        IF p_doctor_count != v_scenario.doctor_count THEN
            RETURN false;
        END IF;
    END IF;
    IF v_scenario.min_doctor_count IS NOT NULL THEN
        IF p_doctor_count < v_scenario.min_doctor_count THEN
            RETURN false;
        END IF;
    END IF;
    
    RETURN true;
END;
$$ LANGUAGE plpgsql;
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
	stmt := `INSERT INTO branches (id, name, code, priority) 
	         VALUES ($1, $2, $3, $4)
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


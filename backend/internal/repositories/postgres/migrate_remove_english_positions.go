package postgres

import (
	"database/sql"
	"fmt"
)

// PositionMapping maps English position IDs to Thai equivalent IDs
var PositionMapping = map[string]string{
	// English → Thai mappings
	"10000000-0000-0000-0000-000000000001": "10000000-0000-0000-0000-000000000008", // Branch Manager → ผู้จัดการสาขา
	"10000000-0000-0000-0000-000000000002": "10000000-0000-0000-0000-000000000009", // Assistant Branch Manager → รองผู้จัดการสาขา
	"10000000-0000-0000-0000-000000000003": "10000000-0000-0000-0000-000000000011", // Service Consultant → ผู้ประสานงานคลินิก (closest match)
	"10000000-0000-0000-0000-000000000004": "10000000-0000-0000-0000-000000000011", // Coordinator → ผู้ประสานงานคลินิก
	"10000000-0000-0000-0000-000000000005": "10000000-0000-0000-0000-000000000010", // Doctor Assistant → ผู้ช่วยแพทย์
	"10000000-0000-0000-0000-000000000006": "10000000-0000-0000-0000-000000000010", // Physiotherapist → ผู้ช่วยแพทย์ (closest match, may need review)
	"10000000-0000-0000-0000-000000000007": "10000000-0000-0000-0000-000000000015", // Nurse → พยาบาล
	"10000000-0000-0000-0000-000000000028": "10000000-0000-0000-0000-000000000027", // Front 3 → ฟร้อนท์วนสาขา (closest match)
	"10000000-0000-0000-0000-000000000029": "10000000-0000-0000-0000-000000000013", // Front Laser → พนักงานต้อนรับ (Laser Receptionist)
	"10000000-0000-0000-0000-000000000030": "10000000-0000-0000-0000-000000000012", // Laser Assistant → ผู้ช่วย Laser Specialist
}

// EnglishPositionIDs lists all English position IDs to be removed
var EnglishPositionIDs = []string{
	"10000000-0000-0000-0000-000000000001", // Branch Manager
	"10000000-0000-0000-0000-000000000002", // Assistant Branch Manager
	"10000000-0000-0000-0000-000000000003", // Service Consultant
	"10000000-0000-0000-0000-000000000004", // Coordinator
	"10000000-0000-0000-0000-000000000005", // Doctor Assistant
	"10000000-0000-0000-0000-000000000006", // Physiotherapist
	"10000000-0000-0000-0000-000000000007", // Nurse
	"10000000-0000-0000-0000-000000000028", // Front 3
	"10000000-0000-0000-0000-000000000029", // Front Laser
	"10000000-0000-0000-0000-000000000030", // Laser Assistant
	"10000000-0000-0000-0000-000000000031", // ฟร้อนท์ 3 (Thai version - to be removed)
	"10000000-0000-0000-0000-000000000032", // ฟร้อนท์ Laser (Thai version - to be removed)
}

// MigrateRemoveEnglishPositions migrates data from English positions to Thai positions and removes English positions
func MigrateRemoveEnglishPositions(db *sql.DB) error {
	// Check if there's anything to migrate first - check if any English positions exist
	var hasEnglishPositions int
	checkQuery := `SELECT COUNT(*) FROM positions WHERE id IN (`
	for i, posID := range EnglishPositionIDs {
		if i > 0 {
			checkQuery += ", "
		}
		checkQuery += "'" + posID + "'"
	}
	checkQuery += ")"
	err := db.QueryRow(checkQuery).Scan(&hasEnglishPositions)
	if err != nil {
		// If query fails, continue anyway (might be first run or table doesn't exist yet)
		hasEnglishPositions = 1
	}

	// If no English positions exist, migration is already complete - skip silently
	if hasEnglishPositions == 0 {
		return nil
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Step 1: Update staff.position_id
	fmt.Println("Updating staff records...")
	totalStaffUpdated := 0
	for englishID, thaiID := range PositionMapping {
		query := `UPDATE staff SET position_id = $1 WHERE position_id = $2`
		result, err := tx.Exec(query, thaiID, englishID)
		if err != nil {
			return fmt.Errorf("failed to update staff for position %s: %w", englishID, err)
		}
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected > 0 {
			fmt.Printf("  Updated %d staff records: %s → %s\n", rowsAffected, englishID, thaiID)
			totalStaffUpdated += int(rowsAffected)
		}
	}
	if totalStaffUpdated == 0 {
		fmt.Println("  No staff records to update")
	}

	// Step 2: Update position_quotas.position_id
	fmt.Println("Updating position_quotas records...")
	totalQuotasUpdated := 0
	for englishID, thaiID := range PositionMapping {
		// Check if quota already exists for Thai position
		var existingCount int
		quotaCheckQuery := `SELECT COUNT(*) FROM position_quotas WHERE branch_id IN (
			SELECT branch_id FROM position_quotas WHERE position_id = $1
		) AND position_id = $2`
		err := tx.QueryRow(quotaCheckQuery, englishID, thaiID).Scan(&existingCount)
		if err != nil {
			return fmt.Errorf("failed to check existing quotas: %w", err)
		}

		if existingCount == 0 {
			// No conflict, update directly
			query := `UPDATE position_quotas SET position_id = $1 WHERE position_id = $2`
			result, err := tx.Exec(query, thaiID, englishID)
			if err != nil {
				return fmt.Errorf("failed to update position_quotas for position %s: %w", englishID, err)
			}
			rowsAffected, _ := result.RowsAffected()
			if rowsAffected > 0 {
				fmt.Printf("  Updated %d quota records: %s → %s\n", rowsAffected, englishID, thaiID)
				totalQuotasUpdated += int(rowsAffected)
			}
		} else {
			// Conflict exists, delete English quotas (Thai ones take precedence)
			query := `DELETE FROM position_quotas WHERE position_id = $1`
			result, err := tx.Exec(query, englishID)
			if err != nil {
				return fmt.Errorf("failed to delete conflicting position_quotas for position %s: %w", englishID, err)
			}
			rowsAffected, _ := result.RowsAffected()
			if rowsAffected > 0 {
				fmt.Printf("  Deleted %d conflicting quota records for position %s (Thai quota exists)\n", rowsAffected, englishID)
				totalQuotasUpdated += int(rowsAffected)
			}
		}
	}
	if totalQuotasUpdated == 0 {
		fmt.Println("  No position_quotas records to update")
	}

	// Step 3: Update allocation_suggestions.position_id (skipped - table removed)
	fmt.Println("Skipping allocation_suggestions update (table has been removed)")

	// Step 4: Update staff_allocation_rules.position_id (if exists)
	fmt.Println("Updating staff_allocation_rules records...")
	totalRulesUpdated := 0
	for englishID, thaiID := range PositionMapping {
		// Check if rule already exists for Thai position
		var existingCount int
		rulesCheckQuery := `SELECT COUNT(*) FROM staff_allocation_rules WHERE position_id = $1`
		err := tx.QueryRow(rulesCheckQuery, thaiID).Scan(&existingCount)
		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("failed to check existing rules: %w", err)
		}

		if existingCount == 0 {
			// No conflict, update directly
			query := `UPDATE staff_allocation_rules SET position_id = $1 WHERE position_id = $2`
			result, err := tx.Exec(query, thaiID, englishID)
			if err != nil {
				// Table might not exist or have data, continue
				continue
			}
			rowsAffected, _ := result.RowsAffected()
			if rowsAffected > 0 {
				fmt.Printf("  Updated %d rule records: %s → %s\n", rowsAffected, englishID, thaiID)
				totalRulesUpdated += int(rowsAffected)
			}
		} else {
			// Conflict exists, delete English rules
			query := `DELETE FROM staff_allocation_rules WHERE position_id = $1`
			result, err := tx.Exec(query, englishID)
			if err != nil {
				continue
			}
			rowsAffected, _ := result.RowsAffected()
			if rowsAffected > 0 {
				fmt.Printf("  Deleted %d conflicting rule records for position %s\n", rowsAffected, englishID)
				totalRulesUpdated += int(rowsAffected)
			}
		}
	}
	if totalRulesUpdated == 0 {
		fmt.Println("  No staff_allocation_rules records to update")
	}

	// Step 5: Migrate Thai versions of Front 3 and Front Laser to existing positions
	fmt.Println("Migrating Thai Front positions...")
	thaiFrontMappings := map[string]string{
		"10000000-0000-0000-0000-000000000031": "10000000-0000-0000-0000-000000000027", // ฟร้อนท์ 3 → ฟร้อนท์วนสาขา
		"10000000-0000-0000-0000-000000000032": "10000000-0000-0000-0000-000000000013", // ฟร้อนท์ Laser → พนักงานต้อนรับ (Laser Receptionist)
	}
	totalThaiFrontUpdated := 0
	for thaiID, targetID := range thaiFrontMappings {
		// Update staff
		query := `UPDATE staff SET position_id = $1 WHERE position_id = $2`
		result, err := tx.Exec(query, targetID, thaiID)
		if err != nil {
			return fmt.Errorf("failed to update staff for position %s: %w", thaiID, err)
		}
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected > 0 {
			fmt.Printf("  Updated %d staff records: %s → %s\n", rowsAffected, thaiID, targetID)
			totalThaiFrontUpdated += int(rowsAffected)
		}

		// Update position_quotas (handle conflicts)
		var existingCount int
		thaiFrontCheckQuery := `SELECT COUNT(*) FROM position_quotas WHERE branch_id IN (
			SELECT branch_id FROM position_quotas WHERE position_id = $1
		) AND position_id = $2`
		err = tx.QueryRow(thaiFrontCheckQuery, thaiID, targetID).Scan(&existingCount)
		if err != nil {
			return fmt.Errorf("failed to check existing quotas: %w", err)
		}

		if existingCount == 0 {
			query = `UPDATE position_quotas SET position_id = $1 WHERE position_id = $2`
			result, err = tx.Exec(query, targetID, thaiID)
			if err != nil {
				return fmt.Errorf("failed to update position_quotas for position %s: %w", thaiID, err)
			}
			rowsAffected, _ = result.RowsAffected()
			if rowsAffected > 0 {
				fmt.Printf("  Updated %d quota records: %s → %s\n", rowsAffected, thaiID, targetID)
				totalThaiFrontUpdated += int(rowsAffected)
			}
		} else {
			query = `DELETE FROM position_quotas WHERE position_id = $1`
			result, err = tx.Exec(query, thaiID)
			if err != nil {
				return fmt.Errorf("failed to delete conflicting position_quotas for position %s: %w", thaiID, err)
			}
			rowsAffected, _ = result.RowsAffected()
			if rowsAffected > 0 {
				fmt.Printf("  Deleted %d conflicting quota records for position %s\n", rowsAffected, thaiID)
				totalThaiFrontUpdated += int(rowsAffected)
			}
		}

		// Update allocation_suggestions (skipped - table removed)
		// Table has been removed, skipping update
		rowsAffected, _ = result.RowsAffected()
		if rowsAffected > 0 {
			fmt.Printf("  Updated %d suggestion records: %s → %s\n", rowsAffected, thaiID, targetID)
			totalThaiFrontUpdated += int(rowsAffected)
		}
	}
	if totalThaiFrontUpdated == 0 {
		fmt.Println("  No Thai Front positions to migrate")
	}

	// Step 6: Delete English positions and Thai Front positions
	fmt.Println("Deleting positions to be removed...")
	totalPositionsDeleted := 0
	for _, positionID := range EnglishPositionIDs {
		query := `DELETE FROM positions WHERE id = $1`
		result, err := tx.Exec(query, positionID)
		if err != nil {
			return fmt.Errorf("failed to delete position %s: %w", positionID, err)
		}
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected > 0 {
			fmt.Printf("  Deleted position: %s\n", positionID)
			totalPositionsDeleted++
		}
	}
	if totalPositionsDeleted == 0 {
		fmt.Println("  No positions to delete")
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	fmt.Println("Migration completed successfully!")
	return nil
}

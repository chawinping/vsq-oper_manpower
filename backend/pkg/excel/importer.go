package excel

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
	"vsq-oper-manpower/backend/internal/domain/models"
)

// ExcelImporter handles importing staff data from Excel files
type ExcelImporter struct{}

func NewExcelImporter() *ExcelImporter {
	return &ExcelImporter{}
}

// ImportStaff parses Excel file and returns staff records
// Expected format:
// - Row 1: Header row (Name, Staff Type, Position ID, Branch ID, Coverage Area)
// - Row 2+: Data rows
// - Column A: Name (required)
// - Column B: Staff Type (branch/rotation) (required)
// - Column C: Position ID (UUID) (required)
// - Column D: Branch ID (UUID, optional for branch staff)
// - Column E: Coverage Area (string, optional for rotation staff)
func (e *ExcelImporter) ImportStaff(fileData []byte) ([]*models.Staff, error) {
	// Open Excel file from byte data
	f, err := excelize.OpenReader(bytes.NewReader(fileData))
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer f.Close()

	// Get the first sheet
	sheetName := f.GetSheetName(0)
	if sheetName == "" {
		return nil, fmt.Errorf("Excel file has no sheets")
	}

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to read rows: %w", err)
	}

	if len(rows) < 2 {
		return nil, fmt.Errorf("Excel file must have at least a header row and one data row")
	}

	var staffList []*models.Staff
	var errors []string

	// Skip header row, start from row 2 (index 1)
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		if len(row) < 3 {
			errors = append(errors, fmt.Sprintf("Row %d: insufficient columns (need at least 3)", i+1))
			continue
		}

		staff := &models.Staff{
			ID: uuid.New(),
		}

		// Column A: Name
		staff.Name = strings.TrimSpace(row[0])
		if staff.Name == "" {
			errors = append(errors, fmt.Sprintf("Row %d: name is required", i+1))
			continue
		}

		// Column B: Staff Type
		staffTypeStr := strings.TrimSpace(strings.ToLower(row[1]))
		if staffTypeStr == "branch" {
			staff.StaffType = models.StaffTypeBranch
		} else if staffTypeStr == "rotation" {
			staff.StaffType = models.StaffTypeRotation
		} else {
			errors = append(errors, fmt.Sprintf("Row %d: invalid staff type '%s' (must be 'branch' or 'rotation')", i+1, row[1]))
			continue
		}

		// Column C: Position ID
		if len(row) > 2 && strings.TrimSpace(row[2]) != "" {
			positionID, err := uuid.Parse(strings.TrimSpace(row[2]))
			if err != nil {
				errors = append(errors, fmt.Sprintf("Row %d: invalid position ID '%s': %v", i+1, row[2], err))
				continue
			}
			staff.PositionID = positionID
		} else {
			errors = append(errors, fmt.Sprintf("Row %d: position ID is required", i+1))
			continue
		}

		// Column D: Branch ID (optional, mainly for branch staff)
		if len(row) > 3 && strings.TrimSpace(row[3]) != "" {
			branchID, err := uuid.Parse(strings.TrimSpace(row[3]))
			if err != nil {
				errors = append(errors, fmt.Sprintf("Row %d: invalid branch ID '%s': %v", i+1, row[3], err))
				continue
			}
			staff.BranchID = &branchID
		}

		// Column E: Coverage Area (optional, mainly for rotation staff)
		if len(row) > 4 && strings.TrimSpace(row[4]) != "" {
			staff.CoverageArea = strings.TrimSpace(row[4])
		}

		// Validate staff data
		if err := e.ValidateStaffData(staff); err != nil {
			errors = append(errors, fmt.Sprintf("Row %d: validation error: %v", i+1, err))
			continue
		}

		staffList = append(staffList, staff)
	}

	if len(errors) > 0 && len(staffList) == 0 {
		return nil, fmt.Errorf("failed to import any staff: %s", strings.Join(errors, "; "))
	}

	if len(errors) > 0 {
		// Return partial success with errors
		return staffList, fmt.Errorf("import completed with errors: %s", strings.Join(errors, "; "))
	}

	return staffList, nil
}

// ValidateStaffData validates imported staff data
func (e *ExcelImporter) ValidateStaffData(staff *models.Staff) error {
	if staff.Name == "" {
		return fmt.Errorf("staff name is required")
	}

	if staff.StaffType != models.StaffTypeBranch && staff.StaffType != models.StaffTypeRotation {
		return fmt.Errorf("invalid staff type: %s", staff.StaffType)
	}

	if staff.StaffType == models.StaffTypeBranch && staff.BranchID == nil {
		// Branch staff should have a branch ID, but we'll allow it to be set later
	}

	return nil
}


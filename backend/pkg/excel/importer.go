package excel

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
	"vsq-oper-manpower/backend/internal/domain/interfaces"
	"vsq-oper-manpower/backend/internal/domain/models"
)

// ExcelImporter handles importing staff data from Excel files
type ExcelImporter struct {
	positionRepo interfaces.PositionRepository
	branchRepo   interfaces.BranchRepository
}

func NewExcelImporter(
	positionRepo interfaces.PositionRepository,
	branchRepo interfaces.BranchRepository,
) *ExcelImporter {
	return &ExcelImporter{
		positionRepo: positionRepo,
		branchRepo:   branchRepo,
	}
}

// ImportStaff parses Excel file and returns staff records
// Expected format:
// - Row 1: Header row (optional, will be skipped)
// - Row 2+: Data rows
// - Column A: Name (required)
// - Column B: Staff Type (branch/rotation) (required)
// - Column C: Position Name (string, required) - e.g., "Nurse", "Doctor"
// - Column D: Branch Code (string, optional for branch staff) - e.g., "TMA", "CPN"
// - Column E: Nickname (string, optional)
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

	if len(rows) < 1 {
		return nil, fmt.Errorf("Excel file is empty")
	}

	var staffList []*models.Staff
	var errors []string

	// Detect if first row is a header row (contains common header keywords)
	startRow := 0
	if len(rows) > 0 {
		firstRow := rows[0]
		headerKeywords := []string{"name", "staff", "type", "position", "branch", "nickname"}
		isHeader := false
		if len(firstRow) > 0 {
			firstCell := strings.ToLower(strings.TrimSpace(firstRow[0]))
			for _, keyword := range headerKeywords {
				if strings.Contains(firstCell, keyword) {
					isHeader = true
					break
				}
			}
		}
		if isHeader {
			startRow = 1
		}
	}

	if len(rows) <= startRow {
		return nil, fmt.Errorf("Excel file must have at least one data row")
	}

	// Process data rows starting from startRow
	for i := startRow; i < len(rows); i++ {
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

		// Column C: Position Name (lookup by name)
		if len(row) > 2 && strings.TrimSpace(row[2]) != "" {
			positionName := strings.TrimSpace(row[2])
			position, err := e.findPositionByName(positionName)
			if err != nil {
				errors = append(errors, fmt.Sprintf("Row %d: position '%s' not found: %v", i+1, positionName, err))
				continue
			}
			staff.PositionID = position.ID
		} else {
			errors = append(errors, fmt.Sprintf("Row %d: position name is required", i+1))
			continue
		}

		// Column D: Branch Code (optional, mainly for branch staff)
		if len(row) > 3 && strings.TrimSpace(row[3]) != "" {
			branchCode := strings.TrimSpace(row[3])
			branch, err := e.findBranchByCode(branchCode)
			if err != nil {
				errors = append(errors, fmt.Sprintf("Row %d: branch code '%s' not found: %v", i+1, branchCode, err))
				continue
			}
			staff.BranchID = &branch.ID
		}

		// Column E: Nickname (optional)
		if len(row) > 4 && strings.TrimSpace(row[4]) != "" {
			staff.Nickname = strings.TrimSpace(row[4])
		}

		// Skill Level defaults to 5 (set during staff creation if not specified)
		staff.SkillLevel = 5

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

// findPositionByName looks up a position by name (case-insensitive)
func (e *ExcelImporter) findPositionByName(name string) (*models.Position, error) {
	positions, err := e.positionRepo.List()
	if err != nil {
		return nil, fmt.Errorf("failed to load positions: %w", err)
	}
	
	nameLower := strings.ToLower(strings.TrimSpace(name))
	for _, pos := range positions {
		if strings.ToLower(pos.Name) == nameLower {
			return pos, nil
		}
	}
	return nil, fmt.Errorf("position '%s' not found", name)
}

// findBranchByCode looks up a branch by code (case-insensitive)
func (e *ExcelImporter) findBranchByCode(code string) (*models.Branch, error) {
	branches, err := e.branchRepo.List()
	if err != nil {
		return nil, fmt.Errorf("failed to load branches: %w", err)
	}
	
	codeLower := strings.ToLower(strings.TrimSpace(code))
	for _, branch := range branches {
		if strings.ToLower(branch.Code) == codeLower {
			return branch, nil
		}
	}
	return nil, fmt.Errorf("branch code '%s' not found", code)
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

	// Validate skill level range
	if staff.SkillLevel < 0 || staff.SkillLevel > 10 {
		return fmt.Errorf("skill level must be between 0 and 10, got %d", staff.SkillLevel)
	}

	return nil
}


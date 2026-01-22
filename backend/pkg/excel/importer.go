package excel

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
	"vsq-oper-manpower/backend/internal/domain/interfaces"
	"vsq-oper-manpower/backend/internal/domain/models"
)

// ExcelImporter handles importing staff data from Excel files
type ExcelImporter struct {
	positionRepo   interfaces.PositionRepository
	branchRepo     interfaces.BranchRepository
	doctorRepo     interfaces.DoctorRepository
	positionQuotaRepo interfaces.PositionQuotaRepository
}

func NewExcelImporter(
	positionRepo interfaces.PositionRepository,
	branchRepo interfaces.BranchRepository,
	doctorRepo interfaces.DoctorRepository,
	positionQuotaRepo interfaces.PositionQuotaRepository,
) *ExcelImporter {
	return &ExcelImporter{
		positionRepo:      positionRepo,
		branchRepo:        branchRepo,
		doctorRepo:        doctorRepo,
		positionQuotaRepo: positionQuotaRepo,
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

// ImportDoctors parses Excel file and returns doctor records
// Expected format:
// - Row 1: Header row (optional, will be skipped)
// - Row 2+: Data rows
// - Column A: Name (required)
// - Column B: Code (optional) - doctor code/nickname
// - Column C: Preferences (optional) - noted remark/preferences
func (e *ExcelImporter) ImportDoctors(fileData []byte) ([]*models.Doctor, error) {
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

	var doctorList []*models.Doctor
	var errors []string

	// Detect if first row is a header row (contains common header keywords)
	startRow := 0
	if len(rows) > 0 {
		firstRow := rows[0]
		headerKeywords := []string{"name", "doctor", "code", "preferences", "preference"}
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
		if len(row) < 1 {
			errors = append(errors, fmt.Sprintf("Row %d: insufficient columns (need at least 1)", i+1))
			continue
		}

		doctor := &models.Doctor{
			ID: uuid.New(),
		}

		// Column A: Name (required)
		doctor.Name = strings.TrimSpace(row[0])
		if doctor.Name == "" {
			errors = append(errors, fmt.Sprintf("Row %d: name is required", i+1))
			continue
		}

		// Column B: Code (optional)
		if len(row) > 1 && strings.TrimSpace(row[1]) != "" {
			doctor.Code = strings.TrimSpace(row[1])
			// Check if code already exists
			existing, err := e.doctorRepo.GetByCode(doctor.Code)
			if err == nil && existing != nil {
				errors = append(errors, fmt.Sprintf("Row %d: doctor code '%s' already exists", i+1, doctor.Code))
				continue
			}
		}

		// Column C: Preferences (optional)
		if len(row) > 2 && strings.TrimSpace(row[2]) != "" {
			doctor.Preferences = strings.TrimSpace(row[2])
		}

		// Validate doctor data
		if err := e.ValidateDoctorData(doctor); err != nil {
			errors = append(errors, fmt.Sprintf("Row %d: validation error: %v", i+1, err))
			continue
		}

		doctorList = append(doctorList, doctor)
	}

	if len(errors) > 0 && len(doctorList) == 0 {
		return nil, fmt.Errorf("failed to import any doctors: %s", strings.Join(errors, "; "))
	}

	if len(errors) > 0 {
		// Return partial success with errors
		return doctorList, fmt.Errorf("import completed with errors: %s", strings.Join(errors, "; "))
	}

	return doctorList, nil
}

// ValidateDoctorData validates imported doctor data
func (e *ExcelImporter) ValidateDoctorData(doctor *models.Doctor) error {
	if doctor.Name == "" {
		return fmt.Errorf("doctor name is required")
	}

	// Code uniqueness is checked during import, not here
	// Preferences can be empty

	return nil
}

// ImportDefaultSchedulesResult contains both schedules to create/update and off days to delete
type ImportDefaultSchedulesResult struct {
	Schedules []*models.DoctorDefaultSchedule
	OffDays   []struct {
		DoctorID  uuid.UUID
		DayOfWeek int
	}
}

// ImportDefaultSchedules parses Excel file and returns doctor default schedule records
// Expected format:
// - Row 1: Header row (optional, will be skipped)
// - Row 2+: Data rows
// - Column A: Doctor Code (required)
// - Column B: Day of Week (required) - 1=Monday, 2=Tuesday, ..., 7=Sunday
// - Column C: Branch Code or Branch Name (optional - empty or "OFF"/"Off Day" means off day)
// Note: If a doctor has duplicate branches on the same workday, the last imported entry will be used.
// Empty branch or "OFF" means the doctor is off on that day (existing schedule will be deleted).
func (e *ExcelImporter) ImportDefaultSchedules(fileData []byte) (*ImportDefaultSchedulesResult, error) {
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

	var scheduleList []*models.DoctorDefaultSchedule
	var offDaysList []struct {
		DoctorID  uuid.UUID
		DayOfWeek int
	}
	var errors []string

	// Detect if first row is a header row (contains common header keywords)
	startRow := 0
	if len(rows) > 0 {
		firstRow := rows[0]
		headerKeywords := []string{"doctor", "code", "day", "week", "branch", "schedule"}
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

	// Use maps to track the last schedule/off day for each doctor+day combination
	// This ensures duplicates are handled by keeping the last imported entry
	scheduleMap := make(map[string]*models.DoctorDefaultSchedule)
	offDaysMap := make(map[string]struct {
		DoctorID  uuid.UUID
		DayOfWeek int
	})

	// Process data rows starting from startRow
	for i := startRow; i < len(rows); i++ {
		row := rows[i]
		if len(row) < 3 {
			errors = append(errors, fmt.Sprintf("Row %d: insufficient columns (need at least 3)", i+1))
			continue
		}

		// Column A: Doctor Code (required)
		doctorCode := strings.TrimSpace(row[0])
		if doctorCode == "" {
			errors = append(errors, fmt.Sprintf("Row %d: doctor code is required", i+1))
			continue
		}

		// Find doctor by code
		doctor, err := e.doctorRepo.GetByCode(doctorCode)
		if err != nil || doctor == nil {
			errors = append(errors, fmt.Sprintf("Row %d: doctor code '%s' not found", i+1, doctorCode))
			continue
		}

		// Column B: Day of Week (required) - 1=Monday, 2=Tuesday, ..., 7=Sunday
		dayOfWeekStr := strings.TrimSpace(row[1])
		if dayOfWeekStr == "" {
			errors = append(errors, fmt.Sprintf("Row %d: day of week is required", i+1))
			continue
		}

		// Parse day of week (user format: 1-7, where 1=Monday, 7=Sunday)
		var userDayOfWeek int
		if _, err := fmt.Sscanf(dayOfWeekStr, "%d", &userDayOfWeek); err != nil {
			errors = append(errors, fmt.Sprintf("Row %d: invalid day of week '%s' (must be 1-7)", i+1, dayOfWeekStr))
			continue
		}

		// Validate day of week range (1-7)
		if userDayOfWeek < 1 || userDayOfWeek > 7 {
			errors = append(errors, fmt.Sprintf("Row %d: day of week must be between 1 and 7 (1=Monday, 7=Sunday), got %d", i+1, userDayOfWeek))
			continue
		}

		// Convert from user format (1-7, Monday-Sunday) to system format (0-6, Sunday-Saturday)
		// User: 1=Monday, 2=Tuesday, 3=Wednesday, 4=Thursday, 5=Friday, 6=Saturday, 7=Sunday
		// System: 0=Sunday, 1=Monday, 2=Tuesday, 3=Wednesday, 4=Thursday, 5=Friday, 6=Saturday
		var systemDayOfWeek int
		if userDayOfWeek == 7 {
			systemDayOfWeek = 0 // Sunday
		} else {
			systemDayOfWeek = userDayOfWeek // Monday-Saturday map directly
		}

		// Column C: Branch Code or Branch Name (optional - empty means off day)
		branchIdentifier := strings.TrimSpace(row[2])
		
		// Check if it's an off day (empty or special values like "OFF", "Off Day", "OFF_DAY")
		branchIdentifierLower := strings.ToLower(branchIdentifier)
		isOffDay := branchIdentifier == "" || 
			branchIdentifierLower == "off" || 
			branchIdentifierLower == "off day" || 
			branchIdentifierLower == "off_day" ||
			branchIdentifierLower == "offday"

		if isOffDay {
			// For off days, track that this day should be off (existing schedule will be deleted)
			// Use a composite key to track duplicates
			key := fmt.Sprintf("%s-%d", doctor.ID.String(), systemDayOfWeek)
			offDaysMap[key] = struct {
				DoctorID  uuid.UUID
				DayOfWeek int
			}{
				DoctorID:  doctor.ID,
				DayOfWeek: systemDayOfWeek,
			}
			// Remove from schedule map if it exists (off day takes precedence)
			delete(scheduleMap, key)
			continue
		}

		// Find branch by code or name
		branch, err := e.findBranchByCodeOrName(branchIdentifier)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Row %d: branch '%s' not found: %v", i+1, branchIdentifier, err))
			continue
		}

		// Create schedule record
		schedule := &models.DoctorDefaultSchedule{
			ID:        uuid.New(),
			DoctorID:  doctor.ID,
			DayOfWeek: systemDayOfWeek,
			BranchID:  branch.ID,
		}

		// Use a composite key to track duplicates: doctorID + dayOfWeek
		// This ensures the last imported entry for the same doctor+day wins
		key := fmt.Sprintf("%s-%d", doctor.ID.String(), systemDayOfWeek)
		scheduleMap[key] = schedule
	}

	// Convert maps to slices (only the last entry for each doctor+day combination)
	for _, schedule := range scheduleMap {
		scheduleList = append(scheduleList, schedule)
	}
	for _, offDay := range offDaysMap {
		offDaysList = append(offDaysList, offDay)
	}

	result := &ImportDefaultSchedulesResult{
		Schedules: scheduleList,
		OffDays:   offDaysList,
	}

	if len(errors) > 0 && len(scheduleList) == 0 && len(offDaysList) == 0 {
		return nil, fmt.Errorf("failed to import any schedules: %s", strings.Join(errors, "; "))
	}

	if len(errors) > 0 {
		// Return partial success with errors
		return result, fmt.Errorf("import completed with errors: %s", strings.Join(errors, "; "))
	}

	return result, nil
}

// findBranchByCodeOrName looks up a branch by code or name (case-insensitive)
// Tries code first, then name if code doesn't match
func (e *ExcelImporter) findBranchByCodeOrName(identifier string) (*models.Branch, error) {
	branches, err := e.branchRepo.List()
	if err != nil {
		return nil, fmt.Errorf("failed to load branches: %w", err)
	}

	identifierLower := strings.ToLower(strings.TrimSpace(identifier))

	// First try to find by code
	for _, branch := range branches {
		if strings.ToLower(branch.Code) == identifierLower {
			return branch, nil
		}
	}

	// If not found by code, try to find by name
	for _, branch := range branches {
		if strings.ToLower(branch.Name) == identifierLower {
			return branch, nil
		}
	}

	return nil, fmt.Errorf("branch code or name '%s' not found", identifier)
}

// ImportPositionQuotasResult contains the import results
type ImportPositionQuotasResult struct {
	Created int      `json:"created"`
	Updated int      `json:"updated"`
	Errors  []string `json:"errors"`
}

// ImportPositionQuotas parses Excel file and imports/updates position quotas
// Expected format:
// - Row 1: Header row (optional, will be skipped)
// - Row 2+: Data rows
// - Column A: Branch Code (required) - e.g., "TMA", "CPN"
// - Column B: Position Code (required) - e.g., "BM", "ABM", "DA"
// - Column C: Preferred No. (required) - designated_quota
// - Column D: Minimum No. (required) - minimum_required
func (e *ExcelImporter) ImportPositionQuotas(fileData []byte, createdBy uuid.UUID) (*ImportPositionQuotasResult, error) {
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

	result := &ImportPositionQuotasResult{
		Errors: []string{},
	}

	// Detect if first row is a header row (contains common header keywords)
	startRow := 0
	if len(rows) > 0 {
		firstRow := rows[0]
		headerKeywords := []string{"branch", "position", "code", "preferred", "minimum", "quota"}
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

	// Load all positions and branches for lookup
	positions, err := e.positionRepo.List()
	if err != nil {
		return nil, fmt.Errorf("failed to load positions: %w", err)
	}
	positionMap := make(map[string]*models.Position) // Map position_code -> Position
	for _, pos := range positions {
		if pos.PositionCode != nil && *pos.PositionCode != "" {
			positionMap[strings.ToLower(strings.TrimSpace(*pos.PositionCode))] = pos
		}
	}

	branches, err := e.branchRepo.List()
	if err != nil {
		return nil, fmt.Errorf("failed to load branches: %w", err)
	}
	branchMap := make(map[string]*models.Branch) // Map branch code -> Branch
	for _, branch := range branches {
		branchMap[strings.ToLower(strings.TrimSpace(branch.Code))] = branch
	}

	// Process data rows starting from startRow
	for i := startRow; i < len(rows); i++ {
		row := rows[i]
		if len(row) < 4 {
			result.Errors = append(result.Errors, fmt.Sprintf("Row %d: insufficient columns (need at least 4: Branch Code, Position Code, Preferred No., Minimum No.)", i+1))
			continue
		}

		// Column A: Branch Code (required)
		branchCode := strings.TrimSpace(row[0])
		if branchCode == "" {
			result.Errors = append(result.Errors, fmt.Sprintf("Row %d: branch code is required", i+1))
			continue
		}
		branch, found := branchMap[strings.ToLower(branchCode)]
		if !found {
			result.Errors = append(result.Errors, fmt.Sprintf("Row %d: branch code '%s' not found", i+1, branchCode))
			continue
		}

		// Column B: Position Code (required)
		positionCode := strings.TrimSpace(row[1])
		if positionCode == "" {
			result.Errors = append(result.Errors, fmt.Sprintf("Row %d: position code is required", i+1))
			continue
		}
		position, found := positionMap[strings.ToLower(positionCode)]
		if !found {
			result.Errors = append(result.Errors, fmt.Sprintf("Row %d: position code '%s' not found", i+1, positionCode))
			continue
		}

		// Column C: Preferred No. (required) - designated_quota
		preferredStr := strings.TrimSpace(row[2])
		if preferredStr == "" {
			result.Errors = append(result.Errors, fmt.Sprintf("Row %d: preferred number is required", i+1))
			continue
		}
		var preferredNo int
		if _, err := fmt.Sscanf(preferredStr, "%d", &preferredNo); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Row %d: invalid preferred number '%s' (must be an integer)", i+1, preferredStr))
			continue
		}
		if preferredNo < 0 {
			result.Errors = append(result.Errors, fmt.Sprintf("Row %d: preferred number must be >= 0, got %d", i+1, preferredNo))
			continue
		}

		// Column D: Minimum No. (required) - minimum_required
		minimumStr := strings.TrimSpace(row[3])
		if minimumStr == "" {
			result.Errors = append(result.Errors, fmt.Sprintf("Row %d: minimum number is required", i+1))
			continue
		}
		var minimumNo int
		if _, err := fmt.Sscanf(minimumStr, "%d", &minimumNo); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Row %d: invalid minimum number '%s' (must be an integer)", i+1, minimumStr))
			continue
		}
		if minimumNo < 0 {
			result.Errors = append(result.Errors, fmt.Sprintf("Row %d: minimum number must be >= 0, got %d", i+1, minimumNo))
			continue
		}

		// Validate that minimum <= preferred
		if minimumNo > preferredNo {
			result.Errors = append(result.Errors, fmt.Sprintf("Row %d: minimum number (%d) cannot be greater than preferred number (%d)", i+1, minimumNo, preferredNo))
			continue
		}

		// Check if quota already exists
		existingQuota, err := e.positionQuotaRepo.GetByBranchAndPosition(branch.ID, position.ID)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Row %d: failed to check existing quota: %v", i+1, err))
			continue
		}

		if existingQuota != nil {
			// Update existing quota
			existingQuota.DesignatedQuota = preferredNo
			existingQuota.MinimumRequired = minimumNo
			existingQuota.IsActive = true
			if err := e.positionQuotaRepo.Update(existingQuota); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Row %d: failed to update quota: %v", i+1, err))
				continue
			}
			result.Updated++
		} else {
			// Create new quota
			quota := &models.PositionQuota{
				ID:              uuid.New(),
				BranchID:        branch.ID,
				PositionID:      position.ID,
				DesignatedQuota:  preferredNo,
				MinimumRequired:  minimumNo,
				IsActive:         true,
				CreatedBy:        createdBy,
			}
			if err := e.positionQuotaRepo.Create(quota); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Row %d: failed to create quota: %v", i+1, err))
				continue
			}
			result.Created++
		}
	}

	return result, nil
}

// ImportBranchRevenueResult contains the import results for branch revenue
type ImportBranchRevenueResult struct {
	Created int      `json:"created"`
	Updated int      `json:"updated"`
	Errors  []string `json:"errors"`
}

// ImportBranchRevenue parses Excel file and imports/updates branch expected revenue
// Expected format:
// - Row 1: Header row (optional, will be skipped)
// - Row 2+: Data rows
// - Column A: Branch Code (required) - e.g., "TMA", "CPN"
// - Column B: Date (required) - Format: YYYY-MM-DD or Excel date format
// - Column C: Skin Revenue (required) - Decimal number (THB)
// - Column D: LS HM Revenue (required) - Decimal number (THB)
// - Column E: Vitamin Cases (required) - Integer (count)
// - Column F: Slim Pen Cases (required) - Integer (count)
// The import overrides all existing revenue for each day of each branch.
// Note: For backward compatibility, if only Column C is provided, it will be treated as Skin Revenue.
func (e *ExcelImporter) ImportBranchRevenue(fileData []byte) ([]*models.RevenueData, []string, error) {
	// Open Excel file from byte data
	f, err := excelize.OpenReader(bytes.NewReader(fileData))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer f.Close()

	// Get the first sheet
	sheetName := f.GetSheetName(0)
	if sheetName == "" {
		return nil, nil, fmt.Errorf("Excel file has no sheets")
	}

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read rows: %w", err)
	}

	if len(rows) < 1 {
		return nil, nil, fmt.Errorf("Excel file is empty")
	}

	var errors []string

		// Detect if first row is a header row (contains common header keywords)
		startRow := 0
		if len(rows) > 0 {
			firstRow := rows[0]
			headerKeywords := []string{"branch", "code", "date", "revenue", "expected", "skin", "ls", "hm", "vitamin", "slim", "pen", "cases"}
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
		return nil, nil, fmt.Errorf("Excel file must have at least one data row")
	}

	// Load all branches for lookup
	branches, err := e.branchRepo.List()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load branches: %w", err)
	}
	branchMap := make(map[string]*models.Branch) // Map branch code -> Branch
	for _, branch := range branches {
		branchMap[strings.ToLower(strings.TrimSpace(branch.Code))] = branch
	}

	var revenueList []*models.RevenueData

	// Process data rows starting from startRow
	for i := startRow; i < len(rows); i++ {
		row := rows[i]
		if len(row) < 3 {
			errors = append(errors, fmt.Sprintf("Row %d: insufficient columns (need at least 3: Branch Code, Date, Skin Revenue)", i+1))
			continue
		}

		// Column A: Branch Code (required)
		branchCode := strings.TrimSpace(row[0])
		if branchCode == "" {
			errors = append(errors, fmt.Sprintf("Row %d: branch code is required", i+1))
			continue
		}
		branch, found := branchMap[strings.ToLower(branchCode)]
		if !found {
			errors = append(errors, fmt.Sprintf("Row %d: branch code '%s' not found", i+1, branchCode))
			continue
		}

		// Column B: Date (required)
		dateStr := strings.TrimSpace(row[1])
		if dateStr == "" {
			errors = append(errors, fmt.Sprintf("Row %d: date is required", i+1))
			continue
		}

		// Try to parse date - support both YYYY-MM-DD and Excel date format
		var date time.Time
		var parseErr error
		
		// First try YYYY-MM-DD format
		date, parseErr = time.Parse("2006-01-02", dateStr)
		if parseErr != nil {
			// Try Excel date format (try parsing as float first, then as various date formats)
			if excelDate, err := strconv.ParseFloat(dateStr, 64); err == nil {
				// Excel date serial number (days since 1900-01-01)
				excelEpoch := time.Date(1899, 12, 30, 0, 0, 0, 0, time.UTC)
				days := int(excelDate)
				date = excelEpoch.AddDate(0, 0, days)
			} else {
				// Try other common date formats
				formats := []string{"2006/01/02", "02/01/2006", "01/02/2006", "2006-01-02 15:04:05"}
				parsed := false
				for _, format := range formats {
					if d, err := time.Parse(format, dateStr); err == nil {
						date = d
						parsed = true
						break
					}
				}
				if !parsed {
					errors = append(errors, fmt.Sprintf("Row %d: invalid date format '%s' (expected YYYY-MM-DD)", i+1, dateStr))
					continue
				}
			}
		}

		// Column C: Skin Revenue (required)
		skinRevenueStr := strings.TrimSpace(row[2])
		if skinRevenueStr == "" {
			errors = append(errors, fmt.Sprintf("Row %d: skin revenue is required", i+1))
			continue
		}
		skinRevenue, err := strconv.ParseFloat(skinRevenueStr, 64)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Row %d: invalid skin revenue '%s': %v", i+1, skinRevenueStr, err))
			continue
		}
		if skinRevenue < 0 {
			errors = append(errors, fmt.Sprintf("Row %d: skin revenue cannot be negative", i+1))
			continue
		}

		// Column D: LS HM Revenue (optional, defaults to 0)
		lsHMRevenue := 0.0
		if len(row) > 3 && strings.TrimSpace(row[3]) != "" {
			lsHMRevenueStr := strings.TrimSpace(row[3])
			lsHMRevenue, err = strconv.ParseFloat(lsHMRevenueStr, 64)
			if err != nil {
				errors = append(errors, fmt.Sprintf("Row %d: invalid LS HM revenue '%s': %v", i+1, lsHMRevenueStr, err))
				continue
			}
			if lsHMRevenue < 0 {
				errors = append(errors, fmt.Sprintf("Row %d: LS HM revenue cannot be negative", i+1))
				continue
			}
		}

		// Column E: Vitamin Cases (optional, defaults to 0)
		vitaminCases := 0
		if len(row) > 4 && strings.TrimSpace(row[4]) != "" {
			vitaminCasesStr := strings.TrimSpace(row[4])
			vitaminCases, err = strconv.Atoi(vitaminCasesStr)
			if err != nil {
				errors = append(errors, fmt.Sprintf("Row %d: invalid vitamin cases '%s': %v", i+1, vitaminCasesStr, err))
				continue
			}
			if vitaminCases < 0 {
				errors = append(errors, fmt.Sprintf("Row %d: vitamin cases cannot be negative", i+1))
				continue
			}
		}

		// Column F: Slim Pen Cases (optional, defaults to 0)
		slimPenCases := 0
		if len(row) > 5 && strings.TrimSpace(row[5]) != "" {
			slimPenCasesStr := strings.TrimSpace(row[5])
			slimPenCases, err = strconv.Atoi(slimPenCasesStr)
			if err != nil {
				errors = append(errors, fmt.Sprintf("Row %d: invalid slim pen cases '%s': %v", i+1, slimPenCasesStr, err))
				continue
			}
			if slimPenCases < 0 {
				errors = append(errors, fmt.Sprintf("Row %d: slim pen cases cannot be negative", i+1))
				continue
			}
		}

		// Create revenue data
		revenue := &models.RevenueData{
			ID:              uuid.New(),
			BranchID:        branch.ID,
			Date:            date,
			ExpectedRevenue: skinRevenue, // Keep for backward compatibility
			SkinRevenue:     skinRevenue,
			LSHMRevenue:     lsHMRevenue,
			VitaminCases:    vitaminCases,
			SlimPenCases:    slimPenCases,
			ActualRevenue:   nil,
			RevenueSource:   "excel",
		}
		revenueList = append(revenueList, revenue)
	}

	return revenueList, errors, nil
}


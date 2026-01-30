# Test Data Generation Guide

## Overview

The Test Data Generation feature allows administrators to automatically generate staff working days and leave days for all branches with configurable randomization rules. This is designed for development and testing purposes to quickly populate schedules without manual entry.

## Accessing the UI

**URL:** `/test-data` (Admin only)

**Navigation:** 
- Go to **System and Administration** → **Test Data Generation** in the sidebar
- Or navigate directly to: `http://localhost:4000/test-data`

**Access Requirements:** Only users with the `admin` role can access this page.

## How It Works

### 1. **Generation Process Flow**

```
User fills form → Submits → Backend generates schedules → Returns statistics
```

The system follows this process:

1. **User Configuration**: Admin selects a month and configures rules via the UI
2. **Data Collection**: Backend retrieves all branches and branch staff
3. **Schedule Generation**: For each staff member, generates schedules day-by-day applying rules
4. **Enforcement Pass**: If enabled, enforces minimum staff per group requirements
5. **Database Storage**: Saves all schedules to the database
6. **Statistics**: Returns summary of generated schedules

### 2. **Randomization Algorithm**

The generator uses a **rule-based randomization** approach with the following priority order:

#### **Rule Priority (Applied in Order)**

1. **Holidays** (if `exclude_holidays` is enabled)
   - Public holidays are automatically set to "off"
   - Thai public holidays for 2026 are pre-configured

2. **Weekends**
   - Uses `weekend_working_ratio` (0.0-1.0) to determine if staff works
   - Example: 0.3 = 30% chance of working on weekends

3. **Leave Probability**
   - Each day has a `leave_probability` chance of being leave
   - Respects `consecutive_leave_max` limit
   - Example: 0.15 = 15% chance of leave on any given day

4. **Working Days Per Week Constraint**
   - Each week gets a random target between `min_working_days_per_week` and `max_working_days_per_week`
   - System ensures this target is met before the week ends
   - Example: Min=4, Max=6 → Each week will have 4-6 working days

5. **Off Days Per Month Constraint**
   - Tracks off days per month per staff member
   - Enforces `min_off_days_per_month` and `max_off_days_per_month`
   - Example: Min=4, Max=8 → Each staff gets 4-8 off days per month

6. **Minimum Staff Per Group Enforcement** (if enabled)
   - After initial generation, checks branch constraints
   - For each day, ensures minimum staff per group requirements are met
   - If below minimum, randomly selects non-working staff to make working

### 3. **Randomization Details**

#### **Consistency**
- Each staff member gets a **deterministic random seed** based on their staff ID
- This means: same staff + same rules = same schedule (reproducible)
- Different staff members get different schedules (variety)

#### **Week-by-Week Randomization**
- Each week gets a random target working days (within min/max range)
- The system ensures this target is met by prioritizing working days when needed
- This creates realistic weekly patterns

#### **Month-by-Month Tracking**
- Off days are tracked per month
- System ensures minimum off days are met (may force off days near month end)
- System prevents exceeding maximum off days (forces working when at max)

### 4. **Example Generation Scenario**

**Input:**
- Month: January 2026
- Min Working Days/Week: 4
- Max Working Days/Week: 6
- Leave Probability: 0.15 (15%)
- Weekend Working Ratio: 0.3 (30%)
- Min Off Days/Month: 4
- Max Off Days/Month: 8
- Enforce Min Staff Per Group: Yes

**Process for Staff Member A:**

**Week 1 (Jan 1-7):**
- Target: 5 working days (randomly chosen between 4-6)
- Jan 1 (Holiday): Off (New Year)
- Jan 2 (Friday): Working (need to meet target)
- Jan 3 (Saturday): Off (70% chance)
- Jan 4 (Sunday): Off (70% chance)
- Jan 5 (Monday): Leave (15% chance hit)
- Jan 6 (Tuesday): Working (need to meet target)
- Jan 7 (Wednesday): Working (need to meet target)
- Result: 3 working, 1 leave, 3 off

**Week 2 (Jan 8-14):**
- Target: 6 working days
- Jan 8 (Thursday): Working
- Jan 9 (Friday): Working
- Jan 10 (Saturday): Working (30% chance hit)
- ... continues with rules applied

**After All Weeks:**
- Total off days: 6 (within 4-8 range)
- Each week has 4-6 working days
- Leave days distributed randomly (15% probability)

**Enforcement Pass:**
- Checks each day for minimum staff per group
- If "Front Staff" group needs minimum 2, but only 1 is working:
  - Randomly selects one off/leave staff from Front Staff group
  - Changes their status to "working" for that day

## UI Components

### **Form Fields**

1. **Select Month** (Month picker)
   - Automatically sets start date to 1st and end date to last day of month
   - Example: Selecting "2026-01" sets dates to Jan 1-31, 2026

2. **Working Days Per Week**
   - Min: Minimum working days per week (0-7)
   - Max: Maximum working days per week (0-7)
   - Each week gets a random target within this range

3. **Leave Settings**
   - **Leave Probability**: 0.0-1.0 (percentage chance of leave)
   - **Max Consecutive Leave**: Maximum consecutive leave days allowed

4. **Weekend Working Ratio**
   - 0.0-1.0 (percentage of weekends that are working days)
   - Example: 0.3 = 30% of weekends will be working

5. **Off Days Per Month**
   - Min: Minimum off days per month per staff
   - Max: Maximum off days per month per staff

6. **Options**
   - **Exclude Public Holidays**: Sets holidays as off days
   - **Enforce Minimum Staff Per Group**: Ensures branch constraints are met
   - **Overwrite Existing Schedules**: Replaces existing schedules if checked

### **Results Display**

After generation, shows:
- **Total Staff**: Number of staff members processed
- **Total Schedules**: Number of schedule entries created
- **Working Days**: Count of working day schedules
- **Leave Days**: Count of leave day schedules
- **Off Days**: Count of off day schedules
- **Errors**: List of any errors encountered (if any)

## Technical Implementation

### **Backend Architecture**

```
Request → Handler → Use Case → Repository → Database
```

**Files:**
- **Handler**: `backend/internal/handlers/test_data_handler.go`
  - Validates request
  - Calls generator use case
  - Returns results

- **Use Case**: `backend/internal/usecases/test_data/schedule_generator.go`
  - Core generation logic
  - Rule application
  - Minimum staff enforcement

- **API Endpoint**: `POST /api/admin/test-data/generate-schedules`
  - Admin-only endpoint
  - Requires authentication

### **Frontend Architecture**

**Files:**
- **Page**: `frontend/src/app/(admin)/test-data/page.tsx`
  - Form UI
  - State management
  - Results display

- **API Client**: `frontend/src/lib/api/test-data.ts`
  - TypeScript interfaces
  - API call function

## Usage Examples

### **Example 1: Generate Schedules for January 2026**

1. Navigate to `/test-data`
2. Select month: "2026-01"
3. Set rules:
   - Min Working Days: 4
   - Max Working Days: 6
   - Leave Probability: 0.15
   - Weekend Ratio: 0.3
   - Min Off Days/Month: 4
   - Max Off Days/Month: 8
4. Check "Enforce Minimum Staff Per Group"
5. Click "Generate Schedules"
6. View results showing statistics

### **Example 2: Quick Test with Minimal Rules**

1. Select current month
2. Set basic rules:
   - Min Working Days: 5
   - Max Working Days: 5
   - Leave Probability: 0.1
   - Weekend Ratio: 0.2
3. Uncheck "Enforce Minimum Staff Per Group" (faster)
4. Generate

## Best Practices

1. **Start Small**: Test with a single month first
2. **Check Constraints**: Ensure branch constraints are configured before enabling "Enforce Minimum Staff Per Group"
3. **Review Results**: Check the statistics to verify the generation meets expectations
4. **Use Overwrite Carefully**: Only overwrite when you want to replace existing data
5. **Adjust Rules**: Fine-tune probabilities and constraints based on your needs

## Limitations

1. **Deterministic Randomness**: Same staff + same rules = same schedule (by design for reproducibility)
2. **Holiday List**: Currently hardcoded for 2026 (can be expanded)
3. **Performance**: Large date ranges may take time (consider generating month-by-month)
4. **No Undo**: Generated schedules are permanent (unless overwritten)

## Troubleshooting

### **Issue: "No schedules generated"**
- Check if branches have staff assigned
- Verify date range is valid
- Check for errors in the results

### **Issue: "Minimum staff not met"**
- Ensure branch constraints are configured
- Check if enough staff exist in required groups
- Verify staff group assignments are correct

### **Issue: "Too many/few off days"**
- Adjust `min_off_days_per_month` and `max_off_days_per_month`
- Check if `exclude_holidays` is adding extra off days

## Future Enhancements

Potential improvements:
- Year selector for holiday lists
- Preview before generation
- Undo/rollback functionality
- Export/import rule presets
- Progress indicator for large generations
- Branch-specific rule configuration UI

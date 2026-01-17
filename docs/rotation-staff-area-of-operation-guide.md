---
title: How to Assign Areas of Operation to Rotation Staff
description: Step-by-step guide for assigning areas of operation to rotation staff via the UI
version: 1.0.0
lastUpdated: 2026-01-17 15:30:00
---

# How to Assign Areas of Operation to Rotation Staff

## Overview

This guide explains how to assign areas of operation to rotation staff members using the Staff Management interface. Areas of operation help categorize and organize rotation staff based on their operational regions.

## Prerequisites

- You must have one of the following roles:
  - **Admin** - Full access to manage all staff
  - **Area Manager** - Can manage rotation staff
  - **District Manager** - Can view rotation staff (read-only)
- Areas of Operation must be created first (see "Managing Areas of Operation" section below)

## Step-by-Step Instructions

### Method 1: Assign Area When Creating New Rotation Staff

1. **Navigate to Staff Management**
   - Log in to the system
   - Go to **Staff Management** from the main menu
   - The page URL is: `/staff-management`

2. **Click "Add Staff" Button**
   - Click the **"Add Staff"** button in the top-right corner of the staff table

3. **Fill in Basic Information**
   - **Nickname** (optional): Enter a nickname for the staff member
   - **Full Name** (required): Enter the full name
   - **Skill Level** (0-10): Set the skill level (default: 5)

4. **Select Staff Type**
   - In the **Staff Type** dropdown, select **"Rotation Staff"**
   - Note: Only Admin and Area Manager roles can create rotation staff
   - Branch Managers can only create branch staff

5. **Select Position**
   - Choose the appropriate **Position** from the dropdown
   - This determines the role/position of the rotation staff member

6. **Assign Area of Operation** ⭐
   - Scroll down to the **"Area of Operation"** field
   - Click the dropdown to see available areas
   - Select the desired area of operation (e.g., "North Region (NR)", "Central Region (CR)")
   - **Note:** This field is optional - you can leave it blank if no area assignment is needed
   - The dropdown shows format: `Area Name (Code)`

7. **Optional: Set Coverage Area (Legacy)**
   - You can also set a **Coverage Area** (legacy field)
   - This is kept for backward compatibility
   - Prefer using Area of Operation instead

8. **Set Effective Branches**
   - Check the boxes for branches this rotation staff can support
   - Set **Level 1** (Priority) or **Level 2** (Reserved) for each branch
   - Level 1 = Priority branches (preferred assignments)
   - Level 2 = Reserved branches (used when Level 1 staff are insufficient)

9. **Save the Staff Member**
   - Click **"Create"** button
   - The rotation staff member will be created with the assigned area of operation

### Method 2: Update Area for Existing Rotation Staff

1. **Navigate to Staff Management**
   - Go to **Staff Management** page

2. **Filter Rotation Staff (Optional)**
   - Use the **"Filter by Type"** dropdown
   - Select **"Rotation Staff"** to see only rotation staff members

3. **Find the Staff Member**
   - Locate the rotation staff member in the table
   - You can see their current Area of Operation in the **"Area of Operation"** column
   - It displays as a badge: `Area Name (Code)` or `-` if not assigned

4. **Click "Edit"**
   - Click the **"Edit"** button next to the staff member
   - Note: Only Admin and Area Manager can edit rotation staff
   - Branch Managers can only edit branch staff

5. **Update Area of Operation**
   - In the edit form, find the **"Area of Operation"** dropdown
   - Select a different area or clear the selection (select "Select Area of Operation (Optional)")
   - The current area will be pre-selected if one is assigned

6. **Save Changes**
   - Click **"Update"** button
   - The area of operation will be updated for the staff member

## Viewing Area of Operation

### In Staff Management Table

The **Staff Management** table displays:
- **Area of Operation** column showing the assigned area
- Format: Badge with `Area Name (Code)` (e.g., "North Region (NR)")
- Shows `-` if no area is assigned

### Filtering by Area of Operation

In the **Rotation Staff List** component (used in rotation assignment views):
- You can filter rotation staff by Area of Operation
- Use the **"Filter by Area of Operation"** dropdown
- Select an area to see only staff assigned to that area

## Managing Areas of Operation

Before assigning areas to rotation staff, you need to create areas of operation:

1. **Navigate to Areas of Operation Management**
   - Go to the admin area (Admin role only)
   - Access the Areas of Operation management interface

2. **Create New Area**
   - Click "Create Area of Operation"
   - Enter:
     - **Name**: Full name (e.g., "North Region")
     - **Code**: Short code (e.g., "NR")
     - **Description**: Optional description
   - Set **Active** status

3. **Areas are now available** in the Staff Management dropdown

## Important Notes

### Role Permissions

- **Admin**: Can create, edit, and assign areas of operation to rotation staff
- **Area Manager**: Can assign areas of operation to rotation staff
- **District Manager**: Can view rotation staff and their areas (read-only)
- **Branch Manager**: Cannot manage rotation staff or areas of operation

### Field Behavior

- **Area of Operation**: Optional field - rotation staff can exist without an area assignment
- **Coverage Area**: Legacy field - kept for backward compatibility
- **Effective Branches**: Separate from Area of Operation - defines which branches the staff can work at

### Best Practices

1. **Use Area of Operation** instead of Coverage Area (legacy field)
2. **Assign areas consistently** - group rotation staff by operational regions
3. **Use areas for filtering** - helps find rotation staff when assigning to branches
4. **Keep area codes short** - makes them easier to identify in dropdowns and tables

## Troubleshooting

### Area of Operation Not Showing in Dropdown

- **Check if areas exist**: Go to Areas of Operation management and verify areas are created
- **Check if area is active**: Only active areas appear in the dropdown
- **Refresh the page**: Try refreshing to reload the areas list

### Cannot Edit Rotation Staff

- **Check your role**: Only Admin and Area Manager can edit rotation staff
- **Verify staff type**: Ensure the staff member is actually a rotation staff member
- **Check permissions**: Contact your administrator if you believe you should have access

### Area Assignment Not Saving

- **Check form validation**: Ensure all required fields are filled
- **Check network**: Verify your internet connection
- **Check console**: Open browser developer tools to see any error messages
- **Try again**: Sometimes a simple retry resolves temporary issues

## Related Features

- **Effective Branches**: Defines which branches rotation staff can work at
- **Rotation Staff Assignment**: Assign rotation staff to specific branches on specific dates
- **Staff Filtering**: Filter rotation staff by area of operation, position, etc.

## See Also

- `SOFTWARE_REQUIREMENTS.md` - Full system requirements
- `docs/business-rules.md` - Business rules and constraints
- Staff Management API documentation

---

**Last Updated:** 2026-01-17 15:30:00

# Create Branch Manager Users

This utility script creates branch manager users for all branches in the system.

## Usage

### Prerequisites

- Database must be running and accessible
- Standard branch codes must be seeded in the database
- Branch manager role must exist in the database

### Running the Script

From the backend directory:

```bash
cd backend
go run cmd/create-branch-managers/main.go
```

### Environment Variables

The script uses the following environment variables (with defaults):

- `DB_HOST` (default: `localhost`)
- `DB_PORT` (default: `5432`)
- `DB_USER` (default: `vsq_user`)
- `DB_PASSWORD` (default: `vsq_password`)
- `DB_NAME` (default: `vsq_manpower`)

### What It Does

For each branch in the system, the script creates two users:

1. **Branch Manager**: `{branchcode}mgr`
   - Example: `cpnmgr` for branch code `CPN`
   - Role: `branch_manager`
   - Password: Same as username (e.g., `cpnmgr`)

2. **Assistant Branch Manager**: `{branchcode}amgr`
   - Example: `cpnamgr` for branch code `CPN`
   - Role: `branch_manager`
   - Password: Same as username (e.g., `cpnamgr`)

### Output

The script will:
- ✅ Show created users with their passwords
- ⏭️ Skip users that already exist
- ❌ Report any errors

### Example Output

```
✅ Created user: cpnmgr (password: cpnmgr) for branch CPN (CPN)
✅ Created user: cpnamgr (password: cpnamgr) for branch CPN (CPN)
⏭️  User ctrmgr already exists, skipping...
...

============================================================
Summary:
  ✅ Created: 60 users
  ⏭️  Skipped: 10 users (already exist)
  ❌ Errors: 0
============================================================
```

### Notes

- Usernames are lowercase versions of branch codes
- Default password is the same as username (for simplicity)
- Users should change their passwords after first login
- Email addresses are auto-generated as `{username}@vsq.local`





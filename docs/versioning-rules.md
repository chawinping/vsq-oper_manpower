---
Date Created: 2025-12-21 09:41:34
Date Updated: 2025-12-21 09:41:34
Version: 1.0.0
---

# Versioning System and Build Time Rules

## Versioning System

**CRITICAL: Automatic Version Updates**

The VSQ Operations Manpower project uses separate versioning for frontend, backend, and database components. **Version numbers and timestamps MUST be updated automatically whenever ANY code change occurs in that component.**

### Version Files

- **Frontend:** `frontend/VERSION.json` and `frontend/public/VERSION.json` (both must be kept in sync)
- **Backend:** `backend/VERSION.json`
- **Database:** `backend/DATABASE_VERSION.json`

### Version Update Rules

**ALWAYS update version files when making code changes:**

1. **Frontend Changes:**
   - Update `frontend/VERSION.json` AND `frontend/public/VERSION.json`
   - Increment version number (semantic versioning: MAJOR.MINOR.PATCH)
   - Update `buildDate` to current date/time (Thailand timezone, UTC+7)
   - Update `buildTime` to current ISO timestamp (Thailand timezone)

2. **Backend Changes:**
   - Update `backend/VERSION.json`
   - Increment version number (semantic versioning: MAJOR.MINOR.PATCH)
   - Update `buildDate` to current date/time (Thailand timezone, UTC+7)
   - Update `buildTime` to current ISO timestamp (Thailand timezone)

3. **Database Changes (migrations, schema changes):**
   - Update `backend/DATABASE_VERSION.json`
   - Increment version number (semantic versioning: MAJOR.MINOR.PATCH)
   - Update `buildDate` to current date/time (Thailand timezone, UTC+7)
   - Update `buildTime` to current ISO timestamp (Thailand timezone)

### Version File Format

```json
{
  "version": "1.0.0",
  "buildDate": "YYYY-MM-DD HH:MM:SS",
  "buildTime": "YYYY-MM-DDTHH:MM:SS+07:00"
}
```

### Version Display

- Versions are displayed on the login page (`frontend/src/app/login/page.tsx`)
- Backend provides version endpoint: `GET /api/v1/version`
- Frontend loads its version from `/VERSION.json` (public folder)
- Backend and database versions are fetched from the API endpoint

### Automatic Update Requirements

**When making ANY code change:**

1. **If changing frontend code** (`frontend/src/**`, `frontend/*.config.*`, `frontend/package.json`, etc.):
   - Update `frontend/VERSION.json` version number and timestamp
   - Update `frontend/public/VERSION.json` version number and timestamp
   - Use current machine date/time (Thailand timezone, UTC+7)

2. **If changing backend code** (`backend/**/*.go`, `backend/go.mod`, `backend/go.sum`, etc.):
   - Update `backend/VERSION.json` version number and timestamp
   - Use current machine date/time (Thailand timezone, UTC+7)

3. **If changing database** (migrations, schema files, `backend/internal/repositories/postgres/migrations/**`):
   - Update `backend/DATABASE_VERSION.json` version number and timestamp
   - Use current machine date/time (Thailand timezone, UTC+7)

4. **If changing multiple components:**
   - Update ALL affected version files
   - Each component version is independent

### Version Numbering

- **MAJOR version:** Breaking changes, incompatible API changes
- **MINOR version:** New features, backward compatible
- **PATCH version:** Bug fixes, small changes

**Example progression:**
- `1.0.0` → `1.0.1` (patch: bug fix)
- `1.0.1` → `1.1.0` (minor: new feature)
- `1.1.0` → `2.0.0` (major: breaking change)

### Getting Current Date/Time

**CRITICAL: ALWAYS use commands to get current date/time - NEVER hardcode dates**

Use PowerShell to get current date/time:

**For Markdown File Headers (Date Created/Date Updated):**
```powershell
powershell -Command "[DateTime]::Now.ToString('yyyy-MM-dd HH:mm:ss')"
```

**For Version Files (buildDate - Thailand timezone UTC+7):**
```powershell
powershell -Command "[DateTime]::Now.AddHours(7).ToString('yyyy-MM-dd HH:mm:ss')"
```

**For Version Files (buildTime - ISO format with timezone):**
```powershell
powershell -Command "[DateTime]::Now.AddHours(7).ToString('yyyy-MM-ddTHH:mm:ss+07:00')"
```

**MANDATORY RULES:**
- ⚠️ **ALWAYS** run the command to get current date/time before updating any date field
- ⚠️ **NEVER** type dates manually (e.g., "2025-01-15 14:30:00")
- ⚠️ **NEVER** copy dates from examples or other files
- ⚠️ **NEVER** use placeholder dates or hardcoded values
- ⚠️ **ALWAYS** use the actual current machine date/time

### Markdown File Standards (Related to Versioning)

**ALWAYS** follow these standards when creating new `.md` files:

1. **Filename Format:**
   - Use descriptive names with kebab-case
   - Examples: `versioning-rules.md`, `deployment-guide.md`, `development-guide.md`
   - Avoid date/time in filenames unless it's a changelog or dated document

2. **File Header:**
   - **MUST** include metadata header at the top of every `.md` file:
   ```markdown
   ---
   Date Created: YYYY-MM-DD HH:MM:SS
   Date Updated: YYYY-MM-DD HH:MM:SS
   Version: 1.0.0
   ---
   ```
   - **CRITICAL: ALWAYS get the current date/time using a command - NEVER use hardcoded or placeholder dates**
     - **MUST** run `powershell -Command "[DateTime]::Now.ToString('yyyy-MM-dd HH:mm:ss')"` to get current date/time
     - **NEVER** type dates manually or use placeholder values like "2025-01-15 14:30:00"
     - **NEVER** copy dates from examples or other files
     - Use the actual current date and time from the machine where the file is being created or updated
   - Update `Date Updated` whenever the file is modified (using the same command to get current time)
   - Increment version number for significant changes:
     - Patch (1.0.1): Minor corrections, typos, formatting
     - Minor (1.1.0): New sections, additional content
     - Major (2.0.0): Structural changes, major rewrites
   - Use local machine timezone for markdown file headers (not Thailand timezone)

3. **Version History:**
   - Include a "Version History" section after the header for tracking changes (optional but recommended for important documents):
   ```markdown
   ## Version History
   - **v1.0.0** (YYYY-MM-DD): Initial creation
   - **v1.0.1** (YYYY-MM-DD): Fixed typo in section X
   - **v1.1.0** (YYYY-MM-DD): Added new section on Y
   ```

## Project-Specific Notes

### VSQ Operations Manpower Project

- **Frontend Technology:** Next.js 14 with TypeScript
- **Backend Technology:** Go (Gin framework) with Clean Architecture
- **Database:** PostgreSQL 15
- **Timezone:** Thailand (UTC+7) for all build timestamps
- **Version Display Location:** Login page (`frontend/src/app/login/page.tsx`)

### Integration Points

1. **Frontend Version Display:**
   - Frontend version is loaded from `/VERSION.json` (public folder)
   - Backend and database versions are fetched via `GET /api/v1/version` endpoint
   - Displayed on login page footer or header

2. **Backend Version Endpoint:**
   - Endpoint: `GET /api/v1/version`
   - Returns JSON with backend and database versions
   - Should be implemented in `backend/internal/handlers/` (create `version_handler.go`)

3. **Version File Locations:**
   - `frontend/VERSION.json` - Source version file
   - `frontend/public/VERSION.json` - Public accessible version file (must match source)
   - `backend/VERSION.json` - Backend version file
   - `backend/DATABASE_VERSION.json` - Database version file

## Summary of Key Principles

1. **Automatic Updates:** Version files MUST be updated whenever code changes occur
2. **Separate Versioning:** Each component (frontend, backend, database) has its own version
3. **Semantic Versioning:** Use MAJOR.MINOR.PATCH format
4. **Never Hardcode Dates:** Always use commands to get current date/time
5. **Timezone Consistency:** Use Thailand timezone (UTC+7) for all version file timestamps
6. **Sync Requirements:** Frontend has two version files that must be kept in sync
7. **Command-Based Dates:** Always use PowerShell commands to get current date/time

## Implementation Checklist

When implementing this system in a new project:

- [x] Create version files for each component (frontend, backend, database)
- [ ] Set up version display endpoints/UI
- [ ] Document version update procedures
- [ ] Create scripts or automation for version updates (optional)
- [ ] Train team on version update requirements
- [ ] Add version update checks to code review process

## Version History

- **v1.0.0** (2025-12-21): Initial creation of versioning rules for VSQ Operations Manpower project





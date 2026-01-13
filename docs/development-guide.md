---
title: Development Guide
description: Guide for developers working on VSQ Operations Manpower
version: 1.3.0
lastUpdated: 2025-12-22 19:14:30
---

# Development Guide

## Overview

This guide provides instructions and best practices for developers working on the VSQ Operations Manpower project.

---

## Prerequisites

### Required Software

- Go 1.21+ (for backend development)
- Node.js 20+ (for frontend development)
- Docker and Docker Compose (for containerized development)
- Git
- IDE/Editor (VS Code recommended)
- PostgreSQL 15+ (or use Docker)

### Required Knowledge

- *To be documented*

---

## Getting Started

### 1. Clone the Repository

```powershell
git clone <repository-url>
cd vsq-oper_manpower
```

### 2. Install Dependencies

```powershell
# Backend dependencies
cd backend
go mod download

# Frontend dependencies
cd ../frontend
npm install
```

### 3. Environment Setup

```powershell
# Copy environment template files
# Backend
cd backend
Copy-Item .env.example .env

# Frontend
cd ../frontend
Copy-Item .env.example .env
```

### 4. Configure Environment Variables

Update `.env` files with appropriate values:
- *To be documented*

### 5. Start Development Servers

#### Option A: Hybrid Development (Recommended for Memory Efficiency)

**Recommended Setup:** Run frontend locally, backend and database in Docker.

This approach saves memory (Next.js dev server uses 1-3GB+), provides faster hot reload, and better debugging experience.

**Step 1: Start Backend and Database in Docker**
```powershell
# Start PostgreSQL and backend-dev with Air live reloading
docker-compose --profile hybrid-dev up -d

# Or use backend-only profile (same thing)
docker-compose --profile backend-only up -d

# View backend logs
docker-compose --profile hybrid-dev logs -f backend-dev
```

**Step 2: Start Frontend Locally**
```powershell
# Navigate to frontend directory
cd frontend

# Install dependencies (first time only)
npm install

# Start Next.js development server
npm run dev
```

**Access Points:**
- Frontend: http://localhost:4000 (or http://127.0.0.1:4000)
- Backend API: http://localhost:8081/api
- Database: localhost:5434

**Benefits:**
- ✅ Lower memory usage (~2-3GB saved vs full Docker)
- ✅ Faster file watching and hot reload
- ✅ Better debugging experience
- ✅ Native Next.js performance
- ✅ Backend still isolated in Docker

#### Option B: Full Docker Development (Alternative)

**Use when:** You want complete isolation or don't have Node.js installed locally.

**Start all services in Docker:**
```powershell
# Start database, backend, and frontend all in Docker
docker-compose --profile fullstack-dev up -d

# View logs
docker-compose --profile fullstack-dev logs -f
```

**Note:** This uses more memory (~4-6GB total) due to Next.js dev server in Docker.

#### Option C: Production Build (Testing)

**For testing production builds:**
```powershell
# Start all services with production builds
docker-compose --profile fullstack up -d
```

#### Option D: Fully Local Development

**Backend with Air (Live Reload):**
```powershell
# Install Air globally (if not already installed)
go install github.com/cosmtrek/air@latest

# Start backend with Air (from backend directory)
cd backend
air
```

**Backend without Air:**
```powershell
cd backend
go run cmd/server/main.go
```

**Frontend:**
```powershell
cd frontend
npm run dev
```

**Note:** Requires PostgreSQL installed locally or running in Docker separately.

---

## Project Structure

```
vsq-oper_manpower/
├── backend/          # Backend application
├── frontend/         # Frontend application
├── docs/            # Documentation
├── requirements.md  # Project requirements
└── CHANGELOG.md     # Project changelog
```

### Backend Structure

```
backend/
├── src/
│   ├── controllers/
│   ├── models/
│   ├── routes/
│   ├── services/
│   ├── middleware/
│   └── utils/
├── tests/
└── package.json
```

### Frontend Structure

```
frontend/
├── src/
│   ├── components/
│   ├── pages/
│   ├── services/
│   ├── utils/
│   └── styles/
├── tests/
└── package.json
```

---

## Live Reloading with Air

The backend uses [Air](https://github.com/cosmtrek/air) for automatic rebuilding when code changes are detected. This provides a seamless development experience.

### How It Works

Air watches for changes in `.go` files and automatically:
1. Rebuilds the application
2. Restarts the server
3. Shows build errors in the console

### Configuration

Air configuration is stored in `backend/.air.toml`. Key settings:
- **Watch directories**: All `.go` files in the project
- **Exclude**: Test files, vendor directory, tmp directory
- **Build command**: `go build -o ./tmp/main ./cmd/server`
- **Output**: Binary is built to `./tmp/main`

### Using Air

**In Docker (Development Mode):**
```powershell
# Start backend-dev service which uses Air
docker-compose --profile dev up -d backend-dev

# View logs to see rebuilds
docker-compose --profile dev logs -f backend-dev
```

**Locally:**
```powershell
cd backend
air
```

### Troubleshooting Air

- **Build errors**: Check `backend/build-errors.log` for detailed error messages
- **Not detecting changes**: Ensure file changes are saved and volumes are properly mounted in Docker
- **Port conflicts**: Make sure port 8081 is not in use by another process

---

## Development Workflow

### 1. Check Requirements

Before starting work on a feature:
- Check `requirements.md` for existing requirements
- If requirement doesn't exist, create it first
- Reference requirement ID in your work

### 2. Create a Branch

```powershell
git checkout -b feature/FR-XX-XX-brief-description
```

### 3. Write Code

- Follow code quality standards (see `.cursorrules`)
- Write tests for business logic
- Add comments referencing requirement IDs
- Follow naming conventions

### 4. Test Your Changes

```powershell
# Backend tests
cd backend
npm test

# Frontend tests
cd ../frontend
npm test
```

### 5. Update Documentation

- Update API documentation if endpoints changed
- Update requirements.md if business logic changed
- Update this guide if processes changed

### 6. Commit Changes

```powershell
git add .
git commit -m "feat FR-XX-XX: Brief description

- Detailed change 1
- Detailed change 2

Related: #issue-number"
```

### 7. Create Pull Request

- Reference requirement IDs in PR description
- Include test results
- Request review

---

## Coding Standards

### General

- Follow the rules defined in `.cursorrules`
- Use consistent naming conventions
- Include error handling
- Validate inputs
- Write meaningful comments

### Backend

- *To be documented based on tech stack*

### Frontend

#### API Response Handling

**CRITICAL: Always handle null/undefined array responses**

When working with API responses that return arrays, always ensure they default to empty arrays to prevent runtime errors.

**Why:** Backend APIs may return `null` instead of empty arrays `[]`, which causes `TypeError: Cannot read properties of null (reading 'map')` when using `.map()` in React components.

**Rule:** Always use null coalescing (`|| []`) when:
1. Extracting arrays from API responses
2. Setting state from API responses
3. Handling errors in API calls

**Examples:**

```typescript
// ✅ CORRECT - API client methods
export const staffApi = {
  list: async () => {
    const response = await apiClient.get('/staff');
    return (response.data.staff || []) as Staff[];
  },
};

// ✅ CORRECT - Component state updates
const loadUsers = async () => {
  try {
    const usersData = await userApi.list();
    setUsers(usersData || []); // Always ensure array
  } catch (error) {
    console.error('Failed to load users:', error);
    setUsers([]); // Set empty array on error
  }
};

// ✅ CORRECT - Promise.all with null coalescing
const [positionsData, branchesData] = await Promise.all([
  positionApi.list(),
  branchApi.list(),
]);
setPositions(positionsData || []);
setBranches(branchesData || []);

// ❌ WRONG - No null handling
const usersData = await userApi.list();
setUsers(usersData); // May be null, causing .map() errors
```

**Files to update when adding new API methods:**
- `frontend/src/lib/api/*.ts` - Add `|| []` to all array-returning methods
- Component files - Add `|| []` when setting state from API responses
- Error handlers - Set empty arrays in catch blocks

**Related Issue:** This prevents `TypeError: Cannot read properties of null (reading 'map')` errors when accessing menu items or pages with empty data.

---

## Testing

### Unit Tests

- Write unit tests for all business logic
- Aim for meaningful coverage
- Test edge cases and error scenarios

### Integration Tests

- Test API endpoints
- Test database interactions
- Test external service integrations

### Running Tests

```powershell
# Backend
cd backend
go test ./...

# Frontend
cd ../frontend
npm test
```

---

## Debugging

### Backend Debugging

**With Air:**
- Air automatically rebuilds on code changes
- Check `backend/build-errors.log` for compilation errors
- View Docker logs: `docker-compose --profile dev logs -f backend-dev`

**Without Air:**
- Use `go run cmd/server/main.go` for manual restarts
- Use `dlv` (Delve) debugger for breakpoints
- Set `GIN_MODE=debug` for verbose logging

### Frontend Debugging

- *To be documented*

---

## Common Tasks

### Adding a New Feature

1. Check/create requirement in `requirements.md`
2. Create feature branch
3. Implement feature with tests
4. Update documentation
5. Commit with requirement ID
6. Create PR

### Adding a New API Endpoint

1. Create route handler
2. Add controller logic
3. Add validation
4. Add error handling
5. Write tests
6. Update API documentation
7. Reference requirement ID

### Adding a New Database Model

1. Create model definition
2. Add migrations if needed
3. Add validation rules
4. Write tests
5. Update database schema documentation

---

## Troubleshooting

### Common Issues

| Issue | Solution |
|-------|----------|
| *To be documented* | *To be documented* |

---

## Resources

- [Requirements](./requirements.md)
- [Business Rules](./business-rules.md)
- [Deployment Guide](./deployment-guide.md)
- [Cursor Rules](../.cursorrules)

---

## Change Log

| Date | Version | Changes | Author |
|------|---------|---------|--------|
| 2025-12-22 | 1.3.0 | Added hybrid development setup (frontend local, backend+db in Docker) for memory efficiency | - |
| 2025-12-18 | 1.2.0 | Added API response handling rules for null array prevention | - |
| 2024-12-19 | 1.0.0 | Initial development guide created | - |

---

## Notes

- Keep this guide updated as the project evolves
- Add project-specific instructions as needed
- Document common issues and solutions


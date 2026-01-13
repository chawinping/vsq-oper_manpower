# VSQ Operations Manpower

Staff allocation system for managing staff across 32 branches of an aesthetic clinic.

## Overview

This system maximizes the efficiency of allocating staff for branches by:
- Managing branch staff and rotation staff
- Scheduling staff workdays and off days
- Assigning rotation staff to branches based on expected revenue
- Providing AI-powered suggestions for optimal staff allocation

## Technology Stack

- **Backend:** Go (Gin framework) with Clean Architecture
- **Frontend:** Next.js 14 with TypeScript
- **Database:** PostgreSQL 15
- **Containerization:** Docker & Docker Compose
- **Testing:** Go unit tests, Playwright E2E tests

## Project Structure

```
vsq-oper_manpower/
├── backend/          # Go backend application
│   ├── cmd/         # Application entry points
│   ├── internal/    # Internal packages
│   ├── pkg/         # Public packages
│   └── tests/       # Test files
├── frontend/         # Next.js frontend application
│   ├── src/         # Source code
│   └── tests/       # E2E tests
├── docs/            # Documentation
├── docker-compose.yml
└── README.md
```

## Getting Started

### Prerequisites

- Docker and Docker Compose
- Go 1.21+ (for local backend development)
- Node.js 20+ (for local frontend development)

### Running with Docker

1. Clone the repository:
```bash
git clone <repository-url>
cd vsq-oper_manpower
```

2. Start all services:
```bash
docker-compose up
```

3. Access the application:
- Frontend: http://localhost:4000
- Backend API: http://localhost:8081
- Database: localhost:5434

### Local Development

#### Option 1: Run Locally (Recommended for Development)

**Backend:**
```bash
cd backend
go mod download
go run cmd/server/main.go
```

**Frontend:**
```bash
cd frontend
npm install
npm run dev
```

#### Option 2: Run with Docker (Hot Reloading Enabled)

**Start all services with hot reloading:**
```bash
docker-compose --profile fullstack-dev up
```

This will:
- Start backend with Air live reloading (auto-restarts on Go file changes)
- Start frontend with Next.js hot module replacement (auto-recompiles on file changes)
- Mount source code as volumes for instant file watching

**Start individual services:**
```bash
# Backend only (with hot reloading)
docker-compose --profile dev up backend-dev

# Frontend only (with hot reloading)
docker-compose --profile fullstack-dev up frontend-dev
```

**Note:** For Windows users, if file watching doesn't work properly, you may need to set `WATCHPACK_POLLING=true` in your environment or `.env` file.

## Environment Variables

### Backend

- `DB_HOST`: Database host (default: localhost)
- `DB_PORT`: Database port (default: 5432, mapped to 5433 on host)
- `DB_USER`: Database user (default: vsq_user)
- `DB_PASSWORD`: Database password (default: vsq_password)
- `DB_NAME`: Database name (default: vsq_manpower)
- `SESSION_SECRET`: Session secret key
- `PORT`: Server port (default: 8080, mapped to 8081 on host)
- `MCP_SERVER_URL`: MCP server URL for AI suggestions
- `MCP_API_KEY`: MCP API key
- `MCP_ENABLED`: Enable MCP integration (true/false)

### Frontend

- `NEXT_PUBLIC_API_URL`: Backend API URL (default: http://localhost:8081/api)
- `PORT`: Server port (default: 3000, mapped to 4000 on host)

## User Roles

1. **Admin:** System configuration, user management
2. **Area Manager/District Manager:** Rotation staff assignment
3. **Branch Manager:** Branch staff scheduling
4. **Viewer:** Read-only access to dashboard and reports

## Features

- ✅ User authentication and authorization
- ✅ Staff management (branch and rotation staff)
- ✅ Branch management with revenue tracking
- ✅ Staff scheduling (monthly calendar view)
- ✅ Rotation staff assignment
- ✅ Business logic for staff allocation
- ✅ MCP integration framework for AI suggestions
- ✅ Multi-language support (Thai/English)
- ✅ Thailand timezone support

## Testing

### Backend Tests

```bash
cd backend
go test ./...
```

### Frontend E2E Tests

```bash
cd frontend
npm run test:e2e
```

## Documentation

- [Software Requirements](SOFTWARE_REQUIREMENTS.md)
- [Software Architecture](SOFTWARE_ARCHITECTURE.md)
- [Development Guide](docs/development-guide.md)
- [Deployment Guide](docs/deployment-guide.md)

## License

[Add license information]

## Contributing

[Add contributing guidelines]


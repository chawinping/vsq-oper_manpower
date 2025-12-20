---
title: Software Architecture Document
description: Architecture documentation for VSQ Operations Manpower System
version: 1.0.0
lastUpdated: 2024-12-19T12:00:00Z
---

# VSQ Operations Manpower - Software Architecture Document

## Document Information

- **Version:** 1.0.0
- **Last Updated:** 2024-12-19T12:00:00Z
- **Status:** Initial Draft

## 1. Architecture Overview

### 1.1 System Architecture

The VSQ Operations Manpower System follows Clean Architecture principles with clear separation of concerns:

```
┌─────────────────────────────────────────────────────────────┐
│                      Frontend (Next.js)                      │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │   Pages/UI   │  │  Components  │  │  API Client   │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
│  ┌──────────────┐  ┌──────────────┐                        │
│  │  i18n (i18n) │  │   Hooks      │                        │
│  └──────────────┘  └──────────────┘                        │
└─────────────────────────────────────────────────────────────┘
                            │
                            │ HTTP/REST API
                            │
┌─────────────────────────────────────────────────────────────┐
│                    Backend (Go/Gin)                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │   Handlers   │  │   Services    │  │ Repositories │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
│  ┌──────────────┐  ┌──────────────┐                        │
│  │  Middleware  │  │   Domain     │                        │
│  └──────────────┘  └──────────────┘                        │
└─────────────────────────────────────────────────────────────┘
                            │
                            │ SQL
                            │
┌─────────────────────────────────────────────────────────────┐
│                  Database (PostgreSQL)                      │
└─────────────────────────────────────────────────────────────┘
                            │
                            │ HTTP
                            │
┌─────────────────────────────────────────────────────────────┐
│              External Services (MCP Server)                 │
└─────────────────────────────────────────────────────────────┘
```

### 1.2 Clean Architecture Layers

#### Backend Layers

1. **Domain Layer** (`internal/domain/`)
   - Models: Core business entities
   - Interfaces: Repository contracts

2. **Use Cases Layer** (`internal/usecases/`)
   - Business logic implementation
   - Allocation engine
   - Business rule enforcement

3. **Handlers Layer** (`internal/handlers/`)
   - HTTP request handling
   - Request/response transformation
   - Input validation

4. **Repository Layer** (`internal/repositories/postgres/`)
   - Database access
   - Data persistence
   - Query implementation

5. **Infrastructure Layer** (`pkg/`)
   - External service clients (MCP)
   - Utility functions
   - Third-party integrations

#### Frontend Layers

1. **Presentation Layer** (`src/app/`)
   - Pages and routes
   - Server-side rendering

2. **Components Layer** (`src/components/`)
   - Reusable UI components
   - Feature-specific components

3. **Business Logic Layer** (`src/lib/`)
   - API clients
   - Business logic hooks
   - Utility functions

4. **Data Layer** (`src/lib/api/`)
   - API communication
   - Data transformation

## 2. Technology Stack

### 2.1 Backend

- **Language:** Go 1.21+
- **Framework:** Gin
- **Database:** PostgreSQL 15
- **Authentication:** Session-based (Gorilla Sessions)
- **Password Hashing:** bcrypt
- **Testing:** Go testing framework

### 2.2 Frontend

- **Framework:** Next.js 14
- **Language:** TypeScript
- **Styling:** CSS (can be extended with Tailwind CSS)
- **Internationalization:** next-i18next
- **HTTP Client:** Axios
- **Date Handling:** date-fns, date-fns-tz
- **Testing:** Playwright

### 2.3 Infrastructure

- **Containerization:** Docker
- **Orchestration:** Docker Compose
- **Database:** PostgreSQL (containerized)
- **Time Zone:** Asia/Bangkok (Thailand)

## 3. Database Schema

### 3.1 Core Tables

```
users
├── id (UUID, PK)
├── username (VARCHAR, UNIQUE)
├── email (VARCHAR, UNIQUE)
├── password_hash (VARCHAR)
├── role_id (UUID, FK → roles)
├── created_at (TIMESTAMP)
└── updated_at (TIMESTAMP)

roles
├── id (UUID, PK)
├── name (VARCHAR, UNIQUE)
└── created_at (TIMESTAMP)

branches
├── id (UUID, PK)
├── name (VARCHAR)
├── code (VARCHAR, UNIQUE)
├── address (TEXT)
├── area_manager_id (UUID, FK → users)
├── expected_revenue (DECIMAL)
├── priority (INTEGER)
├── created_at (TIMESTAMP)
└── updated_at (TIMESTAMP)

staff
├── id (UUID, PK)
├── name (VARCHAR)
├── staff_type (VARCHAR: 'branch' | 'rotation')
├── position_id (UUID, FK → positions)
├── branch_id (UUID, FK → branches, nullable)
├── coverage_area (VARCHAR)
├── created_at (TIMESTAMP)
└── updated_at (TIMESTAMP)

positions
├── id (UUID, PK)
├── name (VARCHAR)
├── min_staff_per_branch (INTEGER)
├── revenue_multiplier (DECIMAL)
└── created_at (TIMESTAMP)

staff_schedules
├── id (UUID, PK)
├── staff_id (UUID, FK → staff)
├── branch_id (UUID, FK → branches)
├── date (DATE)
├── is_working_day (BOOLEAN)
├── created_by (UUID, FK → users)
└── created_at (TIMESTAMP)

rotation_assignments
├── id (UUID, PK)
├── rotation_staff_id (UUID, FK → staff)
├── branch_id (UUID, FK → branches)
├── date (DATE)
├── assignment_level (INTEGER: 1 | 2)
├── assigned_by (UUID, FK → users)
└── created_at (TIMESTAMP)

effective_branches
├── id (UUID, PK)
├── rotation_staff_id (UUID, FK → staff)
├── branch_id (UUID, FK → branches)
├── level (INTEGER: 1 | 2)
└── created_at (TIMESTAMP)

revenue_data
├── id (UUID, PK)
├── branch_id (UUID, FK → branches)
├── date (DATE)
├── expected_revenue (DECIMAL)
├── actual_revenue (DECIMAL, nullable)
├── created_at (TIMESTAMP)
└── updated_at (TIMESTAMP)

system_settings
├── id (UUID, PK)
├── key (VARCHAR, UNIQUE)
├── value (TEXT)
├── description (TEXT)
└── updated_at (TIMESTAMP)

staff_allocation_rules
├── id (UUID, PK)
├── position_id (UUID, FK → positions)
├── min_staff (INTEGER)
├── revenue_threshold (DECIMAL)
├── staff_count_formula (TEXT)
├── created_at (TIMESTAMP)
└── updated_at (TIMESTAMP)
```

## 4. API Design

### 4.1 RESTful API Structure

```
/api
├── /auth
│   ├── POST /login
│   ├── POST /logout
│   └── GET  /me
├── /staff
│   ├── GET    /
│   ├── POST   /
│   ├── PUT    /:id
│   ├── DELETE /:id
│   └── POST   /import
├── /branches
│   ├── GET    /
│   ├── POST   /
│   ├── PUT    /:id
│   └── GET    /:id/revenue
├── /schedules
│   ├── GET  /branch/:branchId
│   ├── POST /
│   └── GET  /monthly
├── /rotation
│   ├── GET    /assignments
│   ├── POST   /assign
│   ├── DELETE /assign/:id
│   ├── GET    /suggestions
│   └── POST   /regenerate-suggestions
├── /settings
│   ├── GET /
│   └── PUT /:key
└── /dashboard
    └── GET /
```

### 4.2 Authentication Flow

1. User submits credentials via `/api/auth/login`
2. Server validates credentials
3. Server creates session and sets cookie
4. Subsequent requests include session cookie
5. Middleware validates session on protected routes

## 5. Security Architecture

### 5.1 Authentication & Authorization

- **Session Management:** Secure cookies with HttpOnly flag
- **Password Security:** bcrypt hashing with salt
- **Role-Based Access:** Middleware enforces role-based permissions
- **CSRF Protection:** SameSite cookie attribute

### 5.2 Data Security

- **SQL Injection Prevention:** Parameterized queries
- **Input Validation:** Request validation at handler level
- **Error Handling:** Generic error messages to prevent information leakage

## 6. Deployment Architecture

### 6.1 Docker Setup

```yaml
Services:
  - postgres: Database service
  - backend: Go API server
  - frontend: Next.js application
```

### 6.2 Environment Configuration

- Environment variables for configuration
- Separate configs for development/production
- Secrets management via environment variables

## 7. Testing Strategy

### 7.1 Backend Testing

- **Unit Tests:** Business logic, repositories
- **Integration Tests:** API endpoints
- **Test Coverage:** Target 70%+

### 7.2 Frontend Testing

- **E2E Tests:** Playwright for critical user flows
- **Component Tests:** (Future) React Testing Library
- **Test Scenarios:**
  - Authentication flow
  - Staff scheduling workflow
  - Rotation assignment workflow
  - Role-based access control

## 8. Scalability Considerations

### 8.1 Database

- Indexes on frequently queried columns
- Connection pooling
- Query optimization

### 8.2 Application

- Stateless API design
- Horizontal scaling capability
- Caching strategy (future)

## 9. Monitoring and Logging

### 9.1 Logging

- Structured logging (to be implemented)
- Error tracking
- Request/response logging

### 9.2 Monitoring

- Health check endpoints
- Performance metrics (future)
- Error rate tracking (future)

## 10. Future Enhancements

- API rate limiting
- Caching layer (Redis)
- Message queue for async processing
- GraphQL API option
- Real-time updates (WebSockets)
- Advanced analytics and reporting

---

## Change Log

| Date | Version | Changes | Author |
|------|---------|---------|--------|
| 2024-12-19 | 1.0.0 | Initial architecture document created | System |




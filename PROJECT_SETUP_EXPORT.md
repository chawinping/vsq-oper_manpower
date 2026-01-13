---
Date Created: 2026-01-08 15:23:26
Date Updated: 2026-01-08 15:23:26
Version: 1.0.0
---

# Project Setup Export Guide

This document exports reusable patterns and configurations for setting up a similar full-stack project with:
- Docker-based development environment
- Hot reload with Air (Go) and Vite (Vue.js)
- Auto-build patterns
- Development rules (date/timezone handling)
- Common framework patterns

---

## Table of Contents

1. [Docker Setup](#1-docker-setup)
2. [Build Patterns](#2-build-patterns)
3. [Hot Reload Configuration](#3-hot-reload-configuration)
4. [Auto-Build Patterns](#4-auto-build-patterns)
5. [Development Rules](#5-development-rules)
6. [Common Framework Patterns](#6-common-framework-patterns)
7. [Quick Start Checklist](#7-quick-start-checklist)

---

## 1. Docker Setup

### 1.1 Docker Compose Structure

**File: `docker-compose.yml`**

```yaml
services:
  # PostgreSQL Database Server
  postgres:
    image: postgres:18-alpine
    container_name: your-project-db
    env_file:
      - env.development
    environment:
      POSTGRES_USER: ${POSTGRES_USER:-db_user}
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD:-secure_password}
      POSTGRES_DB: ${POSTGRES_DB:-your_database}
      POSTGRES_INITDB_ARGS: "--encoding=UTF-8 --lc-collate=C --lc-ctype=C"
    ports:
      - "${DB_PORT_HOST:-5432}:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./backend/setup_database.sql:/docker-entrypoint-initdb.d/setup_database.sql
    networks:
      - app_network
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${POSTGRES_USER:-db_user} -d ${POSTGRES_DB:-your_database}"]
      interval: 10s
      timeout: 5s
      retries: 5
    restart: unless-stopped
    profiles:
      - database
      - fullstack
      - backend

  # Go Backend Server
  backend:
    container_name: your-project-backend
    build:
      context: ./backend
      dockerfile: Dockerfile.dev
    ports:
      - "${BACKEND_PORT_HOST:-8080}:8080"
    environment:
      ENVIRONMENT: ${ENVIRONMENT:-development}
      DATABASE_URL: ${DATABASE_URL:-}
      DB_HOST: ${DB_HOST:-postgres}
      DB_PORT: ${DB_PORT:-5432}
      DB_USER: ${POSTGRES_USER:-db_user}
      DB_PASSWORD: ${POSTGRES_PASSWORD:-secure_password}
      DB_DATABASE: ${POSTGRES_DB:-your_database}
      DB_SSLMODE: ${DB_SSLMODE:-disable}
      CORS_ORIGIN: ${CORS_ORIGIN:-http://localhost:3000}
      GIN_MODE: ${GIN_MODE:-debug}
    volumes:
      # Mount source code for hot reload (air will watch for changes)
      - ./backend:/app
      # Keep backups directory mounted separately
      - ./backend/backups:/app/backups
    depends_on:
      postgres:
        condition: service_healthy
    networks:
      - app_network
    restart: unless-stopped
    profiles:
      - backend
      - fullstack

  # Vue.js Frontend Server
  frontend:
    container_name: your-project-frontend
    build:
      context: ./frontend
      dockerfile: Dockerfile.dev
    ports:
      - "${FRONTEND_PORT_HOST:-3000}:3000"
    environment:
      VITE_API_BASE_URL: ${VITE_API_BASE_URL:-http://localhost:8080}
      VITE_API_URL: ${VITE_API_URL:-http://localhost:8080/api}
    volumes:
      # Mount source code for hot reload (Vite will watch for changes)
      - ./frontend:/app
      # Exclude node_modules from mount (use container's node_modules)
      - /app/node_modules
    depends_on:
      - backend
    networks:
      - app_network
    restart: unless-stopped
    profiles:
      - frontend
      - fullstack

volumes:
  postgres_data:
    driver: local

networks:
  app_network:
    driver: bridge
```

### 1.2 Backend Development Dockerfile

**File: `backend/Dockerfile.dev`**

```dockerfile
# Development Dockerfile with hot reload
FROM golang:alpine

# Install git, ca-certificates, and air for hot reload
RUN apk add --no-cache git ca-certificates

# Install air for hot reload
RUN go install github.com/air-verse/air@latest

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Create backup directory
RUN mkdir -p backups

# Expose port
EXPOSE 8080

# Set environment variables for development
ENV DB_HOST=postgres
ENV DB_PORT=5432
ENV DB_USER=db_user
ENV DB_PASSWORD=secure_password
ENV DB_NAME=your_database
ENV DB_SSLMODE=disable

# Use air for hot reload in development
# Air watches for file changes based on .air.toml configuration
CMD ["air"]
```

### 1.3 Frontend Development Dockerfile

**File: `frontend/Dockerfile.dev`**

```dockerfile
# Development Dockerfile for Vue.js with hot reload
FROM node:20-alpine

# Set working directory
WORKDIR /app

# Copy package files
COPY package*.json ./

# Install all dependencies (including dev dependencies)
RUN npm ci

# Copy source code
COPY . .

# Expose port
EXPOSE 3000

# Set environment variables for development
ENV VITE_API_BASE_URL=http://localhost:8080
ENV VITE_API_URL=http://localhost:8080/api

# Start development server
CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0", "--port", "3000"]
```

### 1.4 Production Dockerfiles

**File: `backend/Dockerfile`**

```dockerfile
# Build stage
FROM golang:alpine AS builder

# Set working directory
WORKDIR /app

# Install git and ca-certificates (needed for go mod download)
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o main ./cmd/server

# Final stage
FROM alpine:latest

# Install ca-certificates, timezone data, and wget for health checks
RUN apk --no-cache add ca-certificates tzdata wget

# Create non-root user for security
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Create app directory and set ownership
WORKDIR /app
RUN mkdir -p backups && \
    chown -R appuser:appgroup /app

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Copy migrations directory (if needed)
COPY --from=builder /app/migrations ./migrations

# Change ownership
RUN chown -R appuser:appgroup main migrations

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/v1/health || exit 1

# Run the application
CMD ["./main"]
```

**File: `frontend/Dockerfile`**

```dockerfile
# Build stage
FROM node:20-alpine AS builder

# Set working directory
WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies (including dev dependencies for build)
RUN npm ci

# Copy source code
COPY . .

# Build the application
RUN npm run build

# Production stage
FROM nginx:alpine

# Install security updates and wget for health checks
RUN apk update && apk upgrade && \
    apk add --no-cache wget

# Copy built application from builder stage
COPY --from=builder /app/dist /usr/share/nginx/html

# Copy nginx configuration template
COPY nginx.conf /etc/nginx/templates/default.conf.template

# Set default BACKEND_URL if not provided
ENV BACKEND_URL=http://backend:8080

# Set proper permissions
RUN chown -R nginx:nginx /usr/share/nginx/html && \
    chown -R nginx:nginx /var/cache/nginx && \
    chown -R nginx:nginx /var/log/nginx

# Expose port
EXPOSE 80

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:80 || exit 1

# Start nginx
CMD ["nginx", "-g", "daemon off;"]
```

---

## 2. Build Patterns

### 2.1 Backend Makefile

**File: `backend/Makefile`**

```makefile
# Backend Makefile

.PHONY: help build build-linux test test-verbose test-coverage clean run dev lint format

# Default target
help:
	@echo "Available targets:"
	@echo "  build         - Build the application"
	@echo "  build-linux   - Build for Linux (for Docker)"
	@echo "  test          - Run tests"
	@echo "  test-verbose  - Run tests with verbose output"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  clean         - Clean build artifacts"
	@echo "  run           - Run the application"
	@echo "  dev           - Run in development mode"
	@echo "  lint          - Run linter"
	@echo "  format        - Format code"

# Build the application
build:
	@echo "Building application..."
	go build -o bin/server ./cmd/server

# Build for Linux (for Docker)
build-linux:
	@echo "Building application for Linux..."
	GOOS=linux GOARCH=amd64 go build -o bin/server-linux ./cmd/server
	@echo "✅ Linux binary built: bin/server-linux"

# Run tests
test:
	@echo "Running tests..."
	go test ./...

# Run tests with verbose output
test-verbose:
	@echo "Running tests with verbose output..."
	go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -f coverage.out coverage.html

# Run the application
run: build
	@echo "Running application..."
	./bin/server

# Run in development mode
dev:
	@echo "Running in development mode with hot reload..."
	@if command -v air > /dev/null 2>&1; then \
		air -c .air.toml; \
	else \
		echo "⚠️  Air not found. Install with: go install github.com/air-verse/air@latest"; \
		echo "Running without hot reload..."; \
		go run ./cmd/server; \
	fi

# Run linter
lint:
	@echo "Running linter..."
	golangci-lint run

# Format code
format:
	@echo "Formatting code..."
	go fmt ./...
	goimports -w .

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Install development tools
install-tools:
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/air-verse/air@latest

# Run all checks
check: lint test
	@echo "All checks passed!"

# Setup development environment
setup: deps install-tools
	@echo "Development environment setup complete!"
```

### 2.2 Frontend Package.json Scripts

**File: `frontend/package.json` (scripts section)**

```json
{
  "scripts": {
    "dev": "vite",
    "prebuild": "node scripts/update-build-time.js",
    "build": "vite build",
    "preview": "vite preview",
    "test:unit": "vitest",
    "test:e2e": "start-server-and-test preview http://localhost:4173 'cypress run --e2e'",
    "test:e2e:dev": "start-server-and-test 'vite dev --port 4173' http://localhost:4173 'cypress open --e2e'",
    "build-only": "vite build",
    "type-check": "vue-tsc --noEmit -p tsconfig.app.json",
    "lint": "eslint . --fix",
    "format": "prettier --write src/"
  }
}
```

---

## 3. Hot Reload Configuration

### 3.1 Air Configuration (Go Hot Reload)

**File: `backend/.air.toml`**

```toml
root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
  args_bin = []
  bin = "./tmp/main"
  # Build command - customize as needed
  cmd = "go build -o ./tmp/main ./cmd/server"
  delay = 500
  exclude_dir = ["assets", "tmp", "vendor", "testdata", "backups"]
  exclude_file = []
  exclude_regex = ["_test.go"]
  exclude_unchanged = false
  follow_symlink = false
  full_bin = ""
  include_dir = []
  include_ext = ["go", "tpl", "tmpl", "html"]
  include_file = []
  kill_delay = "0s"
  log = "build-errors.log"
  # Enable polling for reliable file watching on Windows/Docker
  poll = true
  poll_interval = 500
  rerun = false
  rerun_delay = 500
  send_interrupt = false
  stop_on_root = false

[color]
  app = ""
  build = "yellow"
  main = "magenta"
  runner = "green"
  watcher = "cyan"

[log]
  main_only = false
  time = false

[misc]
  clean_on_exit = false

[screen]
  clear_on_rebuild = false
  keep_scroll = true
```

### 3.2 Vite Configuration (Vue.js Hot Reload)

**File: `frontend/vite.config.ts`**

```typescript
import { fileURLToPath, URL } from 'node:url'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig(({ mode }) => {
  // Determine backend URL based on environment
  const getBackendUrl = () => {
    if (process.env.BACKEND_URL) {
      return process.env.BACKEND_URL;
    }
    if (process.env.DOCKER_ENV === 'true') {
      return 'http://backend:8080';
    }
    return 'http://localhost:8080';
  };

  const backendUrl = getBackendUrl();

  return {
    server: {
      port: 3000,
      host: true,
      // Enable HMR for hot module replacement
      hmr: {
        host: 'localhost',
        port: 3000,
      },
      watch: {
        // Enable polling for reliable file watching on Windows/Docker volumes
        usePolling: true,
        interval: 1000,
        // Exclude directories that cause memory issues
        ignored: [
          '**/node_modules/**',
          '**/dist/**',
          '**/.git/**',
          '**/coverage/**',
          '**/.vite/**',
        ],
      },
      proxy: {
        '/api': {
          target: backendUrl,
          changeOrigin: true,
          secure: false,
          configure: (proxy, _options) => {
            proxy.on('error', (err, _req, res) => {
              console.error('[Vite] Proxy error:', err.message);
              if (res && !res.headersSent) {
                res.writeHead(500, {
                  'Content-Type': 'application/json',
                });
                res.end(JSON.stringify({ 
                  code: 'PROXY_ERROR',
                  message: 'Failed to connect to backend server',
                  details: err.message 
                }));
              }
            });
          },
        },
      },
    },
    plugins: [
      vue(),
    ],
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url))
      },
    },
  };
});
```

---

## 4. Auto-Build Patterns

### 4.1 Pre-Build Scripts

**Backend: Update Build Time Script**

**File: `backend/scripts/update-build-time.go`**

```go
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type Version struct {
	Version  string `json:"version"`
	BuildDate string `json:"buildDate"`
	BuildTime string `json:"buildTime"`
}

func main() {
	// Get current time in Thailand timezone (UTC+7)
	thailandLocation, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		// Fallback to UTC if timezone loading fails
		thailandLocation = time.UTC
	}
	
	now := time.Now().In(thailandLocation)
	
	version := Version{
		Version:   "1.0.0", // Update as needed
		BuildDate: now.Format("2006-01-02 15:04:05"),
		BuildTime: now.Format("2006-01-02T15:04:05+07:00"),
	}
	
	// Write to VERSION.json
	file, err := os.Create("VERSION.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating VERSION.json: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()
	
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(version); err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding VERSION.json: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Println("✅ Build time updated in VERSION.json")
}
```

**Frontend: Update Build Time Script**

**File: `frontend/scripts/update-build-time.js`**

```javascript
import { writeFileSync } from 'fs';
import { join } from 'path';

// Get current time in Thailand timezone (UTC+7)
const now = new Date();
const thailandTime = new Date(now.toLocaleString('en-US', { timeZone: 'Asia/Bangkok' }));

const buildDate = thailandTime.toISOString().slice(0, 19).replace('T', ' ');
const buildTime = thailandTime.toISOString().slice(0, 19) + '+07:00';

const version = {
  version: '1.0.0', // Update as needed
  buildDate: buildDate,
  buildTime: buildTime,
};

// Write to VERSION.json (both locations)
const versionFiles = [
  join(process.cwd(), 'VERSION.json'),
  join(process.cwd(), 'public', 'VERSION.json'),
];

versionFiles.forEach((filePath) => {
  try {
    writeFileSync(filePath, JSON.stringify(version, null, 2) + '\n', 'utf8');
    console.log(`✅ Updated ${filePath}`);
  } catch (error) {
    console.error(`❌ Error writing ${filePath}:`, error.message);
    process.exit(1);
  }
});
```

### 4.2 Vite Plugin for Auto-Versioning

**File: `frontend/scripts/vite-plugin-update-version.js`**

```javascript
/**
 * Vite Plugin: Update Version on Build
 * Automatically updates VERSION.json when Vite rebuilds in dev mode
 */

import { execSync } from 'child_process';
import { fileURLToPath } from 'url';
import { dirname, join } from 'path';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);
const updateScript = join(__dirname, 'update-build-time.js');

let lastUpdateTime = 0;
const UPDATE_THROTTLE = 2000; // Update at most once every 2 seconds

export default function updateVersionPlugin() {
  return {
    name: 'update-version',
    buildStart() {
      // Update version on build start (dev server restart or build)
      try {
        execSync(`node "${updateScript}"`, { stdio: 'inherit' });
        lastUpdateTime = Date.now();
      } catch (error) {
        console.warn('Failed to update version:', error.message);
      }
    },
    handleHotUpdate({ file }) {
      // On file change, update version before rebuild (throttled)
      const now = Date.now();
      if (now - lastUpdateTime < UPDATE_THROTTLE) {
        return; // Skip if updated recently
      }
      
      if (file.endsWith('.vue') || file.endsWith('.ts') || file.endsWith('.js')) {
        try {
          execSync(`node "${updateScript}"`, { stdio: 'pipe' });
          lastUpdateTime = now;
        } catch (error) {
          // Silent fail - don't interrupt dev server
        }
      }
    },
  };
}
```

**Usage in `vite.config.ts`:**

```typescript
import updateVersionPlugin from './scripts/vite-plugin-update-version.js'

export default defineConfig({
  plugins: [
    vue(),
    updateVersionPlugin(), // Update VERSION.json on rebuilds
  ],
});
```

---

## 5. Development Rules

### 5.1 Date/Time Handling Rules

#### ⚠️ CRITICAL RULE

**NEVER use `time.Now()` directly for business logic timestamps.**

**ALWAYS use a timezone helper function instead.**

#### Timezone Helper Package

**File: `backend/pkg/timezone/timezone.go`**

```go
package timezone

import (
	"sync"
	"time"
)

const (
	// BangkokTimezone is the IANA timezone identifier for Bangkok, Thailand (UTC+7)
	BangkokTimezone = "Asia/Bangkok"
)

var (
	bangkokLocation *time.Location
	locationOnce    sync.Once
)

// Location returns the Bangkok timezone location.
// Uses a cached location for performance.
func Location() *time.Location {
	locationOnce.Do(func() {
		loc, err := time.LoadLocation(BangkokTimezone)
		if err != nil {
			// Fallback to UTC if timezone loading fails
			loc = time.UTC
		}
		bangkokLocation = loc
	})
	return bangkokLocation
}

// Now returns the current time in Bangkok timezone (UTC+7).
// Use this instead of time.Now() for all business logic timestamps.
func Now() time.Time {
	return time.Now().In(Location())
}

// LoadLocation loads the Bangkok timezone location.
// Returns error if timezone cannot be loaded.
func LoadLocation() (*time.Location, error) {
	return time.LoadLocation(BangkokTimezone)
}
```

#### Usage Examples

**✅ CORRECT: Use timezone helper**

```go
import "your-project/backend/pkg/timezone"

// Get current time in Bangkok timezone
now := timezone.Now()

// Format timestamp
timestamp := now.Format("2006-01-02 15:04:05")

// Use in database operations
createdAt := timezone.Now()
```

**❌ WRONG: Direct time.Now()**

```go
import "time"

// DON'T DO THIS for business logic
now := time.Now()  // ❌ Uses server timezone, not Bangkok!
```

#### When to Use timezone.Now()

Use `timezone.Now()` for **all business logic timestamps**:
- ✅ Database timestamps (`created_at`, `updated_at`)
- ✅ Booking times
- ✅ Customer creation timestamps
- ✅ Backup filenames
- ✅ Email receipt dates
- ✅ JWT token timestamps
- ✅ Any timestamp displayed to users
- ✅ Any timestamp stored in database

#### When time.Now() is Acceptable

These are **timezone-independent** operations:
- ✅ Duration calculations (`time.Since(start)`)
- ✅ Request ID generation (`time.Now().UnixNano()`)
- ✅ Rate limiting calculations
- ✅ Timeouts (`time.After()`, `time.Sleep()`)
- ✅ Tickers/Timers (`time.NewTicker()`)

**Rule of thumb:** If calculating a **duration** or using time for **scheduling**, `time.Now()` is fine. If creating a **timestamp** for storage or display, use `timezone.Now()`.

#### Date Parsing Rules

**✅ ALWAYS: Use ParseInLocation with timezone**

```go
import "your-project/backend/pkg/timezone"

// Parse date in Bangkok timezone
location := timezone.Location()
date, err := time.ParseInLocation("2006-01-02", dateStr, location)
if err != nil {
    return fmt.Errorf("invalid date format: %w", err)
}
```

**❌ NEVER: Parse dates without timezone context**

```go
// ❌ WRONG - Parses in UTC or server timezone
date, err := time.Parse("2006-01-02", dateStr)
```

#### Frontend Date Handling

**File: `frontend/src/utils/date.ts`**

```typescript
/**
 * Format date in Bangkok timezone
 */
export function formatDate(dateString: string | null | undefined): string {
  if (!dateString) return 'N/A';
  
  try {
    const date = new Date(dateString);
    if (isNaN(date.getTime())) {
      console.warn(`Invalid date string: ${dateString}`);
      return 'Invalid Date';
    }
    
    return date.toLocaleDateString('en-US', {
      timeZone: 'Asia/Bangkok',
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
    });
  } catch (error) {
    console.warn(`Error formatting date: ${dateString}`, error);
    return 'Invalid Date';
  }
}

/**
 * Format datetime in Bangkok timezone
 */
export function formatDateTime(dateString: string | null | undefined): string {
  if (!dateString) return 'N/A';
  
  try {
    const date = new Date(dateString);
    if (isNaN(date.getTime())) {
      console.warn(`Invalid date string: ${dateString}`);
      return 'Invalid Date';
    }
    
    return date.toLocaleString('en-US', {
      timeZone: 'Asia/Bangkok',
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
    });
  } catch (error) {
    console.warn(`Error formatting datetime: ${dateString}`, error);
    return 'Invalid Date';
  }
}

/**
 * Parse date string safely
 */
export function safeParseDate(dateString: string): Date | null {
  if (!dateString) return null;
  
  try {
    const date = new Date(dateString);
    if (isNaN(date.getTime())) {
      console.warn(`Invalid date string: ${dateString}`);
      return null;
    }
    return date;
  } catch (error) {
    console.warn(`Error parsing date: ${dateString}`, error);
    return null;
  }
}
```

### 5.2 Database Timezone Configuration

**File: `backend/internal/database/database.go` (excerpt)**

```go
import (
    "context"
    "database/sql"
    "fmt"
    "time"
    
    _ "github.com/lib/pq"
    "your-project/backend/pkg/timezone"
)

func NewDB(dsn string) (*DB, error) {
    // ... connection setup ...
    
    // Set timezone to Bangkok for all connections
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    _, err = db.ExecContext(ctx, "SET timezone = 'Asia/Bangkok'")
    if err != nil {
        return nil, fmt.Errorf("failed to set timezone: %w", err)
    }
    
    // ... rest of setup ...
}
```

### 5.3 Markdown File Standards

**ALWAYS** follow these standards when creating new `.md` files:

1. **Filename Format:**
   - Include date and time of creation in the filename
   - Format: `YYYY-MM-DD_HH-MM-SS_DESCRIPTIVE_NAME.md`
   - Example: `2026-01-08_15-23-26_DOCKER_SETUP_GUIDE.md`
   - Use 24-hour format for time
   - Use underscores to separate date/time from descriptive name

2. **File Header:**
   ```markdown
   ---
   Date Created: YYYY-MM-DD HH:MM:SS
   Date Updated: YYYY-MM-DD HH:MM:SS
   Version: 1.0.0
   ---
   ```
   - **CRITICAL: ALWAYS get the current date/time using a command - NEVER use hardcoded dates**
   - **MUST** run `powershell -Command "[DateTime]::Now.ToString('yyyy-MM-dd HH:mm:ss')"` to get current date/time
   - Update `Date Updated` whenever the file is modified
   - Increment version number for significant changes

3. **Version History:**
   ```markdown
   ## Version History
   - **v1.0.0** (YYYY-MM-DD): Initial creation
   - **v1.0.1** (YYYY-MM-DD): Fixed typo in section X
   - **v1.1.0** (YYYY-MM-DD): Added new section on Y
   ```

---

## 6. Common Framework Patterns

### 6.1 3-Layer Architecture Pattern

```
┌─────────────────────────────────────────┐
│         HTTP Handlers (v1)              │  ← HTTP concerns only
│  - Request parsing                      │
│  - Response formatting                  │
│  - HTTP status codes                    │
└──────────────┬──────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────┐
│         Service Layer                    │  ← Business logic
│  - Validation & Business Rules           │
│  - Orchestration                        │
│  - Transformations                      │
└──────────────┬──────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────┐
│         Repository Layer                 │  ← Data access
│  - SQL queries                          │
│  - Database operations                  │
│  - Data mapping                         │
└──────────────┬──────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────┐
│         Database (PostgreSQL)            │
└─────────────────────────────────────────┘
```

**Dependency Flow:**
```
Handlers → Services → Repositories → Database
```

**Rules:**
- ✅ Handlers can only call Services
- ✅ Services can call Repositories and other Services
- ✅ Repositories can only access Database
- ❌ Handlers cannot call Repositories directly
- ❌ Services cannot access Database directly
- ❌ Repositories cannot call Services

### 6.2 Error Handling Pattern

**File: `backend/pkg/errors/errors.go`**

```go
package errors

import (
	"fmt"
	"net/http"
)

// AppError represents an application error with HTTP status code
type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"-"`
	Err     error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// New creates a new AppError
func New(code, message string, status int) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Status:  status,
	}
}

// Wrap wraps an error with additional context
func Wrap(err error, code, message string, status int) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Status:  status,
		Err:     err,
	}
}

// Predefined errors
var (
	ErrNotFound     = New("NOT_FOUND", "Resource not found", http.StatusNotFound)
	ErrBadRequest   = New("BAD_REQUEST", "Invalid request", http.StatusBadRequest)
	ErrUnauthorized = New("UNAUTHORIZED", "Unauthorized", http.StatusUnauthorized)
	ErrForbidden    = New("FORBIDDEN", "Forbidden", http.StatusForbidden)
	ErrInternal     = New("INTERNAL_ERROR", "Internal server error", http.StatusInternalServerError)
)
```

**Usage:**

```go
import "your-project/backend/pkg/errors"

// In service layer
func (s *Service) GetByID(ctx context.Context, id int) (*Model, error) {
    model, err := s.repo.GetByID(ctx, id)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, errors.ErrNotFound
        }
        return nil, errors.Wrap(err, "DATABASE_ERROR", "Failed to get model", http.StatusInternalServerError)
    }
    return model, nil
}

// In handler layer
func (h *Handler) GetByID(c *gin.Context) {
    id := c.Param("id")
    model, err := h.service.GetByID(c.Request.Context(), id)
    if err != nil {
        appErr, ok := err.(*errors.AppError)
        if ok {
            c.JSON(appErr.Status, appErr)
            return
        }
        c.JSON(http.StatusInternalServerError, errors.ErrInternal)
        return
    }
    c.JSON(http.StatusOK, model)
}
```

### 6.3 Repository Pattern

**File: `backend/internal/repository/base_repository.go`**

```go
package repository

import (
	"context"
	"database/sql"
	"your-project/backend/internal/database"
)

type BaseRepository struct {
	db *database.DB
}

func NewBaseRepository(db *database.DB) *BaseRepository {
	return &BaseRepository{db: db}
}

func (r *BaseRepository) GetDB() *database.DB {
	return r.db
}

func (r *BaseRepository) GetConnection() *sql.DB {
	return r.db.GetConnection()
}

// WithTx executes a function within a transaction
func (r *BaseRepository) WithTx(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := r.GetConnection().BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()
	
	err = fn(tx)
	return err
}
```

**File: `backend/internal/repository/entity_repository.go`**

```go
package repository

import (
	"context"
	"database/sql"
	"fmt"
	"your-project/backend/internal/models"
)

type EntityRepository interface {
	GetByID(ctx context.Context, id int) (*models.Entity, error)
	List(ctx context.Context) ([]*models.Entity, error)
	Create(ctx context.Context, entity *models.Entity) (*models.Entity, error)
	Update(ctx context.Context, id int, entity *models.Entity) (*models.Entity, error)
	Delete(ctx context.Context, id int) error
}

type entityRepository struct {
	*BaseRepository
}

func NewEntityRepository(baseRepo *BaseRepository) EntityRepository {
	return &entityRepository{BaseRepository: baseRepo}
}

func (r *entityRepository) GetByID(ctx context.Context, id int) (*models.Entity, error) {
	query := `SELECT id, name, created_at, updated_at FROM entities WHERE id = $1`
	
	var entity models.Entity
	err := r.GetConnection().QueryRowContext(ctx, query, id).Scan(
		&entity.ID,
		&entity.Name,
		&entity.CreatedAt,
		&entity.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("entity not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get entity: %w", err)
	}
	
	return &entity, nil
}

// ... other methods ...
```

### 6.4 Service Pattern

**File: `backend/internal/service/entity_service.go`**

```go
package service

import (
	"context"
	"fmt"
	"your-project/backend/internal/models"
	"your-project/backend/internal/repository"
	"your-project/backend/pkg/timezone"
)

type EntityService interface {
	GetByID(ctx context.Context, id int) (*models.Entity, error)
	Create(ctx context.Context, req *CreateEntityRequest) (*models.Entity, error)
	Update(ctx context.Context, id int, req *UpdateEntityRequest) (*models.Entity, error)
	Delete(ctx context.Context, id int) error
}

type entityService struct {
	repo repository.EntityRepository
}

func NewEntityService(repo repository.EntityRepository) EntityService {
	return &entityService{repo: repo}
}

func (s *entityService) GetByID(ctx context.Context, id int) (*models.Entity, error) {
	// Business logic validation
	if id <= 0 {
		return nil, fmt.Errorf("invalid id: %d", id)
	}
	
	// Call repository
	return s.repo.GetByID(ctx, id)
}

func (s *entityService) Create(ctx context.Context, req *CreateEntityRequest) (*models.Entity, error) {
	// Business logic validation
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	
	// Create entity with timestamps
	now := timezone.Now()
	entity := &models.Entity{
		Name:      req.Name,
		CreatedAt: now,
		UpdatedAt: now,
	}
	
	// Call repository
	return s.repo.Create(ctx, entity)
}

// ... other methods ...
```

### 6.5 Handler Pattern

**File: `backend/internal/handlers/v1/entity_handler.go`**

```go
package v1

import (
	"net/http"
	"strconv"
	"your-project/backend/internal/service"
	"your-project/backend/pkg/errors"
	
	"github.com/gin-gonic/gin"
)

type EntityHandler struct {
	service service.EntityService
}

func NewEntityHandler(service service.EntityService) *EntityHandler {
	return &EntityHandler{service: service}
}

func (h *EntityHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.ErrBadRequest)
		return
	}
	
	entity, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		appErr, ok := err.(*errors.AppError)
		if ok {
			c.JSON(appErr.Status, appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, errors.ErrInternal)
		return
	}
	
	c.JSON(http.StatusOK, entity)
}

// ... other handlers ...
```

---

## 7. Quick Start Checklist

### 7.1 Initial Setup

- [ ] Create project structure
- [ ] Set up `docker-compose.yml`
- [ ] Create `backend/Dockerfile.dev` and `backend/Dockerfile`
- [ ] Create `frontend/Dockerfile.dev` and `frontend/Dockerfile`
- [ ] Create `backend/.air.toml` for hot reload
- [ ] Create `frontend/vite.config.ts` with hot reload
- [ ] Set up `backend/Makefile`
- [ ] Configure `frontend/package.json` scripts
- [ ] Create timezone helper package (`backend/pkg/timezone/`)
- [ ] Set up error handling package (`backend/pkg/errors/`)
- [ ] Create base repository pattern
- [ ] Set up environment files (`env.example`, `env.development`)

### 7.2 Development Workflow

- [ ] Install Air: `go install github.com/air-verse/air@latest`
- [ ] Install frontend dependencies: `cd frontend && npm install`
- [ ] Start services: `docker-compose --profile fullstack up -d`
- [ ] Verify hot reload works (make a change, see it reload)
- [ ] Test timezone handling (verify timestamps use correct timezone)
- [ ] Test error handling (verify proper error responses)

### 7.3 Production Build

- [ ] Build backend: `cd backend && make build-linux`
- [ ] Build frontend: `cd frontend && npm run build`
- [ ] Test production Dockerfiles
- [ ] Verify health checks work
- [ ] Test production deployment

---

## Version History

- **v1.0.0** (2026-01-08): Initial export guide creation

---

## Notes

- **Timezone**: This guide uses Bangkok timezone (UTC+7) as an example. Adjust to your project's timezone.
- **Ports**: Default ports are 8080 (backend) and 3000 (frontend). Adjust as needed.
- **Database**: Uses PostgreSQL 18. Adjust version as needed.
- **Framework Versions**: Go 1.23+, Node.js 20+, Vue.js 3. Adjust as needed.

---

## Additional Resources

- [Air Documentation](https://github.com/air-verse/air)
- [Vite Documentation](https://vitejs.dev/)
- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [Go Time Package](https://pkg.go.dev/time)
- [Vue.js Documentation](https://vuejs.org/)

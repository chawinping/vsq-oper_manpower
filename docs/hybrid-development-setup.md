---
title: Hybrid Development Setup
description: Memory-efficient development setup with frontend running locally
version: 1.0.0
lastUpdated: 2025-12-22 19:14:30
---

# Hybrid Development Setup

## Overview

This document explains the recommended hybrid development setup where the frontend runs locally while the backend and database run in Docker containers. This approach optimizes memory usage and provides a better development experience.

---

## Why Hybrid Development?

### Memory Concerns

**Problem:**
- Next.js development server is memory-intensive (typically uses 1-3GB+ RAM)
- Docker adds overhead (~200-500MB per container)
- Running frontend in Docker can consume 4-6GB total memory
- This can be problematic on systems with limited RAM

**Solution:**
- Run frontend locally (uses native Node.js, no Docker overhead)
- Keep backend and database in Docker (lightweight, isolated)
- **Memory savings: ~2-3GB** compared to full Docker setup

### Performance Benefits

| Aspect | Hybrid Setup | Full Docker |
|--------|-------------|-------------|
| **Memory Usage** | ~2-3GB | ~4-6GB |
| **Hot Reload Speed** | Fast (native file watching) | Slower (volume mounts) |
| **File Watching** | Instant (native events) | Delayed (Docker volumes) |
| **Debugging** | Excellent (native tools) | Good (Docker tools) |
| **Build Time** | Faster | Slower |

---

## Architecture

```
┌─────────────────────────────────────────┐
│         Development Environment         │
├─────────────────────────────────────────┤
│                                         │
│  ┌──────────────┐    ┌──────────────┐  │
│  │   Frontend   │    │   Backend    │  │
│  │  (Local)     │───▶│  (Docker)    │  │
│  │  Next.js     │    │  Go + Air    │  │
│  │  Port: 4000  │    │  Port: 8081  │  │
│  └──────────────┘    └──────┬───────┘  │
│                             │          │
│                      ┌──────▼───────┐  │
│                      │  PostgreSQL  │  │
│                      │   (Docker)   │  │
│                      │  Port: 5434  │  │
│                      └──────────────┘  │
│                                         │
└─────────────────────────────────────────┘
```

---

## Setup Instructions

### Prerequisites

1. **Node.js 20+** installed locally
2. **Docker and Docker Compose** installed
3. **Go 1.21+** (optional, only if running backend locally)

### Step 1: Start Backend and Database

```powershell
# Start PostgreSQL and backend-dev in Docker
docker-compose --profile hybrid-dev up -d

# Or use backend-only profile (same thing)
docker-compose --profile backend-only up -d

# Verify services are running
docker-compose --profile hybrid-dev ps

# View backend logs
docker-compose --profile hybrid-dev logs -f backend-dev
```

**What this starts:**
- PostgreSQL database (port 5434)
- Backend with Air live reloading (port 8081)
- Backend automatically connects to database

### Step 2: Start Frontend Locally

```powershell
# Navigate to frontend directory
cd frontend

# Install dependencies (first time only)
npm install

# Start Next.js development server
npm run dev
```

**Frontend will start on:** http://localhost:4000

### Step 3: Verify Setup

1. **Check Backend Health:**
   ```powershell
   curl http://localhost:8081/health
   ```

2. **Check Frontend:**
   - Open browser: http://localhost:4000
   - Should see the application

3. **Check API Connection:**
   - Frontend should connect to backend at `http://localhost:8081/api`
   - Check browser console for any CORS errors

---

## Docker Compose Profiles

The project now supports multiple development profiles:

| Profile | Services | Use Case |
|---------|----------|----------|
| `hybrid-dev` | PostgreSQL + Backend-dev | **Recommended** - Frontend runs locally |
| `backend-only` | PostgreSQL + Backend-dev | Same as hybrid-dev (alias) |
| `dev` | PostgreSQL + Backend-dev | Backend development only |
| `fullstack-dev` | All services in Docker | Full Docker setup (uses more memory) |
| `fullstack` | All services (production builds) | Production testing |

---

## Environment Configuration

### Backend CORS Settings

The backend is configured to accept requests from:
- `http://localhost:4000` (primary)
- `http://localhost:3000` (fallback)
- `http://127.0.0.1:4000` (primary)
- `http://127.0.0.1:3000` (fallback)

This is configured in `docker-compose.yml`:
```yaml
CORS_ALLOWED_ORIGINS: ${CORS_ALLOWED_ORIGINS:-http://localhost:4000,http://localhost:3000,http://127.0.0.1:3000,http://127.0.0.1:4000}
```

### Frontend API URL

The frontend connects to backend at:
- Default: `http://localhost:8081/api`
- Configured in `frontend/next.config.js` or `.env.local`

---

## Development Workflow

### Daily Development

1. **Start backend services:**
   ```powershell
   docker-compose --profile hybrid-dev up -d
   ```

2. **Start frontend (in separate terminal):**
   ```powershell
   cd frontend
   npm run dev
   ```

3. **Make changes:**
   - Frontend: Edit files in `frontend/src/` → Hot reload automatically
   - Backend: Edit files in `backend/` → Air restarts automatically

4. **Stop services:**
   ```powershell
   # Stop frontend: Ctrl+C in frontend terminal
   
   # Stop backend and database
   docker-compose --profile hybrid-dev down
   ```

### Adding Dependencies

**Frontend:**
```powershell
cd frontend
npm install <package-name>
# Restart dev server if needed
```

**Backend:**
```powershell
# Rebuild backend container
docker-compose --profile hybrid-dev build backend-dev
docker-compose --profile hybrid-dev up -d backend-dev
```

---

## Troubleshooting

### CORS Errors

**Problem:** Frontend can't connect to backend API

**Solution:**
1. Check backend CORS settings in `docker-compose.yml`
2. Verify frontend is accessing backend at `http://localhost:8081/api`
3. Check browser console for specific CORS error messages

### Port Conflicts

**Problem:** Port 4000, 5434, or 8081 already in use

**Solution:**
1. **Find process using port:**
   ```powershell
   # PowerShell
   Get-NetTCPConnection -LocalPort 4000
   ```

2. **Stop conflicting service** or change port in:
   - Frontend: `package.json` scripts or `.env.local`
   - Backend: `docker-compose.yml` (BACKEND_PORT)
   - Database: `docker-compose.yml` (DB_PORT)

### Backend Not Starting

**Problem:** Backend container fails to start

**Solution:**
```powershell
# Check logs
docker-compose --profile hybrid-dev logs backend-dev

# Rebuild container
docker-compose --profile hybrid-dev build backend-dev
docker-compose --profile hybrid-dev up -d backend-dev
```

### Frontend Can't Connect to Backend

**Problem:** Network connection issues

**Solution:**
1. Verify backend is running: `curl http://localhost:8081/health`
2. Check `NEXT_PUBLIC_API_URL` in frontend environment
3. Verify Docker network: `docker network ls`
4. Check firewall settings

---

## Memory Usage Comparison

### Typical Memory Usage

| Setup | Memory Usage | Notes |
|-------|-------------|-------|
| **Hybrid (Recommended)** | ~2-3GB | Frontend local, Backend+DB in Docker |
| **Full Docker Dev** | ~4-6GB | All services in Docker |
| **Production Build** | ~1-2GB | Optimized production builds |

### Monitoring Memory

**Check Docker container memory:**
```powershell
docker stats
```

**Check system memory:**
```powershell
# PowerShell
Get-ComputerInfo | Select-Object TotalPhysicalMemory, CsTotalPhysicalMemory
```

---

## When to Use Each Setup

### Use Hybrid Development (Recommended)
- ✅ Active frontend development
- ✅ Memory-constrained systems
- ✅ Fastest iteration cycles needed
- ✅ Better debugging experience desired
- ✅ Node.js already installed locally

### Use Full Docker Development
- ✅ Testing production-like environment
- ✅ Ensuring consistency across team
- ✅ CI/CD pipeline testing
- ✅ Node.js not installed locally
- ✅ Need complete isolation

### Use Production Builds
- ✅ Testing production optimizations
- ✅ Performance testing
- ✅ Pre-deployment verification
- ✅ Staging environment testing

---

## Migration from Full Docker

If you're currently using `fullstack-dev` profile:

1. **Stop current services:**
   ```powershell
   docker-compose --profile fullstack-dev down
   ```

2. **Install Node.js locally** (if not already installed)

3. **Start hybrid setup:**
   ```powershell
   docker-compose --profile hybrid-dev up -d
   cd frontend
   npm install
   npm run dev
   ```

4. **Verify everything works:**
   - Check http://localhost:4000
   - Verify API calls work
   - Check memory usage improvement

---

## Best Practices

1. **Always use hybrid setup** for active development
2. **Keep Docker containers updated** with latest images
3. **Monitor memory usage** regularly
4. **Use production builds** for final testing
5. **Document any custom configurations** in team notes

---

## Related Documentation

- [Development Guide](./development-guide.md) - Complete development guide
- [Docker Dev Hot Reload](./docker-dev-hot-reload.md) - Hot reloading details
- [Docker Setup Analysis](./docker-setup-analysis.md) - Docker configuration analysis

---

## Summary

The hybrid development setup provides:
- ✅ **~2-3GB memory savings** compared to full Docker
- ✅ **Faster hot reload** with native file watching
- ✅ **Better debugging** experience
- ✅ **Same isolation** for backend and database
- ✅ **Flexibility** to switch between setups

**Recommended for:** All developers working on this project, especially those with memory constraints.


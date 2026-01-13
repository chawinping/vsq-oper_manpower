---
title: Quick Start Guide
description: Quick reference for starting development servers
version: 1.0.0
lastUpdated: 2026-01-08 15:42:59
---

# Quick Start Guide - Development Environment

Quick reference for starting frontend, backend, and database servers for development.

---

## Option 1: Hybrid Development (Recommended) ⭐

**Best for:** Active development, memory-constrained systems, fastest iteration

**Setup:** Frontend runs locally, Backend + Database in Docker

### Step 1: Start Backend and Database

```powershell
# Start PostgreSQL and backend-dev in Docker
docker-compose --profile hybrid-dev up -d

# Or use backend-only profile (same thing)
docker-compose --profile backend-only up -d
```

**What starts:**
- ✅ PostgreSQL database (port **5434**)
- ✅ Backend with Air live reloading (port **8081**)

### Step 2: Start Frontend Locally

```powershell
# Navigate to frontend directory
cd frontend

# Install dependencies (first time only)
npm install

# Start Next.js development server
npm run dev
```

**Frontend starts on:** http://localhost:4000

### Verify Everything is Running

```powershell
# Check Docker containers
docker-compose --profile hybrid-dev ps

# Check backend health
curl http://localhost:8081/health

# Check frontend (open in browser)
# http://localhost:4000
```

### Stop Services

```powershell
# Stop frontend: Press Ctrl+C in frontend terminal

# Stop backend and database
docker-compose --profile hybrid-dev down
```

---

## Option 2: Full Docker Development

**Best for:** Complete isolation, testing production-like environment

**Setup:** Everything runs in Docker

### Start All Services

```powershell
# Start PostgreSQL, backend-dev, and frontend-dev in Docker
docker-compose --profile fullstack-dev up -d

# Or run in foreground (see logs)
docker-compose --profile fullstack-dev up
```

**What starts:**
- ✅ PostgreSQL database (port **5434**)
- ✅ Backend with Air live reloading (port **8081**)
- ✅ Frontend with Next.js HMR (port **4000**)

### Verify Everything is Running

```powershell
# Check Docker containers
docker-compose --profile fullstack-dev ps

# Check backend health
curl http://localhost:8081/health

# Check frontend (open in browser)
# http://localhost:4000
```

### Stop Services

```powershell
# Stop all services
docker-compose --profile fullstack-dev down
```

---

## Service URLs

| Service | URL | Port |
|---------|-----|------|
| **Frontend** | http://localhost:4000 | 4000 |
| **Backend API** | http://localhost:8081/api | 8081 |
| **Database** | localhost:5434 | 5434 |

---

## Useful Commands

### View Logs

```powershell
# Backend logs
docker-compose --profile hybrid-dev logs -f backend-dev

# Database logs
docker-compose --profile hybrid-dev logs -f postgres

# All logs
docker-compose --profile hybrid-dev logs -f
```

### Rebuild Containers

```powershell
# Rebuild backend (after dependency changes)
docker-compose --profile hybrid-dev build backend-dev
docker-compose --profile hybrid-dev up -d backend-dev

# Rebuild frontend (after dependency changes)
docker-compose --profile fullstack-dev build frontend-dev
docker-compose --profile fullstack-dev up -d frontend-dev
```

### Check Container Status

```powershell
# List running containers
docker-compose --profile hybrid-dev ps

# Check container resource usage
docker stats
```

---

## Troubleshooting

### Port Already in Use

**Problem:** Port 4000, 5434, or 8081 already in use

**Solution:**
```powershell
# Find process using port (PowerShell)
Get-NetTCPConnection -LocalPort 4000

# Stop conflicting service or change ports in docker-compose.yml
```

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

**Problem:** CORS errors or connection issues

**Solution:**
1. Verify backend is running: `curl http://localhost:8081/health`
2. Check `NEXT_PUBLIC_API_URL` in frontend environment
3. Verify CORS settings in `docker-compose.yml`

### Database Connection Issues

**Problem:** Backend can't connect to database

**Solution:**
```powershell
# Check database is running
docker-compose --profile hybrid-dev ps postgres

# Check database logs
docker-compose --profile hybrid-dev logs postgres

# Restart database
docker-compose --profile hybrid-dev restart postgres
```

---

## Development Workflow

### Daily Development

1. **Start services:**
   ```powershell
   docker-compose --profile hybrid-dev up -d
   cd frontend
   npm run dev
   ```

2. **Make changes:**
   - Frontend: Edit files in `frontend/src/` → Hot reload automatically
   - Backend: Edit files in `backend/` → Air restarts automatically

3. **Stop services:**
   ```powershell
   # Stop frontend: Ctrl+C
   # Stop backend/database:
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

## Memory Usage Comparison

| Setup | Memory Usage | Notes |
|-------|-------------|-------|
| **Hybrid (Recommended)** | ~2-3GB | Frontend local, Backend+DB in Docker |
| **Full Docker Dev** | ~4-6GB | All services in Docker |

---

## Related Documentation

- [Hybrid Development Setup](./hybrid-development-setup.md) - Detailed hybrid setup guide
- [Docker Dev Hot Reload](./docker-dev-hot-reload.md) - Hot reloading details
- [Development Guide](./development-guide.md) - Complete development guide

---

## Quick Reference Card

```
┌─────────────────────────────────────────────────┐
│  HYBRID DEVELOPMENT (RECOMMENDED)              │
├─────────────────────────────────────────────────┤
│  1. docker-compose --profile hybrid-dev up -d  │
│  2. cd frontend                                 │
│  3. npm run dev                                 │
│                                                 │
│  Frontend: http://localhost:4000                │
│  Backend:  http://localhost:8081/api           │
│  Database: localhost:5434                       │
└─────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────┐
│  FULL DOCKER DEVELOPMENT                       │
├─────────────────────────────────────────────────┤
│  docker-compose --profile fullstack-dev up -d   │
│                                                 │
│  Frontend: http://localhost:4000                │
│  Backend:  http://localhost:8081/api           │
│  Database: localhost:5434                       │
└─────────────────────────────────────────────────┘
```

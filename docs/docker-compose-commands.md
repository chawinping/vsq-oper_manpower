---
title: Docker Compose Commands Reference
description: How to restart backend server with docker-compose
version: 1.0.0
lastUpdated: 2025-12-21 09:21:44
---

# Docker Compose Commands - Restarting Backend

## Problem

When you run `docker-compose --profile fullstack up -d`, the backend container won't restart if it's already running. This means the seeding function won't run to populate the branches table.

## Solution Options

### Option 1: Force Recreate Backend (Recommended)

**Command:**
```bash
docker-compose --profile fullstack up -d --force-recreate backend
```

**What it does:**
- Stops and removes the existing backend container
- Creates a new backend container
- Runs migrations and seeding on startup
- Starts the backend service

**Pros:**
- Ensures backend restarts
- Runs migrations/seeding
- Keeps other services running

**Cons:**
- Backend will have brief downtime

---

### Option 2: Rebuild and Recreate Backend

**Command:**
```bash
docker-compose --profile fullstack up -d --build --force-recreate backend
```

**What it does:**
- Rebuilds the backend image (if Dockerfile changed)
- Forces recreation of backend container
- Runs migrations and seeding on startup

**Pros:**
- Picks up code changes
- Ensures fresh start
- Runs migrations/seeding

**Cons:**
- Takes longer (rebuilds image)

---

### Option 3: Restart All Services

**Command:**
```bash
docker-compose --profile fullstack restart backend
```

**What it does:**
- Restarts the backend container without recreating it
- **Note:** This may NOT run migrations/seeding if they already ran

**Pros:**
- Quick restart
- No downtime for other services

**Cons:**
- May not trigger migrations/seeding if container already exists

---

### Option 4: Stop and Start (Full Restart)

**Command:**
```bash
# Stop backend
docker-compose --profile fullstack stop backend

# Start backend (will run migrations/seeding)
docker-compose --profile fullstack up -d backend
```

**What it does:**
- Stops the backend container
- Starts it fresh (runs migrations/seeding)

**Pros:**
- Ensures fresh start
- Runs migrations/seeding

**Cons:**
- Two-step process

---

### Option 5: Recreate All Services

**Command:**
```bash
docker-compose --profile fullstack up -d --force-recreate
```

**What it does:**
- Recreates ALL services (backend, frontend, postgres)
- Runs migrations/seeding on backend startup

**Pros:**
- Ensures everything is fresh
- Runs migrations/seeding

**Cons:**
- Restarts all services (may cause brief downtime)

---

## Recommended Command for Your Use Case

Since you want to ensure the backend restarts to trigger the seeding function:

```bash
docker-compose --profile fullstack up -d --force-recreate backend
```

This will:
1. ✅ Force recreate the backend container
2. ✅ Run migrations (including `SeedStandardBranches()`)
3. ✅ Populate the branches table with 35 standard codes
4. ✅ Keep other services (postgres, frontend) running

---

## Verify Backend Restart Worked

After running the command, verify:

### 1. Check Backend Container Status
```bash
docker ps --filter "name=vsq-manpower-backend"
```

### 2. Check Backend Logs
```bash
docker logs vsq-manpower-backend
```

Look for:
- "Running migrations..."
- "Seeded branch: CPN"
- "Seeded branch: CPN-LS"
- etc.

### 3. Verify Branches Were Created
```bash
docker exec vsq-manpower-db psql -U vsq_user -d vsq_manpower -c "SELECT COUNT(*) FROM branches;"
```

Should return: **35**

### 4. List All Branches
```bash
docker exec vsq-manpower-db psql -U vsq_user -d vsq_manpower -c "SELECT code, name FROM branches ORDER BY code LIMIT 10;"
```

---

## Quick Reference

| Command | Restarts Backend | Runs Migrations | Rebuilds Image |
|---------|-----------------|-----------------|----------------|
| `up -d` | ❌ No | ❌ No | ❌ No |
| `up -d --force-recreate backend` | ✅ Yes | ✅ Yes | ❌ No |
| `up -d --build --force-recreate backend` | ✅ Yes | ✅ Yes | ✅ Yes |
| `restart backend` | ✅ Yes | ❌ Maybe | ❌ No |
| `stop backend && up -d backend` | ✅ Yes | ✅ Yes | ❌ No |

---

## Troubleshooting

### Backend Not Restarting?

1. **Check if container exists:**
   ```bash
   docker ps -a --filter "name=vsq-manpower-backend"
   ```

2. **Force remove and recreate:**
   ```bash
   docker-compose --profile fullstack rm -f backend
   docker-compose --profile fullstack up -d backend
   ```

### Migrations Not Running?

1. **Check backend logs:**
   ```bash
   docker logs vsq-manpower-backend | grep -i migration
   ```

2. **Check for errors:**
   ```bash
   docker logs vsq-manpower-backend | grep -i error
   ```

### Branches Still Empty?

1. **Verify seeding function exists:**
   ```bash
   docker exec vsq-manpower-backend ls -la /app/internal/repositories/postgres/migrations.go
   ```

2. **Manually trigger seeding (if needed):**
   - Connect to backend container
   - Run the seeding function manually
   - Or restart the backend again

---

## Alternative: Create a Script

Create a script `restart-backend.sh`:

```bash
#!/bin/bash
echo "Restarting backend to trigger seeding..."
docker-compose --profile fullstack up -d --force-recreate backend
echo "Waiting for backend to start..."
sleep 5
echo "Checking branches..."
docker exec vsq-manpower-db psql -U vsq_user -d vsq_manpower -c "SELECT COUNT(*) FROM branches;"
```

Make it executable:
```bash
chmod +x restart-backend.sh
```

Run it:
```bash
./restart-backend.sh
```





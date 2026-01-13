---
title: Docker Cleanup Guide
description: Guide for identifying and safely deleting Docker images and volumes associated with this project
version: 1.0.0
lastUpdated: 2025-01-08 15:35:47
---

# Docker Cleanup Guide

This guide helps you identify and safely delete Docker images and volumes associated with the VSQ Operations Manpower project.

---

## Quick Reference: What Can Be Deleted

### ✅ Safe to Delete (Can be Rebuilt)

**Images:**
- All project-built images (backend, frontend, dev variants)
- Old/unused image versions (dangling images)
- Build cache layers

**Volumes:**
- Development volumes (if you don't need the data)
- Staging volumes (if you don't need the data)
- **⚠️ WARNING:** Production volumes contain live data - **DO NOT DELETE** unless you have backups

### ❌ Do NOT Delete

- Base images (`postgres:15-alpine`, `nginx:alpine`) - needed for rebuilds
- Active production volumes (unless you have verified backups)
- Volumes with important data you haven't backed up

---

## 1. Identifying Project Resources

### Check Current Containers

```powershell
# List all containers (running and stopped)
docker ps -a --filter "name=vsq-manpower"

# List containers by environment
docker ps -a | Select-String "vsq-manpower"
```

### Check Current Images

```powershell
# List all images related to this project
docker images | Select-String "vsq-oper_manpower|vsq-manpower"

# List images with more details
docker images --format "table {{.Repository}}\t{{.Tag}}\t{{.Size}}\t{{.CreatedAt}}" | Select-String "vsq"
```

### Check Current Volumes

```powershell
# List all volumes
docker volume ls

# List volumes related to this project
docker volume ls | Select-String "postgres|vsq"
```

### Check Current Networks

```powershell
# List all networks
docker network ls

# List networks related to this project
docker network ls | Select-String "vsq"
```

---

## 2. Docker Images Associated with This Project

### Project-Built Images

Based on the docker-compose files, these images are built for this project:

**Development Images:**
- `vsq-oper_manpower-backend` (or `vsq-oper_manpower-backend-dev`)
- `vsq-oper_manpower-frontend` (or `vsq-oper_manpower-frontend-dev`)

**Staging Images:**
- `vsq-oper_manpower-backend-staging`
- `vsq-oper_manpower-frontend-staging`
- `vsq-manpower-nginx-staging`

**Production Images:**
- `vsq-oper_manpower-backend-production`
- `vsq-oper_manpower-frontend-production`
- `vsq-manpower-nginx-production`

**Base Images (External):**
- `postgres:15-alpine` (from Docker Hub)
- `nginx:alpine` (from Docker Hub)
- `golang:1.24-alpine` (build dependency)
- `node:20-alpine` (build dependency)
- `alpine:latest` (runtime dependency)

### Commands to Delete Images

#### Delete Specific Project Images

```powershell
# Delete development backend image
docker rmi vsq-oper_manpower-backend

# Delete development frontend image
docker rmi vsq-oper_manpower-frontend

# Delete staging images
docker rmi vsq-oper_manpower-backend-staging
docker rmi vsq-oper_manpower-frontend-staging
docker rmi vsq-manpower-nginx-staging

# Delete production images
docker rmi vsq-oper_manpower-backend-production
docker rmi vsq-oper_manpower-frontend-production
docker rmi vsq-manpower-nginx-production
```

#### Delete All Project Images (PowerShell)

```powershell
# Get all project images and delete them
docker images --format "{{.Repository}}:{{.Tag}}" | Where-Object { $_ -match "vsq-oper_manpower|vsq-manpower" } | ForEach-Object { docker rmi $_ }
```

#### Delete Dangling Images (Unused Build Layers)

```powershell
# Delete all dangling images (images without tags)
docker image prune -f

# Delete all unused images (not just dangling)
docker image prune -a -f
```

#### Delete All Unused Images (Including Base Images)

```powershell
# ⚠️ WARNING: This deletes ALL unused images, including base images
docker image prune -a -f
```

---

## 3. Docker Volumes Associated with This Project

### Project Volumes

Based on the docker-compose files, these volumes are created:

**Development Volume:**
- `postgres_data` - Contains development database data

**Staging Volume:**
- `postgres_staging_data` - Contains staging database data

**Production Volume:**
- `postgres_production_data` - Contains production database data

**Anonymous Volumes (from dev containers):**
- `/app/tmp` (backend-dev)
- `/app/vendor` (backend-dev)
- `/app/node_modules` (frontend-dev)
- `/app/.next` (frontend-dev)

### ⚠️ IMPORTANT WARNINGS

**Before deleting volumes:**
1. **Backup your data** - Volumes contain database data that cannot be recovered once deleted
2. **Stop containers first** - Volumes cannot be deleted while containers are using them
3. **Verify backups** - Ensure you have recent backups before deleting production volumes

### Commands to Delete Volumes

#### Stop Containers First

```powershell
# Stop all project containers
docker-compose --profile fullstack --profile fullstack-dev --profile dev down

# Stop staging containers
docker-compose -f docker-compose.yml -f docker-compose.staging.yml down

# Stop production containers
docker-compose -f docker-compose.yml -f docker-compose.production.yml down
```

#### Delete Specific Volumes

```powershell
# Delete development volume (⚠️ WARNING: Deletes all development data)
docker volume rm postgres_data

# Delete staging volume (⚠️ WARNING: Deletes all staging data)
docker volume rm postgres_staging_data

# Delete production volume (⚠️ WARNING: Deletes all production data - BACKUP FIRST!)
docker volume rm postgres_production_data
```

#### Delete All Project Volumes (PowerShell)

```powershell
# ⚠️ WARNING: This deletes all project volumes - BACKUP FIRST!
docker volume ls --format "{{.Name}}" | Where-Object { $_ -match "postgres.*data|vsq" } | ForEach-Object { docker volume rm $_ }
```

#### Delete Unused Volumes

```powershell
# Delete all unused volumes (not attached to any container)
docker volume prune -f
```

---

## 4. Complete Cleanup Scenarios

### Scenario 1: Clean Development Environment (Keep Data)

```powershell
# Stop and remove containers, keep volumes
docker-compose --profile fullstack --profile fullstack-dev --profile dev down

# Remove development images
docker images --format "{{.Repository}}:{{.Tag}}" | Where-Object { $_ -match "vsq-oper_manpower.*dev|vsq-oper_manpower-backend$|vsq-oper_manpower-frontend$" } | ForEach-Object { docker rmi $_ -f }

# Clean up dangling images
docker image prune -f
```

### Scenario 2: Complete Development Cleanup (Delete Everything)

```powershell
# ⚠️ WARNING: This deletes all development data!

# Stop and remove containers
docker-compose --profile fullstack --profile fullstack-dev --profile dev down -v

# Remove development images
docker images --format "{{.Repository}}:{{.Tag}}" | Where-Object { $_ -match "vsq-oper_manpower.*dev|vsq-oper_manpower-backend$|vsq-oper_manpower-frontend$" } | ForEach-Object { docker rmi $_ -f }

# Remove development volume
docker volume rm postgres_data -f

# Clean up dangling images and unused volumes
docker image prune -f
docker volume prune -f
```

### Scenario 3: Clean Staging Environment

```powershell
# Stop staging containers
docker-compose -f docker-compose.yml -f docker-compose.staging.yml down

# Remove staging images
docker images --format "{{.Repository}}:{{.Tag}}" | Where-Object { $_ -match "staging" } | ForEach-Object { docker rmi $_ -f }

# Remove staging volume (⚠️ WARNING: Deletes staging data)
docker volume rm postgres_staging_data -f

# Remove staging network
docker network rm vsq-staging-network
```

### Scenario 4: Clean Production Environment

```powershell
# ⚠️ WARNING: Only run this if you have verified backups!

# Stop production containers
docker-compose -f docker-compose.yml -f docker-compose.production.yml down

# Remove production images
docker images --format "{{.Repository}}:{{.Tag}}" | Where-Object { $_ -match "production" } | ForEach-Object { docker rmi $_ -f }

# Remove production volume (⚠️ WARNING: Deletes production data)
docker volume rm postgres_production_data -f

# Remove production network
docker network rm vsq-production-network
```

### Scenario 5: Nuclear Option - Delete Everything Project-Related

```powershell
# ⚠️ WARNING: This deletes EVERYTHING related to this project!
# ⚠️ Make sure you have backups of all important data!

# Stop all containers
docker-compose --profile fullstack --profile fullstack-dev --profile dev down -v
docker-compose -f docker-compose.yml -f docker-compose.staging.yml down -v
docker-compose -f docker-compose.yml -f docker-compose.production.yml down -v

# Remove all project images
docker images --format "{{.Repository}}:{{.Tag}}" | Where-Object { $_ -match "vsq-oper_manpower|vsq-manpower" } | ForEach-Object { docker rmi $_ -f }

# Remove all project volumes
docker volume ls --format "{{.Name}}" | Where-Object { $_ -match "postgres.*data|vsq" } | ForEach-Object { docker volume rm $_ -f }

# Remove all project networks
docker network ls --format "{{.Name}}" | Where-Object { $_ -match "vsq" } | ForEach-Object { docker network rm $_ -f }

# Clean up everything else
docker system prune -a -f --volumes
```

---

## 5. Checking Disk Space Usage

### Check Image Disk Usage

```powershell
# Show disk usage by images
docker system df

# Show detailed image sizes
docker images --format "table {{.Repository}}\t{{.Tag}}\t{{.Size}}" | Sort-Object
```

### Check Volume Disk Usage

```powershell
# Show volume disk usage
docker system df -v

# Check specific volume size
docker volume inspect postgres_data --format "{{.Mountpoint}}"
```

### Check Total Docker Disk Usage

```powershell
# Show total Docker disk usage
docker system df
```

---

## 6. Safe Cleanup Commands (Recommended)

### Daily/Weekly Cleanup (Safe)

```powershell
# Remove stopped containers
docker container prune -f

# Remove dangling images
docker image prune -f

# Remove unused networks
docker network prune -f

# Remove unused volumes (⚠️ Only if not needed)
docker volume prune -f
```

### Monthly Cleanup (More Aggressive)

```powershell
# Remove all unused images (not just dangling)
docker image prune -a -f

# Remove all unused resources
docker system prune -f
```

---

## 7. Backup Before Cleanup

### Backup Development Database

```powershell
# Backup development database
docker exec vsq-manpower-db pg_dump -U vsq_user vsq_manpower > backup_dev_$(Get-Date -Format "yyyyMMdd_HHmmss").sql
```

### Backup Staging Database

```powershell
# Backup staging database
docker exec vsq-manpower-backend-staging pg_dump -U vsq_user vsq_manpower_staging > backup_staging_$(Get-Date -Format "yyyyMMdd_HHmmss").sql
```

### Backup Production Database

```powershell
# Backup production database (use production script)
.\scripts\backup-database.ps1
```

---

## 8. Summary: What to Delete

### ✅ Safe to Delete Anytime

- **Dangling images** (`docker image prune -f`)
- **Stopped containers** (`docker container prune -f`)
- **Unused networks** (`docker network prune -f`)
- **Project-built images** (can be rebuilt)

### ⚠️ Delete with Caution

- **Development volumes** (only if you don't need the data)
- **Staging volumes** (only if you don't need the data)
- **Unused volumes** (verify they're not needed)

### ❌ Never Delete Without Backup

- **Production volumes** (contains live data)
- **Volumes with important data** (backup first)

---

## 9. Quick Commands Reference

```powershell
# List project containers
docker ps -a --filter "name=vsq-manpower"

# List project images
docker images | Select-String "vsq"

# List project volumes
docker volume ls | Select-String "postgres"

# List project networks
docker network ls | Select-String "vsq"

# Stop all project containers
docker-compose --profile fullstack --profile fullstack-dev --profile dev down

# Remove dangling images (safe)
docker image prune -f

# Remove all unused images (more aggressive)
docker image prune -a -f

# Remove unused volumes (⚠️ verify first)
docker volume prune -f

# Check disk usage
docker system df
```

---

## 10. Troubleshooting

### Volume Won't Delete

**Problem:** `Error: volume is in use`

**Solution:**
```powershell
# Find containers using the volume
docker ps -a --filter volume=postgres_data

# Stop and remove those containers first
docker-compose down -v
```

### Image Won't Delete

**Problem:** `Error: image is referenced in multiple repositories`

**Solution:**
```powershell
# Force remove the image
docker rmi -f <image-id>

# Or remove by ID instead of name
docker images --format "{{.ID}}\t{{.Repository}}" | Select-String "vsq" | ForEach-Object { docker rmi -f $_.Split("`t")[0] }
```

### Network Won't Delete

**Problem:** `Error: network has active endpoints`

**Solution:**
```powershell
# Find containers using the network
docker network inspect vsq-staging-network

# Remove containers first, then network
docker-compose down
docker network rm vsq-staging-network
```

---

## Related Documentation

- `docs/docker-compose-commands.md` - Docker Compose command reference
- `docs/deployment-guide.md` - Deployment procedures
- `scripts/backup-database.ps1` - Database backup script
- `scripts/force-stop-containers.ps1` - Force stop containers utility

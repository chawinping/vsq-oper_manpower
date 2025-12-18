---
title: Deployment Guide
description: Guide for deploying VSQ Operations Manpower to staging and production
version: 2.0.0
lastUpdated: 2025-12-15 19:54:26
---

# Deployment Guide

## Overview

This guide provides comprehensive instructions for deploying the VSQ Operations Manpower application to staging and production environments using Docker Compose.

---

## Prerequisites

### Required Tools

- Docker Engine 20.10+
- Docker Compose 2.0+
- PowerShell 7+ (Windows) or Bash (Linux/macOS)
- Git
- Access to deployment server/environment

### Required Access

- SSH access to deployment server
- Docker daemon access
- Database access (for migrations and backups)
- SSL certificate management (for production)

---

## Environment Configuration

### Development Environment

- **Purpose:** Local development with live reloading
- **URL:** http://localhost:4000
- **Database:** PostgreSQL on localhost:5434
- **Configuration:** `docker-compose.yml` with `dev` or `fullstack-dev` profiles

### Staging Environment

- **Purpose:** Pre-production testing environment
- **URL:** Configured via `NEXT_PUBLIC_API_URL` in `.env.staging`
- **Database:** PostgreSQL container with staging data
- **Configuration:** `docker-compose.staging.yml`
- **Features:** HTTP access, basic rate limiting, health checks

### Production Environment

- **Purpose:** Live production environment
- **URL:** Configured via `NEXT_PUBLIC_API_URL` in `.env.production`
- **Database:** PostgreSQL container with production data
- **Configuration:** `docker-compose.production.yml`
- **Features:** HTTPS only, enhanced security, resource limits, load balancing

---

## Pre-Deployment Checklist

- [ ] All tests passing (`go test ./...` and `npm test`)
- [ ] Code reviewed and approved
- [ ] Requirements documented
- [ ] CHANGELOG.md updated
- [ ] Version numbers updated in `package.json` and `go.mod`
- [ ] Environment variables configured (`.env.staging` or `.env.production`)
- [ ] Database migrations tested
- [ ] Backup strategy verified
- [ ] Rollback plan prepared
- [ ] SSL certificates ready (production only)
- [ ] Monitoring and alerting configured

---

## Environment Setup

### 1. Create Environment Files

**Staging:**
```powershell
# Copy example file
Copy-Item .env.staging.example .env.staging

# Edit with actual values
notepad .env.staging
```

**Production:**
```powershell
# Copy example file
Copy-Item .env.production.example .env.production

# Edit with actual values (use secure editor)
notepad .env.production
```

### 2. Configure Environment Variables

**Critical Variables to Set:**

**Staging (.env.staging):**
```env
DB_PASSWORD=secure_staging_password
SESSION_SECRET=generate_secure_random_string_32_chars_min
NEXT_PUBLIC_API_URL=http://staging.yourdomain.com/api
```

**Production (.env.production):**
```env
DB_PASSWORD=very_secure_production_password
SESSION_SECRET=generate_secure_random_string_64_chars_min
NEXT_PUBLIC_API_URL=https://api.yourdomain.com/api
DB_SSLMODE=require
```

**Generate Secure Secrets:**
```powershell
# Generate SESSION_SECRET (32+ characters)
-join ((48..57) + (65..90) + (97..122) | Get-Random -Count 32 | ForEach-Object {[char]$_})

# Or use OpenSSL
openssl rand -base64 32
```

### 3. SSL Certificates (Production Only)

Place SSL certificates in:
- `nginx/ssl/production/cert.pem` - SSL certificate
- `nginx/ssl/production/key.pem` - SSL private key

**Important:** Never commit SSL certificates to version control!

---

## Deployment Steps

### Staging Deployment

#### Option 1: Using Deployment Script (Recommended)

```powershell
# Deploy with build
.\scripts\deploy-staging.ps1 -Build

# Deploy with pull (if using registry)
.\scripts\deploy-staging.ps1 -Pull

# Standard deployment
.\scripts\deploy-staging.ps1
```

#### Option 2: Manual Deployment

```powershell
# 1. Load environment variables
$env:COMPOSE_FILE = "docker-compose.yml;docker-compose.staging.yml"
Get-Content .env.staging | ForEach-Object {
    if ($_ -match '^\s*([^#][^=]+)=(.*)$') {
        $key = $matches[1].Trim()
        $value = $matches[2].Trim()
        if ($value -and $key) {
            [Environment]::SetEnvironmentVariable($key, $value, "Process")
        }
    }
}

# 2. Stop existing services
docker-compose -f docker-compose.yml -f docker-compose.staging.yml down

# 3. Build images (if needed)
docker-compose -f docker-compose.yml -f docker-compose.staging.yml build

# 4. Start services
docker-compose -f docker-compose.yml -f docker-compose.staging.yml up -d

# 5. Check logs
docker-compose -f docker-compose.yml -f docker-compose.staging.yml logs -f
```

### Production Deployment

#### Option 1: Using Deployment Script (Recommended)

```powershell
# Deploy with confirmation prompt
.\scripts\deploy-production.ps1 -Build

# Deploy with pull (if using registry)
.\scripts\deploy-production.ps1 -Pull -Confirm

# Standard deployment
.\scripts\deploy-production.ps1
```

#### Option 2: Manual Deployment

```powershell
# 1. Create database backup first!
.\scripts\backup-database.ps1 -Environment production

# 2. Load environment variables
$env:COMPOSE_FILE = "docker-compose.yml;docker-compose.production.yml"
Get-Content .env.production | ForEach-Object {
    if ($_ -match '^\s*([^#][^=]+)=(.*)$') {
        $key = $matches[1].Trim()
        $value = $matches[2].Trim()
        if ($value -and $key) {
            [Environment]::SetEnvironmentVariable($key, $value, "Process")
        }
    }
}

# 3. Stop existing services gracefully
docker-compose -f docker-compose.yml -f docker-compose.production.yml down --timeout 30

# 4. Build images
docker-compose -f docker-compose.yml -f docker-compose.production.yml build

# 5. Start services
docker-compose -f docker-compose.yml -f docker-compose.production.yml up -d

# 6. Monitor deployment
docker-compose -f docker-compose.yml -f docker-compose.production.yml logs -f
```

---

## Database Migrations

Migrations run automatically when the backend container starts. To run manually:

```powershell
# Staging
docker exec vsq-manpower-backend-staging ./main migrate

# Production
docker exec vsq-manpower-backend-production ./main migrate
```

---

## Health Checks

### Check Service Health

**Staging:**
```powershell
# Backend
curl http://localhost:8081/health

# Frontend
curl http://localhost:4000/

# Nginx
curl http://localhost/health
```

**Production:**
```powershell
# Backend (via nginx)
curl https://yourdomain.com/api/health

# Frontend
curl https://yourdomain.com/

# Nginx
curl https://yourdomain.com/health
```

### Container Health Status

```powershell
# Check all containers
docker-compose -f docker-compose.yml -f docker-compose.staging.yml ps

# Check specific service
docker inspect vsq-manpower-backend-staging --format='{{.State.Health.Status}}'
```

---

## Monitoring and Logs

### View Logs

**Staging:**
```powershell
# All services
docker-compose -f docker-compose.yml -f docker-compose.staging.yml logs -f

# Specific service
docker-compose -f docker-compose.yml -f docker-compose.staging.yml logs -f backend
```

**Production:**
```powershell
# All services
docker-compose -f docker-compose.yml -f docker-compose.production.yml logs -f

# Specific service with tail
docker-compose -f docker-compose.yml -f docker-compose.production.yml logs --tail=100 backend
```

### Log Locations

- **Backend logs:** Docker logs (json-file driver)
- **Frontend logs:** Docker logs (json-file driver)
- **Nginx logs:** `nginx/logs/` directory
- **Database logs:** Docker logs

---

## Database Backups

### Create Backup

```powershell
# Staging backup
.\scripts\backup-database.ps1 -Environment staging

# Production backup
.\scripts\backup-database.ps1 -Environment production
```

### Restore Backup

```powershell
# Restore from backup file
$backupFile = "backups/production/backup_20241215_120000.sql.zip"
Expand-Archive -Path $backupFile -DestinationPath "backups/temp" -Force
docker exec -i vsq-manpower-db psql -U vsq_user -d vsq_manpower_production < backups/temp/backup_20241215_120000.sql
```

---

## Rollback Procedure

### When to Rollback

- Critical errors in production
- Performance degradation
- Data integrity issues
- Security vulnerabilities
- Failed health checks

### Rollback Steps

1. **Stop current deployment:**
```powershell
docker-compose -f docker-compose.yml -f docker-compose.production.yml down
```

2. **Restore database backup (if needed):**
```powershell
.\scripts\backup-database.ps1 -Environment production
# Then restore from previous backup
```

3. **Checkout previous version:**
```powershell
git checkout <previous-tag-or-commit>
```

4. **Redeploy previous version:**
```powershell
.\scripts\deploy-production.ps1 -Build
```

5. **Verify functionality:**
```powershell
# Run health checks
curl https://yourdomain.com/health
curl https://yourdomain.com/api/health
```

6. **Document the issue:**
- Update CHANGELOG.md
- Create incident report
- Update deployment guide if needed

---

## Troubleshooting

### Common Issues

| Issue | Solution |
|-------|----------|
| **Container won't start** | Check logs: `docker-compose logs <service>` |
| **Database connection failed** | Verify DB credentials in `.env` file |
| **Port already in use** | Change port in `.env` file or stop conflicting service |
| **SSL certificate errors** | Verify certificates in `nginx/ssl/production/` |
| **Health checks failing** | Check service logs and verify dependencies |
| **Out of memory** | Increase Docker memory limit or optimize resource limits |
| **Build failures** | Check Dockerfile and build context |

### Debugging Commands

```powershell
# Check container status
docker ps -a

# Inspect container
docker inspect <container-name>

# Execute command in container
docker exec -it <container-name> /bin/sh

# Check network connectivity
docker network ls
docker network inspect <network-name>

# View resource usage
docker stats

# Check disk space
docker system df
```

---

## Security Checklist

### Pre-Deployment

- [ ] All secrets changed from defaults
- [ ] SSL certificates valid and not expired
- [ ] Database passwords strong (16+ characters)
- [ ] Session secrets random (32+ characters for staging, 64+ for production)
- [ ] Environment variables not committed to git
- [ ] `.env.production` in `.gitignore`
- [ ] SSL certificates not committed to git
- [ ] Security headers configured in nginx
- [ ] Rate limiting enabled
- [ ] HTTPS enforced (production)

### Post-Deployment

- [ ] HTTPS working correctly
- [ ] Security headers present (check with browser dev tools)
- [ ] Rate limiting active
- [ ] Database not exposed publicly
- [ ] Backend API not exposed directly (via nginx only)
- [ ] Logs don't contain sensitive information
- [ ] Monitoring alerts configured

---

## Environment Variables Reference

### Backend Variables

| Variable | Description | Required | Default |
|----------|-------------|----------|---------|
| `DB_HOST` | Database host | Yes | `postgres` |
| `DB_PORT` | Database port | No | `5432` |
| `DB_USER` | Database user | Yes | - |
| `DB_PASSWORD` | Database password | Yes | - |
| `DB_NAME` | Database name | Yes | - |
| `DB_SSLMODE` | SSL mode | No | `disable` (dev), `require` (prod) |
| `SESSION_SECRET` | Session encryption key | Yes | - |
| `PORT` | Backend port | No | `8080` |
| `GIN_MODE` | Gin mode | No | `release` |
| `LOG_LEVEL` | Log level | No | `info` |
| `MCP_SERVER_URL` | MCP server URL | No | - |
| `MCP_API_KEY` | MCP API key | No | - |
| `MCP_ENABLED` | Enable MCP | No | `false` |

### Frontend Variables

| Variable | Description | Required | Default |
|----------|-------------|----------|---------|
| `NEXT_PUBLIC_API_URL` | Backend API URL | Yes | - |
| `NODE_ENV` | Node environment | No | `production` |
| `PORT` | Frontend port | No | `3000` |

---

## Post-Deployment Verification

### Functional Checks

- [ ] Application loads in browser
- [ ] Login functionality works
- [ ] API endpoints respond correctly
- [ ] Database queries execute
- [ ] File uploads work (if applicable)
- [ ] Email notifications work (if applicable)

### Performance Checks

- [ ] Page load times acceptable (< 3s)
- [ ] API response times acceptable (< 500ms)
- [ ] No memory leaks
- [ ] Database queries optimized

### Security Checks

- [ ] HTTPS redirects working
- [ ] Security headers present
- [ ] Rate limiting active
- [ ] No sensitive data in logs
- [ ] Authentication working correctly

---

## Resources

- [Development Guide](./development-guide.md)
- [Docker Setup Analysis](./docker-setup-analysis.md)
- [Requirements](../requirements.md)
- [CHANGELOG](../CHANGELOG.md)
- [Nginx Configuration](../nginx/README.md)

---

## Change Log

| Date | Version | Changes | Author |
|------|---------|---------|--------|
| 2024-12-19 | 1.0.0 | Initial deployment guide created | - |
| 2025-12-15 | 2.0.0 | Added staging and production deployment instructions | - |

---

## Notes

- Always test deployments in staging before production
- Keep deployment procedures updated
- Document environment-specific configurations
- Maintain rollback procedures
- Update this guide as deployment methods change
- Never commit `.env.production` or SSL certificates to version control

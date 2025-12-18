# Deployment Scripts

This directory contains PowerShell scripts for deploying the VSQ Operations Manpower application.

## Scripts

### `deploy-staging.ps1`

Deploys the application to the staging environment.

**Usage:**
```powershell
# Standard deployment
.\scripts\deploy-staging.ps1

# Build images before deployment
.\scripts\deploy-staging.ps1 -Build

# Pull latest images from registry
.\scripts\deploy-staging.ps1 -Pull
```

**Requirements:**
- `.env.staging` file must exist
- Docker and Docker Compose installed
- Sufficient disk space and memory

### `deploy-production.ps1`

Deploys the application to the production environment with safety checks.

**Usage:**
```powershell
# Standard deployment (will prompt for confirmation)
.\scripts\deploy-production.ps1

# Build images before deployment
.\scripts\deploy-production.ps1 -Build

# Skip confirmation prompt (use with caution)
.\scripts\deploy-production.ps1 -Confirm
```

**Safety Features:**
- Requires explicit confirmation before deployment
- Validates critical environment variables
- Creates database backup before deployment
- Checks SSL certificates
- Verifies health checks after deployment

**Requirements:**
- `.env.production` file must exist
- SSL certificates in `nginx/ssl/production/`
- Docker and Docker Compose installed
- Sufficient disk space and memory

### `backup-database.ps1`

Creates a backup of the database for the specified environment.

**Usage:**
```powershell
# Backup staging database
.\scripts\backup-database.ps1 -Environment staging

# Backup production database
.\scripts\backup-database.ps1 -Environment production

# Backup development database
.\scripts\backup-database.ps1 -Environment development
```

**Features:**
- Creates timestamped backup files
- Compresses backups automatically
- Keeps last 10 backups (auto-cleanup)
- Stores backups in `backups/<environment>/` directory

## Environment Files

Before running deployment scripts, ensure you have the appropriate environment file:

- **Staging:** `.env.staging` (copy from `.env.staging.example`)
- **Production:** `.env.production` (copy from `.env.production.example`)

## Prerequisites

1. Docker and Docker Compose installed
2. Appropriate environment file configured
3. Sufficient permissions to run Docker commands
4. Network access to pull images (if using `-Pull` flag)

## Troubleshooting

### Script fails with "file not found"

Ensure you're running the script from the project root directory:
```powershell
cd C:\Users\User\dev_projects\vsq-oper_manpower
.\scripts\deploy-staging.ps1
```

### Environment variables not loading

Check that your `.env` file:
- Exists in the project root
- Has proper format (KEY=VALUE)
- Doesn't have syntax errors
- Has all required variables set

### Docker commands fail

Verify Docker is running:
```powershell
docker ps
```

Check Docker Compose version:
```powershell
docker-compose --version
```

## Best Practices

1. **Always test in staging first** before deploying to production
2. **Create backups** before production deployments
3. **Review changes** in staging environment before production
4. **Monitor logs** after deployment
5. **Verify health checks** after deployment
6. **Keep environment files secure** - never commit them to git



---
title: Staging and Production Setup Summary
description: Quick reference for staging and production deployment setup
version: 1.0.0
lastUpdated: 2025-12-15 20:02:59
---

# Staging and Production Setup Summary

## Quick Start

### Staging Deployment

1. **Create environment file:**
```powershell
Copy-Item .env.staging.example .env.staging
# Edit .env.staging with your values
```

2. **Deploy:**
```powershell
.\scripts\deploy-staging.ps1 -Build
```

3. **Verify:**
```powershell
curl http://localhost:8081/health
```

### Production Deployment

1. **Create environment file:**
```powershell
Copy-Item .env.production.example .env.production
# Edit .env.production with secure values
```

2. **Add SSL certificates:**
```powershell
# Place certificates in:
# nginx/ssl/production/cert.pem
# nginx/ssl/production/key.pem
```

3. **Deploy:**
```powershell
.\scripts\deploy-production.ps1 -Build
```

4. **Verify:**
```powershell
curl https://yourdomain.com/health
```

## File Structure

```
vsq-oper_manpower/
├── docker-compose.yml              # Base configuration
├── docker-compose.staging.yml      # Staging overrides
├── docker-compose.production.yml   # Production overrides
├── .env.staging.example            # Staging env template
├── .env.production.example         # Production env template
├── nginx/
│   ├── staging.conf               # Staging nginx config
│   ├── production.conf            # Production nginx config
│   └── ssl/
│       ├── staging/               # Staging SSL certs
│       └── production/            # Production SSL certs
├── scripts/
│   ├── deploy-staging.ps1         # Staging deployment script
│   ├── deploy-production.ps1      # Production deployment script
│   └── backup-database.ps1        # Database backup script
└── docs/
    ├── deployment-guide.md        # Full deployment guide
    └── docker-setup-analysis.md   # Docker setup analysis
```

## Key Differences: Staging vs Production

| Feature | Staging | Production |
|---------|---------|------------|
| **HTTP Access** | ✅ Allowed | ❌ Redirects to HTTPS |
| **HTTPS** | Optional | Required |
| **Resource Limits** | Basic | Strict |
| **Replicas** | 1 | 2+ |
| **Log Retention** | 3 files, 10MB | 5 files, 50MB |
| **Rate Limiting** | Basic | Enhanced |
| **Security Headers** | Basic | Full |
| **Database SSL** | Prefer | Require |
| **Port Exposure** | Direct | Via Nginx only |

## Environment Variables Checklist

### Required for Both Environments

- [ ] `DB_USER` - Database username
- [ ] `DB_PASSWORD` - Database password (strong!)
- [ ] `DB_NAME` - Database name
- [ ] `SESSION_SECRET` - Session encryption key (32+ chars staging, 64+ production)
- [ ] `NEXT_PUBLIC_API_URL` - Frontend API URL

### Production Only

- [ ] SSL certificates in `nginx/ssl/production/`
- [ ] `DB_SSLMODE=require`
- [ ] Strong passwords (16+ characters)
- [ ] Secure session secret (64+ characters)

## Security Checklist

### Before Deployment

- [ ] All default passwords changed
- [ ] Session secrets generated securely
- [ ] SSL certificates valid and not expired
- [ ] `.env` files not committed to git
- [ ] SSL certificates not committed to git
- [ ] Database credentials strong

### After Deployment

- [ ] HTTPS working (production)
- [ ] Security headers present
- [ ] Rate limiting active
- [ ] Health checks passing
- [ ] No sensitive data in logs
- [ ] Database not publicly accessible

## Common Commands

### View Logs

**Staging:**
```powershell
docker-compose -f docker-compose.yml -f docker-compose.staging.yml logs -f
```

**Production:**
```powershell
docker-compose -f docker-compose.yml -f docker-compose.production.yml logs -f
```

### Check Status

```powershell
docker-compose -f docker-compose.yml -f docker-compose.staging.yml ps
```

### Restart Services

```powershell
docker-compose -f docker-compose.yml -f docker-compose.staging.yml restart
```

### Stop Services

```powershell
docker-compose -f docker-compose.yml -f docker-compose.staging.yml down
```

### Backup Database

```powershell
.\scripts\backup-database.ps1 -Environment staging
.\scripts\backup-database.ps1 -Environment production
```

## Troubleshooting

### Services won't start

1. Check logs: `docker-compose logs <service>`
2. Verify environment variables are set
3. Check port conflicts
4. Verify Docker has enough resources

### Health checks failing

1. Check service logs
2. Verify dependencies are healthy
3. Check network connectivity
4. Verify health check endpoints

### SSL certificate errors

1. Verify certificates exist in correct location
2. Check certificate expiration
3. Verify file permissions
4. Check nginx configuration

## Next Steps

1. **Read the full deployment guide:** [deployment-guide.md](./deployment-guide.md)
2. **Review Docker setup analysis:** [docker-setup-analysis.md](./docker-setup-analysis.md)
3. **Set up monitoring** (recommended)
4. **Configure backups** (automated)
5. **Set up CI/CD pipeline** (optional)

## Support

For issues or questions:
1. Check the [deployment guide](./deployment-guide.md)
2. Review [troubleshooting section](./deployment-guide.md#troubleshooting)
3. Check Docker and service logs
4. Review environment configuration

---

**Remember:** Always test in staging before deploying to production!




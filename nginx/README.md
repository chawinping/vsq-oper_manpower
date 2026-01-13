# Nginx Configuration

This directory contains Nginx reverse proxy configurations for staging and production environments.

## Directory Structure

```
nginx/
├── staging.conf          # Staging environment configuration
├── production.conf       # Production environment configuration
├── ssl/
│   ├── staging/         # SSL certificates for staging
│   └── production/      # SSL certificates for production
└── logs/                # Nginx access and error logs
```

## SSL Certificates

### Staging

Place your SSL certificates in `nginx/ssl/staging/`:
- `cert.pem` - SSL certificate
- `key.pem` - SSL private key

### Production

Place your SSL certificates in `nginx/ssl/production/`:
- `cert.pem` - SSL certificate
- `key.pem` - SSL private key

**Important:** Never commit SSL certificates or private keys to version control.

## Features

### Staging Configuration
- HTTP access allowed
- HTTPS optional (commented out)
- Basic rate limiting
- Health check endpoint

### Production Configuration
- HTTP redirects to HTTPS
- Strict SSL/TLS configuration
- Enhanced security headers
- Rate limiting for API and login endpoints
- Connection limiting
- Static asset caching
- Load balancing with health checks

## Security Headers

Production configuration includes:
- `X-Frame-Options: SAMEORIGIN`
- `X-Content-Type-Options: nosniff`
- `X-XSS-Protection: 1; mode=block`
- `Strict-Transport-Security` (HSTS)
- `Referrer-Policy`

## Rate Limiting

- **API endpoints:** 10 requests/second
- **General routes:** 30 requests/second
- **Login endpoint:** 5 requests/minute
- **Connection limit:** 20 connections per IP

## Health Checks

Health check endpoint available at `/health`:
```bash
curl http://your-domain/health
```

## Updating Configuration

After modifying nginx configuration files, restart the nginx container:

```powershell
docker-compose restart nginx
```

Or rebuild and restart:

```powershell
docker-compose up -d --force-recreate nginx
```







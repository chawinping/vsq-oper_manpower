# Docker Development with Hot Reloading

## Overview

The Docker development configuration supports hot reloading for both frontend and backend. However, **for memory efficiency, we recommend running the frontend locally** while keeping backend and database in Docker (hybrid approach).

## Recommended Setup: Hybrid Development

**Why Hybrid?**
- Next.js dev server uses 1-3GB+ memory
- Docker adds ~200-500MB overhead per container
- Local frontend provides faster file watching and better debugging
- Saves ~2-3GB memory compared to full Docker setup

**How to Use:**
```powershell
# 1. Start backend and database in Docker
docker-compose --profile hybrid-dev up -d

# 2. Start frontend locally (in separate terminal)
cd frontend
npm install  # First time only
npm run dev
```

See [Development Guide](./development-guide.md) for detailed instructions.

---

## Full Docker Development (Alternative)

The Docker development configuration supports hot reloading for both frontend and backend in Docker, allowing you to see changes instantly without rebuilding containers.

## What Changed

### Frontend Hot Reloading

1. **Created `frontend/Dockerfile.dev`**
   - Lightweight development image
   - Runs `npm run dev` instead of production build
   - Enables Next.js Hot Module Replacement (HMR)

2. **Updated `docker-compose.yml`**
   - `frontend-dev` service now uses `Dockerfile.dev`
   - Added volume mounts for source code
   - Set `NODE_ENV=development`
   - Added `WATCHPACK_POLLING` environment variable for Windows compatibility

### Backend Hot Reloading

- Already configured with Air (live reloading tool)
- Uses `Dockerfile.dev` with volume mounts
- Auto-restarts on Go file changes

## How to Use

### Option 1: Hybrid Development (Recommended)

**Backend + Database in Docker, Frontend Local:**

```powershell
# Start backend and database
docker-compose --profile hybrid-dev up -d

# In separate terminal, start frontend locally
cd frontend
npm run dev
```

**Benefits:**
- ✅ Lower memory usage (~2-3GB saved)
- ✅ Faster hot reload
- ✅ Better debugging experience
- ✅ Native Next.js performance

### Option 2: Full Docker Development

**All services in Docker:**

```powershell
docker-compose --profile fullstack-dev up
```

This starts:
- **PostgreSQL** database
- **Backend** with Air live reloading
- **Frontend** with Next.js HMR (in Docker)

**Note:** Uses more memory (~4-6GB total) due to Next.js dev server in Docker.

### Start Individual Services

**Frontend only:**
```powershell
docker-compose --profile fullstack-dev up frontend-dev
```

**Backend only:**
```powershell
docker-compose --profile dev up backend-dev
```

### Rebuild After Dependency Changes

If you add new npm packages or Go modules:

**Frontend:**
```powershell
docker-compose --profile fullstack-dev build frontend-dev
docker-compose --profile fullstack-dev up frontend-dev
```

**Backend:**
```powershell
docker-compose --profile dev build backend-dev
docker-compose --profile dev up backend-dev
```

## How It Works

### Frontend

1. **Volume Mounts:**
   - `./frontend:/app` - Mounts source code
   - `/app/node_modules` - Anonymous volume (preserves installed packages)
   - `/app/.next` - Anonymous volume (preserves build cache)

2. **File Watching:**
   - Next.js watches for changes in mounted source code
   - Changes trigger automatic recompilation
   - Browser updates via Hot Module Replacement

3. **Windows Compatibility:**
   - `WATCHPACK_POLLING=true` enables polling-based file watching
   - Required because Windows file system events don't work well in Docker

### Backend

1. **Air Live Reloading:**
   - Watches Go source files
   - Automatically rebuilds and restarts on changes
   - Uses `.air.toml` configuration

2. **Volume Mounts:**
   - `./backend:/app` - Mounts source code
   - `/app/tmp` - Anonymous volume for temporary files
   - `/app/vendor` - Anonymous volume for Go modules

## Troubleshooting

### File Changes Not Detected (Windows)

If hot reloading doesn't work on Windows:

1. **Set polling explicitly:**
   ```powershell
   $env:WATCHPACK_POLLING="true"
   docker-compose --profile fullstack-dev up
   ```

2. **Or add to `.env` file:**
   ```
   WATCHPACK_POLLING=true
   ```

### Port Already in Use

If port 4000 or 8081 is already in use:

1. **Stop existing containers:**
   ```powershell
   docker-compose --profile fullstack-dev down
   ```

2. **Or change ports in `.env`:**
   ```
   FRONTEND_PORT=4001
   BACKEND_PORT=8082
   ```

### Node Modules Not Found

If you see "module not found" errors:

1. **Rebuild the container:**
   ```powershell
   docker-compose --profile fullstack-dev build frontend-dev
   ```

2. **Or install locally and copy:**
   ```powershell
   cd frontend
   npm install
   ```

## Performance Notes

### Memory Usage Comparison

| Setup | Estimated Memory Usage |
|-------|----------------------|
| **Hybrid (Recommended)** | ~2-3GB (Backend+DB in Docker, Frontend local) |
| **Full Docker Dev** | ~4-6GB (All services in Docker) |
| **Production Build** | ~1-2GB (Optimized builds) |

### Performance Characteristics

- **First Start:** May take 30-60 seconds to compile
- **Subsequent Changes:** Usually < 1 second for hot reload
- **Memory Usage:** Development mode uses more memory than production
- **File Watching:** Uses more CPU than production builds
- **Docker Overhead:** Adds ~200-500MB per container

## Comparison: Hybrid vs Full Docker Dev

| Feature | Hybrid (Recommended) | Full Docker Dev |
|---------|---------------------|-----------------|
| **Memory Usage** | ~2-3GB | ~4-6GB |
| **Hot Reload Speed** | Fast (native) | Slower (volume mounts) |
| **File Watching** | Native (instant) | Via volumes (delayed) |
| **Debugging** | Excellent (native tools) | Good (Docker tools) |
| **Setup Complexity** | Medium (need Node.js) | Low (all in Docker) |
| **Isolation** | Partial | Complete |
| **Consistency** | Varies (Node.js version) | Same everywhere |
| **Port Conflicts** | Possible | Isolated |

## Best Practices

1. **Use Hybrid Development** (Recommended):
   - For active frontend development
   - When memory is a concern
   - For fastest iteration cycles
   - When you have Node.js installed locally

2. **Use Full Docker Dev** for:
   - Testing in production-like environment
   - Ensuring consistency across team
   - CI/CD pipeline testing
   - When you don't want to install Node.js locally
   - When you need complete isolation

3. **Rebuild containers** when:
   - Adding new dependencies
   - Changing Dockerfile
   - Updating base images

4. **Memory Optimization Tips:**
   - Use hybrid setup to save memory
   - Stop unused containers: `docker-compose down`
   - Monitor memory: `docker stats`
   - Consider production builds for testing



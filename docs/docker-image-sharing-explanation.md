---
title: Docker Image Sharing Explained
description: How Docker images are shared across projects
version: 1.0.0
lastUpdated: 2025-01-08
---

# Docker Image Sharing Explained

## Short Answer: YES ✅

**Docker images are shared across ALL projects on your machine.** This is by design and is actually efficient!

## How Docker Image Sharing Works

### Image Storage

Docker images are stored **globally** on your system, not per-project:

```
Your Computer
├── Docker Image Cache (Global)
│   ├── postgres:18-alpine (402MB) ← Shared by ALL projects
│   ├── postgres:15-alpine (392MB) ← Shared by ALL projects
│   └── ... other images ...
│
├── Project 1 (vsq-oper_manpower)
│   └── Uses: postgres:18-alpine
│
├── Project 2 (another-project)
│   └── Uses: postgres:18-alpine ← Same image!
│
└── Project 3 (yet-another-project)
    └── Uses: postgres:18-alpine ← Same image!
```

### How It Works

1. **Image is Pulled Once:**
   ```powershell
   # First project pulls the image
   docker pull postgres:18-alpine
   # Downloads: 402MB
   ```

2. **Other Projects Use Same Image:**
   ```powershell
   # Second project uses the same image
   docker-compose up -d postgres
   # Uses existing image: 0MB download
   ```

3. **Each Container Gets Its Own Data:**
   - Image is **read-only**
   - Each container creates a **writable layer** on top
   - Each container has its own **volume** for data
   - Containers are **completely isolated**

## Benefits of Image Sharing

### ✅ Disk Space Efficiency

**Without sharing (if images were per-project):**
- Project 1: postgres:18-alpine = 402MB
- Project 2: postgres:18-alpine = 402MB
- Project 3: postgres:18-alpine = 402MB
- **Total: 1.2GB**

**With sharing (current Docker behavior):**
- All projects: postgres:18-alpine = 402MB (shared)
- **Total: 402MB** ✅

### ✅ Faster Container Creation

- First project: Downloads image (takes time)
- Other projects: Uses existing image (instant)

### ✅ Consistency

- All projects use the **exact same image**
- Same PostgreSQL version
- Same base configuration

## Container Isolation

Even though images are shared, **containers are completely isolated:**

```
postgres:18-alpine (Shared Image - Read-Only)
│
├── Container 1 (Project A)
│   ├── Writable Layer (container-specific)
│   └── Volume: project_a_data (isolated)
│
├── Container 2 (Project B)
│   ├── Writable Layer (container-specific)
│   └── Volume: project_b_data (isolated)
│
└── Container 3 (Project C)
    ├── Writable Layer (container-specific)
    └── Volume: project_c_data (isolated)
```

**Each container:**
- ✅ Has its own database data (in volumes)
- ✅ Has its own network
- ✅ Has its own environment variables
- ✅ Cannot access other containers' data
- ✅ Runs independently

## Real Example

### Your Current Setup

```powershell
# Check what's using postgres:18-alpine
docker ps --filter "ancestor=postgres:18-alpine" --format "{{.Names}}\t{{.Image}}"
```

**You might see:**
- `vsq-manpower-db` (your project)
- `other-project-db` (another project)
- `yet-another-db` (another project)

All using the **same image**, but:
- Different container names
- Different volumes (data isolation)
- Different ports
- Different networks

## What Gets Shared vs. Isolated

### ✅ Shared (Image Level)

- **Image files** (read-only layers)
- **Base configuration** (PostgreSQL binaries, etc.)
- **Image metadata**

### 🔒 Isolated (Container Level)

- **Database data** (stored in volumes)
- **Environment variables**
- **Port mappings**
- **Network connections**
- **Container state**

## Multiple Projects Using Same Image

### Example: Two Projects Using postgres:18-alpine

**Project 1 (vsq-oper_manpower):**
```yaml
# docker-compose.yml
services:
  postgres:
    image: postgres:18-alpine
    container_name: vsq-manpower-db
    volumes:
      - vsq_postgres_data:/var/lib/postgresql/data  # Isolated volume
    ports:
      - "5434:5432"  # Different port
```

**Project 2 (other-project):**
```yaml
# docker-compose.yml
services:
  postgres:
    image: postgres:18-alpine  # Same image!
    container_name: other-project-db
    volumes:
      - other_postgres_data:/var/lib/postgresql/data  # Different volume
    ports:
      - "5435:5432"  # Different port
```

**Result:**
- ✅ Same image (shared, saves space)
- ✅ Different containers (isolated)
- ✅ Different data (isolated volumes)
- ✅ Different ports (no conflicts)

## Checking Image Usage

### See What Containers Use an Image

```powershell
# Find all containers using postgres:18-alpine
docker ps -a --filter "ancestor=postgres:18-alpine" --format "{{.Names}}\t{{.Image}}\t{{.Status}}"
```

### See Image Size and Usage

```powershell
# Check image size
docker images postgres:18-alpine

# See disk usage
docker system df
```

## Best Practices

### ✅ DO: Share Images

- Use official images like `postgres:18-alpine`
- Let Docker manage image sharing
- Saves disk space and time

### ✅ DO: Isolate Data

- Use separate volumes for each project
- Use different container names
- Use different ports

### ⚠️ DON'T: Worry About Sharing

- Images are read-only
- Containers are isolated
- Data is separate
- No security concerns

## Common Questions

### Q: Will my data mix with other projects?

**A:** No! Each container has its own volume. Data is completely isolated.

### Q: Can I use different PostgreSQL versions?

**A:** Yes! You can have:
- Project 1: `postgres:15-alpine`
- Project 2: `postgres:18-alpine`
- Project 3: `postgres:16-alpine`

All coexist on the same machine.

### Q: What if I delete the image?

**A:** If you delete `postgres:18-alpine`:
- All containers using it will stop working
- You'll need to pull it again
- But volumes (data) are safe - they're separate

### Q: Can I have project-specific images?

**A:** Yes, you can:
- Tag images differently: `myproject-postgres:18-alpine`
- Use different image names
- But sharing is more efficient

## Summary

**✅ YES - Images are shared across projects**

**Benefits:**
- Saves disk space
- Faster setup
- Consistent versions

**Safety:**
- Containers are isolated
- Data is separate (volumes)
- No data mixing
- No security issues

**Your current setup:**
- `postgres:18-alpine` is shared
- `vsq-manpower-db` container is isolated
- `vsq-oper_manpower_postgres_data` volume is isolated
- Safe to use in other projects!

## Verification

Check if other projects are using the same image:

```powershell
# See all containers using postgres:18-alpine
docker ps -a --filter "ancestor=postgres:18-alpine"

# See image details
docker image inspect postgres:18-alpine --format "{{.RepoTags}} | Size: {{.Size}} bytes"
```

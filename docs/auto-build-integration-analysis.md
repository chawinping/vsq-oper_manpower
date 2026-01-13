---
title: Auto-Build Integration Analysis
description: Analysis of PROJECT_SETUP_EXPORT.md patterns for auto-build integration
version: 1.0.0
lastUpdated: 2026-01-08 15:41:49
---

# Auto-Build Integration Analysis

This document analyzes `PROJECT_SETUP_EXPORT.md` (from another project) to identify auto-build patterns that can be integrated into the VSQ Operations Manpower project, focusing on **automatic rebuilds after code changes without explicit rebuild commands**.

---

## Current State

### Backend Auto-Build (✅ Already Implemented)

**Status:** ✅ Fully functional

**Implementation:**
- Uses **Air** (`github.com/cosmtrek/air@v1.49.0`) for live reloading
- Configured in `backend/.air.toml` with:
  - Polling enabled (`poll = true`, `poll_interval = 500`) for Windows/Docker compatibility
  - Watches `.go` files and automatically rebuilds
  - Excludes test files and vendor directories
- Docker volume mounts: `./backend:/app` for source code
- **Result:** Changes to Go files automatically trigger rebuild and restart

**How it works:**
1. Air watches mounted source code directory
2. On `.go` file change → triggers build
3. Build completes → restarts application
4. No manual rebuild needed ✅

### Frontend Auto-Build (✅ Already Implemented)

**Status:** ✅ Fully functional

**Implementation:**
- Uses **Next.js** built-in Hot Module Replacement (HMR)
- Docker volume mounts: `./frontend:/app` with node_modules excluded
- `WATCHPACK_POLLING=true` for Windows compatibility
- **Result:** Changes to frontend files automatically trigger recompilation

**How it works:**
1. Next.js dev server watches mounted source code
2. On file change → triggers HMR
3. Browser automatically updates
4. No manual rebuild needed ✅

---

## Patterns from PROJECT_SETUP_EXPORT.md

### 1. Auto-Version Update on File Changes

**From PROJECT_SETUP_EXPORT.md:**
- Vite plugin that updates `VERSION.json` on every file change (not just build)
- Updates build timestamps automatically during development

**Current Project:**
- ✅ Updates version on `predev` and `prebuild` scripts
- ❌ Does NOT update version on every file change during development

**Gap:** Version timestamps only update when:
- Starting dev server (`predev`)
- Building (`prebuild`)
- NOT when making code changes during development

**Recommendation:** Create a Next.js webpack plugin or use file watcher to update version timestamps on file changes (similar to Vite plugin pattern).

### 2. Enhanced Air Configuration

**From PROJECT_SETUP_EXPORT.md:**
- More detailed `.air.toml` configuration
- Better exclusion patterns
- Clearer documentation

**Current Project:**
- ✅ Has `.air.toml` with polling enabled
- ⚠️ Could benefit from additional optimizations

**Gap:** Minor - current config is functional but could be optimized.

**Recommendation:** Review and optimize `.air.toml` based on PROJECT_SETUP_EXPORT.md patterns.

### 3. Pre-Build Scripts Pattern

**From PROJECT_SETUP_EXPORT.md:**
- Separate scripts for updating build time
- Reusable Go/JavaScript scripts

**Current Project:**
- ✅ Has inline scripts in `package.json` for version updates
- ⚠️ Scripts are inline (harder to maintain)

**Gap:** Scripts are inline in `package.json`, making them harder to maintain and debug.

**Recommendation:** Extract version update scripts to separate files for better maintainability.

---

## Integration Opportunities

### High Priority

#### 1. Auto-Version Update During Development

**Problem:** Version timestamps don't update when making code changes during development - only on server start.

**Solution:** Create a Next.js webpack plugin or file watcher that updates `VERSION.json` on file changes.

**Implementation Options:**

**Option A: Next.js Webpack Plugin (Recommended)**
```javascript
// frontend/plugins/update-version-plugin.js
class UpdateVersionPlugin {
  apply(compiler) {
    compiler.hooks.watchRun.tap('UpdateVersionPlugin', () => {
      // Update VERSION.json on file changes
      this.updateVersion();
    });
  }
  
  updateVersion() {
    // Same logic as predev script
    // Throttle updates (e.g., max once per 2 seconds)
  }
}
```

**Option B: File Watcher Script**
```javascript
// frontend/scripts/watch-version.js
import chokidar from 'chokidar';
import { updateVersion } from './update-build-time.js';

const watcher = chokidar.watch('src/**/*.{ts,tsx,js,jsx}', {
  ignored: /node_modules/,
  persistent: true
});

let lastUpdate = 0;
const THROTTLE_MS = 2000;

watcher.on('change', () => {
  const now = Date.now();
  if (now - lastUpdate > THROTTLE_MS) {
    updateVersion();
    lastUpdate = now;
  }
});
```

**Option C: Use Next.js API Route Hook**
- Less reliable, but simpler
- Update version when API routes are accessed

**Recommendation:** Option A (Webpack Plugin) - most reliable and integrated with Next.js build system.

#### 2. Extract Version Update Scripts

**Problem:** Inline scripts in `package.json` are hard to maintain and debug.

**Solution:** Extract to separate script files.

**Implementation:**
```javascript
// frontend/scripts/update-build-time.js
const fs = require('fs');
const path = require('path');

function updateVersion() {
  const now = new Date();
  const thaiTime = new Date(now.getTime() + 7 * 60 * 60 * 1000);
  const buildDate = thaiTime.toISOString().slice(0, 19).replace('T', ' ');
  const buildTime = thaiTime.toISOString().slice(0, 19) + '+07:00';

  const version = {
    version: require('../package.json').version,
    buildDate,
    buildTime,
  };

  const files = [
    path.join(__dirname, '../VERSION.json'),
    path.join(__dirname, '../public/VERSION.json'),
  ];

  files.forEach((filePath) => {
    try {
      let content = fs.readFileSync(filePath, 'utf8');
      content = content.replace(/^\uFEFF/, '');
      const data = JSON.parse(content);
      data.buildDate = buildDate;
      data.buildTime = buildTime;
      fs.writeFileSync(filePath, JSON.stringify(data, null, 2) + '\n', 'utf8');
    } catch (error) {
      console.error(`Error updating ${filePath}:`, error.message);
    }
  });
}

if (require.main === module) {
  updateVersion();
}

module.exports = { updateVersion };
```

**Update `package.json`:**
```json
{
  "scripts": {
    "predev": "node scripts/update-build-time.js",
    "prebuild": "node scripts/update-build-time.js"
  }
}
```

### Medium Priority

#### 3. Optimize Air Configuration

**Current `.air.toml` is functional, but could be enhanced:**

```toml
# Add to backend/.air.toml
[build]
  # Reduce delay for faster rebuilds (if system can handle it)
  delay = 500  # Current: 1000
  
  # Add backups directory to exclusions
  exclude_dir = ["assets", "tmp", "vendor", "testdata", "tests", "backups"]
  
  # Enable rerun for better reliability
  rerun = true
  rerun_delay = 500
```

#### 4. Backend Pre-Build Script

**From PROJECT_SETUP_EXPORT.md:** Go script to update backend version before build.

**Current Project:** Backend version is updated manually or via scripts.

**Recommendation:** Create `backend/scripts/update-build-time.go` similar to PROJECT_SETUP_EXPORT.md pattern, but integrate with Air build process.

### Low Priority

#### 5. Documentation Improvements

**Enhance existing docs with:**
- Clear explanation of auto-build process
- Troubleshooting guide for file watching issues
- Performance optimization tips

---

## Implementation Plan

### Phase 1: Extract Version Scripts (Quick Win)

1. Create `frontend/scripts/update-build-time.js`
2. Update `package.json` to use extracted script
3. Test that version updates still work

**Estimated Time:** 15 minutes

### Phase 2: Auto-Version Update Plugin (High Value)

1. Create Next.js webpack plugin for auto-version updates
2. Integrate with Next.js config
3. Add throttling to prevent excessive updates
4. Test during development

**Estimated Time:** 1-2 hours

### Phase 3: Optimize Air Configuration (Low Risk)

1. Review and optimize `.air.toml`
2. Test rebuild performance
3. Document changes

**Estimated Time:** 30 minutes

### Phase 4: Backend Version Update Script (Optional)

1. Create `backend/scripts/update-build-time.go`
2. Integrate with Air build process (if possible)
3. Document usage

**Estimated Time:** 1 hour

---

## Key Differences: PROJECT_SETUP_EXPORT.md vs Current Project

| Aspect | PROJECT_SETUP_EXPORT.md | Current Project |
|--------|-------------------------|-----------------|
| **Frontend Framework** | Vue.js + Vite | Next.js |
| **Hot Reload** | Vite HMR | Next.js HMR |
| **Version Update** | Vite plugin (on file change) | Pre-build scripts only |
| **Backend Hot Reload** | Air (same) | Air (same) |
| **Script Organization** | Separate script files | Inline in package.json |
| **Auto-Build** | ✅ On file change | ✅ On file change |
| **Auto-Version** | ✅ On file change | ❌ Only on start/build |

---

## Recommendations Summary

### ✅ Already Working Well
- Backend auto-build with Air
- Frontend auto-build with Next.js HMR
- Docker volume mounts for hot reload
- Polling enabled for Windows compatibility

### 🔧 Should Improve
1. **Extract version update scripts** to separate files (maintainability)
2. **Add auto-version update plugin** for Next.js (update timestamps on file changes)
3. **Optimize Air configuration** (faster rebuilds, better exclusions)

### 📝 Nice to Have
- Backend version update script integration
- Enhanced documentation
- Performance monitoring

---

## Conclusion

**Current Status:** ✅ Auto-build is **already working** for both frontend and backend. Code changes automatically trigger rebuilds without explicit rebuild commands.

**Main Gap:** Version timestamps only update on server start/build, not on every file change during development.

**Priority Actions:**
1. Extract version scripts for maintainability
2. Create Next.js plugin for auto-version updates during development
3. Optimize Air configuration for better performance

**Impact:** These improvements will make version tracking more accurate during development and improve code maintainability.

---

## Version History

- **v1.0.0** (2026-01-08): Initial analysis created

---
title: Auto-Version Update Implementation
description: Implementation guide for automatic version timestamp updates during development
version: 1.0.0
lastUpdated: 2026-01-08 15:42:59
---

# Auto-Version Update Implementation

This document describes the implementation of automatic version timestamp updates during development, ensuring that `VERSION.json` files reflect the latest code changes, not just when the dev server was started.

---

## Overview

**Problem:** Version timestamps (`VERSION.json`) only updated when:
- Starting dev server (`predev` script)
- Building (`prebuild` script)

**Solution:** Implemented a Next.js webpack plugin that automatically updates version timestamps when source files change during development.

---

## Implementation Components

### 1. Extracted Version Update Script

**File:** `frontend/scripts/update-build-time.js`

**Purpose:** Reusable script to update `VERSION.json` files with current build timestamp.

**Features:**
- Reads version from `package.json`
- Calculates current time in Thailand timezone (UTC+7)
- Updates both `VERSION.json` and `public/VERSION.json`
- Can be used as a module or run directly
- Handles UTF-8 BOM removal
- Error handling with optional silent mode

**Usage:**
```bash
# Direct execution
node scripts/update-build-time.js

# As a module
const { updateVersion } = require('./scripts/update-build-time.js');
updateVersion();
```

### 2. Webpack Plugin for Auto-Updates

**File:** `frontend/plugins/update-version-plugin.js`

**Purpose:** Automatically updates version files when source code changes during development.

**Features:**
- Hooks into Next.js webpack compilation process
- Updates version on file changes (via `watchRun` hook)
- Throttles updates (max once per 2 seconds) to prevent excessive writes
- Silent operation (doesn't spam console)
- Only runs in development mode
- Client-side only (doesn't interfere with server-side compilation)

**How it works:**
1. Hooks into webpack's `watchRun` event (file changes detected)
2. Throttles updates to prevent excessive file writes
3. Calls `update-build-time.js` to update version files
4. Runs silently in the background

### 3. Integration in Next.js Config

**File:** `frontend/next.config.js`

**Changes:**
- Added webpack configuration
- Integrated `UpdateVersionPlugin` in development mode only
- Configured for client-side builds only

**Configuration:**
```javascript
webpack: (config, { dev, isServer }) => {
  if (dev && !isServer) {
    const UpdateVersionPlugin = require('./plugins/update-version-plugin');
    config.plugins.push(new UpdateVersionPlugin({
      throttleMs: 2000, // Update at most once per 2 seconds
      silent: true, // Silent operation
    }));
  }
  return config;
}
```

### 4. Updated Package.json Scripts

**File:** `frontend/package.json`

**Changes:**
- Replaced inline scripts with calls to extracted script
- `predev`: `node scripts/update-build-time.js`
- `prebuild`: `node scripts/update-build-time.js`

**Benefits:**
- Cleaner `package.json`
- Easier to maintain and debug
- Reusable script

---

## How It Works

### Development Workflow

1. **Start Dev Server:**
   ```bash
   npm run dev
   ```
   - `predev` script runs → Updates version files
   - Next.js dev server starts
   - Webpack plugin is active

2. **Make Code Changes:**
   - Edit files in `src/`
   - Webpack detects changes
   - `watchRun` hook fires
   - Plugin updates version files (throttled)

3. **Result:**
   - Version timestamps reflect latest changes
   - No manual intervention needed
   - Automatic and silent

### Update Flow

```
File Change
    ↓
Webpack watchRun Hook
    ↓
UpdateVersionPlugin.handleFileChange()
    ↓
Check Throttle (2 seconds)
    ↓
updateVersion() (silent)
    ↓
VERSION.json Updated
```

---

## Version File Format

**Format:**
```json
{
  "version": "1.0.0",
  "buildDate": "YYYY-MM-DD HH:MM:SS",
  "buildTime": "YYYY-MM-DDTHH:MM:SS+07:00"
}
```

**Example:**
```json
{
  "version": "1.0.0",
  "buildDate": "2026-01-08 15:42:59",
  "buildTime": "2026-01-08T15:42:59+07:00"
}
```

**Files Updated:**
- `frontend/VERSION.json`
- `frontend/public/VERSION.json` (must be kept in sync)

---

## Configuration Options

### Plugin Options

```javascript
new UpdateVersionPlugin({
  throttleMs: 2000,  // Minimum milliseconds between updates (default: 2000)
  silent: true,       // Suppress console output (default: true)
})
```

### Script Options

```javascript
updateVersion(silent = false)  // If true, suppresses console output
```

---

## Testing

### Manual Test

1. **Start dev server:**
   ```bash
   cd frontend
   npm run dev
   ```

2. **Check initial version:**
   ```bash
   cat VERSION.json
   ```

3. **Make a code change:**
   - Edit any file in `src/`
   - Save the file

4. **Wait 2 seconds, then check version:**
   ```bash
   cat VERSION.json
   ```
   - `buildDate` and `buildTime` should be updated

### Verify Plugin is Active

Check Next.js dev server output - you should see webpack compilation messages when files change. The plugin runs silently, so no additional output is expected.

---

## Troubleshooting

### Version Not Updating

**Problem:** Version files not updating on file changes.

**Solutions:**
1. **Check plugin is loaded:**
   - Verify `next.config.js` has webpack configuration
   - Check that `dev` mode is active (`NODE_ENV=development`)

2. **Check file permissions:**
   - Ensure `VERSION.json` files are writable
   - Check for file system errors

3. **Check throttling:**
   - Plugin throttles updates (2 seconds)
   - Make changes and wait >2 seconds before checking

4. **Check webpack hooks:**
   - Verify webpack is detecting file changes
   - Check Next.js dev server logs

### Script Errors

**Problem:** `update-build-time.js` fails to run.

**Solutions:**
1. **Check Node.js version:** Requires Node.js 14+
2. **Check file paths:** Ensure script is run from project root
3. **Check package.json:** Verify version field exists
4. **Check file permissions:** Ensure write permissions

### Plugin Not Loading

**Problem:** Webpack plugin not being applied.

**Solutions:**
1. **Check Next.js version:** Requires Next.js 12+
2. **Check webpack config:** Verify webpack function is exported correctly
3. **Check dev mode:** Plugin only runs in development
4. **Check isServer:** Plugin only runs for client-side builds

---

## Performance Considerations

### Throttling

- Updates are throttled to max once per 2 seconds
- Prevents excessive file writes during rapid changes
- Configurable via plugin options

### File System Impact

- Minimal: Only 2 files updated
- Throttled: Max once per 2 seconds
- Silent: No console spam

### Build Performance

- No impact on build time
- Only runs in development mode
- Client-side only (doesn't affect server builds)

---

## Related Documentation

- [Versioning Rules](../docs/versioning-rules.md) - Version file format and rules
- [Version Build Process](../docs/version-build-process.md) - Build-time version updates
- [Auto-Build Integration Analysis](../docs/auto-build-integration-analysis.md) - Analysis of auto-build patterns

---

## Version History

- **v1.0.0** (2026-01-08): Initial implementation
  - Extracted version update script
  - Created webpack plugin for auto-updates
  - Integrated into Next.js config
  - Updated package.json scripts

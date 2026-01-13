# Frontend Scripts

This directory contains utility scripts for the frontend application.

## Scripts

### `update-build-time.js`

Updates `VERSION.json` files with current build timestamp (Thailand timezone UTC+7).

**Usage:**

```bash
# Direct execution
node scripts/update-build-time.js

# As a module
const { updateVersion } = require('./scripts/update-build-time.js');
updateVersion();
```

**What it does:**
- Reads version from `package.json`
- Calculates current time in Thailand timezone (UTC+7)
- Updates `VERSION.json` and `public/VERSION.json` with:
  - `version`: From package.json
  - `buildDate`: Current date/time (format: `YYYY-MM-DD HH:mm:ss`)
  - `buildTime`: Current timestamp (format: `YYYY-MM-DDTHH:mm:ss+07:00`)

**When it runs:**
- Automatically before `npm run dev` (via `predev` script)
- Automatically before `npm run build` (via `prebuild` script)
- Automatically during development when files change (via webpack plugin)

**Files updated:**
- `frontend/VERSION.json`
- `frontend/public/VERSION.json`

---

## Related Files

- `plugins/update-version-plugin.js` - Webpack plugin for auto-updating version during development
- `next.config.js` - Next.js configuration that integrates the webpack plugin

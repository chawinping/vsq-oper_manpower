# Frontend Plugins

This directory contains custom webpack plugins for Next.js.

## Plugins

### `update-version-plugin.js`

**Auto-Update Version Plugin**

Automatically updates `VERSION.json` files when source files change during development. This ensures build timestamps reflect the latest code changes, not just when the dev server was started.

**Features:**
- ✅ Updates version timestamps on file changes
- ✅ Throttles updates (max once per 2 seconds) to prevent excessive file writes
- ✅ Silent operation (doesn't spam console)
- ✅ Only runs in development mode
- ✅ Client-side only (doesn't run during server-side compilation)

**How it works:**

1. Hooks into Next.js webpack compilation process
2. Detects when source files change (via `watchRun` hook)
3. Throttles updates to prevent excessive writes
4. Calls `update-build-time.js` to update version files
5. Runs silently in the background

**Configuration:**

The plugin is configured in `next.config.js`:

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

**Options:**
- `throttleMs` (default: 2000) - Minimum milliseconds between updates
- `silent` (default: true) - Suppress console output

**When it runs:**
- During development mode (`npm run dev`)
- When source files change
- After successful webpack compilation
- Only for client-side builds (not server-side)

**Note:** This plugin complements the `predev` script. The `predev` script updates version when the dev server starts, and this plugin updates it continuously as you make changes.

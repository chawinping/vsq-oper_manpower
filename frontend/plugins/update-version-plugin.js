/**
 * Next.js Webpack Plugin: Auto-Update Version
 * 
 * Automatically updates VERSION.json files when source files change during development.
 * This ensures build timestamps reflect the latest code changes, not just when
 * the dev server was started.
 * 
 * Features:
 * - Updates version timestamps on file changes
 * - Throttles updates to prevent excessive file writes (max once per 2 seconds)
 * - Silent operation (doesn't spam console)
 * - Works with Next.js webpack watch mode
 */

const { updateVersion } = require('../scripts/update-build-time');

class UpdateVersionPlugin {
  constructor(options = {}) {
    this.options = {
      throttleMs: options.throttleMs || 2000, // Update at most once per 2 seconds
      silent: options.silent !== false, // Silent by default
      ...options,
    };
    this.lastUpdateTime = 0;
  }

  apply(compiler) {
    // Hook into webpack's watch mode
    compiler.hooks.watchRun.tap('UpdateVersionPlugin', (compilation) => {
      this.handleFileChange();
    });

    // Also hook into compilation start (for initial build and rebuilds)
    compiler.hooks.compilation.tap('UpdateVersionPlugin', () => {
      this.handleFileChange();
    });

    // Hook into done event (after successful compilation)
    compiler.hooks.done.tap('UpdateVersionPlugin', () => {
      // Update version after successful compilation
      this.handleFileChange();
    });
  }

  handleFileChange() {
    const now = Date.now();
    const timeSinceLastUpdate = now - this.lastUpdateTime;

    // Throttle updates to prevent excessive file writes
    if (timeSinceLastUpdate < this.options.throttleMs) {
      return; // Skip if updated recently
    }

    try {
      // Update version files silently
      updateVersion(true); // silent = true
      this.lastUpdateTime = now;
    } catch (error) {
      // Don't interrupt build process if version update fails
      if (!this.options.silent) {
        console.warn('[UpdateVersionPlugin] Failed to update version:', error.message);
      }
    }
  }
}

module.exports = UpdateVersionPlugin;

/**
 * Update Build Time Script
 * Updates VERSION.json files with current build timestamp (Thailand timezone UTC+7)
 * 
 * Usage:
 *   node scripts/update-build-time.js
 * 
 * Or import as module:
 *   const { updateVersion } = require('./scripts/update-build-time.js');
 *   updateVersion();
 */

const fs = require('fs');
const path = require('path');

/**
 * Updates VERSION.json files with current build timestamp
 * @param {boolean} silent - If true, suppresses console output
 */
function updateVersion(silent = false) {
  try {
    // Get current time in Thailand timezone (UTC+7)
    const now = new Date();
    const thaiTime = new Date(now.getTime() + 7 * 60 * 60 * 1000);
    const buildDate = thaiTime.toISOString().slice(0, 19).replace('T', ' ');
    const buildTime = thaiTime.toISOString().slice(0, 19) + '+07:00';

    // Read current version from package.json
    const packageJsonPath = path.join(__dirname, '../package.json');
    const packageJson = JSON.parse(fs.readFileSync(packageJsonPath, 'utf8'));
    const version = packageJson.version || '1.0.0';

    const versionData = {
      version,
      buildDate,
      buildTime,
    };

    // Files to update
    const files = [
      path.join(__dirname, '../VERSION.json'),
      path.join(__dirname, '../public/VERSION.json'),
    ];

    files.forEach((filePath) => {
      try {
        // Read existing file (if exists)
        let existingData = {};
        if (fs.existsSync(filePath)) {
          let content = fs.readFileSync(filePath, 'utf8');
          // Remove UTF-8 BOM if present
          content = content.replace(/^\uFEFF/, '');
          existingData = JSON.parse(content);
        }

        // Merge with new timestamp data (preserve version if not in package.json)
        const data = {
          version: versionData.version || existingData.version || '1.0.0',
          buildDate: versionData.buildDate,
          buildTime: versionData.buildTime,
        };

        // Write updated data
        fs.writeFileSync(filePath, JSON.stringify(data, null, 2) + '\n', 'utf8');
        
        if (!silent) {
          console.log(`✅ Updated ${path.relative(process.cwd(), filePath)}`);
        }
      } catch (error) {
        console.error(`❌ Error updating ${filePath}:`, error.message);
        if (!silent) {
          throw error;
        }
      }
    });

    return true;
  } catch (error) {
    console.error('❌ Error updating version files:', error.message);
    if (!silent) {
      throw error;
    }
    return false;
  }
}

// If run directly (not imported), execute update
if (require.main === module) {
  updateVersion();
}

module.exports = { updateVersion };

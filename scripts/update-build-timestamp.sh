#!/bin/bash
# VSQ Operations Manpower - Build Timestamp Update Script (Bash version)
# This script updates build timestamps without changing version numbers
# It's designed to be called automatically during the build process

COMPONENT=${1:-all}

# Get current date/time in Thailand timezone (UTC+7)
BUILD_DATE=$(TZ='Asia/Bangkok' date '+%Y-%m-%d %H:%M:%S')
BUILD_TIME=$(TZ='Asia/Bangkok' date '+%Y-%m-%dT%H:%M:%S+07:00')

# Get project root directory (assuming script is in scripts/ directory)
PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$PROJECT_ROOT"

update_timestamp() {
    local file_path=$1
    
    if [ ! -f "$file_path" ]; then
        echo "Warning: Version file not found: $file_path (skipping)"
        return 1
    fi
    
    # Use node or python to update JSON (whichever is available)
    if command -v node &> /dev/null; then
        node -e "
            const fs = require('fs');
            const file = '$file_path';
            const data = JSON.parse(fs.readFileSync(file, 'utf8'));
            data.buildDate = '$BUILD_DATE';
            data.buildTime = '$BUILD_TIME';
            fs.writeFileSync(file, JSON.stringify(data, null, 2) + '\n', 'utf8');
        "
    elif command -v python3 &> /dev/null; then
        python3 << EOF
import json
import sys

file_path = '$file_path'
with open(file_path, 'r') as f:
    data = json.load(f)

data['buildDate'] = '$BUILD_DATE'
data['buildTime'] = '$BUILD_TIME'

with open(file_path, 'w') as f:
    json.dump(data, f, indent=2)
    f.write('\n')
EOF
    else
        echo "Error: Neither node nor python3 is available to update JSON"
        return 1
    fi
    
    echo "Updated build timestamp: $file_path"
    echo "  Build Date: $BUILD_DATE"
    return 0
}

echo ""
echo "Updating build timestamps..."
echo "Build Date: $BUILD_DATE"
echo ""

case $COMPONENT in
    frontend)
        update_timestamp "frontend/VERSION.json"
        update_timestamp "frontend/public/VERSION.json"
        ;;
    backend)
        update_timestamp "backend/VERSION.json"
        ;;
    database)
        update_timestamp "backend/DATABASE_VERSION.json"
        ;;
    all)
        update_timestamp "frontend/VERSION.json"
        update_timestamp "frontend/public/VERSION.json"
        update_timestamp "backend/VERSION.json"
        update_timestamp "backend/DATABASE_VERSION.json"
        ;;
    *)
        echo "Usage: $0 [frontend|backend|database|all]"
        exit 1
        ;;
esac

echo ""
echo "Build timestamp update completed!"




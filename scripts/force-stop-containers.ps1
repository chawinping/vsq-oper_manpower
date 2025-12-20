# Force stop and remove VSQ Manpower containers
# This script forcefully stops containers that are stuck due to restart policies

Write-Host "Stopping VSQ Manpower containers..." -ForegroundColor Yellow

# Update restart policy to 'no' to prevent auto-restart
$containers = @("vsq-manpower-frontend-dev", "vsq-manpower-backend-dev", "vsq-manpower-db")
foreach ($container in $containers) {
    Write-Host "Updating restart policy for $container..." -ForegroundColor Cyan
    docker update --restart=no $container 2>&1 | Out-Null
}

# Wait a moment for the policy to take effect
Start-Sleep -Seconds 2

# Force kill all containers
Write-Host "Force killing containers..." -ForegroundColor Yellow
docker-compose --profile fullstack --profile fullstack-dev --profile dev kill 2>&1 | Out-Null

# Wait a moment
Start-Sleep -Seconds 2

# Force remove containers
Write-Host "Force removing containers..." -ForegroundColor Yellow
foreach ($container in $containers) {
    docker rm -f $container 2>&1 | Out-Null
}

# Run docker-compose down to clean up
Write-Host "Running docker-compose down..." -ForegroundColor Yellow
docker-compose --profile fullstack --profile fullstack-dev --profile dev down --remove-orphans 2>&1 | Out-Null

Write-Host "Done! Containers should be stopped and removed." -ForegroundColor Green



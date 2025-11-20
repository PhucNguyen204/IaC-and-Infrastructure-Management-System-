# Test Docker Service APIs
# Usage: .\test_docker_service.ps1

$ErrorActionPreference = "Stop"

$BASE_URL = "http://localhost:8083/api/v1"
$TOKEN = ""

Write-Host "=== Docker Service API Test ===" -ForegroundColor Cyan

# Get token if not set
if (-not $TOKEN) {
    Write-Host "`n[AUTH] Getting authentication token..." -ForegroundColor Yellow
    $authResponse = Invoke-RestMethod -Uri "http://localhost:8082/auth/login" `
        -Method POST `
        -Headers @{"Content-Type"="application/json"} `
        -Body '{"username":"admin","password":"password123"}'
    $TOKEN = $authResponse.data.access_token
    Write-Host "Token obtained: $($TOKEN.Substring(0, 20))..." -ForegroundColor Green
}

$HEADERS = @{
    "Content-Type" = "application/json"
    "Authorization" = "Bearer $TOKEN"
}

# UC1: Create Docker Service (Simple Nginx)
Write-Host "`n[UC1] Creating Docker Service (nginx:alpine)..." -ForegroundColor Yellow
$createRequest = @{
    name = "test-nginx-service"
    image = "nginx"
    image_tag = "alpine"
    service_type = "web"
    env_vars = @(
        @{
            key = "NGINX_HOST"
            value = "localhost"
            is_secret = $false
        }
    )
    ports = @(
        @{
            container_port = 80
            host_port = 8090
            protocol = "tcp"
        }
    )
    networks = @(
        @{
            network_id = "iaas_iaas-network"
            alias = "nginx-web"
        }
    )
    health_check = @{
        type = "HTTP"
        http_path = "/"
        port = 80
        interval_seconds = 30
        timeout_seconds = 10
        retries = 3
    }
    restart_policy = "always"
    plan = "small"
} | ConvertTo-Json -Depth 10

try {
    $createResponse = Invoke-RestMethod -Uri "$BASE_URL/docker" `
        -Method POST `
        -Headers $HEADERS `
        -Body $createRequest
    
    $SERVICE_ID = $createResponse.data.id
    Write-Host "Docker service created successfully!" -ForegroundColor Green
    Write-Host "Service ID: $SERVICE_ID" -ForegroundColor Cyan
    Write-Host "Container ID: $($createResponse.data.container_id)" -ForegroundColor Cyan
    Write-Host "Internal Endpoint: $($createResponse.data.internal_endpoint)" -ForegroundColor Cyan
    Write-Host "Status: $($createResponse.data.status)" -ForegroundColor Cyan
} catch {
    Write-Host "Failed to create Docker service: $_" -ForegroundColor Red
    exit 1
}

Start-Sleep -Seconds 2

# Get Docker Service Info
Write-Host "`n[GET] Retrieving Docker service info..." -ForegroundColor Yellow
try {
    $getResponse = Invoke-RestMethod -Uri "$BASE_URL/docker/$SERVICE_ID" `
        -Method GET `
        -Headers $HEADERS
    
    Write-Host "Service info retrieved!" -ForegroundColor Green
    Write-Host "Name: $($getResponse.data.infrastructure_name)" -ForegroundColor Cyan
    Write-Host "Image: $($getResponse.data.image):$($getResponse.data.image_tag)" -ForegroundColor Cyan
    Write-Host "Status: $($getResponse.data.status)" -ForegroundColor Cyan
    Write-Host "IP Address: $($getResponse.data.ip_address)" -ForegroundColor Cyan
    Write-Host "Environment Variables:" -ForegroundColor Cyan
    foreach ($env in $getResponse.data.env_vars) {
        if ($env.is_secret) {
            Write-Host "  $($env.key): [REDACTED]" -ForegroundColor Gray
        } else {
            Write-Host "  $($env.key): $($env.value)" -ForegroundColor Gray
        }
    }
} catch {
    Write-Host "Failed to get Docker service: $_" -ForegroundColor Red
}

# UC8: Get Service Logs
Write-Host "`n[UC8] Getting service logs..." -ForegroundColor Yellow
try {
    $logsResponse = Invoke-RestMethod -Uri "$BASE_URL/docker/$SERVICE_ID/logs?tail=10" `
        -Method GET `
        -Headers $HEADERS
    
    Write-Host "Service logs retrieved (last 10 lines):" -ForegroundColor Green
    foreach ($log in $logsResponse.data.logs) {
        Write-Host "  $log" -ForegroundColor Gray
    }
} catch {
    Write-Host "Failed to get service logs: $_" -ForegroundColor Red
}

# Stop Docker Service
Write-Host "`n[STOP] Stopping Docker service..." -ForegroundColor Yellow
try {
    Invoke-RestMethod -Uri "$BASE_URL/docker/$SERVICE_ID/stop" `
        -Method POST `
        -Headers $HEADERS | Out-Null
    Write-Host "Docker service stopped successfully!" -ForegroundColor Green
} catch {
    Write-Host "Failed to stop Docker service: $_" -ForegroundColor Red
}

Start-Sleep -Seconds 1

# Start Docker Service
Write-Host "`n[START] Starting Docker service..." -ForegroundColor Yellow
try {
    Invoke-RestMethod -Uri "$BASE_URL/docker/$SERVICE_ID/start" `
        -Method POST `
        -Headers $HEADERS | Out-Null
    Write-Host "Docker service started successfully!" -ForegroundColor Green
} catch {
    Write-Host "Failed to start Docker service: $_" -ForegroundColor Red
}

Start-Sleep -Seconds 1

# Restart Docker Service
Write-Host "`n[RESTART] Restarting Docker service..." -ForegroundColor Yellow
try {
    Invoke-RestMethod -Uri "$BASE_URL/docker/$SERVICE_ID/restart" `
        -Method POST `
        -Headers $HEADERS | Out-Null
    Write-Host "Docker service restarted successfully!" -ForegroundColor Green
} catch {
    Write-Host "Failed to restart Docker service: $_" -ForegroundColor Red
}

Start-Sleep -Seconds 2

# UC2: Update Environment Variables (including secrets)
Write-Host "`n[UC2] Updating environment variables..." -ForegroundColor Yellow
$updateEnvRequest = @{
    env_vars = @(
        @{
            key = "NGINX_HOST"
            value = "updated-host"
            is_secret = $false
        },
        @{
            key = "DB_PASSWORD"
            value = "secret123"
            is_secret = $true
        }
    )
} | ConvertTo-Json -Depth 10

try {
    Invoke-RestMethod -Uri "$BASE_URL/docker/$SERVICE_ID/env" `
        -Method PUT `
        -Headers $HEADERS `
        -Body $updateEnvRequest | Out-Null
    Write-Host "Environment variables updated successfully!" -ForegroundColor Green
    Write-Host "Note: Container was recreated with new env vars" -ForegroundColor Yellow
} catch {
    Write-Host "Failed to update environment variables: $_" -ForegroundColor Red
}

Start-Sleep -Seconds 2

# Verify updated env vars
Write-Host "`n[VERIFY] Verifying updated environment variables..." -ForegroundColor Yellow
try {
    $getResponse = Invoke-RestMethod -Uri "$BASE_URL/docker/$SERVICE_ID" `
        -Method GET `
        -Headers $HEADERS
    
    Write-Host "Environment Variables after update:" -ForegroundColor Green
    foreach ($env in $getResponse.data.env_vars) {
        if ($env.is_secret) {
            Write-Host "  $($env.key): [REDACTED]" -ForegroundColor Gray
        } else {
            Write-Host "  $($env.key): $($env.value)" -ForegroundColor Gray
        }
    }
} catch {
    Write-Host "Failed to verify env vars: $_" -ForegroundColor Red
}

# UC3: Create Docker Service with PostgreSQL dependency
Write-Host "`n[UC3] Creating Docker Service with PostgreSQL connection..." -ForegroundColor Yellow

# First, create a PostgreSQL instance
Write-Host "Creating PostgreSQL instance..." -ForegroundColor Gray
$pgRequest = @{
    name = "test-postgres-for-docker"
    plan = "small"
} | ConvertTo-Json

try {
    $pgResponse = Invoke-RestMethod -Uri "$BASE_URL/postgres" `
        -Method POST `
        -Headers $HEADERS `
        -Body $pgRequest
    
    $PG_ID = $pgResponse.data.id
    Write-Host "PostgreSQL instance created: $PG_ID" -ForegroundColor Green
} catch {
    Write-Host "Failed to create PostgreSQL instance: $_" -ForegroundColor Red
}

Start-Sleep -Seconds 3

# Get PostgreSQL connection info
$pgInfo = Invoke-RestMethod -Uri "$BASE_URL/postgres/$PG_ID" `
    -Method GET `
    -Headers $HEADERS

Write-Host "Creating Docker service with PostgreSQL connection..." -ForegroundColor Gray
$appRequest = @{
    name = "test-app-with-db"
    image = "postgres"
    image_tag = "15-alpine"
    service_type = "worker"
    command = "psql"
    args = @("-h", $pgInfo.data.ip_address, "-U", "postgres", "-c", "SELECT version();")
    env_vars = @(
        @{
            key = "POSTGRES_HOST"
            value = $pgInfo.data.ip_address
            is_secret = $false
        },
        @{
            key = "POSTGRES_PASSWORD"
            value = "postgres"
            is_secret = $true
        }
    )
    networks = @(
        @{
            network_id = "iaas_iaas-network"
            alias = "app-worker"
        }
    )
    dependencies = @($PG_ID)
    restart_policy = "on-failure"
    plan = "small"
} | ConvertTo-Json -Depth 10

try {
    $appResponse = Invoke-RestMethod -Uri "$BASE_URL/docker" `
        -Method POST `
        -Headers $HEADERS `
        -Body $appRequest
    
    $APP_SERVICE_ID = $appResponse.data.id
    Write-Host "Docker service with PostgreSQL connection created!" -ForegroundColor Green
    Write-Host "Service ID: $APP_SERVICE_ID" -ForegroundColor Cyan
} catch {
    Write-Host "Failed to create Docker service: $_" -ForegroundColor Red
}

# Cleanup
Write-Host "`n[CLEANUP] Cleaning up test resources..." -ForegroundColor Yellow

Write-Host "Deleting first Docker service..." -ForegroundColor Gray
try {
    Invoke-RestMethod -Uri "$BASE_URL/docker/$SERVICE_ID" `
        -Method DELETE `
        -Headers $HEADERS | Out-Null
    Write-Host "First Docker service deleted!" -ForegroundColor Green
} catch {
    Write-Host "Failed to delete first Docker service: $_" -ForegroundColor Red
}

if ($APP_SERVICE_ID) {
    Write-Host "Deleting second Docker service..." -ForegroundColor Gray
    try {
        Invoke-RestMethod -Uri "$BASE_URL/docker/$APP_SERVICE_ID" `
            -Method DELETE `
            -Headers $HEADERS | Out-Null
        Write-Host "Second Docker service deleted!" -ForegroundColor Green
    } catch {
        Write-Host "Failed to delete second Docker service: $_" -ForegroundColor Red
    }
}

if ($PG_ID) {
    Write-Host "Deleting PostgreSQL instance..." -ForegroundColor Gray
    try {
        Invoke-RestMethod -Uri "$BASE_URL/postgres/$PG_ID" `
            -Method DELETE `
            -Headers $HEADERS | Out-Null
        Write-Host "PostgreSQL instance deleted!" -ForegroundColor Green
    } catch {
        Write-Host "Failed to delete PostgreSQL instance: $_" -ForegroundColor Red
    }
}

Write-Host "`n=== Docker Service API Test Completed ===" -ForegroundColor Cyan

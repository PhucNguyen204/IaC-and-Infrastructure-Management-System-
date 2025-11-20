Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Testing Monitoring APIs" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

$baseUrl = "http://localhost:8084/api/v1"
$authUrl = "http://localhost:8082/api/v1"

Write-Host "Step 1: Login to get JWT token..." -ForegroundColor Yellow
$loginBody = @{
    username = "admin"
    password = "password123"
} | ConvertTo-Json

try {
    $loginResponse = Invoke-RestMethod -Uri "$authUrl/auth/login" -Method POST -Body $loginBody -ContentType "application/json"
    $token = $loginResponse.data.access_token
    Write-Host "✓ Login successful" -ForegroundColor Green
    Write-Host "Token: $($token.Substring(0, 50))..." -ForegroundColor Gray
} catch {
    Write-Host "✗ Login failed: $_" -ForegroundColor Red
    exit 1
}

$headers = @{
    "Authorization" = "Bearer $token"
    "Content-Type" = "application/json"
}

Write-Host ""
Write-Host "Step 2: Create a PostgreSQL instance for testing..." -ForegroundColor Yellow
$createPgBody = @{
    name = "test-postgres-monitoring"
    version = "15"
    database_name = "testdb"
    username = "testuser"
    password = "testpass123"
    cpu_limit = 1
    memory_limit = 512000000
    storage_size = 10000000000
} | ConvertTo-Json

try {
    $createResponse = Invoke-RestMethod -Uri "$baseUrl/postgres/single" -Method POST -Body $createPgBody -Headers $headers
    $instanceId = $createResponse.data.id
    Write-Host "✓ PostgreSQL instance created: $instanceId" -ForegroundColor Green
    
    Write-Host "Waiting for container to start..." -ForegroundColor Gray
    Start-Sleep -Seconds 5
} catch {
    Write-Host "✗ Failed to create PostgreSQL instance: $_" -ForegroundColor Red
    Write-Host "Response: $($_.Exception.Response)" -ForegroundColor Gray
    exit 1
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Testing Monitoring APIs" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

Write-Host "Test 1: GET /api/v1/monitoring/metrics/{instance_id}" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/monitoring/metrics/$instanceId" -Method GET
    Write-Host "✓ Success" -ForegroundColor Green
    Write-Host "Response:" -ForegroundColor Gray
    $response | ConvertTo-Json -Depth 5 | Write-Host
} catch {
    Write-Host "✗ Failed: $_" -ForegroundColor Red
    if ($_.Exception.Response) {
        $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
        $responseBody = $reader.ReadToEnd()
        Write-Host "Response: $responseBody" -ForegroundColor Gray
    }
}

Write-Host ""
Write-Host "Test 2: GET /api/v1/monitoring/metrics/{instance_id}/history" -ForegroundColor Yellow
try {
    $uri = "$baseUrl/monitoring/metrics/$instanceId/history?from=0`&size=10"
    $response = Invoke-RestMethod -Uri $uri -Method GET
    Write-Host "✓ Success" -ForegroundColor Green
    Write-Host "Response:" -ForegroundColor Gray
    $response | ConvertTo-Json -Depth 5 | Write-Host
} catch {
    Write-Host "✗ Failed: $_" -ForegroundColor Red
    if ($_.Exception.Response) {
        $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
        $responseBody = $reader.ReadToEnd()
        Write-Host "Response: $responseBody" -ForegroundColor Gray
    }
}

Write-Host ""
Write-Host "Test 3: GET /api/v1/monitoring/metrics/{instance_id}/aggregate" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/monitoring/metrics/$instanceId/aggregate?range=1h" -Method GET
    Write-Host "✓ Success" -ForegroundColor Green
    Write-Host "Response:" -ForegroundColor Gray
    $response | ConvertTo-Json -Depth 5 | Write-Host
} catch {
    Write-Host "✗ Failed: $_" -ForegroundColor Red
    if ($_.Exception.Response) {
        $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
        $responseBody = $reader.ReadToEnd()
        Write-Host "Response: $responseBody" -ForegroundColor Gray
    }
}

Write-Host ""
Write-Host "Test 4: GET /api/v1/monitoring/logs/{instance_id}" -ForegroundColor Yellow
try {
    $uri = "$baseUrl/monitoring/logs/$instanceId?from=0`&size=10"
    $response = Invoke-RestMethod -Uri $uri -Method GET
    Write-Host "✓ Success" -ForegroundColor Green
    Write-Host "Response:" -ForegroundColor Gray
    $response | ConvertTo-Json -Depth 5 | Write-Host
} catch {
    Write-Host "✗ Failed: $_" -ForegroundColor Red
    if ($_.Exception.Response) {
        $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
        $responseBody = $reader.ReadToEnd()
        Write-Host "Response: $responseBody" -ForegroundColor Gray
    }
}

Write-Host ""
Write-Host "Test 5: GET /api/v1/monitoring/health/{instance_id}" -ForegroundColor Yellow
try {
    $pgInfo = Invoke-RestMethod -Uri "$baseUrl/postgres/single/$instanceId" -Method GET -Headers $headers
    $containerId = $pgInfo.data.container_id
    Write-Host "Container ID: $containerId" -ForegroundColor Gray
    $response = Invoke-RestMethod -Uri "$baseUrl/monitoring/health/$containerId" -Method GET
    Write-Host "✓ Success" -ForegroundColor Green
    Write-Host "Response:" -ForegroundColor Gray
    $response | ConvertTo-Json -Depth 5 | Write-Host
} catch {
    Write-Host "✗ Failed: $_" -ForegroundColor Red
    if ($_.Exception.Response) {
        $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
        $responseBody = $reader.ReadToEnd()
        Write-Host "Response: $responseBody" -ForegroundColor Gray
    }
}

Write-Host ""
Write-Host "Test 6: GET /api/v1/monitoring/infrastructure" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/monitoring/infrastructure" -Method GET
    Write-Host "✓ Success" -ForegroundColor Green
    Write-Host "Response:" -ForegroundColor Gray
    $response | ConvertTo-Json -Depth 5 | Write-Host
} catch {
    Write-Host "✗ Failed: $_" -ForegroundColor Red
    if ($_.Exception.Response) {
        $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
        $responseBody = $reader.ReadToEnd()
        Write-Host "Response: $responseBody" -ForegroundColor Gray
    }
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Waiting for health check to collect metrics..." -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Waiting 35 seconds for health check service to collect metrics..." -ForegroundColor Yellow
Start-Sleep -Seconds 35

Write-Host ""
Write-Host "Test 7: GET /api/v1/monitoring/metrics/{instance_id} (after health check)" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/monitoring/metrics/$containerId" -Method GET
    Write-Host "✓ Success" -ForegroundColor Green
    Write-Host "Response:" -ForegroundColor Gray
    $response | ConvertTo-Json -Depth 5 | Write-Host
} catch {
    Write-Host "✗ Failed: $_" -ForegroundColor Red
    if ($_.Exception.Response) {
        $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
        $responseBody = $reader.ReadToEnd()
        Write-Host "Response: $responseBody" -ForegroundColor Gray
    }
}

Write-Host ""
Write-Host "Test 8: GET /api/v1/monitoring/logs/{instance_id} (check for Kafka events)" -ForegroundColor Yellow
try {
    $uri = "$baseUrl/monitoring/logs/$instanceId?from=0`&size=20"
    $response = Invoke-RestMethod -Uri $uri -Method GET
    Write-Host "✓ Success" -ForegroundColor Green
    Write-Host "Total logs found: $($response.data.Count)" -ForegroundColor Gray
    if ($response.data.Count -gt 0) {
        Write-Host "Sample log:" -ForegroundColor Gray
        $response.data[0] | ConvertTo-Json -Depth 3 | Write-Host
    }
} catch {
    Write-Host "✗ Failed: $_" -ForegroundColor Red
    if ($_.Exception.Response) {
        $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
        $responseBody = $reader.ReadToEnd()
        Write-Host "Response: $responseBody" -ForegroundColor Gray
    }
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "Monitoring API Tests Completed" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green


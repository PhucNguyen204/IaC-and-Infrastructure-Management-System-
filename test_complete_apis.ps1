Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Complete API Testing for IaaS System" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

$BASE_URL_AUTH = "http://localhost:8082"
$BASE_URL_PROVISIONING = "http://localhost:8083"
$BASE_URL_MONITORING = "http://localhost:8084"

# Login
Write-Host "`n1. Authentication Test" -ForegroundColor Yellow
$loginBody = @{
    username = "admin"
    password = "password123"
} | ConvertTo-Json

try {
    $loginResponse = Invoke-RestMethod -Uri "$BASE_URL_AUTH/auth/login" -Method Post -Body $loginBody -ContentType "application/json"
    $token = $loginResponse.data.access_token
    Write-Host "   [PASS] Login successful" -ForegroundColor Green
} catch {
    Write-Host "   [FAIL] Login failed: $_" -ForegroundColor Red
    exit 1
}

$headers = @{
    "Authorization" = "Bearer $token"
    "Content-Type" = "application/json"
}

# PostgreSQL Tests
Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "PostgreSQL Single Instance APIs" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

Write-Host "`n2. Create PostgreSQL Instance" -ForegroundColor Yellow
$createPgBody = @{
    name = "test-postgres"
    version = "15"
    port = 15432
    database_name = "testdb"
    username = "testuser"
    password = "testpass123"
    cpu_limit = 1000000000
    memory_limit = 536870912
    storage_size = 10737418240
} | ConvertTo-Json

try {
    $pgResponse = Invoke-RestMethod -Uri "$BASE_URL_PROVISIONING/api/v1/postgres/single" -Method Post -Body $createPgBody -Headers $headers
    $pgId = $pgResponse.data.id
    Write-Host "   [PASS] PostgreSQL created - ID: $pgId" -ForegroundColor Green
} catch {
    Write-Host "   [FAIL] $_" -ForegroundColor Red
    exit 1
}

Start-Sleep -Seconds 3

Write-Host "`n3. Get PostgreSQL Info" -ForegroundColor Yellow
try {
    $pgInfo = Invoke-RestMethod -Uri "$BASE_URL_PROVISIONING/api/v1/postgres/single/$pgId" -Method Get -Headers $headers
    Write-Host "   [PASS] Info retrieved - Status: $($pgInfo.data.status)" -ForegroundColor Green
} catch {
    Write-Host "   [FAIL] $_" -ForegroundColor Red
}

Write-Host "`n4. Get PostgreSQL Stats" -ForegroundColor Yellow
try {
    $pgStats = Invoke-RestMethod -Uri "$BASE_URL_PROVISIONING/api/v1/postgres/single/$pgId/stats" -Method Get -Headers $headers
    Write-Host "   [PASS] Stats - CPU: $($pgStats.data.cpu_usage)%, Memory: $($pgStats.data.memory_usage)" -ForegroundColor Green
} catch {
    Write-Host "   [FAIL] $_" -ForegroundColor Red
}

Write-Host "`n5. Get PostgreSQL Logs" -ForegroundColor Yellow
try {
    $pgLogs = Invoke-RestMethod -Uri "$BASE_URL_PROVISIONING/api/v1/postgres/single/$pgId/logs?tail=5" -Method Get -Headers $headers
    Write-Host "   [PASS] Logs retrieved (5 lines)" -ForegroundColor Green
} catch {
    Write-Host "   [FAIL] $_" -ForegroundColor Red
}

Write-Host "`n6. Stop PostgreSQL" -ForegroundColor Yellow
try {
    Invoke-RestMethod -Uri "$BASE_URL_PROVISIONING/api/v1/postgres/single/$pgId/stop" -Method Post -Headers $headers | Out-Null
    Write-Host "   [PASS] PostgreSQL stopped" -ForegroundColor Green
    Start-Sleep -Seconds 2
} catch {
    Write-Host "   [FAIL] $_" -ForegroundColor Red
}

Write-Host "`n7. Start PostgreSQL" -ForegroundColor Yellow
try {
    Invoke-RestMethod -Uri "$BASE_URL_PROVISIONING/api/v1/postgres/single/$pgId/start" -Method Post -Headers $headers | Out-Null
    Write-Host "   [PASS] PostgreSQL started" -ForegroundColor Green
    Start-Sleep -Seconds 3
} catch {
    Write-Host "   [FAIL] $_" -ForegroundColor Red
}

Write-Host "`n8. Restart PostgreSQL" -ForegroundColor Yellow
try {
    Invoke-RestMethod -Uri "$BASE_URL_PROVISIONING/api/v1/postgres/single/$pgId/restart" -Method Post -Headers $headers | Out-Null
    Write-Host "   [PASS] PostgreSQL restarted" -ForegroundColor Green
    Start-Sleep -Seconds 3
} catch {
    Write-Host "   [FAIL] $_" -ForegroundColor Red
}

Write-Host "`n9. Backup PostgreSQL" -ForegroundColor Yellow
try {
    $backupBody = @{
        backup_path = "/tmp/backup"
    } | ConvertTo-Json
    $backupResponse = Invoke-RestMethod -Uri "$BASE_URL_PROVISIONING/api/v1/postgres/single/$pgId/backup" -Method Post -Body $backupBody -Headers $headers
    Write-Host "   [PASS] Backup created - File: $($backupResponse.data.backup_file)" -ForegroundColor Green
} catch {
    Write-Host "   [FAIL] $_" -ForegroundColor Red
}

# Nginx Tests
Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "Nginx Instance APIs" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

Write-Host "`n10. Create Nginx Instance" -ForegroundColor Yellow
$createNginxBody = @{
    name = "test-nginx"
    port = 18080
    ssl_port = 0
    config = "server { listen 80; location / { return 200 'OK'; } }"
    cpu_limit = 500000000
    memory_limit = 268435456
} | ConvertTo-Json

try {
    $nginxResponse = Invoke-RestMethod -Uri "$BASE_URL_PROVISIONING/api/v1/nginx" -Method Post -Body $createNginxBody -Headers $headers
    $nginxId = $nginxResponse.data.id
    Write-Host "   [PASS] Nginx created - ID: $nginxId" -ForegroundColor Green
} catch {
    Write-Host "   [FAIL] $_" -ForegroundColor Red
    exit 1
}

Start-Sleep -Seconds 3

Write-Host "`n11. Get Nginx Info" -ForegroundColor Yellow
try {
    $nginxInfo = Invoke-RestMethod -Uri "$BASE_URL_PROVISIONING/api/v1/nginx/$nginxId" -Method Get -Headers $headers
    Write-Host "   [PASS] Info retrieved - Status: $($nginxInfo.data.status), Port: $($nginxInfo.data.port)" -ForegroundColor Green
} catch {
    Write-Host "   [FAIL] $_" -ForegroundColor Red
}

Write-Host "`n12. Update Nginx Config" -ForegroundColor Yellow
try {
    $updateBody = @{
        config = "server { listen 80; location / { return 200 'Updated'; } location /health { return 200 'OK'; } }"
    } | ConvertTo-Json
    Invoke-RestMethod -Uri "$BASE_URL_PROVISIONING/api/v1/nginx/$nginxId/config" -Method Put -Body $updateBody -Headers $headers | Out-Null
    Write-Host "   [PASS] Config updated" -ForegroundColor Green
} catch {
    Write-Host "   [FAIL] $_" -ForegroundColor Red
}

Write-Host "`n13. Stop Nginx" -ForegroundColor Yellow
try {
    Invoke-RestMethod -Uri "$BASE_URL_PROVISIONING/api/v1/nginx/$nginxId/stop" -Method Post -Headers $headers | Out-Null
    Write-Host "   [PASS] Nginx stopped" -ForegroundColor Green
    Start-Sleep -Seconds 2
} catch {
    Write-Host "   [FAIL] $_" -ForegroundColor Red
}

Write-Host "`n14. Start Nginx" -ForegroundColor Yellow
try {
    Invoke-RestMethod -Uri "$BASE_URL_PROVISIONING/api/v1/nginx/$nginxId/start" -Method Post -Headers $headers | Out-Null
    Write-Host "   [PASS] Nginx started" -ForegroundColor Green
    Start-Sleep -Seconds 2
} catch {
    Write-Host "   [FAIL] $_" -ForegroundColor Red
}

Write-Host "`n15. Restart Nginx" -ForegroundColor Yellow
try {
    Invoke-RestMethod -Uri "$BASE_URL_PROVISIONING/api/v1/nginx/$nginxId/restart" -Method Post -Headers $headers | Out-Null
    Write-Host "   [PASS] Nginx restarted" -ForegroundColor Green
    Start-Sleep -Seconds 2
} catch {
    Write-Host "   [FAIL] $_" -ForegroundColor Red
}

# Monitoring Tests
Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "Monitoring Service APIs" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

$pgContainerId = $pgInfo.data.container_id

Write-Host "`n16. Get Current Metrics" -ForegroundColor Yellow
try {
    $metrics = Invoke-RestMethod -Uri "$BASE_URL_MONITORING/api/v1/monitoring/metrics/$pgContainerId" -Method Get
    Write-Host "   [PASS] Current metrics retrieved" -ForegroundColor Green
} catch {
    Write-Host "   [FAIL] $_" -ForegroundColor Red
}

Write-Host "`n17. Get Historical Metrics" -ForegroundColor Yellow
try {
    $histMetrics = Invoke-RestMethod -Uri "$BASE_URL_MONITORING/api/v1/monitoring/metrics/$pgContainerId/history?from=0&size=10" -Method Get
    Write-Host "   [PASS] Historical metrics retrieved - Count: $($histMetrics.data.Count)" -ForegroundColor Green
} catch {
    Write-Host "   [FAIL] $_" -ForegroundColor Red
}

Write-Host "`n18. Get Aggregated Metrics" -ForegroundColor Yellow
try {
    $aggMetrics = Invoke-RestMethod -Uri "$BASE_URL_MONITORING/api/v1/monitoring/metrics/$pgContainerId/aggregate?range=1h" -Method Get
    Write-Host "   [PASS] Aggregated metrics retrieved" -ForegroundColor Green
} catch {
    Write-Host "   [FAIL] $_" -ForegroundColor Red
}

Write-Host "`n19. Get Logs from Monitoring" -ForegroundColor Yellow
try {
    $monLogs = Invoke-RestMethod -Uri "$BASE_URL_MONITORING/api/v1/monitoring/logs/$pgId?from=0&size=10" -Method Get
    Write-Host "   [PASS] Monitoring logs retrieved - Count: $($monLogs.data.Count)" -ForegroundColor Green
} catch {
    Write-Host "   [FAIL] $_" -ForegroundColor Red
}

Write-Host "`n20. Get Health Status" -ForegroundColor Yellow
try {
    $health = Invoke-RestMethod -Uri "$BASE_URL_MONITORING/api/v1/monitoring/health/$pgContainerId" -Method Get
    Write-Host "   [PASS] Health status: $($health.data.status)" -ForegroundColor Green
} catch {
    Write-Host "   [FAIL] $_" -ForegroundColor Red
}

Write-Host "`n21. List All Infrastructure" -ForegroundColor Yellow
try {
    $infraList = Invoke-RestMethod -Uri "$BASE_URL_MONITORING/api/v1/monitoring/infrastructure" -Method Get
    Write-Host "   [PASS] Infrastructure list retrieved - Count: $($infraList.data.Count)" -ForegroundColor Green
} catch {
    Write-Host "   [FAIL] $_" -ForegroundColor Red
}

# Cleanup
Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "Cleanup" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

Write-Host "`n22. Delete Nginx Instance" -ForegroundColor Yellow
try {
    Invoke-RestMethod -Uri "$BASE_URL_PROVISIONING/api/v1/nginx/$nginxId" -Method Delete -Headers $headers | Out-Null
    Write-Host "   [PASS] Nginx deleted" -ForegroundColor Green
} catch {
    Write-Host "   [FAIL] $_" -ForegroundColor Red
}

Write-Host "`n23. Delete PostgreSQL Instance" -ForegroundColor Yellow
try {
    Invoke-RestMethod -Uri "$BASE_URL_PROVISIONING/api/v1/postgres/single/$pgId" -Method Delete -Headers $headers | Out-Null
    Write-Host "   [PASS] PostgreSQL deleted" -ForegroundColor Green
} catch {
    Write-Host "   [FAIL] $_" -ForegroundColor Red
}

Write-Host "`n========================================" -ForegroundColor Green
Write-Host "API Testing Complete!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green

Write-Host "`nTested APIs Summary:" -ForegroundColor Cyan
Write-Host "Authentication:" -ForegroundColor White
Write-Host "  - POST /auth/login" -ForegroundColor Gray

Write-Host "`nPostgreSQL Single Instance (Provisioning):" -ForegroundColor White
Write-Host "  - POST   /api/v1/postgres/single" -ForegroundColor Gray
Write-Host "  - GET    /api/v1/postgres/single/:id" -ForegroundColor Gray
Write-Host "  - GET    /api/v1/postgres/single/:id/stats" -ForegroundColor Gray
Write-Host "  - GET    /api/v1/postgres/single/:id/logs" -ForegroundColor Gray
Write-Host "  - POST   /api/v1/postgres/single/:id/start" -ForegroundColor Gray
Write-Host "  - POST   /api/v1/postgres/single/:id/stop" -ForegroundColor Gray
Write-Host "  - POST   /api/v1/postgres/single/:id/restart" -ForegroundColor Gray
Write-Host "  - POST   /api/v1/postgres/single/:id/backup" -ForegroundColor Gray
Write-Host "  - DELETE /api/v1/postgres/single/:id" -ForegroundColor Gray

Write-Host "`nNginx Instance (Provisioning):" -ForegroundColor White
Write-Host "  - POST   /api/v1/nginx" -ForegroundColor Gray
Write-Host "  - GET    /api/v1/nginx/:id" -ForegroundColor Gray
Write-Host "  - PUT    /api/v1/nginx/:id/config" -ForegroundColor Gray
Write-Host "  - POST   /api/v1/nginx/:id/start" -ForegroundColor Gray
Write-Host "  - POST   /api/v1/nginx/:id/stop" -ForegroundColor Gray
Write-Host "  - POST   /api/v1/nginx/:id/restart" -ForegroundColor Gray
Write-Host "  - DELETE /api/v1/nginx/:id" -ForegroundColor Gray

Write-Host "`nMonitoring Service:" -ForegroundColor White
Write-Host "  - GET /api/v1/monitoring/metrics/:instance_id" -ForegroundColor Gray
Write-Host "  - GET /api/v1/monitoring/metrics/:instance_id/history" -ForegroundColor Gray
Write-Host "  - GET /api/v1/monitoring/metrics/:instance_id/aggregate" -ForegroundColor Gray
Write-Host "  - GET /api/v1/monitoring/logs/:instance_id" -ForegroundColor Gray
Write-Host "  - GET /api/v1/monitoring/health/:instance_id" -ForegroundColor Gray
Write-Host "  - GET /api/v1/monitoring/infrastructure" -ForegroundColor Gray

Write-Host "`nUntested APIs (Not Implemented or Need Special Setup):" -ForegroundColor Yellow
Write-Host "  - POST /api/v1/postgres/single/:id/restore (requires backup file)" -ForegroundColor Gray

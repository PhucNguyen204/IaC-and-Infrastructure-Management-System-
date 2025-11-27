# Test Nginx Cluster API - Simple Version
$token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NjQyMzE5NzUsImlhdCI6MTc2NDIzMTA3NSwic2NvcGUiOlsiYWRtaW4iXSwic3ViIjoidGVzdC11c2VyLTEyMyJ9.Z0pIHyNAD_3HIF9UU7bMd8MPgDDYXHSO736d4KlRRbg"
$baseUrl = "http://localhost:8083/api/v1"

$headers = @{
    "Authorization" = "Bearer $token"
    "Content-Type" = "application/json"
}

Write-Host "=== Test Nginx Cluster API ===" -ForegroundColor Cyan

# Test 1: Create Cluster
Write-Host "`n1. Tao Nginx Cluster..." -ForegroundColor Yellow

$createBody = @{
    cluster_name = "demo-nginx"
    node_count = 2
    http_port = 8080
    https_port = 8443
    load_balance_mode = "round_robin"
    virtual_ip = "192.168.100.20"
    worker_processes = 2
    worker_connections = 2048
    ssl_enabled = $false
    gzip_enabled = $true
    health_check_enabled = $true
    health_check_path = "/health"
    cpu_limit = "1"
    memory_limit = "512m"
} | ConvertTo-Json

try {
    $createResp = Invoke-RestMethod -Uri "$baseUrl/nginx/cluster" -Method POST -Headers $headers -Body $createBody
    $clusterId = $createResp.data.id
    Write-Host "   OK - Cluster ID: $clusterId" -ForegroundColor Green
    Write-Host "   Status: $($createResp.data.status)" -ForegroundColor Gray
    Write-Host "   Nodes: $($createResp.data.node_count)" -ForegroundColor Gray
    
    Start-Sleep -Seconds 3
    
    # Test 2: Get Info
    Write-Host "`n2. Lay thong tin Cluster..." -ForegroundColor Yellow
    $infoResp = Invoke-RestMethod -Uri "$baseUrl/nginx/cluster/$clusterId" -Method GET -Headers $headers
    Write-Host "   OK - Status: $($infoResp.data.status)" -ForegroundColor Green
    Write-Host "   Nodes:" -ForegroundColor Gray
    foreach ($node in $infoResp.data.nodes) {
        Write-Host "     - $($node.node_name): $($node.status)" -ForegroundColor Gray
    }
    
    # Test 3: Health Check
    Write-Host "`n3. Kiem tra Health..." -ForegroundColor Yellow
    $healthResp = Invoke-RestMethod -Uri "$baseUrl/nginx/cluster/$clusterId/health" -Method GET -Headers $headers
    Write-Host "   OK - Health: $($healthResp.data.cluster_health)" -ForegroundColor Green
    Write-Host "   Healthy Nodes: $($healthResp.data.healthy_nodes)/$($healthResp.data.total_nodes)" -ForegroundColor Gray
    
    # Test 4: Stop
    Write-Host "`n4. Stop Cluster..." -ForegroundColor Yellow
    $stopResp = Invoke-RestMethod -Uri "$baseUrl/nginx/cluster/$clusterId/stop" -Method POST -Headers $headers
    Write-Host "   OK - Cluster stopped" -ForegroundColor Green
    
    Start-Sleep -Seconds 2
    
    # Test 5: Check status after stop
    Write-Host "`n5. Kiem tra status sau khi Stop..." -ForegroundColor Yellow
    $infoAfterStop = Invoke-RestMethod -Uri "$baseUrl/nginx/cluster/$clusterId" -Method GET -Headers $headers
    Write-Host "   OK - Status: $($infoAfterStop.data.status)" -ForegroundColor Green
    foreach ($node in $infoAfterStop.data.nodes) {
        Write-Host "     - $($node.node_name): $($node.status)" -ForegroundColor Gray
    }
    
    # Test 6: Start
    Write-Host "`n6. Start Cluster..." -ForegroundColor Yellow
    $startResp = Invoke-RestMethod -Uri "$baseUrl/nginx/cluster/$clusterId/start" -Method POST -Headers $headers
    Write-Host "   OK - Cluster started" -ForegroundColor Green
    
    Start-Sleep -Seconds 2
    
    # Test 7: Check status after start
    Write-Host "`n7. Kiem tra status sau khi Start..." -ForegroundColor Yellow
    $infoAfterStart = Invoke-RestMethod -Uri "$baseUrl/nginx/cluster/$clusterId" -Method GET -Headers $headers
    Write-Host "   OK - Status: $($infoAfterStart.data.status)" -ForegroundColor Green
    foreach ($node in $infoAfterStart.data.nodes) {
        Write-Host "     - $($node.node_name): $($node.status)" -ForegroundColor Gray
    }
    
    # Test 8: Delete
    Write-Host "`n8. Xoa Cluster..." -ForegroundColor Yellow
    $deleteResp = Invoke-RestMethod -Uri "$baseUrl/nginx/cluster/$clusterId" -Method DELETE -Headers $headers
    Write-Host "   OK - Cluster deleted" -ForegroundColor Green
    
    Write-Host "`n=== TAT CA TEST PASS! ===" -ForegroundColor Green
    
} catch {
    Write-Host "   FAIL - $($_.Exception.Message)" -ForegroundColor Red
    if ($_.ErrorDetails.Message) {
        Write-Host "   Details: $($_.ErrorDetails.Message)" -ForegroundColor Red
    }
}


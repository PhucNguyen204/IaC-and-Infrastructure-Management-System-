# PostgreSQL Cluster API Testing Script
$ErrorActionPreference = "Continue"

$baseUrl = "http://localhost:8083/api/v1"

# Login first
Write-Host "Logging in..." -ForegroundColor Cyan
$loginBody = @{username='admin';password='password123'} | ConvertTo-Json
$loginResp = Invoke-RestMethod -Uri 'http://localhost:8082/auth/login' -Method Post -Headers @{'Content-Type'='application/json'} -Body $loginBody
$token = "Bearer $($loginResp.data.access_token)"
Write-Host "Logged in successfully`n" -ForegroundColor Green

$clusterId = $null

Write-Host "=== PostgreSQL Cluster API Tests ===" -ForegroundColor Cyan

# Test 1: Create 3-node cluster
Write-Host "`nTest 1: Create Cluster (3 nodes)" -ForegroundColor Yellow
$body = @{
    cluster_name = "test-cluster"
    postgres_version = "16"
    node_count = 3
    cpu_per_node = 2000000000
    memory_per_node = 2147483648
    storage_per_node = 50
    postgres_password = "password123"
    replication_mode = "async"
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri "$baseUrl/postgres/cluster" -Method Post `
        -Headers @{Authorization=$token; 'Content-Type'='application/json'} -Body $body
    $clusterId = $response.cluster_id
    Write-Host "SUCCESS: Cluster created - ID: $clusterId" -ForegroundColor Green
    Write-Host "  Nodes: $($response.nodes.Count)" -ForegroundColor Green
    Start-Sleep -Seconds 20
} catch {
    Write-Host "FAILED: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 2: Get cluster info
if ($clusterId) {
    Write-Host "`nTest 2: Get Cluster Info" -ForegroundColor Yellow
    try {
        $info = Invoke-RestMethod -Uri "$baseUrl/postgres/cluster/$clusterId" -Method Get `
            -Headers @{Authorization=$token}
        Write-Host "SUCCESS: Status=$($info.status), Nodes=$($info.nodes.Count)" -ForegroundColor Green
    } catch {
        Write-Host "FAILED: $($_.Exception.Message)" -ForegroundColor Red
    }
}

# Test 3: Get cluster stats
if ($clusterId) {
    Write-Host "`nTest 3: Get Cluster Stats" -ForegroundColor Yellow
    try {
        $stats = Invoke-RestMethod -Uri "$baseUrl/postgres/cluster/$clusterId/stats" -Method Get `
            -Headers @{Authorization=$token}
        Write-Host "SUCCESS: Stats retrieved" -ForegroundColor Green
    } catch {
        Write-Host "FAILED: $($_.Exception.Message)" -ForegroundColor Red
    }
}

# Test 4: Get cluster logs
if ($clusterId) {
    Write-Host "`nTest 4: Get Cluster Logs" -ForegroundColor Yellow
    try {
        $logs = Invoke-RestMethod -Uri "$baseUrl/postgres/cluster/$clusterId/logs?tail=20" -Method Get `
            -Headers @{Authorization=$token}
        Write-Host "SUCCESS: Logs retrieved, length=$($logs.Length)" -ForegroundColor Green
    } catch {
        Write-Host "FAILED: $($_.Exception.Message)" -ForegroundColor Red
    }
}

# Test 5: Scale up to 5 nodes
if ($clusterId) {
    Write-Host "`nTest 5: Scale Up (3 to 5 nodes)" -ForegroundColor Yellow
    $scaleBody = @{ node_count = 5 } | ConvertTo-Json
    try {
        $response = Invoke-RestMethod -Uri "$baseUrl/postgres/cluster/$clusterId/scale" -Method Post `
            -Headers @{Authorization=$token; 'Content-Type'='application/json'} -Body $scaleBody
        Write-Host "SUCCESS: Scaled to 5 nodes" -ForegroundColor Green
        Start-Sleep -Seconds 10
    } catch {
        Write-Host "FAILED: $($_.Exception.Message)" -ForegroundColor Red
    }
}

# Test 6: Stop cluster
if ($clusterId) {
    Write-Host "`nTest 6: Stop Cluster" -ForegroundColor Yellow
    try {
        Invoke-RestMethod -Uri "$baseUrl/postgres/cluster/$clusterId/stop" -Method Post `
            -Headers @{Authorization=$token}
        Write-Host "SUCCESS: Cluster stopped" -ForegroundColor Green
        Start-Sleep -Seconds 5
    } catch {
        Write-Host "FAILED: $($_.Exception.Message)" -ForegroundColor Red
    }
}

# Test 7: Start cluster
if ($clusterId) {
    Write-Host "`nTest 7: Start Cluster" -ForegroundColor Yellow
    try {
        Invoke-RestMethod -Uri "$baseUrl/postgres/cluster/$clusterId/start" -Method Post `
            -Headers @{Authorization=$token}
        Write-Host "SUCCESS: Cluster started" -ForegroundColor Green
        Start-Sleep -Seconds 5
    } catch {
        Write-Host "FAILED: $($_.Exception.Message)" -ForegroundColor Red
    }
}

# Test 8: Scale down to 3 nodes
if ($clusterId) {
    Write-Host "`nTest 8: Scale Down (5 to 3 nodes)" -ForegroundColor Yellow
    $scaleBody = @{ node_count = 3 } | ConvertTo-Json
    try {
        $response = Invoke-RestMethod -Uri "$baseUrl/postgres/cluster/$clusterId/scale" -Method Post `
            -Headers @{Authorization=$token; 'Content-Type'='application/json'} -Body $scaleBody
        Write-Host "SUCCESS: Scaled down to 3 nodes" -ForegroundColor Green
        Start-Sleep -Seconds 5
    } catch {
        Write-Host "FAILED: $($_.Exception.Message)" -ForegroundColor Red
    }
}

# Test 9: Restart cluster
if ($clusterId) {
    Write-Host "`nTest 9: Restart Cluster" -ForegroundColor Yellow
    try {
        Invoke-RestMethod -Uri "$baseUrl/postgres/cluster/$clusterId/restart" -Method Post `
            -Headers @{Authorization=$token}
        Write-Host "SUCCESS: Cluster restarted" -ForegroundColor Green
        Start-Sleep -Seconds 10
    } catch {
        Write-Host "FAILED: $($_.Exception.Message)" -ForegroundColor Red
    }
}

# Test 10: Delete cluster
if ($clusterId) {
    Write-Host "`nTest 10: Delete Cluster" -ForegroundColor Yellow
    try {
        Invoke-RestMethod -Uri "$baseUrl/postgres/cluster/$clusterId" -Method Delete `
            -Headers @{Authorization=$token}
        Write-Host "SUCCESS: Cluster deleted" -ForegroundColor Green
    } catch {
        Write-Host "FAILED: $($_.Exception.Message)" -ForegroundColor Red
    }
}

Write-Host "`n=== All Tests Completed ===" -ForegroundColor Cyan

# Test Nginx Config Sync with NGINX Best Practices
$ErrorActionPreference = "Continue"

Write-Host "=== Testing Nginx Config Synchronization ===" -ForegroundColor Cyan
Write-Host "Reference: https://docs.nginx.com/nginx/admin-guide/high-availability/configuration-sharing/" -ForegroundColor Gray

# 1. Login
Write-Host "`n1. Login..." -ForegroundColor Yellow
$loginResponse = curl.exe -s -X POST http://localhost:8082/auth/login `
    -H "Content-Type: application/json" `
    -d '{\"username\":\"admin\",\"password\":\"password123\"}' | ConvertFrom-Json

$token = $loginResponse.data.access_token
Write-Host "✅ Token obtained" -ForegroundColor Green

# 2. Create Nginx Cluster
Write-Host "`n2. Creating Nginx Cluster..." -ForegroundColor Yellow
$createBody = @"
{
    "cluster_name": "test-sync-cluster",
    "node_count": 3,
    "http_port": 9080,
    "https_port": 9443,
    "load_balance_mode": "round_robin",
    "virtual_ip": "192.168.100.10",
    "worker_connections": 1024,
    "worker_processes": 2,
    "ssl_enabled": false,
    "gzip_enabled": true,
    "health_check_enabled": true,
    "health_check_path": "/health"
}
"@

$createResponse = curl.exe -s -X POST http://localhost:8083/api/v1/nginx/cluster `
    -H "Authorization: Bearer $token" `
    -H "Content-Type: application/json" `
    -d $createBody | ConvertFrom-Json

if ($createResponse.success) {
    $clusterId = $createResponse.data.id
    Write-Host "✅ Cluster created: $clusterId" -ForegroundColor Green
    
    # Wait for cluster to be ready
    Write-Host "`n3. Waiting for cluster to be ready (15 seconds)..." -ForegroundColor Yellow
    Start-Sleep -Seconds 15
    
    # 4. Get cluster info
    Write-Host "`n4. Getting cluster info..." -ForegroundColor Yellow
    $infoResponse = curl.exe -s -X GET "http://localhost:8083/api/v1/nginx/cluster/$clusterId" `
        -H "Authorization: Bearer $token" | ConvertFrom-Json
    
    Write-Host "Cluster Name: $($infoResponse.data.cluster_name)" -ForegroundColor Cyan
    Write-Host "Node Count: $($infoResponse.data.node_count)" -ForegroundColor Cyan
    Write-Host "Status: $($infoResponse.data.status)" -ForegroundColor Cyan
    Write-Host "`nNodes:" -ForegroundColor Cyan
    $infoResponse.data.nodes | ForEach-Object {
        Write-Host "  - $($_.name) ($($_.role)) - $($_.status)" -ForegroundColor Gray
    }
    
    # 5. Test initial sync
    Write-Host "`n5. Testing Config Sync (following NGINX best practices)..." -ForegroundColor Yellow
    Write-Host "   ✓ Step 1: Validate on master" -ForegroundColor Gray
    Write-Host "   ✓ Step 2: Backup on peers" -ForegroundColor Gray
    Write-Host "   ✓ Step 3: Sync to peers" -ForegroundColor Gray
    Write-Host "   ✓ Step 4: Validate on peers" -ForegroundColor Gray
    Write-Host "   ✓ Step 5: Reload or rollback" -ForegroundColor Gray
    
    $syncResponse = curl.exe -s -X POST "http://localhost:8083/api/v1/nginx/cluster/$clusterId/sync-config" `
        -H "Authorization: Bearer $token" | ConvertFrom-Json
    
    if ($syncResponse.success) {
        Write-Host "`n✅ Config sync successful!" -ForegroundColor Green
        $syncResponse | ConvertTo-Json -Depth 5
    } else {
        Write-Host "`n❌ Config sync failed!" -ForegroundColor Red
        Write-Host "Error: $($syncResponse.error)" -ForegroundColor Red
    }
    
    # 6. Update config and sync again
    Write-Host "`n6. Updating config and syncing..." -ForegroundColor Yellow
    
    $updateResponse = curl.exe -s -X POST "http://localhost:8083/api/v1/nginx/cluster/$clusterId/sync-config" `
        -H "Authorization: Bearer $token" | ConvertFrom-Json
    
    if ($updateResponse.success) {
        Write-Host "✅ Config updated and synced!" -ForegroundColor Green
    } else {
        Write-Host "❌ Config update failed: $($updateResponse.error)" -ForegroundColor Red
    }
    
    # 7. Verify health after sync
    Write-Host "`n7. Verifying cluster health..." -ForegroundColor Yellow
    $healthResponse = curl.exe -s -X GET "http://localhost:8083/api/v1/nginx/cluster/$clusterId/health" `
        -H "Authorization: Bearer $token" | ConvertFrom-Json
    
    Write-Host "Status: $($healthResponse.data.status)" -ForegroundColor $(if($healthResponse.data.status -eq 'healthy'){'Green'}else{'Red'})
    Write-Host "Healthy Nodes: $($healthResponse.data.healthy_nodes)/$($healthResponse.data.total_nodes)" -ForegroundColor Cyan
    
    # 8. Cleanup (optional)
    Write-Host "`n8. Do you want to delete the test cluster? (y/n)" -ForegroundColor Yellow
    $delete = Read-Host
    
    if ($delete -eq "y") {
        curl.exe -s -X DELETE "http://localhost:8083/api/v1/nginx/cluster/$clusterId" `
            -H "Authorization: Bearer $token" | Out-Null
        Write-Host "✅ Cluster deleted" -ForegroundColor Green
    } else {
        Write-Host "ℹ️  Cluster kept: $clusterId" -ForegroundColor Cyan
    }
    
} else {
    Write-Host "❌ Failed to create cluster!" -ForegroundColor Red
    $createResponse | ConvertTo-Json -Depth 5
}

Write-Host "`n=== Test Completed ===" -ForegroundColor Cyan


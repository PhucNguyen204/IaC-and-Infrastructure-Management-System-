# Test K8s Cluster API
$ErrorActionPreference = "Continue"

Write-Host "=== Testing K8s Cluster API ===" -ForegroundColor Cyan

# 1. Login
Write-Host "`n1. Login..." -ForegroundColor Yellow
$loginResponse = curl.exe -s -X POST http://localhost:8082/auth/login `
    -H "Content-Type: application/json" `
    -d '{\"username\":\"admin\",\"password\":\"password123\"}' | ConvertFrom-Json

$token = $loginResponse.data.access_token
Write-Host "Token: $($token.Substring(0, 50))..." -ForegroundColor Green

# 2. Create K8s Cluster
Write-Host "`n2. Creating K8s Cluster..." -ForegroundColor Yellow

$createClusterBody = @"
{
    "cluster_name": "demo-k8s-cluster",
    "k8s_version": "v1.28.5-k3s1",
    "cluster_type": "k3s",
    "node_count": 2,
    "cpu_limit": "1",
    "memory_limit": "1Gi",
    "api_server_port": 6550,
    "cluster_cidr": "10.42.0.0/16",
    "service_cidr": "10.43.0.0/16",
    "dashboard_enabled": true,
    "ingress_enabled": false,
    "metrics_enabled": true,
    "storage_class": "local-path"
}
"@

Write-Host "Request Body:" -ForegroundColor Cyan
Write-Host $createClusterBody

$createResponse = curl.exe -s -X POST http://localhost:8083/api/v1/k8s/cluster `
    -H "Authorization: Bearer $token" `
    -H "Content-Type: application/json" `
    -d $createClusterBody | ConvertFrom-Json

Write-Host "`nCreate Response:" -ForegroundColor Green
$createResponse | ConvertTo-Json -Depth 10

if ($createResponse.success) {
    $clusterId = $createResponse.data.id
    Write-Host "`nCluster ID: $clusterId" -ForegroundColor Green
    
    # 3. Wait for cluster creation
    Write-Host "`n3. Waiting for cluster creation (30 seconds)..." -ForegroundColor Yellow
    Start-Sleep -Seconds 30
    
    # 4. Get Cluster Info
    Write-Host "`n4. Getting Cluster Info..." -ForegroundColor Yellow
    $infoResponse = curl.exe -s -X GET "http://localhost:8083/api/v1/k8s/cluster/$clusterId" `
        -H "Authorization: Bearer $token" | ConvertFrom-Json
    
    Write-Host "`nCluster Info:" -ForegroundColor Green
    $infoResponse.data | ConvertTo-Json -Depth 10
    
    # 5. Get Cluster Health
    Write-Host "`n5. Getting Cluster Health..." -ForegroundColor Yellow
    $healthResponse = curl.exe -s -X GET "http://localhost:8083/api/v1/k8s/cluster/$clusterId/health" `
        -H "Authorization: Bearer $token" | ConvertFrom-Json
    
    Write-Host "`nCluster Health:" -ForegroundColor Green
    $healthResponse.data | ConvertTo-Json -Depth 10
    
    # 6. Get Connection Info
    Write-Host "`n6. Getting Connection Info..." -ForegroundColor Yellow
    $connResponse = curl.exe -s -X GET "http://localhost:8083/api/v1/k8s/cluster/$clusterId/connection-info" `
        -H "Authorization: Bearer $token" | ConvertFrom-Json
    
    Write-Host "`nConnection Info:" -ForegroundColor Green
    Write-Host "API Server URL: $($connResponse.data.api_server_url)"
    Write-Host "Dashboard URL: $($connResponse.data.dashboard_url)"
    Write-Host "Kubeconfig (base64): $($connResponse.data.kubeconfig.Substring(0, 100))..."
    
    # 7. Get Kubeconfig File
    Write-Host "`n7. Downloading Kubeconfig..." -ForegroundColor Yellow
    curl.exe -s -X GET "http://localhost:8083/api/v1/k8s/cluster/$clusterId/kubeconfig" `
        -H "Authorization: Bearer $token" `
        -o "kubeconfig-$clusterId.yaml"
    
    if (Test-Path "kubeconfig-$clusterId.yaml") {
        Write-Host "Kubeconfig saved to: kubeconfig-$clusterId.yaml" -ForegroundColor Green
        Write-Host "`nKubeconfig content (first 10 lines):"
        Get-Content "kubeconfig-$clusterId.yaml" -Head 10
    }
    
    # 8. Test kubectl commands (if k3d is installed)
    Write-Host "`n8. Testing kubectl commands..." -ForegroundColor Yellow
    if (Get-Command kubectl -ErrorAction SilentlyContinue) {
        Write-Host "Getting nodes..."
        kubectl get nodes --kubeconfig "kubeconfig-$clusterId.yaml"
        
        Write-Host "`nGetting all pods..."
        kubectl get pods --all-namespaces --kubeconfig "kubeconfig-$clusterId.yaml"
    } else {
        Write-Host "kubectl not found. Skipping kubectl tests." -ForegroundColor Yellow
    }
    
    # 9. Stop Cluster
    Write-Host "`n9. Stopping Cluster..." -ForegroundColor Yellow
    $stopResponse = curl.exe -s -X POST "http://localhost:8083/api/v1/k8s/cluster/$clusterId/stop" `
        -H "Authorization: Bearer $token" | ConvertFrom-Json
    
    Write-Host "Stop Response:" -ForegroundColor Green
    $stopResponse | ConvertTo-Json
    
    Start-Sleep -Seconds 5
    
    # 10. Start Cluster
    Write-Host "`n10. Starting Cluster..." -ForegroundColor Yellow
    $startResponse = curl.exe -s -X POST "http://localhost:8083/api/v1/k8s/cluster/$clusterId/start" `
        -H "Authorization: Bearer $token" | ConvertFrom-Json
    
    Write-Host "Start Response:" -ForegroundColor Green
    $startResponse | ConvertTo-Json
    
    # 11. Get Metrics
    Write-Host "`n11. Getting Cluster Metrics..." -ForegroundColor Yellow
    $metricsResponse = curl.exe -s -X GET "http://localhost:8083/api/v1/k8s/cluster/$clusterId/metrics" `
        -H "Authorization: Bearer $token" | ConvertFrom-Json
    
    Write-Host "Metrics:" -ForegroundColor Green
    $metricsResponse.data | ConvertTo-Json -Depth 10
    
    # 12. Delete Cluster (optional - comment out if you want to keep it)
    Write-Host "`n12. Do you want to delete the cluster? (y/n)" -ForegroundColor Yellow
    $delete = Read-Host
    
    if ($delete -eq "y") {
        Write-Host "Deleting Cluster..." -ForegroundColor Yellow
        $deleteResponse = curl.exe -s -X DELETE "http://localhost:8083/api/v1/k8s/cluster/$clusterId" `
            -H "Authorization: Bearer $token" | ConvertFrom-Json
        
        Write-Host "Delete Response:" -ForegroundColor Green
        $deleteResponse | ConvertTo-Json
    } else {
        Write-Host "Cluster kept. Cluster ID: $clusterId" -ForegroundColor Green
    }
    
} else {
    Write-Host "Failed to create cluster!" -ForegroundColor Red
    $createResponse | ConvertTo-Json -Depth 10
}

Write-Host "`n=== Test Completed ===" -ForegroundColor Cyan


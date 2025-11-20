# PostgreSQL Cluster API Testing Script
# Test all cluster management endpoints

$baseUrl = "http://localhost:8083/api/v1"
$token = "Bearer test-token"

Write-Host "`n=== PostgreSQL Cluster API Tests ===" -ForegroundColor Cyan
Write-Host "Testing cluster management operations`n" -ForegroundColor Cyan

# Test 1: Create Cluster (3 nodes)
Write-Host "Test 1: Create PostgreSQL Cluster (3 nodes)" -ForegroundColor Yellow
$createClusterBody = @{
    cluster_name = "prod-cluster-01"
    postgres_version = "16"
    node_count = 3
    cpu_per_node = 2000000000
    memory_per_node = 2147483648
    storage_per_node = 50
    postgres_password = "secure_password_123"
    replication_mode = "async"
} | ConvertTo-Json

try {
    $createResponse = Invoke-RestMethod -Uri "$baseUrl/postgres/cluster" -Method Post `
        -Headers @{Authorization=$token; "Content-Type"="application/json"} `
        -Body $createClusterBody
    
    $clusterId = $createResponse.cluster_id
    Write-Host "✓ Cluster created successfully" -ForegroundColor Green
    Write-Host "  Cluster ID: $clusterId"
    Write-Host "  Cluster Name: $($createResponse.cluster_name)"
    Write-Host "  Status: $($createResponse.status)"
    Write-Host "  Nodes: $($createResponse.nodes.Count)"
    
    # Display node info
    foreach ($node in $createResponse.nodes) {
        Write-Host "    - $($node.node_name): $($node.role) (Container: $($node.container_id.Substring(0,12)))"
    }
} catch {
    Write-Host "✗ Create cluster failed: $($_.Exception.Message)" -ForegroundColor Red
    $clusterId = $null
}

Start-Sleep -Seconds 20

# Test 2: Get Cluster Info
if ($clusterId) {
    Write-Host "`nTest 2: Get Cluster Info" -ForegroundColor Yellow
    try {
        $clusterInfo = Invoke-RestMethod -Uri "$baseUrl/postgres/cluster/$clusterId" -Method Get `
            -Headers @{Authorization=$token}
        
        Write-Host "✓ Cluster info retrieved" -ForegroundColor Green
        Write-Host "  Status: $($clusterInfo.status)"
        Write-Host "  Primary Port: $($clusterInfo.haproxy_port)"
        Write-Host "  Node Count: $($clusterInfo.nodes.Count)"
        Write-Host "  Version: $($clusterInfo.postgres_version)"
    } catch {
        Write-Host "✗ Get cluster info failed: $($_.Exception.Message)" -ForegroundColor Red
    }
}

# Test 3: Get Replication Status
if ($clusterId) {
    Write-Host "`nTest 3: Get Replication Status" -ForegroundColor Yellow
    try {
        $repStatus = Invoke-RestMethod -Uri "$baseUrl/postgres/cluster/$clusterId/replication" -Method Get `
            -Headers @{Authorization=$token}
        
        Write-Host "✓ Replication status retrieved" -ForegroundColor Green
        Write-Host "  Primary: $($repStatus.primary)"
        Write-Host "  Replicas:"
        foreach ($replica in $repStatus.replicas) {
            Write-Host "    - $($replica.node_name): State=$($replica.state), Lag=$($replica.lag_bytes) bytes, Healthy=$($replica.is_healthy)"
        }
    } catch {
        Write-Host "✗ Get replication status failed: $($_.Exception.Message)" -ForegroundColor Red
    }
}

# Test 4: Get Cluster Stats
if ($clusterId) {
    Write-Host "`nTest 4: Get Cluster Stats" -ForegroundColor Yellow
    try {
        $stats = Invoke-RestMethod -Uri "$baseUrl/postgres/cluster/$clusterId/stats" -Method Get `
            -Headers @{Authorization=$token}
        
        Write-Host "✓ Cluster stats retrieved" -ForegroundColor Green
        Write-Host "  Total Connections: $($stats.total_connections)"
        Write-Host "  Nodes:"
        foreach ($node in $stats.nodes) {
            Write-Host "    - $($node.node_name) ($($node.role)): $($node.active_connections) connections"
        }
    } catch {
        Write-Host "✗ Get cluster stats failed: $($_.Exception.Message)" -ForegroundColor Red
    }
}

# Test 5: Get Cluster Logs
if ($clusterId) {
    Write-Host "`nTest 5: Get Cluster Logs" -ForegroundColor Yellow
    try {
        $logs = Invoke-RestMethod -Uri "$baseUrl/postgres/cluster/$clusterId/logs?tail=50" -Method Get `
            -Headers @{Authorization=$token}
        
        Write-Host "✓ Cluster logs retrieved" -ForegroundColor Green
        Write-Host "  Logs from $($logs.logs.Count) nodes"
        foreach ($nodeLog in $logs.logs) {
            $logLines = ($nodeLog.logs -split "`n").Count
            Write-Host "    - $($nodeLog.node_name): $logLines lines"
        }
    } catch {
        Write-Host "✗ Get cluster logs failed: $($_.Exception.Message)" -ForegroundColor Red
    }
}

# Test 6: Stop Cluster
if ($clusterId) {
    Write-Host "`nTest 6: Stop Cluster" -ForegroundColor Yellow
    try {
        $stopResponse = Invoke-RestMethod -Uri "$baseUrl/postgres/cluster/$clusterId/stop" -Method Post `
            -Headers @{Authorization=$token}
        
        Write-Host "✓ Cluster stopped: $($stopResponse.message)" -ForegroundColor Green
        Start-Sleep -Seconds 5
    } catch {
        Write-Host "✗ Stop cluster failed: $($_.Exception.Message)" -ForegroundColor Red
    }
}

# Test 7: Start Cluster
if ($clusterId) {
    Write-Host "`nTest 7: Start Cluster" -ForegroundColor Yellow
    try {
        $startResponse = Invoke-RestMethod -Uri "$baseUrl/postgres/cluster/$clusterId/start" -Method Post `
            -Headers @{Authorization=$token}
        
        Write-Host "✓ Cluster started: $($startResponse.message)" -ForegroundColor Green
        Start-Sleep -Seconds 10
    } catch {
        Write-Host "✗ Start cluster failed: $($_.Exception.Message)" -ForegroundColor Red
    }
}

# Test 8: Restart Cluster
if ($clusterId) {
    Write-Host "`nTest 8: Restart Cluster" -ForegroundColor Yellow
    try {
        $restartResponse = Invoke-RestMethod -Uri "$baseUrl/postgres/cluster/$clusterId/restart" -Method Post `
            -Headers @{Authorization=$token}
        
        Write-Host "✓ Cluster restarted: $($restartResponse.message)" -ForegroundColor Green
        Start-Sleep -Seconds 15
    } catch {
        Write-Host "✗ Restart cluster failed: $($_.Exception.Message)" -ForegroundColor Red
    }
}

# Test 9: Scale Up (add 2 more replicas to total 5 nodes)
if ($clusterId) {
    Write-Host "`nTest 9: Scale Up Cluster (3 to 5 nodes)" -ForegroundColor Yellow
    $scaleUpBody = @{
        node_count = 5
    } | ConvertTo-Json
    
    try {
        $scaleResponse = Invoke-RestMethod -Uri "$baseUrl/postgres/cluster/$clusterId/scale" -Method Post `
            -Headers @{Authorization=$token; "Content-Type"="application/json"} `
            -Body $scaleUpBody
        
        Write-Host "✓ Cluster scaled up: $($scaleResponse.message)" -ForegroundColor Green
        Start-Sleep -Seconds 15
        
        # Verify node count
        $clusterInfo = Invoke-RestMethod -Uri "$baseUrl/postgres/cluster/$clusterId" -Method Get `
            -Headers @{Authorization=$token}
        Write-Host "  Current node count: $($clusterInfo.nodes.Count)"
    } catch {
        Write-Host "✗ Scale up failed: $($_.Exception.Message)" -ForegroundColor Red
    }
}

# Test 10: Scale Down (remove 2 replicas to total 3 nodes)
if ($clusterId) {
    Write-Host "`nTest 10: Scale Down Cluster (5 to 3 nodes)" -ForegroundColor Yellow
    $scaleDownBody = @{
        node_count = 3
    } | ConvertTo-Json
    
    try {
        $scaleResponse = Invoke-RestMethod -Uri "$baseUrl/postgres/cluster/$clusterId/scale" -Method Post `
            -Headers @{Authorization=$token; "Content-Type"="application/json"} `
            -Body $scaleDownBody
        
        Write-Host "✓ Cluster scaled down: $($scaleResponse.message)" -ForegroundColor Green
        Start-Sleep -Seconds 5
        
        # Verify node count
        $clusterInfo = Invoke-RestMethod -Uri "$baseUrl/postgres/cluster/$clusterId" -Method Get `
            -Headers @{Authorization=$token}
        Write-Host "  Current node count: $($clusterInfo.nodes.Count)"
    } catch {
        Write-Host "✗ Scale down failed: $($_.Exception.Message)" -ForegroundColor Red
    }
}

# Test 11: Manual Failover (promote replica to primary)
if ($clusterId) {
    Write-Host "`nTest 11: Manual Failover" -ForegroundColor Yellow
    try {
        # Get current cluster info to find a replica
        $clusterInfo = Invoke-RestMethod -Uri "$baseUrl/postgres/cluster/$clusterId" -Method Get `
            -Headers @{Authorization=$token}
        
        $replicaNode = $clusterInfo.nodes | Where-Object { $_.role -eq "replica" } | Select-Object -First 1
        
        if ($replicaNode) {
            $failoverBody = @{
                new_primary_node_id = $replicaNode.node_id
            } | ConvertTo-Json
            
            $failoverResponse = Invoke-RestMethod -Uri "$baseUrl/postgres/cluster/$clusterId/failover" -Method Post `
                -Headers @{Authorization=$token; "Content-Type"="application/json"} `
                -Body $failoverBody
            
            Write-Host "✓ Failover completed: $($failoverResponse.message)" -ForegroundColor Green
            Write-Host "  New primary: $($replicaNode.node_name)"
            Start-Sleep -Seconds 5
        } else {
            Write-Host "⊘ No replica available for failover" -ForegroundColor Gray
        }
    } catch {
        Write-Host "✗ Failover failed: $($_.Exception.Message)" -ForegroundColor Red
    }
}

# Test 12: Delete Cluster
if ($clusterId) {
    Write-Host "`nTest 12: Delete Cluster" -ForegroundColor Yellow
    $confirmation = Read-Host "Do you want to delete the cluster? (yes/no)"
    
    if ($confirmation -eq "yes") {
        try {
            $deleteResponse = Invoke-RestMethod -Uri "$baseUrl/postgres/cluster/$clusterId" -Method Delete `
                -Headers @{Authorization=$token}
            
            Write-Host "✓ Cluster deleted: $($deleteResponse.message)" -ForegroundColor Green
        } catch {
            Write-Host "✗ Delete cluster failed: $($_.Exception.Message)" -ForegroundColor Red
        }
    } else {
        Write-Host "⊘ Cluster deletion skipped" -ForegroundColor Gray
        Write-Host "  Cluster ID: $clusterId (remember to delete manually)"
    }
}

Write-Host "`n=== Test Summary ===" -ForegroundColor Cyan
Write-Host "Cluster API testing completed!" -ForegroundColor Cyan
Write-Host "Review the results above to ensure all operations work correctly.`n"

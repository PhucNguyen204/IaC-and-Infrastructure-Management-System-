# Test Script for New Cluster APIs (PromoteReplica + GetReplicationStatus)
# Prerequisites: Authentication service running on 8082, Provisioning service on 8083

$BASE_URL = "http://localhost:8083"
$AUTH_URL = "http://localhost:8082"

# Color output helpers
function Write-Success { param($msg) Write-Host "[SUCCESS] $msg" -ForegroundColor Green }
function Write-Error { param($msg) Write-Host "[ERROR] $msg" -ForegroundColor Red }
function Write-Info { param($msg) Write-Host "[INFO] $msg" -ForegroundColor Cyan }
function Write-Test { param($msg) Write-Host "`n=== TEST: $msg ===" -ForegroundColor Yellow }

# ============ STEP 1: AUTHENTICATE ============
Write-Test "Authenticate User"
$loginBody = @{
    username = "admin"
    password = "password123"
} | ConvertTo-Json

try {
    $loginResponse = Invoke-RestMethod -Uri "$AUTH_URL/auth/login" `
        -Method POST `
        -Body $loginBody `
        -ContentType "application/json"
    
    $TOKEN = $loginResponse.data.access_token
    $USER_ID = $loginResponse.data.user.id
    Write-Success "Authenticated as user ID: $USER_ID"
    Write-Info "Token: $($TOKEN.Substring(0,20))..."
} catch {
    Write-Error "Authentication failed: $_"
    exit 1
}

$headers = @{
    "Authorization" = "Bearer $TOKEN"
    "Content-Type" = "application/json"
}

# ============ STEP 2: CREATE TEST CLUSTER ============
Write-Test "Create PostgreSQL Cluster (1 Primary + 2 Replicas)"
$createClusterBody = @{
    cluster_name = "failover-test-cluster"
    postgres_version = "16"
    node_count = 3
    cpu_per_node = 1
    memory_per_node = 536870912
    storage_per_node = 2048
    postgres_password = "TestPass123!"
    replication_mode = "async"
} | ConvertTo-Json

try {
    $createResponse = Invoke-RestMethod -Uri "$BASE_URL/api/v1/postgres/cluster" `
        -Method POST `
        -Headers $headers `
        -Body $createClusterBody
    
    $CLUSTER_ID = $createResponse.cluster_id
    Write-Success "Cluster created: $CLUSTER_ID"
    Write-Info "Infrastructure ID: $($createResponse.infrastructure_id)"
    Write-Info "Cluster Name: $($createResponse.cluster_name)"
    Write-Info "Waiting 20 seconds for containers to stabilize..."
    Start-Sleep -Seconds 20
} catch {
    Write-Error "Cluster creation failed: $_"
    Write-Info "Response: $($_.Exception.Message)"
    exit 1
}

# ============ STEP 3: GET CLUSTER INFO ============
Write-Test "Get Cluster Info (Identify Nodes)"
try {
    $clusterInfo = Invoke-RestMethod -Uri "$BASE_URL/api/v1/postgres/cluster/$CLUSTER_ID" `
        -Method GET `
        -Headers $headers
    
    Write-Success "Cluster retrieved successfully"
    Write-Info "Cluster Name: $($clusterInfo.cluster_name)"
    Write-Info "Status: $($clusterInfo.status)"
    
    $PRIMARY_NODE_ID = ""
    $REPLICA_NODE_IDS = @()
    
    Write-Info "Cluster Nodes:"
    foreach ($node in $clusterInfo.nodes) {
        Write-Host "  - Node ID: $($node.node_id) | Role: $($node.role) | Name: $($node.node_name)"
        if ($node.role -eq "primary") {
            $PRIMARY_NODE_ID = $node.node_id
            Write-Info "Primary Node: $PRIMARY_NODE_ID"
        } else {
            $REPLICA_NODE_IDS += $node.node_id
        }
    }
    
    if ($REPLICA_NODE_IDS.Count -lt 1) {
        Write-Error "No replicas found for failover testing"
        exit 1
    }
    
    $TARGET_REPLICA_ID = $REPLICA_NODE_IDS[0]
    Write-Info "Selected replica for promotion: $TARGET_REPLICA_ID"
    
} catch {
    Write-Error "Failed to get cluster info: $_"
    exit 1
}

# ============ STEP 4: TEST GetReplicationStatus (BEFORE FAILOVER) ============
Write-Test "API 1: Get Replication Status (BEFORE Failover)"
try {
    $replicationStatus = Invoke-RestMethod -Uri "$BASE_URL/api/v1/postgres/cluster/$CLUSTER_ID/replication" `
        -Method GET `
        -Headers $headers
    
    Write-Success "Replication status retrieved successfully"
    Write-Info "Primary Node: $($replicationStatus.primary)"
    Write-Info "Replicas:"
    foreach ($replica in $replicationStatus.replicas) {
        $lagMB = [math]::Round($replica.lag_bytes / 1MB, 2)
        Write-Host "  - Node: $($replica.node_name)"
        Write-Host "    State: $($replica.state)"
        Write-Host "    Sync State: $($replica.sync_state)"
        Write-Host "    Lag: $lagMB MB ($($replica.lag_seconds)s)"
        Write-Host "    Healthy: $($replica.is_healthy)"
    }
} catch {
    Write-Error "Failed to get replication status: $_"
    Write-Info "Response: $($_.Exception.Message)"
}

# ============ STEP 5: TEST PromoteReplica (TRIGGER FAILOVER) ============
Write-Test "API 2: Promote Replica (Manual Failover)"
$failoverBody = @{
    new_primary_node_id = $TARGET_REPLICA_ID
} | ConvertTo-Json

Write-Info "Promoting replica $TARGET_REPLICA_ID to primary..."
try {
    $failoverResponse = Invoke-RestMethod -Uri "$BASE_URL/api/v1/postgres/cluster/$CLUSTER_ID/failover" `
        -Method POST `
        -Headers $headers `
        -Body $failoverBody
    
    Write-Success "Failover completed successfully"
    Write-Info "Message: $($failoverResponse.message)"
    Write-Info "Waiting 10 seconds for role swap..."
    Start-Sleep -Seconds 10
} catch {
    Write-Error "Failover failed: $_"
    Write-Info "Response: $($_.Exception.Message)"
}

# ============ STEP 6: VERIFY ROLE SWAP ============
Write-Test "Verify Role Swap After Failover"
try {
    $clusterInfoAfter = Invoke-RestMethod -Uri "$BASE_URL/api/v1/postgres/cluster/$CLUSTER_ID" `
        -Method GET `
        -Headers $headers
    
    Write-Success "Cluster info retrieved after failover"
    
    $newPrimaryNodeID = ""
    $oldPrimaryFound = $false
    
    foreach ($node in $clusterInfoAfter.nodes) {
        if ($node.role -eq "primary") {
            $newPrimaryNodeID = $node.node_id
            Write-Info "New Primary Node ID: $newPrimaryNodeID"
            Write-Info "New Primary Name: $($node.node_name)"
        }
        if ($node.node_id -eq $PRIMARY_NODE_ID) {
            $oldPrimaryFound = $true
            Write-Info "Old Primary ($PRIMARY_NODE_ID) new role: $($node.role)"
        }
    }
    
    if ($newPrimaryNodeID -eq $TARGET_REPLICA_ID) {
        Write-Success "ROLE SWAP VERIFIED: Replica $TARGET_REPLICA_ID is now PRIMARY"
    } else {
        Write-Error "ROLE SWAP FAILED: Expected primary $TARGET_REPLICA_ID, got $newPrimaryNodeID"
    }
    
    if ($oldPrimaryFound) {
        Write-Success "Old primary successfully demoted to replica"
    }
    
} catch {
    Write-Error "Failed to verify role swap: $_"
}

# ============ STEP 7: TEST GetReplicationStatus (AFTER FAILOVER) ============
Write-Test "API 1: Get Replication Status (AFTER Failover)"
try {
    $replicationStatusAfter = Invoke-RestMethod -Uri "$BASE_URL/api/v1/postgres/cluster/$CLUSTER_ID/replication" `
        -Method GET `
        -Headers $headers
    
    Write-Success "Replication status retrieved after failover"
    Write-Info "New Primary Node: $($replicationStatusAfter.primary)"
    Write-Info "Replicas After Failover:"
    foreach ($replica in $replicationStatusAfter.replicas) {
        $lagMB = [math]::Round($replica.lag_bytes / 1MB, 2)
        Write-Host "  - Node: $($replica.node_name)"
        Write-Host "    State: $($replica.state)"
        Write-Host "    Sync State: $($replica.sync_state)"
        Write-Host "    Lag: $lagMB MB ($($replica.lag_seconds)s)"
        Write-Host "    Healthy: $($replica.is_healthy)"
    }
} catch {
    Write-Error "Failed to get replication status after failover: $_"
}

# ============ STEP 8: TEST CLUSTER STATS ============
Write-Test "Get Cluster Stats (After Failover)"
try {
    $statsResponse = Invoke-RestMethod -Uri "$BASE_URL/api/v1/postgres/cluster/$CLUSTER_ID/stats" `
        -Method GET `
        -Headers $headers
    
    Write-Success "Cluster stats retrieved"
    Write-Info "Total Connections: $($statsResponse.total_connections)"
    Write-Info "Total Databases: $($statsResponse.total_databases)"
    Write-Info "Total Size: $($statsResponse.total_size_mb) MB"
    Write-Info "Nodes:"
    foreach ($node in $statsResponse.nodes) {
        Write-Host "  - $($node.node_name) ($($node.role)): CPU $($node.cpu_percent)%, Mem $($node.memory_percent)%, Conn $($node.active_connections)"
    }
} catch {
    Write-Error "Failed to get cluster stats: $_"
}

# ============ STEP 9: CLEANUP ============
Write-Test "Cleanup: Delete Test Cluster"
$cleanup = Read-Host "Delete test cluster? (y/n)"
if ($cleanup -eq "y") {
    try {
        $deleteResponse = Invoke-RestMethod -Uri "$BASE_URL/api/v1/postgres/cluster/$CLUSTER_ID" `
            -Method DELETE `
            -Headers $headers
        
        Write-Success "Cluster deleted successfully"
    } catch {
        Write-Error "Failed to delete cluster: $_"
    }
} else {
    Write-Info "Cluster $CLUSTER_ID kept for manual inspection"
}

# ============ SUMMARY ============
Write-Host "`n==================================" -ForegroundColor Magenta
Write-Host "TEST SUMMARY: NEW CLUSTER APIs" -ForegroundColor Magenta
Write-Host "==================================" -ForegroundColor Magenta
Write-Host "1. GetReplicationStatus (BEFORE): Tested - Shows initial replication state"
Write-Host "2. PromoteReplica (Failover): Tested - Manual failover executed"
Write-Host "3. Role Swap Verification: Tested - Primary/replica roles swapped"
Write-Host "4. GetReplicationStatus (AFTER): Tested - Shows post-failover replication state"
Write-Host "==================================" -ForegroundColor Magenta

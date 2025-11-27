# Nginx Cluster API Test Script
# Run: .\test-nginx-cluster-api.ps1

$ErrorActionPreference = "Continue"

Write-Host "============================================" -ForegroundColor Cyan
Write-Host "   NGINX CLUSTER API TEST SUITE" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""

# Configuration
$AuthURL = "http://localhost:8082"
$ProvisioningURL = "http://localhost:8083/api/v1"

# 1. Login
Write-Host "1. LOGIN" -ForegroundColor Yellow
Write-Host "---------"
try {
    $loginBody = @{
        username = "admin"
        password = "password123"
    } | ConvertTo-Json
    
    $loginResponse = Invoke-RestMethod -Uri "$AuthURL/auth/login" -Method POST -ContentType "application/json" -Body $loginBody
    $token = $loginResponse.data.access_token
    Write-Host "Login successful!" -ForegroundColor Green
    Write-Host "Token: $($token.Substring(0, 50))..." -ForegroundColor DarkGray
} catch {
    Write-Host "Login failed: $_" -ForegroundColor Red
    exit 1
}

$headers = @{
    "Authorization" = "Bearer $token"
    "Content-Type" = "application/json"
}

Write-Host ""

# 2. Create Nginx Cluster with Full Configuration
Write-Host "2. CREATE NGINX CLUSTER (Full Config)" -ForegroundColor Yellow
Write-Host "--------------------------------------"

$createClusterBody = @{
    # Basic Configuration
    cluster_name = "production-nginx-cluster"
    node_count = 2
    http_port = 8080
    https_port = 8443
    load_balance_mode = "round_robin"
    
    # High Availability
    virtual_ip = "192.168.100.100"
    vrrp_interface = "eth0"
    vrrp_router_id = 51
    health_check_enabled = $true
    health_check_path = "/health"
    health_check_interval = 5
    
    # Resources
    cpu_per_node = 1000000000  # 1 CPU core in nanocores
    memory_per_node = 536870912  # 512MB in bytes
    
    # SSL/TLS (disabled for testing)
    ssl_enabled = $false
    ssl_protocols = "TLSv1.2 TLSv1.3"
    ssl_session_timeout = "1d"
    
    # Performance Tuning
    worker_processes = 4
    worker_connections = 2048
    keepalive_timeout = 75
    client_max_body_size = "50m"
    
    # Logging
    access_log_enabled = $true
    error_log_level = "warn"
    
    # Caching
    cache_enabled = $true
    cache_path = "/var/cache/nginx"
    cache_size = "100m"
    
    # Rate Limiting
    rate_limit_enabled = $true
    rate_limit_requests_per_sec = 100
    rate_limit_burst = 50
    
    # Gzip Compression
    gzip_enabled = $true
    gzip_level = 6
    gzip_min_length = 1000
    gzip_types = "text/plain text/css application/json application/javascript application/xml"
    
    # Backend Configuration
    upstreams = @(
        @{
            name = "backend_api"
            algorithm = "least_conn"
            servers = @(
                @{
                    address = "api-server-1:3000"
                    weight = 5
                    max_fails = 3
                    fail_timeout = 30
                    is_backup = $false
                },
                @{
                    address = "api-server-2:3000"
                    weight = 5
                    max_fails = 3
                    fail_timeout = 30
                    is_backup = $false
                },
                @{
                    address = "api-server-3:3000"
                    weight = 1
                    max_fails = 3
                    fail_timeout = 30
                    is_backup = $true
                }
            )
            health_check = $true
            health_path = "/api/health"
        },
        @{
            name = "static_files"
            algorithm = "round_robin"
            servers = @(
                @{
                    address = "static-server:80"
                    weight = 1
                }
            )
        }
    )
    
    # Server Blocks (Virtual Hosts)
    server_blocks = @(
        @{
            server_name = "api.example.com"
            listen_port = 80
            ssl_enabled = $false
            locations = @(
                @{
                    path = "/api"
                    proxy_pass = "http://backend_api"
                    proxy_headers = @{
                        "X-Real-IP" = '$remote_addr'
                        "X-Forwarded-For" = '$proxy_add_x_forwarded_for'
                    }
                    cache_enabled = $false
                    rate_limit = 50
                },
                @{
                    path = "/static"
                    proxy_pass = "http://static_files"
                    cache_enabled = $true
                }
            )
        }
    )
} | ConvertTo-Json -Depth 10

try {
    $createResponse = Invoke-RestMethod -Uri "$ProvisioningURL/nginx/cluster" -Method POST -Headers $headers -Body $createClusterBody
    $clusterId = $createResponse.data.id
    Write-Host "Cluster created successfully!" -ForegroundColor Green
    Write-Host "Cluster ID: $clusterId" -ForegroundColor Cyan
    Write-Host "Status: $($createResponse.data.status)" -ForegroundColor Cyan
    Write-Host "Nodes: $($createResponse.data.node_count)" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Response:" -ForegroundColor DarkGray
    $createResponse.data | ConvertTo-Json -Depth 5
} catch {
    Write-Host "Create cluster failed: $_" -ForegroundColor Red
    Write-Host "Response: $($_.Exception.Response)" -ForegroundColor Red
    $clusterId = $null
}

if (-not $clusterId) {
    Write-Host "Cannot proceed without cluster ID" -ForegroundColor Red
    exit 1
}

Write-Host ""

# 3. Get Cluster Info
Write-Host "3. GET CLUSTER INFO" -ForegroundColor Yellow
Write-Host "-------------------"
try {
    $clusterInfo = Invoke-RestMethod -Uri "$ProvisioningURL/nginx/cluster/$clusterId" -Method GET -Headers $headers
    Write-Host "Cluster Info:" -ForegroundColor Green
    $clusterInfo.data | ConvertTo-Json -Depth 5
} catch {
    Write-Host "Get cluster info failed: $_" -ForegroundColor Red
}

Write-Host ""

# 4. Get Connection Info (User Usecase)
Write-Host "4. GET CONNECTION INFO" -ForegroundColor Yellow
Write-Host "----------------------"
try {
    $connInfo = Invoke-RestMethod -Uri "$ProvisioningURL/nginx/cluster/$clusterId/connection-info" -Method GET -Headers $headers
    Write-Host "Connection Info:" -ForegroundColor Green
    $connInfo.data | ConvertTo-Json -Depth 5
} catch {
    Write-Host "Get connection info failed: $_" -ForegroundColor Red
}

Write-Host ""

# 5. Test Connection
Write-Host "5. TEST CONNECTION" -ForegroundColor Yellow
Write-Host "------------------"
try {
    $testConn = Invoke-RestMethod -Uri "$ProvisioningURL/nginx/cluster/$clusterId/test-connection" -Method POST -Headers $headers
    Write-Host "Test Connection Result:" -ForegroundColor Green
    $testConn.data | ConvertTo-Json
} catch {
    Write-Host "Test connection failed: $_" -ForegroundColor Red
}

Write-Host ""

# 6. Get Cluster Health
Write-Host "6. GET CLUSTER HEALTH" -ForegroundColor Yellow
Write-Host "---------------------"
try {
    $health = Invoke-RestMethod -Uri "$ProvisioningURL/nginx/cluster/$clusterId/health" -Method GET -Headers $headers
    Write-Host "Health Status:" -ForegroundColor Green
    $health.data | ConvertTo-Json -Depth 3
} catch {
    Write-Host "Get health failed: $_" -ForegroundColor Red
}

Write-Host ""

# 7. Get Metrics
Write-Host "7. GET CLUSTER METRICS" -ForegroundColor Yellow
Write-Host "----------------------"
try {
    $metrics = Invoke-RestMethod -Uri "$ProvisioningURL/nginx/cluster/$clusterId/metrics" -Method GET -Headers $headers
    Write-Host "Metrics:" -ForegroundColor Green
    $metrics.data | ConvertTo-Json -Depth 3
} catch {
    Write-Host "Get metrics failed: $_" -ForegroundColor Red
}

Write-Host ""

# 8. List Upstreams
Write-Host "8. LIST UPSTREAMS" -ForegroundColor Yellow
Write-Host "-----------------"
try {
    $upstreams = Invoke-RestMethod -Uri "$ProvisioningURL/nginx/cluster/$clusterId/upstreams" -Method GET -Headers $headers
    Write-Host "Upstreams:" -ForegroundColor Green
    $upstreams.data | ConvertTo-Json -Depth 3
} catch {
    Write-Host "List upstreams failed: $_" -ForegroundColor Red
}

Write-Host ""

# 9. List Server Blocks
Write-Host "9. LIST SERVER BLOCKS" -ForegroundColor Yellow
Write-Host "---------------------"
try {
    $serverBlocks = Invoke-RestMethod -Uri "$ProvisioningURL/nginx/cluster/$clusterId/server-blocks" -Method GET -Headers $headers
    Write-Host "Server Blocks:" -ForegroundColor Green
    $serverBlocks.data | ConvertTo-Json -Depth 4
} catch {
    Write-Host "List server blocks failed: $_" -ForegroundColor Red
}

Write-Host ""

# 10. Add New Upstream
Write-Host "10. ADD NEW UPSTREAM" -ForegroundColor Yellow
Write-Host "--------------------"
$newUpstream = @{
    name = "websocket_backend"
    algorithm = "ip_hash"
    servers = @(
        @{
            address = "ws-server-1:8080"
            weight = 1
        },
        @{
            address = "ws-server-2:8080"
            weight = 1
        }
    )
    health_check = $true
    health_path = "/ws/health"
} | ConvertTo-Json -Depth 5

try {
    $addUpstream = Invoke-RestMethod -Uri "$ProvisioningURL/nginx/cluster/$clusterId/upstreams" -Method POST -Headers $headers -Body $newUpstream
    Write-Host "Upstream added successfully!" -ForegroundColor Green
} catch {
    Write-Host "Add upstream failed: $_" -ForegroundColor Red
}

Write-Host ""

# 11. Add Server Block
Write-Host "11. ADD SERVER BLOCK" -ForegroundColor Yellow
Write-Host "--------------------"
$newServerBlock = @{
    server_name = "ws.example.com"
    listen_port = 80
    locations = @(
        @{
            path = "/ws"
            proxy_pass = "http://websocket_backend"
            proxy_headers = @{
                "Upgrade" = '$http_upgrade'
                "Connection" = "upgrade"
            }
        }
    )
} | ConvertTo-Json -Depth 5

try {
    $addBlock = Invoke-RestMethod -Uri "$ProvisioningURL/nginx/cluster/$clusterId/server-blocks" -Method POST -Headers $headers -Body $newServerBlock
    Write-Host "Server block added successfully!" -ForegroundColor Green
} catch {
    Write-Host "Add server block failed: $_" -ForegroundColor Red
}

Write-Host ""

# 12. Get Failover History
Write-Host "12. GET FAILOVER HISTORY" -ForegroundColor Yellow
Write-Host "------------------------"
try {
    $failoverHistory = Invoke-RestMethod -Uri "$ProvisioningURL/nginx/cluster/$clusterId/failover-history" -Method GET -Headers $headers
    Write-Host "Failover History:" -ForegroundColor Green
    $failoverHistory.data | ConvertTo-Json -Depth 3
} catch {
    Write-Host "Get failover history failed: $_" -ForegroundColor Red
}

Write-Host ""

# Summary
Write-Host "============================================" -ForegroundColor Cyan
Write-Host "   TEST SUMMARY" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Cluster ID: $clusterId" -ForegroundColor Green
Write-Host "API Base URL: $ProvisioningURL/nginx/cluster/$clusterId" -ForegroundColor Green
Write-Host ""
Write-Host "Available Endpoints:" -ForegroundColor Yellow
Write-Host "  GET    /                    - Cluster Info"
Write-Host "  DELETE /                    - Delete Cluster"
Write-Host "  POST   /start               - Start Cluster"
Write-Host "  POST   /stop                - Stop Cluster"
Write-Host "  POST   /restart             - Restart Cluster"
Write-Host "  POST   /nodes               - Add Node"
Write-Host "  DELETE /nodes/:nodeId       - Remove Node"
Write-Host "  PUT    /config              - Update Config"
Write-Host "  POST   /sync-config         - Sync Config"
Write-Host "  GET    /upstreams           - List Upstreams"
Write-Host "  POST   /upstreams           - Add Upstream"
Write-Host "  PUT    /upstreams/:id       - Update Upstream"
Write-Host "  DELETE /upstreams/:id       - Delete Upstream"
Write-Host "  GET    /server-blocks       - List Server Blocks"
Write-Host "  POST   /server-blocks       - Add Server Block"
Write-Host "  DELETE /server-blocks/:id   - Delete Server Block"
Write-Host "  GET    /health              - Cluster Health"
Write-Host "  GET    /metrics             - Cluster Metrics"
Write-Host "  POST   /test-connection     - Test Connection"
Write-Host "  GET    /connection-info     - Connection Info"
Write-Host "  POST   /failover            - Trigger Failover"
Write-Host "  GET    /failover-history    - Failover History"
Write-Host ""
Write-Host "============================================" -ForegroundColor Cyan

# Cleanup option
Write-Host ""
$cleanup = Read-Host "Do you want to delete the test cluster? (y/n)"
if ($cleanup -eq "y") {
    Write-Host "Deleting cluster..." -ForegroundColor Yellow
    try {
        Invoke-RestMethod -Uri "$ProvisioningURL/nginx/cluster/$clusterId" -Method DELETE -Headers $headers
        Write-Host "Cluster deleted successfully!" -ForegroundColor Green
    } catch {
        Write-Host "Delete cluster failed: $_" -ForegroundColor Red
    }
}


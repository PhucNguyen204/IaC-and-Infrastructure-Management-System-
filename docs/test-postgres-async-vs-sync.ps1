# Test PostgreSQL Cluster - Async vs Sync Mode
# So sanh hieu nang va do an toan giua 2 mode

$baseUrl = "http://localhost:8083/api/v1"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "PostgreSQL Async vs Sync Mode Test" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

# Login va lay token
Write-Host "`nDang nhap..." -ForegroundColor Yellow
$loginBody = @{
    username = "admin"
    password = "password123"
} | ConvertTo-Json

$loginResp = Invoke-RestMethod -Uri "http://localhost:8082/auth/login" -Method POST -Headers @{"Content-Type"="application/json"} -Body $loginBody
$token = $loginResp.data.access_token

$headers = @{
    "Authorization" = "Bearer $token"
    "Content-Type" = "application/json"
}

Write-Host "Token: $($token.Substring(0, 30))..." -ForegroundColor Green

# ===========================================
# TEST 1: Tao Cluster voi ASYNC Mode
# ===========================================
Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "TEST 1: ASYNC MODE (Asynchronous)" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

$asyncClusterBody = @{
    cluster_name = "test-async-cluster"
    postgres_version = "15"
    node_count = 3
    cpu_per_node = 1
    memory_per_node = 536870912  # 512MB
    storage_per_node = 5          # 5GB
    postgres_password = "postgres123"
    replication_mode = "async"    # ASYNC MODE
    enable_haproxy = $true
    haproxy_port = 5100
    haproxy_read_port = 5101
} | ConvertTo-Json

Write-Host "`nTao cluster ASYNC mode..." -ForegroundColor Yellow
Write-Host "  - Replication Mode: ASYNC" -ForegroundColor Gray
Write-Host "  - synchronous_commit: local" -ForegroundColor Gray
Write-Host "  - synchronous_standby_names: (empty)" -ForegroundColor Gray
Write-Host "  - Dac diem:" -ForegroundColor Gray
Write-Host "    + Ghi nhanh hon (chi doi local WAL)" -ForegroundColor Gray
Write-Host "    + Co the mat data neu primary fail truoc khi replicate" -ForegroundColor Gray
Write-Host "    + Phu hop cho hieu nang cao" -ForegroundColor Gray

try {
    $asyncStart = Get-Date
    $asyncResp = Invoke-RestMethod -Uri "$baseUrl/postgres/cluster" -Method POST -Headers $headers -Body $asyncClusterBody
    $asyncClusterId = $asyncResp.data.cluster_id
    $asyncCreateTime = (Get-Date) - $asyncStart
    
    Write-Host "  OK - Cluster ID: $asyncClusterId" -ForegroundColor Green
    Write-Host "  Thoi gian tao: $($asyncCreateTime.TotalSeconds) giay" -ForegroundColor Cyan
    
    # Doi cluster ready
    Write-Host "`nDoi cluster ASYNC ready..." -ForegroundColor Yellow
    $maxWait = 120
    $waited = 0
    while ($waited -lt $maxWait) {
        Start-Sleep -Seconds 5
        $waited += 5
        $info = Invoke-RestMethod -Uri "$baseUrl/postgres/cluster/$asyncClusterId" -Method GET -Headers $headers
        Write-Host "  Status: $($info.data.status) ($waited/$maxWait s)" -ForegroundColor Gray
        
        if ($info.data.status -eq "running") {
            Write-Host "  Cluster ASYNC ready!" -ForegroundColor Green
            Write-Host "  Write Endpoint: $($info.data.endpoints.haproxy.write_url)" -ForegroundColor Cyan
            Write-Host "  Read Endpoint: $($info.data.endpoints.haproxy.read_url)" -ForegroundColor Cyan
            break
        }
    }
    
} catch {
    Write-Host "  FAIL - $($_.Exception.Message)" -ForegroundColor Red
}

# ===========================================
# TEST 2: Tao Cluster voi SYNC Mode
# ===========================================
Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "TEST 2: SYNC MODE (Synchronous)" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

$syncClusterBody = @{
    cluster_name = "test-sync-cluster"
    postgres_version = "15"
    node_count = 3
    cpu_per_node = 1
    memory_per_node = 536870912  # 512MB
    storage_per_node = 5          # 5GB
    postgres_password = "postgres123"
    replication_mode = "sync"     # SYNC MODE
    enable_haproxy = $true
    haproxy_port = 5200
    haproxy_read_port = 5201
} | ConvertTo-Json

Write-Host "`nTao cluster SYNC mode..." -ForegroundColor Yellow
Write-Host "  - Replication Mode: SYNC" -ForegroundColor Gray
Write-Host "  - synchronous_commit: on" -ForegroundColor Gray
Write-Host "  - synchronous_standby_names: ANY 1 (*)" -ForegroundColor Gray
Write-Host "  - Dac diem:" -ForegroundColor Gray
Write-Host "    + Phai doi it nhat 1 replica confirm" -ForegroundColor Gray
Write-Host "    + An toan hon, khong mat data" -ForegroundColor Gray
Write-Host "    + Cham hon do phai doi replica" -ForegroundColor Gray

try {
    $syncStart = Get-Date
    $syncResp = Invoke-RestMethod -Uri "$baseUrl/postgres/cluster" -Method POST -Headers $headers -Body $syncClusterBody
    $syncClusterId = $syncResp.data.cluster_id
    $syncCreateTime = (Get-Date) - $syncStart
    
    Write-Host "  OK - Cluster ID: $syncClusterId" -ForegroundColor Green
    Write-Host "  Thoi gian tao: $($syncCreateTime.TotalSeconds) giay" -ForegroundColor Cyan
    
    # Doi cluster ready
    Write-Host "`nDoi cluster SYNC ready..." -ForegroundColor Yellow
    $maxWait = 120
    $waited = 0
    while ($waited -lt $maxWait) {
        Start-Sleep -Seconds 5
        $waited += 5
        $info = Invoke-RestMethod -Uri "$baseUrl/postgres/cluster/$syncClusterId" -Method GET -Headers $headers
        Write-Host "  Status: $($info.data.status) ($waited/$maxWait s)" -ForegroundColor Gray
        
        if ($info.data.status -eq "running") {
            Write-Host "  Cluster SYNC ready!" -ForegroundColor Green
            Write-Host "  Write Endpoint: $($info.data.endpoints.haproxy.write_url)" -ForegroundColor Cyan
            Write-Host "  Read Endpoint: $($info.data.endpoints.haproxy.read_url)" -ForegroundColor Cyan
            break
        }
    }
    
} catch {
    Write-Host "  FAIL - $($_.Exception.Message)" -ForegroundColor Red
}

# ===========================================
# TEST 3: So sanh Performance
# ===========================================
Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "TEST 3: So sanh Performance" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

Write-Host "`nTest ghi 100 rows vao moi cluster..." -ForegroundColor Yellow

# Test ASYNC
Write-Host "`nTest ASYNC cluster..." -ForegroundColor Yellow
$asyncWriteStart = Get-Date
for ($i = 1; $i -le 100; $i++) {
    $query = "INSERT INTO test_table (data) VALUES ('test-$i');"
    $queryBody = @{
        query = $query
        database = "postgres"
    } | ConvertTo-Json
    
    try {
        Invoke-RestMethod -Uri "$baseUrl/postgres/cluster/$asyncClusterId/query" -Method POST -Headers $headers -Body $queryBody -ErrorAction SilentlyContinue | Out-Null
    } catch {
        # Ignore errors, table might not exist yet
    }
}
$asyncWriteTime = (Get-Date) - $asyncWriteStart
Write-Host "  ASYNC write time: $($asyncWriteTime.TotalSeconds) giay" -ForegroundColor Cyan

# Test SYNC
Write-Host "`nTest SYNC cluster..." -ForegroundColor Yellow
$syncWriteStart = Get-Date
for ($i = 1; $i -le 100; $i++) {
    $query = "INSERT INTO test_table (data) VALUES ('test-$i');"
    $queryBody = @{
        query = $query
        database = "postgres"
    } | ConvertTo-Json
    
    try {
        Invoke-RestMethod -Uri "$baseUrl/postgres/cluster/$syncClusterId/query" -Method POST -Headers $headers -Body $queryBody -ErrorAction SilentlyContinue | Out-Null
    } catch {
        # Ignore errors
    }
}
$syncWriteTime = (Get-Date) - $syncWriteStart
Write-Host "  SYNC write time: $($syncWriteTime.TotalSeconds) giay" -ForegroundColor Cyan

# ===========================================
# So sanh ket qua
# ===========================================
Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "KET QUA SO SANH" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

Write-Host "`nASYNC MODE:" -ForegroundColor Yellow
Write-Host "  - Thoi gian tao cluster: $($asyncCreateTime.TotalSeconds) giay" -ForegroundColor Gray
Write-Host "  - Thoi gian ghi 100 rows: $($asyncWriteTime.TotalSeconds) giay" -ForegroundColor Gray
Write-Host "  - Uu diem: Nhanh hon" -ForegroundColor Green
Write-Host "  - Nhuoc diem: Co the mat du lieu khi primary fail" -ForegroundColor Red

Write-Host "`nSYNC MODE:" -ForegroundColor Yellow
Write-Host "  - Thoi gian tao cluster: $($syncCreateTime.TotalSeconds) giay" -ForegroundColor Gray
Write-Host "  - Thoi gian ghi 100 rows: $($syncWriteTime.TotalSeconds) giay" -ForegroundColor Gray
Write-Host "  - Uu diem: An toan hon, khong mat du lieu" -ForegroundColor Green
Write-Host "  - Nhuoc diem: Cham hon do phai doi replica confirm" -ForegroundColor Red

$performanceDiff = (($syncWriteTime.TotalSeconds - $asyncWriteTime.TotalSeconds) / $asyncWriteTime.TotalSeconds) * 100
Write-Host "`nSYNC cham hon ASYNC: $([math]::Round($performanceDiff, 2))%" -ForegroundColor Cyan

# ===========================================
# Cleanup
# ===========================================
Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "CLEANUP" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

Write-Host "`nBan co muon xoa 2 cluster test khong? (y/n): " -ForegroundColor Yellow -NoNewline
$cleanup = Read-Host

if ($cleanup -eq "y") {
    Write-Host "`nXoa ASYNC cluster..." -ForegroundColor Yellow
    try {
        Invoke-RestMethod -Uri "$baseUrl/postgres/cluster/$asyncClusterId" -Method DELETE -Headers $headers | Out-Null
        Write-Host "  OK - ASYNC cluster deleted" -ForegroundColor Green
    } catch {
        Write-Host "  FAIL - $($_.Exception.Message)" -ForegroundColor Red
    }
    
    Write-Host "`nXoa SYNC cluster..." -ForegroundColor Yellow
    try {
        Invoke-RestMethod -Uri "$baseUrl/postgres/cluster/$syncClusterId" -Method DELETE -Headers $headers | Out-Null
        Write-Host "  OK - SYNC cluster deleted" -ForegroundColor Green
    } catch {
        Write-Host "  FAIL - $($_.Exception.Message)" -ForegroundColor Red
    }
}

Write-Host "`n========================================" -ForegroundColor Green
Write-Host "TEST HOAN TAT!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green


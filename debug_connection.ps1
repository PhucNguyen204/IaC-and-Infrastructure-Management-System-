#!/usr/bin/env pwsh
$ErrorActionPreference = "Continue"

$AUTH_URL = "http://localhost:8082"
$INFRA_URL = "http://localhost:8083"
$USERNAME = "admin"
$PASSWORD = "password123"

# --- HELPER FUNCTIONS ---
function Get-AuthToken {
    $body = @{ username = $USERNAME; password = $PASSWORD } | ConvertTo-Json
    try {
        $response = Invoke-RestMethod -Uri "$AUTH_URL/auth/login" -Method POST -ContentType "application/json" -Body $body
        return $response.data.access_token
    } catch {
        Write-Host "Authentication failed: $($_.Exception.Message)" -ForegroundColor Red
        exit 1
    }
}

function Get-Headers {
    param($token)
    return @{ "Authorization" = "Bearer $token"; "Content-Type" = "application/json" }
}

function Show-Section {
    param($title)
    Write-Host "`n========================================================" -ForegroundColor Cyan
    Write-Host " $title" -ForegroundColor Cyan
    Write-Host "========================================================" -ForegroundColor Cyan
}

# --- MAIN ---

$token = Get-AuthToken
$headers = Get-Headers -token $token
Write-Host "Authenticated as $USERNAME" -ForegroundColor Green

# 1. NGINX
Show-Section "TESTING NGINX"
$nginxId = $null
try {
    Write-Host "Creating Nginx..." -ForegroundColor Yellow
    $body = @{ cluster_name = "debug-nginx"; node_count = 2; http_port = 8092 } | ConvertTo-Json
    $res = Invoke-RestMethod -Uri "$INFRA_URL/api/v1/nginx/cluster" -Method POST -Headers $headers -Body $body -TimeoutSec 180
    $nginxId = $res.data.id
    Write-Host "Created ID: $nginxId" -ForegroundColor Green

    $info = Invoke-RestMethod -Uri "$INFRA_URL/api/v1/nginx/cluster/$nginxId" -Method GET -Headers $headers
    $data = $info.data
    
    Write-Host "`n[CONNECTION INFO]" -ForegroundColor Cyan
    Write-Host "HTTP URL: $($data.endpoints.http_url)"
    Write-Host "HTTPS URL: $($data.endpoints.https_url)"
    
    Write-Host "`n[TEST CONNECTIVITY]" -ForegroundColor Cyan
    try {
        $url = $data.endpoints.http_url
        Write-Host "Requesting $url ... " -NoNewline
        $req = Invoke-WebRequest -Uri $url -Method HEAD -TimeoutSec 5
        if ($req.StatusCode -eq 200) { Write-Host "PASS (200 OK)" -ForegroundColor Green }
        else { Write-Host "FAIL ($($req.StatusCode))" -ForegroundColor Red }
    } catch {
        Write-Host "FAIL ($($_.Exception.Message))" -ForegroundColor Red
    }
} catch {
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
} finally {
    if ($nginxId) {
        Write-Host "Cleaning up Nginx..." -ForegroundColor Yellow
        try { Invoke-RestMethod -Uri "$INFRA_URL/api/v1/nginx/cluster/$nginxId" -Method DELETE -Headers $headers | Out-Null } catch {}
    }
}

# 2. POSTGRES
Show-Section "TESTING POSTGRES"
$pgId = $null
try {
    Write-Host "Creating Postgres..." -ForegroundColor Yellow
    $body = @{ 
        cluster_name = "debug-pg"; postgres_version = "15"; node_count = 1; 
        cpu_per_node = 1; memory_per_node = 512; storage_per_node = 1; 
        postgres_password = "password123"; replication_mode = "async" 
    } | ConvertTo-Json
    $res = Invoke-RestMethod -Uri "$INFRA_URL/api/v1/postgres/cluster" -Method POST -Headers $headers -Body $body -TimeoutSec 180
    $pgId = $res.cluster_id
    Write-Host "Created ID: $pgId" -ForegroundColor Green
    
    $info = Invoke-RestMethod -Uri "$INFRA_URL/api/v1/postgres/cluster/$pgId" -Method GET -Headers $headers
    
    Write-Host "`n[CONNECTION INFO]" -ForegroundColor Cyan
    if ($info.connection_info) {
        $c = $info.connection_info
        Write-Host "Host: $($c.host)"
        Write-Host "Port: $($c.port)"
        Write-Host "User: $($c.username)"
        Write-Host "DB:   $($c.database)"
        Write-Host "SSL:  $($c.ssl_mode)"
        
        Write-Host "`n[TEST CONNECTIVITY]" -ForegroundColor Cyan
        if ($c.host -eq "localhost") {
            try {
                Write-Host "Testing TCP localhost:$($c.port) ... " -NoNewline
                $tcp = Test-NetConnection -ComputerName "localhost" -Port $c.port -WarningAction SilentlyContinue
                if ($tcp.TcpTestSucceeded) { Write-Host "PASS" -ForegroundColor Green }
                else { Write-Host "FAIL" -ForegroundColor Red }
            } catch { Write-Host "ERROR ($($_.Exception.Message))" -ForegroundColor Red }
        } else {
             Write-Host "Skipping TCP test (remote host: $($c.host))" -ForegroundColor Yellow
        }
    } elseif ($info.write_endpoint) {
        $we = $info.write_endpoint
        Write-Host "Write Endpoint: $($we.host):$($we.port) (Legacy)"
        # ... (rest of legacy handling omitted for brevity, focusing on new field)
        Write-Host "Please use connection_info field." -ForegroundColor Yellow
    } else {
        Write-Host "No connection info." -ForegroundColor Red
    }
} catch {
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
} finally {
    if ($pgId) {
        Write-Host "Cleaning up Postgres..." -ForegroundColor Yellow
        try { Invoke-RestMethod -Uri "$INFRA_URL/api/v1/postgres/cluster/$pgId" -Method DELETE -Headers $headers | Out-Null } catch {}
    }
}

# 3. CLICKHOUSE
Show-Section "TESTING CLICKHOUSE"
$chId = $null
try {
    Write-Host "Creating ClickHouse..." -ForegroundColor Yellow
    $body = @{ cluster_name = "debug-ch"; password = "password123"; database = "default" } | ConvertTo-Json
    $res = Invoke-RestMethod -Uri "$INFRA_URL/api/v1/clickhouse" -Method POST -Headers $headers -Body $body -TimeoutSec 180
    $chId = $res.cluster_id
    Write-Host "Created ID: $chId" -ForegroundColor Green
    
    $info = Invoke-RestMethod -Uri "$INFRA_URL/api/v1/clickhouse/$chId" -Method GET -Headers $headers
    
    Write-Host "`n[CONNECTION INFO]" -ForegroundColor Cyan
    Write-Host "HTTP Endpoint: http://$($info.http_endpoint.host):$($info.http_endpoint.port)"
    Write-Host "Native Endpoint: tcp://$($info.native_endpoint.host):$($info.native_endpoint.port)"
    
    Write-Host "`n[TEST QUERY API]" -ForegroundColor Cyan
    try {
        Write-Host "Executing SELECT 1 ... " -NoNewline
        $qBody = @{ query = "SELECT 1" } | ConvertTo-Json
        $qRes = Invoke-RestMethod -Uri "$INFRA_URL/api/v1/clickhouse/$chId/query" -Method POST -Headers $headers -Body $qBody
        if ($qRes.success) { Write-Host "PASS (Data: $($qRes.data))" -ForegroundColor Green }
        else { Write-Host "FAIL" -ForegroundColor Red }
    } catch {
        Write-Host "FAIL ($($_.Exception.Message))" -ForegroundColor Red
    }
} catch {
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
} finally {
    if ($chId) {
        Write-Host "Cleaning up ClickHouse..." -ForegroundColor Yellow
        try { Invoke-RestMethod -Uri "$INFRA_URL/api/v1/clickhouse/$chId" -Method DELETE -Headers $headers | Out-Null } catch {}
    }
}

# 4. DinD
Show-Section "TESTING DinD"
$dindId = $null
try {
    Write-Host "Creating DinD..." -ForegroundColor Yellow
    $body = @{ name = "debug-dind"; resource_plan = "small" } | ConvertTo-Json
    $res = Invoke-RestMethod -Uri "$INFRA_URL/api/v1/dind/environments" -Method POST -Headers $headers -Body $body -TimeoutSec 180
    $dindId = $res.data.id
    Write-Host "Created ID: $dindId" -ForegroundColor Green
    
    $info = Invoke-RestMethod -Uri "$INFRA_URL/api/v1/dind/environments/$dindId" -Method GET -Headers $headers
    $data = $info.data
    
    Write-Host "`n[CONNECTION INFO]" -ForegroundColor Cyan
    Write-Host "Docker Host: $($data.docker_host)"
    
    Write-Host "`n[TEST EXEC API]" -ForegroundColor Cyan
    try {
        Write-Host "Executing 'docker -v' ... " -NoNewline
        $execBody = @{ command = "docker -v" } | ConvertTo-Json
        $execRes = Invoke-RestMethod -Uri "$INFRA_URL/api/v1/dind/environments/$dindId/exec" -Method POST -Headers $headers -Body $execBody
        if ($execRes.success) { 
            Write-Host "PASS" -ForegroundColor Green 
            Write-Host "Output: $($execRes.data.output)"
        } else { Write-Host "FAIL" -ForegroundColor Red }
    } catch {
        Write-Host "FAIL ($($_.Exception.Message))" -ForegroundColor Red
    }
} catch {
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
} finally {
    if ($dindId) {
        Write-Host "Cleaning up DinD..." -ForegroundColor Yellow
        try { Invoke-RestMethod -Uri "$INFRA_URL/api/v1/dind/environments/$dindId" -Method DELETE -Headers $headers | Out-Null } catch {}
    }
}

Write-Host "`nDone."

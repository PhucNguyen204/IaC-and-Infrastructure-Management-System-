# Test Docker-in-Docker (DinD) API
# Mo ta: User gui docker command, system chay trong DinD container isolated

$baseUrl = "http://localhost:8083/api/v1"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Test Docker-in-Docker (DinD) API" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

# Login
Write-Host "`n1. Dang nhap..." -ForegroundColor Yellow
$loginBody = @{
    username = "admin"
    password = "password123"
} | ConvertTo-Json

try {
    $loginResp = Invoke-RestMethod -Uri "http://localhost:8082/auth/login" -Method POST -Headers @{"Content-Type"="application/json"} -Body $loginBody
    $token = $loginResp.data.access_token
    Write-Host "   OK - Token: $($token.Substring(0, 30))..." -ForegroundColor Green
} catch {
    Write-Host "   FAIL - Khong the login: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

$headers = @{
    "Authorization" = "Bearer $token"
    "Content-Type" = "application/json"
}

# ===========================================
# TEST 1: Tao DinD Environment
# ===========================================
Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "TEST 1: Tao DinD Environment" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

$createEnvBody = @{
    name = "my-docker-sandbox"
    resource_plan = "medium"
    description = "Sandbox de chay docker commands"
    auto_cleanup = $false
} | ConvertTo-Json

Write-Host "`nTao environment moi..." -ForegroundColor Yellow
Write-Host "  - Name: my-docker-sandbox" -ForegroundColor Gray
Write-Host "  - Plan: medium (2 CPU, 2GB RAM)" -ForegroundColor Gray

try {
    $createResp = Invoke-RestMethod -Uri "$baseUrl/dind/environments" -Method POST -Headers $headers -Body $createEnvBody
    $envId = $createResp.data.id
    Write-Host "   OK - Environment ID: $envId" -ForegroundColor Green
    Write-Host "   Status: $($createResp.data.status)" -ForegroundColor Cyan
    Write-Host "   Docker Host: $($createResp.data.docker_host)" -ForegroundColor Cyan
} catch {
    Write-Host "   FAIL - $($_.Exception.Message)" -ForegroundColor Red
    if ($_.ErrorDetails.Message) {
        Write-Host "   Details: $($_.ErrorDetails.Message)" -ForegroundColor Red
    }
    exit 1
}

# Doi environment ready
Write-Host "`nDoi environment san sang..." -ForegroundColor Yellow
Start-Sleep -Seconds 5

# ===========================================
# TEST 2: Chay Docker Command - Pull Image
# ===========================================
Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "TEST 2: Pull nginx image" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

$pullBody = @{
    image = "nginx:alpine"
} | ConvertTo-Json

Write-Host "`nPull nginx:alpine..." -ForegroundColor Yellow
try {
    $pullResp = Invoke-RestMethod -Uri "$baseUrl/dind/environments/$envId/pull" -Method POST -Headers $headers -Body $pullBody
    Write-Host "   OK - Image: $($pullResp.data.image)" -ForegroundColor Green
    Write-Host "   Status: $($pullResp.data.status)" -ForegroundColor Cyan
} catch {
    Write-Host "   FAIL - $($_.Exception.Message)" -ForegroundColor Red
}

# ===========================================
# TEST 3: Chay Docker Command - docker run
# ===========================================
Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "TEST 3: Chay docker run nginx" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

$execBody = @{
    command = "docker run -d --name my-nginx nginx:alpine"
} | ConvertTo-Json

Write-Host "`nChay: docker run -d --name my-nginx nginx:alpine" -ForegroundColor Yellow
try {
    $execResp = Invoke-RestMethod -Uri "$baseUrl/dind/environments/$envId/exec" -Method POST -Headers $headers -Body $execBody
    Write-Host "   OK - Command executed" -ForegroundColor Green
    Write-Host "   Output: $($execResp.data.output)" -ForegroundColor Cyan
    Write-Host "   Exit Code: $($execResp.data.exit_code)" -ForegroundColor Cyan
    Write-Host "   Duration: $($execResp.data.duration)" -ForegroundColor Gray
} catch {
    Write-Host "   FAIL - $($_.Exception.Message)" -ForegroundColor Red
}

# ===========================================
# TEST 4: List containers trong DinD
# ===========================================
Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "TEST 4: List containers trong DinD" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

Write-Host "`nList containers..." -ForegroundColor Yellow
try {
    $containersResp = Invoke-RestMethod -Uri "$baseUrl/dind/environments/$envId/containers" -Method GET -Headers $headers
    Write-Host "   OK - Total: $($containersResp.data.total) containers" -ForegroundColor Green
    foreach ($c in $containersResp.data.containers) {
        Write-Host "   - $($c.name): $($c.status)" -ForegroundColor Cyan
    }
} catch {
    Write-Host "   FAIL - $($_.Exception.Message)" -ForegroundColor Red
}

# ===========================================
# TEST 5: List images trong DinD
# ===========================================
Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "TEST 5: List images trong DinD" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

Write-Host "`nList images..." -ForegroundColor Yellow
try {
    $imagesResp = Invoke-RestMethod -Uri "$baseUrl/dind/environments/$envId/images" -Method GET -Headers $headers
    Write-Host "   OK - Total: $($imagesResp.data.total) images" -ForegroundColor Green
    foreach ($img in $imagesResp.data.images) {
        Write-Host "   - $($img.repository):$($img.tag) ($($img.size))" -ForegroundColor Cyan
    }
} catch {
    Write-Host "   FAIL - $($_.Exception.Message)" -ForegroundColor Red
}

# ===========================================
# TEST 6: Build Docker Image
# ===========================================
Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "TEST 6: Build custom Docker image" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

$buildBody = @{
    dockerfile = @"
FROM alpine:latest
RUN apk add --no-cache curl
CMD ["echo", "Hello from DinD!"]
"@
    image_name = "my-custom-app"
    tag = "v1"
} | ConvertTo-Json

Write-Host "`nBuild image: my-custom-app:v1" -ForegroundColor Yellow
try {
    $buildResp = Invoke-RestMethod -Uri "$baseUrl/dind/environments/$envId/build" -Method POST -Headers $headers -Body $buildBody
    Write-Host "   OK - Image built: $($buildResp.data.image_name):$($buildResp.data.tag)" -ForegroundColor Green
    Write-Host "   Duration: $($buildResp.data.duration)" -ForegroundColor Cyan
    Write-Host "   Success: $($buildResp.data.success)" -ForegroundColor Cyan
} catch {
    Write-Host "   FAIL - $($_.Exception.Message)" -ForegroundColor Red
}

# ===========================================
# TEST 7: Chay image vua build
# ===========================================
Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "TEST 7: Chay image vua build" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

$runCustomBody = @{
    command = "docker run --rm my-custom-app:v1"
} | ConvertTo-Json

Write-Host "`nChay: docker run --rm my-custom-app:v1" -ForegroundColor Yellow
try {
    $runResp = Invoke-RestMethod -Uri "$baseUrl/dind/environments/$envId/exec" -Method POST -Headers $headers -Body $runCustomBody
    Write-Host "   OK - Output: $($runResp.data.output)" -ForegroundColor Green
} catch {
    Write-Host "   FAIL - $($_.Exception.Message)" -ForegroundColor Red
}

# ===========================================
# TEST 8: Docker Compose
# ===========================================
Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "TEST 8: Docker Compose" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

$composeBody = @{
    compose_content = @"
version: '3'
services:
  web:
    image: nginx:alpine
    ports:
      - "8080:80"
  redis:
    image: redis:alpine
"@
    action = "up"
    detach = $true
} | ConvertTo-Json

Write-Host "`nChay docker-compose up -d..." -ForegroundColor Yellow
try {
    $composeResp = Invoke-RestMethod -Uri "$baseUrl/dind/environments/$envId/compose" -Method POST -Headers $headers -Body $composeBody
    Write-Host "   OK - Services: $($composeResp.data.services -join ', ')" -ForegroundColor Green
    Write-Host "   Success: $($composeResp.data.success)" -ForegroundColor Cyan
} catch {
    Write-Host "   FAIL - $($_.Exception.Message)" -ForegroundColor Red
}

# ===========================================
# TEST 9: Get Stats
# ===========================================
Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "TEST 9: Get Environment Stats" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

Write-Host "`nGet stats..." -ForegroundColor Yellow
try {
    $statsResp = Invoke-RestMethod -Uri "$baseUrl/dind/environments/$envId/stats" -Method GET -Headers $headers
    Write-Host "   OK - Stats:" -ForegroundColor Green
    Write-Host "   - Containers: $($statsResp.data.container_count)" -ForegroundColor Cyan
    Write-Host "   - Images: $($statsResp.data.image_count)" -ForegroundColor Cyan
    Write-Host "   - Memory: $([math]::Round($statsResp.data.memory_usage_percent, 2))%" -ForegroundColor Cyan
} catch {
    Write-Host "   FAIL - $($_.Exception.Message)" -ForegroundColor Red
}

# ===========================================
# TEST 10: List all environments
# ===========================================
Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "TEST 10: List all environments" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

Write-Host "`nList environments..." -ForegroundColor Yellow
try {
    $listResp = Invoke-RestMethod -Uri "$baseUrl/dind/environments" -Method GET -Headers $headers
    Write-Host "   OK - Total: $($listResp.data.Count) environments" -ForegroundColor Green
    foreach ($env in $listResp.data) {
        Write-Host "   - $($env.name): $($env.status) (plan: $($env.resource_plan))" -ForegroundColor Cyan
    }
} catch {
    Write-Host "   FAIL - $($_.Exception.Message)" -ForegroundColor Red
}

# ===========================================
# CLEANUP
# ===========================================
Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "CLEANUP" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

Write-Host "`nBan co muon xoa environment test khong? (y/n): " -ForegroundColor Yellow -NoNewline
$cleanup = Read-Host

if ($cleanup -eq "y") {
    Write-Host "`nXoa environment..." -ForegroundColor Yellow
    try {
        Invoke-RestMethod -Uri "$baseUrl/dind/environments/$envId" -Method DELETE -Headers $headers | Out-Null
        Write-Host "   OK - Environment deleted" -ForegroundColor Green
    } catch {
        Write-Host "   FAIL - $($_.Exception.Message)" -ForegroundColor Red
    }
}

Write-Host "`n========================================" -ForegroundColor Green
Write-Host "TEST HOAN TAT!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green

Write-Host "`nUse Case Summary:" -ForegroundColor Cyan
Write-Host "  1. User tao DinD environment (isolated Docker sandbox)" -ForegroundColor Gray
Write-Host "  2. User gui docker commands qua API" -ForegroundColor Gray
Write-Host "  3. Commands chay trong DinD container (docker in docker)" -ForegroundColor Gray
Write-Host "  4. User co the build images, run containers, docker-compose" -ForegroundColor Gray
Write-Host "  5. Moi user co environment rieng, isolated" -ForegroundColor Gray


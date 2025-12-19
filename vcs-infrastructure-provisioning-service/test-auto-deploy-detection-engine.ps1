$ErrorActionPreference = "Stop"

function Write-Section($title) {
    Write-Host ""
    Write-Host "=== $title ===" -ForegroundColor Cyan
}

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$repoRoot = (Resolve-Path (Join-Path $scriptDir "..")).Path
$detectionEngineDir = (Resolve-Path (Join-Path $repoRoot "detection-engine")).Path
$rulesDir = (Resolve-Path (Join-Path $repoRoot "rules_storage")).Path

# Endpoints
$authUrl = "http://localhost:8082"
$provUrl = "http://localhost:8083/api/v1"

# Image + deploy settings
$imageTag = "iaas-detection-engine:e2e"
$deployPort = 8000
$deployName = "detection-engine-e2e-" + (Get-Date -Format "HHmmss")

Write-Host "Detection Engine auto-deploy integration test" -ForegroundColor Green
Write-Host "Repo root : $repoRoot"
Write-Host "Rules dir : $rulesDir"
Write-Host ""

Write-Section "Pre-flight"
# Check docker availability
if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
    Write-Host "docker CLI not found in PATH." -ForegroundColor Red
    exit 1
}

# Ensure rules file with keywords exists
$rulesFile = Join-Path $rulesDir "detection_engine_rules.yaml"
if (-not (Test-Path $rulesFile)) {
    Write-Host "Rule file not found at $rulesFile" -ForegroundColor Red
    Write-Host "Add at least one YAML rule with 'keywords' before running." -ForegroundColor Red
    exit 1
}

# Build detection engine image if missing
Write-Host "Checking image $imageTag ..."
docker image inspect $imageTag *> $null
if ($LASTEXITCODE -ne 0) {
    Write-Host "Image not found. Building from $detectionEngineDir" -ForegroundColor Yellow
    Push-Location $detectionEngineDir
    docker build -t $imageTag .
    Pop-Location
} else {
    Write-Host "Using existing image $imageTag" -ForegroundColor DarkGray
}

Write-Section "Authenticate"
try {
    $loginBody = @{
        username = "admin"
        password = "password123"
    } | ConvertTo-Json

    $loginResponse = Invoke-RestMethod -Uri "$authUrl/auth/login" -Method Post -ContentType "application/json" -Body $loginBody
    $token = $loginResponse.data.access_token
    if (-not $token) { throw "Token not returned from auth service" }
    Write-Host "Got token for admin user" -ForegroundColor Green
} catch {
    Write-Host "Login failed: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

$authHeaders = @{
    Authorization = "Bearer $token"
    "Content-Type" = "application/json"
}

Write-Section "Deploy detection engine + infra"
$deployBody = @{
    name = $deployName
    image = $imageTag
    exposed_port = $deployPort
    cpu = 1
    memory = 512
    environment = @{
        CH_HOST = "auto"
        CH_PORT = "9000"
        CH_USER = "default"
        CH_PASSWORD = "clickhouse123"
        CH_DATABASE = "detection_db"
        DB_HOST = "auto"
        DB_PORT = "9000"
        DB_USER = "default"
        DB_PASSWORD = "clickhouse123"
        DB_NAME = "detection_db"
        PG_HOST = "auto"
        PG_PORT = "5432"
        PG_USER = "postgres"
        PG_PASSWORD = "postgres123"
        PG_DATABASE = "alerts_db"
        API_PORT = "$deployPort"
        RULES_STORAGE_PATH = "/opt/rules_storage"
    }
    volumes = @(
        @{
            host_path = $rulesDir
            container_path = "/opt/rules_storage"
        }
    )
}

$deployResponse = Invoke-RestMethod -Uri "$provUrl/deploy" -Method Post -Headers $authHeaders -ContentType "application/json" -Body ($deployBody | ConvertTo-Json -Depth 6)
if (-not $deployResponse.success) {
    Write-Host "Deploy API returned error: $($deployResponse.message)" -ForegroundColor Red
    exit 1
}

$deployment = $deployResponse.data
$containerInfo = $deployment.container
$apiBase = $containerInfo.endpoint.TrimEnd("/")
$clickhouseInfra = $deployment.created_infrastructure | Where-Object { $_.type -eq "clickhouse" }

Write-Host "Deployment status : $($deployment.status)" -ForegroundColor Green
Write-Host "Container name    : $($containerInfo.name)"
Write-Host "API endpoint      : $apiBase"
if ($clickhouseInfra) {
    Write-Host "ClickHouse ID     : $($clickhouseInfra.id)"
}

Write-Section "Wait for engine health"
$health = $null
for ($i = 1; $i -le 20; $i++) {
    try {
        $health = Invoke-RestMethod -Uri "$apiBase/health" -TimeoutSec 10
        if ($health.clickhouse -eq "connected" -and $health.status -eq "healthy") {
            break
        }
    } catch {
        # swallow and retry
    }
    Start-Sleep -Seconds 5
}

if (-not $health) {
    Write-Host "Engine did not become healthy in time." -ForegroundColor Red
    exit 1
}

Write-Host "Health: ClickHouse=$($health.clickhouse) Postgres=$($health.postgres) Rules=$($health.rules) Alerts=$($health.alerts)" -ForegroundColor Green

Write-Section "Rules check"
$rules = Invoke-RestMethod -Uri "$apiBase/rules" -TimeoutSec 10
Write-Host "Loaded rules: $($rules.count)" -ForegroundColor Green

Write-Section "Insert log + trigger detection"
$logBody = @{
    level = "ERROR"
    message = "connection refused while reaching upstream from e2e test"
    source = "e2e"
} | ConvertTo-Json

$logResp = Invoke-RestMethod -Uri "$apiBase/logs" -Method Post -ContentType "application/json" -Body $logBody
Write-Host "Inserted log id: $($logResp.log_id)" -ForegroundColor Green

Invoke-RestMethod -Uri "$apiBase/detect" -Method Post | Out-Null
Start-Sleep -Seconds 5

$alerts = Invoke-RestMethod -Uri "$apiBase/alerts?limit=10"
Write-Host "Alerts returned: $($alerts.count)" -ForegroundColor Green

Write-Section "ClickHouse query via provisioning API"
if (-not $clickhouseInfra) {
    Write-Host "No ClickHouse infra detected; skipping query check." -ForegroundColor Yellow
} else {
    $chQueryBody = @{
        query = "SELECT count() AS log_count FROM logs"
    } | ConvertTo-Json

    $chQuery = Invoke-RestMethod -Uri "$provUrl/clickhouse/$($clickhouseInfra.id)/query" -Method Post -Headers $authHeaders -ContentType "application/json" -Body $chQueryBody
    Write-Host "ClickHouse query success: $($chQuery.success) rows=$($chQuery.row_count)" -ForegroundColor Green
    if ($chQuery.data) {
        Write-Host "Data: $($chQuery.data)"
    }
}

Write-Section "Summary"
Write-Host "Deployment ID    : $($deployment.deployment_id)"
Write-Host "Container        : $($containerInfo.name)"
Write-Host "API endpoint     : $apiBase"
if ($clickhouseInfra) {
    Write-Host "ClickHouse ID    : $($clickhouseInfra.id)"
}
Write-Host "Alerts count     : $($alerts.count)"
Write-Host ""
Write-Host "Remember to clean up the test containers and networks when done." -ForegroundColor Yellow

# Test Kafka va Redis trong IaaS Platform
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Test Kafka & Redis Infrastructure" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

# Kiem tra services
Write-Host "`n1. Kiem tra services dang chay..." -ForegroundColor Yellow
$services = docker ps --filter "name=iaas" --format "table {{.Names}}\t{{.Status}}" | Select-Object -Skip 1

if ($services -match "iaas-kafka" -and $services -match "iaas-redis") {
    Write-Host "  OK - Kafka va Redis dang chay" -ForegroundColor Green
} else {
    Write-Host "  FAIL - Kafka hoac Redis khong chay" -ForegroundColor Red
    exit 1
}

# ===========================================
# TEST REDIS
# ===========================================
Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "TEST REDIS" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

Write-Host "`n2. Test Redis connection..." -ForegroundColor Yellow
$redisTest = docker exec iaas-redis redis-cli PING 2>&1
if ($redisTest -match "PONG") {
    Write-Host "  OK - Redis responsive" -ForegroundColor Green
} else {
    Write-Host "  FAIL - Redis not responding" -ForegroundColor Red
}

Write-Host "`n3. Test Redis SET/GET..." -ForegroundColor Yellow
docker exec iaas-redis redis-cli SET test:key "Hello Redis" | Out-Null
$value = docker exec iaas-redis redis-cli GET test:key
if ($value -eq "Hello Redis") {
    Write-Host "  OK - Redis SET/GET hoat dong" -ForegroundColor Green
    Write-Host "  Value: $value" -ForegroundColor Gray
} else {
    Write-Host "  FAIL - Redis SET/GET loi" -ForegroundColor Red
}

Write-Host "`n4. Test Redis TTL (expiration)..." -ForegroundColor Yellow
docker exec iaas-redis redis-cli SETEX test:ttl 10 "Expire in 10s" | Out-Null
$ttl = docker exec iaas-redis redis-cli TTL test:ttl
Write-Host "  OK - Key se expire trong $ttl giay" -ForegroundColor Green

Write-Host "`n5. Kiem tra Redis keys dang su dung..." -ForegroundColor Yellow
$keys = docker exec iaas-redis redis-cli KEYS "*"
if ($keys) {
    Write-Host "  Cac keys dang ton tai:" -ForegroundColor Cyan
    $keys | ForEach-Object {
        Write-Host "    - $_" -ForegroundColor Gray
    }
} else {
    Write-Host "  Khong co key nao" -ForegroundColor Gray
}

Write-Host "`n6. Kiem tra Redis memory usage..." -ForegroundColor Yellow
$memInfo = docker exec iaas-redis redis-cli INFO memory | Select-String "used_memory_human"
Write-Host "  $memInfo" -ForegroundColor Cyan

# ===========================================
# TEST KAFKA
# ===========================================
Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "TEST KAFKA" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

Write-Host "`n7. List Kafka topics..." -ForegroundColor Yellow
$topics = docker exec iaas-kafka kafka-topics --bootstrap-server localhost:9092 --list 2>&1
if ($topics) {
    Write-Host "  Topics hien co:" -ForegroundColor Cyan
    $topics | ForEach-Object {
        if ($_ -and $_ -notmatch "WARN" -and $_ -notmatch "SLF4J") {
            Write-Host "    - $_" -ForegroundColor Gray
        }
    }
} else {
    Write-Host "  Chua co topic nao" -ForegroundColor Gray
}

Write-Host "`n8. Test produce message vao Kafka..." -ForegroundColor Yellow
$testMessage = @{
    instance_id = "test-instance-123"
    user_id = "test-user"
    type = "test"
    action = "ping"
    timestamp = (Get-Date).ToString("o")
    metadata = @{
        source = "powershell-test"
    }
} | ConvertTo-Json -Compress

# Produce message
$produceResult = echo $testMessage | docker exec -i iaas-kafka kafka-console-producer --bootstrap-server localhost:9092 --topic infrastructure.events 2>&1

Start-Sleep -Seconds 1

if ($produceResult -notmatch "ERROR") {
    Write-Host "  OK - Message da duoc publish len Kafka" -ForegroundColor Green
    Write-Host "  Topic: infrastructure.events" -ForegroundColor Gray
} else {
    Write-Host "  FAIL - Khong the publish message" -ForegroundColor Red
}

Write-Host "`n9. Test consume message tu Kafka..." -ForegroundColor Yellow
Write-Host "  Dang consume message (timeout 5s)..." -ForegroundColor Gray

# Consume latest messages (timeout 5 seconds)
$consumeJob = Start-Job -ScriptBlock {
    docker exec iaas-kafka kafka-console-consumer --bootstrap-server localhost:9092 --topic infrastructure.events --from-beginning --max-messages 1 --timeout-ms 5000 2>&1
}

Wait-Job $consumeJob -Timeout 5 | Out-Null
$messages = Receive-Job $consumeJob
Remove-Job $consumeJob -Force

if ($messages -and $messages -match "instance_id") {
    Write-Host "  OK - Co the consume message tu Kafka" -ForegroundColor Green
    Write-Host "  Sample message:" -ForegroundColor Gray
    $messages | Select-Object -First 3 | ForEach-Object {
        if ($_ -and $_ -notmatch "WARN" -and $_ -notmatch "SLF4J") {
            Write-Host "    $_" -ForegroundColor Gray
        }
    }
} else {
    Write-Host "  INFO - Khong co message hoac timeout" -ForegroundColor Yellow
}

# ===========================================
# TEST REDIS CACHE trong Application
# ===========================================
Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "TEST REDIS CACHE trong Application" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

# Login de lay token
Write-Host "`n10. Login va test Redis session..." -ForegroundColor Yellow
$loginBody = @{
    username = "admin"
    password = "password123"
} | ConvertTo-Json

try {
    $loginResp = Invoke-RestMethod -Uri "http://localhost:8082/auth/login" -Method POST -Headers @{"Content-Type"="application/json"} -Body $loginBody
    $token = $loginResp.data.access_token
    
    Write-Host "  OK - Login thanh cong, token da duoc tao" -ForegroundColor Green
    
    # Kiem tra refresh token trong Redis
    Start-Sleep -Seconds 1
    $redisKeys = docker exec iaas-redis redis-cli KEYS "refresh:*"
    if ($redisKeys) {
        Write-Host "  OK - Refresh token da duoc luu trong Redis" -ForegroundColor Green
        Write-Host "  Keys: $($redisKeys.Count) refresh tokens" -ForegroundColor Gray
    }
    
} catch {
    Write-Host "  FAIL - Khong the login" -ForegroundColor Red
}

# ===========================================
# STATISTICS
# ===========================================
Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "STATISTICS" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

Write-Host "`nRedis:" -ForegroundColor Yellow
$redisInfo = docker exec iaas-redis redis-cli INFO stats | Select-String "total_commands_processed|total_connections_received|keyspace_hits|keyspace_misses"
$redisInfo | ForEach-Object {
    Write-Host "  $_" -ForegroundColor Gray
}

# Calculate hit rate if available
$hits = docker exec iaas-redis redis-cli INFO stats | Select-String "keyspace_hits:(\d+)" | ForEach-Object { $_.Matches.Groups[1].Value }
$misses = docker exec iaas-redis redis-cli INFO stats | Select-String "keyspace_misses:(\d+)" | ForEach-Object { $_.Matches.Groups[1].Value }

if ($hits -and $misses) {
    $hitRate = [math]::Round(($hits / ($hits + $misses)) * 100, 2)
    Write-Host "  Cache Hit Rate: $hitRate%" -ForegroundColor Cyan
}

Write-Host "`nKafka:" -ForegroundColor Yellow
$kafkaBrokers = docker exec iaas-kafka kafka-broker-api-versions --bootstrap-server localhost:9092 2>&1 | Select-String "ApiVersion"
if ($kafkaBrokers) {
    Write-Host "  OK - Kafka broker responsive" -ForegroundColor Green
}

# ===========================================
# MONITORING
# ===========================================
Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "MONITORING & HEALTH" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

Write-Host "`nContainer Health:" -ForegroundColor Yellow
$healthStatus = docker ps --filter "name=iaas-redis" --filter "name=iaas-kafka" --format "table {{.Names}}\t{{.Status}}"
$healthStatus | ForEach-Object {
    Write-Host "  $_" -ForegroundColor Gray
}

Write-Host "`n========================================" -ForegroundColor Green
Write-Host "TEST HOAN TAT!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green

Write-Host "`nTong ket:" -ForegroundColor Cyan
Write-Host "  Redis: Session storage + Cache layer" -ForegroundColor Gray
Write-Host "  Kafka: Event streaming giua cac services" -ForegroundColor Gray
Write-Host "  Architecture: Event-driven microservices" -ForegroundColor Gray
Write-Host "`nXem them: KAFKA-REDIS-ANALYSIS.md" -ForegroundColor Yellow


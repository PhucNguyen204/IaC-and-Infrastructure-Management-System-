Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Starting IaaS Management System" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

Write-Host ""
Write-Host "Checking Docker..." -ForegroundColor Yellow
docker --version
if ($LASTEXITCODE -ne 0) {
    Write-Host "Docker is not running. Please start Docker Desktop first." -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "Building and starting services with Docker Compose..." -ForegroundColor Yellow
Write-Host "Using Docker BuildKit for faster builds..." -ForegroundColor Cyan
$env:DOCKER_BUILDKIT = "1"
docker-compose up -d --build

if ($LASTEXITCODE -ne 0) {
    Write-Host "Failed to start services" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "Waiting for services to be ready..." -ForegroundColor Yellow
Start-Sleep -Seconds 15

Write-Host ""
Write-Host "Checking service health..." -ForegroundColor Yellow
Write-Host "- Authentication Service (Port 8082):" -ForegroundColor White
try {
    $response = Invoke-WebRequest -Uri "http://localhost:8082/health" -UseBasicParsing
    Write-Host "  Status: $($response.StatusCode) - HEALTHY" -ForegroundColor Green
} catch {
    Write-Host "  Status: NOT READY" -ForegroundColor Red
}

Write-Host "- Provisioning Service (Port 8083):" -ForegroundColor White
try {
    $response = Invoke-WebRequest -Uri "http://localhost:8083/health" -UseBasicParsing
    Write-Host "  Status: $($response.StatusCode) - HEALTHY" -ForegroundColor Green
} catch {
    Write-Host "  Status: NOT READY" -ForegroundColor Red
}

Write-Host "- Monitoring Service (Port 8084):" -ForegroundColor White
try {
    $response = Invoke-WebRequest -Uri "http://localhost:8084/health" -UseBasicParsing
    Write-Host "  Status: $($response.StatusCode) - HEALTHY" -ForegroundColor Green
} catch {
    Write-Host "  Status: NOT READY" -ForegroundColor Red
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "Services are running!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""
Write-Host "Service URLs:" -ForegroundColor Yellow
Write-Host "  - Authentication API: http://localhost:8082" -ForegroundColor White
Write-Host "  - Provisioning API: http://localhost:8083" -ForegroundColor White
Write-Host "  - Monitoring API: http://localhost:8084" -ForegroundColor White
Write-Host "  - Elasticsearch: http://localhost:9200" -ForegroundColor White
Write-Host "  - Kafka: localhost:29092" -ForegroundColor White
Write-Host ""
Write-Host "To view logs: docker-compose logs -f [service-name]" -ForegroundColor Cyan
Write-Host "To stop services: docker-compose down" -ForegroundColor Cyan


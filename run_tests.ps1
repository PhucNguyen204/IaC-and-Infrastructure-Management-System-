Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "Running tests for IaaS Management System" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan

Write-Host ""
Write-Host "1. Testing Authentication Service..." -ForegroundColor Yellow
Set-Location vcs-authentication-service
go test ./... -v -cover
if ($LASTEXITCODE -ne 0) {
    Write-Host "Authentication service tests failed" -ForegroundColor Red
    Set-Location ..
    exit 1
}
Set-Location ..

Write-Host ""
Write-Host "2. Testing Infrastructure Provisioning Service..." -ForegroundColor Yellow
Set-Location vcs-infrastructure-provisioning-service
go test ./... -v -cover
if ($LASTEXITCODE -ne 0) {
    Write-Host "Provisioning service tests failed" -ForegroundColor Red
    Set-Location ..
    exit 1
}
Set-Location ..

Write-Host ""
Write-Host "3. Testing Infrastructure Monitoring Service..." -ForegroundColor Yellow
Set-Location vcs-infrastructure-monitoring-service
go test ./... -v -cover
if ($LASTEXITCODE -ne 0) {
    Write-Host "Monitoring service tests failed" -ForegroundColor Red
    Set-Location ..
    exit 1
}
Set-Location ..

Write-Host ""
Write-Host "==========================================" -ForegroundColor Green
Write-Host "All tests passed successfully!" -ForegroundColor Green
Write-Host "==========================================" -ForegroundColor Green


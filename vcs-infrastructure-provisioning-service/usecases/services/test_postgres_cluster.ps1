#!/usr/bin/env pwsh
# Test script for PostgreSQL Cluster Service Integration Tests

param(
    [switch]$Verbose,
    [switch]$Short,
    [string]$TestFilter = "",
    [switch]$Coverage
)

Write-Host "=====================================" -ForegroundColor Cyan
Write-Host "PostgreSQL Cluster Integration Tests" -ForegroundColor Cyan
Write-Host "=====================================" -ForegroundColor Cyan
Write-Host ""

# Check Docker is running
Write-Host "Checking Docker status..." -ForegroundColor Yellow
try {
    docker info 2>$null | Out-Null
    if ($LASTEXITCODE -ne 0) {
        Write-Host "✗ Docker is not running. Please start Docker Desktop." -ForegroundColor Red
        exit 1
    }
    Write-Host "✓ Docker is running" -ForegroundColor Green
} catch {
    Write-Host "✗ Docker is not available. Please install and start Docker." -ForegroundColor Red
    exit 1
}

Write-Host ""

# Set working directory
$scriptPath = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $scriptPath

# Build test command
$testCmd = "go test"

if ($Verbose) {
    $testCmd += " -v"
}

if (-not $Short) {
    $testCmd += " -timeout 30m"
    Write-Host "Running FULL integration tests (this may take several minutes)..." -ForegroundColor Yellow
} else {
    $testCmd += " -short"
    Write-Host "Running SHORT tests (integration tests will be skipped)..." -ForegroundColor Yellow
}

if ($TestFilter -ne "") {
    $testCmd += " -run `"$TestFilter`""
    Write-Host "Filter: $TestFilter" -ForegroundColor Cyan
}

if ($Coverage) {
    $testCmd += " -coverprofile=coverage.out"
    Write-Host "Coverage enabled" -ForegroundColor Cyan
}

$testCmd += " ./postgres_cluster_integration_test.go ./postgres_cluster_service.go"

Write-Host ""
Write-Host "Command: $testCmd" -ForegroundColor Gray
Write-Host ""

# Run tests
Invoke-Expression $testCmd

$exitCode = $LASTEXITCODE

Write-Host ""

if ($exitCode -eq 0) {
    Write-Host "=====================================" -ForegroundColor Green
    Write-Host "✓ All tests passed!" -ForegroundColor Green
    Write-Host "=====================================" -ForegroundColor Green
    
    if ($Coverage) {
        Write-Host ""
        Write-Host "Generating coverage report..." -ForegroundColor Yellow
        go tool cover -html=coverage.out -o coverage.html
        Write-Host "Coverage report: coverage.html" -ForegroundColor Cyan
    }
} else {
    Write-Host "=====================================" -ForegroundColor Red
    Write-Host "✗ Tests failed!" -ForegroundColor Red
    Write-Host "=====================================" -ForegroundColor Red
}

Write-Host ""

# Cleanup old containers (optional)
Write-Host "Clean up old test containers? [y/N]" -ForegroundColor Yellow
$cleanup = Read-Host
if ($cleanup -eq "y" -or $cleanup -eq "Y") {
    Write-Host "Cleaning up test containers..." -ForegroundColor Yellow
    docker ps -aq --filter "name=test-cluster-" --filter "name=persist-test-" | ForEach-Object {
        docker rm -f $_ 2>$null
    }
    docker volume ls -q --filter "name=pg-" | ForEach-Object {
        docker volume rm $_ 2>$null
    }
    docker network ls -q --filter "name=pg-cluster-" | ForEach-Object {
        docker network rm $_ 2>$null
    }
    Write-Host "✓ Cleanup completed" -ForegroundColor Green
}

exit $exitCode

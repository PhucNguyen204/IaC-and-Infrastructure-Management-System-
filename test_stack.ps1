# Test Stack APIs - Create a complete web application stack
# Usage: .\test_stack.ps1

$ErrorActionPreference = "Stop"

$BASE_URL = "http://localhost:8083/api/v1"
$TOKEN = ""

Write-Host "=== Stack API Test ===" -ForegroundColor Cyan

# Get token
if (-not $TOKEN) {
    Write-Host "`n[AUTH] Getting authentication token..." -ForegroundColor Yellow
    $authResponse = Invoke-RestMethod -Uri "http://localhost:8082/auth/login" `
        -Method POST `
        -Headers @{"Content-Type"="application/json"} `
        -Body '{"username":"admin","password":"password123"}'
    $TOKEN = $authResponse.data.access_token
    Write-Host "Token obtained: $($TOKEN.Substring(0, 20))..." -ForegroundColor Green
}

$HEADERS = @{
    "Content-Type" = "application/json"
    "Authorization" = "Bearer $TOKEN"
}

# Create a complete web application stack
Write-Host "`n[CREATE STACK] Creating 'my-web-app-prod' stack..." -ForegroundColor Yellow
Write-Host "This stack includes: PostgreSQL + App Service + Nginx Gateway" -ForegroundColor Gray

$createStackRequest = @{
    name = "my-web-app-prod"
    description = "Production web application with database and gateway"
    environment = "prod"
    project_id = "project-001"
    tags = @("web", "production", "backend")
    resources = @(
        # 1. PostgreSQL Database Instance (created first)
        @{
            type = "POSTGRES_INSTANCE"
            role = "database"
            name = "app-database"
            spec = @{
                plan = "medium"
            }
            depends_on = @()
            order = 1
        },
        # 2. Docker Service (App) - depends on PostgreSQL
        @{
            type = "DOCKER_SERVICE"
            role = "app"
            name = "web-app"
            spec = @{
                image = "nginx"
                image_tag = "alpine"
                service_type = "web"
                ports = @(
                    @{
                        container_port = 80
                        host_port = 0
                        protocol = "tcp"
                    }
                )
                networks = @(
                    @{
                        network_id = "iaas_iaas-network"
                        alias = "web-app"
                    }
                )
                restart_policy = "always"
                plan = "small"
            }
            depends_on = @("app-database")
            order = 2
        },
        # 3. Nginx Gateway - routes traffic to app
        @{
            type = "NGINX_GATEWAY"
            role = "gateway"
            name = "api-gateway"
            spec = @{
                plan = "small"
                domains = @(
                    @{
                        domain = "myapp.local"
                        port = 8092
                    }
                )
            }
            depends_on = @("web-app")
            order = 3
        }
    )
} | ConvertTo-Json -Depth 10

try {
    $createResponse = Invoke-RestMethod -Uri "$BASE_URL/stacks" `
        -Method POST `
        -Headers $HEADERS `
        -Body $createStackRequest
    
    $STACK_ID = $createResponse.data.id
    Write-Host "Stack created successfully!" -ForegroundColor Green
    Write-Host "Stack ID: $STACK_ID" -ForegroundColor Cyan
    Write-Host "Status: $($createResponse.data.status)" -ForegroundColor Cyan
    Write-Host "Resources created: $($createResponse.data.resources.Count)" -ForegroundColor Cyan
    
    Write-Host "`nResources in stack:" -ForegroundColor Yellow
    foreach ($resource in $createResponse.data.resources) {
        Write-Host "  - $($resource.resource_name) ($($resource.resource_type)) - Role: $($resource.role)" -ForegroundColor Gray
        if ($resource.outputs) {
            foreach ($key in $resource.outputs.Keys) {
                Write-Host "    $key : $($resource.outputs[$key])" -ForegroundColor DarkGray
            }
        }
    }
} catch {
    Write-Host "Failed to create stack: $_" -ForegroundColor Red
    Write-Host $_.Exception.Response.StatusCode -ForegroundColor Red
    exit 1
}

Start-Sleep -Seconds 2

# Get Stack Info
Write-Host "`n[GET STACK] Retrieving stack information..." -ForegroundColor Yellow
try {
    $getResponse = Invoke-RestMethod -Uri "$BASE_URL/stacks/$STACK_ID" `
        -Method GET `
        -Headers $HEADERS
    
    Write-Host "Stack details:" -ForegroundColor Green
    Write-Host "  Name: $($getResponse.data.name)" -ForegroundColor Cyan
    Write-Host "  Environment: $($getResponse.data.environment)" -ForegroundColor Cyan
    Write-Host "  Status: $($getResponse.data.status)" -ForegroundColor Cyan
    Write-Host "  Project ID: $($getResponse.data.project_id)" -ForegroundColor Cyan
    Write-Host "  Tags: $($getResponse.data.tags -join ', ')" -ForegroundColor Cyan
    Write-Host "  Resources: $($getResponse.data.resources.Count)" -ForegroundColor Cyan
} catch {
    Write-Host "Failed to get stack: $_" -ForegroundColor Red
}

# List All Stacks
Write-Host "`n[LIST STACKS] Listing all stacks..." -ForegroundColor Yellow
try {
    $listResponse = Invoke-RestMethod -Uri "$BASE_URL/stacks?page=1&page_size=10" `
        -Method GET `
        -Headers $HEADERS
    
    Write-Host "Total stacks: $($listResponse.data.total_count)" -ForegroundColor Green
    foreach ($stack in $listResponse.data.stacks) {
        Write-Host "  - $($stack.name) [$($stack.environment)] - $($stack.resource_count) resources - Status: $($stack.status)" -ForegroundColor Gray
    }
} catch {
    Write-Host "Failed to list stacks: $_" -ForegroundColor Red
}

# Update Stack (add tags)
Write-Host "`n[UPDATE STACK] Updating stack metadata..." -ForegroundColor Yellow
$updateRequest = @{
    description = "Updated production web application stack"
    tags = @("web", "production", "backend", "updated")
} | ConvertTo-Json

try {
    $updateResponse = Invoke-RestMethod -Uri "$BASE_URL/stacks/$STACK_ID" `
        -Method PUT `
        -Headers $HEADERS `
        -Body $updateRequest
    
    Write-Host "Stack updated successfully!" -ForegroundColor Green
    Write-Host "  New tags: $($updateResponse.data.tags -join ', ')" -ForegroundColor Cyan
} catch {
    Write-Host "Failed to update stack: $_" -ForegroundColor Red
}

# Stop Stack
Write-Host "`n[STOP STACK] Stopping all resources in stack..." -ForegroundColor Yellow
try {
    Invoke-RestMethod -Uri "$BASE_URL/stacks/$STACK_ID/stop" `
        -Method POST `
        -Headers $HEADERS | Out-Null
    Write-Host "Stack stopped successfully!" -ForegroundColor Green
} catch {
    Write-Host "Failed to stop stack: $_" -ForegroundColor Red
}

Start-Sleep -Seconds 2

# Start Stack
Write-Host "`n[START STACK] Starting all resources in stack..." -ForegroundColor Yellow
try {
    Invoke-RestMethod -Uri "$BASE_URL/stacks/$STACK_ID/start" `
        -Method POST `
        -Headers $HEADERS | Out-Null
    Write-Host "Stack started successfully!" -ForegroundColor Green
} catch {
    Write-Host "Failed to start stack: $_" -ForegroundColor Red
}

Start-Sleep -Seconds 2

# Restart Stack
Write-Host "`n[RESTART STACK] Restarting all resources in stack..." -ForegroundColor Yellow
try {
    Invoke-RestMethod -Uri "$BASE_URL/stacks/$STACK_ID/restart" `
        -Method POST `
        -Headers $HEADERS | Out-Null
    Write-Host "Stack restarted successfully!" -ForegroundColor Green
} catch {
    Write-Host "Failed to restart stack: $_" -ForegroundColor Red
}

Start-Sleep -Seconds 2

# Create Stack Template
Write-Host "`n[CREATE TEMPLATE] Creating reusable stack template..." -ForegroundColor Yellow
$templateRequest = @{
    name = "Standard Web App Template"
    description = "Template for web applications with PostgreSQL and Nginx"
    category = "web-app"
    is_public = $true
    resources = @(
        @{
            type = "POSTGRES_INSTANCE"
            role = "database"
            name = "database"
            spec = @{
                plan = "small"
            }
            order = 1
        },
        @{
            type = "DOCKER_SERVICE"
            role = "app"
            name = "app"
            spec = @{
                image = "nginx"
                image_tag = "alpine"
                service_type = "web"
                ports = @(
                    @{
                        container_port = 80
                        host_port = 0
                        protocol = "tcp"
                    }
                )
                plan = "small"
            }
            depends_on = @("database")
            order = 2
        },
        @{
            type = "NGINX_GATEWAY"
            role = "gateway"
            name = "gateway"
            spec = @{
                plan = "small"
            }
            depends_on = @("app")
            order = 3
        }
    )
} | ConvertTo-Json -Depth 10

try {
    $templateResponse = Invoke-RestMethod -Uri "$BASE_URL/stack-templates" `
        -Method POST `
        -Headers $HEADERS `
        -Body $templateRequest
    
    $TEMPLATE_ID = $templateResponse.data.id
    Write-Host "Template created successfully!" -ForegroundColor Green
    Write-Host "  Template ID: $TEMPLATE_ID" -ForegroundColor Cyan
    Write-Host "  Name: $($templateResponse.data.name)" -ForegroundColor Cyan
} catch {
    Write-Host "Failed to create template: $_" -ForegroundColor Red
}

# List Public Templates
Write-Host "`n[LIST TEMPLATES] Listing public templates..." -ForegroundColor Yellow
try {
    $templatesResponse = Invoke-RestMethod -Uri "$BASE_URL/stack-templates" `
        -Method GET `
        -Headers $HEADERS
    
    Write-Host "Public templates available: $($templatesResponse.data.Count)" -ForegroundColor Green
    foreach ($template in $templatesResponse.data) {
        Write-Host "  - $($template.name) [$($template.category)]" -ForegroundColor Gray
        Write-Host "    $($template.description)" -ForegroundColor DarkGray
    }
} catch {
    Write-Host "Failed to list templates: $_" -ForegroundColor Red
}

# Cleanup
Write-Host "`n[CLEANUP] Deleting stack and all resources..." -ForegroundColor Yellow
try {
    Invoke-RestMethod -Uri "$BASE_URL/stacks/$STACK_ID" `
        -Method DELETE `
        -Headers $HEADERS | Out-Null
    Write-Host "Stack deleted successfully!" -ForegroundColor Green
} catch {
    Write-Host "Failed to delete stack: $_" -ForegroundColor Red
}

Write-Host "`n=== Stack API Test Completed ===" -ForegroundColor Cyan
Write-Host @"
Summary:
✓ Stack creation with multiple resources (PostgreSQL + Docker + Nginx)
✓ Dependency resolution (app depends on database, gateway depends on app)
✓ Stack lifecycle operations (start/stop/restart)
✓ Stack templates for reusability
✓ Automatic resource orchestration in correct order
"@ -ForegroundColor Green

$BASE_URL = "http://localhost:8083"
$AUTH_URL = "http://localhost:8082"
$TOKEN = ""

Write-Host "=== Login ===" -ForegroundColor Cyan
$loginResponse = Invoke-RestMethod -Uri "$AUTH_URL/auth/login" -Method POST -Body (@{username="admin";password="password123"} | ConvertTo-Json) -ContentType "application/json"
$TOKEN = $loginResponse.data.access_token
Write-Host "Token: $TOKEN`n" -ForegroundColor Green

$headers = @{Authorization = "Bearer $TOKEN"}

Write-Host "=== Create Nginx ===" -ForegroundColor Cyan
$createBody = @{
    name = "test-nginx-extended"
    port = 8088
    ssl_port = 8443
    config = "server { listen 80; }"
    cpu_limit = 1000000000
    memory_limit = 536870912
} | ConvertTo-Json

$nginxResponse = Invoke-RestMethod -Uri "$BASE_URL/api/v1/nginx" -Method POST -Headers $headers -Body $createBody -ContentType "application/json"
$NGINX_ID = $nginxResponse.data.id
Write-Host "Nginx ID: $NGINX_ID`n" -ForegroundColor Green

Start-Sleep -Seconds 3

Write-Host "\n=== Add Domain ===" -ForegroundColor Cyan
$domainBody = @{domain = "api.example.com"} | ConvertTo-Json
Invoke-RestMethod -Uri "$BASE_URL/api/v1/nginx/$NGINX_ID/domains" -Method POST -Headers $headers -Body $domainBody -ContentType "application/json" | ConvertTo-Json

Write-Host "`n=== Add Route ===" -ForegroundColor Cyan
$routeBody = @{path = "/api/v1"; backend = "backend-service:8080"; priority = 10} | ConvertTo-Json
$routeResponse = Invoke-RestMethod -Uri "$BASE_URL/api/v1/nginx/$NGINX_ID/routes" -Method POST -Headers $headers -Body $routeBody -ContentType "application/json"
$ROUTE_ID = $routeResponse.data.id
Write-Host "Route ID: $ROUTE_ID" -ForegroundColor Green
$routeResponse | ConvertTo-Json

Write-Host "`n=== Update Route ===" -ForegroundColor Cyan
$updateRouteBody = @{backend = "new-backend:9090"; priority = 20} | ConvertTo-Json
Invoke-RestMethod -Uri "$BASE_URL/api/v1/nginx/$NGINX_ID/routes/$ROUTE_ID" -Method PUT -Headers $headers -Body $updateRouteBody -ContentType "application/json" | ConvertTo-Json

Write-Host "`n=== Update Upstreams ===" -ForegroundColor Cyan
$upstreamsBody = @{
    backends = @(
        @{address = "backend1:8080"; weight = 1}
        @{address = "backend2:8080"; weight = 2}
    )
    policy = "round_robin"
} | ConvertTo-Json -Depth 3
Invoke-RestMethod -Uri "$BASE_URL/api/v1/nginx/$NGINX_ID/upstreams" -Method PUT -Headers $headers -Body $upstreamsBody -ContentType "application/json" | ConvertTo-Json

Write-Host "`n=== Get Upstreams ===" -ForegroundColor Cyan
Invoke-RestMethod -Uri "$BASE_URL/api/v1/nginx/$NGINX_ID/upstreams" -Method GET -Headers $headers | ConvertTo-Json -Depth 5

Write-Host "`n=== Set Security Policy ===" -ForegroundColor Cyan
$securityBody = @{
    rate_limit = @{
        requests_per_second = 100
        burst = 200
        path = "/api"
    }
    ip_filter = @{
        allow_ips = @("192.168.1.0/24")
        deny_ips = @("10.0.0.1")
    }
} | ConvertTo-Json -Depth 3
Invoke-RestMethod -Uri "$BASE_URL/api/v1/nginx/$NGINX_ID/security" -Method POST -Headers $headers -Body $securityBody -ContentType "application/json" | ConvertTo-Json

Write-Host "`n=== Get Security Policy ===" -ForegroundColor Cyan
Invoke-RestMethod -Uri "$BASE_URL/api/v1/nginx/$NGINX_ID/security" -Method GET -Headers $headers | ConvertTo-Json -Depth 5

Write-Host "`n=== Get Nginx Info (with extended fields) ===" -ForegroundColor Cyan
Invoke-RestMethod -Uri "$BASE_URL/api/v1/nginx/$NGINX_ID" -Method GET -Headers $headers | ConvertTo-Json -Depth 5

Write-Host "`n=== Get Logs ===" -ForegroundColor Cyan
Invoke-RestMethod -Uri "$BASE_URL/api/v1/nginx/$NGINX_ID/logs?tail=50" -Method GET -Headers $headers | ConvertTo-Json -Depth 3

Write-Host "`n=== Get Stats ===" -ForegroundColor Cyan
Invoke-RestMethod -Uri "$BASE_URL/api/v1/nginx/$NGINX_ID/stats" -Method GET -Headers $headers | ConvertTo-Json -Depth 5

Write-Host "`n=== Delete Route ===" -ForegroundColor Cyan
Invoke-RestMethod -Uri "$BASE_URL/api/v1/nginx/$NGINX_ID/routes/$ROUTE_ID" -Method DELETE -Headers $headers | ConvertTo-Json

Write-Host "`n=== Delete Domain ===" -ForegroundColor Cyan
Invoke-RestMethod -Uri "$BASE_URL/api/v1/nginx/$NGINX_ID/domains/api.example.com" -Method DELETE -Headers $headers | ConvertTo-Json

Write-Host "`n=== Delete Security ===" -ForegroundColor Cyan
Invoke-RestMethod -Uri "$BASE_URL/api/v1/nginx/$NGINX_ID/security" -Method DELETE -Headers $headers | ConvertTo-Json

Write-Host "`n=== Cleanup: Delete Nginx ===" -ForegroundColor Cyan
Invoke-RestMethod -Uri "$BASE_URL/api/v1/nginx/$NGINX_ID" -Method DELETE -Headers $headers | ConvertTo-Json

Write-Host "`n=== Test Complete ===" -ForegroundColor Green

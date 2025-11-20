$BASE_URL = "http://localhost:8083"

Write-Host "`n=== Login ===" -ForegroundColor Cyan
$loginResponse = Invoke-RestMethod -Uri "http://localhost:8082/auth/login" -Method Post -ContentType "application/json" -Body (@{
    username = "admin"
    password = "password123"
} | ConvertTo-Json)
$TOKEN = $loginResponse.data.access_token
Write-Host "Token: $TOKEN`n"

$headers = @{
    "Authorization" = "Bearer $TOKEN"
    "Content-Type" = "application/json"
}

Write-Host "=== Create Shared PostgreSQL Instance ===" -ForegroundColor Cyan
$pgResponse = Invoke-RestMethod -Uri "$BASE_URL/api/v1/postgres/single" -Method Post -Headers $headers -Body (@{
    name = "shared-postgres-instance"
    version = "15"
    port = 5434
    database_name = "postgres"
    username = "postgres"
    password = "postgres123"
    cpu_limit = 1000000000
    memory_limit = 536870912
    storage_size = 10737418240
} | ConvertTo-Json)
$INSTANCE_ID = $pgResponse.data.id
Write-Host "Instance ID: $INSTANCE_ID`n"

Start-Sleep -Seconds 10

Write-Host "=== Usecase 2: Create Database for Project A ===" -ForegroundColor Cyan
$db1Response = Invoke-RestMethod -Uri "$BASE_URL/api/v1/postgres/$INSTANCE_ID/databases" -Method Post -Headers $headers -Body (@{
    db_name = "project_a_db"
    owner_username = "project_a_user"
    owner_password = "projecta123"
    project_id = "proj-001"
    tenant_id = "tenant-alpha"
    environment_id = "dev"
    max_size_gb = 5
    max_connections = 20
    init_schema = "CREATE TABLE users (id SERIAL PRIMARY KEY, name VARCHAR(100));"
} | ConvertTo-Json)
$DB1_ID = $db1Response.data.id
Write-Host "Database 1 ID: $DB1_ID"
$db1Response.data | ConvertTo-Json -Depth 5

Write-Host "`n=== Create Database for Project B ===" -ForegroundColor Cyan
$db2Response = Invoke-RestMethod -Uri "$BASE_URL/api/v1/postgres/$INSTANCE_ID/databases" -Method Post -Headers $headers -Body (@{
    db_name = "project_b_db"
    owner_username = "project_b_user"
    owner_password = "projectb123"
    project_id = "proj-002"
    tenant_id = "tenant-beta"
    environment_id = "prod"
    max_size_gb = 10
    max_connections = 50
} | ConvertTo-Json)
$DB2_ID = $db2Response.data.id
Write-Host "Database 2 ID: $DB2_ID"
$db2Response.data | ConvertTo-Json -Depth 5

Write-Host "`n=== List All Databases on Instance ===" -ForegroundColor Cyan
Invoke-RestMethod -Uri "$BASE_URL/api/v1/postgres/$INSTANCE_ID/databases" -Method Get -Headers $headers | ConvertTo-Json -Depth 5

Write-Host "`n=== Usecase 3: Get Metrics for DB1 ===" -ForegroundColor Cyan
Invoke-RestMethod -Uri "$BASE_URL/api/v1/databases/$DB1_ID/metrics" -Method Get -Headers $headers | ConvertTo-Json -Depth 5

Write-Host "`n=== Usecase 3: Update Quota for DB1 ===" -ForegroundColor Cyan
Invoke-RestMethod -Uri "$BASE_URL/api/v1/databases/$DB1_ID/quota" -Method Put -Headers $headers -Body (@{
    max_size_gb = 8
    max_connections = 30
} | ConvertTo-Json) | ConvertTo-Json -Depth 5

Write-Host "`n=== Get Updated Metrics ===" -ForegroundColor Cyan
Invoke-RestMethod -Uri "$BASE_URL/api/v1/databases/$DB1_ID/metrics" -Method Get -Headers $headers | ConvertTo-Json -Depth 5

Write-Host "`n=== Usecase 4: Backup Database 1 ===" -ForegroundColor Cyan
$backupResponse = Invoke-RestMethod -Uri "$BASE_URL/api/v1/databases/$DB1_ID/backup" -Method Post -Headers $headers -Body (@{
    backup_type = "LOGICAL"
    mode = "MANUAL"
} | ConvertTo-Json)
$BACKUP_ID = $backupResponse.data.id
Write-Host "Backup ID: $BACKUP_ID"
$backupResponse.data | ConvertTo-Json -Depth 5

Start-Sleep -Seconds 3

Write-Host "`n=== Usecase 4: Restore Database (Clone Mode) ===" -ForegroundColor Cyan
Invoke-RestMethod -Uri "$BASE_URL/api/v1/databases/$DB1_ID/restore" -Method Post -Headers $headers -Body (@{
    backup_id = $BACKUP_ID
    mode = "CLONE"
    new_db_name = "project_a_db_clone"
} | ConvertTo-Json) | ConvertTo-Json -Depth 5

Write-Host "`n=== Usecase 5: Lock Database 2 (Read-Only) ===" -ForegroundColor Cyan
Invoke-RestMethod -Uri "$BASE_URL/api/v1/databases/$DB2_ID/lifecycle" -Method Post -Headers $headers -Body (@{
    action = "LOCK"
    require_backup = $false
} | ConvertTo-Json) | ConvertTo-Json -Depth 5

Write-Host "`n=== Check DB2 Status After Lock ===" -ForegroundColor Cyan
Invoke-RestMethod -Uri "$BASE_URL/api/v1/databases/$DB2_ID" -Method Get -Headers $headers | ConvertTo-Json -Depth 5

Write-Host "`n=== Unlock Database 2 ===" -ForegroundColor Cyan
Invoke-RestMethod -Uri "$BASE_URL/api/v1/databases/$DB2_ID/lifecycle" -Method Post -Headers $headers -Body (@{
    action = "UNLOCK"
    require_backup = $false
} | ConvertTo-Json) | ConvertTo-Json -Depth 5

Write-Host "`n=== Usecase 6: Get Instance Overview ===" -ForegroundColor Cyan
Invoke-RestMethod -Uri "$BASE_URL/api/v1/postgres/$INSTANCE_ID/overview" -Method Get -Headers $headers | ConvertTo-Json -Depth 10

Write-Host "`n=== Usecase 5: Drop Database with Backup ===" -ForegroundColor Cyan
Invoke-RestMethod -Uri "$BASE_URL/api/v1/databases/$DB2_ID/lifecycle" -Method Post -Headers $headers -Body (@{
    action = "DROP"
    require_backup = $true
} | ConvertTo-Json) | ConvertTo-Json -Depth 5

Write-Host "`n=== Final Instance Overview ===" -ForegroundColor Cyan
Invoke-RestMethod -Uri "$BASE_URL/api/v1/postgres/$INSTANCE_ID/overview" -Method Get -Headers $headers | ConvertTo-Json -Depth 10

Write-Host "`n=== Cleanup: Delete Instance ===" -ForegroundColor Cyan
Invoke-RestMethod -Uri "$BASE_URL/api/v1/postgres/single/$INSTANCE_ID" -Method Delete -Headers $headers | ConvertTo-Json -Depth 5

Write-Host "`n=== Test Complete ===" -ForegroundColor Green

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Setting up test user in database" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

Write-Host ""
Write-Host "Creating admin user (username: admin, password: password123)..." -ForegroundColor Yellow

$sql = @"
INSERT INTO users (id, username, hash, email)
VALUES (
    'test-user-123',
    'admin',
    '`$2a`$10`$rjDj8A.xMhKpbq3z5m5b9.ey4xKq4vGJz5O8gH8jZ9xK9tJ6eI9Gu',
    'admin@iaas.local'
) ON CONFLICT (username) DO NOTHING;

INSERT INTO user_scopes (name)
VALUES ('admin')
ON CONFLICT (name) DO NOTHING;

INSERT INTO user_scope_mapping (user_id, user_scope_id)
SELECT 'test-user-123', id FROM user_scopes WHERE name = 'admin'
ON CONFLICT DO NOTHING;

SELECT 'User created successfully' as status;
"@

try {
    docker exec -i iaas-metadata-postgres psql -U iaas_user -d iaas_metadata -c $sql
    Write-Host ""
    Write-Host "Test user created successfully!" -ForegroundColor Green
    Write-Host "Username: admin" -ForegroundColor Cyan
    Write-Host "Password: password123" -ForegroundColor Cyan
} catch {
    Write-Host "Error creating user: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "Make sure docker-compose is running" -ForegroundColor Yellow
}


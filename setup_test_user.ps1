Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Setting up test user in database" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

Write-Host ""
Write-Host "Generating bcrypt hash for password..." -ForegroundColor Yellow

# Generate bcrypt hash using Python
$pythonCode = @"
import bcrypt
password = 'password123'
hashed = bcrypt.hashpw(password.encode('utf-8'), bcrypt.gensalt())
print(hashed.decode('utf-8'))
"@

try {
    # Generate hash
    $hash = python -c $pythonCode
    
    if (-not $hash) {
        Write-Host "Error: Failed to generate hash. Make sure bcrypt is installed (pip install bcrypt)" -ForegroundColor Red
        exit 1
    }
    
    Write-Host "Hash generated: $($hash.Substring(0, 20))..." -ForegroundColor Gray
    Write-Host ""
    Write-Host "Creating admin user (username: admin, password: password123)..." -ForegroundColor Yellow
    
    # Create SQL file with proper escaping
    $sqlContent = @"
-- Delete existing user if exists
DELETE FROM user_scope_mapping WHERE user_id = 'test-user-123';
DELETE FROM users WHERE username = 'admin';

-- Insert new user with fresh hash
INSERT INTO users (id, username, hash, email)
VALUES (
    'test-user-123',
    'admin',
    '$hash',
    'admin@iaas.local'
);

-- Ensure admin scope exists
INSERT INTO user_scopes (name)
VALUES ('admin')
ON CONFLICT (name) DO NOTHING;

-- Map user to admin scope
INSERT INTO user_scope_mapping (user_id, user_scope_id)
SELECT 'test-user-123', id FROM user_scopes WHERE name = 'admin';

SELECT 'User created successfully' as status;
"@
    
    # Write SQL to temp file and execute
    $sqlContent | Out-File -FilePath "temp_setup.sql" -Encoding UTF8
    Get-Content "temp_setup.sql" | docker exec -i iaas-metadata-postgres psql -U iaas_user -d iaas_metadata
    Remove-Item "temp_setup.sql"
    
    Write-Host ""
    Write-Host "Test user created successfully!" -ForegroundColor Green
    Write-Host "Username: admin" -ForegroundColor Cyan
    Write-Host "Password: password123" -ForegroundColor Cyan
    
} catch {
    Write-Host ""
    Write-Host "Error creating user: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "Make sure:" -ForegroundColor Yellow
    Write-Host "  1. docker-compose is running" -ForegroundColor Yellow
    Write-Host "  2. Python bcrypt is installed (pip install bcrypt)" -ForegroundColor Yellow
    if (Test-Path "temp_setup.sql") {
        Remove-Item "temp_setup.sql"
    }
}


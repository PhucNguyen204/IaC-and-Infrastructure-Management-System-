-- 005_postgres_databases.sql
CREATE TABLE IF NOT EXISTS postgres_databases (
    id VARCHAR(36) PRIMARY KEY,
    instance_id VARCHAR(36) NOT NULL,
    db_name VARCHAR(100) NOT NULL,
    owner_username VARCHAR(100) NOT NULL,
    owner_password VARCHAR(255) NOT NULL,
    project_id VARCHAR(100),
    tenant_id VARCHAR(100),
    environment_id VARCHAR(100),
    max_size_gb INTEGER DEFAULT 10,
    max_connections INTEGER DEFAULT 50,
    current_size_mb BIGINT DEFAULT 0,
    active_conns INTEGER DEFAULT 0,
    status VARCHAR(20) DEFAULT 'CREATING',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (instance_id) REFERENCES postgre_sql_instances(id) ON DELETE CASCADE,
    UNIQUE(instance_id, db_name)
);

CREATE INDEX idx_postgres_databases_instance ON postgres_databases(instance_id);
CREATE INDEX idx_postgres_databases_project ON postgres_databases(project_id);
CREATE INDEX idx_postgres_databases_tenant ON postgres_databases(tenant_id);

CREATE TABLE IF NOT EXISTS postgres_backups (
    id VARCHAR(36) PRIMARY KEY,
    database_id VARCHAR(36) NOT NULL,
    backup_type VARCHAR(20) DEFAULT 'LOGICAL',
    size_mb BIGINT DEFAULT 0,
    location TEXT,
    status VARCHAR(20) DEFAULT 'PENDING',
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (database_id) REFERENCES postgres_databases(id) ON DELETE CASCADE
);

CREATE INDEX idx_postgres_backups_database ON postgres_backups(database_id);
CREATE INDEX idx_postgres_backups_status ON postgres_backups(status);

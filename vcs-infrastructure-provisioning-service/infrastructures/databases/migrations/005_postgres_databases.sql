-- 005_postgres_databases.sql
-- Tables for PostgreSQL database and backup management within clusters

CREATE TABLE IF NOT EXISTS postgres_databases (
    id VARCHAR(36) PRIMARY KEY,
    cluster_id VARCHAR(36) NOT NULL,
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
    FOREIGN KEY (cluster_id) REFERENCES postgre_sql_clusters(id) ON DELETE CASCADE,
    UNIQUE(cluster_id, db_name)
);

CREATE INDEX IF NOT EXISTS idx_postgres_databases_cluster ON postgres_databases(cluster_id);
CREATE INDEX IF NOT EXISTS idx_postgres_databases_project ON postgres_databases(project_id);
CREATE INDEX IF NOT EXISTS idx_postgres_databases_tenant ON postgres_databases(tenant_id);
CREATE INDEX IF NOT EXISTS idx_postgres_databases_status ON postgres_databases(status);

CREATE TABLE IF NOT EXISTS postgres_backups (
    id VARCHAR(36) PRIMARY KEY,
    cluster_id VARCHAR(36) NOT NULL,
    database_name VARCHAR(100),
    backup_type VARCHAR(20) DEFAULT 'FULL',
    size_mb BIGINT DEFAULT 0,
    location TEXT,
    status VARCHAR(20) DEFAULT 'PENDING',
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (cluster_id) REFERENCES postgre_sql_clusters(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_postgres_backups_cluster ON postgres_backups(cluster_id);
CREATE INDEX IF NOT EXISTS idx_postgres_backups_status ON postgres_backups(status);

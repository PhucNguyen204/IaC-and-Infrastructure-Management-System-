-- Migration: 007_dind_environments.sql
-- Description: Create tables for Docker-in-Docker environments

-- DinD Environments table
CREATE TABLE IF NOT EXISTS dind_environments (
    id VARCHAR(36) PRIMARY KEY,
    infrastructure_id VARCHAR(36) NOT NULL,
    name VARCHAR(255) NOT NULL,
    container_id VARCHAR(100),
    container_name VARCHAR(255),
    status VARCHAR(50) DEFAULT 'creating',
    docker_host VARCHAR(255),
    ip_address VARCHAR(50),
    resource_plan VARCHAR(20) DEFAULT 'medium',
    cpu_limit VARCHAR(20),
    memory_limit VARCHAR(20),
    storage_driver VARCHAR(50) DEFAULT 'overlay2',
    network_id VARCHAR(100),
    description TEXT,
    auto_cleanup BOOLEAN DEFAULT FALSE,
    ttl_hours INTEGER DEFAULT 0,
    expires_at TIMESTAMP,
    user_id VARCHAR(36) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (infrastructure_id) REFERENCES infrastructures(id) ON DELETE CASCADE
);

-- Create indexes for DinD environments
CREATE INDEX IF NOT EXISTS idx_dind_user_id ON dind_environments(user_id);
CREATE INDEX IF NOT EXISTS idx_dind_status ON dind_environments(status);
CREATE INDEX IF NOT EXISTS idx_dind_expires_at ON dind_environments(expires_at);
CREATE INDEX IF NOT EXISTS idx_dind_container_id ON dind_environments(container_id);

-- DinD Command History table
CREATE TABLE IF NOT EXISTS dind_command_history (
    id VARCHAR(36) PRIMARY KEY,
    environment_id VARCHAR(36) NOT NULL,
    command TEXT NOT NULL,
    output TEXT,
    exit_code INTEGER DEFAULT 0,
    duration INTEGER DEFAULT 0,
    executed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (environment_id) REFERENCES dind_environments(id) ON DELETE CASCADE
);

-- Create index for command history
CREATE INDEX IF NOT EXISTS idx_dind_command_history_env ON dind_command_history(environment_id);
CREATE INDEX IF NOT EXISTS idx_dind_command_history_time ON dind_command_history(executed_at);

-- Add comment to explain the table purpose
COMMENT ON TABLE dind_environments IS 'Docker-in-Docker isolated environments for users to run docker commands';
COMMENT ON TABLE dind_command_history IS 'History of commands executed in DinD environments';


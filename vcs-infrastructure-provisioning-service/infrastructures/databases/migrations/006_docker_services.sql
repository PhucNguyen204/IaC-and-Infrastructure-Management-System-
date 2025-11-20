-- Create docker_services table
CREATE TABLE IF NOT EXISTS docker_services (
    id VARCHAR(36) PRIMARY KEY,
    infrastructure_id VARCHAR(36) NOT NULL,
    image VARCHAR(255) NOT NULL,
    image_tag VARCHAR(100) NOT NULL DEFAULT 'latest',
    service_type VARCHAR(50) NOT NULL,
    command TEXT,
    args TEXT,
    container_id VARCHAR(255),
    status VARCHAR(50) NOT NULL,
    ip_address VARCHAR(45),
    internal_endpoint VARCHAR(255),
    restart_policy VARCHAR(50) NOT NULL DEFAULT 'no',
    cpu_limit BIGINT,
    memory_limit BIGINT,
    resources JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (infrastructure_id) REFERENCES infrastructures(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_docker_services_infrastructure_id ON docker_services(infrastructure_id);
CREATE INDEX IF NOT EXISTS idx_docker_services_container_id ON docker_services(container_id);
CREATE INDEX IF NOT EXISTS idx_docker_services_status ON docker_services(status);

-- Create docker_env_vars table
CREATE TABLE IF NOT EXISTS docker_env_vars (
    id VARCHAR(36) PRIMARY KEY,
    service_id VARCHAR(36) NOT NULL,
    key VARCHAR(255) NOT NULL,
    value TEXT NOT NULL,
    is_secret BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (service_id) REFERENCES docker_services(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_docker_env_vars_service_id ON docker_env_vars(service_id);

-- Create docker_ports table
CREATE TABLE IF NOT EXISTS docker_ports (
    id VARCHAR(36) PRIMARY KEY,
    service_id VARCHAR(36) NOT NULL,
    container_port INT NOT NULL,
    host_port INT NOT NULL,
    protocol VARCHAR(10) NOT NULL DEFAULT 'tcp',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (service_id) REFERENCES docker_services(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_docker_ports_service_id ON docker_ports(service_id);
CREATE INDEX IF NOT EXISTS idx_docker_ports_host_port ON docker_ports(host_port);

-- Create docker_networks table
CREATE TABLE IF NOT EXISTS docker_networks (
    id VARCHAR(36) PRIMARY KEY,
    service_id VARCHAR(36) NOT NULL,
    network_id VARCHAR(255) NOT NULL,
    alias VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (service_id) REFERENCES docker_services(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_docker_networks_service_id ON docker_networks(service_id);

-- Create docker_health_checks table
CREATE TABLE IF NOT EXISTS docker_health_checks (
    id VARCHAR(36) PRIMARY KEY,
    service_id VARCHAR(36) NOT NULL UNIQUE,
    type VARCHAR(50) NOT NULL,
    http_path VARCHAR(255),
    port INT,
    command TEXT,
    interval_seconds INT NOT NULL DEFAULT 30,
    timeout_seconds INT NOT NULL DEFAULT 10,
    retries INT NOT NULL DEFAULT 3,
    status VARCHAR(50) NOT NULL DEFAULT 'unknown',
    last_check TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (service_id) REFERENCES docker_services(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_docker_health_checks_service_id ON docker_health_checks(service_id);
CREATE INDEX IF NOT EXISTS idx_docker_health_checks_status ON docker_health_checks(status);

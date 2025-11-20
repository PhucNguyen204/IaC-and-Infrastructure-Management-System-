CREATE TABLE IF NOT EXISTS nginx_domains (
    id VARCHAR(255) PRIMARY KEY,
    nginx_id VARCHAR(255) NOT NULL,
    domain VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (nginx_id) REFERENCES nginx_instances(id) ON DELETE CASCADE,
    UNIQUE (nginx_id, domain)
);

CREATE TABLE IF NOT EXISTS nginx_routes (
    id VARCHAR(255) PRIMARY KEY,
    nginx_id VARCHAR(255) NOT NULL,
    path VARCHAR(500) NOT NULL,
    backend VARCHAR(500) NOT NULL,
    priority INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (nginx_id) REFERENCES nginx_instances(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS nginx_upstreams (
    id VARCHAR(255) PRIMARY KEY,
    nginx_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    policy VARCHAR(50) DEFAULT 'round_robin',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (nginx_id) REFERENCES nginx_instances(id) ON DELETE CASCADE,
    UNIQUE (nginx_id, name)
);

CREATE TABLE IF NOT EXISTS nginx_upstream_backends (
    id VARCHAR(255) PRIMARY KEY,
    upstream_id VARCHAR(255) NOT NULL,
    address VARCHAR(500) NOT NULL,
    weight INT DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (upstream_id) REFERENCES nginx_upstreams(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS nginx_certificates (
    id VARCHAR(255) PRIMARY KEY,
    nginx_id VARCHAR(255) NOT NULL UNIQUE,
    domain VARCHAR(255) NOT NULL,
    certificate TEXT NOT NULL,
    private_key TEXT NOT NULL,
    status VARCHAR(50) DEFAULT 'valid',
    expires_at TIMESTAMP,
    issuer VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (nginx_id) REFERENCES nginx_instances(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS nginx_securities (
    id VARCHAR(255) PRIMARY KEY,
    nginx_id VARCHAR(255) NOT NULL UNIQUE,
    rate_limit_rps INT DEFAULT 0,
    rate_limit_burst INT DEFAULT 0,
    rate_limit_path VARCHAR(500),
    allow_ips TEXT,
    deny_ips TEXT,
    basic_auth_enabled BOOLEAN DEFAULT FALSE,
    basic_auth_username VARCHAR(255),
    basic_auth_password VARCHAR(255),
    basic_auth_realm VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (nginx_id) REFERENCES nginx_instances(id) ON DELETE CASCADE
);

CREATE INDEX idx_nginx_routes_priority ON nginx_routes(nginx_id, priority DESC);
CREATE INDEX idx_nginx_domains_nginx_id ON nginx_domains(nginx_id);
CREATE INDEX idx_nginx_upstreams_nginx_id ON nginx_upstreams(nginx_id);

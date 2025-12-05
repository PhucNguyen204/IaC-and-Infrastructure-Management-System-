-- ClickHouse clusters table
CREATE TABLE IF NOT EXISTS clickhouse_clusters (
    id VARCHAR(36) PRIMARY KEY,
    infrastructure_id VARCHAR(36) NOT NULL REFERENCES infrastructures(id),
    cluster_name VARCHAR(255) NOT NULL,
    version VARCHAR(20) NOT NULL DEFAULT 'latest',
    node_count INTEGER NOT NULL DEFAULT 1,
    username VARCHAR(100) NOT NULL DEFAULT 'default',
    password VARCHAR(255) NOT NULL,
    database_name VARCHAR(100) NOT NULL,
    http_port INTEGER DEFAULT 8123,
    native_port INTEGER DEFAULT 9000,
    network_id VARCHAR(255),
    data_directory VARCHAR(255),
    storage_size INTEGER DEFAULT 0,
    cpu_limit BIGINT DEFAULT 0,
    memory_limit BIGINT DEFAULT 0,
    replication_enabled BOOLEAN DEFAULT FALSE,
    shard_count INTEGER DEFAULT 1,
    replica_count INTEGER DEFAULT 1,
    zoo_keeper_endpoints VARCHAR(500),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ClickHouse nodes table
CREATE TABLE IF NOT EXISTS clickhouse_nodes (
    id VARCHAR(36) PRIMARY KEY,
    cluster_id VARCHAR(36) NOT NULL REFERENCES clickhouse_clusters(id) ON DELETE CASCADE,
    node_name VARCHAR(100) NOT NULL,
    container_id VARCHAR(100),
    http_port INTEGER NOT NULL,
    native_port INTEGER NOT NULL,
    volume_id VARCHAR(255),
    shard_num INTEGER DEFAULT 1,
    replica_num INTEGER DEFAULT 1,
    is_healthy BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_clickhouse_clusters_infra ON clickhouse_clusters(infrastructure_id);
CREATE INDEX IF NOT EXISTS idx_clickhouse_nodes_cluster ON clickhouse_nodes(cluster_id);

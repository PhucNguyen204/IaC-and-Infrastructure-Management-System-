-- Performance indexes for high-traffic queries

-- Infrastructure indexes
CREATE INDEX IF NOT EXISTS idx_infra_user_type ON infrastructures(user_id, type);
CREATE INDEX IF NOT EXISTS idx_infra_user_status ON infrastructures(user_id, status);

-- Stack indexes
CREATE INDEX IF NOT EXISTS idx_stack_user_env ON stacks(user_id, environment);
CREATE INDEX IF NOT EXISTS idx_stack_tenant_proj ON stacks(tenant_id, project_id);
CREATE INDEX IF NOT EXISTS idx_stacks_tags_gin ON stacks USING GIN (tags jsonb_path_ops);

-- PostgreSQL Cluster indexes
CREATE INDEX IF NOT EXISTS idx_cluster_nodes_health ON cluster_nodes(cluster_id, is_healthy);
CREATE INDEX IF NOT EXISTS idx_cluster_nodes_role ON cluster_nodes(cluster_id, role);

-- Postgres databases indexes
CREATE INDEX IF NOT EXISTS idx_postgres_databases_cluster ON postgres_databases(cluster_id, status);
CREATE INDEX IF NOT EXISTS idx_postgres_databases_tenant ON postgres_databases(tenant_id, project_id);

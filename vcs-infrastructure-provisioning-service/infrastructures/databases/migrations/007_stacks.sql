-- Create stacks table
CREATE TABLE IF NOT EXISTS stacks (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    environment VARCHAR(50),
    project_id VARCHAR(36),
    tenant_id VARCHAR(36),
    user_id VARCHAR(36) NOT NULL,
    status VARCHAR(50) NOT NULL,
    tags JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_stacks_user_id ON stacks(user_id);
CREATE INDEX IF NOT EXISTS idx_stacks_name ON stacks(name);
CREATE INDEX IF NOT EXISTS idx_stacks_environment ON stacks(environment);
CREATE INDEX IF NOT EXISTS idx_stacks_project_id ON stacks(project_id);
CREATE INDEX IF NOT EXISTS idx_stacks_tenant_id ON stacks(tenant_id);

-- Create stack_resources table (links stacks to infrastructures)
CREATE TABLE IF NOT EXISTS stack_resources (
    id VARCHAR(36) PRIMARY KEY,
    stack_id VARCHAR(36) NOT NULL,
    infrastructure_id VARCHAR(36) NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    role VARCHAR(50),
    depends_on JSONB,
    "order" INT DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (stack_id) REFERENCES stacks(id) ON DELETE CASCADE,
    FOREIGN KEY (infrastructure_id) REFERENCES infrastructures(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_stack_resources_stack_id ON stack_resources(stack_id);
CREATE INDEX IF NOT EXISTS idx_stack_resources_infrastructure_id ON stack_resources(infrastructure_id);
CREATE INDEX IF NOT EXISTS idx_stack_resources_order ON stack_resources("order");

-- Create stack_templates table
CREATE TABLE IF NOT EXISTS stack_templates (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    category VARCHAR(50),
    is_public BOOLEAN DEFAULT FALSE,
    user_id VARCHAR(36),
    spec JSONB NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_stack_templates_category ON stack_templates(category);
CREATE INDEX IF NOT EXISTS idx_stack_templates_is_public ON stack_templates(is_public);
CREATE INDEX IF NOT EXISTS idx_stack_templates_user_id ON stack_templates(user_id);

-- Create stack_operations table (tracks stack operations)
CREATE TABLE IF NOT EXISTS stack_operations (
    id VARCHAR(36) PRIMARY KEY,
    stack_id VARCHAR(36) NOT NULL,
    operation_type VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    started_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP,
    error_message TEXT,
    details JSONB,
    FOREIGN KEY (stack_id) REFERENCES stacks(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_stack_operations_stack_id ON stack_operations(stack_id);
CREATE INDEX IF NOT EXISTS idx_stack_operations_status ON stack_operations(status);
CREATE INDEX IF NOT EXISTS idx_stack_operations_user_id ON stack_operations(user_id);

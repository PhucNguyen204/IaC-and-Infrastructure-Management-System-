# Stack Management System

## Overview

Stack là một **logical grouping** của các infrastructure resources để phục vụ một mục đích cụ thể. Ví dụ: "tenant-123-prod", "project-abc-staging".

### Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                         STACK                                │
│  (my-web-app-prod)                                          │
│                                                              │
│  ┌────────────────┐  ┌────────────────┐  ┌───────────────┐ │
│  │ POSTGRES_      │  │ DOCKER_        │  │ NGINX_        │ │
│  │ INSTANCE       │  │ SERVICE        │  │ GATEWAY       │ │
│  │ (database)     │──│ (app)          │──│ (gateway)     │ │
│  │                │  │                │  │               │ │
│  │ Role: database │  │ Role: app      │  │ Role: gateway │ │
│  │ Order: 1       │  │ Order: 2       │  │ Order: 3      │ │
│  └────────────────┘  └────────────────┘  └───────────────┘ │
│         │                    │                   │          │
│         └────────────────────┴───────────────────┘          │
│                    Dependencies                              │
└─────────────────────────────────────────────────────────────┘
```

## Resource Types

Stack hỗ trợ các resource types sau:

1. **NGINX_GATEWAY** - Reverse proxy / API Gateway
2. **POSTGRES_INSTANCE** - PostgreSQL database instance (1 instance + nhiều databases)
3. **POSTGRES_DATABASE** - Logical database trong một POSTGRES_INSTANCE
4. **POSTGRES_CLUSTER** - PostgreSQL HA cluster (primary + replicas)
5. **DOCKER_SERVICE** - Container service

## API Endpoints

### Stack Management

#### 1. Create Stack
**POST** `/api/v1/stacks`

Tạo một stack mới với nhiều resources.

**Request:**
```json
{
  "name": "my-web-app-prod",
  "description": "Production web application",
  "environment": "prod",
  "project_id": "project-001",
  "tenant_id": "tenant-123",
  "tags": ["web", "production"],
  "resources": [
    {
      "type": "POSTGRES_INSTANCE",
      "role": "database",
      "name": "app-database",
      "spec": {
        "plan": "medium"
      },
      "depends_on": [],
      "order": 1
    },
    {
      "type": "DOCKER_SERVICE",
      "role": "app",
      "name": "web-app",
      "spec": {
        "image": "myapp",
        "image_tag": "v1.0",
        "service_type": "web",
        "ports": [
          {
            "container_port": 3000,
            "host_port": 8080,
            "protocol": "tcp"
          }
        ],
        "networks": [
          {
            "network_id": "iaas_iaas-network",
            "alias": "web-app"
          }
        ],
        "plan": "small"
      },
      "depends_on": ["app-database"],
      "order": 2
    },
    {
      "type": "NGINX_GATEWAY",
      "role": "gateway",
      "name": "api-gateway",
      "spec": {
        "plan": "small",
        "domains": [
          {
            "domain": "myapp.com",
            "port": 80
          }
        ]
      },
      "depends_on": ["web-app"],
      "order": 3
    }
  ]
}
```

**Response:**
```json
{
  "success": true,
  "code": "SUCCESS",
  "message": "Stack created successfully",
  "data": {
    "id": "stack-uuid",
    "name": "my-web-app-prod",
    "status": "running",
    "environment": "prod",
    "resources": [
      {
        "id": "resource-uuid",
        "infrastructure_id": "infra-uuid",
        "resource_type": "POSTGRES_INSTANCE",
        "resource_name": "app-database",
        "role": "database",
        "status": "running",
        "outputs": {
          "connection_string": "postgresql://postgres:postgres@172.18.0.5:5432",
          "ip_address": "172.18.0.5"
        }
      }
    ]
  }
}
```

#### 2. Get Stack
**GET** `/api/v1/stacks/:id`

Lấy thông tin chi tiết của một stack.

#### 3. List Stacks
**GET** `/api/v1/stacks?page=1&page_size=20`

Liệt kê tất cả stacks của user.

#### 4. Update Stack
**PUT** `/api/v1/stacks/:id`

Cập nhật metadata hoặc resources của stack.

```json
{
  "name": "my-web-app-prod-v2",
  "description": "Updated description",
  "tags": ["web", "production", "v2"]
}
```

#### 5. Delete Stack
**DELETE** `/api/v1/stacks/:id`

Xóa stack và tất cả resources trong đó (theo reverse order).

### Stack Operations

#### Start Stack
**POST** `/api/v1/stacks/:id/start`

Khởi động tất cả resources trong stack (theo order).

#### Stop Stack
**POST** `/api/v1/stacks/:id/stop`

Dừng tất cả resources trong stack (reverse order).

#### Restart Stack
**POST** `/api/v1/stacks/:id/restart`

Restart tất cả resources trong stack.

### Stack Templates

#### Create Template
**POST** `/api/v1/stack-templates`

Tạo template để tái sử dụng stack configuration.

```json
{
  "name": "Standard Web App Template",
  "description": "Template for web applications",
  "category": "web-app",
  "is_public": true,
  "resources": [
    {
      "type": "POSTGRES_INSTANCE",
      "role": "database",
      "name": "database",
      "spec": {
        "plan": "small"
      },
      "order": 1
    }
  ]
}
```

#### List Public Templates
**GET** `/api/v1/stack-templates`

Liệt kê tất cả public templates.

#### Get Template
**GET** `/api/v1/stack-templates/:id`

Lấy chi tiết một template.

## Key Features

### 1. Dependency Management
Resources có thể khai báo `depends_on` để đảm bảo thứ tự tạo:
- App depends on Database
- Gateway depends on App

### 2. Automatic Resource Injection
Khi tạo Docker service với dependency là PostgreSQL:
- Tự động inject `DATABASE_HOST` env variable
- Connection string được tạo tự động

### 3. Ordered Creation/Deletion
Resources được tạo theo `order` field (ascending).
Resources được xóa theo reverse order (descending).

### 4. Resource Outputs
Mỗi resource cung cấp outputs:
- **POSTGRES_INSTANCE**: `connection_string`, `ip_address`
- **DOCKER_SERVICE**: `internal_endpoint`, `ip_address`
- **NGINX_GATEWAY**: `internal_endpoint`, `ip_address`
- **POSTGRES_CLUSTER**: `primary_endpoint`, `replica_endpoints`

### 5. Stack Operations Tracking
Mọi operation (CREATE, UPDATE, DELETE, START, STOP) đều được track:
- Operation ID
- Status: PENDING → IN_PROGRESS → COMPLETED/FAILED
- Start time, completion time
- Error messages nếu có

## Use Cases

### Use Case 1: Simple Web Application
```
Stack: my-app-dev
├── POSTGRES_INSTANCE (database)
├── DOCKER_SERVICE (app)
└── NGINX_GATEWAY (gateway)
```

### Use Case 2: Microservices Application
```
Stack: microservices-prod
├── POSTGRES_CLUSTER (shared database)
├── DOCKER_SERVICE (user-service)
├── DOCKER_SERVICE (order-service)
├── DOCKER_SERVICE (payment-service)
└── NGINX_GATEWAY (api-gateway)
```

### Use Case 3: Development Environment Clone
```
# Clone from production stack
POST /api/v1/stacks/clone
{
  "source_stack_id": "prod-stack-id",
  "name": "dev-clone",
  "environment": "dev"
}
```

## Database Schema

### Tables

1. **stacks** - Stack records
   - id, name, description, environment
   - project_id, tenant_id, user_id
   - status, tags
   - created_at, updated_at

2. **stack_resources** - Links stacks to infrastructures
   - id, stack_id, infrastructure_id
   - resource_type, role
   - depends_on (JSONB), order
   - created_at, updated_at

3. **stack_templates** - Reusable templates
   - id, name, description, category
   - is_public, user_id
   - spec (JSONB)
   - created_at, updated_at

4. **stack_operations** - Operation tracking
   - id, stack_id, operation_type
   - status, user_id
   - started_at, completed_at
   - error_message, details (JSONB)

## Benefits

1. **Simplified Management**: Quản lý theo "cụm" thay vì từng resource riêng lẻ
2. **Consistency**: Đảm bảo tất cả resources trong stack được tạo/xóa cùng nhau
3. **Reusability**: Templates cho phép tái sử dụng configurations
4. **Environment Cloning**: Dễ dàng clone stack từ prod → staging → dev
5. **Dependency Resolution**: Tự động resolve dependencies giữa resources
6. **Audit Trail**: Track tất cả operations trên stack
7. **Tenant Isolation**: Stack có thể thuộc về project/tenant cụ thể

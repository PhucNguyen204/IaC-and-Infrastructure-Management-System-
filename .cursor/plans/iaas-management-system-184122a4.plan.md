<!-- 184122a4-761c-4d25-bde2-c00e4cac08fc f07ba7fc-0864-4031-a61e-0068b000d173 -->
# Phase 2: Infrastructure Monitoring Service + PostgreSQL Cluster

## Nhóm 1: Monitoring - Metric Collectors & Kafka Consumer

### 1.1 Docker Collector Implementation

**File**: `vcs-infrastructure-monitoring-service/pkg/collectors/docker_collector.go`

- Implement `dockerCollector` struct với Docker client
- Function `CollectContainerStats`: Parse Docker stats JSON, tính CPU/Memory/Network/Disk I/O
- Sử dụng logic tương tự `postgres_service.go` GetPostgreSQLStats

### 1.2 PostgreSQL Metrics Collector

**File**: `vcs-infrastructure-monitoring-service/pkg/collectors/postgres_collector.go`

- Interface `IPostgreSQLCollector`
- Connect vào PostgreSQL container qua `psql` hoặc `pg_stat_*` views
- Collect metrics: active connections, transactions, commits/rollbacks, replication lag
- Struct `PostgreSQLMetrics` với các field tương ứng

### 1.3 Nginx Metrics Collector

**File**: `vcs-infrastructure-monitoring-service/pkg/collectors/nginx_collector.go`

- Interface `INginxCollector`
- Parse nginx status page (`/nginx_status`)
- Collect: active connections, requests, reading/writing/waiting
- Struct `NginxMetrics`

### 1.4 Kafka Consumer Hoàn Chỉnh

**File**: `vcs-infrastructure-monitoring-service/infrastructures/kafka/consumer.go`

- Update `StartConsuming` để subscribe topics: `postgres.*`, `nginx.*`
- Parse Kafka event từ provisioning service
- Index event vào Elasticsearch qua `esClient.IndexLog`
- Handle error và retry logic

### 1.5 Unit Tests - Collectors & Kafka

**Files**:

- `pkg/collectors/docker_collector_test.go`
- `pkg/collectors/postgres_collector_test.go`
- `pkg/collectors/nginx_collector_test.go`
- `infrastructures/kafka/consumer_test.go`
- Mock Docker client, database connections, Kafka messages
- Coverage target: >90%

**Testing**: Start docker-compose, verify Kafka messages được consume và index vào Elasticsearch

---

## Nhóm 2: Monitoring - APIs Hoàn Chỉnh

### 2.1 Update Metrics Service

**File**: `vcs-infrastructure-monitoring-service/usecases/services/metrics_service.go`

- Add `AggregateMetrics(ctx, instanceID, timeRange, aggregationType)` - aggregate metrics (avg, max, min)
- Update `GetCurrentMetrics` để include PostgreSQL/Nginx specific metrics

### 2.2 Update Health Check Service

**File**: `vcs-infrastructure-monitoring-service/usecases/services/health_check_service.go`

- Add alert logic khi status = unhealthy (log alert hoặc send event)
- Update health check để include PostgreSQL/Nginx specific checks

### 2.3 Infrastructure List API

**File**: `vcs-infrastructure-monitoring-service/api/http/monitoring_handler.go`

- Update `ListInfrastructure` để query từ PostgreSQL metadata database
- Connect tới database của provisioning service hoặc shared database
- Join với Redis để lấy current status

### 2.4 DTOs Update

**File**: `vcs-infrastructure-monitoring-service/dto/monitoring.go`

- Add `PostgreSQLMetricsResponse`, `NginxMetricsResponse`
- Add `AggregatedMetricsResponse`

### 2.5 Unit Tests - Services & Handlers

**Files**:

- `usecases/services/metrics_service_test.go`
- `usecases/services/health_check_service_test.go`
- `api/http/monitoring_handler_test.go` (update existing)
- Coverage target: >90%

**Testing**: Test từng API:

- `GET /api/v1/metrics/{instance_id}` 
- `GET /api/v1/metrics/{instance_id}/history`
- `GET /api/v1/logs/{instance_id}`
- `GET /api/v1/health/{instance_id}`
- `GET /api/v1/infrastructure`

---

## Nhóm 3: PostgreSQL Cluster - Setup & Basic Operations

### 3.1 DTOs cho Cluster

**File**: `vcs-infrastructure-provisioning-service/dto/postgres_cluster.go`

- `CreateClusterRequest`: node_count, version, resources, replication_config
- `ClusterInfoResponse`: cluster topology, primary node, replicas, replication lag
- `ClusterNodeInfo`: node details
- `ScaleClusterRequest`, `FailoverRequest`

### 3.2 Repositories

**File**: `vcs-infrastructure-provisioning-service/usecases/repositories/postgres_cluster_repository.go`

- Interface `IPostgreSQLClusterRepository`
- CRUD operations cho `PostgreSQLCluster`, `ClusterNode`, `EtcdNode`
- Functions: Create, FindByID, Update, Delete, ListNodes

### 3.3 Cluster Service - Docker Setup

**File**: `vcs-infrastructure-provisioning-service/usecases/services/postgres_cluster_service.go`

- Interface `IPostgreSQLClusterService`
- Helper functions:
- `createEtcdCluster`: Deploy 3 etcd nodes
- `createPatroniNode`: Deploy Patroni container với PostgreSQL
- `createHAProxy`: Deploy HAProxy container
- `setupPatroniConfig`: Generate Patroni YAML config
- `setupHAProxyConfig`: Generate HAProxy config

### 3.4 Cluster Service - Create Cluster

**Function**: `CreateCluster(ctx, req)`

1. Create Infrastructure entity
2. Create PostgreSQLCluster entity
3. Create network: `iaas-cluster-{cluster_id}`
4. Deploy 3 etcd nodes
5. Deploy primary Patroni node (role=primary)
6. Deploy N replica Patroni nodes (role=replica)
7. Deploy HAProxy
8. Wait for cluster initialization
9. Publish Kafka event: `postgres.cluster.created`

### 3.5 Cluster Service - Lifecycle Operations

**Functions**:

- `StartCluster`: Start tất cả containers (etcd → patroni → haproxy)
- `StopCluster`: Stop tất cả containers
- `RestartCluster`: Restart từng node theo thứ tự
- `DeleteCluster`: Remove tất cả containers, volumes, network
- `GetClusterInfo`: Query database + Docker inspect, return topology

### 3.6 Unit Tests - Cluster Basic Operations

**File**: `usecases/services/postgres_cluster_service_test.go`

- Test CreateCluster, StartCluster, StopCluster, DeleteCluster, GetClusterInfo
- Mock Docker service, repositories, Kafka producer
- Coverage target: >90%

### 3.7 HTTP Handlers - Cluster Basic

**File**: `vcs-infrastructure-provisioning-service/api/http/postgres_cluster_handler.go`

- Routes:
- `POST /api/v1/postgres/cluster` - CreateCluster
- `GET /api/v1/postgres/cluster/{id}` - GetClusterInfo
- `POST /api/v1/postgres/cluster/{id}/start` - StartCluster
- `POST /api/v1/postgres/cluster/{id}/stop` - StopCluster
- `POST /api/v1/postgres/cluster/{id}/restart` - RestartCluster
- `DELETE /api/v1/postgres/cluster/{id}` - DeleteCluster

### 3.8 Unit Tests - Handlers

**File**: `api/http/postgres_cluster_handler_test.go`

- Test từng handler với mock service
- Coverage target: >90%

**Testing**: Test từng API với docker-compose:

- Create cluster với 1 primary + 2 replicas
- Get cluster info
- Start/Stop/Restart cluster
- Delete cluster

---

## Nhóm 4: PostgreSQL Cluster - Advanced Operations

### 4.1 Scale Cluster

**Function**: `ScaleCluster(ctx, clusterID, action, nodeCount)`

- Action: `add` hoặc `remove`
- Add: Deploy thêm replica nodes, update cluster config
- Remove: Remove replica nodes (không remove primary), update config
- Publish Kafka event: `postgres.cluster.scaled`

### 4.2 Failover

**Function**:

- `TriggerFailover(ctx, clusterID, newPrimaryNodeID)` - Manual failover
- `MonitorAutoFailover(ctx, clusterID)` - Check Patroni auto-failover status
- Update primary node trong database
- Publish Kafka event: `postgres.cluster.failover`

### 4.3 Backup & Restore

**Functions**:

- `BackupCluster(ctx, clusterID, req)` - Backup từ primary node, tương tự single node
- `RestoreCluster(ctx, clusterID, req)` - Restore vào primary, replicas sẽ sync
- Publish Kafka events

### 4.4 Cluster Metrics & Logs

**Functions**:

- `GetClusterStats(ctx, clusterID)` - Stats cho tất cả nodes
- `GetClusterLogs(ctx, clusterID)` - Logs từ tất cả nodes
- Return aggregated và per-node data

### 4.5 HTTP Handlers - Advanced

**File**: Update `postgres_cluster_handler.go`

- Routes:
- `POST /api/v1/postgres/cluster/{id}/scale` - ScaleCluster
- `POST /api/v1/postgres/cluster/{id}/failover` - TriggerFailover
- `POST /api/v1/postgres/cluster/{id}/backup` - BackupCluster
- `POST /api/v1/postgres/cluster/{id}/restore` - RestoreCluster
- `GET /api/v1/postgres/cluster/{id}/stats` - GetClusterStats
- `GET /api/v1/postgres/cluster/{id}/logs` - GetClusterLogs

### 4.6 Unit Tests - Advanced Operations

**File**: Update `postgres_cluster_service_test.go`

- Test Scale, Failover, Backup/Restore, Metrics, Logs
- Coverage target: >90%

### 4.7 Unit Tests - Advanced Handlers

**File**: Update `postgres_cluster_handler_test.go`

- Test handlers cho advanced operations
- Coverage target: >90%

**Testing**: Test từng API với docker-compose:

- Scale up (add 1 replica)
- Scale down (remove 1 replica)
- Trigger manual failover
- Backup cluster
- Restore cluster
- Get cluster stats
- Get cluster logs

---

## Nhóm 5: Integration Testing & Coverage Report

### 5.1 Run All Unit Tests

- Run tests cho monitoring service: `go test ./... -cover`
- Run tests cho provisioning service cluster: `go test ./... -cover`
- Verify coverage >90% cho tất cả packages

### 5.2 Integration Testing Script

**File**: `test_phase2_apis.ps1`

- Test monitoring APIs với real Elasticsearch/Kafka
- Test cluster APIs với real Docker containers
- Verify Kafka events được consume
- Verify metrics/logs được index vào Elasticsearch

### 5.3 Update docker-compose.yml

- Add etcd image configuration
- Add Patroni image configuration (sử dụng `patroni` Docker image)
- Verify tất cả services có thể communicate

### 5.4 Phase 2 Completion Report

- Tạo `PHASE2_COMPLETION_REPORT.md`
- Document tất cả APIs, test results, coverage report

### To-dos

- [ ] Tạo entities và DTOs cho PostgreSQL single node
- [ ] Viết unit tests cho provisioning service
- [ ] Implement Elasticsearch integration cho logs và metrics indexing
- [ ] Implement metric collectors (Docker stats, PostgreSQL metrics)
- [ ] Implement Kafka consumer để consume infrastructure events
- [ ] Implement health check service với periodic checks
- [ ] Implement REST API handlers cho metrics, logs, health status
- [ ] Viết unit tests cho monitoring service
- [ ] Tạo entities và DTOs cho PostgreSQL cluster
- [ ] Implement PostgreSQL cluster service với Patroni + etcd + HAProxy
- [ ] Implement cluster scaling và failover operations
- [ ] Implement REST API handlers cho PostgreSQL cluster operations
- [ ] Tạo entities và DTOs cho Nginx
- [ ] Implement REST API handlers cho Nginx operations
- [ ] Tích hợp Authentication với Provisioning và Monitoring services
- [ ] Viết integration và E2E tests cho toàn bộ system
# IaaS Platform Architecture Design

## 📊 System Overview

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              IaaS Platform                                       │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                  │
│  ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐              │
│  │   Frontend      │    │  Auth Service   │    │  Provisioning   │              │
│  │   (React)       │───▶│  (8082)         │───▶│  Service (8083) │              │
│  │                 │    │                 │    │                 │              │
│  └─────────────────┘    └─────────────────┘    └────────┬────────┘              │
│                                                          │                       │
│                                                          │ Kafka Events          │
│                                                          ▼                       │
│                              ┌───────────────────────────────────────┐          │
│                              │           Apache Kafka                 │          │
│                              │  Topics:                               │          │
│                              │  • infrastructure.lifecycle            │          │
│                              │  • infrastructure.metrics              │          │
│                              └───────────────┬───────────────────────┘          │
│                                              │                                   │
│                    ┌─────────────────────────┼─────────────────────────┐        │
│                    │                         │                         │        │
│                    ▼                         ▼                         ▼        │
│  ┌─────────────────────┐   ┌─────────────────────┐   ┌─────────────────────┐   │
│  │   Health Check      │   │   Metrics           │   │   Monitoring        │   │
│  │   Service (8085)    │   │   Collector         │   │   Service (8084)    │   │
│  │                     │   │   (Background)      │   │                     │   │
│  │  • Consume events   │   │                     │   │  • Query ES         │   │
│  │  • Health probes    │   │  • Docker stats     │   │  • Build dashboard  │   │
│  │  • Index to ES      │   │  • System metrics   │   │  • Calculate uptime │   │
│  └──────────┬──────────┘   └──────────┬──────────┘   └──────────┬──────────┘   │
│             │                         │                         │               │
│             └─────────────────────────┼─────────────────────────┘               │
│                                       ▼                                          │
│                         ┌───────────────────────────┐                           │
│                         │      Elasticsearch        │                           │
│                         │  Indices:                 │                           │
│                         │  • infra-events-*         │                           │
│                         │  • infra-metrics-*        │                           │
│                         │  • infra-health-*         │                           │
│                         └───────────────────────────┘                           │
│                                                                                  │
│  ┌─────────────────┐    ┌─────────────────┐                                     │
│  │   PostgreSQL    │    │     Redis       │                                     │
│  │   (Metadata)    │    │   (Cache/Lock)  │                                     │
│  └─────────────────┘    └─────────────────┘                                     │
└─────────────────────────────────────────────────────────────────────────────────┘
```

## 🎯 Service Responsibilities

### 1. Provisioning Service (8083)
- Create/Delete/Start/Stop infrastructure
- Publish lifecycle events to Kafka
- CRUD operations on PostgreSQL

### 2. Health Check Service (8085) - NEW
- Consume lifecycle events from Kafka
- Perform periodic health checks on containers
- Collect container metrics (CPU, Memory, Network, Disk)
- Index events & metrics to Elasticsearch

### 3. Monitoring Service (8084)
- Query Elasticsearch for events/metrics
- Calculate uptime statistics
- Build dashboards and charts
- Provide analytics API

## 📨 Kafka Topics Design

### Topic: `infrastructure.lifecycle`
```json
{
  "event_id": "uuid",
  "infrastructure_id": "uuid",
  "cluster_id": "uuid",
  "user_id": "string",
  "type": "postgres_cluster | nginx_cluster | dind | clickhouse",
  "action": "created | started | stopped | deleted | node_added | node_removed | failover | backup | restore",
  "status": "running | stopped | failed | deleted",
  "previous_status": "string",
  "timestamp": "2025-12-15T07:00:00Z",
  "metadata": {
    "cluster_name": "mytest",
    "node_count": 3,
    "version": "17",
    "triggered_by": "user | system | auto_failover"
  }
}
```

### Topic: `infrastructure.metrics`
```json
{
  "metric_id": "uuid",
  "infrastructure_id": "uuid",
  "container_id": "string",
  "container_name": "string",
  "type": "postgres_cluster | nginx_cluster",
  "timestamp": "2025-12-15T07:00:00Z",
  "cpu": {
    "usage_percent": 25.5,
    "cores": 2
  },
  "memory": {
    "used_bytes": 536870912,
    "limit_bytes": 1073741824,
    "usage_percent": 50.0
  },
  "network": {
    "rx_bytes": 1024000,
    "tx_bytes": 512000,
    "rx_packets": 1000,
    "tx_packets": 500
  },
  "disk": {
    "read_bytes": 10240000,
    "write_bytes": 5120000
  },
  "health": {
    "status": "healthy | unhealthy | unknown",
    "last_check": "2025-12-15T07:00:00Z",
    "message": "All checks passed"
  }
}
```

## 🗄️ Elasticsearch Indices

### Index: `infra-events-YYYY.MM.DD`
- Lifecycle events (created, started, stopped, deleted)
- Node events (added, removed, failover)
- Used for audit trail and uptime calculation

### Index: `infra-metrics-YYYY.MM.DD`
- Container metrics (CPU, Memory, Network, Disk)
- Collected every 30 seconds
- Retained for 30 days

### Index: `infra-health-YYYY.MM.DD`
- Health check results
- Container status changes
- Used for SLA reporting

## 🔴 Redis Usage

### Purpose: Minimal, Essential Only

1. **Distributed Locks** (for health check coordination)
   - Key: `lock:healthcheck:{infrastructure_id}`
   - TTL: 30 seconds

2. **Rate Limiting** (API protection)
   - Key: `ratelimit:{user_id}:{endpoint}`
   - TTL: 1 minute

3. **Session/Token Cache** (Auth service only)
   - Key: `session:{user_id}`
   - TTL: 24 hours

### NOT Using Redis For:
- ❌ Metrics storage (use ES)
- ❌ Event queuing (use Kafka)
- ❌ Infrastructure state (use PostgreSQL)

## 🐘 PostgreSQL Schema

### Tables:
```sql
-- Core infrastructure metadata
infrastructures (id, name, type, user_id, status, created_at, updated_at)

-- Cluster-specific tables
postgre_sql_clusters (id, infrastructure_id, version, node_count, ...)
nginx_clusters (id, infrastructure_id, node_count, ...)
clickhouse_clusters (id, infrastructure_id, ...)

-- Node tracking
cluster_nodes (id, cluster_id, container_id, role, status, ...)

-- NOT storing in PostgreSQL:
-- ❌ Metrics (too much data, use ES)
-- ❌ Logs (use ES)
-- ❌ Events history (use ES)
```

## 🏗️ Health Check Service Design

### Folder Structure:
```
vcs-healthcheck-service/
├── cmd/
│   └── main.go
├── api/
│   └── http/
│       └── health_handler.go
├── dto/
│   ├── events.go
│   └── metrics.go
├── entities/
│   └── health.go
├── infrastructures/
│   ├── docker/
│   │   └── stats_collector.go
│   ├── kafka/
│   │   ├── consumer.go
│   │   └── producer.go
│   └── elasticsearch/
│       └── client.go
├── usecases/
│   └── services/
│       ├── health_checker.go
│       └── metrics_collector.go
├── pkg/
│   ├── env/
│   └── logger/
├── config.yaml
├── Dockerfile
└── go.mod
```

### Core Components:

1. **Event Consumer**: Listen to `infrastructure.lifecycle`
2. **Health Checker**: Periodic container health probes
3. **Metrics Collector**: Docker stats collection
4. **ES Indexer**: Push to Elasticsearch

## 📈 Monitoring Service Enhancements

### New APIs:

```
GET /api/v1/dashboard/overview
  - Total infrastructures by type
  - Overall uptime percentage
  - Active alerts count

GET /api/v1/dashboard/metrics/{infrastructure_id}
  - CPU/Memory charts (time series)
  - Network I/O graphs
  - Disk usage trends

GET /api/v1/uptime/summary
  - Uptime by infrastructure type
  - SLA compliance percentage
  - Downtime incidents list

GET /api/v1/uptime/{infrastructure_id}/timeline
  - Status change history
  - Calculated uptime percentage
  - Incident details
```

## 🔄 Event Flow

```
1. User creates cluster via API
   └─▶ Provisioning Service
       └─▶ Create containers
       └─▶ Save to PostgreSQL
       └─▶ Publish to Kafka: infrastructure.lifecycle (action=created)

2. Health Check Service (running continuously)
   └─▶ Consume from Kafka
   └─▶ Start monitoring new infrastructure
   └─▶ Every 30s: Collect Docker stats
   └─▶ Index to Elasticsearch

3. User views dashboard
   └─▶ Monitoring Service
       └─▶ Query Elasticsearch
       └─▶ Aggregate metrics
       └─▶ Return charts/statistics
```

## 🎛️ Simplified docker-compose.yml

Key changes:
1. Remove Zookeeper (use KRaft mode for Kafka)
2. Minimize Redis usage
3. Add Health Check Service
4. Cleaner environment variables

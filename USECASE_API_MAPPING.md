# API Coverage Analysis - PostgreSQL Infrastructure Usecases

## Usecase Coverage Summary

| Usecase | Required APIs | Status | Missing APIs |
|---------|---------------|--------|--------------|
| UC1: Provision Single (dev/staging) | Create, Get, Lifecycle | ✅ COMPLETE | None |
| UC2: Single (production) | Create, Backup, Resize, Read Replica | ⚠️ PARTIAL | Resize, Read Replica for Single |
| UC3: Provision Cluster HA | Create, Get, Lifecycle, Failover | ⚠️ PARTIAL | Auto-failover, Promote, Replication Status |
| UC4: Multi-tenant | Tenant DB/Schema Management | ❌ MISSING | Create Tenant DB, List Tenants, Delete Tenant |
| UC5: Read Replica | Add Replica, Remove Replica, Replication Status | ⚠️ PARTIAL | Replication Status (TODO) |
| UC6: Backup/Restore/PITR | Backup, Restore, PITR, List Backups | ⚠️ PARTIAL | PITR, List Backups |
| UC7: Scale | Vertical Scale, Horizontal Scale | ⚠️ PARTIAL | Vertical Scale (CPU/RAM change) |
| UC8: Upgrade & Migration | Version Upgrade, Single→Cluster Migration | ❌ MISSING | All upgrade/migration APIs |
| UC9: Monitoring & Lifecycle | Metrics, Quota, Pause/Resume | ⚠️ PARTIAL | Quota management, Pause/Resume |

---

## Detailed API Analysis

### ✅ POSTGRES_SINGLE APIs (Already Implemented)

| API | Endpoint | Usecase Coverage |
|-----|----------|------------------|
| Create Single | POST `/api/v1/postgres/single` | UC1, UC2 |
| Get Single Info | GET `/api/v1/postgres/single/:id` | UC1, UC2 |
| Start Single | POST `/api/v1/postgres/single/:id/start` | UC1, UC2, UC9 |
| Stop Single | POST `/api/v1/postgres/single/:id/stop` | UC1, UC2, UC9 |
| Restart Single | POST `/api/v1/postgres/single/:id/restart` | UC1, UC2 |
| Delete Single | DELETE `/api/v1/postgres/single/:id` | UC1, UC2, UC9 |
| Get Stats | GET `/api/v1/postgres/single/:id/stats` | UC9 |
| Get Logs | GET `/api/v1/postgres/single/:id/logs` | UC9 |
| Backup | POST `/api/v1/postgres/single/:id/backup` | UC2, UC6 |

**Coverage**: 9/9 APIs ✅

---

### ⚠️ POSTGRES_CLUSTER APIs (Partially Implemented)

| API | Endpoint | Status | Usecase Coverage |
|-----|----------|--------|------------------|
| Create Cluster | POST `/api/v1/postgres/cluster` | ✅ DONE | UC3, UC5 |
| Get Cluster Info | GET `/api/v1/postgres/cluster/:id` | ✅ DONE | UC3, UC5 |
| Start Cluster | POST `/api/v1/postgres/cluster/:id/start` | ✅ DONE | UC3, UC9 |
| Stop Cluster | POST `/api/v1/postgres/cluster/:id/stop` | ✅ DONE | UC3, UC9 |
| Restart Cluster | POST `/api/v1/postgres/cluster/:id/restart` | ✅ DONE | UC3 |
| Delete Cluster | DELETE `/api/v1/postgres/cluster/:id` | ✅ DONE | UC3, UC9 |
| Scale Cluster | POST `/api/v1/postgres/cluster/:id/scale` | ✅ DONE | UC5, UC7 |
| Get Stats | GET `/api/v1/postgres/cluster/:id/stats` | ✅ DONE | UC9 |
| Get Logs | GET `/api/v1/postgres/cluster/:id/logs` | ⚠️ ISSUE | UC9 |
| Promote Replica | POST `/api/v1/postgres/cluster/:id/promote` | ⏳ TODO | UC3, UC5 |
| Replication Status | GET `/api/v1/postgres/cluster/:id/replication/status` | ⏳ TODO | UC3, UC5 |

**Coverage**: 9/11 APIs (81.8%)

---

## ❌ Missing APIs for Complete Usecase Support

### Priority 1: Critical for Production

#### 1. **UC2 - Resize Single Instance**
```
POST /api/v1/postgres/single/:id/resize
Body: {
  "cpu": 4,
  "memory": 2048,
  "storage": 500
}
```
**Usecase**: Scale CPU/RAM/Storage cho single instance production

---

#### 2. **UC3 - Promote Replica (Manual Failover)**
```
POST /api/v1/postgres/cluster/:id/promote
Body: {
  "node_id": "replica-node-id"
}
```
**Usecase**: Manual failover khi primary có vấn đề

---

#### 3. **UC3 - Get Replication Status**
```
GET /api/v1/postgres/cluster/:id/replication/status
Response: {
  "primary": "node-primary",
  "replicas": [
    {
      "node_name": "replica-1",
      "lag_bytes": 0,
      "lag_seconds": 0.02,
      "state": "streaming",
      "is_healthy": true
    }
  ]
}
```
**Usecase**: Monitoring replication health

---

#### 4. **UC6 - List Backups**
```
GET /api/v1/postgres/single/:id/backups
GET /api/v1/postgres/cluster/:id/backups
Response: {
  "backups": [
    {
      "backup_id": "...",
      "timestamp": "2025-11-19T10:00:00Z",
      "size_mb": 1024,
      "type": "full",
      "status": "completed"
    }
  ]
}
```
**Usecase**: Xem lịch sử backup trước khi restore

---

#### 5. **UC6 - Restore from Backup**
```
POST /api/v1/postgres/single/:id/restore
POST /api/v1/postgres/cluster/:id/restore
Body: {
  "backup_id": "...",
  "restore_mode": "overwrite|new_instance",
  "pitr_target": "2025-11-19T10:30:00Z" // optional
}
```
**Usecase**: Restore DB từ backup hoặc PITR

---

### Priority 2: Multi-tenant Support

#### 6. **UC4 - Create Tenant Database**
```
POST /api/v1/postgres/cluster/:id/tenants
Body: {
  "tenant_id": "tenant-abc",
  "mode": "database_per_tenant|schema_per_tenant",
  "initial_user": "tenant_abc_user",
  "initial_password": "...",
  "quota": {
    "max_storage_mb": 10240,
    "max_connections": 50
  }
}
```
**Usecase**: Tạo database/schema riêng cho tenant mới

---

#### 7. **UC4 - List Tenants**
```
GET /api/v1/postgres/cluster/:id/tenants
Response: {
  "tenants": [
    {
      "tenant_id": "tenant-abc",
      "database_name": "tenant_abc",
      "user": "tenant_abc_user",
      "created_at": "...",
      "quota_used": {
        "storage_mb": 256,
        "connections": 12
      }
    }
  ]
}
```
**Usecase**: Quản lý danh sách tenant trên cluster

---

#### 8. **UC4 - Delete Tenant**
```
DELETE /api/v1/postgres/cluster/:id/tenants/:tenant_id
Query: backup_before_delete=true
```
**Usecase**: Xóa tenant database/schema (với backup tự động)

---

### Priority 3: Advanced Operations

#### 9. **UC2 - Add Read Replica to Single**
```
POST /api/v1/postgres/single/:id/replicas
Body: {
  "replica_count": 1,
  "cpu": 1,
  "memory": 512
}
```
**Usecase**: Thêm read replica cho single instance production

---

#### 10. **UC8 - Upgrade PostgreSQL Version**
```
POST /api/v1/postgres/single/:id/upgrade
POST /api/v1/postgres/cluster/:id/upgrade
Body: {
  "target_version": "16",
  "mode": "in_place|blue_green"
}
```
**Usecase**: Nâng cấp version PostgreSQL

---

#### 11. **UC8 - Migrate Single to Cluster**
```
POST /api/v1/postgres/migrate
Body: {
  "source_id": "single-instance-id",
  "target_type": "cluster",
  "node_count": 3,
  "cutover_mode": "manual|automatic"
}
```
**Usecase**: Migrate từ single sang cluster khi app phát triển

---

#### 12. **UC9 - Pause/Resume Instance**
```
POST /api/v1/postgres/single/:id/pause
POST /api/v1/postgres/single/:id/resume
```
**Usecase**: Tạm dừng instance để tiết kiệm chi phí

---

#### 13. **UC9 - Set Quota**
```
PUT /api/v1/postgres/single/:id/quota
PUT /api/v1/postgres/cluster/:id/quota
Body: {
  "max_storage_mb": 102400,
  "max_connections": 200,
  "alert_thresholds": {
    "storage_percent": 80,
    "connections_percent": 90
  }
}
```
**Usecase**: Thiết lập giới hạn resource và cảnh báo

---

## Implementation Priority

### Phase 1: Complete Core Cluster Features (1-2 days)
- ✅ Scale Cluster (DONE)
- ⏳ Promote Replica (manual failover)
- ⏳ Replication Status
- ⏳ Fix Get Logs issue

### Phase 2: Backup/Restore (2-3 days)
- List Backups API
- Restore from Backup API
- PITR support
- Backup scheduling

### Phase 3: Scaling & Resize (1-2 days)
- Resize Single Instance (CPU/RAM/Storage)
- Add Read Replica to Single
- Vertical scaling for Cluster nodes

### Phase 4: Multi-tenant (2-3 days)
- Create Tenant Database/Schema
- List Tenants
- Delete Tenant with backup
- Quota management per tenant

### Phase 5: Advanced Operations (3-4 days)
- Version Upgrade
- Single → Cluster Migration
- Pause/Resume
- Quota management
- Advanced monitoring

---

## Next Steps

1. **Immediate**: Implement Promote Replica and Replication Status (complete UC3 & UC5)
2. **Short term**: Backup/Restore APIs (complete UC6)
3. **Medium term**: Multi-tenant support (complete UC4)
4. **Long term**: Migration and upgrade features (complete UC8)

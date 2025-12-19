# PostgreSQL HA Cluster - Tài liệu Kỹ thuật

## 🏗️ KIẾN TRÚC POSTGRESQL HA CLUSTER

### Thành phần chính

```
┌─────────────────────────────────────────────────────────────┐
│                    PostgreSQL HA Cluster                     │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│   ┌─────────┐     ┌─────────┐     ┌─────────┐               │
│   │ HAProxy │────▶│  etcd   │     │         │               │
│   │:5000/01 │     │ :2379   │     │         │               │
│   └────┬────┘     └────┬────┘     │         │               │
│        │               │          │         │               │
│   ┌────▼────┐     ┌────▼────┐     ┌────▼────┐              │
│   │Patroni  │◀───▶│Patroni  │◀───▶│Patroni  │              │
│   │+Postgres│     │+Postgres│     │+Postgres│              │
│   │ PRIMARY │     │ REPLICA │     │ REPLICA │              │
│   │ :5432   │     │ :5432   │     │ :5432   │              │
│   └─────────┘     └─────────┘     └─────────┘              │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### Vai trò từng thành phần

| Thành phần | Vai trò |
|------------|---------|
| **etcd** | Distributed key-value store - lưu trạng thái cluster, leader election |
| **Patroni** | Cluster manager - quản lý failover, replication, configuration |
| **PostgreSQL** | Database engine - lưu trữ và xử lý dữ liệu |
| **HAProxy** | Load balancer - điều phối traffic đến Primary/Replica |

---

## 🔄 CÁCH ĐỒNG BỘ DỮ LIỆU HOẠT ĐỘNG

### Streaming Replication (WAL-based)

PostgreSQL sử dụng **Write-Ahead Logging (WAL)** để đồng bộ dữ liệu:

```
┌─────────────┐     WAL Stream      ┌─────────────┐
│   PRIMARY   │ ═══════════════════▶│   REPLICA   │
│             │     (liên tục)      │             │
│  Data Files │                     │  Data Files │
│     WAL     │                     │     WAL     │
└─────────────┘                     └─────────────┘
```

**Giải thích:**
1. Mọi thay đổi (INSERT, UPDATE, DELETE) được ghi vào WAL trước
2. WAL được stream liên tục đến các Replica qua TCP
3. Replica apply WAL để có dữ liệu giống Primary

### Cấu hình trong patroni.yml

```yaml
postgresql:
  parameters:
    wal_level: replica           # Bật WAL logging cho replication
    hot_standby: on              # Cho phép replica phục vụ read queries
    max_wal_senders: 20          # Tối đa 20 connections gửi WAL
    max_replication_slots: 20    # 20 replication slots
    archive_mode: on             # Lưu WAL để recovery
```

---

## 🆕 QUY TRÌNH KHI NODE MỚI JOIN VÀO CLUSTER

### Bước 1: Tạo Container với Environment Variables

Code trong `postgres_cluster_service.go` - hàm `AddNode()`:

```go
env := []string{
    fmt.Sprintf("SCOPE=%s", scope),           // Cluster name - để Patroni biết thuộc cluster nào
    fmt.Sprintf("NAMESPACE=%s", namespace),   // Namespace trong etcd
    fmt.Sprintf("PATRONI_NAME=%s", nodeName), // Tên node unique
    fmt.Sprintf("ETCD_HOST=%s", etcdHost),    // etcd để lấy cluster state
    fmt.Sprintf("POSTGRES_PASSWORD=%s", cluster.Password),
    "REPLICATION_PASSWORD=replicator_pass",   // Password cho replication user
    "CLONEFROM=true",                         // Cho phép clone từ node khác
    "PGDATA=/data/patroni",
}
```

### Bước 2: Container Start và Chạy entrypoint.sh

```bash
#!/bin/bash
# 1. Đợi etcd sẵn sàng
wait_for_etcd

# 2. Tạo pgpass file cho replication authentication
cat > /opt/secretpg/pgpass <<PGPASS_EOF
*:5432:*:replicator:${REPLICATION_PASSWORD:-replicator_pass}
*:5432:*:postgres:${POSTGRES_PASSWORD}
PGPASS_EOF
chmod 600 /opt/secretpg/pgpass

# 3. Generate patroni.yml từ environment variables
cat > /etc/patroni/patroni.yml <<EOF
scope: "${SCOPE}"
namespace: "${NAMESPACE:-percona_lab}"
name: "${PATRONI_NAME}"
...
EOF

# 4. Start Patroni
exec patroni /etc/patroni/patroni.yml
```

### Bước 3: Patroni Tự động Clone Dữ liệu từ Primary

Khi Patroni phát hiện PGDATA rỗng, nó chạy `pg_basebackup`:

```bash
# Script tự động được Patroni gọi
export PGPASSWORD="${REPLICATION_PASSWORD:-replicator_pass}"

# Lấy địa chỉ Primary từ Patroni
MASTER_HOST=$(echo "${PATRONI_MASTER_CONNECT_ADDRESS}" | cut -d: -f1)
MASTER_PORT=$(echo "${PATRONI_MASTER_CONNECT_ADDRESS}" | cut -d: -f2)

# Clone toàn bộ database từ Primary
/usr/lib/postgresql/17/bin/pg_basebackup \
  -h "${MASTER_HOST}" \
  -p "${MASTER_PORT:-5432}" \
  -U replicator \           # User có quyền replication
  -D "${PGDATA}" \          # Thư mục data của replica
  -X stream \               # Stream WAL trong quá trình backup
  -c fast \                 # Checkpoint nhanh
  -R \                      # Tự động tạo standby.signal và config
  -v
```

### Bước 4: PostgreSQL Start ở chế độ Replica

Sau khi `pg_basebackup` hoàn thành:
- Tạo file `standby.signal` (đánh dấu đây là replica)
- Cấu hình `primary_conninfo` trong `postgresql.auto.conf`
- PostgreSQL start và bắt đầu streaming WAL từ Primary

---

## 📊 FLOW DIAGRAM CHI TIẾT

```
1. Container start
   │
   ▼
2. entrypoint.sh chạy
   │
   ▼
3. Đợi etcd ready (wait_for_etcd)
   │
   ▼
4. Generate /etc/patroni/patroni.yml từ ENV vars
   │
   ▼
5. Patroni start
   │
   ▼
6. Patroni kết nối etcd, đọc cluster state
   │  - Tìm ai là Primary
   │  - Lấy connection info của Primary
   │
   ▼
7. Patroni phát hiện PGDATA rỗng → chạy pg_basebackup
   │  - Kết nối đến Primary:5432
   │  - User: replicator / Password: replicator_pass
   │  - Clone toàn bộ database
   │
   ▼
8. pg_basebackup hoàn thành
   │  - Tạo standby.signal file
   │  - Cấu hình primary_conninfo trong postgresql.auto.conf
   │
   ▼
9. PostgreSQL start ở chế độ REPLICA
   │  - Kết nối streaming replication đến Primary
   │  - Bắt đầu nhận WAL liên tục
   │
   ▼
10. Patroni đăng ký node mới vào etcd
    │  - Node xuất hiện trong cluster member list
    │  - HAProxy tự động phát hiện và thêm vào backend
```

---

## ⚙️ CẤU HÌNH QUAN TRỌNG ĐỂ REPLICA ĐỌC ĐƯỢC DỮ LIỆU

### 1. pg_hba.conf - Cho phép Replication Connection

```
# Cho phép user replicator kết nối từ mọi IP để replication
host    replication     replicator   0.0.0.0/0    scram-sha-256

# Cho phép tất cả users kết nối từ mọi IP
host    all             all          0.0.0.0/0    scram-sha-256
```

### 2. Replication User

```yaml
# Trong patroni.yml bootstrap
users:
  replicator:
    password: ${REPLICATION_PASSWORD}
    options:
      - replication    # Quyền đặc biệt cho streaming replication
```

### 3. PostgreSQL Parameters

```yaml
parameters:
  wal_level: replica              # BẮT BUỘC: Log đủ thông tin cho replication
  hot_standby: on                 # BẮT BUỘC: Cho phép replica xử lý read queries
  max_wal_senders: 20             # Số connection gửi WAL đồng thời
  max_replication_slots: 20       # Số slots cho replicas
  wal_log_hints: on               # Cần cho pg_rewind khi failover
  max_wal_size: '10GB'            # Giữ đủ WAL cho replica catch up
```

### 4. File được tạo trên Replica sau pg_basebackup

```ini
# postgresql.auto.conf (tự động tạo)
primary_conninfo = 'host=pg-cluster-node1 port=5432 user=replicator password=replicator_pass'
```

```
# standby.signal (file rỗng)
# Chỉ cần file này tồn tại, PostgreSQL biết chạy ở chế độ standby
```

---

## 🔐 AUTHENTICATION FLOW

```
┌────────────────┐                        ┌────────────────┐
│    REPLICA     │                        │    PRIMARY     │
├────────────────┤                        ├────────────────┤
│                │  1. TCP Connect        │                │
│ pg_basebackup  │───────────────────────▶│  Port 5432     │
│                │                        │                │
│                │  2. SCRAM-SHA-256 Auth │                │
│ user=replicator│◀──────────────────────▶│ pg_hba check   │
│ pass=replicator│                        │                │
│                │                        │                │
│                │  3. REPLICATION Stream │                │
│                │◀═══════════════════════│ Send WAL       │
│                │     (liên tục)         │                │
└────────────────┘                        └────────────────┘
```

---

## 📋 KIỂM TRA REPLICATION STATUS

### Trên PRIMARY

```sql
-- Xem các replica đang kết nối
SELECT 
    client_addr,
    state,
    sent_lsn,
    write_lsn,
    flush_lsn,
    replay_lsn,
    sync_state
FROM pg_stat_replication;
```

### Trên REPLICA

```sql
-- Xem trạng thái streaming
SELECT 
    status,
    received_lsn,
    latest_end_lsn,
    sender_host,
    sender_port
FROM pg_stat_wal_receiver;
```

---

## 🎯 TÓM TẮT FLOW KHI ADD NODE

| Bước | Thành phần | Hành động |
|------|------------|-----------|
| 1 | Go Service | Gọi `AddNode()` → tạo container với ENV vars |
| 2 | Docker | Start container với network + volumes |
| 3 | entrypoint.sh | Đợi etcd → generate patroni.yml |
| 4 | Patroni | Query etcd → tìm Primary address |
| 5 | Patroni | Chạy pg_basebackup → clone data |
| 6 | PostgreSQL | Start ở standby mode → streaming WAL |
| 7 | Patroni | Đăng ký vào etcd → HAProxy phát hiện |

---

## 📁 CẤU TRÚC FILE QUAN TRỌNG

```
vcs-infrastructure-provisioning-service/
├── docker/
│   ├── patroni/
│   │   ├── Dockerfile          # Build Patroni + PostgreSQL 17 image
│   │   ├── entrypoint.sh       # Script khởi tạo và start Patroni
│   │   └── patroni.yml         # Template cấu hình Patroni
│   ├── etcd/
│   │   ├── Dockerfile
│   │   └── entrypoint.sh
│   └── haproxy/
│       ├── Dockerfile
│       ├── haproxy.cfg
│       └── docker-entrypoint.sh
└── usecases/services/
    └── postgres_cluster_service.go   # Logic tạo/quản lý cluster
```

---

## 🔧 COMMANDS THƯỜNG DÙNG

```bash
# Kiểm tra trạng thái Patroni cluster
docker exec <container> patronictl -c /etc/patroni/patroni.yml list

# Xem logs Patroni
docker logs <container>

# Kết nối PostgreSQL
docker exec -it <container> psql -U postgres

# Kiểm tra replication lag
docker exec <container> psql -U postgres -c "SELECT * FROM pg_stat_replication;"
```

---

*Tài liệu được tạo tự động từ source code dự án VCS IaaS*


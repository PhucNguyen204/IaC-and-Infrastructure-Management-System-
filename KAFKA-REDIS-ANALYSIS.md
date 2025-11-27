# Phân tích Kafka và Redis trong Dự án IaaS Platform

## 📊 Tổng quan Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    IaaS Platform Architecture                    │
└─────────────────────────────────────────────────────────────────┘

┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│ Auth Service │     │ Provisioning │     │  Monitoring  │
│   (8082)     │────▶│   Service    │────▶│   Service    │
│              │     │    (8083)    │     │   (8084)     │
└──────┬───────┘     └──────┬───────┘     └──────┬───────┘
       │                    │                     │
       │ Redis              │ Kafka               │ Kafka
       │ (Session)          │ (Events)            │ (Consumer)
       │                    │                     │
       ▼                    ▼                     ▼
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│    Redis     │     │    Kafka     │     │ Elasticsearch│
│   (6379)     │◀────│   (9092)     │────▶│   (9200)     │
└──────────────┘     └──────────────┘     └──────────────┘
       ▲                    ▲
       │                    │
       │ Cache              │ Events
       │                    │
┌──────┴──────────────────┬─┴────────────────────┐
│         Docker Events    │   Infrastructure     │
│      (Container Lifecycle)   (State Changes)    │
└──────────────────────────────────────────────────┘
```

---

## 🔴 REDIS - Distributed Cache & Session Store

### 1. Vai trò chính

Redis được sử dụng làm:
- **Session Storage** - Lưu trữ refresh tokens
- **Cache Layer** - Cache thông tin cluster để giảm tải database
- **Fast Read/Write** - Key-value store nhanh chóng

### 2. Sử dụng trong Authentication Service

**File:** `vcs-authentication-service/usecases/services/auth.go`

#### a) Lưu Refresh Token
```go
// Khi user login thành công
func (s *authService) Login(ctx context.Context, username, password string) (string, string, error) {
    // ... validate user ...
    
    accessToken, _ := s.generateAccessToken(user.ID, scopes)
    refreshToken, _ := s.generateRefreshToken()
    
    // LƯU refresh token vào Redis với TTL 7 ngày
    s.redisClient.Set(ctx, "refresh:"+refreshToken, user.ID, time.Hour*24*7)
    
    return accessToken, refreshToken, nil
}
```

**Key Pattern:** `refresh:{token}` → `user_id`  
**TTL:** 7 ngày  
**Purpose:** Cho phép renew access token mà không cần login lại

#### b) Refresh Access Token
```go
func (s *authService) RefreshAccessToken(ctx context.Context, refreshToken string) (string, error) {
    // ĐỌC user_id từ Redis
    userId, err := s.redisClient.Get(ctx, "refresh:"+refreshToken)
    if err != nil {
        return "", err // Token expired hoặc không tồn tại
    }
    
    // Generate access token mới
    accessToken, _ := s.generateAccessToken(userId, scopes)
    return accessToken, nil
}
```

#### c) Invalidate Token khi đổi password
```go
func (s *authService) UpdatePassword(ctx context.Context, userId, currentPassword, newPassword string) error {
    // ... update password ...
    
    // XÓA refresh token khỏi Redis
    s.redisClient.Del(ctx, "refresh:"+user.ID)
    
    return nil
}
```

**Tại sao dùng Redis?**
- ⚡ **Fast lookup** - O(1) time complexity
- ⏰ **Auto expiration** - TTL tự động xóa token cũ
- 🔒 **Secure** - Token không lưu trong database chính
- 📈 **Scalable** - Dễ dàng scale horizontal

---

### 3. Sử dụng trong Provisioning Service

**File:** `vcs-infrastructure-provisioning-service/usecases/services/cache_service.go`

#### a) Cache Cluster Information
```go
type ICacheService interface {
    GetClusterInfo(ctx context.Context, clusterID string) (*dto.ClusterInfoResponse, bool)
    SetClusterInfo(ctx context.Context, clusterID string, info *dto.ClusterInfoResponse, ttl time.Duration) error
    InvalidateCluster(ctx context.Context, clusterID string) error
}
```

**Cache Keys:**
- `cluster:info:{cluster_id}` - Thông tin cluster (nodes, endpoints, status)
- `cluster:stats:{cluster_id}` - Số liệu thống kê (CPU, RAM, connections)
- `cluster:replication:{cluster_id}` - Replication status (primary, replicas, lag)

#### b) Flow sử dụng Cache

```go
// 1. Đọc từ cache trước
func (s *Service) GetClusterInfo(clusterID string) (*ClusterInfo, error) {
    // Check cache
    if cached, found := s.cache.GetClusterInfo(ctx, clusterID); found {
        return cached, nil  // ✅ Cache HIT - nhanh!
    }
    
    // Cache MISS - query database
    info := s.queryDatabase(clusterID)
    
    // Lưu vào cache với TTL 5 phút
    s.cache.SetClusterInfo(ctx, clusterID, info, 5*time.Minute)
    
    return info, nil
}
```

#### c) Cache Invalidation
```go
// Khi cluster thay đổi → xóa cache
func (s *cacheService) InvalidateCluster(ctx context.Context, clusterID string) error {
    keys := []string{
        fmt.Sprintf("cluster:info:%s", clusterID),
        fmt.Sprintf("cluster:stats:%s", clusterID),
        fmt.Sprintf("cluster:replication:%s", clusterID),
    }
    return s.redis.Del(ctx, keys...).Err()
}
```

**Khi nào invalidate cache?**
- ✅ Cluster start/stop/restart
- ✅ Node thêm/xóa
- ✅ Configuration thay đổi
- ✅ Failover xảy ra

**Performance Improvement:**
```
Without Cache:  
  Database Query: ~50-200ms
  Complex JOIN: ~100-500ms

With Redis Cache:
  Cache Hit: ~1-5ms (100x faster!)
  TTL: 5 minutes
  Cache Hit Rate: ~80-90%
```

---

## 🟢 KAFKA - Event Streaming Platform

### 1. Vai trò chính

Kafka được sử dụng làm:
- **Event Bus** - Truyền tải events giữa các services
- **Decoupling** - Tách biệt services (loose coupling)
- **Async Processing** - Xử lý bất đồng bộ
- **Event Sourcing** - Lưu lại lịch sử các sự kiện

### 2. Architecture Kafka trong Dự án

```
┌─────────────────────────────────────────────────────────────┐
│                     Kafka Architecture                       │
└─────────────────────────────────────────────────────────────┘

Docker Events          User Actions
     │                      │
     ▼                      ▼
┌────────────────────────────────────┐
│   Provisioning Service (Producer)   │
│   - Publish infrastructure events   │
└─────────────┬──────────────────────┘
              │
              ▼
      ┌─────────────┐
      │   Kafka     │  Topic: infrastructure.events
      │  (Broker)   │  Partitions: Auto
      └──────┬──────┘  Replication: 1
             │
             ├──────────────────────────────────┐
             ▼                                  ▼
┌────────────────────────┐      ┌──────────────────────────┐
│  Monitoring Service    │      │  Provisioning Service    │
│  (Consumer)            │      │  (Consumer)              │
│  → Elasticsearch       │      │  → Cache Invalidation    │
└────────────────────────┘      └──────────────────────────┘
```

### 3. Producer - Provisioning Service

**File:** `vcs-infrastructure-provisioning-service/infrastructures/kafka/producer.go`

#### a) Kafka Producer Configuration
```go
type kafkaProducer struct {
    writer *kafka.Writer
}

func NewKafkaProducer(env env.KafkaEnv, logger logger.ILogger) IKafkaProducer {
    writer := &kafka.Writer{
        Addr:                   kafka.TCP(env.Brokers...),
        Topic:                  env.Topic,               // infrastructure.events
        Balancer:               &kafka.Hash{},           // Hash by key
        AllowAutoTopicCreation: true,                    // Auto create topic
        Async:                  true,                    // ⚡ Async publish
        BatchSize:              100,                     // Batch 100 messages
        BatchTimeout:           10 * time.Millisecond,   // hoặc 10ms
        RequiredAcks:           1,                       // Leader ack
        MaxAttempts:            3,                       // Retry 3 lần
        Compression:            kafka.Snappy,            // Nén dữ liệu
    }
    return &kafkaProducer{writer: writer}
}
```

**Tại sao Async = true?**
- ⚡ **Non-blocking** - Không đợi Kafka confirm
- 🚀 **High throughput** - Có thể publish hàng ngàn events/giây
- 📦 **Batching** - Gom nhiều messages lại gửi 1 lần

#### b) Event Structure
```go
type InfrastructureEvent struct {
    InstanceID string                 // Cluster/Instance ID
    UserID     string                 // User thực hiện action
    Type       string                 // "postgres", "nginx", "k8s"
    Action     string                 // "created", "started", "stopped", "deleted"
    Timestamp  time.Time              // Thời gian event
    Metadata   map[string]interface{} // Thông tin bổ sung
}
```

#### c) Publish Event
```go
func (kp *kafkaProducer) PublishEvent(ctx context.Context, event InfrastructureEvent) error {
    event.Timestamp = time.Now()
    
    eventBytes, _ := json.Marshal(event)
    
    msg := kafka.Message{
        Key:   []byte(event.InstanceID),  // Key = instance_id (same key → same partition)
        Value: eventBytes,
        Time:  event.Timestamp,
    }
    
    // Gửi đến Kafka (async)
    if err := kp.writer.WriteMessages(ctx, msg); err != nil {
        return err
    }
    
    return nil
}
```

### 4. Docker Event Listener → Kafka

**File:** `vcs-infrastructure-provisioning-service/usecases/services/docker_event_listener_service.go`

**Flow hoạt động:**

```
1. Docker Engine phát ra event (container start/stop/die)
        ↓
2. Docker Event Listener Service lắng nghe
        ↓
3. Parse event → Cập nhật database status
        ↓
4. Publish event lên Kafka
        ↓
5. Kafka broadcast đến các consumers
```

#### Code:
```go
func (s *dockerEventListenerService) handleEvent(ctx context.Context, event events.Message) {
    containerID := event.ID
    action := string(event.Action)
    
    // Map Docker action → Infrastructure status
    var status entities.InfrastructureStatus
    switch event.Action {
    case events.ActionStart:
        status = entities.StatusRunning
    case events.ActionStop, events.ActionDie:
        status = entities.StatusStopped
    case events.ActionDestroy:
        status = entities.StatusDeleted
    }
    
    // Cập nhật database
    infra, _ := s.infraRepo.FindByContainerID(ctx, containerID)
    infra.Status = status
    s.infraRepo.Update(infra)
    
    // 🔥 PUBLISH LÊN KAFKA
    kafkaEvent := kafka.InfrastructureEvent{
        InstanceID: infra.ID,
        UserID:     infra.UserID,
        Type:       string(infra.Type),  // "postgres", "nginx"
        Action:     action,               // "start", "stop", "die"
        Timestamp:  time.Now(),
        Metadata: map[string]interface{}{
            "container_id":   containerID,
            "container_name": containerName,
            "status":         string(status),
        },
    }
    
    s.kafkaProducer.PublishEvent(ctx, kafkaEvent)
    
    // Bonus: Broadcast qua WebSocket (real-time UI update)
    if s.wsBroadcaster != nil {
        s.wsBroadcaster.BroadcastUpdate(update)
    }
}
```

**Các loại events được publish:**
- `postgres.created` - PostgreSQL cluster/instance tạo mới
- `postgres.started` - PostgreSQL start
- `postgres.stopped` - PostgreSQL stop
- `postgres.deleted` - PostgreSQL xóa
- `nginx.created` - Nginx cluster tạo mới
- `nginx.started` - Nginx start
- ... tương tự cho K8s, Docker services

### 5. Consumer - Monitoring Service

**File:** `vcs-infrastructure-monitoring-service/infrastructures/kafka/consumer.go`

#### a) Kafka Consumer Configuration
```go
func NewKafkaConsumer(env env.KafkaEnv, esClient elasticsearch.IElasticsearchClient) IKafkaConsumer {
    readers := make([]*kafka.Reader, 0)
    
    // Subscribe nhiều topics
    topics := []string{
        "postgres.created", "postgres.started", "postgres.stopped",
        "nginx.created", "nginx.started", "nginx.stopped",
        // ... more topics
    }
    
    for _, topic := range topics {
        reader := kafka.NewReader(kafka.ReaderConfig{
            Brokers: env.Brokers,            // kafka:9092
            GroupID: "monitoring-consumer-group",  // Consumer group
            Topic:   topic,
        })
        readers = append(readers, reader)
    }
    
    return &kafkaConsumer{
        readers:  readers,
        esClient: esClient,
    }
}
```

#### b) Consume Messages và Index vào Elasticsearch
```go
func (kc *kafkaConsumer) consumeMessages(ctx context.Context, reader *kafka.Reader) {
    for {
        // Đọc message từ Kafka
        msg, err := reader.ReadMessage(ctx)
        if err != nil {
            continue
        }
        
        // Parse event
        var event InfrastructureEvent
        json.Unmarshal(msg.Value, &event)
        
        // Tạo log entry
        logEntry := elasticsearch.LogEntry{
            InstanceID: event.InstanceID,
            UserID:     event.UserID,
            Type:       event.Type,
            Action:     event.Action,
            Message:    fmt.Sprintf("%s %s", event.Type, event.Action),
            Level:      "info",
            Metadata:   event.Metadata,
        }
        
        // 📊 INDEX VÀO ELASTICSEARCH
        kc.esClient.IndexLog(ctx, logEntry)
    }
}
```

**Flow:**
```
Kafka Event → Monitoring Service → Elasticsearch → Kibana Dashboard
```

### 6. Consumer - Cache Invalidation

**File:** `vcs-infrastructure-provisioning-service/infrastructures/kafka/consumer.go`

```go
func (c *eventConsumer) handleEvent(ctx context.Context, event InfrastructureEvent) error {
    // Khi có event thay đổi cluster
    // → Invalidate Redis cache
    
    if event.Action == "started" || 
       event.Action == "stopped" || 
       event.Action == "deleted" {
        
        // Xóa cache của cluster này
        return c.cache.InvalidateCluster(ctx, event.InstanceID)
    }
    
    return nil
}
```

**Lợi ích:**
- ✅ Cache luôn consistent với database
- ✅ Không cần manual invalidation
- ✅ Event-driven architecture

---

## 🔄 So sánh Redis vs Kafka

| Tiêu chí | Redis | Kafka |
|----------|-------|-------|
| **Purpose** | Cache, Session Store | Event Streaming |
| **Data Type** | Key-Value | Message Stream |
| **Persistence** | In-memory (+ RDB/AOF) | Disk (Log) |
| **Speed** | Cực nhanh (~1ms) | Nhanh (~5-10ms) |
| **Retention** | TTL (giây → giờ) | Days/Weeks/Forever |
| **Pattern** | Request/Response | Pub/Sub |
| **Use Case** | Fast reads, Cache | Event processing, Logging |
| **Scalability** | Vertical | Horizontal (partitions) |

---

## 📈 Metrics & Monitoring

### Redis Metrics cần theo dõi:
```bash
# Memory usage
redis-cli INFO memory

# Cache hit rate
hits = GET commands that found the key
misses = GET commands that didn't find the key
hit_rate = hits / (hits + misses)

# Eviction count
redis-cli INFO stats | grep evicted_keys
```

**Target:**
- Hit Rate: > 80%
- Eviction Rate: < 5%
- Memory Usage: < 70%

### Kafka Metrics cần theo dõi:
```bash
# Consumer lag (얼마나 chậm trễ)
kafka-consumer-groups --bootstrap-server kafka:9092 \
  --describe --group monitoring-consumer-group

# Messages per second
kafka-run-class kafka.tools.GetOffsetShell \
  --broker-list kafka:9092 \
  --topic infrastructure.events
```

**Target:**
- Consumer Lag: < 100 messages
- Publish Latency: < 10ms
- Processing Time: < 50ms

---

## 🎯 Best Practices

### Redis:
1. **Set appropriate TTL** - Không để cache cũ quá lâu
2. **Use pipeline** - Batch multiple commands
3. **Monitor memory** - Tránh out of memory
4. **Use connection pool** - Tái sử dụng connections
5. **Compression** - Nén data lớn trước khi cache

### Kafka:
1. **Use async producer** - Non-blocking publish
2. **Batch messages** - Giảm network overhead
3. **Compression** - Snappy hoặc LZ4
4. **Monitor consumer lag** - Đảm bảo consumer kịp xử lý
5. **Partition by key** - Đảm bảo ordering
6. **Set retention policy** - Xóa message cũ tự động

---

## 🚀 Kết luận

### Redis trong dự án:
✅ **Authentication:** Session storage cho refresh tokens (7 ngày TTL)  
✅ **Provisioning:** Cache cluster info/stats/replication (5 phút TTL)  
✅ **Performance:** Giảm 80-90% database queries  

### Kafka trong dự án:
✅ **Event Bus:** Truyền tải infrastructure events giữa services  
✅ **Monitoring:** Stream events vào Elasticsearch để analytics  
✅ **Cache Sync:** Trigger cache invalidation khi có thay đổi  
✅ **Decoupling:** Services không phụ thuộc trực tiếp vào nhau  

**Architecture Pattern:** Event-Driven Microservices với Cache Layer

```
User Request → Service → Database
                ↓
            Kafka Event → [Monitoring, Cache Invalidation, Webhooks, ...]
                ↓
            Redis Cache → Fast Response
```


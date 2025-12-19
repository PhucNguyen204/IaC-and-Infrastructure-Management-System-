package docs

// ============================================================================
// CÂU HỎI PHỎNG VẤN CHUYÊN SÂU VỀ KAFKA TRONG DỰ ÁN IaaS
// ============================================================================
//
// File này chứa các câu hỏi phỏng vấn về Kafka được sử dụng trong hệ thống
// IaaS (Infrastructure as a Service), bao gồm cả các câu hỏi về mở rộng
// và tối ưu hóa trong tương lai.
//
// ============================================================================

// ============================================================================
// PHẦN 1: KIẾN TRÚC VÀ THIẾT KẾ HIỆN TẠI
// ============================================================================

/*
Câu hỏi 1.1: Kiến trúc Kafka hiện tại
- Trong dự án này, Kafka được sử dụng với mục đích gì?
- Tại sao chọn Kafka thay vì các message queue khác (RabbitMQ, Redis Pub/Sub)?
- Hiện tại dự án sử dụng single broker hay cluster? Giải thích lý do.
- Zookeeper được sử dụng để làm gì? Có cần thiết không với Kafka version hiện tại?

Câu hỏi 1.2: Event-Driven Architecture
- Dự án sử dụng pattern nào: Event Sourcing, CQRS, hay Event-Driven?
- Các service nào đóng vai trò Producer và Consumer?
- Làm thế nào để đảm bảo event ordering trong hệ thống hiện tại?
- Có sử dụng Saga pattern để quản lý distributed transactions không?

Câu hỏi 1.3: Topic Design
- Hiện tại có bao nhiêu topics? Cách đặt tên topics như thế nào?
- Tại sao sử dụng nhiều topics (postgres.created, postgres.started, etc.) 
  thay vì một topic duy nhất với partition key?
- Có sử dụng topic compaction không? Khi nào nên sử dụng?
- Làm thế nào để quản lý lifecycle của topics (creation, deletion, retention)?
*/

// ============================================================================
// PHẦN 2: PRODUCER IMPLEMENTATION
// ============================================================================

/*
Câu hỏi 2.1: Producer Configuration
- Tại sao sử dụng Hash balancer thay vì RoundRobin hoặc LeastBytes?
- Async mode được sử dụng với mục đích gì? Có trade-off gì không?
- BatchSize=100 và BatchTimeout=10ms có phù hợp không? Làm sao tối ưu?
- RequiredAcks=RequireOne có đủ đảm bảo reliability không? Khi nào cần RequireAll?

Câu hỏi 2.2: Message Key và Partitioning
- Tại sao sử dụng InstanceID làm message key?
- Làm thế nào để đảm bảo messages của cùng một infrastructure 
  được gửi vào cùng partition?
- Có vấn đề gì nếu InstanceID không phân bố đều? Làm sao giải quyết?
- Khi nào nên sử dụng null key thay vì specific key?

Câu hỏi 2.3: Error Handling và Retry
- Hiện tại có retry mechanism không? MaxAttempts=3 có đủ không?
- WriteBackoffMin/Max được thiết lập như thế nào? Có cần exponential backoff?
- Khi Kafka broker down, producer sẽ làm gì? Có dead letter queue không?
- Làm thế nào để handle message too large errors?

Câu hỏi 2.4: Compression và Performance
- Snappy compression được chọn vì lý do gì? So sánh với gzip, lz4?
- Async mode có thể gây mất message không? Làm sao đảm bảo?
- Có monitoring producer metrics không? (throughput, latency, error rate)
- Làm thế nào để optimize producer throughput cho high-volume events?
*/

// ============================================================================
// PHẦN 3: CONSUMER IMPLEMENTATION
// ============================================================================

/*
Câu hỏi 3.1: Consumer Group và Parallelism
- Tại sao sử dụng Consumer Group? GroupID được đặt như thế nào?
- Có bao nhiêu consumer instances trong cùng một group?
- Làm thế nào để scale consumers khi load tăng?
- Có vấn đề gì nếu số consumers > số partitions?

Câu hỏi 3.2: Message Processing
- Consumer xử lý message như thế nào? Có batching không?
- Có sử dụng manual commit offset không? Tại sao?
- Làm thế nào để handle duplicate messages?
- Có idempotency check trong consumer không?

Câu hỏi 3.3: Error Handling
- Khi consumer xử lý message fail, sẽ làm gì?
- Có retry mechanism trong consumer không?
- Làm thế nào để handle poison messages?
- Có dead letter topic không? Khi nào nên sử dụng?

Câu hỏi 3.4: Multiple Topics Consumption
- Consumer đọc từ nhiều topics cùng lúc. Làm thế nào quản lý?
- Có vấn đề gì về ordering khi đọc từ nhiều topics?
- Làm thế nào để đảm bảo processing order cho messages từ cùng topic?
- Có cần separate consumer groups cho từng topic không?
*/

// ============================================================================
// PHẦN 4: RELIABILITY VÀ CONSISTENCY
// ============================================================================

/*
Câu hỏi 4.1: Message Delivery Guarantees
- Hệ thống đảm bảo at-least-once hay exactly-once delivery?
- Làm thế nào để đảm bảo không mất message?
- Có vấn đề duplicate messages không? Làm sao handle?
- Khi nào cần exactly-once semantics? Có trade-off gì?

Câu hỏi 4.2: Transaction và Idempotency
- Có sử dụng Kafka Transactions không? Khi nào cần?
- Làm thế nào để đảm bảo idempotency trong producer?
- Consumer có cần idempotent processing không?
- Làm thế nào để handle out-of-order messages?

Câu hỏi 4.3: Data Consistency
- Làm thế nào đảm bảo consistency giữa Kafka và Elasticsearch?
- Có sử dụng two-phase commit không?
- Khi Elasticsearch indexing fails, message có bị mất không?
- Làm thế nào để handle eventual consistency?
*/

// ============================================================================
// PHẦN 5: PERFORMANCE VÀ SCALABILITY
// ============================================================================

/*
Câu hỏi 5.1: Throughput Optimization
- Làm thế nào để tăng producer throughput?
- Consumer có bottleneck nào không? Làm sao optimize?
- Có sử dụng compression ở consumer side không?
- Batch processing vs individual processing: trade-off?

Câu hỏi 5.2: Latency Optimization
- Làm thế nào để giảm end-to-end latency?
- Có trade-off giữa latency và throughput không?
- BatchTimeout=10ms có ảnh hưởng latency không?
- Làm thế nào để handle real-time requirements?

Câu hỏi 5.3: Partitioning Strategy
- Làm thế nào quyết định số partitions cho một topic?
- Có thể thay đổi số partitions sau khi tạo topic không?
- Làm thế nào để rebalance partitions khi scale?
- Có vấn đề gì về partition skew không?

Câu hỏi 5.4: Resource Management
- Kafka producer/consumer sử dụng bao nhiêu memory?
- Có connection pooling không?
- Làm thế nào để handle backpressure?
- Có rate limiting mechanism không?
*/

// ============================================================================
// PHẦN 6: MONITORING VÀ OBSERVABILITY
// ============================================================================

/*
Câu hỏi 6.1: Metrics và Monitoring
- Có monitoring Kafka metrics không? (lag, throughput, error rate)
- Làm thế nào để detect consumer lag?
- Có alerting khi consumer lag quá cao không?
- Làm thế nào để monitor producer performance?

Câu hỏi 6.2: Logging và Tracing
- Có distributed tracing cho Kafka messages không?
- Làm thế nào để trace một message từ producer đến consumer?
- Có correlation ID trong messages không?
- Làm thế nào để debug message flow issues?

Câu hỏi 6.3: Health Checks
- Làm thế nào để check Kafka broker health?
- Producer có health check endpoint không?
- Consumer có health check không?
- Làm thế nào để detect và handle broker failures?
*/

// ============================================================================
// PHẦN 7: MỞ RỘNG VÀ TƯƠNG LAI
// ============================================================================

/*
Câu hỏi 7.1: Kafka Cluster Setup
- Khi nào cần chuyển từ single broker sang cluster?
- Làm thế nào để setup Kafka cluster với replication?
- Cần bao nhiêu brokers cho high availability?
- Làm thế nào để handle broker failures trong cluster?

Câu hỏi 7.2: Schema Registry
- Có cần Schema Registry không? Khi nào?
- Làm thế nào để handle schema evolution?
- Có backward/forward compatibility requirements không?
- Làm thế nào để versioning message schemas?

Câu hỏi 7.3: Kafka Streams
- Có cần Kafka Streams cho real-time processing không?
- Làm thế nào để aggregate events trong real-time?
- Có cần stateful processing không?
- Làm thế nào để handle windowing operations?

Câu hỏi 7.4: Kafka Connect
- Có cần Kafka Connect để integrate với external systems không?
- Làm thế nào để sync data với databases?
- Có cần CDC (Change Data Capture) không?
- Làm thế nào để handle connector failures?

Câu hỏi 7.5: Multi-Region Deployment
- Làm thế nào để deploy Kafka across multiple regions?
- Có cần mirroring giữa các regions không?
- Làm thế nào để handle network partitions?
- Có cần active-active hay active-passive setup?

Câu hỏi 7.6: Security và Compliance
- Có cần Kafka security (SASL, SSL/TLS) không?
- Làm thế nào để implement access control (ACLs)?
- Có cần encryption at rest không?
- Làm thế nào để audit Kafka access?

Câu hỏi 7.7: Advanced Features
- Có cần exactly-once semantics không?
- Làm thế nào để implement event sourcing với Kafka?
- Có cần CQRS pattern với Kafka không?
- Làm thế nào để handle event replay?
*/

// ============================================================================
// PHẦN 8: TROUBLESHOOTING VÀ BEST PRACTICES
// ============================================================================

/*
Câu hỏi 8.1: Common Issues
- Làm thế nào để handle consumer lag spikes?
- Khi nào xảy ra rebalancing? Làm sao minimize?
- Làm thế nào để handle partition leader election?
- Có vấn đề gì về message ordering khi rebalance?

Câu hỏi 8.2: Data Retention và Cleanup
- Retention policy được thiết lập như thế nào?
- Làm thế nào để cleanup old messages?
- Có cần log compaction cho state topics không?
- Làm thế nào để handle disk space issues?

Câu hỏi 8.3: Testing
- Làm thế nào để test Kafka producers/consumers?
- Có integration tests với embedded Kafka không?
- Làm thế nào để test failure scenarios?
- Có load testing cho Kafka không?

Câu hỏi 8.4: Best Practices
- Những best practices nào nên follow khi design Kafka topics?
- Làm thế nào để design message schemas?
- Có anti-patterns nào cần tránh không?
- Làm thế nào để optimize Kafka for production?
*/

// ============================================================================
// PHẦN 9: CODE REVIEW VÀ IMPROVEMENTS
// ============================================================================

/*
Câu hỏi 9.1: Producer Code Review
- Trong producer.go, có vấn đề gì về error handling không?
- IsConnected() check có race condition không?
- Async mode có thể gây message loss không?
- Có cần circuit breaker pattern không?

Câu hỏi 9.2: Consumer Code Review
- Consumer có handle context cancellation đúng cách không?
- Có vấn đề gì về error handling trong consumeMessages?
- Có cần backoff strategy khi Elasticsearch fails?
- Có cần batch processing thay vì individual processing?

Câu hỏi 9.3: Architecture Improvements
- Có cần separate topics cho different event types không?
- Có cần event versioning không?
- Làm thế nào để handle schema changes?
- Có cần event store pattern không?
*/

// ============================================================================
// PHẦN 10: SCENARIO-BASED QUESTIONS
// ============================================================================

/*
Câu hỏi 10.1: High Load Scenario
- Nếu có 1 triệu events/giờ, làm thế nào để handle?
- Làm thế nào để scale consumers để process faster?
- Có cần partition increase không?
- Làm thế nào để prevent consumer lag?

Câu hỏi 10.2: Failure Scenario
- Nếu Kafka broker crashes, hệ thống sẽ như thế nào?
- Làm thế nào để recover từ broker failure?
- Có cần message buffering khi Kafka down không?
- Làm thế nào để prevent data loss?

Câu hỏi 10.3: Migration Scenario
- Làm thế nào để migrate từ single broker sang cluster?
- Có downtime không khi migrate?
- Làm thế nào để migrate consumers?
- Có cần blue-green deployment không?

Câu hỏi 10.4: Multi-Tenant Scenario
- Làm thế nào để isolate data giữa các tenants?
- Có cần separate topics per tenant không?
- Làm thế nào để implement tenant-based access control?
- Có vấn đề gì về resource sharing?
*/

// ============================================================================
// KẾT THÚC
// ============================================================================








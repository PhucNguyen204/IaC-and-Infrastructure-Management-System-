package dto

type MetricsResponse struct {
	InstanceID    string                 `json:"instance_id"`
	Timestamp     string                 `json:"timestamp"`
	CPUPercent    float64                `json:"cpu_percent"`
	MemoryUsed    int64                  `json:"memory_used"`
	MemoryLimit   int64                  `json:"memory_limit"`
	MemoryPercent float64                `json:"memory_percent"`
	NetworkRx     int64                  `json:"network_rx"`
	NetworkTx     int64                  `json:"network_tx"`
	DiskRead      int64                  `json:"disk_read"`
	DiskWrite     int64                  `json:"disk_write"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

type PostgreSQLMetricsResponse struct {
	InstanceID        string  `json:"instance_id"`
	ActiveConnections int64   `json:"active_connections"`
	TotalConnections  int64   `json:"total_connections"`
	Transactions      int64   `json:"transactions"`
	Commits           int64   `json:"commits"`
	Rollbacks         int64   `json:"rollbacks"`
	BlocksRead        int64   `json:"blocks_read"`
	BlocksHit         int64   `json:"blocks_hit"`
	CacheHitRatio     float64 `json:"cache_hit_ratio"`
	TuplesReturned    int64   `json:"tuples_returned"`
	TuplesFetched     int64   `json:"tuples_fetched"`
	TuplesInserted    int64   `json:"tuples_inserted"`
	TuplesUpdated     int64   `json:"tuples_updated"`
	TuplesDeleted     int64   `json:"tuples_deleted"`
	ReplicationLag    int64   `json:"replication_lag,omitempty"`
}

type NginxMetricsResponse struct {
	InstanceID        string  `json:"instance_id"`
	ActiveConnections int64   `json:"active_connections"`
	Accepts           int64   `json:"accepts"`
	Handled           int64   `json:"handled"`
	Requests          int64   `json:"requests"`
	Reading           int64   `json:"reading"`
	Writing           int64   `json:"writing"`
	Waiting           int64   `json:"waiting"`
	RequestsPerSecond float64 `json:"requests_per_second,omitempty"`
}

type AggregatedMetricsResponse struct {
	InstanceID    string                   `json:"instance_id"`
	TimeRange     string                   `json:"time_range"`
	CPUPercent    AggregatedValue          `json:"cpu_percent"`
	MemoryPercent AggregatedValue          `json:"memory_percent"`
	NetworkRx     AggregatedValue          `json:"network_rx"`
	NetworkTx     AggregatedValue          `json:"network_tx"`
	DiskRead      AggregatedValue          `json:"disk_read"`
	DiskWrite     AggregatedValue          `json:"disk_write"`
	DataPoints    int                      `json:"data_points"`
}

type AggregatedValue struct {
	Avg float64 `json:"avg"`
	Max float64 `json:"max"`
	Min float64 `json:"min"`
}

type LogsResponse struct {
	InstanceID string                 `json:"instance_id"`
	Timestamp  string                 `json:"timestamp"`
	Message    string                 `json:"message"`
	Level      string                 `json:"level"`
	Action     string                 `json:"action"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

type HealthStatusResponse struct {
	InstanceID string `json:"instance_id"`
	Status     string `json:"status"`
	Timestamp  string `json:"timestamp"`
}

type InfrastructureListResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

type APIResponse struct {
	Success bool        `json:"success"`
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}


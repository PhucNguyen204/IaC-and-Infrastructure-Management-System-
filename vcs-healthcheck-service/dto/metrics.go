package dto

import "time"

// ContainerMetrics represents collected container metrics
type ContainerMetrics struct {
	MetricID         string    `json:"metric_id"`
	InfrastructureID string    `json:"infrastructure_id"`
	ClusterID        string    `json:"cluster_id,omitempty"`
	ContainerID      string    `json:"container_id"`
	ContainerName    string    `json:"container_name"`
	Type             string    `json:"type"` // postgres_cluster, nginx_cluster, etc.
	Timestamp        time.Time `json:"timestamp"`

	// CPU metrics
	CPU CPUMetrics `json:"cpu"`

	// Memory metrics
	Memory MemoryMetrics `json:"memory"`

	// Network metrics
	Network NetworkMetrics `json:"network"`

	// Disk I/O metrics
	Disk DiskMetrics `json:"disk"`

	// Health status
	Health HealthStatus `json:"health"`
}

type CPUMetrics struct {
	UsagePercent float64 `json:"usage_percent"`
	Cores        int     `json:"cores"`
	ThrottledNs  int64   `json:"throttled_ns,omitempty"`
}

type MemoryMetrics struct {
	UsedBytes    int64   `json:"used_bytes"`
	LimitBytes   int64   `json:"limit_bytes"`
	UsagePercent float64 `json:"usage_percent"`
	CacheBytes   int64   `json:"cache_bytes,omitempty"`
}

type NetworkMetrics struct {
	RxBytes   int64 `json:"rx_bytes"`
	TxBytes   int64 `json:"tx_bytes"`
	RxPackets int64 `json:"rx_packets"`
	TxPackets int64 `json:"tx_packets"`
	RxErrors  int64 `json:"rx_errors,omitempty"`
	TxErrors  int64 `json:"tx_errors,omitempty"`
}

type DiskMetrics struct {
	ReadBytes  int64 `json:"read_bytes"`
	WriteBytes int64 `json:"write_bytes"`
	ReadOps    int64 `json:"read_ops,omitempty"`
	WriteOps   int64 `json:"write_ops,omitempty"`
}

type HealthStatus struct {
	Status     string    `json:"status"` // healthy, unhealthy, unknown
	LastCheck  time.Time `json:"last_check"`
	Message    string    `json:"message,omitempty"`
	CheckType  string    `json:"check_type,omitempty"` // tcp, http, exec
	ResponseMs int64     `json:"response_ms,omitempty"`
}

// HealthCheckResult represents a single health check result
type HealthCheckResult struct {
	InfrastructureID string                 `json:"infrastructure_id"`
	ContainerID      string                 `json:"container_id"`
	ContainerName    string                 `json:"container_name"`
	Type             string                 `json:"type"`
	Status           string                 `json:"status"`
	Timestamp        time.Time              `json:"timestamp"`
	CheckDuration    int64                  `json:"check_duration_ms"`
	Message          string                 `json:"message,omitempty"`
	Details          map[string]interface{} `json:"details,omitempty"`
}

// AggregatedMetrics for dashboard display
type AggregatedMetrics struct {
	InfrastructureID string    `json:"infrastructure_id"`
	Type             string    `json:"type"`
	Period           string    `json:"period"` // 1h, 24h, 7d
	Timestamp        time.Time `json:"timestamp"`

	AvgCPUPercent    float64 `json:"avg_cpu_percent"`
	MaxCPUPercent    float64 `json:"max_cpu_percent"`
	AvgMemoryPercent float64 `json:"avg_memory_percent"`
	MaxMemoryPercent float64 `json:"max_memory_percent"`
	TotalNetworkRx   int64   `json:"total_network_rx"`
	TotalNetworkTx   int64   `json:"total_network_tx"`
	TotalDiskRead    int64   `json:"total_disk_read"`
	TotalDiskWrite   int64   `json:"total_disk_write"`

	UptimePercent float64 `json:"uptime_percent"`
	HealthChecks  int     `json:"health_checks"`
	FailedChecks  int     `json:"failed_checks"`
}

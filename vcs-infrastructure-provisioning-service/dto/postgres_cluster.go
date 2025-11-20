package dto

// CreateClusterRequest for creating PostgreSQL cluster
type CreateClusterRequest struct {
	ClusterName        string `json:"cluster_name" binding:"required"`
	PostgreSQLVersion  string `json:"postgres_version" binding:"required"`
	NodeCount          int    `json:"node_count" binding:"required,min=1,max=10"`
	CPUPerNode         int64  `json:"cpu_per_node" binding:"required"`
	MemoryPerNode      int64  `json:"memory_per_node" binding:"required"`
	StoragePerNode     int    `json:"storage_per_node" binding:"required"`
	PostgreSQLPassword string `json:"postgres_password" binding:"required,min=8"`
	ReplicationMode    string `json:"replication_mode" binding:"required,oneof=async sync"`
}

// ClusterInfoResponse returns cluster details
type ClusterInfoResponse struct {
	ClusterID         string            `json:"cluster_id"`
	InfrastructureID  string            `json:"infrastructure_id"`
	ClusterName       string            `json:"cluster_name"`
	PostgreSQLVersion string            `json:"postgres_version"`
	Status            string            `json:"status"`
	HAProxyPort       int               `json:"haproxy_port"` // Primary node port
	Nodes             []ClusterNodeInfo `json:"nodes"`
	CreatedAt         string            `json:"created_at"`
	UpdatedAt         string            `json:"updated_at"`
}

// ClusterNodeInfo details for each node
type ClusterNodeInfo struct {
	NodeID           string `json:"node_id"`
	NodeName         string `json:"node_name"`
	ContainerID      string `json:"container_id"`
	Role             string `json:"role"` // primary or replica
	Status           string `json:"status"`
	ReplicationDelay int    `json:"replication_delay"` // bytes
	IsHealthy        bool   `json:"is_healthy"`
}

// ScaleClusterRequest for scaling up/down
type ScaleClusterRequest struct {
	NodeCount int `json:"node_count" binding:"required,min=1,max=10"`
}

// TriggerFailoverRequest for manual failover
type TriggerFailoverRequest struct {
	NewPrimaryNodeID string `json:"new_primary_node_id" binding:"required"`
}

// ReplicationStatusResponse shows replication health
type ReplicationStatusResponse struct {
	Primary  string          `json:"primary"`
	Replicas []ReplicaStatus `json:"replicas"`
}

// ReplicaStatus for each replica node
type ReplicaStatus struct {
	NodeName   string  `json:"node_name"`
	State      string  `json:"state"`      // streaming, catchup, etc
	SyncState  string  `json:"sync_state"` // async or sync
	LagBytes   int     `json:"lag_bytes"`
	LagSeconds float64 `json:"lag_seconds"`
	IsHealthy  bool    `json:"is_healthy"`
}

// ClusterStatsResponse aggregated stats
type ClusterStatsResponse struct {
	ClusterID        string      `json:"cluster_id"`
	TotalConnections int         `json:"total_connections"`
	TotalDatabases   int         `json:"total_databases"`
	TotalSizeMB      int         `json:"total_size_mb"`
	Nodes            []NodeStats `json:"nodes"`
}

// NodeStats per node statistics
type NodeStats struct {
	NodeName          string `json:"node_name"`
	Role              string `json:"role"`
	CPUPercent        int    `json:"cpu_percent"`
	MemoryPercent     int    `json:"memory_percent"`
	ActiveConnections int    `json:"active_connections"`
}

// ClusterLogsResponse logs from all nodes
type ClusterLogsResponse struct {
	ClusterID string    `json:"cluster_id"`
	Logs      []NodeLog `json:"logs"`
}

// NodeLog individual node logs
type NodeLog struct {
	NodeName  string `json:"node_name"`
	Timestamp string `json:"timestamp"`
	Logs      string `json:"logs"`
}

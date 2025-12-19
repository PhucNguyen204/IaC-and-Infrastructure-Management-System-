package dto

import "time"

// LifecycleEvent represents infrastructure lifecycle events from Kafka
// Note: Uses instance_id to match the provisioning service's event format
type LifecycleEvent struct {
	EventID          string                 `json:"event_id"`
	InfrastructureID string                 `json:"instance_id"` // Maps from provisioning's instance_id
	ClusterID        string                 `json:"cluster_id,omitempty"`
	UserID           string                 `json:"user_id"`
	Type             string                 `json:"type"`   // postgres_cluster, nginx_cluster, dind, clickhouse
	Action           string                 `json:"action"` // created, started, stopped, deleted, node_added, node_removed, failover
	Status           string                 `json:"status"` // running, stopped, failed, deleted
	PreviousStatus   string                 `json:"previous_status,omitempty"`
	Timestamp        time.Time              `json:"timestamp"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// Actions constants
const (
	ActionCreated     = "created"
	ActionStarted     = "started"
	ActionStopped     = "stopped"
	ActionDeleted     = "deleted"
	ActionNodeAdded   = "node_added"
	ActionNodeRemoved = "node_removed"
	ActionFailover    = "failover"
	ActionBackup      = "backup"
	ActionRestore     = "restore"
	ActionHealthCheck = "health_check"

	// Cluster-specific actions from provisioning service (postgres)
	ActionClusterCreated = "cluster.created"
	ActionClusterStarted = "cluster.started"
	ActionClusterStopped = "cluster.stopped"
	ActionClusterDeleted = "cluster.deleted"

	// Nginx cluster actions
	ActionNginxCreated = "nginx_cluster.created"
	ActionNginxStarted = "nginx_cluster.started"
	ActionNginxStopped = "nginx_cluster.stopped"
	ActionNginxDeleted = "nginx_cluster.deleted"

	// DinD environment actions
	ActionDinDCreated = "dind.created"
	ActionDinDStarted = "dind.started"
	ActionDinDStopped = "dind.stopped"
	ActionDinDDeleted = "dind.deleted"

	// ClickHouse actions
	ActionClickHouseCreated = "clickhouse.created"
	ActionClickHouseStarted = "clickhouse.started"
	ActionClickHouseStopped = "clickhouse.stopped"
	ActionClickHouseDeleted = "clickhouse.deleted"
)

// Status constants
const (
	StatusRunning   = "running"
	StatusStopped   = "stopped"
	StatusFailed    = "failed"
	StatusDeleted   = "deleted"
	StatusHealthy   = "healthy"
	StatusUnhealthy = "unhealthy"
	StatusUnknown   = "unknown"
)

// InfraType constants
const (
	TypePostgresCluster = "postgres_cluster"
	TypeNginxCluster    = "nginx_cluster"
	TypeDinD            = "dind"
	TypeClickHouse      = "clickhouse"
)

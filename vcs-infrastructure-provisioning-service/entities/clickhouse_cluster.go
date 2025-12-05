package entities

import (
	"time"
)

// ClickHouseCluster represents a ClickHouse database cluster
type ClickHouseCluster struct {
	ID               string         `gorm:"primaryKey;type:varchar(36)"`
	InfrastructureID string         `gorm:"type:varchar(36);not null;index"`
	Infrastructure   Infrastructure `gorm:"foreignKey:InfrastructureID"`
	ClusterName      string         `gorm:"type:varchar(255)"`
	Version          string         `gorm:"type:varchar(20);not null"` // 23.8, 24.1, etc
	NodeCount        int            `gorm:"not null"`
	Username         string         `gorm:"type:varchar(100);not null"`
	Password         string         `gorm:"type:varchar(255);not null"`
	DatabaseName     string         `gorm:"type:varchar(100);not null"`
	HTTPPort         int            `gorm:"default:8123"` // HTTP interface
	NativePort       int            `gorm:"default:9000"` // Native TCP protocol
	NetworkID        string         `gorm:"type:varchar(255)"`
	DataDirectory    string         `gorm:"type:varchar(255)"`
	StorageSize      int            `gorm:"default:0"` // GB
	CPULimit         int64          `gorm:"default:0"`
	MemoryLimit      int64          `gorm:"default:0"`

	// Cluster configuration
	ReplicationEnabled bool   `gorm:"default:false"`
	ShardCount         int    `gorm:"default:1"`
	ReplicaCount       int    `gorm:"default:1"`
	ZooKeeperEndpoints string `gorm:"type:varchar(500)"` // For replication

	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

// ClickHouseNode represents a single node in ClickHouse cluster
type ClickHouseNode struct {
	ID          string            `gorm:"primaryKey;type:varchar(36)"`
	ClusterID   string            `gorm:"type:varchar(36);not null;index"`
	Cluster     ClickHouseCluster `gorm:"foreignKey:ClusterID"`
	NodeName    string            `gorm:"type:varchar(100)"`
	ContainerID string            `gorm:"type:varchar(100)"`
	HTTPPort    int               `gorm:"not null"` // Exposed HTTP port
	NativePort  int               `gorm:"not null"` // Exposed native port
	VolumeID    string            `gorm:"type:varchar(255)"`
	ShardNum    int               `gorm:"default:1"`
	ReplicaNum  int               `gorm:"default:1"`
	IsHealthy   bool              `gorm:"default:true"`
	CreatedAt   time.Time         `gorm:"autoCreateTime"`
	UpdatedAt   time.Time         `gorm:"autoUpdateTime"`
}

func (ClickHouseCluster) TableName() string {
	return "clickhouse_clusters"
}

func (ClickHouseNode) TableName() string {
	return "clickhouse_nodes"
}

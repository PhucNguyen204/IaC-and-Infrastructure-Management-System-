package entities

import (
	"time"
)

type PostgreSQLCluster struct {
	ID               string    `gorm:"primaryKey;type:varchar(36)"`
	InfrastructureID string    `gorm:"type:varchar(36);not null;index"`
	Infrastructure   Infrastructure `gorm:"foreignKey:InfrastructureID"`
	NodeCount        int       `gorm:"not null"`
	PrimaryNodeID    string    `gorm:"type:varchar(100)"`
	Version          string    `gorm:"type:varchar(20);not null"`
	DatabaseName     string    `gorm:"type:varchar(100);not null"`
	Username         string    `gorm:"type:varchar(100);not null"`
	Password         string    `gorm:"type:varchar(255);not null"`
	HAProxyPort      int       `gorm:"not null"`
	NetworkID        string    `gorm:"type:varchar(255)"`
	CPULimit         int64     `gorm:"default:0"`
	MemoryLimit      int64     `gorm:"default:0"`
	CreatedAt        time.Time `gorm:"autoCreateTime"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime"`
}

type ClusterNode struct {
	ID                string    `gorm:"primaryKey;type:varchar(36)"`
	ClusterID         string    `gorm:"type:varchar(36);not null;index"`
	Cluster           PostgreSQLCluster `gorm:"foreignKey:ClusterID"`
	ContainerID       string    `gorm:"type:varchar(100)"`
	Role              string    `gorm:"type:varchar(20);not null"`
	Port              int       `gorm:"not null"`
	VolumeID          string    `gorm:"type:varchar(255)"`
	ReplicationDelay  int64     `gorm:"default:0"`
	IsHealthy         bool      `gorm:"default:true"`
	CreatedAt         time.Time `gorm:"autoCreateTime"`
	UpdatedAt         time.Time `gorm:"autoUpdateTime"`
}

type EtcdNode struct {
	ID          string    `gorm:"primaryKey;type:varchar(36)"`
	ClusterID   string    `gorm:"type:varchar(36);not null;index"`
	ContainerID string    `gorm:"type:varchar(100)"`
	Port        int       `gorm:"not null"`
	VolumeID    string    `gorm:"type:varchar(255)"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}


package entities

import (
	"time"
)

type InfrastructureType string
type InfrastructureStatus string

const (
	TypePostgreSQLSingle  InfrastructureType = "postgres_single"
	TypePostgreSQLCluster InfrastructureType = "postgres_cluster"
	TypeNginx             InfrastructureType = "nginx"
	TypeNginxCluster      InfrastructureType = "nginx_cluster"
	TypeDockerService     InfrastructureType = "docker_service"
	TypeDinD              InfrastructureType = "dind" // Docker-in-Docker
	InfraTypeK8sCluster   InfrastructureType = "k8s_cluster"
	TypeClickHouse        InfrastructureType = "clickhouse"
)

const (
	StatusCreating InfrastructureStatus = "creating"
	StatusRunning  InfrastructureStatus = "running"
	StatusStopped  InfrastructureStatus = "stopped"
	StatusFailed   InfrastructureStatus = "failed"
	StatusDeleting InfrastructureStatus = "deleting"
	StatusDeleted  InfrastructureStatus = "deleted"
)

type Infrastructure struct {
	ID        string               `gorm:"primaryKey;type:varchar(36)"`
	Name      string               `gorm:"type:varchar(255);not null"`
	Type      InfrastructureType   `gorm:"type:varchar(50);not null"`
	Status    InfrastructureStatus `gorm:"type:varchar(50);not null"`
	UserID    string               `gorm:"type:varchar(36);not null;index"`
	CreatedAt time.Time            `gorm:"autoCreateTime"`
	UpdatedAt time.Time            `gorm:"autoUpdateTime"`
}

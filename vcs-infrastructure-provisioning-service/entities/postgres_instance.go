package entities

import (
	"time"
)

type PostgreSQLInstance struct {
	ID               string         `gorm:"primaryKey;type:varchar(36)"`
	InfrastructureID string         `gorm:"type:varchar(36);not null;index"`
	Infrastructure   Infrastructure `gorm:"foreignKey:InfrastructureID"`
	ContainerID      string         `gorm:"type:varchar(100)"`
	Version          string         `gorm:"type:varchar(20);not null"`
	Port             int            `gorm:"not null"`
	DatabaseName     string         `gorm:"type:varchar(100);not null"`
	Username         string         `gorm:"type:varchar(100);not null"`
	Password         string         `gorm:"type:varchar(255);not null"`
	CPULimit         int64          `gorm:"default:0"`
	MemoryLimit      int64          `gorm:"default:0"`
	StorageSize      int64          `gorm:"default:10737418240"`
	VolumeID         string         `gorm:"type:varchar(255)"`
	NetworkID        string         `gorm:"type:varchar(255)"`
	CreatedAt        time.Time      `gorm:"autoCreateTime"`
	UpdatedAt        time.Time      `gorm:"autoUpdateTime"`
}

package entities

import (
	"time"
)

// DinDEnvironment - Docker-in-Docker environment entity
type DinDEnvironment struct {
	ID               string    `gorm:"primaryKey;type:varchar(36)"`
	InfrastructureID string    `gorm:"type:varchar(36);not null;index"`
	Name             string    `gorm:"type:varchar(255);not null"`
	ContainerID      string    `gorm:"type:varchar(100)"`
	ContainerName    string    `gorm:"type:varchar(255)"`
	Status           string    `gorm:"type:varchar(50);default:'creating'"` // creating, running, stopped, failed
	DockerHost       string    `gorm:"type:varchar(255)"`                   // Docker host endpoint inside DinD
	IPAddress        string    `gorm:"type:varchar(50)"`
	ResourcePlan     string    `gorm:"type:varchar(20);default:'medium'"` // small, medium, large
	CPULimit         string    `gorm:"type:varchar(20)"`
	MemoryLimit      string    `gorm:"type:varchar(20)"`
	StorageDriver    string    `gorm:"type:varchar(50);default:'overlay2'"`
	NetworkID        string    `gorm:"type:varchar(100)"`
	Description      string    `gorm:"type:text"`
	AutoCleanup      bool      `gorm:"default:false"`
	TTLHours         int       `gorm:"default:0"` // 0 = no expiration
	ExpiresAt        time.Time `gorm:"index"`
	UserID           string    `gorm:"type:varchar(36);index"`
	CreatedAt        time.Time `gorm:"autoCreateTime"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime"`
}

// TableName - Tên bảng trong database
func (DinDEnvironment) TableName() string {
	return "dind_environments"
}

// DinDCommandHistory - Lịch sử command đã chạy
type DinDCommandHistory struct {
	ID            string    `gorm:"primaryKey;type:varchar(36)"`
	EnvironmentID string    `gorm:"type:varchar(36);not null;index"`
	Command       string    `gorm:"type:text;not null"`
	Output        string    `gorm:"type:text"`
	ExitCode      int       `gorm:"default:0"`
	Duration      int       `gorm:"default:0"` // milliseconds
	ExecutedAt    time.Time `gorm:"autoCreateTime"`
}

// TableName - Tên bảng trong database
func (DinDCommandHistory) TableName() string {
	return "dind_command_history"
}


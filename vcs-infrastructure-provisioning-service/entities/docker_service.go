package entities

import (
	"time"
)

type DockerService struct {
	ID               string             `gorm:"primaryKey;type:varchar(36)"`
	InfrastructureID string             `gorm:"type:varchar(36);not null;index"`
	Infrastructure   Infrastructure     `gorm:"foreignKey:InfrastructureID"`
	Name             string             `gorm:"type:varchar(255);not null"`
	Image            string             `gorm:"type:varchar(500);not null"`
	ImageTag         string             `gorm:"type:varchar(100);not null"`
	ServiceType      string             `gorm:"type:varchar(50);default:'web'"`
	ContainerID      string             `gorm:"type:varchar(100)"`
	ContainerName    string             `gorm:"type:varchar(255)"`
	Command          string             `gorm:"type:text"`
	Args             string             `gorm:"type:text"`
	EnvVars          []DockerEnvVar     `gorm:"foreignKey:ServiceID"`
	Ports            []DockerPort       `gorm:"foreignKey:ServiceID"`
	Networks         []DockerNetwork    `gorm:"foreignKey:ServiceID"`
	HealthCheck      *DockerHealthCheck `gorm:"foreignKey:ServiceID"`
	RestartPolicy    string             `gorm:"type:varchar(50);default:'unless-stopped'"`
	MaxRetries       int                `gorm:"default:3"`
	CPULimit         int64              `gorm:"default:0"`
	MemoryLimit      int64              `gorm:"default:0"`
	Status           string             `gorm:"type:varchar(50);default:'creating'"`
	IPAddress        string             `gorm:"type:varchar(50)"`
	InternalEndpoint string             `gorm:"type:varchar(255)"`
	CreatedAt        time.Time          `gorm:"autoCreateTime"`
	UpdatedAt        time.Time          `gorm:"autoUpdateTime"`
}

type DockerEnvVar struct {
	ID        string    `gorm:"primaryKey;type:varchar(36)"`
	ServiceID string    `gorm:"type:varchar(36);not null;index"`
	Key       string    `gorm:"type:varchar(255);not null"`
	Value     string    `gorm:"type:text"`
	IsSecret  bool      `gorm:"default:false"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

type DockerPort struct {
	ID            string    `gorm:"primaryKey;type:varchar(36)"`
	ServiceID     string    `gorm:"type:varchar(36);not null;index"`
	ContainerPort int       `gorm:"not null"`
	HostPort      int       `gorm:"default:0"`
	Protocol      string    `gorm:"type:varchar(10);default:'tcp'"`
	CreatedAt     time.Time `gorm:"autoCreateTime"`
}

type DockerNetwork struct {
	ID        string    `gorm:"primaryKey;type:varchar(36)"`
	ServiceID string    `gorm:"type:varchar(36);not null;index"`
	NetworkID string    `gorm:"type:varchar(255);not null"`
	Alias     string    `gorm:"type:varchar(255)"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

type DockerHealthCheck struct {
	ID                 string `gorm:"primaryKey;type:varchar(36)"`
	ServiceID          string `gorm:"type:varchar(36);not null;index;unique"`
	Type               string `gorm:"type:varchar(50);not null"`
	HTTPPath           string `gorm:"type:varchar(255)"`
	Port               int    `gorm:"default:0"`
	Command            string `gorm:"type:text"`
	Interval           int    `gorm:"default:30"`
	Timeout            int    `gorm:"default:10"`
	HealthyThreshold   int    `gorm:"default:3"`
	UnhealthyThreshold int    `gorm:"default:3"`
	Status             string `gorm:"type:varchar(50);default:'unknown'"`
	LastCheck          time.Time
	CreatedAt          time.Time `gorm:"autoCreateTime"`
	UpdatedAt          time.Time `gorm:"autoUpdateTime"`
}

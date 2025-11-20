package entities

import (
	"time"
)

type NginxInstance struct {
	ID               string            `gorm:"primaryKey;type:varchar(36)"`
	InfrastructureID string            `gorm:"type:varchar(36);not null;index"`
	Infrastructure   Infrastructure    `gorm:"foreignKey:InfrastructureID"`
	ContainerID      string            `gorm:"type:varchar(100)"`
	Port             int               `gorm:"not null"`
	SSLPort          int               `gorm:"default:0"`
	Config           string            `gorm:"type:text"`
	VolumeID         string            `gorm:"type:varchar(255)"`
	NetworkID        string            `gorm:"type:varchar(255)"`
	CPULimit         int64             `gorm:"default:0"`
	MemoryLimit      int64             `gorm:"default:0"`
	Domains          []NginxDomain     `gorm:"foreignKey:NginxID"`
	Routes           []NginxRoute      `gorm:"foreignKey:NginxID"`
	Upstreams        []NginxUpstream   `gorm:"foreignKey:NginxID"`
	Certificate      *NginxCertificate `gorm:"foreignKey:NginxID"`
	SecurityPolicy   *NginxSecurity    `gorm:"foreignKey:NginxID"`
	CreatedAt        time.Time         `gorm:"autoCreateTime"`
	UpdatedAt        time.Time         `gorm:"autoUpdateTime"`
}

type NginxDomain struct {
	ID        string    `gorm:"primaryKey;type:varchar(36)"`
	NginxID   string    `gorm:"type:varchar(36);not null;index"`
	Domain    string    `gorm:"type:varchar(255);not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

type NginxRoute struct {
	ID        string    `gorm:"primaryKey;type:varchar(36)"`
	NginxID   string    `gorm:"type:varchar(36);not null;index"`
	Path      string    `gorm:"type:varchar(255);not null"`
	Backend   string    `gorm:"type:varchar(255);not null"`
	Priority  int       `gorm:"default:0"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

type NginxUpstream struct {
	ID        string                 `gorm:"primaryKey;type:varchar(36)"`
	NginxID   string                 `gorm:"type:varchar(36);not null;index"`
	Name      string                 `gorm:"type:varchar(100);not null"`
	Policy    string                 `gorm:"type:varchar(50);default:'round_robin'"`
	Backends  []NginxUpstreamBackend `gorm:"foreignKey:UpstreamID"`
	CreatedAt time.Time              `gorm:"autoCreateTime"`
	UpdatedAt time.Time              `gorm:"autoUpdateTime"`
}

type NginxUpstreamBackend struct {
	ID         string    `gorm:"primaryKey;type:varchar(36)"`
	UpstreamID string    `gorm:"type:varchar(36);not null;index"`
	Address    string    `gorm:"type:varchar(255);not null"`
	Weight     int       `gorm:"default:1"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`
}

type NginxCertificate struct {
	ID          string `gorm:"primaryKey;type:varchar(36)"`
	NginxID     string `gorm:"type:varchar(36);not null;index;unique"`
	Domain      string `gorm:"type:varchar(255);not null"`
	Certificate string `gorm:"type:text;not null"`
	PrivateKey  string `gorm:"type:text;not null"`
	Status      string `gorm:"type:varchar(50);default:'valid'"`
	ExpiresAt   time.Time
	Issuer      string    `gorm:"type:varchar(255)"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}

type NginxSecurity struct {
	ID                string    `gorm:"primaryKey;type:varchar(36)"`
	NginxID           string    `gorm:"type:varchar(36);not null;index;unique"`
	RateLimitRPS      int       `gorm:"default:0"`
	RateLimitBurst    int       `gorm:"default:0"`
	RateLimitPath     string    `gorm:"type:varchar(255)"`
	AllowIPs          string    `gorm:"type:text"`
	DenyIPs           string    `gorm:"type:text"`
	BasicAuthUsername string    `gorm:"type:varchar(255)"`
	BasicAuthPassword string    `gorm:"type:varchar(255)"`
	BasicAuthRealm    string    `gorm:"type:varchar(255)"`
	CreatedAt         time.Time `gorm:"autoCreateTime"`
	UpdatedAt         time.Time `gorm:"autoUpdateTime"`
}

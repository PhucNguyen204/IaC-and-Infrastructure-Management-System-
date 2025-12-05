package entities

import (
	"time"
)

type StackStatus string

const (
	StackStatusCreating StackStatus = "creating"
	StackStatusRunning  StackStatus = "running"
	StackStatusUpdating StackStatus = "updating"
	StackStatusStopped  StackStatus = "stopped"
	StackStatusFailed   StackStatus = "failed"
	StackStatusDeleting StackStatus = "deleting"
	StackStatusDeleted  StackStatus = "deleted"
)

// Stack represents a logical grouping of infrastructure resources
// Examples: "tenant-123-prod", "project-abc-staging"
type Stack struct {
	ID          string      `gorm:"primaryKey;type:varchar(36)"`
	Name        string      `gorm:"type:varchar(255);not null;index"`
	Description string      `gorm:"type:text"`
	Environment string      `gorm:"type:varchar(50);index"` // dev, staging, prod
	ProjectID   string      `gorm:"type:varchar(36);index"`
	TenantID    string      `gorm:"type:varchar(36);index"`
	UserID      string      `gorm:"type:varchar(36);not null;index"`
	Status      StackStatus `gorm:"type:varchar(50);not null"`
	Tags        string      `gorm:"type:jsonb"` // JSON array of tags
	CreatedAt   time.Time   `gorm:"autoCreateTime"`
	UpdatedAt   time.Time   `gorm:"autoUpdateTime"`

	// Relations
	Resources []StackResource `gorm:"foreignKey:StackID"`
}

// StackResource links a Stack to its Infrastructure resources
type StackResource struct {
	ID               string    `gorm:"primaryKey;type:varchar(36)"`
	StackID          string    `gorm:"type:varchar(36);not null;index"`
	InfrastructureID string    `gorm:"type:varchar(36);not null;index"`
	ResourceType     string    `gorm:"type:varchar(50);not null"` // NGINX_GATEWAY, POSTGRES_INSTANCE, etc.
	Role             string    `gorm:"type:varchar(50)"`          // gateway, database, app, cache, queue
	DependsOn        string    `gorm:"type:jsonb"`                // JSON array of resource IDs this depends on
	Order            int       `gorm:"type:int;default:0"`        // Creation order (lower first)
	CreatedAt        time.Time `gorm:"autoCreateTime"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime"`

	// Relations
	Stack          Stack          `gorm:"foreignKey:StackID"`
	Infrastructure Infrastructure `gorm:"foreignKey:InfrastructureID"`
}

// StackTemplate represents a reusable stack configuration
type StackTemplate struct {
	ID          string    `gorm:"primaryKey;type:varchar(36)"`
	Name        string    `gorm:"type:varchar(255);not null;uniqueIndex"`
	Description string    `gorm:"type:text"`
	Category    string    `gorm:"type:varchar(50);index"` // web-app, microservice, data-pipeline
	IsPublic    bool      `gorm:"default:false"`
	UserID      string    `gorm:"type:varchar(36);index"`
	Spec        string    `gorm:"type:jsonb;not null"` // JSON template specification
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}

// StackOperation tracks operations performed on stacks
type StackOperation struct {
	ID            string    `gorm:"primaryKey;type:varchar(36)"`
	StackID       string    `gorm:"type:varchar(36);not null;index"`
	OperationType string    `gorm:"type:varchar(50);not null"` 
	Status        string    `gorm:"type:varchar(50);not null"` 
	UserID        string    `gorm:"type:varchar(36);not null"`
	StartedAt     time.Time `gorm:"autoCreateTime"`
	CompletedAt   *time.Time
	ErrorMessage  string `gorm:"type:text"`
	Details       string `gorm:"type:jsonb"` 

	// Relations
	Stack Stack `gorm:"foreignKey:StackID"`
}

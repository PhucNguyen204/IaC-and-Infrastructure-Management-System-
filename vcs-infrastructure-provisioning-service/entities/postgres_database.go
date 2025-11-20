package entities

import "time"

type PostgresDatabase struct {
	ID             string             `gorm:"primaryKey;type:varchar(36)"`
	InstanceID     string             `gorm:"type:varchar(36);not null;index"`
	Instance       PostgreSQLInstance `gorm:"foreignKey:InstanceID;references:ID"`
	DBName         string             `gorm:"type:varchar(100);not null;uniqueIndex:idx_instance_dbname"`
	OwnerUsername  string             `gorm:"type:varchar(100);not null"`
	OwnerPassword  string             `gorm:"type:varchar(255);not null"`
	ProjectID      string             `gorm:"type:varchar(100);index"`
	TenantID       string             `gorm:"type:varchar(100);index"`
	EnvironmentID  string             `gorm:"type:varchar(100)"`
	MaxSizeGB      int                `gorm:"default:10"`
	MaxConnections int                `gorm:"default:50"`
	CurrentSizeMB  int64              `gorm:"default:0"`
	ActiveConns    int                `gorm:"default:0"`
	Status         string             `gorm:"type:varchar(20);default:'CREATING'"`
	CreatedAt      time.Time          `gorm:"autoCreateTime"`
	UpdatedAt      time.Time          `gorm:"autoUpdateTime"`
}

type PostgresBackup struct {
	ID           string           `gorm:"primaryKey;type:varchar(36)"`
	DatabaseID   string           `gorm:"type:varchar(36);not null;index"`
	Database     PostgresDatabase `gorm:"foreignKey:DatabaseID"`
	BackupType   string           `gorm:"type:varchar(20);default:'LOGICAL'"`
	SizeMB       int64            `gorm:"default:0"`
	Location     string           `gorm:"type:text"`
	Status       string           `gorm:"type:varchar(20);default:'PENDING'"`
	StartedAt    time.Time
	CompletedAt  *time.Time
	ErrorMessage string    `gorm:"type:text"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
}

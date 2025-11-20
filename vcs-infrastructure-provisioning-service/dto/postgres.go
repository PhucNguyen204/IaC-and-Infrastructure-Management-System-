package dto

type CreatePostgreSQLRequest struct {
	Name         string `json:"name" binding:"required"`
	Version      string `json:"version" binding:"required"`
	Port         int    `json:"port" binding:"required,min=1024,max=65535"`
	DatabaseName string `json:"database_name" binding:"required"`
	Username     string `json:"username" binding:"required"`
	Password     string `json:"password" binding:"required,min=8"`
	CPULimit     int64  `json:"cpu_limit"`
	MemoryLimit  int64  `json:"memory_limit"`
	StorageSize  int64  `json:"storage_size"`
}

type PostgreSQLInfoResponse struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Status       string `json:"status"`
	ContainerID  string `json:"container_id"`
	Version      string `json:"version"`
	Port         int    `json:"port"`
	DatabaseName string `json:"database_name"`
	Username     string `json:"username"`
	CPULimit     int64  `json:"cpu_limit"`
	MemoryLimit  int64  `json:"memory_limit"`
	StorageSize  int64  `json:"storage_size"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

type BackupPostgreSQLRequest struct {
	BackupPath string `json:"backup_path" binding:"required"`
}

type BackupPostgreSQLResponse struct {
	BackupFile string `json:"backup_file"`
	Size       int64  `json:"size"`
}

type RestorePostgreSQLRequest struct {
	BackupFile string `json:"backup_file" binding:"required"`
}

type CreateDatabaseRequest struct {
	DBName         string `json:"db_name" binding:"required"`
	OwnerUsername  string `json:"owner_username"`
	OwnerPassword  string `json:"owner_password" binding:"min=8"`
	ProjectID      string `json:"project_id" binding:"required"`
	TenantID       string `json:"tenant_id"`
	EnvironmentID  string `json:"environment_id"`
	MaxSizeGB      int    `json:"max_size_gb"`
	MaxConnections int    `json:"max_connections"`
	InitSchema     string `json:"init_schema"`
}

type DatabaseInfo struct {
	ID             string         `json:"id"`
	InstanceID     string         `json:"instance_id"`
	DBName         string         `json:"db_name"`
	OwnerUsername  string         `json:"owner_username"`
	ProjectID      string         `json:"project_id"`
	TenantID       string         `json:"tenant_id"`
	EnvironmentID  string         `json:"environment_id"`
	MaxSizeGB      int            `json:"max_size_gb"`
	MaxConnections int            `json:"max_connections"`
	CurrentSizeMB  int64          `json:"current_size_mb"`
	ActiveConns    int            `json:"active_conns"`
	Status         string         `json:"status"`
	ConnectionInfo ConnectionInfo `json:"connection_info"`
	CreatedAt      string         `json:"created_at"`
	UpdatedAt      string         `json:"updated_at"`
}

type ConnectionInfo struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type UpdateQuotaRequest struct {
	MaxSizeGB      int `json:"max_size_gb"`
	MaxConnections int `json:"max_connections"`
}

type DatabaseMetrics struct {
	DatabaseID       string  `json:"database_id"`
	DBName           string  `json:"db_name"`
	CurrentSizeMB    int64   `json:"current_size_mb"`
	MaxSizeGB        int     `json:"max_size_gb"`
	SizeUsagePercent float64 `json:"size_usage_percent"`
	ActiveConns      int     `json:"active_conns"`
	MaxConnections   int     `json:"max_connections"`
	ConnUsagePercent float64 `json:"conn_usage_percent"`
	Status           string  `json:"status"`
	QueryPerSecond   float64 `json:"query_per_second"`
}

type BackupDatabaseRequest struct {
	BackupType string `json:"backup_type"`
	Mode       string `json:"mode"`
}

type BackupInfo struct {
	ID          string `json:"id"`
	DatabaseID  string `json:"database_id"`
	BackupType  string `json:"backup_type"`
	SizeMB      int64  `json:"size_mb"`
	Location    string `json:"location"`
	Status      string `json:"status"`
	StartedAt   string `json:"started_at"`
	CompletedAt string `json:"completed_at,omitempty"`
}

type RestoreDatabaseRequest struct {
	BackupID  string `json:"backup_id" binding:"required"`
	Mode      string `json:"mode"`
	NewDBName string `json:"new_db_name"`
}

type ManageLifecycleRequest struct {
	Action        string `json:"action" binding:"required"`
	RequireBackup bool   `json:"require_backup"`
}

type InstanceOverview struct {
	InstanceID       string            `json:"instance_id"`
	TotalDatabases   int               `json:"total_databases"`
	TotalSizeGB      float64           `json:"total_size_gb"`
	TotalConnections int               `json:"total_connections"`
	TopDatabases     []DatabaseMetrics `json:"top_databases"`
	CapacityStatus   string            `json:"capacity_status"`
}

type ProvisionSharedInstanceRequest struct {
	Plan             string `json:"plan" binding:"required"`
	Version          string `json:"version" binding:"required"`
	Region           string `json:"region"`
	StorageSizeGB    int    `json:"storage_size_gb"`
	IOPS             int    `json:"iops"`
	BackupPolicy     string `json:"backup_policy"`
	MonitoringPolicy string `json:"monitoring_policy"`
}

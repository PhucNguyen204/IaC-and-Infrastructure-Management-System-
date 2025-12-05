package dto

type CreateClickHouseRequest struct {
	ClusterName string `json:"cluster_name" binding:"required"`
	Version     string `json:"version"`  // latest, 24.1, 23.8 (default: latest)
	Username    string `json:"username"` // default: default
	Password    string `json:"password" binding:"required,min=6"`
	Database    string `json:"database" binding:"required"`

	CPULimit    int64 `json:"cpu_limit"`    // CPU cores
	MemoryLimit int64 `json:"memory_limit"` // MB
	StorageSize int   `json:"storage_size"` // GB

	Tables []ClickHouseTableDef `json:"tables,omitempty"`
}

type ClickHouseTableDef struct {
	Name        string                `json:"name" binding:"required"`
	Engine      string                `json:"engine,omitempty"` // MergeTree (default), ReplacingMergeTree, etc
	Columns     []ClickHouseColumnDef `json:"columns" binding:"required"`
	OrderBy     []string              `json:"order_by,omitempty"`
	PartitionBy string                `json:"partition_by,omitempty"`
	TTL         string                `json:"ttl,omitempty"` // e.g., "timestamp + INTERVAL 30 DAY"
}

// ClickHouseColumnDef defines a column
type ClickHouseColumnDef struct {
	Name         string `json:"name" binding:"required"`
	Type         string `json:"type" binding:"required"` // String, UInt64, DateTime64, etc
	DefaultValue string `json:"default,omitempty"`
	Nullable     bool   `json:"nullable,omitempty"`
}

// ClickHouseResponse returns cluster info
type ClickHouseResponse struct {
	ClusterID        string             `json:"cluster_id"`
	InfrastructureID string             `json:"infrastructure_id"`
	ClusterName      string             `json:"cluster_name"`
	Version          string             `json:"version"`
	Status           string             `json:"status"`
	Database         string             `json:"database"`
	Username         string             `json:"username"`
	HTTPEndpoint     ClickHouseEndpoint `json:"http_endpoint"`   // Port 8123 - for HTTP queries
	NativeEndpoint   ClickHouseEndpoint `json:"native_endpoint"` // Port 9000 - for native protocol
	ContainerID      string             `json:"container_id"`
	ContainerName    string             `json:"container_name"`
	NetworkName      string             `json:"network_name"`
	CreatedAt        string             `json:"created_at"`
	UpdatedAt        string             `json:"updated_at"`
}

// ClickHouseEndpoint connection info
type ClickHouseEndpoint struct {
	Host         string `json:"host"`          // localhost for external access
	Port         int    `json:"port"`          // Exposed port
	InternalHost string `json:"internal_host"` // Container name for internal network access
	InternalPort int    `json:"internal_port"` // Internal port (8123 or 9000)
}

// ClickHouseQueryRequest for running SQL on ClickHouse
type ClickHouseQueryRequest struct {
	Query    string `json:"query" binding:"required"`
	Database string `json:"database,omitempty"`
}

// ClickHouseQueryResponse query result
type ClickHouseQueryResponse struct {
	Success  bool        `json:"success"`
	Data     interface{} `json:"data,omitempty"`
	Columns  []string    `json:"columns,omitempty"`
	RowCount int         `json:"row_count"`
	Elapsed  float64     `json:"elapsed_ms"`
	Error    string      `json:"error,omitempty"`
}

// ClickHouseInsertRequest for inserting data
type ClickHouseInsertRequest struct {
	Table string                   `json:"table" binding:"required"`
	Data  []map[string]interface{} `json:"data" binding:"required"`
}

// ================== Deploy Container with Auto-Provisioned Dependencies ==================

// DeployContainerRequest for deploying any container image with dependencies
type DeployContainerRequest struct {
	Name      string `json:"name" binding:"required"`
	ImageName string `json:"image_name" binding:"required"`

	// Container configuration
	Ports       map[int]int       `json:"ports,omitempty"`       // container_port -> host_port (0 = auto assign)
	Environment map[string]string `json:"environment,omitempty"` // Static ENV vars
	Volumes     []VolumeMount     `json:"volumes,omitempty"`
	Command     []string          `json:"command,omitempty"`
	Entrypoint  []string          `json:"entrypoint,omitempty"`

	// Dependencies - will be auto-created and connected
	Dependencies []DependencySpec `json:"dependencies,omitempty"`

	// Network - will be auto-created if not exists
	NetworkName string `json:"network_name,omitempty"`
}

// VolumeMount for container volumes
type VolumeMount struct {
	HostPath      string `json:"host_path,omitempty"` // If empty, creates named volume
	ContainerPath string `json:"container_path" binding:"required"`
	VolumeName    string `json:"volume_name,omitempty"` // Named volume
	Content       string `json:"content,omitempty"`     // File content to write (for config files)
	ReadOnly      bool   `json:"read_only,omitempty"`
}

// DependencySpec for auto-provisioning dependencies
type DependencySpec struct {
	Type   string                 `json:"type" binding:"required"` // clickhouse, postgres, nginx, redis
	Name   string                 `json:"name" binding:"required"`
	Config map[string]interface{} `json:"config,omitempty"` // Type-specific config

	// ENV mapping - automatically inject dependency endpoints into container ENV
	// Keys are ENV var names, values are what to inject:
	// - "host": internal hostname
	// - "port": native port (9000 for clickhouse, 5432 for postgres)
	// - "http_port": HTTP port if available
	// - "username": username
	// - "password": password
	// - "database": database name
	EnvMapping map[string]string `json:"env_mapping,omitempty"`
}

// DeployContainerResponse result
type DeployContainerResponse struct {
	ContainerID   string               `json:"container_id"`
	ContainerName string               `json:"container_name"`
	Status        string               `json:"status"`
	Ports         map[int]int          `json:"ports"` // container_port -> host_port
	NetworkName   string               `json:"network_name"`
	Dependencies  []DependencyResponse `json:"dependencies"`
}

// DependencyResponse created dependency info
type DependencyResponse struct {
	Type            string            `json:"type"`
	Name            string            `json:"name"`
	InfraID         string            `json:"infra_id"`
	Status          string            `json:"status"`
	ContainerName   string            `json:"container_name"`
	Endpoints       map[string]string `json:"endpoints"`         // host, port, http_port, etc
	InjectedEnvVars map[string]string `json:"injected_env_vars"` // What ENV vars were set
}

package dto

// =============================================================================
// AUTO INFRASTRUCTURE DEPLOYMENT
// User provides: Image + Config → System auto-creates required infrastructure
// =============================================================================

// AutoDeployRequest - User gửi image và config, system tự động tạo infra
type AutoDeployRequest struct {
	// Tên deployment
	Name string `json:"name" binding:"required" example:"my-detection-app"`

	// Docker image cần deploy
	Image string `json:"image" binding:"required" example:"iaas-detection-engine:1.0"`

	// Environment variables từ image (system sẽ phân tích và tạo infra tương ứng)
	Environment map[string]string `json:"environment" example:"DB_HOST=clickhouse,DB_PORT=9000"`

	// Volumes cần mount
	Volumes []DeployVolumeMount `json:"volumes,omitempty"`

	// Port mà container expose
	ExposedPort int `json:"exposed_port" example:"8000"`

	// Resource limits
	CPU    float64 `json:"cpu" example:"1"`
	Memory int     `json:"memory" example:"512"` // MB
}

type DeployVolumeMount struct {
	HostPath      string `json:"host_path" example:"/home/ubuntu/rules"`
	ContainerPath string `json:"container_path" example:"/opt/rules_storage"`
}

// AutoDeployResponse - Kết quả deploy
type AutoDeployResponse struct {
	DeploymentID string `json:"deployment_id"`
	Name         string `json:"name"`
	Status       string `json:"status"` // creating, running, failed

	// Infra được tự động tạo
	CreatedInfrastructure []CreatedInfra `json:"created_infrastructure"`

	// Container info
	Container ContainerInfo `json:"container"`

	// Endpoints để access
	Endpoints map[string]string `json:"endpoints"`
}

type CreatedInfra struct {
	Type     string `json:"type"`     // clickhouse, postgresql, nginx, redis, etc.
	ID       string `json:"id"`       // ID của infra
	Name     string `json:"name"`     // Tên container/cluster
	Endpoint string `json:"endpoint"` // Connection string
	Status   string `json:"status"`   // running, failed
}

type ContainerInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Status   string `json:"status"`
	Endpoint string `json:"endpoint"` // API endpoint nếu có
}

// =============================================================================
// INFRASTRUCTURE DETECTION RULES
// Dựa vào env vars để detect loại infra cần tạo
// =============================================================================

// InfraDetectionRule - Rule để detect infra từ env
type InfraDetectionRule struct {
	InfraType    string   // clickhouse, postgresql, mysql, redis, nginx, kafka, etc.
	EnvPatterns  []string // DB_HOST, CLICKHOUSE_HOST, POSTGRES_HOST, REDIS_HOST, etc.
	DefaultPort  string   // 9000, 5432, 3306, 6379, etc.
	PortEnvNames []string // DB_PORT, CLICKHOUSE_PORT, etc.
}

// Predefined detection rules
var InfraDetectionRules = []InfraDetectionRule{
	{
		InfraType:    "clickhouse",
		EnvPatterns:  []string{"CLICKHOUSE_HOST", "CH_HOST", "DB_HOST"},
		DefaultPort:  "9000",
		PortEnvNames: []string{"CLICKHOUSE_PORT", "CH_PORT", "DB_PORT"},
	},
	{
		InfraType:    "postgresql",
		EnvPatterns:  []string{"POSTGRES_HOST", "PG_HOST", "PGHOST"},
		DefaultPort:  "5432",
		PortEnvNames: []string{"POSTGRES_PORT", "PG_PORT", "PGPORT"},
	},
	{
		InfraType:    "mysql",
		EnvPatterns:  []string{"MYSQL_HOST", "MARIADB_HOST"},
		DefaultPort:  "3306",
		PortEnvNames: []string{"MYSQL_PORT", "MARIADB_PORT"},
	},
	{
		InfraType:    "redis",
		EnvPatterns:  []string{"REDIS_HOST", "REDIS_URL"},
		DefaultPort:  "6379",
		PortEnvNames: []string{"REDIS_PORT"},
	},
	{
		InfraType:    "kafka",
		EnvPatterns:  []string{"KAFKA_BROKER", "KAFKA_HOST", "KAFKA_BOOTSTRAP_SERVERS"},
		DefaultPort:  "9092",
		PortEnvNames: []string{"KAFKA_PORT"},
	},
	{
		InfraType:    "elasticsearch",
		EnvPatterns:  []string{"ELASTICSEARCH_HOST", "ES_HOST", "ELASTIC_HOST"},
		DefaultPort:  "9200",
		PortEnvNames: []string{"ELASTICSEARCH_PORT", "ES_PORT"},
	},
	{
		InfraType:    "mongodb",
		EnvPatterns:  []string{"MONGO_HOST", "MONGODB_HOST", "MONGO_URL"},
		DefaultPort:  "27017",
		PortEnvNames: []string{"MONGO_PORT", "MONGODB_PORT"},
	},
}

// =============================================================================
// STATUS & MANAGEMENT
// =============================================================================

// GetDeploymentResponse - Chi tiết một deployment
type GetDeploymentResponse struct {
	DeploymentID   string            `json:"deployment_id"`
	Name           string            `json:"name"`
	Status         string            `json:"status"`
	Image          string            `json:"image"`
	Infrastructure []CreatedInfra    `json:"infrastructure"`
	Container      ContainerInfo     `json:"container"`
	Endpoints      map[string]string `json:"endpoints"`
	CreatedAt      string            `json:"created_at"`
	UpdatedAt      string            `json:"updated_at"`
}

// ListDeploymentsResponse - Danh sách deployments
type ListDeploymentsResponse struct {
	Total       int                     `json:"total"`
	Deployments []GetDeploymentResponse `json:"deployments"`
}

// DeploymentLogsResponse - Logs của deployment
type DeploymentLogsResponse struct {
	DeploymentID string   `json:"deployment_id"`
	Logs         []string `json:"logs"`
}

// DeploymentHealthResponse - Health check
type DeploymentHealthResponse struct {
	DeploymentID string            `json:"deployment_id"`
	Overall      string            `json:"overall"` // healthy, unhealthy, degraded
	Components   map[string]string `json:"components"`
	LastChecked  string            `json:"last_checked"`
}

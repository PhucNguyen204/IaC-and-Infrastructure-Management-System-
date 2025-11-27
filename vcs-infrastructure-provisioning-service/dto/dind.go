package dto

// CreateDinDEnvironmentRequest - Tạo môi trường Docker-in-Docker mới
type CreateDinDEnvironmentRequest struct {
	Name         string `json:"name" binding:"required"`         // Tên environment
	ResourcePlan string `json:"resource_plan"`                   // small, medium, large
	Description  string `json:"description"`                     // Mô tả
	AutoCleanup  bool   `json:"auto_cleanup"`                    // Tự động xóa sau TTL
	TTLHours     int    `json:"ttl_hours"`                       // Thời gian sống (giờ)
}

// DinDEnvironmentInfo - Thông tin môi trường DinD
type DinDEnvironmentInfo struct {
	ID               string `json:"id"`
	InfrastructureID string `json:"infrastructure_id"`
	Name             string `json:"name"`
	ContainerID      string `json:"container_id"`
	Status           string `json:"status"` // creating, running, stopped, failed
	DockerHost       string `json:"docker_host"`
	IPAddress        string `json:"ip_address"`
	ResourcePlan     string `json:"resource_plan"`
	CPULimit         string `json:"cpu_limit"`
	MemoryLimit      string `json:"memory_limit"`
	Description      string `json:"description"`
	AutoCleanup      bool   `json:"auto_cleanup"`
	TTLHours         int    `json:"ttl_hours"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
	ExpiresAt        string `json:"expires_at,omitempty"`
}

// ExecCommandRequest - Chạy docker command trong DinD environment
type ExecCommandRequest struct {
	Command string `json:"command" binding:"required"` // Docker command (e.g., "docker run nginx")
	Timeout int    `json:"timeout"`                    // Timeout in seconds (default 60)
}

// ExecCommandResponse - Kết quả thực thi command
type ExecCommandResponse struct {
	Command    string `json:"command"`
	Output     string `json:"output"`
	ExitCode   int    `json:"exit_code"`
	Duration   string `json:"duration"`
	ExecutedAt string `json:"executed_at"`
}

// BuildImageRequest - Build Docker image trong DinD environment
type BuildImageRequest struct {
	Dockerfile string            `json:"dockerfile" binding:"required"` // Nội dung Dockerfile
	ImageName  string            `json:"image_name" binding:"required"` // Tên image
	Tag        string            `json:"tag"`                           // Tag (default: latest)
	BuildArgs  map[string]string `json:"build_args"`                    // Build arguments
	NoCache    bool              `json:"no_cache"`                      // Build without cache
}

// BuildImageResponse - Kết quả build image
type BuildImageResponse struct {
	ImageName string   `json:"image_name"`
	Tag       string   `json:"tag"`
	ImageID   string   `json:"image_id"`
	Size      string   `json:"size"`
	BuildLogs []string `json:"build_logs"`
	Duration  string   `json:"duration"`
	Success   bool     `json:"success"`
}

// ComposeRequest - Chạy docker-compose trong DinD environment
type ComposeRequest struct {
	ComposeContent string `json:"compose_content" binding:"required"` // Nội dung docker-compose.yml
	Action         string `json:"action" binding:"required"`          // up, down, restart, logs, ps
	ServiceName    string `json:"service_name"`                       // Tên service cụ thể (optional)
	Detach         bool   `json:"detach"`                             // Run in background
}

// ComposeResponse - Kết quả docker-compose
type ComposeResponse struct {
	Action   string   `json:"action"`
	Output   string   `json:"output"`
	Services []string `json:"services"`
	Success  bool     `json:"success"`
}

// DinDContainerInfo - Thông tin container trong DinD environment
type DinDContainerInfo struct {
	ContainerID string `json:"container_id"`
	Name        string `json:"name"`
	Image       string `json:"image"`
	Status      string `json:"status"`
	Ports       string `json:"ports"`
	Created     string `json:"created"`
}

// DinDImageInfo - Thông tin image trong DinD environment
type DinDImageInfo struct {
	ImageID    string   `json:"image_id"`
	Repository string   `json:"repository"`
	Tag        string   `json:"tag"`
	Size       string   `json:"size"`
	Created    string   `json:"created"`
	Labels     []string `json:"labels"`
}

// ListContainersResponse - Danh sách containers
type ListContainersResponse struct {
	Containers []DinDContainerInfo `json:"containers"`
	Total      int                 `json:"total"`
}

// ListImagesResponse - Danh sách images
type ListImagesResponse struct {
	Images []DinDImageInfo `json:"images"`
	Total  int             `json:"total"`
}

// DinDLogsRequest - Request logs
type DinDLogsRequest struct {
	Tail       int  `form:"tail"`       // Số dòng cuối
	Follow     bool `form:"follow"`     // Stream logs
	Timestamps bool `form:"timestamps"` // Hiện timestamps
}

// DinDLogsResponse - Response logs
type DinDLogsResponse struct {
	Logs []string `json:"logs"`
}

// DinDStatsResponse - Thống kê resource của DinD environment
type DinDStatsResponse struct {
	CPUUsage      float64 `json:"cpu_usage_percent"`
	MemoryUsage   int64   `json:"memory_usage_bytes"`
	MemoryLimit   int64   `json:"memory_limit_bytes"`
	MemoryPercent float64 `json:"memory_usage_percent"`
	NetworkRx     int64   `json:"network_rx_bytes"`
	NetworkTx     int64   `json:"network_tx_bytes"`
	ContainerCount int    `json:"container_count"`
	ImageCount    int     `json:"image_count"`
}

// PullImageRequest - Pull image từ registry
type PullImageRequest struct {
	Image    string `json:"image" binding:"required"` // Image name (e.g., nginx:latest)
	Username string `json:"username"`                 // Registry username (optional)
	Password string `json:"password"`                 // Registry password (optional)
}

// PullImageResponse - Kết quả pull image
type PullImageResponse struct {
	Image   string `json:"image"`
	Status  string `json:"status"`
	Digest  string `json:"digest"`
	Success bool   `json:"success"`
}


package dto

type CreateDockerServiceRequest struct {
	Name          string            `json:"name" binding:"required"`
	Image         string            `json:"image" binding:"required"`
	ImageTag      string            `json:"image_tag" binding:"required"`
	ServiceType   string            `json:"service_type"`
	Command       string            `json:"command"`
	Args          string            `json:"args"`
	EnvVars       []EnvVarInput     `json:"env_vars"`
	Ports         []PortInput       `json:"ports" binding:"required"`
	Networks      []NetworkInput    `json:"networks"`
	Dependencies  []string          `json:"dependencies"`
	HealthCheck   *HealthCheckInput `json:"health_check"`
	RestartPolicy string            `json:"restart_policy"`
	Plan          string            `json:"plan"`
}

type EnvVarInput struct {
	Key      string `json:"key" binding:"required"`
	Value    string `json:"value"`
	IsSecret bool   `json:"is_secret"`
}

type NetworkInput struct {
	NetworkID string `json:"network_id" binding:"required"`
	Alias     string `json:"alias"`
}

type PortInput struct {
	ContainerPort int    `json:"container_port" binding:"required"`
	HostPort      int    `json:"host_port"`
	Protocol      string `json:"protocol"`
}

type HealthCheckInput struct {
	Type               string `json:"type" binding:"required"`
	HTTPPath           string `json:"http_path"`
	Port               int    `json:"port"`
	Command            string `json:"command"`
	Interval           int    `json:"interval"`
	Timeout            int    `json:"timeout"`
	HealthyThreshold   int    `json:"healthy_threshold"`
	UnhealthyThreshold int    `json:"unhealthy_threshold"`
}

type DockerServiceInfo struct {
	ID               string           `json:"id"`
	InfrastructureID string           `json:"infrastructure_id"`
	Name             string           `json:"name"`
	Image            string           `json:"image"`
	ImageTag         string           `json:"image_tag"`
	ServiceType      string           `json:"service_type"`
	ContainerID      string           `json:"container_id"`
	Status           string           `json:"status"`
	IPAddress        string           `json:"ip_address"`
	InternalEndpoint string           `json:"internal_endpoint"`
	EnvVars          []EnvVarInfo     `json:"env_vars"`
	Ports            []PortInfo       `json:"ports"`
	Networks         []NetworkInfo    `json:"networks"`
	HealthCheck      *HealthCheckInfo `json:"health_check"`
	RestartPolicy    string           `json:"restart_policy"`
	CPULimit         int64            `json:"cpu_limit"`
	MemoryLimit      int64            `json:"memory_limit"`
	CreatedAt        string           `json:"created_at"`
	UpdatedAt        string           `json:"updated_at"`
}

type EnvVarInfo struct {
	Key      string `json:"key"`
	Value    string `json:"value,omitempty"`
	IsSecret bool   `json:"is_secret"`
}

type PortInfo struct {
	ContainerPort int    `json:"container_port"`
	HostPort      int    `json:"host_port"`
	Protocol      string `json:"protocol"`
}

type NetworkInfo struct {
	NetworkID string `json:"network_id"`
	Alias     string `json:"alias"`
}

type HealthCheckInfo struct {
	Type               string `json:"type"`
	HTTPPath           string `json:"http_path,omitempty"`
	Port               int    `json:"port,omitempty"`
	Command            string `json:"command,omitempty"`
	Interval           int    `json:"interval"`
	Timeout            int    `json:"timeout"`
	HealthyThreshold   int    `json:"healthy_threshold"`
	UnhealthyThreshold int    `json:"unhealthy_threshold"`
	Status             string `json:"status"`
	LastCheck          string `json:"last_check,omitempty"`
}

type UpdateDockerEnvRequest struct {
	EnvVars []EnvVarInput `json:"env_vars" binding:"required"`
}

type UpdateDockerHealthCheckRequest struct {
	HealthCheck HealthCheckInput `json:"health_check" binding:"required"`
}

type DockerServiceLogsRequest struct {
	Tail       int  `form:"tail"`
	Follow     bool `form:"follow"`
	Timestamps bool `form:"timestamps"`
}

type DockerServiceLogsResponse struct {
	Logs []string `json:"logs"`
}

package dto

import "time"

// CreateK8sClusterRequest represents the request to create a Kubernetes cluster
type CreateK8sClusterRequest struct {
	ClusterName      string `json:"cluster_name" binding:"required"`
	K8sVersion       string `json:"k8s_version"`                    // Default: latest k3s
	ClusterType      string `json:"cluster_type"`                   // "k3s" or "k8s", default: "k3s"
	NodeCount        int    `json:"node_count" binding:"required,min=1,max=10"`
	
	// Resource Limits per Node
	CPULimit         string `json:"cpu_limit"`                      // Default: "2"
	MemoryLimit      string `json:"memory_limit"`                   // Default: "2Gi"
	
	// Network Configuration
	APIServerPort    int    `json:"api_server_port"`                // Default: auto-assign
	ClusterCIDR      string `json:"cluster_cidr"`                   // Default: "10.42.0.0/16"
	ServiceCIDR      string `json:"service_cidr"`                   // Default: "10.43.0.0/16"
	
	// Addons
	DashboardEnabled bool   `json:"dashboard_enabled"`              // Default: true
	IngressEnabled   bool   `json:"ingress_enabled"`                // Default: false
	MetricsEnabled   bool   `json:"metrics_enabled"`                // Default: true
	StorageClass     string `json:"storage_class"`                  // Default: "local-path"
}

// K8sClusterInfoResponse represents the response for cluster information
type K8sClusterInfoResponse struct {
	ID               string          `json:"id"`
	InfrastructureID string          `json:"infrastructure_id"`
	ClusterName      string          `json:"cluster_name"`
	K8sVersion       string          `json:"k8s_version"`
	ClusterType      string          `json:"cluster_type"`
	Status           string          `json:"status"`
	NodeCount        int             `json:"node_count"`
	
	// Nodes
	Nodes            []K8sNodeInfo   `json:"nodes"`
	
	// Connection Info
	APIServerURL     string          `json:"api_server_url"`
	APIServerPort    int             `json:"api_server_port"`
	Kubeconfig       string          `json:"kubeconfig,omitempty"`
	
	// Dashboard
	DashboardEnabled bool            `json:"dashboard_enabled"`
	DashboardURL     string          `json:"dashboard_url,omitempty"`
	DashboardToken   string          `json:"dashboard_token,omitempty"`
	
	// Network
	ClusterCIDR      string          `json:"cluster_cidr"`
	ServiceCIDR      string          `json:"service_cidr"`
	
	// Addons
	IngressEnabled   bool            `json:"ingress_enabled"`
	MetricsEnabled   bool            `json:"metrics_enabled"`
	StorageClass     string          `json:"storage_class"`
	
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

// K8sNodeInfo represents node information
type K8sNodeInfo struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Role           string    `json:"role"`
	Status         string    `json:"status"`
	IsReady        bool      `json:"is_ready"`
	IPAddress      string    `json:"ip_address"`
	ContainerID    string    `json:"container_id"`
	
	// Resources
	CPUCapacity    string    `json:"cpu_capacity"`
	MemoryCapacity string    `json:"memory_capacity"`
	PodCapacity    int       `json:"pod_capacity"`
	
	// Node Info
	KubeletVersion string    `json:"kubelet_version"`
	OSImage        string    `json:"os_image"`
	KernelVersion  string    `json:"kernel_version"`
	
	CreatedAt      time.Time `json:"created_at"`
}

// K8sClusterHealthResponse represents cluster health status
type K8sClusterHealthResponse struct {
	ClusterID      string           `json:"cluster_id"`
	ClusterName    string           `json:"cluster_name"`
	Status         string           `json:"status"`
	HealthyNodes   int              `json:"healthy_nodes"`
	TotalNodes     int              `json:"total_nodes"`
	APIServerReady bool             `json:"api_server_ready"`
	NodeHealth     []K8sNodeHealth  `json:"node_health"`
	LastCheck      time.Time        `json:"last_check"`
}

// K8sNodeHealth represents node health status
type K8sNodeHealth struct {
	NodeID         string    `json:"node_id"`
	NodeName       string    `json:"node_name"`
	Role           string    `json:"role"`
	IsReady        bool      `json:"is_ready"`
	Status         string    `json:"status"`
	Conditions     []string  `json:"conditions,omitempty"`
	LastHeartbeat  time.Time `json:"last_heartbeat"`
}

// K8sClusterMetricsResponse represents cluster metrics
type K8sClusterMetricsResponse struct {
	ClusterID         string              `json:"cluster_id"`
	TotalPods         int                 `json:"total_pods"`
	RunningPods       int                 `json:"running_pods"`
	TotalDeployments  int                 `json:"total_deployments"`
	TotalServices     int                 `json:"total_services"`
	TotalNamespaces   int                 `json:"total_namespaces"`
	NodeMetrics       []K8sNodeMetrics    `json:"node_metrics"`
	CollectedAt       time.Time           `json:"collected_at"`
}

// K8sNodeMetrics represents node resource metrics
type K8sNodeMetrics struct {
	NodeName       string  `json:"node_name"`
	CPUUsage       string  `json:"cpu_usage"`
	MemoryUsage    string  `json:"memory_usage"`
	CPUPercent     float64 `json:"cpu_percent"`
	MemoryPercent  float64 `json:"memory_percent"`
	PodCount       int     `json:"pod_count"`
}

// K8sConnectionInfoResponse represents connection information
type K8sConnectionInfoResponse struct {
	ClusterID      string `json:"cluster_id"`
	ClusterName    string `json:"cluster_name"`
	APIServerURL   string `json:"api_server_url"`
	Kubeconfig     string `json:"kubeconfig"`
	DashboardURL   string `json:"dashboard_url,omitempty"`
	DashboardToken string `json:"dashboard_token,omitempty"`
	
	// kubectl commands
	KubectlCommands struct {
		GetNodes       string `json:"get_nodes"`
		GetPods        string `json:"get_pods"`
		GetServices    string `json:"get_services"`
		GetDeployments string `json:"get_deployments"`
	} `json:"kubectl_commands"`
}

// ScaleK8sClusterRequest represents request to scale cluster
type ScaleK8sClusterRequest struct {
	NodeCount int `json:"node_count" binding:"required,min=1,max=10"`
}


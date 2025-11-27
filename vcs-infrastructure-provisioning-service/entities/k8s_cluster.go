package entities

import (
	"time"

	"gorm.io/gorm"
)

// K8sCluster represents a Kubernetes cluster
type K8sCluster struct {
	ID               string         `gorm:"type:uuid;primaryKey" json:"id"`
	InfrastructureID string         `gorm:"type:uuid;not null;index" json:"infrastructure_id"`
	ClusterName      string         `gorm:"type:varchar(255);not null;uniqueIndex" json:"cluster_name"`
	K8sVersion       string         `gorm:"type:varchar(50);not null" json:"k8s_version"` // e.g., "v1.28.5-k3s1"
	ClusterType      string         `gorm:"type:varchar(50);not null" json:"cluster_type"` // "k3s" or "k8s"
	NodeCount        int            `gorm:"not null" json:"node_count"`
	Status           string         `gorm:"type:varchar(50);not null" json:"status"` // creating, running, stopped, failed
	
	// Network Configuration
	APIServerPort    int            `gorm:"not null" json:"api_server_port"`
	NetworkName      string         `gorm:"type:varchar(255)" json:"network_name"`
	ClusterCIDR      string         `gorm:"type:varchar(100)" json:"cluster_cidr"` // Pod network CIDR
	ServiceCIDR      string         `gorm:"type:varchar(100)" json:"service_cidr"` // Service network CIDR
	
	// Resource Limits per Node
	CPULimit         string         `gorm:"type:varchar(50)" json:"cpu_limit"`     // e.g., "2"
	MemoryLimit      string         `gorm:"type:varchar(50)" json:"memory_limit"`  // e.g., "2Gi"
	
	// Kubeconfig
	Kubeconfig       string         `gorm:"type:text" json:"kubeconfig"`
	
	// Dashboard
	DashboardEnabled bool           `gorm:"default:true" json:"dashboard_enabled"`
	DashboardPort    int            `gorm:"default:0" json:"dashboard_port"`
	DashboardToken   string         `gorm:"type:text" json:"dashboard_token"`
	
	// Addons
	IngressEnabled   bool           `gorm:"default:false" json:"ingress_enabled"`
	MetricsEnabled   bool           `gorm:"default:true" json:"metrics_enabled"`
	StorageClass     string         `gorm:"type:varchar(100)" json:"storage_class"` // local-path, nfs, etc.
	
	// Metadata
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
	
	// Relations
	Nodes            []K8sNode      `gorm:"foreignKey:ClusterID;constraint:OnDelete:CASCADE" json:"nodes,omitempty"`
}

// K8sNode represents a node in the Kubernetes cluster
type K8sNode struct {
	ID          string         `gorm:"type:uuid;primaryKey" json:"id"`
	ClusterID   string         `gorm:"type:uuid;not null;index" json:"cluster_id"`
	Name        string         `gorm:"type:varchar(255);not null" json:"name"`
	Role        string         `gorm:"type:varchar(50);not null" json:"role"` // "server" (master) or "agent" (worker)
	ContainerID string         `gorm:"type:varchar(255);uniqueIndex" json:"container_id"`
	IPAddress   string         `gorm:"type:varchar(50)" json:"ip_address"`
	Status      string         `gorm:"type:varchar(50);not null" json:"status"` // running, stopped, failed
	IsReady     bool           `gorm:"default:false" json:"is_ready"`
	
	// Node Resources
	CPUCapacity    string         `gorm:"type:varchar(50)" json:"cpu_capacity"`
	MemoryCapacity string         `gorm:"type:varchar(50)" json:"memory_capacity"`
	PodCapacity    int            `gorm:"default:110" json:"pod_capacity"`
	
	// Node Info
	KubeletVersion string         `gorm:"type:varchar(50)" json:"kubelet_version"`
	OSImage        string         `gorm:"type:varchar(255)" json:"os_image"`
	KernelVersion  string         `gorm:"type:varchar(100)" json:"kernel_version"`
	
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for K8sCluster
func (K8sCluster) TableName() string {
	return "k8s_clusters"
}

// TableName specifies the table name for K8sNode
func (K8sNode) TableName() string {
	return "k8s_nodes"
}


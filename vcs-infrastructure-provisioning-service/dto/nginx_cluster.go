package dto

// ================== Create Nginx Cluster ==================

// CreateNginxClusterRequest request to create a new Nginx HA cluster
type CreateNginxClusterRequest struct {
	// Basic Configuration
	ClusterName     string `json:"cluster_name" binding:"required"`
	NodeCount       int    `json:"node_count" binding:"required,min=2,max=10"`
	HTTPPort        int    `json:"http_port" binding:"required,min=1024,max=65535"`
	HTTPSPort       int    `json:"https_port"`
	LoadBalanceMode string `json:"load_balance_mode"` // round_robin, least_conn, ip_hash, random

	// High Availability Configuration
	VirtualIP           string `json:"virtual_ip"`            // Optional VIP for Keepalived
	VRRPInterface       string `json:"vrrp_interface"`        // Network interface for VRRP (default: eth0)
	VRRPRouterID        int    `json:"vrrp_router_id"`        // VRRP Router ID (1-255)
	HealthCheckEnabled  bool   `json:"health_check_enabled"`  // Enable health checks
	HealthCheckPath     string `json:"health_check_path"`     // Health check endpoint (default: /health)
	HealthCheckInterval int    `json:"health_check_interval"` // Interval in seconds (default: 5)

	// Resource Limits
	CPUPerNode    int64 `json:"cpu_per_node"`    // CPU limit in nanocores
	MemoryPerNode int64 `json:"memory_per_node"` // Memory limit in bytes

	// SSL/TLS Configuration
	SSLEnabled        bool   `json:"ssl_enabled"`
	SSLCertificate    string `json:"ssl_certificate,omitempty"`
	SSLPrivateKey     string `json:"ssl_private_key,omitempty"`
	SSLProtocols      string `json:"ssl_protocols,omitempty"`       // e.g., "TLSv1.2 TLSv1.3"
	SSLCiphers        string `json:"ssl_ciphers,omitempty"`         // SSL cipher suites
	SSLSessionTimeout string `json:"ssl_session_timeout,omitempty"` // e.g., "1d"

	// Performance Tuning
	WorkerProcesses   int    `json:"worker_processes"`     // Number of worker processes (default: auto)
	WorkerConnections int    `json:"worker_connections"`   // Max connections per worker (default: 1024)
	KeepaliveTimeout  int    `json:"keepalive_timeout"`    // Keepalive timeout in seconds (default: 65)
	ClientMaxBodySize string `json:"client_max_body_size"` // Max request body size (default: 1m)

	// Logging Configuration
	AccessLogEnabled bool   `json:"access_log_enabled"` // Enable access logging
	ErrorLogLevel    string `json:"error_log_level"`    // error, warn, info, debug

	// Caching Configuration
	CacheEnabled bool   `json:"cache_enabled"`
	CachePath    string `json:"cache_path"`
	CacheSize    string `json:"cache_size"` // e.g., "100m"

	// Rate Limiting
	RateLimitEnabled        bool `json:"rate_limit_enabled"`
	RateLimitRequestsPerSec int  `json:"rate_limit_requests_per_sec"`
	RateLimitBurst          int  `json:"rate_limit_burst"`

	// Gzip Compression
	GzipEnabled   bool   `json:"gzip_enabled"`
	GzipLevel     int    `json:"gzip_level"`      // 1-9 (default: 6)
	GzipMinLength int    `json:"gzip_min_length"` // Min size to compress (default: 1000)
	GzipTypes     string `json:"gzip_types"`      // MIME types to compress

	// Backend Configuration
	Upstreams    []CreateUpstreamRequest    `json:"upstreams,omitempty"`
	ServerBlocks []CreateServerBlockRequest `json:"server_blocks,omitempty"`
}

// CreateUpstreamRequest defines an upstream group
type CreateUpstreamRequest struct {
	Name        string                        `json:"name" binding:"required"`
	Algorithm   string                        `json:"algorithm"` // round_robin, least_conn, ip_hash
	Servers     []CreateUpstreamServerRequest `json:"servers" binding:"required"`
	HealthCheck bool                          `json:"health_check"`
	HealthPath  string                        `json:"health_path"`
}

// CreateUpstreamServerRequest defines a backend server
type CreateUpstreamServerRequest struct {
	Address     string `json:"address" binding:"required"` // host:port
	Weight      int    `json:"weight"`
	MaxFails    int    `json:"max_fails"`
	FailTimeout int    `json:"fail_timeout"`
	IsBackup    bool   `json:"is_backup"`
}

// CreateServerBlockRequest defines a virtual host
type CreateServerBlockRequest struct {
	ServerName string                  `json:"server_name" binding:"required"`
	ListenPort int                     `json:"listen_port"`
	SSLEnabled bool                    `json:"ssl_enabled"`
	RootPath   string                  `json:"root_path"`
	Locations  []CreateLocationRequest `json:"locations"`
}

// CreateLocationRequest defines a location block
type CreateLocationRequest struct {
	Path         string            `json:"path" binding:"required"`
	ProxyPass    string            `json:"proxy_pass"`
	ProxyHeaders map[string]string `json:"proxy_headers"`
	CacheEnabled bool              `json:"cache_enabled"`
	RateLimit    int               `json:"rate_limit"`
}

// ================== Responses ==================

// NginxClusterInfoResponse detailed cluster information
type NginxClusterInfoResponse struct {
	ID               string            `json:"id"`
	InfrastructureID string            `json:"infrastructure_id"`
	ClusterName      string            `json:"cluster_name"`
	Status           string            `json:"status"`
	NodeCount        int               `json:"node_count"`
	MasterNode       *NginxNodeInfo    `json:"master_node,omitempty"`
	Nodes            []NginxNodeInfo   `json:"nodes"`
	VirtualIP        string            `json:"virtual_ip,omitempty"`
	HTTPPort         int               `json:"http_port"`
	HTTPSPort        int               `json:"https_port,omitempty"`
	LoadBalanceMode  string            `json:"load_balance_mode"`
	SSLEnabled       bool              `json:"ssl_enabled"`
	Upstreams        []UpstreamInfo    `json:"upstreams,omitempty"`
	ServerBlocks     []ServerBlockInfo `json:"server_blocks,omitempty"`
	Endpoints        NginxEndpoints    `json:"endpoints"`
	CreatedAt        string            `json:"created_at"`
	UpdatedAt        string            `json:"updated_at"`
}

// NginxNodeInfo represents a node in the cluster
type NginxNodeInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Role        string `json:"role"` // master, backup
	Priority    int    `json:"priority"`
	Status      string `json:"status"`
	ContainerID string `json:"container_id"`
	IPAddress   string `json:"ip_address,omitempty"`
	HTTPPort    int    `json:"http_port"`
	HTTPSPort   int    `json:"https_port,omitempty"`
	IsHealthy   bool   `json:"is_healthy"`
	IsMaster    bool   `json:"is_master"`
}

// NginxEndpoints connection endpoints
type NginxEndpoints struct {
	VirtualIP  string `json:"virtual_ip,omitempty"`
	HTTPURL    string `json:"http_url"`
	HTTPSURL   string `json:"https_url,omitempty"`
	MasterNode string `json:"master_node"`
}

// ServerBlockInfo virtual host information
type ServerBlockInfo struct {
	ID         string         `json:"id"`
	ServerName string         `json:"server_name"`
	ListenPort int            `json:"listen_port"`
	SSLEnabled bool           `json:"ssl_enabled"`
	Locations  []LocationInfo `json:"locations,omitempty"`
}

// LocationInfo location block information
type LocationInfo struct {
	ID           string            `json:"id"`
	Path         string            `json:"path"`
	ProxyPass    string            `json:"proxy_pass,omitempty"`
	ProxyHeaders map[string]string `json:"proxy_headers,omitempty"`
	CacheEnabled bool              `json:"cache_enabled"`
	RateLimit    int               `json:"rate_limit"`
}

// ================== Cluster Operations ==================

// AddNginxNodeRequest add a new node to cluster
type AddNginxNodeRequest struct {
	Priority int `json:"priority"` // Keepalived priority
}

// UpdateNginxConfigRequest update nginx configuration
type UpdateNginxClusterConfigRequest struct {
	NginxConfig string `json:"nginx_config" binding:"required"`
	ReloadAll   bool   `json:"reload_all"` // Reload all nodes
}

// AddUpstreamRequest add upstream to cluster
type AddNginxUpstreamRequest struct {
	Name        string                        `json:"name" binding:"required"`
	Algorithm   string                        `json:"algorithm"`
	Servers     []CreateUpstreamServerRequest `json:"servers" binding:"required"`
	HealthCheck bool                          `json:"health_check"`
	HealthPath  string                        `json:"health_path"`
}

// UpdateUpstreamRequest update upstream configuration
type UpdateNginxUpstreamRequest struct {
	Algorithm   string                        `json:"algorithm"`
	Servers     []CreateUpstreamServerRequest `json:"servers"`
	HealthCheck bool                          `json:"health_check"`
	HealthPath  string                        `json:"health_path"`
}

// AddServerBlockRequest add a server block
type AddNginxServerBlockRequest struct {
	ServerName string                  `json:"server_name" binding:"required"`
	ListenPort int                     `json:"listen_port"`
	SSLEnabled bool                    `json:"ssl_enabled"`
	RootPath   string                  `json:"root_path"`
	Locations  []CreateLocationRequest `json:"locations"`
}

// AddLocationRequest add a location to server block
type AddNginxLocationRequest struct {
	Path         string            `json:"path" binding:"required"`
	ProxyPass    string            `json:"proxy_pass"`
	ProxyHeaders map[string]string `json:"proxy_headers"`
	CacheEnabled bool              `json:"cache_enabled"`
	RateLimit    int               `json:"rate_limit"`
}

// ================== Health & Monitoring ==================

// NginxClusterHealthResponse health check response
type NginxClusterHealthResponse struct {
	ClusterID      string               `json:"cluster_id"`
	ClusterName    string               `json:"cluster_name"`
	Status         string               `json:"status"` // healthy, degraded, unhealthy
	HealthyNodes   int                  `json:"healthy_nodes"`
	TotalNodes     int                  `json:"total_nodes"`
	MasterNode     string               `json:"master_node"`
	VIPActive      bool                 `json:"vip_active"`
	NodeHealth     []NodeHealthInfo     `json:"node_health"`
	UpstreamHealth []UpstreamHealthInfo `json:"upstream_health,omitempty"`
}

// NodeHealthInfo individual node health
type NodeHealthInfo struct {
	NodeID           string `json:"node_id"`
	NodeName         string `json:"node_name"`
	Role             string `json:"role"`
	IsHealthy        bool   `json:"is_healthy"`
	NginxStatus      string `json:"nginx_status"` // running, stopped, error
	KeepalivedStatus string `json:"keepalived_status"`
	LastCheck        string `json:"last_check"`
}

// UpstreamHealthInfo upstream health status
type UpstreamHealthInfo struct {
	Name           string   `json:"name"`
	HealthyServers int      `json:"healthy_servers"`
	TotalServers   int      `json:"total_servers"`
	UnhealthyList  []string `json:"unhealthy_list,omitempty"`
}

// ================== Failover ==================

// TriggerFailoverRequest manual failover request
type TriggerNginxFailoverRequest struct {
	TargetNodeID string `json:"target_node_id" binding:"required"`
	Reason       string `json:"reason"`
}

// FailoverResponse failover result
type NginxFailoverResponse struct {
	Success       bool   `json:"success"`
	Message       string `json:"message"`
	OldMasterID   string `json:"old_master_id"`
	OldMasterName string `json:"old_master_name"`
	NewMasterID   string `json:"new_master_id"`
	NewMasterName string `json:"new_master_name"`
	Duration      string `json:"duration"`
}

// FailoverHistoryResponse failover history
type NginxFailoverHistoryResponse struct {
	ClusterID string               `json:"cluster_id"`
	Events    []NginxFailoverEvent `json:"events"`
}

// NginxFailoverEvent single failover event
type NginxFailoverEvent struct {
	ID            string `json:"id"`
	OldMasterID   string `json:"old_master_id"`
	OldMasterName string `json:"old_master_name"`
	NewMasterID   string `json:"new_master_id"`
	NewMasterName string `json:"new_master_name"`
	Reason        string `json:"reason"`
	TriggeredBy   string `json:"triggered_by"`
	OccurredAt    string `json:"occurred_at"`
}

// ================== Metrics & Stats ==================

// NginxClusterMetricsResponse cluster-wide metrics
type NginxClusterMetricsResponse struct {
	ClusterID      string             `json:"cluster_id"`
	TotalRequests  int64              `json:"total_requests"`
	RequestsPerSec float64            `json:"requests_per_sec"`
	ActiveConns    int                `json:"active_connections"`
	Status2xx      int64              `json:"status_2xx"`
	Status4xx      int64              `json:"status_4xx"`
	Status5xx      int64              `json:"status_5xx"`
	AvgLatencyMs   float64            `json:"avg_latency_ms"`
	NodeMetrics    []NginxNodeMetrics `json:"node_metrics"`
}

// NginxNodeMetrics individual node metrics
type NginxNodeMetrics struct {
	NodeID         string  `json:"node_id"`
	NodeName       string  `json:"node_name"`
	Role           string  `json:"role"`
	TotalRequests  int64   `json:"total_requests"`
	RequestsPerSec float64 `json:"requests_per_sec"`
	ActiveConns    int     `json:"active_connections"`
	Status2xx      int64   `json:"status_2xx"`
	Status4xx      int64   `json:"status_4xx"`
	Status5xx      int64   `json:"status_5xx"`
}

// ================== Connection Info ==================

// NginxConnectionInfoResponse connection details for users
type NginxConnectionInfoResponse struct {
	ClusterID   string                   `json:"cluster_id"`
	ClusterName string                   `json:"cluster_name"`
	Status      string                   `json:"status"`
	Endpoints   NginxConnectionEndpoints `json:"endpoints"`
	ServerNames []string                 `json:"server_names"`
}

// NginxConnectionEndpoints all available endpoints
type NginxConnectionEndpoints struct {
	VirtualIP   *VIPEndpoint   `json:"virtual_ip,omitempty"`
	MasterNode  NodeEndpoint   `json:"master_node"`
	BackupNodes []NodeEndpoint `json:"backup_nodes,omitempty"`
}

// VIPEndpoint Virtual IP endpoint
type VIPEndpoint struct {
	IP        string `json:"ip"`
	HTTPPort  int    `json:"http_port"`
	HTTPSPort int    `json:"https_port,omitempty"`
	HTTPURL   string `json:"http_url"`
	HTTPSURL  string `json:"https_url,omitempty"`
}

// NodeEndpoint individual node endpoint
type NodeEndpoint struct {
	NodeID    string `json:"node_id"`
	NodeName  string `json:"node_name"`
	Role      string `json:"role"`
	IP        string `json:"ip"`
	HTTPPort  int    `json:"http_port"`
	HTTPSPort int    `json:"https_port,omitempty"`
	HTTPURL   string `json:"http_url"`
	HTTPSURL  string `json:"https_url,omitempty"`
	IsHealthy bool   `json:"is_healthy"`
}

// TestNginxConnectionResponse test connection result
type TestNginxConnectionResponse struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	Latency      string `json:"latency"`
	NodeName     string `json:"node_name"`
	NodeRole     string `json:"node_role"`
	NginxVersion string `json:"nginx_version,omitempty"`
	StatusCode   int    `json:"status_code"`
}

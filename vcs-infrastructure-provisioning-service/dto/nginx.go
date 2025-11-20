package dto

type CreateNginxRequest struct {
	Name        string            `json:"name" binding:"required"`
	Port        int               `json:"port" binding:"required,min=1024,max=65535"`
	SSLPort     int               `json:"ssl_port"`
	Config      string            `json:"config" binding:"required"`
	Upstreams   map[string]string `json:"upstreams"`
	CPULimit    int64             `json:"cpu_limit"`
	MemoryLimit int64             `json:"memory_limit"`
}

type NginxInfoResponse struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Status      string           `json:"status"`
	ContainerID string           `json:"container_id"`
	Port        int              `json:"port"`
	SSLPort     int              `json:"ssl_port"`
	Config      string           `json:"config"`
	Domains     []string         `json:"domains"`
	Routes      []RouteInfo      `json:"routes"`
	Upstreams   []UpstreamInfo   `json:"upstreams"`
	Certificate *CertificateInfo `json:"certificate,omitempty"`
	Security    *SecurityPolicy  `json:"security,omitempty"`
	CPULimit    int64            `json:"cpu_limit"`
	MemoryLimit int64            `json:"memory_limit"`
	CreatedAt   string           `json:"created_at"`
	UpdatedAt   string           `json:"updated_at"`
}

type UpdateNginxConfigRequest struct {
	Config string `json:"config" binding:"required"`
}

type AddDomainRequest struct {
	Domain string `json:"domain" binding:"required"`
}

type AddRouteRequest struct {
	Path     string `json:"path" binding:"required"`
	Backend  string `json:"backend" binding:"required"`
	Priority int    `json:"priority"`
}

type UpdateRouteRequest struct {
	Backend  string `json:"backend"`
	Priority int    `json:"priority"`
}

type RouteInfo struct {
	ID       string `json:"id"`
	Path     string `json:"path"`
	Backend  string `json:"backend"`
	Priority int    `json:"priority"`
}

type UploadCertificateRequest struct {
	Certificate string `json:"certificate" binding:"required"`
	PrivateKey  string `json:"private_key" binding:"required"`
	Domain      string `json:"domain" binding:"required"`
}

type CertificateInfo struct {
	Domain    string `json:"domain"`
	Status    string `json:"status"`
	ExpiresAt string `json:"expires_at"`
	Issuer    string `json:"issuer"`
}

type UpdateUpstreamsRequest struct {
	Backends []BackendServer `json:"backends" binding:"required"`
	Policy   string          `json:"policy"`
}

type BackendServer struct {
	Address string `json:"address" binding:"required"`
	Weight  int    `json:"weight"`
}

type UpstreamInfo struct {
	Name     string          `json:"name"`
	Backends []BackendServer `json:"backends"`
	Policy   string          `json:"policy"`
}

type SecurityPolicyRequest struct {
	RateLimit *RateLimitConfig `json:"rate_limit,omitempty"`
	IPFilter  *IPFilterConfig  `json:"ip_filter,omitempty"`
	BasicAuth *BasicAuthConfig `json:"basic_auth,omitempty"`
}

type RateLimitConfig struct {
	RequestsPerSecond int    `json:"requests_per_second"`
	Burst             int    `json:"burst"`
	Path              string `json:"path"`
}

type IPFilterConfig struct {
	AllowIPs []string `json:"allow_ips"`
	DenyIPs  []string `json:"deny_ips"`
}

type BasicAuthConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Realm    string `json:"realm"`
}

type SecurityPolicy struct {
	RateLimit *RateLimitConfig `json:"rate_limit,omitempty"`
	IPFilter  *IPFilterConfig  `json:"ip_filter,omitempty"`
	BasicAuth *BasicAuthConfig `json:"basic_auth,omitempty"`
}

type NginxLogsResponse struct {
	InstanceID string   `json:"instance_id"`
	Logs       []string `json:"logs"`
	Tail       int      `json:"tail"`
}

type NginxMetricsResponse struct {
	InstanceID     string  `json:"instance_id"`
	TotalRequests  int64   `json:"total_requests"`
	RequestsPerSec float64 `json:"requests_per_sec"`
	Status2xx      int64   `json:"status_2xx"`
	Status4xx      int64   `json:"status_4xx"`
	Status5xx      int64   `json:"status_5xx"`
	AvgLatencyMs   float64 `json:"avg_latency_ms"`
}

type NginxStatsResponse struct {
	InstanceID     string           `json:"instance_id"`
	ActiveConns    int              `json:"active_connections"`
	TotalRequests  int64            `json:"total_requests"`
	Status2xx      int64            `json:"status_2xx"`
	Status4xx      int64            `json:"status_4xx"`
	Status5xx      int64            `json:"status_5xx"`
	UpstreamHealth []UpstreamHealth `json:"upstream_health"`
}

type UpstreamHealth struct {
	Name    string `json:"name"`
	Healthy bool   `json:"healthy"`
	Address string `json:"address"`
}

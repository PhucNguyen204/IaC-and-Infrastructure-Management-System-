package entities

import (
	"time"
)

// NginxCluster represents a high-availability Nginx cluster with Keepalived
type NginxCluster struct {
	ID               string         `gorm:"primaryKey;type:varchar(36)"`
	InfrastructureID string         `gorm:"type:varchar(36);not null;index"`
	Infrastructure   Infrastructure `gorm:"foreignKey:InfrastructureID"`
	ClusterName      string         `gorm:"type:varchar(255)"`
	NodeCount        int            `gorm:"not null;default:2"`
	MasterNodeID     string         `gorm:"type:varchar(36)"`

	// High Availability
	VirtualIP           string `gorm:"type:varchar(45)"` // VIP managed by Keepalived
	VRRPInterface       string `gorm:"type:varchar(20);default:'eth0'"`
	VRRPRouterID        int    `gorm:"default:51"`
	HealthCheckEnabled  bool   `gorm:"default:true"`
	HealthCheckPath     string `gorm:"type:varchar(255);default:'/health'"`
	HealthCheckInterval int    `gorm:"default:5"`

	// Network
	HTTPPort  int    `gorm:"not null;default:80"`
	HTTPSPort int    `gorm:"default:443"`
	NetworkID string `gorm:"type:varchar(255)"`

	// Load Balancing
	LoadBalanceMode string `gorm:"type:varchar(20);default:'round_robin'"` // round_robin, least_conn, ip_hash

	// SSL/TLS
	SSLEnabled        bool   `gorm:"default:false"`
	SSLCertificate    string `gorm:"type:text"`
	SSLPrivateKey     string `gorm:"type:text"`
	SSLProtocols      string `gorm:"type:varchar(100);default:'TLSv1.2 TLSv1.3'"`
	SSLSessionTimeout string `gorm:"type:varchar(20);default:'1d'"`

	// Performance
	WorkerProcesses   int    `gorm:"default:0"` // 0 = auto
	WorkerConnections int    `gorm:"default:1024"`
	KeepaliveTimeout  int    `gorm:"default:65"`
	ClientMaxBodySize string `gorm:"type:varchar(20);default:'10m'"`

	// Logging
	AccessLogEnabled bool   `gorm:"default:true"`
	ErrorLogLevel    string `gorm:"type:varchar(20);default:'warn'"`

	// Caching
	CacheEnabled bool   `gorm:"default:false"`
	CachePath    string `gorm:"type:varchar(255)"`
	CacheSize    string `gorm:"type:varchar(20)"`

	// Rate Limiting
	RateLimitEnabled        bool `gorm:"default:false"`
	RateLimitRequestsPerSec int  `gorm:"default:100"`
	RateLimitBurst          int  `gorm:"default:50"`

	// Gzip
	GzipEnabled   bool   `gorm:"default:true"`
	GzipLevel     int    `gorm:"default:6"`
	GzipMinLength int    `gorm:"default:1000"`
	GzipTypes     string `gorm:"type:varchar(500);default:'text/plain text/css application/json application/javascript'"`

	// Config
	NginxConfig string `gorm:"type:text"` // Generated nginx.conf

	// Resources
	CPULimit    int64 `gorm:"default:0"`
	MemoryLimit int64 `gorm:"default:0"`

	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (NginxCluster) TableName() string {
	return "nginx_clusters"
}

type NginxNode struct {
	ID           string       `gorm:"primaryKey;type:varchar(36)"`
	ClusterID    string       `gorm:"type:varchar(36);not null;index"`
	Cluster      NginxCluster `gorm:"foreignKey:ClusterID"`
	Name         string       `gorm:"type:varchar(100)"`
	ContainerID  string       `gorm:"type:varchar(100)"`
	Role         string       `gorm:"type:varchar(20);not null"` // master, backup
	Priority     int          `gorm:"not null"`                  // Keepalived priority (higher = more preferred)
	HTTPPort     int          `gorm:"not null"`
	HTTPSPort    int          `gorm:"default:0"`
	Status       string       `gorm:"type:varchar(20);default:'pending'"`
	IPAddress    string       `gorm:"type:varchar(45)"`
	IsHealthy    bool         `gorm:"default:true"`
	LastHealthAt time.Time    `gorm:"autoCreateTime"`
	CreatedAt    time.Time    `gorm:"autoCreateTime"`
	UpdatedAt    time.Time    `gorm:"autoUpdateTime"`
}

func (NginxNode) TableName() string {
	return "nginx_nodes"
}

type NginxClusterUpstream struct {
	ID          string       `gorm:"primaryKey;type:varchar(36)"`
	ClusterID   string       `gorm:"type:varchar(36);not null;index"`
	Cluster     NginxCluster `gorm:"foreignKey:ClusterID"`
	Name        string       `gorm:"type:varchar(100);not null"`
	Algorithm   string       `gorm:"type:varchar(20);default:'round_robin'"` // round_robin, least_conn, ip_hash
	HealthCheck bool         `gorm:"default:true"`
	HealthPath  string       `gorm:"type:varchar(255);default:'/health'"`
	CreatedAt   time.Time    `gorm:"autoCreateTime"`
	UpdatedAt   time.Time    `gorm:"autoUpdateTime"`
}

func (NginxClusterUpstream) TableName() string {
	return "nginx_cluster_upstreams"
}

type NginxUpstreamServer struct {
	ID          string               `gorm:"primaryKey;type:varchar(36)"`
	UpstreamID  string               `gorm:"type:varchar(36);not null;index"`
	Upstream    NginxClusterUpstream `gorm:"foreignKey:UpstreamID"`
	Address     string               `gorm:"type:varchar(255);not null"` // host:port
	Weight      int                  `gorm:"default:1"`
	MaxFails    int                  `gorm:"default:3"`
	FailTimeout int                  `gorm:"default:30"` // seconds
	IsBackup    bool                 `gorm:"default:false"`
	IsDown      bool                 `gorm:"default:false"`
	CreatedAt   time.Time            `gorm:"autoCreateTime"`
	UpdatedAt   time.Time            `gorm:"autoUpdateTime"`
}

func (NginxUpstreamServer) TableName() string {
	return "nginx_upstream_servers"
}

type NginxServerBlock struct {
	ID         string       `gorm:"primaryKey;type:varchar(36)"`
	ClusterID  string       `gorm:"type:varchar(36);not null;index"`
	Cluster    NginxCluster `gorm:"foreignKey:ClusterID"`
	ServerName string       `gorm:"type:varchar(255);not null"` // domain name
	ListenPort int          `gorm:"default:80"`
	SSLEnabled bool         `gorm:"default:false"`
	SSLCertID  string       `gorm:"type:varchar(36)"`
	RootPath   string       `gorm:"type:varchar(255)"`
	IndexFiles string       `gorm:"type:varchar(255);default:'index.html index.htm'"`
	CreatedAt  time.Time    `gorm:"autoCreateTime"`
	UpdatedAt  time.Time    `gorm:"autoUpdateTime"`
}

func (NginxServerBlock) TableName() string {
	return "nginx_server_blocks"
}

type NginxLocation struct {
	ID            string           `gorm:"primaryKey;type:varchar(36)"`
	ServerBlockID string           `gorm:"type:varchar(36);not null;index"`
	ServerBlock   NginxServerBlock `gorm:"foreignKey:ServerBlockID"`
	Path          string           `gorm:"type:varchar(255);not null"` // /api, /static, etc.
	ProxyPass     string           `gorm:"type:varchar(255)"`          // upstream name or URL
	ProxyHeaders  string           `gorm:"type:text"`                  // JSON of proxy headers
	CacheEnabled  bool             `gorm:"default:false"`
	RateLimit     int              `gorm:"default:0"` // requests per second
	CreatedAt     time.Time        `gorm:"autoCreateTime"`
	UpdatedAt     time.Time        `gorm:"autoUpdateTime"`
}

func (NginxLocation) TableName() string {
	return "nginx_locations"
}

type NginxFailoverEvent struct {
	ID            string       `gorm:"primaryKey;type:varchar(36)"`
	ClusterID     string       `gorm:"type:varchar(36);not null;index"`
	Cluster       NginxCluster `gorm:"foreignKey:ClusterID"`
	OldMasterID   string       `gorm:"type:varchar(36)"`
	OldMasterName string       `gorm:"type:varchar(100)"`
	NewMasterID   string       `gorm:"type:varchar(36)"`
	NewMasterName string       `gorm:"type:varchar(100)"`
	Reason        string       `gorm:"type:varchar(50)"` // manual, automatic, node_failure
	TriggeredBy   string       `gorm:"type:varchar(50)"` // system, user, keepalived
	OccurredAt    time.Time    `gorm:"autoCreateTime"`
}

func (NginxFailoverEvent) TableName() string {
	return "nginx_failover_events"
}

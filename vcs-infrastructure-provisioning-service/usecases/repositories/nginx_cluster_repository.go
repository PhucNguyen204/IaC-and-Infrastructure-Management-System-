package repositories

import (
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/entities"
	"gorm.io/gorm"
)

// INginxClusterRepository interface for Nginx cluster data operations
type INginxClusterRepository interface {
	// Cluster operations
	Create(cluster *entities.NginxCluster) error
	FindByID(id string) (*entities.NginxCluster, error)
	FindByInfrastructureID(infraID string) (*entities.NginxCluster, error)
	Update(cluster *entities.NginxCluster) error
	Delete(id string) error
	ListAll() ([]entities.NginxCluster, error)

	// Node operations
	CreateNode(node *entities.NginxNode) error
	FindNodeByID(nodeID string) (*entities.NginxNode, error)
	ListNodes(clusterID string) ([]entities.NginxNode, error)
	UpdateNode(node *entities.NginxNode) error
	DeleteNode(id string) error
	FindMasterNode(clusterID string) (*entities.NginxNode, error)

	// Upstream operations
	CreateUpstream(upstream *entities.NginxClusterUpstream) error
	FindUpstreamByID(id string) (*entities.NginxClusterUpstream, error)
	ListUpstreams(clusterID string) ([]entities.NginxClusterUpstream, error)
	UpdateUpstream(upstream *entities.NginxClusterUpstream) error
	DeleteUpstream(id string) error

	// Upstream server operations
	CreateUpstreamServer(server *entities.NginxUpstreamServer) error
	ListUpstreamServers(upstreamID string) ([]entities.NginxUpstreamServer, error)
	UpdateUpstreamServer(server *entities.NginxUpstreamServer) error
	DeleteUpstreamServer(id string) error
	DeleteUpstreamServersByUpstreamID(upstreamID string) error

	// Server block operations
	CreateServerBlock(block *entities.NginxServerBlock) error
	FindServerBlockByID(id string) (*entities.NginxServerBlock, error)
	ListServerBlocks(clusterID string) ([]entities.NginxServerBlock, error)
	UpdateServerBlock(block *entities.NginxServerBlock) error
	DeleteServerBlock(id string) error

	// Location operations
	CreateLocation(loc *entities.NginxLocation) error
	FindLocationByID(id string) (*entities.NginxLocation, error)
	ListLocations(serverBlockID string) ([]entities.NginxLocation, error)
	UpdateLocation(loc *entities.NginxLocation) error
	DeleteLocation(id string) error
	DeleteLocationsByServerBlockID(serverBlockID string) error

	// Failover event operations
	CreateFailoverEvent(event *entities.NginxFailoverEvent) error
	ListFailoverEvents(clusterID string) ([]entities.NginxFailoverEvent, error)
}

type nginxClusterRepository struct {
	db *gorm.DB
}

// NewNginxClusterRepository creates a new nginx cluster repository
func NewNginxClusterRepository(db *gorm.DB) INginxClusterRepository {
	return &nginxClusterRepository{db: db}
}

// ================== Cluster Operations ==================

func (r *nginxClusterRepository) Create(cluster *entities.NginxCluster) error {
	return r.db.Create(cluster).Error
}

func (r *nginxClusterRepository) FindByID(id string) (*entities.NginxCluster, error) {
	var cluster entities.NginxCluster
	err := r.db.Preload("Infrastructure").First(&cluster, "id = ?", id).Error
	return &cluster, err
}

func (r *nginxClusterRepository) FindByInfrastructureID(infraID string) (*entities.NginxCluster, error) {
	var cluster entities.NginxCluster
	err := r.db.Preload("Infrastructure").First(&cluster, "infrastructure_id = ?", infraID).Error
	return &cluster, err
}

func (r *nginxClusterRepository) Update(cluster *entities.NginxCluster) error {
	return r.db.Save(cluster).Error
}

func (r *nginxClusterRepository) Delete(id string) error {
	return r.db.Delete(&entities.NginxCluster{}, "id = ?", id).Error
}

func (r *nginxClusterRepository) ListAll() ([]entities.NginxCluster, error) {
	var clusters []entities.NginxCluster
	err := r.db.Preload("Infrastructure").Find(&clusters).Error
	return clusters, err
}

// ================== Node Operations ==================

func (r *nginxClusterRepository) CreateNode(node *entities.NginxNode) error {
	return r.db.Create(node).Error
}

func (r *nginxClusterRepository) FindNodeByID(nodeID string) (*entities.NginxNode, error) {
	var node entities.NginxNode
	err := r.db.First(&node, "id = ?", nodeID).Error
	return &node, err
}

func (r *nginxClusterRepository) ListNodes(clusterID string) ([]entities.NginxNode, error) {
	var nodes []entities.NginxNode
	err := r.db.Order("priority DESC").Find(&nodes, "cluster_id = ?", clusterID).Error
	return nodes, err
}

func (r *nginxClusterRepository) UpdateNode(node *entities.NginxNode) error {
	return r.db.Save(node).Error
}

func (r *nginxClusterRepository) DeleteNode(id string) error {
	return r.db.Delete(&entities.NginxNode{}, "id = ?", id).Error
}

func (r *nginxClusterRepository) FindMasterNode(clusterID string) (*entities.NginxNode, error) {
	var node entities.NginxNode
	err := r.db.Where("cluster_id = ? AND role = ?", clusterID, "master").First(&node).Error
	return &node, err
}

// ================== Upstream Operations ==================

func (r *nginxClusterRepository) CreateUpstream(upstream *entities.NginxClusterUpstream) error {
	return r.db.Create(upstream).Error
}

func (r *nginxClusterRepository) FindUpstreamByID(id string) (*entities.NginxClusterUpstream, error) {
	var upstream entities.NginxClusterUpstream
	err := r.db.First(&upstream, "id = ?", id).Error
	return &upstream, err
}

func (r *nginxClusterRepository) ListUpstreams(clusterID string) ([]entities.NginxClusterUpstream, error) {
	var upstreams []entities.NginxClusterUpstream
	err := r.db.Find(&upstreams, "cluster_id = ?", clusterID).Error
	return upstreams, err
}

func (r *nginxClusterRepository) UpdateUpstream(upstream *entities.NginxClusterUpstream) error {
	return r.db.Save(upstream).Error
}

func (r *nginxClusterRepository) DeleteUpstream(id string) error {
	return r.db.Delete(&entities.NginxClusterUpstream{}, "id = ?", id).Error
}

// ================== Upstream Server Operations ==================

func (r *nginxClusterRepository) CreateUpstreamServer(server *entities.NginxUpstreamServer) error {
	return r.db.Create(server).Error
}

func (r *nginxClusterRepository) ListUpstreamServers(upstreamID string) ([]entities.NginxUpstreamServer, error) {
	var servers []entities.NginxUpstreamServer
	err := r.db.Find(&servers, "upstream_id = ?", upstreamID).Error
	return servers, err
}

func (r *nginxClusterRepository) UpdateUpstreamServer(server *entities.NginxUpstreamServer) error {
	return r.db.Save(server).Error
}

func (r *nginxClusterRepository) DeleteUpstreamServer(id string) error {
	return r.db.Delete(&entities.NginxUpstreamServer{}, "id = ?", id).Error
}

func (r *nginxClusterRepository) DeleteUpstreamServersByUpstreamID(upstreamID string) error {
	return r.db.Delete(&entities.NginxUpstreamServer{}, "upstream_id = ?", upstreamID).Error
}

// ================== Server Block Operations ==================

func (r *nginxClusterRepository) CreateServerBlock(block *entities.NginxServerBlock) error {
	return r.db.Create(block).Error
}

func (r *nginxClusterRepository) FindServerBlockByID(id string) (*entities.NginxServerBlock, error) {
	var block entities.NginxServerBlock
	err := r.db.First(&block, "id = ?", id).Error
	return &block, err
}

func (r *nginxClusterRepository) ListServerBlocks(clusterID string) ([]entities.NginxServerBlock, error) {
	var blocks []entities.NginxServerBlock
	err := r.db.Find(&blocks, "cluster_id = ?", clusterID).Error
	return blocks, err
}

func (r *nginxClusterRepository) UpdateServerBlock(block *entities.NginxServerBlock) error {
	return r.db.Save(block).Error
}

func (r *nginxClusterRepository) DeleteServerBlock(id string) error {
	return r.db.Delete(&entities.NginxServerBlock{}, "id = ?", id).Error
}

// ================== Location Operations ==================

func (r *nginxClusterRepository) CreateLocation(loc *entities.NginxLocation) error {
	return r.db.Create(loc).Error
}

func (r *nginxClusterRepository) FindLocationByID(id string) (*entities.NginxLocation, error) {
	var loc entities.NginxLocation
	err := r.db.First(&loc, "id = ?", id).Error
	return &loc, err
}

func (r *nginxClusterRepository) ListLocations(serverBlockID string) ([]entities.NginxLocation, error) {
	var locs []entities.NginxLocation
	err := r.db.Find(&locs, "server_block_id = ?", serverBlockID).Error
	return locs, err
}

func (r *nginxClusterRepository) UpdateLocation(loc *entities.NginxLocation) error {
	return r.db.Save(loc).Error
}

func (r *nginxClusterRepository) DeleteLocation(id string) error {
	return r.db.Delete(&entities.NginxLocation{}, "id = ?", id).Error
}

func (r *nginxClusterRepository) DeleteLocationsByServerBlockID(serverBlockID string) error {
	return r.db.Delete(&entities.NginxLocation{}, "server_block_id = ?", serverBlockID).Error
}

// ================== Failover Event Operations ==================

func (r *nginxClusterRepository) CreateFailoverEvent(event *entities.NginxFailoverEvent) error {
	return r.db.Create(event).Error
}

func (r *nginxClusterRepository) ListFailoverEvents(clusterID string) ([]entities.NginxFailoverEvent, error) {
	var events []entities.NginxFailoverEvent
	err := r.db.Order("occurred_at DESC").Find(&events, "cluster_id = ?", clusterID).Error
	return events, err
}

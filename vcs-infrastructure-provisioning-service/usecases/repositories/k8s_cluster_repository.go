package repositories

import (
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/entities"
	"gorm.io/gorm"
)

// IK8sClusterRepository defines the interface for K8s cluster repository
type IK8sClusterRepository interface {
	// Cluster operations
	Create(cluster *entities.K8sCluster) error
	FindByID(id string) (*entities.K8sCluster, error)
	FindByName(name string) (*entities.K8sCluster, error)
	FindByInfrastructureID(infraID string) (*entities.K8sCluster, error)
	Update(cluster *entities.K8sCluster) error
	Delete(id string) error
	List() ([]entities.K8sCluster, error)
	
	// Node operations
	CreateNode(node *entities.K8sNode) error
	FindNodeByID(nodeID string) (*entities.K8sNode, error)
	ListNodes(clusterID string) ([]entities.K8sNode, error)
	UpdateNode(node *entities.K8sNode) error
	DeleteNode(nodeID string) error
	DeleteNodesByClusterID(clusterID string) error
}

// k8sClusterRepository implements IK8sClusterRepository
type k8sClusterRepository struct {
	db *gorm.DB
}

// NewK8sClusterRepository creates a new K8s cluster repository
func NewK8sClusterRepository(db *gorm.DB) IK8sClusterRepository {
	return &k8sClusterRepository{db: db}
}

// Create creates a new K8s cluster
func (r *k8sClusterRepository) Create(cluster *entities.K8sCluster) error {
	return r.db.Create(cluster).Error
}

// FindByID finds a cluster by ID
func (r *k8sClusterRepository) FindByID(id string) (*entities.K8sCluster, error) {
	var cluster entities.K8sCluster
	err := r.db.Preload("Nodes").Where("id = ?", id).First(&cluster).Error
	return &cluster, err
}

// FindByName finds a cluster by name
func (r *k8sClusterRepository) FindByName(name string) (*entities.K8sCluster, error) {
	var cluster entities.K8sCluster
	err := r.db.Preload("Nodes").Where("cluster_name = ?", name).First(&cluster).Error
	return &cluster, err
}

// FindByInfrastructureID finds a cluster by infrastructure ID
func (r *k8sClusterRepository) FindByInfrastructureID(infraID string) (*entities.K8sCluster, error) {
	var cluster entities.K8sCluster
	err := r.db.Preload("Nodes").Where("infrastructure_id = ?", infraID).First(&cluster).Error
	return &cluster, err
}

// Update updates a cluster
func (r *k8sClusterRepository) Update(cluster *entities.K8sCluster) error {
	return r.db.Save(cluster).Error
}

// Delete deletes a cluster
func (r *k8sClusterRepository) Delete(id string) error {
	return r.db.Delete(&entities.K8sCluster{}, "id = ?", id).Error
}

// List lists all clusters
func (r *k8sClusterRepository) List() ([]entities.K8sCluster, error) {
	var clusters []entities.K8sCluster
	err := r.db.Preload("Nodes").Find(&clusters).Error
	return clusters, err
}

// CreateNode creates a new node
func (r *k8sClusterRepository) CreateNode(node *entities.K8sNode) error {
	return r.db.Create(node).Error
}

// FindNodeByID finds a node by ID
func (r *k8sClusterRepository) FindNodeByID(nodeID string) (*entities.K8sNode, error) {
	var node entities.K8sNode
	err := r.db.Where("id = ?", nodeID).First(&node).Error
	return &node, err
}

// ListNodes lists all nodes in a cluster
func (r *k8sClusterRepository) ListNodes(clusterID string) ([]entities.K8sNode, error) {
	var nodes []entities.K8sNode
	err := r.db.Where("cluster_id = ?", clusterID).Order("role DESC, name ASC").Find(&nodes).Error
	return nodes, err
}

// UpdateNode updates a node
func (r *k8sClusterRepository) UpdateNode(node *entities.K8sNode) error {
	return r.db.Save(node).Error
}

// DeleteNode deletes a node
func (r *k8sClusterRepository) DeleteNode(nodeID string) error {
	return r.db.Delete(&entities.K8sNode{}, "id = ?", nodeID).Error
}

// DeleteNodesByClusterID deletes all nodes in a cluster
func (r *k8sClusterRepository) DeleteNodesByClusterID(clusterID string) error {
	return r.db.Where("cluster_id = ?", clusterID).Delete(&entities.K8sNode{}).Error
}


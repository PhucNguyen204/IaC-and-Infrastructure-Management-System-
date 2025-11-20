package repositories

import (
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/entities"
	"gorm.io/gorm"
)

type IPostgreSQLClusterRepository interface {
	Create(cluster *entities.PostgreSQLCluster) error
	FindByID(id string) (*entities.PostgreSQLCluster, error)
	FindByInfrastructureID(infraID string) (*entities.PostgreSQLCluster, error)
	Update(cluster *entities.PostgreSQLCluster) error
	Delete(id string) error
	ListNodes(clusterID string) ([]entities.ClusterNode, error)
	CreateNode(node *entities.ClusterNode) error
	UpdateNode(node *entities.ClusterNode) error
	DeleteNode(id string) error
	CreateEtcdNode(node *entities.EtcdNode) error
	ListEtcdNodes(clusterID string) ([]entities.EtcdNode, error)
	DeleteEtcdNode(id string) error
}

type postgreSQLClusterRepository struct {
	db *gorm.DB
}

func NewPostgreSQLClusterRepository(db *gorm.DB) IPostgreSQLClusterRepository {
	return &postgreSQLClusterRepository{db: db}
}

func (r *postgreSQLClusterRepository) Create(cluster *entities.PostgreSQLCluster) error {
	return r.db.Create(cluster).Error
}

func (r *postgreSQLClusterRepository) FindByID(id string) (*entities.PostgreSQLCluster, error) {
	var cluster entities.PostgreSQLCluster
	err := r.db.Preload("Infrastructure").First(&cluster, "id = ?", id).Error
	return &cluster, err
}

func (r *postgreSQLClusterRepository) FindByInfrastructureID(infraID string) (*entities.PostgreSQLCluster, error) {
	var cluster entities.PostgreSQLCluster
	err := r.db.Preload("Infrastructure").First(&cluster, "infrastructure_id = ?", infraID).Error
	return &cluster, err
}

func (r *postgreSQLClusterRepository) Update(cluster *entities.PostgreSQLCluster) error {
	return r.db.Save(cluster).Error
}

func (r *postgreSQLClusterRepository) Delete(id string) error {
	return r.db.Delete(&entities.PostgreSQLCluster{}, "id = ?", id).Error
}

func (r *postgreSQLClusterRepository) ListNodes(clusterID string) ([]entities.ClusterNode, error) {
	var nodes []entities.ClusterNode
	err := r.db.Find(&nodes, "cluster_id = ?", clusterID).Error
	return nodes, err
}

func (r *postgreSQLClusterRepository) CreateNode(node *entities.ClusterNode) error {
	return r.db.Create(node).Error
}

func (r *postgreSQLClusterRepository) UpdateNode(node *entities.ClusterNode) error {
	return r.db.Save(node).Error
}

func (r *postgreSQLClusterRepository) DeleteNode(id string) error {
	return r.db.Delete(&entities.ClusterNode{}, "id = ?", id).Error
}

func (r *postgreSQLClusterRepository) CreateEtcdNode(node *entities.EtcdNode) error {
	return r.db.Create(node).Error
}

func (r *postgreSQLClusterRepository) ListEtcdNodes(clusterID string) ([]entities.EtcdNode, error) {
	var nodes []entities.EtcdNode
	err := r.db.Find(&nodes, "cluster_id = ?", clusterID).Error
	return nodes, err
}

func (r *postgreSQLClusterRepository) DeleteEtcdNode(id string) error {
	return r.db.Delete(&entities.EtcdNode{}, "id = ?", id).Error
}


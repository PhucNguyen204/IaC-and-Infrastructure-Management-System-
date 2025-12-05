package repositories

import (
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/entities"
	"gorm.io/gorm"
)

type ClickHouseRepository struct {
	db *gorm.DB
}

func NewClickHouseRepository(db *gorm.DB) *ClickHouseRepository {
	return &ClickHouseRepository{db: db}
}

// Cluster operations
func (r *ClickHouseRepository) CreateCluster(cluster *entities.ClickHouseCluster) error {
	return r.db.Create(cluster).Error
}

func (r *ClickHouseRepository) GetClusterByID(id string) (*entities.ClickHouseCluster, error) {
	var cluster entities.ClickHouseCluster
	err := r.db.Preload("Infrastructure").First(&cluster, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &cluster, nil
}

func (r *ClickHouseRepository) GetClusterByInfraID(infraID string) (*entities.ClickHouseCluster, error) {
	var cluster entities.ClickHouseCluster
	err := r.db.Preload("Infrastructure").First(&cluster, "infrastructure_id = ?", infraID).Error
	if err != nil {
		return nil, err
	}
	return &cluster, nil
}

func (r *ClickHouseRepository) GetClusterByName(name string) (*entities.ClickHouseCluster, error) {
	var cluster entities.ClickHouseCluster
	err := r.db.Preload("Infrastructure").First(&cluster, "cluster_name = ?", name).Error
	if err != nil {
		return nil, err
	}
	return &cluster, nil
}

func (r *ClickHouseRepository) UpdateCluster(cluster *entities.ClickHouseCluster) error {
	return r.db.Save(cluster).Error
}

func (r *ClickHouseRepository) DeleteCluster(id string) error {
	return r.db.Delete(&entities.ClickHouseCluster{}, "id = ?", id).Error
}

func (r *ClickHouseRepository) ListClusters() ([]entities.ClickHouseCluster, error) {
	var clusters []entities.ClickHouseCluster
	err := r.db.Preload("Infrastructure").Find(&clusters).Error
	return clusters, err
}

// Node operations
func (r *ClickHouseRepository) CreateNode(node *entities.ClickHouseNode) error {
	return r.db.Create(node).Error
}

func (r *ClickHouseRepository) GetNodeByID(id string) (*entities.ClickHouseNode, error) {
	var node entities.ClickHouseNode
	err := r.db.First(&node, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &node, nil
}

func (r *ClickHouseRepository) GetNodesByClusterID(clusterID string) ([]entities.ClickHouseNode, error) {
	var nodes []entities.ClickHouseNode
	err := r.db.Where("cluster_id = ?", clusterID).Find(&nodes).Error
	return nodes, err
}

func (r *ClickHouseRepository) UpdateNode(node *entities.ClickHouseNode) error {
	return r.db.Save(node).Error
}

func (r *ClickHouseRepository) DeleteNode(id string) error {
	return r.db.Delete(&entities.ClickHouseNode{}, "id = ?", id).Error
}

func (r *ClickHouseRepository) DeleteNodesByClusterID(clusterID string) error {
	return r.db.Delete(&entities.ClickHouseNode{}, "cluster_id = ?", clusterID).Error
}

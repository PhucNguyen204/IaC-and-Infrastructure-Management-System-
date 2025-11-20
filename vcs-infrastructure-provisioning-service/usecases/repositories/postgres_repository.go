package repositories

import (
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/entities"
	"gorm.io/gorm"
)

type IPostgreSQLRepository interface {
	Create(instance *entities.PostgreSQLInstance) error
	FindByID(id string) (*entities.PostgreSQLInstance, error)
	FindByInfrastructureID(infraID string) (*entities.PostgreSQLInstance, error)
	Update(instance *entities.PostgreSQLInstance) error
	Delete(id string) error
}

type postgreSQLRepository struct {
	db *gorm.DB
}

func NewPostgreSQLRepository(db *gorm.DB) IPostgreSQLRepository {
	return &postgreSQLRepository{db: db}
}

func (r *postgreSQLRepository) Create(instance *entities.PostgreSQLInstance) error {
	return r.db.Create(instance).Error
}

func (r *postgreSQLRepository) FindByID(id string) (*entities.PostgreSQLInstance, error) {
	var instance entities.PostgreSQLInstance
	if err := r.db.Preload("Infrastructure").Where("id = ?", id).First(&instance).Error; err != nil {
		return nil, err
	}
	return &instance, nil
}

func (r *postgreSQLRepository) FindByInfrastructureID(infraID string) (*entities.PostgreSQLInstance, error) {
	var instance entities.PostgreSQLInstance
	if err := r.db.Preload("Infrastructure").Where("infrastructure_id = ?", infraID).First(&instance).Error; err != nil {
		return nil, err
	}
	return &instance, nil
}

func (r *postgreSQLRepository) Update(instance *entities.PostgreSQLInstance) error {
	return r.db.Save(instance).Error
}

func (r *postgreSQLRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&entities.PostgreSQLInstance{}).Error
}


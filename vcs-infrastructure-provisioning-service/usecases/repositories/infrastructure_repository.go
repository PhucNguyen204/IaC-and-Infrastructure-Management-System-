package repositories

import (
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/entities"
	"gorm.io/gorm"
)

type IInfrastructureRepository interface {
	Create(infra *entities.Infrastructure) error
	FindByID(id string) (*entities.Infrastructure, error)
	FindByUserID(userID string) ([]*entities.Infrastructure, error)
	Update(infra *entities.Infrastructure) error
	Delete(id string) error
}

type infrastructureRepository struct {
	db *gorm.DB
}

func NewInfrastructureRepository(db *gorm.DB) IInfrastructureRepository {
	return &infrastructureRepository{db: db}
}

func (r *infrastructureRepository) Create(infra *entities.Infrastructure) error {
	return r.db.Create(infra).Error
}

func (r *infrastructureRepository) FindByID(id string) (*entities.Infrastructure, error) {
	var infra entities.Infrastructure
	if err := r.db.Where("id = ?", id).First(&infra).Error; err != nil {
		return nil, err
	}
	return &infra, nil
}

func (r *infrastructureRepository) FindByUserID(userID string) ([]*entities.Infrastructure, error) {
	var infras []*entities.Infrastructure
	if err := r.db.Where("user_id = ?", userID).Find(&infras).Error; err != nil {
		return nil, err
	}
	return infras, nil
}

func (r *infrastructureRepository) Update(infra *entities.Infrastructure) error {
	return r.db.Save(infra).Error
}

func (r *infrastructureRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&entities.Infrastructure{}).Error
}


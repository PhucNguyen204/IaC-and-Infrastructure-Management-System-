package repositories

import (
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/entities"
	"gorm.io/gorm"
)

type IDinDRepository interface {
	Create(env *entities.DinDEnvironment) error
	Update(env *entities.DinDEnvironment) error
	Delete(id string) error
	FindByID(id string) (*entities.DinDEnvironment, error)
	FindByInfrastructureID(infraID string) (*entities.DinDEnvironment, error)
	FindByUserID(userID string) ([]entities.DinDEnvironment, error)
	FindByContainerID(containerID string) (*entities.DinDEnvironment, error)
	FindExpired() ([]entities.DinDEnvironment, error)
	CreateCommandHistory(history *entities.DinDCommandHistory) error
	GetCommandHistory(environmentID string, limit int) ([]entities.DinDCommandHistory, error)
}

type dinDRepository struct {
	db *gorm.DB
}

func NewDinDRepository(db *gorm.DB) IDinDRepository {
	return &dinDRepository{db: db}
}

func (r *dinDRepository) Create(env *entities.DinDEnvironment) error {
	return r.db.Create(env).Error
}

func (r *dinDRepository) Update(env *entities.DinDEnvironment) error {
	return r.db.Save(env).Error
}

func (r *dinDRepository) Delete(id string) error {
	// Delete command history first
	r.db.Where("environment_id = ?", id).Delete(&entities.DinDCommandHistory{})
	return r.db.Delete(&entities.DinDEnvironment{}, "id = ?", id).Error
}

func (r *dinDRepository) FindByID(id string) (*entities.DinDEnvironment, error) {
	var env entities.DinDEnvironment
	err := r.db.Where("id = ?", id).First(&env).Error
	if err != nil {
		return nil, err
	}
	return &env, nil
}

func (r *dinDRepository) FindByInfrastructureID(infraID string) (*entities.DinDEnvironment, error) {
	var env entities.DinDEnvironment
	err := r.db.Where("infrastructure_id = ?", infraID).First(&env).Error
	if err != nil {
		return nil, err
	}
	return &env, nil
}

func (r *dinDRepository) FindByUserID(userID string) ([]entities.DinDEnvironment, error) {
	var envs []entities.DinDEnvironment
	err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&envs).Error
	return envs, err
}

func (r *dinDRepository) FindByContainerID(containerID string) (*entities.DinDEnvironment, error) {
	var env entities.DinDEnvironment
	err := r.db.Where("container_id = ?", containerID).First(&env).Error
	if err != nil {
		return nil, err
	}
	return &env, nil
}

func (r *dinDRepository) FindExpired() ([]entities.DinDEnvironment, error) {
	var envs []entities.DinDEnvironment
	err := r.db.Where("auto_cleanup = ? AND expires_at < NOW() AND status != ?", true, "deleted").Find(&envs).Error
	return envs, err
}

func (r *dinDRepository) CreateCommandHistory(history *entities.DinDCommandHistory) error {
	return r.db.Create(history).Error
}

func (r *dinDRepository) GetCommandHistory(environmentID string, limit int) ([]entities.DinDCommandHistory, error) {
	var history []entities.DinDCommandHistory
	query := r.db.Where("environment_id = ?", environmentID).Order("executed_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&history).Error
	return history, err
}


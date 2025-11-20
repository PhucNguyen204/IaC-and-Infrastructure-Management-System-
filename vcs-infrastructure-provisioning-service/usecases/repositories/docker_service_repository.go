package repositories

import (
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/entities"
	"gorm.io/gorm"
)

type IDockerServiceRepository interface {
	Create(service *entities.DockerService) error
	FindByID(id string) (*entities.DockerService, error)
	FindByInfrastructureID(infraID string) (*entities.DockerService, error)
	Update(service *entities.DockerService) error
	Delete(id string) error
	CreateEnvVar(envVar *entities.DockerEnvVar) error
	UpdateEnvVars(serviceID string, envVars []entities.DockerEnvVar) error
	DeleteEnvVarsByServiceID(serviceID string) error
	CreatePort(port *entities.DockerPort) error
	CreateNetwork(network *entities.DockerNetwork) error
	CreateHealthCheck(healthCheck *entities.DockerHealthCheck) error
	UpdateHealthCheck(healthCheck *entities.DockerHealthCheck) error
	FindHealthCheckByServiceID(serviceID string) (*entities.DockerHealthCheck, error)
}

type dockerServiceRepository struct {
	db *gorm.DB
}

func NewDockerServiceRepository(db *gorm.DB) IDockerServiceRepository {
	return &dockerServiceRepository{db: db}
}

func (r *dockerServiceRepository) Create(service *entities.DockerService) error {
	return r.db.Create(service).Error
}

func (r *dockerServiceRepository) FindByID(id string) (*entities.DockerService, error) {
	var service entities.DockerService
	err := r.db.Preload("EnvVars").Preload("Ports").Preload("Networks").Preload("HealthCheck").First(&service, "id = ?", id).Error
	return &service, err
}

func (r *dockerServiceRepository) FindByInfrastructureID(infraID string) (*entities.DockerService, error) {
	var service entities.DockerService
	err := r.db.Preload("EnvVars").Preload("Ports").Preload("Networks").Preload("HealthCheck").First(&service, "infrastructure_id = ?", infraID).Error
	return &service, err
}

func (r *dockerServiceRepository) Update(service *entities.DockerService) error {
	return r.db.Save(service).Error
}

func (r *dockerServiceRepository) Delete(id string) error {
	return r.db.Delete(&entities.DockerService{}, "id = ?", id).Error
}

func (r *dockerServiceRepository) CreateEnvVar(envVar *entities.DockerEnvVar) error {
	return r.db.Create(envVar).Error
}

func (r *dockerServiceRepository) UpdateEnvVars(serviceID string, envVars []entities.DockerEnvVar) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("service_id = ?", serviceID).Delete(&entities.DockerEnvVar{}).Error; err != nil {
			return err
		}
		if len(envVars) > 0 {
			if err := tx.Create(&envVars).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *dockerServiceRepository) DeleteEnvVarsByServiceID(serviceID string) error {
	return r.db.Where("service_id = ?", serviceID).Delete(&entities.DockerEnvVar{}).Error
}

func (r *dockerServiceRepository) CreatePort(port *entities.DockerPort) error {
	return r.db.Create(port).Error
}

func (r *dockerServiceRepository) CreateNetwork(network *entities.DockerNetwork) error {
	return r.db.Create(network).Error
}

func (r *dockerServiceRepository) CreateHealthCheck(healthCheck *entities.DockerHealthCheck) error {
	return r.db.Create(healthCheck).Error
}

func (r *dockerServiceRepository) UpdateHealthCheck(healthCheck *entities.DockerHealthCheck) error {
	return r.db.Save(healthCheck).Error
}

func (r *dockerServiceRepository) FindHealthCheckByServiceID(serviceID string) (*entities.DockerHealthCheck, error) {
	var healthCheck entities.DockerHealthCheck
	err := r.db.First(&healthCheck, "service_id = ?", serviceID).Error
	return &healthCheck, err
}

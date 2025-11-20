package repositories

import (
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/entities"
	"gorm.io/gorm"
)

type IPostgresDatabaseRepository interface {
	Create(db *entities.PostgresDatabase) error
	FindByID(id string) (*entities.PostgresDatabase, error)
	FindByInstanceID(instanceID string) ([]*entities.PostgresDatabase, error)
	FindByDBName(instanceID, dbName string) (*entities.PostgresDatabase, error)
	FindByProjectID(projectID string) ([]*entities.PostgresDatabase, error)
	Update(db *entities.PostgresDatabase) error
	Delete(id string) error
	CreateBackup(backup *entities.PostgresBackup) error
	FindBackupByID(id string) (*entities.PostgresBackup, error)
	ListBackups(databaseID string) ([]*entities.PostgresBackup, error)
	UpdateBackup(backup *entities.PostgresBackup) error
}

type postgresDatabaseRepository struct {
	db *gorm.DB
}

func NewPostgresDatabaseRepository(db *gorm.DB) IPostgresDatabaseRepository {
	return &postgresDatabaseRepository{db: db}
}

func (r *postgresDatabaseRepository) Create(database *entities.PostgresDatabase) error {
	return r.db.Create(database).Error
}

func (r *postgresDatabaseRepository) FindByID(id string) (*entities.PostgresDatabase, error) {
	var database entities.PostgresDatabase
	if err := r.db.Preload("Instance").Where("id = ?", id).First(&database).Error; err != nil {
		return nil, err
	}
	return &database, nil
}

func (r *postgresDatabaseRepository) FindByInstanceID(instanceID string) ([]*entities.PostgresDatabase, error) {
	var databases []*entities.PostgresDatabase
	if err := r.db.Where("instance_id = ?", instanceID).Find(&databases).Error; err != nil {
		return nil, err
	}
	return databases, nil
}

func (r *postgresDatabaseRepository) FindByDBName(instanceID, dbName string) (*entities.PostgresDatabase, error) {
	var database entities.PostgresDatabase
	if err := r.db.Where("instance_id = ? AND db_name = ?", instanceID, dbName).First(&database).Error; err != nil {
		return nil, err
	}
	return &database, nil
}

func (r *postgresDatabaseRepository) FindByProjectID(projectID string) ([]*entities.PostgresDatabase, error) {
	var databases []*entities.PostgresDatabase
	if err := r.db.Preload("Instance").Where("project_id = ?", projectID).Find(&databases).Error; err != nil {
		return nil, err
	}
	return databases, nil
}

func (r *postgresDatabaseRepository) Update(database *entities.PostgresDatabase) error {
	return r.db.Save(database).Error
}

func (r *postgresDatabaseRepository) Delete(id string) error {
	return r.db.Delete(&entities.PostgresDatabase{}, "id = ?", id).Error
}

func (r *postgresDatabaseRepository) CreateBackup(backup *entities.PostgresBackup) error {
	return r.db.Create(backup).Error
}

func (r *postgresDatabaseRepository) FindBackupByID(id string) (*entities.PostgresBackup, error) {
	var backup entities.PostgresBackup
	if err := r.db.Preload("Database").Where("id = ?", id).First(&backup).Error; err != nil {
		return nil, err
	}
	return &backup, nil
}

func (r *postgresDatabaseRepository) ListBackups(databaseID string) ([]*entities.PostgresBackup, error) {
	var backups []*entities.PostgresBackup
	if err := r.db.Where("database_id = ?", databaseID).Order("created_at DESC").Find(&backups).Error; err != nil {
		return nil, err
	}
	return backups, nil
}

func (r *postgresDatabaseRepository) UpdateBackup(backup *entities.PostgresBackup) error {
	return r.db.Save(backup).Error
}

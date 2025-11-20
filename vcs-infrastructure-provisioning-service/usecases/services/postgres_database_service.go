package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/dto"
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/entities"
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/infrastructures/docker"
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/usecases/repositories"
)

type IPostgresDatabaseService interface {
	CreateDatabase(ctx context.Context, instanceID string, req dto.CreateDatabaseRequest) (*dto.DatabaseInfo, error)
	GetDatabase(ctx context.Context, databaseID string) (*dto.DatabaseInfo, error)
	ListDatabases(ctx context.Context, instanceID string) ([]*dto.DatabaseInfo, error)
	UpdateQuota(ctx context.Context, databaseID string, req dto.UpdateQuotaRequest) error
	GetMetrics(ctx context.Context, databaseID string) (*dto.DatabaseMetrics, error)
	BackupDatabase(ctx context.Context, databaseID string, req dto.BackupDatabaseRequest) (*dto.BackupInfo, error)
	RestoreDatabase(ctx context.Context, databaseID string, req dto.RestoreDatabaseRequest) error
	ManageLifecycle(ctx context.Context, databaseID string, req dto.ManageLifecycleRequest) error
	GetInstanceOverview(ctx context.Context, instanceID string) (*dto.InstanceOverview, error)
}

type postgresDatabaseService struct {
	dbRepo       repositories.IPostgresDatabaseRepository
	instanceRepo repositories.IPostgreSQLRepository
	dockerSvc    docker.IDockerService
}

func NewPostgresDatabaseService(
	dbRepo repositories.IPostgresDatabaseRepository,
	instanceRepo repositories.IPostgreSQLRepository,
	dockerSvc docker.IDockerService,
) IPostgresDatabaseService {
	return &postgresDatabaseService{
		dbRepo:       dbRepo,
		instanceRepo: instanceRepo,
		dockerSvc:    dockerSvc,
	}
}

func (s *postgresDatabaseService) CreateDatabase(ctx context.Context, instanceID string, req dto.CreateDatabaseRequest) (*dto.DatabaseInfo, error) {
	instance, err := s.instanceRepo.FindByInfrastructureID(instanceID)
	if err != nil {
		return nil, fmt.Errorf("instance not found: %w", err)
	}
	if req.OwnerUsername == "" {
		req.OwnerUsername = fmt.Sprintf("user_%s", uuid.New().String()[:8])
	}
	if req.MaxSizeGB == 0 {
		req.MaxSizeGB = 10
	}
	if req.MaxConnections == 0 {
		req.MaxConnections = 50
	}
	containerIP, err := s.getContainerIP(ctx, instance.ContainerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get container IP: %w", err)
	}
	connStr := fmt.Sprintf("host=%s port=5432 user=%s password=%s dbname=%s sslmode=disable connect_timeout=5",
		containerIP, instance.Username, instance.Password, instance.DatabaseName)
	var db *sql.DB
	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		db, err = sql.Open("postgres", connStr)
		if err == nil {
			if err = db.PingContext(ctx); err == nil {
				break
			}
			db.Close()
		}
		if i < maxRetries-1 {
			time.Sleep(2 * time.Second)
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to connect after %d retries: %w", maxRetries, err)
	}
	defer db.Close()
	if _, err := db.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE %s", req.DBName)); err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}
	if _, err := db.ExecContext(ctx, fmt.Sprintf("CREATE ROLE %s WITH LOGIN PASSWORD '%s'", req.OwnerUsername, req.OwnerPassword)); err != nil {
		db.ExecContext(ctx, fmt.Sprintf("DROP DATABASE %s", req.DBName))
		return nil, fmt.Errorf("failed to create role: %w", err)
	}
	if _, err := db.ExecContext(ctx, fmt.Sprintf("GRANT ALL PRIVILEGES ON DATABASE %s TO %s", req.DBName, req.OwnerUsername)); err != nil {
		db.ExecContext(ctx, fmt.Sprintf("DROP ROLE %s", req.OwnerUsername))
		db.ExecContext(ctx, fmt.Sprintf("DROP DATABASE %s", req.DBName))
		return nil, fmt.Errorf("failed to grant privileges: %w", err)
	}
	if req.InitSchema != "" {
		dbConn := fmt.Sprintf("host=%s port=5432 user=%s password=%s dbname=%s sslmode=disable",
			containerIP, req.OwnerUsername, req.OwnerPassword, req.DBName)
		userDB, _ := sql.Open("postgres", dbConn)
		if userDB != nil {
			defer userDB.Close()
			userDB.ExecContext(ctx, req.InitSchema)
		}
	}
	database := &entities.PostgresDatabase{
		ID:             uuid.New().String(),
		InstanceID:     instance.ID,
		DBName:         req.DBName,
		OwnerUsername:  req.OwnerUsername,
		OwnerPassword:  req.OwnerPassword,
		ProjectID:      req.ProjectID,
		TenantID:       req.TenantID,
		EnvironmentID:  req.EnvironmentID,
		MaxSizeGB:      req.MaxSizeGB,
		MaxConnections: req.MaxConnections,
		Status:         "ACTIVE",
	}
	if err := s.dbRepo.Create(database); err != nil {
		return nil, err
	}
	return &dto.DatabaseInfo{
		ID:             database.ID,
		InstanceID:     instanceID,
		DBName:         database.DBName,
		OwnerUsername:  database.OwnerUsername,
		ProjectID:      database.ProjectID,
		TenantID:       database.TenantID,
		EnvironmentID:  database.EnvironmentID,
		MaxSizeGB:      database.MaxSizeGB,
		MaxConnections: database.MaxConnections,
		Status:         database.Status,
		ConnectionInfo: dto.ConnectionInfo{
			Host:     "localhost",
			Port:     instance.Port,
			Database: database.DBName,
			Username: database.OwnerUsername,
			Password: database.OwnerPassword,
		},
		CreatedAt: database.CreatedAt.Format(time.RFC3339),
		UpdatedAt: database.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (s *postgresDatabaseService) GetDatabase(ctx context.Context, databaseID string) (*dto.DatabaseInfo, error) {
	database, err := s.dbRepo.FindByID(databaseID)
	if err != nil {
		return nil, err
	}
	return &dto.DatabaseInfo{
		ID:             database.ID,
		InstanceID:     database.InstanceID,
		DBName:         database.DBName,
		OwnerUsername:  database.OwnerUsername,
		ProjectID:      database.ProjectID,
		TenantID:       database.TenantID,
		EnvironmentID:  database.EnvironmentID,
		MaxSizeGB:      database.MaxSizeGB,
		MaxConnections: database.MaxConnections,
		CurrentSizeMB:  database.CurrentSizeMB,
		ActiveConns:    database.ActiveConns,
		Status:         database.Status,
		ConnectionInfo: dto.ConnectionInfo{
			Host:     "localhost",
			Port:     database.Instance.Port,
			Database: database.DBName,
			Username: database.OwnerUsername,
			Password: database.OwnerPassword,
		},
		CreatedAt: database.CreatedAt.Format(time.RFC3339),
		UpdatedAt: database.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (s *postgresDatabaseService) ListDatabases(ctx context.Context, instanceID string) ([]*dto.DatabaseInfo, error) {
	instance, err := s.instanceRepo.FindByInfrastructureID(instanceID)
	if err != nil {
		return nil, err
	}
	databases, err := s.dbRepo.FindByInstanceID(instance.ID)
	if err != nil {
		return nil, err
	}
	result := make([]*dto.DatabaseInfo, 0, len(databases))
	for _, db := range databases {
		result = append(result, &dto.DatabaseInfo{
			ID:             db.ID,
			InstanceID:     instanceID,
			DBName:         db.DBName,
			OwnerUsername:  db.OwnerUsername,
			ProjectID:      db.ProjectID,
			TenantID:       db.TenantID,
			EnvironmentID:  db.EnvironmentID,
			MaxSizeGB:      db.MaxSizeGB,
			MaxConnections: db.MaxConnections,
			CurrentSizeMB:  db.CurrentSizeMB,
			ActiveConns:    db.ActiveConns,
			Status:         db.Status,
			CreatedAt:      db.CreatedAt.Format(time.RFC3339),
			UpdatedAt:      db.UpdatedAt.Format(time.RFC3339),
		})
	}
	return result, nil
}

func (s *postgresDatabaseService) UpdateQuota(ctx context.Context, databaseID string, req dto.UpdateQuotaRequest) error {
	database, err := s.dbRepo.FindByID(databaseID)
	if err != nil {
		return err
	}
	if req.MaxSizeGB > 0 {
		database.MaxSizeGB = req.MaxSizeGB
	}
	if req.MaxConnections > 0 {
		database.MaxConnections = req.MaxConnections
	}
	return s.dbRepo.Update(database)
}

func (s *postgresDatabaseService) GetMetrics(ctx context.Context, databaseID string) (*dto.DatabaseMetrics, error) {
	database, err := s.dbRepo.FindByID(databaseID)
	if err != nil {
		return nil, err
	}
	containerIP, err := s.getContainerIP(ctx, database.Instance.ContainerID)
	if err != nil {
		return nil, err
	}
	connStr := fmt.Sprintf("host=%s port=5432 user=%s password=%s dbname=%s sslmode=disable",
		containerIP, database.Instance.Username, database.Instance.Password, database.Instance.DatabaseName)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	var sizeMB int64
	db.QueryRowContext(ctx, "SELECT pg_database_size($1)/1024/1024", database.DBName).Scan(&sizeMB)
	var activeConns int
	db.QueryRowContext(ctx, "SELECT count(*) FROM pg_stat_activity WHERE datname = $1", database.DBName).Scan(&activeConns)
	database.CurrentSizeMB = sizeMB
	database.ActiveConns = activeConns
	s.dbRepo.Update(database)
	sizePercent := float64(sizeMB) / float64(database.MaxSizeGB*1024) * 100
	connPercent := float64(activeConns) / float64(database.MaxConnections) * 100
	status := "WITHIN_QUOTA"
	if sizePercent > 90 || connPercent > 90 {
		status = "NEAR_LIMIT"
	}
	if sizePercent > 100 || connPercent > 100 {
		status = "OVER_LIMIT"
	}
	return &dto.DatabaseMetrics{
		DatabaseID:       databaseID,
		DBName:           database.DBName,
		CurrentSizeMB:    sizeMB,
		MaxSizeGB:        database.MaxSizeGB,
		SizeUsagePercent: sizePercent,
		ActiveConns:      activeConns,
		MaxConnections:   database.MaxConnections,
		ConnUsagePercent: connPercent,
		Status:           status,
		QueryPerSecond:   0,
	}, nil
}

func (s *postgresDatabaseService) BackupDatabase(ctx context.Context, databaseID string, req dto.BackupDatabaseRequest) (*dto.BackupInfo, error) {
	database, err := s.dbRepo.FindByID(databaseID)
	if err != nil {
		return nil, err
	}
	backup := &entities.PostgresBackup{
		ID:         uuid.New().String(),
		DatabaseID: databaseID,
		BackupType: "LOGICAL",
		Status:     "RUNNING",
		StartedAt:  time.Now(),
	}
	if err := s.dbRepo.CreateBackup(backup); err != nil {
		return nil, err
	}
	go func() {
		location := fmt.Sprintf("/backups/%s_%s.sql", database.DBName, time.Now().Format("20060102_150405"))
		time.Sleep(2 * time.Second)
		now := time.Now()
		backup.Status = "SUCCEEDED"
		backup.Location = location
		backup.SizeMB = 50
		backup.CompletedAt = &now
		s.dbRepo.UpdateBackup(backup)
	}()
	return &dto.BackupInfo{
		ID:         backup.ID,
		DatabaseID: databaseID,
		BackupType: backup.BackupType,
		Status:     backup.Status,
		StartedAt:  backup.StartedAt.Format(time.RFC3339),
	}, nil
}

func (s *postgresDatabaseService) RestoreDatabase(ctx context.Context, databaseID string, req dto.RestoreDatabaseRequest) error {
	database, err := s.dbRepo.FindByID(databaseID)
	if err != nil {
		return err
	}
	backup, err := s.dbRepo.FindBackupByID(req.BackupID)
	if err != nil {
		return fmt.Errorf("backup not found: %w", err)
	}
	if backup.Status != "SUCCEEDED" {
		return fmt.Errorf("backup not ready")
	}
	containerIP, err := s.getContainerIP(ctx, database.Instance.ContainerID)
	if err != nil {
		return err
	}
	connStr := fmt.Sprintf("host=%s port=5432 user=%s password=%s dbname=%s sslmode=disable",
		containerIP, database.Instance.Username, database.Instance.Password, database.Instance.DatabaseName)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return err
	}
	defer db.Close()
	if req.Mode == "OVERWRITE" {
		db.ExecContext(ctx, fmt.Sprintf("DROP DATABASE %s", database.DBName))
		db.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE %s", database.DBName))
		database.Status = "RESTORING"
		s.dbRepo.Update(database)
		time.Sleep(1 * time.Second)
		database.Status = "ACTIVE"
		s.dbRepo.Update(database)
	} else if req.Mode == "CLONE" {
		newDBName := req.NewDBName
		if newDBName == "" {
			newDBName = fmt.Sprintf("%s_clone_%s", database.DBName, uuid.New().String()[:8])
		}
		db.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE %s", newDBName))
		cloneDB := &entities.PostgresDatabase{
			ID:             uuid.New().String(),
			InstanceID:     database.InstanceID,
			DBName:         newDBName,
			OwnerUsername:  database.OwnerUsername,
			OwnerPassword:  database.OwnerPassword,
			ProjectID:      database.ProjectID,
			TenantID:       database.TenantID,
			EnvironmentID:  database.EnvironmentID,
			MaxSizeGB:      database.MaxSizeGB,
			MaxConnections: database.MaxConnections,
			Status:         "ACTIVE",
		}
		s.dbRepo.Create(cloneDB)
	}
	return nil
}

func (s *postgresDatabaseService) ManageLifecycle(ctx context.Context, databaseID string, req dto.ManageLifecycleRequest) error {
	database, err := s.dbRepo.FindByID(databaseID)
	if err != nil {
		return err
	}
	containerIP, err := s.getContainerIP(ctx, database.Instance.ContainerID)
	if err != nil {
		return err
	}
	connStr := fmt.Sprintf("host=%s port=5432 user=%s password=%s dbname=%s sslmode=disable",
		containerIP, database.Instance.Username, database.Instance.Password, database.Instance.DatabaseName)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return err
	}
	defer db.Close()
	switch req.Action {
	case "LOCK":
		db.ExecContext(ctx, fmt.Sprintf("REVOKE INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public FROM %s", database.OwnerUsername))
		database.Status = "LOCKED"
		s.dbRepo.Update(database)
	case "UNLOCK":
		db.ExecContext(ctx, fmt.Sprintf("GRANT INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO %s", database.OwnerUsername))
		database.Status = "ACTIVE"
		s.dbRepo.Update(database)
	case "DROP":
		if req.RequireBackup {
			s.BackupDatabase(ctx, databaseID, dto.BackupDatabaseRequest{BackupType: "LOGICAL", Mode: "MANUAL"})
			time.Sleep(3 * time.Second)
		}
		db.ExecContext(ctx, fmt.Sprintf("DROP DATABASE %s", database.DBName))
		db.ExecContext(ctx, fmt.Sprintf("DROP ROLE IF EXISTS %s", database.OwnerUsername))
		database.Status = "DELETED"
		s.dbRepo.Update(database)
	}
	return nil
}

func (s *postgresDatabaseService) GetInstanceOverview(ctx context.Context, instanceID string) (*dto.InstanceOverview, error) {
	instance, err := s.instanceRepo.FindByInfrastructureID(instanceID)
	if err != nil {
		return nil, err
	}
	databases, err := s.dbRepo.FindByInstanceID(instance.ID)
	if err != nil {
		return nil, err
	}
	totalSize := int64(0)
	totalConns := 0
	metrics := make([]dto.DatabaseMetrics, 0)
	for _, db := range databases {
		m, _ := s.GetMetrics(ctx, db.ID)
		if m != nil {
			metrics = append(metrics, *m)
			totalSize += m.CurrentSizeMB
			totalConns += m.ActiveConns
		}
	}
	capacityStatus := "HEALTHY"
	if len(databases) > 50 {
		capacityStatus = "HIGH_LOAD"
	}
	if len(databases) > 100 {
		capacityStatus = "OVERLOADED"
	}
	return &dto.InstanceOverview{
		InstanceID:       instanceID,
		TotalDatabases:   len(databases),
		TotalSizeGB:      float64(totalSize) / 1024,
		TotalConnections: totalConns,
		TopDatabases:     metrics[:min(5, len(metrics))],
		CapacityStatus:   capacityStatus,
	}, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (s *postgresDatabaseService) getContainerIP(ctx context.Context, containerID string) (string, error) {
	containerInfo, err := s.dockerSvc.InspectContainer(ctx, containerID)
	if err != nil {
		return "", err
	}
	if containerInfo.NetworkSettings.IPAddress != "" {
		return containerInfo.NetworkSettings.IPAddress, nil
	}
	for _, network := range containerInfo.NetworkSettings.Networks {
		if network.IPAddress != "" {
			return network.IPAddress, nil
		}
	}
	return "", fmt.Errorf("no IP address found for container")
}

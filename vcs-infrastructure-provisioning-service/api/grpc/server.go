package grpc

import (
	"context"

	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/dto"
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/usecases/services"
)

type ProvisioningServer struct {
	pgService services.IPostgreSQLService
}

func NewProvisioningServer(pgService services.IPostgreSQLService) *ProvisioningServer {
	return &ProvisioningServer{
		pgService: pgService,
	}
}

func (s *ProvisioningServer) CreatePostgreSQL(ctx context.Context, userID string, req dto.CreatePostgreSQLRequest) (*dto.PostgreSQLInfoResponse, error) {
	return s.pgService.CreatePostgreSQL(ctx, userID, req)
}

func (s *ProvisioningServer) GetPostgreSQLInfo(ctx context.Context, id string) (*dto.PostgreSQLInfoResponse, error) {
	return s.pgService.GetPostgreSQLInfo(ctx, id)
}

func (s *ProvisioningServer) StartPostgreSQL(ctx context.Context, id string) error {
	return s.pgService.StartPostgreSQL(ctx, id)
}

func (s *ProvisioningServer) StopPostgreSQL(ctx context.Context, id string) error {
	return s.pgService.StopPostgreSQL(ctx, id)
}

func (s *ProvisioningServer) DeletePostgreSQL(ctx context.Context, id string) error {
	return s.pgService.DeletePostgreSQL(ctx, id)
}


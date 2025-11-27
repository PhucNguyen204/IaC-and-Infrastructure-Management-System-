package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/dto"
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/entities"
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/infrastructures/docker"
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/infrastructures/kafka"
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/pkg/logger"
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/usecases/repositories"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type IDinDService interface {
	// Environment management
	CreateEnvironment(ctx context.Context, userID string, req dto.CreateDinDEnvironmentRequest) (*dto.DinDEnvironmentInfo, error)
	GetEnvironment(ctx context.Context, id string) (*dto.DinDEnvironmentInfo, error)
	GetEnvironmentByInfraID(ctx context.Context, infraID string) (*dto.DinDEnvironmentInfo, error)
	ListEnvironments(ctx context.Context, userID string) ([]dto.DinDEnvironmentInfo, error)
	DeleteEnvironment(ctx context.Context, id string) error
	StartEnvironment(ctx context.Context, id string) error
	StopEnvironment(ctx context.Context, id string) error

	// Docker operations inside DinD
	ExecCommand(ctx context.Context, id string, req dto.ExecCommandRequest) (*dto.ExecCommandResponse, error)
	BuildImage(ctx context.Context, id string, req dto.BuildImageRequest) (*dto.BuildImageResponse, error)
	RunCompose(ctx context.Context, id string, req dto.ComposeRequest) (*dto.ComposeResponse, error)
	PullImage(ctx context.Context, id string, req dto.PullImageRequest) (*dto.PullImageResponse, error)

	// Info retrieval
	ListContainers(ctx context.Context, id string) (*dto.ListContainersResponse, error)
	ListImages(ctx context.Context, id string) (*dto.ListImagesResponse, error)
	GetLogs(ctx context.Context, id string, tail int) (*dto.DinDLogsResponse, error)
	GetStats(ctx context.Context, id string) (*dto.DinDStatsResponse, error)
}

type dinDService struct {
	dinDRepo      repositories.IDinDRepository
	infraRepo     repositories.IInfrastructureRepository
	dockerSvc     docker.IDockerService
	kafkaProducer kafka.IKafkaProducer
	logger        logger.ILogger
}

func NewDinDService(
	dinDRepo repositories.IDinDRepository,
	infraRepo repositories.IInfrastructureRepository,
	dockerSvc docker.IDockerService,
	kafkaProducer kafka.IKafkaProducer,
	logger logger.ILogger,
) IDinDService {
	return &dinDService{
		dinDRepo:      dinDRepo,
		infraRepo:     infraRepo,
		dockerSvc:     dockerSvc,
		kafkaProducer: kafkaProducer,
		logger:        logger,
	}
}

// getResourceLimits returns CPU and Memory limits based on plan
func (s *dinDService) getResourceLimits(plan string) (string, string) {
	switch plan {
	case "small":
		return "1", "1g"
	case "large":
		return "4", "4g"
	default: // medium
		return "2", "2g"
	}
}

// CreateEnvironment creates a new Docker-in-Docker environment
func (s *dinDService) CreateEnvironment(ctx context.Context, userID string, req dto.CreateDinDEnvironmentRequest) (*dto.DinDEnvironmentInfo, error) {
	infraID := uuid.New().String()
	envID := uuid.New().String()

	s.logger.Info("creating DinD environment",
		zap.String("name", req.Name),
		zap.String("user_id", userID))

	// Create infrastructure record
	infra := &entities.Infrastructure{
		ID:     infraID,
		Name:   req.Name,
		Type:   entities.TypeDinD,
		Status: entities.StatusCreating,
		UserID: userID,
	}
	if err := s.infraRepo.Create(infra); err != nil {
		s.logger.Error("failed to create infrastructure", zap.Error(err))
		return nil, err
	}

	// Get resource limits
	if req.ResourcePlan == "" {
		req.ResourcePlan = "medium"
	}
	cpuLimit, memoryLimit := s.getResourceLimits(req.ResourcePlan)

	// Create DinD environment record
	env := &entities.DinDEnvironment{
		ID:               envID,
		InfrastructureID: infraID,
		Name:             req.Name,
		Status:           "creating",
		ResourcePlan:     req.ResourcePlan,
		CPULimit:         cpuLimit,
		MemoryLimit:      memoryLimit,
		StorageDriver:    "overlay2",
		Description:      req.Description,
		AutoCleanup:      req.AutoCleanup,
		TTLHours:         req.TTLHours,
		UserID:           userID,
	}

	if req.AutoCleanup && req.TTLHours > 0 {
		env.ExpiresAt = time.Now().Add(time.Duration(req.TTLHours) * time.Hour)
	}

	if err := s.dinDRepo.Create(env); err != nil {
		s.logger.Error("failed to create DinD environment record", zap.Error(err))
		return nil, err
	}

	// Create network for DinD
	networkName := fmt.Sprintf("dind-network-%s", envID)
	networkID, err := s.dockerSvc.CreateNetwork(ctx, networkName)
	if err != nil {
		s.logger.Warn("failed to create network, using default", zap.Error(err))
		networkID = ""
	}
	env.NetworkID = networkID

	// Create DinD container
	containerName := fmt.Sprintf("iaas-dind-%s", envID)

	// Convert limits to proper format
	cpuNano := int64(0)
	memBytes := int64(0)
	switch cpuLimit {
	case "1":
		cpuNano = 1000000000 // 1 CPU
	case "2":
		cpuNano = 2000000000 // 2 CPUs
	case "4":
		cpuNano = 4000000000 // 4 CPUs
	}
	switch memoryLimit {
	case "1g":
		memBytes = 1073741824 // 1GB
	case "2g":
		memBytes = 2147483648 // 2GB
	case "4g":
		memBytes = 4294967296 // 4GB
	}

	containerConfig := docker.ContainerConfig{
		Name:  containerName,
		Image: "docker:dind", // Official Docker-in-Docker image
		Env: []string{
			"DOCKER_TLS_CERTDIR=", // Disable TLS for simplicity
		},
		Network: networkName,
		Resources: docker.ResourceConfig{
			CPULimit:    cpuNano,
			MemoryLimit: memBytes,
		},
		Privileged: true, // Required for DinD
	}

	containerID, err := s.dockerSvc.CreateDinDContainer(ctx, containerConfig)
	if err != nil {
		s.logger.Error("failed to create DinD container", zap.Error(err))
		env.Status = "failed"
		s.dinDRepo.Update(env)
		infra.Status = entities.StatusFailed
		s.infraRepo.Update(infra)
		return nil, fmt.Errorf("failed to create DinD container: %w", err)
	}

	env.ContainerID = containerID
	env.ContainerName = containerName

	// Start the container
	if err := s.dockerSvc.StartContainer(ctx, containerID); err != nil {
		s.logger.Error("failed to start DinD container", zap.Error(err))
		env.Status = "failed"
		s.dinDRepo.Update(env)
		infra.Status = entities.StatusFailed
		s.infraRepo.Update(infra)
		return nil, fmt.Errorf("failed to start DinD container: %w", err)
	}

	// Wait for Docker daemon to be ready inside DinD
	s.logger.Info("waiting for Docker daemon to be ready inside DinD...")
	if err := s.waitForDinDReady(ctx, containerID, 30*time.Second); err != nil {
		s.logger.Warn("Docker daemon may not be fully ready", zap.Error(err))
	}

	// Get container IP
	if containerInfo, err := s.dockerSvc.InspectContainer(ctx, containerID); err == nil {
		for _, network := range containerInfo.NetworkSettings.Networks {
			if network.IPAddress != "" {
				env.IPAddress = network.IPAddress
				env.DockerHost = fmt.Sprintf("tcp://%s:2375", network.IPAddress)
				break
			}
		}
	}

	env.Status = "running"
	s.dinDRepo.Update(env)

	infra.Status = entities.StatusRunning
	s.infraRepo.Update(infra)

	// Publish event
	s.publishEvent(ctx, env, "created")

	s.logger.Info("DinD environment created successfully",
		zap.String("id", envID),
		zap.String("container_id", containerID))

	return s.GetEnvironment(ctx, envID)
}

// GetEnvironment retrieves environment info
func (s *dinDService) GetEnvironment(ctx context.Context, id string) (*dto.DinDEnvironmentInfo, error) {
	env, err := s.dinDRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// Sync status from container
	if env.ContainerID != "" {
		if containerInfo, err := s.dockerSvc.InspectContainer(ctx, env.ContainerID); err == nil {
			if containerInfo.State.Running {
				env.Status = "running"
			} else {
				env.Status = "stopped"
			}
			s.dinDRepo.Update(env)
		}
	}

	return s.toDTO(env), nil
}

// GetEnvironmentByInfraID retrieves environment by infrastructure ID
func (s *dinDService) GetEnvironmentByInfraID(ctx context.Context, infraID string) (*dto.DinDEnvironmentInfo, error) {
	env, err := s.dinDRepo.FindByInfrastructureID(infraID)
	if err != nil {
		return nil, err
	}

	// Sync status from container
	if env.ContainerID != "" {
		if containerInfo, err := s.dockerSvc.InspectContainer(ctx, env.ContainerID); err == nil {
			if containerInfo.State.Running {
				env.Status = "running"
			} else {
				env.Status = "stopped"
			}
			s.dinDRepo.Update(env)
		}
	}

	return s.toDTO(env), nil
}

// ListEnvironments lists all environments for a user
func (s *dinDService) ListEnvironments(ctx context.Context, userID string) ([]dto.DinDEnvironmentInfo, error) {
	envs, err := s.dinDRepo.FindByUserID(userID)
	if err != nil {
		return nil, err
	}

	result := make([]dto.DinDEnvironmentInfo, 0, len(envs))
	for _, env := range envs {
		result = append(result, *s.toDTO(&env))
	}
	return result, nil
}

// DeleteEnvironment deletes a DinD environment
func (s *dinDService) DeleteEnvironment(ctx context.Context, id string) error {
	env, err := s.dinDRepo.FindByID(id)
	if err != nil {
		return err
	}

	// Stop and remove container
	if env.ContainerID != "" {
		s.dockerSvc.StopContainer(ctx, env.ContainerID)
		s.dockerSvc.RemoveContainer(ctx, env.ContainerID)
	}

	// Remove network
	if env.NetworkID != "" {
		s.dockerSvc.RemoveNetwork(ctx, env.NetworkID)
	}

	// Delete from database
	if err := s.dinDRepo.Delete(id); err != nil {
		return err
	}

	// Update infrastructure
	if infra, err := s.infraRepo.FindByID(env.InfrastructureID); err == nil {
		infra.Status = entities.StatusDeleted
		s.infraRepo.Update(infra)
	}

	s.publishEvent(ctx, env, "deleted")

	s.logger.Info("DinD environment deleted", zap.String("id", id))
	return nil
}

// StartEnvironment starts a stopped environment
func (s *dinDService) StartEnvironment(ctx context.Context, id string) error {
	env, err := s.dinDRepo.FindByID(id)
	if err != nil {
		return err
	}

	if err := s.dockerSvc.StartContainer(ctx, env.ContainerID); err != nil {
		return err
	}

	env.Status = "running"
	s.dinDRepo.Update(env)

	s.publishEvent(ctx, env, "started")
	return nil
}

// StopEnvironment stops a running environment
func (s *dinDService) StopEnvironment(ctx context.Context, id string) error {
	env, err := s.dinDRepo.FindByID(id)
	if err != nil {
		return err
	}

	if err := s.dockerSvc.StopContainer(ctx, env.ContainerID); err != nil {
		return err
	}

	env.Status = "stopped"
	s.dinDRepo.Update(env)

	s.publishEvent(ctx, env, "stopped")
	return nil
}

// ExecCommand executes a docker command inside the DinD environment
func (s *dinDService) ExecCommand(ctx context.Context, id string, req dto.ExecCommandRequest) (*dto.ExecCommandResponse, error) {
	env, err := s.dinDRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if env.Status != "running" {
		return nil, fmt.Errorf("environment is not running")
	}

	startTime := time.Now()

	// Execute command inside DinD container
	// The command should be a docker command, we execute it directly
	cmd := strings.Fields(req.Command)
	if len(cmd) == 0 {
		return nil, fmt.Errorf("empty command")
	}

	output, err := s.dockerSvc.ExecCommand(ctx, env.ContainerID, cmd)

	duration := time.Since(startTime)
	exitCode := 0
	if err != nil {
		exitCode = 1
		output = err.Error()
	}

	// Save command history
	history := &entities.DinDCommandHistory{
		ID:            uuid.New().String(),
		EnvironmentID: id,
		Command:       req.Command,
		Output:        output,
		ExitCode:      exitCode,
		Duration:      int(duration.Milliseconds()),
	}
	s.dinDRepo.CreateCommandHistory(history)

	return &dto.ExecCommandResponse{
		Command:    req.Command,
		Output:     output,
		ExitCode:   exitCode,
		Duration:   duration.String(),
		ExecutedAt: time.Now().Format(time.RFC3339),
	}, nil
}

// BuildImage builds a Docker image inside DinD environment
func (s *dinDService) BuildImage(ctx context.Context, id string, req dto.BuildImageRequest) (*dto.BuildImageResponse, error) {
	env, err := s.dinDRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if env.Status != "running" {
		return nil, fmt.Errorf("environment is not running")
	}

	startTime := time.Now()

	// Create Dockerfile inside container
	dockerfileCmd := []string{"sh", "-c", fmt.Sprintf("mkdir -p /build && cat > /build/Dockerfile << 'DOCKERFILE'\n%s\nDOCKERFILE", req.Dockerfile)}
	if _, err := s.dockerSvc.ExecCommand(ctx, env.ContainerID, dockerfileCmd); err != nil {
		return nil, fmt.Errorf("failed to create Dockerfile: %w", err)
	}

	// Build image
	tag := req.Tag
	if tag == "" {
		tag = "latest"
	}
	imageName := fmt.Sprintf("%s:%s", req.ImageName, tag)

	buildCmd := []string{"docker", "build", "-t", imageName, "/build"}
	if req.NoCache {
		buildCmd = append(buildCmd, "--no-cache")
	}

	output, err := s.dockerSvc.ExecCommand(ctx, env.ContainerID, buildCmd)
	success := err == nil

	duration := time.Since(startTime)

	// Get image info if build succeeded
	imageID := ""
	size := ""
	if success {
		inspectCmd := []string{"docker", "image", "inspect", imageName, "--format", "{{.Id}} {{.Size}}"}
		if inspectOutput, err := s.dockerSvc.ExecCommand(ctx, env.ContainerID, inspectCmd); err == nil {
			parts := strings.Fields(inspectOutput)
			if len(parts) >= 1 {
				imageID = parts[0]
			}
			if len(parts) >= 2 {
				size = parts[1]
			}
		}
	}

	return &dto.BuildImageResponse{
		ImageName: req.ImageName,
		Tag:       tag,
		ImageID:   imageID,
		Size:      size,
		BuildLogs: strings.Split(output, "\n"),
		Duration:  duration.String(),
		Success:   success,
	}, nil
}

// RunCompose runs docker-compose inside DinD environment
func (s *dinDService) RunCompose(ctx context.Context, id string, req dto.ComposeRequest) (*dto.ComposeResponse, error) {
	env, err := s.dinDRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if env.Status != "running" {
		return nil, fmt.Errorf("environment is not running")
	}

	// Create docker-compose.yml inside container
	composeCmd := []string{"sh", "-c", fmt.Sprintf("mkdir -p /compose && cat > /compose/docker-compose.yml << 'COMPOSEFILE'\n%s\nCOMPOSEFILE", req.ComposeContent)}
	if _, err := s.dockerSvc.ExecCommand(ctx, env.ContainerID, composeCmd); err != nil {
		return nil, fmt.Errorf("failed to create docker-compose.yml: %w", err)
	}

	// Run docker-compose command
	var cmd []string
	switch req.Action {
	case "up":
		cmd = []string{"docker-compose", "-f", "/compose/docker-compose.yml", "up"}
		if req.Detach {
			cmd = append(cmd, "-d")
		}
	case "down":
		cmd = []string{"docker-compose", "-f", "/compose/docker-compose.yml", "down"}
	case "restart":
		cmd = []string{"docker-compose", "-f", "/compose/docker-compose.yml", "restart"}
	case "logs":
		cmd = []string{"docker-compose", "-f", "/compose/docker-compose.yml", "logs"}
	case "ps":
		cmd = []string{"docker-compose", "-f", "/compose/docker-compose.yml", "ps"}
	default:
		return nil, fmt.Errorf("unknown action: %s", req.Action)
	}

	if req.ServiceName != "" {
		cmd = append(cmd, req.ServiceName)
	}

	output, err := s.dockerSvc.ExecCommand(ctx, env.ContainerID, cmd)
	success := err == nil

	// Get list of services
	servicesCmd := []string{"docker-compose", "-f", "/compose/docker-compose.yml", "config", "--services"}
	servicesOutput, _ := s.dockerSvc.ExecCommand(ctx, env.ContainerID, servicesCmd)
	services := strings.Fields(servicesOutput)

	return &dto.ComposeResponse{
		Action:   req.Action,
		Output:   output,
		Services: services,
		Success:  success,
	}, nil
}

// PullImage pulls an image inside DinD environment
func (s *dinDService) PullImage(ctx context.Context, id string, req dto.PullImageRequest) (*dto.PullImageResponse, error) {
	env, err := s.dinDRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if env.Status != "running" {
		return nil, fmt.Errorf("environment is not running")
	}

	cmd := []string{"docker", "pull", req.Image}
	_, err = s.dockerSvc.ExecCommand(ctx, env.ContainerID, cmd)

	success := err == nil
	status := "pulled"
	if !success {
		status = "failed"
	}

	// Get digest
	digest := ""
	if success {
		inspectCmd := []string{"docker", "image", "inspect", req.Image, "--format", "{{.RepoDigests}}"}
		if digestOutput, err := s.dockerSvc.ExecCommand(ctx, env.ContainerID, inspectCmd); err == nil {
			digest = strings.TrimSpace(digestOutput)
		}
	}

	return &dto.PullImageResponse{
		Image:   req.Image,
		Status:  status,
		Digest:  digest,
		Success: success,
	}, nil
}

// ListContainers lists containers inside DinD environment
func (s *dinDService) ListContainers(ctx context.Context, id string) (*dto.ListContainersResponse, error) {
	env, err := s.dinDRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if env.Status != "running" {
		return nil, fmt.Errorf("environment is not running")
	}

	cmd := []string{"docker", "ps", "-a", "--format", "{{.ID}}|{{.Names}}|{{.Image}}|{{.Status}}|{{.Ports}}|{{.CreatedAt}}"}
	output, err := s.dockerSvc.ExecCommand(ctx, env.ContainerID, cmd)
	if err != nil {
		return nil, err
	}

	containers := make([]dto.DinDContainerInfo, 0)
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) >= 6 {
			containers = append(containers, dto.DinDContainerInfo{
				ContainerID: parts[0],
				Name:        parts[1],
				Image:       parts[2],
				Status:      parts[3],
				Ports:       parts[4],
				Created:     parts[5],
			})
		}
	}

	return &dto.ListContainersResponse{
		Containers: containers,
		Total:      len(containers),
	}, nil
}

// ListImages lists images inside DinD environment
func (s *dinDService) ListImages(ctx context.Context, id string) (*dto.ListImagesResponse, error) {
	env, err := s.dinDRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if env.Status != "running" {
		return nil, fmt.Errorf("environment is not running")
	}

	cmd := []string{"docker", "images", "--format", "{{.ID}}|{{.Repository}}|{{.Tag}}|{{.Size}}|{{.CreatedAt}}"}
	output, err := s.dockerSvc.ExecCommand(ctx, env.ContainerID, cmd)
	if err != nil {
		return nil, err
	}

	images := make([]dto.DinDImageInfo, 0)
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) >= 5 {
			images = append(images, dto.DinDImageInfo{
				ImageID:    parts[0],
				Repository: parts[1],
				Tag:        parts[2],
				Size:       parts[3],
				Created:    parts[4],
			})
		}
	}

	return &dto.ListImagesResponse{
		Images: images,
		Total:  len(images),
	}, nil
}

// GetLogs gets logs from DinD container
func (s *dinDService) GetLogs(ctx context.Context, id string, tail int) (*dto.DinDLogsResponse, error) {
	env, err := s.dinDRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if tail <= 0 {
		tail = 100
	}

	logs, err := s.dockerSvc.GetContainerLogs(ctx, env.ContainerID, tail)
	if err != nil {
		return nil, err
	}

	return &dto.DinDLogsResponse{
		Logs: logs,
	}, nil
}

// GetStats gets resource stats of DinD environment
func (s *dinDService) GetStats(ctx context.Context, id string) (*dto.DinDStatsResponse, error) {
	_, err := s.dinDRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// Get container count
	containerCount := 0
	if containers, err := s.ListContainers(ctx, id); err == nil {
		containerCount = containers.Total
	}

	// Get image count
	imageCount := 0
	if images, err := s.ListImages(ctx, id); err == nil {
		imageCount = images.Total
	}

	// Return basic stats (memory stats requires parsing the docker stats stream)
	return &dto.DinDStatsResponse{
		CPUUsage:       0,
		MemoryUsage:    0,
		MemoryLimit:    0,
		MemoryPercent:  0,
		ContainerCount: containerCount,
		ImageCount:     imageCount,
	}, nil
}

// waitForDinDReady waits for Docker daemon inside DinD container to be ready
func (s *dinDService) waitForDinDReady(ctx context.Context, containerID string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	checkCmd := []string{"docker", "info"}

	for time.Now().Before(deadline) {
		output, err := s.dockerSvc.ExecCommand(ctx, containerID, checkCmd)
		if err == nil && strings.Contains(output, "Server Version") {
			s.logger.Info("Docker daemon is ready inside DinD")
			return nil
		}

		s.logger.Debug("Docker daemon not ready yet, retrying...",
			zap.String("output", output))
		time.Sleep(2 * time.Second)
	}

	return fmt.Errorf("timeout waiting for Docker daemon to be ready")
}

// toDTO converts entity to DTO
func (s *dinDService) toDTO(env *entities.DinDEnvironment) *dto.DinDEnvironmentInfo {
	info := &dto.DinDEnvironmentInfo{
		ID:               env.ID,
		InfrastructureID: env.InfrastructureID,
		Name:             env.Name,
		ContainerID:      env.ContainerID,
		Status:           env.Status,
		DockerHost:       env.DockerHost,
		IPAddress:        env.IPAddress,
		ResourcePlan:     env.ResourcePlan,
		CPULimit:         env.CPULimit,
		MemoryLimit:      env.MemoryLimit,
		Description:      env.Description,
		AutoCleanup:      env.AutoCleanup,
		TTLHours:         env.TTLHours,
		CreatedAt:        env.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        env.UpdatedAt.Format(time.RFC3339),
	}

	if !env.ExpiresAt.IsZero() {
		info.ExpiresAt = env.ExpiresAt.Format(time.RFC3339)
	}

	return info
}

// publishEvent publishes event to Kafka
func (s *dinDService) publishEvent(ctx context.Context, env *entities.DinDEnvironment, action string) {
	if s.kafkaProducer == nil {
		return
	}

	event := kafka.InfrastructureEvent{
		InstanceID: env.ID,
		UserID:     env.UserID,
		Type:       "dind",
		Action:     action,
		Metadata: map[string]interface{}{
			"name":          env.Name,
			"container_id":  env.ContainerID,
			"resource_plan": env.ResourcePlan,
		},
	}

	s.kafkaProducer.PublishEvent(ctx, event)
}

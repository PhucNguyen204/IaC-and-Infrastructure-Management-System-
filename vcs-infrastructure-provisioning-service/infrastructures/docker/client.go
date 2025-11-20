package docker

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/pkg/logger"
	"go.uber.org/zap"
)

type IDockerService interface {
	CreateContainer(ctx context.Context, config ContainerConfig) (string, error)
	StartContainer(ctx context.Context, containerID string) error
	StopContainer(ctx context.Context, containerID string) error
	RestartContainer(ctx context.Context, containerID string) error
	RemoveContainer(ctx context.Context, containerID string) error
	GetContainerStats(ctx context.Context, containerID string) (types.ContainerStats, error)
	GetContainerLogs(ctx context.Context, containerID string, tail int) ([]string, error)
	ExecCommand(ctx context.Context, containerID string, cmd []string) (string, error)
	InspectContainer(ctx context.Context, containerID string) (*types.ContainerJSON, error)
	CreateNetwork(ctx context.Context, networkName string) (string, error)
	RemoveNetwork(ctx context.Context, networkID string) error
	CreateVolume(ctx context.Context, volumeName string) error
	RemoveVolume(ctx context.Context, volumeName string) error
}

type ContainerConfig struct {
	Name         string
	Image        string
	Env          []string
	Ports        map[string]string
	Volumes      map[string]string
	Network      string
	NetworkAlias string
	Cmd          []string
	Resources    ResourceConfig
}

type ResourceConfig struct {
	CPULimit    int64
	MemoryLimit int64
}

type dockerService struct {
	client *client.Client
	logger logger.ILogger
}

func NewDockerService(logger logger.ILogger) (IDockerService, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	return &dockerService{
		client: cli,
		logger: logger,
	}, nil
}

func (ds *dockerService) CreateContainer(ctx context.Context, config ContainerConfig) (string, error) {
	ds.logger.Info("pulling image", zap.String("image", config.Image))
	reader, err := ds.client.ImagePull(ctx, config.Image, types.ImagePullOptions{})
	if err != nil {
		ds.logger.Error("failed to pull image", zap.Error(err))
		return "", err
	}
	defer reader.Close()
	io.Copy(io.Discard, reader)

	portBindings := nat.PortMap{}
	exposedPorts := nat.PortSet{}
	for containerPort, hostPort := range config.Ports {
		port, err := nat.NewPort("tcp", containerPort)
		if err != nil {
			return "", err
		}
		portBindings[port] = []nat.PortBinding{{HostPort: hostPort}}
		exposedPorts[port] = struct{}{}
	}

	binds := []string{}
	for hostPath, containerPath := range config.Volumes {
		binds = append(binds, fmt.Sprintf("%s:%s", hostPath, containerPath))
	}

	containerConfig := &container.Config{
		Image:        config.Image,
		Env:          config.Env,
		ExposedPorts: exposedPorts,
		Cmd:          config.Cmd,
	}

	hostConfig := &container.HostConfig{
		PortBindings: portBindings,
		Binds:        binds,
		Resources: container.Resources{
			NanoCPUs: config.Resources.CPULimit,
			Memory:   config.Resources.MemoryLimit,
		},
		RestartPolicy: container.RestartPolicy{
			Name: "unless-stopped",
		},
	}

	networkConfig := &network.NetworkingConfig{}
	if config.Network != "" {
		networkConfig.EndpointsConfig = map[string]*network.EndpointSettings{
			config.Network: {
				Aliases: []string{config.NetworkAlias},
			},
		}
	}

	resp, err := ds.client.ContainerCreate(ctx, containerConfig, hostConfig, networkConfig, nil, config.Name)
	if err != nil {
		ds.logger.Error("failed to create container", zap.Error(err))
		return "", err
	}

	ds.logger.Info("container created", zap.String("container_id", resp.ID), zap.String("name", config.Name))
	return resp.ID, nil
}

func (ds *dockerService) StartContainer(ctx context.Context, containerID string) error {
	if err := ds.client.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		ds.logger.Error("failed to start container", zap.String("container_id", containerID), zap.Error(err))
		return err
	}
	ds.logger.Info("container started", zap.String("container_id", containerID))
	return nil
}

func (ds *dockerService) StopContainer(ctx context.Context, containerID string) error {
	timeout := 30
	if err := ds.client.ContainerStop(ctx, containerID, container.StopOptions{Timeout: &timeout}); err != nil {
		ds.logger.Error("failed to stop container", zap.String("container_id", containerID), zap.Error(err))
		return err
	}
	ds.logger.Info("container stopped", zap.String("container_id", containerID))
	return nil
}

func (ds *dockerService) RestartContainer(ctx context.Context, containerID string) error {
	timeout := 30
	if err := ds.client.ContainerRestart(ctx, containerID, container.StopOptions{Timeout: &timeout}); err != nil {
		ds.logger.Error("failed to restart container", zap.String("container_id", containerID), zap.Error(err))
		return err
	}
	ds.logger.Info("container restarted", zap.String("container_id", containerID))
	return nil
}

func (ds *dockerService) RemoveContainer(ctx context.Context, containerID string) error {
	if err := ds.client.ContainerRemove(ctx, containerID, container.RemoveOptions{
		Force:         true,
		RemoveVolumes: true,
	}); err != nil {
		ds.logger.Error("failed to remove container", zap.String("container_id", containerID), zap.Error(err))
		return err
	}
	ds.logger.Info("container removed", zap.String("container_id", containerID))
	return nil
}

func (ds *dockerService) GetContainerStats(ctx context.Context, containerID string) (types.ContainerStats, error) {
	stats, err := ds.client.ContainerStats(ctx, containerID, false)
	if err != nil {
		ds.logger.Error("failed to get container stats", zap.String("container_id", containerID), zap.Error(err))
		return types.ContainerStats{}, err
	}

	return stats, nil
}

func (ds *dockerService) GetContainerLogs(ctx context.Context, containerID string, tail int) ([]string, error) {
	tailStr := "all"
	if tail > 0 {
		tailStr = fmt.Sprintf("%d", tail)
	}

	options := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Tail:       tailStr,
	}

	logs, err := ds.client.ContainerLogs(ctx, containerID, options)
	if err != nil {
		ds.logger.Error("failed to get container logs", zap.String("container_id", containerID), zap.Error(err))
		return nil, err
	}
	defer logs.Close()

	logBytes, err := io.ReadAll(logs)
	if err != nil {
		return nil, err
	}

	logLines := strings.Split(string(logBytes), "\n")
	result := []string{}
	for _, line := range logLines {
		if len(line) > 8 {
			result = append(result, line[8:])
		}
	}

	return result, nil
}

func (ds *dockerService) ExecCommand(ctx context.Context, containerID string, cmd []string) (string, error) {
	execConfig := types.ExecConfig{
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          cmd,
	}

	execResp, err := ds.client.ContainerExecCreate(ctx, containerID, execConfig)
	if err != nil {
		ds.logger.Error("failed to create exec", zap.String("container_id", containerID), zap.Error(err))
		return "", err
	}

	attachResp, err := ds.client.ContainerExecAttach(ctx, execResp.ID, types.ExecStartCheck{})
	if err != nil {
		ds.logger.Error("failed to attach exec", zap.String("exec_id", execResp.ID), zap.Error(err))
		return "", err
	}
	defer attachResp.Close()

	output, err := io.ReadAll(attachResp.Reader)
	if err != nil {
		return "", err
	}

	return string(output), nil
}

func (ds *dockerService) InspectContainer(ctx context.Context, containerID string) (*types.ContainerJSON, error) {
	inspect, err := ds.client.ContainerInspect(ctx, containerID)
	if err != nil {
		ds.logger.Error("failed to inspect container", zap.String("container_id", containerID), zap.Error(err))
		return nil, err
	}
	return &inspect, nil
}

func (ds *dockerService) CreateNetwork(ctx context.Context, networkName string) (string, error) {
	filter := filters.NewArgs()
	filter.Add("name", networkName)
	networks, err := ds.client.NetworkList(ctx, types.NetworkListOptions{Filters: filter})
	if err != nil {
		return "", err
	}

	if len(networks) > 0 {
		ds.logger.Info("network already exists", zap.String("network", networkName))
		return networks[0].ID, nil
	}

	resp, err := ds.client.NetworkCreate(ctx, networkName, types.NetworkCreate{
		Driver: "bridge",
	})
	if err != nil {
		ds.logger.Error("failed to create network", zap.String("network", networkName), zap.Error(err))
		return "", err
	}

	ds.logger.Info("network created", zap.String("network_id", resp.ID), zap.String("network", networkName))
	return resp.ID, nil
}

func (ds *dockerService) RemoveNetwork(ctx context.Context, networkID string) error {
	if err := ds.client.NetworkRemove(ctx, networkID); err != nil {
		ds.logger.Error("failed to remove network", zap.String("network_id", networkID), zap.Error(err))
		return err
	}
	ds.logger.Info("network removed", zap.String("network_id", networkID))
	return nil
}

func (ds *dockerService) CreateVolume(ctx context.Context, volumeName string) error {
	filter := filters.NewArgs()
	filter.Add("name", volumeName)
	volumes, err := ds.client.VolumeList(ctx, volume.ListOptions{Filters: filter})
	if err != nil {
		return err
	}

	if len(volumes.Volumes) > 0 {
		ds.logger.Info("volume already exists", zap.String("volume", volumeName))
		return nil
	}

	_, err = ds.client.VolumeCreate(ctx, volume.CreateOptions{
		Name: volumeName,
	})
	if err != nil {
		ds.logger.Error("failed to create volume", zap.String("volume", volumeName), zap.Error(err))
		return err
	}

	ds.logger.Info("volume created", zap.String("volume", volumeName))
	return nil
}

func (ds *dockerService) RemoveVolume(ctx context.Context, volumeName string) error {
	if err := ds.client.VolumeRemove(ctx, volumeName, true); err != nil {
		ds.logger.Error("failed to remove volume", zap.String("volume", volumeName), zap.Error(err))
		return err
	}
	ds.logger.Info("volume removed", zap.String("volume", volumeName))
	return nil
}

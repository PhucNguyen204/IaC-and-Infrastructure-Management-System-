package collectors

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/PhucNguyen204/vcs-infrastructure-monitoring-service/pkg/logger"
	"go.uber.org/zap"
)

type INginxCollector interface {
	CollectMetrics(ctx context.Context, statusURL string) (*NginxMetrics, error)
}

type NginxMetrics struct {
	ActiveConnections int64
	Accepts           int64
	Handled           int64
	Requests          int64
	Reading           int64
	Writing           int64
	Waiting           int64
}

type nginxCollector struct {
	logger logger.ILogger
}

func NewNginxCollector(logger logger.ILogger) INginxCollector {
	return &nginxCollector{
		logger: logger,
	}
}

func (nc *nginxCollector) CollectMetrics(ctx context.Context, statusURL string) (*NginxMetrics, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", statusURL, nil)
	if err != nil {
		nc.logger.Error("failed to create request", zap.Error(err))
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		nc.logger.Error("failed to fetch nginx status", zap.Error(err))
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		nc.logger.Error("nginx status returned non-200", zap.Int("status_code", resp.StatusCode))
		return nil, fmt.Errorf("nginx status returned %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		nc.logger.Error("failed to read response body", zap.Error(err))
		return nil, err
	}

	return nc.parseNginxStatus(string(body))
}

func (nc *nginxCollector) parseNginxStatus(status string) (*NginxMetrics, error) {
	metrics := &NginxMetrics{}
	lines := strings.Split(status, "\n")

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Active connections:") {
			parts := strings.Split(line, ":")
			if len(parts) == 2 {
				val, err := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
				if err == nil {
					metrics.ActiveConnections = val
				}
			}
		} else if strings.Contains(line, "accepts") && strings.Contains(line, "handled") && strings.Contains(line, "requests") {
			if i+1 < len(lines) {
				fields := strings.Fields(strings.TrimSpace(lines[i+1]))
				if len(fields) >= 3 {
					if val, err := strconv.ParseInt(fields[0], 10, 64); err == nil {
						metrics.Accepts = val
					}
					if val, err := strconv.ParseInt(fields[1], 10, 64); err == nil {
						metrics.Handled = val
					}
					if val, err := strconv.ParseInt(fields[2], 10, 64); err == nil {
						metrics.Requests = val
					}
				}
			}
		} else if strings.HasPrefix(line, "Reading:") {
			fields := strings.Fields(line)
			if len(fields) >= 6 {
				if val, err := strconv.ParseInt(fields[1], 10, 64); err == nil {
					metrics.Reading = val
				}
				if val, err := strconv.ParseInt(fields[3], 10, 64); err == nil {
					metrics.Writing = val
				}
				if val, err := strconv.ParseInt(fields[5], 10, 64); err == nil {
					metrics.Waiting = val
				}
			}
		}
	}

	return metrics, nil
}

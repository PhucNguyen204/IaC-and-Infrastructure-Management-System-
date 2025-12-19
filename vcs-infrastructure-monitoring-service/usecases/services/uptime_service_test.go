package services

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/PhucNguyen204/vcs-infrastructure-monitoring-service/infrastructures/elasticsearch"
	"github.com/PhucNguyen204/vcs-infrastructure-monitoring-service/pkg/logger"
	"go.uber.org/zap"
)

// stubElasticsearchClient cung cấp dữ liệu tĩnh cho uptime service.
type stubElasticsearchClient struct {
	events []elasticsearch.UptimeEvent
}

func (s *stubElasticsearchClient) IndexLog(ctx context.Context, log elasticsearch.LogEntry) error     { return nil }
func (s *stubElasticsearchClient) IndexMetric(ctx context.Context, metric elasticsearch.MetricEntry) error {
	return nil
}
func (s *stubElasticsearchClient) QueryLogs(ctx context.Context, instanceID string, from, size int) ([]elasticsearch.LogEntry, error) {
	return nil, nil
}
func (s *stubElasticsearchClient) QueryMetrics(ctx context.Context, instanceID string, from, size int) ([]elasticsearch.MetricEntry, error) {
	return nil, nil
}
func (s *stubElasticsearchClient) IndexUptimeEvent(ctx context.Context, event elasticsearch.UptimeEvent) error {
	return nil
}

func (s *stubElasticsearchClient) QueryUptimeEvents(ctx context.Context, filter elasticsearch.UptimeFilter) ([]elasticsearch.UptimeEvent, error) {
	out := []elasticsearch.UptimeEvent{}
	for _, e := range s.events {
		if filter.InstanceID != "" && e.InstanceID != filter.InstanceID {
			continue
		}
		if filter.UserID != "" && e.UserID != filter.UserID {
			continue
		}
		if filter.Type != "" && e.Type != filter.Type {
			continue
		}
		if !filter.From.IsZero() && e.Timestamp.Before(filter.From) {
			continue
		}
		if !filter.To.IsZero() && e.Timestamp.After(filter.To) {
			continue
		}
		out = append(out, e)
	}
	return out, nil
}

func (s *stubElasticsearchClient) QueryUptimeByUser(ctx context.Context, userID string, from, to time.Time) ([]elasticsearch.UptimeEvent, error) {
	return s.QueryUptimeEvents(ctx, elasticsearch.UptimeFilter{UserID: userID, From: from, To: to})
}

func (s *stubElasticsearchClient) QueryUptimeByType(ctx context.Context, infraType string, from, to time.Time) ([]elasticsearch.UptimeEvent, error) {
	return s.QueryUptimeEvents(ctx, elasticsearch.UptimeFilter{Type: infraType, From: from, To: to})
}

func (s *stubElasticsearchClient) QueryAllUptimeEvents(ctx context.Context, from, to time.Time, size int) ([]elasticsearch.UptimeEvent, error) {
	return s.QueryUptimeEvents(ctx, elasticsearch.UptimeFilter{From: from, To: to})
}

// fakeLogger tránh phụ thuộc vào logger thật.
type fakeLogger struct{}

func (fakeLogger) Debug(msg string, fields ...zap.Field) {}
func (fakeLogger) Info(msg string, fields ...zap.Field)  {}
func (fakeLogger) Warn(msg string, fields ...zap.Field)  {}
func (fakeLogger) Error(msg string, fields ...zap.Field) {}
func (fakeLogger) Fatal(msg string, fields ...zap.Field) {}
func (fakeLogger) Sync() error                     { return nil }
func (fakeLogger) With(fields ...zap.Field) logger.ILogger {
	return fakeLogger{}
}

func approxEqual(t *testing.T, got, want, delta float64) {
	t.Helper()
	if math.Abs(got-want) > delta {
		t.Fatalf("want %.3f, got %.3f (delta %.3f)", want, got, delta)
	}
}

func TestGetInfrastructureUptime_NoEvents(t *testing.T) {
	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := from.Add(2 * time.Hour)

	stub := &stubElasticsearchClient{events: nil}
	svc := NewUptimeService(stub, fakeLogger{})

	res, err := svc.GetInfrastructureUptime(context.Background(), "inst-1", from, to)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.TotalUptime != 0 || res.TotalDowntime != int64((2*time.Hour).Seconds()) {
		t.Fatalf("unexpected uptime/downtime: %+v", res)
	}
	if res.CurrentStatus != "unknown" {
		t.Fatalf("expected status unknown, got %s", res.CurrentStatus)
	}
}

func TestGetInfrastructureUptime_MixedStatus(t *testing.T) {
	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := from.Add(3 * time.Hour)

	events := []elasticsearch.UptimeEvent{
		{InstanceID: "inst-1", UserID: "user-1", Type: "postgres_cluster", Action: "created", Status: "running", Timestamp: from},
		{InstanceID: "inst-1", UserID: "user-1", Type: "postgres_cluster", Action: "stopped", Status: "stopped", Timestamp: from.Add(1 * time.Hour)},
	}

	stub := &stubElasticsearchClient{events: events}
	svc := NewUptimeService(stub, fakeLogger{})

	res, err := svc.GetInfrastructureUptime(context.Background(), "inst-1", from, to)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.TotalUptime != int64((1 * time.Hour).Seconds()) {
		t.Fatalf("expected uptime 1h, got %d", res.TotalUptime)
	}
	if res.TotalDowntime != int64((2 * time.Hour).Seconds()) {
		t.Fatalf("expected downtime 2h, got %d", res.TotalDowntime)
	}
	approxEqual(t, res.UptimePercent, 33.333, 0.1)
	if res.CurrentStatus != "stopped" {
		t.Fatalf("expected current status stopped, got %s", res.CurrentStatus)
	}
	if len(res.OutageEvents) != 1 {
		t.Fatalf("expected 1 outage event, got %d", len(res.OutageEvents))
	}
}

func TestGetOverallUptimeSummary_ByTypeAggregation(t *testing.T) {
	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := from.Add(4 * time.Hour)

	events := []elasticsearch.UptimeEvent{
		// inst A: 3h up, 1h down
		{InstanceID: "inst-A", UserID: "user-1", Type: "postgres_cluster", Action: "start", Status: "running", Timestamp: from},
		{InstanceID: "inst-A", UserID: "user-1", Type: "postgres_cluster", Action: "stop", Status: "stopped", Timestamp: from.Add(3 * time.Hour)},
		// inst B: always up 4h
		{InstanceID: "inst-B", UserID: "user-2", Type: "nginx_cluster", Action: "start", Status: "running", Timestamp: from},
	}

	stub := &stubElasticsearchClient{events: events}
	svc := NewUptimeService(stub, fakeLogger{})

	res, err := svc.GetOverallUptimeSummary(context.Background(), from, to)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.TotalInfra != 2 {
		t.Fatalf("expected 2 infrastructures, got %d", res.TotalInfra)
	}
	if len(res.ByType) != 2 {
		t.Fatalf("expected 2 types, got %d", len(res.ByType))
	}

	pg := res.ByType["postgres_cluster"]
	// inst A uptime 3h of 4h => 75%
	approxEqual(t, pg.AverageUptime, 75.0, 0.1)

	ng := res.ByType["nginx_cluster"]
	approxEqual(t, ng.AverageUptime, 100.0, 0.1)

	// overall average = (75 + 100) / 2 = 87.5
	approxEqual(t, res.AverageUptime, 87.5, 0.1)
}


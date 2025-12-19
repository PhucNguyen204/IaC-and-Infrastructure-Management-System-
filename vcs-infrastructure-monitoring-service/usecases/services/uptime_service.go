package services

import (
	"context"
	"sort"
	"time"

	"github.com/PhucNguyen204/vcs-infrastructure-monitoring-service/dto"
	"github.com/PhucNguyen204/vcs-infrastructure-monitoring-service/infrastructures/elasticsearch"
	"github.com/PhucNguyen204/vcs-infrastructure-monitoring-service/pkg/logger"
	"go.uber.org/zap"
)

type IUptimeService interface {
	// Record uptime event
	RecordStatusChange(ctx context.Context, event elasticsearch.UptimeEvent) error

	// Get uptime for single infrastructure
	GetInfrastructureUptime(ctx context.Context, instanceID string, from, to time.Time) (*dto.UptimeResponse, error)

	// Get uptime summary for user
	GetUserUptimeSummary(ctx context.Context, userID string, from, to time.Time) (*dto.UptimeSummaryResponse, error)

	// Get uptime by infrastructure type
	GetUptimeByType(ctx context.Context, infraType string, from, to time.Time) (*dto.UptimeSummaryResponse, error)

	// Get overall uptime summary
	GetOverallUptimeSummary(ctx context.Context, from, to time.Time) (*dto.UptimeSummaryResponse, error)
}

type uptimeService struct {
	esClient elasticsearch.IElasticsearchClient
	logger   logger.ILogger
}

func NewUptimeService(esClient elasticsearch.IElasticsearchClient, logger logger.ILogger) IUptimeService {
	return &uptimeService{
		esClient: esClient,
		logger:   logger,
	}
}

func (us *uptimeService) RecordStatusChange(ctx context.Context, event elasticsearch.UptimeEvent) error {
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	us.logger.Info("recording uptime event",
		zap.String("instance_id", event.InstanceID),
		zap.String("action", event.Action),
		zap.String("status", event.Status))

	return us.esClient.IndexUptimeEvent(ctx, event)
}

func (us *uptimeService) GetInfrastructureUptime(ctx context.Context, instanceID string, from, to time.Time) (*dto.UptimeResponse, error) {
	events, err := us.esClient.QueryUptimeEvents(ctx, elasticsearch.UptimeFilter{
		InstanceID: instanceID,
		From:       from,
		To:         to,
		Size:       10000,
	})
	if err != nil {
		us.logger.Error("failed to query uptime events", zap.Error(err))
		return nil, err
	}

	return us.calculateUptime(events, instanceID, from, to)
}

func (us *uptimeService) GetUserUptimeSummary(ctx context.Context, userID string, from, to time.Time) (*dto.UptimeSummaryResponse, error) {
	events, err := us.esClient.QueryUptimeByUser(ctx, userID, from, to)
	if err != nil {
		us.logger.Error("failed to query user uptime events", zap.Error(err))
		return nil, err
	}

	return us.calculateSummary(events, from, to)
}

func (us *uptimeService) GetUptimeByType(ctx context.Context, infraType string, from, to time.Time) (*dto.UptimeSummaryResponse, error) {
	events, err := us.esClient.QueryUptimeByType(ctx, infraType, from, to)
	if err != nil {
		us.logger.Error("failed to query type uptime events", zap.Error(err))
		return nil, err
	}

	return us.calculateSummary(events, from, to)
}

func (us *uptimeService) GetOverallUptimeSummary(ctx context.Context, from, to time.Time) (*dto.UptimeSummaryResponse, error) {
	events, err := us.esClient.QueryAllUptimeEvents(ctx, from, to, 10000)
	if err != nil {
		us.logger.Error("failed to query all uptime events", zap.Error(err))
		return nil, err
	}

	return us.calculateSummary(events, from, to)
}

// calculateUptime tính toán uptime cho một infrastructure trong khoảng thời gian [from, to]
//
// LOGIC ĐƠN GIẢN:
// 1. Sắp xếp events theo thời gian
// 2. Với mỗi event, tính thời gian từ event trước đến event hiện tại
// 3. Nếu status trước đó là "up" -> cộng vào uptime, ngược lại cộng vào downtime
// 4. Cuối cùng, tính thời gian từ event cuối đến "to"
//
// FIX BUGS:
// - Bug #1: Thời gian từ "from" đến event đầu tiên giờ được tính (giả định down nếu chưa có status)
// - Bug #2: Nếu không có event, trả về unknown (đúng behavior)
// - Bug #3: Outage tracking hoạt động từ đầu period
func (us *uptimeService) calculateUptime(events []elasticsearch.UptimeEvent, instanceID string, from, to time.Time) (*dto.UptimeResponse, error) {
	// Tổng thời gian của period
	totalPeriodSeconds := int64(to.Sub(from).Seconds())

	// Không có events -> không biết trạng thái -> downtime = 100%
	if len(events) == 0 {
		return &dto.UptimeResponse{
			InfrastructureID: instanceID,
			TotalUptime:      0,
			TotalDowntime:    totalPeriodSeconds,
			UptimePercent:    0,
			CurrentStatus:    "unknown",
			Period:           formatPeriod(from, to),
			From:             from.Format(time.RFC3339),
			To:               to.Format(time.RFC3339),
		}, nil
	}

	// Sắp xếp events theo thời gian
	sort.Slice(events, func(i, j int) bool {
		return events[i].Timestamp.Before(events[j].Timestamp)
	})

	// Khởi tạo biến
	var totalUptime int64
	var totalDowntime int64
	var statusHistory []dto.StatusChange
	var outageEvents []dto.OutageEvent

	// Lấy thông tin infra từ event đầu tiên
	infraType := events[0].Type
	infraName := events[0].InstanceName
	userID := events[0].UserID

	// Bắt đầu từ thời điểm "from"
	lastEventTime := from
	lastStatus := "unknown" // Giả định ban đầu là unknown (sẽ tính vào downtime)
	var outageStartTime time.Time
	isInOutage := false

	// Duyệt qua từng event
	for _, event := range events {
		// Clamp eventTime vào trong khoảng [from, to]
		eventTime := event.Timestamp
		if eventTime.Before(from) {
			eventTime = from
		}
		if eventTime.After(to) {
			eventTime = to
		}

		// Tính thời gian từ event trước đến event hiện tại
		duration := int64(eventTime.Sub(lastEventTime).Seconds())

		// Cộng vào uptime hoặc downtime dựa trên status TRƯỚC ĐÓ
		if isStatusUp(lastStatus) {
			totalUptime += duration
		} else {
			// lastStatus là "unknown", "stopped", "failed", etc. -> downtime
			totalDowntime += duration
		}

		// Xác định status mới từ event
		newStatus := determineStatus(event.Action, event.Status)

		// Theo dõi outages
		if isStatusUp(lastStatus) && !isStatusUp(newStatus) {
			// Bắt đầu outage mới
			isInOutage = true
			outageStartTime = eventTime
		} else if !isStatusUp(lastStatus) && isStatusUp(newStatus) {
			// Kết thúc outage (nếu có)
			if isInOutage {
				outageEvents = append(outageEvents, dto.OutageEvent{
					StartTime: outageStartTime.Format(time.RFC3339),
					EndTime:   eventTime.Format(time.RFC3339),
					Duration:  int64(eventTime.Sub(outageStartTime).Seconds()),
					Reason:    lastStatus,
				})
				isInOutage = false
			}
		}

		// Ghi lại lịch sử thay đổi status
		statusHistory = append(statusHistory, dto.StatusChange{
			Timestamp:  event.Timestamp.Format(time.RFC3339),
			FromStatus: lastStatus,
			ToStatus:   newStatus,
			Action:     event.Action,
			Duration:   duration,
		})

		// Cập nhật cho vòng lặp tiếp theo
		lastStatus = newStatus
		lastEventTime = eventTime
	}

	// Tính thời gian còn lại từ event cuối đến "to"
	remainingDuration := int64(to.Sub(lastEventTime).Seconds())
	if isStatusUp(lastStatus) {
		totalUptime += remainingDuration
	} else {
		totalDowntime += remainingDuration
	}

	// Đóng outage nếu còn đang mở
	if isInOutage {
		outageEvents = append(outageEvents, dto.OutageEvent{
			StartTime: outageStartTime.Format(time.RFC3339),
			EndTime:   to.Format(time.RFC3339),
			Duration:  int64(to.Sub(outageStartTime).Seconds()),
			Reason:    "ongoing",
		})
	}

	// Tính phần trăm uptime
	// FIX: Sử dụng totalPeriodSeconds để đảm bảo đúng tổng thời gian
	var uptimePercent float64
	if totalPeriodSeconds > 0 {
		uptimePercent = float64(totalUptime) / float64(totalPeriodSeconds) * 100
	}

	return &dto.UptimeResponse{
		InfrastructureID:   instanceID,
		InfrastructureName: infraName,
		InfrastructureType: infraType,
		UserID:             userID,
		TotalUptime:        totalUptime,
		TotalDowntime:      totalDowntime,
		UptimePercent:      uptimePercent,
		CurrentStatus:      lastStatus,
		Period:             formatPeriod(from, to),
		From:               from.Format(time.RFC3339),
		To:                 to.Format(time.RFC3339),
		StatusHistory:      statusHistory,
		OutageEvents:       outageEvents,
	}, nil
}

func (us *uptimeService) calculateSummary(events []elasticsearch.UptimeEvent, from, to time.Time) (*dto.UptimeSummaryResponse, error) {
	// Group events by instance
	instanceEvents := make(map[string][]elasticsearch.UptimeEvent)
	instanceTypes := make(map[string]string)

	for _, event := range events {
		instanceEvents[event.InstanceID] = append(instanceEvents[event.InstanceID], event)
		if event.Type != "" {
			instanceTypes[event.InstanceID] = event.Type
		}
	}

	byType := make(map[string]*dto.TypeUptime)
	infrastructures := []dto.UptimeResponse{}
	var totalUptimePercent float64

	for instanceID, evts := range instanceEvents {
		uptime, err := us.calculateUptime(evts, instanceID, from, to)
		if err != nil {
			continue
		}
		infrastructures = append(infrastructures, *uptime)
		totalUptimePercent += uptime.UptimePercent

		// Aggregate by type
		infraType := instanceTypes[instanceID]
		if infraType == "" {
			infraType = "unknown"
		}

		if byType[infraType] == nil {
			byType[infraType] = &dto.TypeUptime{
				Type: infraType,
			}
		}

		byType[infraType].Count++
		byType[infraType].TotalUptime += uptime.TotalUptime
		byType[infraType].TotalDowntime += uptime.TotalDowntime

		if uptime.CurrentStatus == "running" || uptime.CurrentStatus == "healthy" {
			byType[infraType].ActiveCount++
		}
	}

	// Calculate average uptime per type
	for _, typeUptime := range byType {
		totalPeriod := typeUptime.TotalUptime + typeUptime.TotalDowntime
		if totalPeriod > 0 {
			typeUptime.AverageUptime = float64(typeUptime.TotalUptime) / float64(totalPeriod) * 100
		}
	}

	// Convert byType map to final format
	byTypeResult := make(map[string]dto.TypeUptime)
	for k, v := range byType {
		byTypeResult[k] = *v
	}

	// Sort infrastructures by uptime percent
	sort.Slice(infrastructures, func(i, j int) bool {
		return infrastructures[i].UptimePercent > infrastructures[j].UptimePercent
	})

	// Get top and worst performers
	var topPerformers, worstPerformers []dto.UptimeResponse
	if len(infrastructures) > 0 {
		topCount := 5
		if len(infrastructures) < topCount {
			topCount = len(infrastructures)
		}
		topPerformers = infrastructures[:topCount]

		// Reverse sort for worst
		worstInfras := make([]dto.UptimeResponse, len(infrastructures))
		copy(worstInfras, infrastructures)
		sort.Slice(worstInfras, func(i, j int) bool {
			return worstInfras[i].UptimePercent < worstInfras[j].UptimePercent
		})
		if len(worstInfras) < topCount {
			topCount = len(worstInfras)
		}
		worstPerformers = worstInfras[:topCount]
	}

	// Calculate overall average
	var avgUptime float64
	if len(infrastructures) > 0 {
		avgUptime = totalUptimePercent / float64(len(infrastructures))
	}

	return &dto.UptimeSummaryResponse{
		Period:          formatPeriod(from, to),
		From:            from.Format(time.RFC3339),
		To:              to.Format(time.RFC3339),
		TotalInfra:      len(infrastructures),
		AverageUptime:   avgUptime,
		ByType:          byTypeResult,
		TopPerformers:   topPerformers,
		WorstPerformers: worstPerformers,
		Infrastructures: infrastructures,
	}, nil
}

// Helper functions

func isStatusUp(status string) bool {
	upStatuses := map[string]bool{
		"running": true,
		"healthy": true,
		"started": true,
		"created": true,
		"active":  true,
	}
	return upStatuses[status]
}

func determineStatus(action, status string) string {
	if status != "" {
		return status
	}

	actionToStatus := map[string]string{
		"created":   "running",
		"started":   "running",
		"stopped":   "stopped",
		"deleted":   "deleted",
		"failed":    "failed",
		"healthy":   "running",
		"unhealthy": "failed",
		"start":     "running",
		"stop":      "stopped",
		"create":    "running",
		"delete":    "deleted",
	}

	if s, ok := actionToStatus[action]; ok {
		return s
	}
	return action
}

func formatPeriod(from, to time.Time) string {
	duration := to.Sub(from)
	hours := duration.Hours()

	if hours <= 1 {
		return "1h"
	} else if hours <= 24 {
		return "24h"
	} else if hours <= 168 { // 7 days
		return "7d"
	} else if hours <= 720 { // 30 days
		return "30d"
	}
	return "custom"
}

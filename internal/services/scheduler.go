package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/leofvo/bridgr/internal/domain"
	"github.com/leofvo/bridgr/pkg/logger"
)

// SchedulerService manages the scheduling of source polling
type SchedulerService struct {
	notificationService *NotificationService
	sources            []domain.Source
	exporters          []domain.Exporter
	wg                 sync.WaitGroup
}

// NewSchedulerService creates a new scheduler service
func NewSchedulerService(notificationService *NotificationService, sources []domain.Source, exporters []domain.Exporter) *SchedulerService {
	return &SchedulerService{
		notificationService: notificationService,
		sources:            sources,
		exporters:          exporters,
	}
}

// Start starts the scheduler
func (s *SchedulerService) Start(ctx context.Context) error {
	for _, source := range s.sources {
		s.wg.Add(1)
		go func(source domain.Source) {
			defer s.wg.Done()
			s.scheduleSource(ctx, source)
		}(source)
	}

	return nil
}

// Stop stops the scheduler
func (s *SchedulerService) Stop() {
	s.wg.Wait()
}

// scheduleSource schedules a source for polling
func (s *SchedulerService) scheduleSource(ctx context.Context, source domain.Source) {
	ticker := time.NewTicker(source.GetInterval())
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.pollSource(ctx, source); err != nil {
				logger.Error("Failed to poll source: source=%s error=%v", source.GetType(), err)
			}
		}
	}
}

// pollSource polls a source for new items
func (s *SchedulerService) pollSource(ctx context.Context, source domain.Source) error {
	items, err := source.Fetch()
	if err != nil {
		return fmt.Errorf("failed to fetch items: %w", err)
	}

	if len(items) == 0 {
		return nil
	}

	return s.notificationService.ProcessItems(ctx, items, s.exporters)
} 
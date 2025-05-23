package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/leofvo/bridgr/internal/domain"
	"github.com/leofvo/bridgr/pkg/logger"
)

// NotificationService handles the notification processing
type NotificationService struct {
	store domain.Store
}

// NewNotificationService creates a new notification service
func NewNotificationService(store domain.Store) *NotificationService {
	return &NotificationService{
		store: store,
	}
}

// ProcessItems processes items and sends notifications
func (s *NotificationService) ProcessItems(ctx context.Context, items []domain.Item, exporters []domain.Exporter) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(items)*len(exporters))

	for _, item := range items {
		// Filter exporters by group
		groupExporters := make([]domain.Exporter, 0)
		for _, exporter := range exporters {
			if exporter.GetGroup() == item.Group {
				groupExporters = append(groupExporters, exporter)
			}
		}

		for _, exporter := range groupExporters {
			wg.Add(1)
			go func(item domain.Item, exporter domain.Exporter) {
				defer wg.Done()

				// Check if item has been processed
				processed, err := s.store.HasProcessed(item.ID, exporter.GetType())
				if err != nil {
					errChan <- fmt.Errorf("failed to check if item was processed: item=%s exporter=%s error=%w", item.ID, exporter.GetType(), err)
					return
				}

				if processed {
					logger.Debug("Item already processed: item=%s exporter=%s", item.ID, exporter.GetType())
					return
				}

				// Send notification
				if err := exporter.Export(item); err != nil {
					errChan <- fmt.Errorf("failed to export item: item=%s exporter=%s error=%w", item.ID, exporter.GetType(), err)
					return
				}

				// Get source TTL if available
				var sourceTTL *time.Duration
				if source, ok := exporter.(interface{ GetSourceTTL() *time.Duration }); ok {
					sourceTTL = source.GetSourceTTL()
				}

				// Mark as processed
				if err := s.store.MarkProcessed(item.ID, exporter.GetType(), sourceTTL); err != nil {
					errChan <- fmt.Errorf("failed to mark item as processed: item=%s exporter=%s error=%w", item.ID, exporter.GetType(), err)
					return
				}

				logger.Info("Processed item: item=%s exporter=%s", item.ID, exporter.GetType())
			}(item, exporter)
		}
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(errChan)

	// Collect errors
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
		logger.Error("Error processing item: %v", err)
	}

	if len(errors) > 0 {
		return fmt.Errorf("encountered %d errors while processing items: %v", len(errors), errors)
	}

	return nil
} 
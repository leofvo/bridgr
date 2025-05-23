package exporters

import (
	"fmt"

	"github.com/leofvo/bridgr/internal/config"
	"github.com/leofvo/bridgr/internal/domain"
)

// Factory creates new exporter instances
type Factory struct{}

// NewFactory creates a new exporter factory
func NewFactory() *Factory {
	return &Factory{}
}

// CreateExporter creates a new exporter based on the configuration
func (f *Factory) CreateExporter(cfg *config.ExporterConfig, group string) (domain.Exporter, error) {
	switch cfg.Type {
	case "webhook":
		return NewWebhookExporter(cfg, group), nil
	default:
		return nil, fmt.Errorf("unknown exporter type: %s", cfg.Type)
	}
} 
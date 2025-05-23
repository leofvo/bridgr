package sources

import (
	"fmt"

	"github.com/leofvo/bridgr/internal/config"
	"github.com/leofvo/bridgr/internal/domain"
)

// Factory creates new source instances
type Factory struct{}

// NewFactory creates a new source factory
func NewFactory() *Factory {
	return &Factory{}
}

// CreateSource creates a new source based on the configuration
func (f *Factory) CreateSource(cfg *config.SourceConfig, group string) (domain.Source, error) {
	switch cfg.Type {
	case "rss":
		return NewRSSSource(cfg, group), nil
	default:
		return nil, fmt.Errorf("unknown source type: %s", cfg.Type)
	}
} 
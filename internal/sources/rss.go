package sources

import (
	"fmt"
	"time"

	"github.com/leofvo/bridgr/internal/config"
	"github.com/leofvo/bridgr/internal/domain"
	"github.com/leofvo/bridgr/internal/utils"
	"github.com/leofvo/bridgr/pkg/logger"
	"github.com/mmcdole/gofeed"
)

// RSSSource implements the Source interface for RSS feeds
type RSSSource struct {
	config  *config.SourceConfig
	parser  *gofeed.Parser
	group   string
	lastRun time.Time
}

// NewRSSSource creates a new RSS source
func NewRSSSource(cfg *config.SourceConfig, group string) *RSSSource {
	return &RSSSource{
		config: cfg,
		parser: gofeed.NewParser(),
		group:  group,
	}
}

// Fetch retrieves items from the RSS feed
func (s *RSSSource) Fetch() ([]domain.Item, error) {
	feed, err := s.parser.ParseURL(s.config.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RSS feed: url=%s error=%w", s.config.URL, err)
	}

	items := make([]domain.Item, 0, len(feed.Items))
	for _, item := range feed.Items {
		// Skip items older than last run
		if item.PublishedParsed != nil && s.lastRun.After(*item.PublishedParsed) {
			continue
		}

		publishedAt := time.Now()
		if item.PublishedParsed != nil {
			publishedAt = *item.PublishedParsed
		}

		// Validate required fields
		if item.GUID == "" {
			logger.Warn("Skipping item with empty GUID: url=%s title=%s", s.config.URL, item.Title)
			continue
		}

		if item.Title == "" {
			logger.Warn("Skipping item with empty title: url=%s guid=%s", s.config.URL, item.GUID)
			continue
		}

		// Clean HTML content
		title := utils.StripHTML(item.Title)
		description := utils.StripHTML(item.Description)

		items = append(items, domain.Item{
			ID:          item.GUID,
			Title:       title,
			Description: description,
			Link:        item.Link,
			PublishedAt: publishedAt,
			Source:      s.config.URL,
			Group:       s.group,
		})
	}

	s.lastRun = time.Now()
	logger.Info("Fetched RSS feed: url=%s items=%d", s.config.URL, len(items))
	return items, nil
}

// GetType returns the source type
func (s *RSSSource) GetType() string {
	return "rss"
}

// GetInterval returns the polling interval
func (s *RSSSource) GetInterval() time.Duration {
	return s.config.Interval
}

// GetGroup returns the group name
func (s *RSSSource) GetGroup() string {
	return s.group
}

// GetSourceTTL returns the source-specific TTL if configured
func (s *RSSSource) GetSourceTTL() *time.Duration {
	if s.config.TTL > 0 {
		return &s.config.TTL
	}
	return nil
} 
package domain

import "time"

// Item represents a feed item from any source
type Item struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Link        string    `json:"link"`
	PublishedAt time.Time `json:"published_at"`
	Source      string    `json:"source"`
	Group       string    `json:"group"`
}

// Group represents a collection of sources and exporters
type Group struct {
	Name      string     `json:"name"`
	Sources   []Source   `json:"sources"`
	Exporters []Exporter `json:"exporters"`
}

// Source represents a data source (e.g., RSS feed)
type Source interface {
	Fetch() ([]Item, error)
	GetType() string
	GetInterval() time.Duration
	GetGroup() string
}

// Exporter represents a notification target
type Exporter interface {
	Export(item Item) error
	GetType() string
	GetGroup() string
}

// Store represents the data persistence layer
type Store interface {
	HasProcessed(itemID, exporterID string) (bool, error)
	MarkProcessed(itemID, exporterID string, sourceTTL *time.Duration) error
	Close() error
}

// Notification represents a processed notification
type Notification struct {
	Item      Item      `json:"item"`
	Exporter  string    `json:"exporter"`
	Group     string    `json:"group"`
	CreatedAt time.Time `json:"created_at"`
} 
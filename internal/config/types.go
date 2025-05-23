package config

import "time"

// Config represents the root configuration structure
type Config struct {
	Groups []GroupConfig `yaml:"groups"`
	Redis  RedisConfig   `yaml:"redis"`
	Server ServerConfig  `yaml:"server"`
}

// GroupConfig represents a group configuration
type GroupConfig struct {
	Name      string           `yaml:"name"`
	Sources   []SourceConfig   `yaml:"sources"`
	Exporters []ExporterConfig `yaml:"exporters"`
}

// SourceConfig represents a source configuration
type SourceConfig struct {
	Type     string        `yaml:"type"`
	URL      string        `yaml:"url"`
	Interval time.Duration `yaml:"interval"`
	TTL      time.Duration `yaml:"ttl,omitempty"`
}

// ExporterConfig represents an exporter configuration
type ExporterConfig struct {
	Type       string                 `yaml:"type"`
	Value      string                 `yaml:"value"`
	Options    map[string]interface{} `yaml:"options"`
	RateLimit  *RateLimitConfig      `yaml:"rate_limit,omitempty"`
}

// RateLimitConfig represents rate limiting configuration
type RateLimitConfig struct {
	RequestsPerSecond float64 `yaml:"requests_per_second"`
}

// RedisConfig represents Redis connection configuration
type RedisConfig struct {
	Address  string        `yaml:"address" env:"REDIS_ADDRESS"`
	Password string        `yaml:"password" env:"REDIS_PASSWORD"`
	DB       int          `yaml:"db" env:"REDIS_DB"`
	TTL      time.Duration `yaml:"ttl" env:"REDIS_TTL"`
}

// ServerConfig represents HTTP server configuration
type ServerConfig struct {
	Port int `yaml:"port"`
} 
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/spf13/viper"
)

// LoadConfig loads the configuration from the specified path and environment variables
func LoadConfig(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Override with environment variables
	overrideFromEnv(&config)

	// Set defaults
	if config.Server.Port == 0 {
		config.Server.Port = 8080
	}

	if config.Redis.TTL == 0 {
		config.Redis.TTL = 7 * 24 * time.Hour // 7 days default TTL
	}

	return &config, nil
}

// overrideFromEnv overrides configuration values with environment variables
func overrideFromEnv(config *Config) {
	// Override Redis config
	if addr := os.Getenv("REDIS_ADDRESS"); addr != "" {
		config.Redis.Address = addr
	}
	if pass := os.Getenv("REDIS_PASSWORD"); pass != "" {
		config.Redis.Password = pass
	}
	if db := os.Getenv("REDIS_DB"); db != "" {
		if dbNum, err := strconv.Atoi(db); err == nil {
			config.Redis.DB = dbNum
		}
	}
	if ttl := os.Getenv("REDIS_TTL"); ttl != "" {
		if duration, err := time.ParseDuration(ttl); err == nil {
			config.Redis.TTL = duration
		}
	}
}

// ValidateConfig validates the configuration
func ValidateConfig(config *Config) error {
	if len(config.Groups) == 0 {
		return fmt.Errorf("no groups configured")
	}

	for _, group := range config.Groups {
		if group.Name == "" {
			return fmt.Errorf("group name cannot be empty")
		}

		if len(group.Sources) == 0 {
			return fmt.Errorf("group %s has no sources", group.Name)
		}

		if len(group.Exporters) == 0 {
			return fmt.Errorf("group %s has no exporters", group.Name)
		}

		for _, source := range group.Sources {
			if source.Type == "" {
				return fmt.Errorf("source type cannot be empty in group %s", group.Name)
			}
			if source.URL == "" {
				return fmt.Errorf("source URL cannot be empty in group %s", group.Name)
			}
			if source.Interval == 0 {
				return fmt.Errorf("source interval cannot be zero in group %s", group.Name)
			}
		}

		for _, exporter := range group.Exporters {
			if exporter.Type == "" {
				return fmt.Errorf("exporter type cannot be empty in group %s", group.Name)
			}
			if exporter.Value == "" {
				return fmt.Errorf("exporter value cannot be empty in group %s", group.Name)
			}
		}
	}

	return nil
} 
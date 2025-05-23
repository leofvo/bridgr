package exporters

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/leofvo/bridgr/internal/config"
	"github.com/leofvo/bridgr/internal/domain"
	"github.com/leofvo/bridgr/internal/ratelimit"
	"github.com/leofvo/bridgr/pkg/logger"
)

// WebhookExporter implements the Exporter interface for webhooks
type WebhookExporter struct {
	config  *config.ExporterConfig
	client  *http.Client
	group   string
	limiter *ratelimit.Limiter
}

// DiscordWebhook represents a Discord webhook payload
type DiscordWebhook struct {
	Content string         `json:"content,omitempty"`
	Embeds  []DiscordEmbed `json:"embeds,omitempty"`
}

// DiscordEmbed represents a Discord embed
type DiscordEmbed struct {
	Title       string         `json:"title"`
	Description string         `json:"description"`
	URL         string         `json:"url"`
	Color       int            `json:"color"`
	Timestamp   string         `json:"timestamp"`
	Footer      *DiscordFooter `json:"footer,omitempty"`
}

// DiscordFooter represents a Discord embed footer
type DiscordFooter struct {
	Text string `json:"text"`
}

// TeamsWebhook represents a Microsoft Teams webhook payload
type TeamsWebhook struct {
	Type        string           `json:"type"`
	Attachments []TeamsAttachment `json:"attachments"`
}

// TeamsAttachment represents a Teams message attachment
type TeamsAttachment struct {
	ContentType string      `json:"contentType"`
	ContentURL  interface{} `json:"contentUrl"`
	Content     TeamsContent `json:"content"`
}

// TeamsContent represents the content of a Teams message
type TeamsContent struct {
	Schema  string        `json:"$schema"`
	Type    string        `json:"type"`
	Version string        `json:"version"`
	Body    []TeamsBlock  `json:"body"`
}

// TeamsBlock represents a block in a Teams message
type TeamsBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
	Size string `json:"size,omitempty"`
	Weight string `json:"weight,omitempty"`
	Color string `json:"color,omitempty"`
}

// NewWebhookExporter creates a new webhook exporter
func NewWebhookExporter(cfg *config.ExporterConfig, group string) *WebhookExporter {
	// Create rate limiter if configured
	var limiter *ratelimit.Limiter
	if cfg.RateLimit != nil {
		limiter = ratelimit.NewLimiter(cfg.RateLimit.RequestsPerSecond)
	}

	return &WebhookExporter{
		config:  cfg,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		group:   group,
		limiter: limiter,
	}
}

// Export sends an item to the webhook
func (e *WebhookExporter) Export(item domain.Item) error {
	// Wait for rate limit if configured
	if e.limiter != nil {
		e.limiter.Wait()
	}

	var payload interface{}

	// Check the webhook format
	if format, ok := e.config.Options["format"].(string); ok {
		switch format {
		case "discord":
			payload = e.createDiscordPayload(item)
		case "teams":
			payload = e.createTeamsPayload(item)
		default:
			// Default to simple JSON payload
			payload = item
		}
	} else {
		// Default to simple JSON payload
		payload = item
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: item=%s error=%w", item.ID, err)
	}

	req, err := http.NewRequest("POST", e.config.Value, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: item=%s error=%w", item.ID, err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := e.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook: item=%s url=%s error=%w", item.ID, e.config.Value, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		// Check for Discord rate limit response
		if resp.StatusCode == 429 {
			var rateLimitResp struct {
				Message    string  `json:"message"`
				RetryAfter float64 `json:"retry_after"`
				Global     bool    `json:"global"`
			}
			if err := json.Unmarshal(body, &rateLimitResp); err == nil {
				// Wait for the specified retry_after duration
				retryDuration := time.Duration(rateLimitResp.RetryAfter * float64(time.Second))
				logger.Debug("Rate limit hit: waiting for %v before retry", retryDuration)
				time.Sleep(retryDuration)
				// Retry the request
				return e.Export(item)
			}
		}
		return fmt.Errorf("webhook request failed: item=%s url=%s status=%d body=%s", item.ID, e.config.Value, resp.StatusCode, string(body))
	}

	logger.Info("Sent webhook: url=%s item=%s", e.config.Value, item.ID)
	return nil
}

// GetType returns the exporter type
func (e *WebhookExporter) GetType() string {
	return "webhook"
}

// GetGroup returns the group name
func (e *WebhookExporter) GetGroup() string {
	return e.group
}

// createDiscordPayload creates a Discord webhook payload
func (e *WebhookExporter) createDiscordPayload(item domain.Item) DiscordWebhook {
	// Extract domain from source URL
	sourceDomain := "Unknown Source"
	if parsedURL, err := url.Parse(item.Source); err == nil {
		sourceDomain = strings.TrimPrefix(parsedURL.Hostname(), "www.")
	}

	return DiscordWebhook{
		Embeds: []DiscordEmbed{
			{
				Title:       item.Title,
				Description: item.Description,
				URL:         item.Link,
				Color:       3447003, // Blue color
				Timestamp:   item.PublishedAt.Format(time.RFC3339),
				Footer: &DiscordFooter{
					Text: fmt.Sprintf("Source: %s", sourceDomain),
				},
			},
		},
	}
}

// createTeamsPayload creates a Microsoft Teams webhook payload
func (e *WebhookExporter) createTeamsPayload(item domain.Item) TeamsWebhook {
	// Extract domain from source URL
	sourceDomain := "Unknown Source"
	if parsedURL, err := url.Parse(item.Source); err == nil {
		sourceDomain = strings.TrimPrefix(parsedURL.Hostname(), "www.")
	}

	return TeamsWebhook{
		Type: "message",
		Attachments: []TeamsAttachment{
			{
				ContentType: "application/vnd.microsoft.card.adaptive",
				ContentURL:  nil,
				Content: TeamsContent{
					Schema:  "http://adaptivecards.io/schemas/adaptive-card.json",
					Type:    "AdaptiveCard",
					Version: "1.2",
					Body: []TeamsBlock{
						{
							Type:   "TextBlock",
							Text:   item.Title,
							Size:   "Large",
							Weight: "Bolder",
						},
						{
							Type: "TextBlock",
							Text: item.Description,
						},
						{
							Type:  "TextBlock",
							Text:  fmt.Sprintf("[Read more](%s)", item.Link),
							Color: "Accent",
						},
						{
							Type: "TextBlock",
							Text: fmt.Sprintf("Source: %s", sourceDomain),
							Size: "Small",
						},
					},
				},
			},
		},
	}
} 
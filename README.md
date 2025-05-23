# Bridgr

Bridgr is a real-time notification service for RSS feeds deployed on Kubernetes. It monitors RSS feeds periodically, processes new items, and sends notifications via webhooks with deduplication support.

## Features

- Monitor multiple RSS feeds with configurable polling intervals
- Send notifications via webhooks (with Discord support)
- Deduplicate notifications (one notification per item per exporter)
- YAML configuration for flexible setup
- Support for multiple groups with their own sources and exporters
- Health check endpoint for monitoring
- Kubernetes deployment support
- Redis-based state management

## Architecture

Bridgr follows clean architecture principles with clear separation of concerns:

- **Factory Pattern** for Sources and Exporters
- **Port/Adapter Pattern** for data store interactions
- **Strategy Pattern** for different notification formats
- **Repository Pattern** for state management

## Prerequisites

- Go 1.23 or later
- Redis server
- Docker (for containerization)
- Kubernetes cluster (for deployment)

## Configuration

Bridgr uses a YAML configuration file located at `/etc/bridgr/config.yaml`. Here's an example configuration:

```yaml
groups:
  - name: "tech-news"
    sources:
      - type: "rss"
        url: "https://example.com/feed.xml"
        interval: "5m"
    exporters:
      - type: "webhook"
        value: "https://discord.com/api/webhooks/..."
        options:
          format: "discord"
        rate_limit:
          requests_per_second: 2.0
  - name: "security-alerts"
    sources:
      - type: "rss"
        url: "https://security.example.com/feed.xml"
        interval: "1m"
    exporters:
      - type: "webhook"
        value: "https://xxx.webhook.office.com/webhookb2/..."
        options:
          format: "teams"
          
server:
  port: 8080
```

## Development

1. Clone the repository:

    ```bash
    git clone https://github.com/leofvo/bridgr.git
    cd bridgr
    ```

2. Build the application with hot-reloading:

    ```bash
    docker compose -f docker-compose.yml up --build
    ```

## Docker

Build the Docker image:

```bash
docker build -t ghcr.io/leofvo/bridgr:latest .
```

## Health Check

Bridgr exposes a health check endpoint at `/health`. The endpoint returns a JSON response with the current status and timestamp:

```json
{
  "status": "ok",
  "timestamp": "2024-02-20T12:00:00Z"
}
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'feat: add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

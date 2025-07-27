# IP Geo Service

A high-performance IP geolocation service that validates IP addresses against allowed countries. Built with Go, it provides HTTP and gRPC APIs, automatic database updates, comprehensive monitoring, and a web UI for testing.

## Features

- REST HTTP and gRPC APIs
- Web UI for testing and monitoring
- Demo mode for testing without real GeoIP data
- Concurrent request handling
- Automatic MaxMind GeoLite2 database updates
- Comprehensive logging, metrics, and health checks
- Docker and Kubernetes ready
- Security best practices

## Quick Start

### Prerequisites

- Go 1.21 or higher
- Docker (optional)
- MaxMind license key (for production use)

### Running Locally

1. Clone the repository:
```bash
git clone https://github.com/yourusername/geo-service.git
cd geo-service
```

2. Download the GeoIP database:
```bash
# For testing (downloads sample database)
./scripts/download-geodb.sh

# For production (requires MaxMind license)
./scripts/download-geodb.sh YOUR_MAXMIND_LICENSE_KEY
```

3. Run the service:
```bash
go run cmd/server/main.go
```

4. Open the web UI:
```
http://localhost:8080
```

### Using Docker

```bash
# Build the image
docker build -t geo-service:latest .

# Run the container
docker run -p 8080:8080 -p 9090:9090 \
  -e MAXMIND_LICENSE_KEY=your_license_key \
  geo-service:latest
```

## Web UI

The service includes a web interface for testing:

- Test IP validation with different countries
- Demo mode simulates real GeoIP data without a license
- View request logs and statistics
- Quick presets for common scenarios

Access the UI at `http://localhost:8080` after starting the service.

## API Documentation

### REST API

#### Validate IP Address

```http
POST /api/v1/validate
Content-Type: application/json

{
  "ip": "1.2.3.4",
  "allowed_countries": ["US", "CA", "GB"]
}
```

Response:
```json
{
  "allowed": true,
  "country": "US",
  "ip": "1.2.3.4"
}
```

#### Health Check

```http
GET /health
```

### gRPC API

See `pkg/proto/geo.proto` for the gRPC service definition.

## Configuration

Environment variables:

- `HTTP_PORT` - HTTP server port (default: 8080)
- `GRPC_PORT` - gRPC server port (default: 9090)
- `DB_PATH` - Path to GeoIP database (default: ./data/GeoLite2-Country.mmdb)
- `UPDATE_INTERVAL` - Database update interval (default: 24h)
- `MAXMIND_LICENSE_KEY` - MaxMind license key for database updates
- `LOG_LEVEL` - Logging level: debug, info, warn, error (default: info)
- `DEMO_MODE` - Enable demo mode for testing (default: false)

## Development

### Building

```bash
make build
```

### Testing

```bash
make test
```

### Generate gRPC Code

```bash
make proto
```

## Deployment

### Kubernetes

```bash
kubectl apply -f deployments/kubernetes/
```

### Docker Compose

```bash
docker-compose up
```

## Monitoring

The service exposes Prometheus metrics at `/metrics`:

- `geo_requests_total` - Total number of requests
- `geo_request_duration_seconds` - Request duration histogram
- `geo_allowed_total` - Total allowed requests
- `geo_blocked_total` - Total blocked requests

## License

MIT License
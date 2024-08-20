## Code Structure

```bash
.
├── Dockerfile
├── config
│   └── config.go
├── go.mod
├── go.sum
├── handlers
│   └── cert_handler.go
├── main.go
├── metrics
│   └── metrics.go
├── redisclient
│   └── redis_client.go
├── tracing
│   └── tracing.go
├── ulfflogger
│   └── ulfflogger.go
└── utils
    └── cert_utils.go
```

- **`config/config.go`**: Manages the configuration loading via environment variables using viper. It supports custom configurations for Redis connections, circuit breakers, and TLS versions.
- **`handlers/cert_handler.go`**: Contains the core logic for certificate processing, including handling HTTP requests, domain validation, retrieving certificate details, and interfacing with Redis.
- **`metrics/metrics.go`**: Handles Prometheus metrics, exposing the certificate expiration details for domain monitoring.
- **`redisclient/redis_client.go`**: Implements the Redis client for caching certificate data and keeping track of processed domains.
- **`tracing/tracing.go`**: Implements OpenTelemetry distributed tracing for tracking and monitoring requests and external dependencies.
- **`ulfflogger/ulfflogger.go`**: Provides structured logging for better observability and debugging, encapsulating logs with relevant context and metadata.
- **`utils/cert_utils.go`**: Utility functions to validate domain names, retrieve SSL certificate details, and calculate expiration dates.

## API Endpoints

### `POST /check`
**Description:**: Accepts a JSON array of domains, processes each domain, and returns certificate details for each domain.

Request:

- Body: JSON array of domains.
- Method: `POST`

```json
{
    "domains": ["example.com", "test.com"]
}
```

Response:

- Status Code: 200
- Body: JSON array of certificate details.

```json
{
    "results": [
        {
            "domain": "example.com",
            "expiry_date": "2022-12-31T23:59:59Z",
            "issuer": "Let's Encrypt",
            "days_until_expiry": 365
        },
        {
            "domain": "test.com",
            "expiry_date": "2022-12-31T23:59:59Z",
            "issuer": "DigiCert",
            "days_until_expiry": 365
        }
    ]
}
```

### `GET /check?domain={domain}`

**Description:** Fetches certificate information for a single domain using a query parameter.

Request:

- Query Parameter: `domain`
- Method: `GET`

```plaintext
GET /check?domain=example.com
```

Response:

- Status Code: 200
- Body: JSON object with certificate details.

```json
{
    "domain": "example.com",
    "expiry_date": "2022-12-31T23:59:59Z",
    "issuer": "Let's Encrypt",
    "days_until_expiry": 365
}
```

### `GET /metrics`

- **Description:** Exposes Prometheus metrics.
- **Response:** Returns Prometheus metrics in plain text format.

### `GET /healthz`

- **Description:** Health check endpoint.
- **Response:** Returns a `200 OK` status code if the service is healthy.

### `GET /readiness`

- **Description:** Readiness check endpoint.
- **Response:** Returns a `200 OK`  if Redis is connected, otherwise `503 Service Unavailable`.

## Error Handling

Cert-Tracker utilizes structured error handling mechanisms to ensure that errors are logged effectively and clients receive meaningful error responses. The error responses follow a consistent structure:

- **Invalid Input:** Returns` 400 Bad Request` for malformed requests, invalid JSON, or missing domains.
- **Rate Limiting:** Returns `429 Too Many Requests` if the rate limit is exceeded.
- **Method Not Allowed:** Returns `405 Method Not Allowed` for unsupported HTTP methods.
- **Internal Errors:** Returns `500 Internal Server Error` for unexpected issues, including failure to retrieve certificate information or Redis errors.

Each error response follows a consistent schema:

```json
{
  "code": "error_code",
  "message": "Error description"
}
```

## Usage Examples

Example POST /check Request

```bash
curl -X POST http://localhost:8080/check \
     -H "Content-Type: application/json" \
     -d '["example.com", "anotherdomain.com"]'
```

Example GET /check Request

```bash
curl -X GET http://localhost:8080/check?domain=example.com
```

## Security Considerations

Cert-Tracker implements several key security mechanisms to ensure the safety and integrity of the system:

- **Rate Limiting:** Protects the service from abuse by limiting the rate at which clients can send requests. Implemented using Go's `x/time/rate` package.
- **Circuit Breaker:** Utilizes the `sony/gobreaker` package to prevent cascading failures by managing Redis operations. When Redis fails or becomes unresponsive, the circuit breaker opens to protect the system.
- **TLS:** The service mandates a minimum TLS version (configurable via environment variables) when fetching SSL certificate details from domains.
- **Security Headers:** The service enforces several HTTP security headers:
  - `Strict-Transport-Security`
  - `X-Frame-Options`
  - `X-XSS-Protection`
  - `X-Content-Type-Options`
- **Input Validation:** Before processing, domain inputs are validated using regular expressions to prevent invalid domains from entering the system.
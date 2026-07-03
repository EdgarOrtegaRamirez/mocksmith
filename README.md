# MockSmith

A config-driven HTTP mock server for API development, testing, and prototyping. Define routes in YAML or JSON, and MockSmith serves realistic responses with matching, delays, and hot-reload.

## Features

- **YAML/JSON Configuration** — Define mock routes in human-readable config files
- **Path Parameter Matching** — Use `:param` syntax for dynamic path segments (`/users/:id`)
- **Wildcard Matching** — Use `*` to match any remaining path
- **Query Parameter Matching** — Match requests by query string values
- **Header Matching** — Match requests by HTTP headers
- **Body Matching** — Match by exact body, substring, or JSON field values
- **Response Templates** — Dynamic responses with path parameter substitution
- **Request Delays** — Simulate network latency with configurable min/max delays
- **CORS Support** — Configurable Cross-Origin Resource Sharing headers
- **Hot Reload** — Watch config file for changes and reload automatically
- **HTTPS Support** — Serve over TLS with custom certificates
- **Multiple Responses** — Define multiple responses per route (future: scenario switching)
- **Request Logging** — Colored text or JSON structured logging
- **Single Binary** — No runtime dependencies, just download and run

## Quick Start

### Install

```bash
go install github.com/EdgarOrtegaRamirez/mocksmith/cmd/mocksmith@latest
```

### Generate a Sample Config

```bash
mocksmith sample > mock.yaml
```

### Start the Server

```bash
mocksmith serve mock.yaml
```

### Test It

```bash
curl http://localhost:3000/api/users
```

## Usage

### Serve

```bash
# Start with a config file
mocksmith serve config.yaml

# Start with file watching (auto-reload on changes)
mocksmith serve --watch config.yaml

# Verbose logging
mocksmith serve -v config.yaml

# JSON structured logs
mocksmith serve --json-log config.yaml
```

### Validate

```bash
# Check if a config file is valid
mocksmith validate config.yaml
```

### Routes

```bash
# List all routes in a table
mocksmith routes config.yaml

# List routes as JSON
mocksmith routes --json config.yaml
```

### Sample

```bash
# Generate a sample config file
mocksmith sample > mock.yaml
```

## Configuration

### Server Config

```yaml
server:
  host: "0.0.0.0"
  port: 3000
  log_level: info
```

### Routes

```yaml
routes:
  - name: List Users
    method: GET
    path: /api/users
    responses:
      - status_code: 200
        content_type: application/json
        body: |
          [{"id": 1, "name": "Alice"}, {"id": 2, "name": "Bob"}]
```

### Path Parameters

```yaml
routes:
  - name: Get User
    method: GET
    path: /api/users/:id
    responses:
      - status_code: 200
        body: '{"id": ":id", "name": "User :id"}'
```

Parameters are extracted from the URL and substituted into the response body.

### Query Parameters

```yaml
routes:
  - name: Search
    method: GET
    path: /api/search
    query:
      q: test
      limit: "10"
    responses:
      - status_code: 200
        body: '{"results": []}'
```

### Headers

```yaml
routes:
  - name: Auth Required
    method: GET
    path: /api/protected
    headers:
      Authorization: "Bearer my-token"
    responses:
      - status_code: 200
        body: '{"authorized": true}'
```

### Body Matching

```yaml
routes:
  - name: Webhook
    method: POST
    path: /api/webhook
    body:
      contains: '"event":"push"'  # Substring match
    responses:
      - status_code: 200
        body: '{"received": true}'

  - name: Exact Body
    method: POST
    path: /api/exact
    body:
      exact: '{"key":"value"}'
    responses:
      - status_code: 200
```

### Delays

```yaml
routes:
  - name: Slow API
    method: GET
    path: /api/slow
    delay:
      min_ms: 1000
      max_ms: 3000
    responses:
      - status_code: 200
        body: '{"status": "ok"}'
```

### CORS

```yaml
cors:
  allow_origins: ["*"]
  allow_methods: ["GET", "POST", "PUT", "DELETE"]
  allow_headers: ["Content-Type", "Authorization"]
  expose_headers: ["X-Request-Id"]
  allow_credentials: true
  max_age: 3600
```

### Multiple Responses

```yaml
routes:
  - name: Get User
    method: GET
    path: /api/users/:id
    responses:
      - name: success
        status_code: 200
        body: '{"id": ":id", "name": "Alice"}'
      - name: not_found
        status_code: 404
        body: '{"error": "User not found"}'
```

The first response is used by default. (Scenario switching coming soon.)

## Architecture

```
mocksmith/
├── cmd/mocksmith/     # CLI entry point (Cobra)
├── internal/
│   ├── models/        # Data structures (Config, Route, Response)
│   ├── parser/        # YAML/JSON config parsing & validation
│   ├── matcher/       # Route matching engine (path, query, header, body)
│   ├── server/        # HTTP server with hot-reload support
│   └── logger/        # Request logging (text & JSON)
└── tests/             # Test suite
```

### Route Matching Priority

1. **Method** — HTTP method must match exactly
2. **Path** — Path must match template (with `:param` and `*` wildcards)
3. **Query** — All specified query parameters must be present and match
4. **Headers** — All specified headers must be present and match
5. **Body** — Request body must match the body matcher (if specified)

## Use Cases

- **Frontend Development** — Mock backend APIs while frontend is being built
- **API Testing** — Test API clients against realistic responses
- **CI/CD Pipelines** — Mock external services in integration tests
- **Demos & Prototyping** — Quickly stand up a fake API for presentations
- **Contract Testing** — Verify API consumers match expected request/response formats

## Development

```bash
# Build
go build -o mocksmith ./cmd/mocksmith

# Run tests
go test ./tests/... -v

# Run specific test
go test ./tests/... -run TestServer -v
```

## License

MIT

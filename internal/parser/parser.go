// Package parser loads mock server configuration from YAML or JSON files.
package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/EdgarOrtegaRamirez/mocksmith/internal/models"
	"gopkg.in/yaml.v3"
)

// Load reads a config file and returns a parsed Config.
// Supports .yaml, .yml, and .json extensions.
func Load(path string) (*models.Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(path))
	cfg := models.NewDefaultConfig()

	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("parsing YAML config: %w", err)
		}
	case ".json":
		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("parsing JSON config: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported config format: %s (use .yaml, .yml, or .json)", ext)
	}

	cfg.ApplyDefaults()

	if err := validate(cfg); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return cfg, nil
}

// LoadFromBytes parses config content directly (useful for testing).
func LoadFromBytes(data []byte, format string) (*models.Config, error) {
	cfg := models.NewDefaultConfig()

	switch strings.ToLower(format) {
	case "yaml", "yml":
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("parsing YAML config: %w", err)
		}
	case "json":
		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("parsing JSON config: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}

	cfg.ApplyDefaults()

	if err := validate(cfg); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return cfg, nil
}

// Validate checks a config for common errors.
func validate(cfg *models.Config) error {
	if len(cfg.Routes) == 0 {
		return fmt.Errorf("no routes defined")
	}

	for i, route := range cfg.Routes {
		if route.Method == "" {
			return fmt.Errorf("route %d: method is required", i)
		}
		method := strings.ToUpper(route.Method)
		if !isValidMethod(method) {
			return fmt.Errorf("route %d: invalid method %q", i, route.Method)
		}
		if route.Path == "" {
			return fmt.Errorf("route %d: path is required", i)
		}
		if !strings.HasPrefix(route.Path, "/") {
			return fmt.Errorf("route %d: path must start with /", i)
		}
		if len(route.Responses) == 0 {
			return fmt.Errorf("route %d: at least one response is required", i)
		}
		for j, resp := range route.Responses {
			if resp.StatusCode < 100 || resp.StatusCode > 599 {
				return fmt.Errorf("route %d, response %d: invalid status code %d", i, j, resp.StatusCode)
			}
		}
	}

	return nil
}

func isValidMethod(method string) bool {
	valid := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS", "TRACE"}
	for _, v := range valid {
		if method == v {
			return true
		}
	}
	return false
}

// SampleConfig generates a sample YAML configuration string.
func SampleConfig() string {
	return `# MockSmith - Sample Configuration
# HTTP Mock Server Configuration

server:
  host: "0.0.0.0"
  port: 3000
  log_level: info

cors:
  allow_origins: ["*"]
  allow_methods: ["GET", "POST", "PUT", "DELETE"]
  allow_headers: ["Content-Type", "Authorization"]

routes:
  - name: List Users
    method: GET
    path: /api/users
    responses:
      - status_code: 200
        content_type: application/json
        body: |
          [
            {"id": 1, "name": "Alice", "email": "alice@example.com"},
            {"id": 2, "name": "Bob", "email": "bob@example.com"}
          ]

  - name: Get User by ID
    method: GET
    path: /api/users/:id
    responses:
      - name: success
        status_code: 200
        content_type: application/json
        body: |
          {"id": 1, "name": "Alice", "email": "alice@example.com"}
      - name: not_found
        status_code: 404
        content_type: application/json
        body: |
          {"error": "User not found"}

  - name: Create User
    method: POST
    path: /api/users
    headers:
      Content-Type: application/json
    responses:
      - status_code: 201
        content_type: application/json
        body: |
          {"id": 3, "name": "New User", "email": "new@example.com"}

  - name: Slow Response
    method: GET
    path: /api/slow
    delay:
      min_ms: 1000
      max_ms: 3000
    responses:
      - status_code: 200
        body: '{"status": "ok"}'

  - name: Health Check
    method: GET
    path: /health
    responses:
      - status_code: 200
        body: '{"status": "healthy"}'
`
}

package mocksmith_test

import (
	"testing"

	"github.com/EdgarOrtegaRamirez/mocksmith/internal/parser"
)

func TestLoadFromBytes_ValidYAML(t *testing.T) {
	yaml := `
server:
  port: 8080
  host: "localhost"
routes:
  - method: GET
    path: /api/test
    responses:
      - status_code: 200
        body: '{"ok": true}'
`
	cfg, err := parser.LoadFromBytes([]byte(yaml), "yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("expected port 8080, got %d", cfg.Server.Port)
	}
	if cfg.Server.Host != "localhost" {
		t.Errorf("expected host localhost, got %q", cfg.Server.Host)
	}
	if len(cfg.Routes) != 1 {
		t.Fatalf("expected 1 route, got %d", len(cfg.Routes))
	}
}

func TestLoadFromBytes_ValidJSON(t *testing.T) {
	json := `{
		"server": {"port": 9090},
		"routes": [
			{
				"method": "POST",
				"path": "/api/data",
				"responses": [
					{"status_code": 201, "body": "{}"}
				]
			}
		]
	}`
	cfg, err := parser.LoadFromBytes([]byte(json), "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Server.Port != 9090 {
		t.Errorf("expected port 9090, got %d", cfg.Server.Port)
	}
}

func TestLoadFromBytes_NoRoutes(t *testing.T) {
	yaml := `
server:
  port: 8080
routes: []
`
	_, err := parser.LoadFromBytes([]byte(yaml), "yaml")
	if err == nil {
		t.Fatal("expected error for empty routes")
	}
}

func TestLoadFromBytes_MissingMethod(t *testing.T) {
	yaml := `
routes:
  - path: /api/test
    responses:
      - status_code: 200
`
	_, err := parser.LoadFromBytes([]byte(yaml), "yaml")
	if err == nil {
		t.Fatal("expected error for missing method")
	}
}

func TestLoadFromBytes_InvalidMethod(t *testing.T) {
	yaml := `
routes:
  - method: INVALID
    path: /api/test
    responses:
      - status_code: 200
`
	_, err := parser.LoadFromBytes([]byte(yaml), "yaml")
	if err == nil {
		t.Fatal("expected error for invalid method")
	}
}

func TestLoadFromBytes_MissingPath(t *testing.T) {
	yaml := `
routes:
  - method: GET
    responses:
      - status_code: 200
`
	_, err := parser.LoadFromBytes([]byte(yaml), "yaml")
	if err == nil {
		t.Fatal("expected error for missing path")
	}
}

func TestLoadFromBytes_PathNotStartingWithSlash(t *testing.T) {
	yaml := `
routes:
  - method: GET
    path: api/test
    responses:
      - status_code: 200
`
	_, err := parser.LoadFromBytes([]byte(yaml), "yaml")
	if err == nil {
		t.Fatal("expected error for path not starting with /")
	}
}

func TestLoadFromBytes_InvalidStatusCode(t *testing.T) {
	yaml := `
routes:
  - method: GET
    path: /test
    responses:
      - status_code: 999
`
	_, err := parser.LoadFromBytes([]byte(yaml), "yaml")
	if err == nil {
		t.Fatal("expected error for invalid status code")
	}
}

func TestLoadFromBytes_NoResponses(t *testing.T) {
	yaml := `
routes:
  - method: GET
    path: /test
`
	_, err := parser.LoadFromBytes([]byte(yaml), "yaml")
	if err == nil {
		t.Fatal("expected error for no responses")
	}
}

func TestLoadFromBytes_DefaultsApplied(t *testing.T) {
	yaml := `
routes:
  - method: GET
    path: /test
    responses:
      - body: "hello"
`
	cfg, err := parser.LoadFromBytes([]byte(yaml), "yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Server.Port != 3000 {
		t.Errorf("expected default port 3000, got %d", cfg.Server.Port)
	}
	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("expected default host 0.0.0.0, got %q", cfg.Server.Host)
	}
	if cfg.Routes[0].Responses[0].StatusCode != 200 {
		t.Errorf("expected default status 200, got %d", cfg.Routes[0].Responses[0].StatusCode)
	}
}

func TestLoadFromBytes_AllHTTPMethods(t *testing.T) {
	methods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS", "TRACE"}
	for _, method := range methods {
		yaml := `
routes:
  - method: ` + method + `
    path: /test
    responses:
      - status_code: 200
`
		_, err := parser.LoadFromBytes([]byte(yaml), "yaml")
		if err != nil {
			t.Errorf("method %s should be valid: %v", method, err)
		}
	}
}

func TestLoadFromBytes_WithCORS(t *testing.T) {
	yaml := `
cors:
  allow_origins: ["*"]
  allow_methods: ["GET", "POST"]
  allow_headers: ["Content-Type"]
routes:
  - method: GET
    path: /test
    responses:
      - status_code: 200
`
	cfg, err := parser.LoadFromBytes([]byte(yaml), "yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.CORS == nil {
		t.Fatal("expected CORS config")
	}
	if len(cfg.CORS.AllowOrigins) != 1 || cfg.CORS.AllowOrigins[0] != "*" {
		t.Errorf("expected allow_origins ['*'], got %v", cfg.CORS.AllowOrigins)
	}
}

func TestLoadFromBytes_WithDelay(t *testing.T) {
	yaml := `
routes:
  - method: GET
    path: /slow
    delay:
      min_ms: 100
      max_ms: 500
    responses:
      - status_code: 200
`
	cfg, err := parser.LoadFromBytes([]byte(yaml), "yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Routes[0].Delay == nil {
		t.Fatal("expected delay config")
	}
	if cfg.Routes[0].Delay.Min != 100 || cfg.Routes[0].Delay.Max != 500 {
		t.Errorf("expected delay 100-500ms, got %d-%d", cfg.Routes[0].Delay.Min, cfg.Routes[0].Delay.Max)
	}
}

func TestSampleConfig(t *testing.T) {
	sample := parser.SampleConfig()
	if sample == "" {
		t.Fatal("sample config should not be empty")
	}
	// Verify it can be parsed
	_, err := parser.LoadFromBytes([]byte(sample), "yaml")
	if err != nil {
		t.Fatalf("sample config should be valid: %v", err)
	}
}

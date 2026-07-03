package mocksmith_test

import (
	"testing"

	"github.com/EdgarOrtegaRamirez/mocksmith/internal/models"
)

func TestNewDefaultConfig(t *testing.T) {
	cfg := models.NewDefaultConfig()
	if cfg.Server.Port != 3000 {
		t.Errorf("expected default port 3000, got %d", cfg.Server.Port)
	}
	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("expected default host 0.0.0.0, got %q", cfg.Server.Host)
	}
	if cfg.Server.LogLevel != "info" {
		t.Errorf("expected default log level 'info', got %q", cfg.Server.LogLevel)
	}
}

func TestApplyDefaults(t *testing.T) {
	cfg := &models.Config{}
	cfg.ApplyDefaults()

	if cfg.Server.Port != 3000 {
		t.Errorf("expected default port 3000, got %d", cfg.Server.Port)
	}
	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("expected default host, got %q", cfg.Server.Host)
	}
	if cfg.Server.LogLevel != "info" {
		t.Errorf("expected default log level, got %q", cfg.Server.LogLevel)
	}
}

func TestApplyDefaults_PreservesExisting(t *testing.T) {
	cfg := &models.Config{
		Server: models.ServerConfig{
			Port:     8080,
			Host:     "localhost",
			LogLevel: "debug",
		},
	}
	cfg.ApplyDefaults()

	if cfg.Server.Port != 8080 {
		t.Errorf("expected port 8080, got %d", cfg.Server.Port)
	}
	if cfg.Server.Host != "localhost" {
		t.Errorf("expected host localhost, got %q", cfg.Server.Host)
	}
	if cfg.Server.LogLevel != "debug" {
		t.Errorf("expected log level debug, got %q", cfg.Server.LogLevel)
	}
}

func TestApplyDefaults_ResponseDefaults(t *testing.T) {
	cfg := &models.Config{
		Routes: []models.Route{
			{
				Method: "GET",
				Path:   "/test",
				Responses: []models.Response{
					{Body: "hello"},
				},
			},
		},
	}
	cfg.ApplyDefaults()

	resp := cfg.Routes[0].Responses[0]
	if resp.StatusCode != 200 {
		t.Errorf("expected default status 200, got %d", resp.StatusCode)
	}
	if resp.ContentType != "application/json" {
		t.Errorf("expected default content type, got %q", resp.ContentType)
	}
}

func TestApplyDefaults_PreservesStatusCode(t *testing.T) {
	cfg := &models.Config{
		Routes: []models.Route{
			{
				Method: "GET",
				Path:   "/test",
				Responses: []models.Response{
					{StatusCode: 418, Body: "I'm a teapot"},
				},
			},
		},
	}
	cfg.ApplyDefaults()

	if cfg.Routes[0].Responses[0].StatusCode != 418 {
		t.Errorf("expected status 418 preserved, got %d", cfg.Routes[0].Responses[0].StatusCode)
	}
}

func TestRequestLogFields(t *testing.T) {
	entry := models.RequestLog{
		Method:     "POST",
		Path:       "/api/test",
		StatusCode: 201,
		Matched:    true,
		RouteName:  "create",
	}

	if entry.Method != "POST" {
		t.Errorf("expected method POST, got %q", entry.Method)
	}
	if entry.StatusCode != 201 {
		t.Errorf("expected status 201, got %d", entry.StatusCode)
	}
	if !entry.Matched {
		t.Error("expected matched=true")
	}
}

func TestMatchResultDefaults(t *testing.T) {
	result := &models.MatchResult{}
	if result.Matched {
		t.Error("expected Matched=false by default")
	}
	if result.Params == nil {
		// nil params is fine
	}
}

func TestRouteWithMultipleResponses(t *testing.T) {
	cfg := &models.Config{
		Routes: []models.Route{
			{
				Method: "GET",
				Path:   "/test",
				Responses: []models.Response{
					{Name: "success", StatusCode: 200, Body: "ok"},
					{Name: "error", StatusCode: 500, Body: "fail"},
				},
			},
		},
	}

	if len(cfg.Routes[0].Responses) != 2 {
		t.Errorf("expected 2 responses, got %d", len(cfg.Routes[0].Responses))
	}
	if cfg.Routes[0].Responses[0].Name != "success" {
		t.Errorf("expected first response named 'success', got %q", cfg.Routes[0].Responses[0].Name)
	}
}

func TestCORSConfig(t *testing.T) {
	cfg := &models.Config{
		CORS: &models.CORSConfig{
			AllowOrigins:     []string{"*"},
			AllowMethods:     []string{"GET", "POST"},
			AllowHeaders:     []string{"Content-Type"},
			AllowCredentials: true,
			MaxAge:           3600,
		},
	}

	if cfg.CORS == nil {
		t.Fatal("expected CORS config")
	}
	if len(cfg.CORS.AllowOrigins) != 1 || cfg.CORS.AllowOrigins[0] != "*" {
		t.Errorf("expected allow_origins ['*'], got %v", cfg.CORS.AllowOrigins)
	}
	if !cfg.CORS.AllowCredentials {
		t.Error("expected AllowCredentials=true")
	}
}

func TestDelayConfig(t *testing.T) {
	delay := &models.DelayConfig{
		Min: 100,
		Max: 500,
	}
	if delay.Min != 100 || delay.Max != 500 {
		t.Errorf("expected delay 100-500ms, got %d-%d", delay.Min, delay.Max)
	}
}

func TestBodyMatcher(t *testing.T) {
	tests := []struct {
		name   string
		matcher models.BodyMatcher
	}{
		{"exact", models.BodyMatcher{Exact: `{"key":"value"}`}},
		{"contains", models.BodyMatcher{Contains: `"key"`}},
		{"json_path", models.BodyMatcher{JSONPath: "status=active"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just verify it can be created
			if tt.matcher.Exact == "" && tt.matcher.Contains == "" && tt.matcher.JSONPath == "" {
				t.Error("expected at least one matcher field to be set")
			}
		})
	}
}

// Package models defines the core data structures for MockSmith.
package models

import (
	"net/http"
	"time"
)

// Config represents the top-level mock server configuration.
type Config struct {
	Server   ServerConfig            `yaml:"server" json:"server"`
	CORS     *CORSConfig             `yaml:"cors,omitempty" json:"cors,omitempty"`
	Routes   []Route                 `yaml:"routes" json:"routes"`
	Scenarios map[string][]Route     `yaml:"scenarios,omitempty" json:"scenarios,omitempty"`
}

// ServerConfig holds server startup options.
type ServerConfig struct {
	Port      int    `yaml:"port" json:"port"`
	Host      string `yaml:"host" json:"host"`
	TLS       *TLSConfig `yaml:"tls,omitempty" json:"tls,omitempty"`
	LogLevel  string `yaml:"log_level" json:"log_level"`
}

// TLSConfig for HTTPS support.
type TLSConfig struct {
	CertFile string `yaml:"cert_file" json:"cert_file"`
	KeyFile  string `yaml:"key_file" json:"key_file"`
}

// CORSConfig configures Cross-Origin Resource Sharing headers.
type CORSConfig struct {
	AllowOrigins     []string `yaml:"allow_origins" json:"allow_origins"`
	AllowMethods     []string `yaml:"allow_methods" json:"allow_methods"`
	AllowHeaders     []string `yaml:"allow_headers" json:"allow_headers"`
	ExposeHeaders    []string `yaml:"expose_headers" json:"expose_headers"`
	AllowCredentials bool     `yaml:"allow_credentials" json:"allow_credentials"`
	MaxAge           int      `yaml:"max_age" json:"max_age"`
}

// Route defines a mock endpoint with matching criteria and responses.
type Route struct {
	Name        string            `yaml:"name,omitempty" json:"name,omitempty"`
	Method      string            `yaml:"method" json:"method"`
	Path        string            `yaml:"path" json:"path"`
	Query       map[string]string `yaml:"query,omitempty" json:"query,omitempty"`
	Headers     map[string]string `yaml:"headers,omitempty" json:"headers,omitempty"`
	Body        *BodyMatcher      `yaml:"body,omitempty" json:"body,omitempty"`
	Responses   []Response        `yaml:"responses" json:"responses"`
	Delay       *DelayConfig      `yaml:"delay,omitempty" json:"delay,omitempty"`
	Tags        []string          `yaml:"tags,omitempty" json:"tags,omitempty"`
}

// BodyMatcher matches request body content.
type BodyMatcher struct {
	JSONPath  string `yaml:"json_path,omitempty" json:"json_path,omitempty"`
	Contains  string `yaml:"contains,omitempty" json:"contains,omitempty"`
	Exact     string `yaml:"exact,omitempty" json:"exact,omitempty"`
}

// Response defines what to return when a route matches.
type Response struct {
	Name        string            `yaml:"name,omitempty" json:"name,omitempty"`
	StatusCode  int               `yaml:"status_code" json:"status_code"`
	Headers     map[string]string `yaml:"headers,omitempty" json:"headers,omitempty"`
	Body        string            `yaml:"body" json:"body"`
	ContentType string            `yaml:"content_type,omitempty" json:"content_type,omitempty"`
}

// DelayConfig simulates network latency.
type DelayConfig struct {
	Min int `yaml:"min_ms" json:"min_ms"`
	Max int `yaml:"max_ms" json:"max_ms"`
}

// MatchResult holds the outcome of route matching.
type MatchResult struct {
	Matched  bool
	Route    *Route
	Response *Response
	Params   map[string]string // :id style path params
}

// RequestLog captures incoming request details.
type RequestLog struct {
	Timestamp   time.Time         `json:"timestamp"`
	Method      string            `json:"method"`
	Path        string            `json:"path"`
	RemoteAddr  string            `json:"remote_addr"`
	StatusCode  int               `json:"status_code"`
	Matched     bool              `json:"matched"`
	RouteName   string            `json:"route_name,omitempty"`
	Headers     map[string]string `json:"headers"`
	Params      map[string]string `json:"params,omitempty"`
	Duration    time.Duration     `json:"duration"`
}

// NewDefaultConfig returns a Config with sensible defaults.
func NewDefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port:     3000,
			Host:     "0.0.0.0",
			LogLevel: "info",
		},
		Routes: []Route{},
	}
}

// ApplyDefaults fills in missing values with defaults.
func (c *Config) ApplyDefaults() {
	if c.Server.Port == 0 {
		c.Server.Port = 3000
	}
	if c.Server.Host == "" {
		c.Server.Host = "0.0.0.0"
	}
	if c.Server.LogLevel == "" {
		c.Server.LogLevel = "info"
	}
	for i := range c.Routes {
		for j := range c.Routes[i].Responses {
			if c.Routes[i].Responses[j].StatusCode == 0 {
				c.Routes[i].Responses[j].StatusCode = http.StatusOK
			}
			if c.Routes[i].Responses[j].ContentType == "" {
				c.Routes[i].Responses[j].ContentType = "application/json"
			}
		}
	}
}

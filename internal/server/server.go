// Package server implements the HTTP mock server for MockSmith.
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/EdgarOrtegaRamirez/mocksmith/internal/logger"
	"github.com/EdgarOrtegaRamirez/mocksmith/internal/matcher"
	"github.com/EdgarOrtegaRamirez/mocksmith/internal/models"
)

// MockServer is the main HTTP server.
type MockServer struct {
	config     *models.Config
	httpServer *http.Server
	logger     *logger.Logger
	watcher    *ConfigWatcher
	mu         sync.RWMutex
}

// New creates a new MockServer.
func New(cfg *models.Config, log *logger.Logger) *MockServer {
	return &MockServer{
		config: cfg,
		logger: log,
	}
}

// Start begins serving HTTP requests.
func (s *MockServer) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleRequest)

	addr := fmt.Sprintf("%s:%d", s.config.Server.Host, s.config.Server.Port)
	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	s.logger.LogEvent(logger.LevelInfo, "MockSmith starting on %s", addr)
	s.logger.LogEvent(logger.LevelInfo, "Loaded %d route(s)", len(s.config.Routes))

	var err error
	if s.config.Server.TLS != nil && s.config.Server.TLS.CertFile != "" {
		s.logger.LogEvent(logger.LevelInfo, "HTTPS enabled with %s", s.config.Server.TLS.CertFile)
		err = s.httpServer.ListenAndServeTLS(s.config.Server.TLS.CertFile, s.config.Server.TLS.KeyFile)
	} else {
		err = s.httpServer.ListenAndServe()
	}
	if err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}
	return nil
}

// Stop gracefully shuts down the server.
func (s *MockServer) Stop(ctx context.Context) error {
	if s.watcher != nil {
		s.watcher.Stop()
	}
	return s.httpServer.Shutdown(ctx)
}

// ReloadConfig hot-reloads a new configuration.
func (s *MockServer) ReloadConfig(cfg *models.Config) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config = cfg
	s.logger.LogEvent(logger.LevelInfo, "Config reloaded: %d route(s)", len(cfg.Routes))
}

// StartWatcher starts a file watcher for hot-reloading config.
func (s *MockServer) StartWatcher(configPath string) error {
	watcher, err := NewConfigWatcher(configPath, s)
	if err != nil {
		return err
	}
	s.watcher = watcher
	go watcher.Watch()
	return nil
}

// GetRequestCount returns the total number of requests handled.
func (s *MockServer) GetRequestCount() int {
	return s.logger.GetCount()
}

// HandleRequest is the public entry point for HTTP request handling.
func (s *MockServer) HandleRequest(w http.ResponseWriter, r *http.Request) {
	s.handleRequest(w, r)
}

func (s *MockServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Read body (limited to 1MB)
	var bodyStr string
	if r.Body != nil {
		bodyBytes, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
		if err == nil {
			bodyStr = string(bodyBytes)
		}
		r.Body.Close()
	}

	// Parse query parameters
	query := make(map[string][]string)
	for key, values := range r.URL.Query() {
		query[key] = values
	}

	// Normalize headers
	headers := make(map[string]string)
	for key := range r.Header {
		headers[key] = r.Header.Get(key)
	}

	// Match route
	s.mu.RLock()
	routes := s.config.Routes
	result := matcher.MatchRoute(r.Method, r.URL.Path, query, headers, bodyStr, routes)
	corsConfig := s.config.CORS
	s.mu.RUnlock()

	entry := models.RequestLog{
		Timestamp:  start,
		Method:     r.Method,
		Path:       r.URL.Path,
		RemoteAddr: r.RemoteAddr,
		Headers:    headers,
		Duration:   time.Since(start),
	}

	if !result.Matched {
		entry.StatusCode = http.StatusNotFound
		entry.Matched = false

		s.logger.LogEvent(logger.LevelDebug, "No match for %s %s", r.Method, r.URL.Path)

		// Apply CORS headers even for unmatched routes
		s.applyCORS(w, corsConfig)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "No matching route found", "method": "` + r.Method + `", "path": "` + r.URL.Path + `"}`))
		s.logger.LogRequest(entry)
		return
	}

	// Apply delay if configured
	if result.Route.Delay != nil {
		delay := calculateDelay(result.Route.Delay)
		time.Sleep(delay)
	}

	resp := result.Response
	entry.Matched = true
	entry.StatusCode = resp.StatusCode
	entry.RouteName = result.Route.Name
	entry.Params = result.Params

	// Apply CORS headers
	s.applyCORS(w, corsConfig)

	// Set response headers
	if resp.ContentType != "" {
		w.Header().Set("Content-Type", resp.ContentType)
	}
	for key, value := range resp.Headers {
		w.Header().Set(key, value)
	}

	w.WriteHeader(resp.StatusCode)

	// Apply template substitution for path params
	body := resp.Body
	if result.Params != nil {
		for key, value := range result.Params {
			body = strings.ReplaceAll(body, ":"+key, value)
		}
	}

	w.Write([]byte(body))

	entry.Duration = time.Since(start)
	s.logger.LogRequest(entry)
}

func (s *MockServer) applyCORS(w http.ResponseWriter, cors *models.CORSConfig) {
	if cors == nil {
		return
	}
	if len(cors.AllowOrigins) > 0 {
		w.Header().Set("Access-Control-Allow-Origin", strings.Join(cors.AllowOrigins, ", "))
	}
	if len(cors.AllowMethods) > 0 {
		w.Header().Set("Access-Control-Allow-Methods", strings.Join(cors.AllowMethods, ", "))
	}
	if len(cors.AllowHeaders) > 0 {
		w.Header().Set("Access-Control-Allow-Headers", strings.Join(cors.AllowHeaders, ", "))
	}
	if len(cors.ExposeHeaders) > 0 {
		w.Header().Set("Access-Control-Expose-Headers", strings.Join(cors.ExposeHeaders, ", "))
	}
	if cors.AllowCredentials {
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	}
	if cors.MaxAge > 0 {
		w.Header().Set("Access-Control-Max-Age", fmt.Sprintf("%d", cors.MaxAge))
	}
}

func calculateDelay(delay *models.DelayConfig) time.Duration {
	min := delay.Min
	max := delay.Max
	if min > max {
		min, max = max, min
	}
	if min == max {
		return time.Duration(min) * time.Millisecond
	}
	ms := min + rand.Intn(max-min+1)
	return time.Duration(ms) * time.Millisecond
}

// ConfigWatcher watches a config file for changes and triggers reload.
type ConfigWatcher struct {
	configPath string
	server     *MockServer
	stopCh     chan struct{}
}

// NewConfigWatcher creates a new ConfigWatcher.
func NewConfigWatcher(configPath string, server *MockServer) (*ConfigWatcher, error) {
	return &ConfigWatcher{
		configPath: configPath,
		server:     server,
		stopCh:     make(chan struct{}),
	}, nil
}

// Watch monitors the config file for changes.
func (w *ConfigWatcher) Watch() {
	// Simple poll-based watching (cross-platform)
	var lastMod time.Time
	if info, err := os.Stat(w.configPath); err == nil {
		lastMod = info.ModTime()
	}

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-w.stopCh:
			return
		case <-ticker.C:
			info, err := os.Stat(w.configPath)
			if err != nil {
				continue
			}
			if info.ModTime().After(lastMod) {
				lastMod = info.ModTime()
				w.reload()
			}
		}
	}
}

// Stop halts the watcher.
func (w *ConfigWatcher) Stop() {
	close(w.stopCh)
}

func (w *ConfigWatcher) reload() {
	cfg, err := LoadConfigFile(w.configPath)
	if err != nil {
		w.server.logger.LogEvent(logger.LevelError, "Failed to reload config: %v", err)
		return
	}

	w.server.ReloadConfig(cfg)
	w.server.logger.LogEvent(logger.LevelInfo, "Config hot-reloaded successfully")
}

// LoadConfigFile loads a config from a file path.
func LoadConfigFile(path string) (*models.Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	cfg := &models.Config{}
	if strings.HasSuffix(path, ".json") {
		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("parsing JSON: %w", err)
		}
	} else {
		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("parsing config: %w", err)
		}
	}

	cfg.ApplyDefaults()
	return cfg, nil
}

// ServeCmd is an entry point for the serve command.
func ServeCmd(cfg *models.Config, log *logger.Logger, watch bool, configPath string) error {
	srv := New(cfg, log)

	if watch && configPath != "" {
		if err := srv.StartWatcher(configPath); err != nil {
			log.LogEvent(logger.LevelWarn, "Failed to start file watcher: %v", err)
		}
	}

	return srv.Start()
}

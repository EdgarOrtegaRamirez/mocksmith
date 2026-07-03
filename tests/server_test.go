package mocksmith_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/EdgarOrtegaRamirez/mocksmith/internal/logger"
	"github.com/EdgarOrtegaRamirez/mocksmith/internal/parser"
	"github.com/EdgarOrtegaRamirez/mocksmith/internal/server"
)

func newTestServer(t *testing.T, yaml string) *httptest.Server {
	t.Helper()
	cfg, err := parser.LoadFromBytes([]byte(yaml), "yaml")
	if err != nil {
		t.Fatalf("failed to parse config: %v", err)
	}

	log := logger.NewWithWriter("error", false, &bytes.Buffer{})
	srv := server.New(cfg, log)

	return httptest.NewServer(http.HandlerFunc(srv.HandleRequest))
}

func TestServer_SimpleRoute(t *testing.T) {
	yaml := `
routes:
  - method: GET
    path: /hello
    responses:
      - status_code: 200
        body: '{"message": "hello world"}'
`
	ts := newTestServer(t, yaml)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/hello")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var body map[string]string
	json.NewDecoder(resp.Body).Decode(&body)
	if body["message"] != "hello world" {
		t.Errorf("expected 'hello world', got %q", body["message"])
	}
}

func TestServer_PathParams(t *testing.T) {
	yaml := `
routes:
  - method: GET
    path: /api/users/:id
    responses:
      - status_code: 200
        body: '{"id": ":id"}'
`
	ts := newTestServer(t, yaml)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/users/42")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	var body map[string]string
	json.NewDecoder(resp.Body).Decode(&body)
	if body["id"] != "42" {
		t.Errorf("expected id=42, got %q", body["id"])
	}
}

func TestServer_MultipleRoutes(t *testing.T) {
	yaml := `
routes:
  - method: GET
    path: /api/users
    name: list
    responses:
      - status_code: 200
        body: '[]'
  - method: POST
    path: /api/users
    name: create
    responses:
      - status_code: 201
        body: '{"created": true}'
`
	ts := newTestServer(t, yaml)
	defer ts.Close()

	// GET
	resp1, err := http.Get(ts.URL + "/api/users")
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	resp1.Body.Close()
	if resp1.StatusCode != 200 {
		t.Errorf("GET: expected 200, got %d", resp1.StatusCode)
	}

	// POST
	resp2, err := http.Post(ts.URL+"/api/users", "application/json", bytes.NewBufferString(`{}`))
	if err != nil {
		t.Fatalf("POST failed: %v", err)
	}
	resp2.Body.Close()
	if resp2.StatusCode != 201 {
		t.Errorf("POST: expected 201, got %d", resp2.StatusCode)
	}
}

func TestServer_NotFound(t *testing.T) {
	yaml := `
routes:
  - method: GET
    path: /exists
    responses:
      - status_code: 200
`
	ts := newTestServer(t, yaml)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/does-not-exist")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 404 {
		t.Errorf("expected 404 for unknown route, got %d", resp.StatusCode)
	}
}

func TestServer_MethodNotAllowed(t *testing.T) {
	yaml := `
routes:
  - method: POST
    path: /only-post
    responses:
      - status_code: 201
`
	ts := newTestServer(t, yaml)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/only-post")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 404 {
		t.Errorf("expected 404 for wrong method, got %d", resp.StatusCode)
	}
}

func TestServer_QueryParams(t *testing.T) {
	yaml := `
routes:
  - method: GET
    path: /api/search
    query:
      q: test
    responses:
      - status_code: 200
        body: '{"results": []}'
`
	ts := newTestServer(t, yaml)
	defer ts.Close()

	// With correct query
	resp1, err := http.Get(ts.URL + "/api/search?q=test")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	resp1.Body.Close()
	if resp1.StatusCode != 200 {
		t.Errorf("expected 200 with correct query, got %d", resp1.StatusCode)
	}

	// Without required query
	resp2, err := http.Get(ts.URL + "/api/search?q=wrong")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	resp2.Body.Close()
	if resp2.StatusCode != 404 {
		t.Errorf("expected 404 without correct query, got %d", resp2.StatusCode)
	}
}

func TestServer_CustomHeaders(t *testing.T) {
	yaml := `
routes:
  - method: GET
    path: /api/auth
    headers:
      Authorization: Bearer secret-token
    responses:
      - status_code: 200
        body: '{"authorized": true}'
`
	ts := newTestServer(t, yaml)
	defer ts.Close()

	// With correct header
	req, _ := http.NewRequest("GET", ts.URL+"/api/auth", nil)
	req.Header.Set("Authorization", "Bearer secret-token")
	resp1, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	resp1.Body.Close()
	if resp1.StatusCode != 200 {
		t.Errorf("expected 200 with correct header, got %d", resp1.StatusCode)
	}

	// Without header
	resp2, err := http.Get(ts.URL + "/api/auth")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	resp2.Body.Close()
	if resp2.StatusCode != 404 {
		t.Errorf("expected 404 without header, got %d", resp2.StatusCode)
	}
}

func TestServer_BodyMatch(t *testing.T) {
	yaml := `
routes:
  - method: POST
    path: /api/webhook
    body:
      contains: '"event":"push"'
    responses:
      - status_code: 200
        body: '{"received": true}'
`
	ts := newTestServer(t, yaml)
	defer ts.Close()

	// Matching body
	resp1, err := http.Post(ts.URL+"/api/webhook", "application/json",
		bytes.NewBufferString(`{"event":"push","repo":"test"}`))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	resp1.Body.Close()
	if resp1.StatusCode != 200 {
		t.Errorf("expected 200 with matching body, got %d", resp1.StatusCode)
	}

	// Non-matching body
	resp2, err := http.Post(ts.URL+"/api/webhook", "application/json",
		bytes.NewBufferString(`{"event":"pull_request"}`))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	resp2.Body.Close()
	if resp2.StatusCode != 404 {
		t.Errorf("expected 404 with non-matching body, got %d", resp2.StatusCode)
	}
}

func TestServer_CORSHeaders(t *testing.T) {
	yaml := `
cors:
  allow_origins: ["https://example.com"]
  allow_methods: ["GET", "POST"]
  allow_headers: ["Content-Type"]
  max_age: 3600
routes:
  - method: GET
    path: /api/cors
    responses:
      - status_code: 200
`
	ts := newTestServer(t, yaml)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/cors")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	resp.Body.Close()

	if resp.Header.Get("Access-Control-Allow-Origin") != "https://example.com" {
		t.Errorf("expected CORS origin header, got %q", resp.Header.Get("Access-Control-Allow-Origin"))
	}
	if resp.Header.Get("Access-Control-Allow-Methods") != "GET, POST" {
		t.Errorf("expected CORS methods header, got %q", resp.Header.Get("Access-Control-Allow-Methods"))
	}
}

func TestServer_MultipleResponses(t *testing.T) {
	yaml := `
routes:
  - method: GET
    path: /api/status
    responses:
      - name: healthy
        status_code: 200
        body: '{"status": "ok"}'
      - name: degraded
        status_code: 503
        body: '{"status": "degraded"}'
`
	cfg, err := parser.LoadFromBytes([]byte(yaml), "yaml")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	// Verify first response is the default
	if cfg.Routes[0].Responses[0].StatusCode != 200 {
		t.Errorf("expected first response to be 200, got %d", cfg.Routes[0].Responses[0].StatusCode)
	}
}

func TestServer_JSONResponse(t *testing.T) {
	yaml := `
routes:
  - method: GET
    path: /api/json
    responses:
      - status_code: 200
        content_type: application/json
        body: '{"key": "value", "number": 42}'
`
	ts := newTestServer(t, yaml)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/json")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	ct := resp.Header.Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected content-type application/json, got %q", ct)
	}

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	if body["key"] != "value" {
		t.Errorf("expected key=value, got %q", body["key"])
	}
}

func TestServer_CustomContentType(t *testing.T) {
	yaml := `
routes:
  - method: GET
    path: /api/xml
    responses:
      - status_code: 200
        content_type: application/xml
        body: '<root><item>test</item></root>'
`
	ts := newTestServer(t, yaml)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/xml")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	ct := resp.Header.Get("Content-Type")
	if ct != "application/xml" {
		t.Errorf("expected content-type application/xml, got %q", ct)
	}
}

func TestServer_RootPath(t *testing.T) {
	yaml := `
routes:
  - method: GET
    path: /
    responses:
      - status_code: 200
        body: '{"root": true}'
`
	ts := newTestServer(t, yaml)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("expected 200 for root path, got %d", resp.StatusCode)
	}
}

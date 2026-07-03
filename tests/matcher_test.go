package mocksmith_test

import (
	"testing"

	"github.com/EdgarOrtegaRamirez/mocksmith/internal/matcher"
	"github.com/EdgarOrtegaRamirez/mocksmith/internal/models"
)

func TestMatchRoute_BasicGet(t *testing.T) {
	routes := []models.Route{
		{
			Method: "GET",
			Path:   "/api/users",
			Responses: []models.Response{
				{StatusCode: 200, Body: `{"users": []}`},
			},
		},
	}

	result := matcher.MatchRoute("GET", "/api/users", nil, nil, "", routes)
	if !result.Matched {
		t.Fatal("expected match")
	}
	if result.Response.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", result.Response.StatusCode)
	}
}

func TestMatchRoute_MethodMismatch(t *testing.T) {
	routes := []models.Route{
		{
			Method: "POST",
			Path:   "/api/users",
			Responses: []models.Response{
				{StatusCode: 201},
			},
		},
	}

	result := matcher.MatchRoute("GET", "/api/users", nil, nil, "", routes)
	if result.Matched {
		t.Fatal("expected no match")
	}
}

func TestMatchRoute_PathParams(t *testing.T) {
	routes := []models.Route{
		{
			Method: "GET",
			Path:   "/api/users/:id",
			Responses: []models.Response{
				{StatusCode: 200, Body: `{"id": ":id"}`},
			},
		},
	}

	result := matcher.MatchRoute("GET", "/api/users/42", nil, nil, "", routes)
	if !result.Matched {
		t.Fatal("expected match")
	}
	if result.Params["id"] != "42" {
		t.Errorf("expected param id=42, got %q", result.Params["id"])
	}
}

func TestMatchRoute_PathParamsMultiple(t *testing.T) {
	routes := []models.Route{
		{
			Method: "GET",
			Path:   "/api/users/:userId/posts/:postId",
			Responses: []models.Response{
				{StatusCode: 200},
			},
		},
	}

	result := matcher.MatchRoute("GET", "/api/users/1/posts/99", nil, nil, "", routes)
	if !result.Matched {
		t.Fatal("expected match")
	}
	if result.Params["userId"] != "1" {
		t.Errorf("expected userId=1, got %q", result.Params["userId"])
	}
	if result.Params["postId"] != "99" {
		t.Errorf("expected postId=99, got %q", result.Params["postId"])
	}
}

func TestMatchRoute_QueryParams(t *testing.T) {
	routes := []models.Route{
		{
			Method: "GET",
			Path:   "/api/search",
			Query:  map[string]string{"q": "hello"},
			Responses: []models.Response{
				{StatusCode: 200},
			},
		},
	}

	query := map[string][]string{"q": {"hello"}}
	result := matcher.MatchRoute("GET", "/api/search", query, nil, "", routes)
	if !result.Matched {
		t.Fatal("expected match with correct query")
	}

	// Wrong query value
	query2 := map[string][]string{"q": {"world"}}
	result2 := matcher.MatchRoute("GET", "/api/search", query2, nil, "", routes)
	if result2.Matched {
		t.Fatal("expected no match with wrong query")
	}
}

func TestMatchRoute_Headers(t *testing.T) {
	routes := []models.Route{
		{
			Method:  "POST",
			Path:    "/api/data",
			Headers: map[string]string{"Content-Type": "application/json"},
			Responses: []models.Response{
				{StatusCode: 201},
			},
		},
	}

	headers := map[string]string{"Content-Type": "application/json"}
	result := matcher.MatchRoute("POST", "/api/data", nil, headers, "", routes)
	if !result.Matched {
		t.Fatal("expected match with correct headers")
	}

	// Wrong content type
	headers2 := map[string]string{"Content-Type": "text/plain"}
	result2 := matcher.MatchRoute("POST", "/api/data", nil, headers2, "", routes)
	if result2.Matched {
		t.Fatal("expected no match with wrong content type")
	}
}

func TestMatchRoute_BodyExact(t *testing.T) {
	routes := []models.Route{
		{
			Method: "POST",
			Path:   "/api/hook",
			Body:   &models.BodyMatcher{Exact: `{"event":"push"}`},
			Responses: []models.Response{
				{StatusCode: 200},
			},
		},
	}

	body := `{"event":"push"}`
	result := matcher.MatchRoute("POST", "/api/hook", nil, nil, body, routes)
	if !result.Matched {
		t.Fatal("expected match with exact body")
	}
}

func TestMatchRoute_BodyContains(t *testing.T) {
	routes := []models.Route{
		{
			Method: "POST",
			Path:   "/api/webhook",
			Body:   &models.BodyMatcher{Contains: `"type":"push"`},
			Responses: []models.Response{
				{StatusCode: 200},
			},
		},
	}

	body := `{"type":"push","repo":"test","commits":[]}`
	result := matcher.MatchRoute("POST", "/api/webhook", nil, nil, body, routes)
	if !result.Matched {
		t.Fatal("expected match with body contains")
	}
}

func TestMatchRoute_NoMatch(t *testing.T) {
	routes := []models.Route{
		{
			Method: "GET",
			Path:   "/api/users",
			Responses: []models.Response{
				{StatusCode: 200},
			},
		},
	}

	result := matcher.MatchRoute("GET", "/api/unknown", nil, nil, "", routes)
	if result.Matched {
		t.Fatal("expected no match")
	}
}

func TestMatchRoute_Wildcard(t *testing.T) {
	routes := []models.Route{
		{
			Method: "GET",
			Path:   "/api/*",
			Responses: []models.Response{
				{StatusCode: 200},
			},
		},
	}

	result := matcher.MatchRoute("GET", "/api/anything/here", nil, nil, "", routes)
	if !result.Matched {
		t.Fatal("expected wildcard to match")
	}
}

func TestMatchRoute_RootPath(t *testing.T) {
	routes := []models.Route{
		{
			Method: "GET",
			Path:   "/",
			Responses: []models.Response{
				{StatusCode: 200, Body: "root"},
			},
		},
	}

	result := matcher.MatchRoute("GET", "/", nil, nil, "", routes)
	if !result.Matched {
		t.Fatal("expected root path to match")
	}
}

func TestMatchRoute_TrailingSlash(t *testing.T) {
	routes := []models.Route{
		{
			Method: "GET",
			Path:   "/api/users",
			Responses: []models.Response{
				{StatusCode: 200},
			},
		},
	}

	// Should match with trailing slash (normalized)
	result := matcher.MatchRoute("GET", "/api/users/", nil, nil, "", routes)
	if !result.Matched {
		t.Fatal("expected trailing slash to be normalized")
	}
}

func TestMatchRoute_MultipleRoutes(t *testing.T) {
	routes := []models.Route{
		{Method: "GET", Path: "/api/users", Responses: []models.Response{{StatusCode: 200, Body: "list"}}},
		{Method: "POST", Path: "/api/users", Responses: []models.Response{{StatusCode: 201, Body: "created"}}},
		{Method: "GET", Path: "/api/users/:id", Responses: []models.Response{{StatusCode: 200, Body: "detail"}}},
	}

	// GET list
	r1 := matcher.MatchRoute("GET", "/api/users", nil, nil, "", routes)
	if !r1.Matched || r1.Response.Body != "list" {
		t.Error("expected list response")
	}

	// POST create
	r2 := matcher.MatchRoute("POST", "/api/users", nil, nil, "", routes)
	if !r2.Matched || r2.Response.Body != "created" {
		t.Error("expected created response")
	}

	// GET detail
	r3 := matcher.MatchRoute("GET", "/api/users/5", nil, nil, "", routes)
	if !r3.Matched || r3.Response.Body != "detail" {
		t.Error("expected detail response")
	}
}

func TestExtractPathParams(t *testing.T) {
	tests := []struct {
		path     string
		expected []string
	}{
		{"/api/users/:id", []string{"id"}},
		{"/api/users/:userId/posts/:postId", []string{"userId", "postId"}},
		{"/api/health", nil},
		{"/api/users/:id/posts/:postId/comments/:commentId", []string{"id", "postId", "commentId"}},
	}

	for _, tt := range tests {
		params := matcher.ExtractPathParams(tt.path)
		if len(params) != len(tt.expected) {
			t.Errorf("path %q: expected %d params, got %d", tt.path, len(tt.expected), len(params))
			continue
		}
		for i, p := range params {
			if p != tt.expected[i] {
				t.Errorf("path %q: param[%d] = %q, want %q", tt.path, i, p, tt.expected[i])
			}
		}
	}
}

func TestCompilePathRegex(t *testing.T) {
	tests := []struct {
		path    string
		request string
		match   bool
	}{
		{"/api/users/:id", "/api/users/42", true},
		{"/api/users/:id", "/api/users/abc", true},
		{"/api/users/:id", "/api/users/42/extra", false},
		{"/api/*", "/api/anything/here", true},
	}

	for _, tt := range tests {
		re, err := matcher.CompilePathRegex(tt.path)
		if err != nil {
			t.Errorf("path %q: compile error: %v", tt.path, err)
			continue
		}
		matched := re.MatchString(tt.request)
		if matched != tt.match {
			t.Errorf("path %q, request %q: match = %v, want %v", tt.path, tt.request, matched, tt.match)
		}
	}
}

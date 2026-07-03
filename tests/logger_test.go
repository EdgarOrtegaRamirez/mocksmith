package mocksmith_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/EdgarOrtegaRamirez/mocksmith/internal/logger"
	"github.com/EdgarOrtegaRamirez/mocksmith/internal/models"
)

func TestLogger_TextOutput(t *testing.T) {
	var buf bytes.Buffer
	log := logger.NewWithWriter("info", false, &buf)

	entry := models.RequestLog{
		Timestamp:  time.Date(2026, 7, 3, 12, 30, 45, 0, time.UTC),
		Method:     "GET",
		Path:       "/api/users",
		StatusCode: 200,
		Matched:    true,
		RouteName:  "list-users",
		Duration:   5 * time.Millisecond,
	}

	log.LogRequest(entry)
	output := buf.String()

	if !strings.Contains(output, "GET") {
		t.Error("output should contain method")
	}
	if !strings.Contains(output, "/api/users") {
		t.Error("output should contain path")
	}
	if !strings.Contains(output, "200") {
		t.Error("output should contain status code")
	}
	if !strings.Contains(output, "list-users") {
		t.Error("output should contain route name")
	}
}

func TestLogger_JSONOutput(t *testing.T) {
	var buf bytes.Buffer
	log := logger.NewWithWriter("info", true, &buf)

	entry := models.RequestLog{
		Timestamp:  time.Date(2026, 7, 3, 12, 30, 45, 0, time.UTC),
		Method:     "POST",
		Path:       "/api/data",
		StatusCode: 201,
		Matched:    true,
		Duration:   10 * time.Millisecond,
	}

	log.LogRequest(entry)
	output := buf.String()

	var parsed models.RequestLog
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &parsed); err != nil {
		t.Fatalf("output should be valid JSON: %v", err)
	}
	if parsed.Method != "POST" {
		t.Errorf("expected method POST, got %q", parsed.Method)
	}
	if parsed.StatusCode != 201 {
		t.Errorf("expected status 201, got %d", parsed.StatusCode)
	}
}

func TestLogger_Unmatched(t *testing.T) {
	var buf bytes.Buffer
	log := logger.NewWithWriter("info", false, &buf)

	entry := models.RequestLog{
		Timestamp:  time.Now(),
		Method:     "GET",
		Path:       "/unknown",
		StatusCode: 404,
		Matched:    false,
		Duration:   1 * time.Millisecond,
	}

	log.LogRequest(entry)
	output := buf.String()

	if !strings.Contains(output, "miss") {
		t.Error("output should contain 'miss' for unmatched routes")
	}
}

func TestLogger_Count(t *testing.T) {
	var buf bytes.Buffer
	log := logger.NewWithWriter("info", false, &buf)

	for i := 0; i < 5; i++ {
		log.LogRequest(models.RequestLog{
			Timestamp: time.Now(),
			Method:    "GET",
			Path:      "/test",
			Duration:  time.Millisecond,
		})
	}

	if log.GetCount() != 5 {
		t.Errorf("expected 5 requests logged, got %d", log.GetCount())
	}
}

func TestLogger_LevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	log := logger.NewWithWriter("warn", false, &buf)

	// These should not appear
	log.LogEvent(logger.LevelDebug, "debug message")
	log.LogEvent(logger.LevelInfo, "info message")

	// These should appear
	log.LogEvent(logger.LevelWarn, "warn message")
	log.LogEvent(logger.LevelError, "error message")

	output := buf.String()
	if strings.Contains(output, "debug message") {
		t.Error("debug should be filtered out")
	}
	if strings.Contains(output, "info message") {
		t.Error("info should be filtered out")
	}
	if !strings.Contains(output, "warn message") {
		t.Error("warn should appear")
	}
	if !strings.Contains(output, "error message") {
		t.Error("error should appear")
	}
}

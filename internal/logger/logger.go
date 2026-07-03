// Package logger provides request logging for MockSmith.
package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/EdgarOrtegaRamirez/mocksmith/internal/models"
)

// Level represents log severity.
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

// Logger captures and formats request logs.
type Logger struct {
	level   Level
	output  io.Writer
	jsonOut bool
	mu      sync.Mutex
	count   int
}

// New creates a Logger with the given level and output.
func New(level string, jsonOutput bool) *Logger {
	l := &Logger{
		level:   parseLevel(level),
		output:  os.Stdout,
		jsonOut: jsonOutput,
	}
	return l
}

// NewWithWriter creates a Logger writing to a specific writer.
func NewWithWriter(level string, jsonOutput bool, w io.Writer) *Logger {
	l := &Logger{
		level:   parseLevel(level),
		output:  w,
		jsonOut: jsonOutput,
	}
	return l
}

// LogRequest logs an incoming request.
func (l *Logger) LogRequest(entry models.RequestLog) {
	l.mu.Lock()
	l.count++
	l.mu.Unlock()

	if l.jsonOut {
		l.logJSON(entry)
	} else {
		l.logText(entry)
	}
}

func (l *Logger) logText(entry models.RequestLog) {
	sc := getStatusColor(entry.StatusCode)
	method := padRight(entry.Method, 7)
	path := entry.Path
	statusCode := fmt.Sprintf("%d", entry.StatusCode)

	matchIndicator := "miss"
	if entry.Matched {
		matchIndicator = "hit "
		routeName := entry.RouteName
		if routeName == "" {
			routeName = "-"
		}
		fmt.Fprintf(l.output, "%s %s%s\033[0m %s%-3s\033[0m %s %s [%s] %v\n",
			entry.Timestamp.Format("15:04:05"),
			sc, statusCode,
			sc, method,
			path,
			matchIndicator,
			routeName,
			entry.Duration.Round(time.Microsecond),
		)
	} else {
		fmt.Fprintf(l.output, "%s %s%s\033[0m %s%-3s\033[0m %s %s [-] %v\n",
			entry.Timestamp.Format("15:04:05"),
			sc, statusCode,
			sc, method,
			path,
			matchIndicator,
			entry.Duration.Round(time.Microsecond),
		)
	}
}

func (l *Logger) logJSON(entry models.RequestLog) {
	data, _ := json.Marshal(entry)
	fmt.Fprintf(l.output, "%s\n", data)
}

// GetCount returns the total number of logged requests.
func (l *Logger) GetCount() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.count
}

// LogEvent logs a general server event.
func (l *Logger) LogEvent(level Level, format string, args ...interface{}) {
	if level < l.level {
		return
	}
	prefix := levelPrefix(level)
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(l.output, "%s %s\n", prefix, msg)
}

func parseLevel(s string) Level {
	switch strings.ToLower(s) {
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	case "warn", "warning":
		return LevelWarn
	case "error":
		return LevelError
	default:
		return LevelInfo
	}
}

func levelPrefix(l Level) string {
	switch l {
	case LevelDebug:
		return "\033[36mDEBUG\033[0m"
	case LevelInfo:
		return "\033[32mINFO\033[0m"
	case LevelWarn:
		return "\033[33mWARN\033[0m"
	case LevelError:
		return "\033[31mERROR\033[0m"
	default:
		return "INFO"
	}
}

func getStatusColor(code int) string {
	switch {
	case code >= 200 && code < 300:
		return "\033[32m" // green
	case code >= 300 && code < 400:
		return "\033[36m" // cyan
	case code >= 400 && code < 500:
		return "\033[33m" // yellow
	case code >= 500:
		return "\033[31m" // red
	default:
		return "\033[0m"
	}
}

func padRight(s string, width int) string {
	if len(s) >= width {
		return s[:width]
	}
	return s + strings.Repeat(" ", width-len(s))
}

# AGENTS.md

## Project Overview
MockSmith is a config-driven HTTP mock server written in Go. It loads route definitions from YAML/JSON config files and serves mock responses with path parameter matching, query/header/body matching, delays, CORS, and hot-reload.

## Architecture

### Core Modules

1. **models/** - Data structures (Config, Route, Response, MatchResult, RequestLog)
2. **parser/** - YAML/JSON config loading, validation, sample generation
3. **matcher/** - Route matching engine (path with :param/*, query, headers, body)
4. **server/** - HTTP server with hot-reload file watching
5. **logger/** - Request logging (colored text or JSON structured output)

### CLI Commands
- `mocksmith serve <config>` - Start the mock server
- `mocksmith validate <config>` - Validate a config file
- `mocksmith routes <config>` - List configured routes
- `mocksmith sample` - Generate sample config
- `mocksmith version` - Print version

## Testing
- Run all tests: `go test ./tests/... -v`
- Run specific package: `go test ./tests/... -run TestServer -v`
- Tests use `httptest.NewServer` for integration testing

## Key Design Decisions
- Route matching is ordered — first match wins
- Path parameters use `:name` syntax (like Express.js)
- Wildcards use `*` to match any remaining path
- Body matching supports exact, contains, and JSON path
- Config hot-reload uses polling (2s interval) for cross-platform compatibility

## Adding New Features
- New route matchers: add to `matcher/matcher.go`
- New CLI commands: add to `cmd/mocksmith/main.go`
- New config options: add to `models/models.go`

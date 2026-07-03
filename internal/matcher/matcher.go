// Package matcher implements route matching logic for MockSmith.
package matcher

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/EdgarOrtegaRamirez/mocksmith/internal/models"
)

// pathSegment represents a parsed path template segment.
type pathSegment struct {
	literal  string // literal text
	param    string // :param name, empty for literal segments
	wildcard bool   // matches any remaining path
}

// parsePathTemplate parses a path like /api/users/:id/posts into segments.
func parsePathTemplate(path string) []pathSegment {
	var segments []pathSegment
	parts := strings.Split(strings.Trim(path, "/"), "/")
	for _, part := range parts {
		if part == "*" {
			segments = append(segments, pathSegment{wildcard: true})
		} else if strings.HasPrefix(part, ":") {
			segments = append(segments, pathSegment{param: part[1:]})
		} else {
			segments = append(segments, pathSegment{literal: part})
		}
	}
	return segments
}

// MatchRoute checks a request against all routes and returns the best match.
func MatchRoute(method, path string, query map[string][]string, headers map[string]string, body string, routes []models.Route) *models.MatchResult {
	method = strings.ToUpper(method)
	path = strings.TrimRight(path, "/")
	if path == "" {
		path = "/"
	}

	for i := range routes {
		route := &routes[i]
		if !matchMethod(route, method) {
			continue
		}
		params, pathMatched := matchPath(route, path)
		if !pathMatched {
			continue
		}
		if !matchQuery(route, query) {
			continue
		}
		if !matchHeaders(route, headers) {
			continue
		}
		if !matchBody(route, body) {
			continue
		}

		// Pick the first response (or "default" named response)
		resp := pickResponse(route)

		return &models.MatchResult{
			Matched:  true,
			Route:    route,
			Response: resp,
			Params:   params,
		}
	}

	return &models.MatchResult{Matched: false}
}

func matchMethod(route *models.Route, method string) bool {
	return strings.ToUpper(route.Method) == method
}

func matchPath(route *models.Route, reqPath string) (map[string]string, bool) {
	templateSegs := parsePathTemplate(route.Path)
	reqParts := strings.Split(strings.Trim(reqPath, "/"), "/")

	params := make(map[string]string)
	tIdx, rIdx := 0, 0

	for tIdx < len(templateSegs) {
		seg := templateSegs[tIdx]

		if seg.wildcard {
			// Matches everything remaining
			return params, true
		}

		if seg.param != "" {
			if rIdx >= len(reqParts) {
				return nil, false
			}
			params[seg.param] = reqParts[rIdx]
			rIdx++
			tIdx++
			continue
		}

		// Literal match
		if rIdx >= len(reqParts) {
			return nil, false
		}
		if seg.literal != reqParts[rIdx] {
			return nil, false
		}
		rIdx++
		tIdx++
	}

	return params, rIdx == len(reqParts)
}

func matchQuery(route *models.Route, query map[string][]string) bool {
	if len(route.Query) == 0 {
		return true
	}
	for key, expected := range route.Query {
		values, ok := query[key]
		if !ok {
			return false
		}
		found := false
		for _, v := range values {
			if v == expected {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func matchHeaders(route *models.Route, headers map[string]string) bool {
	if len(route.Headers) == 0 {
		return true
	}
	for key, expected := range route.Headers {
		actual, ok := headers[key]
		if !ok {
			actual = headers[strings.ToLower(key)]
		}
		if !ok && actual == "" {
			return false
		}
		if !strings.EqualFold(actual, expected) {
			return false
		}
	}
	return true
}

func matchBody(route *models.Route, body string) bool {
	if route.Body == nil {
		return true
	}
	if body == "" {
		return false
	}

	bm := route.Body
	if bm.Exact != "" {
		return body == bm.Exact
	}
	if bm.Contains != "" {
		return strings.Contains(body, bm.Contains)
	}
	if bm.JSONPath != "" {
		return matchJSONPath(body, bm.JSONPath)
	}
	return true
}

// matchJSONPath does simple top-level JSON field matching.
// e.g., json_path: "status=active" matches {"status": "active", ...}
func matchJSONPath(body, path string) bool {
	eqIdx := strings.Index(path, "=")
	if eqIdx < 0 {
		return false
	}
	key := path[:eqIdx]
	value := path[eqIdx+1:]

	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(body), &obj); err != nil {
		return false
	}

	val, ok := obj[key]
	if !ok {
		return false
	}

	// Handle quoted values
	if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
		value = value[1 : len(value)-1]
	}

	return fmt.Sprintf("%v", val) == value
}

func pickResponse(route *models.Route) *models.Response {
	if len(route.Responses) == 0 {
		return nil
	}
	return &route.Responses[0]
}

// ExtractPathParams returns the param names from a route path template.
func ExtractPathParams(pathTemplate string) []string {
	var params []string
	for _, seg := range parsePathTemplate(pathTemplate) {
		if seg.param != "" {
			params = append(params, seg.param)
		}
	}
	return params
}

// CompilePathRegex compiles a path template into a regex for advanced matching.
func CompilePathRegex(pathTemplate string) (*regexp.Regexp, error) {
	segments := parsePathTemplate(pathTemplate)
	pattern := "^/"
	for i, seg := range segments {
		if seg.wildcard {
			pattern += ".*"
		} else if seg.param != "" {
			pattern += "([^/]+)"
		} else {
			pattern += regexp.QuoteMeta(seg.literal)
		}
		if i < len(segments)-1 {
			pattern += "/"
		}
	}
	pattern += "$"
	return regexp.Compile(pattern)
}

package router

import (
	"context"
	"strings"
)

type Params map[string]string

type HandlerFunc func(context.Context, Params, any) any

type Handlers map[string]HandlerFunc

func (h Handlers) Match(subject string) (HandlerFunc, Params, bool) {
	for route, handler := range h {
		params, ok := h.match(route, subject)
		if ok {
			return handler, params, true
		}
	}
	return nil, nil, false
}

func (h Handlers) match(route, subject string) (map[string]string, bool) {
	// Split route and subject into segments
	routeParts := strings.Split(route, ".")
	subjectParts := strings.Split(subject, ".")

	// Check if lengths match
	if len(routeParts) != len(subjectParts) {
		return nil, false
	}

	// Initialize result map for variables
	params := make(map[string]string)

	// Compare each segment
	for i, routePart := range routeParts {
		// Check if routePart is a variable (starts with < and ends with >)
		if strings.HasPrefix(routePart, "<") && strings.HasSuffix(routePart, ">") {
			// Extract variable name (remove < and >)
			varName := strings.TrimPrefix(strings.TrimSuffix(routePart, ">"), "<")
			// Validate variable name doesn't contain < or >
			if strings.Contains(varName, "<") || strings.Contains(varName, ">") {
				return nil, false
			}
			// Store variable name and corresponding subject value
			params[varName] = subjectParts[i]
		} else {
			// If not a variable, segments must match exactly
			if routePart != subjectParts[i] {
				return nil, false
			}
		}
	}

	return params, true
}

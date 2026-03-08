package api

import "strings"

// Common error extension codes returned by the Linear GraphQL API.
const (
	CodeEntityNotFound  = "ENTITY_NOT_FOUND"
	CodeRateLimited     = "RATELIMITED"
	CodeForbidden       = "FORBIDDEN"
	CodeValidationError = "VALIDATION_ERROR"
)

// GraphQLError represents a single error entry in a GraphQL response.
type GraphQLError struct {
	Message    string         `json:"message"`
	Path       []any          `json:"path,omitempty"`
	Extensions map[string]any `json:"extensions,omitempty"`
}

// Code returns the error code from extensions, or empty string if absent.
func (e GraphQLError) Code() string {
	if e.Extensions == nil {
		return ""
	}
	code, _ := e.Extensions["code"].(string)
	return code
}

// GraphQLErrors is a slice of GraphQLError that implements the error interface.
type GraphQLErrors []GraphQLError

func (e GraphQLErrors) Error() string {
	msgs := make([]string, len(e))
	for i, err := range e {
		msgs[i] = err.Message
	}
	return strings.Join(msgs, "; ")
}

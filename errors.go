package grid

import (
	"errors"
	"fmt"
	"strings"
)

// ErrorType represents GRID GraphQL error types documented in the GRID API reference.
type ErrorType string

const (
	ErrorTypeUnknown            ErrorType = "UNKNOWN"
	ErrorTypeInternal           ErrorType = "INTERNAL"
	ErrorTypeNotFound           ErrorType = "NOT_FOUND"
	ErrorTypeUnauthenticated    ErrorType = "UNAUTHENTICATED"
	ErrorTypeUnauthorized       ErrorType = "UNAUTHORIZED"
	ErrorTypePermissionDenied   ErrorType = "PERMISSION_DENIED"
	ErrorTypeBadRequest         ErrorType = "BAD_REQUEST"
	ErrorTypeUnavailable        ErrorType = "UNAVAILABLE"
	ErrorTypeFailedPrecondition ErrorType = "FAILED_PRECONDITION"
)

// ErrorDetail represents GRID GraphQL error detail codes.
type ErrorDetail string

const (
	ErrorDetailUnknown              ErrorDetail = "UNKNOWN"
	ErrorDetailFieldNotFound        ErrorDetail = "FIELD_NOT_FOUND"
	ErrorDetailInvalidCursor        ErrorDetail = "INVALID_CURSOR"
	ErrorDetailUnimplemented        ErrorDetail = "UNIMPLEMENTED"
	ErrorDetailInvalidArgument      ErrorDetail = "INVALID_ARGUMENT"
	ErrorDetailDeadlineExceeded     ErrorDetail = "DEADLINE_EXCEEDED"
	ErrorDetailServiceError         ErrorDetail = "SERVICE_ERROR"
	ErrorDetailThrottledCPU         ErrorDetail = "THROTTLED_CPU"
	ErrorDetailThrottledConcurrency ErrorDetail = "THROTTLED_CONCURRENCY"
	ErrorDetailEnhanceYourCalm      ErrorDetail = "ENHANCE_YOUR_CALM"
	ErrorDetailTooManyRequests      ErrorDetail = "TOO_MANY_REQUESTS"
	ErrorDetailTCPFailure           ErrorDetail = "TCP_FAILURE"
	ErrorDetailMissingResource      ErrorDetail = "MISSING_RESOURCE"
)

// ErrorLocation identifies a line and column in a GraphQL query.
type ErrorLocation struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

// ErrorExtensions holds GRID-specific GraphQL error metadata.
type ErrorExtensions struct {
	Code        string      `json:"code,omitempty"`
	ServiceName string      `json:"serviceName,omitempty"`
	ErrorType   ErrorType   `json:"errorType,omitempty"`
	ErrorDetail ErrorDetail `json:"errorDetail,omitempty"`
}

// GraphQLError represents a single error entry returned in a GraphQL response.
type GraphQLError struct {
	Message    string          `json:"message"`
	Locations  []ErrorLocation `json:"locations,omitempty"`
	Path       []any           `json:"path,omitempty"`
	Extensions ErrorExtensions `json:"extensions,omitempty"`
}

func (e GraphQLError) Error() string {
	var parts []string
	if e.Extensions.ErrorType != "" {
		parts = append(parts, string(e.Extensions.ErrorType))
	}
	if e.Extensions.ErrorDetail != "" {
		parts = append(parts, string(e.Extensions.ErrorDetail))
	}
	if len(parts) > 0 {
		return fmt.Sprintf("%s: %s", strings.Join(parts, "/"), e.Message)
	}
	return e.Message
}

// APIError aggregates one or more GraphQL errors returned by a GRID endpoint.
type APIError struct {
	API        API
	StatusCode int
	Errors     []GraphQLError
}

func (e *APIError) Error() string {
	if len(e.Errors) == 0 {
		return fmt.Sprintf("grid %s api error (status %d)", e.API, e.StatusCode)
	}
	msgs := make([]string, len(e.Errors))
	for i, ge := range e.Errors {
		msgs[i] = ge.Error()
	}
	return fmt.Sprintf("grid %s api error: %s", e.API, strings.Join(msgs, "; "))
}

// HasType reports whether any error in the response has the specified ErrorType.
func (e *APIError) HasType(t ErrorType) bool {
	for _, ge := range e.Errors {
		if ge.Extensions.ErrorType == t {
			return true
		}
	}
	return false
}

// HasDetail reports whether any error in the response has the specified ErrorDetail.
func (e *APIError) HasDetail(d ErrorDetail) bool {
	for _, ge := range e.Errors {
		if ge.Extensions.ErrorDetail == d {
			return true
		}
	}
	return false
}

// HTTPError represents a non-2xx HTTP response that did not contain a valid GraphQL error payload.
type HTTPError struct {
	API        API
	StatusCode int
	Body       string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("grid %s http error %d: %s", e.API, e.StatusCode, strings.TrimSpace(e.Body))
}

// IsUnauthenticated reports whether err indicates an authentication failure.
func IsUnauthenticated(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.HasType(ErrorTypeUnauthenticated) || apiErr.StatusCode == 401
	}
	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		return httpErr.StatusCode == 401
	}
	return false
}

// IsRateLimited reports whether err indicates a GRID rate-limit rejection.
func IsRateLimited(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.HasDetail(ErrorDetailTooManyRequests) ||
			apiErr.HasDetail(ErrorDetailEnhanceYourCalm) ||
			apiErr.StatusCode == 429
	}
	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		return httpErr.StatusCode == 429
	}
	return false
}

// IsNotFound reports whether err indicates a missing GRID resource.
func IsNotFound(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.HasType(ErrorTypeNotFound)
	}
	return false
}

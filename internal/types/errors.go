// Package types provides internal shared types for the GreenWeb API.
//
// This package contains error types, response wrappers, and other internal
// data structures that are not exposed to external API consumers.
package types

import (
	"fmt"
	"net/http"
	"time"
)

// ErrorCode represents a specific error code for categorizing errors.
type ErrorCode string

const (
	// Carbon intensity related errors
	ErrorCodeCarbonIntensityFetch      ErrorCode = "CARBON_INTENSITY_FETCH_ERROR"
	ErrorCodeCarbonIntensityInvalid    ErrorCode = "CARBON_INTENSITY_INVALID"
	ErrorCodeCarbonIntensityTimeout    ErrorCode = "CARBON_INTENSITY_TIMEOUT"
	ErrorCodeCarbonIntensityUnavailable ErrorCode = "CARBON_INTENSITY_UNAVAILABLE"

	// Optimization related errors
	ErrorCodeOptimizationGeneration ErrorCode = "OPTIMIZATION_GENERATION_ERROR"
	ErrorCodeOptimizationInvalid    ErrorCode = "OPTIMIZATION_INVALID"
	ErrorCodeOptimizationRuleError  ErrorCode = "OPTIMIZATION_RULE_ERROR"

	// Location and geolocation errors
	ErrorCodeLocationInvalid    ErrorCode = "LOCATION_INVALID"
	ErrorCodeLocationNotFound   ErrorCode = "LOCATION_NOT_FOUND"
	ErrorCodeGeolocationFailed  ErrorCode = "GEOLOCATION_FAILED"

	// External API errors
	ErrorCodeExternalAPIError     ErrorCode = "EXTERNAL_API_ERROR"
	ErrorCodeExternalAPITimeout   ErrorCode = "EXTERNAL_API_TIMEOUT"
	ErrorCodeExternalAPIRateLimit ErrorCode = "EXTERNAL_API_RATE_LIMIT"
	ErrorCodeExternalAPIAuth      ErrorCode = "EXTERNAL_API_AUTH_ERROR"

	// Configuration and setup errors
	ErrorCodeConfigurationError ErrorCode = "CONFIGURATION_ERROR"
	ErrorCodeServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"
	ErrorCodeInternalError      ErrorCode = "INTERNAL_ERROR"

	// Validation errors
	ErrorCodeValidationError ErrorCode = "VALIDATION_ERROR"
	ErrorCodeInvalidRequest  ErrorCode = "INVALID_REQUEST"
	ErrorCodeMissingParameter ErrorCode = "MISSING_PARAMETER"

	// Cache related errors
	ErrorCodeCacheError     ErrorCode = "CACHE_ERROR"
	ErrorCodeCacheMiss      ErrorCode = "CACHE_MISS"
	ErrorCodeCacheCorrupted ErrorCode = "CACHE_CORRUPTED"

	// Rate limiting errors
	ErrorCodeRateLimitExceeded ErrorCode = "RATE_LIMIT_EXCEEDED"
	ErrorCodeQuotaExceeded     ErrorCode = "QUOTA_EXCEEDED"
)

// GreenWebError represents a structured error with additional context.
type GreenWebError struct {
	// Code is the specific error code for categorization
	Code ErrorCode `json:"code"`

	// Message is a human-readable error message
	Message string `json:"message"`

	// Details provides additional context about the error
	Details string `json:"details,omitempty"`

	// Cause is the underlying error that caused this error
	Cause error `json:"-"`

	// Timestamp is when the error occurred
	Timestamp time.Time `json:"timestamp"`

	// RequestID helps track the error across logs
	RequestID string `json:"request_id,omitempty"`

	// Location is the location context where the error occurred
	Location string `json:"location,omitempty"`

	// HTTPStatus is the suggested HTTP status code for this error
	HTTPStatus int `json:"-"`

	// Retryable indicates if the operation can be retried
	Retryable bool `json:"retryable"`

	// Metadata contains additional error-specific data
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Error implements the error interface.
func (e *GreenWebError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("[%s] %s: %s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the underlying cause error.
func (e *GreenWebError) Unwrap() error {
	return e.Cause
}

// Is checks if this error matches the target error code.
func (e *GreenWebError) Is(target error) bool {
	if targetErr, ok := target.(*GreenWebError); ok {
		return e.Code == targetErr.Code
	}
	return false
}

// WithCause adds a cause error to this error.
func (e *GreenWebError) WithCause(cause error) *GreenWebError {
	e.Cause = cause
	return e
}

// WithDetails adds additional details to this error.
func (e *GreenWebError) WithDetails(details string) *GreenWebError {
	e.Details = details
	return e
}

// WithRequestID adds a request ID to this error.
func (e *GreenWebError) WithRequestID(requestID string) *GreenWebError {
	e.RequestID = requestID
	return e
}

// WithLocation adds location context to this error.
func (e *GreenWebError) WithLocation(location string) *GreenWebError {
	e.Location = location
	return e
}

// WithMetadata adds metadata to this error.
func (e *GreenWebError) WithMetadata(key string, value interface{}) *GreenWebError {
	if e.Metadata == nil {
		e.Metadata = make(map[string]interface{})
	}
	e.Metadata[key] = value
	return e
}

// NewGreenWebError creates a new GreenWeb error.
func NewGreenWebError(code ErrorCode, message string) *GreenWebError {
	return &GreenWebError{
		Code:       code,
		Message:    message,
		Timestamp:  time.Now(),
		HTTPStatus: getDefaultHTTPStatus(code),
		Retryable:  isRetryableError(code),
	}
}

// NewCarbonIntensityError creates a carbon intensity related error.
func NewCarbonIntensityError(message string, cause error) *GreenWebError {
	return NewGreenWebError(ErrorCodeCarbonIntensityFetch, message).WithCause(cause)
}

// NewCarbonIntensityTimeoutError creates a carbon intensity timeout error.
func NewCarbonIntensityTimeoutError(location string) *GreenWebError {
	return NewGreenWebError(ErrorCodeCarbonIntensityTimeout, "Carbon intensity request timed out").
		WithLocation(location).
		WithDetails("The carbon intensity data provider took too long to respond")
}

// NewCarbonIntensityUnavailableError creates a carbon intensity unavailable error.
func NewCarbonIntensityUnavailableError(location string) *GreenWebError {
	return NewGreenWebError(ErrorCodeCarbonIntensityUnavailable, "Carbon intensity data unavailable").
		WithLocation(location).
		WithDetails("Carbon intensity data is temporarily unavailable for this location")
}

// NewOptimizationError creates an optimization related error.
func NewOptimizationError(message string, cause error) *GreenWebError {
	return NewGreenWebError(ErrorCodeOptimizationGeneration, message).WithCause(cause)
}

// NewLocationError creates a location related error.
func NewLocationError(location string) *GreenWebError {
	return NewGreenWebError(ErrorCodeLocationInvalid, "Invalid location").
		WithLocation(location).
		WithDetails("The specified location is not recognized or supported")
}

// NewExternalAPIError creates an external API error.
func NewExternalAPIError(apiName, message string, cause error) *GreenWebError {
	return NewGreenWebError(ErrorCodeExternalAPIError, fmt.Sprintf("%s API error", apiName)).
		WithCause(cause).
		WithDetails(message).
		WithMetadata("api_name", apiName)
}

// NewExternalAPITimeoutError creates an external API timeout error.
func NewExternalAPITimeoutError(apiName string) *GreenWebError {
	return NewGreenWebError(ErrorCodeExternalAPITimeout, fmt.Sprintf("%s API timeout", apiName)).
		WithDetails("The external API request timed out").
		WithMetadata("api_name", apiName)
}

// NewExternalAPIRateLimitError creates an external API rate limit error.
func NewExternalAPIRateLimitError(apiName string, retryAfter time.Duration) *GreenWebError {
	err := NewGreenWebError(ErrorCodeExternalAPIRateLimit, fmt.Sprintf("%s API rate limit exceeded", apiName)).
		WithDetails("The external API rate limit has been exceeded").
		WithMetadata("api_name", apiName)
	
	if retryAfter > 0 {
		err = err.WithMetadata("retry_after_seconds", int(retryAfter.Seconds()))
	}
	
	return err
}

// NewExternalAPIAuthError creates an external API authentication error.
func NewExternalAPIAuthError(apiName string) *GreenWebError {
	return NewGreenWebError(ErrorCodeExternalAPIAuth, fmt.Sprintf("%s API authentication failed", apiName)).
		WithDetails("Invalid or missing API key for external service").
		WithMetadata("api_name", apiName)
}

// NewValidationError creates a validation error.
func NewValidationError(field, message string) *GreenWebError {
	return NewGreenWebError(ErrorCodeValidationError, fmt.Sprintf("Validation failed for field '%s'", field)).
		WithDetails(message).
		WithMetadata("field", field)
}

// NewConfigurationError creates a configuration error.
func NewConfigurationError(component, message string) *GreenWebError {
	return NewGreenWebError(ErrorCodeConfigurationError, fmt.Sprintf("Configuration error in %s", component)).
		WithDetails(message).
		WithMetadata("component", component)
}

// NewCacheError creates a cache related error.
func NewCacheError(operation, message string, cause error) *GreenWebError {
	return NewGreenWebError(ErrorCodeCacheError, fmt.Sprintf("Cache %s failed", operation)).
		WithCause(cause).
		WithDetails(message).
		WithMetadata("operation", operation)
}

// NewRateLimitError creates a rate limiting error.
func NewRateLimitError(limit int, window time.Duration) *GreenWebError {
	return NewGreenWebError(ErrorCodeRateLimitExceeded, "Rate limit exceeded").
		WithDetails(fmt.Sprintf("Exceeded %d requests per %v", limit, window)).
		WithMetadata("limit", limit).
		WithMetadata("window_seconds", int(window.Seconds()))
}

// getDefaultHTTPStatus returns the default HTTP status code for an error code.
func getDefaultHTTPStatus(code ErrorCode) int {
	switch code {
	case ErrorCodeValidationError, ErrorCodeInvalidRequest, ErrorCodeMissingParameter:
		return http.StatusBadRequest
	case ErrorCodeExternalAPIAuth:
		return http.StatusUnauthorized
	case ErrorCodeLocationNotFound:
		return http.StatusNotFound
	case ErrorCodeRateLimitExceeded, ErrorCodeQuotaExceeded:
		return http.StatusTooManyRequests
	case ErrorCodeExternalAPITimeout, ErrorCodeCarbonIntensityTimeout:
		return http.StatusRequestTimeout
	case ErrorCodeServiceUnavailable, ErrorCodeCarbonIntensityUnavailable:
		return http.StatusServiceUnavailable
	case ErrorCodeExternalAPIError, ErrorCodeCarbonIntensityFetch, ErrorCodeOptimizationGeneration, 
		 ErrorCodeGeolocationFailed, ErrorCodeCacheError, ErrorCodeInternalError:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// isRetryableError determines if an error is retryable.
func isRetryableError(code ErrorCode) bool {
	switch code {
	case ErrorCodeExternalAPITimeout, ErrorCodeCarbonIntensityTimeout, ErrorCodeExternalAPIError,
		 ErrorCodeServiceUnavailable, ErrorCodeCarbonIntensityUnavailable, ErrorCodeCacheError:
		return true
	case ErrorCodeExternalAPIRateLimit, ErrorCodeRateLimitExceeded:
		return true // Retryable after delay
	default:
		return false
	}
}

// ErrorResponse represents a structured error response for APIs.
type ErrorResponse struct {
	// Error contains the error details
	Error *GreenWebError `json:"error"`

	// RequestID helps track the error across logs
	RequestID string `json:"request_id,omitempty"`

	// Timestamp is when the error response was generated
	Timestamp time.Time `json:"timestamp"`

	// Path is the API path that generated the error
	Path string `json:"path,omitempty"`

	// Method is the HTTP method that generated the error
	Method string `json:"method,omitempty"`

	// TraceID helps with distributed tracing
	TraceID string `json:"trace_id,omitempty"`
}

// NewErrorResponse creates a new error response.
func NewErrorResponse(err *GreenWebError, requestID string) *ErrorResponse {
	return &ErrorResponse{
		Error:     err,
		RequestID: requestID,
		Timestamp: time.Now(),
	}
}

// WithPath adds the API path to the error response.
func (er *ErrorResponse) WithPath(path string) *ErrorResponse {
	er.Path = path
	return er
}

// WithMethod adds the HTTP method to the error response.
func (er *ErrorResponse) WithMethod(method string) *ErrorResponse {
	er.Method = method
	return er
}

// WithTraceID adds a trace ID to the error response.
func (er *ErrorResponse) WithTraceID(traceID string) *ErrorResponse {
	er.TraceID = traceID
	return er
}

// MultiError represents multiple errors that occurred together.
type MultiError struct {
	// Errors is the list of individual errors
	Errors []*GreenWebError `json:"errors"`

	// Message is an overall message describing the multi-error
	Message string `json:"message"`

	// Count is the number of errors
	Count int `json:"count"`
}

// Error implements the error interface.
func (me *MultiError) Error() string {
	if me.Message != "" {
		return fmt.Sprintf("%s (%d errors)", me.Message, me.Count)
	}
	return fmt.Sprintf("Multiple errors occurred (%d errors)", me.Count)
}

// Add adds an error to the multi-error.
func (me *MultiError) Add(err *GreenWebError) {
	me.Errors = append(me.Errors, err)
	me.Count = len(me.Errors)
}

// HasErrors returns true if there are any errors.
func (me *MultiError) HasErrors() bool {
	return me.Count > 0
}

// FirstError returns the first error or nil if there are no errors.
func (me *MultiError) FirstError() *GreenWebError {
	if me.Count == 0 {
		return nil
	}
	return me.Errors[0]
}

// NewMultiError creates a new multi-error.
func NewMultiError(message string) *MultiError {
	return &MultiError{
		Message: message,
		Errors:  make([]*GreenWebError, 0),
		Count:   0,
	}
}

// ErrorCollector helps collect multiple errors during processing.
type ErrorCollector struct {
	errors *MultiError
}

// NewErrorCollector creates a new error collector.
func NewErrorCollector(message string) *ErrorCollector {
	return &ErrorCollector{
		errors: NewMultiError(message),
	}
}

// Add adds an error to the collector.
func (ec *ErrorCollector) Add(err *GreenWebError) {
	ec.errors.Add(err)
}

// AddError adds a generic error to the collector by wrapping it.
func (ec *ErrorCollector) AddError(code ErrorCode, message string, cause error) {
	gwErr := NewGreenWebError(code, message)
	if cause != nil {
		gwErr = gwErr.WithCause(cause)
	}
	ec.Add(gwErr)
}

// HasErrors returns true if there are any errors.
func (ec *ErrorCollector) HasErrors() bool {
	return ec.errors.HasErrors()
}

// Error returns the multi-error if there are errors, nil otherwise.
func (ec *ErrorCollector) Error() error {
	if ec.errors.HasErrors() {
		return ec.errors
	}
	return nil
}

// Count returns the number of collected errors.
func (ec *ErrorCollector) Count() int {
	return ec.errors.Count
}

// FirstError returns the first collected error.
func (ec *ErrorCollector) FirstError() *GreenWebError {
	return ec.errors.FirstError()
}
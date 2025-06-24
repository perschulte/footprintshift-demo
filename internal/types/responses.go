// Package types provides internal HTTP response types for the GreenWeb API.
package types

import (
	"time"

	"github.com/perschulte/greenweb-api/pkg/carbon"
	"github.com/perschulte/greenweb-api/pkg/optimization"
)

// APIResponse is a generic wrapper for all API responses.
type APIResponse[T any] struct {
	// Data contains the actual response data
	Data T `json:"data"`

	// Success indicates if the request was successful
	Success bool `json:"success"`

	// Message provides additional context about the response
	Message string `json:"message,omitempty"`

	// RequestID helps track the request across logs
	RequestID string `json:"request_id,omitempty"`

	// Timestamp is when the response was generated
	Timestamp time.Time `json:"timestamp"`

	// Version is the API version that generated this response
	Version string `json:"version,omitempty"`

	// Metadata contains additional response metadata
	Metadata ResponseMetadata `json:"metadata,omitempty"`
}

// ResponseMetadata contains additional metadata about the response.
type ResponseMetadata struct {
	// ProcessingTime is how long the request took to process
	ProcessingTime time.Duration `json:"processing_time,omitempty"`

	// CacheHit indicates if the response was served from cache
	CacheHit bool `json:"cache_hit,omitempty"`

	// CacheAge is the age of cached data if served from cache
	CacheAge time.Duration `json:"cache_age,omitempty"`

	// DataSource indicates where the data came from (e.g., "electricity_maps", "mock")
	DataSource string `json:"data_source,omitempty"`

	// Location is the resolved location for location-based requests
	Location string `json:"location,omitempty"`

	// RateLimitRemaining indicates remaining rate limit quota
	RateLimitRemaining int `json:"rate_limit_remaining,omitempty"`

	// RateLimitReset indicates when the rate limit resets
	RateLimitReset time.Time `json:"rate_limit_reset,omitempty"`

	// Warnings contains non-fatal warnings about the response
	Warnings []string `json:"warnings,omitempty"`
}

// NewAPIResponse creates a new API response wrapper.
func NewAPIResponse[T any](data T, requestID string) *APIResponse[T] {
	return &APIResponse[T]{
		Data:      data,
		Success:   true,
		RequestID: requestID,
		Timestamp: time.Now(),
	}
}

// WithMessage adds a message to the response.
func (r *APIResponse[T]) WithMessage(message string) *APIResponse[T] {
	r.Message = message
	return r
}

// WithVersion adds a version to the response.
func (r *APIResponse[T]) WithVersion(version string) *APIResponse[T] {
	r.Version = version
	return r
}

// WithProcessingTime adds processing time metadata.
func (r *APIResponse[T]) WithProcessingTime(duration time.Duration) *APIResponse[T] {
	r.Metadata.ProcessingTime = duration
	return r
}

// WithCacheInfo adds cache information metadata.
func (r *APIResponse[T]) WithCacheInfo(hit bool, age time.Duration) *APIResponse[T] {
	r.Metadata.CacheHit = hit
	r.Metadata.CacheAge = age
	return r
}

// WithDataSource adds data source metadata.
func (r *APIResponse[T]) WithDataSource(source string) *APIResponse[T] {
	r.Metadata.DataSource = source
	return r
}

// WithLocation adds location metadata.
func (r *APIResponse[T]) WithLocation(location string) *APIResponse[T] {
	r.Metadata.Location = location
	return r
}

// WithRateLimit adds rate limit metadata.
func (r *APIResponse[T]) WithRateLimit(remaining int, reset time.Time) *APIResponse[T] {
	r.Metadata.RateLimitRemaining = remaining
	r.Metadata.RateLimitReset = reset
	return r
}

// WithWarning adds a warning to the response.
func (r *APIResponse[T]) WithWarning(warning string) *APIResponse[T] {
	r.Metadata.Warnings = append(r.Metadata.Warnings, warning)
	return r
}

// HealthResponse represents the health check response.
type HealthResponse struct {
	// Status indicates the overall health status
	Status string `json:"status"` // "healthy", "degraded", "unhealthy"

	// Service is the service name
	Service string `json:"service"`

	// Version is the service version
	Version string `json:"version"`

	// Timestamp is when the health check was performed
	Timestamp time.Time `json:"timestamp"`

	// Checks contains individual health check results
	Checks map[string]HealthCheck `json:"checks"`

	// Uptime is how long the service has been running
	Uptime time.Duration `json:"uptime,omitempty"`
}

// HealthCheck represents an individual health check result.
type HealthCheck struct {
	// Status indicates the health status of this component
	Status string `json:"status"` // "healthy", "degraded", "unhealthy"

	// ResponseTime is how long the health check took
	ResponseTime time.Duration `json:"response_time,omitempty"`

	// Message provides additional context
	Message string `json:"message,omitempty"`

	// LastChecked is when this component was last checked
	LastChecked time.Time `json:"last_checked"`

	// Metadata contains component-specific health data
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// CarbonIntensityResponse wraps carbon intensity data with API metadata.
type CarbonIntensityResponse struct {
	*carbon.CarbonIntensity

	// API-specific metadata
	APIMetadata struct {
		// RequestedLocation is the original location requested
		RequestedLocation string `json:"requested_location,omitempty"`

		// ResolvedLocation is the actual location used for the query
		ResolvedLocation string `json:"resolved_location,omitempty"`

		// DataFreshness indicates how fresh the data is
		DataFreshness time.Duration `json:"data_freshness,omitempty"`

		// FallbackUsed indicates if fallback/mock data was used
		FallbackUsed bool `json:"fallback_used,omitempty"`

		// GridZone is the electricity grid zone identifier
		GridZone string `json:"grid_zone,omitempty"`
	} `json:"api_metadata,omitempty"`
}

// GreenHoursForecastResponse wraps green hours forecast data with API metadata.
type GreenHoursForecastResponse struct {
	*carbon.GreenHoursForecast

	// API-specific metadata
	APIMetadata struct {
		// RequestedLocation is the original location requested
		RequestedLocation string `json:"requested_location,omitempty"`

		// ResolvedLocation is the actual location used for the query
		ResolvedLocation string `json:"resolved_location,omitempty"`

		// RequestedHours is the number of hours originally requested
		RequestedHours int `json:"requested_hours,omitempty"`

		// ForecastMethod describes how the forecast was generated
		ForecastMethod string `json:"forecast_method,omitempty"`

		// DataQuality indicates the quality/reliability of the forecast
		DataQuality string `json:"data_quality,omitempty"` // "high", "medium", "low"

		// FallbackUsed indicates if fallback/mock data was used
		FallbackUsed bool `json:"fallback_used,omitempty"`
	} `json:"api_metadata,omitempty"`
}

// OptimizationResponse wraps optimization profile data with API metadata.
type OptimizationResponse struct {
	*optimization.OptimizationResponse

	// API-specific metadata
	APIMetadata struct {
		// RequestedLocation is the original location requested
		RequestedLocation string `json:"requested_location,omitempty"`

		// OptimizationAlgorithm describes the algorithm used
		OptimizationAlgorithm string `json:"optimization_algorithm,omitempty"`

		// RulesApplied indicates how many rules were applied
		RulesApplied int `json:"rules_applied,omitempty"`

		// ProfileVersion is the version of the optimization profile format
		ProfileVersion string `json:"profile_version,omitempty"`

		// EstimatedImpact provides estimated impact metrics
		EstimatedImpact struct {
			EnergySavings     float64 `json:"energy_savings_percent,omitempty"`
			PerformanceImpact float64 `json:"performance_impact_percent,omitempty"`
			UserExperience    string  `json:"user_experience_impact,omitempty"` // "minimal", "moderate", "significant"
		} `json:"estimated_impact,omitempty"`
	} `json:"api_metadata,omitempty"`
}

// PaginatedResponse wraps paginated data with pagination metadata.
type PaginatedResponse[T any] struct {
	// Data contains the paginated items
	Data []T `json:"data"`

	// Pagination contains pagination metadata
	Pagination PaginationMetadata `json:"pagination"`

	// Total is the total number of items available
	Total int64 `json:"total"`
}

// PaginationMetadata contains pagination information.
type PaginationMetadata struct {
	// Page is the current page number (1-based)
	Page int `json:"page"`

	// PerPage is the number of items per page
	PerPage int `json:"per_page"`

	// TotalPages is the total number of pages
	TotalPages int `json:"total_pages"`

	// HasNext indicates if there's a next page
	HasNext bool `json:"has_next"`

	// HasPrev indicates if there's a previous page
	HasPrev bool `json:"has_prev"`

	// NextPage is the next page number (if available)
	NextPage *int `json:"next_page,omitempty"`

	// PrevPage is the previous page number (if available)
	PrevPage *int `json:"prev_page,omitempty"`
}

// NewPaginatedResponse creates a new paginated response.
func NewPaginatedResponse[T any](data []T, page, perPage int, total int64) *PaginatedResponse[T] {
	totalPages := int((total + int64(perPage) - 1) / int64(perPage))
	
	response := &PaginatedResponse[T]{
		Data:  data,
		Total: total,
		Pagination: PaginationMetadata{
			Page:       page,
			PerPage:    perPage,
			TotalPages: totalPages,
			HasNext:    page < totalPages,
			HasPrev:    page > 1,
		},
	}

	if response.Pagination.HasNext {
		nextPage := page + 1
		response.Pagination.NextPage = &nextPage
	}

	if response.Pagination.HasPrev {
		prevPage := page - 1
		response.Pagination.PrevPage = &prevPage
	}

	return response
}

// ValidationErrorResponse represents validation error details.
type ValidationErrorResponse struct {
	// Message is the overall validation error message
	Message string `json:"message"`

	// Errors contains field-specific validation errors
	Errors map[string][]string `json:"errors"`

	// Code is the error code
	Code string `json:"code"`
}

// AddFieldError adds a validation error for a specific field.
func (ver *ValidationErrorResponse) AddFieldError(field, message string) {
	if ver.Errors == nil {
		ver.Errors = make(map[string][]string)
	}
	ver.Errors[field] = append(ver.Errors[field], message)
}

// HasErrors returns true if there are any validation errors.
func (ver *ValidationErrorResponse) HasErrors() bool {
	return len(ver.Errors) > 0
}

// NewValidationErrorResponse creates a new validation error response.
func NewValidationErrorResponse(message string) *ValidationErrorResponse {
	return &ValidationErrorResponse{
		Message: message,
		Code:    "VALIDATION_ERROR",
		Errors:  make(map[string][]string),
	}
}

// BatchResponse represents a response for batch operations.
type BatchResponse[T any] struct {
	// Results contains the results for each item in the batch
	Results []BatchResult[T] `json:"results"`

	// Summary provides a summary of the batch operation
	Summary BatchSummary `json:"summary"`

	// RequestID helps track the batch request
	RequestID string `json:"request_id,omitempty"`

	// ProcessedAt is when the batch was processed
	ProcessedAt time.Time `json:"processed_at"`
}

// BatchResult represents the result for a single item in a batch.
type BatchResult[T any] struct {
	// ID is the identifier for this batch item
	ID string `json:"id,omitempty"`

	// Success indicates if this item was processed successfully
	Success bool `json:"success"`

	// Data contains the result data (if successful)
	Data *T `json:"data,omitempty"`

	// Error contains error information (if unsuccessful)
	Error *GreenWebError `json:"error,omitempty"`
}

// BatchSummary provides a summary of batch operation results.
type BatchSummary struct {
	// Total is the total number of items in the batch
	Total int `json:"total"`

	// Successful is the number of successfully processed items
	Successful int `json:"successful"`

	// Failed is the number of failed items
	Failed int `json:"failed"`

	// ProcessingTime is the total time taken to process the batch
	ProcessingTime time.Duration `json:"processing_time"`
}

// NewBatchResponse creates a new batch response.
func NewBatchResponse[T any](requestID string) *BatchResponse[T] {
	return &BatchResponse[T]{
		Results:     make([]BatchResult[T], 0),
		RequestID:   requestID,
		ProcessedAt: time.Now(),
	}
}

// AddResult adds a result to the batch response.
func (br *BatchResponse[T]) AddResult(id string, data *T, err *GreenWebError) {
	result := BatchResult[T]{
		ID:      id,
		Success: err == nil,
		Data:    data,
		Error:   err,
	}
	
	br.Results = append(br.Results, result)
	br.updateSummary()
}

// updateSummary updates the batch summary based on current results.
func (br *BatchResponse[T]) updateSummary() {
	br.Summary.Total = len(br.Results)
	br.Summary.Successful = 0
	br.Summary.Failed = 0

	for _, result := range br.Results {
		if result.Success {
			br.Summary.Successful++
		} else {
			br.Summary.Failed++
		}
	}
}

// StreamResponse represents a server-sent events response.
type StreamResponse struct {
	// Event is the type of event
	Event string `json:"event"`

	// Data contains the event data
	Data interface{} `json:"data"`

	// ID is a unique identifier for this event
	ID string `json:"id,omitempty"`

	// Timestamp is when the event occurred
	Timestamp time.Time `json:"timestamp"`

	// Retry indicates how long the client should wait before reconnecting (in milliseconds)
	Retry int `json:"retry,omitempty"`
}

// NewStreamResponse creates a new stream response.
func NewStreamResponse(event string, data interface{}) *StreamResponse {
	return &StreamResponse{
		Event:     event,
		Data:      data,
		Timestamp: time.Now(),
	}
}

// WithID adds an ID to the stream response.
func (sr *StreamResponse) WithID(id string) *StreamResponse {
	sr.ID = id
	return sr
}

// WithRetry adds a retry interval to the stream response.
func (sr *StreamResponse) WithRetry(retryMs int) *StreamResponse {
	sr.Retry = retryMs
	return sr
}

// Format formats the stream response as a server-sent event string.
func (sr *StreamResponse) Format() string {
	var result string

	if sr.Event != "" {
		result += "event: " + sr.Event + "\n"
	}
	
	if sr.ID != "" {
		result += "id: " + sr.ID + "\n"
	}

	if sr.Retry > 0 {
		result += "retry: " + string(rune(sr.Retry)) + "\n"
	}

	// TODO: Properly format the data field based on content type
	result += "data: " + "JSON_DATA_HERE" + "\n\n"

	return result
}
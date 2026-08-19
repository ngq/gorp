// Package integration provides CloudEvents 1.0 standardized event models for gorp framework.
package integration

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"
)

// CloudEventsVersion is the CloudEvents specification version string.
const CloudEventsVersion = "1.0"

// CloudEvent represents a standardized CloudEvents v1.0.2 event payload.
type CloudEvent[T any] struct {
	SpecVersion     string         `json:"specversion"`               // "1.0"
	ID              string         `json:"id"`                        // Unique Event ID
	Source          string         `json:"source"`                    // Event Source URI (e.g. "urn:gorp:service:order")
	Type            string         `json:"type"`                      // Event Type (e.g. "com.gorp.order.created")
	DataContentType string         `json:"datacontenttype,omitempty"` // "application/json"
	Time            time.Time      `json:"time"`                      // Event Timestamp
	Subject         string         `json:"subject,omitempty"`         // Optional Subject
	Data            T              `json:"data"`                      // Business Payload
	Extensions      map[string]any `json:"extensions,omitempty"`      // CloudEvents extensions (e.g. traceparent, trace_id)
}

// Name returns the CloudEvent Type as the Event Name for EventBus compatibility.
func (e CloudEvent[T]) Name() string { return e.Type }

// Payload returns the Data payload.
func (e CloudEvent[T]) Payload() any { return e.Data }

// OccurredAt returns the Time property.
func (e CloudEvent[T]) OccurredAt() time.Time { return e.Time }

// ExtensionsMap returns the Extensions map property for Trace Context restoration.
func (e CloudEvent[T]) ExtensionsMap() map[string]any { return e.Extensions }

// NewCloudEvent constructs a new CloudEvents v1.0 instance with auto-generated UUID and Trace Context propagation.
func NewCloudEvent[T any](ctx context.Context, eventType, source string, data T) CloudEvent[T] {
	id := generateUUID()
	now := time.Now()

	ext := make(map[string]any)

	// Extract TraceID / RequestID from context if available
	if ctx != nil {
		if traceID, ok := ctx.Value("trace_id").(string); ok && traceID != "" {
			ext["trace_id"] = traceID
		}
		if reqID, ok := ctx.Value("request_id").(string); ok && reqID != "" {
			ext["request_id"] = reqID
		}
		if traceparent, ok := ctx.Value("traceparent").(string); ok && traceparent != "" {
			ext["traceparent"] = traceparent
		}
	}

	return CloudEvent[T]{
		SpecVersion:     CloudEventsVersion,
		ID:              id,
		Source:          source,
		Type:            eventType,
		DataContentType: "application/json",
		Time:            now,
		Data:            data,
		Extensions:      ext,
	}
}

func generateUUID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

package gocloak

import (
	"context"

	"go.opentelemetry.io/otel/propagation"
)

type contextKey string

var propagatorContextKey = contextKey("propagator")

// Ptr returns a pointer to the given value of any type.
func Ptr[T any](v T) *T {
	return &v
}

// StringP returns a pointer of a string variable
func StringP(value string) *string {
	return &value
}

// PString returns a string value from a pointer
func PString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

// BoolP returns a pointer of a boolean variable
func BoolP(value bool) *bool {
	return &value
}

// PBool returns a boolean value from a pointer
func PBool(value *bool) bool {
	if value == nil {
		return false
	}
	return *value
}

// IntP returns a pointer of an integer variable
func IntP(value int) *int {
	return &value
}

// Int32P returns a pointer of an int32 variable
func Int32P(value int32) *int32 {
	return &value
}

// Int64P returns a pointer of an int64 variable
func Int64P(value int64) *int64 {
	return &value
}

// PInt returns an integer value from a pointer
func PInt(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}

// PInt32 returns an int32 value from a pointer
func PInt32(value *int32) int32 {
	if value == nil {
		return 0
	}
	return *value
}

// PInt64 returns an int64 value from a pointer
func PInt64(value *int64) int64 {
	if value == nil {
		return 0
	}
	return *value
}

// Float32P returns a pointer of a float32 variable
func Float32P(value float32) *float32 {
	return &value
}

// Float64P returns a pointer of a float64 variable
func Float64P(value float64) *float64 {
	return &value
}

// PFloat32 returns a float32 value from a pointer
func PFloat32(value *float32) float32 {
	if value == nil {
		return 0
	}
	return *value
}

// PFloat64 returns a float64 value from a pointer
func PFloat64(value *float64) float64 {
	if value == nil {
		return 0
	}
	return *value
}

// NilOrEmpty returns true if string pointer is nil or the string is empty.
func NilOrEmpty(value *string) bool {
	return value == nil || len(*value) == 0
}

// NilOrEmptyArray returns true if the slice is empty or its first element is an empty string.
func NilOrEmptyArray(value []string) bool {
	if len(value) == 0 {
		return true
	}
	return value[0] == ""
}

// PStringSlice returns the slice as-is, or an empty slice if nil.
func PStringSlice(value []string) []string {
	if value == nil {
		return []string{}
	}
	return value
}

// NilOrEmptySlice returns true if the slice is nil or empty.
func NilOrEmptySlice(value []string) bool {
	return len(value) == 0
}

// WithPropagator generates a context that has an OpenTelemetry propagator attached.
// The propagator is used to inject the span from the context into the request headers.
// If no propagator is attached, the global propagator (otel.GetTextMapPropagator) is used.
func WithPropagator(ctx context.Context, propagator propagation.TextMapPropagator) context.Context {
	return context.WithValue(ctx, propagatorContextKey, propagator)
}

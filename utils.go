package gocloak

import (
	"context"

	"go.opentelemetry.io/otel/propagation"
)

type contextKey string

var propagatorContextKey = contextKey("propagator")

// PString returns the value of a string pointer, or "" if it is nil.
// Input: value [*string].
// Output: [string] dereferenced value.
func PString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

// PBool returns the value of a bool pointer, or false if it is nil.
// Input: value [*bool].
// Output: [bool] dereferenced value.
func PBool(value *bool) bool {
	if value == nil {
		return false
	}
	return *value
}

// PInt returns the value of an int pointer, or 0 if it is nil.
// Input: value [*int].
// Output: [int] dereferenced value.
func PInt(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}

// PInt32 returns the value of an int32 pointer, or 0 if it is nil.
// Input: value [*int32].
// Output: [int32] dereferenced value.
func PInt32(value *int32) int32 {
	if value == nil {
		return 0
	}
	return *value
}

// PInt64 returns the value of an int64 pointer, or 0 if it is nil.
// Input: value [*int64].
// Output: [int64] dereferenced value.
func PInt64(value *int64) int64 {
	if value == nil {
		return 0
	}
	return *value
}

// PFloat32 returns the value of a float32 pointer, or 0 if it is nil.
// Input: value [*float32].
// Output: [float32] dereferenced value.
func PFloat32(value *float32) float32 {
	if value == nil {
		return 0
	}
	return *value
}

// PFloat64 returns the value of a float64 pointer, or 0 if it is nil.
// Input: value [*float64].
// Output: [float64] dereferenced value.
func PFloat64(value *float64) float64 {
	if value == nil {
		return 0
	}
	return *value
}

// NilOrEmpty reports whether the string pointer is nil or points to an empty string.
// Input: value [*string].
// Output: [bool] true if nil or empty.
func NilOrEmpty(value *string) bool {
	return value == nil || len(*value) == 0
}

// NilOrEmptyArray reports whether the slice is empty or its first element is an empty string.
// Input: value [[]string].
// Output: [bool] true if empty or the first element is "".
func NilOrEmptyArray(value []string) bool {
	if len(value) == 0 {
		return true
	}
	return value[0] == ""
}

// WithPropagator generates a context that has an OpenTelemetry propagator attached.
// The propagator is used to inject the span from the context into the request headers.
// If no propagator is attached, the global propagator (otel.GetTextMapPropagator) is used.
// Input: ctx [context.Context], propagator [propagation.TextMapPropagator].
// Output: [context.Context] context carrying the propagator.
func WithPropagator(ctx context.Context, propagator propagation.TextMapPropagator) context.Context {
	return context.WithValue(ctx, propagatorContextKey, propagator)
}

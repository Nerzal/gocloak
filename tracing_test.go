package gocloak_test

import (
	"context"
	"testing"

	"github.com/Nerzal/gocloak/v14"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

func newSpanContext(t *testing.T) trace.SpanContext {
	t.Helper()
	traceID, err := trace.TraceIDFromHex("4bf92f3577b34da6a3ce929d0e0e4736")
	require.NoError(t, err)
	spanID, err := trace.SpanIDFromHex("00f067aa0ba902b7")
	require.NoError(t, err)

	return trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: trace.FlagsSampled,
		Remote:     true,
	})
}

func TestGetRequest_InjectsTraceHeadersWithPropagator(t *testing.T) {
	t.Parallel()
	client := gocloak.NewClient("http://localhost:8080")

	ctx := trace.ContextWithSpanContext(context.Background(), newSpanContext(t))
	ctx = gocloak.WithPropagator(ctx, propagation.TraceContext{})

	req := client.GetRequest(ctx)

	assert.Equal(t,
		"00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
		req.Header.Get("traceparent"),
	)
}

func TestGetRequest_NoTraceHeadersWithoutSpan(t *testing.T) {
	t.Parallel()
	client := gocloak.NewClient("http://localhost:8080")

	ctx := gocloak.WithPropagator(context.Background(), propagation.TraceContext{})

	req := client.GetRequest(ctx)

	assert.Empty(t, req.Header.Get("traceparent"))
}

// Not parallel: it changes the global OpenTelemetry propagator.
func TestGetRequest_UsesGlobalPropagator(t *testing.T) {
	previous := otel.GetTextMapPropagator()
	otel.SetTextMapPropagator(propagation.TraceContext{})
	t.Cleanup(func() { otel.SetTextMapPropagator(previous) })

	client := gocloak.NewClient("http://localhost:8080")
	ctx := trace.ContextWithSpanContext(context.Background(), newSpanContext(t))

	req := client.GetRequest(ctx)

	assert.Equal(t,
		"00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
		req.Header.Get("traceparent"),
	)
}

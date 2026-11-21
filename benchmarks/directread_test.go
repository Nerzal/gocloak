//go:build !baseline

package benchmarks

import (
	"context"
	"sync/atomic"

	"github.com/Nerzal/gocloak/v15"
)

// The "baseline" build tag excludes this file, so the benchmarks also build against gocloak versions without
// SetDirectBodyRead (ab.sh ref mode).
func init() {
	clientKinds = append(clientKinds, clientKind{"gocloak-direct", gocloakDirectCaller})
}

// gocloakDirectCaller builds a gocloak client with SetDirectBodyRead.
// Input: t [*target], dials [*atomic.Int64]. Output: [caller].
func gocloakDirectCaller(t *target, dials *atomic.Int64) caller {
	g := gocloak.NewClient(t.base, gocloak.SetDirectBodyRead())
	configureTransport(g, t, dials, keepPool)
	return func(ctx context.Context, ep *endpoint) error { return ep.viaGocloak(ctx, g, t) }
}

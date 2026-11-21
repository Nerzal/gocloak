//go:build keycloak

package benchmarks

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Nerzal/gocloak/v15"
)

const (
	liveUsers     = 1000 // users seeded for the users1000 endpoint
	liveGroups    = 20
	liveSubGroups = 5
	livePrefix    = "bench-"
	warmupCalls   = 300 // per endpoint, lets Keycloak's JIT and caches settle before measuring
)

// liveConfig is the subset of testdata/config.json the live benchmarks need.
type liveConfig struct {
	HostName string `json:"hostname"`
	Admin    struct {
		UserName string `json:"username"`
		Password string `json:"password"`
		Realm    string `json:"realm"`
	} `json:"admin"`
	GoCloak struct {
		Realm        string `json:"realm"`
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
	} `json:"gocloak"`
}

var (
	live      *target
	liveAdmin *gocloak.GoCloak
	liveCfg   liveConfig
)

// TestMain seeds Keycloak once, warms it up, runs the benchmarks and removes the seeded data again.
func TestMain(m *testing.M) {
	ctx, cancel := context.WithCancel(context.Background())
	if err := setupLive(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "keycloak setup:", err)
		os.Exit(1)
	}
	code := m.Run()
	cleanupLive(context.Background())
	cancel()
	os.Exit(code)
}

// setupLive reads the config, starts the admin token refresher, seeds users and groups and warms Keycloak up.
// Input: ctx [context.Context] stops the token refresher. Output: [error].
func setupLive(ctx context.Context) error {
	path := os.Getenv("GOCLOAK_TEST_CONFIG")
	if path == "" {
		path = "../testdata/config.json"
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(raw, &liveCfg); err != nil {
		return err
	}
	liveAdmin = gocloak.NewClient(liveCfg.HostName)

	var token atomic.Pointer[string]
	login := func() error {
		jwt, err := liveAdmin.LoginAdmin(ctx, liveCfg.Admin.UserName, liveCfg.Admin.Password, liveCfg.Admin.Realm)
		if err == nil {
			token.Store(&jwt.AccessToken)
		}
		return err
	}
	if err := login(); err != nil {
		return err
	}
	// master admin tokens live 60s; refresh off the measured path
	go func() {
		t := time.NewTicker(20 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				_ = login()
			}
		}
	}()

	live = &target{
		base: liveCfg.HostName, realm: liveCfg.GoCloak.Realm,
		clientID: liveCfg.GoCloak.ClientID, clientSecret: liveCfg.GoCloak.ClientSecret,
		token:    func() string { return *token.Load() },
		cooldown: 15 * time.Second,
	}
	cleanupLive(ctx) // leftovers of an interrupted run would change payload sizes
	if live.userID, err = seedLive(ctx); err != nil {
		return err
	}
	return warmup(ctx)
}

// seedLive creates liveUsers users and a group tree, in parallel.
// Input: ctx [context.Context]. Output: [string] ID of one seeded user, [error].
func seedLive(ctx context.Context) (string, error) {
	var (
		wg       sync.WaitGroup
		firstErr atomic.Value
		userID   atomic.Pointer[string]
		jobs     = make(chan int)
	)
	for range 16 {
		wg.Go(func() {
			for i := range jobs {
				u := makeUser(i)
				u.ID, u.CreatedTimestamp, u.Access = nil, nil, nil
				u.Username = new(fmt.Sprintf("%s%05d", livePrefix, i))
				id, err := liveAdmin.CreateUser(ctx, live.token(), live.realm, u)
				if err != nil {
					firstErr.CompareAndSwap(nil, err)
					continue
				}
				userID.CompareAndSwap(nil, &id)
			}
		})
	}
	for i := range liveUsers {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
	if err, ok := firstErr.Load().(error); ok {
		return "", err
	}
	for _, g := range makeGroups(liveGroups, liveSubGroups) {
		parent := gocloak.Group{Name: new(livePrefix + *g.Name), Attributes: g.Attributes}
		id, err := liveAdmin.CreateGroup(ctx, live.token(), live.realm, parent)
		if err != nil {
			return "", err
		}
		for _, s := range g.SubGroups {
			if _, err := liveAdmin.CreateChildGroup(ctx, live.token(), live.realm, id, gocloak.Group{Name: s.Name}); err != nil {
				return "", err
			}
		}
	}
	return *userID.Load(), nil
}

// warmup calls every endpoint warmupCalls times. Input: ctx [context.Context]. Output: [error].
func warmup(ctx context.Context) error {
	var dials atomic.Int64
	call := clientKinds[0].build(live, &dials)
	for i := range endpoints {
		for range warmupCalls {
			if err := call(ctx, &endpoints[i]); err != nil {
				return fmt.Errorf("warmup %s: %w", endpoints[i].name, err)
			}
		}
	}
	return nil
}

// cleanupLive deletes all seeded users and groups (prefix livePrefix). Input: ctx [context.Context].
func cleanupLive(ctx context.Context) {
	for {
		users, err := liveAdmin.GetUsers(ctx, live.token(), live.realm, gocloak.GetUsersParams{Search: new(livePrefix), Max: new(500)})
		if err != nil || len(users) == 0 {
			break
		}
		for _, u := range users {
			_ = liveAdmin.DeleteUser(ctx, live.token(), live.realm, *u.ID)
		}
	}
	groups, err := liveAdmin.GetGroups(ctx, live.token(), live.realm, gocloak.GetGroupsParams{Search: new(livePrefix), Max: new(500)})
	if err == nil {
		for _, g := range groups {
			_ = liveAdmin.DeleteGroup(ctx, live.token(), live.realm, *g.ID)
		}
	}
}

// BenchmarkKeycloakHTTP compares HTTP stacks per endpoint against the real Keycloak (plain HTTP, docker).
func BenchmarkKeycloakHTTP(b *testing.B) {
	runEndpointBenchmarks(b, live, nil)
}

// BenchmarkKeycloakPool measures many concurrent callers against the real Keycloak.
func BenchmarkKeycloakPool(b *testing.B) {
	runPoolBenchmark(b, live, "user", poolParallelism)
}

// BenchmarkKeycloakDecodeAccessToken verifies a real Keycloak token locally with warm certs.
func BenchmarkKeycloakDecodeAccessToken(b *testing.B) {
	ctx := context.Background()
	g := gocloak.NewClient(live.base)
	jwt, err := g.LoginClient(ctx, live.clientID, live.clientSecret, live.realm)
	if err != nil {
		b.Fatal(err)
	}
	if _, _, err := g.DecodeAccessToken(ctx, jwt.AccessToken, live.realm); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for b.Loop() {
		if _, _, err := g.DecodeAccessToken(ctx, jwt.AccessToken, live.realm); err != nil {
			b.Fatal(err)
		}
	}
}

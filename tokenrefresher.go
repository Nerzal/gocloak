package gocloak

import (
	"context"
	"sync"
	"time"
)

// RefreshConfig is the configuration for the token refresher.
type RefreshConfig struct {
	Domain, ClientID, ClientSecret, Realm string

	// The number of seconds early to refresh a token.
	EarlyRefreshSecs int
}

// NewTokenRefresher starts a background service that refreshes our jwt for us
// whenever it's going to expire. It's thread-safe.
func NewTokenRefresher(ctx context.Context, config *RefreshConfig) (*TokenRefresher, error) {
	keycloakClient := NewClient(config.Domain)
	jwt, err := keycloakClient.LoginClient(ctx, config.ClientID, config.ClientSecret, config.Realm)
	if err != nil {
		return nil, err
	}
	t := &TokenRefresher{
		ctx:      ctx,
		config:   config,
		keycloak: keycloakClient,
		jwt:      jwt,
	}
	t.startBackgroundRefresh()
	return t, nil
}

// TokenRefresher is an object that refreshes tokens in the background.
// To use it, please use NewTokenRefresher because it begins the
// background refresh process.
type TokenRefresher struct {
	ctx      context.Context
	config   *RefreshConfig
	keycloak GoCloak

	// The mu protects the jwt from a race. We need it
	// because it's accessed on one goroutine, but refreshed on another.
	mu  sync.RWMutex
	jwt *JWT
}

// AccessToken returns a jwt access token for use in client calls.
func (t *TokenRefresher) AccessToken() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.jwt.AccessToken
}

// startBackgroundRefresh begins a service that is responsible for always
// keeping the access token up-to-date.
func (t *TokenRefresher) startBackgroundRefresh() {
	// This goroutine is the only thread _writing_ the jwt, down inside
	// its token refreshing methods.
	earlyRefresh := time.Duration(t.config.EarlyRefreshSecs) * time.Second
	go func() {
		for {
			t.mu.RLock()
			expirationTimer := time.NewTimer(time.Duration(t.jwt.ExpiresIn)*time.Second - earlyRefresh)
			refreshTimer := time.NewTimer(time.Duration(t.jwt.RefreshExpiresIn)*time.Second - earlyRefresh)
			t.mu.RUnlock()

			select {
			case <-expirationTimer.C:
				if err := t.newToken(); err != nil {
					t.mu.Lock()
					t.jwt.ExpiresIn = 1
					t.mu.Unlock()
				}
				continue

			case <-refreshTimer.C:
				if err := t.refreshToken(); err != nil {
					t.mu.Lock()
					t.jwt.RefreshExpiresIn = 1
					t.mu.Unlock()
				}
				continue

			case <-t.ctx.Done():
				return
			}
		}
	}()
}

func (t *TokenRefresher) newToken() error {
	t.mu.RLock()
	jwt, err := t.keycloak.LoginClient(t.ctx, t.config.ClientID, t.config.ClientSecret, t.config.Realm)
	t.mu.RUnlock()
	if err != nil {
		return err
	}

	t.mu.Lock()
	t.jwt = jwt
	t.mu.Unlock()

	return nil
}

func (t *TokenRefresher) refreshToken() error {
	t.mu.RLock()
	jwt, err := t.keycloak.RefreshToken(t.ctx, t.jwt.RefreshToken, t.config.ClientID, t.config.ClientSecret, t.config.Realm)
	t.mu.RUnlock()
	if err != nil {
		return err
	}

	t.mu.Lock()
	t.jwt = jwt
	t.mu.Unlock()

	return nil
}

package gocloak

import (
	"context"
	"sync"
	"time"
)

type RefreshConfig struct {
	Domain, ClientId, ClientSecret, Realm string

	// The number of seconds early to refresh a token.
	EarlyRefreshSecs int
}

// newTokenRefresher starts a background service that refreshes our jwt for us
// whenever it's going to expire. It's thread-safe.
func NewTokenRefresher(ctx context.Context, config *RefreshConfig) (*TokenRefresher, error) {
	keycloakClient := NewClient(config.Domain)
	jwt, err := keycloakClient.LoginClient(ctx, config.ClientId, config.ClientSecret, config.Realm)
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

type TokenRefresher struct {
	ctx      context.Context
	config   *RefreshConfig
	keycloak GoCloak

	// The mu protects the jwt from a race. We need it
	// because it's accessed on one goroutine, but refreshed on another.
	mu  sync.RWMutex
	jwt *JWT
}

func (t *TokenRefresher) AccessToken() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.jwt.AccessToken
}

// startBackgroundRefresh begins a service that is responsible for always
// keeping the access token up-to-date.
// If it's unable to refresh or renew a token, it logs at error level and
// tries again each second until it's either successful, or it's not and
// the caller will begin to experience authorization errors.
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
	jwt, err := t.keycloak.LoginClient(t.ctx, t.config.ClientId, t.config.ClientSecret, t.config.Realm)
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
	jwt, err := t.keycloak.RefreshToken(t.ctx, t.jwt.RefreshToken, t.config.ClientId, t.config.ClientSecret, t.config.Realm)
	t.mu.RUnlock()
	if err != nil {
		return err
	}

	t.mu.Lock()
	t.jwt = jwt
	t.mu.Unlock()

	return nil
}

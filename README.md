# gocloak

[![Go Report Card](https://goreportcard.com/badge/github.com/Nerzal/gocloak)](https://goreportcard.com/report/github.com/Nerzal/gocloak)
[![Go Reference](https://pkg.go.dev/badge/github.com/Nerzal/gocloak/v15.svg)](https://pkg.go.dev/github.com/Nerzal/gocloak/v15)
[![Build Status](https://github.com/Nerzal/gocloak/workflows/Tests/badge.svg)](https://github.com/Nerzal/gocloak/actions?query=branch%3Amain+event%3Apush)
[![GitHub release](https://img.shields.io/github/tag/Nerzal/gocloak.svg)](https://github.com/Nerzal/gocloak/releases/)
[![codecov](https://codecov.io/gh/Nerzal/gocloak/branch/main/graph/badge.svg)](https://codecov.io/gh/Nerzal/gocloak)
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2FNerzal%2Fgocloak.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2FNerzal%2Fgocloak?ref=badge_shield)

gocloak is a Go client for the Keycloak admin REST API and its OpenID Connect endpoints. It logs in users and
clients, verifies tokens, and manages realms, users, groups, roles, clients, identity providers, organizations and
authorization resources. It is based on [go-keycloak](https://github.com/PhilippHeuer/go-keycloak).

gocloak needs Go 1.26 or newer and supports current Keycloak releases. CI runs the integration tests against
Keycloak 26.8.2 and the nightly build.

Questions go to the [issue tracker](https://github.com/Nerzal/gocloak/issues) or to
[#gocloak](https://gophers.slack.com/app_redirect?channel=gocloak) on the
[Gophers Slack](https://invite.slack.golangbridge.org/). For the echo framework there is
[gocloak-echo](https://github.com/Nerzal/gocloak-echo).

## Installation

```shell
go get github.com/Nerzal/gocloak/v15
```

```go
import "github.com/Nerzal/gocloak/v15"
```

Release notes for every version are on the [releases page](https://github.com/Nerzal/gocloak/releases). v15 is a
major release with breaking changes; read its notes before you upgrade from v14.

## Usage

### Log in and create a user

```go
client := gocloak.NewClient("https://keycloak.example.com")
ctx := context.Background()

token, err := client.LoginAdmin(ctx, "admin", "password", "master")
if err != nil {
	return fmt.Errorf("admin login: %w", err)
}

userID, err := client.CreateUser(ctx, token.AccessToken, "my-realm", gocloak.User{
	Username:  new("bob"),
	Email:     new("bob@example.com"),
	FirstName: new("Bob"),
	Enabled:   new(true),
})
if err != nil {
	return fmt.Errorf("create user: %w", err)
}
```

Fields of the structs you send to Keycloak (`User`, `Client`, `RealmRepresentation` and so on) and of the query
parameter structs (`GetUsersParams` and so on) are pointers. Keycloak keeps a field you leave out on an update, so a
pointer is the only way to tell "set to `false`" from "don't touch". Build them with the builtin `new(expr)`, for
example `new("bob")` or `new(int32(5))`, and read them with `gocloak.PString`, `gocloak.PBool`, `gocloak.PInt` and
friends, which return the zero value for `nil`.

Slices and maps are left out when they are `nil`. An empty slice is sent as `[]`, so
`user.RequiredActions = []string{}` clears the required actions of a user on update.

Structs that Keycloak only returns, and the token and form options, use plain values.

### Handle errors

Failed calls return a `*gocloak.APIError` with the HTTP status code and Keycloak's error message:

```go
_, err := client.GetUserByID(ctx, token.AccessToken, "my-realm", userID)

var apiErr *gocloak.APIError
if errors.As(err, &apiErr) && apiErr.Code == http.StatusNotFound {
	// the user does not exist
}
```

`apiErr.Type` is `gocloak.APIErrTypeInvalidGrant` when Keycloak rejected credentials or a refresh token. Network
errors and canceled contexts have `Code` 0 and wrap the original error, so `errors.Is(err, context.Canceled)` works.

### Validate an access token

`ValidateAccessToken` checks the signature against the realm's public keys, then `exp`, `nbf`, `iss`, `typ`, `aud` and
`azp`. ID tokens, tokens of other realms and tokens issued to other clients fail. Only the path of `iss` is compared
(`.../realms/my-realm`), so a public hostname that differs from the base URL is fine.

```go
_, claims, err := client.ValidateAccessToken(ctx, accessToken, "my-realm", gocloak.TokenValidation{
	Audience:        "my-api",        // must be in aud
	AuthorizedParty: "web-app",       // must equal azp
	Leeway:          5 * time.Second, // tolerated clock skew
})
if err != nil {
	// reject the request
}
```

You have to set at least one of `Audience` and `AuthorizedParty`. To decode into your own claims struct, use
`ValidateAccessTokenCustomClaims`.

`DecodeAccessToken` and `DecodeAccessTokenCustomClaims` check only the signature, `exp` and `nbf`. Use them when
you check the remaining claims yourself.

The public keys are cached per realm for 10 minutes (`SetCertCacheInvalidationTime` changes that). A token signed
with an unknown key id makes gocloak fetch the keys again, at most every 30 seconds, so key rotation works without a
restart. Pass only realm names from your configuration, never one read from the token: the cache keeps an entry for
every realm it has seen.

### Authorization code flow with PKCE

```go
// before redirecting to Keycloak, keep the verifier in the user's session
verifier, challenge := gocloak.NewPKCE()
// add code_challenge=<challenge>&code_challenge_method=S256 to the authorization URL

// in the redirect handler
token, err := client.GetToken(ctx, "my-realm", gocloak.TokenOptions{
	ClientID:     "web-app",
	ClientSecret: clientSecret, // empty for public clients
	GrantType:    "authorization_code",
	Code:         code,
	RedirectURI:  redirectURI,
	CodeVerifier: verifier,
})
```

### Introspect a token

```go
token, err := client.LoginClient(ctx, clientID, clientSecret, "my-realm")
if err != nil {
	return err
}

result, err := client.RetrospectToken(ctx, token.AccessToken, clientID, clientSecret, "my-realm")
if err != nil {
	return err
}
if !result.Active {
	return errors.New("token is not active")
}
```

Introspection asks Keycloak on every call, so it also sees revoked sessions. `ValidateAccessToken` works offline
with the cached keys.

### Find a client by its client ID

A client has two identifiers. `clientId` is the name you choose in the admin console; `id` is the internal ID that
Keycloak generates and that most client endpoints take. To get the `id` for a `clientId`:

```go
clients, err := client.GetClients(ctx, token.AccessToken, "my-realm", gocloak.GetClientsParams{
	ClientID: new("web-app"),
})
if err != nil {
	return "", err
}
if len(clients) == 0 {
	return "", errors.New("client not found")
}
return gocloak.PString(clients[0].ID), nil
```

### Upgrading from v14

- IDs, names and realms are path-escaped by gocloak. Pass them unescaped: code that called `url.PathEscape(roleName)`
  itself now sends a double-escaped name and gets a 404. IDs returned by `Create*` functions are unescaped too.
- `UpdateAuthenticationFlow` returns only an `error`; read the flow back with `GetAuthenticationFlow`.
- `ValidateAccessToken` rejects tokens without `exp`.

## Configuration

`NewClient` takes functional options:

```go
client := gocloak.NewClient("https://keycloak.example.com",
	gocloak.SetCertCacheInvalidationTime(30*time.Minute),
	gocloak.SetServerVersion("26.8.2"), // skips the serverinfo call in GetServerVersion
)
```

Keycloak versions before 17 served everything under `/auth`; `gocloak.SetLegacyWildFlySupport()` adds that prefix.
Other paths can be changed with `SetAuthRealms`, `SetAuthAdminRealms`, `SetTokenEndpoint` and the other `Set*`
options.

### HTTP client, timeouts and TLS

Requests go through a [resty](https://github.com/go-resty/resty) client and time out after 30 seconds. A deadline on
the `ctx` you pass also ends the request. Response bodies are limited to 100 MiB (`SetResponseBodyLimit`), and
redirects are only followed within the same scheme, host and port, so credentials never go to another server
(`SetRedirectPolicy` replaces this). `RestyClient()` returns the client for timeouts, proxies, TLS and retries:

```go
restyClient := client.RestyClient()
restyClient.SetTimeout(time.Minute)
restyClient.SetRootCertificate("/etc/ssl/keycloak-ca.pem")
```

For a local Keycloak with a self-signed certificate you can turn off certificate checks. Never do this in
production:

```go
client.RestyClient().SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
```

`RestyClient().SetDebug(true)` logs every request and response, including passwords, client secrets and tokens.
Keep it out of production too.

### Tracing

gocloak writes the OpenTelemetry trace context of `ctx` into the request headers. It uses the global propagator
(`otel.GetTextMapPropagator()`) unless you attach another one:

```go
ctx = gocloak.WithPropagator(ctx, propagation.TraceContext{})
```

### Mocking

`GoCloakIface` in [gocloak_iface.go](gocloak_iface.go) lists every method of `*gocloak.GoCloak`. Depend on the
interface in your code and pass a mock in tests.

### Inspecting models

The models have no `String()` method, so passwords and secrets in them don't end up in logs by accident. Printing
one with `fmt.Println` shows pointer addresses. Marshal it to JSON instead:

```go
out, _ := json.MarshalIndent(realm, "", "  ")
fmt.Println(string(out))
```

```json
{
  "clients": [
    {
      "name": "someClient",
      "protocolMappers": [
        {
          "config": {
            "bar": "foo",
            "ping": "pong"
          },
          "name": "someMapper"
        }
      ]
    }
  ],
  "displayName": "someRealm"
}
```

Fields that are `nil` don't appear in the output.

## Examples

- [Add client roles to a user](./examples/ADD_CLIENT_ROLE_TO_USER.md)
- [Create an LDAP user federation and sync it](./examples/USER_FEDERATION.md)
- [User federation with a group LDAP mapper](./examples/USER_FEDERATION_GROUP_LDAP_MAPPER.md)
- [User federation with a role LDAP mapper](./examples/USER_FEDERATION_ROLE_LDAP_MAPPER.md)
- [User federation with a user attribute LDAP mapper](./examples/USER_FEDERATION_USER_ATTRIBUTE_LDAP_MAPPER.md)

## Development and testing

The integration tests need a running Keycloak. `run-tests.sh` starts Keycloak and MailHog with Docker Compose,
imports the test realm from `testdata/gocloak-realm.json`, runs the tests and stops the containers again:

```shell
./run-tests.sh                          # all tests
./run-tests.sh Test_CreateUser          # one test
KEYCLOAK_VERSION=nightly ./run-tests.sh # another Keycloak image tag
```

To keep Keycloak running between test runs:

```shell
make start-keycloak
go test -run '^Test_CreateUser$' ./...
make stop-keycloak
```

The tests read their settings from `testdata/config.json`. To test against your own Keycloak, point
`GOCLOAK_TEST_CONFIG` at another config file:

```shell
export GOCLOAK_TEST_CONFIG=/path/to/config.json
```

Tests delete everything they create. Only the test user from the config file stays.

Lint with `golangci-lint run` (configuration in `.golangci.yml`). `gocloak_iface.go` is generated from `client.go`;
after changing a method, run `make generate-gocloak-interface` instead of editing it by hand.

## Contributing

Pull requests are welcome. The integration tests in `client_test.go` show how tests create and clean up their
Keycloak resources. Some older notes are in the [wiki](https://github.com/Nerzal/gocloak/wiki/Contribute).

## License

[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2FNerzal%2Fgocloak.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2FNerzal%2Fgocloak?ref=badge_large)

## Related projects

- [gocloak-echo](https://github.com/Nerzal/gocloak-echo) for the echo framework
- [GocloakSession](https://github.com/Clarilab/gocloaksession)

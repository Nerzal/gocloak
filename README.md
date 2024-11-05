# gocloak

[![codebeat badge](https://codebeat.co/badges/18a37f35-6a95-4e40-9e78-272233892332)](https://codebeat.co/projects/github-com-nerzal-gocloak-main)
[![Go Report Card](https://goreportcard.com/badge/github.com/Nerzal/gocloak)](https://goreportcard.com/report/github.com/Nerzal/gocloak)
[![Go Doc](https://godoc.org/github.com/Nerzal/gocloak?status.svg)](https://godoc.org/github.com/Nerzal/gocloak)
[![Build Status](https://github.com/Nerzal/gocloak/workflows/Tests/badge.svg)](https://github.com/Nerzal/gocloak/actions?query=branch%3Amain+event%3Apush)
[![GitHub release](https://img.shields.io/github/tag/Nerzal/gocloak.svg)](https://GitHub.com/Nerzal/gocloak/releases/)
[![codecov](https://codecov.io/gh/Nerzal/gocloak/branch/master/graph/badge.svg)](https://codecov.io/gh/Nerzal/gocloak)
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2FNerzal%2Fgocloak.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2FNerzal%2Fgocloak?ref=badge_shield)

Golang Keycloak API Package

This client is based on: [go-keycloak](https://github.com/PhilippHeuer/go-keycloak)

For Questions either raise an issue, or come to the [gopher-slack](https://invite.slack.golangbridge.org/) into the channel [#gocloak](https://gophers.slack.com/app_redirect?channel=gocloak)

If u are using the echo framework have a look at [gocloak-echo](https://github.com/Nerzal/gocloak-echo)

Benchmarks can be found [here](https://nerzal.github.io/gocloak/dev/bench/)

## Contribution

(WIP) <https://github.com/Nerzal/gocloak/wiki/Contribute>

## Changelog

For release notes please consult the specific releases [here](https://github.com/Nerzal/gocloak/releases)


## Usage

### Installation

```shell
go get github.com/Nerzal/gocloak/v13
```

### Importing

```go
 import "github.com/Nerzal/gocloak/v13"
```

### Create New User

```go
 client := gocloak.NewClient("https://mycool.keycloak.instance")
 ctx := context.Background()
 token, err := client.LoginAdmin(ctx, "user", "password", "realmName")
 if err != nil {
  panic("Something wrong with the credentials or url")
 }

 user := gocloak.User{
  FirstName: gocloak.StringP("Bob"),
  LastName:  gocloak.StringP("Uncle"),
  Email:     gocloak.StringP("something@really.wrong"),
  Enabled:   gocloak.BoolP(true),
  Username:  gocloak.StringP("CoolGuy"),
 }

 _, err = client.CreateUser(ctx, token.AccessToken, "realm", user)
 if err != nil {
  panic("Oh no!, failed to create user :(")
 }
```

### Introspect Token

```go
 client := gocloak.NewClient(hostname)
 ctx := context.Background()
 token, err := client.LoginClient(ctx, clientID, clientSecret, realm)
 if err != nil {
  panic("Login failed:"+ err.Error())
 }

 rptResult, err := client.RetrospectToken(ctx, token.AccessToken, clientID, clientSecret, realm)
 if err != nil {
  panic("Inspection failed:"+ err.Error())
 }

 if !*rptResult.Active {
  panic("Token is not active")
 }

 permissions := rptResult.Permissions
 // Do something with the permissions ;)
```

### Get Client id

Client has 2 identity fields- `id` and `clientId` and both are unique in one realm.

- `id` is generated automatically by Keycloak.
- `clientId` is configured by users in `Add client` page.

To get the `clientId` from `id`, use `GetClients` method with `GetClientsParams{ClientID: &clientName}`.

```go
 clients, err := c.Client.GetClients(
  c.Ctx,
  c.JWT.AccessToken,
  c.Realm,
  gocloak.GetClientsParams{
   ClientID: &clientName,
  },
 )
 if err != nil {
  panic("List clients failed:"+ err.Error())
 }
 for _, client := range clients {
  return *client.ID, nil
 }
```

## Features

[GoCloakIface](gocloak_iface.go) holds all methods a client should fulfil.


## Configure gocloak to skip TLS Insecure Verification

```go
    client := gocloak.NewClient(serverURL)
    restyClient := client.RestyClient()
    restyClient.SetDebug(true)
    restyClient.SetTLSClientConfig(&tls.Config{ InsecureSkipVerify: true })
```

## developing & testing

For local testing you need to start a docker container. Simply run following commands prior to starting the tests:

```shell
docker pull quay.io/keycloak/keycloak
docker run -d \
 -e KEYCLOAK_USER=admin \
 -e KEYCLOAK_PASSWORD=secret \
 -e KEYCLOAK_IMPORT=/tmp/gocloak-realm.json \
 -v "`pwd`/testdata/gocloak-realm.json:/tmp/gocloak-realm.json" \
 -p 8080:8080 \
 --name gocloak-test \
 quay.io/keycloak/keycloak:latest -Dkeycloak.profile.feature.upload_scripts=enabled

go test
```

Or you can run with docker compose using the run-tests script

```shell
./run-tests.sh
```

or

```shell
./run-tests.sh <TestCase>
```

Or you can run the tests on you own keycloak:

```shell
export GOCLOAK_TEST_CONFIG=/path/to/gocloak/config.json
```

All resources created as a result of unit tests will be deleted, except for the test user defined in the configuration file.

To remove running docker container after completion of tests:

```shell
docker stop gocloak-test
docker rm gocloak-test
```

### Inspecting custom types

The custom types contain many pointers, so printing them yields mostly pointer values, which aren't much help when debugging your application. For example

```go
someRealmRepresentation := gocloak.RealmRepresentation{
   <snip>
}

fmt.Println(someRealmRepresentation)

```

yields a large set of pointer values

```bash
{<nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> 0xc00000e960 <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> 0xc000093cf0 <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> null <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil>}
```

For convenience, the ```String()``` interface has been added so you can easily see the contents, even for nested custom types. For example,

```go
fmt.Println(someRealmRepresentation.String())
```

yields

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
  },
  {
   "name": "AnotherClient"
  }
 ],
 "displayName": "someRealm"
}
```

Note that empty parameters are not included, because of the use of ```omitempty``` in the type definitions.

## Examples

* [Add client role to user](./examples/ADD_CLIENT_ROLE_TO_USER.md)

* [Create User Federation & Sync](./examples/USER_FEDERATION.md)

* [Create User Federation & Sync with group ldap mapper](./examples/USER_FEDERATION_GROUP_LDAP_MAPPER.md)

* [Create User Federation & Sync with role ldap mapper](./examples/USER_FEDERATION_ROLE_LDAP_MAPPER.md)

* [Create User Federation & Sync with user attribute ldap mapper](./examples/USER_FEDERATION_USER_ATTRIBUTE_LDAP_MAPPER.md)

## License

[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2FNerzal%2Fgocloak.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2FNerzal%2Fgocloak?ref=badge_large)

## Related Projects

[GocloakSession](https://github.com/Clarilab/gocloaksession)

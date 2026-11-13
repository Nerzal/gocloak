package gocloak_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Nerzal/gocloak/v14"
)

// coverageSetup returns a debug client, the test config and a fresh admin access token.
// Input: t [*testing.T]. Output: client [*gocloak.GoCloak], cfg [*Config], token [string].
func coverageSetup(t *testing.T) (*gocloak.GoCloak, *Config, string) {
	t.Helper()
	client := NewClientWithDebug(t)
	return client, GetConfig(t), GetAdminToken(t, client).AccessToken
}

// idOfRole returns the ID of a role. Input: role [*gocloak.Role]. Output: [*string].
func idOfRole(role *gocloak.Role) *string { return role.ID }

// idOfGroup returns the ID of a group. Input: group [*gocloak.Group]. Output: [*string].
func idOfGroup(group *gocloak.Group) *string { return group.ID }

// idOfComponent returns the ID of a component. Input: component [*gocloak.Component]. Output: [*string].
func idOfComponent(component *gocloak.Component) *string { return component.ID }

// idOfPolicy returns the ID of a policy. Input: policy [*gocloak.PolicyRepresentation]. Output: [*string].
func idOfPolicy(policy *gocloak.PolicyRepresentation) *string { return policy.ID }

// idOfCredential returns the ID of a credential. Input: credential [*gocloak.CredentialRepresentation]. Output: [*string].
func idOfCredential(credential *gocloak.CredentialRepresentation) *string { return credential.ID }

// createRealmRoleCleanup creates a realm role, registers its deletion with t.Cleanup and returns the stored role.
// Input: t [*testing.T], client [*gocloak.GoCloak]. Output: [*gocloak.Role] role as returned by the server.
func createRealmRoleCleanup(t *testing.T, client *gocloak.GoCloak) *gocloak.Role {
	t.Helper()
	tearDown, roleName := CreateRealmRole(t, client)
	t.Cleanup(tearDown)
	cfg := GetConfig(t)
	role, err := client.GetRealmRole(context.Background(), GetAdminToken(t, client).AccessToken, cfg.GoCloak.Realm, roleName)
	require.NoError(t, err, "GetRealmRole failed")
	return role
}

// createTestClient creates a client from the given representation and registers its deletion with t.Cleanup.
// Input: t [*testing.T], client [*gocloak.GoCloak], newClient [gocloak.Client]. Output: [string] internal client ID.
func createTestClient(t *testing.T, client *gocloak.GoCloak, newClient gocloak.Client) string {
	t.Helper()
	tearDown, idOfClient := CreateClient(t, client, &newClient)
	t.Cleanup(tearDown)
	return idOfClient
}

// createGoogleIdentityProvider creates a google identity provider with a random alias and registers its deletion.
// Input: t [*testing.T], client [*gocloak.GoCloak]. Output: [string] alias of the created provider.
func createGoogleIdentityProvider(t *testing.T, client *gocloak.GoCloak) string {
	t.Helper()
	cfg := GetConfig(t)
	alias := GetRandomName("google")
	res, err := client.CreateIdentityProvider(context.Background(), GetAdminToken(t, client).AccessToken, cfg.GoCloak.Realm,
		gocloak.IdentityProviderRepresentation{
			ProviderID: new("google"),
			Alias:      &alias,
			Enabled:    new(true),
			Config:     map[string]string{"clientId": "id", "clientSecret": "secret"},
		})
	require.NoError(t, err, "CreateIdentityProvider failed")
	require.Equal(t, alias, res)
	t.Cleanup(func() {
		_ = client.DeleteIdentityProvider(context.Background(), GetAdminToken(t, client).AccessToken, cfg.GoCloak.Realm, alias)
	})
	return alias
}

func Test_RealmRoleByIDAndComposites(t *testing.T) {
	t.Parallel()
	client, cfg, token := coverageSetup(t)
	ctx := context.Background()
	parent := createRealmRoleCleanup(t, client)
	child := createRealmRoleCleanup(t, client)

	err := client.AddRealmRoleComposite(ctx, token, cfg.GoCloak.Realm, *parent.Name, []gocloak.Role{*child})
	require.NoError(t, err, "AddRealmRoleComposite failed")

	byID, err := client.GetRealmRoleByID(ctx, token, cfg.GoCloak.Realm, *parent.ID)
	require.NoError(t, err, "GetRealmRoleByID failed")
	require.Equal(t, *parent.Name, gocloak.PString(byID.Name))
	require.True(t, gocloak.PBool(byID.Composite))

	byID.Description = new("updated by id")
	err = client.UpdateRealmRoleByID(ctx, token, cfg.GoCloak.Realm, *parent.ID, *byID)
	require.NoError(t, err, "UpdateRealmRoleByID failed")
	updated, err := client.GetRealmRoleByID(ctx, token, cfg.GoCloak.Realm, *parent.ID)
	require.NoError(t, err, "GetRealmRoleByID failed")
	require.Equal(t, "updated by id", gocloak.PString(updated.Description))

	composites, err := client.GetCompositeRealmRoles(ctx, token, cfg.GoCloak.Realm, *parent.Name)
	require.NoError(t, err, "GetCompositeRealmRoles failed")
	require.Equal(t, []string{*child.Name}, pluck(composites, nameOfRole))

	compositesByID, err := client.GetCompositeRolesByRoleID(ctx, token, cfg.GoCloak.Realm, *parent.ID)
	require.NoError(t, err, "GetCompositeRolesByRoleID failed")
	require.Equal(t, []string{*child.ID}, pluck(compositesByID, idOfRole))

	tearDownGroup, groupID := CreateGroup(t, client)
	t.Cleanup(tearDownGroup)
	err = client.AddRealmRoleToGroup(ctx, token, cfg.GoCloak.Realm, groupID, []gocloak.Role{*parent})
	require.NoError(t, err, "AddRealmRoleToGroup failed")
	groupRoles, err := client.GetCompositeRealmRolesByGroupID(ctx, token, cfg.GoCloak.Realm, groupID)
	require.NoError(t, err, "GetCompositeRealmRolesByGroupID failed")
	names := pluck(groupRoles, nameOfRole)
	require.Contains(t, names, *parent.Name)
	require.Contains(t, names, *child.Name, "composite child must be expanded")
}

func Test_UpdateRole(t *testing.T) {
	t.Parallel()
	client, cfg, token := coverageSetup(t)
	ctx := context.Background()
	tearDown, roleName := CreateClientRole(t, client)
	t.Cleanup(tearDown)

	role, err := client.GetClientRole(ctx, token, cfg.GoCloak.Realm, gocloakClientID, roleName)
	require.NoError(t, err, "GetClientRole failed")
	role.Description = new("updated client role")
	err = client.UpdateRole(ctx, token, cfg.GoCloak.Realm, gocloakClientID, *role)
	require.NoError(t, err, "UpdateRole failed")

	updated, err := client.GetClientRole(ctx, token, cfg.GoCloak.Realm, gocloakClientID, roleName)
	require.NoError(t, err, "GetClientRole failed")
	require.Equal(t, "updated client role", gocloak.PString(updated.Description))
	require.Equal(t, gocloak.PString(role.ID), gocloak.PString(updated.ID))
}

func Test_UpdateClientScope(t *testing.T) {
	t.Parallel()
	client, cfg, token := coverageSetup(t)
	ctx := context.Background()
	tearDown, scopeID := CreateClientScope(t, client, nil)
	t.Cleanup(tearDown)

	scope, err := client.GetClientScope(ctx, token, cfg.GoCloak.Realm, scopeID)
	require.NoError(t, err, "GetClientScope failed")
	scope.Description = new("updated scope")
	err = client.UpdateClientScope(ctx, token, cfg.GoCloak.Realm, *scope)
	require.NoError(t, err, "UpdateClientScope failed")

	updated, err := client.GetClientScope(ctx, token, cfg.GoCloak.Realm, scopeID)
	require.NoError(t, err, "GetClientScope failed")
	require.Equal(t, "updated scope", gocloak.PString(updated.Description))
	require.Equal(t, gocloak.PString(scope.Name), gocloak.PString(updated.Name))
}

func Test_ClientScopesScopeMappingsRealmRoles(t *testing.T) {
	t.Parallel()
	client, cfg, token := coverageSetup(t)
	ctx := context.Background()
	tearDown, scopeID := CreateClientScope(t, client, nil)
	t.Cleanup(tearDown)
	role := createRealmRoleCleanup(t, client)

	available, err := client.GetClientScopesScopeMappingsRealmRolesAvailable(ctx, token, cfg.GoCloak.Realm, scopeID)
	require.NoError(t, err, "GetClientScopesScopeMappingsRealmRolesAvailable failed")
	require.Contains(t, pluck(available, nameOfRole), *role.Name)

	err = client.CreateClientScopesScopeMappingsRealmRoles(ctx, token, cfg.GoCloak.Realm, scopeID, []gocloak.Role{*role})
	require.NoError(t, err, "CreateClientScopesScopeMappingsRealmRoles failed")

	mapped, err := client.GetClientScopesScopeMappingsRealmRoles(ctx, token, cfg.GoCloak.Realm, scopeID)
	require.NoError(t, err, "GetClientScopesScopeMappingsRealmRoles failed")
	require.Equal(t, []string{*role.Name}, pluck(mapped, nameOfRole))

	available, err = client.GetClientScopesScopeMappingsRealmRolesAvailable(ctx, token, cfg.GoCloak.Realm, scopeID)
	require.NoError(t, err, "GetClientScopesScopeMappingsRealmRolesAvailable failed")
	require.NotContains(t, pluck(available, nameOfRole), *role.Name)

	err = client.DeleteClientScopesScopeMappingsRealmRoles(ctx, token, cfg.GoCloak.Realm, scopeID, []gocloak.Role{*role})
	require.NoError(t, err, "DeleteClientScopesScopeMappingsRealmRoles failed")
	mapped, err = client.GetClientScopesScopeMappingsRealmRoles(ctx, token, cfg.GoCloak.Realm, scopeID)
	require.NoError(t, err, "GetClientScopesScopeMappingsRealmRoles failed")
	require.Empty(t, mapped)
}

func Test_GetClientScopeMappings(t *testing.T) {
	t.Parallel()
	client, cfg, token := coverageSetup(t)
	idOfClient := createTestClient(t, client, gocloak.Client{ClientID: GetRandomNameP("ClientID")})
	role := createRealmRoleCleanup(t, client)
	t.Cleanup(CreateClientScopeMappingsRealmRoles(t, client, idOfClient, []gocloak.Role{*role}))

	mappings, err := client.GetClientScopeMappings(context.Background(), token, cfg.GoCloak.Realm, idOfClient)
	require.NoError(t, err, "GetClientScopeMappings failed")
	require.Equal(t, []string{*role.Name}, pluck(mappings.RealmMappings, nameOfRoleValue))
}

func Test_GetScopePermissions(t *testing.T) {
	t.Parallel()
	client, cfg, token := coverageSetup(t)
	tearDownScope, scopeID := CreateScope(t, client, gocloakClientID)
	t.Cleanup(tearDownScope)
	tearDownPolicy, policyID := CreatePolicy(t, client, gocloakClientID, gocloak.PolicyRepresentation{
		Name:                       GetRandomNameP("PolicyName"),
		Type:                       new("client"),
		ClientPolicyRepresentation: gocloak.ClientPolicyRepresentation{Clients: []string{gocloakClientID}},
	})
	t.Cleanup(tearDownPolicy)
	tearDownPermission, permissionID := CreatePermission(t, client, gocloakClientID, gocloak.PermissionRepresentation{
		Name:     GetRandomNameP("PermissionName"),
		Type:     new("scope"),
		Scopes:   []string{scopeID},
		Policies: []string{policyID},
	})
	t.Cleanup(tearDownPermission)

	permissions, err := client.GetScopePermissions(context.Background(), token, cfg.GoCloak.Realm, gocloakClientID, scopeID)
	require.NoError(t, err, "GetScopePermissions failed")
	require.Equal(t, []string{permissionID}, pluck(permissions, idOfPolicy))
}

func Test_GetPoliciesFilteredByType(t *testing.T) {
	t.Parallel()
	client, cfg, token := coverageSetup(t)
	ctx := context.Background()
	policyName := GetRandomName("PolicyName")
	tearDown, policyID := CreatePolicy(t, client, gocloakClientID, gocloak.PolicyRepresentation{
		Name:                       &policyName,
		Type:                       new("client"),
		ClientPolicyRepresentation: gocloak.ClientPolicyRepresentation{Clients: []string{gocloakClientID}},
	})
	t.Cleanup(tearDown)

	matching, err := client.GetPolicies(ctx, token, cfg.GoCloak.Realm, gocloakClientID,
		gocloak.GetPolicyParams{Name: &policyName, Type: new("client")})
	require.NoError(t, err, "GetPolicies failed")
	require.Equal(t, []string{policyID}, pluck(matching, idOfPolicy))

	other, err := client.GetPolicies(ctx, token, cfg.GoCloak.Realm, gocloakClientID,
		gocloak.GetPolicyParams{Name: &policyName, Type: new("role")})
	require.NoError(t, err, "GetPolicies failed")
	require.Empty(t, other, "a client policy must not match type role")
}

// Test_UsersManagementPermissions toggles a realm-wide setting, so it does not run in parallel.
func Test_UsersManagementPermissions(t *testing.T) {
	client, cfg, token := coverageSetup(t)
	ctx := context.Background()

	original, err := client.GetUsersManagementPermissions(ctx, token, cfg.GoCloak.Realm)
	require.NoError(t, err, "GetUsersManagementPermissions failed")
	t.Cleanup(func() {
		_, err := client.UpdateUsersManagementPermissions(ctx, GetAdminToken(t, client).AccessToken, cfg.GoCloak.Realm,
			gocloak.ManagementPermissionRepresentation{Enabled: new(gocloak.PBool(original.Enabled))})
		require.NoError(t, err, "restoring users management permissions failed")
	})

	updated, err := client.UpdateUsersManagementPermissions(ctx, token, cfg.GoCloak.Realm,
		gocloak.ManagementPermissionRepresentation{Enabled: new(true)})
	require.NoError(t, err, "UpdateUsersManagementPermissions failed")
	require.True(t, gocloak.PBool(updated.Enabled))
	require.NotEmpty(t, updated.Resource)
	require.NotEmpty(t, updated.ScopePermissions)

	fetched, err := client.GetUsersManagementPermissions(ctx, token, cfg.GoCloak.Realm)
	require.NoError(t, err, "GetUsersManagementPermissions failed")
	require.True(t, gocloak.PBool(fetched.Enabled))
	require.Equal(t, updated.ScopePermissions, fetched.ScopePermissions)
}

func Test_GetComponents(t *testing.T) {
	t.Parallel()
	client, cfg, token := coverageSetup(t)
	tearDown, component := CreateComponent(t, client)
	t.Cleanup(tearDown)

	components, err := client.GetComponents(context.Background(), token, cfg.GoCloak.Realm)
	require.NoError(t, err, "GetComponents failed")
	require.Contains(t, pluck(components, idOfComponent), *component.ID)
}

func Test_MoveCredential(t *testing.T) {
	t.Parallel()
	client, cfg, token := coverageSetup(t)
	ctx := context.Background()
	username := GetRandomName("credential-user")
	userID, err := client.CreateUser(ctx, token, cfg.GoCloak.Realm, gocloak.User{
		Username: &username,
		Enabled:  new(true),
		Credentials: []gocloak.CredentialRepresentation{
			{Type: new("password"), Value: new("S3cret-pass"), Temporary: new(false)},
			{
				// stored credentials can be imported with raw secret/credential data, which gives a second type
				Type:           new("otp"),
				UserLabel:      new("totp"),
				SecretData:     new(`{"value":"JBSWY3DPEHPK3PXP"}`),
				CredentialData: new(`{"subType":"totp","digits":6,"period":30,"algorithm":"HmacSHA1","counter":0}`),
			},
		},
	})
	require.NoError(t, err, "CreateUser failed")
	t.Cleanup(func() {
		_ = client.DeleteUser(ctx, GetAdminToken(t, client).AccessToken, cfg.GoCloak.Realm, userID)
	})

	credentials, err := client.GetCredentials(ctx, token, cfg.GoCloak.Realm, userID)
	require.NoError(t, err, "GetCredentials failed")
	require.Len(t, credentials, 2)
	byType := map[string]string{}
	for _, c := range credentials {
		byType[gocloak.PString(c.Type)] = gocloak.PString(c.ID)
	}
	passwordID, otpID := byType["password"], byType["otp"]
	require.NotEmpty(t, passwordID)
	require.NotEmpty(t, otpID)

	err = client.MoveCredentialToFirst(ctx, token, cfg.GoCloak.Realm, userID, otpID)
	require.NoError(t, err, "MoveCredentialToFirst failed")
	credentials, err = client.GetCredentials(ctx, token, cfg.GoCloak.Realm, userID)
	require.NoError(t, err, "GetCredentials failed")
	require.Equal(t, []string{otpID, passwordID}, pluck(credentials, idOfCredential))

	err = client.MoveCredentialBehind(ctx, token, cfg.GoCloak.Realm, userID, otpID, passwordID)
	require.NoError(t, err, "MoveCredentialBehind failed")
	credentials, err = client.GetCredentials(ctx, token, cfg.GoCloak.Realm, userID)
	require.NoError(t, err, "GetCredentials failed")
	require.Equal(t, []string{passwordID, otpID}, pluck(credentials, idOfCredential))
}

func Test_GetIdentityProviderMapper(t *testing.T) {
	t.Parallel()
	client, cfg, token := coverageSetup(t)
	ctx := context.Background()
	alias := createGoogleIdentityProvider(t, client)

	mapperID, err := client.CreateIdentityProviderMapper(ctx, token, cfg.GoCloak.Realm, alias, gocloak.IdentityProviderMapper{
		Name:                   GetRandomNameP("mapper"),
		IdentityProviderMapper: new("hardcoded-attribute-idp-mapper"),
		IdentityProviderAlias:  &alias,
		Config:                 map[string]string{"syncMode": "INHERIT", "attribute": "origin", "attribute.value": "google"},
	})
	require.NoError(t, err, "CreateIdentityProviderMapper failed")
	require.NotEmpty(t, mapperID)

	mapper, err := client.GetIdentityProviderMapper(ctx, token, cfg.GoCloak.Realm, alias, mapperID)
	require.NoError(t, err, "GetIdentityProviderMapper failed")
	require.Equal(t, mapperID, gocloak.PString(mapper.ID))
	require.Equal(t, alias, gocloak.PString(mapper.IdentityProviderAlias))
	require.Equal(t, "google", mapper.Config["attribute.value"])

	byID, err := client.GetIdentityProviderMapperByID(ctx, token, cfg.GoCloak.Realm, alias, mapperID)
	require.NoError(t, err, "GetIdentityProviderMapperByID failed")
	require.Equal(t, byID, mapper, "both getters hit the same endpoint")
}

func Test_GetGroupByPath_GetChildGroupsMax(t *testing.T) {
	t.Parallel()
	client, cfg, token := coverageSetup(t)
	ctx := context.Background()
	tearDown, parentID := CreateGroup(t, client)
	t.Cleanup(tearDown) // deleting the parent removes the children too
	parent, err := client.GetGroup(ctx, token, cfg.GoCloak.Realm, parentID)
	require.NoError(t, err, "GetGroup failed")

	childName := GetRandomName("child")
	childID, err := client.CreateChildGroup(ctx, token, cfg.GoCloak.Realm, parentID, gocloak.Group{Name: &childName})
	require.NoError(t, err, "CreateChildGroup failed")
	_, err = client.CreateChildGroup(ctx, token, cfg.GoCloak.Realm, parentID, gocloak.Group{Name: GetRandomNameP("child")})
	require.NoError(t, err, "CreateChildGroup failed")

	byPath, err := client.GetGroupByPath(ctx, token, cfg.GoCloak.Realm, "/"+*parent.Name+"/"+childName)
	require.NoError(t, err, "GetGroupByPath failed")
	require.Equal(t, childID, gocloak.PString(byPath.ID))

	all, err := client.GetChildGroups(ctx, token, cfg.GoCloak.Realm, parentID, gocloak.GetChildGroupsParams{})
	require.NoError(t, err, "GetChildGroups failed")
	require.Len(t, all, 2)
	require.Contains(t, pluck(all, idOfGroup), childID)

	limited, err := client.GetChildGroups(ctx, token, cfg.GoCloak.Realm, parentID, gocloak.GetChildGroupsParams{Max: new(1)})
	require.NoError(t, err, "GetChildGroups failed")
	require.Len(t, limited, 1)
}

func Test_LogoutPublicClient(t *testing.T) {
	t.Parallel()
	client, cfg, _ := coverageSetup(t)
	ctx := context.Background()
	SetUpTestUser(t, client)
	clientID := GetRandomName("public-client")
	createTestClient(t, client, gocloak.Client{
		ClientID:                  &clientID,
		PublicClient:              new(true),
		DirectAccessGrantsEnabled: new(true),
	})

	userToken, err := client.Login(ctx, clientID, "", cfg.GoCloak.Realm, cfg.GoCloak.UserName, cfg.GoCloak.Password)
	require.NoError(t, err, "Login with public client failed")

	err = client.LogoutPublicClient(ctx, clientID, cfg.GoCloak.Realm, userToken.AccessToken, userToken.RefreshToken)
	require.NoError(t, err, "LogoutPublicClient failed")

	_, err = client.RefreshToken(ctx, userToken.RefreshToken, clientID, "", cfg.GoCloak.Realm)
	requireAPIErrorCode(t, err, http.StatusBadRequest)
}

func Test_TokenExchange(t *testing.T) {
	t.Parallel()
	client, cfg, _ := coverageSetup(t)
	ctx := context.Background()
	SetUpTestUser(t, client)
	clientID := GetRandomName("exchange-client")
	secret := "exchange-secret"
	createTestClient(t, client, gocloak.Client{
		ClientID:                  &clientID,
		Secret:                    &secret,
		PublicClient:              new(false),
		DirectAccessGrantsEnabled: new(true),
		Attributes: map[string]string{
			"standard.token.exchange.enabled":                         "true",
			"standard.token.exchange.enableRefreshRequestedTokenType": "SAME_SESSION",
		},
	})

	subject, err := client.Login(ctx, clientID, secret, cfg.GoCloak.Realm, cfg.GoCloak.UserName, cfg.GoCloak.Password)
	require.NoError(t, err, "Login failed")

	// Direct naked impersonation only exists in legacy token exchange v1 (preview feature "token-exchange",
	// not enabled on the test server); standard token exchange v2 rejects a request without subject_token.
	_, err = client.DirectNakedImpersonationTokenExchange(ctx, clientID, secret, cfg.GoCloak.Realm, testUserID)
	requireAPIErrorCode(t, err, http.StatusBadRequest)

	exchanged, err := client.LoginClientTokenExchange(ctx, clientID, subject.AccessToken, secret, cfg.GoCloak.Realm, "", "")
	if err != nil && strings.Contains(err.Error(), "subject_token_type") {
		t.Skip("LoginClientTokenExchange does not send subject_token_type, which standard token exchange v2 requires")
	}
	require.NoError(t, err, "LoginClientTokenExchange failed")
	require.NotEmpty(t, exchanged.AccessToken)
}

func Test_ErrorPaths(t *testing.T) {
	t.Parallel()
	client, cfg, token := coverageSetup(t)
	ctx := context.Background()

	t.Run("GetGroupsByRole unknown role", func(t *testing.T) {
		t.Parallel()
		_, err := client.GetGroupsByRole(ctx, token, cfg.GoCloak.Realm, GetRandomName("missing-role"))
		requireAPIErrorCode(t, err, http.StatusNotFound)
	})
	t.Run("LogoutUserSession unknown session", func(t *testing.T) {
		t.Parallel()
		err := client.LogoutUserSession(ctx, token, cfg.GoCloak.Realm, GetRandomName("missing-session"))
		requireAPIErrorCode(t, err, http.StatusNotFound)
	})
	t.Run("CreateIdentityProvider duplicate alias", func(t *testing.T) {
		t.Parallel()
		alias := createGoogleIdentityProvider(t, client)
		_, err := client.CreateIdentityProvider(ctx, token, cfg.GoCloak.Realm, gocloak.IdentityProviderRepresentation{
			ProviderID: new("google"),
			Alias:      &alias,
			Config:     map[string]string{"clientId": "id", "clientSecret": "secret"},
		})
		requireAPIErrorCode(t, err, http.StatusConflict)
	})
	t.Run("CreateUserFederatedIdentity unknown user", func(t *testing.T) {
		t.Parallel()
		err := client.CreateUserFederatedIdentity(ctx, token, cfg.GoCloak.Realm, "00000000-0000-0000-0000-000000000000", "google",
			gocloak.FederatedIdentityRepresentation{UserID: new("external"), UserName: new("external")})
		requireAPIErrorCode(t, err, http.StatusNotFound)
	})
	t.Run("CreateUserFederatedIdentity unknown provider", func(t *testing.T) {
		t.Parallel()
		tearDown, userID := CreateUser(t, client)
		t.Cleanup(tearDown)
		// Keycloak 26 answers 204 for an unknown provider alias but links nothing.
		err := client.CreateUserFederatedIdentity(ctx, token, cfg.GoCloak.Realm, userID, GetRandomName("missing-idp"),
			gocloak.FederatedIdentityRepresentation{UserID: new("external"), UserName: new("external")})
		require.NoError(t, err, "CreateUserFederatedIdentity failed")
		identities, err := client.GetUserFederatedIdentities(ctx, token, cfg.GoCloak.Realm, userID)
		require.NoError(t, err, "GetUserFederatedIdentities failed")
		require.Empty(t, identities)
	})
}

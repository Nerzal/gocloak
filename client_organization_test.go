package gocloak_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Nerzal/gocloak/v14"
)

// idOfMember returns the ID of an organization member.
// Input: member [*gocloak.MemberRepresentation]. Output: [*string].
func idOfMember(member *gocloak.MemberRepresentation) *string { return member.ID }

// aliasOfIdentityProvider returns the alias of an identity provider.
// Input: idp [*gocloak.IdentityProviderRepresentation]. Output: [*string].
func aliasOfIdentityProvider(idp *gocloak.IdentityProviderRepresentation) *string { return idp.Alias }

// addOrganizationMember creates a user, adds it to the organization and registers cleanup for the user.
// Input: t [*testing.T], client [gocloak.GoCloakIface], token [string], orgID [string]. Output: [string] user ID.
func addOrganizationMember(t *testing.T, client gocloak.GoCloakIface, token, orgID string) string {
	t.Helper()
	cfg := GetConfig(t)
	tearDownUser, userID := CreateUser(t, client)
	t.Cleanup(tearDownUser)
	err := client.AddUserToOrganization(context.Background(), token, cfg.GoCloak.Realm, orgID, userID)
	require.NoError(t, err, "AddUserToOrganization failed")
	return userID
}

// createOrganizationGroup creates a top-level group in the organization and registers its cleanup.
// Input: t [*testing.T], client [gocloak.GoCloakIface], token [string], orgID [string].
// Output: [string] group ID, [string] group name.
func createOrganizationGroup(t *testing.T, client gocloak.GoCloakIface, token, orgID string) (groupID, name string) {
	t.Helper()
	cfg := GetConfig(t)
	name = GetRandomName("OrgGroup")
	groupID, err := client.AddOrganizationGroup(context.Background(), token, cfg.GoCloak.Realm, orgID, gocloak.Group{Name: &name})
	require.NoError(t, err, "AddOrganizationGroup failed")
	require.NotEmpty(t, groupID)
	t.Cleanup(func() {
		// deleting the organization removes its groups too, so a 404 here is fine
		_ = client.DeleteOrganizationGroup(context.Background(), token, cfg.GoCloak.Realm, orgID, groupID)
	})
	return groupID, name
}

// createOrganizationTestIdentityProvider creates an OIDC identity provider with a random alias and registers its cleanup.
// Input: t [*testing.T], client [gocloak.GoCloakIface], token [string]. Output: [string] alias.
func createOrganizationTestIdentityProvider(t *testing.T, client gocloak.GoCloakIface, token string) string {
	t.Helper()
	cfg := GetConfig(t)
	alias := strings.ToLower(GetRandomName("orgidp"))
	_, err := client.CreateIdentityProvider(context.Background(), token, cfg.GoCloak.Realm, gocloak.IdentityProviderRepresentation{
		ProviderID: new("oidc"),
		Alias:      &alias,
		Enabled:    new(true),
		Config: map[string]string{
			"clientId":         "org-idp-client",
			"clientSecret":     "org-idp-secret",
			"authorizationUrl": "http://localhost:8080/auth",
			"tokenUrl":         "http://localhost:8080/token",
		},
	})
	require.NoError(t, err, "CreateIdentityProvider failed")
	t.Cleanup(func() {
		_ = client.DeleteIdentityProvider(context.Background(), token, cfg.GoCloak.Realm, alias)
	})
	return alias
}

func Test_GetOrganizationCount(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)
	ctx := context.Background()

	tearDown, orgID := CreateOrganization(t, client)
	t.Cleanup(tearDown)

	org, err := client.GetOrganizationByID(ctx, token.AccessToken, cfg.GoCloak.Realm, orgID)
	require.NoError(t, err, "GetOrganizationByID failed")

	count, err := client.GetOrganizationCount(ctx, token.AccessToken, cfg.GoCloak.Realm,
		gocloak.GetOrganizationCountParams{Search: org.Name, Exact: new(true)})
	require.NoError(t, err, "GetOrganizationCount failed")
	require.Equal(t, int64(1), count)

	count, err = client.GetOrganizationCount(ctx, token.AccessToken, cfg.GoCloak.Realm,
		gocloak.GetOrganizationCountParams{Search: GetRandomNameP("no-such-org"), Exact: new(true)})
	require.NoError(t, err, "GetOrganizationCount failed")
	require.Equal(t, int64(0), count)
}

func Test_GetOrganizationMemberOrganizations(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)
	ctx := context.Background()

	tearDown, orgID := CreateOrganization(t, client)
	t.Cleanup(tearDown)
	userID := addOrganizationMember(t, client, token.AccessToken, orgID)

	orgs, err := client.GetOrganizationMemberOrganizations(ctx, token.AccessToken, cfg.GoCloak.Realm, orgID, userID)
	require.NoError(t, err, "GetOrganizationMemberOrganizations failed")
	require.Contains(t, pluck(orgs, idOfOrganization), orgID)
}

func Test_GetOrganizationMemberGroups(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)
	ctx := context.Background()

	tearDown, orgID := CreateOrganization(t, client)
	t.Cleanup(tearDown)
	userID := addOrganizationMember(t, client, token.AccessToken, orgID)
	groupID, name := createOrganizationGroup(t, client, token.AccessToken, orgID)

	err := client.AddMemberToOrganizationGroup(ctx, token.AccessToken, cfg.GoCloak.Realm, orgID, groupID, userID)
	require.NoError(t, err, "AddMemberToOrganizationGroup failed")

	groups, err := client.GetOrganizationMemberGroups(ctx, token.AccessToken, cfg.GoCloak.Realm, orgID, userID, true)
	require.NoError(t, err, "GetOrganizationMemberGroups failed")
	require.Contains(t, pluck(groups, idOfGroup), groupID)
	require.Contains(t, pluck(groups, func(g *gocloak.Group) *string { return g.Name }), name)
}

func Test_OrganizationIdentityProviders(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)
	ctx := context.Background()

	tearDown, orgID := CreateOrganization(t, client)
	t.Cleanup(tearDown)
	alias := createOrganizationTestIdentityProvider(t, client, token.AccessToken)

	err := client.AddIdentityProviderToOrganization(ctx, token.AccessToken, cfg.GoCloak.Realm, orgID, alias)
	require.NoError(t, err, "AddIdentityProviderToOrganization failed")
	t.Cleanup(func() {
		_ = client.RemoveIdentityProviderFromOrganization(ctx, token.AccessToken, cfg.GoCloak.Realm, orgID, alias)
	})

	idps, err := client.GetOrganizationIdentityProviders(ctx, token.AccessToken, cfg.GoCloak.Realm, orgID)
	require.NoError(t, err, "GetOrganizationIdentityProviders failed")
	require.Equal(t, []string{alias}, pluck(idps, aliasOfIdentityProvider))

	idp, err := client.GetOrganizationIdentityProvider(ctx, token.AccessToken, cfg.GoCloak.Realm, orgID, alias)
	require.NoError(t, err, "GetOrganizationIdentityProvider failed")
	require.Equal(t, alias, gocloak.PString(idp.Alias))
	require.Equal(t, "oidc", gocloak.PString(idp.ProviderID))

	err = client.RemoveIdentityProviderFromOrganization(ctx, token.AccessToken, cfg.GoCloak.Realm, orgID, alias)
	require.NoError(t, err, "RemoveIdentityProviderFromOrganization failed")

	idps, err = client.GetOrganizationIdentityProviders(ctx, token.AccessToken, cfg.GoCloak.Realm, orgID)
	require.NoError(t, err, "GetOrganizationIdentityProviders failed")
	require.NotContains(t, pluck(idps, aliasOfIdentityProvider), alias)

	_, err = client.GetOrganizationIdentityProvider(ctx, token.AccessToken, cfg.GoCloak.Realm, orgID, alias)
	requireAPIErrorCode(t, err, http.StatusNotFound)
}

func Test_OrganizationInvitations(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)
	ctx := context.Background()

	tearDown, orgID := CreateOrganization(t, client)
	t.Cleanup(tearDown)

	email := strings.ToLower(GetRandomName("invitee")) + "@localhost.com"
	firstName := GetRandomName("FirstName")
	lastName := GetRandomName("LastName")
	err := client.InviteUserToOrganization(ctx, token.AccessToken, cfg.GoCloak.Realm, orgID,
		gocloak.OrganizationInviteUserParams{Email: &email, FirstName: &firstName, LastName: &lastName})
	require.NoError(t, err, "InviteUserToOrganization failed")

	invitations, err := client.GetOrganizationInvitations(ctx, token.AccessToken, cfg.GoCloak.Realm, orgID,
		gocloak.GetOrganizationInvitationsParams{Email: &email})
	require.NoError(t, err, "GetOrganizationInvitations failed")
	require.Len(t, invitations, 1)
	invitationID := gocloak.PString(invitations[0].ID)
	require.NotEmpty(t, invitationID)
	t.Cleanup(func() {
		_ = client.DeleteOrganizationInvitation(ctx, token.AccessToken, cfg.GoCloak.Realm, orgID, invitationID)
	})

	invitation, err := client.GetOrganizationInvitationByID(ctx, token.AccessToken, cfg.GoCloak.Realm, orgID, invitationID)
	require.NoError(t, err, "GetOrganizationInvitationByID failed")
	require.Equal(t, invitationID, gocloak.PString(invitation.ID))
	require.Equal(t, orgID, gocloak.PString(invitation.OrganizationID))
	require.Equal(t, email, gocloak.PString(invitation.Email))
	require.Equal(t, firstName, gocloak.PString(invitation.FirstName))
	require.Equal(t, lastName, gocloak.PString(invitation.LastName))
	require.NotNil(t, invitation.Status)
	require.Equal(t, gocloak.InvitationStatusPending, *invitation.Status)

	err = client.ResendOrganizationInvitation(ctx, token.AccessToken, cfg.GoCloak.Realm, orgID, invitationID)
	require.NoError(t, err, "ResendOrganizationInvitation failed")

	// Keycloak 26.8 resend replaces the invitation: the old ID is gone, a new one is pending
	_, err = client.GetOrganizationInvitationByID(ctx, token.AccessToken, cfg.GoCloak.Realm, orgID, invitationID)
	requireAPIErrorCode(t, err, http.StatusNotFound)
	invitations, err = client.GetOrganizationInvitations(ctx, token.AccessToken, cfg.GoCloak.Realm, orgID,
		gocloak.GetOrganizationInvitationsParams{Email: &email})
	require.NoError(t, err, "GetOrganizationInvitations failed")
	require.Len(t, invitations, 1)
	invitationID = gocloak.PString(invitations[0].ID) // the cleanup closure picks up the new ID
	require.NotEmpty(t, invitationID)

	err = client.DeleteOrganizationInvitation(ctx, token.AccessToken, cfg.GoCloak.Realm, orgID, invitationID)
	require.NoError(t, err, "DeleteOrganizationInvitation failed")

	_, err = client.GetOrganizationInvitationByID(ctx, token.AccessToken, cfg.GoCloak.Realm, orgID, invitationID)
	requireAPIErrorCode(t, err, http.StatusNotFound)
}

func Test_OrganizationGroups(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)
	ctx := context.Background()

	tearDown, orgID := CreateOrganization(t, client)
	t.Cleanup(tearDown)
	groupID, name := createOrganizationGroup(t, client, token.AccessToken, orgID)

	groups, err := client.GetOrganizationGroups(ctx, token.AccessToken, cfg.GoCloak.Realm, orgID, gocloak.GetOrganizationGroupsParams{})
	require.NoError(t, err, "GetOrganizationGroups failed")
	require.Equal(t, []string{groupID}, pluck(groups, idOfGroup))

	group, err := client.GetOrganizationGroupByID(ctx, token.AccessToken, cfg.GoCloak.Realm, orgID, groupID)
	require.NoError(t, err, "GetOrganizationGroupByID failed")
	require.Equal(t, name, gocloak.PString(group.Name))

	group, err = client.GetOrganizationGroupByPath(ctx, token.AccessToken, cfg.GoCloak.Realm, orgID, "/"+name, false)
	require.NoError(t, err, "GetOrganizationGroupByPath failed")
	require.Equal(t, groupID, gocloak.PString(group.ID))

	newName := GetRandomName("OrgGroupRenamed")
	err = client.UpdateOrganizationGroup(ctx, token.AccessToken, cfg.GoCloak.Realm, orgID, gocloak.Group{ID: &groupID, Name: &newName})
	require.NoError(t, err, "UpdateOrganizationGroup failed")
	group, err = client.GetOrganizationGroupByID(ctx, token.AccessToken, cfg.GoCloak.Realm, orgID, groupID)
	require.NoError(t, err, "GetOrganizationGroupByID failed")
	require.Equal(t, newName, gocloak.PString(group.Name))

	subName := GetRandomName("OrgSubGroup")
	subGroupID, err := client.AddOrganizationSubgroup(ctx, token.AccessToken, cfg.GoCloak.Realm, orgID, groupID, gocloak.Group{Name: &subName})
	require.NoError(t, err, "AddOrganizationSubgroup failed")
	require.NotEmpty(t, subGroupID)

	subGroups, err := client.GetOrganizationGroupSubgroups(ctx, token.AccessToken, cfg.GoCloak.Realm, orgID, groupID, gocloak.GetOrganizationGroupSubgroupsParams{})
	require.NoError(t, err, "GetOrganizationGroupSubgroups failed")
	require.Equal(t, []string{subGroupID}, pluck(subGroups, idOfGroup))
	require.Equal(t, subName, gocloak.PString(subGroups[0].Name))

	err = client.DeleteOrganizationGroup(ctx, token.AccessToken, cfg.GoCloak.Realm, orgID, groupID)
	require.NoError(t, err, "DeleteOrganizationGroup failed")

	_, err = client.GetOrganizationGroupByID(ctx, token.AccessToken, cfg.GoCloak.Realm, orgID, groupID)
	requireAPIErrorCode(t, err, http.StatusNotFound)
}

func Test_OrganizationGroupMembers(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)
	ctx := context.Background()

	tearDown, orgID := CreateOrganization(t, client)
	t.Cleanup(tearDown)
	userID := addOrganizationMember(t, client, token.AccessToken, orgID)
	groupID, _ := createOrganizationGroup(t, client, token.AccessToken, orgID)

	err := client.AddMemberToOrganizationGroup(ctx, token.AccessToken, cfg.GoCloak.Realm, orgID, groupID, userID)
	require.NoError(t, err, "AddMemberToOrganizationGroup failed")

	members, err := client.GetOrganizationGroupMembers(ctx, token.AccessToken, cfg.GoCloak.Realm, orgID, groupID, gocloak.GetOrganizationMembersParams{})
	require.NoError(t, err, "GetOrganizationGroupMembers failed")
	require.Equal(t, []string{userID}, pluck(members, idOfMember))

	err = client.RemoveMemberFromOrganizationGroup(ctx, token.AccessToken, cfg.GoCloak.Realm, orgID, groupID, userID)
	require.NoError(t, err, "RemoveMemberFromOrganizationGroup failed")

	members, err = client.GetOrganizationGroupMembers(ctx, token.AccessToken, cfg.GoCloak.Realm, orgID, groupID, gocloak.GetOrganizationMembersParams{})
	require.NoError(t, err, "GetOrganizationGroupMembers failed")
	require.Empty(t, members)
}

func Test_OrganizationValidationErrors(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	ctx := context.Background()

	// client-side checks fail before any request, so no token is needed
	_, err := client.GetOrganizationMemberOrganizations(ctx, "", cfg.GoCloak.Realm, "", "user")
	require.ErrorContains(t, err, "organizationID shall not be empty")

	_, err = client.GetOrganizationMemberOrganizations(ctx, "", cfg.GoCloak.Realm, "org", "")
	require.ErrorContains(t, err, "userID shall not be empty")

	err = client.InviteUserToOrganization(ctx, "", cfg.GoCloak.Realm, "org", gocloak.OrganizationInviteUserParams{})
	require.ErrorContains(t, err, "email required")

	err = client.UpdateOrganizationGroup(ctx, "", cfg.GoCloak.Realm, "org", gocloak.Group{Name: new("name")})
	require.ErrorContains(t, err, "ID of a group required")
}

func Test_OrganizationServerErrors(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)
	ctx := context.Background()

	tearDown, orgID := CreateOrganization(t, client)
	t.Cleanup(tearDown)
	missingID := "00000000-0000-0000-0000-000000000000"

	_, err := client.GetOrganizationMemberOrganizations(ctx, token.AccessToken, cfg.GoCloak.Realm, orgID, missingID)
	requireAPIErrorCode(t, err, http.StatusNotFound)

	_, err = client.GetOrganizationIdentityProvider(ctx, token.AccessToken, cfg.GoCloak.Realm, orgID, GetRandomName("missing-idp"))
	requireAPIErrorCode(t, err, http.StatusNotFound)

	_, err = client.GetOrganizationInvitationByID(ctx, token.AccessToken, cfg.GoCloak.Realm, orgID, missingID)
	requireAPIErrorCode(t, err, http.StatusNotFound)

	groupID, _ := createOrganizationGroup(t, client, token.AccessToken, orgID)
	_, err = client.GetOrganizationGroupByID(ctx, token.AccessToken, cfg.GoCloak.Realm, orgID, missingID)
	requireAPIErrorCode(t, err, http.StatusNotFound)

	// only organization members may join organization groups
	tearDownUser, outsiderID := CreateUser(t, client)
	t.Cleanup(tearDownUser)
	err = client.AddMemberToOrganizationGroup(ctx, token.AccessToken, cfg.GoCloak.Realm, orgID, groupID, outsiderID)
	requireAPIErrorCode(t, err, http.StatusBadRequest)
}

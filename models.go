package gocloak

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// GetQueryParams converts the struct to map[string]string
// The fields tags must have `json:"<name>,string,omitempty"` format for all types, except strings
// The string fields must have: `json:"<name>,omitempty"`. The `json:"<name>,string,omitempty"` tag for string field
// will add additional double quotes.
// "string" tag allows to convert the non-string fields of a structure to map[string]string.
// "omitempty" allows to skip the fields with default values.
func GetQueryParams(s any) (map[string]string, error) {
	b, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	var res map[string]string
	err = json.Unmarshal(b, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// StringOrArray represents a value that can either be a string or an array of strings
type StringOrArray []string

// UnmarshalJSON decodes either a JSON array of strings or a single JSON string into the StringOrArray.
// Input: data [[]byte] JSON array or string.
// Output: [error] if data is neither.
func (s *StringOrArray) UnmarshalJSON(data []byte) error {
	if len(data) > 1 && data[0] == '[' {
		var obj []string
		if err := json.Unmarshal(data, &obj); err != nil {
			return err
		}
		*s = StringOrArray(obj)
		return nil
	}

	var obj string
	if err := json.Unmarshal(data, &obj); err != nil {
		return err
	}
	*s = StringOrArray([]string{obj})
	return nil
}

// MarshalJSON encodes a single-item StringOrArray as a JSON string, otherwise as a JSON array.
// Value receiver, so non-addressable values (e.g. json.Marshal(StringOrArray{"x"})) use it too.
// Output: [[]byte] JSON, [error] if encoding fails.
func (s StringOrArray) MarshalJSON() ([]byte, error) {
	if len(s) == 1 {
		return json.Marshal(s[0])
	}
	return json.Marshal([]string(s))
}

// EnforcedString can be used when the expected value is string but Keycloak in some cases gives you mixed types
type EnforcedString string

// UnmarshalJSON keeps a JSON string as its value and any other JSON value (number, bool, object, array)
// as its raw JSON text; null leaves the value unchanged.
// Input: data [[]byte] JSON value. Output: [error] if a JSON string is malformed.
func (s *EnforcedString) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
		return nil
	}
	if data[0] != '"' {
		*s = EnforcedString(data)
		return nil
	}

	var val string
	err := json.Unmarshal(data, &val)
	*s = EnforcedString(val)
	return err
}

// MarshalJSON encodes the EnforcedString as a plain JSON string.
// Output: [[]byte] JSON, [error] if encoding fails.
func (s *EnforcedString) MarshalJSON() ([]byte, error) {
	return json.Marshal(*s)
}

// APIErrType is a field containing more specific API error types
// that may be checked by the receiver.
type APIErrType string

const (
	// APIErrTypeUnknown is for API errors that are not strongly
	// typed.
	APIErrTypeUnknown APIErrType = "unknown"

	// APIErrTypeInvalidGrant corresponds with Keycloak's
	// OAuthErrorException due to "invalid_grant".
	APIErrTypeInvalidGrant APIErrType = "oauth: invalid grant"
)

// ParseAPIErrType is a convenience method for returning strongly typed API errors.
// Input: err [error].
// Output: [APIErrType] APIErrTypeInvalidGrant if the message contains "invalid_grant", otherwise APIErrTypeUnknown.
func ParseAPIErrType(err error) APIErrType {
	if err == nil {
		return APIErrTypeUnknown
	}
	switch {
	case strings.Contains(err.Error(), "invalid_grant"):
		return APIErrTypeInvalidGrant
	default:
		return APIErrTypeUnknown
	}
}

// APIError holds message and statusCode for api errors
type APIError struct {
	Code    int        `json:"code"`
	Message string     `json:"message"`
	Type    APIErrType `json:"type"`
	// Err is the underlying transport error (e.g. context.Canceled), nil for HTTP errors
	Err error `json:"-"`
}

// Unwrap returns the underlying transport error so errors.Is/As work on it.
// Output: [error] wrapped error or nil.
func (apiError APIError) Unwrap() error {
	return apiError.Err
}

// Error returns the message of the APIError.
// Output: [string] error message.
func (apiError APIError) Error() string {
	return apiError.Message
}

// CertResponseKey is returned by the certs endpoint.
// JSON Web Key structure is described here:
// https://self-issued.info/docs/draft-ietf-jose-json-web-key.html#JWKContents
type CertResponseKey struct {
	Kid     string   `json:"kid,omitzero"`
	Kty     string   `json:"kty,omitzero"`
	Alg     string   `json:"alg,omitzero"`
	Use     string   `json:"use,omitzero"`
	N       string   `json:"n,omitzero"`
	E       string   `json:"e,omitzero"`
	X       string   `json:"x,omitzero"`
	Y       string   `json:"y,omitzero"`
	Crv     string   `json:"crv,omitzero"`
	KeyOps  []string `json:"key_ops,omitzero"`
	X5u     string   `json:"x5u,omitzero"`
	X5c     []string `json:"x5c,omitzero"`
	X5t     string   `json:"x5t,omitzero"`
	X5tS256 string   `json:"x5t#S256,omitzero"`
}

// CertResponse is returned by the certs endpoint
type CertResponse struct {
	Keys []CertResponseKey `json:"keys,omitzero"`
}

// IssuerResponse is returned by the issuer endpoint
type IssuerResponse struct {
	Realm           string `json:"realm,omitzero"`
	PublicKey       string `json:"public_key,omitzero"`
	TokenService    string `json:"token-service,omitzero"`
	AccountService  string `json:"account-service,omitzero"`
	TokensNotBefore int    `json:"tokens-not-before,omitzero"`
}

// ResourcePermission represents a permission granted to a resource
type ResourcePermission struct {
	RSID           string   `json:"rsid,omitzero"`
	ResourceID     string   `json:"resource_id,omitzero"`
	RSName         string   `json:"rsname,omitzero"`
	Scopes         []string `json:"scopes,omitzero"`
	ResourceScopes []string `json:"resource_scopes,omitzero"`
}

// PermissionResource represents a resources asscoiated with a permission
type PermissionResource struct {
	ResourceID   string `json:"_id,omitzero"`
	ResourceName string `json:"name,omitzero"`
}

// PermissionScope represents scopes associated with a permission
type PermissionScope struct {
	ScopeID   string `json:"id,omitzero"`
	ScopeName string `json:"name,omitzero"`
}

// IntroSpectTokenResult is returned when a token was checked
type IntroSpectTokenResult struct {
	Permissions []ResourcePermission `json:"permissions,omitzero"`
	Exp         int                  `json:"exp,omitzero"`
	Nbf         int                  `json:"nbf,omitzero"`
	Iat         int                  `json:"iat,omitzero"`
	Aud         StringOrArray        `json:"aud,omitzero"`
	Active      bool                 `json:"active,omitzero"`
	AuthTime    int                  `json:"auth_time,omitzero"`
	Jti         string               `json:"jti,omitzero"`
	Type        string               `json:"typ,omitzero"`
	Azp         string               `json:"azp,omitzero"`
}

// User represents the Keycloak User Structure
type User struct {
	ID                         *string                    `json:"id,omitempty"`
	CreatedTimestamp           *int64                     `json:"createdTimestamp,omitempty"`
	Username                   *string                    `json:"username,omitempty"`
	Enabled                    *bool                      `json:"enabled,omitempty"`
	Totp                       *bool                      `json:"totp,omitempty"`
	EmailVerified              *bool                      `json:"emailVerified,omitempty"`
	FirstName                  *string                    `json:"firstName,omitempty"`
	LastName                   *string                    `json:"lastName,omitempty"`
	Email                      *string                    `json:"email,omitempty"`
	FederationLink             *string                    `json:"federationLink,omitempty"`
	Attributes                 map[string][]string        `json:"attributes,omitzero"`
	DisableableCredentialTypes []any                      `json:"disableableCredentialTypes,omitzero"`
	RequiredActions            []string                   `json:"requiredActions,omitzero"`
	Access                     map[string]bool            `json:"access,omitzero"`
	ClientRoles                map[string][]string        `json:"clientRoles,omitzero"`
	RealmRoles                 []string                   `json:"realmRoles,omitzero"`
	Groups                     []string                   `json:"groups,omitzero"`
	ServiceAccountClientID     *string                    `json:"serviceAccountClientId,omitempty"`
	Credentials                []CredentialRepresentation `json:"credentials,omitzero"`
}

// SetPasswordRequest sets a new password
type SetPasswordRequest struct {
	Type      string `json:"type,omitzero"`
	Temporary bool   `json:"temporary,omitzero"`
	Password  string `json:"value,omitzero"`
}

// Component is a component
type Component struct {
	ID              *string             `json:"id,omitempty"`
	Name            *string             `json:"name,omitempty"`
	ProviderID      *string             `json:"providerId,omitempty"`
	ProviderType    *string             `json:"providerType,omitempty"`
	ParentID        *string             `json:"parentId,omitempty"`
	ComponentConfig map[string][]string `json:"config,omitzero"`
	SubType         *string             `json:"subType,omitempty"`
}

// KeyStoreConfig holds the keyStoreConfig
type KeyStoreConfig struct {
	ActiveKeys ActiveKeys `json:"active,omitzero"`
	Key        []Key      `json:"keys,omitzero"`
}

// ActiveKeys holds the active keys
type ActiveKeys struct {
	HS256 string `json:"HS256,omitzero"`
	RS256 string `json:"RS256,omitzero"`
	AES   string `json:"AES,omitzero"`
}

// Key is a key
type Key struct {
	ProviderID       string `json:"providerId,omitzero"`
	ProviderPriority int    `json:"providerPriority,omitzero"`
	Kid              string `json:"kid,omitzero"`
	Status           string `json:"status,omitzero"`
	Type             string `json:"type,omitzero"`
	Algorithm        string `json:"algorithm,omitzero"`
	PublicKey        string `json:"publicKey,omitzero"`
	Certificate      string `json:"certificate,omitzero"`
}

// GetUsersParams represents the optional parameters for getting users
type GetUsersParams struct {
	BriefRepresentation *bool   `json:"briefRepresentation,string,omitempty"`
	Email               *string `json:"email,omitempty"`
	EmailVerified       *bool   `json:"emailVerified,string,omitempty"`
	Enabled             *bool   `json:"enabled,string,omitempty"`
	Exact               *bool   `json:"exact,string,omitempty"`
	First               *int    `json:"first,string,omitempty"`
	FirstName           *string `json:"firstName,omitempty"`
	IDPAlias            *string `json:"idpAlias,omitempty"`
	IDPUserID           *string `json:"idpUserId,omitempty"`
	LastName            *string `json:"lastName,omitempty"`
	Max                 *int    `json:"max,string,omitempty"`
	Q                   *string `json:"q,omitempty"`
	Search              *string `json:"search,omitempty"`
	Username            *string `json:"username,omitempty"`
}

// GetComponentsParams represents the optional parameters for getting components
type GetComponentsParams struct {
	Name         *string `json:"name,omitempty"`
	ProviderType *string `json:"type,omitempty"`
	ParentID     *string `json:"parent,omitempty"`
}

// ExecuteActionsEmail represents parameters for executing action emails
type ExecuteActionsEmail struct {
	UserID      *string  `json:"-"`
	ClientID    *string  `json:"client_id,omitempty"`
	Lifespan    *int     `json:"lifespan,string,omitempty"`
	RedirectURI *string  `json:"redirect_uri,omitempty"`
	Actions     []string `json:"-"`
}

// SendVerificationMailParams is being used to send verification params
type SendVerificationMailParams struct {
	ClientID    *string
	RedirectURI *string
}

// Group is a Group
type Group struct {
	ID          *string             `json:"id,omitempty"`
	Name        *string             `json:"name,omitempty"`
	Path        *string             `json:"path,omitempty"`
	SubGroups   []Group             `json:"subGroups,omitzero"`
	Attributes  map[string][]string `json:"attributes,omitzero"`
	Access      map[string]bool     `json:"access,omitzero"`
	ClientRoles map[string][]string `json:"clientRoles,omitzero"`
	RealmRoles  []string            `json:"realmRoles,omitzero"`
}

// GroupsCount represents the groups count response from keycloak
type GroupsCount struct {
	Count int `json:"count,omitempty"`
}

// GetGroupsParams represents the optional parameters for getting groups
type GetGroupsParams struct {
	BriefRepresentation *bool   `json:"briefRepresentation,string,omitempty"`
	Exact               *bool   `json:"exact,string,omitempty"`
	First               *int    `json:"first,string,omitempty"`
	Full                *bool   `json:"full,string,omitempty"`
	Max                 *int    `json:"max,string,omitempty"`
	Q                   *string `json:"q,omitempty"`
	Search              *string `json:"search,omitempty"`
}

// GetChildGroupsParams represents the optional parameters for getting child groups
type GetChildGroupsParams struct {
	BriefRepresentation *bool   `json:"briefRepresentation,string,omitempty"`
	Exact               *bool   `json:"exact,string,omitempty"`
	First               *int    `json:"first,string,omitempty"`
	Max                 *int    `json:"max,string,omitempty"`
	Search              *string `json:"search,omitempty"`
}

// MarshalJSON is a custom json marshaling function that keeps Full and BriefRepresentation consistent (one is the
// negation of the other) for backward compatibility.
// Output: [[]byte] JSON, [error] if encoding fails.
func (obj GetGroupsParams) MarshalJSON() ([]byte, error) {
	type Alias GetGroupsParams
	a := Alias(obj)
	if a.BriefRepresentation != nil {
		a.Full = new(!*a.BriefRepresentation)
	} else if a.Full != nil {
		a.BriefRepresentation = new(!*a.Full)
	}
	return json.Marshal(a)
}

// CompositesRepresentation represents the composite roles of a role
type CompositesRepresentation struct {
	Client map[string][]string `json:"client,omitzero"`
	Realm  []string            `json:"realm,omitzero"`
}

// Role is a role
type Role struct {
	ID                 *string                  `json:"id,omitempty"`
	Name               *string                  `json:"name,omitempty"`
	ScopeParamRequired *bool                    `json:"scopeParamRequired,omitempty"`
	Composite          *bool                    `json:"composite,omitempty"`
	Composites         CompositesRepresentation `json:"composites,omitzero"`
	ClientRole         *bool                    `json:"clientRole,omitempty"`
	ContainerID        *string                  `json:"containerId,omitempty"`
	Description        *string                  `json:"description,omitempty"`
	Attributes         map[string][]string      `json:"attributes,omitzero"`
}

// GetRoleParams represents the optional parameters for getting roles
type GetRoleParams struct {
	First               *int    `json:"first,string,omitempty"`
	Max                 *int    `json:"max,string,omitempty"`
	Search              *string `json:"search,omitempty"`
	BriefRepresentation *bool   `json:"briefRepresentation,string,omitempty"`
}

// ClientMappingsRepresentation is a client role mappings
type ClientMappingsRepresentation struct {
	ID       string `json:"id,omitzero"`
	Client   string `json:"client,omitzero"`
	Mappings []Role `json:"mappings,omitzero"`
}

// MappingsRepresentation is a representation of role mappings
type MappingsRepresentation struct {
	ClientMappings map[string]ClientMappingsRepresentation `json:"clientMappings,omitzero"`
	RealmMappings  []Role                                  `json:"realmMappings,omitzero"`
}

// ClientScope is a ClientScope
type ClientScope struct {
	ID                    *string               `json:"id,omitempty"`
	Name                  *string               `json:"name,omitempty"`
	Type                  *string               `json:"type,omitempty"`
	Description           *string               `json:"description,omitempty"`
	Protocol              *string               `json:"protocol,omitempty"`
	ClientScopeAttributes ClientScopeAttributes `json:"attributes,omitzero"`
	ProtocolMappers       []ProtocolMappers     `json:"protocolMappers,omitzero"`
}

// ClientScopeAttributes are attributes of client scopes
type ClientScopeAttributes struct {
	ConsentScreenText      *string `json:"consent.screen.text,omitempty"`
	DisplayOnConsentScreen *string `json:"display.on.consent.screen,omitempty"`
	IncludeInTokenScope    *string `json:"include.in.token.scope,omitempty"`
}

// ProtocolMappers are protocolmappers
type ProtocolMappers struct {
	ID                    *string               `json:"id,omitempty"`
	Name                  *string               `json:"name,omitempty"`
	Protocol              *string               `json:"protocol,omitempty"`
	ProtocolMapper        *string               `json:"protocolMapper,omitempty"`
	ConsentRequired       *bool                 `json:"consentRequired,omitempty"`
	ProtocolMappersConfig ProtocolMappersConfig `json:"config,omitzero"`
}

// ProtocolMappersConfig is a config of a protocol mapper
type ProtocolMappersConfig struct {
	IntrospectionTokenClaim            *string `json:"introspection.token.claim,omitempty"`
	UserinfoTokenClaim                 *string `json:"userinfo.token.claim,omitempty"`
	UserAttribute                      *string `json:"user.attribute,omitempty"`
	IDTokenClaim                       *string `json:"id.token.claim,omitempty"`
	AccessTokenClaim                   *string `json:"access.token.claim,omitempty"`
	ClaimName                          *string `json:"claim.name,omitempty"`
	ClaimValue                         *string `json:"claim.value,omitempty"`
	JSONTypeLabel                      *string `json:"jsonType.label,omitempty"`
	Multivalued                        *string `json:"multivalued,omitempty"`
	AggregateAttrs                     *string `json:"aggregate.attrs,omitempty"`
	UsermodelClientRoleMappingClientID *string `json:"usermodel.clientRoleMapping.clientId,omitempty"`
	IncludedClientAudience             *string `json:"included.client.audience,omitempty"`
	IncludedCustomAudience             *string `json:"included.custom.audience,omitempty"`
	FullPath                           *string `json:"full.path,omitempty"`
	AttributeName                      *string `json:"attribute.name,omitempty"`
	AttributeNameFormat                *string `json:"attribute.nameformat,omitempty"`
	Single                             *string `json:"single,omitempty"`
	Script                             *string `json:"script,omitempty"`
	UserSessionNote                    *string `json:"user.session.note,omitempty"`
	AddOrganizationAttributes          *string `json:"addOrganizationAttributes,omitempty"`
	AddOrganizationID                  *string `json:"addOrganizationId,omitempty"`
}

// Client is a ClientRepresentation
type Client struct {
	Access                             map[string]any                 `json:"access,omitzero"`
	AdminURL                           *string                        `json:"adminUrl,omitempty"`
	Attributes                         map[string]string              `json:"attributes,omitzero"`
	AuthenticationFlowBindingOverrides map[string]string              `json:"authenticationFlowBindingOverrides,omitzero"`
	AuthorizationServicesEnabled       *bool                          `json:"authorizationServicesEnabled,omitempty"`
	AuthorizationSettings              ResourceServerRepresentation   `json:"authorizationSettings,omitzero"`
	BaseURL                            *string                        `json:"baseUrl,omitempty"`
	BearerOnly                         *bool                          `json:"bearerOnly,omitempty"`
	ClientAuthenticatorType            *string                        `json:"clientAuthenticatorType,omitempty"`
	ClientID                           *string                        `json:"clientId,omitempty"`
	ClientTemplate                     *string                        `json:"clientTemplate,omitempty"`
	ConsentRequired                    *bool                          `json:"consentRequired,omitempty"`
	DefaultClientScopes                []string                       `json:"defaultClientScopes,omitzero"`
	DefaultRoles                       []string                       `json:"defaultRoles,omitzero"`
	Description                        *string                        `json:"description,omitempty"`
	DirectAccessGrantsEnabled          *bool                          `json:"directAccessGrantsEnabled,omitempty"`
	Enabled                            *bool                          `json:"enabled,omitempty"`
	FrontChannelLogout                 *bool                          `json:"frontchannelLogout,omitempty"`
	FullScopeAllowed                   *bool                          `json:"fullScopeAllowed,omitempty"`
	ID                                 *string                        `json:"id,omitempty"`
	ImplicitFlowEnabled                *bool                          `json:"implicitFlowEnabled,omitempty"`
	Name                               *string                        `json:"name,omitempty"`
	NodeReRegistrationTimeout          *int32                         `json:"nodeReRegistrationTimeout,omitempty"`
	NotBefore                          *int32                         `json:"notBefore,omitempty"`
	OptionalClientScopes               []string                       `json:"optionalClientScopes,omitzero"`
	Origin                             *string                        `json:"origin,omitempty"`
	Protocol                           *string                        `json:"protocol,omitempty"`
	ProtocolMappers                    []ProtocolMapperRepresentation `json:"protocolMappers,omitzero"`
	PublicClient                       *bool                          `json:"publicClient,omitempty"`
	RedirectURIs                       []string                       `json:"redirectUris,omitzero"`
	RegisteredNodes                    map[string]int                 `json:"registeredNodes,omitzero"`
	RegistrationAccessToken            *string                        `json:"registrationAccessToken,omitempty"`
	RootURL                            *string                        `json:"rootUrl,omitempty"`
	Secret                             *string                        `json:"secret,omitempty"`
	ServiceAccountsEnabled             *bool                          `json:"serviceAccountsEnabled,omitempty"`
	ServiceAccountUserID               *string                        `json:"serviceAccountUserId,omitempty"`
	StandardFlowEnabled                *bool                          `json:"standardFlowEnabled,omitempty"`
	SurrogateAuthRequired              *bool                          `json:"surrogateAuthRequired,omitempty"`
	UseTemplateConfig                  *bool                          `json:"useTemplateConfig,omitempty"`
	UseTemplateMappers                 *bool                          `json:"useTemplateMappers,omitempty"`
	UseTemplateScope                   *bool                          `json:"useTemplateScope,omitempty"`
	WebOrigins                         []string                       `json:"webOrigins,omitzero"`
}

// ResourceServerRepresentation represents the resources of a Server
type ResourceServerRepresentation struct {
	AllowRemoteResourceManagement *bool                    `json:"allowRemoteResourceManagement,omitempty"`
	ClientID                      *string                  `json:"clientId,omitempty"`
	ID                            *string                  `json:"id,omitempty"`
	Name                          *string                  `json:"name,omitempty"`
	Policies                      []PolicyRepresentation   `json:"policies,omitzero"`
	PolicyEnforcementMode         PolicyEnforcementMode    `json:"policyEnforcementMode,omitzero"`
	Resources                     []ResourceRepresentation `json:"resources,omitzero"`
	Scopes                        []ScopeRepresentation    `json:"scopes,omitzero"`
	DecisionStrategy              DecisionStrategy         `json:"decisionStrategy,omitzero"`
}

// RoleDefinition represents a role in a RolePolicyRepresentation
type RoleDefinition struct {
	ID       *string `json:"id,omitempty"`
	Private  *bool   `json:"private,omitempty"`
	Required *bool   `json:"required,omitempty"`
}

// AdapterConfiguration represents adapter configuration of a client
type AdapterConfiguration struct {
	Realm            string `json:"realm"`
	AuthServerURL    string `json:"auth-server-url"`
	SSLRequired      string `json:"ssl-required"`
	Resource         string `json:"resource"`
	Credentials      any    `json:"credentials"`
	ConfidentialPort int    `json:"confidential-port"`
}

// PolicyEnforcementMode is an enum type for PolicyEnforcementMode of ResourceServerRepresentation
type PolicyEnforcementMode string

// PolicyEnforcementMode values
const (
	Enforcing  PolicyEnforcementMode = "ENFORCING"
	Permissive PolicyEnforcementMode = "PERMISSIVE"
	Disabled   PolicyEnforcementMode = "DISABLED"
)

// Logic is an enum type for policy logic
type Logic string

// Logic values
const (
	Positive Logic = "POSITIVE"
	Negative Logic = "NEGATIVE"
)

// DecisionStrategy is an enum type for DecisionStrategy of PolicyRepresentation
type DecisionStrategy string

// DecisionStrategy values
const (
	Affirmative DecisionStrategy = "AFFIRMATIVE"
	Unanimous   DecisionStrategy = "UNANIMOUS"
	Consensus   DecisionStrategy = "CONSENSUS"
)

// PolicyRepresentation is a representation of a Policy
type PolicyRepresentation struct {
	Config           map[string]string `json:"config,omitzero"`
	DecisionStrategy DecisionStrategy  `json:"decisionStrategy,omitzero"`
	Description      *string           `json:"description,omitempty"`
	ID               *string           `json:"id,omitempty"`
	Logic            Logic             `json:"logic,omitzero"`
	Name             *string           `json:"name,omitempty"`
	Owner            *string           `json:"owner,omitempty"`
	Policies         []string          `json:"policies,omitzero"`
	Resources        []string          `json:"resources,omitzero"`
	Scopes           []string          `json:"scopes,omitzero"`
	Type             *string           `json:"type,omitempty"`
	RolePolicyRepresentation
	JSPolicyRepresentation
	ClientPolicyRepresentation
	TimePolicyRepresentation
	UserPolicyRepresentation
	AggregatedPolicyRepresentation
	GroupPolicyRepresentation
}

// RolePolicyRepresentation represents role based policies
type RolePolicyRepresentation struct {
	FetchRoles *bool            `json:"fetchRoles,omitempty"`
	Roles      []RoleDefinition `json:"roles,omitzero"`
}

// JSPolicyRepresentation represents js based policies
type JSPolicyRepresentation struct {
	Code *string `json:"code,omitempty"`
}

// ClientPolicyRepresentation represents client based policies
type ClientPolicyRepresentation struct {
	Clients []string `json:"clients,omitzero"`
}

// TimePolicyRepresentation represents time based policies
type TimePolicyRepresentation struct {
	NotBefore    *string `json:"notBefore,omitempty"`
	NotOnOrAfter *string `json:"notOnOrAfter,omitempty"`
	DayMonth     *string `json:"dayMonth,omitempty"`
	DayMonthEnd  *string `json:"dayMonthEnd,omitempty"`
	Month        *string `json:"month,omitempty"`
	MonthEnd     *string `json:"monthEnd,omitempty"`
	Year         *string `json:"year,omitempty"`
	YearEnd      *string `json:"yearEnd,omitempty"`
	Hour         *string `json:"hour,omitempty"`
	HourEnd      *string `json:"hourEnd,omitempty"`
	Minute       *string `json:"minute,omitempty"`
	MinuteEnd    *string `json:"minuteEnd,omitempty"`
}

// UserPolicyRepresentation represents user based policies
type UserPolicyRepresentation struct {
	Users []string `json:"users,omitzero"`
}

// AggregatedPolicyRepresentation represents aggregated policies
type AggregatedPolicyRepresentation struct {
	Policies []string `json:"policies,omitzero"`
}

// GroupPolicyRepresentation represents group based policies
type GroupPolicyRepresentation struct {
	Groups      []GroupDefinition `json:"groups,omitzero"`
	GroupsClaim *string           `json:"groupsClaim,omitempty"`
}

// GroupDefinition represents a group in a GroupPolicyRepresentation
type GroupDefinition struct {
	ID             *string `json:"id,omitempty"`
	Path           *string `json:"path,omitempty"`
	ExtendChildren *bool   `json:"extendChildren,omitempty"`
}

// ResourceRepresentation is a representation of a Resource
type ResourceRepresentation struct {
	ID                 *string                     `json:"_id,omitempty"` // TODO: is marked "_optional" in template, input error or deliberate?
	Attributes         map[string][]string         `json:"attributes,omitzero"`
	DisplayName        *string                     `json:"displayName,omitempty"`
	IconURI            *string                     `json:"icon_uri,omitempty"` // TODO: With "_" because that's how it's written down in the template
	Name               *string                     `json:"name,omitempty"`
	Owner              ResourceOwnerRepresentation `json:"owner,omitzero"`
	OwnerManagedAccess *bool                       `json:"ownerManagedAccess,omitempty"`
	ResourceScopes     []ScopeRepresentation       `json:"resource_scopes,omitzero"`
	Scopes             []ScopeRepresentation       `json:"scopes,omitzero"`
	Type               *string                     `json:"type,omitempty"`
	URIs               []string                    `json:"uris,omitzero"`
}

// ResourceOwnerRepresentation represents a resource's owner
type ResourceOwnerRepresentation struct {
	ID   *string `json:"id,omitempty"`
	Name *string `json:"name,omitempty"`
}

// ScopeRepresentation is a represents a Scope
type ScopeRepresentation struct {
	DisplayName *string                  `json:"displayName,omitempty"`
	IconURI     *string                  `json:"iconUri,omitempty"`
	ID          *string                  `json:"id,omitempty"`
	Name        *string                  `json:"name,omitempty"`
	Policies    []PolicyRepresentation   `json:"policies,omitzero"`
	Resources   []ResourceRepresentation `json:"resources,omitzero"`
}

// ProtocolMapperRepresentation represents....
type ProtocolMapperRepresentation struct {
	Config          map[string]string `json:"config,omitzero"`
	ID              *string           `json:"id,omitempty"`
	Name            *string           `json:"name,omitempty"`
	Protocol        *string           `json:"protocol,omitempty"`
	ProtocolMapper  *string           `json:"protocolMapper,omitempty"`
	ConsentRequired *bool             `json:"consentRequired,omitempty"`
}

// GetClientsParams represents the query parameters
type GetClientsParams struct {
	ClientID             *string `json:"clientId,omitempty"`
	ViewableOnly         *bool   `json:"viewableOnly,string,omitempty"`
	First                *int    `json:"first,string,omitempty"`
	Max                  *int    `json:"max,string,omitempty"`
	Search               *bool   `json:"search,string,omitempty"`
	SearchableAttributes *string `json:"q,omitempty"`
}

// UserInfoAddress is representation of the address sub-filed of UserInfo
// https://openid.net/specs/openid-connect-core-1_0.html#AddressClaim
type UserInfoAddress struct {
	Formatted     string `json:"formatted,omitzero"`
	StreetAddress string `json:"street_address,omitzero"`
	Locality      string `json:"locality,omitzero"`
	Region        string `json:"region,omitzero"`
	PostalCode    string `json:"postal_code,omitzero"`
	Country       string `json:"country,omitzero"`
}

// UserInfo is returned by the userinfo endpoint
// https://openid.net/specs/openid-connect-core-1_0.html#StandardClaims
type UserInfo struct {
	Sub                 string          `json:"sub,omitzero"`
	Name                string          `json:"name,omitzero"`
	GivenName           string          `json:"given_name,omitzero"`
	FamilyName          string          `json:"family_name,omitzero"`
	MiddleName          string          `json:"middle_name,omitzero"`
	Nickname            string          `json:"nickname,omitzero"`
	PreferredUsername   string          `json:"preferred_username,omitzero"`
	Profile             string          `json:"profile,omitzero"`
	Picture             string          `json:"picture,omitzero"`
	Website             string          `json:"website,omitzero"`
	Email               string          `json:"email,omitzero"`
	EmailVerified       bool            `json:"email_verified,omitzero"`
	Gender              string          `json:"gender,omitzero"`
	ZoneInfo            string          `json:"zoneinfo,omitzero"`
	Locale              string          `json:"locale,omitzero"`
	PhoneNumber         string          `json:"phone_number,omitzero"`
	PhoneNumberVerified bool            `json:"phone_number_verified,omitzero"`
	Address             UserInfoAddress `json:"address,omitzero"`
	UpdatedAt           int             `json:"updated_at,omitzero"`
}

// RolesRepresentation represents the roles of a realm
type RolesRepresentation struct {
	Client map[string][]Role `json:"client,omitzero"`
	Realm  []Role            `json:"realm,omitzero"`
}

// RealmRepresentation represents a realm
type RealmRepresentation struct {
	AccessCodeLifespan                                        *int                             `json:"accessCodeLifespan,omitempty"`
	AccessCodeLifespanLogin                                   *int                             `json:"accessCodeLifespanLogin,omitempty"`
	AccessCodeLifespanUserAction                              *int                             `json:"accessCodeLifespanUserAction,omitempty"`
	AccessTokenLifespan                                       *int                             `json:"accessTokenLifespan,omitempty"`
	AccessTokenLifespanForImplicitFlow                        *int                             `json:"accessTokenLifespanForImplicitFlow,omitempty"`
	AccountTheme                                              *string                          `json:"accountTheme,omitempty"`
	ActionTokenGeneratedByAdminLifespan                       *int                             `json:"actionTokenGeneratedByAdminLifespan,omitempty"`
	ActionTokenGeneratedByUserLifespan                        *int                             `json:"actionTokenGeneratedByUserLifespan,omitempty"`
	AdminEventsDetailsEnabled                                 *bool                            `json:"adminEventsDetailsEnabled,omitempty"`
	AdminEventsEnabled                                        *bool                            `json:"adminEventsEnabled,omitempty"`
	AdminTheme                                                *string                          `json:"adminTheme,omitempty"`
	Attributes                                                map[string]string                `json:"attributes,omitzero"`
	AuthenticationFlows                                       []any                            `json:"authenticationFlows,omitzero"`
	AuthenticatorConfig                                       []any                            `json:"authenticatorConfig,omitzero"`
	BruteForceProtected                                       *bool                            `json:"bruteForceProtected,omitempty"`
	BruteForceStrategy                                        *string                          `json:"bruteForceStrategy,omitempty"`
	BrowserFlow                                               *string                          `json:"browserFlow,omitempty"`
	BrowserSecurityHeaders                                    map[string]string                `json:"browserSecurityHeaders,omitzero"`
	ClientOfflineSessionIdleTimeout                           *int                             `json:"clientOfflineSessionIdleTimeout,omitempty"`
	ClientOfflineSessionMaxLifespan                           *int                             `json:"clientOfflineSessionMaxLifespan,omitempty"`
	ClientAuthenticationFlow                                  *string                          `json:"clientAuthenticationFlow,omitempty"`
	ClientPolicies                                            map[string][]any                 `json:"clientPolicies,omitzero"`
	ClientProfiles                                            map[string][]any                 `json:"clientProfiles,omitzero"`
	ClientScopeMappings                                       map[string][]any                 `json:"clientScopeMappings,omitzero"`
	ClientScopes                                              []ClientScope                    `json:"clientScopes,omitzero"`
	ClientSessionIdleTimeout                                  *int                             `json:"clientSessionIdleTimeout,omitempty"`
	ClientSessionMaxLifespan                                  *int                             `json:"clientSessionMaxLifespan,omitempty"`
	Clients                                                   []Client                         `json:"clients,omitzero"`
	Components                                                map[string][]Component           `json:"components,omitzero"`
	DefaultDefaultClientScopes                                []string                         `json:"defaultDefaultClientScopes,omitzero"`
	DefaultGroups                                             []string                         `json:"defaultGroups,omitzero"`
	DefaultLocale                                             *string                          `json:"defaultLocale,omitempty"`
	DefaultOptionalClientScopes                               []string                         `json:"defaultOptionalClientScopes,omitzero"`
	DefaultRole                                               Role                             `json:"defaultRole,omitzero"`
	DefaultRoles                                              []string                         `json:"defaultRoles,omitzero"`
	DefaultSignatureAlgorithm                                 *string                          `json:"defaultSignatureAlgorithm,omitempty"`
	DirectGrantFlow                                           *string                          `json:"directGrantFlow,omitempty"`
	DisplayName                                               *string                          `json:"displayName,omitempty"`
	DisplayNameHTML                                           *string                          `json:"displayNameHtml,omitempty"`
	DuplicateEmailsAllowed                                    *bool                            `json:"duplicateEmailsAllowed,omitempty"`
	EditUsernameAllowed                                       *bool                            `json:"editUsernameAllowed,omitempty"`
	EmailTheme                                                *string                          `json:"emailTheme,omitempty"`
	Enabled                                                   *bool                            `json:"enabled,omitempty"`
	EnabledEventTypes                                         []string                         `json:"enabledEventTypes,omitzero"`
	EventsEnabled                                             *bool                            `json:"eventsEnabled,omitempty"`
	EventsListeners                                           []string                         `json:"eventsListeners,omitzero"`
	FailureFactor                                             *int                             `json:"failureFactor,omitempty"`
	FederatedUsers                                            []any                            `json:"federatedUsers,omitzero"`
	Groups                                                    []Group                          `json:"groups,omitzero"`
	ID                                                        *string                          `json:"id,omitempty"`
	IdentityProviderMappers                                   []IdentityProviderMapper         `json:"identityProviderMappers,omitzero"`
	IdentityProviders                                         []IdentityProviderRepresentation `json:"identityProviders,omitzero"`
	InternationalizationEnabled                               *bool                            `json:"internationalizationEnabled,omitempty"`
	KeycloakVersion                                           *string                          `json:"keycloakVersion,omitempty"`
	LoginTheme                                                *string                          `json:"loginTheme,omitempty"`
	LocalizationTexts                                         map[string]map[string]string     `json:"localizationTexts,omitzero"`
	LoginWithEmailAllowed                                     *bool                            `json:"loginWithEmailAllowed,omitempty"`
	MaxDeltaTimeSeconds                                       *int                             `json:"maxDeltaTimeSeconds,omitempty"`
	MaxFailureWaitSeconds                                     *int                             `json:"maxFailureWaitSeconds,omitempty"`
	MaxTemporaryLockouts                                      *int                             `json:"maxTemporaryLockouts,omitempty"`
	MinimumQuickLoginWaitSeconds                              *int                             `json:"minimumQuickLoginWaitSeconds,omitempty"`
	NotBefore                                                 *int                             `json:"notBefore,omitempty"`
	OAuth2DeviceCodeLifespan                                  *int                             `json:"oauth2DeviceCodeLifespan,omitempty"`
	OAuth2DevicePollingInterval                               *int                             `json:"oauth2DevicePollingInterval,omitempty"`
	OfflineSessionIdleTimeout                                 *int                             `json:"offlineSessionIdleTimeout,omitempty"`
	OfflineSessionMaxLifespan                                 *int                             `json:"offlineSessionMaxLifespan,omitempty"`
	OfflineSessionMaxLifespanEnabled                          *bool                            `json:"offlineSessionMaxLifespanEnabled,omitempty"`
	OrganizationsEnabled                                      *bool                            `json:"organizationsEnabled,omitempty"`
	OTPPolicyAlgorithm                                        *string                          `json:"otpPolicyAlgorithm,omitempty"`
	OTPPolicyCodeReusable                                     *bool                            `json:"otpPolicyCodeReusable,omitempty"`
	OTPPolicyDigits                                           *int                             `json:"otpPolicyDigits,omitempty"`
	OTPPolicyInitialCounter                                   *int                             `json:"otpPolicyInitialCounter,omitempty"`
	OTPPolicyLookAheadWindow                                  *int                             `json:"otpPolicyLookAheadWindow,omitempty"`
	OTPPolicyPeriod                                           *int                             `json:"otpPolicyPeriod,omitempty"`
	OTPPolicyType                                             *string                          `json:"otpPolicyType,omitempty"`
	OTPSupportedApplications                                  []string                         `json:"otpSupportedApplications,omitzero"`
	PasswordPolicy                                            *string                          `json:"passwordPolicy,omitempty"`
	PermanentLockout                                          *bool                            `json:"permanentLockout,omitempty"`
	ProtocolMappers                                           []any                            `json:"protocolMappers,omitzero"`
	QuickLoginCheckMilliSeconds                               *int                             `json:"quickLoginCheckMilliSeconds,omitempty"`
	Realm                                                     *string                          `json:"realm,omitempty"`
	RevokeRefreshToken                                        *bool                            `json:"revokeRefreshToken,omitempty"`
	RefreshTokenMaxReuse                                      *int                             `json:"refreshTokenMaxReuse,omitempty"`
	RegistrationAllowed                                       *bool                            `json:"registrationAllowed,omitempty"`
	RegistrationEmailAsUsername                               *bool                            `json:"registrationEmailAsUsername,omitempty"`
	RegistrationFlow                                          *string                          `json:"registrationFlow,omitempty"`
	RememberMe                                                *bool                            `json:"rememberMe,omitempty"`
	RequiredActions                                           []any                            `json:"requiredActions,omitzero"`
	ResetCredentialsFlow                                      *string                          `json:"resetCredentialsFlow,omitempty"`
	RequiredCredentials                                       []string                         `json:"requiredCredentials,omitzero"`
	ResetPasswordAllowed                                      *bool                            `json:"resetPasswordAllowed,omitempty"`
	Roles                                                     RolesRepresentation              `json:"roles,omitzero"`
	SSOSessionIdleTimeout                                     *int                             `json:"ssoSessionIdleTimeout,omitempty"`
	SSOSessionIdleTimeoutRememberMe                           *int                             `json:"ssoSessionIdleTimeoutRememberMe,omitempty"`
	SSOSessionMaxLifespan                                     *int                             `json:"ssoSessionMaxLifespan,omitempty"`
	SSOSessionMaxLifespanRememberMe                           *int                             `json:"ssoSessionMaxLifespanRememberMe,omitempty"`
	SMTPServer                                                map[string]string                `json:"smtpServer,omitzero"`
	ScopeMappings                                             []any                            `json:"scopeMappings,omitzero"`
	SSLRequired                                               *string                          `json:"sslRequired,omitempty"`
	SupportedLocales                                          []string                         `json:"supportedLocales,omitzero"`
	UserFederationMappers                                     []any                            `json:"userFederationMappers,omitzero"`
	UserFederationProviders                                   []any                            `json:"userFederationProviders,omitzero"`
	UserManagedAccessAllowed                                  *bool                            `json:"userManagedAccessAllowed,omitempty"`
	Users                                                     []User                           `json:"users,omitzero"`
	VerifyEmail                                               *bool                            `json:"verifyEmail,omitempty"`
	WebAuthnPolicyAcceptableAaguids                           []string                         `json:"webAuthnPolicyAcceptableAaguids,omitzero"`
	WebAuthnPolicyAttestationConveyancePreference             *string                          `json:"webAuthnPolicyAttestationConveyancePreference,omitempty"`
	WebAuthnPolicyAuthenticatorAttachment                     *string                          `json:"webAuthnPolicyAuthenticatorAttachment,omitempty"`
	WebAuthnPolicyAvoidSameAuthenticatorRegister              *bool                            `json:"webAuthnPolicyAvoidSameAuthenticatorRegister,omitempty"`
	WebAuthnPolicyCreateTimeout                               *int                             `json:"webAuthnPolicyCreateTimeout,omitempty"`
	WebAuthnPolicyExtraOrigins                                []string                         `json:"webAuthnPolicyExtraOrigins,omitzero"`
	WebAuthnPolicyRequireResidentKey                          *string                          `json:"webAuthnPolicyRequireResidentKey,omitempty"`
	WebAuthnPolicyRpEntityName                                *string                          `json:"webAuthnPolicyRpEntityName,omitempty"`
	WebAuthnPolicyRpID                                        *string                          `json:"webAuthnPolicyRpId,omitempty"`
	WebAuthnPolicySignatureAlgorithms                         []string                         `json:"webAuthnPolicySignatureAlgorithms,omitzero"`
	WebAuthnPolicyUserVerificationRequirement                 *string                          `json:"webAuthnPolicyUserVerificationRequirement,omitempty"`
	WebAuthnPolicyPasswordlessAcceptableAaguids               []string                         `json:"webAuthnPolicyPasswordlessAcceptableAaguids,omitzero"`
	WebAuthnPolicyPasswordlessAttestationConveyancePreference *string                          `json:"webAuthnPolicyPasswordlessAttestationConveyancePreference,omitempty"`
	WebAuthnPolicyPasswordlessAuthenticatorAttachment         *string                          `json:"webAuthnPolicyPasswordlessAuthenticatorAttachment,omitempty"`
	WebAuthnPolicyPasswordlessAvoidSameAuthenticatorRegister  *bool                            `json:"webAuthnPolicyPasswordlessAvoidSameAuthenticatorRegister,omitempty"`
	WebAuthnPolicyPasswordlessCreateTimeout                   *int                             `json:"webAuthnPolicyPasswordlessCreateTimeout,omitempty"`
	WebAuthnPolicyPasswordlessExtraOrigins                    []string                         `json:"webAuthnPolicyPasswordlessExtraOrigins,omitzero"`
	WebAuthnPolicyPasswordlessRequireResidentKey              *string                          `json:"webAuthnPolicyPasswordlessRequireResidentKey,omitempty"`
	WebAuthnPolicyPasswordlessRpEntityName                    *string                          `json:"webAuthnPolicyPasswordlessRpEntityName,omitempty"`
	WebAuthnPolicyPasswordlessRpID                            *string                          `json:"webAuthnPolicyPasswordlessRpId,omitempty"`
	WebAuthnPolicyPasswordlessSignatureAlgorithms             []string                         `json:"webAuthnPolicyPasswordlessSignatureAlgorithms,omitzero"`
	WebAuthnPolicyPasswordlessUserVerificationRequirement     *string                          `json:"webAuthnPolicyPasswordlessUserVerificationRequirement,omitempty"`
	WaitIncrementSeconds                                      *int                             `json:"waitIncrementSeconds,omitempty"`
}

// AuthenticationFlowRepresentation represents an authentication flow of a realm
type AuthenticationFlowRepresentation struct {
	Alias                    *string                                 `json:"alias,omitempty"`
	AuthenticationExecutions []AuthenticationExecutionRepresentation `json:"authenticationExecutions,omitzero"`
	BuiltIn                  *bool                                   `json:"builtIn,omitempty"`
	Description              *string                                 `json:"description,omitempty"`
	ID                       *string                                 `json:"id,omitempty"`
	ProviderID               *string                                 `json:"providerId,omitempty"`
	TopLevel                 *bool                                   `json:"topLevel,omitempty"`
}

// ConfigPropertyRepresentation represents a configuration property
type ConfigPropertyRepresentation struct {
	Name         string   `json:"name,omitzero"`
	Label        string   `json:"label,omitzero"`
	HelpText     string   `json:"helpText,omitzero"`
	Type         string   `json:"type,omitzero"`
	DefaultValue any      `json:"defaultValue,omitempty"`
	Options      []string `json:"options,omitzero"`
	Secret       bool     `json:"secret,omitzero"`
	Required     bool     `json:"required,omitzero"`
	ReadOnly     bool     `json:"readOnly,omitzero"`
}

// AuthenticatorConfigInfoRepresentation represents an authenticator config description
type AuthenticatorConfigInfoRepresentation struct {
	Name       string                         `json:"name,omitzero"`
	ProviderID string                         `json:"providerId,omitzero"`
	HelpText   string                         `json:"helpText,omitzero"`
	Properties []ConfigPropertyRepresentation `json:"properties,omitzero"`
}

// AuthenticatorConfigRepresentation represents an authenticator config of a realm
type AuthenticatorConfigRepresentation struct {
	ID     *string           `json:"id,omitempty"`
	Alias  *string           `json:"alias,omitempty"`
	Config map[string]string `json:"config,omitzero"`
}

// AuthenticationExecutionRepresentation represents the authentication execution of an AuthenticationFlowRepresentation
type AuthenticationExecutionRepresentation struct {
	Authenticator           *string `json:"authenticator,omitempty"`
	AuthenticatorConfig     *string `json:"authenticatorConfig,omitempty"`
	AuthenticatorFlow       *bool   `json:"authenticatorFlow,omitempty"`
	AuthenticatorFlowLegacy *bool   `json:"autheticatorFlow,omitempty"`
	FlowAlias               *string `json:"flowAlias,omitempty"`
	Priority                *int    `json:"priority,omitempty"`
	Requirement             *string `json:"requirement,omitempty"`
	UserSetupAllowed        *bool   `json:"userSetupAllowed,omitempty"`
}

// CreateAuthenticationExecutionRepresentation contains the provider to be used for a new authentication representation
type CreateAuthenticationExecutionRepresentation struct {
	Provider string `json:"provider,omitzero"`
}

// CreateAuthenticationExecutionFlowRepresentation contains the provider to be used for a new authentication representation
type CreateAuthenticationExecutionFlowRepresentation struct {
	Alias       string `json:"alias,omitzero"`
	Description string `json:"description,omitzero"`
	Provider    string `json:"provider,omitzero"`
	Type        string `json:"type,omitzero"`
}

// ModifyAuthenticationExecutionRepresentation is the payload for updating an execution representation
type ModifyAuthenticationExecutionRepresentation struct {
	ID                   *string  `json:"id,omitempty"`
	ProviderID           *string  `json:"providerId,omitempty"`
	AuthenticationConfig *string  `json:"authenticationConfig,omitempty"`
	AuthenticationFlow   *bool    `json:"authenticationFlow,omitempty"`
	Requirement          *string  `json:"requirement,omitempty"`
	FlowID               *string  `json:"flowId"`
	DisplayName          *string  `json:"displayName,omitempty"`
	Alias                *string  `json:"alias,omitempty"`
	RequirementChoices   []string `json:"requirementChoices,omitzero"`
	Configurable         *bool    `json:"configurable,omitempty"`
	Level                *int     `json:"level,omitempty"`
	Index                *int     `json:"index,omitempty"`
	Description          *string  `json:"description"`
}

// MultiValuedHashMap represents something
type MultiValuedHashMap struct {
	Empty      bool    `json:"empty,omitzero"`
	LoadFactor float32 `json:"loadFactor,omitzero"`
	Threshold  int32   `json:"threshold,omitzero"`
}

// TokenOptions represents the options to obtain a token
type TokenOptions struct {
	ClientID            string   `json:"client_id,omitzero"`
	ClientSecret        string   `json:"-"`
	GrantType           string   `json:"grant_type,omitzero"`
	RefreshToken        string   `json:"refresh_token,omitzero"`
	Scopes              []string `json:"-"`
	Scope               string   `json:"scope,omitzero"`
	ResponseTypes       []string `json:"-"`
	ResponseType        string   `json:"response_type,omitzero"`
	Permission          string   `json:"permission,omitzero"`
	Username            string   `json:"username,omitzero"`
	Password            string   `json:"password,omitzero"`
	Totp                string   `json:"totp,omitzero"`
	Code                string   `json:"code,omitzero"`
	RedirectURI         string   `json:"redirect_uri,omitzero"`
	CodeVerifier        string   `json:"code_verifier,omitzero"` // PKCE (RFC 7636), see NewPKCE
	ClientAssertionType string   `json:"client_assertion_type,omitzero"`
	ClientAssertion     string   `json:"client_assertion,omitzero"`
	SubjectToken        string   `json:"subject_token,omitzero"`
	SubjectTokenType    string   `json:"subject_token_type,omitzero"`
	RequestedSubject    string   `json:"requested_subject,omitzero"`
	Audience            string   `json:"audience,omitzero"`
	RequestedTokenType  string   `json:"requested_token_type,omitzero"`
}

// TokenValidation lists the claims ValidateAccessToken checks in addition to signature, exp and nbf.
type TokenValidation struct {
	// Audience must be contained in the aud claim; empty skips the check.
	Audience string
	// AuthorizedParty must equal the azp claim (the client the token was issued to); empty skips the check.
	AuthorizedParty string
	// Type must equal the typ claim; empty means "Bearer", which rejects ID and refresh tokens.
	Type string
	// Leeway tolerates clock skew between Keycloak and this service when checking exp and nbf.
	Leeway time.Duration
}

// FormData returns the options as form fields for SetFormData, joining Scopes and ResponseTypes and defaulting
// response_type to "token". The receiver is not modified.
// Output: [map[string]string] form fields.
func (t *TokenOptions) FormData() map[string]string {
	o := *t
	t = &o
	if len(t.Scopes) > 0 {
		t.Scope = strings.Join(t.Scopes, " ")
	}

	if len(t.ResponseTypes) > 0 {
		t.ResponseType = strings.Join(t.ResponseTypes, " ")
	}

	if t.ResponseType == "" {
		t.ResponseType = "token"
	}

	m, _ := json.Marshal(t) //nolint:gosec // form data must carry the credentials; marshaling string pointers cannot fail
	var res map[string]string
	_ = json.Unmarshal(m, &res)

	return res
}

// RequestingPartyTokenOptions represents the options to obtain a requesting party token
type RequestingPartyTokenOptions struct {
	GrantType                     string   `json:"grant_type,omitzero"`
	Ticket                        string   `json:"ticket,omitzero"`
	ClaimToken                    string   `json:"claim_token,omitzero"`
	ClaimTokenFormat              string   `json:"claim_token_format,omitzero"`
	RPT                           string   `json:"rpt,omitzero"`
	Permissions                   []string `json:"-"`
	PermissionResourceFormat      string   `json:"permission_resource_format,omitzero"`
	PermissionResourceMatchingURI bool     `json:"permission_resource_matching_uri,string,omitzero"`
	Audience                      string   `json:"audience,omitzero"`
	ResponseIncludeResourceName   *bool    `json:"response_include_resource_name,string,omitempty"`
	ResponsePermissionsLimit      uint32   `json:"response_permissions_limit,omitzero"`
	SubmitRequest                 bool     `json:"submit_request,string,omitzero"`
	ResponseMode                  string   `json:"response_mode,omitzero"`
	SubjectToken                  string   `json:"subject_token,omitzero"`
}

// FormData returns the options as form fields for SetFormData, defaulting grant_type to the UMA ticket grant and
// response_include_resource_name to true. The receiver is not modified.
// Output: [map[string]string] form fields.
func (t *RequestingPartyTokenOptions) FormData() map[string]string {
	o := *t
	t = &o
	if t.GrantType == "" { // required grant type for RPT
		t.GrantType = "urn:ietf:params:oauth:grant-type:uma-ticket"
	}

	if t.ResponseIncludeResourceName == nil { // defaults to true if no value set
		t.ResponseIncludeResourceName = new(true)
	}

	res := make(map[string]string)
	v := reflect.ValueOf(t).Elem()
	typeOf := v.Type()

	for i := range v.NumField() {
		field := v.Field(i)
		jsonTag := typeOf.Field(i).Tag.Get("json")
		if jsonTag == "-" || field.IsZero() {
			continue
		}
		field = reflect.Indirect(field)

		jsonName, _, _ := strings.Cut(jsonTag, ",")
		if jsonName == "" {
			jsonName = typeOf.Field(i).Name
		}

		switch field.Kind() {
		case reflect.String:
			res[jsonName] = field.String()
		case reflect.Bool:
			res[jsonName] = strconv.FormatBool(field.Bool())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			res[jsonName] = strconv.FormatUint(field.Uint(), 10)
		default:
			// other kinds are not used in RequestingPartyTokenOptions
		}
	}
	return res
}

// RequestingPartyPermission is returned by request party token with response type set to "permissions"
type RequestingPartyPermission struct {
	Claims       map[string]string `json:"claims,omitzero"`
	ResourceID   string            `json:"rsid,omitzero"`
	ResourceName string            `json:"rsname,omitzero"`
	Scopes       []string          `json:"scopes,omitzero"`
}

// RequestingPartyPermissionDecision is returned by request party token with response type set to "decision"
type RequestingPartyPermissionDecision struct {
	Result bool `json:"result,omitzero"`
}

// UserSessionRepresentation represents a list of user's sessions
type UserSessionRepresentation struct {
	Clients    map[string]string `json:"clients,omitzero"`
	ID         string            `json:"id,omitzero"`
	IPAddress  string            `json:"ipAddress,omitzero"`
	LastAccess int64             `json:"lastAccess,omitzero"`
	Start      int64             `json:"start,omitzero"`
	UserID     string            `json:"userId,omitzero"`
	Username   string            `json:"username,omitzero"`
}

// SystemInfoRepresentation represents a system info
type SystemInfoRepresentation struct {
	FileEncoding   string `json:"fileEncoding,omitzero"`
	JavaHome       string `json:"javaHome,omitzero"`
	JavaRuntime    string `json:"javaRuntime,omitzero"`
	JavaVendor     string `json:"javaVendor,omitzero"`
	JavaVersion    string `json:"javaVersion,omitzero"`
	JavaVM         string `json:"javaVm,omitzero"`
	JavaVMVersion  string `json:"javaVmVersion,omitzero"`
	OSArchitecture string `json:"osArchitecture,omitzero"`
	OSName         string `json:"osName,omitzero"`
	OSVersion      string `json:"osVersion,omitzero"`
	ServerTime     string `json:"serverTime,omitzero"`
	Uptime         string `json:"uptime,omitzero"`
	UptimeMillis   int    `json:"uptimeMillis,omitzero"`
	UserDir        string `json:"userDir,omitzero"`
	UserLocale     string `json:"userLocale,omitzero"`
	UserName       string `json:"userName,omitzero"`
	UserTimezone   string `json:"userTimezone,omitzero"`
	Version        string `json:"version,omitzero"`
}

// MemoryInfoRepresentation represents a memory info
type MemoryInfoRepresentation struct {
	Free           int    `json:"free,omitzero"`
	FreeFormatted  string `json:"freeFormated,omitzero"`
	FreePercentage int    `json:"freePercentage,omitzero"`
	Total          int    `json:"total,omitzero"`
	TotalFormatted string `json:"totalFormated,omitzero"`
	Used           int    `json:"used,omitzero"`
	UsedFormatted  string `json:"usedFormated,omitzero"`
}

// PasswordPolicy represents the configuration for a supported password policy
type PasswordPolicy struct {
	ConfigType        string `json:"configType,omitempty"`
	DefaultValue      string `json:"defaultValue,omitempty"`
	DisplayName       string `json:"displayName,omitempty"`
	ID                string `json:"id,omitempty"`
	MultipleSupported bool   `json:"multipleSupported,omitempty"`
}

// ProtocolMapperTypeProperty represents a property of a ProtocolMapperType
type ProtocolMapperTypeProperty struct {
	Name         string         `json:"name,omitempty"`
	Label        string         `json:"label,omitempty"`
	HelpText     string         `json:"helpText,omitempty"`
	Type         string         `json:"type,omitempty"`
	Options      []string       `json:"options,omitzero"`
	DefaultValue EnforcedString `json:"defaultValue,omitempty"`
	Secret       bool           `json:"secret,omitempty"`
	ReadOnly     bool           `json:"readOnly,omitempty"`
}

// ProtocolMapperType represents a type of protocol mapper
type ProtocolMapperType struct {
	ID         string                       `json:"id,omitempty"`
	Name       string                       `json:"name,omitempty"`
	Category   string                       `json:"category,omitempty"`
	HelpText   string                       `json:"helpText,omitempty"`
	Priority   int                          `json:"priority,omitempty"`
	Properties []ProtocolMapperTypeProperty `json:"properties,omitzero"`
}

// ProtocolMapperTypes holds the currently available ProtocolMapperType-s grouped by protocol
type ProtocolMapperTypes struct {
	DockerV2      []ProtocolMapperType `json:"docker-v2,omitzero"`
	SAML          []ProtocolMapperType `json:"saml,omitzero"`
	OpenIDConnect []ProtocolMapperType `json:"openid-connect,omitzero"`
}

// BuiltinProtocolMappers holds the currently available built-in blueprints of ProtocolMapper-s grouped by protocol
type BuiltinProtocolMappers struct {
	SAML          []ProtocolMapperRepresentation `json:"saml,omitzero"`
	OpenIDConnect []ProtocolMapperRepresentation `json:"openid-connect,omitzero"`
}

// ServerInfoRepresentation represents a server info
type ServerInfoRepresentation struct {
	SystemInfo             SystemInfoRepresentation `json:"systemInfo,omitzero"`
	MemoryInfo             MemoryInfoRepresentation `json:"memoryInfo,omitzero"`
	PasswordPolicies       []PasswordPolicy         `json:"passwordPolicies,omitzero"`
	ProtocolMapperTypes    ProtocolMapperTypes      `json:"protocolMapperTypes,omitzero"`
	BuiltinProtocolMappers BuiltinProtocolMappers   `json:"builtinProtocolMappers,omitzero"`
	Themes                 Themes                   `json:"themes,omitzero"`
}

// ThemeRepresentation contains the theme name and locales
type ThemeRepresentation struct {
	Name    string   `json:"name,omitempty"`
	Locales []string `json:"locales,omitzero"`
}

// Themes contains the available keycloak themes with locales
type Themes struct {
	Accounts []ThemeRepresentation `json:"account,omitzero"`
	Admin    []ThemeRepresentation `json:"admin,omitzero"`
	Common   []ThemeRepresentation `json:"common,omitzero"`
	Email    []ThemeRepresentation `json:"email,omitzero"`
	Login    []ThemeRepresentation `json:"login,omitzero"`
	Welcome  []ThemeRepresentation `json:"welcome,omitzero"`
}

// FederatedIdentityRepresentation represents an user federated identity
type FederatedIdentityRepresentation struct {
	IdentityProvider string `json:"identityProvider,omitzero"`
	UserID           string `json:"userId,omitzero"`
	UserName         string `json:"userName,omitzero"`
}

// IdentityProviderRepresentation represents an identity provider
type IdentityProviderRepresentation struct {
	AddReadTokenRoleOnCreate  *bool             `json:"addReadTokenRoleOnCreate,omitempty"`
	Alias                     *string           `json:"alias,omitempty"`
	Config                    map[string]string `json:"config,omitzero"`
	DisplayName               *string           `json:"displayName,omitempty"`
	Enabled                   *bool             `json:"enabled,omitempty"`
	FirstBrokerLoginFlowAlias *string           `json:"firstBrokerLoginFlowAlias,omitempty"`
	InternalID                *string           `json:"internalId,omitempty"`
	LinkOnly                  *bool             `json:"linkOnly,omitempty"`
	PostBrokerLoginFlowAlias  *string           `json:"postBrokerLoginFlowAlias,omitempty"`
	ProviderID                *string           `json:"providerId,omitempty"`
	StoreToken                *bool             `json:"storeToken,omitempty"`
	TrustEmail                *bool             `json:"trustEmail,omitempty"`
	HideOnLogin               *bool             `json:"hideOnLogin,omitempty"`
}

// IdentityProviderMapper represents the body of a call to add a mapper to
// an identity provider
type IdentityProviderMapper struct {
	ID                     *string           `json:"id,omitempty"`
	Name                   *string           `json:"name,omitempty"`
	IdentityProviderMapper *string           `json:"identityProviderMapper,omitempty"`
	IdentityProviderAlias  *string           `json:"identityProviderAlias,omitempty"`
	Config                 map[string]string `json:"config"`
}

// GetResourceParams represents the optional parameters for getting resources
type GetResourceParams struct {
	First       *int    `json:"first,string,omitempty"`
	Max         *int    `json:"max,string,omitempty"`
	Name        *string `json:"name,omitempty"`
	Owner       *string `json:"owner,omitempty"`
	Type        *string `json:"type,omitempty"`
	URI         *string `json:"uri,omitempty"`
	Scope       *string `json:"scope,omitempty"`
	MatchingURI *bool   `json:"matchingUri,string,omitempty"`
	ExactName   *bool   `json:"exactName,string,omitempty"`
}

// GetScopeParams represents the optional parameters for getting scopes
type GetScopeParams struct {
	Deep  *bool   `json:"deep,string,omitempty"`
	First *int    `json:"first,string,omitempty"`
	Max   *int    `json:"max,string,omitempty"`
	Name  *string `json:"name,omitempty"`
}

// GetPolicyParams represents the optional parameters for getting policies
// TODO: more policy params?
type GetPolicyParams struct {
	First      *int    `json:"first,string,omitempty"`
	Max        *int    `json:"max,string,omitempty"`
	Name       *string `json:"name,omitempty"`
	Permission *bool   `json:"permission,string,omitempty"`
	Type       *string `json:"type,omitempty"`
}

// GetPermissionParams represents the optional parameters for getting permissions
type GetPermissionParams struct {
	First    *int    `json:"first,string,omitempty"`
	Max      *int    `json:"max,string,omitempty"`
	Name     *string `json:"name,omitempty"`
	Resource *string `json:"resource,omitempty"`
	Scope    *string `json:"scope,omitempty"`
	Type     *string `json:"type,omitempty"`
}

// GetUsersByRoleParams represents the optional parameters for getting users by role
type GetUsersByRoleParams struct {
	First *int `json:"first,string,omitempty"`
	Max   *int `json:"max,string,omitempty"`
}

// PermissionRepresentation is a representation of a RequestingPartyPermission
type PermissionRepresentation struct {
	DecisionStrategy DecisionStrategy `json:"decisionStrategy,omitzero"`
	Description      *string          `json:"description,omitempty"`
	ID               *string          `json:"id,omitempty"`
	Logic            Logic            `json:"logic,omitzero"`
	Name             *string          `json:"name,omitempty"`
	Policies         []string         `json:"policies,omitzero"`
	Resources        []string         `json:"resources,omitzero"`
	ResourceType     *string          `json:"resourceType,omitempty"`
	Scopes           []string         `json:"scopes,omitzero"`
	Type             *string          `json:"type,omitempty"`
}

// CreatePermissionTicketParams represents the optional parameters for getting a permission ticket
type CreatePermissionTicketParams struct {
	ResourceID     string              `json:"resource_id,omitzero"`
	ResourceScopes []string            `json:"resource_scopes,omitzero"`
	Claims         map[string][]string `json:"claims,omitzero"`
}

// PermissionTicketDescriptionRepresentation represents the parameters returned along with a permission ticket
type PermissionTicketDescriptionRepresentation struct {
	ID                     string               `json:"id,omitzero"`
	CreatedTimeStamp       int64                `json:"createdTimestamp,omitzero"`
	UserName               string               `json:"username,omitzero"`
	Enabled                bool                 `json:"enabled,omitzero"`
	TOTP                   bool                 `json:"totp,omitzero"`
	EmailVerified          bool                 `json:"emailVerified,omitzero"`
	FirstName              string               `json:"firstName,omitzero"`
	LastName               string               `json:"lastName,omitzero"`
	Email                  string               `json:"email,omitzero"`
	DisableCredentialTypes []string             `json:"disableCredentialTypes,omitzero"`
	RequiredActions        []string             `json:"requiredActions,omitzero"`
	NotBefore              int64                `json:"notBefore,omitzero"`
	Access                 AccessRepresentation `json:"access,omitzero"`
}

// AccessRepresentation represents the access parameters returned in the permission ticket description
type AccessRepresentation struct {
	ManageGroupMembership bool `json:"manageGroupMembership,omitzero"`
	View                  bool `json:"view,omitzero"`
	MapRoles              bool `json:"mapRoles,omitzero"`
	Impersonate           bool `json:"impersonate,omitzero"`
	Manage                bool `json:"manage,omitzero"`
}

// PermissionTicketResponseRepresentation represents the keycloak response containing the permission ticket
type PermissionTicketResponseRepresentation struct {
	Ticket string `json:"ticket,omitzero"`
}

// PermissionTicketRepresentation represents the permission ticket contents
type PermissionTicketRepresentation struct {
	AZP         string                                     `json:"azp,omitzero"`
	Claims      map[string][]string                        `json:"claims,omitzero"`
	Permissions []PermissionTicketPermissionRepresentation `json:"permissions,omitzero"`
	jwt.RegisteredClaims
}

// PermissionTicketPermissionRepresentation represents the individual permissions in a permission ticket
type PermissionTicketPermissionRepresentation struct {
	Scopes []string `json:"scopes,omitzero"`
	RSID   string   `json:"rsid,omitzero"`
}

// PermissionGrantParams represents the permission which the resource owner is granting to a specific user
type PermissionGrantParams struct {
	ResourceID  string `json:"resource,omitzero"`
	RequesterID string `json:"requester,omitzero"`
	Granted     *bool  `json:"granted,omitempty"`
	ScopeName   string `json:"scopeName,omitzero"`
	TicketID    string `json:"id,omitzero"`
}

// PermissionGrantResponseRepresentation represents the reply from Keycloack after granting permission
type PermissionGrantResponseRepresentation struct {
	ID          string `json:"id,omitzero"`
	Owner       string `json:"owner,omitzero"`
	ResourceID  string `json:"resource,omitzero"`
	Scope       string `json:"scope,omitzero"`
	Granted     bool   `json:"granted,omitzero"`
	RequesterID string `json:"requester,omitzero"`
}

// GetUserPermissionParams represents the optional parameters for getting user permissions
type GetUserPermissionParams struct {
	ScopeID     *string `json:"scopeId,omitempty"`
	ResourceID  *string `json:"resourceId,omitempty"`
	Owner       *string `json:"owner,omitempty"`
	Requester   *string `json:"requester,omitempty"`
	Granted     *bool   `json:"granted,string,omitempty"`
	ReturnNames *string `json:"returnNames,omitempty"`
	First       *int    `json:"first,string,omitempty"`
	Max         *int    `json:"max,string,omitempty"`
}

// ResourcePolicyRepresentation is a representation of a Policy applied to a resource
type ResourcePolicyRepresentation struct {
	Name             *string          `json:"name,omitempty"`
	Description      *string          `json:"description,omitempty"`
	Scopes           []string         `json:"scopes,omitzero"`
	Roles            []string         `json:"roles,omitzero"`
	Groups           []string         `json:"groups,omitzero"`
	Clients          []string         `json:"clients,omitzero"`
	ID               *string          `json:"id,omitempty"`
	Logic            Logic            `json:"logic,omitzero"`
	DecisionStrategy DecisionStrategy `json:"decisionStrategy,omitzero"`
	Owner            *string          `json:"owner,omitempty"`
	Type             *string          `json:"type,omitempty"`
	Users            []string         `json:"users,omitzero"`
}

// PolicyScopeRepresentation is a representation of a scopes of specific policy
type PolicyScopeRepresentation struct {
	ID   string `json:"id,omitzero"`
	Name string `json:"name,omitzero"`
}

// PolicyResourceRepresentation is a representation of a resource of specific policy
type PolicyResourceRepresentation struct {
	ID   string `json:"_id,omitzero"`
	Name string `json:"name,omitzero"`
}

// GetResourcePoliciesParams is a representation of the query params for getting policies
type GetResourcePoliciesParams struct {
	ResourceID *string `json:"resource,omitempty"`
	Name       *string `json:"name,omitempty"`
	Scope      *string `json:"scope,omitempty"`
	First      *int    `json:"first,string,omitempty"`
	Max        *int    `json:"max,string,omitempty"`
}

// GetEventsParams represents the optional parameters for getting events
type GetEventsParams struct {
	Client    *string  `json:"client,omitempty"`
	DateFrom  *string  `json:"dateFrom,omitempty"`
	DateTo    *string  `json:"dateTo,omitempty"`
	First     *int32   `json:"first,string,omitempty"`
	IPAddress *string  `json:"ipAddress,omitempty"`
	Max       *int32   `json:"max,string,omitempty"`
	Type      []string `json:"-"` // sent as repeated "type" query parameter
	UserID    *string  `json:"user,omitempty"`
}

// EventRepresentation is a representation of a Event
type EventRepresentation struct {
	Time      int64             `json:"time,omitempty"`
	Type      string            `json:"type,omitzero"`
	RealmID   string            `json:"realmId,omitzero"`
	ClientID  string            `json:"clientId,omitzero"`
	UserID    string            `json:"userId,omitzero"`
	SessionID string            `json:"sessionId,omitzero"`
	IPAddress string            `json:"ipAddress,omitzero"`
	Details   map[string]string `json:"details,omitzero"`
}

// GetAdminEventsParams represents the optional parameters for getting events
type GetAdminEventsParams struct {
	AuthClient     *string  `json:"authClient,omitempty"`
	AuthIPAddress  *string  `json:"authIpAddress,omitempty"`
	AuthRealm      *string  `json:"authRealm,omitempty"`
	AuthUser       *string  `json:"authUser,omitempty"`
	DateFrom       *string  `json:"dateFrom,omitempty"`
	DateTo         *string  `json:"dateTo,omitempty"`
	First          *int32   `json:"first,string,omitempty"`
	Max            *int32   `json:"max,string,omitempty"`
	OperationTypes []string `json:"-"` // sent as repeated "operationTypes" query parameter
	ResourcePath   *string  `json:"resourcePath,omitempty"`
	ResourceTypes  []string `json:"-"` // sent as repeated "resourceTypes" query parameter
}

// AdminEventRepresentation is a representation of an Admin Event
type AdminEventRepresentation struct {
	Time           int64                               `json:"time,omitempty"`
	OperationType  string                              `json:"operationType,omitzero"`
	RealmID        string                              `json:"realmId,omitzero"`
	ResourceType   string                              `json:"resourceType,omitzero"`
	ResourcePath   string                              `json:"resourcePath,omitzero"`
	AuthDetails    AdminEventAuthDetailsRepresentation `json:"authDetails,omitzero"`
	Representation string                              `json:"representation,omitzero"`
}

// AdminEventAuthDetailsRepresentation is a representation of an Admin Event Details
type AdminEventAuthDetailsRepresentation struct {
	RealmID   string `json:"realmId,omitzero"`
	ClientID  string `json:"clientId,omitzero"`
	UserID    string `json:"userId,omitzero"`
	IPAddress string `json:"ipAddress,omitzero"`
}

// CredentialRepresentation is a representations of the credentials
// v7: https://www.keycloak.org/docs-api/7.0/rest-api/index.html#_credentialrepresentation
// v8: https://www.keycloak.org/docs-api/8.0/rest-api/index.html#_credentialrepresentation
type CredentialRepresentation struct {
	// Common part
	CreatedDate *int64  `json:"createdDate,omitempty"`
	Temporary   *bool   `json:"temporary,omitempty"`
	Type        *string `json:"type,omitempty"`
	Value       *string `json:"value,omitempty"`

	// <= v7
	Algorithm         *string            `json:"algorithm,omitempty"`
	Config            MultiValuedHashMap `json:"config,omitzero"`
	Counter           *int32             `json:"counter,omitempty"`
	Device            *string            `json:"device,omitempty"`
	Digits            *int32             `json:"digits,omitempty"`
	HashIterations    *int32             `json:"hashIterations,omitempty"`
	HashedSaltedValue *string            `json:"hashedSaltedValue,omitempty"`
	Period            *int32             `json:"period,omitempty"`
	Salt              *string            `json:"salt,omitempty"`

	// >= v8
	CredentialData *string `json:"credentialData,omitempty"`
	ID             *string `json:"id,omitempty"`
	Priority       *int32  `json:"priority,omitempty"`
	SecretData     *string `json:"secretData,omitempty"`
	UserLabel      *string `json:"userLabel,omitempty"`
}

// BruteForceStatus is a representation of realm user regarding brute force attack
type BruteForceStatus struct {
	NumFailures   int    `json:"numFailures,omitzero"`
	Disabled      bool   `json:"disabled,omitzero"`
	LastIPFailure string `json:"lastIPFailure,omitzero"`
	LastFailure   int    `json:"lastFailure,omitzero"`
}

// RequiredActionProviderRepresentation is a representation of required actions
// v15: https://www.keycloak.org/docs-api/15.0/rest-api/index.html#_requiredactionproviderrepresentation
type RequiredActionProviderRepresentation struct {
	Alias         *string           `json:"alias,omitempty"`
	Config        map[string]string `json:"config,omitzero"`
	DefaultAction *bool             `json:"defaultAction,omitempty"`
	Enabled       *bool             `json:"enabled,omitempty"`
	Name          *string           `json:"name,omitempty"`
	Priority      *int32            `json:"priority,omitempty"`
	ProviderID    *string           `json:"providerId,omitempty"`
}

// UnregisteredRequiredActionProviderRepresentation is a representation of unregistered required actions.
type UnregisteredRequiredActionProviderRepresentation struct {
	Name       string `json:"name,omitzero"`
	ProviderID string `json:"providerId,omitzero"`
}

// ManagementPermissionRepresentation is a representation of management permissions
// v18: https://www.keycloak.org/docs-api/18.0/rest-api/#_managementpermissionreference
type ManagementPermissionRepresentation struct {
	Enabled          *bool             `json:"enabled,omitempty"`
	Resource         *string           `json:"resource,omitempty"`
	ScopePermissions map[string]string `json:"scopePermissions,omitzero"`
}

// GetClientUserSessionsParams represents the optional parameters for getting user sessions associated with the client
type GetClientUserSessionsParams struct {
	First *int `json:"first,string,omitempty"`
	Max   *int `json:"max,string,omitempty"`
}

// OrganizationInviteUserParams represents the parameters for inviting a new user to an organization
type OrganizationInviteUserParams struct {
	Email     string `json:"email,omitzero"`
	FirstName string `json:"firstName,omitzero"`
	LastName  string `json:"lastName,omitzero"`
}

// FormData returns the OrganizationInviteUserParams as form data, keyed by their JSON field names.
// Output: [map[string]string] form fields.
func (v *OrganizationInviteUserParams) FormData() map[string]string {
	m, _ := json.Marshal(v)
	var res map[string]string
	_ = json.Unmarshal(m, &res)
	return res
}

// MembershipType represent the membership type of an organization member.
// v26: https://www.keycloak.org/docs-api/latest/rest-api/index.html#MembershipType
type MembershipType string

const (
	// MembershipTypeManaged indicates the member cannot exist without the organization.
	MembershipTypeManaged MembershipType = "MANAGED"
	// MembershipTypeUnmanaged indicates the member can exist without the organization.
	MembershipTypeUnmanaged MembershipType = "UNMANAGED"
)

// MemberRepresentation represents a member of an organization
// v26: https://www.keycloak.org/docs-api/latest/rest-api/index.html#MemberRepresentation
type MemberRepresentation struct {
	User
	MembershipType MembershipType `json:"membershipType,omitzero"`
}

// GetMembersParams represents the optional parameters for getting members of an organization
type GetMembersParams struct {
	Exact          *bool           `json:"exact,string,omitempty"`
	First          *int            `json:"first,string,omitempty"`
	Max            *int            `json:"max,string,omitempty"`
	MembershipType *MembershipType `json:"membershipType,omitempty"`
	Search         *string         `json:"search,omitempty"`
}

// GetOrganizationMembersParams represents the optional parameters for getting members of an organization group
type GetOrganizationMembersParams struct {
	BriefRepresentation *bool `json:"briefRepresentation,string,omitempty"`
	First               *int  `json:"first,string,omitempty"`
	Max                 *int  `json:"max,string,omitempty"`
}

// GetOrganizationsParams represents the optional parameters for getting organizations
type GetOrganizationsParams struct {
	BriefRepresentation *bool   `json:"briefRepresentation,string,omitempty"`
	Exact               *bool   `json:"exact,string,omitempty"`
	First               *int    `json:"first,string,omitempty"`
	Max                 *int    `json:"max,string,omitempty"`
	Q                   *string `json:"q,omitempty"`
	Search              *string `json:"search,omitempty"`
}

// GetOrganizationCountParams represents the optional parameters for counting organizations
type GetOrganizationCountParams struct {
	Exact  *bool   `json:"exact,string,omitempty"`
	Q      *string `json:"q,omitempty"`
	Search *string `json:"search,omitempty"`
}

// GetOrganizationInvitationsParams represents the optional parameters for getting organization invitations
type GetOrganizationInvitationsParams struct {
	Email     *string           `json:"email,omitempty"`
	First     *int              `json:"first,string,omitempty"`
	FirstName *string           `json:"firstName,omitempty"`
	LastName  *string           `json:"lastName,omitempty"`
	Max       *int              `json:"max,string,omitempty"`
	Search    *string           `json:"search,omitempty"`
	Status    *InvitationStatus `json:"status,omitempty"`
}

// GetOrganizationGroupsParams represents the optional parameters for getting organization groups
type GetOrganizationGroupsParams struct {
	BriefRepresentation *bool   `json:"briefRepresentation,string,omitempty"`
	Exact               *bool   `json:"exact,string,omitempty"`
	First               *int    `json:"first,string,omitempty"`
	Max                 *int    `json:"max,string,omitempty"`
	PopulateHierarchy   *bool   `json:"populateHierarchy,string,omitempty"`
	Q                   *string `json:"q,omitempty"`
	Search              *string `json:"search,omitempty"`
	SubGroupsCount      *bool   `json:"subGroupsCount,string,omitempty"`
}

// GetOrganizationGroupSubgroupsParams represents the optional parameters for getting subgroups of an organization group
type GetOrganizationGroupSubgroupsParams struct {
	Exact          *bool   `json:"exact,string,omitempty"`
	First          *int    `json:"first,string,omitempty"`
	Max            *int    `json:"max,string,omitempty"`
	Search         *string `json:"search,omitempty"`
	SubGroupsCount *bool   `json:"subGroupsCount,string,omitempty"`
}

// OrganizationDomainRepresentation is a representation of an organization's domain
// v26: https://www.keycloak.org/docs-api/latest/rest-api/index.html#OrganizationDomainRepresentation
type OrganizationDomainRepresentation struct {
	Name     *string `json:"name,omitempty"`
	Verified *bool   `json:"verified,omitempty"`
}

// OrganizationRepresentation is a representation of an organization
// v26: https://www.keycloak.org/docs-api/latest/rest-api/index.html#OrganizationRepresentation
type OrganizationRepresentation struct {
	ID                *string                            `json:"id,omitempty"`
	Name              *string                            `json:"name,omitempty"`
	Alias             *string                            `json:"alias,omitempty"`
	Enabled           *bool                              `json:"enabled,omitempty"`
	Description       *string                            `json:"description,omitempty"`
	RedirectURL       *string                            `json:"redirectUrl,omitempty"`
	Attributes        map[string][]string                `json:"attributes,omitzero"`
	Domains           []OrganizationDomainRepresentation `json:"domains,omitzero"`
	Members           []MemberRepresentation             `json:"members,omitzero"`
	IdentityProviders []IdentityProviderRepresentation   `json:"identityProviders,omitzero"`
	Groups            []Group                            `json:"groups,omitzero"`
}

// InvitationStatus represents the status of an organization invitation
type InvitationStatus string

const (
	// InvitationStatusPending indicates the invitation has not yet been accepted
	InvitationStatusPending InvitationStatus = "PENDING"
	// InvitationStatusExpired indicates the invitation has expired
	InvitationStatusExpired InvitationStatus = "EXPIRED"
)

// OrganizationInvitationRepresentation represents an invitation to join an organization
// v26: https://www.keycloak.org/docs-api/latest/rest-api/index.html#OrganizationInvitationRepresentation
type OrganizationInvitationRepresentation struct {
	ID             string           `json:"id,omitzero"`
	OrganizationID string           `json:"organizationId,omitzero"`
	Email          string           `json:"email,omitzero"`
	FirstName      string           `json:"firstName,omitzero"`
	LastName       string           `json:"lastName,omitzero"`
	SentDate       int64            `json:"sentDate,omitzero"`
	ExpiresAt      int64            `json:"expiresAt,omitzero"`
	Status         InvitationStatus `json:"status,omitzero"`
	InviteLink     string           `json:"inviteLink,omitzero"`
}

// UserProfileAttribute represents a user profile attribute
type UserProfileAttribute struct {
	Name        *string                         `json:"name,omitempty"`
	DisplayName *string                         `json:"displayName,omitempty"`
	Validations map[string]any                  `json:"validations,omitzero"`
	Annotations map[string]any                  `json:"annotations,omitzero"`
	Required    UserProfileAttributeRequired    `json:"required,omitzero"`
	Permissions UserProfileAttributePermissions `json:"permissions,omitzero"`
	Selector    UserProfileAttributeSelector    `json:"selector,omitzero"`
	Group       *string                         `json:"group,omitempty"`
	MultiValued *bool                           `json:"multivalued,omitempty"`
}

// UserProfileAttributePermissions represents the permissions of a user profile attribute
type UserProfileAttributePermissions struct {
	View []string `json:"view,omitzero"`
	Edit []string `json:"edit,omitzero"`
}

// UserProfileAttributeRequired represents the required roles and scopes for a user profile attribute
type UserProfileAttributeRequired struct {
	Roles  []string `json:"roles,omitzero"`
	Scopes []string `json:"scopes,omitzero"`
}

// UserProfileAttributeSelector represents the scopes for selecting a user profile attribute
type UserProfileAttributeSelector struct {
	Scopes []string `json:"scopes,omitzero"`
}

// UserProfileConfig represents the configuration of user profile
type UserProfileConfig struct {
	Attributes               []UserProfileAttribute   `json:"attributes,omitzero"`
	Groups                   []UserProfileGroup       `json:"groups,omitzero"`
	UnmanagedAttributePolicy UnmanagedAttributePolicy `json:"unmanagedAttributePolicy,omitzero"`
}

// UserProfileGroup represents a user profile group
type UserProfileGroup struct {
	Name               *string        `json:"name,omitempty"`
	DisplayHeader      *string        `json:"displayHeader,omitempty"`
	DisplayDescription *string        `json:"displayDescription,omitempty"`
	Annotations        map[string]any `json:"annotations,omitzero"`
}

// UnmanagedAttributePolicy represents the policy for unmanaged attributes in user profile
type UnmanagedAttributePolicy string

const (
	// UnmanagedAttributePolicyEnabled indicates that unmanaged attributes are enabled and can be used in user profile
	UnmanagedAttributePolicyEnabled UnmanagedAttributePolicy = "ENABLED"
	// UnmanagedAttributePolicyAdminView indicates that unmanaged attributes are disabled and cannot be used in user profile
	UnmanagedAttributePolicyAdminView UnmanagedAttributePolicy = "ADMIN_VIEW"
	// UnmanagedAttributePolicyAdminEdit indicates that unmanaged attributes are enabled but can only be viewed and edited by admin in user profile
	UnmanagedAttributePolicyAdminEdit UnmanagedAttributePolicy = "ADMIN_EDIT"
)

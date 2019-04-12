package gocloak

import "encoding/json"

// GetQueryParams converts the struct to map[string]string
// The fields tags must have `json:"<name>,string,omitempty"` format for all types, except strings
// The string fields must have: `json:"<name>,omitempty"`. The `json:"<name>,string,omitempty"` tag for string field
// will add additional double quotes.
// "string" tag allows to convert the non-string fields of a structure to map[string]string.
// "omitempty" allows to skip the fields with default values.
func GetQueryParams(s interface{}) (map[string]string, error) {
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

// APIError represents an api error
type APIError struct {
	Code    int
	Message string
}

// Error stringifies the APIError
func (apiError APIError) Error() string {
	return apiError.Message
}

// CertResponseKey is returned by the certs endpoint
type CertResponseKey struct {
	Kid string `json:"kid,omitempty"`
	Kty string `json:"kty,omitempty"`
	Alg string `json:"alg,omitempty"`
	Use string `json:"use,omitempty"`
	N   string `json:"n,omitempty"`
	E   string `json:"e,omitempty"`
}

// CertResponse is returned by the certs endpoint
type CertResponse struct {
	Keys []CertResponseKey `json:"keys,omitempty"`
}

// IssuerResponse is returned by the issuer endpoint
type IssuerResponse struct {
	Realm           string `json:"realm,omitempty"`
	PublicKey       string `json:"public_key,omitempty"`
	TokenService    string `json:"token-service,omitempty"`
	AccountService  string `json:"account-service,omitempty"`
	TokensNotBefore int    `json:"tokens-not-before,omitempty"`
}

// RetrospecTokenResult is returned when a token was checked
type RetrospecTokenResult struct {
	Permissions map[string]string `json:"permissions,omitempty,omitempty"`
	Exp         int               `json:"exp,omitempty,omitempty"`
	Nbf         int               `json:"nbf,omitempty,omitempty"`
	Iat         int               `json:"iat,omitempty,omitempty"`
	Aud         string            `json:"aud,omitempty,omitempty"`
	Active      bool              `json:"active,omitempty,omitempty"`
	AuthTime    int               `json:"auth_time,omitempty,omitempty"`
	Jti         string            `json:"jti,omitempty,omitempty"`
	Type        string            `json:"typ,omitempty,omitempty"`
}

// User represents the Keycloak User Structure
type User struct {
	ID                         string              `json:"id,omitempty,omitempty"`
	CreatedTimestamp           int64               `json:"createdTimestamp,omitempty,omitempty"`
	Username                   string              `json:"username,omitempty,omitempty"`
	Enabled                    bool                `json:"enabled,omitempty,omitempty"`
	Totp                       bool                `json:"totp,omitempty,omitempty"`
	EmailVerified              bool                `json:"emailVerified,omitempty,omitempty"`
	FirstName                  string              `json:"firstName,omitempty,omitempty"`
	LastName                   string              `json:"lastName,omitempty,omitempty"`
	Email                      string              `json:"email,omitempty,omitempty"`
	FederationLink             string              `json:"federationLink,omitempty,omitempty"`
	Attributes                 map[string][]string `json:"attributes,omitempty,omitempty"`
	DisableableCredentialTypes []interface{}       `json:"disableableCredentialTypes,omitempty,omitempty"`
	RequiredActions            []interface{}       `json:"requiredActions,omitempty,omitempty"`
	Access                     map[string]bool     `json:"access,omitempty,omitempty"`
}

// SetPasswordRequest sets a new password
type SetPasswordRequest struct {
	Type      string `json:"type,omitempty"`
	Temporary bool   `json:"temporary,omitempty"`
	Password  string `json:"value,omitempty"`
}

// Component is a component
type Component struct {
	ID              string          `json:"id,omitempty"`
	Name            string          `json:"name,omitempty"`
	ProviderID      string          `json:"providerId,omitempty"`
	ProviderType    string          `json:"providerType,omitempty"`
	ParentID        string          `json:"parentId,omitempty"`
	ComponentConfig ComponentConfig `json:"config,omitempty"`
	SubType         string          `json:"subType,omitempty,omitempty"`
}

// ComponentConfig is a componentconfig
type ComponentConfig struct {
	Priority  []string `json:"priority,omitempty"`
	Algorithm []string `json:"algorithm,omitempty"`
}

// KeyStoreConfig holds the keyStoreConfig
type KeyStoreConfig struct {
	ActiveKeys ActiveKeys `json:"active,omitempty"`
	Key        []Key      `json:"keys,omitempty"`
}

// ActiveKeys holds the active keys
type ActiveKeys struct {
	HS256 string `json:"HS256,omitempty"`
	RS256 string `json:"RS256,omitempty"`
	AES   string `json:"AES,omitempty"`
}

// Key is a key
type Key struct {
	ProviderID       string `json:"providerId,omitempty"`
	ProviderPriority int    `json:"providerPriority,omitempty"`
	Kid              string `json:"kid,omitempty"`
	Status           string `json:"status,omitempty"`
	Type             string `json:"type,omitempty"`
	Algorithm        string `json:"algorithm,omitempty"`
	PublicKey        string `json:"publicKey,omitempty,omitempty"`
	Certificate      string `json:"certificate,omitempty,omitempty"`
}

// Attributes holds Attributes
type Attributes struct {
	LDAPENTRYDN []string `json:"LDAP_ENTRY_DN,omitempty"`
	LDAPID      []string `json:"LDAP_ID,omitempty"`
}

// Access represents access
type Access struct {
	ManageGroupMembership bool `json:"manageGroupMembership,omitempty"`
	View                  bool `json:"view,omitempty"`
	MapRoles              bool `json:"mapRoles,omitempty"`
	Impersonate           bool `json:"impersonate,omitempty"`
	Manage                bool `json:"manage,omitempty"`
}

// UserGroup is a UserGroup
type UserGroup struct {
	ID   string `json:"id,omitempty,omitempty"`
	Name string `json:"name,omitempty,omitempty"`
	Path string `json:"path,omitempty,omitempty"`
}

// GetUsersParams represents the optional parameters for getting users
type GetUsersParams struct {
	BriefRepresentation *bool  `json:"briefRepresentation,string,omitempty,omitempty"`
	Email               string `json:"email,omitempty,omitempty"`
	First               int    `json:"first,string,omitempty,omitempty"`
	FirstName           string `json:"firstName,omitempty,omitempty"`
	LastName            string `json:"lastName,omitempty,omitempty"`
	Max                 int    `json:"max,string,omitempty,omitempty"`
	Search              string `json:"search,omitempty,omitempty"`
	Username            string `json:"username,omitempty,omitempty"`
}

// ExecuteActionsEmail represents parameters for executing action emails
type ExecuteActionsEmail struct {
	UserID      string   `json:"-,omitempty"`
	ClientID    string   `json:"client_id,omitempty,omitempty"`
	Lifespan    int      `json:"lifespan,string,omitempty,omitempty"`
	RedirectURI string   `json:"redirect_uri,omitempty,omitempty"`
	Actions     []string `json:"-,omitempty"`
}

// Group is a Group
type Group struct {
	ID        string        `json:"id,omitempty,omitempty"`
	Name      string        `json:"name,omitempty,omitempty"`
	Path      string        `json:"path,omitempty,omitempty"`
	SubGroups []interface{} `json:"subGroups,omitempty,omitempty"`
}

// GetGroupsParams represents the optional parameters for getting groups
type GetGroupsParams struct {
	First  int    `json:"first,string,omitempty,omitempty"`
	Max    int    `json:"max,string,omitempty,omitempty"`
	Search string `json:"search,omitempty,omitempty"`
}

// Role is a role
type Role struct {
	ID                 string              `json:"id,omitempty,omitempty"`
	Name               string              `json:"name,omitempty,omitempty"`
	ScopeParamRequired bool                `json:"scopeParamRequired,omitempty,omitempty"`
	Composite          bool                `json:"composite,omitempty,omitempty"`
	ClientRole         bool                `json:"clientRole,omitempty,omitempty"`
	ContainerID        string              `json:"containerId,omitempty,omitempty"`
	Description        string              `json:"description,omitempty,omitempty"`
	Attributes         map[string][]string `json:"attributes,omitempty,omitempty"`
}

// ClientMappingsRepresentation is a client role mappings
type ClientMappingsRepresentation struct {
	ID       string `json:"id,omitempty"`
	Client   string `json:"client,omitempty"`
	Mappings []Role `json:"mappings,omitempty"`
}

// MappingsRepresentation is a representation of role mappings
type MappingsRepresentation struct {
	ClientMappings map[string]ClientMappingsRepresentation `json:"clientMappings,omitempty,omitempty"`
	RealmMappings  []Role                                  `json:"realmMappings,omitempty,omitempty"`
}

// ClientScope is a ClientScope
type ClientScope struct {
	ID                    string                `json:"id,omitempty"`
	Name                  string                `json:"name,omitempty"`
	Description           string                `json:"description,omitempty"`
	Protocol              string                `json:"protocol,omitempty"`
	ClientScopeAttributes ClientScopeAttributes `json:"attributes,omitempty"`
	ProtocolMappers       ProtocolMappers       `json:"protocolMappers,omitempty,omitempty"`
}

// ClientScopeAttributes are attributes of client scopes
type ClientScopeAttributes struct {
	ConsentScreenText      string `json:"consent.screen.text,omitempty"`
	DisplayOnConsentScreen string `json:"display.on.consent.screen,omitempty"`
}

// ProtocolMappers are protocolmappers
type ProtocolMappers struct {
	ID                    string                `json:"id,omitempty"`
	Name                  string                `json:"name,omitempty"`
	Protocol              string                `json:"protocol,omitempty"`
	ProtocolMapper        string                `json:"protocolMapper,omitempty"`
	ConsentRequired       bool                  `json:"consentRequired,omitempty"`
	ProtocolMappersConfig ProtocolMappersConfig `json:"config,omitempty"`
}

// ProtocolMappersConfig is a config of a protocol mapper
type ProtocolMappersConfig struct {
	UserinfoTokenClaim string `json:"userinfo.token.claim,omitempty"`
	UserAttribute      string `json:"user.attribute,omitempty"`
	IDTokenClaim       string `json:"id.token.claim,omitempty"`
	AccessTokenClaim   string `json:"access.token.claim,omitempty"`
	ClaimName          string `json:"claim.name,omitempty"`
	JSONTypeLabel      string `json:"jsonType.label,omitempty"`
}

// Client is a Client
type Client struct {
	ID       string `json:"id,omitempty"`
	ClientID string `json:"clientId,omitempty"`
}

// GetClientsParams represents the query parameters
type GetClientsParams struct {
	ClientID     string `json:"clientId,omitempty,omitempty"`
	ViewableOnly bool   `json:"viewableOnly,string,omitempty,omitempty"`
}

// UserInfo is returned by the userinfo endpoint
type UserInfo struct {
	Sub               string      `json:"sub,omitempty"`
	EmailVerified     bool        `json:"email_verified,omitempty"`
	Address           interface{} `json:"address,omitempty"`
	PreferredUsername string      `json:"preferred_username,omitempty"`
	Email             string      `json:"email,omitempty"`
}

// RealmRepresentation represent a realm
type RealmRepresentation struct {
	AccessCodeLifespan                  int               `json:"accessCodeLifespan,omitempty"`
	AccessCodeLifespanLogin             int               `json:"accessCodeLifespanLogin,omitempty"`
	AccessCodeLifespanUserAction        int               `json:"accessCodeLifespanUserAction,omitempty"`
	AccessTokenLifespan                 int               `json:"accessTokenLifespan,omitempty"`
	AccessTokenLifespanForImplicitFlow  int               `json:"accessTokenLifespanForImplicitFlow,omitempty"`
	AccountTheme                        string            `json:"accountTheme,omitempty"`
	ActionTokenGeneratedByAdminLifespan int               `json:"actionTokenGeneratedByAdminLifespan,omitempty"`
	ActionTokenGeneratedByUserLifespan  int               `json:"actionTokenGeneratedByUserLifespan,omitempty"`
	AdminEventsDetailsEnabled           bool              `json:"adminEventsDetailsEnabled,omitempty"`
	AdminEventsEnabled                  bool              `json:"adminEventsEnabled,omitempty"`
	AdminTheme                          string            `json:"adminTheme,omitempty"`
	Attributes                          map[string]string `json:"attributes,omitempty"`
	AuthenticationFlows                 []interface{}     `json:"authenticationFlows,omitempty"`
	AuthenticatorConfig                 []interface{}     `json:"authenticatorConfig,omitempty"`
	BrowserFlow                         string            `json:"browserFlow,omitempty"`
	BrowserSecurityHeaders              map[string]string `json:"browserSecurityHeaders,omitempty"`
	BruteForceProtected                 bool              `json:"bruteForceProtected,omitempty"`
	ClientAuthenticationFlow            string            `json:"clientAuthenticationFlow,omitempty"`
	ClientScopeMappings                 map[string]string `json:"clientScopeMappings,omitempty"`
	ClientScopes                        []interface{}     `json:"clientScopes,omitempty"`
	Clients                             []interface{}     `json:"clients,omitempty"`
	Components                          interface{}       `json:"components,omitempty"`
	DefaultDefaultClientScopes          []string          `json:"defaultDefaultClientScopes,omitempty"`
	DefaultGroups                       []string          `json:"defaultGroups,omitempty"`
	DefaultLocale                       string            `json:"defaultLocale,omitempty"`
	DefaultOptionalClientScopes         []string          `json:"defaultOptionalClientScopes,omitempty"`
	DefaultRoles                        []string          `json:"defaultRoles,omitempty"`
	DefaultSignatureAlgorithm           string            `json:"defaultSignatureAlgorithm,omitempty"`
	DirectGrantFlow                     string            `json:"directGrantFlow,omitempty"`
	DisplayName                         string            `json:"displayName,omitempty"`
	DisplayNameHTML                     string            `json:"displayNameHtml,omitempty"`
	DockerAuthenticationFlow            string            `json:"dockerAuthenticationFlow,omitempty"`
	DuplicateEmailsAllowed              bool              `json:"duplicateEmailsAllowed,omitempty"`
	EditUsernameAllowed                 bool              `json:"editUsernameAllowed,omitempty"`
	EmailTheme                          string            `json:"emailTheme,omitempty"`
	Enabled                             bool              `json:"enabled,omitempty"`
	EnabledEventTypes                   []string          `json:"enabledEventTypes,omitempty"`
	EventsEnabled                       bool              `json:"eventsEnabled,omitempty"`
	EventsExpiration                    int64             `json:"eventsExpiration,omitempty"`
	EventsListeners                     []string          `json:"eventsListeners,omitempty"`
	FailureFactor                       int               `json:"failureFactor,omitempty"`
	FederatedUsers                      []interface{}     `json:"federatedUsers,omitempty"`
	Groups                              []interface{}     `json:"groups,omitempty"`
	ID                                  string            `json:"id,omitempty"`
	IdentityProviderMappers             []interface{}     `json:"identityProviderMappers,omitempty"`
	IdentityProviders                   []interface{}     `json:"identityProviders,omitempty"`
	InternationalizationEnabled         bool              `json:"internationalizationEnabled,omitempty"`
	KeycloakVersion                     string            `json:"keycloakVersion,omitempty"`
	LoginTheme                          string            `json:"loginTheme,omitempty"`
	LoginWithEmailAllowed               bool              `json:"loginWithEmailAllowed,omitempty"`
	MaxDeltaTimeSeconds                 int               `json:"maxDeltaTimeSeconds,omitempty"`
	MaxFailureWaitSeconds               int               `json:"maxFailureWaitSeconds,omitempty"`
	MinimumQuickLoginWaitSeconds        int               `json:"minimumQuickLoginWaitSeconds,omitempty"`
	NotBefore                           int               `json:"notBefore,omitempty"`
	OfflineSessionIdleTimeout           int               `json:"offlineSessionIdleTimeout,omitempty"`
	OfflineSessionMaxLifespan           int               `json:"offlineSessionMaxLifespan,omitempty"`
	OfflineSessionMaxLifespanEnabled    bool              `json:"offlineSessionMaxLifespanEnabled,omitempty"`
	OtpPolicyAlgorithm                  string            `json:"otpPolicyAlgorithm,omitempty"`
	OtpPolicyDigits                     int               `json:"otpPolicyDigits,omitempty"`
	OtpPolicyInitialCounter             int               `json:"otpPolicyInitialCounter,omitempty"`
	OtpPolicyLookAheadWindow            int               `json:"otpPolicyLookAheadWindow,omitempty"`
	OtpPolicyPeriod                     int               `json:"otpPolicyPeriod,omitempty"`
	OtpPolicyType                       string            `json:"otpPolicyType,omitempty"`
	OtpSupportedApplications            []string          `json:"otpSupportedApplications,omitempty"`
	PasswordPolicy                      string            `json:"passwordPolicy,omitempty"`
	PermanentLockout                    bool              `json:"permanentLockout,omitempty"`
	ProtocolMappers                     []interface{}     `json:"protocolMappers,omitempty"`
	QuickLoginCheckMilliSeconds         int64             `json:"quickLoginCheckMilliSeconds,omitempty"`
	Realm                               string            `json:"realm,omitempty"`
	RefreshTokenMaxReuse                int               `json:"refreshTokenMaxReuse,omitempty"`
	RegistrationAllowed                 bool              `json:"registrationAllowed,omitempty"`
	RegistrationEmailAsUsername         bool              `json:"registrationEmailAsUsername,omitempty"`
	RegistrationFlow                    string            `json:"registrationFlow,omitempty"`
	RememberMe                          bool              `json:"rememberMe,omitempty"`
	RequiredActions                     []interface{}     `json:"requiredActions,omitempty"`
	ResetCredentialsFlow                string            `json:"resetCredentialsFlow,omitempty"`
	ResetPasswordAllowed                bool              `json:"resetPasswordAllowed,omitempty"`
	RevokeRefreshToken                  bool              `json:"revokeRefreshToken,omitempty"`
	Roles                               interface{}       `json:"roles,omitempty"`
	ScopeMappings                       []interface{}     `json:"scopeMappings,omitempty"`
	SMTPServer                          map[string]string `json:"smtpServer,omitempty"`
	SslRequired                         string            `json:"sslRequired,omitempty"`
	SsoSessionIdleTimeout               int               `json:"ssoSessionIdleTimeout,omitempty"`
	SsoSessionIdleTimeoutRememberMe     int               `json:"ssoSessionIdleTimeoutRememberMe,omitempty"`
	SsoSessionMaxLifespan               int               `json:"ssoSessionMaxLifespan,omitempty"`
	SsoSessionMaxLifespanRememberMe     int               `json:"ssoSessionMaxLifespanRememberMe,omitempty"`
	SupportedLocales                    []string          `json:"supportedLocales,omitempty"`
	UserFederationMappers               []interface{}     `json:"userFederationMappers,omitempty"`
	UserFederationProviders             []interface{}     `json:"userFederationProviders,omitempty"`
	UserManagedAccessAllowed            bool              `json:"userManagedAccessAllowed,omitempty"`
	Users                               []interface{}     `json:"users,omitempty"`
	VerifyEmail                         bool              `json:"verifyEmail,omitempty"`
	WaitIncrementSeconds                int               `json:"waitIncrementSeconds,omitempty"`
}

// MultivaluedHashMap represents something
type MultivaluedHashMap struct {
	Empty      bool    `json:"empty,omitempty"`
	LoadFactor float32 `json:"loadFactor,omitempty"`
	Threshold  int32   `json:"threshold,omitempty"`
}

// CredentialRepresentation represents credentials
type CredentialRepresentation struct {
	Algorithm         string             `json:"algorithm,omitempty"`
	Config            MultivaluedHashMap `json:"config,omitempty"`
	Counter           int32              `json:"counter,omitempty"`
	CreatedDate       int64              `json:"createdDate,omitempty"`
	Device            string             `json:"device,omitempty"`
	Digits            int32              `json:"digits,omitempty"`
	HashIterations    int32              `json:"hashIterations,omitempty"`
	HashedSaltedValue string             `json:"hashedSaltedValue,omitempty"`
	Period            int32              `json:"period,omitempty"`
	Salt              string             `json:"salt,omitempty"`
	Temporary         bool               `json:"temporary,omitempty"`
	Type              string             `json:"type,omitempty"`
	Value             string             `json:"value,omitempty"`
}

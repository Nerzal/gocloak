```go
type Tkc struct {
	realm  string
	token  *gocloak.JWT
	client *gocloak.GoCloak
}

func (tkc *Tkc) getRealms(ctx context.Context) (realms map[string]gocloak.RealmRepresentation, getErr error) {
	var realmList []*gocloak.RealmRepresentation

	// init map
	realms = make(map[string]gocloak.RealmRepresentation)

	// get all realms
	if realmList, getErr = tkc.client.GetRealms(ctx, tkc.token.AccessToken); getErr != nil {
		getErr = fmt.Errorf("get realms failed (error: %s)", getErr.Error())
		return realms, getErr
	}

	// transform to map with realmID (meaning realm name) as map key
	for _, r := range realmList {
		realms[*r.Realm] = *r
	}

	return realms, nil
}

func (tkc *Tkc) newUserFederation(ctx context.Context) (userFederation gocloak.Component, newErr error) {
	var idRealm string

	// get realm ID, is needed as ParentID in gocloak.Component
	if realms, newErr := tkc.getRealms(ctx); newErr != nil {
		return userFederation, newErr
	} else {
		idRealm = *realms[tkc.realm].ID
	}

	userFederationConfig := make(map[string][]string)

	// keycloak self
	userFederationConfig["enabled"] = []string{"true"}
	userFederationConfig["priority"] = []string{"1"}
	userFederationConfig["importUsers"] = []string{"true"}

	// sync options
	userFederationConfig["fullSyncPeriod"] = []string{"-1"}
	userFederationConfig["changedSyncPeriod"] = []string{"300"}
	userFederationConfig["batchSizeForSync"] = []string{"1000"}

	// ldap connection
	userFederationConfig["editMode"] = []string{"READ_ONLY"}
	userFederationConfig["vendor"] = []string{"other"}
	userFederationConfig["connectionUrl"] = []string{"ldap://ldap"}
	userFederationConfig["bindDn"] = []string{"cn=XXX,dc=example,dc=com"}
	userFederationConfig["bindCredential"] = []string{"YYYYYY"}
	userFederationConfig["usersDn"] = []string{"ou=users,dc=example,dc=com"}
	userFederationConfig["usernameLDAPAttribute"] = []string{"uid"}
	userFederationConfig["uuidLDAPAttribute"] = []string{"entryUUID"}
	userFederationConfig["authType"] = []string{"simple"}
	userFederationConfig["userObjectClasses"] = []string{"person, uidObject"}
	userFederationConfig["rdnLDAPAttribute"] = []string{"cn"}
	userFederationConfig["searchScope"] = []string{"1"}
	userFederationConfig["pagination"] = []string{"true"}

	userFederation = gocloak.Component{
		Name:            gocloak.StringP("ldap"),
		ProviderID:      gocloak.StringP("ldap"),
		ProviderType:    gocloak.StringP("org.keycloak.storage.UserStorageProvider"),
		ParentID:        gocloak.StringP(idRealm),
		ComponentConfig: &userFederationConfig,
	}

	return userFederation, nil
}

func (tkc *Tkc) RunCreateUserFederation() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer func() {
		cancel()
	}()

	// user federation is a keycloak component with type 'org.keycloak.storage.UserStorageProvider'
	var userFederation gocloak.Component

	// get default new legacy user federation to create or update an existing one
	if newUserFederation, runErr := tkc.newUserFederation(ctx); runErr != nil {
		return runErr
	} else {
		userFederation = newUserFederation
	}

	// search parameter to get exactly legacy user federation which belongs to given realm (parent)
	ldapGetParams := gocloak.GetComponentsParams{
		Name:         userFederation.Name,
		ProviderType: userFederation.ProviderType,
		ParentID:     userFederation.ParentID,
	}

	if comps, getErr := tkc.client.GetComponentsWithParams(ctx, tkc.token.AccessToken, tkc.realm, ldapGetParams); getErr != nil {
		return getErr
	} else {
		// means user federation not found for given realm
		if len(comps) == 0 {
			if idUserFederation, createErr := tkc.client.CreateComponent(ctx, tkc.token.AccessToken, tkc.realm, userFederation); createErr != nil {
				return createErr
			} else {
				// do full sync after creating new one
				if syncErr := tkc.syncUserFederation(ctx, idUserFederation, true); syncErr != nil {
					syncErr = fmt.Errorf("legacy user federation '%s' created (%s), but sync failed", *userFederation.Name, idUserFederation)
					return syncErr
				}
			}
		} else {
			// set ID of user federation for update exactly existing user federation
			userFederation.ID = comps[0].ID
			if updateErr := tkc.client.UpdateComponent(ctx, tkc.token.AccessToken, tkc.baseConfig.Realm, userFederation); updateErr != nil {
				return updateErr
			} else {
				// do change sync only after update exiting one
				if syncErr := tkc.syncUserFederation(ctx, *userFederation.ID, false); syncErr != nil {
					syncErr = fmt.Errorf("legacy user federation '%s' updated (%s), but sync failed", *userFederation.Name, *userFederation.ID)
					return syncErr
				}
			}
		}
	}

	return nil
}

func (tkc *Tkc) syncUserFederation(ctx context.Context, idUserFederation string, fullSync bool) error {
	var url string

	url = tkc.baseConfig.Url + "/admin/realms/" + tkc.realm + "/user-storage/" + idUserFederation + "/sync"

	if fullSync {
		url += "?action=triggerFullSync"
	} else {
		url += "?action=triggerChangedUsersSync"
	}

	if response, postErr := tkc.client.RestyClient().NewRequest().SetAuthToken(tkc.token.AccessToken).Post(url); postErr != nil {
		return postErr
	} else {
		if response.StatusCode() != 200 {
			postErr = fmt.Errorf("got status code '%d' with response body '%s'", response.StatusCode(), response.String())
			return postErr
		} 
	}

	return nil
}
}
```

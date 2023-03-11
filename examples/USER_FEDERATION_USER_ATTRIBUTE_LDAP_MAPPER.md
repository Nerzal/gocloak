```go
type Tkc struct {
	realm  string
	token  *gocloak.JWT
	client *gocloak.GoCloak
}

// get all ldap mapper for given user federation
func (tk *Tkc) getLdapMapperOfUserFederation(ctx context.Context, idUserFederation string) (mappers map[string]gocloak.Component, getErr error) {
	var mapperList []*gocloak.Component
	mappers = make(map[string]gocloak.Component)

	// get only typed components for given id
	mapperGetParameter := gocloak.GetComponentsParams{
		ProviderType: gocloak.StringP("org.keycloak.storage.ldap.mappers.LDAPStorageMapper"),
		ParentID:     gocloak.StringP(idUserFederation),
	}

	if mapperList, getErr = tk.client.GetComponentsWithParams(ctx, tk.token.AccessToken, tk.realm, mapperGetParameter); getErr != nil {
		return mappers, getErr
	}

	// transform to map with mapper name as map key
	for _, mapper := range mapperList {
		mappers[*mapper.Name] = *mapper
	}

	return mappers, nil
}

// create given user ldap attribute mappers for given user federation
func (tk *Tkc) createUserModelAttributeMapper(ctx context.Context, idUserFederation string, mappers map[string]string) (createErr error) {
	var existingMappers map[string]gocloak.Component
	var mapper gocloak.Component
	var idMapper string

	if existingMappers, createErr = tk.getLdapMapperOfUserFederation(ctx, idUserFederation); createErr != nil {
		return createErr
	}

	for userModelAttribute, ldapAttribute := range mappers {
		mapper = prepareUserModelAttributeMapper(userModelAttribute, ldapAttribute, idUserFederation)
		if existingMapper, exists := existingMappers[userModelAttribute]; exists {
			mapper.ID = existingMapper.ID
			if createErr = tk.client.UpdateComponent(ctx, tk.token.AccessToken, tk.realm, mapper); createErr != nil {
				return createErr
			}
			log.Infof("user attribute ldap mapper '%s' updated (%s)", *mapper.Name, *mapper.ID)
		} else {
			if idMapper, createErr = tk.client.CreateComponent(ctx, tk.token.AccessToken, tk.realm, mapper); createErr != nil {
				return createErr
			}
			log.Infof("user attribute ldap mapper '%s' created (%s)", *mapper.Name, idMapper)
		}
	}

	return nil
}


// prepare component object of type 'org.keycloak.storage.ldap.mappers.LDAPStorageMapper' and ProviderId 'user-attribute-ldap-mapper' with config
func prepareUserModelAttributeMapper(userModelAttribute, ldapAttribute, idUserFederation string) (mapper gocloak.Component) {
	mapperConfig := make(map[string][]string)

	mapperConfig["ldap.attribute"] = []string{ldapAttribute}
	mapperConfig["user.model.attribute"] = []string{userModelAttribute}
	mapperConfig["is.mandatory.in.ldap"] = []string{"false"}
	mapperConfig["always.read.value.from.ldap"] = []string{"false"}
	mapperConfig["read.only"] = []string{"true"}

	mapper = gocloak.Component{
		Name:            gocloak.StringP(userModelAttribute),
		ProviderID:      gocloak.StringP("user-attribute-ldap-mapper"),
		ProviderType:    gocloak.StringP("org.keycloak.storage.ldap.mappers.LDAPStorageMapper"),
		ParentID:        gocloak.StringP(idUserFederation),
		ComponentConfig: &mapperConfig,
	}

	return mapper
}

// From last example: But with Mappers
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

	// sync options
	userFederationConfig["fullSyncPeriod"] = []string{"-1"}
	userFederationConfig["changedSyncPeriod"] = []string{"300"}
	userFederationConfig["batchSizeForSync"] = []string{"1000"}
	userFederationConfig["importUsers"] = []string{"true"}

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

	// ldap mappers for legacy user federation
	userModelAttributeMappers := make(map[string]string)
	userModelAttributeMappers["office"] = "physicalDeliveryOfficeName"
	userModelAttributeMappers["department"] = "destinationIndicator"
	userModelAttributeMappers["address"] = "postalAddress"
	userModelAttributeMappers["telephone"] = "telephoneNumber"
	userModelAttributeMappers["fax"] = "facsimileTelephoneNumber"
	userModelAttributeMappers["title"] = "title"
	userModelAttributeMappers["firstName"] = "cn"

	// ldap group mapper name
	groupMapper := "group"

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
				// do create federation user attribute ldap mappers
				if createErr := tkc.createUserModelAttributeMapper(ctx, idUserFederation, userModelAttributeMappers); createErr != nil {
					return createErr
				}
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
```
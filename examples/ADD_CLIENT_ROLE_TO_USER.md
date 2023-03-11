```go
type Tkc struct {
	realm  string
	token  *gocloak.JWT
	client *gocloak.GoCloak
}

func (tkc *Tkc) getClientRolesByList(ctx context.Context, idClient string, roles []string) (clientRoles []gocloak.Role, getErr error) {
	var notFoundRoles []string

	if roleObjects, tmpErr := tkc.client.GetClientRoles(ctx, tkc.token.AccessToken, tkc.realm, idClient, gocloak.GetRoleParams{}); tmpErr != nil {
		getErr = fmt.Errorf("failed to get roles for client (error: %s)", tmpErr.Error())

		return nil, getErr
	} else {
	searchRole:
		for _, r := range roles {
			for _, rb := range roleObjects {
				if r == *rb.Name {
					clientRoles = append(clientRoles, *rb)
					continue searchRole
				}
			}
			notFoundRoles = append(notFoundRoles, r)
		}
	}

	if len(notFoundRoles) > 0 {
		getErr = fmt.Errorf("failed to found role(s) '%s' for client", strings.Join(notFoundRoles, ", "))
	}

	return clientRoles, getErr
}

func (tkc *Tkc) getClients(ctx context.Context) (clients map[string]gocloak.Client, getErr error) {
	var clientList []*gocloak.Client

	// init map
	clients = make(map[string]gocloak.Client)

	// get all clients of realm
	if clientList, getErr = tkc.client.GetClients(ctx, tkc.token.AccessToken, tkc.realm, gocloak.GetClientsParams{}); getErr != nil {
		getErr = fmt.Errorf("get clients of realm failed (error: %s)", getErr.Error())
		return clients, getErr
	}

	// transform to map with clientID as map key
	for _, c := range clientList {
		clients[*c.ClientID] = *c
	}

	return clients, nil
}

func (tkc *Tkc) addClientRolesToUser(idUser string, clientRoles map[string][]string) (addErr error) {

	var clients map[string]gocloak.Client
	var userMappedClientRoles *gocloak.MappingsRepresentation
	var newRoles []string
	var clientRolesToAdd []gocloak.Role
	// ctx for current try
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	// get mapped roles of user to check if some role mus be added
	if userMappedClientRoles, addErr = tkc.client.GetRoleMappingByUserID(ctx, tkc.token.AccessToken, tkc.realm, idUser); addErr != nil {
		cancel()
	}

	// get clients to check if needed clients already exist
	if clients, addErr = tkc.getClients(ctx); addErr != nil {
		cancel()
	}

	// loop through given client role combination
	for client, roles := range clientRoles {
		// check client exist
		if _, exist := clients[client]; !exist {
			addErr = fmt.Errorf("client '%s' does not exist", client)
			break
		}

		// check if given role must be added
		if mappedClientRoles, exist := userMappedClientRoles.ClientMappings[client]; exist {
		SearchForMappedRoles:
			for _, role := range roles {
				for _, mappedClientRole := range *mappedClientRoles.Mappings {
					if role == *mappedClientRole.Name {
						// when role already mapped, continue with next role
						continue SearchForMappedRoles
					}
				}
				newRoles = append(newRoles, role)
			}
		} else {
			newRoles = roles
		}

		// add new roles otherwise do nothing
		if len(newRoles) > 0 {
			// get roles of client which should be added
			if clientRolesToAdd, addErr = tkc.getClientRolesByList(ctx, *clients[client].ID, roles); addErr != nil {
				break
			}

			// add roles to user
			if addErr = tkc.client.AddClientRolesToUser(ctx, tkc.token.AccessToken, tkc.realm, *clients[client].ID, idUser, clientRolesToAdd); addErr != nil {
				break
			}
		}
	}

	// cancel ctx
	cancel()

	return nil
}
```
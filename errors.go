package gocloak

// ObjectAllreadyExists is used when keycloak answers with 409
type ObjectAllreadyExists struct{}

func (o *ObjectAllreadyExists) Error() string {
	return "Conflict: Object allready exists"
}

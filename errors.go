package gocloak

// ObjectAlreadyExists is used when keycloak answers with 409
type ObjectAlreadyExists struct{}

func (o *ObjectAlreadyExists) Error() string {
	return "Conflict: Object already exists"
}

package gocloak

// ObjectAlreadyExists is used when keycloak answers with 409
type ObjectAlreadyExists struct {
	ErrorMessage string
}

func (e *ObjectAlreadyExists) Error() string {
	return e.ErrorMessage
}

// IsObjectAlreadyExists is a helper to verify tht the err is ObjectAlreadyExists
func IsObjectAlreadyExists(err error) bool {
	_, ok := err.(*ObjectAlreadyExists)
	return ok
}

// HTTPErrorResponse is a model of an error response
type HTTPErrorResponse struct {
	ErrorMessage string `json:"errorMessage,omitempty"`
}

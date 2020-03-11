package gocloak

// HTTPErrorResponse is a model of an error response
type HTTPErrorResponse struct {
	ErrorMessage string `json:"error_description,omitempty"`
	Error        string `json:"error,omitempty"`
}

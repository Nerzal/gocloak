package gocloak

import (
	"strings"
)

// HTTPErrorResponse is a model of an error response
type HTTPErrorResponse struct {
	Error       string `json:"error,omitempty"`
	Message     string `json:"errorMessage,omitempty"`
	Description string `json:"error_description,omitempty"`
}

// String joins the non-empty Error, Message and Description fields of the response with ": ".
// Output: [string] error text.
func (e HTTPErrorResponse) String() string {
	var res strings.Builder
	if len(e.Error) > 0 {
		res.WriteString(e.Error)
	}
	if len(e.Message) > 0 {
		if res.Len() > 0 {
			res.WriteString(": ")
		}
		res.WriteString(e.Message)
	}
	if len(e.Description) > 0 {
		if res.Len() > 0 {
			res.WriteString(": ")
		}
		res.WriteString(e.Description)
	}
	return res.String()
}

// NotEmpty reports whether any of Error, Message or Description is set.
// Output: [bool].
func (e HTTPErrorResponse) NotEmpty() bool {
	return len(e.Error) > 0 || len(e.Message) > 0 || len(e.Description) > 0
}

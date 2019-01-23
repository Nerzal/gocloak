package jwx

// DecodedAccessTokenHeader is the decoded header from the access token
type DecodedAccessTokenHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
	Kid string `json:"kid"`
}

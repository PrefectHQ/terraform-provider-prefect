package api

// CSRFTokenResponse represents the JSON response from the /csrf-token endpoint.
type CSRFTokenResponse struct {
	Token string `json:"token"`
}

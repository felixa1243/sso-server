package dto

type RegisterClientRequest struct {
	Name         string `json:"name" validate:"required"`
	RedirectURIs string `json:"redirect_uris" validate:"required"` // Comma-separated
	Scopes       string `json:"scopes" validate:"required"`
}

type ClientAppResponse struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	ClientID     string `json:"client_id"`
	RedirectURIs string `json:"redirect_uris"`
	Scopes       string `json:"scopes"`
	CreatedAt    string `json:"created_at"`
}

type ClientAppSecretResponse struct {
	ClientAppResponse
	ClientSecret string `json:"client_secret"`
}

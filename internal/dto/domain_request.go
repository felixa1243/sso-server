package dto

type RegisterDomainRequest struct {
	Name string `json:"name" validate:"required"`
	URL  string `json:"url" validate:"required"`
}

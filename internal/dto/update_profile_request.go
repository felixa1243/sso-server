package dto

type UpdateProfileRequest struct {
	FullName string `json:"full_name" validate:"required"`
}

type UpdateAvatarRequest struct {
	AvatarURI string `json:"avatar_uri" validate:"required"`
}

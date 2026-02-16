package dto

import "sso-server/internal/models"

type JoinUser struct {
	Fullname  string        `json:"fullname"`
	Email     string        `json:"email"`
	Role      []models.Role `json:"role"`
	AvatarURI string        `json:"avatar_uri"`
}

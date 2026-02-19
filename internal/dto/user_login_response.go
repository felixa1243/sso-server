package dto

type UserLoginResponse struct {
	Fullname  string `json:"fullname"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	AvatarURI string `json:"avatar_uri"`
}

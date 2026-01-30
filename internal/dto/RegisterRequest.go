package dto

type RegisterRequest struct {
	FullName        string `json:"fullname" form:"fullname" validate:"required,min=3,max=50"`
	Email           string `json:"email" form:"email" validate:"required,email"`
	Password        string `json:"password" form:"password" validate:"required,min=8,strong_password"`
	PasswordConfirm string `json:"password_confirm" form:"password_confirm" validate:"required,min=8,strong_password"`
}

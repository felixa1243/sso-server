package dto

type ChangePasswordRequest struct {
	OldPassword        string `json:"old_password" validate:"required"`
	NewPassword        string `json:"new_password" validate:"required"`
	NewPasswordConfirm string `json:"new_password_confirm" validate:"required"`
}

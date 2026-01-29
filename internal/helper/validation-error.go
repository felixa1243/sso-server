package helper

import "github.com/go-playground/validator/v10"

type ErrorResponse struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func GetCustomMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email format"
	case "min":
		return "Too short (minimum " + fe.Param() + " characters)"
	case "max":
		return "Too long (maximum " + fe.Param() + " characters)"
	}
	return "Invalid value"
}

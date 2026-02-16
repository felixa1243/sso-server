package helper

import (
	"unicode"

	"github.com/go-playground/validator/v10"
)

func IsStrongPassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	var (
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}
	return hasUpper && hasLower && hasNumber && hasSpecial
}

var validate *validator.Validate

func init() {
	validate = validator.New()
	validate.RegisterValidation("strong_password", IsStrongPassword)
}

func ValidateStruct(req interface{}) map[string]string {
	err := validate.Struct(req)
	if err == nil {
		return nil
	}

	errors := make(map[string]string)
	for _, fe := range err.(validator.ValidationErrors) {
		errors[fe.Field()] = GetCustomMessage(fe)
	}
	return errors
}

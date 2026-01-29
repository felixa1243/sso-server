package controllers

import (
	"crypto/rsa"
	"sso-server/internal/database"
	"sso-server/internal/dto"
	"sso-server/internal/helper"
	"sso-server/internal/models"
	"unicode"

	"github.com/go-playground/validator/v10" // Import validator
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func isStrongPassword(fl validator.FieldLevel) bool {
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

	// You can adjust these requirements as needed
	return hasUpper && hasLower && hasNumber && hasSpecial
}

var validate *validator.Validate

func init() {
	validate = validator.New()
	validate.RegisterValidation("strong_password", isStrongPassword)
}

type AuthController struct {
	DB         *gorm.DB
	PrivateKey *rsa.PrivateKey
}

func validateStruct(req interface{}) map[string]string {
	err := validate.Struct(req)
	if err == nil {
		return nil
	}

	errors := make(map[string]string)
	for _, fe := range err.(validator.ValidationErrors) {
		errors[fe.Field()] = helper.GetCustomMessage(fe)
	}
	return errors
}
func (ac *AuthController) createUser(c *fiber.Ctx, roleName string) (*models.User, interface{}, error) {
	var req dto.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return nil, err, nil
	}
	if errs := validateStruct(req); errs != nil {
		return nil, errs, nil
	}

	var role models.Role
	if err := ac.DB.Where("name = ?", roleName).First(&role).Error; err != nil {
		return nil, nil, err
	}

	user := models.User{
		Email:        req.Email,
		ID:           uuid.New(),
		PasswordHash: helper.GeneratePassword(req.Password),
		RoleID:       role.ID,
	}

	if err := database.New().GetDB().Create(&user).Error; err != nil {
		return nil, nil, err
	}

	return &user, nil, nil
}

func (ac *AuthController) ReaderRegister(c *fiber.Ctx) error {
	user, valErrors, err := ac.createUser(c, "Reader")
	if valErrors != nil {
		return c.Status(400).JSON(fiber.Map{
			"message": "validation error",
			"errors":  valErrors,
		})
	}
	if err != nil {
		return nil
	}
	return c.Status(201).JSON(ac.mapUser(*user))
}

func (ac *AuthController) EditorRegister(c *fiber.Ctx) error {
	user, valErrors, err := ac.createUser(c, "Editor")
	if valErrors != nil {
		return c.Status(400).JSON(fiber.Map{
			"message": "validation error",
			"errors":  valErrors,
		})
	}
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": err.Error()})
	}
	return c.Status(201).JSON(ac.mapUser(*user))
}
func (ac *AuthController) mapUser(user models.User) fiber.Map {
	var role models.Role
	ac.DB.Where("id = ?", user.RoleID).First(&role)
	return fiber.Map{
		"id":    user.ID,
		"email": user.Email,
		"role":  role.Name,
	}
}
func (ac *AuthController) Login(c *fiber.Ctx) error {
	var req dto.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": err.Error()})
	}
	if valErrors := validateStruct(req); valErrors != nil {
		return c.Status(400).JSON(fiber.Map{
			"message": "validation error",
			"errors":  valErrors,
		})
	}

	var user models.User

	res := ac.DB.Preload("Role").Where("email = ?", req.Email).First(&user)

	if res.Error != nil {
		return c.Status(400).JSON(fiber.Map{"message": "user not found"})
	}
	if !helper.ComparePassword(user.PasswordHash, req.Password) {
		return c.Status(400).JSON(fiber.Map{"message": "incorrect password"})
	}
	token, err := helper.GenerateToken(user, ac.PrivateKey)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "failed to generate secure token"})
	}

	return c.JSON(fiber.Map{
		"message": "user logged in successfully",
		"token":   token,
		"user": fiber.Map{
			"email": user.Email,
			"role":  user.Role.Name,
		},
	})
}

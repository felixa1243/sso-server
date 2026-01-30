package controllers

import (
	"crypto/rsa"
	"sso-server/internal/database"
	"sso-server/internal/dto"
	"sso-server/internal/helper"
	"sso-server/internal/models"
	"time"
	"unicode"

	"github.com/go-playground/validator/v10" // Import validator
	"github.com/go-redis/redis/v8"
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
	Redis      *redis.Client
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
	user, valErrors, err := ac.createUser(c, "Blog:Reader")
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
	user, valErrors, err := ac.createUser(c, "Blog:Editor")
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

	redirectURL := c.Query("redirect_url")

	var user models.User
	res := ac.DB.Preload("Role").Where("email = ?", req.Email).First(&user)

	if res.Error != nil {
		return c.Status(400).JSON(fiber.Map{"message": "user not found"})
	}
	if !helper.ComparePassword(user.PasswordHash, req.Password) {
		return c.Status(400).JSON(fiber.Map{"message": "incorrect password"})
	}
	authCode := uuid.New().String()
	err := ac.Redis.Set(c.Context(), "auth_code:"+authCode, user.ID.String(), 5*time.Minute).Err()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "failed to store session"})
	}

	// 3. Redirect back to Next.js Callback with the CODE
	return c.Redirect(redirectURL + "?code=" + authCode)
}
func (ac *AuthController) ExchangeCode(c *fiber.Ctx) error {
	var req struct {
		Code string `json:"code"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}
	userID, err := ac.Redis.Get(c.Context(), "auth_code:"+req.Code).Result()
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "code expired or invalid"})
	}
	ac.Redis.Del(c.Context(), "auth_code:"+req.Code)
	var user models.User
	if err := ac.DB.Preload("Role").First(&user, "id = ?", userID).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "user not found"})
	}

	token, err := helper.GenerateToken(user, ac.PrivateKey)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "token generation failed"})
	}

	return c.JSON(fiber.Map{
		"token": token,
		"user": fiber.Map{
			"email": user.Email,
			"role":  user.Role.Name,
		},
	})
}
func (ac *AuthController) ShowRegister(c *fiber.Ctx) error {
	return c.Render("register", fiber.Map{})
}

func (ac *AuthController) ShowLogin(c *fiber.Ctx) error {
	return c.Render("login", fiber.Map{
		"RedirectURL": c.Query("redirect_url"),
	})
}

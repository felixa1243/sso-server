package services

import (
	"sso-server/internal/database"
	"sso-server/internal/dto"
	"sso-server/internal/helper"
	"sso-server/internal/models"
	"sso-server/internal/repositories"
	"time"
	"unicode"

	"github.com/go-playground/validator/v10"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type UserServices struct {
	userRepository repositories.UserRepository
	Redis          *redis.Client
}

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
	return hasUpper && hasLower && hasNumber && hasSpecial
}

var validate *validator.Validate

func init() {
	validate = validator.New()
	validate.RegisterValidation("strong_password", isStrongPassword)
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
func NewUserServices(userRepository repositories.UserRepository, redis *redis.Client) *UserServices {
	return &UserServices{userRepository: userRepository, Redis: redis}
}

func (u *UserServices) CreateUser(c *fiber.Ctx, roleName string) (*models.User, any, error) {
	var req dto.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return nil, err, nil
	}
	if errs := validateStruct(req); errs != nil {
		return nil, errs, nil
	}
	var role models.Role
	if req.Password != req.PasswordConfirm {
		return nil, nil, fiber.ErrBadRequest
	}
	if err := database.New().GetDB().Where("name = ?", roleName).First(&role).Error; err != nil {
		return nil, nil, err
	}
	userExists := database.New().GetDB().Where("email = ?", req.Email).First(&models.User{})
	if userExists.RowsAffected > 0 {
		return nil, nil, fiber.ErrBadRequest
	}
	UserCreated := models.User{
		ID:           uuid.New(),
		Email:        req.Email,
		PasswordHash: helper.GeneratePassword(req.Password),
		RoleID:       role.ID,
	}
	err := u.userRepository.Create(c.UserContext(), &UserCreated)
	if err != nil {
		return nil, nil, err
	}

	return &UserCreated, nil, nil
}
func (u *UserServices) Login(c *fiber.Ctx) error {
	var req dto.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": err.Error()})
	}

	redirectURL := c.Query("redirect_url")
	var user models.User
	res := database.New().GetDB().Preload("Role").Where("email = ?", req.Email).First(&user)

	if res.Error != nil {
		return c.Status(400).JSON(fiber.Map{"message": "user not found"})
	}
	if !helper.ComparePassword(user.PasswordHash, req.Password) {
		return c.Status(400).JSON(fiber.Map{"message": "incorrect password"})
	}
	authCode := uuid.New().String()
	err := u.Redis.Set(c.Context(), "auth_code:"+authCode, user.ID.String(), 5*time.Minute).Err()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "failed to store session"})
	}

	// 3. Redirect back to Next.js Callback with the CODE
	return c.Redirect(redirectURL + "?code=" + authCode)
}

func (u *UserServices) Logout(c *fiber.Ctx) error {

	token := c.Cookies("access_token")

	if token != "" {
		err := u.Redis.Del(c.Context(), "access_token:"+token).Err()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Failed to invalidate session in cache",
			})
		}
	}
	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
		SameSite: "Lax",
	})

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Logged out successfully",
	})
}

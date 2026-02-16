package services

import (
	"crypto/rsa"
	"encoding/base64"
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
	userRepository    repositories.UserRepository
	profileRepository repositories.ProfileRepository
	Redis             *redis.Client
	PrivateKey        *rsa.PrivateKey
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
func NewUserServices(userRepository repositories.UserRepository, profileRepository repositories.ProfileRepository, redis *redis.Client, privateKey *rsa.PrivateKey) *UserServices {
	return &UserServices{userRepository: userRepository, profileRepository: profileRepository, Redis: redis, PrivateKey: privateKey}
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
		Role:         []models.Role{role},
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
	userMap, errMap := u.GetUser(user.ID.String())
	if errMap != nil {
		return errMap
	}
	if redirectURL == "" {
		token, err := u.GetToken(user.ID.String(), c)
		if err != nil {
			return err
		}
		return c.JSON(fiber.Map{
			"access_token": token,
			"user":         userMap,
		})
	}
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
func (u *UserServices) ChangePassword(c *fiber.Ctx) error {
	var req dto.ChangePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": err.Error()})
	}
	user, err := helper.GetUserFromContext(c)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": err.Error()})
	}
	var userID string
	userID = base64.StdEncoding.EncodeToString(user.ID[:])
	if !helper.ComparePassword(user.PasswordHash, req.OldPassword) {
		return c.Status(400).JSON(fiber.Map{"message": "incorrect password"})
	}
	if req.NewPassword != req.NewPasswordConfirm {
		return c.Status(400).JSON(fiber.Map{"message": "passwords do not match"})
	}
	user.PasswordHash = helper.GeneratePassword(req.NewPassword)
	err = u.userRepository.Update(userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": err.Error()})
	}
	return c.Status(200).JSON(fiber.Map{"message": "password changed successfully"})
}
func (u *UserServices) GetUser(uid string) (*dto.JoinUser, error) {
	joinUser, queryError := u.profileRepository.FindByUserID(uid)
	if queryError != nil {
		return nil, fiber.NewError(fiber.StatusNotFound, "Profile record not found")
	}
	return &joinUser, nil
}
func (u *UserServices) GetToken(userID string, ctx *fiber.Ctx) (string, error) {
	joinUser, queryError := u.GetUser(userID)
	if queryError != nil {
		return "", fiber.NewError(fiber.StatusNotFound, "Profile record not found")
	}
	user, err := u.userRepository.FindByID(userID)
	if err != nil {
		return "", fiber.NewError(fiber.StatusUnauthorized, "User account is not exists")
	}
	fullname := joinUser.Fullname
	token, err := helper.GenerateToken(user, fullname, u.PrivateKey)
	if err != nil {
		return "", fiber.NewError(fiber.StatusInternalServerError, "Security signing failed")
	}
	err = u.Redis.Set(ctx.Context(), "access_token:"+token, user.ID.String(), 24*time.Hour).Err()
	if err != nil {
		return "", fiber.NewError(fiber.StatusServiceUnavailable, "Session Storage Failed")
	}
	return token, nil
}

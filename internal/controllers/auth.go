package controllers

import (
	"crypto/rsa"
	"os"
	"sso-server/internal/helper"
	"sso-server/internal/models"
	"sso-server/internal/repositories"
	"sso-server/internal/services"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type AuthController struct {
	DB          *gorm.DB
	PrivateKey  *rsa.PrivateKey
	Redis       *redis.Client
	UserService *services.UserServices
}

func NewAuthController(db *gorm.DB, privateKey *rsa.PrivateKey, redis *redis.Client) *AuthController {
	return &AuthController{
		DB:          db,
		PrivateKey:  privateKey,
		Redis:       redis,
		UserService: services.NewUserServices(repositories.NewUserRepository(db), redis),
	}
}
func (ac *AuthController) createUser(c *fiber.Ctx, roleName string) (*models.User, interface{}, error) {
	return ac.UserService.CreateUser(c, roleName)
}

func (ac *AuthController) ReaderRegister(c *fiber.Ctx) error {
	_, valErrors, err := ac.createUser(c, "Blog:Reader")
	if valErrors != nil {
		return c.Status(400).JSON(fiber.Map{
			"message": "validation error",
			"errors":  valErrors,
		})
	}
	if err != nil {
		return nil
	}
	return c.Redirect("/login")
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
	return ac.UserService.Login(c)
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
		return c.Status(404).JSON(fiber.Map{"error": "user not found"})
	}
	var userProfie models.UserProfile
	ac.DB.Where("user_id = ?", user.ID).First(&userProfie)
	fullname := userProfie.FullName
	token, err := helper.GenerateToken(user, fullname, ac.PrivateKey)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "token generation failed"})
	}
	err = ac.Redis.Set(c.Context(), "access_token:"+token, user.ID.String(), 24*time.Hour).Err()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to store session"})
	}
	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    token,
		Expires:  time.Now().Add(24 * time.Hour),
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Lax",
	})
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
		"AppUrl":      os.Getenv("APP_URL"),
	})
}
func (ac *AuthController) Logout(c *fiber.Ctx) error {
	return ac.UserService.Logout(c)
}

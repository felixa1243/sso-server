package controllers

import (
	"crypto/rsa"
	"os"
	"sso-server/internal/models"
	"sso-server/internal/services"
	"strings"
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

func NewAuthController(db *gorm.DB, privateKey *rsa.PrivateKey, redis *redis.Client, userService *services.UserServices) *AuthController {
	return &AuthController{
		DB:          db,
		PrivateKey:  privateKey,
		Redis:       redis,
		UserService: userService,
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
	rolesString := ""
	for _, role := range user.Role {
		rolesString += role.Name
	}
	return fiber.Map{
		"id":    user.ID,
		"email": user.Email,
		"role":  rolesString,
	}
}
func (ac *AuthController) Login(c *fiber.Ctx) error {
	return ac.UserService.Login(c)
}
func (ac *AuthController) ExchangeCode(c *fiber.Ctx) error {
	var req struct {
		Code string `json:"code"`
	}

	// Parse request body first
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}

	// Get userID from Redis
	userID, err := ac.Redis.Get(c.Context(), "auth_code:"+req.Code).Result()
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "code expired or invalid"})
	}

	// Fetch user by ID
	var user models.User
	result := ac.DB.Preload("Role").Where("id = ?", userID).First(&user)
	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{"error": "user not found"})
	}

	// Generate token
	token, err := ac.UserService.GetToken(userID, c)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Delete code after successful exchange
	ac.Redis.Del(c.Context(), "auth_code:"+req.Code)

	// Set cookie
	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    token,
		Expires:  time.Now().Add(24 * time.Hour),
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Lax",
	})

	// Build role string
	var roleString strings.Builder
	for _, role := range user.Role {
		roleString.WriteString(role.Name)
	}

	return c.JSON(fiber.Map{
		"token": token,
		"user": fiber.Map{
			"email": user.Email,
			"role":  roleString.String(),
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

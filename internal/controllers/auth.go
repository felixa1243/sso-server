package controllers

import (
	"crypto/rsa"
	"sso-server/internal/database"
	"sso-server/internal/dto"
	"sso-server/internal/helper"
	"sso-server/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuthController struct {
	DB         *gorm.DB
	PrivateKey *rsa.PrivateKey
}

func createUser(c *fiber.Ctx, dto dto.RegisterRequest, roleName string) (*models.User, error) {
	if err := c.BodyParser(&dto); err != nil {
		return &models.User{}, err
	}
	var role models.Role
	if err := database.New().GetDB().Where("name = ?", roleName).First(&role).Error; err != nil {
		return &models.User{}, err
	}
	user := models.User{
		Email:        dto.Email,
		ID:           uuid.New(),
		PasswordHash: helper.GeneratePassword(dto.Password),
		RoleID:       role.ID,
	}
	if err := database.New().GetDB().Create(&user).Error; err != nil {
		return &models.User{}, err
	}

	return &user, nil
}
func (ac *AuthController) ReaderRegister(c *fiber.Ctx) error {
	var req dto.RegisterRequest
	user, err := createUser(c, req, "Reader")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": err.Error()})
	}
	return c.Status(201).JSON(fiber.Map{
		"message": "user created",
		"user":    user,
	})
}
func (ac *AuthController) EditorRegister(c *fiber.Ctx) error {
	var req dto.RegisterRequest
	user, err := createUser(c, req, "Editor")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": err.Error()})
	}
	return c.Status(201).JSON(fiber.Map{
		"message": "user created",
		"user":    user,
	})
}
func (ac *AuthController) Login(c *fiber.Ctx) error {
	var req dto.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": err.Error()})
	}

	var user models.User
	// Use the DB from the struct receiver
	res := ac.DB.Preload("Role").Where("email = ?", req.Email).First(&user)

	if res.Error != nil {
		return c.Status(400).JSON(fiber.Map{"message": "user not found"})
	}
	if !helper.ComparePassword(user.PasswordHash, req.Password) {
		return c.Status(400).JSON(fiber.Map{"message": "incorrect password"})
	}

	// FIX: Access PrivateKey directly from the struct receiver (ac)
	// No more c.Locals lookup needed!
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

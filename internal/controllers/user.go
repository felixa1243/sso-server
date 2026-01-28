package controllers

import (
	"crypto/rsa"
	"sso-server/internal/dto"
	"sso-server/internal/helper"
	"sso-server/internal/models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type UserController struct {
	DB         *gorm.DB
	PrivateKey *rsa.PrivateKey
}

func (u *UserController) SetBasicProfile(c *fiber.Ctx) error {
	var req dto.ProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": err.Error()})
	}
	user, err := helper.GetUserFromContext(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"message": "Unauthorized: " + err.Error()})
	}

	profile := models.UserProfile{
		UserID:   user.ID,
		FullName: req.FullName,
	}
	if err := u.DB.Create(&profile).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "could not create profile: " + err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "profile updated",
		"profile": profile,
	})
}

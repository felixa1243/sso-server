package controllers

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type UserProfileController interface {
	ChangePassword(c *fiber.Ctx) error
	ChangeProfilePicture(c *fiber.Ctx) error
	ChangeUserProfile(c *fiber.Ctx) error
}

type UserProfileControllerImpl struct {
	db *gorm.DB
}

func (u *UserProfileControllerImpl) ChangePassword(c *fiber.Ctx) error {
	return nil
}

func (u *UserProfileControllerImpl) ChangeProfilePicture(c *fiber.Ctx) error {
	return nil
}

func (u *UserProfileControllerImpl) ChangeUserProfile(c *fiber.Ctx) error {
	return nil
}

func NewUserProfileController(db *gorm.DB) UserProfileController {
	return &UserProfileControllerImpl{db: db}
}

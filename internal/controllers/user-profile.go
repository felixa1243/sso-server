package controllers

import (
	"sso-server/internal/dto"
	"sso-server/internal/helper"
	"sso-server/internal/services"

	"github.com/gofiber/fiber/v2"
)

type UserProfileController interface {
	ChangePassword(c *fiber.Ctx) error
	ChangeProfilePicture(c *fiber.Ctx) error
	ChangeUserProfile(c *fiber.Ctx) error
}

type UserProfileControllerImpl struct {
	UserService services.UserService
}

func (u *UserProfileControllerImpl) ChangePassword(c *fiber.Ctx) error {
	var req dto.ChangePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": err.Error()})
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		return c.Status(400).JSON(fiber.Map{
			"message": "validation error",
			"errors":  errs,
		})
	}

	user, err := helper.GetUserFromContext(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"message": "Unauthorized"})
	}

	err = u.UserService.ChangePassword(c.Context(), user.ID.String(), req)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"message": err.Error()})
	}

	return c.Status(200).JSON(fiber.Map{"message": "password changed successfully"})
}

func (u *UserProfileControllerImpl) ChangeProfilePicture(c *fiber.Ctx) error {
	var req dto.UpdateAvatarRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": err.Error()})
	}
	if errs := helper.ValidateStruct(req); errs != nil {
		return c.Status(400).JSON(fiber.Map{
			"message": "validation error",
			"errors":  errs,
		})
	}

	user, err := helper.GetUserFromContext(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"message": "Unauthorized"})
	}

	if err := u.UserService.ChangeAvatar(user, req.AvatarURI); err != nil {
		return err
	}

	return c.Status(200).JSON(fiber.Map{"message": "avatar updated successfully", "avatar_uri": req.AvatarURI})
}

func (u *UserProfileControllerImpl) ChangeUserProfile(c *fiber.Ctx) error {
	var req dto.UpdateProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": err.Error()})
	}
	if errs := helper.ValidateStruct(req); errs != nil {
		return c.Status(400).JSON(fiber.Map{
			"message": "validation error",
			"errors":  errs,
		})
	}

	user, err := helper.GetUserFromContext(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"message": "Unauthorized"})
	}

	if err := u.UserService.UpdateFullName(user, req.FullName); err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "failed to update profile"})
	}

	return c.Status(200).JSON(fiber.Map{"message": "profile updated successfully", "full_name": req.FullName})
}

func NewUserProfileController(userService services.UserService) UserProfileController {
	return &UserProfileControllerImpl{UserService: userService}
}

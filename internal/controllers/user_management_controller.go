package controllers

import (
	"sso-server/internal/services"

	"github.com/gofiber/fiber/v2"
)

type UserManagementController struct {
	UserService services.UserService
}

func NewUserManagementController(userService services.UserService) *UserManagementController {
	return &UserManagementController{
		UserService: userService,
	}
}

func (c *UserManagementController) ListUsers(ctx *fiber.Ctx) error {
	users, err := c.UserService.ListUsers(ctx.Context())
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(users)
}

func (c *UserManagementController) DeleteUser(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if err := c.UserService.DeleteUser(ctx.Context(), id); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.SendStatus(fiber.StatusOK)
}

func (c *UserManagementController) UpdateUserRoles(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	var req struct {
		RoleIDs []string `json:"role_ids"`
	}
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := c.UserService.UpdateUserRoles(ctx.Context(), id, req.RoleIDs); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return ctx.SendStatus(fiber.StatusOK)
}

func (c *UserManagementController) BanUser(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if err := c.UserService.BanUser(ctx.Context(), id); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.SendStatus(fiber.StatusOK)
}

func (c *UserManagementController) UnbanUser(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if err := c.UserService.UnbanUser(ctx.Context(), id); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.SendStatus(fiber.StatusOK)
}

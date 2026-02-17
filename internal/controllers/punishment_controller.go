package controllers

import (
	"sso-server/internal/helper"
	"sso-server/internal/services"

	"github.com/gofiber/fiber/v2"
)

type PunishmentController struct {
	PunishmentService services.PunishmentService
}

func NewPunishmentController(punishmentService services.PunishmentService) *PunishmentController {
	return &PunishmentController{
		PunishmentService: punishmentService,
	}
}

func (c *PunishmentController) PunishUser(ctx *fiber.Ctx) error {
	var req services.PunishRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	admin, err := helper.GetUserFromContext(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	punishment, err := c.PunishmentService.PunishUser(ctx.Context(), admin.ID.String(), req)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return ctx.Status(fiber.StatusCreated).JSON(punishment)
}

func (c *PunishmentController) GetUserPunishments(ctx *fiber.Ctx) error {
	userID := ctx.Params("id")
	punishments, err := c.PunishmentService.GetPunishments(ctx.Context(), userID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(punishments)
}

func (c *PunishmentController) RevokePunishment(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if err := c.PunishmentService.RevokePunishment(ctx.Context(), id); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.SendStatus(fiber.StatusOK)
}

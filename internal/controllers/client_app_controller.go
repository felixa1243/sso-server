package controllers

import (
	"sso-server/internal/services"
	"sso-server/internal/helper"
	"github.com/gofiber/fiber/v2"
)

type ClientAppController struct {
	ClientService services.ClientAppService
}

func NewClientAppController(clientService services.ClientAppService) *ClientAppController {
	return &ClientAppController{
		ClientService: clientService,
	}
}

func (c *ClientAppController) CreateClient(ctx *fiber.Ctx) error {
	var req struct {
		Name         string `json:"name"`
		RedirectURIs string `json:"redirect_uris"`
		Scopes       string `json:"scopes"`
	}
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	user, err := helper.GetUserFromContext(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	client, secret, err := c.ClientService.RegisterClient(ctx.Context(), user.ID.String(), req.Name, req.RedirectURIs, req.Scopes)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":            client.ID,
		"name":          client.Name,
		"redirect_uris": client.RedirectURIs,
		"scopes":        client.AllowedScopes,
		"secret":        secret,
	})
}

func (c *ClientAppController) ListClients(ctx *fiber.Ctx) error {
	user, err := helper.GetUserFromContext(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	clients, err := c.ClientService.ListClients(ctx.Context(), user.ID.String())
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(clients)
}

func (c *ClientAppController) DeleteClient(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	user, err := helper.GetUserFromContext(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	if err := c.ClientService.DeleteClient(ctx.Context(), id, user.ID.String()); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.SendStatus(fiber.StatusOK)
}

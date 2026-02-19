package controllers

import (
	"sso-server/internal/dto"
	"sso-server/internal/helper"
	"sso-server/internal/services"

	"github.com/gofiber/fiber/v2"
)

type ClientAppController struct {
	Service services.ClientAppService
}

func NewClientAppController(service services.ClientAppService) *ClientAppController {
	return &ClientAppController{Service: service}
}

func (c *ClientAppController) Register(ctx *fiber.Ctx) error {
	var req dto.RegisterClientRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	user, err := helper.GetUserFromContext(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	app, err := c.Service.RegisterClient(ctx.Context(), user.ID.String(), req)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":            app.ID,
		"name":          app.Name,
		"client_id":     app.ClientID,
		"client_secret": app.ClientSecret,
		"redirect_uris": app.RedirectURIs,
		"scopes":        app.Scopes,
	})
}

func (c *ClientAppController) List(ctx *fiber.Ctx) error {
	user, err := helper.GetUserFromContext(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	apps, err := c.Service.ListClients(ctx.Context(), user.ID.String())
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	response := []fiber.Map{}
	for _, app := range apps {
		response = append(response, fiber.Map{
			"id":            app.ID,
			"name":          app.Name,
			"client_id":     app.ClientID,
			"redirect_uris": app.RedirectURIs,
			"scopes":        app.Scopes,
		})
	}

	return ctx.JSON(fiber.Map{
		"status": "success",
		"data":   response,
	})
}

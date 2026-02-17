package controllers

import (
	"sso-server/internal/services"

	"github.com/gofiber/fiber/v2"
)

type ScopeController struct {
	ScopeService services.ScopeService
}

func NewScopeController(scopeService services.ScopeService) *ScopeController {
	return &ScopeController{
		ScopeService: scopeService,
	}
}

func (c *ScopeController) ListScopes(ctx *fiber.Ctx) error {
	scopes, err := c.ScopeService.ListScopes(ctx.Context())
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(scopes)
}

func (c *ScopeController) CreateScope(ctx *fiber.Ctx) error {
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	scope, err := c.ScopeService.CreateScope(ctx.Context(), req.Name, req.Description)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.Status(fiber.StatusCreated).JSON(scope)
}

func (c *ScopeController) UpdateScope(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	scope, err := c.ScopeService.UpdateScope(ctx.Context(), id, req.Name, req.Description)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(scope)
}

func (c *ScopeController) DeleteScope(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if err := c.ScopeService.DeleteScope(ctx.Context(), id); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.SendStatus(fiber.StatusOK)
}

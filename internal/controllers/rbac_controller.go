package controllers

import (
	"sso-server/internal/services"

	"github.com/gofiber/fiber/v2"
)

type RBACController struct {
	RBACService services.RBACService
}

func NewRBACController(rbacService services.RBACService) *RBACController {
	return &RBACController{
		RBACService: rbacService,
	}
}

func (c *RBACController) CreateRole(ctx *fiber.Ctx) error {
	var req struct {
		Name     string  `json:"name"`
		DomainID *string `json:"domain_id"`
	}
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	role, err := c.RBACService.CreateRole(ctx.Context(), req.Name, req.DomainID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return ctx.Status(fiber.StatusCreated).JSON(role)
}

func (c *RBACController) ListRoles(ctx *fiber.Ctx) error {
	roles, err := c.RBACService.ListRoles(ctx.Context())
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(roles)
}

func (c *RBACController) UpdateRole(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	var req struct {
		Name     string  `json:"name"`
		DomainID *string `json:"domain_id"`
	}
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	role, err := c.RBACService.UpdateRole(ctx.Context(), id, req.Name, req.DomainID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(role)
}

func (c *RBACController) DeleteRole(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if err := c.RBACService.DeleteRole(ctx.Context(), id); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.SendStatus(fiber.StatusOK)
}

func (c *RBACController) AssignPermissions(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	var req struct {
		PermissionIDs []string `json:"permission_ids"`
	}
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := c.RBACService.AssignPermissions(ctx.Context(), id, req.PermissionIDs); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.SendStatus(fiber.StatusOK)
}

func (c *RBACController) CreatePermission(ctx *fiber.Ctx) error {
	var req struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	}
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	perm, err := c.RBACService.CreatePermission(ctx.Context(), req.Name, req.Slug)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.Status(fiber.StatusCreated).JSON(perm)
}

func (c *RBACController) ListPermissions(ctx *fiber.Ctx) error {
	perms, err := c.RBACService.ListPermissions(ctx.Context())
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.JSON(perms)
}

package controllers

import (
	"sso-server/internal/dto"
	"sso-server/internal/helper"
	"sso-server/internal/services"

	"github.com/gofiber/fiber/v2"
)

type DomainController struct {
	DomainService services.DomainService
}

func NewDomainController(domainService services.DomainService) *DomainController {
	return &DomainController{
		DomainService: domainService,
	}
}

func (c *DomainController) Register(ctx *fiber.Ctx) error {
	var req dto.RegisterDomainRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(400).JSON(fiber.Map{"message": err.Error()})
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		return ctx.Status(400).JSON(fiber.Map{
			"message": "validation error",
			"errors":  errs,
		})
	}

	user, err := helper.GetUserFromContext(ctx)
	if err != nil {
		return ctx.Status(401).JSON(fiber.Map{"message": "Unauthorized"})
	}

	domain, err := c.DomainService.CreateDomain(ctx.Context(), req.Name, req.URL, user.ID.String())
	if err != nil {
		return ctx.Status(400).JSON(fiber.Map{"message": err.Error()})
	}

	return ctx.Status(201).JSON(domain)
}

func (c *DomainController) List(ctx *fiber.Ctx) error {
	user, err := helper.GetUserFromContext(ctx)
	if err != nil {
		return ctx.Status(401).JSON(fiber.Map{"message": "Unauthorized"})
	}

	domains, err := c.DomainService.ListDomains(ctx.Context(), user.ID.String())
	if err != nil {
		return ctx.Status(500).JSON(fiber.Map{"message": err.Error()})
	}
	return ctx.Status(200).JSON(domains)
}

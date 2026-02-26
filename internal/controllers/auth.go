package controllers

import (
	"sso-server/internal/dto"
	"sso-server/internal/helper"
	"sso-server/internal/models"
	"sso-server/internal/services"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

type AuthController struct {
	UserService      services.UserService
	ClientAppService services.ClientAppService
}

func NewAuthController(userService services.UserService, clientAppService services.ClientAppService) *AuthController {
	return &AuthController{
		UserService:      userService,
		ClientAppService: clientAppService,
	}
}

func (ac *AuthController) Register(c *fiber.Ctx) error {
	var req dto.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": err.Error()})
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		return c.Status(400).JSON(fiber.Map{
			"message": "validation error",
			"errors":  errs,
		})
	}

	roleName := "User" // Default fallback

	user, err := ac.UserService.CreateUser(c.Context(), req, roleName)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"message": err.Error()})
	}
	if user == nil {
		return c.Status(500).JSON(fiber.Map{"message": "User creation failed unexpectedly"})
	}

	return c.Status(201).JSON(ac.mapUser(*user))
}

func (ac *AuthController) createUser(c *fiber.Ctx, roleName string) (*models.User, map[string]string, error) {
	var req dto.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return nil, nil, err
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		return nil, errs, nil
	}

	user, err := ac.UserService.CreateUser(c.Context(), req, roleName)
	if err != nil {
		return nil, nil, err
	}
	return user, nil, nil
}

func (ac *AuthController) mapUser(user models.User) fiber.Map {
	rolesString := ""
	for _, role := range user.Role {
		rolesString += role.Name
	}
	return fiber.Map{
		"id":    user.ID,
		"email": user.Email,
		"role":  rolesString,
	}
}

func (ac *AuthController) Login(c *fiber.Ctx) error {
	var req dto.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": err.Error()})
	}

	user, err := ac.UserService.Authenticate(c.Context(), req.Email, req.Password)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"message": err.Error()})
	}

	var clientApp *models.ClientApp
	if req.ClientID != "" {
		app, err := ac.ClientAppService.GetClientByClientID(c.Context(), req.ClientID)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"message": "Invalid client_id"})
		}
		clientApp = app
	}

	redirectURL := c.Query("redirect_url")
	if redirectURL != "" {
		if clientApp != nil {
			allowed := false
			for _, uri := range strings.Split(clientApp.RedirectURIs, ",") {
				if strings.TrimSpace(uri) == redirectURL {
					allowed = true
					break
				}
			}
			if !allowed {
				return c.Status(400).JSON(fiber.Map{"message": "Invalid redirect_url"})
			}
		}

		code, err := ac.UserService.GenerateAuthCode(c.Context(), user.ID.String(), req.Scope)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"message": "failed to generate auth code"})
		}
		return c.Redirect(redirectURL + "?code=" + code)
	}

	token, err := ac.UserService.GetToken(c.Context(), user.ID.String(), "", req.Scope)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": err.Error()})
	}

	userLogin, err := ac.UserService.GetUser(c.Context(), user.ID.String())
	var roleString strings.Builder
	for i, role := range userLogin.Role {
		if i > 0 && i < len(user.Role)-1 {
			roleString.WriteString(",")
		}
		roleString.WriteString(role.Name)
	}
	response := &dto.UserLoginResponse{
		Fullname:  userLogin.Fullname,
		Email:     userLogin.Email,
		Role:      roleString.String(),
		AvatarURI: userLogin.AvatarURI,
	}
	if err != nil {
		response = &dto.UserLoginResponse{}
	}

	return c.JSON(fiber.Map{
		"access_token": token,
		"user":         response,
	})
}

func (ac *AuthController) ExchangeCode(c *fiber.Ctx) error {
	var req struct {
		Code string `json:"code"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}

	user, scope, err := ac.UserService.ExchangeCode(c.Context(), req.Code)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	token, err := ac.UserService.GetToken(c.Context(), user.ID.String(), "", scope)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    token,
		Expires:  time.Now().Add(24 * time.Hour),
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Lax",
	})

	var roleString strings.Builder
	for _, role := range user.Role {
		roleString.WriteString(role.Name)
	}

	return c.JSON(fiber.Map{
		"token": token,
		"user": fiber.Map{
			"email": user.Email,
			"role":  roleString.String(),
		},
	})
}

func (ac *AuthController) Logout(c *fiber.Ctx) error {
	token := c.Cookies("access_token")
	if token != "" {
		_ = ac.UserService.Logout(c.Context(), token)
	}

	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
		SameSite: "Lax",
	})

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Logged out successfully",
	})
}

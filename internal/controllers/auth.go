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

func (ac *AuthController) ReaderRegister(c *fiber.Ctx) error {
	user, valErrors, err := ac.createUser(c, "Blog:Reader")
	if valErrors != nil {
		return c.Status(400).JSON(fiber.Map{
			"message": "validation error",
			"errors":  valErrors,
		})
	}
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"message": err.Error()})
	}
	if user == nil {
		return c.Status(500).JSON(fiber.Map{"message": "User creation failed unexpectedly"})
	}
	return c.Status(201).JSON(ac.mapUser(*user))
}

func (ac *AuthController) EditorRegister(c *fiber.Ctx) error {
	user, valErrors, err := ac.createUser(c, "Blog:Editor")
	if valErrors != nil {
		return c.Status(400).JSON(fiber.Map{
			"message": "validation error",
			"errors":  valErrors,
		})
	}
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": err.Error()})
	}
	if user == nil {
		return c.Status(500).JSON(fiber.Map{"message": "User creation failed unexpectedly"})
	}
	return c.Status(201).JSON(ac.mapUser(*user))
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

	// Simple validation for login if needed, though usually just email/pass presence
	// helper.ValidateStruct(req) could be used if dto has tags. Assuming it does or minimal validation.

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

	userDto, err := ac.UserService.GetUser(c.Context(), user.ID.String())
	if err != nil {
		// Fallback if profile not found but user exists
		userDto = &dto.JoinUser{Email: user.Email}
	}

	return c.JSON(fiber.Map{
		"access_token": token,
		"user":         userDto,
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

func (ac *AuthController) ChangePassword(c *fiber.Ctx) error {
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

	err = ac.UserService.ChangePassword(c.Context(), user.ID.String(), req)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"message": err.Error()})
	}

	return c.Status(200).JSON(fiber.Map{"message": "password changed successfully"})
}

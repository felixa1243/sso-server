package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func AdminMiddleware(c *fiber.Ctx) error {
	token, ok := c.Locals("user").(*jwt.Token)
	if !ok || token == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Invalid token claims"})
	}

	rolesStr, ok := claims["role"].(string)
	if !ok || !strings.Contains(rolesStr, "Administrator") {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "Forbidden: Admin access required",
		})
	}

	return c.Next()
}

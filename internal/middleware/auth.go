package middleware

import (
	"crypto/rsa"
	"sso-server/internal/helper"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func AuthMiddleware(publicKey *rsa.PublicKey) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tokenString := strings.Split(c.Get("Authorization"), "Bearer ")[1]
		token, err := helper.VerifyToken(tokenString, publicKey)
		if err != nil || !token.Valid {
			return c.Status(401).JSON(fiber.Map{"message": "Invalid token"})
		}
		c.Locals("user", token)

		return c.Next()
	}
}

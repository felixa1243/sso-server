package middleware

import (
	"crypto/rsa"
	"sso-server/internal/helper"
	"strings"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func AuthMiddleware(publicKey *rsa.PublicKey, rdb *redis.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Missing or invalid token"})
		}
		tokenString := strings.Split(authHeader, "Bearer ")[1]
		token, err := helper.VerifyToken(tokenString, publicKey)
		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Invalid token"})
		}

		// Check if token exists in Redis (Whitelist/Revocation Check)
		// Assuming the key format "access_token:<token>" -> userID
		// If key doesn't exist, token is revoked/logged out.
		ctx := c.Context()
		userIDStr, err := rdb.Get(ctx, "access_token:"+tokenString).Result()
		if err == redis.Nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Token revoked or expired"})
		} else if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
		}

		// Check if user is banned
		// We use a separate key for banned users to avoid checking token payload first?
		// Actually, token payload has userID, but Redis key value is also userID.
		// Let's check "user:<userID>:banned"
		isBanned, err := rdb.Get(ctx, "user:"+userIDStr+":banned").Result()
		if err == nil && isBanned == "true" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "User is banned"})
		}

		c.Locals("user", token)

		// Also verify user_id in token matches redis value for extra safety?
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Invalid token claims"})
		}
		claimUserID, ok := claims["user_id"].(string)
		if !ok || claimUserID != userIDStr {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Token user mismatch"})
		}

		// Check user ID validity format
		if _, err := uuid.Parse(claimUserID); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Invalid user ID"})
		}

		return c.Next()
	}
}

package helper

import (
	"crypto/rsa"
	"fmt"
	"sso-server/internal/models"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func GenerateToken(user *models.User, fullname string, privateKey *rsa.PrivateKey, domain *models.Domain) (string, error) {
	var rolesString strings.Builder
	var permissionsString strings.Builder

	for _, role := range user.Role {
		// If domain provided, check if role belongs to this domain or is global
		if domain != nil && role.DomainID != nil && *role.DomainID != domain.ID {
			continue
		}

		rolesString.WriteString(role.Name)

		for _, perm := range role.Permissions {
			if permissionsString.Len() > 0 {
				permissionsString.WriteString(" ")
			}
			permissionsString.WriteString(perm.Slug)
		}
	}

	claims := jwt.MapClaims{
		"sub":        user.ID.String(),
		"jti":        uuid.New().String(),
		"exp":        jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
		"iat":        jwt.NewNumericDate(time.Now()),
		"fullname":   fullname,
		"email":      user.Email,
		"role":       rolesString.String(),
		"scope":      permissionsString.String(),
		"user_id":    user.ID.String(),
		"avatar_uri": user.Profile.AvatarURI,
	}

	if domain != nil {
		claims["aud"] = domain.URL
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = "sso-key-1"
	return token.SignedString(privateKey)
}

func GetUserFromContext(c *fiber.Ctx) (models.User, error) {
	var user models.User
	token, ok := c.Locals("user").(*jwt.Token)
	if !ok || token == nil {
		return user, fmt.Errorf("token not found in context")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return user, fmt.Errorf("invalid token claims")
	}
	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return user, fmt.Errorf("user_id claim missing")
	}

	id, err := uuid.Parse(userIDStr)
	if err != nil {
		return user, fmt.Errorf("invalid uuid format")
	}
	user.ID = id
	if email, ok := claims["email"].(string); ok {
		user.Email = email
	}

	if roleName, ok := claims["role"].(string); ok {
		user.Role = []models.Role{{Name: roleName}}
	}

	return user, nil
}
func VerifyToken(tokenString string, publicKey *rsa.PublicKey) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {

		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})
}

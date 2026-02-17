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

func GenerateToken(user *models.User, fullname string, privateKey *rsa.PrivateKey, scope string) (string, error) {
	var rolesString strings.Builder
	var permissionsString strings.Builder

	// In new logic, scope param dictates what permissions are granted/requested.
	// But usually we just grant what the user has.
	// We can use scope to filter roles if we want, but "backend figured itself" might mean
	// just give everything allowed.
	// Let's assume we populate all roles for now, or filter by requested scope if implemented.

	// Previous logic filtered by Domain. Now we don't pass Domain object.
	// We just iterate all roles.
	// If we want to support legacy domain scoping via 'scope' param (e.g. "domain:blog"), we could parse it.

	for _, role := range user.Role {
		rolesString.WriteString(role.Name)

		for _, perm := range role.Permissions {
			if permissionsString.Len() > 0 {
				permissionsString.WriteString(" ")
			}
			permissionsString.WriteString(perm.Slug)
		}
	}

	// If scope is provided, we might want to put it in the token claim "scope"
	// OR use it to limit permissionsString.
	// For OIDC/OAuth, "scope" claim is what was granted.
	finalScope := permissionsString.String()
	if scope != "" {
		// Just append requested scope or replace?
		// Standard: 'scope' claim contains space-separated scopes.
		// We'll use the user's permissions as the scope.
	}

	claims := jwt.MapClaims{
		"sub":        user.ID.String(),
		"jti":        uuid.New().String(),
		"exp":        jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
		"iat":        jwt.NewNumericDate(time.Now()),
		"fullname":   fullname,
		"email":      user.Email,
		"role":       rolesString.String(),
		"scope":      finalScope,
		"user_id":    user.ID.String(),
		"avatar_uri": user.Profile.AvatarURI,
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

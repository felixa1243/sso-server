package helper

import (
	"crypto/rsa"
	"fmt"
	"sso-server/internal/models"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func GenerateToken(user models.User, privateKey *rsa.PrivateKey) (string, error) {
	claims := jwt.MapClaims{
		"sub":     user.ID.String(),
		"jti":     uuid.New().String(),
		"exp":     jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
		"iat":     jwt.NewNumericDate(time.Now()),
		"role":    user.Role.Name,
		"user_id": user.ID.String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(privateKey)
}

func VerifyToken(tokenString string, publicKey *rsa.PublicKey) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {

		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})
}

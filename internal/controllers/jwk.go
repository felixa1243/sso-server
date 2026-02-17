package controllers

import (
	"crypto/rsa"
	"encoding/base64"
	"math/big"

	"github.com/gofiber/fiber/v2"
)

type JWKController struct {
	PublicKey *rsa.PublicKey
}

func NewJWKController(publicKey *rsa.PublicKey) *JWKController {
	return &JWKController{
		PublicKey: publicKey,
	}
}

type JWK struct {
	Kty string `json:"kty"`
	E   string `json:"e"`
	N   string `json:"n"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	Kid string `json:"kid,omitempty"`
}

type JWKS struct {
	Keys []JWK `json:"keys"`
}

func (c *JWKController) GetJWKs(ctx *fiber.Ctx) error {
	// Encode N and E
	nBytes := c.PublicKey.N.Bytes()
	eBytes := big.NewInt(int64(c.PublicKey.E)).Bytes()

	jwk := JWK{
		Kty: "RSA",
		Use: "sig",
		Alg: "RS256",
		E:   base64.RawURLEncoding.EncodeToString(eBytes),
		N:   base64.RawURLEncoding.EncodeToString(nBytes),
		Kid: "sso-key-1",
	}

	return ctx.JSON(JWKS{
		Keys: []JWK{jwk},
	})
}

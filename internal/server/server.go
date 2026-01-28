package server

import (
	"crypto/rsa"

	"github.com/gofiber/fiber/v2"

	"sso-server/internal/database"
)

type FiberServer struct {
	*fiber.App
	db         database.Service
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
}

func New(privKey *rsa.PrivateKey, pubKey *rsa.PublicKey) *FiberServer {
	server := &FiberServer{
		App: fiber.New(fiber.Config{
			ServerHeader: "sso-server",
			AppName:      "sso-server",
		}),
		PrivateKey: privKey,
		PublicKey:  pubKey,
		db:         database.New(),
	}

	return server
}

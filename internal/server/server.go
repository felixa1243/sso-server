package server

import (
	"crypto/rsa"
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"

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
			ErrorHandler: func(c *fiber.Ctx, err error) error {
				code := fiber.StatusInternalServerError
				message := "An Internal server error occured"
				var e *fiber.Error
				if errors.As(err, &e) {
					code = e.Code
					message = e.Message
				}

				fmt.Printf("[ERROR] %d: %v\n", code, err)
				return c.Status(code).JSON(fiber.Map{
					"status":  false,
					"message": message,
				})
			},
		}),
		PrivateKey: privKey,
		PublicKey:  pubKey,
		db:         database.New(),
	}
	server.App.Use(logger.New(logger.Config{
		Format: "[${ip}]:${port} ${status} - ${method} ${path}\n",
	}))
	server.App.Static("/uploads", "./uploads")
	return server
}

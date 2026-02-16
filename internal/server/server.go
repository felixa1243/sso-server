package server

import (
	"crypto/rsa"
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/template/html/v2"

	"sso-server/internal/database"
)

type FiberServer struct {
	*fiber.App
	db         database.Service
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
}

func New(privKey *rsa.PrivateKey, pubKey *rsa.PublicKey) *FiberServer {
	engine := html.New("./resources/views", ".html")
	server := &FiberServer{
		App: fiber.New(fiber.Config{
			ServerHeader: "sso-server",
			AppName:      "sso-server",
			Views:        engine,
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
	server.Static("/assets", "./resources/assets")
	return server
}

package server

import (
	"github.com/gofiber/fiber/v2"

	"sso-server/internal/database"
)

type FiberServer struct {
	*fiber.App

	db database.Service
}

func New() *FiberServer {
	server := &FiberServer{
		App: fiber.New(fiber.Config{
			ServerHeader: "sso-server",
			AppName:      "sso-server",
		}),

		db: database.New(),
	}

	return server
}

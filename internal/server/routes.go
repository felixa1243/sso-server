package server

import (
	"sso-server/internal/controllers"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/samber/do/v2"
)

func (s *FiberServer) RegisterFiberRoutes() {
	s.App.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS,PATCH",
	}))

	injector := SetupDI(s.db.GetDB(), s.db.GetRedis(), s.PrivateKey, s.PublicKey)
	// Resolve the AuthController
	authControllers := do.MustInvoke[*controllers.AuthController](injector)
	// Resolve Middleware
	authMiddleware := do.MustInvoke[fiber.Handler](injector)

	// Routes
	s.App.Post("/register/reader", authControllers.ReaderRegister)
	s.App.Post("/register/editor", authControllers.EditorRegister)
	s.App.Post("/login", authControllers.Login)
	s.App.Get("/login", authControllers.ShowLogin)
	s.App.Post("/logout", authControllers.Logout)
	s.App.Post("/exchange", authControllers.ExchangeCode)

	// Protected Routes
	s.App.Post("/change-password", authMiddleware, authControllers.ChangePassword)
}

func (s *FiberServer) HelloWorldHandler(c *fiber.Ctx) error {
	resp := fiber.Map{
		"message": "Hello World",
	}

	return c.JSON(resp)
}

func (s *FiberServer) healthHandler(c *fiber.Ctx) error {
	return c.JSON(s.db.Health())
}

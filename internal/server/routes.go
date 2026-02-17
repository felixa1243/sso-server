package server

import (
	"sso-server/internal/controllers"
	"sso-server/internal/middleware"

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
	jwkController := do.MustInvoke[*controllers.JWKController](injector)
	// Resolve Middleware
	authMiddleware := do.MustInvoke[fiber.Handler](injector)

	// Routes
	s.App.Get("/.well-known/jwks.json", jwkController.GetJWKs)
	s.App.Post("/register/reader", authControllers.ReaderRegister)
	s.App.Post("/register/editor", authControllers.EditorRegister)
	s.App.Post("/login", authControllers.Login)
	s.App.Get("/login", authControllers.ShowLogin)
	s.App.Post("/logout", authControllers.Logout)
	s.App.Post("/exchange", authControllers.ExchangeCode)

	// Protected Routes
	s.App.Post("/change-password", authMiddleware, authControllers.ChangePassword)

	// Admin Routes
	userManagementController := do.MustInvoke[*controllers.UserManagementController](injector)
	admin := s.App.Group("/admin", authMiddleware, middleware.AdminMiddleware)
	admin.Get("/users", userManagementController.ListUsers)
	admin.Delete("/users/:id", userManagementController.DeleteUser)
	admin.Put("/users/:id/roles", userManagementController.UpdateUserRoles)
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

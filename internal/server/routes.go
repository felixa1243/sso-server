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
	clientAppController := do.MustInvoke[*controllers.ClientAppController](injector)
	domainController := do.MustInvoke[*controllers.DomainController](injector)

	s.App.Post("/change-password", authMiddleware, authControllers.ChangePassword)
	s.App.Post("/clients", authMiddleware, clientAppController.Register)
	s.App.Get("/clients", authMiddleware, clientAppController.List)
	s.App.Post("/domains", authMiddleware, domainController.Register)
	s.App.Get("/domains", authMiddleware, domainController.List)

	// Admin Routes
	userManagementController := do.MustInvoke[*controllers.UserManagementController](injector)
	punishmentController := do.MustInvoke[*controllers.PunishmentController](injector)
	rbacController := do.MustInvoke[*controllers.RBACController](injector)

	admin := s.App.Group("/admin", authMiddleware, middleware.AdminMiddleware)
	admin.Get("/users", userManagementController.ListUsers)
	admin.Delete("/users/:id", userManagementController.DeleteUser)
	admin.Put("/users/:id/roles", userManagementController.UpdateUserRoles)
	admin.Put("/users/:id/ban", userManagementController.BanUser)
	admin.Put("/users/:id/unban", userManagementController.UnbanUser)

	// Punishment
	admin.Post("/punishments", punishmentController.PunishUser)
	admin.Get("/users/:id/punishments", punishmentController.GetUserPunishments)
	admin.Delete("/punishments/:id", punishmentController.RevokePunishment)

	// RBAC
	admin.Get("/roles", rbacController.ListRoles)
	admin.Post("/roles", rbacController.CreateRole)
	admin.Put("/roles/:id", rbacController.UpdateRole)
	admin.Delete("/roles/:id", rbacController.DeleteRole)
	admin.Post("/roles/:id/permissions", rbacController.AssignPermissions)

	admin.Get("/permissions", rbacController.ListPermissions)
	admin.Post("/permissions", rbacController.CreatePermission)

	// SPA Fallback - Must be last
	s.App.Get("/*", func(c *fiber.Ctx) error {
		return c.SendFile("./resources/public/index.html")
	})
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

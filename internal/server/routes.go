package server

import (
	"log"
	"sso-server/internal/controllers"
	"sso-server/internal/database"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func (s *FiberServer) RegisterFiberRoutes() {
	// Apply CORS middleware
	s.App.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS,PATCH",
		AllowHeaders:     "Accept,Authorization,Content-Type",
		AllowCredentials: false, // credentials require explicit origins
		MaxAge:           300,
	}))

	s.App.Get("/", s.HelloWorldHandler)
	db := database.New().GetDB()
	authControllers := &controllers.AuthController{
		DB:         db,
		PrivateKey: s.PrivateKey,
		Redis:      s.db.GetRedis(),
	}
	s.App.Post("/register/reader", authControllers.ReaderRegister)
	s.App.Post("/register/editor", authControllers.EditorRegister)
	if s.PublicKey == nil {
		log.Fatal("CRITICAL: RSA Public Key is nil. Check Your Configuration")
	}
	s.App.Post("/login", authControllers.Login)
	s.App.Get("/login", authControllers.ShowLogin)
	s.App.Get("/register/reader", authControllers.ShowRegister)
	s.App.Post("/exchange", authControllers.ExchangeCode)
	s.App.Get("/health", s.healthHandler)

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

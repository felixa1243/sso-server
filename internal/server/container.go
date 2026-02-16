package server

import (
	"crypto/rsa"
	"sso-server/internal/controllers"
	"sso-server/internal/middleware"
	"sso-server/internal/repositories"
	"sso-server/internal/services"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/samber/do/v2"
	"gorm.io/gorm"
)

func SetupDI(db *gorm.DB, rdb *redis.Client, key *rsa.PrivateKey, publicKey *rsa.PublicKey) do.Injector {
	injector := do.New()

	// External dependencies
	do.ProvideValue(injector, db)
	do.ProvideValue(injector, rdb)
	do.ProvideValue(injector, publicKey)
	do.ProvideValue(injector, key)

	// Middleware
	do.Provide(injector, func(i do.Injector) (fiber.Handler, error) {
		pk := do.MustInvoke[*rsa.PublicKey](i)
		return middleware.AuthMiddleware(pk), nil
	})

	// Repositories
	do.Provide(injector, func(i do.Injector) (repositories.UserRepository, error) {
		return repositories.NewUserRepository(do.MustInvoke[*gorm.DB](i)), nil
	})
	do.Provide(injector, func(i do.Injector) (repositories.ProfileRepository, error) {
		return repositories.NewProfileRepository(do.MustInvoke[*gorm.DB](i)), nil
	})
	do.Provide(injector, func(i do.Injector) (repositories.RoleRepository, error) {
		return repositories.NewRoleRepository(do.MustInvoke[*gorm.DB](i)), nil
	})

	// Services
	do.Provide(injector, func(i do.Injector) (services.UserService, error) {
		repo := do.MustInvoke[repositories.UserRepository](i)
		profileRepo := do.MustInvoke[repositories.ProfileRepository](i)
		roleRepo := do.MustInvoke[repositories.RoleRepository](i)
		redisInstance := do.MustInvoke[*redis.Client](i)
		privateKey := do.MustInvoke[*rsa.PrivateKey](i)
		return services.NewUserServices(repo, profileRepo, roleRepo, redisInstance, privateKey), nil
	})

	// Controllers
	do.Provide(injector, func(i do.Injector) (*controllers.AuthController, error) {
		return controllers.NewAuthController(
			do.MustInvoke[services.UserService](i),
		), nil
	})

	return injector
}

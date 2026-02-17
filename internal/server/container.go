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
		rdb := do.MustInvoke[*redis.Client](i)
		return middleware.AuthMiddleware(pk, rdb), nil
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
	do.Provide(injector, func(i do.Injector) (repositories.DomainRepository, error) {
		return repositories.NewDomainRepository(do.MustInvoke[*gorm.DB](i)), nil
	})
	do.Provide(injector, func(i do.Injector) (repositories.PunishmentRepository, error) {
		return repositories.NewPunishmentRepository(do.MustInvoke[*gorm.DB](i)), nil
	})
	do.Provide(injector, func(i do.Injector) (repositories.PermissionRepository, error) {
		return repositories.NewPermissionRepository(do.MustInvoke[*gorm.DB](i)), nil
	})

	// Services
	do.Provide(injector, func(i do.Injector) (services.EventService, error) {
		return services.NewEventService(do.MustInvoke[*redis.Client](i)), nil
	})
	do.Provide(injector, func(i do.Injector) (services.UserService, error) {
		repo := do.MustInvoke[repositories.UserRepository](i)
		profileRepo := do.MustInvoke[repositories.ProfileRepository](i)
		roleRepo := do.MustInvoke[repositories.RoleRepository](i)
		domainRepo := do.MustInvoke[repositories.DomainRepository](i)
		eventService := do.MustInvoke[services.EventService](i)
		redisInstance := do.MustInvoke[*redis.Client](i)
		privateKey := do.MustInvoke[*rsa.PrivateKey](i)
		return services.NewUserServices(repo, profileRepo, roleRepo, domainRepo, eventService, redisInstance, privateKey), nil
	})

	// Services (RBAC & Punishment)
	do.Provide(injector, func(i do.Injector) (services.PunishmentService, error) {
		return services.NewPunishmentService(
			do.MustInvoke[repositories.PunishmentRepository](i),
			do.MustInvoke[repositories.UserRepository](i),
			do.MustInvoke[services.EventService](i),
			do.MustInvoke[*redis.Client](i),
		), nil
	})
	do.Provide(injector, func(i do.Injector) (services.RBACService, error) {
		return services.NewRBACService(
			do.MustInvoke[repositories.RoleRepository](i),
			do.MustInvoke[repositories.PermissionRepository](i),
			do.MustInvoke[services.EventService](i),
		), nil
	})

	// Controllers
	do.Provide(injector, func(i do.Injector) (*controllers.AuthController, error) {
		return controllers.NewAuthController(
			do.MustInvoke[services.UserService](i),
		), nil
	})
	do.Provide(injector, func(i do.Injector) (*controllers.JWKController, error) {
		return controllers.NewJWKController(
			do.MustInvoke[*rsa.PublicKey](i),
		), nil
	})
	do.Provide(injector, func(i do.Injector) (*controllers.UserManagementController, error) {
		return controllers.NewUserManagementController(
			do.MustInvoke[services.UserService](i),
		), nil
	})
	do.Provide(injector, func(i do.Injector) (*controllers.PunishmentController, error) {
		return controllers.NewPunishmentController(
			do.MustInvoke[services.PunishmentService](i),
		), nil
	})
	do.Provide(injector, func(i do.Injector) (*controllers.RBACController, error) {
		return controllers.NewRBACController(
			do.MustInvoke[services.RBACService](i),
		), nil
	})

	return injector
}

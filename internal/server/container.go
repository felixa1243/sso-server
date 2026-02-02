package server

import (
	"crypto/rsa"
	"sso-server/internal/controllers"
	"sso-server/internal/repositories"
	"sso-server/internal/services"

	"github.com/go-redis/redis/v8"
	"github.com/samber/do/v2" // Ensure v2 is used
	"gorm.io/gorm"
)

// Changed return type from *do.Injector to do.Injector
func SetupDI(db *gorm.DB, rdb *redis.Client, key *rsa.PrivateKey, publicKey *rsa.PublicKey) do.Injector {
	injector := do.New()
	do.ProvideValue(injector, db)
	do.ProvideValue(injector, rdb)
	do.ProvideValue(injector, publicKey)
	do.ProvideValue(injector, key)
	do.Provide(injector, func(i do.Injector) (repositories.UserRepository, error) {
		dbInstance := do.MustInvoke[*gorm.DB](i) // Use 'i' as value, not '*i'
		return repositories.NewUserRepository(dbInstance), nil
	})

	// 3. Register Service
	do.Provide(injector, func(i do.Injector) (*services.UserServices, error) {
		repo := do.MustInvoke[repositories.UserRepository](i)
		// Fixed: Only pass 'repo' as per your user-services.go definition
		return services.NewUserServices(repo, rdb), nil
	})

	// 4. Register Controller
	do.Provide(injector, func(i do.Injector) (*controllers.AuthController, error) {
		userService := do.MustInvoke[*services.UserServices](i)
		dbInstance := do.MustInvoke[*gorm.DB](i)
		redisInstance := do.MustInvoke[*redis.Client](i)
		privateKey := do.MustInvoke[*rsa.PrivateKey](i)

		return &controllers.AuthController{
			DB:          dbInstance,
			PrivateKey:  privateKey,
			Redis:       redisInstance,
			UserService: userService,
		}, nil
	})

	return injector
}

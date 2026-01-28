package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"sso-server/internal/helper"
	"sso-server/internal/models"
	"strconv"
	"time"

	"github.com/google/uuid"
	_ "github.com/joho/godotenv/autoload"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Service interface {
	Health() map[string]string
	Close() error
	GetDB() *gorm.DB
	SeedPermissionsAndRoles() error
}

type service struct {
	db *gorm.DB
}

var (
	database   = os.Getenv("BLUEPRINT_DB_DATABASE")
	password   = os.Getenv("BLUEPRINT_DB_PASSWORD")
	username   = os.Getenv("BLUEPRINT_DB_USERNAME")
	port       = os.Getenv("BLUEPRINT_DB_PORT")
	host       = os.Getenv("BLUEPRINT_DB_HOST")
	schema     = os.Getenv("BLUEPRINT_DB_SCHEMA")
	dbInstance *service
)

func New() Service {
	if dbInstance != nil {
		return dbInstance
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable search_path=%s",
		host, username, password, database, port, schema)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	db.AutoMigrate(&models.User{}, &models.Role{}, &models.Permission{}, &models.UserProfile{})
	dbInstance = &service{
		db: db,
	}
	return dbInstance
}

func (s *service) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	stats := make(map[string]string)

	// GORM uses the standard database/sql pool internally
	sqlDB, err := s.db.DB()
	if err != nil {
		stats["status"] = "down"
		stats["error"] = "could not get underlying sql.DB"
		return stats
	}

	err = sqlDB.PingContext(ctx)
	if err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db down: %v", err)
		return stats
	}

	stats["status"] = "up"
	stats["message"] = "It's healthy"

	dbStats := sqlDB.Stats()
	stats["open_connections"] = strconv.Itoa(dbStats.OpenConnections)
	stats["in_use"] = strconv.Itoa(dbStats.InUse)
	stats["idle"] = strconv.Itoa(dbStats.Idle)
	stats["wait_count"] = strconv.FormatInt(dbStats.WaitCount, 10)
	stats["wait_duration"] = dbStats.WaitDuration.String()
	stats["max_idle_closed"] = strconv.FormatInt(dbStats.MaxIdleClosed, 10)
	stats["max_lifetime_closed"] = strconv.FormatInt(dbStats.MaxLifetimeClosed, 10)

	if dbStats.OpenConnections > 40 {
		stats["message"] = "The database is experiencing heavy load."
	}

	return stats
}

func (s *service) Close() error {
	log.Printf("Disconnected from database: %s", database)
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (s *service) GetDB() *gorm.DB {
	return s.db
}

func (s *service) SeedPermissionsAndRoles() error {
	perms := []models.Permission{
		{Name: "View Blog", Slug: "blog:read"},
		{Name: "Write Blog", Slug: "blog:write"},
		{Name: "View Library", Slug: "library:read"},
		{Name: "Write Library", Slug: "library:write"},
	}

	for _, p := range perms {
		if err := s.db.Where(models.Permission{Slug: p.Slug}).FirstOrCreate(&p).Error; err != nil {
			return err
		}
	}

	var blogRead models.Permission
	s.db.Where("slug = ?", "blog:read").First(&blogRead)

	var blogWrite models.Permission
	s.db.Where("slug = ?", "blog:write").First(&blogWrite)

	roles := []models.Role{
		{
			Name:        "Reader",
			Permissions: []models.Permission{blogRead},
		},
		{
			Name:        "Editor",
			Permissions: []models.Permission{blogRead, blogWrite},
		},
		{
			Name:        "Administrator",
			Permissions: []models.Permission{blogRead, blogWrite},
		},
	}
	var UserCreated models.User
	s.db.Where("email = ?", "felixarajiph@gmail.com").First(&UserCreated)

	for _, r := range roles {
		if err := s.db.Where(models.Role{Name: r.Name}).FirstOrCreate(&r).Error; err != nil {
			return err
		}
	}
	var adminRole models.Role
	s.db.Where("name=?", "Administrator").First(&adminRole)
	if UserCreated.ID == uuid.Nil {
		UserCreated = models.User{
			Email:        "felixarajiph@gmail.com",
			ID:           uuid.New(),
			PasswordHash: helper.GeneratePassword("password"),
			RoleID:       adminRole.ID,
		}
		if err := s.db.Create(&UserCreated).Error; err != nil {
			return err
		}
	}
	return nil
}

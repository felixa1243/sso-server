package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ClientApp struct {
	ID            uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name          string         `gorm:"type:varchar(100);not null"`
	Secret        string         `gorm:"not null"` // Hashed?
	RedirectURIs  string         `gorm:"type:text;not null"` // Comma separated or JSON
	AllowedScopes string         `gorm:"type:text"`          // Space separated permissions/roles
	UserID        uuid.UUID      `gorm:"type:uuid;not null;index"` // Owner
	User          User           `gorm:"foreignKey:UserID"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"`
}

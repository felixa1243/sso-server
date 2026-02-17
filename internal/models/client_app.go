package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ClientApp struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name         string         `gorm:"not null"`
	ClientID     string         `gorm:"uniqueIndex;not null"`
	ClientSecret string         `gorm:"not null" json:"-"`
	RedirectURIs string         `gorm:"not null"` // Comma-separated
	Scopes       string         `gorm:"not null"` // Space-separated allowed scopes
	ScopeEntities []*Scope      `gorm:"many2many:client_app_scopes;"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
	UserID       uuid.UUID      `gorm:"type:uuid;not null"` // Owner
	User         User           `gorm:"foreignKey:UserID"`
}

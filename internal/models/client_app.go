package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ClientApp struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name         string    `gorm:"not null"`
	ClientID     string    `gorm:"uniqueIndex;not null"`
	ClientSecret string    `gorm:"not null" json:"-"`
	RedirectURIs string    `gorm:"not null"`
	Scopes       []Scope   `gorm:"many2many:client_app_scopes;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
	UserID       uuid.UUID      `gorm:"type:uuid;not null"`
	User         User           `gorm:"foreignKey:UserID"`
}

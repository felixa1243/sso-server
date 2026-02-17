package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Scope struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name        string         `gorm:"uniqueIndex;not null"`
	Description string         `gorm:"not null"`
	IsDefault   bool           `gorm:"default:false"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

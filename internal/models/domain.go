package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Domain struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name      string         `gorm:"type:varchar(100);uniqueIndex;not null"`
	URL       string         `gorm:"type:varchar(255);not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	UserID    uuid.UUID      `gorm:"type:uuid;not null"`
	User      User           `gorm:"foreignKey:UserID"`
}

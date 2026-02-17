package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Punishment struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    uuid.UUID      `gorm:"type:uuid;not null;index"`
	User      User           `gorm:"foreignKey:UserID"`
	AdminID   uuid.UUID      `gorm:"type:uuid;not null"` // Who punished
	Type      string         `gorm:"type:varchar(50);not null"` // BAN, MUTE
	Reason    string         `gorm:"type:text"`
	StartTime time.Time      `gorm:"not null"`
	EndTime   time.Time      `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

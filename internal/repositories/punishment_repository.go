package repositories

import (
	"sso-server/internal/models"
	"time"

	"gorm.io/gorm"
)

type PunishmentRepository interface {
	Create(punishment *models.Punishment) error
	FindActiveByUserID(userID string, punishmentType string) (*models.Punishment, error)
	FindAllByUserID(userID string) ([]models.Punishment, error)
	Revoke(id string) error
}

type punishmentRepositoryImpl struct {
	db *gorm.DB
}

func NewPunishmentRepository(db *gorm.DB) PunishmentRepository {
	return &punishmentRepositoryImpl{db: db}
}

func (r *punishmentRepositoryImpl) Create(punishment *models.Punishment) error {
	return r.db.Create(punishment).Error
}

func (r *punishmentRepositoryImpl) FindActiveByUserID(userID string, punishmentType string) (*models.Punishment, error) {
	var punishment models.Punishment
	now := time.Now()
	if err := r.db.Where("user_id = ? AND type = ? AND start_time <= ? AND end_time > ?", userID, punishmentType, now, now).First(&punishment).Error; err != nil {
		return nil, err
	}
	return &punishment, nil
}

func (r *punishmentRepositoryImpl) FindAllByUserID(userID string) ([]models.Punishment, error) {
	var punishments []models.Punishment
	if err := r.db.Where("user_id = ?", userID).Find(&punishments).Error; err != nil {
		return nil, err
	}
	return punishments, nil
}

func (r *punishmentRepositoryImpl) Revoke(id string) error {
	return r.db.Model(&models.Punishment{}).Where("id = ?", id).Update("end_time", time.Now()).Error
}

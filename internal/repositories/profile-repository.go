package repositories

import (
	"errors"
	"sso-server/internal/dto"
	"sso-server/internal/models"

	"gorm.io/gorm"
)

type ProfileRepository interface {
	FindByUserID(id string) (dto.JoinUser, error)
}

type ProfileRepositoryImpl struct {
	db *gorm.DB
}

func (p *ProfileRepositoryImpl) FindByUserID(id string) (dto.JoinUser, error) {
	var user models.User
	result := p.db.
		Preload("Profile").
		Preload("Role").
		Where("id = ?", id).
		First(&user)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return dto.JoinUser{}, nil
		}
		return dto.JoinUser{}, result.Error
	}
	joinUser := dto.JoinUser{
		Fullname:  user.Profile.FullName,
		Email:     user.Email,
		Role:      user.Role,
		AvatarURI: user.Profile.AvatarURI,
	}

	return joinUser, nil
}
func NewProfileRepository(db *gorm.DB) ProfileRepository {
	return &ProfileRepositoryImpl{db: db}
}

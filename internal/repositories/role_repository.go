package repositories

import (
	"sso-server/internal/models"

	"gorm.io/gorm"
)

type RoleRepository interface {
	FindByName(name string) (*models.Role, error)
}

type roleRepositoryImpl struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) RoleRepository {
	return &roleRepositoryImpl{db: db}
}

func (r *roleRepositoryImpl) FindByName(name string) (*models.Role, error) {
	var role models.Role
	if err := r.db.Where("name = ?", name).First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

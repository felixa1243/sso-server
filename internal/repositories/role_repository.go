package repositories

import (
	"sso-server/internal/models"

	"gorm.io/gorm"
)

type RoleRepository interface {
	FindByName(name string) (*models.Role, error)
	FindByID(id string) (*models.Role, error)
	FindAll() ([]models.Role, error)
	FindByIds(ids []string) ([]models.Role, error)
	Create(role *models.Role) error
	Update(role *models.Role) error
	Delete(id string) error
	UpdatePermissions(role *models.Role, permissions []models.Permission) error
}

type roleRepositoryImpl struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) RoleRepository {
	return &roleRepositoryImpl{db: db}
}

func (r *roleRepositoryImpl) FindByName(name string) (*models.Role, error) {
	var role models.Role
	if err := r.db.Preload("Domain").Preload("Permissions").Where("name = ?", name).First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepositoryImpl) FindAll() ([]models.Role, error) {
	var roles []models.Role
	if err := r.db.Preload("Domain").Preload("Permissions").Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *roleRepositoryImpl) FindByID(id string) (*models.Role, error) {
	var role models.Role
	if err := r.db.Preload("Domain").Preload("Permissions").Where("id = ?", id).First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepositoryImpl) Create(role *models.Role) error {
	return r.db.Create(role).Error
}

func (r *roleRepositoryImpl) Update(role *models.Role) error {
	return r.db.Save(role).Error
}

func (r *roleRepositoryImpl) Delete(id string) error {
	return r.db.Delete(&models.Role{}, "id = ?", id).Error
}

func (r *roleRepositoryImpl) UpdatePermissions(role *models.Role, permissions []models.Permission) error {
	return r.db.Model(role).Association("Permissions").Replace(permissions)
}

func (r *roleRepositoryImpl) FindByIds(ids []string) ([]models.Role, error) {
	var roles []models.Role
	if err := r.db.Preload("Domain").Preload("Permissions").Where("id IN ?", ids).Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

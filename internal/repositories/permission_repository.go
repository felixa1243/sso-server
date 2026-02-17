package repositories

import (
	"sso-server/internal/models"

	"gorm.io/gorm"
)

type PermissionRepository interface {
	FindByName(name string) (*models.Permission, error)
	FindAll() ([]models.Permission, error)
	Create(permission *models.Permission) error
	Delete(id string) error
	Update(permission *models.Permission) error
	FindByIds(ids []string) ([]models.Permission, error)
}

type permissionRepositoryImpl struct {
	db *gorm.DB
}

func NewPermissionRepository(db *gorm.DB) PermissionRepository {
	return &permissionRepositoryImpl{db: db}
}

func (r *permissionRepositoryImpl) FindByName(name string) (*models.Permission, error) {
	var permission models.Permission
	if err := r.db.Where("name = ?", name).First(&permission).Error; err != nil {
		return nil, err
	}
	return &permission, nil
}

func (r *permissionRepositoryImpl) FindAll() ([]models.Permission, error) {
	var permissions []models.Permission
	if err := r.db.Find(&permissions).Error; err != nil {
		return nil, err
	}
	return permissions, nil
}

func (r *permissionRepositoryImpl) Create(permission *models.Permission) error {
	return r.db.Create(permission).Error
}

func (r *permissionRepositoryImpl) Delete(id string) error {
	return r.db.Delete(&models.Permission{}, "id = ?", id).Error
}

func (r *permissionRepositoryImpl) Update(permission *models.Permission) error {
	return r.db.Save(permission).Error
}

func (r *permissionRepositoryImpl) FindByIds(ids []string) ([]models.Permission, error) {
	var permissions []models.Permission
	if err := r.db.Where("id IN ?", ids).Find(&permissions).Error; err != nil {
		return nil, err
	}
	return permissions, nil
}

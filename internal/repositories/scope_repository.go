package repositories

import (
	"sso-server/internal/models"

	"gorm.io/gorm"
)

type ScopeRepository interface {
	FindAll() ([]models.Scope, error)
	FindByName(name string) (*models.Scope, error)
	FindByID(id string) (*models.Scope, error)
	FindByIds(ids []string) ([]models.Scope, error)
	Create(scope *models.Scope) error
	Update(scope *models.Scope) error
	Delete(id string) error
	FindByNames(names []string) ([]models.Scope, error)
}

type scopeRepositoryImpl struct {
	db *gorm.DB
}

func NewScopeRepository(db *gorm.DB) ScopeRepository {
	return &scopeRepositoryImpl{db: db}
}

func (r *scopeRepositoryImpl) FindAll() ([]models.Scope, error) {
	var scopes []models.Scope
	if err := r.db.Find(&scopes).Error; err != nil {
		return nil, err
	}
	return scopes, nil
}

func (r *scopeRepositoryImpl) FindByName(name string) (*models.Scope, error) {
	var scope models.Scope
	if err := r.db.Where("name = ?", name).First(&scope).Error; err != nil {
		return nil, err
	}
	return &scope, nil
}

func (r *scopeRepositoryImpl) FindByID(id string) (*models.Scope, error) {
	var scope models.Scope
	if err := r.db.Where("id = ?", id).First(&scope).Error; err != nil {
		return nil, err
	}
	return &scope, nil
}

func (r *scopeRepositoryImpl) FindByIds(ids []string) ([]models.Scope, error) {
	var scopes []models.Scope
	if err := r.db.Where("id IN ?", ids).Find(&scopes).Error; err != nil {
		return nil, err
	}
	return scopes, nil
}

func (r *scopeRepositoryImpl) Create(scope *models.Scope) error {
	return r.db.Create(scope).Error
}

func (r *scopeRepositoryImpl) Update(scope *models.Scope) error {
	return r.db.Save(scope).Error
}

func (r *scopeRepositoryImpl) Delete(id string) error {
	return r.db.Delete(&models.Scope{}, "id = ?", id).Error
}
func (r *scopeRepositoryImpl) FindByNames(names []string) ([]models.Scope, error) {
	var scopes []models.Scope
	if err := r.db.Where("name IN ?", names).Find(&scopes).Error; err != nil {
		return nil, err
	}
	return scopes, nil
}

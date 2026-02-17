package repositories

import (
	"sso-server/internal/models"

	"gorm.io/gorm"
)

type DomainRepository interface {
	FindByName(name string) (*models.Domain, error)
	FindAll() ([]models.Domain, error)
	FindAllByUserID(userID string) ([]models.Domain, error)
	Create(domain *models.Domain) error
	Delete(id string) error
	Update(domain *models.Domain) error
}

type domainRepositoryImpl struct {
	db *gorm.DB
}

func NewDomainRepository(db *gorm.DB) DomainRepository {
	return &domainRepositoryImpl{db: db}
}

func (r *domainRepositoryImpl) FindByName(name string) (*models.Domain, error) {
	var domain models.Domain
	if err := r.db.Where("name = ?", name).First(&domain).Error; err != nil {
		return nil, err
	}
	return &domain, nil
}

func (r *domainRepositoryImpl) FindAll() ([]models.Domain, error) {
	var domains []models.Domain
	if err := r.db.Find(&domains).Error; err != nil {
		return nil, err
	}
	return domains, nil
}

func (r *domainRepositoryImpl) FindAllByUserID(userID string) ([]models.Domain, error) {
	var domains []models.Domain
	if err := r.db.Where("user_id = ?", userID).Find(&domains).Error; err != nil {
		return nil, err
	}
	return domains, nil
}

func (r *domainRepositoryImpl) Create(domain *models.Domain) error {
	return r.db.Create(domain).Error
}

func (r *domainRepositoryImpl) Delete(id string) error {
	return r.db.Delete(&models.Domain{}, "id = ?", id).Error
}

func (r *domainRepositoryImpl) Update(domain *models.Domain) error {
	return r.db.Save(domain).Error
}

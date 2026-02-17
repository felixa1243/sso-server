package repositories

import (
	"sso-server/internal/models"

	"gorm.io/gorm"
)

type ClientAppRepository interface {
	Create(client *models.ClientApp) error
	FindByID(id string) (*models.ClientApp, error)
	FindAllByUserID(userID string) ([]models.ClientApp, error)
	Update(client *models.ClientApp) error
	Delete(id string) error
}

type clientAppRepositoryImpl struct {
	db *gorm.DB
}

func NewClientAppRepository(db *gorm.DB) ClientAppRepository {
	return &clientAppRepositoryImpl{db: db}
}

func (r *clientAppRepositoryImpl) Create(client *models.ClientApp) error {
	return r.db.Create(client).Error
}

func (r *clientAppRepositoryImpl) FindByID(id string) (*models.ClientApp, error) {
	var client models.ClientApp
	if err := r.db.Where("id = ?", id).First(&client).Error; err != nil {
		return nil, err
	}
	return &client, nil
}

func (r *clientAppRepositoryImpl) FindAllByUserID(userID string) ([]models.ClientApp, error) {
	var clients []models.ClientApp
	if err := r.db.Where("user_id = ?", userID).Find(&clients).Error; err != nil {
		return nil, err
	}
	return clients, nil
}

func (r *clientAppRepositoryImpl) Update(client *models.ClientApp) error {
	return r.db.Save(client).Error
}

func (r *clientAppRepositoryImpl) Delete(id string) error {
	return r.db.Delete(&models.ClientApp{}, "id = ?", id).Error
}

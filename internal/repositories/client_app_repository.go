package repositories

import (
	"context"
	"sso-server/internal/models"

	"gorm.io/gorm"
)

type ClientAppRepository interface {
	Create(ctx context.Context, clientApp *models.ClientApp) error
	FindByID(ctx context.Context, id string) (*models.ClientApp, error)
	FindByClientID(ctx context.Context, clientID string) (*models.ClientApp, error)
	FindByUserID(ctx context.Context, userID string) ([]models.ClientApp, error)
	Delete(ctx context.Context, id string) error
}

type clientAppRepositoryImpl struct {
	db *gorm.DB
}

func NewClientAppRepository(db *gorm.DB) ClientAppRepository {
	return &clientAppRepositoryImpl{db: db}
}

func (r *clientAppRepositoryImpl) Create(ctx context.Context, clientApp *models.ClientApp) error {
	return r.db.WithContext(ctx).Create(clientApp).Error
}

func (r *clientAppRepositoryImpl) FindByID(ctx context.Context, id string) (*models.ClientApp, error) {
	var clientApp models.ClientApp
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&clientApp).Error; err != nil {
		return nil, err
	}
	return &clientApp, nil
}

func (r *clientAppRepositoryImpl) FindByClientID(ctx context.Context, clientID string) (*models.ClientApp, error) {
	var clientApp models.ClientApp
	if err := r.db.WithContext(ctx).Where("client_id = ?", clientID).First(&clientApp).Error; err != nil {
		return nil, err
	}
	return &clientApp, nil
}

func (r *clientAppRepositoryImpl) FindByUserID(ctx context.Context, userID string) ([]models.ClientApp, error) {
	var apps []models.ClientApp
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&apps).Error; err != nil {
		return nil, err
	}
	return apps, nil
}

func (r *clientAppRepositoryImpl) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.ClientApp{}).Error
}

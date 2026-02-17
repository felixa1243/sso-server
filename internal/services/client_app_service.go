package services

import (
	"context"
	"errors"
	"sso-server/internal/helper"
	"sso-server/internal/models"
	"sso-server/internal/repositories"

	"github.com/google/uuid"
)

type ClientAppService interface {
	RegisterClient(ctx context.Context, userID string, name string, redirectURIs string, scopes string) (*models.ClientApp, string, error)
	GetClient(ctx context.Context, clientID string) (*models.ClientApp, error)
	ListClients(ctx context.Context, userID string) ([]models.ClientApp, error)
	DeleteClient(ctx context.Context, clientID string, userID string) error
	ValidateClient(ctx context.Context, clientID string, clientSecret string) (*models.ClientApp, error)
}

type clientAppServiceImpl struct {
	clientRepo repositories.ClientAppRepository
}

func NewClientAppService(clientRepo repositories.ClientAppRepository) ClientAppService {
	return &clientAppServiceImpl{
		clientRepo: clientRepo,
	}
}

func (s *clientAppServiceImpl) RegisterClient(ctx context.Context, userID string, name string, redirectURIs string, scopes string) (*models.ClientApp, string, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, "", errors.New("invalid user id")
	}

	// Generate Secret
	secret := uuid.New().String() // Simple secret
	hashedSecret := helper.GeneratePassword(secret)

	client := &models.ClientApp{
		ID:            uuid.New(),
		Name:          name,
		Secret:        hashedSecret,
		RedirectURIs:  redirectURIs,
		AllowedScopes: scopes,
		UserID:        userUUID,
	}

	if err := s.clientRepo.Create(client); err != nil {
		return nil, "", err
	}

	// Return the plain secret only on creation
	return client, secret, nil
}

func (s *clientAppServiceImpl) GetClient(ctx context.Context, clientID string) (*models.ClientApp, error) {
	return s.clientRepo.FindByID(clientID)
}

func (s *clientAppServiceImpl) ListClients(ctx context.Context, userID string) ([]models.ClientApp, error) {
	return s.clientRepo.FindAllByUserID(userID)
}

func (s *clientAppServiceImpl) DeleteClient(ctx context.Context, clientID string, userID string) error {
	client, err := s.clientRepo.FindByID(clientID)
	if err != nil {
		return err
	}
	if client.UserID.String() != userID {
		return errors.New("unauthorized")
	}
	return s.clientRepo.Delete(clientID)
}

func (s *clientAppServiceImpl) ValidateClient(ctx context.Context, clientID string, clientSecret string) (*models.ClientApp, error) {
	client, err := s.clientRepo.FindByID(clientID)
	if err != nil {
		return nil, errors.New("client not found")
	}
	if !helper.ComparePassword(client.Secret, clientSecret) {
		return nil, errors.New("invalid client secret")
	}
	return client, nil
}

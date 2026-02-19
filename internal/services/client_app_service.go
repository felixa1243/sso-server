package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sso-server/internal/dto"
	"sso-server/internal/models"
	"sso-server/internal/repositories"
	"strings"
	"time"

	"github.com/google/uuid"
)

type ClientAppService interface {
	RegisterClient(ctx context.Context, userID string, req dto.RegisterClientRequest) (*models.ClientApp, error)
	ListClients(ctx context.Context, userID string) ([]models.ClientApp, error)
	GetClient(ctx context.Context, userID string, clientID string) (*models.ClientApp, error)
	GetClientByClientID(ctx context.Context, clientID string) (*models.ClientApp, error)
}

type clientAppServiceImpl struct {
	repo            repositories.ClientAppRepository
	scopeRepository repositories.ScopeRepository
}

func NewClientAppService(repo repositories.ClientAppRepository, scopeRepository repositories.ScopeRepository) ClientAppService {
	return &clientAppServiceImpl{repo: repo, scopeRepository: scopeRepository}
}

func (s *clientAppServiceImpl) RegisterClient(ctx context.Context, userID string, req dto.RegisterClientRequest) (*models.ClientApp, error) {
	clientID, err := generateRandomString(32)
	if err != nil {
		return nil, err
	}
	clientSecret, err := generateRandomString(64)
	if err != nil {
		return nil, err
	}
	scopeStrings := strings.Split(req.Scopes, ",")
	scopes, err := s.scopeRepository.FindByNames(scopeStrings)
	if err != nil {
		return nil, err
	}
	app := &models.ClientApp{
		ID:           uuid.New(),
		Name:         req.Name,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURIs: req.RedirectURIs,
		Scopes:       scopes,
		UserID:       uuid.MustParse(userID),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.repo.Create(ctx, app); err != nil {
		return nil, err
	}

	return app, nil
}

func (s *clientAppServiceImpl) ListClients(ctx context.Context, userID string) ([]models.ClientApp, error) {
	return s.repo.FindByUserID(ctx, userID)
}

func (s *clientAppServiceImpl) GetClient(ctx context.Context, userID string, clientID string) (*models.ClientApp, error) {
	app, err := s.repo.FindByClientID(ctx, clientID)
	if err != nil {
		return nil, err
	}
	if app.UserID.String() != userID {
		return nil, errors.New("unauthorized")
	}
	return app, nil
}

func (s *clientAppServiceImpl) GetClientByClientID(ctx context.Context, clientID string) (*models.ClientApp, error) {
	return s.repo.FindByClientID(ctx, clientID)
}

func generateRandomString(n int) (string, error) {
	bytes := make([]byte, n/2)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

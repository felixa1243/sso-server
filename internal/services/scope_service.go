package services

import (
	"context"
	"sso-server/internal/models"
	"sso-server/internal/repositories"

	"github.com/google/uuid"
)

type ScopeService interface {
	ListScopes(ctx context.Context) ([]models.Scope, error)
	GetScope(ctx context.Context, id string) (*models.Scope, error)
	CreateScope(ctx context.Context, name string, description string) (*models.Scope, error)
	UpdateScope(ctx context.Context, id string, name string, description string) (*models.Scope, error)
	DeleteScope(ctx context.Context, id string) error
}

type scopeServiceImpl struct {
	scopeRepo repositories.ScopeRepository
}

func NewScopeService(scopeRepo repositories.ScopeRepository) ScopeService {
	return &scopeServiceImpl{
		scopeRepo: scopeRepo,
	}
}

func (s *scopeServiceImpl) ListScopes(ctx context.Context) ([]models.Scope, error) {
	return s.scopeRepo.FindAll()
}

func (s *scopeServiceImpl) GetScope(ctx context.Context, id string) (*models.Scope, error) {
	return s.scopeRepo.FindByID(id)
}

func (s *scopeServiceImpl) CreateScope(ctx context.Context, name string, description string) (*models.Scope, error) {
	scope := &models.Scope{
		ID:          uuid.New(),
		Name:        name,
		Description: description,
	}
	if err := s.scopeRepo.Create(scope); err != nil {
		return nil, err
	}
	return scope, nil
}

func (s *scopeServiceImpl) UpdateScope(ctx context.Context, id string, name string, description string) (*models.Scope, error) {
	scope, err := s.scopeRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	scope.Name = name
	scope.Description = description
	if err := s.scopeRepo.Update(scope); err != nil {
		return nil, err
	}
	return scope, nil
}

func (s *scopeServiceImpl) DeleteScope(ctx context.Context, id string) error {
	return s.scopeRepo.Delete(id)
}

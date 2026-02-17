package services

import (
	"context"
	"errors"
	"sso-server/internal/models"
	"sso-server/internal/repositories"
	"time"

	"github.com/google/uuid"
)

type DomainService interface {
	CreateDomain(ctx context.Context, name, url string, userID string) (*models.Domain, error)
	ListDomains(ctx context.Context, userID string) ([]models.Domain, error)
}

type domainServiceImpl struct {
	domainRepository repositories.DomainRepository
}

func NewDomainService(domainRepository repositories.DomainRepository) DomainService {
	return &domainServiceImpl{
		domainRepository: domainRepository,
	}
}

func (s *domainServiceImpl) CreateDomain(ctx context.Context, name, url string, userID string) (*models.Domain, error) {
	if name == "" || url == "" {
		return nil, errors.New("name and url are required")
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("invalid user id")
	}

	existing, _ := s.domainRepository.FindByName(name)
	if existing != nil {
		return nil, errors.New("domain already exists")
	}

	domain := &models.Domain{
		ID:        uuid.New(),
		Name:      name,
		URL:       url,
		UserID:    uid,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.domainRepository.Create(domain); err != nil {
		return nil, err
	}

	return domain, nil
}

func (s *domainServiceImpl) ListDomains(ctx context.Context, userID string) ([]models.Domain, error) {
	return s.domainRepository.FindAllByUserID(userID)
}

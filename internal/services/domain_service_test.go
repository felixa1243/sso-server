package services_test

import (
	"context"
	"errors"
	"sso-server/internal/models"
	"sso-server/internal/services"
	"testing"
)

type mockDomainRepository struct {
	domains map[string]*models.Domain
}

func newMockDomainRepository() *mockDomainRepository {
	return &mockDomainRepository{
		domains: make(map[string]*models.Domain),
	}
}

func (m *mockDomainRepository) FindByName(name string) (*models.Domain, error) {
	if d, ok := m.domains[name]; ok {
		return d, nil
	}
	return nil, errors.New("not found")
}

func (m *mockDomainRepository) FindAll() ([]models.Domain, error) {
	var list []models.Domain
	for _, d := range m.domains {
		list = append(list, *d)
	}
	return list, nil
}

func (m *mockDomainRepository) FindAllByUserID(userID string) ([]models.Domain, error) {
	var list []models.Domain
	for _, d := range m.domains {
		if d.UserID.String() == userID {
			list = append(list, *d)
		}
	}
	return list, nil
}

func (m *mockDomainRepository) Create(domain *models.Domain) error {
	if _, ok := m.domains[domain.Name]; ok {
		return errors.New("already exists")
	}
	m.domains[domain.Name] = domain
	return nil
}

func (m *mockDomainRepository) Delete(id string) error {
	return nil
}

func (m *mockDomainRepository) Update(domain *models.Domain) error {
	m.domains[domain.Name] = domain
	return nil
}

func TestCreateDomain(t *testing.T) {
	repo := newMockDomainRepository()
	service := services.NewDomainService(repo)
	ctx := context.Background()
	uid := "123e4567-e89b-12d3-a456-426614174000"

	t.Run("Create new domain", func(t *testing.T) {
		domain, err := service.CreateDomain(ctx, "example", "https://example.com", uid)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if domain.Name != "example" {
			t.Errorf("expected name example, got %s", domain.Name)
		}
		if domain.UserID.String() != uid {
			t.Errorf("expected user id %s, got %s", uid, domain.UserID.String())
		}
	})

	t.Run("Create existing domain", func(t *testing.T) {
		_, err := service.CreateDomain(ctx, "example", "https://example.com", uid)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("Create with empty name", func(t *testing.T) {
		_, err := service.CreateDomain(ctx, "", "https://example.com", uid)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestListDomains(t *testing.T) {
	repo := newMockDomainRepository()
	service := services.NewDomainService(repo)
	ctx := context.Background()
	uid1 := "123e4567-e89b-12d3-a456-426614174000"
	uid2 := "223e4567-e89b-12d3-a456-426614174000"

	service.CreateDomain(ctx, "d1", "u1", uid1)
	service.CreateDomain(ctx, "d2", "u2", uid2)

	t.Run("List user 1 domains", func(t *testing.T) {
		domains, err := service.ListDomains(ctx, uid1)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(domains) != 1 {
			t.Errorf("expected 1 domain, got %d", len(domains))
		}
		if domains[0].Name != "d1" {
			t.Errorf("expected d1, got %s", domains[0].Name)
		}
	})

	t.Run("List user 2 domains", func(t *testing.T) {
		domains, err := service.ListDomains(ctx, uid2)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(domains) != 1 {
			t.Errorf("expected 1 domain, got %d", len(domains))
		}
		if domains[0].Name != "d2" {
			t.Errorf("expected d2, got %s", domains[0].Name)
		}
	})
}

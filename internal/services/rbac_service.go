package services

import (
	"context"
	"errors"
	"sso-server/internal/models"
	"sso-server/internal/repositories"

	"github.com/google/uuid"
)

type RBACService interface {
	CreateRole(ctx context.Context, name string, domainID *string) (*models.Role, error)
	UpdateRole(ctx context.Context, id string, name string, domainID *string) (*models.Role, error)
	DeleteRole(ctx context.Context, id string) error
	ListRoles(ctx context.Context) ([]models.Role, error)
	GetRole(ctx context.Context, id string) (*models.Role, error)

	CreatePermission(ctx context.Context, name string, slug string) (*models.Permission, error)
	ListPermissions(ctx context.Context) ([]models.Permission, error)

	AssignPermissions(ctx context.Context, roleID string, permissionIDs []string) error
}

type rbacServiceImpl struct {
	roleRepo       repositories.RoleRepository
	permissionRepo repositories.PermissionRepository
	eventService   EventService
}

func NewRBACService(
	roleRepo repositories.RoleRepository,
	permissionRepo repositories.PermissionRepository,
	eventService EventService,
) RBACService {
	return &rbacServiceImpl{
		roleRepo:       roleRepo,
		permissionRepo: permissionRepo,
		eventService:   eventService,
	}
}

func (s *rbacServiceImpl) CreateRole(ctx context.Context, name string, domainID *string) (*models.Role, error) {
	var domUUID *uuid.UUID
	if domainID != nil && *domainID != "" {
		parsed, err := uuid.Parse(*domainID)
		if err != nil {
			return nil, errors.New("invalid domain id")
		}
		domUUID = &parsed
	}

	role := &models.Role{
		ID:       uuid.New(),
		Name:     name,
		DomainID: domUUID,
	}

	if err := s.roleRepo.Create(role); err != nil {
		return nil, err
	}

	s.eventService.Publish(ctx, "role.created", role)
	return role, nil
}

func (s *rbacServiceImpl) UpdateRole(ctx context.Context, id string, name string, domainID *string) (*models.Role, error) {
	role, err := s.roleRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if name != "" {
		role.Name = name
	}

	if domainID != nil {
		if *domainID == "" {
			role.DomainID = nil
		} else {
			parsed, err := uuid.Parse(*domainID)
			if err != nil {
				return nil, errors.New("invalid domain id")
			}
			role.DomainID = &parsed
		}
	}

	if err := s.roleRepo.Update(role); err != nil {
		return nil, err
	}

	s.eventService.Publish(ctx, "role.updated", role)
	return role, nil
}

func (s *rbacServiceImpl) DeleteRole(ctx context.Context, id string) error {
	if err := s.roleRepo.Delete(id); err != nil {
		return err
	}
	s.eventService.Publish(ctx, "role.deleted", map[string]string{"id": id})
	return nil
}

func (s *rbacServiceImpl) ListRoles(ctx context.Context) ([]models.Role, error) {
	return s.roleRepo.FindAll()
}

func (s *rbacServiceImpl) GetRole(ctx context.Context, id string) (*models.Role, error) {
	return s.roleRepo.FindByID(id)
}

func (s *rbacServiceImpl) CreatePermission(ctx context.Context, name string, slug string) (*models.Permission, error) {
	perm := &models.Permission{
		ID:   uuid.New(),
		Name: name,
		Slug: slug,
	}
	if err := s.permissionRepo.Create(perm); err != nil {
		return nil, err
	}
	return perm, nil
}

func (s *rbacServiceImpl) ListPermissions(ctx context.Context) ([]models.Permission, error) {
	return s.permissionRepo.FindAll()
}

func (s *rbacServiceImpl) AssignPermissions(ctx context.Context, roleID string, permissionIDs []string) error {
	role, err := s.roleRepo.FindByID(roleID)
	if err != nil {
		return err
	}

	perms, err := s.permissionRepo.FindByIds(permissionIDs)
	if err != nil {
		return err
	}

	if err := s.roleRepo.UpdatePermissions(role, perms); err != nil {
		return err
	}

	s.eventService.Publish(ctx, "role.permissions_updated", map[string]interface{}{
		"role_id": roleID,
		"permissions": perms,
	})

	return nil
}

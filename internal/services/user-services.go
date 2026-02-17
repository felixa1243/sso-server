package services

import (
	"context"
	"crypto/rsa"
	"errors"
	"sso-server/internal/dto"
	"sso-server/internal/helper"
	"sso-server/internal/models"
	"sso-server/internal/repositories"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

type UserService interface {
	CreateUser(ctx context.Context, req dto.RegisterRequest, roleName string) (*models.User, error)
	Authenticate(ctx context.Context, email, password string) (*models.User, error)
	ExchangeCode(ctx context.Context, code string) (*models.User, string, error)
	GenerateAuthCode(ctx context.Context, userID string, scope string) (string, error)
	GetToken(ctx context.Context, userID string, domainName string, scope string) (string, error)
	GetUser(ctx context.Context, userID string) (*dto.JoinUser, error)
	Logout(ctx context.Context, token string) error
	ChangePassword(ctx context.Context, userID string, req dto.ChangePasswordRequest) error
	ListUsers(ctx context.Context) ([]models.User, error)
	DeleteUser(ctx context.Context, userID string) error
	UpdateUserRoles(ctx context.Context, userID string, roleIDs []string) error
	BanUser(ctx context.Context, userID string) error
	UnbanUser(ctx context.Context, userID string) error
}

type userServiceImpl struct {
	userRepository       repositories.UserRepository
	profileRepository    repositories.ProfileRepository
	roleRepository       repositories.RoleRepository
	domainRepository     repositories.DomainRepository
	punishmentRepository repositories.PunishmentRepository
	eventService         EventService
	redis                *redis.Client
	privateKey           *rsa.PrivateKey
}

func NewUserServices(
	userRepository repositories.UserRepository,
	profileRepository repositories.ProfileRepository,
	roleRepository repositories.RoleRepository,
	domainRepository repositories.DomainRepository,
	punishmentRepository repositories.PunishmentRepository,
	eventService EventService,
	redis *redis.Client,
	privateKey *rsa.PrivateKey,
) UserService {
	return &userServiceImpl{
		userRepository:       userRepository,
		profileRepository:    profileRepository,
		roleRepository:       roleRepository,
		domainRepository:     domainRepository,
		punishmentRepository: punishmentRepository,
		eventService:         eventService,
		redis:                redis,
		privateKey:           privateKey,
	}
}

func (u *userServiceImpl) CreateUser(ctx context.Context, req dto.RegisterRequest, roleName string) (*models.User, error) {
	if req.Password != req.PasswordConfirm {
		return nil, errors.New("passwords do not match")
	}

	role, err := u.roleRepository.FindByName(roleName)
	if err != nil {
		return nil, err
	}

	existingUser, _ := u.userRepository.FindByEmail(req.Email)
	if existingUser != nil {
		return nil, errors.New("user already exists")
	}

	userCreated := models.User{
		ID:           uuid.New(),
		Email:        req.Email,
		PasswordHash: helper.GeneratePassword(req.Password),
		Role:         []models.Role{*role},
	}

	err = u.userRepository.Create(ctx, &userCreated)
	if err != nil {
		return nil, err
	}

	return &userCreated, nil
}

func (u *userServiceImpl) Authenticate(ctx context.Context, email, password string) (*models.User, error) {
	user, err := u.userRepository.FindByEmail(email)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if !helper.ComparePassword(user.PasswordHash, password) {
		return nil, errors.New("incorrect password")
	}

	punishment, _ := u.punishmentRepository.FindActiveByUserID(user.ID.String(), "BAN")
	if punishment != nil {
		return nil, errors.New("user is banned: " + punishment.Reason)
	}

	return user, nil
}

func (u *userServiceImpl) ExchangeCode(ctx context.Context, code string) (*models.User, string, error) {
	val, err := u.redis.Get(ctx, "auth_code:"+code).Result()
	if err != nil {
		return nil, "", errors.New("code expired or invalid")
	}

	parts := strings.SplitN(val, ":", 2)
	userID := parts[0]
	scope := ""
	if len(parts) > 1 {
		scope = parts[1]
	}

	user, err := u.userRepository.FindByID(userID)
	if err != nil {
		return nil, "", errors.New("user not found")
	}

	u.redis.Del(ctx, "auth_code:"+code)
	return user, scope, nil
}

func (u *userServiceImpl) GenerateAuthCode(ctx context.Context, userID string, scope string) (string, error) {
	authCode := uuid.New().String()
	val := userID
	if scope != "" {
		val = userID + ":" + scope
	}
	err := u.redis.Set(ctx, "auth_code:"+authCode, val, 5*time.Minute).Err()
	if err != nil {
		return "", errors.New("failed to store session")
	}
	return authCode, nil
}

func (u *userServiceImpl) Logout(ctx context.Context, token string) error {
	if token != "" {
		err := u.redis.Del(ctx, "access_token:"+token).Err()
		if err != nil {
			return errors.New("failed to invalidate session in cache")
		}
	}
	return nil
}

func (u *userServiceImpl) ChangePassword(ctx context.Context, userID string, req dto.ChangePasswordRequest) error {
	user, err := u.userRepository.FindByID(userID)
	if err != nil {
		return err
	}

	if !helper.ComparePassword(user.PasswordHash, req.OldPassword) {
		return errors.New("incorrect password")
	}

	if req.NewPassword != req.NewPasswordConfirm {
		return errors.New("passwords do not match")
	}

	user.PasswordHash = helper.GeneratePassword(req.NewPassword)
	return u.userRepository.Update(user)
}

func (u *userServiceImpl) ListUsers(ctx context.Context) ([]models.User, error) {
	return u.userRepository.FindAll()
}

func (u *userServiceImpl) DeleteUser(ctx context.Context, userID string) error {
	return u.userRepository.Delete(userID)
}

func (u *userServiceImpl) UpdateUserRoles(ctx context.Context, userID string, roleIDs []string) error {
	user, err := u.userRepository.FindByID(userID)
	if err != nil {
		return err
	}

	roles, err := u.roleRepository.FindByIds(roleIDs)
	if err != nil {
		return err
	}

	if err := u.userRepository.UpdateRoles(user, roles); err != nil {
		return err
	}

	u.eventService.Publish(ctx, "user.promoted", map[string]interface{}{
		"user_id": userID,
		"roles": roles,
	})

	return nil
}

func (u *userServiceImpl) BanUser(ctx context.Context, userID string) error {
	// Simple ban implementation that creates an indefinite punishment
	punishment := models.Punishment{
		ID:        uuid.New(),
		UserID:    uuid.MustParse(userID),
		Type:      "BAN",
		Reason:    "Banned by admin",
		StartTime: time.Now(),
		EndTime:   time.Now().Add(24 * 365 * 100 * time.Hour), // Indefinite
		AdminID:   uuid.Nil, // Unknown admin
	}

	if err := u.punishmentRepository.Create(&punishment); err != nil {
		return err
	}

	// Set banned status in Redis for immediate effect in middleware
	return u.redis.Set(ctx, "user:"+userID+":banned", "true", 0).Err()
}

func (u *userServiceImpl) UnbanUser(ctx context.Context, userID string) error {
	// Revoke active bans
	punishment, _ := u.punishmentRepository.FindActiveByUserID(userID, "BAN")
	if punishment != nil {
		u.punishmentRepository.Revoke(punishment.ID.String())
	}

	// Remove banned status from Redis
	return u.redis.Del(ctx, "user:"+userID+":banned").Err()
}

func (u *userServiceImpl) GetUser(ctx context.Context, userID string) (*dto.JoinUser, error) {
	joinUser, queryError := u.profileRepository.FindByUserID(userID)
	if queryError != nil {
		return nil, errors.New("profile record not found")
	}
	return &joinUser, nil
}

func (u *userServiceImpl) GetToken(ctx context.Context, userID string, domainName string, scope string) (string, error) {
	joinUser, queryError := u.GetUser(ctx, userID)
	if queryError != nil {
		return "", queryError
	}
	user, err := u.userRepository.FindByID(userID)
	if err != nil {
		return "", errors.New("user account does not exist")
	}

	var domain *models.Domain
	if domainName != "" {
		d, err := u.domainRepository.FindByName(domainName)
		if err == nil {
			domain = d
		}
	}

	fullname := joinUser.Fullname
	token, err := helper.GenerateToken(user, fullname, u.privateKey, domain, scope)
	if err != nil {
		return "", errors.New("security signing failed")
	}
	err = u.redis.Set(ctx, "access_token:"+token, user.ID.String(), 24*time.Hour).Err()
	if err != nil {
		return "", errors.New("session storage failed")
	}
	return token, nil
}

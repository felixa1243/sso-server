package services

import (
	"context"
	"crypto/rsa"
	"errors"
	"sso-server/internal/dto"
	"sso-server/internal/helper"
	"sso-server/internal/models"
	"sso-server/internal/repositories"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

type UserService interface {
	CreateUser(ctx context.Context, req dto.RegisterRequest, roleName string) (*models.User, error)
	Authenticate(ctx context.Context, email, password string) (*models.User, error)
	ExchangeCode(ctx context.Context, code string) (*models.User, error)
	GenerateAuthCode(ctx context.Context, userID string) (string, error)
	GetToken(ctx context.Context, userID string) (string, error)
	GetUser(ctx context.Context, userID string) (*dto.JoinUser, error)
	Logout(ctx context.Context, token string) error
	ChangePassword(ctx context.Context, userID string, req dto.ChangePasswordRequest) error
}

type userServiceImpl struct {
	userRepository    repositories.UserRepository
	profileRepository repositories.ProfileRepository
	roleRepository    repositories.RoleRepository
	redis             *redis.Client
	privateKey        *rsa.PrivateKey
}

func NewUserServices(
	userRepository repositories.UserRepository,
	profileRepository repositories.ProfileRepository,
	roleRepository repositories.RoleRepository,
	redis *redis.Client,
	privateKey *rsa.PrivateKey,
) UserService {
	return &userServiceImpl{
		userRepository:    userRepository,
		profileRepository: profileRepository,
		roleRepository:    roleRepository,
		redis:             redis,
		privateKey:        privateKey,
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

	return user, nil
}

func (u *userServiceImpl) ExchangeCode(ctx context.Context, code string) (*models.User, error) {
	userID, err := u.redis.Get(ctx, "auth_code:"+code).Result()
	if err != nil {
		return nil, errors.New("code expired or invalid")
	}

	user, err := u.userRepository.FindByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Fetch roles to include them, similar to original controller logic which used Preload("Role")
	// The repository FindByID doesn't preload roles explicitly in my reading of it?
	// Let's check user-repository. FindByID: r.db.Where("id = ?", id).First(&user)
	// It does NOT Preload roles.
	// But `models.User` has `Role []Role`.
	// I should probably ensure roles are loaded if needed.
	// The original controller did: ac.DB.Preload("Role").Where("id = ?", userID).First(&user)
	// So I need to update UserRepository.FindByID to Preload roles or add a method for it.
	// Or I can just return the user and let the repository handle loading if configured (it's not by default).

	// I'll update UserRepository.FindByID to Preload("Role") as it seems essential for this app (RBAC).
	// But first, finish this method.

	u.redis.Del(ctx, "auth_code:"+code)
	return user, nil
}

func (u *userServiceImpl) GenerateAuthCode(ctx context.Context, userID string) (string, error) {
	authCode := uuid.New().String()
	err := u.redis.Set(ctx, "auth_code:"+authCode, userID, 5*time.Minute).Err()
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

func (u *userServiceImpl) GetUser(ctx context.Context, userID string) (*dto.JoinUser, error) {
	joinUser, queryError := u.profileRepository.FindByUserID(userID)
	if queryError != nil {
		return nil, errors.New("profile record not found")
	}
	return &joinUser, nil
}

func (u *userServiceImpl) GetToken(ctx context.Context, userID string) (string, error) {
	joinUser, queryError := u.GetUser(ctx, userID)
	if queryError != nil {
		return "", queryError
	}
	user, err := u.userRepository.FindByID(userID)
	if err != nil {
		return "", errors.New("user account does not exist")
	}
	fullname := joinUser.Fullname
	token, err := helper.GenerateToken(user, fullname, u.privateKey)
	if err != nil {
		return "", errors.New("security signing failed")
	}
	err = u.redis.Set(ctx, "access_token:"+token, user.ID.String(), 24*time.Hour).Err()
	if err != nil {
		return "", errors.New("session storage failed")
	}
	return token, nil
}

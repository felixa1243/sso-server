package services

import (
	"context"
	"errors"
	"sso-server/internal/models"
	"sso-server/internal/repositories"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

type PunishmentService interface {
	PunishUser(ctx context.Context, adminID string, req PunishRequest) (*models.Punishment, error)
	RevokePunishment(ctx context.Context, punishmentID string) error
	GetPunishments(ctx context.Context, userID string) ([]models.Punishment, error)
	UnbanUser(ctx context.Context, userID string) error
	BanUser(ctx context.Context, adminID string, userID string, reason string, durationSec int64) (*models.Punishment, error)
}

type PunishRequest struct {
	UserID      string `json:"user_id"`
	Type        string `json:"type"` // BAN, MUTE
	Reason      string `json:"reason"`
	DurationSec int64  `json:"duration_sec"`
}

type punishmentServiceImpl struct {
	punishmentRepo repositories.PunishmentRepository
	userRepo       repositories.UserRepository
	eventService   EventService
	redis          *redis.Client
}

func NewPunishmentService(
	punishmentRepo repositories.PunishmentRepository,
	userRepo repositories.UserRepository,
	eventService EventService,
	redis *redis.Client,
) PunishmentService {
	return &punishmentServiceImpl{
		punishmentRepo: punishmentRepo,
		userRepo:       userRepo,
		eventService:   eventService,
		redis:          redis,
	}
}

func (s *punishmentServiceImpl) PunishUser(ctx context.Context, adminID string, req PunishRequest) (*models.Punishment, error) {
	userUUID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, errors.New("invalid user id")
	}
	adminUUID, err := uuid.Parse(adminID)
	if err != nil {
		return nil, errors.New("invalid admin id")
	}

	endTime := time.Now().Add(time.Duration(req.DurationSec) * time.Second)
	// If 0, assume infinite? Or default 1 day? Let's say 100 years for infinite.
	if req.DurationSec <= 0 {
		endTime = time.Now().Add(24 * 365 * 100 * time.Hour)
	}

	punishment := &models.Punishment{
		ID:        uuid.New(),
		UserID:    userUUID,
		AdminID:   adminUUID,
		Type:      req.Type,
		Reason:    req.Reason,
		StartTime: time.Now(),
		EndTime:   endTime,
	}

	if err := s.punishmentRepo.Create(punishment); err != nil {
		return nil, err
	}

	if req.Type == "BAN" {
		ttl := time.Until(endTime)
		if ttl < 0 {
			ttl = 0
		}
		// Convert duration to time.Duration
		s.redis.Set(ctx, "user:"+req.UserID+":banned", "true", ttl)

		// Update User IsBanned flag in DB as well for persistence
		user, err := s.userRepo.FindByID(req.UserID)
		if err == nil {
			user.IsBanned = true
			s.userRepo.Update(user)
		}
	}

	s.eventService.Publish(ctx, "user.punished", map[string]interface{}{
		"user_id":    req.UserID,
		"type":       req.Type,
		"reason":     req.Reason,
		"end_time":   endTime,
		"admin_id":   adminID,
		"punishment": punishment,
	})

	return punishment, nil
}

func (s *punishmentServiceImpl) RevokePunishment(ctx context.Context, punishmentID string) error {
	// Ideally check if exists
	// Then revoke
	if err := s.punishmentRepo.Revoke(punishmentID); err != nil {
		return err
	}
	return nil
}

func (s *punishmentServiceImpl) GetPunishments(ctx context.Context, userID string) ([]models.Punishment, error) {
	return s.punishmentRepo.FindAllByUserID(userID)
}

func (s *punishmentServiceImpl) UnbanUser(ctx context.Context, userID string) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}
	user.IsBanned = false
	if err := s.userRepo.Update(user); err != nil {
		return err
	}

	s.redis.Del(ctx, "user:"+userID+":banned")

	// Revoke active bans
	punishment, _ := s.punishmentRepo.FindActiveByUserID(userID, "BAN")
	if punishment != nil {
		s.punishmentRepo.Revoke(punishment.ID.String())
	}

	s.eventService.Publish(ctx, "user.unbanned", map[string]interface{}{
		"user_id": userID,
	})

	return nil
}

func (s *punishmentServiceImpl) BanUser(ctx context.Context, adminID string, userID string, reason string, durationSec int64) (*models.Punishment, error) {
	req := PunishRequest{
		UserID:      userID,
		Type:        "BAN",
		Reason:      reason,
		DurationSec: durationSec,
	}
	return s.PunishUser(ctx, adminID, req)
}

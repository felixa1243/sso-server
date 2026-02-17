package services

import (
	"context"
	"encoding/json"

	"github.com/go-redis/redis/v8"
)

type EventService interface {
	Publish(ctx context.Context, channel string, message interface{}) error
}

type eventServiceImpl struct {
	redis *redis.Client
}

func NewEventService(redis *redis.Client) EventService {
	return &eventServiceImpl{redis: redis}
}

func (s *eventServiceImpl) Publish(ctx context.Context, channel string, message interface{}) error {
	msgBytes, err := json.Marshal(message)
	if err != nil {
		return err
	}
	return s.redis.Publish(ctx, channel, msgBytes).Err()
}

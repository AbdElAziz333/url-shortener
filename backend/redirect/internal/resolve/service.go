package resolve

import (
	"context"
	"errors"
	"fmt"
	"time"

	sqlcgen "aziz.dev/redirect/internal/postgres/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

var (
	ErrNotFound = errors.New("link not found")
	ErrLinkInactive = errors.New("link is not active or expired")
)

type RedisClient interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd
}

type Service interface {
	ResolveCode(ctx context.Context, code string) (*Dto, error)
}

type service struct {
	querier sqlcgen.Querier
	cache RedisClient
}

func NewService(querier sqlcgen.Querier, cache RedisClient) Service {
	return &service{
		querier: querier,
		cache:   cache,
	}
}

func (s *service) ResolveCode(ctx context.Context, code string) (*Dto, error) {
	log := logrus.WithField("code", code)
	log.Info("Resolving code")

	// L1: Redis cache
	url, err := s.cache.Get(ctx, code).Result()
	switch {
	case err == nil && url != "":
		log.Info("Cache hit")
		return &Dto{Code: code, OriginalURL: url}, nil
	case errors.Is(err, redis.Nil) || (err == nil && url == ""):
		log.Info("Cache miss")
	default:
		log.WithError(err).Warn("Cache lookup error")
	}

	// L2: database
	link, err := s.querier.FindByCode(ctx, code)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Warn("Code not found in database")
			return nil, ErrNotFound // map it, so handlers don't depend on pgx
		}
		log.WithError(err).Error("Database lookup error")
		return nil, fmt.Errorf("find by code: %w", err)
	}

	// Validate: must be active and not past its expiry.
	expiresAt := expiryInstant(link.ExpiresAt)
	ttl := time.Until(expiresAt)

	if !link.IsActive || ttl <= 0 {
		log.Warn("Link is inactive or expired")
		return nil, ErrLinkInactive
	}

	// Write-back to cache; TTL matches the link's remaining lifetime.
	// Best-effort: a cache failure must never break the redirect.
	if err := s.cache.Set(ctx, code, link.OriginalUrl, ttl).Err(); err != nil {
		log.WithError(err).Warn("Failed to write to cache")
	} else {
		log.Info("Successfully wrote to cache")
	}

	return &Dto{
		Code:        link.Code,
		OriginalURL: link.OriginalUrl,
	}, nil
}

// expiryInstant returns the first moment the link is no longer valid:
// midnight UTC at the start of the day AFTER the stored expiry date.
func expiryInstant(d time.Time) time.Time {
	y, m, day := d.UTC().Date()
	return time.Date(y, m, day+1, 0, 0, 0, 0, time.UTC)
}
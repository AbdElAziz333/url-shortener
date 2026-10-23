package link

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	sqlcgen "aziz.dev/shortener/internal/postgres/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrNotFound      = errors.New("link not found")
	ErrCodeTaken     = errors.New("code already in use")
	ErrInvalidExpiry = errors.New("expires_at must be in the future")
	ErrCodeGen       = errors.New("could not generate a unique code")
)

const (
	codeLength      = 8
	maxCodeAttempts = 5
	charset         = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

type Service interface {
	GetAll(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]Dto, error)
	Get(ctx context.Context, userID uuid.UUID, code string) (*Dto, error)
	Create(ctx context.Context, userID uuid.UUID, req CreateRequest) (*Dto, error)
	UpdateExpiry(ctx context.Context, userID uuid.UUID, code string, req UpdateExpiryRequest) (*Dto, error)
	Delete(ctx context.Context, userID uuid.UUID, code string) error
}

type service struct {
	q sqlcgen.Querier
}

func NewService(q sqlcgen.Querier) Service {
	return &service{
		q: q,
	}
}

func (s *service) GetAll(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]Dto, error) {
	links, err := s.q.ListLinksByUser(ctx, sqlcgen.ListLinksByUserParams{
		UserID: userID,
		Limit: limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list links: %w", err)
	}

	dtos := make([]Dto, 0, len(links)) // never nil -> JSON []
	for _, l := range links {
		dtos = append(dtos, toDto(l))
	}
	return dtos, nil
}

func (s *service) Get(ctx context.Context, userID uuid.UUID, code string) (*Dto, error) {
	row, err := s.q.GetLinkByCode(ctx, sqlcgen.GetLinkByCodeParams{Code: code, UserID: userID})

	if err != nil {
		return nil, wrapErr("get link", err)
	}

	return ptr(toDto(row)), nil
}

func (s *service) Create(ctx context.Context, userID uuid.UUID, req CreateRequest) (*Dto, error) {
	if req.ExpiresAt == nil {
		return nil, ErrInvalidExpiry // or apply a default, e.g. now + 30 days
	}

	expiresAt, err := normalizeExpiry(*req.ExpiresAt)
	if err !=  nil {
		return nil, fmt.Errorf("normalizing expiry: %w", err)
	}

	for i := 0; i < maxCodeAttempts; i++ {
		code, err := generateRandomCode(codeLength)
		if err != nil {
			return nil, fmt.Errorf("generate code: %w", err)
		}

		row, err := s.q.CreateLink(ctx, sqlcgen.CreateLinkParams{
			UserID:      userID,
			Code:        code,
			OriginalUrl: req.OriginalURL,
			ExpiresAt:   expiresAt,
		})
		
		if err == nil {
			return ptr(toDto(row)), nil
		}

		if !isUniqueViolation(err) {
			return nil, fmt.Errorf("create link: %w", err)
		}

		// code collision: try again with a new code
	}

	return nil, ErrCodeGen
}

func (s *service) UpdateExpiry(ctx context.Context, userID uuid.UUID, code string, req UpdateExpiryRequest) (*Dto, error) {
	if req.ExpiresAt == nil {
		return nil, ErrInvalidExpiry // or apply a default, e.g. now + 30 days
	}

	expiresAt, err := normalizeExpiry(*req.ExpiresAt)
	if err !=  nil {
		return nil, fmt.Errorf("normalizing expiry: %w", err)
	}

	row, err := s.q.UpdateLinkExpiry(ctx, sqlcgen.UpdateLinkExpiryParams{
		Code:      code,
		UserID:    userID,
		ExpiresAt: expiresAt,
	})

	if err != nil {
		return nil, wrapErr("update expiry", err)
	}

	return ptr(toDto(row)), nil
}

func (s *service) Delete(ctx context.Context, userID uuid.UUID, code string) error {
	n, err := s.q.DeactivateLink(ctx, sqlcgen.DeactivateLinkParams{Code: code, UserID: userID})
	if err != nil {
		return fmt.Errorf("delete link: %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func normalizeExpiry(t time.Time) (time.Time, error) {
	y, m, d := t.UTC().Date()
	expiry := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)

	ny, nm, nd := time.Now().UTC().Date()
	today := time.Date(ny, nm, nd, 0, 0, 0, 0, time.UTC)

	// Must be strictly in the future (tomorrow or later).
	// Use expiry.Before(today) instead if "today" should be allowed.
	if !expiry.After(today) {
		return time.Time{}, ErrInvalidExpiry
	}

	return expiry, nil
}

// wrapErr maps "no rows" to ErrNotFound and wraps anything else with context.
func wrapErr(op string, err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	
	return fmt.Errorf("%s: %w", op, err)
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func generateRandomCode(length int) (string, error) {
	b := make([]byte, length)
	max := big.NewInt(int64(len(charset)))
	for i := range b {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		b[i] = charset[n.Int64()]
	}
	return string(b), nil
}

func ptr[T any](v T) *T {
	return &v
}
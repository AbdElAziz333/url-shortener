package user

import (
	"context"
	"errors"
	"time"

	sqlcgen "aziz.dev/gateway/internal/postgres/sqlc"
	"aziz.dev/gateway/internal/security"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmailTaken         = errors.New("email already in use")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInactiveAccount    = errors.New("account is not active")
	ErrNotImplemented     = errors.New("not implemented")
)

type Service interface {
	Signup(ctx context.Context, r *SignupRequest) (error)
	Login(ctx context.Context, r *LoginRequest) (*security.TokenPair, error)
	Refresh(ctx context.Context) (*security.TokenPair, error)
	Me(ctx context.Context, id uuid.UUID) (*sqlcgen.FindByIDRow, error)
}

type service struct {
	querier sqlcgen.Querier
	jwtService security.Service
}

func NewService(querier sqlcgen.Querier, jwtService security.Service) Service {
	return &service{
		querier: querier,
		jwtService: jwtService,
	}
}

func (s *service) Signup(ctx context.Context, r *SignupRequest) error {
	_, err := s.querier.FindByEmail(ctx, r.Email)
	switch {
	case err == nil:
		return ErrEmailTaken
	case !errors.Is(err, pgx.ErrNoRows):
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(r.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	now := time.Now()

	return s.querier.CreateUser(ctx, sqlcgen.CreateUserParams{
		ID:           uuid.New(),
		Email:        r.Email,
		PasswordHash: string(hashedPassword),
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	})
}

func (s *service) Login(ctx context.Context, r *LoginRequest) (*security.TokenPair, error) {
	user, err := s.querier.FindByEmail(ctx, r.Email)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrInvalidCredentials
	}

	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(r.Password)); err != nil {
		logrus.WithField("user_id", user.ID).Warn("invalid password attempt")
		return nil, ErrInvalidCredentials
	}

	if !user.IsActive {
		return nil, ErrInactiveAccount
	}

	token, err := s.jwtService.GenerateToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &security.TokenPair{Token: token}, nil
}

// TODO: validate refresh token, issue a new token pair, store in redis.
func (s *service) Refresh(ctx context.Context) (*security.TokenPair, error) {
	return nil, ErrNotImplemented
}

func (s *service) Me(ctx context.Context, id uuid.UUID) (*sqlcgen.FindByIDRow, error)  {
	user, err := s.querier.FindByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrInvalidCredentials
	}

	if err != nil {
		return nil, err
	}

	return &user, nil
}
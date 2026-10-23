package security

import (
	"errors"
	"time"

	"aziz.dev/gateway/internal/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const issuer = "gateway"

type Service interface {
	GenerateToken(userID uuid.UUID) (string, error)
}

type Claim struct {
	UserID uuid.UUID `json:"user_id"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	Token  string `json:"token"`
}

type service struct {
	config *config.JWTConfig
}

func NewService(cfg *config.JWTConfig) (Service, error) {
	if len(cfg.Secret) == 0 {
		return nil, errors.New("jwt secret key is required")
	}

	if cfg.RefreshExpiry <= cfg.AccessExpiry {
		return nil, errors.New("refresh expiry must be greater than access expiry")
	}

	return &service{config: cfg}, nil
}

func (s *service) GenerateToken(userID uuid.UUID) (string, error) {
	return sign(userID, s.config.Secret, s.config.AccessExpiry)
}

func (s *service) GenerateRefreshToken(userID uuid.UUID) (string, error) {
	return sign(userID, s.config.Secret, s.config.RefreshExpiry)
}

func sign(userID uuid.UUID, secret []byte, ttl time.Duration) (string, error) {
	now := time.Now()

	claim := Claim{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    issuer,
			Subject:   userID.String(),
		},
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS256, claim).SignedString(secret)
}

// func ValidateAccessToken(cfg *config.JWTConfig, token string) (*Claim, error) {
// 	return validate(token, cfg.AccessSecret)
// }

func ValidateToken(cfg *config.JWTConfig, token string) (*Claim, error) {
	return validate(token, cfg.Secret)
}

func validate(token string, secret []byte) (*Claim, error) {
	claims := &Claim{}
	_, err := jwt.ParseWithClaims(token, claims, func(*jwt.Token) (any, error) {
		return secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return nil, err
	}
	return claims, nil
}
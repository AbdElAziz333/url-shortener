package user

import (
	"time"

	"github.com/google/uuid"
)

type LoginRequest struct {
	Email string `json:"email"`
	Password string `json:"password"`
}

type SignupRequest struct {
	Email string `json:"email"`
	Password string `json:"password"`
}

type Dto struct {
	ID           uuid.UUID
	Email        string
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

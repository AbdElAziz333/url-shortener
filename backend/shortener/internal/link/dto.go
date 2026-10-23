package link

import (
	"time"

	sqlcgen "aziz.dev/shortener/internal/postgres/sqlc"
)

type Dto struct {
	Code string `json:"code"`
	OriginalURL string `json:"original_url"`
	ExpiresAt time.Time `json:"expires_at"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
}

type CreateRequest struct {
	OriginalURL string     `json:"original_url" binding:"required,url"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

type UpdateExpiryRequest struct {
	ExpiresAt *time.Time `json:"expires_at" binding:"required"`
}

func toDto(l sqlcgen.Link) Dto {
	return Dto{
		Code:        l.Code,
		OriginalURL: l.OriginalUrl,
		ExpiresAt:   l.ExpiresAt,
		IsActive:    l.IsActive,
		CreatedAt:   l.CreatedAt,
	}
}
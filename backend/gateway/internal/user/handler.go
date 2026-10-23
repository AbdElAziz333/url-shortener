package user

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const (
	refreshCookieName   = "refresh_token"
	refreshCookieMaxAge = 60 * 60 * 24 * 7 // seconds
	cookieDomain        = "localhost"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	authRoutes := r.Group("/auth")
	authRoutes.POST("/signup", h.Signup)
	authRoutes.POST("/login", h.Login)
	authRoutes.POST("/refresh", h.Refresh)
	authRoutes.POST("/logout", h.Logout)
}

func (h *Handler) Signup(c *gin.Context) {
	var req SignupRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.Signup(c.Request.Context(), &req); err != nil {
		respondError(c, err)
		return
	}

	c.Status(http.StatusCreated)
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tokens, err := h.service.Login(c.Request.Context(), &req)
	if err != nil {
		respondError(c, err)
		return
	}

	setTokens(c, tokens.Token)
}

func (h *Handler) Refresh(c *gin.Context) {
	tokens, err := h.service.Refresh(c.Request.Context())

	if err != nil {
		respondError(c, err)
		return
	}

	setTokens(c, tokens.Token)
}

func (h *Handler) Logout(c *gin.Context) {
	c.SetCookie(refreshCookieName, "", -1, "/", cookieDomain, false, true)
	c.Status(http.StatusOK)
}

func setTokens(c *gin.Context, token string) {
	c.Header("access_token", token)
	c.SetCookie(refreshCookieName, token, refreshCookieMaxAge, "/", cookieDomain, false, true)
	c.Status(http.StatusOK)
}

// respondError maps service errors to HTTP responses without leaking internals.
func respondError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrEmailTaken):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, ErrInvalidCredentials):
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
	case errors.Is(err, ErrInactiveAccount):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, ErrNotImplemented):
		c.JSON(http.StatusNotImplemented, gin.H{"error": err.Error()})
	default:
		logrus.WithError(err).WithField("path", c.Request.URL.Path).Error("request failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
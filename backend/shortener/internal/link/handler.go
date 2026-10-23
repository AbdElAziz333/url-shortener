package link

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	links := r.Group("/links")
	links.GET("", h.GetAll)
	links.GET("/:code", h.Get)
	links.POST("", h.Create)
	links.PATCH("/:code/expiry", h.UpdateExpiry)
	links.DELETE("/:code", h.Delete)
}

func (h *Handler) GetAll(c *gin.Context) {
	userID, err := uuid.Parse(c.GetHeader("User-ID"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or missing user id"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	links, err := h.service.GetAll(c.Request.Context(), userID, int32(limit), int32(offset))
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, ErrInvalidExpiry):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		default:
			logrus.WithError(err).WithField("user_id", userID).Error("request failed")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": links})
}

func (h *Handler) Get(c *gin.Context) {
	userID, err := uuid.Parse(c.GetHeader("User-ID"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or missing user id"})
		return
	}

	link, err := h.service.Get(c.Request.Context(), userID, c.Param("code"))
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, ErrInvalidExpiry):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		default:
			logrus.WithError(err).WithField("user_id", userID).Error("request failed")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": link})
}

func (h *Handler) Create(c *gin.Context) {
	userID, err := uuid.Parse(c.GetHeader("User-ID"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or missing user id"})
		return
	}

	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dto, err := h.service.Create(c.Request.Context(), userID, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, ErrInvalidExpiry):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		default:
			logrus.WithError(err).WithField("user_id", userID).Error("request failed")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": dto})
}

func (h *Handler) UpdateExpiry(c *gin.Context) {
	userID, err := uuid.Parse(c.GetHeader("User-ID"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or missing user id"})
		return
	}

	var req UpdateExpiryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dto, err := h.service.UpdateExpiry(c.Request.Context(), userID, c.Param("code"), req)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, ErrInvalidExpiry):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		default:
			logrus.WithError(err).WithField("user_id", userID).Error("request failed")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": dto})
}

func (h *Handler) Delete(c *gin.Context) {
	userID, err := uuid.Parse(c.GetHeader("User-ID"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or missing user id"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), userID, c.Param("code")); err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, ErrInvalidExpiry):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		default:
			logrus.WithError(err).WithField("user_id", userID).Error("request failed")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.Status(http.StatusNoContent)
}
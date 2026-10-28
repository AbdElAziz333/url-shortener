package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func NewRouter() *gin.Engine {
// func NewRouter(linkHandler *link.Handler) *gin.Engine {
	router := gin.New()
	router.Use(
		gin.Recovery(),
		gin.Logger(),
	)

	router.GET("/shortener/health", healthHandler)
	// shortenerGroup := router.Group("/shortener")
	// shortenerGroup.GET("/health", healthHandler)

	// linkGroup := shortenerGroup.Group("/api")
	// linkHandler.RegisterRoutes(linkGroup)

	return router
}

func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "OK",
	})
}

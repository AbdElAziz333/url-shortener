package server

import (
	"net/http"

	"aziz.dev/gateway/internal/config"
	"aziz.dev/gateway/internal/grpcclient"
	"aziz.dev/gateway/internal/middleware"
	"aziz.dev/gateway/internal/user"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func NewRouter(userHandler *user.Handler, jwtConfig *config.JWTConfig, grpcHandler *grpcclient.Handler) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery(), gin.Logger())

	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:5173"},
		AllowMethods: []string{"GET", "POST", "PATCH", "PUT", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization", "Access-Control-Allow-Origin"},
		ExposeHeaders: []string{"Content-Length", "access_token"},
		AllowCredentials: true,
	}))

	router.GET("/health", healthHandler)

	authRoutes := router.Group("/auth")
    authRoutes.POST("/signup", userHandler.Signup)
    authRoutes.POST("/login", userHandler.Login)
    authRoutes.POST("/refresh", userHandler.Refresh)
    authRoutes.POST("/logout", userHandler.Logout)

    // Protected auth routes (like /me)
    protectedAuth := router.Group("/auth", middleware.AccessTokenMiddleware(jwtConfig))
    protectedAuth.GET("/me", userHandler.Me)

	// shortener microservice endpoints
	links := router.Group("/shortener/api/links", middleware.AccessTokenMiddleware(jwtConfig))
	links.GET("", grpcHandler.ListLinks)
	links.GET("/:code", grpcHandler.GetLink)
	links.POST("", grpcHandler.CreateLink)
	links.PATCH("/:code/expiry", grpcHandler.UpdateExpiry)
	links.DELETE("/:code", grpcHandler.DeleteLink)

	return router;
}

func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "up"})
}
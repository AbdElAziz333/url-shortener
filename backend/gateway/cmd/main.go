package main

import (
	"context"
	"os"

	"aziz.dev/gateway/internal/configloader"
	"aziz.dev/gateway/internal/grpcclient"
	"aziz.dev/gateway/internal/middleware"
	"aziz.dev/gateway/internal/postgres"
	sqlcgen "aziz.dev/gateway/internal/postgres/sqlc"
	"aziz.dev/gateway/internal/redis"
	"aziz.dev/gateway/internal/security"
	"aziz.dev/gateway/internal/server"
	"aziz.dev/gateway/internal/user"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.InfoLevel)

	config, err := configloader.LoadFromEnv()
	if err != nil {
		logrus.WithError(err).Fatal("failed to load config")
	}

	ctx := context.Background()

	postgresDB, err := postgres.NewClient(ctx, &config.Postgres)
	if err != nil {
		logrus.WithError(err).Fatal("failed to connect to PostgreSQL")
	}

	// Not used yet (refresh tokens are a TODO); kept so startup fails early on bad config.
	if _, err := redis.NewClient(ctx, &config.Redis); err != nil {
		logrus.WithError(err).Fatal("failed to connect to Redis")
	}

	jwtService, err := security.NewService(&config.JWT)
	if err != nil {
		logrus.WithError(err).Fatal("failed to initialize JWT service")
	}

	userService := user.NewService(sqlcgen.New(postgresDB), jwtService)
	userHandler := user.NewHandler(userService)

	grpcClients, err := grpcclient.NewClients(config.Service.ShortenerServiceURL)
	if err != nil {
		logrus.WithError(err).Fatal("failed to initialize gRPC clients")
	}
	defer grpcClients.Close()

	grpcHandler := grpcclient.NewHandler(grpcClients)
	router := server.NewRouter(userHandler)

	links := router.Group("/shortener/api/links", middleware.AccessTokenMiddleware(&config.JWT))
	links.GET("", grpcHandler.GetAllLinks)
	links.GET("/:code", grpcHandler.GetLinkByCode)
	links.POST("", grpcHandler.CreateLink)
	links.PATCH("/:code/expiry", grpcHandler.UpdateExpiry)
	links.DELETE("/:code", grpcHandler.DeleteLink)

	logrus.Infof("gateway listening on port %s", config.Service.Port)
	if err := router.Run(":" + config.Service.Port); err != nil {
		logrus.WithError(err).Fatal("server stopped")
	}
}
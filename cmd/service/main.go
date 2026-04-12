package main

import (
	"go.uber.org/fx"

	"playlist-service/internal/config"
	"playlist-service/internal/handler"
	"playlist-service/internal/logger"
	"playlist-service/internal/postgres"
	"playlist-service/internal/redis"
	"playlist-service/internal/repository"
	"playlist-service/internal/server"
	"playlist-service/internal/tracing"
	"playlist-service/internal/usecase"
)

// @title           Playlist Service API
// @version         1.0
// @description     API сервиса управления плейлистами.
// @host            localhost:8080
// @BasePath        /api/v1
func main() {
	cfg := config.Load()

	fx.New(
		fx.Supply(cfg),
		fx.Provide(
			logger.New,
			postgres.New,
			redis.New,
			fx.Annotate(
				repository.NewPostgresRepo,
				fx.As(new(repository.Playlist)),
			),
			fx.Annotate(
				repository.NewRedisCache,
				fx.As(new(repository.Cache)),
			),
			usecase.NewPlaylistUseCase,
			handler.New,
			handler.NewPlaylistHandler,
		),
		fx.Invoke(
			tracing.Register,
			server.Register,
		),
		fx.StopTimeout(cfg.ShutdownTimeout),
	).Run()
}

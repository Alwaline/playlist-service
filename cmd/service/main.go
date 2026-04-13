package main

import (
	"context"
	"log/slog"
	"playlist-service/internal/config"
	"playlist-service/internal/consumer"
	"playlist-service/internal/handler"
	"playlist-service/internal/kafka"
	"playlist-service/internal/logger"
	"playlist-service/internal/postgres"
	"playlist-service/internal/redis"
	"playlist-service/internal/repository"
	"playlist-service/internal/server"
	"playlist-service/internal/tracing"
	"playlist-service/internal/usecase"

	"go.uber.org/fx"
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
			func(cfg config.Config) postgres.Config { return cfg.Postgres },
			func(cfg config.Config) redis.Config { return cfg.Redis },
			func(cfg config.Config) kafka.ProducerConfig { return cfg.Kafka },
			postgres.New,
			redis.New,
			kafka.NewProducer,
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
			func(cfg config.Config, log *slog.Logger) *kafka.Consumer {
				return kafka.NewConsumer(cfg.KafkaConsumer, log)
			},
			consumer.NewTrackDeletedConsumer,
		),
		fx.Invoke(
			tracing.Register,
			server.Register,
			func(lc fx.Lifecycle, c *kafka.Consumer, h *consumer.TrackDeletedConsumer, log *slog.Logger) {
				lc.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						go func() {
							if err := c.Run(ctx, h.Handle); err != nil {
								log.Error("kafka consumer error", "error", err)
							}
						}()
						return nil
					},
					OnStop: func(ctx context.Context) error {
						return c.Close()
					},
				})
			},
		),
		fx.StopTimeout(cfg.ShutdownTimeout),
	).Run()
}

package main

import (
	"go.uber.org/fx"

	"playlist-service/internal/config"
	"playlist-service/internal/handler"
	"playlist-service/internal/logger"
	"playlist-service/internal/server"
	"playlist-service/internal/tracing"
)

// @title           Service API
// @version         1.0
// @description     API шаблона микросервиса.
// @host            localhost:8080
// @BasePath        /api/v1
func main() {
	cfg := config.Load()

	fx.New(
		fx.Supply(cfg),
		fx.Provide(
			logger.New,
			handler.New,
		),
		fx.Invoke(
			tracing.Register,
			server.Register,
		),
		fx.StopTimeout(cfg.ShutdownTimeout),
	).Run()
}

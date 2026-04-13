package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"playlist-service/internal/config"
	"playlist-service/internal/handler"
	"playlist-service/internal/middleware"
	"time"

	"github.com/fasthttp/router"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
	"go.uber.org/fx"

	_ "playlist-service/docs"
)

func withRoute(pattern string, next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		ctx.SetUserValue(middleware.RoutePatternKey, pattern)
		next(ctx)
	}
}

// Register creates the fasthttp server and binds its start/stop to the fx lifecycle.
func Register(lc fx.Lifecycle, cfg config.Config, logger *slog.Logger, h *handler.Handler, ph *handler.PlaylistHandler) {
	r := router.New()

	r.GET("/healthz", withRoute("/healthz", h.Healthz))
	r.GET("/readyz", withRoute("/readyz", h.Readyz))
	r.GET("/metrics", withRoute("/metrics", fasthttpadaptor.NewFastHTTPHandler(promhttp.Handler())))
	r.GET("/api/v1/example", withRoute("/api/v1/example", h.Example))
	r.GET("/swagger/{filepath:*}", withRoute("/swagger", fasthttpadaptor.NewFastHTTPHandler(httpSwagger.WrapHandler)))

	// Playlist routes
	r.POST("/api/v1/playlists", withRoute("/api/v1/playlists", ph.CreatePlaylist))
	r.GET("/api/v1/playlists/{id}/tracks", withRoute("/api/v1/playlists/{id}/tracks", ph.GetPlaylistTracks))
	r.POST("/api/v1/playlists/{id}/tracks", withRoute("/api/v1/playlists/{id}/tracks", ph.AddTrack))
	r.DELETE("/api/v1/playlists/{id}/tracks/{track_id}", withRoute("/api/v1/playlists/{id}/tracks/{track_id}", ph.RemoveTrack))

	chain := middleware.Chain(
		middleware.Recoverer(logger),
		middleware.RequestID,
		middleware.Tracing,
		middleware.Metrics,
		middleware.Logger(logger),
	)

	srv := &fasthttp.Server{
		Handler:            chain(r.Handler),
		Name:               cfg.ServiceName,
		ReadTimeout:        10 * time.Second,
		WriteTimeout:       10 * time.Second,
		IdleTimeout:        120 * time.Second,
		MaxRequestBodySize: 4 * 1024 * 1024,
	}

	addr := ":" + cfg.HTTPPort

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			lc := &net.ListenConfig{}
			ln, err := lc.Listen(ctx, "tcp", addr)
			if err != nil {
				return fmt.Errorf("listen %s: %w", addr, err)
			}
			logger.Info("http server listening", "addr", addr)
			go func() {
				if err := srv.Serve(ln); err != nil {
					logger.Error("http server error", "error", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			h.SetReady(false)
			logger.Info("shutting down http server")

			done := make(chan error, 1)
			go func() {
				done <- srv.Shutdown()
			}()

			select {
			case err := <-done:
				return err
			case <-ctx.Done():
				return fmt.Errorf("shutdown interrupted: %w", ctx.Err())
			}
		},
	})
}

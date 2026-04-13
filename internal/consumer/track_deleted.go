package consumer

import (
	"context"
	"encoding/json"
	"log/slog"
	"playlist-service/internal/repository"

	"github.com/segmentio/kafka-go"
)

type TrackDeletedConsumer struct {
	repo  repository.Playlist
	cache repository.Cache
	log   *slog.Logger
}

func NewTrackDeletedConsumer(repo repository.Playlist, cache repository.Cache, log *slog.Logger) *TrackDeletedConsumer {
	return &TrackDeletedConsumer{repo: repo, cache: cache, log: log}
}

type trackDeletedPayload struct {
	TrackID string `json:"track_id"`
}

func (c *TrackDeletedConsumer) Handle(ctx context.Context, msg kafka.Message) error {
	var envelope struct {
		Type    string              `json:"type"`
		Payload trackDeletedPayload `json:"payload"`
	}

	if err := json.Unmarshal(msg.Value, &envelope); err != nil {
		c.log.Error("failed to unmarshal track.deleted event", "error", err)
		return nil
	}

	trackID := envelope.Payload.TrackID
	if trackID == "" {
		c.log.Warn("track.deleted event has empty track_id")
		return nil
	}

	playlistIDs, err := c.repo.RemoveTrackFromAllPlaylists(ctx, trackID)
	if err != nil {
		c.log.Error("failed to remove track from playlists", "track_id", trackID, "error", err)
		return nil
	}

	for _, playlistID := range playlistIDs {
		if err := c.cache.InvalidatePlaylist(ctx, playlistID); err != nil {
			c.log.Warn("failed to invalidate playlist cache", "playlist_id", playlistID, "error", err)
		}
	}

	c.log.Info("track deleted from all playlists", "track_id", trackID, "affected_playlists", len(playlistIDs))
	return nil
}

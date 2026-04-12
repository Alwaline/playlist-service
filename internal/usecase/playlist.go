package usecase

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/redis/go-redis/v9"

	"playlist-service/internal/domain"
	"playlist-service/internal/kafka"
	"playlist-service/internal/repository"
)

type PlaylistUseCase struct {
	repo     repository.Playlist
	cache    repository.Cache
	producer *kafka.Producer
}

func NewPlaylistUseCase(repo repository.Playlist, cache repository.Cache, producer *kafka.Producer) *PlaylistUseCase {
	return &PlaylistUseCase{repo: repo, cache: cache, producer: producer}
}

func (uc *PlaylistUseCase) CreatePlaylist(ctx context.Context, ownerID, name string) (*domain.Playlist, error) {
	if name == "" {
		return nil, errors.New("name cannot be empty")
	}

	playlist := &domain.Playlist{
		OwnerID: ownerID,
		Name:    name,
	}

	if err := uc.repo.CreatePlaylist(ctx, playlist); err != nil {
		return nil, err
	}
	return playlist, nil
}

func (uc *PlaylistUseCase) AddTrack(ctx context.Context, ownerID, playlistID string, meta *domain.TrackMeta) error {
	playlist, err := uc.repo.GetPlaylist(ctx, playlistID)
	if err != nil {
		return err
	}
	if playlist == nil {
		return ErrNotFound
	}
	if playlist.OwnerID != ownerID {
		return ErrForbidden
	}

	if err := uc.repo.UpsertTrackMeta(ctx, meta); err != nil {
		return err
	}

	if err := uc.repo.AddTrack(ctx, playlistID, meta); err != nil {
		return err
	}

	_ = uc.cache.InvalidatePlaylist(ctx, playlistID)

	payload, _ := json.Marshal(map[string]string{
		"playlist_id": playlistID,
		"track_id":    meta.TrackID,
	})

	event, _ := json.Marshal(map[string]any{
		"type":    "playlist.track_added",
		"version": 1,
		"payload": json.RawMessage(payload),
	})

	_ = uc.producer.Publish(ctx, "playlists", []byte(playlistID), event)

	return nil
}

func (uc *PlaylistUseCase) RemoveTrack(ctx context.Context, ownerID, playlistID, trackID string) error {
	playlist, err := uc.repo.GetPlaylist(ctx, playlistID)
	if err != nil {
		return err
	}
	if playlist == nil {
		return ErrNotFound
	}
	if playlist.OwnerID != ownerID {
		return ErrForbidden
	}

	if err := uc.repo.RemoveTrack(ctx, playlistID, trackID); err != nil {
		return err
	}

	_ = uc.cache.InvalidatePlaylist(ctx, playlistID)

	return nil
}

func (uc *PlaylistUseCase) GetPlaylistTracks(ctx context.Context, ownerID, playlistID string) ([]domain.PlaylistTrack, error) {
	playlist, err := uc.repo.GetPlaylist(ctx, playlistID)
	if err != nil {
		return nil, err
	}
	if playlist == nil {
		return nil, ErrNotFound
	}
	if playlist.OwnerID != ownerID {
		return nil, ErrForbidden
	}

	tracks, err := uc.cache.GetPlaylistTracks(ctx, playlistID)
	if err == nil {
		return tracks, nil
	}
	if !errors.Is(err, redis.Nil) {
		return nil, err
	}

	tracks, err = uc.repo.GetPlaylistTracks(ctx, playlistID)
	if err != nil {
		return nil, err
	}

	_ = uc.cache.SetPlaylistTracks(ctx, playlistID, tracks)

	return tracks, nil
}

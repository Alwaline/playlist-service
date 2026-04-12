package repository

import (
	"context"

	"playlist-service/internal/domain"
)

type Playlist interface {
	CreatePlaylist(ctx context.Context, playlist *domain.Playlist) error
	GetPlaylist(ctx context.Context, playlistID string) (*domain.Playlist, error)
	AddTrack(ctx context.Context, playlistID string, meta *domain.TrackMeta) error
	RemoveTrack(ctx context.Context, playlistID string, trackID string) error
	RemoveTrackFromAllPlaylists(ctx context.Context, trackID string) ([]string, error)
	GetPlaylistTracks(ctx context.Context, playlistID string) ([]domain.PlaylistTrack, error)
	UpsertTrackMeta(ctx context.Context, meta *domain.TrackMeta) error
}

type Cache interface {
	GetPlaylistTracks(ctx context.Context, playlistID string) ([]domain.PlaylistTrack, error)
	SetPlaylistTracks(ctx context.Context, playlistID string, tracks []domain.PlaylistTrack) error
	InvalidatePlaylist(ctx context.Context, playlistID string) error
}

package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"

	"playlist-service/internal/domain"
)

const playlistTTL = 5 * time.Minute

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(client *redis.Client) *RedisCache {
	return &RedisCache{client: client}
}

func cacheKey(playlistID string) string {
	return "playlist:" + playlistID
}

func (c *RedisCache) GetPlaylistTracks(ctx context.Context, playlistID string) ([]domain.PlaylistTrack, error) {
	data, err := c.client.Get(ctx, cacheKey(playlistID)).Bytes()
	if err != nil {
		return nil, err
	}

	var tracks []domain.PlaylistTrack
	if err := json.Unmarshal(data, &tracks); err != nil {
		return nil, err
	}
	return tracks, nil
}

func (c *RedisCache) SetPlaylistTracks(ctx context.Context, playlistID string, tracks []domain.PlaylistTrack) error {
	data, err := json.Marshal(tracks)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, cacheKey(playlistID), data, playlistTTL).Err()
}

func (c *RedisCache) InvalidatePlaylist(ctx context.Context, playlistID string) error {
	return c.client.Del(ctx, cacheKey(playlistID)).Err()
}

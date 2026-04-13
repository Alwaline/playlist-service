package repository

import (
	"context"
	"errors"
	"playlist-service/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepo struct {
	db *pgxpool.Pool
}

func NewPostgresRepo(db *pgxpool.Pool) *PostgresRepo {
	return &PostgresRepo{db: db}
}

func (r *PostgresRepo) CreatePlaylist(ctx context.Context, playlist *domain.Playlist) error {
	query := `
		INSERT INTO playlists (owner_id, name)
		VALUES ($1, $2)
		RETURNING id, created_at, updated_at`

	return r.db.QueryRow(ctx, query, playlist.OwnerID, playlist.Name).
		Scan(&playlist.ID, &playlist.CreatedAt, &playlist.UpdatedAt)
}

func (r *PostgresRepo) GetPlaylist(ctx context.Context, playlistID string) (*domain.Playlist, error) {
	query := `SELECT id, owner_id, name, created_at, updated_at FROM playlists WHERE id = $1`

	p := &domain.Playlist{}
	err := r.db.QueryRow(ctx, query, playlistID).
		Scan(&p.ID, &p.OwnerID, &p.Name, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return p, err
}

func (r *PostgresRepo) UpsertTrackMeta(ctx context.Context, meta *domain.TrackMeta) error {
	query := `
		INSERT INTO track_meta (track_id, title, artist, duration_sec)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (track_id) DO UPDATE
		SET title = EXCLUDED.title,
		    artist = EXCLUDED.artist,
		    duration_sec = EXCLUDED.duration_sec`

	_, err := r.db.Exec(ctx, query, meta.TrackID, meta.Title, meta.Artist, meta.DurationSec)
	return err
}

func (r *PostgresRepo) AddTrack(ctx context.Context, playlistID string, meta *domain.TrackMeta) error {
	query := `
		INSERT INTO playlist_tracks (playlist_id, track_id, position)
		VALUES ($1, $2, (
			SELECT COALESCE(MAX(position) + 1, 0)
			FROM playlist_tracks
			WHERE playlist_id = $1
		))`

	_, err := r.db.Exec(ctx, query, playlistID, meta.TrackID)
	return err
}

func (r *PostgresRepo) RemoveTrack(ctx context.Context, playlistID, trackID string) error {
	query := `DELETE FROM playlist_tracks WHERE playlist_id = $1 AND track_id = $2`
	_, err := r.db.Exec(ctx, query, playlistID, trackID)
	return err
}

func (r *PostgresRepo) RemoveTrackFromAllPlaylists(ctx context.Context, trackID string) ([]string, error) {
	query := `
		DELETE FROM playlist_tracks
		WHERE track_id = $1
		RETURNING playlist_id`

	rows, err := r.db.Query(ctx, query, trackID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var playlistIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		playlistIDs = append(playlistIDs, id)
	}
	return playlistIDs, nil
}

func (r *PostgresRepo) GetPlaylistTracks(ctx context.Context, playlistID string) ([]domain.PlaylistTrack, error) {
	query := `
		SELECT pt.playlist_id, pt.track_id, pt.position, pt.added_at,
		       tm.title, tm.artist, tm.duration_sec
		FROM playlist_tracks pt
		JOIN track_meta tm ON tm.track_id = pt.track_id
		WHERE pt.playlist_id = $1
		ORDER BY pt.position`

	rows, err := r.db.Query(ctx, query, playlistID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tracks []domain.PlaylistTrack
	for rows.Next() {
		var t domain.PlaylistTrack
		err := rows.Scan(
			&t.PlaylistID, &t.TrackID, &t.Position, &t.AddedAt,
			&t.Meta.Title, &t.Meta.Artist, &t.Meta.DurationSec,
		)
		if err != nil {
			return nil, err
		}
		t.Meta.TrackID = t.TrackID
		tracks = append(tracks, t)
	}
	return tracks, nil
}

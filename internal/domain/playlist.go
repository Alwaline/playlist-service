package domain

import "time"

type Playlist struct {
	ID        string    `json:"id"`
	OwnerID   string    `json:"owner_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TrackMeta struct {
	TrackID     string `json:"track_id"`
	Title       string `json:"title"`
	Artist      string `json:"artist"`
	DurationSec int    `json:"duration_sec"`
}

type PlaylistTrack struct {
	PlaylistID string    `json:"playlist_id"`
	TrackID    string    `json:"track_id"`
	Position   int       `json:"position"`
	AddedAt    time.Time `json:"added_at"`
	Meta       TrackMeta `json:"meta"`
}

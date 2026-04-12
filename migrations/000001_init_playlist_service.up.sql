CREATE TABLE IF NOT EXISTS playlists (
                                         id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id   TEXT NOT NULL,
    name       TEXT NOT NULL CHECK (char_length(name) > 0 AND char_length(name) <= 255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
    );

CREATE TABLE IF NOT EXISTS track_meta (
                                          track_id    TEXT PRIMARY KEY,
                                          title       TEXT NOT NULL,
                                          artist      TEXT NOT NULL,
                                          duration_sec INT NOT NULL
);

CREATE TABLE IF NOT EXISTS playlist_tracks (
                                               playlist_id UUID NOT NULL REFERENCES playlists(id) ON DELETE CASCADE,
    track_id    TEXT NOT NULL REFERENCES track_meta(track_id) ON DELETE CASCADE,
    position    INT NOT NULL DEFAULT 0,
    added_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (playlist_id, track_id)
    );

CREATE INDEX IF NOT EXISTS idx_playlist_tracks_playlist_id ON playlist_tracks(playlist_id);
CREATE INDEX IF NOT EXISTS idx_playlist_tracks_track_id ON playlist_tracks(track_id);
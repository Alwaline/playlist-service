package handler

import (
	"errors"

	"github.com/valyala/fasthttp"

	"playlist-service/internal/apperror"
	"playlist-service/internal/domain"
	"playlist-service/internal/usecase"
	"playlist-service/internal/validator"
)

type PlaylistHandler struct {
	uc *usecase.PlaylistUseCase
}

func NewPlaylistHandler(uc *usecase.PlaylistUseCase) *PlaylistHandler {
	return &PlaylistHandler{uc: uc}
}

type createPlaylistRequest struct {
	Name string `json:"name" validate:"required,min=1,max=255"`
}

type addTrackRequest struct {
	TrackID     string `json:"track_id" validate:"required"`
	Title       string `json:"title" validate:"required"`
	Artist      string `json:"artist" validate:"required"`
	DurationSec int    `json:"duration_sec" validate:"required,min=1"`
}

func (h *PlaylistHandler) CreatePlaylist(ctx *fasthttp.RequestCtx) {
	ownerID := string(ctx.Request.Header.Peek("X-User-ID"))
	if ownerID == "" {
		WriteError(ctx, apperror.NewUnauthorized("missing X-User-ID header"))
		return
	}

	var req createPlaylistRequest
	if err := validator.BindJSON(ctx, &req); err != nil {
		WriteError(ctx, err)
		return
	}

	playlist, err := h.uc.CreatePlaylist(ctx, ownerID, req.Name)
	if err != nil {
		WriteError(ctx, apperror.NewInternal("failed to create playlist", err))
		return
	}

	WriteJSON(ctx, fasthttp.StatusCreated, playlist)
}

func (h *PlaylistHandler) AddTrack(ctx *fasthttp.RequestCtx) {
	ownerID := string(ctx.Request.Header.Peek("X-User-ID"))
	if ownerID == "" {
		WriteError(ctx, apperror.NewUnauthorized("missing X-User-ID header"))
		return
	}

	playlistID := ctx.UserValue("id").(string)

	var req addTrackRequest
	if err := validator.BindJSON(ctx, &req); err != nil {
		WriteError(ctx, err)
		return
	}

	meta := &domain.TrackMeta{
		TrackID:     req.TrackID,
		Title:       req.Title,
		Artist:      req.Artist,
		DurationSec: req.DurationSec,
	}

	err := h.uc.AddTrack(ctx, ownerID, playlistID, meta)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrNotFound):
			WriteError(ctx, apperror.NewNotFound("playlist not found"))
		case errors.Is(err, usecase.ErrForbidden):
			WriteError(ctx, apperror.NewForbidden("access denied"))
		default:
			WriteError(ctx, apperror.NewInternal("failed to add track", err))
		}
		return
	}

	WriteJSON(ctx, fasthttp.StatusOK, map[string]string{"status": "ok"})
}

func (h *PlaylistHandler) RemoveTrack(ctx *fasthttp.RequestCtx) {
	ownerID := string(ctx.Request.Header.Peek("X-User-ID"))
	if ownerID == "" {
		WriteError(ctx, apperror.NewUnauthorized("missing X-User-ID header"))
		return
	}

	playlistID := ctx.UserValue("id").(string)
	trackID := ctx.UserValue("track_id").(string)

	err := h.uc.RemoveTrack(ctx, ownerID, playlistID, trackID)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrNotFound):
			WriteError(ctx, apperror.NewNotFound("playlist not found"))
		case errors.Is(err, usecase.ErrForbidden):
			WriteError(ctx, apperror.NewForbidden("access denied"))
		default:
			WriteError(ctx, apperror.NewInternal("failed to remove track", err))
		}
		return
	}

	ctx.SetStatusCode(fasthttp.StatusNoContent)
}

func (h *PlaylistHandler) GetPlaylistTracks(ctx *fasthttp.RequestCtx) {
	ownerID := string(ctx.Request.Header.Peek("X-User-ID"))
	if ownerID == "" {
		WriteError(ctx, apperror.NewUnauthorized("missing X-User-ID header"))
		return
	}

	playlistID := ctx.UserValue("id").(string)

	tracks, err := h.uc.GetPlaylistTracks(ctx, ownerID, playlistID)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrNotFound):
			WriteError(ctx, apperror.NewNotFound("playlist not found"))
		case errors.Is(err, usecase.ErrForbidden):
			WriteError(ctx, apperror.NewForbidden("access denied"))
		default:
			WriteError(ctx, apperror.NewInternal("failed to get tracks", err))
		}
		return
	}

	WriteJSON(ctx, fasthttp.StatusOK, tracks)
}

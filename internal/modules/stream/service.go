package stream

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/vkrishna03/streamz/db/sqlc"
	apperr "github.com/vkrishna03/streamz/internal/errors"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// Start creates a new stream
func (s *Service) Start(ctx context.Context, userID uuid.UUID, req StartRequest) (*Response, error) {
	// Check concurrent stream limit
	settings, err := s.repo.GetUserSettings(ctx, userID)
	if err != nil {
		return nil, apperr.Wrap(apperr.ErrInternal, "failed to get user settings")
	}

	count, err := s.repo.CountActiveByUser(ctx, userID)
	if err != nil {
		return nil, apperr.Wrap(apperr.ErrInternal, "failed to count active streams")
	}

	if count >= int64(settings.MaxConcurrentStreams.Int32) {
		return nil, apperr.Wrap(apperr.ErrValidation, "concurrent stream limit reached (max %d)", settings.MaxConcurrentStreams.Int32)
	}

	// Parse stream type
	streamType, err := parseStreamType(req.StreamType)
	if err != nil {
		return nil, err
	}

	// Parse quality (default to auto)
	quality := sqlc.StreamQualityAuto
	if req.Quality != "" {
		q, err := parseStreamQuality(req.Quality)
		if err != nil {
			return nil, err
		}
		quality = q
	}

	// Create stream
	var targetDeviceID *uuid.UUID
	if req.TargetDeviceID != uuid.Nil {
		targetDeviceID = &req.TargetDeviceID
	}

	stream, err := s.repo.Create(ctx, userID, &req.SourceDeviceID, targetDeviceID, streamType, quality)
	if err != nil {
		return nil, apperr.Wrap(apperr.ErrInternal, "failed to create stream")
	}

	// Increment stream count
	_ = s.repo.IncrementStreamCount(ctx, userID)

	resp := toResponse(stream)
	return &resp, nil
}

// List returns all streams for the user
func (s *Service) List(ctx context.Context, userID uuid.UUID) ([]Response, error) {
	streams, err := s.repo.ListByUser(ctx, userID)
	if err != nil {
		return nil, apperr.Wrap(apperr.ErrInternal, "failed to list streams")
	}

	return toResponseList(streams), nil
}

// ListActive returns all active streams for the user
func (s *Service) ListActive(ctx context.Context, userID uuid.UUID) ([]Response, error) {
	streams, err := s.repo.ListActiveByUser(ctx, userID)
	if err != nil {
		return nil, apperr.Wrap(apperr.ErrInternal, "failed to list active streams")
	}

	return toResponseList(streams), nil
}

// GetByID returns a stream by ID
func (s *Service) GetByID(ctx context.Context, userID, streamID uuid.UUID) (*Response, error) {
	stream, err := s.repo.GetByID(ctx, streamID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperr.Wrap(apperr.ErrNotFound, "stream not found")
		}
		return nil, apperr.Wrap(apperr.ErrInternal, "failed to get stream")
	}

	// Verify ownership
	if stream.UserID != userID {
		return nil, apperr.Wrap(apperr.ErrForbidden, "stream not owned by user")
	}

	resp := toResponse(stream)
	return &resp, nil
}

// UpdateStatus updates the stream status
func (s *Service) UpdateStatus(ctx context.Context, userID, streamID uuid.UUID, req UpdateStatusRequest) error {
	stream, err := s.repo.GetByID(ctx, streamID)
	if err != nil {
		if err == sql.ErrNoRows {
			return apperr.Wrap(apperr.ErrNotFound, "stream not found")
		}
		return apperr.Wrap(apperr.ErrInternal, "failed to get stream")
	}

	if stream.UserID != userID {
		return apperr.Wrap(apperr.ErrForbidden, "stream not owned by user")
	}

	status, err := parseStreamStatus(req.Status)
	if err != nil {
		return err
	}

	return s.repo.UpdateStatus(ctx, streamID, status)
}

// UpdateLatency updates the stream latency
func (s *Service) UpdateLatency(ctx context.Context, userID, streamID uuid.UUID, latencyMs int) error {
	stream, err := s.repo.GetByID(ctx, streamID)
	if err != nil {
		if err == sql.ErrNoRows {
			return apperr.Wrap(apperr.ErrNotFound, "stream not found")
		}
		return apperr.Wrap(apperr.ErrInternal, "failed to get stream")
	}

	if stream.UserID != userID {
		return apperr.Wrap(apperr.ErrForbidden, "stream not owned by user")
	}

	return s.repo.UpdateLatency(ctx, streamID, latencyMs)
}

// UpdateConnectionType updates the connection type
func (s *Service) UpdateConnectionType(ctx context.Context, userID, streamID uuid.UUID, req UpdateConnectionTypeRequest) error {
	stream, err := s.repo.GetByID(ctx, streamID)
	if err != nil {
		if err == sql.ErrNoRows {
			return apperr.Wrap(apperr.ErrNotFound, "stream not found")
		}
		return apperr.Wrap(apperr.ErrInternal, "failed to get stream")
	}

	if stream.UserID != userID {
		return apperr.Wrap(apperr.ErrForbidden, "stream not owned by user")
	}

	connType, err := parseConnectionType(req.ConnectionType)
	if err != nil {
		return err
	}

	return s.repo.UpdateConnectionType(ctx, streamID, connType)
}

// End ends a stream
func (s *Service) End(ctx context.Context, userID, streamID uuid.UUID) error {
	stream, err := s.repo.GetByID(ctx, streamID)
	if err != nil {
		if err == sql.ErrNoRows {
			return apperr.Wrap(apperr.ErrNotFound, "stream not found")
		}
		return apperr.Wrap(apperr.ErrInternal, "failed to get stream")
	}

	if stream.UserID != userID {
		return apperr.Wrap(apperr.ErrForbidden, "stream not owned by user")
	}

	// Calculate duration and update metrics
	if stream.StartedAt.Valid {
		duration := time.Since(stream.StartedAt.Time)
		minutes := int(duration.Minutes())
		if minutes > 0 {
			_ = s.repo.IncrementStreamMinutes(ctx, userID, minutes)
		}
	}

	return s.repo.End(ctx, streamID)
}

// Helpers

func toResponse(s sqlc.Stream) Response {
	resp := Response{
		ID:         s.ID,
		UserID:     s.UserID,
		StreamType: string(s.StreamType),
		Status:     string(s.Status.StreamStatus),
		StartedAt:  s.StartedAt.Time,
	}

	if s.SourceDeviceID.Valid {
		resp.SourceDeviceID = &s.SourceDeviceID.UUID
	}
	if s.TargetDeviceID.Valid {
		resp.TargetDeviceID = &s.TargetDeviceID.UUID
	}
	if s.ConnectionType.Valid {
		ct := string(s.ConnectionType.ConnectionType)
		resp.ConnectionType = &ct
	}
	if s.Quality.Valid {
		resp.Quality = string(s.Quality.StreamQuality)
	}
	if s.LatencyMs.Valid {
		lm := int(s.LatencyMs.Int32)
		resp.LatencyMs = &lm
	}
	if s.EndedAt.Valid {
		resp.EndedAt = &s.EndedAt.Time
	}

	return resp
}

func toResponseList(streams []sqlc.Stream) []Response {
	resp := make([]Response, len(streams))
	for i, s := range streams {
		resp[i] = toResponse(s)
	}
	return resp
}

func parseStreamType(s string) (sqlc.StreamType, error) {
	switch s {
	case "video":
		return sqlc.StreamTypeVideo, nil
	case "audio":
		return sqlc.StreamTypeAudio, nil
	case "both":
		return sqlc.StreamTypeBoth, nil
	default:
		return "", apperr.Wrap(apperr.ErrValidation, "invalid stream type")
	}
}

func parseStreamQuality(s string) (sqlc.StreamQuality, error) {
	switch s {
	case "low":
		return sqlc.StreamQualityLow, nil
	case "medium":
		return sqlc.StreamQualityMedium, nil
	case "high":
		return sqlc.StreamQualityHigh, nil
	case "auto":
		return sqlc.StreamQualityAuto, nil
	default:
		return "", apperr.Wrap(apperr.ErrValidation, "invalid stream quality")
	}
}

func parseStreamStatus(s string) (sqlc.StreamStatus, error) {
	switch s {
	case "connecting":
		return sqlc.StreamStatusConnecting, nil
	case "active":
		return sqlc.StreamStatusActive, nil
	case "paused":
		return sqlc.StreamStatusPaused, nil
	case "ended":
		return sqlc.StreamStatusEnded, nil
	case "failed":
		return sqlc.StreamStatusFailed, nil
	default:
		return "", apperr.Wrap(apperr.ErrValidation, "invalid stream status")
	}
}

func parseConnectionType(s string) (sqlc.ConnectionType, error) {
	switch s {
	case "p2p":
		return sqlc.ConnectionTypeP2p, nil
	case "relay":
		return sqlc.ConnectionTypeRelay, nil
	default:
		return "", apperr.Wrap(apperr.ErrValidation, "invalid connection type")
	}
}

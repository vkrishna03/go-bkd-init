package stream

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/vkrishna03/streamz/db/sqlc"
)

type Repository struct {
	db *sql.DB
	q  *sqlc.Queries
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db, q: sqlc.New(db)}
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (sqlc.Stream, error) {
	return r.q.GetStreamByID(ctx, id)
}

func (r *Repository) ListByUser(ctx context.Context, userID uuid.UUID) ([]sqlc.Stream, error) {
	return r.q.ListUserStreams(ctx, userID)
}

func (r *Repository) ListActiveByUser(ctx context.Context, userID uuid.UUID) ([]sqlc.Stream, error) {
	return r.q.ListActiveUserStreams(ctx, userID)
}

func (r *Repository) Create(ctx context.Context, userID uuid.UUID, sourceDeviceID, targetDeviceID *uuid.UUID, streamType sqlc.StreamType, quality sqlc.StreamQuality) (sqlc.Stream, error) {
	return r.q.CreateStream(ctx, sqlc.CreateStreamParams{
		UserID:         userID,
		SourceDeviceID: toNullUUID(sourceDeviceID),
		TargetDeviceID: toNullUUID(targetDeviceID),
		StreamType:     streamType,
		Quality:        sqlc.NullStreamQuality{StreamQuality: quality, Valid: true},
	})
}

func (r *Repository) UpdateStatus(ctx context.Context, id uuid.UUID, status sqlc.StreamStatus) error {
	return r.q.UpdateStreamStatus(ctx, sqlc.UpdateStreamStatusParams{
		ID:     id,
		Status: sqlc.NullStreamStatus{StreamStatus: status, Valid: true},
	})
}

func (r *Repository) UpdateLatency(ctx context.Context, id uuid.UUID, latencyMs int) error {
	return r.q.UpdateStreamLatency(ctx, sqlc.UpdateStreamLatencyParams{
		ID:        id,
		LatencyMs: sql.NullInt32{Int32: int32(latencyMs), Valid: true},
	})
}

func (r *Repository) UpdateConnectionType(ctx context.Context, id uuid.UUID, connType sqlc.ConnectionType) error {
	return r.q.UpdateStreamConnectionType(ctx, sqlc.UpdateStreamConnectionTypeParams{
		ID:             id,
		ConnectionType: sqlc.NullConnectionType{ConnectionType: connType, Valid: true},
	})
}

func (r *Repository) End(ctx context.Context, id uuid.UUID) error {
	return r.q.EndStream(ctx, id)
}

func (r *Repository) CountActiveByUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	return r.q.CountActiveUserStreams(ctx, userID)
}

// User settings

func (r *Repository) GetUserSettings(ctx context.Context, userID uuid.UUID) (sqlc.UserSetting, error) {
	return r.q.GetUserSettings(ctx, userID)
}

func (r *Repository) IncrementStreamCount(ctx context.Context, userID uuid.UUID) error {
	return r.q.IncrementStreamCount(ctx, userID)
}

func (r *Repository) IncrementStreamMinutes(ctx context.Context, userID uuid.UUID, minutes int) error {
	return r.q.IncrementStreamMinutes(ctx, sqlc.IncrementStreamMinutesParams{
		UserID:             userID,
		TotalStreamMinutes: sql.NullInt32{Int32: int32(minutes), Valid: true},
	})
}

// Helpers

func toNullUUID(id *uuid.UUID) uuid.NullUUID {
	if id == nil {
		return uuid.NullUUID{}
	}
	return uuid.NullUUID{UUID: *id, Valid: true}
}

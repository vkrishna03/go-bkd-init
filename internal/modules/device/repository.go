package device

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

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (sqlc.Device, error) {
	return r.q.GetDeviceByID(ctx, id)
}

func (r *Repository) GetByUserAndDeviceID(ctx context.Context, userID uuid.UUID, deviceID string) (sqlc.Device, error) {
	return r.q.GetDeviceByUserAndDeviceID(ctx, sqlc.GetDeviceByUserAndDeviceIDParams{
		UserID:   userID,
		DeviceID: deviceID,
	})
}

func (r *Repository) ListByUser(ctx context.Context, userID uuid.UUID) ([]sqlc.Device, error) {
	return r.q.ListUserDevices(ctx, userID)
}

func (r *Repository) ListOnlineByUser(ctx context.Context, userID uuid.UUID) ([]sqlc.Device, error) {
	return r.q.ListOnlineUserDevices(ctx, userID)
}

func (r *Repository) Create(ctx context.Context, userID uuid.UUID, deviceID, deviceName string, deviceType sqlc.DeviceType, hasCamera, hasMic bool) (sqlc.Device, error) {
	return r.q.CreateDevice(ctx, sqlc.CreateDeviceParams{
		UserID:        userID,
		DeviceID:      deviceID,
		DeviceName:    deviceName,
		DeviceType:    deviceType,
		HasCamera:     sql.NullBool{Bool: hasCamera, Valid: true},
		HasMicrophone: sql.NullBool{Bool: hasMic, Valid: true},
	})
}

func (r *Repository) Update(ctx context.Context, id uuid.UUID, name *string, hasCamera, hasMic *bool) (sqlc.Device, error) {
	// Get current device to use existing values for nil params
	current, err := r.q.GetDeviceByID(ctx, id)
	if err != nil {
		return sqlc.Device{}, err
	}

	deviceName := current.DeviceName
	if name != nil {
		deviceName = *name
	}

	return r.q.UpdateDevice(ctx, sqlc.UpdateDeviceParams{
		ID:            id,
		DeviceName:    deviceName,
		HasCamera:     toNullBool(hasCamera),
		HasMicrophone: toNullBool(hasMic),
	})
}

func (r *Repository) UpdateOnlineStatus(ctx context.Context, id uuid.UUID, isOnline bool) error {
	return r.q.UpdateDeviceOnlineStatus(ctx, sqlc.UpdateDeviceOnlineStatusParams{
		ID:       id,
		IsOnline: sql.NullBool{Bool: isOnline, Valid: true},
	})
}

func (r *Repository) UpdateLastSeen(ctx context.Context, id uuid.UUID) error {
	return r.q.UpdateDeviceLastSeen(ctx, id)
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.q.DeleteDevice(ctx, id)
}

func (r *Repository) CountByUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	return r.q.CountUserDevices(ctx, userID)
}

// User settings

func (r *Repository) GetUserSettings(ctx context.Context, userID uuid.UUID) (sqlc.UserSetting, error) {
	return r.q.GetUserSettings(ctx, userID)
}

// Helpers

func toNullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *s, Valid: true}
}

func toNullBool(b *bool) sql.NullBool {
	if b == nil {
		return sql.NullBool{}
	}
	return sql.NullBool{Bool: *b, Valid: true}
}

package auth

import (
	"context"
	"database/sql"
	"time"

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

// User methods

func (r *Repository) GetUserByID(ctx context.Context, id uuid.UUID) (sqlc.User, error) {
	return r.q.GetUserByID(ctx, id)
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (sqlc.User, error) {
	return r.q.GetUserByEmail(ctx, email)
}

func (r *Repository) CreateUser(ctx context.Context, email, passwordHash string, firstName, lastName *string) (sqlc.User, error) {
	return r.q.CreateUser(ctx, sqlc.CreateUserParams{
		Email:        email,
		PasswordHash: passwordHash,
		FirstName:    toNullString(firstName),
		LastName:     toNullString(lastName),
	})
}

func (r *Repository) UpdateUserPassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	return r.q.UpdateUserPassword(ctx, sqlc.UpdateUserPasswordParams{
		ID:           id,
		PasswordHash: passwordHash,
	})
}

// User settings methods

func (r *Repository) CreateUserSettings(ctx context.Context, userID uuid.UUID) (sqlc.UserSetting, error) {
	return r.q.CreateUserSettings(ctx, userID)
}

func (r *Repository) GetUserSettings(ctx context.Context, userID uuid.UUID) (sqlc.UserSetting, error) {
	return r.q.GetUserSettings(ctx, userID)
}

// Session methods

func (r *Repository) CreateSession(ctx context.Context, userID uuid.UUID, deviceID *uuid.UUID, refreshToken string, expiresAt time.Time) (sqlc.Session, error) {
	return r.q.CreateSession(ctx, sqlc.CreateSessionParams{
		UserID:       userID,
		DeviceID:     toNullUUID(deviceID),
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
	})
}

func (r *Repository) GetSessionByToken(ctx context.Context, token string) (sqlc.Session, error) {
	return r.q.GetSessionByToken(ctx, token)
}

func (r *Repository) DeleteSessionByToken(ctx context.Context, token string) error {
	return r.q.DeleteSessionByToken(ctx, token)
}

func (r *Repository) DeleteUserSessions(ctx context.Context, userID uuid.UUID) error {
	return r.q.DeleteUserSessions(ctx, userID)
}

// Password reset methods

func (r *Repository) CreatePasswordReset(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) (sqlc.PasswordReset, error) {
	return r.q.CreatePasswordReset(ctx, sqlc.CreatePasswordResetParams{
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
	})
}

func (r *Repository) GetPasswordResetByToken(ctx context.Context, token string) (sqlc.PasswordReset, error) {
	return r.q.GetPasswordResetByToken(ctx, token)
}

func (r *Repository) MarkPasswordResetUsed(ctx context.Context, id uuid.UUID) error {
	return r.q.MarkPasswordResetUsed(ctx, id)
}

// Helpers

func toNullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *s, Valid: true}
}

func toNullUUID(id *uuid.UUID) uuid.NullUUID {
	if id == nil {
		return uuid.NullUUID{}
	}
	return uuid.NullUUID{UUID: *id, Valid: true}
}

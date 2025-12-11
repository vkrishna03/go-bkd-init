package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/vkrishna03/streamz/db/sqlc"
	apperr "github.com/vkrishna03/streamz/internal/errors"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo              *Repository
	jwtSecret         []byte
	jwtExpiry         time.Duration
	refreshExpiry     time.Duration
	passwordResetExp  time.Duration
}

type Config struct {
	JWTSecret        string
	JWTExpiry        time.Duration
	RefreshExpiry    time.Duration
	PasswordResetExp time.Duration
}

func NewService(repo *Repository, cfg Config) *Service {
	return &Service{
		repo:             repo,
		jwtSecret:        []byte(cfg.JWTSecret),
		jwtExpiry:        cfg.JWTExpiry,
		refreshExpiry:    cfg.RefreshExpiry,
		passwordResetExp: cfg.PasswordResetExp,
	}
}

// Register creates a new user account
func (s *Service) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	// Check if user exists
	_, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err == nil {
		return nil, apperr.Wrap(apperr.ErrConflict, "email already registered")
	}
	if err != sql.ErrNoRows {
		return nil, apperr.Wrap(apperr.ErrInternal, "failed to check existing user")
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apperr.Wrap(apperr.ErrInternal, "failed to hash password")
	}

	// Create user
	user, err := s.repo.CreateUser(ctx, req.Email, string(hash), strPtr(req.FirstName), strPtr(req.LastName))
	if err != nil {
		return nil, apperr.Wrap(apperr.ErrInternal, "failed to create user")
	}

	// Create user settings
	_, err = s.repo.CreateUserSettings(ctx, user.ID)
	if err != nil {
		return nil, apperr.Wrap(apperr.ErrInternal, "failed to create user settings")
	}

	// Generate tokens
	return s.generateAuthResponse(ctx, user, nil)
}

// Login authenticates a user
func (s *Service) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperr.Wrap(apperr.ErrUnauthorized, "invalid credentials")
		}
		return nil, apperr.Wrap(apperr.ErrInternal, "failed to get user")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, apperr.Wrap(apperr.ErrUnauthorized, "invalid credentials")
	}

	return s.generateAuthResponse(ctx, user, nil)
}

// Refresh generates new tokens from a valid refresh token
func (s *Service) Refresh(ctx context.Context, req RefreshRequest) (*AuthResponse, error) {
	session, err := s.repo.GetSessionByToken(ctx, req.RefreshToken)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperr.Wrap(apperr.ErrUnauthorized, "invalid or expired refresh token")
		}
		return nil, apperr.Wrap(apperr.ErrInternal, "failed to get session")
	}

	// Delete old session
	_ = s.repo.DeleteSessionByToken(ctx, req.RefreshToken)

	// Get user
	user, err := s.repo.GetUserByID(ctx, session.UserID)
	if err != nil {
		return nil, apperr.Wrap(apperr.ErrInternal, "failed to get user")
	}

	// Get device ID if present
	var deviceID *uuid.UUID
	if session.DeviceID.Valid {
		deviceID = &session.DeviceID.UUID
	}

	return s.generateAuthResponse(ctx, user, deviceID)
}

// Logout invalidates the refresh token
func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	return s.repo.DeleteSessionByToken(ctx, refreshToken)
}

// ForgotPassword initiates password reset
func (s *Service) ForgotPassword(ctx context.Context, req ForgotPasswordRequest) (*MessageResponse, error) {
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		// Don't reveal if email exists
		return &MessageResponse{Message: "If the email exists, a reset link has been sent"}, nil
	}

	// Generate reset token
	token, err := generateSecureToken(32)
	if err != nil {
		return nil, apperr.Wrap(apperr.ErrInternal, "failed to generate reset token")
	}

	expiresAt := time.Now().Add(s.passwordResetExp)
	_, err = s.repo.CreatePasswordReset(ctx, user.ID, token, expiresAt)
	if err != nil {
		return nil, apperr.Wrap(apperr.ErrInternal, "failed to create password reset")
	}

	// TODO: Send email with reset link
	// For now, just return success message

	return &MessageResponse{Message: "If the email exists, a reset link has been sent"}, nil
}

// ResetPassword completes the password reset
func (s *Service) ResetPassword(ctx context.Context, req ResetPasswordRequest) (*MessageResponse, error) {
	reset, err := s.repo.GetPasswordResetByToken(ctx, req.Token)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperr.Wrap(apperr.ErrValidation, "invalid or expired reset token")
		}
		return nil, apperr.Wrap(apperr.ErrInternal, "failed to get reset token")
	}

	// Hash new password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, apperr.Wrap(apperr.ErrInternal, "failed to hash password")
	}

	// Update password
	if err := s.repo.UpdateUserPassword(ctx, reset.UserID, string(hash)); err != nil {
		return nil, apperr.Wrap(apperr.ErrInternal, "failed to update password")
	}

	// Mark token as used
	_ = s.repo.MarkPasswordResetUsed(ctx, reset.ID)

	// Invalidate all sessions
	_ = s.repo.DeleteUserSessions(ctx, reset.UserID)

	return &MessageResponse{Message: "Password has been reset successfully"}, nil
}

// Helper methods

func (s *Service) generateAuthResponse(ctx context.Context, user sqlc.User, deviceID *uuid.UUID) (*AuthResponse, error) {
	// Generate access token
	expiresAt := time.Now().Add(s.jwtExpiry)
	accessToken, err := s.generateJWT(user.ID, expiresAt)
	if err != nil {
		return nil, apperr.Wrap(apperr.ErrInternal, "failed to generate access token")
	}

	// Generate refresh token
	refreshToken, err := generateSecureToken(32)
	if err != nil {
		return nil, apperr.Wrap(apperr.ErrInternal, "failed to generate refresh token")
	}

	// Save session
	refreshExpiresAt := time.Now().Add(s.refreshExpiry)
	_, err = s.repo.CreateSession(ctx, user.ID, deviceID, refreshToken, refreshExpiresAt)
	if err != nil {
		return nil, apperr.Wrap(apperr.ErrInternal, "failed to create session")
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt.Unix(),
		User:         toUserResponse(user),
	}, nil
}

func (s *Service) generateJWT(userID uuid.UUID, expiresAt time.Time) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID.String(),
		"exp": expiresAt.Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func generateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func toUserResponse(u sqlc.User) UserResponse {
	resp := UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		CreatedAt: u.CreatedAt.Time.Format(time.RFC3339),
	}
	if u.FirstName.Valid {
		resp.FirstName = u.FirstName.String
	}
	if u.LastName.Valid {
		resp.LastName = u.LastName.String
	}
	return resp
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

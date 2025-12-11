package user

import (
	"context"
	"database/sql"

	"github.com/vkrishna03/go-bkd-init/db/sqlc"
	apperr "github.com/vkrishna03/go-bkd-init/internal/errors"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context) ([]Response, error) {
	users, err := s.repo.List(ctx)
	if err != nil {
		return nil, apperr.Wrap(apperr.ErrInternal, "failed to retrieve users")
	}
	return toResponseList(users), nil
}

func (s *Service) GetByID(ctx context.Context, id int32) (*Response, error) {
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperr.Wrap(apperr.ErrNotFound, "user with id %d not found", id)
		}
		return nil, apperr.Wrap(apperr.ErrInternal, "failed to retrieve user")
	}
	resp := toResponse(u)
	return &resp, nil
}

func (s *Service) Create(ctx context.Context, req CreateRequest) (*Response, error) {
	u, err := s.repo.Create(ctx, req.Name, req.Email)
	if err != nil {
		// TODO: check for unique constraint violation → ErrConflict
		return nil, apperr.Wrap(apperr.ErrInternal, "failed to create user")
	}
	resp := toResponse(u)
	return &resp, nil
}

func (s *Service) Delete(ctx context.Context, id int32) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return apperr.Wrap(apperr.ErrInternal, "failed to delete user")
	}
	return nil
}

// Converters
func toResponse(u sqlc.User) Response {
	resp := Response{
		ID:    u.ID,
		Name:  u.Name,
		Email: u.Email,
	}
	if u.CreatedAt.Valid {
		resp.CreatedAt = u.CreatedAt.Time.Format("2006-01-02T15:04:05Z")
	}
	return resp
}

func toResponseList(users []sqlc.User) []Response {
	resp := make([]Response, len(users))
	for i, u := range users {
		resp[i] = toResponse(u)
	}
	return resp
}

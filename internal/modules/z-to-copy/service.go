package ztocopy

import "context"

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// TODO: Add service methods
func (s *Service) GetByID(ctx context.Context, id int32) (*Response, error) {
	// Use service when:
	// - Multiple repos needed
	// - Complex business logic
	// - External API calls
	return nil, nil
}

package ztocopy

import (
	"context"
	"database/sql"

	"github.com/vkrishna03/go-bkd-init/db/sqlc"
)

type Repository struct {
	db *sql.DB
	q  *sqlc.Queries
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db, q: sqlc.New(db)}
}

// TODO: Add repository methods
func (r *Repository) GetByID(ctx context.Context, id int32) (interface{}, error) {
	// return r.q.GetXxx(ctx, id)
	return nil, nil
}

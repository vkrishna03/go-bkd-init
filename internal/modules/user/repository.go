package user

import (
	"context"
	"database/sql"

	"github.com/vkrishna03/streamz/db/sqlc"
)

type Repository struct {
	db *sql.DB
	q  *sqlc.Queries
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db, q: sqlc.New(db)}
}

func (r *Repository) List(ctx context.Context) ([]sqlc.User, error) {
	return r.q.ListUsers(ctx)
}

func (r *Repository) GetByID(ctx context.Context, id int32) (sqlc.User, error) {
	return r.q.GetUser(ctx, id)
}

func (r *Repository) Create(ctx context.Context, name, email string) (sqlc.User, error) {
	return r.q.CreateUser(ctx, sqlc.CreateUserParams{
		Name:  name,
		Email: email,
	})
}

func (r *Repository) Delete(ctx context.Context, id int32) error {
	return r.q.DeleteUser(ctx, id)
}

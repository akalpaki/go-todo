package user

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/noquark/nanoid"
)

var (
	errInsertFailed = errors.New("insert failed")
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool: pool,
	}
}

func (r *Repository) Register(ctx context.Context, data UserRequest) (User, error) {
	var err error
	var u User

	data.Password, err = hashPassword(data.Password)
	if err != nil {
		return User{}, err
	}

	id, err := nanoid.New(21)
	if err != nil {
		return User{}, err
	}
	u.ID = id
	u.Email = data.Email
	u.Password = data.Password

	res, err := r.pool.Exec(ctx, insert, u.ID, u.Email, u.Password)
	if res.RowsAffected() == 0 {
		return User{}, errInsertFailed
	}
	if err != nil {
		return User{}, err
	}
	return u, nil
}

func (r *Repository) GetByEmail(ctx context.Context, email string) (User, error) {
	var err error
	var u User

	row := r.pool.QueryRow(ctx, queryByEmail, email)
	if err = row.Scan(&u.ID, &u.Email, &u.Password); err != nil {
		return User{}, err
	}
	return u, nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (User, error) {
	var err error
	var u User

	row := r.pool.QueryRow(ctx, queryByID, id)
	if err = row.Scan(&u.ID, &u.Email, &u.Password); err != nil {
		return User{}, err
	}
	return u, nil
}

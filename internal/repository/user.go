package repository

import (
	"context"
	"database/sql"

	"github.com/yeboahd24/personal-finance-manager/internal/model"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *model.User) error
	GetUserByID(ctx context.Context, id string) (*model.User, error)
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	GetUserByUsername(ctx context.Context, username string) (*model.User, error)
	UpdateUser(ctx context.Context, user *model.User) error
	DeleteUser(ctx context.Context, id string) error
	GetUserCount(ctx context.Context) (int64, error)
}

type UserSQL struct {
	db *sql.DB
	tx *sql.Tx
}

func (r *UserSQL) query() QueryExecutor {
	if r.tx != nil {
		return r.tx
	}
	return r.db
}

func (r *UserSQL) CreateUser(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (email, username, password_hash, first_name, last_name)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at`
	
	return r.query().QueryRowContext(
		ctx,
		query,
		user.Email,
		user.Username,
		user.PasswordHash,
		user.FirstName,
		user.LastName,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

func (r *UserSQL) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	user := &model.User{}
	query := `
		SELECT id, email, username, password_hash, first_name, last_name, created_at, updated_at
		FROM users
		WHERE id = $1`
	
	err := r.query().QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserSQL) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	user := &model.User{}
	query := `
		SELECT id, email, username, password_hash, first_name, last_name, created_at, updated_at
		FROM users
		WHERE email = $1`
	
	err := r.query().QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserSQL) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	user := &model.User{}
	query := `
		SELECT id, email, username, password_hash, first_name, last_name, created_at, updated_at
		FROM users
		WHERE username = $1`
	
	err := r.query().QueryRowContext(ctx, query, username).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserSQL) UpdateUser(ctx context.Context, user *model.User) error {
	query := `
		UPDATE users
		SET first_name = $2, last_name = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING updated_at`
	
	return r.query().QueryRowContext(
		ctx,
		query,
		user.ID,
		user.FirstName,
		user.LastName,
	).Scan(&user.UpdatedAt)
}

func (r *UserSQL) DeleteUser(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = $1`
	result, err := r.query().ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	
	return nil
}

// GetUserCount returns the total number of users in the system
func (r *UserSQL) GetUserCount(ctx context.Context) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM users`
	err := r.query().QueryRowContext(ctx, query).Scan(&count)
	return count, err
}

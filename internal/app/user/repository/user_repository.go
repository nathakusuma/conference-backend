package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/nathakusuma/astungkara/domain/contract"
	"github.com/nathakusuma/astungkara/domain/entity"
)

type userRepository struct {
	conn *sqlx.DB
}

func NewUserRepository(conn *sqlx.DB) contract.IUserRepository {
	return &userRepository{
		conn: conn,
	}
}

func (r *userRepository) CreateUser(ctx context.Context, user *entity.User) error {
	return r.createUser(ctx, r.conn, user)
}

func (r *userRepository) createUser(ctx context.Context, tx sqlx.ExtContext, user *entity.User) error {
	_, err := sqlx.NamedExecContext(
		ctx,
		tx,
		`INSERT INTO users (
                   id, name, password_hash, role, email
                   ) VALUES (:id, :name, :password_hash, :role, :email)`,
		user,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *userRepository) GetUserByField(ctx context.Context, field, value string) (*entity.User, error) {
	var user entity.User

	statement := `SELECT
			id,
			name,
			email,
			password_hash,
			role,
			bio,
			created_at,
			updated_at,
			deleted_at
		FROM users
		WHERE ` + field + ` = $1
		AND deleted_at IS NULL
		`

	err := r.conn.GetContext(ctx, &user, statement, value)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) updateUser(ctx context.Context, tx sqlx.ExtContext, user *entity.User) error {
	_, err := sqlx.NamedExecContext(
		ctx,
		tx,
		`UPDATE users
		SET name = :name,
			email = :email,
			password_hash = :password_hash,
			role = :role,
			bio = :bio,
			updated_at = now()
		WHERE id = :id`,
		user,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *userRepository) UpdateUser(ctx context.Context, user *entity.User) error {
	return r.updateUser(ctx, r.conn, user)
}

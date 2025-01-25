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
                   id, name, password_hash, role, email, auth_method
                   ) VALUES (:id, :name, :password_hash, :role, :email, :auth_method)`,
		user,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *userRepository) GetUserByField(ctx context.Context, field, value string) (*entity.User, error) {
	return nil, nil
}

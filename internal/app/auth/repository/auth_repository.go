package repository

import (
	"context"
	"github.com/jmoiron/sqlx"
	"github.com/nathakusuma/astungkara/domain/contract"
	"github.com/nathakusuma/astungkara/domain/entity"
	"github.com/redis/go-redis/v9"
	"time"
)

type authRepository struct {
	db  *sqlx.DB
	rds *redis.Client
}

func NewAuthRepository(db *sqlx.DB, rds *redis.Client) contract.IAuthRepository {
	return &authRepository{
		db:  db,
		rds: rds,
	}
}

func (r *authRepository) SetUserRegisterOTP(ctx context.Context, email string, otp string) error {
	return r.rds.Set(ctx, "auth:"+email+":register_otp", otp, 10*time.Minute).Err()
}

func (r *authRepository) GetUserRegisterOTP(ctx context.Context, email string) (string, error) {
	return r.rds.Get(ctx, "auth:"+email+":register_otp").Result()
}

func (r *authRepository) DeleteUserRegisterOTP(ctx context.Context, email string) error {
	return r.rds.Del(ctx, "auth:"+email+":register_otp").Err()
}

func (r *authRepository) CreateSession(ctx context.Context, session *entity.Session) error {
	return r.createSession(ctx, r.db, session)
}

func (r *authRepository) createSession(ctx context.Context, tx sqlx.ExtContext, session *entity.Session) error {
	query := `INSERT INTO sessions (user_id, token, expires_at)
				VALUES (:user_id, :token, :expires_at)
				ON CONFLICT (user_id) DO UPDATE SET token = :token, expires_at = :expires_at`

	_, err := sqlx.NamedExecContext(ctx, tx, query, session)
	if err != nil {
		return err
	}

	return nil
}

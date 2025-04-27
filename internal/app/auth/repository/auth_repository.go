package repository

import (
	"context"
	"github.com/jmoiron/sqlx"
	"github.com/nathakusuma/astungkara/domain/contract"
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
	return r.rds.Set(ctx, "auth:"+email+":otp", otp, 10*time.Minute).Err()
}

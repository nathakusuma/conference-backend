package contract

import (
	"context"
	"github.com/nathakusuma/astungkara/domain/dto"
	"github.com/nathakusuma/astungkara/domain/entity"
)

type IAuthRepository interface {
	SetUserRegisterOTP(ctx context.Context, email, otp string) error
	GetUserRegisterOTP(ctx context.Context, email string) (string, error)

	CreateSession(ctx context.Context, session *entity.Session) error
}

type IAuthService interface {
	RequestOTPRegisterUser(ctx context.Context, email string) error
	CheckOTPRegisterUser(ctx context.Context, email, otp string) error
	LoginUser(ctx context.Context, req dto.LoginUserRequest) (dto.LoginResponse, error)
}

package contract

import (
	"context"
	"github.com/nathakusuma/astungkara/domain/dto"
	"github.com/nathakusuma/astungkara/domain/entity"
)

type IAuthRepository interface {
	SetUserRegisterOTP(ctx context.Context, email, otp string) error
	GetUserRegisterOTP(ctx context.Context, email string) (string, error)
	DeleteUserRegisterOTP(ctx context.Context, email string) error

	CreateSession(ctx context.Context, session *entity.Session) error
	GetSessionByToken(ctx context.Context, token string) (*entity.Session, error)
}

type IAuthService interface {
	RequestOTPRegisterUser(ctx context.Context, email string) error
	CheckOTPRegisterUser(ctx context.Context, email, otp string) error
	RegisterUser(ctx context.Context, req dto.RegisterUserRequest) (dto.LoginResponse, error)
	LoginUser(ctx context.Context, req dto.LoginUserRequest) (dto.LoginResponse, error)

	RefreshToken(ctx context.Context, refreshToken string) (dto.LoginResponse, error)
}

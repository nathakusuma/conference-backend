package contract

import (
	"context"
	"github.com/google/uuid"
	"github.com/nathakusuma/astungkara/domain/dto"
	"github.com/nathakusuma/astungkara/domain/entity"
)

type IAuthRepository interface {
	SetUserRegisterOTP(ctx context.Context, email, otp string) error
	GetUserRegisterOTP(ctx context.Context, email string) (string, error)
	DeleteUserRegisterOTP(ctx context.Context, email string) error

	CreateSession(ctx context.Context, session *entity.Session) error
	GetSessionByToken(ctx context.Context, token string) (*entity.Session, error)
	DeleteSession(ctx context.Context, userID uuid.UUID) error

	SetUserResetPasswordOTP(ctx context.Context, email, otp string) error
	GetUserResetPasswordOTP(ctx context.Context, email string) (string, error)
	DeleteUserResetPasswordOTP(ctx context.Context, email string) error
}

type IAuthService interface {
	RequestOTPRegisterUser(ctx context.Context, email string) error
	CheckOTPRegisterUser(ctx context.Context, email, otp string) error
	RegisterUser(ctx context.Context, req dto.RegisterUserRequest) (dto.LoginResponse, error)
	LoginUser(ctx context.Context, req dto.LoginUserRequest) (dto.LoginResponse, error)

	RefreshToken(ctx context.Context, refreshToken string) (dto.LoginResponse, error)
	Logout(ctx context.Context) error

	RequestOTPResetPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, req dto.ResetPasswordRequest) (dto.LoginResponse, error)
}

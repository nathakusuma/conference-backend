package contract

import (
	"context"
)

type IAuthRepository interface {
	SetUserRegisterOTP(ctx context.Context, email, otp string) error
}

type IAuthService interface {
	RequestOTPRegisterUser(ctx context.Context, email string) error
}

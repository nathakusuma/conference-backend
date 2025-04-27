package service

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/nathakusuma/astungkara/domain/contract"
	"github.com/nathakusuma/astungkara/domain/entity"
	"github.com/nathakusuma/astungkara/domain/errorpkg"
	"github.com/nathakusuma/astungkara/internal/app/auth/service"
	appmocks "github.com/nathakusuma/astungkara/test/unit/mocks/app"
	pkgmocks "github.com/nathakusuma/astungkara/test/unit/mocks/pkg"
	_ "github.com/nathakusuma/astungkara/test/unit/setup"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

type authServiceMocks struct {
	authRepo *appmocks.MockIAuthRepository
	userSvc  *appmocks.MockIUserService
	bcrypt   *pkgmocks.MockIBcrypt
	jwt      *pkgmocks.MockIJwt
	mailer   *pkgmocks.MockIMailer
	uuid     *pkgmocks.MockIUUID
}

func setupAuthServiceMocks(t *testing.T) (contract.IAuthService, *authServiceMocks) {
	mocks := &authServiceMocks{
		authRepo: appmocks.NewMockIAuthRepository(t),
		userSvc:  appmocks.NewMockIUserService(t),
		bcrypt:   pkgmocks.NewMockIBcrypt(t),
		jwt:      pkgmocks.NewMockIJwt(t),
		mailer:   pkgmocks.NewMockIMailer(t),
		uuid:     pkgmocks.NewMockIUUID(t),
	}

	svc := service.NewAuthService(mocks.authRepo, mocks.userSvc, mocks.bcrypt, mocks.jwt, mocks.mailer, mocks.uuid)

	return svc, mocks
}

func Test_AuthService_RequestOTPRegisterUser(t *testing.T) {
	ctx := context.Background()
	email := "test@example.com"

	t.Run("success", func(t *testing.T) {
		svc, mocks := setupAuthServiceMocks(t)
		emailSent := make(chan struct{}, 1)

		// Expect user not found (which is good for registration)
		mocks.userSvc.EXPECT().
			GetUserByEmail(ctx, email).
			Return(nil, errorpkg.ErrNotFound)

		// Expect OTP to be set
		mocks.authRepo.EXPECT().
			SetUserRegisterOTP(ctx, email, mock.AnythingOfType("string")).
			Return(nil)

		// Mock email sending with channel notification
		mocks.mailer.EXPECT().
			Send(
				email,
				"[Class Manager] Verify Your Account",
				"otp_register_user.html",
				mock.AnythingOfType("map[string]interface {}"),
			).RunAndReturn(func(_, _, _ string, _ map[string]interface{}) error {
			emailSent <- struct{}{}
			return nil
		})

		err := svc.RequestOTPRegisterUser(ctx, email)
		assert.NoError(t, err)

		// Wait for email sending goroutine to complete
		<-emailSent
	})

	t.Run("error - email already registered", func(t *testing.T) {
		svc, mocks := setupAuthServiceMocks(t)

		// Return existing user
		mocks.userSvc.EXPECT().
			GetUserByEmail(ctx, email).
			Return(&entity.User{ID: uuid.New()}, nil)

		err := svc.RequestOTPRegisterUser(ctx, email)
		assert.ErrorIs(t, err, errorpkg.ErrEmailAlreadyRegistered)
	})

	t.Run("error - get user unexpected error", func(t *testing.T) {
		svc, mocks := setupAuthServiceMocks(t)

		mocks.userSvc.EXPECT().
			GetUserByEmail(ctx, email).
			Return(nil, errors.New("unexpected error"))

		err := svc.RequestOTPRegisterUser(ctx, email)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), errorpkg.ErrInternalServer.Error())
	})

	t.Run("error - set OTP fails", func(t *testing.T) {
		svc, mocks := setupAuthServiceMocks(t)

		mocks.userSvc.EXPECT().
			GetUserByEmail(ctx, email).
			Return(nil, errorpkg.ErrNotFound)

		mocks.authRepo.EXPECT().
			SetUserRegisterOTP(ctx, email, mock.AnythingOfType("string")).
			Return(errors.New("redis error"))

		err := svc.RequestOTPRegisterUser(ctx, email)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), errorpkg.ErrInternalServer.Error())
	})
}

func Test_AuthService_CheckOTPRegisterUser(t *testing.T) {
	ctx := context.Background()
	email := "test@example.com"
	otp := "123456"

	t.Run("success", func(t *testing.T) {
		svc, mocks := setupAuthServiceMocks(t)

		mocks.authRepo.EXPECT().
			GetUserRegisterOTP(ctx, email).
			Return(otp, nil)

		err := svc.CheckOTPRegisterUser(ctx, email, otp)
		assert.NoError(t, err)
	})

	t.Run("error - OTP not found", func(t *testing.T) {
		svc, mocks := setupAuthServiceMocks(t)

		mocks.authRepo.EXPECT().
			GetUserRegisterOTP(ctx, email).
			Return("", redis.Nil)

		err := svc.CheckOTPRegisterUser(ctx, email, otp)
		assert.ErrorIs(t, err, errorpkg.ErrInvalidOTP)
	})

	t.Run("error - get OTP fails", func(t *testing.T) {
		svc, mocks := setupAuthServiceMocks(t)

		mocks.authRepo.EXPECT().
			GetUserRegisterOTP(ctx, email).
			Return("", errors.New("redis error"))

		err := svc.CheckOTPRegisterUser(ctx, email, otp)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), errorpkg.ErrInternalServer.Error())
	})

	t.Run("error - invalid OTP", func(t *testing.T) {
		svc, mocks := setupAuthServiceMocks(t)

		mocks.authRepo.EXPECT().
			GetUserRegisterOTP(ctx, email).
			Return("654321", nil)

		err := svc.CheckOTPRegisterUser(ctx, email, otp)
		assert.ErrorIs(t, err, errorpkg.ErrInvalidOTP)
	})
}

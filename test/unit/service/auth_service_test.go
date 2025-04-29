package service

import (
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"github.com/nathakusuma/astungkara/domain/contract"
	"github.com/nathakusuma/astungkara/domain/dto"
	"github.com/nathakusuma/astungkara/domain/entity"
	"github.com/nathakusuma/astungkara/domain/enum"
	"github.com/nathakusuma/astungkara/domain/errorpkg"
	"github.com/nathakusuma/astungkara/internal/app/auth/service"
	appmocks "github.com/nathakusuma/astungkara/test/unit/mocks/app"
	pkgmocks "github.com/nathakusuma/astungkara/test/unit/mocks/pkg"
	_ "github.com/nathakusuma/astungkara/test/unit/setup"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
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
			SetOTPRegisterUser(ctx, email, mock.AnythingOfType("string")).
			Return(nil)

		// Mock email sending with channel notification
		mocks.mailer.EXPECT().
			Send(
				email,
				"[Astungkara] Verify Your Account",
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

	t.Run("error - email sending fails", func(t *testing.T) {
		svc, mocks := setupAuthServiceMocks(t)
		emailSent := make(chan struct{}, 1)

		// Expect user not found (which is good for registration)
		mocks.userSvc.EXPECT().
			GetUserByEmail(ctx, email).
			Return(nil, errorpkg.ErrNotFound)

		// Expect OTP to be set
		mocks.authRepo.EXPECT().
			SetOTPRegisterUser(ctx, email, mock.AnythingOfType("string")).
			Return(nil)

		// Mock email sending to fail
		mocks.mailer.EXPECT().
			Send(
				email,
				"[Astungkara] Verify Your Account",
				"otp_register_user.html",
				mock.AnythingOfType("map[string]interface {}"),
			).RunAndReturn(func(_, _, _ string, _ map[string]interface{}) error {
			emailSent <- struct{}{}
			return errors.New("email sending error")
		})

		err := svc.RequestOTPRegisterUser(ctx, email)
		assert.NoError(t, err)

		// Wait for email sending goroutine to complete
		<-emailSent

		// It should not return an error because the email sending is done in a goroutine
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
		assert.ErrorIs(t, err, errorpkg.ErrInternalServer)
	})

	t.Run("error - set OTP fails", func(t *testing.T) {
		svc, mocks := setupAuthServiceMocks(t)

		mocks.userSvc.EXPECT().
			GetUserByEmail(ctx, email).
			Return(nil, errorpkg.ErrNotFound)

		mocks.authRepo.EXPECT().
			SetOTPRegisterUser(ctx, email, mock.AnythingOfType("string")).
			Return(errors.New("redis error"))

		err := svc.RequestOTPRegisterUser(ctx, email)
		assert.Error(t, err)
		assert.ErrorIs(t, err, errorpkg.ErrInternalServer)
	})
}

func Test_AuthService_CheckOTPRegisterUser(t *testing.T) {
	ctx := context.Background()
	email := "test@example.com"
	otp := "123456"

	t.Run("success", func(t *testing.T) {
		svc, mocks := setupAuthServiceMocks(t)

		mocks.authRepo.EXPECT().
			GetOTPRegisterUser(ctx, email).
			Return(otp, nil)

		err := svc.CheckOTPRegisterUser(ctx, email, otp)
		assert.NoError(t, err)
	})

	t.Run("error - OTP not found", func(t *testing.T) {
		svc, mocks := setupAuthServiceMocks(t)

		mocks.authRepo.EXPECT().
			GetOTPRegisterUser(ctx, email).
			Return("", redis.Nil)

		err := svc.CheckOTPRegisterUser(ctx, email, otp)
		assert.ErrorIs(t, err, errorpkg.ErrInvalidOTP)
	})

	t.Run("error - get OTP fails", func(t *testing.T) {
		svc, mocks := setupAuthServiceMocks(t)

		mocks.authRepo.EXPECT().
			GetOTPRegisterUser(ctx, email).
			Return("", errors.New("redis error"))

		err := svc.CheckOTPRegisterUser(ctx, email, otp)
		assert.Error(t, err)
		assert.ErrorIs(t, err, errorpkg.ErrInternalServer)
	})

	t.Run("error - invalid OTP", func(t *testing.T) {
		svc, mocks := setupAuthServiceMocks(t)

		mocks.authRepo.EXPECT().
			GetOTPRegisterUser(ctx, email).
			Return("654321", nil)

		err := svc.CheckOTPRegisterUser(ctx, email, otp)
		assert.ErrorIs(t, err, errorpkg.ErrInvalidOTP)
	})
}

func Test_AuthService_LoginUser(t *testing.T) {
	ctx := context.Background()
	req := dto.LoginUserRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	t.Run("success", func(t *testing.T) {
		svc, mocks := setupAuthServiceMocks(t)

		userID := uuid.New()
		user := &entity.User{
			ID:           userID,
			Email:        req.Email,
			PasswordHash: req.Password,
			Role:         enum.RoleUser,
		}

		mockLoginExpectations(mocks, ctx, req.Email, req.Password, userID)

		resp, err := svc.Login(ctx, req)
		assert.NoError(t, err)
		assert.NotEmpty(t, resp.AccessToken)
		assert.NotEmpty(t, resp.RefreshToken)
		assert.Equal(t, user.Email, resp.User.Email)
	})

	t.Run("error - user not found", func(t *testing.T) {
		svc, mocks := setupAuthServiceMocks(t)

		mocks.userSvc.EXPECT().
			GetUserByEmail(ctx, req.Email).
			Return(nil, errorpkg.ErrNotFound)

		resp, err := svc.Login(ctx, req)
		assert.Empty(t, resp)
		assert.ErrorIs(t, err, errorpkg.ErrNotFound)
	})

	t.Run("error - get user unexpected error", func(t *testing.T) {
		svc, mocks := setupAuthServiceMocks(t)

		mocks.userSvc.EXPECT().
			GetUserByEmail(ctx, req.Email).
			Return(nil, errors.New("db error"))

		resp, err := svc.Login(ctx, req)
		assert.Empty(t, resp)
		assert.Error(t, err)
		assert.ErrorIs(t, err, errorpkg.ErrInternalServer)
	})

	t.Run("error - invalid credentials", func(t *testing.T) {
		svc, mocks := setupAuthServiceMocks(t)

		passwordHash := "hashed_password"
		user := &entity.User{
			Email:        req.Email,
			PasswordHash: passwordHash,
		}

		mocks.userSvc.EXPECT().
			GetUserByEmail(ctx, req.Email).
			Return(user, nil)

		mocks.bcrypt.EXPECT().
			Compare(req.Password, passwordHash).
			Return(false)

		resp, err := svc.Login(ctx, req)
		assert.Empty(t, resp)
		assert.ErrorIs(t, err, errorpkg.ErrCredentialsNotMatch)
	})

	t.Run("error - jwt creation fails", func(t *testing.T) {
		svc, mocks := setupAuthServiceMocks(t)

		passwordHash := "hashed_password"
		user := &entity.User{
			ID:           uuid.New(),
			Email:        req.Email,
			PasswordHash: passwordHash,
			Role:         enum.RoleUser,
		}

		// Setup expectations
		mocks.userSvc.EXPECT().
			GetUserByEmail(ctx, req.Email).
			Return(user, nil)

		mocks.bcrypt.EXPECT().
			Compare(req.Password, passwordHash).
			Return(true)

		// JWT creation will fail
		mocks.jwt.EXPECT().
			Create(user.ID, user.Role).
			Return("", errors.New("jwt error"))

		// Expect CreateAuthSession to be called but we don't care about the result
		// since the JWT error should be returned first
		mocks.authRepo.EXPECT().
			CreateAuthSession(ctx, mock.MatchedBy(func(authSession *entity.AuthSession) bool {
				return authSession.UserID == user.ID &&
					len(authSession.Token) == 32 &&
					!authSession.ExpiresAt.IsZero()
			})).
			Return(nil).
			Maybe() // This may or may not be called depending on goroutine scheduling

		resp, err := svc.Login(ctx, req)
		assert.Empty(t, resp)
		assert.Error(t, err)
		assert.ErrorIs(t, err, errorpkg.ErrInternalServer)
	})

	t.Run("error - create auth session fails", func(t *testing.T) {
		svc, mocks := setupAuthServiceMocks(t)

		passwordHash := "hashed_password"
		user := &entity.User{
			ID:           uuid.New(),
			Email:        req.Email,
			PasswordHash: passwordHash,
			Role:         enum.RoleUser,
		}

		// Setup expectations
		mocks.userSvc.EXPECT().
			GetUserByEmail(ctx, req.Email).
			Return(user, nil)

		mocks.bcrypt.EXPECT().
			Compare(req.Password, passwordHash).
			Return(true)

		// JWT creation succeeds
		mocks.jwt.EXPECT().
			Create(user.ID, user.Role).
			Return("access_token", nil)

		// AuthSession creation fails
		mocks.authRepo.EXPECT().
			CreateAuthSession(ctx, mock.MatchedBy(func(authSession *entity.AuthSession) bool {
				return authSession.UserID == user.ID &&
					len(authSession.Token) == 32 &&
					!authSession.ExpiresAt.IsZero()
			})).
			Return(errors.New("db error"))

		resp, err := svc.Login(ctx, req)
		assert.Empty(t, resp)
		assert.Error(t, err)
		assert.ErrorIs(t, err, errorpkg.ErrInternalServer)
	})

	t.Run("both operations fail", func(t *testing.T) {
		svc, mocks := setupAuthServiceMocks(t)

		passwordHash := "hashed_password"
		user := &entity.User{
			ID:           uuid.New(),
			Email:        req.Email,
			PasswordHash: passwordHash,
			Role:         enum.RoleUser,
		}

		// Setup expectations
		mocks.userSvc.EXPECT().
			GetUserByEmail(ctx, req.Email).
			Return(user, nil)

		mocks.bcrypt.EXPECT().
			Compare(req.Password, passwordHash).
			Return(true)

		// Both operations fail
		mocks.jwt.EXPECT().
			Create(user.ID, user.Role).
			Return("", errors.New("jwt error"))

		mocks.authRepo.EXPECT().
			CreateAuthSession(ctx, mock.MatchedBy(func(authSession *entity.AuthSession) bool {
				return authSession.UserID == user.ID &&
					len(authSession.Token) == 32 &&
					!authSession.ExpiresAt.IsZero()
			})).
			Return(errors.New("db error")).
			Maybe() // This may or may not be called depending on goroutine scheduling

		resp, err := svc.Login(ctx, req)
		assert.Empty(t, resp)
		assert.Error(t, err)
		assert.ErrorIs(t, err, errorpkg.ErrInternalServer)
	})
}

func Test_AuthService_RegisterUser(t *testing.T) {
	ctx := context.Background()
	req := dto.RegisterUserRequest{
		Email:    "test@example.com",
		OTP:      "123456",
		Name:     "Test User",
		Password: "password123",
	}

	t.Run("success", func(t *testing.T) {
		svc, mocks := setupAuthServiceMocks(t)

		// Setup expectations
		mocks.authRepo.EXPECT().
			GetOTPRegisterUser(ctx, req.Email).
			Return(req.OTP, nil)

		mocks.authRepo.EXPECT().
			DeleteOTPRegisterUser(ctx, req.Email).
			Return(nil)

		userID := uuid.New()
		mocks.userSvc.EXPECT().
			CreateUser(ctx, &dto.CreateUserRequest{
				Name:     req.Name,
				Email:    req.Email,
				Password: req.Password,
				Role:     enum.RoleUser,
			}).Return(userID, nil)

		// Mock login expectations
		mockLoginExpectations(mocks, ctx, req.Email, req.Password, userID)

		resp, err := svc.RegisterUser(ctx, req)
		assert.NoError(t, err)
		assert.NotEmpty(t, resp.AccessToken)
		assert.NotEmpty(t, resp.RefreshToken)
		assert.NotNil(t, resp.User)
	})

	t.Run("error - no OTP found", func(t *testing.T) {
		svc, mocks := setupAuthServiceMocks(t)

		mocks.authRepo.EXPECT().
			GetOTPRegisterUser(ctx, req.Email).
			Return("", redis.Nil)

		resp, err := svc.RegisterUser(ctx, req)
		assert.Empty(t, resp)
		assert.ErrorIs(t, err, errorpkg.ErrInvalidOTP)
	})

	t.Run("error - invalid OTP", func(t *testing.T) {
		svc, mocks := setupAuthServiceMocks(t)

		mocks.authRepo.EXPECT().
			GetOTPRegisterUser(ctx, req.Email).
			Return("different-otp", nil)

		resp, err := svc.RegisterUser(ctx, req)
		assert.Empty(t, resp)
		assert.ErrorIs(t, err, errorpkg.ErrInvalidOTP)
	})

	t.Run("error - get OTP fails", func(t *testing.T) {
		svc, mocks := setupAuthServiceMocks(t)

		mocks.authRepo.EXPECT().
			GetOTPRegisterUser(ctx, req.Email).
			Return("", errors.New("redis error"))

		resp, err := svc.RegisterUser(ctx, req)
		assert.Empty(t, resp)
		assert.Error(t, err)
		assert.ErrorIs(t, err, errorpkg.ErrInternalServer)
	})

	t.Run("error - delete OTP fails", func(t *testing.T) {
		svc, mocks := setupAuthServiceMocks(t)

		mocks.authRepo.EXPECT().
			GetOTPRegisterUser(ctx, req.Email).
			Return(req.OTP, nil)

		mocks.authRepo.EXPECT().
			DeleteOTPRegisterUser(ctx, req.Email).
			Return(errors.New("redis error"))

		resp, err := svc.RegisterUser(ctx, req)
		assert.Empty(t, resp)
		assert.Error(t, err)
		assert.ErrorIs(t, err, errorpkg.ErrInternalServer)
	})

	t.Run("error - create user fails", func(t *testing.T) {
		svc, mocks := setupAuthServiceMocks(t)

		mocks.authRepo.EXPECT().
			GetOTPRegisterUser(ctx, req.Email).
			Return(req.OTP, nil)

		mocks.authRepo.EXPECT().
			DeleteOTPRegisterUser(ctx, req.Email).
			Return(nil)

		mocks.userSvc.EXPECT().
			CreateUser(ctx, mock.Anything).
			Return(uuid.UUID{}, errors.New("db error"))

		resp, err := svc.RegisterUser(ctx, req)
		assert.Empty(t, resp)
		assert.Error(t, err)
		assert.ErrorIs(t, err, errorpkg.ErrInternalServer)
	})
}

// Helper function to set up common login expectations
func mockLoginExpectations(mocks *authServiceMocks, ctx context.Context, email, password string, userID uuid.UUID) {
	passwordHash := "hashed_password"
	user := &entity.User{
		ID:           userID,
		Email:        email,
		PasswordHash: passwordHash,
		Role:         enum.RoleUser,
		Name:         "Test User",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	mocks.userSvc.EXPECT().
		GetUserByEmail(ctx, email).
		Return(user, nil)

	mocks.bcrypt.EXPECT().
		Compare(password, passwordHash).
		Return(true)

	mocks.jwt.EXPECT().
		Create(user.ID, user.Role).
		Return("access_token", nil)

	mocks.authRepo.EXPECT().
		CreateAuthSession(ctx, mock.MatchedBy(func(authSession *entity.AuthSession) bool {
			return authSession.UserID == user.ID &&
				len(authSession.Token) == 32 && // Check refresh token length
				!authSession.ExpiresAt.IsZero() // Check expiration is set
		})).
		Return(nil)
}

func Test_AuthService_RefreshToken(t *testing.T) {
	ctx := context.Background()
	refreshToken := "valid_refresh_token"
	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		svc, mocks := setupAuthServiceMocks(t)

		// Setup auth session
		authSession := &entity.AuthSession{
			Token:     refreshToken,
			UserID:    userID,
			ExpiresAt: time.Now().Add(time.Hour), // Valid future expiration
		}

		// Setup user
		user := &entity.User{
			ID:        userID,
			Email:     "test@example.com",
			Name:      "Test User",
			Role:      enum.RoleUser,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Set up expectations
		mocks.authRepo.EXPECT().
			GetAuthSessionByToken(ctx, refreshToken).
			Return(authSession, nil)

		mocks.userSvc.EXPECT().
			GetUserByID(ctx, userID).
			Return(user, nil)

		mocks.jwt.EXPECT().
			Create(user.ID, user.Role).
			Return("new_access_token", nil)

		// Execute and verify
		resp, err := svc.RefreshToken(ctx, refreshToken)
		assert.NoError(t, err)
		assert.NotEmpty(t, resp.AccessToken)
		assert.Equal(t, refreshToken, resp.RefreshToken)
		assert.NotNil(t, resp.User)
		assert.Equal(t, user.ID, resp.User.ID)
		assert.Equal(t, user.Email, resp.User.Email)
		assert.Equal(t, user.Role, resp.User.Role)
	})

	t.Run("error - auth session not found", func(t *testing.T) {
		svc, mocks := setupAuthServiceMocks(t)

		mocks.authRepo.EXPECT().
			GetAuthSessionByToken(ctx, refreshToken).
			Return(nil, sql.ErrNoRows)

		resp, err := svc.RefreshToken(ctx, refreshToken)
		assert.Empty(t, resp)
		assert.ErrorIs(t, err, errorpkg.ErrInvalidRefreshToken)
	})

	t.Run("error - get auth session unexpected error", func(t *testing.T) {
		svc, mocks := setupAuthServiceMocks(t)

		mocks.authRepo.EXPECT().
			GetAuthSessionByToken(ctx, refreshToken).
			Return(nil, errors.New("db error"))

		resp, err := svc.RefreshToken(ctx, refreshToken)
		assert.Empty(t, resp)
		assert.Error(t, err)
		assert.ErrorIs(t, err, errorpkg.ErrInternalServer)
	})

	t.Run("error - expired auth session", func(t *testing.T) {
		svc, mocks := setupAuthServiceMocks(t)

		// Setup expired auth session
		authSession := &entity.AuthSession{
			Token:     refreshToken,
			UserID:    userID,
			ExpiresAt: time.Now().Add(-time.Hour), // Expired
		}

		mocks.authRepo.EXPECT().
			GetAuthSessionByToken(ctx, refreshToken).
			Return(authSession, nil)

		resp, err := svc.RefreshToken(ctx, refreshToken)
		assert.Empty(t, resp)
		assert.ErrorIs(t, err, errorpkg.ErrInvalidRefreshToken)
	})

	t.Run("error - user not found", func(t *testing.T) {
		svc, mocks := setupAuthServiceMocks(t)

		// Setup valid auth session
		authSession := &entity.AuthSession{
			Token:     refreshToken,
			UserID:    userID,
			ExpiresAt: time.Now().Add(time.Hour),
		}

		mocks.authRepo.EXPECT().
			GetAuthSessionByToken(ctx, refreshToken).
			Return(authSession, nil)

		mocks.userSvc.EXPECT().
			GetUserByID(ctx, userID).
			Return(nil, errorpkg.ErrNotFound)

		resp, err := svc.RefreshToken(ctx, refreshToken)
		assert.Empty(t, resp)
		assert.ErrorIs(t, err, errorpkg.ErrNotFound)
	})

	t.Run("error - get user unexpected error", func(t *testing.T) {
		svc, mocks := setupAuthServiceMocks(t)

		// Setup valid auth session
		authSession := &entity.AuthSession{
			Token:     refreshToken,
			UserID:    userID,
			ExpiresAt: time.Now().Add(time.Hour),
		}

		mocks.authRepo.EXPECT().
			GetAuthSessionByToken(ctx, refreshToken).
			Return(authSession, nil)

		mocks.userSvc.EXPECT().
			GetUserByID(ctx, userID).
			Return(nil, errors.New("db error"))

		resp, err := svc.RefreshToken(ctx, refreshToken)
		assert.Empty(t, resp)
		assert.Error(t, err)
		assert.ErrorIs(t, err, errorpkg.ErrInternalServer)
	})

	t.Run("error - jwt creation fails", func(t *testing.T) {
		svc, mocks := setupAuthServiceMocks(t)

		// Setup valid auth session
		authSession := &entity.AuthSession{
			Token:     refreshToken,
			UserID:    userID,
			ExpiresAt: time.Now().Add(time.Hour),
		}

		// Setup user
		user := &entity.User{
			ID:   userID,
			Role: enum.RoleUser,
		}

		mocks.authRepo.EXPECT().
			GetAuthSessionByToken(ctx, refreshToken).
			Return(authSession, nil)

		mocks.userSvc.EXPECT().
			GetUserByID(ctx, userID).
			Return(user, nil)

		mocks.jwt.EXPECT().
			Create(user.ID, user.Role).
			Return("", errors.New("jwt error"))

		resp, err := svc.RefreshToken(ctx, refreshToken)
		assert.Empty(t, resp)
		assert.Error(t, err)
		assert.ErrorIs(t, err, errorpkg.ErrInternalServer)
	})
}

func Test_AuthService_Logout(t *testing.T) {
	userID := uuid.New()
	ctx := context.WithValue(context.Background(), "user.id", userID)

	t.Run("success", func(t *testing.T) {
		svc, mocks := setupAuthServiceMocks(t)

		mocks.authRepo.EXPECT().
			DeleteAuthSession(ctx, userID).
			Return(nil)

		err := svc.Logout(ctx)
		assert.NoError(t, err)
	})

	t.Run("error - auth session not found", func(t *testing.T) {
		svc, mocks := setupAuthServiceMocks(t)

		mocks.authRepo.EXPECT().
			DeleteAuthSession(ctx, userID).
			Return(sql.ErrNoRows)

		err := svc.Logout(ctx)
		assert.ErrorIs(t, err, errorpkg.ErrInvalidBearerToken)
	})

	t.Run("error - delete auth session fails", func(t *testing.T) {
		svc, mocks := setupAuthServiceMocks(t)

		mocks.authRepo.EXPECT().
			DeleteAuthSession(ctx, userID).
			Return(errors.New("db error"))

		err := svc.Logout(ctx)
		assert.Error(t, err)
		assert.ErrorIs(t, err, errorpkg.ErrInternalServer)
	})
}

func Test_AuthService_RequestOTPResetPassword(t *testing.T) {
	ctx := context.Background()
	email := "test@example.com"

	t.Run("success", func(t *testing.T) {
		svc, mocks := setupAuthServiceMocks(t)
		emailSent := make(chan struct{}, 1)

		// Expect user to be found
		mocks.userSvc.EXPECT().
			GetUserByEmail(ctx, email).
			Return(&entity.User{ID: uuid.New()}, nil)

		// Expect OTP to be set
		mocks.authRepo.EXPECT().
			SetOTPResetPassword(ctx, email, mock.AnythingOfType("string")).
			Return(nil)

		// Mock email sending with channel notification
		mocks.mailer.EXPECT().
			Send(
				email,
				"[Astungkara] Reset Password",
				"otp_reset_password.html",
				mock.AnythingOfType("map[string]interface {}"),
			).RunAndReturn(func(_, _, _ string, _ map[string]interface{}) error {
			emailSent <- struct{}{}
			return nil
		})

		err := svc.RequestOTPResetPassword(ctx, email)
		assert.NoError(t, err)

		// Wait for email sending goroutine to complete
		<-emailSent
	})

	t.Run("error - user not found", func(t *testing.T) {
		svc, mocks := setupAuthServiceMocks(t)

		mocks.userSvc.EXPECT().
			GetUserByEmail(ctx, email).
			Return(nil, errorpkg.ErrNotFound)

		err := svc.RequestOTPResetPassword(ctx, email)
		assert.ErrorIs(t, err, errorpkg.ErrNotFound)
		assert.Contains(t, err.Error(), "User not found. Please register.")
	})

	t.Run("error - get user unexpected error", func(t *testing.T) {
		svc, mocks := setupAuthServiceMocks(t)

		mocks.userSvc.EXPECT().
			GetUserByEmail(ctx, email).
			Return(nil, errors.New("unexpected error"))

		err := svc.RequestOTPResetPassword(ctx, email)
		assert.Error(t, err)
		assert.ErrorIs(t, err, errorpkg.ErrInternalServer)
	})

	t.Run("error - set OTP fails", func(t *testing.T) {
		svc, mocks := setupAuthServiceMocks(t)

		mocks.userSvc.EXPECT().
			GetUserByEmail(ctx, email).
			Return(&entity.User{ID: uuid.New()}, nil)

		mocks.authRepo.EXPECT().
			SetOTPResetPassword(ctx, email, mock.AnythingOfType("string")).
			Return(errors.New("redis error"))

		err := svc.RequestOTPResetPassword(ctx, email)
		assert.Error(t, err)
		assert.ErrorIs(t, err, errorpkg.ErrInternalServer)
	})

	t.Run("error - email sending fails", func(t *testing.T) {
		svc, mocks := setupAuthServiceMocks(t)
		emailSent := make(chan struct{}, 1)

		mocks.userSvc.EXPECT().
			GetUserByEmail(ctx, email).
			Return(&entity.User{ID: uuid.New()}, nil)

		mocks.authRepo.EXPECT().
			SetOTPResetPassword(ctx, email, mock.AnythingOfType("string")).
			Return(nil)

		mocks.mailer.EXPECT().
			Send(
				email,
				"[Astungkara] Reset Password",
				"otp_reset_password.html",
				mock.AnythingOfType("map[string]interface {}"),
			).RunAndReturn(func(_, _, _ string, _ map[string]interface{}) error {
			emailSent <- struct{}{}
			return errors.New("email sending error")
		})

		err := svc.RequestOTPResetPassword(ctx, email)
		assert.NoError(t, err) // Should not return error as email is sent in goroutine

		// Wait for email sending goroutine to complete
		<-emailSent
	})
}

func Test_AuthService_ResetPassword(t *testing.T) {
	ctx := context.Background()
	req := dto.ResetPasswordRequest{
		Email:       "test@example.com",
		OTP:         "123456",
		NewPassword: "newpassword123",
	}

	t.Run("success", func(t *testing.T) {
		svc, mocks := setupAuthServiceMocks(t)

		// Setup expectations for password reset
		mocks.authRepo.EXPECT().
			GetOTPResetPassword(ctx, req.Email).
			Return(req.OTP, nil)

		mocks.authRepo.EXPECT().
			SetOTPResetPassword(ctx, req.Email, "").
			Return(nil)

		mocks.userSvc.EXPECT().
			UpdatePassword(ctx, req.Email, req.NewPassword).
			Return(nil)

		// Setup expectations for subsequent login
		mockLoginExpectations(mocks, ctx, req.Email, req.NewPassword, uuid.New())

		resp, err := svc.ResetPassword(ctx, req)
		assert.NoError(t, err)
		assert.NotEmpty(t, resp.AccessToken)
		assert.NotEmpty(t, resp.RefreshToken)
		assert.NotNil(t, resp.User)
	})

	t.Run("error - OTP not found", func(t *testing.T) {
		svc, mocks := setupAuthServiceMocks(t)

		mocks.authRepo.EXPECT().
			GetOTPResetPassword(ctx, req.Email).
			Return("", redis.Nil)

		resp, err := svc.ResetPassword(ctx, req)
		assert.Empty(t, resp)
		assert.ErrorIs(t, err, errorpkg.ErrInvalidOTP)
	})

	t.Run("error - get OTP unexpected error", func(t *testing.T) {
		svc, mocks := setupAuthServiceMocks(t)

		mocks.authRepo.EXPECT().
			GetOTPResetPassword(ctx, req.Email).
			Return("", errors.New("redis error"))

		resp, err := svc.ResetPassword(ctx, req)
		assert.Empty(t, resp)
		assert.Error(t, err)
		assert.ErrorIs(t, err, errorpkg.ErrInternalServer)
	})

	t.Run("error - invalid OTP", func(t *testing.T) {
		svc, mocks := setupAuthServiceMocks(t)

		mocks.authRepo.EXPECT().
			GetOTPResetPassword(ctx, req.Email).
			Return("different-otp", nil)

		resp, err := svc.ResetPassword(ctx, req)
		assert.Empty(t, resp)
		assert.ErrorIs(t, err, errorpkg.ErrInvalidOTP)
	})

	t.Run("error - delete OTP fails", func(t *testing.T) {
		svc, mocks := setupAuthServiceMocks(t)

		mocks.authRepo.EXPECT().
			GetOTPResetPassword(ctx, req.Email).
			Return(req.OTP, nil)

		mocks.authRepo.EXPECT().
			SetOTPResetPassword(ctx, req.Email, "").
			Return(errors.New("redis error"))

		resp, err := svc.ResetPassword(ctx, req)
		assert.Empty(t, resp)
		assert.Error(t, err)
		assert.ErrorIs(t, err, errorpkg.ErrInternalServer)
	})

	t.Run("error - update password fails", func(t *testing.T) {
		svc, mocks := setupAuthServiceMocks(t)

		mocks.authRepo.EXPECT().
			GetOTPResetPassword(ctx, req.Email).
			Return(req.OTP, nil)

		mocks.authRepo.EXPECT().
			SetOTPResetPassword(ctx, req.Email, "").
			Return(nil)

		mocks.userSvc.EXPECT().
			UpdatePassword(ctx, req.Email, req.NewPassword).
			Return(errors.New("db error"))

		resp, err := svc.ResetPassword(ctx, req)
		assert.Empty(t, resp)
		assert.Error(t, err)
		assert.ErrorIs(t, err, errorpkg.ErrInternalServer)
	})

	t.Run("error - user not found during update", func(t *testing.T) {
		svc, mocks := setupAuthServiceMocks(t)

		mocks.authRepo.EXPECT().
			GetOTPResetPassword(ctx, req.Email).
			Return(req.OTP, nil)

		mocks.authRepo.EXPECT().
			SetOTPResetPassword(ctx, req.Email, "").
			Return(nil)

		mocks.userSvc.EXPECT().
			UpdatePassword(ctx, req.Email, req.NewPassword).
			Return(errorpkg.ErrNotFound)

		resp, err := svc.ResetPassword(ctx, req)
		assert.Empty(t, resp)
		assert.ErrorIs(t, err, errorpkg.ErrNotFound)
	})

	t.Run("error - login after reset fails", func(t *testing.T) {
		svc, mocks := setupAuthServiceMocks(t)

		// Success up to password update
		mocks.authRepo.EXPECT().
			GetOTPResetPassword(ctx, req.Email).
			Return(req.OTP, nil)

		mocks.authRepo.EXPECT().
			SetOTPResetPassword(ctx, req.Email, "").
			Return(nil)

		mocks.userSvc.EXPECT().
			UpdatePassword(ctx, req.Email, req.NewPassword).
			Return(nil)

		// Fail during login
		mocks.userSvc.EXPECT().
			GetUserByEmail(ctx, req.Email).
			Return(nil, errors.New("db error"))

		resp, err := svc.ResetPassword(ctx, req)
		assert.Empty(t, resp)
		assert.Error(t, err)
		assert.ErrorIs(t, err, errorpkg.ErrInternalServer)
	})
}

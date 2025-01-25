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

	t.Run("error - email sending fails", func(t *testing.T) {
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

		// Mock email sending to fail
		mocks.mailer.EXPECT().
			Send(
				email,
				"[Class Manager] Verify Your Account",
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
			SetUserRegisterOTP(ctx, email, mock.AnythingOfType("string")).
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
		assert.ErrorIs(t, err, errorpkg.ErrInternalServer)
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

		resp, err := svc.LoginUser(ctx, req)
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

		resp, err := svc.LoginUser(ctx, req)
		assert.Empty(t, resp)
		assert.ErrorIs(t, err, errorpkg.ErrNotFound)
	})

	t.Run("error - get user unexpected error", func(t *testing.T) {
		svc, mocks := setupAuthServiceMocks(t)

		mocks.userSvc.EXPECT().
			GetUserByEmail(ctx, req.Email).
			Return(nil, errors.New("db error"))

		resp, err := svc.LoginUser(ctx, req)
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

		resp, err := svc.LoginUser(ctx, req)
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

		mocks.userSvc.EXPECT().
			GetUserByEmail(ctx, req.Email).
			Return(user, nil)

		mocks.bcrypt.EXPECT().
			Compare(req.Password, passwordHash).
			Return(true)

		mocks.jwt.EXPECT().
			Create(user.ID, user.Role).
			Return("", errors.New("jwt error"))

		mocks.authRepo.EXPECT().
			CreateSession(ctx, mock.Anything).
			Return(nil)

		resp, err := svc.LoginUser(ctx, req)
		assert.Empty(t, resp)
		assert.Error(t, err)
		assert.ErrorIs(t, err, errorpkg.ErrInternalServer)
	})

	t.Run("error - create session fails", func(t *testing.T) {
		svc, mocks := setupAuthServiceMocks(t)

		passwordHash := "hashed_password"
		user := &entity.User{
			ID:           uuid.New(),
			Email:        req.Email,
			PasswordHash: passwordHash,
			Role:         enum.RoleUser,
		}

		mocks.userSvc.EXPECT().
			GetUserByEmail(ctx, req.Email).
			Return(user, nil)

		mocks.bcrypt.EXPECT().
			Compare(req.Password, passwordHash).
			Return(true)

		mocks.jwt.EXPECT().
			Create(user.ID, user.Role).
			Return("access_token", nil)

		mocks.authRepo.EXPECT().
			CreateSession(ctx, mock.MatchedBy(func(session *entity.Session) bool {
				return session.UserID == user.ID
			})).
			Return(errors.New("db error"))

		resp, err := svc.LoginUser(ctx, req)
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
			GetUserRegisterOTP(ctx, req.Email).
			Return(req.OTP, nil)

		mocks.authRepo.EXPECT().
			DeleteUserRegisterOTP(ctx, req.Email).
			Return(nil)

		hashedPassword := "hashed_password"
		mocks.bcrypt.EXPECT().
			Hash(req.Password).
			Return(hashedPassword, nil)

		userID := uuid.New()
		mocks.userSvc.EXPECT().
			CreateUser(ctx, mock.MatchedBy(func(createReq *dto.CreateUserRequest) bool {
				return createReq.Email == req.Email &&
					createReq.Name == req.Name &&
					createReq.PasswordHash == hashedPassword &&
					createReq.Role == enum.RoleUser
			})).
			Return(userID, nil)

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
			GetUserRegisterOTP(ctx, req.Email).
			Return("", redis.Nil)

		resp, err := svc.RegisterUser(ctx, req)
		assert.Empty(t, resp)
		assert.ErrorIs(t, err, errorpkg.ErrInvalidOTP)
	})

	t.Run("error - invalid OTP", func(t *testing.T) {
		svc, mocks := setupAuthServiceMocks(t)

		mocks.authRepo.EXPECT().
			GetUserRegisterOTP(ctx, req.Email).
			Return("different-otp", nil)

		resp, err := svc.RegisterUser(ctx, req)
		assert.Empty(t, resp)
		assert.ErrorIs(t, err, errorpkg.ErrInvalidOTP)
	})

	t.Run("error - get OTP fails", func(t *testing.T) {
		svc, mocks := setupAuthServiceMocks(t)

		mocks.authRepo.EXPECT().
			GetUserRegisterOTP(ctx, req.Email).
			Return("", errors.New("redis error"))

		resp, err := svc.RegisterUser(ctx, req)
		assert.Empty(t, resp)
		assert.Error(t, err)
		assert.ErrorIs(t, err, errorpkg.ErrInternalServer)
	})

	t.Run("error - delete OTP fails", func(t *testing.T) {
		svc, mocks := setupAuthServiceMocks(t)

		mocks.authRepo.EXPECT().
			GetUserRegisterOTP(ctx, req.Email).
			Return(req.OTP, nil)

		mocks.authRepo.EXPECT().
			DeleteUserRegisterOTP(ctx, req.Email).
			Return(errors.New("redis error"))

		resp, err := svc.RegisterUser(ctx, req)
		assert.Empty(t, resp)
		assert.Error(t, err)
		assert.ErrorIs(t, err, errorpkg.ErrInternalServer)
	})

	t.Run("error - password hash fails", func(t *testing.T) {
		svc, mocks := setupAuthServiceMocks(t)

		mocks.authRepo.EXPECT().
			GetUserRegisterOTP(ctx, req.Email).
			Return(req.OTP, nil)

		mocks.authRepo.EXPECT().
			DeleteUserRegisterOTP(ctx, req.Email).
			Return(nil)

		mocks.bcrypt.EXPECT().
			Hash(req.Password).
			Return("", errors.New("bcrypt error"))

		resp, err := svc.RegisterUser(ctx, req)
		assert.Empty(t, resp)
		assert.Error(t, err)
		assert.ErrorIs(t, err, errorpkg.ErrInternalServer)
	})

	t.Run("error - create user fails", func(t *testing.T) {
		svc, mocks := setupAuthServiceMocks(t)

		mocks.authRepo.EXPECT().
			GetUserRegisterOTP(ctx, req.Email).
			Return(req.OTP, nil)

		mocks.authRepo.EXPECT().
			DeleteUserRegisterOTP(ctx, req.Email).
			Return(nil)

		hashedPassword := "hashed_password"
		mocks.bcrypt.EXPECT().
			Hash(req.Password).
			Return(hashedPassword, nil)

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
		CreateSession(ctx, mock.MatchedBy(func(session *entity.Session) bool {
			return session.UserID == user.ID &&
				len(session.Token) == 32 && // Check refresh token length
				!session.ExpiresAt.IsZero() // Check expiration is set
		})).
		Return(nil)
}

func Test_AuthService_RefreshToken(t *testing.T) {
	ctx := context.Background()
	refreshToken := "valid_refresh_token"
	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		svc, mocks := setupAuthServiceMocks(t)

		// Setup session
		session := &entity.Session{
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
			GetSessionByToken(ctx, refreshToken).
			Return(session, nil)

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

	t.Run("error - session not found", func(t *testing.T) {
		svc, mocks := setupAuthServiceMocks(t)

		mocks.authRepo.EXPECT().
			GetSessionByToken(ctx, refreshToken).
			Return(nil, sql.ErrNoRows)

		resp, err := svc.RefreshToken(ctx, refreshToken)
		assert.Empty(t, resp)
		assert.ErrorIs(t, err, errorpkg.ErrInvalidRefreshToken)
	})

	t.Run("error - get session unexpected error", func(t *testing.T) {
		svc, mocks := setupAuthServiceMocks(t)

		mocks.authRepo.EXPECT().
			GetSessionByToken(ctx, refreshToken).
			Return(nil, errors.New("db error"))

		resp, err := svc.RefreshToken(ctx, refreshToken)
		assert.Empty(t, resp)
		assert.Error(t, err)
		assert.ErrorIs(t, err, errorpkg.ErrInternalServer)
	})

	t.Run("error - expired session", func(t *testing.T) {
		svc, mocks := setupAuthServiceMocks(t)

		// Setup expired session
		session := &entity.Session{
			Token:     refreshToken,
			UserID:    userID,
			ExpiresAt: time.Now().Add(-time.Hour), // Expired
		}

		mocks.authRepo.EXPECT().
			GetSessionByToken(ctx, refreshToken).
			Return(session, nil)

		resp, err := svc.RefreshToken(ctx, refreshToken)
		assert.Empty(t, resp)
		assert.ErrorIs(t, err, errorpkg.ErrInvalidRefreshToken)
	})

	t.Run("error - user not found", func(t *testing.T) {
		svc, mocks := setupAuthServiceMocks(t)

		// Setup valid session
		session := &entity.Session{
			Token:     refreshToken,
			UserID:    userID,
			ExpiresAt: time.Now().Add(time.Hour),
		}

		mocks.authRepo.EXPECT().
			GetSessionByToken(ctx, refreshToken).
			Return(session, nil)

		mocks.userSvc.EXPECT().
			GetUserByID(ctx, userID).
			Return(nil, errorpkg.ErrNotFound)

		resp, err := svc.RefreshToken(ctx, refreshToken)
		assert.Empty(t, resp)
		assert.ErrorIs(t, err, errorpkg.ErrNotFound)
	})

	t.Run("error - get user unexpected error", func(t *testing.T) {
		svc, mocks := setupAuthServiceMocks(t)

		// Setup valid session
		session := &entity.Session{
			Token:     refreshToken,
			UserID:    userID,
			ExpiresAt: time.Now().Add(time.Hour),
		}

		mocks.authRepo.EXPECT().
			GetSessionByToken(ctx, refreshToken).
			Return(session, nil)

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

		// Setup valid session
		session := &entity.Session{
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
			GetSessionByToken(ctx, refreshToken).
			Return(session, nil)

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
			DeleteSession(ctx, userID).
			Return(nil)

		err := svc.Logout(ctx)
		assert.NoError(t, err)
	})

	t.Run("error - session not found", func(t *testing.T) {
		svc, mocks := setupAuthServiceMocks(t)

		mocks.authRepo.EXPECT().
			DeleteSession(ctx, userID).
			Return(sql.ErrNoRows)

		err := svc.Logout(ctx)
		assert.ErrorIs(t, err, errorpkg.ErrInvalidBearerToken)
	})

	t.Run("error - delete session fails", func(t *testing.T) {
		svc, mocks := setupAuthServiceMocks(t)

		mocks.authRepo.EXPECT().
			DeleteSession(ctx, userID).
			Return(errors.New("db error"))

		err := svc.Logout(ctx)
		assert.Error(t, err)
		assert.ErrorIs(t, err, errorpkg.ErrInternalServer)
	})
}

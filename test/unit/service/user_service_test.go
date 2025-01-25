package service

import (
	"context"
	"database/sql"
	"errors"
	"github.com/nathakusuma/astungkara/domain/contract"
	"github.com/nathakusuma/astungkara/internal/app/user/service"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/nathakusuma/astungkara/domain/dto"
	"github.com/nathakusuma/astungkara/domain/entity"
	"github.com/nathakusuma/astungkara/domain/enum"
	"github.com/nathakusuma/astungkara/domain/errorpkg"
	appmocks "github.com/nathakusuma/astungkara/test/unit/mocks/app"
	pkgmocks "github.com/nathakusuma/astungkara/test/unit/mocks/pkg"
	_ "github.com/nathakusuma/astungkara/test/unit/setup" // Initialize test environment
	"github.com/stretchr/testify/assert"
)

type userServiceMocks struct {
	userRepo *appmocks.MockIUserRepository
	uuid     *pkgmocks.MockIUUID
	bcrypt   *pkgmocks.MockIBcrypt
}

func setupUserServiceTest(t *testing.T) (contract.IUserService, *userServiceMocks) {
	mocks := &userServiceMocks{
		userRepo: appmocks.NewMockIUserRepository(t),
		uuid:     pkgmocks.NewMockIUUID(t),
		bcrypt:   pkgmocks.NewMockIBcrypt(t),
	}

	svc := service.NewUserService(mocks.userRepo, mocks.bcrypt, mocks.uuid)

	return svc, mocks
}

func Test_UserService_CreateUser(t *testing.T) {
	ctx := context.Background()
	hashedPassword := "hashed_password"
	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		svc, mocks := setupUserServiceTest(t)

		req := &dto.CreateUserRequest{
			Name:     "Test User",
			Email:    "test@example.com",
			Password: hashedPassword,
			Role:     enum.RoleUser,
		}

		// Expect UUID generation
		mocks.uuid.EXPECT().
			NewV7().
			Return(userID, nil)

		// Expect password hashing
		mocks.bcrypt.EXPECT().
			Hash(req.Password).
			Return(hashedPassword, nil)

		// Expect user creation
		mocks.userRepo.EXPECT().
			CreateUser(ctx, &entity.User{
				ID:           userID,
				Name:         req.Name,
				Email:        req.Email,
				PasswordHash: req.Password,
				Role:         req.Role,
			}).
			Return(nil)

		resultID, err := svc.CreateUser(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, userID, resultID)
	})

	t.Run("error - uuid generation fails", func(t *testing.T) {
		svc, mocks := setupUserServiceTest(t)

		req := &dto.CreateUserRequest{
			Name:     "Test User",
			Email:    "test@example.com",
			Password: hashedPassword,
		}

		// Expect UUID generation to fail
		mocks.uuid.EXPECT().
			NewV7().
			Return(uuid.UUID{}, errors.New("uuid error"))

		resultID, err := svc.CreateUser(ctx, req)
		assert.Equal(t, uuid.Nil, resultID)
		assert.Error(t, err)
		assert.ErrorIs(t, err, errorpkg.ErrInternalServer)
	})

	t.Run("error - email already exists", func(t *testing.T) {
		svc, mocks := setupUserServiceTest(t)

		req := &dto.CreateUserRequest{
			Name:     "Test User",
			Email:    "test@example.com",
			Password: hashedPassword,
		}

		// Expect UUID generation
		mocks.uuid.EXPECT().
			NewV7().
			Return(userID, nil)

		// Expect password hashing
		mocks.bcrypt.EXPECT().
			Hash(req.Password).
			Return(hashedPassword, nil)

		// Expect user creation to fail with unique constraint violation
		pgErr := &pgconn.PgError{
			ConstraintName: "users_email_key",
		}
		mocks.userRepo.EXPECT().
			CreateUser(ctx, &entity.User{
				ID:           userID,
				Name:         req.Name,
				Email:        req.Email,
				PasswordHash: req.Password,
				Role:         req.Role,
			}).
			Return(pgErr)

		resultID, err := svc.CreateUser(ctx, req)
		assert.Equal(t, uuid.Nil, resultID)
		assert.ErrorIs(t, err, errorpkg.ErrEmailAlreadyRegistered)
	})

	t.Run("error - repository error", func(t *testing.T) {
		svc, mocks := setupUserServiceTest(t)

		req := &dto.CreateUserRequest{
			Name:     "Test User",
			Email:    "test@example.com",
			Password: hashedPassword,
		}

		// Expect UUID generation
		mocks.uuid.EXPECT().
			NewV7().
			Return(userID, nil)

		hashedPassword := "hashed_password"
		mocks.bcrypt.EXPECT().
			Hash(req.Password).
			Return(hashedPassword, nil)

		// Expect user creation to fail
		mocks.userRepo.EXPECT().
			CreateUser(ctx, &entity.User{
				ID:           userID,
				Name:         req.Name,
				Email:        req.Email,
				PasswordHash: hashedPassword,
				Role:         req.Role,
			}).
			Return(errors.New("db error"))

		resultID, err := svc.CreateUser(ctx, req)
		assert.Equal(t, uuid.Nil, resultID)
		assert.Error(t, err)
		assert.ErrorIs(t, err, errorpkg.ErrInternalServer)
	})
}

func Test_UserService_GetUserByEmail(t *testing.T) {
	ctx := context.Background()
	email := "test@example.com"

	t.Run("success", func(t *testing.T) {
		svc, mocks := setupUserServiceTest(t)

		expectedUser := &entity.User{
			ID:        uuid.New(),
			Name:      "Test User",
			Email:     email,
			Role:      enum.RoleUser,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mocks.userRepo.EXPECT().
			GetUserByField(ctx, "email", email).
			Return(expectedUser, nil)

		user, err := svc.GetUserByEmail(ctx, email)
		assert.NoError(t, err)
		assert.Equal(t, expectedUser, user)
	})

	t.Run("error - user not found", func(t *testing.T) {
		svc, mocks := setupUserServiceTest(t)

		mocks.userRepo.EXPECT().
			GetUserByField(ctx, "email", email).
			Return(nil, sql.ErrNoRows)

		user, err := svc.GetUserByEmail(ctx, email)
		assert.Nil(t, user)
		assert.ErrorIs(t, err, errorpkg.ErrNotFound)
	})

	t.Run("error - repository error", func(t *testing.T) {
		svc, mocks := setupUserServiceTest(t)

		mocks.userRepo.EXPECT().
			GetUserByField(ctx, "email", email).
			Return(nil, errors.New("db error"))

		user, err := svc.GetUserByEmail(ctx, email)
		assert.Nil(t, user)
		assert.Error(t, err)
		assert.ErrorIs(t, err, errorpkg.ErrInternalServer)
	})
}

func Test_UserService_GetUserByID(t *testing.T) {
	ctx := context.Background()
	id := uuid.New()

	t.Run("success", func(t *testing.T) {
		svc, mocks := setupUserServiceTest(t)

		expectedUser := &entity.User{
			ID:        id,
			Name:      "Test User",
			Email:     "test@example.com",
			Role:      enum.RoleUser,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mocks.userRepo.EXPECT().
			GetUserByField(ctx, "id", id.String()).
			Return(expectedUser, nil)

		user, err := svc.GetUserByID(ctx, id)
		assert.NoError(t, err)
		assert.Equal(t, expectedUser, user)
	})

	t.Run("error - user not found", func(t *testing.T) {
		svc, mocks := setupUserServiceTest(t)

		mocks.userRepo.EXPECT().
			GetUserByField(ctx, "id", id.String()).
			Return(nil, sql.ErrNoRows)

		user, err := svc.GetUserByID(ctx, id)
		assert.Nil(t, user)
		assert.ErrorIs(t, err, errorpkg.ErrNotFound)
	})

	t.Run("error - repository error", func(t *testing.T) {
		svc, mocks := setupUserServiceTest(t)

		mocks.userRepo.EXPECT().
			GetUserByField(ctx, "id", id.String()).
			Return(nil, errors.New("db error"))

		user, err := svc.GetUserByID(ctx, id)
		assert.Nil(t, user)
		assert.Error(t, err)
		assert.ErrorIs(t, err, errorpkg.ErrInternalServer)
	})
}

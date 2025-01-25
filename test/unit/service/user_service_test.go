package service

import (
	"context"
	"errors"
	"github.com/nathakusuma/astungkara/domain/contract"
	"github.com/nathakusuma/astungkara/internal/app/user/service"
	"testing"

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
			Name:         "Test User",
			Email:        "test@example.com",
			PasswordHash: hashedPassword,
			Role:         enum.RoleUser,
		}

		// Expect UUID generation
		mocks.uuid.EXPECT().
			NewV7().
			Return(userID, nil)

		// Expect user creation
		mocks.userRepo.EXPECT().
			CreateUser(ctx, &entity.User{
				ID:           userID,
				Name:         req.Name,
				Email:        req.Email,
				PasswordHash: req.PasswordHash,
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
			Name:         "Test User",
			Email:        "test@example.com",
			PasswordHash: hashedPassword,
		}

		// Expect UUID generation to fail
		mocks.uuid.EXPECT().
			NewV7().
			Return(uuid.UUID{}, errors.New("uuid error"))

		resultID, err := svc.CreateUser(ctx, req)
		assert.Equal(t, uuid.Nil, resultID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), errorpkg.ErrInternalServer.Error())
	})

	t.Run("error - email already exists", func(t *testing.T) {
		svc, mocks := setupUserServiceTest(t)

		req := &dto.CreateUserRequest{
			Name:         "Test User",
			Email:        "test@example.com",
			PasswordHash: hashedPassword,
		}

		// Expect UUID generation
		mocks.uuid.EXPECT().
			NewV7().
			Return(userID, nil)

		// Expect user creation to fail with unique constraint violation
		pgErr := &pgconn.PgError{
			ConstraintName: "users_email_key",
		}
		mocks.userRepo.EXPECT().
			CreateUser(ctx, &entity.User{
				ID:           userID,
				Name:         req.Name,
				Email:        req.Email,
				PasswordHash: req.PasswordHash,
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
			Name:         "Test User",
			Email:        "test@example.com",
			PasswordHash: hashedPassword,
		}

		// Expect UUID generation
		mocks.uuid.EXPECT().
			NewV7().
			Return(userID, nil)

		// Expect user creation to fail
		mocks.userRepo.EXPECT().
			CreateUser(ctx, &entity.User{
				ID:           userID,
				Name:         req.Name,
				Email:        req.Email,
				PasswordHash: req.PasswordHash,
				Role:         req.Role,
			}).
			Return(errors.New("db error"))

		resultID, err := svc.CreateUser(ctx, req)
		assert.Equal(t, uuid.Nil, resultID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), errorpkg.ErrInternalServer.Error())
	})
}

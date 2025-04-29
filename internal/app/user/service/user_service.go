package service

import (
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/nathakusuma/astungkara/domain/errorpkg"
	"github.com/nathakusuma/astungkara/pkg/log"
	"github.com/nathakusuma/astungkara/pkg/uuidpkg"

	"github.com/nathakusuma/astungkara/domain/contract"
	"github.com/nathakusuma/astungkara/domain/dto"
	"github.com/nathakusuma/astungkara/domain/entity"
	"github.com/nathakusuma/astungkara/pkg/bcrypt"
)

type userService struct {
	userRepo contract.IUserRepository
	bcrypt   bcrypt.IBcrypt
	uuid     uuidpkg.IUUID
}

func NewUserService(
	userRepo contract.IUserRepository,
	bcrypt bcrypt.IBcrypt,
	uuid uuidpkg.IUUID,
) contract.IUserService {
	return &userService{
		userRepo: userRepo,
		bcrypt:   bcrypt,
		uuid:     uuid,
	}
}

func (s *userService) CreateUser(ctx context.Context, req *dto.CreateUserRequest) (uuid.UUID, error) {
	loggableReq := *req
	loggableReq.Password = ""

	creatorID := ctx.Value("user.id")
	if creatorID == nil {
		creatorID = "system"
	}

	// generate user ID
	userID, err := s.uuid.NewV7()
	if err != nil {
		traceID := log.ErrorWithTraceID(map[string]interface{}{
			"error":        err.Error(),
			"request":      loggableReq,
			"requester.id": creatorID,
		}, "[UserService][CreateUser] Failed to generate user ID")

		return uuid.Nil, errorpkg.ErrInternalServer.WithTraceID(traceID)
	}

	passwordHash, err := s.bcrypt.Hash(req.Password)
	if err != nil {
		traceID := log.ErrorWithTraceID(map[string]interface{}{
			"error":        err.Error(),
			"request":      loggableReq,
			"requester.id": creatorID,
		}, "[UserService][CreateUser] Failed to hash password")

		return uuid.Nil, errorpkg.ErrInternalServer.WithTraceID(traceID)
	}

	// create user data
	user := &entity.User{
		ID:           userID,
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: passwordHash,
		Role:         req.Role,
	}

	err = s.userRepo.CreateUser(ctx, user)
	if err != nil {
		// if error is due to conflict in unique constraint in email column
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.ConstraintName == "users_email_key" {
			return uuid.Nil, errorpkg.ErrEmailAlreadyRegistered
		}

		// other error
		traceID := log.ErrorWithTraceID(map[string]interface{}{
			"error":        err.Error(),
			"request":      loggableReq,
			"requester.id": creatorID,
		}, "[UserService][CreateUser] Failed to create user")

		return uuid.Nil, errorpkg.ErrInternalServer.WithTraceID(traceID)
	}

	log.Info(map[string]interface{}{
		"user":         user,
		"requester.id": creatorID,
	}, "[UserService][CreateUser] User created")

	return userID, nil
}

func (s *userService) getUserByField(ctx context.Context, field, value string) (*entity.User, error) {
	// get from repository
	user, err := s.userRepo.GetUserByField(ctx, field, value)
	if err != nil {
		// if user not found
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errorpkg.ErrNotFound.WithMessage("User not found.")
		}

		// other error
		traceID := log.ErrorWithTraceID(map[string]interface{}{
			"error": err.Error(),
			"field": field,
			"value": value,
		}, "[UserService][getUserByField] Failed to get user by field")

		return nil, errorpkg.ErrInternalServer.WithTraceID(traceID)
	}

	return user, nil
}

func (s *userService) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	return s.getUserByField(ctx, "email", email)
}

func (s *userService) GetUserByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	return s.getUserByField(ctx, "id", id.String())
}

func (s *userService) UpdatePassword(ctx context.Context, email, newPassword string) error {
	// get user by email
	user, err := s.GetUserByEmail(ctx, email)
	if err != nil {
		return err
	}

	// hash new password
	newPasswordHash, err := s.bcrypt.Hash(newPassword)
	if err != nil {
		traceID := log.ErrorWithTraceID(map[string]interface{}{
			"error":      err.Error(),
			"user.email": email,
		}, "[UserService][UpdatePassword] Failed to hash password")

		return errorpkg.ErrInternalServer.WithTraceID(traceID)
	}

	// update user password
	user.PasswordHash = newPasswordHash
	err = s.userRepo.UpdateUser(ctx, user)
	if err != nil {
		traceID := log.ErrorWithTraceID(map[string]interface{}{
			"error":      err.Error(),
			"user.email": email,
		}, "[UserService][UpdatePassword] Failed to update user password")

		return errorpkg.ErrInternalServer.WithTraceID(traceID)
	}

	log.Info(map[string]interface{}{
		"user.email": email,
	}, "[UserService][UpdatePassword] Password updated")

	return nil
}

func (s *userService) UpdateUser(ctx context.Context, id uuid.UUID, req dto.UpdateUserRequest) error {
	// get user by ID
	user, err := s.GetUserByID(ctx, id)
	if err != nil {
		return err
	}

	// update user data
	if req.Name != nil {
		user.Name = *req.Name
	}
	if req.Bio != nil {
		user.Bio = req.Bio
	}

	// update user
	err = s.userRepo.UpdateUser(ctx, user)
	if err != nil {
		traceID := log.ErrorWithTraceID(map[string]interface{}{
			"error": err.Error(),
			"user":  user,
		}, "[UserService][UpdateUser] Failed to update user")

		return errorpkg.ErrInternalServer.WithTraceID(traceID)
	}

	log.Info(map[string]interface{}{
		"user": user,
	}, "[UserService][UpdateUser] User updated")

	return nil
}

func (s *userService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	requesterID := ctx.Value("user.id")
	if requesterID == nil {
		requesterID = "system"
	}

	// delete user
	err := s.userRepo.DeleteUser(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errorpkg.ErrNotFound
		}

		traceID := log.ErrorWithTraceID(map[string]interface{}{
			"error":        err.Error(),
			"user.id":      id,
			"requester.id": requesterID,
		}, "[UserService][DeleteUser] Failed to delete user")
		return errorpkg.ErrInternalServer.WithTraceID(traceID)
	}

	log.Info(map[string]interface{}{
		"user.id":      id,
		"requester.id": requesterID,
	}, "[UserService][DeleteUser] User deleted")

	return nil
}

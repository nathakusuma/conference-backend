package contract

import (
	"context"

	"github.com/google/uuid"
	"github.com/nathakusuma/astungkara/domain/dto"
	"github.com/nathakusuma/astungkara/domain/entity"
)

type IUserRepository interface {
	CreateUser(ctx context.Context, user *entity.User) error
	GetUserByField(ctx context.Context, field, value string) (*entity.User, error)
	UpdateUser(ctx context.Context, user *entity.User) error
}

type IUserService interface {
	CreateUser(ctx context.Context, req *dto.CreateUserRequest) (uuid.UUID, error)
	GetUserByEmail(ctx context.Context, email string) (*entity.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	UpdatePassword(ctx context.Context, email, newPassword string) error
}

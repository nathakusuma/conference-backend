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
}

type IUserService interface {
	CreateUser(ctx context.Context, req *dto.CreateUserRequest) (uuid.UUID, error)
	GetUserByEmail(ctx context.Context, email string) (*entity.User, error)
}

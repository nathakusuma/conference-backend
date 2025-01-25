package dto

import (
	"github.com/google/uuid"
	"github.com/nathakusuma/astungkara/domain/entity"
	"github.com/nathakusuma/astungkara/domain/enum"
	"time"
)

type UserResponse struct {
	ID        uuid.UUID     `json:"id,omitempty"`
	Name      string        `json:"name,omitempty"`
	Email     string        `json:"email,omitempty,omitempty"`
	Role      enum.UserRole `json:"role,omitempty,omitempty"`
	CreatedAt time.Time     `json:"created_at,omitempty"`
	UpdatedAt time.Time     `json:"updated_at,omitempty"`
}

func (u *UserResponse) PopulateFromEntity(user *entity.User) *UserResponse {
	u.ID = user.ID
	u.Name = user.Name
	u.Email = user.Email
	u.Role = user.Role
	u.CreatedAt = user.CreatedAt
	u.UpdatedAt = user.UpdatedAt
	return u
}

type CreateUserRequest struct {
	Name         string        `json:"name" validate:"required,min=3,max=100,ascii"`
	Email        string        `json:"email" validate:"required,email,max=320"`
	PasswordHash string        `json:"password" validate:"required,len=60"`
	Role         enum.UserRole `json:"role" validate:"required,oneof=user admin superadmin"`
}

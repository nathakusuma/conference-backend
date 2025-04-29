package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/nathakusuma/astungkara/domain/contract"
	"github.com/nathakusuma/astungkara/domain/dto"
	"github.com/nathakusuma/astungkara/domain/errorpkg"
	"github.com/nathakusuma/astungkara/internal/middleware"
	"github.com/nathakusuma/astungkara/pkg/validator"
)

type userHandler struct {
	val validator.IValidator
	svc contract.IUserService
}

func InitUserHandler(
	router fiber.Router,
	middlewareInstance *middleware.Middleware,
	validator validator.IValidator,
	userSvc contract.IUserService,
) {
	handler := userHandler{
		svc: userSvc,
		val: validator,
	}

	userGroup := router.Group("/users")
	userGroup.Patch("", middlewareInstance.RequireAuthenticated(), handler.updateUser())
}

func (c *userHandler) updateUser() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		var req dto.UpdateUserRequest
		if err := ctx.BodyParser(&req); err != nil {
			return errorpkg.ErrFailParseRequest
		}

		if err := c.val.ValidateStruct(req); err != nil {
			return err
		}

		if err := c.svc.UpdateUser(ctx.Context(), ctx.Locals("user.id").(uuid.UUID), req); err != nil {
			return err
		}

		return ctx.SendStatus(fiber.StatusNoContent)
	}
}

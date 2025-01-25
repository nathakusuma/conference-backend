package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nathakusuma/astungkara/domain/contract"
	"github.com/nathakusuma/astungkara/domain/dto"
	"github.com/nathakusuma/astungkara/domain/errorpkg"
	"github.com/nathakusuma/astungkara/pkg/validator"
	"net/http"
)

type authHandler struct {
	val validator.IValidator
	svc contract.IAuthService
}

func InitAuthHandler(
	router fiber.Router,
	validator validator.IValidator,
	authSvc contract.IAuthService,
) {

	handler := authHandler{
		svc: authSvc,
		val: validator,
	}

	authGroup := router.Group("/auth")
	authGroup.Post("/register/otp", handler.requestOTPRegisterUser())
	authGroup.Post("/register/otp/check", handler.checkOTPRegisterUser())
	authGroup.Post("/register", handler.registerUser())
	authGroup.Post("/login", handler.loginUser())
}

func (c *authHandler) requestOTPRegisterUser() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		var req dto.RequestOTPRegisterUserRequest
		if err := ctx.BodyParser(&req); err != nil {
			return errorpkg.ErrFailParseRequest
		}

		if err := c.val.ValidateStruct(req); err != nil {
			return err
		}

		err := c.svc.RequestOTPRegisterUser(ctx.Context(), req.Email)
		if err != nil {
			return err
		}

		return ctx.SendStatus(http.StatusNoContent)
	}
}

func (c *authHandler) checkOTPRegisterUser() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		var req dto.CheckOTPRegisterUserRequest
		if err := ctx.BodyParser(&req); err != nil {
			return errorpkg.ErrFailParseRequest
		}

		if err := c.val.ValidateStruct(req); err != nil {
			return err
		}

		err := c.svc.CheckOTPRegisterUser(ctx.Context(), req.Email, req.OTP)
		if err != nil {
			return err
		}

		return ctx.SendStatus(http.StatusNoContent)
	}
}

func (c *authHandler) registerUser() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		var req dto.RegisterUserRequest
		if err := ctx.BodyParser(&req); err != nil {
			return errorpkg.ErrFailParseRequest
		}

		if err := c.val.ValidateStruct(req); err != nil {
			return err
		}

		resp, err := c.svc.RegisterUser(ctx.Context(), req)
		if err != nil {
			return err
		}

		return ctx.Status(http.StatusCreated).JSON(resp)
	}
}

func (c *authHandler) loginUser() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		var req dto.LoginUserRequest
		if err := ctx.BodyParser(&req); err != nil {
			return errorpkg.ErrFailParseRequest
		}

		if err := c.val.ValidateStruct(req); err != nil {
			return err
		}

		resp, err := c.svc.LoginUser(ctx.Context(), req)
		if err != nil {
			return err
		}

		return ctx.Status(http.StatusOK).JSON(resp)
	}
}

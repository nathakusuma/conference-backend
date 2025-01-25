package service

import (
	"context"
	"errors"
	"github.com/nathakusuma/astungkara/domain/contract"
	"github.com/nathakusuma/astungkara/domain/errorpkg"
	"github.com/nathakusuma/astungkara/internal/infra/env"
	"github.com/nathakusuma/astungkara/pkg/bcrypt"
	"github.com/nathakusuma/astungkara/pkg/jwt"
	"github.com/nathakusuma/astungkara/pkg/log"
	"github.com/nathakusuma/astungkara/pkg/mail"
	"github.com/nathakusuma/astungkara/pkg/randgen"
	"github.com/nathakusuma/astungkara/pkg/uuidpkg"
	"net/url"
	"strconv"
)

type authService struct {
	repo    contract.IAuthRepository
	userSvc contract.IUserService
	bcrypt  bcrypt.IBcrypt
	jwt     jwt.IJwt
	mailer  mail.IMailer
	uuid    uuidpkg.IUUID
}

func NewAuthService(
	authRepo contract.IAuthRepository,
	userSvc contract.IUserService,
	bcrypt bcrypt.IBcrypt,
	jwt jwt.IJwt,
	mailer mail.IMailer,
	uuid uuidpkg.IUUID,
) contract.IAuthService {
	return &authService{
		repo:    authRepo,
		userSvc: userSvc,
		bcrypt:  bcrypt,
		jwt:     jwt,
		mailer:  mailer,
		uuid:    uuid,
	}
}

func (s *authService) RequestOTPRegisterUser(ctx context.Context, email string) error {
	// check if email is already registered
	_, err := s.userSvc.GetUserByEmail(ctx, email)
	if err == nil {
		return errorpkg.ErrEmailAlreadyRegistered
	}

	if !errors.Is(err, errorpkg.ErrNotFound) {
		traceID := log.ErrorWithTraceID(map[string]interface{}{
			"error": err.Error(),
			"email": email,
		}, "[AuthService][RequestOTPRegisterUser] failed to get user by email")

		return errorpkg.ErrInternalServer.WithTraceID(traceID)
	}

	// generate otp
	otp := strconv.Itoa(randgen.RandomNumber(6))

	// save otp
	err = s.repo.SetUserRegisterOTP(ctx, email, otp)
	if err != nil {
		traceID := log.ErrorWithTraceID(map[string]interface{}{
			"error": err.Error(),
			"email": email,
		}, "[AuthService][RequestOTPRegisterUser] failed to save otp")

		return errorpkg.ErrInternalServer.WithTraceID(traceID)
	}

	// send otp to email
	go func() {
		err = s.mailer.Send(
			email,
			"[Class Manager] Verify Your Account",
			"otp_register_user.html",
			map[string]interface{}{
				"otp":  otp,
				"href": env.GetEnv().FrontendURL + "/verify-email?email=" + url.QueryEscape(email) + "&otp=" + otp,
			})

		if err != nil {
			log.Error(map[string]interface{}{
				"error": err.Error(),
			}, "[AuthService][RequestOTPRegisterUser] failed to send email")
		}
	}()

	return nil
}

func (s *authService) CheckOTPRegisterUser(ctx context.Context, email, otp string) error {
	return nil
}

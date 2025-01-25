package service

import (
	"context"
	"errors"
	"github.com/nathakusuma/astungkara/domain/contract"
	"github.com/nathakusuma/astungkara/domain/dto"
	"github.com/nathakusuma/astungkara/domain/entity"
	"github.com/nathakusuma/astungkara/domain/errorpkg"
	"github.com/nathakusuma/astungkara/internal/infra/env"
	"github.com/nathakusuma/astungkara/pkg/bcrypt"
	"github.com/nathakusuma/astungkara/pkg/jwt"
	"github.com/nathakusuma/astungkara/pkg/log"
	"github.com/nathakusuma/astungkara/pkg/mail"
	"github.com/nathakusuma/astungkara/pkg/randgen"
	"github.com/nathakusuma/astungkara/pkg/uuidpkg"
	"github.com/redis/go-redis/v9"
	"net/url"
	"strconv"
	"time"
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
	savedOtp, err := s.repo.GetUserRegisterOTP(ctx, email)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return errorpkg.ErrInvalidOTP
		}

		traceID := log.ErrorWithTraceID(map[string]interface{}{
			"error": err.Error(),
			"email": email,
		}, "[AuthService][CheckOTPRegisterUser] failed to get otp")

		return errorpkg.ErrInternalServer.WithTraceID(traceID)
	}

	if savedOtp != otp {
		return errorpkg.ErrInvalidOTP
	}

	return nil
}

func (s *authService) LoginUser(ctx context.Context,
	req dto.LoginUserRequest) (dto.LoginResponse, error) {

	var resp dto.LoginResponse

	// get user by email
	user, err := s.userSvc.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, errorpkg.ErrNotFound) {
			return resp, errorpkg.ErrNotFound.WithMessage("User not found. Please register first.")
		}

		traceID := log.ErrorWithTraceID(map[string]interface{}{
			"error": err.Error(),
			"email": req.Email,
		}, "[AuthService][LoginUser] failed to get user by email")

		return resp, errorpkg.ErrInternalServer.WithTraceID(traceID)
	}

	// check password
	ok := s.bcrypt.Compare(req.Password, user.PasswordHash)
	if !ok {
		return resp, errorpkg.ErrCredentialsNotMatch
	}

	// Create channels for token generation results
	type tokenResult struct {
		token string
		err   error
	}
	accessTokenCh := make(chan tokenResult)
	refreshTokenCh := make(chan tokenResult)

	// Generate access token
	go func() {
		token, err := s.jwt.Create(user.ID, user.Role)
		accessTokenCh <- tokenResult{token: token, err: err}
	}()

	// Generate and store refresh token
	go func() {
		refreshToken := randgen.RandomString(64)
		err := s.repo.CreateSession(ctx, &entity.Session{
			UserID:    user.ID,
			Token:     refreshToken,
			ExpiresAt: time.Now().Add(env.GetEnv().JwtRefreshExpireDuration),
		})
		refreshTokenCh <- tokenResult{token: refreshToken, err: err}
	}()

	var accessResult, refreshResult tokenResult
	resultsReceived := 0

	// Wait for both operations to complete in any order
	for resultsReceived < 2 {
		select {
		case accessResult = <-accessTokenCh:
			if accessResult.err != nil {
				traceID := log.ErrorWithTraceID(map[string]interface{}{
					"error": accessResult.err.Error(),
					"email": req.Email,
				}, "[AuthService][LoginUser] failed to generate access token")
				return resp, errorpkg.ErrInternalServer.WithTraceID(traceID)
			}
			resultsReceived++

		case refreshResult = <-refreshTokenCh:
			if refreshResult.err != nil {
				traceID := log.ErrorWithTraceID(map[string]interface{}{
					"error": refreshResult.err.Error(),
					"email": req.Email,
				}, "[AuthService][LoginUser] failed to store session")
				return resp, errorpkg.ErrInternalServer.WithTraceID(traceID)
			}
			resultsReceived++
		}
	}

	userResp := dto.UserResponse{}
	userResp.PopulateFromEntity(user)
	resp = dto.LoginResponse{
		AccessToken:  accessResult.token,
		RefreshToken: refreshResult.token,
		User:         &userResp,
	}

	log.Info(map[string]interface{}{
		"email": req.Email,
	}, "[AuthService][LoginUser] user logged in")

	return resp, nil
}

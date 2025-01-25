package service

import (
	"context"
	"github.com/nathakusuma/astungkara/domain/contract"
	"github.com/nathakusuma/astungkara/pkg/bcrypt"
	"github.com/nathakusuma/astungkara/pkg/jwt"
	"github.com/nathakusuma/astungkara/pkg/mail"
	"github.com/nathakusuma/astungkara/pkg/uuidpkg"
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
	return nil
}

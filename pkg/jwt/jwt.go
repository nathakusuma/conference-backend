package jwt

import (
	"github.com/nathakusuma/conference-backend/domain/enum"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type IJwt interface {
	Create(userID uuid.UUID, role enum.UserRole) (string, error)
	Decode(tokenString string, claims *Claims) error
}

type Claims struct {
	jwt.RegisteredClaims
	Role enum.UserRole `json:"role"`
}

type JwtStruct struct {
	exp    time.Duration
	secret []byte
}

func NewJwt(exp time.Duration, secret []byte) IJwt {
	return &JwtStruct{
		exp:    exp,
		secret: secret,
	}
}

func (j *JwtStruct) Create(userID uuid.UUID, role enum.UserRole) (string, error) {
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "conference-backend",
			Subject:   userID.String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.exp)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		Role: role,
	}

	unsignedJWT := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedJWT, err := unsignedJWT.SignedString(j.secret)
	if err != nil {
		return "", err
	}

	return signedJWT, nil
}

func (j *JwtStruct) Decode(tokenString string, claims *Claims) error {
	token, err := jwt.ParseWithClaims(tokenString, claims, func(_ *jwt.Token) (any, error) {
		return j.secret, nil
	})

	if err != nil {
		return err
	}

	if !token.Valid {
		return jwt.ErrSignatureInvalid
	}

	return nil
}

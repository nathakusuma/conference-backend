package middleware

import "github.com/nathakusuma/conference-backend/pkg/jwt"

type Middleware struct {
	jwt jwt.IJwt
}

func NewMiddleware(
	jwt jwt.IJwt,
) *Middleware {
	return &Middleware{
		jwt: jwt,
	}
}

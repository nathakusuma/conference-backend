package middleware

import (
	"github.com/google/uuid"
	"github.com/nathakusuma/astungkara/domain/enum"
	"github.com/nathakusuma/astungkara/domain/errorpkg"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/nathakusuma/astungkara/pkg/jwt"
)

func (m *Middleware) RequireAuthenticated() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		header := ctx.Get("Authorization")
		if header == "" {
			return errorpkg.ErrNoBearerToken
		}

		headerSlice := strings.Split(header, " ")
		if len(headerSlice) != 2 && headerSlice[0] != "Bearer" {
			return errorpkg.ErrInvalidBearerToken
		}

		token := headerSlice[1]
		var claims jwt.Claims
		err := m.jwt.Decode(token, &claims)
		if err != nil {
			return errorpkg.ErrInvalidBearerToken
		}

		expirationTime, err := claims.GetExpirationTime()
		if err != nil {
			return errorpkg.ErrInvalidBearerToken
		}

		if expirationTime.Before(time.Now()) {
			return errorpkg.ErrInvalidBearerToken
		}

		ctx.Locals("user.id", uuid.MustParse(claims.Subject))
		ctx.Locals("user.role", claims.Role)

		return ctx.Next()
	}
}

// RequireOneOfRoles dependency: RequireAuthenticated
func (m *Middleware) RequireOneOfRoles(roles ...enum.UserRole) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		userRole := ctx.Locals("user.role").(enum.UserRole)

		for _, role := range roles {
			if userRole == role {
				return ctx.Next()
			}
		}

		return errorpkg.ErrForbiddenRole
	}
}

package server

import (
	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/nathakusuma/astungkara/internal/app/auth/handler"
	authrepo "github.com/nathakusuma/astungkara/internal/app/auth/repository"
	authsvc "github.com/nathakusuma/astungkara/internal/app/auth/service"
	userrepo "github.com/nathakusuma/astungkara/internal/app/user/repository"
	usersvc "github.com/nathakusuma/astungkara/internal/app/user/service"
	"github.com/nathakusuma/astungkara/internal/infra/env"
	"github.com/nathakusuma/astungkara/internal/middleware"
	"github.com/nathakusuma/astungkara/pkg/bcrypt"
	"github.com/nathakusuma/astungkara/pkg/jwt"
	"github.com/nathakusuma/astungkara/pkg/log"
	"github.com/nathakusuma/astungkara/pkg/mail"
	"github.com/nathakusuma/astungkara/pkg/uuidpkg"
	"github.com/nathakusuma/astungkara/pkg/validator"
	"github.com/redis/go-redis/v9"
)

type HttpServer interface {
	Start(part string)
	MountMiddlewares()
	MountRoutes(db *sqlx.DB, rds *redis.Client)
	GetApp() *fiber.App
}

type httpServer struct {
	app *fiber.App
}

func NewHttpServer() HttpServer {
	config := fiber.Config{
		AppName:      "Astungkara",
		JSONEncoder:  sonic.Marshal,
		JSONDecoder:  sonic.Unmarshal,
		ErrorHandler: ErrorHandler(),
	}

	app := fiber.New(config)

	return &httpServer{
		app: app,
	}
}

func (s *httpServer) GetApp() *fiber.App {
	return s.app
}

func (s *httpServer) Start(port string) {
	if port[0] != ':' {
		port = ":" + port
	}

	err := s.app.Listen(port)

	if err != nil {
		log.Fatal(map[string]interface{}{
			"error": err.Error(),
		}, "[SERVER][Start] failed to start server")
	}
}

func (s *httpServer) MountMiddlewares() {
	s.app.Use(middleware.LoggerConfig())
	s.app.Use(middleware.Helmet())
	s.app.Use(middleware.Compress())
	s.app.Use(middleware.Cors())
	s.app.Use(middleware.RecoverConfig())
}

func (s *httpServer) MountRoutes(db *sqlx.DB, rds *redis.Client) {
	bcryptInstance := bcrypt.GetBcrypt()
	jwtAccess := jwt.NewJwt(env.GetEnv().JwtAccessExpireDuration, env.GetEnv().JwtAccessSecretKey)
	mailer := mail.NewMailDialer()
	uuidInstance := uuidpkg.GetUUID()
	validatorInstance := validator.NewValidator()
	middlewareInstance := middleware.NewMiddleware(jwtAccess)

	s.app.Get("/", func(ctx *fiber.Ctx) error {
		return ctx.Status(fiber.StatusOK).SendString("Astungkara Healthy")
	})

	api := s.app.Group("/api")
	v1 := api.Group("/v1")

	userRepository := userrepo.NewUserRepository(db)
	authRepository := authrepo.NewAuthRepository(db, rds)

	userService := usersvc.NewUserService(userRepository, bcryptInstance, uuidInstance)
	authService := authsvc.NewAuthService(authRepository, userService, bcryptInstance, jwtAccess, mailer, uuidInstance)

	handler.InitAuthHandler(v1, middlewareInstance, validatorInstance, authService)
}

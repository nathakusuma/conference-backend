package server

import (
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/nathakusuma/astungkara/internal/app/user/repository"
	"github.com/nathakusuma/astungkara/internal/app/user/service"
	"github.com/nathakusuma/astungkara/internal/middleware"
	"github.com/nathakusuma/astungkara/pkg/bcrypt"
	"github.com/nathakusuma/astungkara/pkg/log"
	"github.com/nathakusuma/astungkara/pkg/uuidpkg"
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
	uuid := uuidpkg.GetUUID()
	bcryptInstance := bcrypt.GetBcrypt()

	s.app.Get("/", func(ctx *fiber.Ctx) error {
		return ctx.Status(fiber.StatusOK).SendString("Astungkara Healthy")
	})

	api := s.app.Group("/api")
	v1 := api.Group("/v1")

	userRepository := repository.NewUserRepository(db)

	userService := service.NewUserService(userRepository, bcryptInstance, uuid)

	fmt.Println(v1, userService) // TODO: Just to avoid unused variable error
}

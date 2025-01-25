package main

import (
	"github.com/nathakusuma/astungkara/internal/infra/database"
	"github.com/nathakusuma/astungkara/internal/infra/env"
	"github.com/nathakusuma/astungkara/internal/infra/redis"
	"github.com/nathakusuma/astungkara/internal/infra/server"
	"github.com/nathakusuma/astungkara/pkg/log"
)

func main() {
	env.NewEnv()
	log.NewLogger()

	srv := server.NewHttpServer()
	postgresDB := database.NewPostgresPool(
		env.GetEnv().DBHost,
		env.GetEnv().DBPort,
		env.GetEnv().DBUser,
		env.GetEnv().DBPass,
		env.GetEnv().DBName,
	)
	redisClient := redis.NewRedisPool(
		env.GetEnv().RedisHost,
		env.GetEnv().RedisPort,
		env.GetEnv().RedisPass,
		env.GetEnv().RedisDB,
	)
	defer postgresDB.Close()
	defer redisClient.Close()

	srv.MountMiddlewares()
	srv.MountRoutes(postgresDB, redisClient)
	srv.Start(env.GetEnv().AppPort)
}

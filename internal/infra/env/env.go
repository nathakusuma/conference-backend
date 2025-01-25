package env

import (
	"github.com/iamolegga/enviper"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"os"
	"sync"
)

type Env struct {
	AppEnv    string `mapstructure:"APP_ENV"`
	AppPort   string `mapstructure:"APP_PORT"`
	AppURL    string `mapstructure:"APP_URL"`
	DBHost    string `mapstructure:"DB_HOST"`
	DBPort    string `mapstructure:"DB_PORT"`
	DBUser    string `mapstructure:"DB_USER"`
	DBPass    string `mapstructure:"DB_PASS"`
	DBName    string `mapstructure:"DB_NAME"`
	RedisHost string `mapstructure:"REDIS_HOST"`
	RedisPort string `mapstructure:"REDIS_PORT"`
	RedisPass string `mapstructure:"REDIS_PASS"`
	RedisDB   int    `mapstructure:"REDIS_DB"`
}

var (
	viperInstance *viper.Viper
	env           *Env
	once          sync.Once
)

// NewEnv initializes and returns the environment configuration
func NewEnv() *Env {
	once.Do(func() {
		viperInstance = viper.New()

		env = &Env{}

		// Enable environment variables first
		viperInstance.AutomaticEnv()

		// Check if APP_ENV is set in environment variables
		if appEnv := os.Getenv("APP_ENV"); appEnv != "" {
			log.Info().Msgf("[ENV] Using %s environment variables", appEnv)

			// Unmarshal configuration with enviper due to issue with viper
			if err := enviper.New(viperInstance).Unmarshal(env); err != nil {
				log.Fatal().Msgf("[ENV] failed to unmarshal configuration: %s", err.Error())
				return
			}
		} else {
			// If APP_ENV not found in environment, try .env file
			if _, err := os.Stat(".env"); err != nil {
				log.Fatal().Msg("[ENV] APP_ENV is not set in environment variables")
				return
			}

			viperInstance.SetConfigFile(".env")
			if err := viperInstance.ReadInConfig(); err != nil {
				log.Fatal().Msg("[ENV] Failed to read .env file")
				return
			}

			log.Info().Msg("[ENV] Using .env file")

			// Unmarshal configuration
			if err := viperInstance.Unmarshal(env); err != nil {
				log.Fatal().Msgf("[ENV] failed to unmarshal configuration: %s", err.Error())
				return
			}
		}
	})

	return env
}

func GetEnv() *Env {
	return env
}

// SetEnv is used in testing to set the environment
func SetEnv(mockEnv *Env) {
	env = mockEnv
}

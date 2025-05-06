package setup

import (
	"github.com/nathakusuma/conference-backend/internal/infra/env"
	"github.com/nathakusuma/conference-backend/pkg/log"
)

func init() {
	// Set up test environment
	env.SetEnv(&env.Env{
		AppEnv: "test",
	})

	log.NewLogger()
}

package setup

import (
	"github.com/nathakusuma/astungkara/internal/infra/env"
	"github.com/nathakusuma/astungkara/pkg/log"
)

func init() {
	// Set up test environment
	env.SetEnv(&env.Env{
		AppEnv: "test",
	})

	log.NewLogger()
}

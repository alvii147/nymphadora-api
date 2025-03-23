package testkitinternal

import (
	"github.com/alvii147/nymphadora-api/internal/config"
	"github.com/alvii147/nymphadora-api/pkg/env"
	"github.com/alvii147/nymphadora-api/pkg/errutils"
)

// MustCreateConfig creates a new Config and panics on error.
func MustCreateConfig() *config.Config {
	cfg := &config.Config{}
	err := env.NewConfig(cfg)
	if err != nil {
		panic(errutils.FormatError(err))
	}

	return cfg
}

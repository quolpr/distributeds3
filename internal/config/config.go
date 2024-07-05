package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	DBURL string `envconfig:"DB_URL" required:"true"`
}

func FromEnv() (*Config, error) {
	cfg := new(Config)

	if err := envconfig.Process("", cfg); err != nil {
		return nil, fmt.Errorf("error while parse env config | %w", err)
	}

	return cfg, nil
}

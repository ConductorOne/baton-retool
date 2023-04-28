package main

import (
	"context"
	"fmt"

	"github.com/conductorone/baton-sdk/pkg/cli"
)

// config defines the external configuration required for the connector to run.
type config struct {
	cli.BaseConfig `mapstructure:",squash"` // Puts the base config options in the same place as the connector options

	ConnectionString string `mapstructure:"connection-string"`
}

// validateConfig is run after the configuration is loaded, and should return an error if it isn't valid.
func validateConfig(ctx context.Context, cfg *config) error {
	if cfg.ConnectionString == "" {
		return fmt.Errorf("--connection-string is required")
	}

	return nil
}

package config

import (
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

const (
	registryProvider = "registry_provider"
	registryURL      = "registry_url"
)

var validProviders = []string{"ecr"}

func init() {
	viper.SetDefault(registryProvider, "ecr")
	viper.SetDefault(registryURL, "215023620964.dkr.ecr.us-east-1.amazonaws.com")
	viper.SetDefault("aws_region", "us-east-1")
	viper.AutomaticEnv()
}

// Config represents all possible configurations for the application
type Config struct {
	RegistryProvider string
	RegistryURL      string
}

// Parse environment variables into a config struct
func Parse() (Config, error) {
	provider := viper.GetString(registryProvider)

	if err := validateRegistryProvider(provider); err != nil {
		return Config{}, errors.Wrap(err, "invalid configuration: ")
	}

	return Config{
		RegistryProvider: provider,
		RegistryURL:      viper.GetString(registryURL),
	}, nil
}

func validateRegistryProvider(provider string) error {
	for _, valid := range validProviders {
		if valid == provider {
			return nil
		}
	}

	return errors.Errorf("invalid registry provider '%s'.", provider)
}

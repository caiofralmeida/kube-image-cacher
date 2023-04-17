package config

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
)

type ConfigTestSuite struct {
	suite.Suite
}

func TestRunConfigTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}

func (s *ConfigTestSuite) SetupTest() {
	viper.Reset()
}

func (s *ConfigTestSuite) TestParse_InvalidRegistryProvider() {
	viper.Set("registry_provider", "foo")
	_, err := Parse()
	s.Error(err)
}

func (s *ConfigTestSuite) TestParse_EnvVarWithValues() {
	viper.Set("registry_provider", "ecr")
	viper.Set("registry_url", "my.registry.foo")

	c, err := Parse()

	s.NoError(err)
	s.Equal("ecr", c.RegistryProvider)
	s.Equal("my.registry.foo", c.RegistryURL)
}

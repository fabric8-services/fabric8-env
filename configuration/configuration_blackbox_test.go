package configuration_test

import (
	"os"
	"testing"

	testsuite "github.com/fabric8-services/fabric8-common/test/suite"
	"github.com/fabric8-services/fabric8-env/configuration"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ConfigurationTestSuite struct {
	testsuite.UnitTestSuite
}

func TestConfiguration(t *testing.T) {
	suite.Run(t, &ConfigurationTestSuite{})
}

func (s *ConfigurationTestSuite) TestAuthURL() {
	existingAuthURL := os.Getenv("F8_AUTH_URL")
	existingDevMode := os.Getenv("F8_DEVELOPER_MODE_ENABLED")
	defer func() {
		os.Setenv("F8_AUTH_URL", existingAuthURL)
		os.Setenv("F8_DEVELOPER_MODE_ENABLED", existingDevMode)
	}()

	// Default in dev mode
	os.Setenv("F8_DEVELOPER_MODE_ENABLED", "true")
	os.Unsetenv("F8_AUTH_URL")
	config, err := configuration.New("")
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "https://auth.prod-preview.openshift.io", config.GetAuthServiceURL())

	// Explicitly set via AUTH_WIT_URL env var
	os.Setenv("F8_AUTH_URL", "https://api.some.io")
	config, err = configuration.New("")
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "https://api.some.io", config.GetAuthServiceURL())

	os.Unsetenv("F8_DEVELOPER_MODE_ENABLED")
	config, err = configuration.New("")
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "https://api.some.io", config.GetAuthServiceURL())
}

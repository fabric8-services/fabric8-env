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

func (s *ConfigurationTestSuite) TestConfigErr() {
	t := s.T()
	currPgPass := os.Getenv("F8_POSTGRES_PASSWORD")
	currSentryDSN := os.Getenv("F8_SENTRY_DSN")
	currAuthURL := os.Getenv("F8_AUTH_URL")
	currClusterURL := os.Getenv("F8_CLUSTER_URL")
	defer func() {
		os.Setenv("F8_POSTGRES_PASSWORD", currPgPass)
		os.Setenv("F8_SENTRY_DSN", currSentryDSN)
		os.Setenv("F8_AUTH_URL", currAuthURL)
		os.Setenv("F8_CLUSTER_URL", currClusterURL)
	}()

	configErr := createConfigGetConfigErr(t)
	assert.Equal(t, "default DB password is used; Sentry DSN is empty; Auth service url is empty; Cluster service url is empty", configErr.Error())

	os.Setenv("F8_POSTGRES_PASSWORD", "abcd1234")
	configErr = createConfigGetConfigErr(t)
	assert.Equal(t, "Sentry DSN is empty; Auth service url is empty; Cluster service url is empty", configErr.Error())

	os.Setenv("F8_SENTRY_DSN", "https://somedsn.com")
	configErr = createConfigGetConfigErr(t)
	assert.Equal(t, "Auth service url is empty; Cluster service url is empty", configErr.Error())

	os.Setenv("F8_AUTH_URL", "https://someauth.com")
	configErr = createConfigGetConfigErr(t)
	assert.Equal(t, "Cluster service url is empty", configErr.Error())

	os.Unsetenv("F8_AUTH_URL")
	os.Setenv("F8_CLUSTER_URL", "https://somecluster.com")
	configErr = createConfigGetConfigErr(t)
	assert.Equal(t, "Auth service url is empty", configErr.Error())
}

func createConfigGetConfigErr(t *testing.T) error {
	config, err := configuration.New("")
	require.NoError(t, err)
	configErr := config.DefaultConfigError()
	assert.NotNil(t, configErr)
	return configErr
}

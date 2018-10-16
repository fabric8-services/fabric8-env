package configuration

import (
	"fmt"
	"os"
	"strings"
	"time"

	errs "github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

func (c *Registry) String() string {
	allSettings := c.v.AllSettings()
	y, err := yaml.Marshal(&allSettings)
	if err != nil {
		log.WithFields(map[string]interface{}{
			"settings": allSettings,
			"err":      err,
		}).Panicln("Failed to marshall config to string")
	}
	return fmt.Sprintf("%s\n", y)
}

const (
	varLogLevel             = "log.level"
	varLogJSON              = "log.json"
	varAuthURL              = "auth.url"
	varKeysTokenPath        = "auth.keys.token.path"
	varHTTPAddress          = "http.address"
	varMetricsHTTPAddress   = "metrics.http.address"
	varDiagnoseHTTPAddress  = "diagnose.http.address"
	varEnvironment          = "environment"
	varDeveloperModeEnabled = "developer.mode.enabled"
	varSentryDSN            = "sentry.dsn"

	// postgres
	varPostgresHost                 = "postgres.host"
	varPostgresPort                 = "postgres.port"
	varPostgresUser                 = "postgres.user"
	varPostgresDatabase             = "postgres.database"
	varPostgresPassword             = "postgres.password"
	varPostgresSSLMode              = "postgres.sslmode"
	varPostgresConnectionTimeout    = "postgres.connection.timeout"
	varPostgresConnectionRetrySleep = "postgres.connection.retrysleep"
	varPostgresConnectionMaxIdle    = "postgres.connection.maxidle"
	varPostgresConnectionMaxOpen    = "postgres.connection.maxopen"

	defaultLogLevel = "info"
)

type Registry struct {
	v *viper.Viper
}

func New(configFilePath string) (*Registry, error) {
	c := Registry{
		v: viper.New(),
	}
	c.v.SetEnvPrefix("F8")
	c.v.AutomaticEnv()
	c.v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	c.v.SetTypeByDefaultValue(true)
	c.setConfigDefaults()

	if configFilePath != "" {
		c.v.SetConfigType("yaml")
		c.v.SetConfigFile(configFilePath)
		err := c.v.ReadInConfig()
		if err != nil {
			return nil, errs.Errorf("Fatal error config file: %s \n", err)
		}
	}
	return &c, nil
}

func getConfigFilePath() string {
	// This was either passed as a env var Or, set inside main.go from --config
	envConfigPath, ok := os.LookupEnv("F8_CONFIG_FILE_PATH")
	if !ok {
		return ""
	}
	return envConfigPath
}

// Get is a wrapper over New() which reads configuration file path from the environment variable.
func Get() (*Registry, error) {
	cd, err := New(getConfigFilePath())
	return cd, err
}

func (c *Registry) setConfigDefaults() {
	c.v.SetDefault(varLogLevel, defaultLogLevel)
	c.v.SetDefault(varHTTPAddress, "0.0.0.0:8080")
	c.v.SetDefault(varMetricsHTTPAddress, "0.0.0.0:8080")
	c.v.SetDefault(varDeveloperModeEnabled, false)
}

func (c *Registry) GetLogLevel() string {
	return c.v.GetString(varLogLevel)
}

func (c *Registry) IsLogJSON() bool {
	if c.v.IsSet(varLogJSON) {
		return c.v.GetBool(varLogJSON)
	}
	if c.DeveloperModeEnabled() {
		return false
	}
	return true
}

func (c *Registry) GetKeysTokenPath() string {
	return c.v.GetString(varKeysTokenPath)
}

func (c *Registry) GetAuthServiceURL() string {
	return c.v.GetString(varAuthURL)
}

func (c *Registry) GetHTTPAddress() string {
	return c.v.GetString(varHTTPAddress)
}

func (c *Registry) GetMetricsHTTPAddress() string {
	return c.v.GetString(varMetricsHTTPAddress)
}

func (c *Registry) GetDiagnoseHTTPAddress() string {
	if c.v.IsSet(varDiagnoseHTTPAddress) {
		return c.v.GetString(varDiagnoseHTTPAddress)
	} else if c.DeveloperModeEnabled() {
		return "127.0.0.1:0"
	}
	return ""
}

func (c *Registry) GetEnvironment() string {
	if c.v.IsSet(varEnvironment) {
		return c.v.GetString(varEnvironment)
	}
	return "local"
}

func (c *Registry) DeveloperModeEnabled() bool {
	return c.v.GetBool(varDeveloperModeEnabled)
}

func (c *Registry) GetPostgresHost() string {
	return c.v.GetString(varPostgresHost)
}

func (c *Registry) GetPostgresPort() int64 {
	return c.v.GetInt64(varPostgresPort)
}

func (c *Registry) GetPostgresUser() string {
	return c.v.GetString(varPostgresUser)
}

func (c *Registry) GetPostgresPassword() string {
	return c.v.GetString(varPostgresPassword)
}

func (c *Registry) GetPostgresDatabase() string {
	return c.v.GetString(varPostgresDatabase)
}

func (c *Registry) GetPostgresSSLMode() string {
	return c.v.GetString(varPostgresSSLMode)
}

func (c *Registry) GetPostgresConnectionTimeout() int64 {
	return c.v.GetInt64(varPostgresConnectionTimeout)
}

func (c *Registry) GetPostgresConfigString() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s connect_timeout=%d",
		c.GetPostgresHost(),
		c.GetPostgresPort(),
		c.GetPostgresUser(),
		c.GetPostgresPassword(),
		c.GetPostgresDatabase(),
		c.GetPostgresSSLMode(),
		c.GetPostgresConnectionTimeout(),
	)
}

func (c *Registry) GetPostgresConnectionRetrySleep() time.Duration {
	return c.v.GetDuration(varPostgresConnectionRetrySleep)
}

func (c *Registry) IsPostgresDeveloperModeEnabled() bool {
	return c.v.GetBool(varDeveloperModeEnabled)
}

func (c *Registry) GetPostgresConnectionMaxIdle() int {
	return c.v.GetInt(varPostgresConnectionMaxIdle)
}

func (c *Registry) GetPostgresConnectionMaxOpen() int {
	return c.v.GetInt(varPostgresConnectionMaxOpen)
}

func (c *Registry) GetSentryDSN() string {
	return c.v.GetString(varSentryDSN)
}

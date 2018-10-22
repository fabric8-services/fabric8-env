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
	varAuthKeysPath         = "auth.keys.path"
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
	varPostgresTransactionTimeout   = "postgres.transaction.timeout"
)

const (
	defaultLogLevel   = "info"
	defaultDBPassword = "mysecretpassword"
)

// DevModePrivateKey the private key used to sign some tokens
// that are used in the tests.
var DevModePrivateKey = []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA40yB6SNoU4SpWxTfG5ilu+BlLYikRyyEcJIGg//w/GyqtjvT
/CVo92DRTh/DlrgwjSitmZrhauBnrCOoUBMin0/TXeSo3w2M5tEiiIFPbTDRf2jM
fbSGEOke9O0USCCR+bM2TncrgZR74qlSwq38VCND4zHc89rAzqJ2LVM2aXkuBbO7
TcgLNyooBrpOK9khVHAD64cyODAdJY4esUjcLdlcB7TMDGOgxGGn2RARU7+TUf32
gZZbTMikbuPM5gXuzGlo/22ECbQSKuZpbGwgPIAZ5NN9QA4D1NRz9+KDoiXZ6deZ
TTVCrZykJJ6RyLNfRh+XS+6G5nvcqAmfBpyOWwIDAQABAoIBAE5pBie23zZwfTu+
Z3jNn96/+idLC+DBqq5qsXS3xhpOIlXbLbW98gfkjk+1BXPo9la7wadLlpeX8iuf
4WA+OaNblj69ssO/mOvHGXKdqRixzpN1Q5XZwKX0xYkYf/ahxbmt6P4IfimlX1dB
shsWigU8ZR7rBJ3ayMh/ouTf39ViIbXsHYpEubmACcLaOlXbEuZNr7ofkFQKl/mh
XLWUeOoM97xY6Agw/gv60GIcxIC5OAg7iNqS+XNzhba7f2nf2YqodbN9H1BmEJsf
RRaTTWlZAiQXC8lpZOKwP7DiMLOT78lfmlYtquEBhwRbXazfzsdf67Mr4Kdl2Cej
Jy0EGwECgYEA/DZWB0Lb0tPdT1FmORNrBfGg3PjhX9FOilhbtUgX3nNKp8Zsi3yO
yN6hf0/98qIGlmAQi5C92cXpdhqTiVAGktWD+q0a1W99udIjinS1tFrKgNtOyBWN
uwDBZyhw8RrwpQinMe7B966SVDaphvvOWlB1TadMDh5kReJCYpvRCrMCgYEA5rZj
djCU2UqMw6jIP07nCFjWgxPPjg7jP8aRo07oW2mv1sEA0doCyoZaMrdNeGd3fB0B
sm+IvlQtWD7r0tWZI1GkYpdRkDFurdkIzVPV5pMwH4ByOq/Jf5ZqtjIpoMaRBirA
whJyjmiGU3yDyPDLtEFpNgqM3mIyxS6M6UGKYbkCgYEAg6w+d6YBK+1uQiXGD5BC
tKS0jgjlaOfWcEW3A0qzI3Dfjf3610vdI6OPfu8dLppGhCV9HdAgPdykiQNQ+UQt
WmVcdPgA5WNCqUu7QGK0Joer52AXnkAacYHwdtHXPRkKf66n01rKK2wZexvan91A
m0gcJcFs5IYbZZy9ecvNdB8CgYEAo4JZ5Vay93j1YGnLWcrixDCp/wXYUJbOidGC
QBpZZQf3Hh11JkT7O2uSm2T727yAmw63uC2B3VotNOCLI8ZMHRLsjQ8vOCFAjqdF
rLeg3iQss/bFfkA9b1Y8VNoiVJbGC3fbWu/WDoWXxa12fL/jruG43hsGEUnJL6Q5
K8tOdskCgYABpoHFRxsvJ5Sp9CUS3BBTicVSkpAjoX2O3+cS9XL8IsIqZEMW7VKb
16/H2BRvI0uUq12t+UCc0P0SyrWRGxwGR5zSYHVDOot5EDHqE8aYSbX4jiXtAAiu
qCn3Rug8QWyBjjxnU3CxPRiLSmEllQAAVlzfRWn6kL4RKSyruUhZaA==
-----END RSA PRIVATE KEY-----`)

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

	// db
	c.v.SetDefault(varPostgresHost, "localhost")
	c.v.SetDefault(varPostgresPort, 5432)
	c.v.SetDefault(varPostgresUser, "postgres")
	c.v.SetDefault(varPostgresDatabase, "postgres")
	c.v.SetDefault(varPostgresPassword, defaultDBPassword)
	c.v.SetDefault(varPostgresSSLMode, "disable")
	c.v.SetDefault(varPostgresConnectionTimeout, 5)
	c.v.SetDefault(varPostgresConnectionMaxIdle, -1)
	c.v.SetDefault(varPostgresConnectionMaxOpen, -1)
	c.v.SetDefault(varPostgresConnectionRetrySleep, time.Duration(time.Second))
	c.v.SetDefault(varPostgresTransactionTimeout, time.Duration(5*time.Minute))

	// others
	c.v.SetDefault(varAuthKeysPath, "/api/token/keys")
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

func (c *Registry) GetAuthKeysPath() string {
	return c.v.GetString(varAuthKeysPath)
}

func (c *Registry) GetAuthServiceURL() string {
	return c.v.GetString(varAuthURL)
}

func (c *Registry) GetDevModePrivateKey() []byte {
	if c.DeveloperModeEnabled() {
		return DevModePrivateKey
	}
	return nil
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

func (c *Registry) GetPostgresTransactionTimeout() time.Duration {
	return c.v.GetDuration(varPostgresTransactionTimeout)
}

package controller_test

import (
	"errors"
	"testing"
	"time"

	testsuite "github.com/fabric8-services/fabric8-common/test/suite"
	"github.com/fabric8-services/fabric8-env/app"
	"github.com/fabric8-services/fabric8-env/app/test"
	"github.com/fabric8-services/fabric8-env/configuration"
	"github.com/fabric8-services/fabric8-env/controller"
	"github.com/goadesign/goa"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type StatusControllerSuite struct {
	testsuite.DBTestSuite
}

type testStatusConfig struct {
	devMode   bool
	configErr error
}

func (s *testStatusConfig) DeveloperModeEnabled() bool {
	return s.devMode
}

func (s *testStatusConfig) DefaultConfigError() error {
	return s.configErr
}

type testDBChecker struct {
}

func (t *testDBChecker) Ping() error {
	return errors.New("DB is unreachable")
}

func TestStatusController(t *testing.T) {
	config, err := configuration.New("")
	require.NoError(t, err)
	suite.Run(t, &StatusControllerSuite{DBTestSuite: testsuite.NewDBTestSuite(config)})
}

func (s *StatusControllerSuite) UnSecuredController(config *testStatusConfig) (*goa.Service, *controller.StatusController) {
	svc := goa.New("status-test")
	ctrl := controller.NewStatusController(svc, controller.NewGormDBChecker(s.DBTestSuite.DB), config)
	return svc, ctrl
}

func (s *StatusControllerSuite) UnSecuredControllerWithUnreachableDB(config *testStatusConfig) (*goa.Service, *controller.StatusController) {
	svc := goa.New("status-test")
	ctrl := controller.NewStatusController(svc, &testDBChecker{}, config)
	return svc, ctrl
}

func (s *StatusControllerSuite) TestShow() {

	s.T().Run("ok_with_dev_mode", func(t *testing.T) {
		config := &testStatusConfig{
			devMode:   true,
			configErr: nil,
		}
		svc, ctrl := s.UnSecuredController(config)

		_, got := test.ShowStatusOK(t, svc.Context, svc, ctrl)

		checkStatus(t, got)
		assert.Equal(t, true, *got.DevMode)
		assert.Equal(t, "OK", got.DatabaseStatus)
		assert.Equal(t, "OK", got.ConfigurationStatus)
	})

	s.T().Run("ok_with_dev_mode_with_config_issue", func(t *testing.T) {
		config := &testStatusConfig{
			devMode:   true,
			configErr: errors.New("configuration error"),
		}
		svc, ctrl := s.UnSecuredController(config)

		_, got := test.ShowStatusOK(t, svc.Context, svc, ctrl)

		checkStatus(t, got)
		assert.Equal(t, true, *got.DevMode)
		assert.Equal(t, "OK", got.DatabaseStatus)
		assert.Equal(t, "Error: configuration error", got.ConfigurationStatus)
	})

	s.T().Run("service_unavailable_with_config_issue", func(t *testing.T) {
		config := &testStatusConfig{
			devMode:   false,
			configErr: errors.New("configuration error"),
		}
		svc, ctrl := s.UnSecuredController(config)

		_, got := test.ShowStatusServiceUnavailable(t, svc.Context, svc, ctrl)

		assert.Nil(t, got.DevMode)
		assert.Equal(t, "OK", got.DatabaseStatus)
		assert.Equal(t, "Error: configuration error", got.ConfigurationStatus)
	})

	s.T().Run("service_unavailable_with_db_issue", func(t *testing.T) {
		config := &testStatusConfig{
			devMode:   true,
			configErr: nil,
		}
		svc, ctrl := s.UnSecuredControllerWithUnreachableDB(config)

		_, got := test.ShowStatusServiceUnavailable(t, svc.Context, svc, ctrl)

		assert.Equal(t, true, *got.DevMode)
		assert.Equal(t, "Error: DB is unreachable", got.DatabaseStatus)
		assert.Equal(t, "OK", got.ConfigurationStatus)
	})

}

func checkStatus(t *testing.T, got *app.Status) {
	t.Helper()
	assert.Equal(t, app.Commit, got.Commit)
	assert.Equal(t, app.StartTime, got.StartTime)
	_, err := time.Parse("2006-01-02T15:04:05Z", got.StartTime)
	assert.Nil(t, err, "Incorrect layout of StartTime")
}

package controller_test

import (
	"errors"
	"os"
	"strconv"
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

const (
	wantConfigErrMsgForDevMode  = "Error: developer Mode is enabled; default DB password is used; Sentry DSN is empty"
	wantConfigErrMsgForProdMode = "Error: default DB password is used; Sentry DSN is empty; Auth service url is empty; Cluster service url is empty"
)

type StatusControllerSuite struct {
	testsuite.DBTestSuite
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

func (s *StatusControllerSuite) UnSecuredController() (*goa.Service, *controller.StatusController) {
	config, err := configuration.New("")
	require.NoError(s.T(), err)
	svc := goa.New("status-test")
	ctrl := controller.NewStatusController(svc, controller.NewGormDBChecker(s.DBTestSuite.DB), config)
	return svc, ctrl
}

func (s *StatusControllerSuite) UnSecuredControllerWithUnreachableDB() (*goa.Service, *controller.StatusController) {
	config, err := configuration.New("")
	require.NoError(s.T(), err)
	svc := goa.New("status-test")
	ctrl := controller.NewStatusController(svc, &testDBChecker{}, config)
	return svc, ctrl
}

func (s *StatusControllerSuite) TestShow() {

	s.T().Run("ok_with_dev_mode", func(t *testing.T) {
		currDevMode := os.Getenv("F8_DEVELOPER_MODE_ENABLED")
		defer os.Setenv("F8_DEVELOPER_MODE_ENABLED", currDevMode)

		wantDevMode := true
		os.Setenv("F8_DEVELOPER_MODE_ENABLED", strconv.FormatBool(wantDevMode))
		svc, ctrl := s.UnSecuredController()

		_, got := test.ShowStatusOK(t, svc.Context, svc, ctrl)

		checkStatus(t, got)
		assert.Equal(t, wantDevMode, *got.DevMode)
		assert.Equal(t, "OK", got.DatabaseStatus)
		assert.Equal(t, wantConfigErrMsgForDevMode, got.ConfigurationStatus)
	})

	s.T().Run("ok_with_dev_mode_with_config_issue", func(t *testing.T) {
		currDevMode := os.Getenv("F8_DEVELOPER_MODE_ENABLED")
		defer os.Setenv("F8_DEVELOPER_MODE_ENABLED", currDevMode)

		wantDevMode := true
		os.Setenv("F8_DEVELOPER_MODE_ENABLED", strconv.FormatBool(wantDevMode))
		svc, ctrl := s.UnSecuredController()

		_, got := test.ShowStatusOK(t, svc.Context, svc, ctrl)

		checkStatus(t, got)
		assert.Equal(t, wantDevMode, *got.DevMode)
		assert.Equal(t, "OK", got.DatabaseStatus)
		assert.Equal(t, wantConfigErrMsgForDevMode, got.ConfigurationStatus)
	})

	s.T().Run("service_unavailable_with_config_issue", func(t *testing.T) {
		currDevMode := os.Getenv("F8_DEVELOPER_MODE_ENABLED")
		defer os.Setenv("F8_DEVELOPER_MODE_ENABLED", currDevMode)

		os.Setenv("F8_DEVELOPER_MODE_ENABLED", "false")
		svc, ctrl := s.UnSecuredController()

		_, got := test.ShowStatusServiceUnavailable(t, svc.Context, svc, ctrl)

		assert.Nil(t, got.DevMode)
		assert.Equal(t, "OK", got.DatabaseStatus)
		assert.Equal(t, wantConfigErrMsgForProdMode, got.ConfigurationStatus)
	})

	s.T().Run("service_unavailable_with_db_issue", func(t *testing.T) {
		currDevMode := os.Getenv("F8_DEVELOPER_MODE_ENABLED")
		defer os.Setenv("F8_DEVELOPER_MODE_ENABLED", currDevMode)

		wantDevMode := true
		os.Setenv("F8_DEVELOPER_MODE_ENABLED", strconv.FormatBool(wantDevMode))
		svc, ctrl := s.UnSecuredControllerWithUnreachableDB()

		_, got := test.ShowStatusServiceUnavailable(t, svc.Context, svc, ctrl)

		assert.Equal(t, wantDevMode, *got.DevMode)
		assert.Equal(t, "Error: DB is unreachable", got.DatabaseStatus)
		assert.Equal(t, wantConfigErrMsgForDevMode, got.ConfigurationStatus)
	})

}

func checkStatus(t *testing.T, got *app.Status) {
	t.Helper()
	assert.Equal(t, app.Commit, got.Commit)
	assert.Equal(t, app.StartTime, got.StartTime)
	_, err := time.Parse("2006-01-02T15:04:05Z", got.StartTime)
	assert.Nil(t, err, "Incorrect layout of StartTime")
}

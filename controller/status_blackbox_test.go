package controller_test

import (
	"context"
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

func TestStatusController(t *testing.T) {
	config, err := configuration.New("")
	require.NoError(t, err)
	suite.Run(t, &StatusControllerSuite{DBTestSuite: testsuite.NewDBTestSuite(config)})
}

func (s *StatusControllerSuite) TestShowStatus() {
	service := goa.New("status-test")
	ctrl := controller.NewStatusController(service, controller.NewGormDBChecker(s.DBTestSuite.DB))

	_, res := test.ShowStatusOK(s.T(), context.Background(), service, ctrl)

	assert.Equal(s.T(), app.Commit, res.Commit, "Commit is not correct")
	assert.Equal(s.T(), app.StartTime, res.StartTime, "StartTime is not correct")
	_, err := time.Parse("2006-01-02T15:04:05Z", res.StartTime)
	assert.Nil(s.T(), err, "Incorrect layout of StartTime")
	assert.Equal(s.T(), "OK", res.DatabaseStatus)
}

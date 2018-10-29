package controller_test

import (
	"context"
	"testing"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"

	testsuite "github.com/fabric8-services/fabric8-common/test/suite"
	"github.com/fabric8-services/fabric8-env/app"
	"github.com/fabric8-services/fabric8-env/app/test"
	"github.com/fabric8-services/fabric8-env/configuration"
	"github.com/fabric8-services/fabric8-env/controller"
	"github.com/fabric8-services/fabric8-env/gormapp"
	"github.com/fabric8-services/fabric8-env/resource"
	"github.com/goadesign/goa"
	"github.com/stretchr/testify/suite"
)

type EnvironmentControllerSuite struct {
	testsuite.DBTestSuite
}

func TestEnvironmentController(t *testing.T) {
	resource.Require(t, resource.Database)
	config, err := configuration.New("")
	require.NoError(t, err)
	suite.Run(t, &EnvironmentControllerSuite{DBTestSuite: testsuite.NewDBTestSuite(config)})
}

func (s *EnvironmentControllerSuite) TestCreateEnvironment() {
	service := goa.New("enviroment-test")
	appDB := gormapp.NewGormDB(s.DB)
	ctrl := controller.NewEnvironmentController(service, appDB)

	spaceID, err := uuid.FromString("f03f023b-0427-4cdb-924b-fb2369018ab6")
	require.NoError(s.T(), err)
	payload := newCreateEnvironmentPayload("osio-stage", "stage", "cluster1.com")
	test.CreateEnvironmentCreated(s.T(), context.Background(), service, ctrl, spaceID, payload)
}

func (s *EnvironmentControllerSuite) TestListEnvironment() {
	service := goa.New("enviroment-test")
	appDB := gormapp.NewGormDB(s.DB)
	ctrl := controller.NewEnvironmentController(service, appDB)

	spaceID, err := uuid.FromString("f03f023b-0427-4cdb-924b-fb2369018ab6")
	require.NoError(s.T(), err)
	payload := newCreateEnvironmentPayload("osio-stage", "stage", "cluster1.com")
	test.CreateEnvironmentCreated(s.T(), context.Background(), service, ctrl, spaceID, payload)

	_, list := test.ListEnvironmentOK(s.T(), context.Background(), service, ctrl, spaceID, nil, nil)
	require.NotNil(s.T(), list)
	require.NotEmpty(s.T(), list.Data)
}

func newCreateEnvironmentPayload(name, envType, clusterURL string) *app.CreateEnvironmentPayload {
	payload := &app.CreateEnvironmentPayload{
		Data: &app.Environment{
			Attributes: &app.EnvironmentAttributes{
				Name:       &name,
				Type:       &envType,
				ClusterURL: &clusterURL,
			},
			Type: "environments",
		},
	}
	return payload
}

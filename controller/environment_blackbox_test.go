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
	"github.com/goadesign/goa"
	"github.com/stretchr/testify/suite"
)

type EnvironmentControllerSuite struct {
	testsuite.DBTestSuite
	service *goa.Service
	db      *gormapp.GormDB
	ctrl    *controller.EnvironmentController
}

func TestEnvironmentController(t *testing.T) {
	config, err := configuration.New("")
	require.NoError(t, err)
	suite.Run(t, &EnvironmentControllerSuite{DBTestSuite: testsuite.NewDBTestSuite(config)})
}

func (s *EnvironmentControllerSuite) SetupSuite() {
	s.DBTestSuite.SetupSuite()

	s.service = goa.New("enviroment-test")
	s.db = gormapp.NewGormDB(s.DB)
	s.ctrl = controller.NewEnvironmentController(s.service, s.db)
}
func (s *EnvironmentControllerSuite) TestCreateEnvironment() {
	spaceID, err := uuid.FromString("f03f023b-0427-4cdb-924b-fb2369018ab6")
	require.NoError(s.T(), err)
	payload := newCreateEnvironmentPayload("osio-stage", "stage", "cluster1.com")
	test.CreateEnvironmentCreated(s.T(), context.Background(), s.service, s.ctrl, spaceID, payload)
}

func (s *EnvironmentControllerSuite) TestListEnvironment() {
	spaceID, err := uuid.FromString("f03f023b-0427-4cdb-924b-fb2369018ab6")
	require.NoError(s.T(), err)
	payload := newCreateEnvironmentPayload("osio-stage", "stage", "cluster1.com")
	test.CreateEnvironmentCreated(s.T(), context.Background(), s.service, s.ctrl, spaceID, payload)

	_, list := test.ListEnvironmentOK(s.T(), context.Background(), s.service, s.ctrl, spaceID)
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

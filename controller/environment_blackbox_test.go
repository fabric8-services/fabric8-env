package controller_test

import (
	"context"
	"testing"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
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
	// TODO add test for all different response code
}

func (s *EnvironmentControllerSuite) SetupSuite() {
	s.DBTestSuite.SetupSuite()

	s.service = goa.New("enviroment-test")
	s.db = gormapp.NewGormDB(s.DB)
	s.ctrl = controller.NewEnvironmentController(s.service, s.db)
}
func (s *EnvironmentControllerSuite) TestCreate() {
	spaceID := uuid.NewV4()
	payload := newCreateEnvironmentPayload("osio-stage", "stage", "cluster1.com")

	_, newEnv := test.CreateEnvironmentCreated(s.T(), context.Background(), s.service, s.ctrl, spaceID, payload)

	assert.NotNil(s.T(), newEnv)
	assert.NotNil(s.T(), newEnv.Data.ID)
	_, env := test.ShowEnvironmentOK(s.T(), context.Background(), s.service, s.ctrl, *newEnv.Data.ID)
	require.NotNil(s.T(), env)
	assert.Equal(s.T(), env.Data.ID, newEnv.Data.ID)
}

func (s *EnvironmentControllerSuite) TestList() {
	spaceID := uuid.NewV4()
	payload := newCreateEnvironmentPayload("osio-stage", "stage", "cluster1.com")
	_, newEnv := test.CreateEnvironmentCreated(s.T(), context.Background(), s.service, s.ctrl, spaceID, payload)
	require.NotNil(s.T(), newEnv)

	_, list := test.ListEnvironmentOK(s.T(), context.Background(), s.service, s.ctrl, spaceID)
	assert.NotNil(s.T(), list)
	assert.NotEmpty(s.T(), list.Data)
	assert.Equal(s.T(), newEnv.Data.ID, list.Data[0].ID)
}

func (s *EnvironmentControllerSuite) TestShow() {
	spaceID := uuid.NewV4()
	payload := newCreateEnvironmentPayload("osio-stage", "stage", "cluster1.com")
	_, newEnv := test.CreateEnvironmentCreated(s.T(), context.Background(), s.service, s.ctrl, spaceID, payload)
	require.NotNil(s.T(), newEnv)

	_, env := test.ShowEnvironmentOK(s.T(), context.Background(), s.service, s.ctrl, *newEnv.Data.ID)
	assert.NotNil(s.T(), env)
	assert.Equal(s.T(), newEnv.Data.ID, env.Data.ID)
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

package controller_test

import (
	"context"
	"testing"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	testauth "github.com/fabric8-services/fabric8-common/test/auth"
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
	db *gormapp.GormDB

	svc      *goa.Service // secure
	svc2     *goa.Service // unsecure
	ctx      context.Context
	ctx2     context.Context
	ctrl     *controller.EnvironmentController
	ctrl2    *controller.EnvironmentController
	prodCtrl *controller.EnvironmentController
}

func TestEnvironmentController(t *testing.T) {
	config, err := configuration.New("")
	require.NoError(t, err)
	suite.Run(t, &EnvironmentControllerSuite{DBTestSuite: testsuite.NewDBTestSuite(config)})
}

func (s *EnvironmentControllerSuite) SetupSuite() {
	s.DBTestSuite.SetupSuite()

	s.db = gormapp.NewGormDB(s.DB)
	svc, err := testauth.ServiceAsUser("enviroment-test1", testauth.NewIdentity())
	require.NoError(s.T(), err)
	s.svc = svc
	s.svc2 = testauth.UnsecuredService("enviroment-test2")
	s.ctx = s.svc.Context
	s.ctx2 = s.svc2.Context
	s.ctrl = controller.NewEnvironmentController(s.svc, s.db, true)
	s.ctrl2 = controller.NewEnvironmentController(s.svc2, s.db, true)
	s.prodCtrl = controller.NewEnvironmentController(s.svc, s.db, false)
}
func (s *EnvironmentControllerSuite) TestCreate() {
	s.T().Run("ok", func(t *testing.T) {
		spaceID := uuid.NewV4()
		payload := newCreateEnvironmentPayload("osio-stage", "stage", "cluster1.com")

		_, newEnv := test.CreateEnvironmentCreated(t, s.ctx, s.svc, s.ctrl, spaceID, payload)

		assert.NotNil(t, newEnv)
		assert.NotNil(t, newEnv.Data.ID)
		_, env := test.ShowEnvironmentOK(t, s.ctx, s.svc, s.ctrl, *newEnv.Data.ID)
		require.NotNil(t, env)
		assert.Equal(t, env.Data.ID, newEnv.Data.ID)
	})

	s.T().Run("unauthorized", func(t *testing.T) {
		spaceID := uuid.NewV4()
		payload := newCreateEnvironmentPayload("osio-stage", "stage", "cluster1.com")
		_, err := test.CreateEnvironmentUnauthorized(t, s.ctx2, s.svc2, s.ctrl2, spaceID, payload)
		assert.NotNil(t, err)
	})
}

func (s *EnvironmentControllerSuite) TestCreateNeg() {
	spaceID := uuid.NewV4()
	payload := newCreateEnvironmentPayload("osio-stage", "stage", "cluster1.com")
	_, err := test.CreateEnvironmentMethodNotAllowed(s.T(), s.ctx, s.svc, s.prodCtrl, spaceID, payload)
	assert.NotNil(s.T(), err)
}

func (s *EnvironmentControllerSuite) TestList() {
	s.T().Run("ok", func(t *testing.T) {
		spaceID := uuid.NewV4()
		payload := newCreateEnvironmentPayload("osio-stage", "stage", "cluster1.com")
		_, newEnv := test.CreateEnvironmentCreated(t, s.ctx, s.svc, s.ctrl, spaceID, payload)
		require.NotNil(t, newEnv)

		_, list := test.ListEnvironmentOK(t, s.ctx, s.svc, s.ctrl, spaceID)
		assert.NotNil(t, list)
		assert.NotEmpty(t, list.Data)
		assert.Equal(t, newEnv.Data.ID, list.Data[0].ID)
	})

	s.T().Run("unauthorized", func(t *testing.T) {
		spaceID := uuid.NewV4()
		_, err := test.ListEnvironmentUnauthorized(t, s.ctx2, s.svc2, s.ctrl2, spaceID)
		assert.NotNil(t, err)
	})
}

func (s *EnvironmentControllerSuite) TestShow() {
	s.T().Run("ok", func(t *testing.T) {
		spaceID := uuid.NewV4()
		payload := newCreateEnvironmentPayload("osio-stage", "stage", "cluster1.com")
		_, newEnv := test.CreateEnvironmentCreated(t, s.ctx, s.svc, s.ctrl, spaceID, payload)
		require.NotNil(t, newEnv)

		_, env := test.ShowEnvironmentOK(t, s.ctx2, s.svc2, s.ctrl2, *newEnv.Data.ID)
		assert.NotNil(t, env)
		assert.Equal(t, newEnv.Data.ID, env.Data.ID)
	})

	s.T().Run("not_found", func(t *testing.T) {
		envID := uuid.NewV4()
		_, err := test.ShowEnvironmentNotFound(t, s.ctx2, s.svc2, s.ctrl2, envID)
		assert.NotNil(t, err)
	})
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

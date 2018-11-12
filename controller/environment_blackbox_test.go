package controller_test

import (
	"context"
	"testing"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gock "gopkg.in/h2non/gock.v1"

	"github.com/fabric8-services/fabric8-auth-client/auth"
	testauth "github.com/fabric8-services/fabric8-common/test/auth"
	testsuite "github.com/fabric8-services/fabric8-common/test/suite"
	"github.com/fabric8-services/fabric8-env/app"
	"github.com/fabric8-services/fabric8-env/app/test"
	"github.com/fabric8-services/fabric8-env/client"
	"github.com/fabric8-services/fabric8-env/configuration"
	"github.com/fabric8-services/fabric8-env/controller"
	"github.com/fabric8-services/fabric8-env/gormapp"
	"github.com/goadesign/goa"
	"github.com/stretchr/testify/suite"
)

type EnvironmentControllerSuite struct {
	testsuite.DBTestSuite
	db *gormapp.GormDB

	svc  *goa.Service
	ctx  context.Context
	ctrl *controller.EnvironmentController
}

func TestEnvironmentController(t *testing.T) {
	config, err := configuration.New("")
	require.NoError(t, err)
	suite.Run(t, &EnvironmentControllerSuite{DBTestSuite: testsuite.NewDBTestSuite(config)})
}

func (s *EnvironmentControllerSuite) SetupSuite() {
	s.DBTestSuite.SetupSuite()

	s.db = gormapp.NewGormDB(s.DB)
	authClient, err := client.NewAuthClient("https://auth.prod-preview.openshift.io")
	require.NoError(s.T(), err)

	svc := testauth.UnsecuredService("enviroment-test")
	s.svc = svc
	s.ctx = s.svc.Context
	s.ctrl = controller.NewEnvironmentController(s.svc, s.db, authClient)
}

func (s *EnvironmentControllerSuite) TestCreate() {
	s.T().Run("ok", func(t *testing.T) {
		defer gock.Off()
		spaceID := newMockedUUID(2)
		payload := newCreateEnvironmentPayload("osio-stage", "stage", "cluster1.com")

		_, newEnv := test.CreateEnvironmentCreated(t, s.ctx, s.svc, s.ctrl, spaceID, payload)

		assert.NotNil(t, newEnv)
		assert.NotNil(t, newEnv.Data.ID)

		_, env := test.ShowEnvironmentOK(t, s.ctx, s.svc, s.ctrl, *newEnv.Data.ID)
		require.NotNil(t, env)
		assert.Equal(t, env.Data.ID, newEnv.Data.ID)
	})

	s.T().Run("unauthorized", func(t *testing.T) {
		defer gock.Off()
		spaceID := newMockedUUID(1, "manage")
		payload := newCreateEnvironmentPayload("osio-stage", "stage", "cluster1.com")
		_, err := test.CreateEnvironmentUnauthorized(t, s.ctx, s.svc, s.ctrl, spaceID, payload)
		assert.NotNil(t, err)
	})
}

func (s *EnvironmentControllerSuite) TestList() {
	s.T().Run("ok", func(t *testing.T) {
		defer gock.Off()
		spaceID := newMockedUUID(2)
		payload := newCreateEnvironmentPayload("osio-stage", "stage", "cluster1.com")
		_, newEnv := test.CreateEnvironmentCreated(t, s.ctx, s.svc, s.ctrl, spaceID, payload)
		require.NotNil(t, newEnv)

		_, list := test.ListEnvironmentOK(t, s.ctx, s.svc, s.ctrl, spaceID)
		assert.NotNil(t, list)
		assert.NotEmpty(t, list.Data)
		assert.Equal(t, newEnv.Data.ID, list.Data[0].ID)
	})

	s.T().Run("unauthorized", func(t *testing.T) {
		defer gock.Off()
		spaceID := newMockedUUID(1, "contribute")
		_, err := test.ListEnvironmentUnauthorized(t, s.ctx, s.svc, s.ctrl, spaceID)
		assert.NotNil(t, err)
	})
}

func (s *EnvironmentControllerSuite) TestShow() {
	s.T().Run("ok", func(t *testing.T) {
		defer gock.Off()
		spaceID := newMockedUUID(2)
		payload := newCreateEnvironmentPayload("osio-stage", "stage", "cluster1.com")
		_, newEnv := test.CreateEnvironmentCreated(t, s.ctx, s.svc, s.ctrl, spaceID, payload)
		require.NotNil(t, newEnv)

		_, env := test.ShowEnvironmentOK(t, s.ctx, s.svc, s.ctrl, *newEnv.Data.ID)
		assert.NotNil(t, env)
		assert.Equal(t, newEnv.Data.ID, env.Data.ID)
	})

	s.T().Run("unauthorized", func(t *testing.T) {
		defer gock.Off()
		spaceID := newMockedUUID(2, "contribute")
		payload := newCreateEnvironmentPayload("osio-stage", "stage", "cluster1.com")
		_, newEnv := test.CreateEnvironmentCreated(t, s.ctx, s.svc, s.ctrl, spaceID, payload)
		require.NotNil(t, newEnv)

		_, err := test.ShowEnvironmentUnauthorized(t, s.ctx, s.svc, s.ctrl, *newEnv.Data.ID)
		assert.NotNil(t, err)
	})

	s.T().Run("not_found", func(t *testing.T) {
		defer gock.Off()
		envID := newMockedUUID(0)
		_, err := test.ShowEnvironmentNotFound(t, s.ctx, s.svc, s.ctrl, envID)
		assert.NotNil(t, err)
	})
}

func newMockedUUID(mockedTimes int, removeScope ...string) uuid.UUID {
	id := uuid.NewV4()
	if mockedTimes > 0 {
		body := ""
		if len(removeScope) == 0 {
			body = `{"data":[{"id":"view","type":"user_resource_scope"},{"id":"contribute","type":"user_resource_scope"},{"id":"manage","type":"user_resource_scope"}]}`
		} else if removeScope[0] == "manage" {
			body = `{"data":[{"id":"view","type":"user_resource_scope"},{"id":"contribute","type":"user_resource_scope"}]}`
		} else if removeScope[0] == "contribute" {
			body = `{"data":[{"id":"view","type":"user_resource_scope"},{"id":"manage","type":"user_resource_scope"}]}`
		}
		gock.New("https://auth.prod-preview.openshift.io").Times(mockedTimes).
			Get(auth.ScopesResourcePath(id.String())).
			Reply(200).
			BodyString(body)
	}
	return id
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

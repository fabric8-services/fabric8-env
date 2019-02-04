package controller_test

import (
	"context"
	"testing"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	clusterclient "github.com/fabric8-services/fabric8-cluster-client/cluster"
	"github.com/fabric8-services/fabric8-common/convert/ptr"
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

	svc  *goa.Service
	ctx  context.Context
	ctrl *controller.EnvironmentController
}

type testAuthService struct{}

func (s *testAuthService) RequireScope(ctx context.Context, resourceID, requiredScope string) error {
	return nil
}

type testClusterService struct{}

func (s *testClusterService) UserClusters(ctx context.Context) (*clusterclient.ClusterList, error) {
	return &clusterclient.ClusterList{
		Data: []*clusterclient.ClusterData{
			{
				Name:   "cluster1",
				APIURL: "cluster1.com",
			},
		},
	}, nil
}

func TestEnvironmentController(t *testing.T) {
	config, err := configuration.New("")
	require.NoError(t, err)
	suite.Run(t, &EnvironmentControllerSuite{DBTestSuite: testsuite.NewDBTestSuite(config)})
}

func (s *EnvironmentControllerSuite) SetupSuite() {
	s.DBTestSuite.SetupSuite()

	s.db = gormapp.NewGormDB(s.DB)

	svc := testauth.UnsecuredService("enviroment-test")
	s.svc = svc
	s.ctx = s.svc.Context
	s.ctrl = controller.NewEnvironmentController(s.svc, s.db, &testAuthService{}, &testClusterService{})
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

	s.T().Run("cluster_not_linked", func(t *testing.T) {
		spaceID := uuid.NewV4()
		payload := newCreateEnvironmentPayload("osio-stage", "stage", "cluster2.com")

		_, err := test.CreateEnvironmentForbidden(t, s.ctx, s.svc, s.ctrl, spaceID, payload)
		assert.NotNil(t, err)
	})
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
}

func (s *EnvironmentControllerSuite) TestShow() {
	s.T().Run("ok", func(t *testing.T) {
		spaceID := uuid.NewV4()
		payload := newCreateEnvironmentPayload("osio-stage", "stage", "cluster1.com")
		_, newEnv := test.CreateEnvironmentCreated(t, s.ctx, s.svc, s.ctrl, spaceID, payload)
		require.NotNil(t, newEnv)

		_, env := test.ShowEnvironmentOK(t, s.ctx, s.svc, s.ctrl, *newEnv.Data.ID)
		assert.NotNil(t, env)
		assert.Equal(t, newEnv.Data.ID, env.Data.ID)
	})

	s.T().Run("not_found", func(t *testing.T) {
		envID := uuid.NewV4()
		_, err := test.ShowEnvironmentNotFound(t, s.ctx, s.svc, s.ctrl, envID)
		assert.NotNil(t, err)
	})
}

func (s *EnvironmentControllerSuite) TestValidate() {
	s.T().Run("ok", func(t *testing.T) {
		env := app.CreateEnvironmentPayload{
			Data: &app.Environment{
				Type: "environments",
				Attributes: &app.EnvironmentAttributes{
					Name:          "osio-stage",
					Type:          "stage",
					ClusterURL:    "cluster1.com",
					NamespaceName: ptr.String("osio-stage"),
				},
			},
		}
		err := env.Validate()
		assert.NoError(t, err)
	})

	s.T().Run("missing_name_failed", func(t *testing.T) {
		env := app.CreateEnvironmentPayload{
			Data: &app.Environment{
				Type: "environments",
				Attributes: &app.EnvironmentAttributes{
					Type:          "stage",
					ClusterURL:    "cluster1.com",
					NamespaceName: ptr.String("osio-stage"),
				},
			},
		}
		err := env.Validate()
		assert.Error(t, err)
		cause := err.(*goa.ErrorResponse)
		assert.Equal(t, 400, cause.Status)
	})

	s.T().Run("missing_type_failed", func(t *testing.T) {
		env := app.CreateEnvironmentPayload{
			Data: &app.Environment{
				Type: "environments",
				Attributes: &app.EnvironmentAttributes{
					Name:          "osio-stage",
					ClusterURL:    "cluster1.com",
					NamespaceName: ptr.String("osio-stage"),
				},
			},
		}
		err := env.Validate()
		assert.Error(t, err)
		cause := err.(*goa.ErrorResponse)
		assert.Equal(t, 400, cause.Status)
	})

	s.T().Run("missing_cluster_url_failed", func(t *testing.T) {
		env := app.CreateEnvironmentPayload{
			Data: &app.Environment{
				Type: "environments",
				Attributes: &app.EnvironmentAttributes{
					Name:          "osio-stage",
					Type:          "stage",
					NamespaceName: ptr.String("osio-stage"),
				},
			},
		}
		err := env.Validate()
		assert.Error(t, err)
		cause := err.(*goa.ErrorResponse)
		assert.Equal(t, 400, cause.Status)
	})

	s.T().Run("wrong_type_failed", func(t *testing.T) {
		env := app.CreateEnvironmentPayload{
			Data: &app.Environment{
				Type: "environments",
				Attributes: &app.EnvironmentAttributes{
					Name:          "osio-stage",
					Type:          "STAGE",
					ClusterURL:    "cluster1.com",
					NamespaceName: ptr.String("osio-stage"),
				},
			},
		}
		err := env.Validate()
		assert.Error(t, err)
		cause := err.(*goa.ErrorResponse)
		assert.Equal(t, 400, cause.Status)
	})
}

func newCreateEnvironmentPayload(name, envType, clusterURL string) *app.CreateEnvironmentPayload {
	payload := &app.CreateEnvironmentPayload{
		Data: &app.Environment{
			Attributes: &app.EnvironmentAttributes{
				Name:       name,
				Type:       envType,
				ClusterURL: clusterURL,
			},
			Type: "environments",
		},
	}
	return payload
}

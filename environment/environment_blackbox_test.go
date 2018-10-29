package environment_test

import (
	"context"
	"testing"

	testsuite "github.com/fabric8-services/fabric8-common/test/suite"
	"github.com/fabric8-services/fabric8-env/configuration"
	"github.com/fabric8-services/fabric8-env/environment"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type EnvironmentRepositorySuite struct {
	testsuite.DBTestSuite
	envRepo *environment.GormRepository
}

func TestEnvironmentRepository(t *testing.T) {
	config, err := configuration.New("")
	require.NoError(t, err)
	suite.Run(t, &EnvironmentRepositorySuite{DBTestSuite: testsuite.NewDBTestSuite(config)})
}

func (s *EnvironmentRepositorySuite) SetupSuite() {
	s.DBTestSuite.SetupSuite()

	s.envRepo = environment.NewRepository(s.DB)
}

func (s *EnvironmentRepositorySuite) TestCreateEnvironment() {
	newEnv := newEnvironment("osio-prod", "prod", "cluster1.com", uuid.NewV4())

	env, err := s.envRepo.Create(context.Background(), newEnv)

	require.NoError(s.T(), err)
	assert.NotNil(s.T(), env)
	assert.NotNil(s.T(), env.ID)
}

func (s *EnvironmentRepositorySuite) TestListEnvironment() {
	spaceID := uuid.NewV4()
	newEnv1 := newEnvironment("osio-prod", "prod", "cluster1.com", spaceID)
	env, err := s.envRepo.Create(context.Background(), newEnv1)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), env)
	require.NotNil(s.T(), env.ID)
	envID := env.ID

	newEnv2 := newEnvironment("osio-stage", "stage", "cluster2.com", uuid.NewV4())
	env, err = s.envRepo.Create(context.Background(), newEnv2)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), env)

	envs, err := s.envRepo.List(context.Background(), spaceID)

	require.NoError(s.T(), err)
	assert.NotNil(s.T(), envs)
	assert.Equal(s.T(), 1, len(envs))
	assert.Equal(s.T(), envID.String(), envs[0].ID.String())
}

func newEnvironment(name, envType, clusterURL string, spaceID uuid.UUID) *environment.Environment {
	env := &environment.Environment{
		Name:       &name,
		Type:       &envType,
		SpaceID:    &spaceID,
		ClusterURL: &clusterURL,
	}
	return env
}

package controller_test

import (
	"context"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/fabric8-services/fabric8-common/service"
	testauth "github.com/fabric8-services/fabric8-common/test/auth"
	testsuite "github.com/fabric8-services/fabric8-common/test/suite"
	"github.com/fabric8-services/fabric8-common/token"
	"github.com/fabric8-services/fabric8-env/app"
	"github.com/fabric8-services/fabric8-env/app/test"
	"github.com/fabric8-services/fabric8-env/configuration"
	"github.com/fabric8-services/fabric8-env/controller"
	"github.com/fabric8-services/fabric8-env/gormapp"
	"github.com/goadesign/goa"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

/*
Setup:

One space 	=> space1
Three users	=> user1, user2, user3

** user_space_scope matrix **
================================
user 	manage 	contribute 	view
================================
user1	yes		yes			yes
user2	no		yes			yes
user3	no		no			yes

** user operation matrix **
================================
user 	create 	list		show
================================
user1	yes		yes			yes
user2	no		yes			yes
user3	no		no			no
*/

var testUser1 = &testauth.Identity{ID: uuid.NewV4(), Email: "user1@test.com", Username: "user1"} // user1
var testUser2 = &testauth.Identity{ID: uuid.NewV4(), Email: "user2@test.com", Username: "user2"} // user2
var testUser3 = &testauth.Identity{ID: uuid.NewV4(), Email: "user3@test.com", Username: "user3"} // user3

type EnvironmentSpaceScopeSuite struct {
	testsuite.DBTestSuite
	db *gormapp.GormDB

	svc        *goa.Service
	ctrl       *controller.EnvironmentController
	authServer *httptest.Server

	ctx1      context.Context
	ctx2      context.Context
	ctx3      context.Context
	spaceID   uuid.UUID // space1
	publicKey *rsa.PublicKey
}

func TestEnvironmentSpaceScopeSuite(t *testing.T) {
	config, err := configuration.New("")
	require.NoError(t, err)
	suite.Run(t, &EnvironmentSpaceScopeSuite{DBTestSuite: testsuite.NewDBTestSuite(config)})
}

func (s *EnvironmentSpaceScopeSuite) SetupSuite() {
	s.DBTestSuite.SetupSuite()

	s.db = gormapp.NewGormDB(s.DB)
	s.authServer = s.startAuthServer()
	authURL := "http://" + s.authServer.Listener.Addr().String() + "/"
	authService, err := service.NewAuthService(authURL)
	require.NoError(s.T(), err)

	s.svc = testauth.UnsecuredService("enviroment-test")
	s.ctrl = controller.NewEnvironmentController(s.svc, s.db, authService)
	s.spaceID = uuid.NewV4()

	s.ctx1, _, err = testauth.EmbedUserTokenInContext(context.Background(), testUser1)
	require.NoError(s.T(), err)
	s.ctx2, _, err = testauth.EmbedUserTokenInContext(context.Background(), testUser2)
	require.NoError(s.T(), err)
	s.ctx3, _, err = testauth.EmbedUserTokenInContext(context.Background(), testUser3)
	require.NoError(s.T(), err)

	tokenMgr, err := token.ReadManagerFromContext(s.ctx1)
	require.NoError(s.T(), err)
	keys := tokenMgr.PublicKeys()
	require.NotNil(s.T(), keys)
	require.NotEmpty(s.T(), keys)
	s.publicKey = keys[0]
}

func (s *EnvironmentSpaceScopeSuite) TearDownSuite() {
	s.DBTestSuite.TearDownSuite()
	s.authServer.Close()
}

func (s *EnvironmentSpaceScopeSuite) TestAllScope() {
	payload := newCreateEnvironmentPayload("osio-stage", "stage", "cluster1.com")
	var newEnv *app.EnvironmentSingle

	s.T().Run("user1_create", func(t *testing.T) {
		_, newEnv = test.CreateEnvironmentCreated(t, s.ctx1, s.svc, s.ctrl, s.spaceID, payload)
		assert.NotNil(t, newEnv)
	})

	s.T().Run("user1_list", func(t *testing.T) {
		require.NotNil(t, newEnv)
		_, list := test.ListEnvironmentOK(t, s.ctx1, s.svc, s.ctrl, s.spaceID)
		assert.NotNil(t, list)
	})

	s.T().Run("user1_show", func(t *testing.T) {
		require.NotNil(t, newEnv)
		_, env := test.ShowEnvironmentOK(t, s.ctx1, s.svc, s.ctrl, *newEnv.Data.ID)
		assert.NotNil(t, env)
	})

	s.T().Run("user2_create", func(t *testing.T) {
		require.NotNil(t, newEnv)
		_, err := test.CreateEnvironmentUnauthorized(t, s.ctx2, s.svc, s.ctrl, s.spaceID, payload)
		assert.NotNil(t, err)
	})

	s.T().Run("user2_list", func(t *testing.T) {
		require.NotNil(t, newEnv)
		_, list := test.ListEnvironmentOK(t, s.ctx2, s.svc, s.ctrl, s.spaceID)
		assert.NotNil(t, list)
	})

	s.T().Run("user2_show", func(t *testing.T) {
		require.NotNil(t, newEnv)
		_, env := test.ShowEnvironmentOK(t, s.ctx2, s.svc, s.ctrl, *newEnv.Data.ID)
		assert.NotNil(t, env)
	})

	s.T().Run("user3_create", func(t *testing.T) {
		require.NotNil(t, newEnv)
		_, err := test.CreateEnvironmentUnauthorized(t, s.ctx3, s.svc, s.ctrl, s.spaceID, payload)
		assert.NotNil(t, err)
	})

	s.T().Run("user3_list", func(t *testing.T) {
		require.NotNil(t, newEnv)
		_, err := test.ListEnvironmentUnauthorized(t, s.ctx3, s.svc, s.ctrl, s.spaceID)
		assert.NotNil(t, err)
	})

	s.T().Run("user3_show", func(t *testing.T) {
		require.NotNil(t, newEnv)
		_, err := test.ShowEnvironmentUnauthorized(t, s.ctx3, s.svc, s.ctrl, *newEnv.Data.ID)
		assert.NotNil(t, err)
	})

}

func (s *EnvironmentSpaceScopeSuite) startAuthServer() *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleSpaceScopeRequest)
	return httptest.NewServer(mux)
}

func (s *EnvironmentSpaceScopeSuite) handleSpaceScopeRequest(rw http.ResponseWriter, req *http.Request) {
	authHeader := req.Header.Get("Authorization")
	tokenRaw := strings.Split(authHeader, " ")[1]
	jwtToken, err := s.toJwtToken(tokenRaw)
	require.NoError(s.T(), err)
	username, _ := jwtToken.Claims.(jwt.MapClaims)["preferred_username"].(string)

	body := ""
	switch username {
	case "user1":
		body = `{"data":[{"id":"view","type":"user_resource_scope"},{"id":"contribute","type":"user_resource_scope"},{"id":"manage","type":"user_resource_scope"}]}`
	case "user2":
		body = `{"data":[{"id":"view","type":"user_resource_scope"},{"id":"contribute","type":"user_resource_scope"}]}`
	case "user3":
		body = `{"data":[{"id":"view","type":"user_resource_scope"}]}`
	}

	rw.Write([]byte(body))
}

func (s *EnvironmentSpaceScopeSuite) toJwtToken(tokenStr string) (*jwt.Token, error) {
	jwtToken, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return s.publicKey, nil
	})
	return jwtToken, err
}

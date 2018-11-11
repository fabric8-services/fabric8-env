package controller

import (
	"github.com/fabric8-services/fabric8-common/errors"
	"github.com/fabric8-services/fabric8-common/httpsupport"
	"github.com/fabric8-services/fabric8-common/log"
	"github.com/fabric8-services/fabric8-common/token"
	"github.com/fabric8-services/fabric8-env/app"
	"github.com/fabric8-services/fabric8-env/application"
	"github.com/fabric8-services/fabric8-env/client"
	"github.com/fabric8-services/fabric8-env/environment"
	"github.com/goadesign/goa"
	errs "github.com/pkg/errors"
)

const (
	APIStringTypeEnvironment = "environments"
)

type EnvironmentController struct {
	*goa.Controller
	db         application.DB
	authClient *client.AuthClient
}

func NewEnvironmentController(service *goa.Service, db application.DB, authClient *client.AuthClient) *EnvironmentController {
	return &EnvironmentController{
		Controller: service.NewController("EnvironmentController"),
		db:         db,
		authClient: authClient,
	}
}

func ConvertEnvironment(env *environment.Environment) *app.Environment {
	respEnv := &app.Environment{
		ID:   env.ID,
		Type: APIStringTypeEnvironment,
		Attributes: &app.EnvironmentAttributes{
			Name:          env.Name,
			Type:          env.Type,
			NamespaceName: env.NamespaceName,
			ClusterURL:    env.ClusterURL,
		},
		// TODO add links, relations
	}
	return respEnv
}

func ConvertEnvironments(envs []*environment.Environment) *app.EnvironmentsList {
	res := &app.EnvironmentsList{Data: make([]*app.Environment, len(envs), len(envs))}
	for ind, env := range envs {
		res.Data[ind] = ConvertEnvironment(env)
	}
	return res
}

func (c *EnvironmentController) Create(ctx *app.CreateEnvironmentContext) error {
	tokenMgr, err := token.ReadManagerFromContext(ctx)
	if err != nil {
		return app.JSONErrorResponse(ctx, err)
	}
	_, err = tokenMgr.Locate(ctx)
	if err != nil {
		return app.JSONErrorResponse(ctx, errors.NewUnauthorizedError(err.Error()))
	}

	reqEnv := ctx.Payload.Data
	if reqEnv == nil {
		return app.JSONErrorResponse(ctx, errors.NewBadParameterError("data", nil).Expected("not nil"))
	}

	spaceID := ctx.SpaceID
	err = c.authClient.CheckSpaceScope(ctx, spaceID.String(), "manage")
	if err != nil {
		return app.JSONErrorResponse(ctx, err)
	}

	var env *environment.Environment
	err = application.Transactional(c.db, func(appl application.Application) error {
		newEnv := environment.Environment{
			Name:          reqEnv.Attributes.Name,
			Type:          reqEnv.Attributes.Type,
			SpaceID:       &spaceID,
			NamespaceName: reqEnv.Attributes.NamespaceName,
			ClusterURL:    reqEnv.Attributes.ClusterURL,
		}

		env, err = appl.Environments().Create(ctx, &newEnv)
		if err != nil {
			log.Error(ctx, map[string]interface{}{"err": err},
				"failed to create environment: %s", newEnv.Name)
			return errs.Wrapf(err, "failed to create environment: %s", newEnv.Name)
		}
		return nil
	})
	if err != nil {
		return app.JSONErrorResponse(ctx, err)
	}

	envData := ConvertEnvironment(env)
	res := &app.EnvironmentSingle{
		Data: envData,
	}
	ctx.ResponseData.Header().Set("Location", httpsupport.AbsoluteURL(&goa.RequestData{Request: ctx.Request},
		app.EnvironmentHref(res.Data.ID), nil))
	return ctx.Created(res)
}

func (c *EnvironmentController) List(ctx *app.ListEnvironmentContext) error {
	tokenMgr, err := token.ReadManagerFromContext(ctx)
	if err != nil {
		return app.JSONErrorResponse(ctx, err)
	}
	_, err = tokenMgr.Locate(ctx)
	if err != nil {
		return app.JSONErrorResponse(ctx, errors.NewUnauthorizedError(err.Error()))
	}

	spaceID := ctx.SpaceID
	err = c.authClient.CheckSpaceScope(ctx, spaceID.String(), "contribute")
	if err != nil {
		return app.JSONErrorResponse(ctx, err)
	}

	envs, err := c.db.Environments().List(ctx, spaceID)
	if err != nil {
		return app.JSONErrorResponse(ctx, err)
	}

	res := ConvertEnvironments(envs)
	return ctx.OK(res)
}

func (c *EnvironmentController) Show(ctx *app.ShowEnvironmentContext) error {
	envID := ctx.EnvID

	env, err := c.db.Environments().Load(ctx, envID)
	if err != nil {
		return app.JSONErrorResponse(ctx, err)
	}

	spaceID := env.SpaceID
	err = c.authClient.CheckSpaceScope(ctx, spaceID.String(), "contribute")
	if err != nil {
		return app.JSONErrorResponse(ctx, err)
	}

	envData := ConvertEnvironment(env)
	res := &app.EnvironmentSingle{
		Data: envData,
	}
	return ctx.OK(res)
}

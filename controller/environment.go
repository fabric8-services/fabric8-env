package controller

import (
	"github.com/fabric8-services/fabric8-common/errors"
	"github.com/fabric8-services/fabric8-env/app"
	"github.com/fabric8-services/fabric8-env/application"
	"github.com/fabric8-services/fabric8-env/environment"
	"github.com/goadesign/goa"
	errs "github.com/pkg/errors"
)

const (
	APIStringTypeEnvironment = "environments"
)

type EnvironmentController struct {
	*goa.Controller
	db application.DB
}

func NewEnvironmentController(service *goa.Service, db application.DB) *EnvironmentController {
	return &EnvironmentController{
		Controller: service.NewController("EnvironmentController"),
		db:         db,
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
			CreatedAt:     &env.CreatedAt,
			UpdatedAt:     &env.UpdatedAt,
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
	// TODO use from f8-common
	// _, err := token.ContextIdentity(ctx)
	// if err != nil {
	// 	return app.JSONErrorResponse(ctx, goa.ErrUnauthorized(err.Error()))
	// }

	reqEnv := ctx.Payload.Data
	if reqEnv == nil {
		return app.JSONErrorResponse(ctx, errors.NewBadParameterError("data", nil).Expected("not nil"))
	}

	var err error
	var env *environment.Environment
	err = application.Transactional(c.db, func(appl application.Application) error {
		newEnv := environment.Environment{
			Name:          reqEnv.Attributes.Name,
			Type:          reqEnv.Attributes.Type,
			NamespaceName: reqEnv.Attributes.NamespaceName,
			ClusterURL:    reqEnv.Attributes.ClusterURL,
		}

		env, err = appl.Environments().Create(ctx, &newEnv)
		if err != nil {
			return errs.Wrapf(err, "failed to create space: %s", newEnv.Name)
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
	// TODO use from f8-common
	// ctx.ResponseData.Header().Set("Location", httpsupport.AbsoluteURL(ctx.Request, app.SpaceHref(res.Data.ID)))
	return ctx.Created(res)
}

func (c *EnvironmentController) List(ctx *app.ListEnvironmentContext) error {
	var err error
	var envs []*environment.Environment
	err = application.Transactional(c.db, func(appl application.Application) error {
		envs, err = appl.Environments().List(ctx)
		return err
	})
	if err != nil {
		return app.JSONErrorResponse(ctx, err)
	}

	res := ConvertEnvironments(envs)
	return ctx.OK(res)
}

package controller

import (
	"github.com/fabric8-services/fabric8-env/app"
	"github.com/goadesign/goa"
)

// StatusController implements the status resource.
type StatusController struct {
	*goa.Controller
}

// NewStatusController creates a status controller.
func NewStatusController(service *goa.Service) *StatusController {
	return &StatusController{Controller: service.NewController("StatusController")}
}

// Show runs the show action.
func (c *StatusController) Show(ctx *app.ShowStatusContext) error {
	res := &app.Status{}
	res.Commit = app.Commit
	res.BuildTime = app.BuildTime
	res.StartTime = app.StartTime

	return ctx.OK(res)
}

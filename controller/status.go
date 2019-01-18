package controller

import (
	"fmt"

	"github.com/fabric8-services/fabric8-common/log"
	"github.com/fabric8-services/fabric8-env/app"
	"github.com/goadesign/goa"
	"github.com/jinzhu/gorm"
)

type statusConfig interface {
	DeveloperModeEnabled() bool
	DefaultConfigError() error
}

type StatusController struct {
	*goa.Controller
	dbChecker DBChecker
	config    statusConfig
}

func NewStatusController(service *goa.Service, dbChecker DBChecker, config statusConfig) *StatusController {
	return &StatusController{
		Controller: service.NewController("StatusController"),
		dbChecker:  dbChecker,
		config:     config,
	}
}

func (c *StatusController) Show(ctx *app.ShowStatusContext) error {
	res := &app.Status{}
	res.Commit = app.Commit
	res.BuildTime = app.BuildTime
	res.StartTime = app.StartTime

	devMode := c.config.DeveloperModeEnabled()
	if devMode {
		res.DevMode = &devMode
	}

	dbErr := c.dbChecker.Ping()
	if dbErr != nil {
		log.Error(ctx, map[string]interface{}{
			"db_error": dbErr.Error(),
		}, "database configuration error")
		res.DatabaseStatus = fmt.Sprintf("Error: %s", dbErr.Error())
	} else {
		res.DatabaseStatus = "OK"
	}

	configErr := c.config.DefaultConfigError()
	if configErr != nil {
		log.Error(ctx, map[string]interface{}{
			"config_error": configErr.Error(),
		}, "configuration error")
		res.ConfigurationStatus = fmt.Sprintf("Error: %s", configErr.Error())
	} else {
		res.ConfigurationStatus = "OK"
	}

	if dbErr != nil || (configErr != nil && !devMode) {
		return ctx.ServiceUnavailable(res)
	}

	return ctx.OK(res)
}

type DBChecker interface {
	Ping() error
}

type GormDBChecker struct {
	db *gorm.DB
}

func NewGormDBChecker(db *gorm.DB) DBChecker {
	return &GormDBChecker{
		db: db,
	}
}

func (c *GormDBChecker) Ping() error {
	_, err := c.db.DB().Exec("select 1")
	return err
}

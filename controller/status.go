package controller

import (
	"fmt"

	"github.com/fabric8-services/fabric8-common/log"
	"github.com/fabric8-services/fabric8-env/app"
	"github.com/goadesign/goa"
	"github.com/jinzhu/gorm"
)

type StatusController struct {
	*goa.Controller
	dbChecker DBChecker
}

func NewStatusController(service *goa.Service, dbChecker DBChecker) *StatusController {
	return &StatusController{
		Controller: service.NewController("StatusController"),
		dbChecker:  dbChecker,
	}
}

func (c *StatusController) Show(ctx *app.ShowStatusContext) error {
	res := &app.Status{}
	res.Commit = app.Commit
	res.BuildTime = app.BuildTime
	res.StartTime = app.StartTime

	err := c.dbChecker.Ping()
	if err != nil {
		log.Error(ctx, map[string]interface{}{
			"db_error": err.Error(),
		}, "database configuration error")
		res.DatabaseStatus = fmt.Sprintf("Error: %s", err.Error())
	} else {
		res.DatabaseStatus = "OK"
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

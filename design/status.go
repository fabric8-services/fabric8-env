package design

import (
	d "github.com/goadesign/goa/design"
	a "github.com/goadesign/goa/design/apidsl"
)

// Status defines the status of the current running instance
var Status = a.MediaType("application/vnd.status+json", func() {
	a.Description("The status of the current running instance")
	a.Attributes(func() {
		a.Attribute("commit", d.String, "Commit SHA this build is based on", func() {
			a.Example("35c4ca5aea9003f492cfa02f1211491385fe2ed7")
		})
		a.Attribute("buildTime", d.String, "The time when built", func() {
			a.Example("2018-11-02T07:11:53Z")
		})
		a.Attribute("startTime", d.String, "The time when started", func() {
			a.Example("2018-11-02T07:47:46Z")
		})
		a.Attribute("devMode", d.Boolean, "'True' if the Developer Mode is enabled")
		a.Attribute("databaseStatus", d.String, "The status of Database connection. 'OK' or an error message is displayed.")
		a.Attribute("configurationStatus", d.String, "The status of the used configuration. 'OK' or an error message if there is something wrong with the configuration used by service.")
		a.Required("commit", "buildTime", "startTime", "databaseStatus", "configurationStatus")
	})
	a.View("default", func() {
		a.Attribute("commit")
		a.Attribute("buildTime")
		a.Attribute("startTime")
		a.Attribute("devMode")
		a.Attribute("databaseStatus")
		a.Attribute("configurationStatus")
	})
})

var _ = a.Resource("status", func() {

	a.DefaultMedia(Status)
	a.BasePath("/status")

	a.Action("show", func() {
		a.Routing(
			a.GET(""),
		)
		a.Description("Show the status of the current running instance")
		a.Response(d.OK)
		a.Response(d.ServiceUnavailable, Status)
	})

})

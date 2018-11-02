package design

import (
	d "github.com/goadesign/goa/design"
	a "github.com/goadesign/goa/design/apidsl"
)

var env = a.Type("Environment", func() {
	a.Description(`JSONAPI store for data of environment.`)
	a.Attribute("type", d.String, func() {
		a.Enum("environments")
	})
	a.Attribute("id", d.UUID, "ID of environment", func() {
		a.Example("40bbdd3d-8b5d-4fd6-ac90-7236b669af04")
	})
	a.Attribute("attributes", envAttrs)
	// a.Attribute("relationships", envRelationships)
	a.Attribute("links", genericLinks)
	a.Required("type", "attributes")
})

var envAttrs = a.Type("EnvironmentAttributes", func() {
	a.Description(`JSONAPI store for all the "attributes" of environment.`)
	a.Attribute("name", d.String, "The environment name", func() {
		a.Example("myapp-stage")
	})
	a.Attribute("type", d.String, "The environment type", func() {
		a.Example("stage")
	})
	a.Attribute("namespaceName", d.String, "The namespace name", func() {
		a.Example("myapp-stage")
	})
	a.Attribute("cluster-url", d.String, "The cluster url", func() {
		a.Example("https://api.starter-us-east-2a.openshift.com")
	})
})

// var envRelationships = a.Type("EnvironmentRelations", func() {
// a.Attribute("space", relationGeneric, "Environment associated with one space")
// TODO for type
// })

var envListMeta = a.Type("EnvironmentListMeta", func() {
	a.Attribute("totalCount", d.Integer)
	a.Required("totalCount")
})

var envList = JSONList(
	"Environments", "Holds the list of environments",
	env,
	pagingLinks,
	envListMeta)

var envSingle = JSONSingle(
	"Environment", "Holds a single environment",
	env,
	nil)

var _ = a.Resource("environment", func() {

	a.Action("list", func() {
		a.Security("jwt")
		a.Routing(
			a.GET("/spaces/:spaceID/environments"),
		)
		a.Description("List environments for the given space ID.")
		a.Params(func() {
			a.Param("spaceID", d.UUID, "ID of the space")
		})
		a.Response(d.OK, envList)
		a.Response(d.BadRequest, JSONAPIErrors)
		a.Response(d.InternalServerError, JSONAPIErrors)
		a.Response(d.Unauthorized, JSONAPIErrors)
	})

	a.Action("create", func() {
		a.Security("jwt")
		a.Routing(
			a.POST("/spaces/:spaceID/environments"),
		)
		a.Description("Create environment")
		a.Params(func() {
			a.Param("spaceID", d.UUID, "ID of the space")
		})
		a.Payload(envSingle)
		a.Response(d.Created, envSingle)
		a.Response(d.BadRequest, JSONAPIErrors)
		a.Response(d.InternalServerError, JSONAPIErrors)
		a.Response(d.Unauthorized, JSONAPIErrors)
		// a.Response(d.Forbidden, JSONAPIErrors)
		a.Response(d.MethodNotAllowed, JSONAPIErrors)
	})

	a.Action("show", func() {
		a.Security("jwt")
		a.Routing(
			a.GET("/environments/:envID"),
		)
		a.Description("Retrieve environment (as JSONAPI) for the given ID.")
		a.Params(func() {
			a.Param("envID", d.UUID, "ID of the environment")
		})
		a.Response(d.OK, envSingle)
		a.Response(d.InternalServerError, JSONAPIErrors)
		a.Response(d.NotFound, JSONAPIErrors)
	})

})

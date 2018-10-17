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
	a.Attribute("relationships", envRelationships)
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
	a.Attribute("created-at", d.DateTime, "When the environment was created", func() {
		a.Example("2016-11-29T23:18:14Z")
	})
	a.Attribute("updated-at", d.DateTime, "When the environment was updated", func() {
		a.Example("2016-11-29T23:18:14Z")
	})
})

var envRelationships = a.Type("EnvironmentRelations", func() {
	a.Attribute("space", relationGeneric, "Environment associated with one space")
	// TODO for type
})

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
	a.BasePath("/environments")

	a.Action("list", func() {
		a.Security("jwt")
		a.Routing(
			a.GET(""),
		)
		a.Description("List environments.")
		// TODO required ??
		// a.Params(func() {
		// 	a.Param("page[offset]", d.String, "Paging start position")
		// 	a.Param("page[limit]", d.Integer, "Paging size")
		// })
		a.UseTrait("conditional")
		a.Response(d.OK, func() {
			a.Media(envList)
		})
		a.Response(d.NotModified)
		a.Response(d.BadRequest, JSONAPIErrors)
		a.Response(d.InternalServerError, JSONAPIErrors)
		a.Response(d.Unauthorized, JSONAPIErrors)
	})

	a.Action("create", func() {
		a.Security("jwt")
		a.Routing(
			a.POST(""),
		)
		a.Description("Create environment")
		a.Payload(envSingle)
		a.Response(d.Created, "/environments/.*", func() {
			a.Media(envSingle)
		})
		a.Response(d.BadRequest, JSONAPIErrors)
		a.Response(d.NotFound, JSONAPIErrors)
		a.Response(d.InternalServerError, JSONAPIErrors)
		a.Response(d.Unauthorized, JSONAPIErrors)
		a.Response(d.Forbidden, JSONAPIErrors)
		a.Response(d.Conflict, JSONAPIErrors)
	})

})

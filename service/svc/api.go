package svc

import (
	"github.com/gridsx/micro-conf/service/app"
	"github.com/kataras/iris/v12"
)

func RouteSvc(party iris.Party) {
	party.Use(app.RequireToken)
	party.Post("/instances", getServiceInstanceList)
}

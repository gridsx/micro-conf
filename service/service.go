package service

import (
	"github.com/gridsx/micro-conf/service/app"
	"github.com/gridsx/micro-conf/service/base"
	"github.com/gridsx/micro-conf/service/cfg"
	"github.com/gridsx/micro-conf/service/svc"
	"github.com/kataras/iris/v12"
)

func RegisterAPI(a *iris.Application) {
	base.RouteRaft(a.Party("/api/raft"))
	base.RouteStore(a.Party("/api/store"))
	svc.RouteSvc(a.Party("/api/svc"))
	app.RouteApp(a.Party("/api/app"))
	cfg.RoutConfig(a.Party("/api/cfg"))
}

func RouteInner(app *iris.Application) {
	base.RouteInner(app)
	cfg.RouteInner(app)
}

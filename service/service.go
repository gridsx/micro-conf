package service

import (
	"github.com/gridsx/micro-conf/service/app"
	"github.com/gridsx/micro-conf/service/base"
	"github.com/gridsx/micro-conf/service/cfg"
	"github.com/gridsx/micro-conf/service/conn"
	"github.com/gridsx/micro-conf/service/svc"
	"github.com/kataras/iris/v12"
)

func RegisterAPI(a *iris.Application) {
	base.RouteRaft(a.Party("/api/raft"))
	base.RouteStore(a.Party("/api/store"))
	app.RouteApp(a.Party("/api/app"))
	cfg.RoutConfig(a.Party("/api/cfg"))
}

func RouteInner(a *iris.Application) {
	base.RouteInner(a) // 需要校验来源IP是否是集群内部
	cfg.RouteInner(a)  // 需要校验来源IP是否是集群内部
	// TODO 开启token校验
	svc.RouteSvc(a.Party("/api/svc"))
	app.RouteAPI(a) // 需要token
	conn.RouteWs(a) //需要token
}

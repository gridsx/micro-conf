package base

import (
	"github.com/gridsx/micro-conf/service/app"
	"github.com/kataras/iris/v12"
)

// RouteInner 暂时不能加权限， 只能校验请求是否为集群内部IP而来
func RouteInner(app *iris.Application) {
	app.Post("/api/raft/cluster", clusterOperation)
	app.Post("/api/store/key", keyOperation)
}

func RouteRaft(party iris.Party) {
	party.Get("/info", getClusterInfo)
}

func RouteStore(party iris.Party) {
	party.Use(app.RequireAdmin)
	party.Get("/key", get)
	party.Get("/scan", scan)
}

package cfg

import (
	"github.com/gridsx/micro-conf/service/app"
	"github.com/kataras/iris/v12"
)

func RouteInner(app *iris.Application) {
	app.Post("/api/cfg/listen", acceptConfigChange)
	app.Post("/api/cfg/app", appConfig)
}

// RoutConfig /api/cfg/
func RoutConfig(party iris.Party) {
	// 管理后台的接口，则走权限控制模式
	adminParty := party.Party("/admin")
	routeAdminAPI(adminParty)
}

func routeAdminAPI(party iris.Party) {

	// namespace 相关API -------------------------------------------------
	// 获取所有namespace
	party.Get("/app/{appId:string}/group/{groupId:string}/namespaces",
		func(ctx iris.Context) { app.RequirePermission(ctx, app.Viewer) }, getNamespaces)
	// 查看某namespace 配置
	party.Get("/app/{appId:string}/group/{groupId:string}/namespace/{namespace:string}",
		func(ctx iris.Context) { app.RequirePermission(ctx, app.Viewer) }, getNamespaceContent)

	// 查看某namespace 配置
	party.Get("/app/{appId:string}/group/{groupId:string}/namespace/{namespace:string}",
		func(ctx iris.Context) { app.RequirePermission(ctx, app.Viewer) }, getNamespaceContent)

	// 新增 namespace
	party.Post("/app/{appId:string}/group/{groupId:string}/namespace",
		func(ctx iris.Context) { app.RequirePermission(ctx, app.Owner) },
		func(ctx iris.Context) { namespaceOperation(ctx, createNamespace) })
	// 删除某namespace 配置
	party.Delete("/app/{appId:string}/group/{groupId:string}/namespace",
		func(ctx iris.Context) { app.RequirePermission(ctx, app.Owner) },
		func(ctx iris.Context) { namespaceOperation(ctx, removeNamespace) })
	// namespace history
	party.Get("/app/{appId:string}/group/{group:string}/namespace/{namespace:string}/history",
		func(ctx iris.Context) { app.RequirePermission(ctx, app.Developer) }, namespaceEditHistory)
	// 修改某namespace内容
	party.Put("/app/{appId:string}/group/{group:string}/namespace/{namespace:string}",
		func(ctx iris.Context) { app.RequirePermission(ctx, app.Developer) }, changeNamespaceContent)
	// 发布某namespace功能
	party.Post("/app/{appId:string}/group/{group:string}/namespace/{namespace:string}/release",
		func(ctx iris.Context) { app.RequirePermission(ctx, app.Owner) }, releaseNamespace)
}

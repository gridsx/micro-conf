package svc

/// 注册中心的一些操作， 主要是保持心跳，上下线服务等
/// 可以上线，也可以下线， 也可以用来保持心跳
/// meta 的过期时间是state的10倍
import (
	"github.com/gridsx/micro-conf/service/app"
	"github.com/kataras/iris/v12"
	"github.com/winjeg/irisword/ret"
)

// 作为client获取某个注册的服务列表， 根据appId 和group来获取列表信息
func getServiceInstanceList(ctx iris.Context) {
	svc := new(app.ServiceInfo)
	if err := ctx.ReadJSON(svc); err != nil {
		ret.BadRequest(ctx, "service info not correct")
		return
	}
	// 参数校验
	if len(svc.App) == 0 {
		ret.BadRequest(ctx, "service info not correct")
		return
	}
	if len(svc.Group) == 0 {
		svc.Group = defaultGroup
	}
	ret.Ok(ctx, app.ServiceInfos(svc))
}

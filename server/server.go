package server

import (
	"fmt"

	"github.com/gridsx/micro-conf/service"
	"github.com/gridsx/micro-conf/service/conn"
	"github.com/gridsx/micro-conf/user"
	"github.com/kataras/iris/v12"
	"github.com/winjeg/go-commons/log"
	"github.com/winjeg/irisword/middleware"
)

func Serve(port int) {
	app := iris.New()

	servWebsocket(app)

	app.Get("/ping", func(ctx iris.Context) { ctx.Text("pong") })
	app.UseRouter(middleware.NewRequestLogger(nil))
	// 暂时不能进行权限验证的，只能验证来源IP来确定是否拦截
	service.RouteInner(app)

	// 涉及到jwt的应该要先初始化
	user.RegisterAPI(app)
	service.RegisterAPI(app)

	err := app.Listen(fmt.Sprintf(":%d", port))
	if err != nil {
		log.GetLogger(nil).Errorln("listen port error: %s\n" + err.Error())
		return
	}
}

func servWebsocket(app *iris.Application) {
	app.Get("/ws", conn.ServeWs)
}

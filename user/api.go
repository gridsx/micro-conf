package user

import (
	"encoding/json"

	"github.com/gridsx/micro-conf/config"
	"github.com/gridsx/micro-conf/user/session"
	"github.com/kataras/iris/v12"
	"github.com/winjeg/irisword/middleware"
	"github.com/winjeg/irisword/ret"
)

var cfg = config.App.JWT

func dec(d []byte) (interface{}, error) {
	m := new(session.UserInfo)
	err := json.Unmarshal(d, &m)
	return m, err
}

func RegisterAPI(app *iris.Application) {
	// 配置
	jwtCfg := cfg
	jwtCfg.Claims = login
	jwtCfg.Deserializer = dec
	jwtLogin := middleware.NewJWT(&jwtCfg)

	party := app.Party("/api/user")
	party.Post("/signup", register)
	party.Post("/login", jwtLogin)

	// 其他接口校验 session
	app.Use(middleware.JWTSession)
	party.Get("/logout", logout)
	party.Get("/info", func(ctx iris.Context) {
		userInfo := middleware.GetFromJWT(ctx)
		ret.Ok(ctx, userInfo)
		return
	})
}

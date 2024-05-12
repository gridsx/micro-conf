package app

import (
	"strings"

	"github.com/gridsx/micro-conf/api"
	"github.com/gridsx/micro-conf/config"
	"github.com/gridsx/micro-conf/user"
	"github.com/gridsx/micro-conf/user/session"
	"github.com/kataras/iris/v12"
	"github.com/winjeg/go-commons/cryptos"
	"github.com/winjeg/irisword/ret"
)

var token = cryptos.Sha1([]byte(config.App.JWT.Secret))

// RequirePermission 需要不同角色的role
func RequirePermission(ctx iris.Context, r role) {
	userinfo := session.GetUserInfo(ctx)
	if userinfo == nil {
		ret.Unauthorized(ctx)
		ctx.StopExecution()
		return
	}
	username := userinfo.Username
	if user.IsAdmin(username) {
		ctx.Next()
		return
	}
	// 获取appId和用户名
	appId := ctx.URLParam("app")
	appInfo, err := FindApp(appId)
	if err != nil || appInfo == nil {
		ret.ServerError(ctx, "no app found!")
		return
	}
	roleStr := ""
	// 权限 Owner > Developer > Viewer
	switch r {
	case Viewer:
		roleStr = appInfo.Owners + "," + appInfo.Developers + "," + appInfo.Viewers
	case Developer:
		roleStr = appInfo.Developers + "," + appInfo.Owners
	case Owner:
		roleStr = appInfo.Owners
	}

	if len(roleStr) == 0 {
		ret.Unauthorized(ctx, "require privilege of "+string(r))
		return
	}
	roles := strings.Split(roleStr, ",")
	for _, v := range roles {
		if strings.EqualFold(v, username) {
			ctx.Next()
			return
		}
	}
	ret.Unauthorized(ctx, "require privilege of "+string(r))
}

// RequireToken 从参数中拿到app，根据app获取token, 构造SecretProvider
func RequireToken(ctx iris.Context) {
	appId := ctx.URLParam("app")
	if len(appId) == 0 {
		ret.BadRequest(ctx, "param app should present")
		ctx.StopExecution()
		return
	}
	appInfo, err := FindApp(appId)
	if err != nil || appInfo == nil {
		ret.ServerError(ctx, "app info error")
		return
	}

	req := ctx.Request()
	// you can put the key somewhere in the header or url params
	r, err := api.CheckValid(req,
		// default implementation is via sql, to fetch the secrect
		api.DefaultProvider{
			AppKey:    appId,
			AppSecret: appInfo.Token,
		})
	if r {
		// verfy success, continue the request
		ctx.Next()
	} else {
		// verify fail, stop the request and return
		ret.Unauthorized(ctx, err.Error())
		ctx.StopExecution()
		return
	}
}

func RequireAdmin(ctx iris.Context) {
	authToken := ctx.GetHeader("_inner_auth")
	if strings.EqualFold(token, authToken) {
		ctx.Next()
		return
	}
	userinfo := session.GetUserInfo(ctx)
	if userinfo != nil {
		username := userinfo.Username
		if user.IsAdmin(username) {
			ctx.Next()
			return
		}
	}
	ret.Unauthorized(ctx)
	ctx.StopExecution()
	return
}

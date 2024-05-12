package user

import (
	"errors"

	"github.com/gridsx/micro-conf/user/session"
	"github.com/kataras/iris/v12"
	"github.com/winjeg/irisword/ret"
)

func register(ctx iris.Context) {
	// 已经登录的不能注册
	if session.GetUserInfo(ctx) != nil {
		ctx.Redirect("/", 301)
		return
	}
	info := new(User)
	if err := ctx.ReadJSON(info); err != nil {
		ret.BadRequest(ctx, err.Error())
		return
	}
	if _, err := info.Valid(); err != nil {
		ret.BadRequest(ctx, err.Error())
		return
	}
	if u, _ := FindUser(info.Username); u != nil {
		ret.BizError(ctx, "1002", "user already registered")
		return
	}
	if err := storeUserInfo(info); err != nil {
		ret.ServerError(ctx, err.Error())
		return
	}
	ret.Ok(ctx)
}

// 除此之外，需要更新用户上次登录时间和IP的话，需要此处处理
func login(ctx iris.Context) (interface{}, error) {
	info := new(User)
	if err := ctx.ReadJSON(info); err != nil {
		ret.BadRequest(ctx, err.Error())
		return nil, err
	}
	if ok, user := loginValid(info); ok {
		return *user, nil
	}
	return nil, errors.New("password not correct")
}

// 调用此接口后，需要刷新页面， 如果是app，则需要进行app内部做一定的处理
func logout(ctx iris.Context) {
	ctx.Values().Remove(cfg.Name)
	ctx.RemoveCookie(cfg.Name)
	ret.Ok(ctx, "logout successful, refresh your page")
	return
}

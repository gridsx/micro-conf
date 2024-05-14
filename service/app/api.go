package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/dgraph-io/badger/v4"
	"github.com/gridsx/micro-conf/user"
	"github.com/gridsx/micro-conf/user/session"
	"github.com/kataras/iris/v12"
	"github.com/winjeg/irisword/ret"
)

// RouteApp 以 /api/app 开头
// 暂时不支持删除APP， 仅有管理员可以管理

func RouteAPI(app *iris.Application) {
	app.Post("/api/app/reg", appStart)
}

func RouteApp(party iris.Party) {

	// 这段必须要放在前面，否则都需要admin权限了
	userParty := party.Party("/user")
	userParty.Get("/list", getUserApps)
	userParty.Post("/bookmark", modifyUserBookmark)
	userParty.Get("/{appId:string}", func(ctx iris.Context) { RequirePermission(ctx, Viewer) }, appDetail)

	// 角色管理
	party.Post("/{app:string}/role", func(ctx iris.Context) { RequirePermission(ctx, Owner) }, manageRole)
	// 应用启动的时候注册app到配置中心的接口， RequireToken

	party.Use(RequireAdmin)
	party.Get("/list", RequireAdmin, allAppList)                                                    // 获取APP列表
	party.Post("/new", createApp)                                                                   // 创建APP
	party.Put("/{app:string}", func(ctx iris.Context) { RequirePermission(ctx, Owner) }, modifyApp) // 修改APP信息
}

// 创建App， 目前也只有管理员有权限
func createApp(ctx iris.Context) {
	appInfo := new(AppInfo)
	if err := ctx.ReadJSON(appInfo); err != nil {
		ret.BadRequest(ctx, err.Error())
		return
	}
	userInfo := session.GetUserInfo(ctx)
	appInfo.Creator = userInfo.Username
	if err := newApp(appInfo); err != nil {
		ret.ServerError(ctx, err.Error())
		return
	}
	ret.Ok(ctx, appInfo)
}

// 仅支持修改App名称描述和部门信息
func modifyApp(ctx iris.Context) {
	appId := ctx.Params().Get("app")
	appInfo := new(AppInfo)
	if err := ctx.ReadJSON(appInfo); err != nil {
		ret.BadRequest(ctx, err.Error())
		return
	}
	appInfo.AppId = appId
	if err := modifyAppInfo(appInfo); err != nil {
		ret.ServerError(ctx, err.Error())
		return
	}
	ret.Ok(ctx)
}

type roleRequest struct {
	Action string `json:"action"`
	Role   string `json:"role"`
	User   string `json:"user"`
}

// 管理App对应的角色， 目前只有管理员有权限
func manageRole(ctx iris.Context) {
	appId := ctx.Params().Get("app")
	req := new(roleRequest)
	if err := ctx.ReadJSON(req); err != nil {
		ret.BadRequest(ctx, err.Error())
		return
	}
	if u, err := user.FindUser(req.User); u == nil || err != nil {
		ret.BadRequest(ctx, "user does not exist")
		return
	}

	switch req.Action {
	case "add":
		if err := addRole(role(req.Role), appId, req.User); err != nil {
			ret.ServerError(ctx, err.Error())
			return
		} else {
			if err := addUserApp(req.User, appId); err != nil {
				logger.Errorln("manageRole - add user error: " + err.Error())
			}
		}
	case "remove", "del":
		if err := removeRole(role(req.Role), appId, req.User); err != nil {
			ret.ServerError(ctx, err.Error())
			return
		} else {
			if err := removeUserApp(req.User, appId); err != nil {
				logger.Errorln("manageRole - remove user error: " + err.Error())
			}
		}
	default:
		ret.BadRequest(ctx, "unknown action")
	}
	ret.Ok(ctx)
}

func appStart(ctx iris.Context) {
	instInfo := new(InstanceInfo)
	if err := ctx.ReadJSON(instInfo); err != nil {
		ret.BadRequest(ctx, err.Error())
		return
	}
	if err := regApp(instInfo); err != nil {
		ret.ServerError(ctx, err.Error())
		return
	}
	ret.Ok(ctx)
}

func getUserApps(ctx iris.Context) {
	userInfo := session.GetUserInfo(ctx)
	ret.Ok(ctx, appList(userInfo.Username))
}

type bookmarkAction struct {
	AppId  string `json:"appId"`
	Action string `json:"action"`
}

func modifyUserBookmark(ctx iris.Context) {
	userInfo := session.GetUserInfo(ctx)
	if userInfo == nil {
		ret.Unauthorized(ctx)
		return
	}
	action := new(bookmarkAction)
	if err := ctx.ReadJSON(action); err != nil {
		ret.BadRequest(ctx, err.Error())
		return
	}
	switch action.Action {
	case "add":
		if err := bookmarkApp(userInfo.Username, action.AppId); err != nil {
			ret.ServerError(ctx, err.Error())
			return
		}
	case "remove", "del":
		if err := removeAppBookMark(userInfo.Username, action.AppId); err != nil {
			ret.ServerError(ctx, err.Error())
			return
		}
	default:
		ret.BadRequest(ctx, "illegal action")
		return
	}
	ret.Ok(ctx)
}

// 包括服务注册的实例列表， 以及监听namespace的IP列表
func appDetail(ctx iris.Context) {
	appId := ctx.Params().Get("appId")
	if len(appId) == 0 {
		ret.BadRequest(ctx)
		return
	}
	appKey := fmt.Sprintf(appInfoPattern, appId)
	appStr, err := rs.Get(appKey)
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			ret.BadRequest(ctx, "app does no exist")
			return
		}
		ret.ServerError(ctx, err.Error())
		return
	}
	// 获取基础信息, 包括group的名称等
	appInfo := new(AppInfo)
	if err := json.Unmarshal([]byte(appStr), appInfo); err != nil {
		ret.ServerError(ctx, err.Error())
		return
	}

	result := make(map[string]interface{}, defaultSize)

	// 获取注册的app实例列表， 有可能是空的，对应应用无任何实例
	// state 和 meta 还有 ns的
	groups := strings.Split(appInfo.Groups, ",")
	groupMap := make(map[string]interface{}, defaultSize)
	nsMap := make(map[string]interface{}, defaultSize)
	for _, g := range groups {
		infos := ServiceInfos(&ServiceInfo{App: appId, Group: g})
		groupMap[g] = infos
		nsInstances := getNamespaces(appId, g)
		nsMap[g] = nsInstances
	}
	result["app"] = appInfo
	result["groups"] = groupMap
	result["namespaces"] = nsMap
	ret.Ok(ctx, result)
}

func allAppList(ctx iris.Context) {
	appMap := make(map[string]*AppInfo, 16)
	appInfoMap := rs.ScanKvs(appInfoScanPattern)
	for k, v := range appInfoMap {
		info := new(AppInfo)
		if err := json.Unmarshal([]byte(v), info); err != nil {
			continue
		}
		appMap[k] = info
	}
	ret.Ok(ctx, appMap)
	return
}

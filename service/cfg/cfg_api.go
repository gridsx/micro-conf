package cfg

import (
	"fmt"

	"github.com/gridsx/micro-conf/service/app"
	"github.com/kataras/iris/v12"
	"github.com/winjeg/irisword/ret"
)

type NamespaceClientRequest struct {
	AppId            string   `json:"appId,omitempty"`
	Group            string   `json:"group,omitempty"`
	Namespaces       []string `json:"namespaces,omitempty"`
	SharedNamespaces []string `json:"sharedNamespaces,omitempty"`
}

// Keys 获取一个应用锁监听的所有Namespace对应的 数据KEY列表
func (r *NamespaceClientRequest) Keys() []string {
	result := make([]string, 0, defaultSize)
	for _, v := range r.Namespaces {
		namespaceKey := fmt.Sprintf(appConfigKeyPattern, r.AppId, r.Group, v)
		result = append(result, namespaceKey)
	}
	for _, v := range r.SharedNamespaces {
		if app.IsNsShared(v) {
			appId, group, ns := app.ExtractAppGroupNs(v)
			namespaceKey := fmt.Sprintf(appConfigKeyPattern, appId, group, ns)
			result = append(result, namespaceKey)
		}
	}
	return result
}

func appConfig(ctx iris.Context) {
	appId := ctx.Params().Get("appId")
	req := new(NamespaceClientRequest)
	if err := ctx.ReadJSON(req); err != nil || len(appId) == 0 {
		ret.BadRequest(ctx)
		return
	}
	req.AppId = appId
	nsWithContent := queryNamespaceContent(req.Keys())
	ret.Ok(ctx, nsWithContent)
}

func queryNamespaceContent(keys []string) map[string]string {
	result := make(map[string]string, defaultSize)
	for _, key := range keys {
		content, err := rs.Get(key)
		if err != nil {
			continue
		}
		result[key] = content
	}
	return result
}

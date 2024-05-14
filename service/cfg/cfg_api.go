package cfg

import (
	"fmt"
	"strings"

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
	req := new(NamespaceClientRequest)
	if err := ctx.ReadJSON(req); err != nil {
		ret.BadRequest(ctx)
		return
	}
	nsWithContent := queryNamespaceContent(req.Keys())
	ret.Ok(ctx, nsWithContent)
}

// 这里要求一个应用下面的namespace不能重名
func queryNamespaceContent(keys []string) map[string]map[string]string {
	result := make(map[string]map[string]string, defaultSize)
	for _, key := range keys {
		content, err := rs.Get(key)
		if err != nil {
			continue
		}
		idx := strings.LastIndex(key, ".")
		if idx < 0 {
			continue
		}

		nsIdx := strings.LastIndex(key[:idx], ".")
		if nsIdx < 0 {
			continue
		}
		namespace := key[nsIdx+1:]

		groupIdx := strings.LastIndex(key[:nsIdx], ".")
		if groupIdx < 0 {
			continue
		}
		group := key[groupIdx+1 : nsIdx]

		appIdx := strings.LastIndex(key[:groupIdx], ".")
		if appIdx < 0 {
			continue
		}
		app := key[appIdx+1 : groupIdx]
		format := key[idx+1:]
		nsKey := fmt.Sprintf("%s.%s.%s", app, group, namespace)
		switch format {
		case typeYaml:
			result[nsKey], _ = YamlToFlatMap(content)
		case typeProps:
			result[nsKey], _ = PropertiesToMap(content)
		case typeJson:
			result[nsKey], _ = JsonToFlatMap(content)
		}
	}
	return result
}

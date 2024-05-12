package cfg

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/gridsx/micro-conf/service/app"
	"github.com/gridsx/micro-conf/user/session"
	json "github.com/json-iterator/go"
	"github.com/kataras/iris/v12"
	"github.com/winjeg/irisword/ret"
)

func namespaceOperation(ctx iris.Context, f func(req *NamespaceReq) error) {
	ns := new(NamespaceReq)
	if err := ctx.ReadJSON(ns); err != nil {
		ret.BadRequest(ctx, err.Error())
		return
	}
	if err := f(ns); err != nil {
		ret.ServerError(ctx, err.Error())
		return
	}
	ret.Ok(ctx)
}

func createNamespace(ns *NamespaceReq) error {
	if err := ns.Valid(); err != nil {
		return err
	}

	nsKey := fmt.Sprintf(appConfigKeyPattern, ns.AppId, ns.Group, ns.Namespace)
	existed, err := rs.Get(nsKey)
	if err != nil && !errors.Is(err, badger.ErrKeyNotFound) {
		return err
	}
	if len(existed) > 0 {
		return errors.New("app already exists")
	}
	if errors.Is(err, badger.ErrKeyNotFound) {
		if err := rs.Set(nsKey, "", -1); err != nil {
			return err
		} else {
			return nil
		}
	}
	return errors.New("app already exists")
}

// 如果要删除一个namespace 要求这个namespace要没有人用才行
// 需要删除目前生效，待发布，以及历史版本
// 删除只是为了回收磁盘空间，因此就不必做一些强一致性的东西， 一次删除不完全可以偿失多次
func removeNamespace(ns *NamespaceReq) error {
	if err := ns.Valid(); err != nil {
		return err
	}
	if hasInstance(ns) {
		return errors.New("namespace has active peers listening")
	}
	// 历史的和现在的直接删除
	nsKey := fmt.Sprintf(appConfigKeyPattern, ns.AppId, ns.Group, ns.Namespace)
	nsReleaseKey := fmt.Sprintf(appUnreleasedKeyPattern, ns.AppId, ns.Group, ns.Namespace)

	if err := rs.Delete(nsKey); err != nil {
		logger.Errorln("removeNamespace - delete current config error")
		return err
	}
	if err := rs.Delete(nsReleaseKey); err != nil {
		logger.Errorln("removeNamespace - delete unreleased config error")
	}

	//历史版本比较多需要scan后删除
	nsHistoryPrefix := fmt.Sprintf(appHistoryScanPattern, ns.AppId, ns.Group, ns.Namespace)
	historyKeys := rs.ScanKeys(nsHistoryPrefix)
	for _, v := range historyKeys {
		if err := rs.Delete(v); err != nil {
			logger.Errorln("removeNamespace - delete history config error")
		}
	}
	return nil
}

func hasInstance(ns *NamespaceReq) bool {
	return len(app.GetNamespaceInstances(ns.AppId, ns.Group, ns.Namespace)) > 0
}

func namespaceEditHistory(ctx iris.Context) {
	appId := ctx.Params().Get("appId")
	group := ctx.Params().Get("group")
	namespace := ctx.Params().Get("namespace")
	ret.Ok(ctx, queryNamespaceHistory(appId, group, namespace))
}

func queryNamespaceHistory(appId, group, namespace string) NamespaceHistory {
	nsHistoryPrefix := fmt.Sprintf(appHistoryScanPattern, appId, group, namespace)
	namespaceMap := rs.ScanKvs(nsHistoryPrefix)
	configs := make([]*NamespaceEditHistory, 0, defaultSize)
	for _, v := range namespaceMap {
		nsh := new(NamespaceEditHistory)
		if err := json.Unmarshal([]byte(v), nsh); err != nil {
			continue
		}
		configs = append(configs, nsh)
	}
	result := NamespaceHistory(configs)
	sort.Sort(result)
	return result
}

func getNamespaces(ctx iris.Context) {
	appId := ctx.Params().Get("appId")
	group := ctx.Params().Get("group")
	scanPattern := fmt.Sprintf(appConfigKeyScanPattern, appId, group)
	keys := rs.ScanKeys(scanPattern)

	namespaces := make([]string, 0, defaultSize)
	for _, v := range keys {
		fileTypeIdx := strings.LastIndex(v, ".")
		idx := strings.LastIndex(v[:fileTypeIdx], ".")
		namespaces = append(namespaces, v[:idx])
	}
	ret.Ok(ctx, namespaces)
}

// 由于是管理后台使用，因此，需要给出的内容就不是普通内容
// 需要有几个方面的内容，包括当前配置，如果存在未发布的配置，则是未发布的配置
func getNamespaceContent(ctx iris.Context) {
	appId := ctx.Params().Get("appId")
	group := ctx.Params().Get("group")
	namespace := ctx.Params().Get("namespace")
	toRelease := fmt.Sprintf(appUnreleasedKeyPattern, appId, group, namespace)
	currentKey := fmt.Sprintf(appConfigKeyPattern, appId, group, namespace)
	toReleaseContent, err := rs.Get(toRelease)
	if err != nil && !errors.Is(err, badger.ErrKeyNotFound) {
		ret.ServerError(ctx, err.Error())
		return
	}
	// 当前的一定要有，否则就不对了
	currentContent, err := rs.Get(currentKey)
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			ret.BadRequest(ctx, "namespace does not exist")
			return
		} else {
			ret.ServerError(ctx, err.Error())
			return
		}
	}
	nsDiff, err := diff(namespace, currentContent, toReleaseContent)

	if err != nil {
		ret.ServerError(ctx, err.Error())
		return
	}
	ret.Ok(ctx, nsDiff)
}

type namespaceChangeReq struct {
	Content string `json:"content"`
}

func changeNamespaceContent(ctx iris.Context) {
	appId := ctx.Params().Get("appId")
	group := ctx.Params().Get("group")
	namespace := ctx.Params().Get("namespace")
	content := new(namespaceChangeReq)
	err := ctx.ReadJSON(content)
	if err != nil {
		ret.BadRequest(ctx, "new content is empty")
		return
	}
	currentKey := fmt.Sprintf(appConfigKeyPattern, appId, group, namespace)
	if _, err := rs.Get(currentKey); err != nil {
		ret.BadRequest(ctx, "namespace does not exist")
		return
	}
	toRelease := fmt.Sprintf(appUnreleasedKeyPattern, appId, group, namespace)
	if err := rs.Set(toRelease, content.Content, -1); err != nil {
		ret.ServerError(ctx, err.Error())
		return
	}
	ret.Ok(ctx)
}

// 1. 新增历史数据
// 2. 把toRelease 设置到 current
// 3. 推送变更
// 4. 把toRelease 删掉
func releaseNamespace(ctx iris.Context) {
	appId := ctx.Params().Get("appId")
	group := ctx.Params().Get("group")
	namespace := ctx.Params().Get("namespace")
	userInfo := session.GetUserInfo(ctx)
	// 当前内容
	currentKey := fmt.Sprintf(appConfigKeyPattern, appId, group, namespace)
	currentContent, err := rs.Get(currentKey)
	if err != nil {
		ret.BadRequest(ctx, "namespace does not exist")
		return
	}

	//  待发布内容
	toRelease := fmt.Sprintf(appUnreleasedKeyPattern, appId, group, namespace)
	toReleaseContent, err := rs.Get(toRelease)
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			ret.BadRequest(ctx, "no content to release")
			return
		}
		ret.ServerError(ctx, err.Error())
		return
	}

	// 当前与待发布进行比较
	namespaceDiff, err := diff(namespace, currentContent, toReleaseContent)
	if err != nil {
		ret.ServerError(ctx, err.Error())
		return
	}

	if namespaceDiff.Same {
		ret.Ok(ctx, "config same, nothing  changed!")
		return
	}

	// 历史数据记录
	now := time.Now()
	editHistory := NamespaceEditHistory{
		Time:       now,
		ModifiedBy: userInfo.Username,
		Content:    currentContent,
	}
	historyKey := fmt.Sprintf(appHistoryKeyPattern, appId, group, namespace, now.Format(time.RFC3339))
	d, jsonErr := json.Marshal(editHistory)
	if jsonErr != nil {
		ret.ServerError(ctx, jsonErr.Error())
		return
	}
	if err := rs.Set(historyKey, string(d), -1); err != nil {
		ret.ServerError(ctx, err.Error())
		return
	}

	if err := rs.Set(currentKey, toReleaseContent, -1); err != nil {
		ret.ServerError(ctx, err.Error())
		return
	}

	// 推送变更内容
	pushChange(appId, group, namespaceDiff)

	// 删除待发布的key
	if err := rs.Delete(toRelease); err != nil {
		ret.ServerError(ctx, err.Error())
		return
	}
	ret.Ok(ctx)
}

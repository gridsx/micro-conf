package cfg

/// 每个namespace 对应一个文件，每个文件都要有相应的文件解析器，把文件内容解释成KV形式， 主要是为了对比本版本和上个版本的差异
/// 计算差异后，可以用于展示， 也可以用于差异推送

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/gridsx/micro-conf/service/app"
	"github.com/gridsx/micro-conf/store"
	"github.com/winjeg/go-commons/log"
	"github.com/winjeg/go-commons/str"
)

const (
	defaultSize = 8

	// 单APP相关配置, app, group, namespace
	appConfigKeyPattern     = "app.cfg.current.%s.%s.%s"    // 当前版本
	appUnreleasedKeyPattern = "app.cfg.future.%s.%s.%s"     // 未发布的版本
	appHistoryKeyPattern    = "app.cfg.history.%s.%s.%s.%s" // 历史版本 每个版本会额外再添加时间
	appHistoryScanPattern   = "app.cfg.history.%s.%s.%s."
	appConfigKeyScanPattern = "app.cfg.current.%s.%s."
)

var (
	namePattern, _ = regexp.Compile("[a-zA-Z\\d\\-_]+")
	rs             = store.GetRaftStore()

	logger = log.GetLogger(nil)

	typeYaml  = "yaml"
	typeJson  = "json"
	typeProps = "props"
)

type NamespaceReq struct {
	AppId     string `json:"appId,omitempty"`
	Group     string `json:"group,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	Content   string `json:"content,omitempty"`
}

func (ns *NamespaceReq) Valid() error {
	if !namePattern.Match([]byte(ns.AppId)) || !namePattern.Match([]byte(ns.Group)) {
		return errors.New("invalid appId or group")
	}
	if !str.EndsWith(ns.Namespace, ".yaml") && !str.EndsWith(ns.Namespace, ".yml") &&
		!str.EndsWith(ns.Namespace, ".json") && !str.EndsWith(ns.Namespace, ".props") {
		return errors.New("invalid namespace name")
	}

	// 校验下app与group是否存在
	// 校验下app, group是否存在
	app, err := app.FindApp(ns.AppId)
	if app == nil || err != nil {
		return errors.New("app doesn't exist")
	}
	arr := strings.Split(app.Groups, ",")
	if !str.Contains(arr, ns.Group) {
		return errors.New("group doesn't exist")
	}
	return nil
}

type NamespaceHistory []*NamespaceEditHistory

type NamespaceEditHistory struct {
	Time       time.Time `json:"time"`
	ModifiedBy string    `json:"modifiedBy,omitempty"`
	Content    string    `json:"content,omitempty"`
}

func (n NamespaceHistory) Len() int {
	return len(n)
}

func (n NamespaceHistory) Swap(i, j int) {
	n[i], n[j] = n[j], n[i]
}

func (n NamespaceHistory) Less(i, j int) bool {
	return n[i].Time.After(n[j].Time)
}

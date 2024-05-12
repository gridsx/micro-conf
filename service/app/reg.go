package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/winjeg/go-commons/log"
	"github.com/winjeg/go-commons/str"
)

type InstanceInfo struct {
	AppId      string            `json:"appId,omitempty"`
	Group      string            `json:"group,omitempty"`
	IP         string            `json:"ip,omitempty"`
	Port       int               `json:"port,omitempty"`
	Meta       map[string]string `json:"meta,omitempty"`
	State      string            `json:"state,omitempty"` // UP/DOWN/DISABLED
	Namespaces []string          `json:"namespaces,omitempty"`
}

func (a *InstanceInfo) Valid() error {
	if len(a.AppId) == 0 {
		return errors.New("appId is empty")
	}
	if len(a.Group) == 0 {
		return errors.New("group is empty")
	}
	if !namePattern.Match([]byte(a.AppId)) || !namePattern.Match([]byte(a.Group)) {
		return errors.New("appId or group invalid")
	}
	if len(a.IP) < 7 {
		return errors.New("ip illegal")
	}
	if a.Port < 1 {
		return errors.New("port illegal")
	}
	if len(a.State) == 0 {
		return errors.New("state illegal")
	}
	// 校验下app, group是否存在
	app, err := FindApp(a.AppId)
	if app == nil || err != nil {
		return errors.New("app doesn't exist")
	}
	arr := strings.Split(app.Groups, ",")
	if !str.Contains(arr, a.Group) {
		return errors.New("group doesn't exist")
	}
	return nil
}

func (a *InstanceInfo) MetaKey() string {
	return fmt.Sprintf(MetaPattern, a.AppId, a.Group, a.IP, a.Port)
}

func (a *InstanceInfo) InstKey() string {
	return fmt.Sprintf(InstPattern, a.AppId, a.Group, a.IP, a.Port)
}
func (a *InstanceInfo) NSKeys() []string {
	result := make([]string, 0, 1)
	for _, v := range a.Namespaces {
		result = append(result, fmt.Sprintf(NamespacePattern, a.AppId, a.Group, v, a.IP, a.Port))
	}
	return result
}

func (a *InstanceInfo) MetaString() string {
	if len(a.Meta) == 0 {
		return ""
	}
	d, _ := json.Marshal(a.Meta)
	return string(d)
}

// app.instance.meta.{appId}.{group}.ip:port
// meta 有效期是1天， 每次应用重启，都用应用配置覆盖
// 注册 instance 和 namespace
func regApp(appInfo *InstanceInfo) error {
	if err := appInfo.Valid(); err != nil {
		return err
	}
	// 第一次需严格按照传递过来的执行
	metaKey := appInfo.MetaKey()
	if err := rs.Set(metaKey, appInfo.MetaString(), MetaExpire); err != nil {
		return err
	}
	nsKeys := appInfo.NSKeys()
	for _, v := range nsKeys {
		err := rs.Set(v, StateUp, instanceExpire)
		if err != nil {
			logger.Errorln("set config namespace instance error")
		}
	}
	instKey := appInfo.InstKey()
	return rs.Set(instKey, appInfo.State, instanceExpire)
}

var logger = log.GetLogger(nil)

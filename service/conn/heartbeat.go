package conn

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gridsx/micro-conf/service/app"
	"github.com/gridsx/micro-conf/store"
)

var rs = store.GetRaftStore()

// HeartBeat 客户端上报的心跳包数据  （上行）
type HeartBeat struct {
	AppId      string            `json:"appId,omitempty"`
	Group      string            `json:"group,omitempty"`
	IP         string            `json:"ip,omitempty"`
	Port       int               `json:"port,omitempty"`
	EnableSvc  bool              `json:"enableSvc"`
	EnableCfg  bool              `json:"enableCfg"`
	Meta       map[string]string `json:"meta,omitempty"`
	Namespaces []string          `json:"namespaces"` //此实例监听了哪些 配置文件， 如果是自己的话，取自己的appId和group， 如果是他人的，则取他人的
	Timeout    int64             `json:"timeout,omitempty"`
}

func (b *HeartBeat) Valid() error {
	if len(b.AppId) == 0 {
		return errors.New("invalid appId")
	}
	if len(b.Group) == 0 {
		return errors.New("invalid group")
	}
	if len(b.IP) < 7 {
		return errors.New("invalid IP")
	}
	if b.Port < 1 {
		return errors.New("invalid port")
	}
	return nil
}

func (b *HeartBeat) InstanceKey() string {
	return fmt.Sprintf(app.InstPattern, b.AppId, b.Group, b.IP, b.Port)
}

func (b *HeartBeat) MetaKey() string {
	return fmt.Sprintf(app.MetaPattern, b.AppId, b.Group, b.IP, b.Port)
}

func (b *HeartBeat) MetaString() string {
	if len(b.Meta) == 0 {
		return ""
	}
	d, _ := json.Marshal(b.Meta)
	return string(d)
}

func routeHeartbeat(info *HeartBeat) {
	// 首先把app的心跳给搞起来，凡是连接了这个app的，都安排上
	if err := info.Valid(); err != nil {
		logger.WithField("err", err.Error()).Errorln("routeHeartbeat - payload err")
		return
	}
	setAppHeartBeat(info)
	setCfgNsHeartBeat(info)
}

// 维护两个KEY： 一个是meta， 另外一个是 inst
// 本身并不需要改KEY对应的内容， 如果对应内容不存在的时候，由heartbeat 的内容进行填充
// instance key 是跟app相连的，  metaKey 则是为了维持起来用作其他功能
func setAppHeartBeat(info *HeartBeat) {
	// 拿到meta， 更新meta过期时间
	metaKey := info.MetaKey()
	metaVal, metaErr := rs.Get(metaKey)
	if metaErr != nil {
		if err := rs.Set(metaKey, info.MetaString(), app.MetaExpire); err != nil {
			logger.Errorln("setAppHeartBeat- set meta none error: " + err.Error())
		}
	} else {
		if err := rs.Set(metaKey, metaVal, app.MetaExpire); err != nil {
			logger.Errorln("setAppHeartBeat- set meta existed error: " + err.Error())
		}
	}
	// 拿到 instance， 更新instance过期时间， 如果instance没有，则默认更新为UP
	instanceKey := info.InstanceKey()
	instanceState, _ := rs.Get(instanceKey)
	// 根据 instance state 判断是否注册了服务
	if info.EnableSvc {
		if len(instanceState) == 0 {
			instanceState = app.StateUp
		}
	} else {
		instanceState = app.StateDisabled
	}
	timeout := info.Timeout
	if info.Timeout < int64(time.Second) {
		timeout = int64(time.Second * 10)
	}
	if err := rs.Set(instanceKey, instanceState, timeout); err != nil {
		logger.Errorln("setAppHeartBeat- set state error: " + err.Error())
	}
}

func setCfgNsHeartBeat(info *HeartBeat) {
	if !info.EnableCfg {
		return
	}
	timeout := info.Timeout
	if info.Timeout < int64(time.Second) {
		timeout = int64(time.Second * 10)
	}
	for _, v := range info.Namespaces {
		var nsKey string
		if app.IsNsShared(v) {
			appId, group, ns := app.ExtractAppGroupNs(v)
			nsKey = fmt.Sprintf(app.NamespacePattern, appId, group, ns, info.IP, info.Port)
		} else {
			nsKey = fmt.Sprintf(app.NamespacePattern, info.AppId, info.Group, v, info.IP, info.Port)
		}
		if err := rs.Set(nsKey, app.StateUp, timeout); err != nil {
			logger.Errorln("setCfgNsHeartBeat - set app ns state failed, err: " + err.Error())
		}
	}
}

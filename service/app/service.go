package app

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/winjeg/go-commons/log"
)

const appPartStartIdx = 18

// ServiceInfo 注册服务的信息
// 服务注册后一般会有两个key
// svc.${appId}.ip:port.${group}.state: UP   代表服务是否在线
// svc.${appId}.ip:port.${group}.meta: {}	 服务中的一些元数据信息， 比如 tag, cluster, region
type ServiceInfo struct {
	App              string                 `json:"app"`
	Group            string                 `json:"group"`
	IP               string                 `json:"ip,omitempty"`
	Port             int                    `json:"port,omitempty"`
	State            string                 `json:"state,omitempty"`
	Meta             map[string]interface{} `json:"meta,omitempty"`
	HeartbeatTimeout int                    `json:"timeout,omitempty"`
}

func ServiceInfos(svc *ServiceInfo) []ServiceInfo {
	stateMap := rs.ScanKvs(fmt.Sprintf(InstanceScanPattern, svc.App, svc.Group))
	services := make([]ServiceInfo, 0, defaultSize)
	metaMap := rs.ScanKvs(fmt.Sprintf(MetaScanPattern, svc.App, svc.Group))
	for k, v := range stateMap {
		appInfo := extractAppInfo(k)
		if appInfo == nil || len(appInfo.IP) == 0 || appInfo.Port == 0 {
			continue
		}
		mm := make(map[string]interface{}, defaultSize)
		metaKey := fmt.Sprintf(MetaPattern, svc.App, svc.Group, appInfo.IP, appInfo.Port)
		if v, ok := metaMap[metaKey]; ok {
			err := json.Unmarshal([]byte(v), &mm)
			if err != nil {
				log.GetLogger(nil).Errorf("serviceInfos - get service info err: %s\n", err.Error())
			}
		}

		services = append(services, ServiceInfo{
			App:   svc.App,
			Group: svc.Group,
			IP:    appInfo.IP,
			Port:  appInfo.Port,
			State: v,
			Meta:  mm,
		})
	}
	return services
}

// "app.instance.info.%s.%s.%s:%d"
func extractAppInfo(key string) *ServiceInfo {
	if len(key) < appPartStartIdx {
		return nil
	}
	key = key[appPartStartIdx:]
	appIdx := strings.Index(key, ".")
	if (appIdx) < 0 {
		return nil
	}
	app := key[:appIdx]
	if len(app) == 0 {
		return nil
	}
	groupIdx := strings.Index(key[appIdx+1:], ".")
	if groupIdx < 0 {
		return nil
	}
	group := key[appIdx+1 : appIdx+1+groupIdx]
	if len(group) == 0 {
		return nil
	}
	ipIdx := strings.Index(key, ":")
	if ipIdx < 0 {
		return nil
	}
	ip := key[appIdx+2+groupIdx : ipIdx]
	if len(ip) == 0 {
		return nil
	}
	portStr := key[ipIdx+1:]
	port, err := strconv.Atoi(portStr)
	if err != nil || port < 1 {
		return nil
	}
	return &ServiceInfo{
		App:   app,
		Group: group,
		IP:    ip,
		Port:  port,
	}
}

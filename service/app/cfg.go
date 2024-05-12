package app

import (
	"fmt"
	"strconv"
	"strings"
)

const ClientKeyFormat = "%s:%s:%d"

// 扫描的是待编辑的内容
// 有待发布的，则看待发布的，没有，则取current
const namespaceScanPattern = "app.cfg.future.%s.%s."
const namespaceScanCurrentPattern = "app.cfg.current.%s.%s."

type NamespaceInstance struct {
	IP        string `json:"ip,omitempty"`
	Port      int    `json:"port,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	Format    string `json:"format"`
}

type NamespaceInfo struct {
	Namespace string               `json:"namespace,omitempty"`
	Format    string               `json:"format,omitempty"`
	Instances []*NamespaceInstance `json:"instances,omitempty"`
	Content   string               `json:"content,omitempty"`
	Original  string               `json:"original,omitempty"`
}

func GetAllNamespaceInstances(appId, group string) map[string][]*NamespaceInstance {
	namespacePrefix := fmt.Sprintf(NamespaceScanPattern, appId, group)
	nsKeys := rs.ScanKeys(namespacePrefix)
	nsInfos := make([]*NamespaceInstance, 0, defaultSize)
	for _, ns := range nsKeys {
		nsInfos = append(nsInfos, extractNamespaceInfo(ns))
	}
	result := make(map[string][]*NamespaceInstance, defaultSize)
	for _, info := range nsInfos {
		if _, ok := result[info.Namespace]; ok {
			infos := result[info.Namespace]
			infos = append(infos, info)
			result[info.Namespace] = infos
		} else {
			result[info.Namespace] = []*NamespaceInstance{info}
		}
	}
	return result
}

func getNamespaces(appId, group string) []*NamespaceInfo {
	nsKeyPrefix := fmt.Sprintf(namespaceScanPattern, appId, group)
	toReleaseMap := rs.ScanKvs(nsKeyPrefix)
	currentNsKeyPrefix := fmt.Sprintf(namespaceScanCurrentPattern, appId, group)
	currentCfgNsMap := rs.ScanKvs(currentNsKeyPrefix)
	instanceMap := GetAllNamespaceInstances(appId, group)
	result := make([]*NamespaceInfo, 0, defaultSize)
	for k, v := range currentCfgNsMap {
		namespace := extractNamespace(k)
		if len(namespace) == 0 {
			continue
		}

		// 这里先拿待发布的配置，如果存在，则是待发布的配置，否则用已发布的配置
		toReleaseContent := toReleaseMap[fmt.Sprintf(namespaceScanPattern+"%s", appId, group, namespace)]
		if len(toReleaseContent) == 0 {
			toReleaseContent = v
		}
		info := &NamespaceInfo{
			Namespace: namespace,
			Format:    namespace[strings.LastIndex(namespace, ".")+1:],
			Instances: instanceMap[namespace],
			Content:   toReleaseContent,
			Original:  v,
		}
		result = append(result, info)
	}
	return result
}

func GetNamespaceInstances(appId, group, namespace string) []*NamespaceInstance {
	infos := GetAllNamespaceInstances(appId, group)
	return infos[namespace]
}

// app.ns.appId.group.namespace.format.ip:port
func extractNamespaceInfo(key string) *NamespaceInstance {
	count := 0
	idx := -1
	for i := range key {
		if key[i] == '.' {
			count++
		}
		if count == 4 {
			idx = i
			break
		}
	}
	if idx < 0 {
		return nil
	}
	remainKey := key[idx+1:]
	nsIdx := strings.Index(remainKey, ".")
	if nsIdx < 0 {
		return nil
	}
	nsFormatIdx := strings.Index(remainKey[nsIdx+1:], ".")
	if nsFormatIdx < 0 {
		return nil
	}
	namespace := remainKey[:nsIdx+nsFormatIdx+1]
	format := remainKey[nsIdx+1 : nsIdx+nsFormatIdx+1]
	arr := strings.Split(remainKey[nsIdx+nsFormatIdx+2:], ":")
	if len(arr) != 2 {
		return nil
	}
	port, _ := strconv.Atoi(arr[1])
	return &NamespaceInstance{
		IP:        arr[0],
		Port:      port,
		Format:    format,
		Namespace: namespace,
	}
}

// IsNsShared 要求所有共享的ns 必须要写清楚  appId.group.namespace.format
func IsNsShared(v string) bool {
	return len(strings.Split(v, ".")) == 4
}

func ExtractAppGroupNs(v string) (string, string, string) {
	arr := strings.Split(v, ".")
	return arr[0], arr[1], arr[2] + "." + arr[3]
}

func extractNamespace(k string) string {
	idx := strings.LastIndex(k, ".")
	if idx < 0 {
		return ""
	}
	idx2 := strings.LastIndex(k[:idx], ".")
	if idx2 < 0 {
		return ""
	}
	return k[idx2+1:]
}

package cfg

import (
	"errors"
	"strings"
)

type ConfigChangeType string

const (
	ConfigAdd    = ConfigChangeType("add")
	ConfigRemove = ConfigChangeType("remove")
	ConfigChange = ConfigChangeType("change")
)

type ConfigChangeEvent struct {
	Namespace string           `json:"namespace,omitempty"`
	Key       string           `json:"key,omitempty"`
	Type      ConfigChangeType `json:"type,omitempty"`
	Current   string           `json:"current,omitempty"`
	Before    string           `json:"before,omitempty"`
}

type NamespaceDiff struct {
	Namespace string                `json:"namespace,omitempty"`
	Same      bool                  `json:"same,omitempty"`
	Added     map[string]string     `json:"added,omitempty"`
	Removed   map[string]string     `json:"removed,omitempty"`
	Changed   map[string]StringPair `json:"changed,omitempty"`
	Unchanged map[string]string     `json:"unchanged,omitempty"`
}

func diff(namespace, old, new string) (*NamespaceDiff, error) {
	idx := strings.LastIndex(namespace, ".")
	if idx < 0 {
		return nil, errors.New("namespace illegal")
	}
	format := namespace[idx+1:]
	// 统一转换成map, 对立面的key统一对比
	var oldMap, newMap map[string]string
	var err error
	switch format {
	case typeYaml:
		oldMap, err = YamlToFlatMap(old)
		newMap, err = YamlToFlatMap(new)
	case typeJson:
		oldMap, err = JsonToFlatMap(new)
		newMap, err = JsonToFlatMap(new)
	case typeProps:
		oldMap, err = PropertiesToMap(old)
		newMap, err = PropertiesToMap(new)
	}
	if err != nil {
		return nil, err
	}
	return buildNamespaceDiff(namespace, oldMap, newMap), nil
}

func buildNamespaceDiff(namespace string, oldMap, newMap map[string]string) *NamespaceDiff {
	addedMap := make(map[string]string, len(oldMap))
	removedMap := make(map[string]string, len(newMap))
	changedMap := make(map[string]StringPair, len(oldMap))
	unchangedMap := make(map[string]string, len(oldMap))

	if len(newMap) == 0 {
		return &NamespaceDiff{
			Namespace: namespace,
			Same:      true,
			Added:     nil,
			Removed:   nil,
			Changed:   nil,
			Unchanged: oldMap,
		}
	}

	for k, v := range oldMap {
		if nv, ok := newMap[k]; ok {
			//  新老均有
			if strings.EqualFold(nv, v) {
				// 新老均一样，则是未变动的
				unchangedMap[k] = v
			} else {
				// 新老不一样，则是变动的值
				changedMap[k] = StringPair{v, nv}
			}
		} else {
			// 老有新没有， 则是删除的值
			removedMap[k] = v
		}
	}
	for k, v := range newMap {
		if _, ok := oldMap[k]; !ok {
			// 新有， 老没有， 则是新增的
			addedMap[k] = v
		}
	}
	isSame := len(addedMap) == 0 && len(removedMap) == 0 && len(changedMap) == 0
	return &NamespaceDiff{
		Namespace: namespace,
		Same:      isSame,
		Added:     addedMap,
		Removed:   removedMap,
		Changed:   changedMap,
		Unchanged: unchangedMap,
	}
}

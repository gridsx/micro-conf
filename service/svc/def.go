package svc

import (
	"github.com/gridsx/micro-conf/store"
)

const (
	// 注册中心的key pattern

	// 默认的一些耗时和group
	defaultGroup = "default"
)

var rs = store.GetRaftStore()

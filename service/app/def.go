package app

import (
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
)

const (
	appInfoPattern     = "app.info.%s"
	appInfoScanPattern = "app.info."
	MetaPattern        = "app.instance.meta.%s.%s.%s:%d"
	InstPattern        = "app.instance.info.%s.%s.%s:%d"
	NamespacePattern   = "app.ns.%s.%s.%s.%s:%d"

	InstanceScanPattern  = "app.instance.info.%s.%s."
	MetaScanPattern      = "app.instance.meta.%s.%s."
	NamespaceScanPattern = "app.ns.%s.%s."

	UserAppPattern  = "user.app.list.%s"
	UserAppBookMark = "user.app.bookmark.%s"

	StateUp       = "UP"
	StateDown     = "DOWN"
	StateDisabled = "DISABLED"

	MetaExpire     = int64(time.Hour * 24)
	instanceExpire = int64(time.Second * 30)

	defaultSize = 10
)

var namePattern, _ = regexp.Compile("[a-zA-Z\\d\\-_]+")

// AppInfo   the meta info of an app, won't change often
type AppInfo struct {
	AppId      string `json:"appId,omitempty"`
	Name       string `json:"name,omitempty"`
	Creator    string `json:"creator,omitempty"`
	CreateTime string `json:"createTime"`
	Token      string `json:"token,omitempty"`
	Department string `json:"department,omitempty"`
	Detail     string `json:"detail,omitempty"`
	Owners     string `json:"owners,omitempty"`
	Developers string `json:"developers,omitempty"`
	Viewers    string `json:"viewers,omitempty"`
	Groups     string `json:"groups,omitempty"`
}

func (i *AppInfo) Valid() error {
	if !namePattern.Match([]byte(i.AppId)) {
		return errors.New("appId illegal")
	}
	return nil
}

func (i *AppInfo) String() string {
	d, _ := json.Marshal(i)
	return string(d)
}

type EventType string

func (t *EventType) Byte() byte {
	if strings.EqualFold(string(*t), string(ConfigChange)) {
		return 1
	}
	if strings.EqualFold(string(*t), string(InfoChange)) {
		return 2
	}
	if strings.EqualFold(string(*t), string(SvcInfoChange)) {
		return 3
	}
	return 0
}

const (
	ConfigChange  = EventType("cfg")
	InfoChange    = EventType("info")
	SvcInfoChange = EventType("svc")
)

type AppEvent struct {
	Type    EventType   `json:"type,omitempty"`
	Content interface{} `json:"content,omitempty"`
}

func (e *AppEvent) String() string {
	d, _ := jsoniter.Marshal(e.Content)
	d = append(d, e.Type.Byte())
	return string(d)
}

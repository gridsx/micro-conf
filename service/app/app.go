package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/gridsx/micro-conf/store"
	"github.com/winjeg/go-commons/str"
)

type role string

const (
	tokenLen = 32

	Developer role = "developer"
	Viewer    role = "viewer"
	Owner     role = "owner"
)

var rs = store.GetRaftStore()

func newApp(info *AppInfo) error {
	if info == nil || info.Valid() != nil {
		return errors.New("app info invalid")
	}
	if app, _ := FindApp(info.AppId); app != nil {
		return errors.New("app already exist")
	}
	appToken := str.RandomNumAlphabets(tokenLen)
	info.Token = appToken
	info.CreateTime = time.Now().Format(time.RFC3339)
	info.Groups = "default"
	appKey := fmt.Sprintf(appInfoPattern, info.AppId)
	createDefaultNamespace(info.AppId, "default", "app.props")
	return rs.Set(appKey, info.String(), -1)
}

const appConfigKeyPattern = "app.cfg.current.%s.%s.%s" // 当前版本

func createDefaultNamespace(appId, group, namespace string) error {
	nsKey := fmt.Sprintf(appConfigKeyPattern, appId, group, namespace)
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
		}
	}
	return errors.New("app already exists")
}

func modifyAppInfo(info *AppInfo) error {
	if info == nil {
		return errors.New("app info nil")
	}
	app, _ := FindApp(info.AppId)
	if app == nil {
		return errors.New("app does not exist")
	}
	if len(info.Name) > 0 {
		app.Name = info.Name
	}
	if len(info.Department) > 0 {
		app.Department = info.Department
	}
	if len(info.Detail) > 0 {
		app.Detail = info.Detail
	}
	if len(info.Groups) > 0 {
		app.Groups = info.Groups
	}
	appKey := fmt.Sprintf(appInfoPattern, info.AppId)
	return rs.Set(appKey, app.String(), -1)
}

func FindApp(id string) (*AppInfo, error) {
	str, err := rs.Get(fmt.Sprintf(appInfoPattern, id))
	if err != nil {
		return nil, err
	}
	info := new(AppInfo)
	if err := json.Unmarshal([]byte(str), info); err != nil {
		return nil, err
	}
	return info, nil
}

func addRole(r role, appId, username string) error {
	if app, _ := FindApp(appId); app != nil {
		switch r {
		case Developer:
			if contains(app.Developers, username) {
				return nil
			} else {
				if len(app.Developers) == 0 {
					app.Developers = username
				} else {
					app.Developers = app.Developers + "," + username
				}
			}
		case Owner:
			if contains(app.Owners, username) {
				return nil
			} else {
				if len(app.Owners) == 0 {
					app.Owners = username
				} else {
					app.Owners = app.Owners + "," + username
				}
			}
		case Viewer:
			if contains(app.Viewers, username) {
				return nil
			} else {
				if len(app.Viewers) == 0 {
					app.Viewers = username
				} else {
					app.Viewers = app.Viewers + "," + username
				}
			}
		default:
			return errors.New("unknown role")
		}
		appKey := fmt.Sprintf(appInfoPattern, appId)
		return rs.Set(appKey, app.String(), -1)
	}
	return errors.New("app not found")
}

func removeRole(r role, appId, username string) error {
	if app, _ := FindApp(appId); app != nil {
		switch r {
		case Developer:
			if !contains(app.Developers, username) {
				return nil
			} else {
				app.Developers = removePart(app.Developers, username)
			}
		case Owner:
			if !contains(app.Owners, username) {
				return nil
			} else {
				app.Owners = removePart(app.Owners, username)
			}
		case Viewer:
			if !contains(app.Viewers, username) {
				return nil
			} else {
				app.Viewers = removePart(app.Viewers, username)
			}
		default:
			return errors.New("unknown role")
		}
		appKey := fmt.Sprintf(appInfoPattern, appId)
		return rs.Set(appKey, app.String(), -1)
	}
	return errors.New("app does not exist")
}

func removePart(whole, part string) string {
	if len(whole) == 0 || len(part) == 0 {
		return whole
	}
	arr := strings.Split(whole, ",")
	remains := make([]string, 0, len(arr))
	for _, v := range arr {
		if !strings.EqualFold(v, part) {
			remains = append(remains, v)
		}
	}
	return str.Join(remains, ",")
}

func contains(whole, part string) bool {
	if len(whole) == 0 {
		return false
	}
	arr := strings.Split(whole, ",")
	for _, v := range arr {
		if strings.EqualFold(v, part) {
			return true
		}
	}
	return false
}

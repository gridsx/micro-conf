package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/dgraph-io/badger/v4"
	"github.com/winjeg/go-commons/str"
)

// 给用户添加app， 不管权限， 有权限的app都会增加
func addUserApp(username, appId string) error {
	return addApp(username, appId, UserAppPattern)
}

// 删除用户APP， 有权限的才删除了
func removeUserApp(username, appId string) error {
	return removeApp(username, appId, UserAppPattern)
}

// 收藏App
func bookmarkApp(username, appId string) error {
	return addApp(username, appId, UserAppBookMark)
}

// 取消收藏App
func removeAppBookMark(username, appId string) error {
	return removeApp(username, appId, UserAppBookMark)
}

func addApp(username, appId, pattern string) error {
	userAppKey := fmt.Sprintf(pattern, username)
	appStr, err := rs.Get(userAppKey)
	if errors.Is(err, badger.ErrKeyNotFound) {
		if err := rs.Set(userAppKey, appId, -1); err != nil {
			return err
		}
	}
	apps := strings.Split(appStr, ",")
	for _, v := range apps {
		if strings.EqualFold(v, appId) {
			return nil
		}
	}
	v := str.TrimComma(appStr + "," + appId)
	if err := rs.Set(userAppKey, v, -1); err != nil {
		return err
	}
	return nil
}

// 删除用户APP， 有权限的才删除了
func removeApp(username, appId, pattern string) error {
	userAppKey := fmt.Sprintf(pattern, username)
	appStr, err := rs.Get(userAppKey)
	if errors.Is(err, badger.ErrKeyNotFound) {
		return nil
	}
	apps := strings.Split(appStr, ",")
	remainApps := make([]string, 0, len(apps))
	for _, v := range apps {
		if !strings.EqualFold(v, appId) {
			remainApps = append(remainApps, appId)
		}
	}
	modifiedStr := str.Join(remainApps, ",")
	if len(modifiedStr) == len(appStr) {
		return nil
	}
	return rs.Set(userAppKey, modifiedStr, -1)
}

type userApps struct {
	Apps      map[string]*AppInfo `json:"apps,omitempty"`
	Bookmarks map[string]*AppInfo `json:"bookmarks,omitempty"`
}

func appList(username string) *userApps {

	return &userApps{
		Apps:      getApps(username, UserAppPattern),
		Bookmarks: getApps(username, UserAppBookMark),
	}
}

func getApps(username, pattern string) map[string]*AppInfo {
	userAppKey := fmt.Sprintf(pattern, username)
	appStr, err := rs.Get(userAppKey)
	if errors.Is(err, badger.ErrKeyNotFound) {
		return map[string]*AppInfo{}
	}
	appIds := strings.Split(appStr, ",")
	resultMap := make(map[string]*AppInfo, len(appIds))
	for _, v := range appIds {
		appInfoStr, err := rs.Get(fmt.Sprintf(appInfoPattern, v))
		if err != nil {
			continue
		}
		appInfo := new(AppInfo)
		if err := json.Unmarshal([]byte(appInfoStr), appInfo); err != nil {
			continue
		}
		resultMap[v] = appInfo
	}
	return resultMap
}

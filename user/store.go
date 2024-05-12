package user

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gridsx/micro-conf/store"
	"github.com/gridsx/micro-conf/user/session"
	"github.com/winjeg/go-commons/cryptos"
)

const (
	userKeyPattern = "user.info.%s"
	defaultAvatar  = "https://z3.ax1x.com/2021/01/03/spXCkT.png"
)

type User = session.UserInfo

var (
	rs = store.GetRaftStore()
)

// 从存储中取出，校验sha1值是否相等， 不等则证明密码错误, 验证通过返回用户信息
func loginValid(loginInfo *User) (bool, *User) {
	if loginInfo == nil || len(loginInfo.Password) == 0 {
		return false, nil
	}
	userInfo, err := FindUser(loginInfo.Username)
	if err != nil {
		return false, nil
	}
	sha1 := cryptos.Sha1([]byte(loginInfo.Password))
	if len(sha1) == 0 {
		return false, nil
	}
	if strings.EqualFold(userInfo.Password, sha1) {
		userInfo.Password = ""
		return true, userInfo
	}
	return false, nil
}

func storeUserInfo(info *User) error {
	if ok, err := info.Valid(); !ok {
		return err
	}
	info.Password = cryptos.Sha1([]byte(info.Password))
	if len(info.Avatar) == 0 {
		info.Avatar = defaultAvatar
	}
	userInfoKey := fmt.Sprintf(userKeyPattern, info.Username)
	return rs.Set(userInfoKey, info.String(), -1)
}

func FindUser(username string) (*User, error) {
	userInfoStr, err := rs.Get(fmt.Sprintf(userKeyPattern, username))
	if err != nil {
		return nil, err
	}
	userInfo := new(User)
	if jsonErr := json.Unmarshal([]byte(userInfoStr), userInfo); jsonErr != nil {
		return nil, jsonErr
	}
	return userInfo, nil
}

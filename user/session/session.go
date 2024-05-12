package session

import (
	"encoding/json"
	"errors"
	"regexp"

	"github.com/kataras/iris/v12"
	"github.com/winjeg/irisword/middleware"
)

const (
	minPasswdLen  = 6
	defaultAvatar = "https://z3.ax1x.com/2021/01/03/spXp7V.png"
)

var (
	usernamePattern, _ = regexp.Compile("[a-zA-Z\\d_]+")
	emailPattern, _    = regexp.Compile("[a-zA-Z\\d_]+@[a-zA-Z\\d_]+")
)

type UserInfo struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Avatar   string `json:"avatar"`
	Password string `json:"password"`
}

func GetUserInfo(ctx iris.Context) *UserInfo {
	session := middleware.GetFromJWT(ctx)
	if session == nil {
		return nil
	}
	v, ok := session.(*UserInfo)
	if !ok {
		return nil
	}
	return v
}

func (u *UserInfo) Valid() (bool, error) {
	if len(u.Username) == 0 || !usernamePattern.Match([]byte(u.Username)) {
		return false, errors.New("username illegal")
	}
	if len(u.Password) == 0 || len(u.Password) < minPasswdLen {
		return false, errors.New("password illegal")
	}
	if len(u.Avatar) == 0 {
		u.Avatar = defaultAvatar
	}
	if len(u.Email) == 0 || !emailPattern.Match([]byte(u.Email)) {
		return false, errors.New("email illegal")
	}
	return true, nil
}

func (u *UserInfo) String() string {
	d, _ := json.Marshal(u)
	return string(d)
}

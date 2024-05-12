package user

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gridsx/micro-conf/config"
	"github.com/winjeg/go-commons/log"
)

const (
	adminKeyPattern = "user.admin.%s"
)

var (
	admin  = config.App.Admin
	logger = log.GetLogger(nil)
)

func initAdminUser() {
	u, _ := FindUser(admin.User)
	if u != nil {
		return
	}
	if err := storeUserInfo(&User{Username: admin.User, Password: admin.Password, Email: admin.Email}); err != nil {
		logger.WithField("error", err.Error()).Infoln("error init admin user")
		return
	} else {
		if err := addAdmin(admin.User); err != nil {
			logger.WithField("error", err.Error()).Infoln("error add user to admin")
			return
		}
	}

}

func init() {
	// TODO  优化这个逻辑， 需要等到选主完毕才能用
	go func() {
		c := time.After(time.Second * 5)
		for {
			select {
			case <-c:
				initAdminUser()
			}
		}
	}()
}

func IsAdmin(username string) bool {
	s, err := rs.Get(fmt.Sprintf(adminKeyPattern, username))
	if err != nil || !strings.EqualFold("true", s) {
		return false
	}
	return strings.EqualFold("true", s)
}

func addAdmin(username string) error {
	u, _ := FindUser(admin.User)
	if u == nil {
		return errors.New("error user not found")
	}
	return rs.Set(fmt.Sprintf(adminKeyPattern, username), "true", -1)
}

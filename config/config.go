package config

import (
	"sync"

	"github.com/winjeg/go-commons/conf"
	"github.com/winjeg/go-commons/log"
	"github.com/winjeg/irisword/middleware"
)

type (
	SeverConfig struct {
		Port int `yaml:"port" json:"port"`
	}

	RaftConfig struct {
		StoreDir string `yaml:"storeDir"`
		PeerId   string `yaml:"peerId"`
		Address  string `yaml:"address"`
	}

	AdminConfig struct {
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Email    string `yaml:"email"`
	}

	Settings struct {
		Raft    RaftConfig               `json:"raft" yaml:"raft"`
		Server  SeverConfig              `json:"server" yaml:"server"`
		Log     log.LogSettings          `json:"log" yaml:"log"`
		JWT     middleware.JWTConfig     `json:"jwt" yaml:"jwt"`
		Admin   AdminConfig              `json:"admin" yaml:"admin"`
		Monitor middleware.MonitorConfig `json:"monitor" yaml:"monitor"`
	}
)

var (
	once         sync.Once
	confFileName = "conf.yaml"
	projConf     *Settings
	App          = getSettings()
)

func initConf() {
	projConf = new(Settings)
	err := conf.Yaml2Object(confFileName, &projConf)
	if err != nil {
		panic(err)
	}
}

func getSettings() *Settings {
	if projConf != nil {
		return projConf
	} else {
		once.Do(initConf)
	}
	return projConf
}

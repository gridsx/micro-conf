package store

import (
	"fmt"

	"github.com/gridsx/micro-conf/config"
	"github.com/gridsx/micro-conf/store/raft"
)

var app = config.App

var rs = New(app.Raft.StoreDir+"/data", app.Raft.StoreDir+"/raft",
	fmt.Sprintf(app.Raft.Address+":%d", app.Server.Port+1000))

func init() {
	err := rs.Open(true, app.Raft.PeerId)

	if err != nil {
		panic(err)
	}
	fmt.Printf("init - current cluster state: %s\n", raft.NewClusterInfo(rs.Raft()))
}

func GetRaftStore() RaftStore {
	return rs
}

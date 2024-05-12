package base

/// 主要是加入或者删除集群节点，这个是非常必要的，因此，此处不可省略
/// 集群启动组建高可用服务的时候，均需要这里的接口来操作

import (
	"strings"

	"github.com/gridsx/micro-conf/config"
	"github.com/gridsx/micro-conf/store"
	"github.com/gridsx/micro-conf/store/raft"
	"github.com/kataras/iris/v12"
	"github.com/winjeg/irisword/ret"
)

var rs = store.GetRaftStore()

type RaftCmd struct {
	Cmd    string `json:"cmd,omitempty"`
	NodeId string `json:"nodeId,omitempty"`
	Addr   string `json:"addr,omitempty"`
}

func clusterOperation(ctx iris.Context) {
	cmd := new(RaftCmd)
	err := ctx.ReadJSON(cmd)
	if err != nil {
		ret.BadRequest(ctx, err.Error())
		return
	}
	if len(cmd.Cmd) == 0 || len(cmd.Addr) == 0 || len(cmd.NodeId) == 0 {
		ret.BadRequest(ctx, "wrong params")
		return
	}
	switch cmd.Cmd {
	case "remove":
		if err := rs.Remove(cmd.NodeId, cmd.Addr); err != nil {
			ret.ServerError(ctx, err.Error())
			return
		}
	case "join":
		if err := rs.Join(cmd.NodeId, cmd.Addr); err != nil {
			ret.ServerError(ctx, err.Error())
			return
		}
	default:
		ret.BadRequest(ctx, "wrong command")
		return
	}
	ret.Ok(ctx)
}

func getClusterInfo(ctx iris.Context) {
	ret.Ok(ctx, raft.NewClusterInfo(rs.Raft()))
}

func IsSelf() bool {
	info := raft.NewClusterInfo(rs.Raft())
	if info == nil {
		return false
	}
	for _, p := range info.Peers {
		if strings.EqualFold(config.App.Raft.PeerId, p.Id) {
			return true
		}
	}
	return false
}

func GetClusterInfo() *raft.ClusterState {
	return raft.NewClusterInfo(rs.Raft())
}

package base

/// 直接操作raft集群里面的底层存储， 可以从底层解决一些问题
/// 这些接口只在修复一些无法预料的bug的时候才使用， 需要对架构有清晰的认知

import (
	"github.com/gridsx/micro-conf/store/raft"
	"github.com/kataras/iris/v12"
	"github.com/winjeg/irisword/ret"
)

type KeyCmd struct {
	Cmd   string `json:"cmd,omitempty"`
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
	Exp   int64  `json:"exp,omitempty"`
}

func keyOperation(ctx iris.Context) {
	cmd := new(KeyCmd)
	err := ctx.ReadJSON(cmd)
	if err != nil {
		ret.BadRequest(ctx)
		return
	}
	if len(cmd.Key) == 0 || len(cmd.Cmd) == 0 {
		ret.BadRequest(ctx)
		return
	}

	switch cmd.Cmd {
	case raft.CmdDel:
		if err := rs.Delete(cmd.Key); err != nil {
			ret.ServerError(ctx, err.Error())
			return
		}
	case raft.CmdSet:
		if err := rs.Set(cmd.Key, cmd.Value, -1); err != nil {
			ret.ServerError(ctx, err.Error())
			return
		}
	case raft.CmdSetEx:
		if err := rs.Set(cmd.Key, cmd.Value, cmd.Exp); err != nil {
			ret.ServerError(ctx, err.Error())
			return
		}
	default:
		ret.BadRequest(ctx, "unknown cmd: "+cmd.Cmd)
		return
	}
	ret.Ok(ctx)
}

func get(ctx iris.Context) {
	key := ctx.URLParam("key")
	val, err := rs.Get(key)
	if err != nil {
		ret.ServerError(ctx, err.Error())
		return
	}
	ret.Ok(ctx, val)
}

func scan(ctx iris.Context) {
	key := ctx.URLParam("prefix")
	val := rs.ScanKvs(key)
	ret.Ok(ctx, val)
}

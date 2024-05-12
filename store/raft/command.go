package raft

type command struct {
	Op    string `json:"op,omitempty"`
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
	Exp   uint64 `json:"exp,omitempty"`
}

const (
	CmdSet   = "set"
	CmdSetEx = "setex"
	CmdDel   = "del"
)

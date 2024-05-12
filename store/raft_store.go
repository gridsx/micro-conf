package store

import (
	"github.com/hashicorp/raft"

	ra "github.com/gridsx/micro-conf/store/raft"
)

type RaftStore interface {
	Open(enableSingle bool, localID string) error

	// Get returns the value for the given key
	Get(key string) (string, error)

	// ScanKvs 根据前缀获取符合前缀的所有Key value
	ScanKvs(p string) map[string]string

	// ScanKeys 仅扫描KEY
	ScanKeys(p string) []string

	// Set sets the value for the given key, via distributed consensus
	Set(key, value string, exp int64) error

	// Delete removes the given key, via distributed consensus
	Delete(key string) error

	// Join joins the node, identified by nodeID and reachable at addr, to the cluster
	Join(nodeID string, addr string) error

	Remove(nodeID, addr string) error

	Raft() *raft.Raft
}

func New(dataDir, raftDir, raftBindAddr string) RaftStore {
	return &ra.Store{
		DataDir:      dataDir,
		RaftDir:      raftDir,
		RaftBindAddr: raftBindAddr,
	}
}

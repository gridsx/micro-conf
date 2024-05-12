package raft

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/hashicorp/raft"
	"github.com/winjeg/go-commons/log"
)

const (
	raftTimeout         = 10 * time.Second
	retainSnapshotCount = 2
	defaultSize         = 8
)

var logger = log.GetLogger(nil)

type Store struct {
	DataDir      string
	RaftDir      string
	RaftBindAddr string

	mu   sync.Mutex
	data *badger.DB
	//data map[string]string // the key-value store for the system

	raft *raft.Raft // the consensus mechanism
}

func (s *Store) Raft() *raft.Raft {
	return s.raft
}

// Open opens the store. If enableSingle is set, and there are no existing peers,
// then this node becomes the first node, and therefore leader, of the cluster.
// localID should be the server identifier for this node.
func (s *Store) Open(enableSingle bool, localID string) error {
	// Open data storage
	opts := badgerOpts(s.DataDir)
	db, err := badger.Open(opts)
	if err != nil {
		return err
	}
	s.data = db
	go runBadgerGC(db)

	// Setup Raft configuration
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(localID)
	config.Logger = NewCustomLogger(logger, localID)

	// Setup Raft communication
	addr, err := net.ResolveTCPAddr("tcp", s.RaftBindAddr)
	if err != nil {
		return err
	}
	transport, err := raft.NewTCPTransport(s.RaftBindAddr, addr, 3, 10*time.Second, os.Stderr)
	if err != nil {
		return err
	}

	// Create the snapshot store. This allows the Raft to truncate the log.
	snapshots, err := raft.NewFileSnapshotStore(s.RaftDir, retainSnapshotCount, os.Stderr)
	if err != nil {
		return fmt.Errorf("file snapshot store: %s", err)
	}

	// Create the log store and stable store
	var logStore raft.LogStore
	var stableStore raft.StableStore
	logStore, err = newBadgerStore(Options{Path: s.RaftDir + "/logs"})
	if err != nil {
		return fmt.Errorf("new badger store: %s", err)
	}
	stableStore, err = newBadgerStore(Options{Path: s.RaftDir + "/config"})
	if err != nil {
		return fmt.Errorf("new badger store: %s", err)
	}

	// Instantiate the Raft system
	ra, err := raft.NewRaft(config, (*fsm)(s), logStore, stableStore, snapshots, transport)
	if err != nil {
		return fmt.Errorf("new raft: %s", err)
	}
	s.raft = ra
	if enableSingle {
		configuration := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      config.LocalID,
					Address: transport.LocalAddr(),
				},
			},
		}
		ra.BootstrapCluster(configuration)
	}
	return nil
}

// Get returns the value for the given key.
func (s *Store) Get(key string) (string, error) {
	var value []byte
	err := s.data.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		value, err = item.ValueCopy(value)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return string(value[:]), nil
}

// ScanKvs is a prefix related pattern
func (s *Store) ScanKvs(p string) map[string]string {
	result := make(map[string]string, defaultSize)
	err := s.data.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte(p)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			err := item.Value(func(v []byte) error {
				result[string(item.Key())] = string(v)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		log.GetLogger(nil).Errorln("ScanKvs - error getting this pattern " + err.Error())
	}
	return result
}

// ScanKeys scan Keys 要快一个数量级, 可以先找到key然后在进行稀疏的值读取
// 大部分key都在内存中
func (s *Store) ScanKeys(p string) []string {
	result := make([]string, 0, defaultSize)
	prefix := []byte(p)
	err := s.data.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			result = append(result, string(it.Item().Key()))
		}
		return nil
	})
	if err != nil {
		log.GetLogger(nil).Errorln("ScanKeys - error getting this pattern " + err.Error())
	}
	return result
}

// Set sets the value for the given key.
func (s *Store) Set(key, value string, exp int64) error {
	cmd := CmdSet
	if exp > 0 {
		cmd = CmdSetEx
	}
	if s.raft.State() != raft.Leader {
		return RedirectKeyRequest(s.raft, cmd, key, value, exp)
	}
	op := CmdSet
	if exp > 0 {
		op = CmdSetEx
	}
	c := &command{
		Op:    op,
		Key:   key,
		Value: value,
		Exp:   uint64(time.Now().UnixMilli() + exp/int64(time.Millisecond)),
	}
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}

	f := s.raft.Apply(b, raftTimeout)
	return f.Error()
}

// Delete deletes the given key.
func (s *Store) Delete(key string) error {
	cmd := CmdDel
	if s.raft.State() != raft.Leader {
		return RedirectKeyRequest(s.raft, cmd, key, "", 0)
	}
	c := &command{
		Op:  "delete",
		Key: key,
	}
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}

	f := s.raft.Apply(b, raftTimeout)
	return f.Error()
}

// Join joins a node, identified by nodeID and located at addr, to this store.
// The node must be ready to respond to Raft communications at that address.
func (s *Store) Join(nodeID, addr string) error {
	logger.Printf("received join request for remote node %s at %s", nodeID, addr)

	configFuture := s.raft.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		logger.Printf("failed to get raft configuration: %v", err)
		return err
	}
	for _, srv := range configFuture.Configuration().Servers {
		// If a node already exists with either the joining node's ID or address,
		// that node may need to be removed from the config first.
		if srv.ID == raft.ServerID(nodeID) || srv.Address == raft.ServerAddress(addr) {
			// However if *both* the ID and the address are the same, then nothing -- not even
			// a join operation -- is needed.
			logger.Printf("node %s at %s already member of cluster, ignoring join request", nodeID, addr)
			return nil
			//if srv.Address == raft.ServerAddress(addr) && srv.ID == raft.ServerID(nodeID) {
			//
			//}
			//future := s.raft.RemoveServer(srv.ID, 0, 0)
			//if err := future.Error(); err != nil {
			//	return fmt.Errorf("error removing existing node %s at %s: %s", nodeID, addr, err)
			//}
		}
	}
	f := s.raft.AddVoter(raft.ServerID(nodeID), raft.ServerAddress(addr), 0, 0)
	if f.Error() != nil {
		return f.Error()
	}
	logger.Printf("node %s at %s joined successfully", nodeID, addr)
	return nil
}

func (s *Store) Remove(nodeId, addr string) error {
	if s.raft.State() != raft.Leader {
		return RedirectRaftRequest(s.raft, nodeId, addr)
	}
	logger.Printf("received remove request for remote node %s at %s", nodeId, addr)
	configFuture := s.raft.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		logger.Printf("failed to get raft configuration: %v", err)
		return err
	}
	for _, srv := range configFuture.Configuration().Servers {
		// If a node already exists with either the joining node's ID or address,
		// that node may need to be removed from the config first.
		if srv.ID == raft.ServerID(nodeId) || srv.Address == raft.ServerAddress(addr) {
			future := s.raft.RemoveServer(srv.ID, 0, 0)
			if err := future.Error(); err != nil {
				return fmt.Errorf("error removing existing node %s at %s: %s", nodeId, addr, err)
			}
		}
	}
	logger.Printf("node %s at %s removed successfully", nodeId, addr)
	return nil
}

func badgerOpts(path string) badger.Options {
	opts := badger.DefaultOptions(path)
	opts = opts.WithLogger(logger)
	opts = opts.WithMetricsEnabled(true)
	opts = opts.WithNumMemtables(4)
	opts = opts.WithMemTableSize(32 << 20)
	opts = opts.WithNumLevelZeroTables(5)
	opts = opts.WithNumLevelZeroTablesStall(8)
	opts = opts.WithNumCompactors(4)
	opts = opts.WithValueLogFileSize(128 << 20)
	return opts
}

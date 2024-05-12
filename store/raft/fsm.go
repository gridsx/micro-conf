package raft

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"runtime"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/hashicorp/raft"
)

// 有限状态机
type fsm Store // TODO extract to db.go with the database => balloon

type fsmGenericResponse struct {
	error error
}

// Apply applies a Raft log entry to the key-value store.
func (f *fsm) Apply(l *raft.Log) interface{} {
	var c command
	if err := json.Unmarshal(l.Data, &c); err != nil {
		panic(fmt.Sprintf("failed to unmarshal command: %s", err.Error()))
	}
	switch c.Op {
	case CmdSet:
		return f.applySet(c.Key, c.Value)
	case CmdSetEx:
		return f.applySet(c.Key, c.Value, c.Exp)
	case CmdDel:
		return f.applyDelete(c.Key)
	default:
		return &fsmGenericResponse{error: errors.New("unknown command")}
	}
}

// Snapshot returns a snapshot of the key-value store.
func (f *fsm) Snapshot() (raft.FSMSnapshot, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return &fsmSnapshot{store: f.data}, nil
}

// Restore stores the key-value store to a previous state.
func (f *fsm) Restore(rc io.ReadCloser) error {
	// Set the state from the snapshot, no lock required according to
	// Hashicorp docs.
	err := f.data.Load(rc, runtime.NumCPU()*2)
	if err != nil {
		return err
	}
	return nil
}

func (f *fsm) applySet(key, value string, exp ...uint64) interface{} {
	f.mu.Lock()
	defer f.mu.Unlock()

	// 如果要设置的key已经过期，那么就不再设置了
	if len(exp) > 0 && time.Now().UnixMilli() > int64(exp[0]) {
		return &fsmGenericResponse{error: nil}
	}

	err := f.data.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry([]byte(key), []byte(value))
		if len(exp) > 0 {
			millis := (exp[0] - uint64(time.Now().UnixMilli())) * uint64(time.Millisecond)
			e = e.WithTTL(time.Duration(millis))
		}
		return txn.SetEntry(e)
	})
	return &fsmGenericResponse{error: err}
}

func (f *fsm) applyDelete(key string) interface{} {
	f.mu.Lock()
	defer f.mu.Unlock()
	err := f.data.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
	return &fsmGenericResponse{error: err}
}

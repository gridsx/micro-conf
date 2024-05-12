package raft

import (
	"github.com/dgraph-io/badger/v4"
	"github.com/hashicorp/raft"
)

type fsmSnapshot struct {
	store *badger.DB
}

func (f *fsmSnapshot) Persist(sink raft.SnapshotSink) error {
	err := func() error {
		_, err := f.store.Backup(sink, 0)
		if err == nil {
			return err
		}
		return sink.Close()
	}()

	if err != nil {
		cancelErr := sink.Cancel()
		if cancelErr != nil {
			return err
		}
	}
	return err
}

func (f *fsmSnapshot) Release() {}

package raft

import (
	"errors"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/hashicorp/raft"
)

// 主要用作raft的 snapshot store 和 stable store

var (
	// ErrKeyNotFound is an error indicating a given key does not exist
	ErrKeyNotFound = errors.New("not found")
)

type BadgerStore struct {
	// conn is the underlying handle to the db.
	conn *badger.DB

	// The path to the Badger database directory.
	path string
}

// Options contains all the configuration used to open the Badger db
type Options struct {
	// Path is the directory path to the Badger db to use.
	Path string

	// BadgerOptions contains any specific Badger options you might
	// want to specify.
	BadgerOptions *badger.Options

	// NoSync causes the database to skip fsync calls after each
	// write to the log. This is unsafe, so it should be used
	// with caution.
	NoSync bool
}

// init a badger store instance
func newBadgerStore(options Options) (*BadgerStore, error) {
	if options.BadgerOptions == nil {
		defaultOpts := badgerOpts(options.Path)
		options.BadgerOptions = &defaultOpts
		options.BadgerOptions.SyncWrites = !options.NoSync
	}
	db, err := badger.Open(*options.BadgerOptions)
	if err != nil {
		return nil, err
	}
	go runBadgerGC(db)
	// Create the new store
	store := &BadgerStore{
		conn: db,
		path: options.Path,
	}
	return store, nil
}

// Close is used to gracefully close the DB connection.
func (b *BadgerStore) Close() error {
	return b.conn.Close()
}

// FirstIndex returns the first known index from the Raft log.
func (b *BadgerStore) FirstIndex() (uint64, error) {
	return b.firstIndex(false)
}

// LastIndex returns the last known index from the Raft log.
func (b *BadgerStore) LastIndex() (uint64, error) {
	return b.firstIndex(true)
}

func (b *BadgerStore) firstIndex(reverse bool) (uint64, error) {
	var value uint64
	err := b.conn.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.IteratorOptions{
			PrefetchValues: false,
			Reverse:        reverse,
		})
		defer it.Close()
		it.Rewind()
		if it.Valid() {
			value = bytesToUint64(it.Item().Key())
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return value, nil
}

// GetLog gets a log entry from Badger at a given index.
func (b *BadgerStore) GetLog(index uint64, log *raft.Log) error {
	err := b.conn.View(func(txn *badger.Txn) error {
		item, err := txn.Get(uint64ToBytes(index))
		if err != nil {
			switch {
			case errors.Is(err, badger.ErrKeyNotFound):
				return raft.ErrLogNotFound
			default:
				return err
			}
		}
		var val []byte
		vErr := item.Value(func(v []byte) error {
			val = v
			return nil
		})
		if vErr != nil {
			return vErr
		}
		return decodeMsgPack(val, log)
	})
	if err != nil {
		return err
	}
	return nil
}

// StoreLog stores a single raft log.
func (b *BadgerStore) StoreLog(log *raft.Log) error {
	return b.StoreLogs([]*raft.Log{log})
}

// StoreLogs stores a set of raft logs.
func (b *BadgerStore) StoreLogs(logs []*raft.Log) error {
	err := b.conn.Update(func(txn *badger.Txn) error {
		for _, log := range logs {
			key := uint64ToBytes(log.Index)
			val, err := encodeMsgPack(log)
			if err != nil {
				return err
			}
			if err := txn.Set(key, val.Bytes()); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// DeleteRange deletes logs within a given range inclusively.
func (b *BadgerStore) DeleteRange(min, max uint64) error {
	// we manage the transaction manually in order to avoid ErrTxnTooBig errors
	txn := b.conn.NewTransaction(true)
	it := txn.NewIterator(badger.IteratorOptions{
		PrefetchValues: false,
		Reverse:        false,
	})
	start := uint64ToBytes(min)
	for it.Seek(start); it.Valid(); it.Next() {
		key := make([]byte, 8)
		it.Item().KeyCopy(key)
		// Handle out-of-range log index
		if bytesToUint64(key) > max {
			break
		}
		// Delete in-range log index, 如果某个key造成txn过大，那么剩余的则从后面的继续
		if err := txn.Delete(key); err != nil {
			if errors.Is(err, badger.ErrTxnTooBig) {
				it.Close()
				err = txn.Commit()
				if err != nil {
					return err
				}
				return b.DeleteRange(bytesToUint64(key), max)
			}
			return err
		}
	}
	it.Close()
	err := txn.Commit()
	if err != nil {
		return err
	}
	return nil
}

// Set is used to set a key/value set outside the raft log.
func (b *BadgerStore) Set(key []byte, val []byte) error {
	return b.conn.Update(func(txn *badger.Txn) error {
		return txn.SetEntry(badger.NewEntry(key, val))
	})
}

// Get is used to retrieve a value from the k/v store by key
func (b *BadgerStore) Get(key []byte) ([]byte, error) {
	var value []byte
	err := b.conn.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			switch {
			case errors.Is(err, badger.ErrKeyNotFound):
				return ErrKeyNotFound
			default:
				return err
			}
		}
		value, err = item.ValueCopy(value)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return value, nil
}

// SetUint64 is like Set, but handles uint64 values
func (b *BadgerStore) SetUint64(key []byte, val uint64) error {
	return b.Set(key, uint64ToBytes(val))
}

// GetUint64 is like Get, but handles uint64 values
func (b *BadgerStore) GetUint64(key []byte) (uint64, error) {
	val, err := b.Get(key)
	if err != nil {
		return 0, err
	}
	return bytesToUint64(val), nil
}

// Badger 依赖客户端在他们选择的时间执行垃圾收集, 所以每个badger都要进行GC使得磁盘不会越来越大
// 每次GC都会造成一次LSM tree的突刺，因此得考虑下GC的时候耗时和吞吐量怎么样
func runBadgerGC(db *badger.DB) {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
	again:
		err := db.RunValueLogGC(0.6)
		if err == nil {
			goto again
		}
		// 已经close了
		if err == badger.ErrRejected {
			ticker.Stop()
		}
	}
}

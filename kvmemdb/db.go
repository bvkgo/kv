// Copyright (c) 2023 BVK Chaitanya

package kvmemdb

import (
	"context"
	"sync"

	"github.com/bvkgo/kv"
	"github.com/bvkgo/kv/kvmemdb/internal/multival"
	"github.com/bvkgo/kv/kvmemdb/internal/syncmap"
)

type DB struct {
	mu sync.Mutex

	// pins holds a commit version and total number of snapshot and transaction
	// references to it.
	pins map[int64]int

	// maxCommitVersion holds the last committed transaction version.
	maxCommitVersion int64

	// lastTxVersion holds the most recent tx version.
	lastTxVersion int64

	// store holds the key-value data for multiple committed versions. Each value
	// can hold data for multiple versions cause snapshots may need access to
	// older data, while newer transactions have updated the DB values. Note that
	// store never has dirty (uncommitted) values.
	store syncmap.Map[string, *multival.MultiValue]
}

func New() *DB {
	return &DB{
		pins: make(map[int64]int),
	}
}

func (db *DB) keys(skip map[string]*multival.Value) []string {
	var keys []string
	db.store.Range(func(key string, _ *multival.MultiValue) bool {
		if skip != nil {
			if _, ok := skip[key]; ok {
				return true
			}
		}
		keys = append(keys, key)
		return true
	})
	return keys
}

func (db *DB) WithSnapshot(ctx context.Context, f func(context.Context, kv.Snapshot) error) error {
	snap := db.NewSnapshot()
	defer kv.Close(snap)

	if err := f(ctx, snap); err != nil {
		return err
	}
	return nil
}

func (db *DB) WithTransaction(ctx context.Context, f func(context.Context, kv.Transaction) error) error {
	tx := db.NewTransaction()
	defer tx.Rollback(ctx)

	if err := f(ctx, tx); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

func (db *DB) WithReadOnlyTransaction(ctx context.Context, f func(context.Context, kv.ReadOnlyTransaction) error) error {
	tx := db.NewTransaction()
	defer tx.Rollback(ctx)

	if err := f(ctx, tx); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

func (db *DB) NewSnapshot() *Snapshot {
	db.mu.Lock()
	defer db.mu.Unlock()

	s := &Snapshot{
		db:                db,
		lastCommitVersion: db.maxCommitVersion,
	}

	db.pins[db.maxCommitVersion]++
	return s
}

func (db *DB) NewTransaction() *Transaction {
	db.mu.Lock()
	defer db.mu.Unlock()

	version := db.lastTxVersion + 1
	db.lastTxVersion++

	t := &Transaction{
		db:                db,
		lastCommitVersion: db.maxCommitVersion,
		version:           version,
		accesses:          make(map[string]*multival.Value),
	}

	db.pins[db.maxCommitVersion]++
	return t
}

func (db *DB) releasePin(version int64) {
	db.mu.Lock()
	defer db.mu.Unlock()

	// Release the pin on the committed version, so that it can be discarded
	// later.

	n := db.pins[version]
	if n == 1 {
		delete(db.pins, version)
	} else {
		db.pins[version] = n - 1
	}
}

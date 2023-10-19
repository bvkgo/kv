// Copyright (c) 2023 BVK Chaitanya

package kvmemdb

import (
	"context"
	"fmt"
	"math"
	"os"

	"github.com/bvkgo/kv/kvmemdb/internal/multival"
)

func (db *DB) commit(tx *Transaction) error {
	if tx.db != db {
		return os.ErrInvalid
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	minVersion := db.maxCommitVersion
	for k := range db.pins {
		if k < minVersion {
			minVersion = k
		}
	}

	// Check that all items accessed are unmodified in the database.
	for key, txval := range tx.accesses {
		if mv, ok := db.store.Load(key); ok {
			curval, cok := mv.Fetch(math.MaxInt64)
			begval, bok := mv.Fetch(tx.lastCommitVersion)
			// log.Printf("precommit %v key %s min-ver %d max-ver %d last-ver %d tx-ver %d curval %v begval %v txval %v", tx, key, minVersion, db.maxCommitVersion, tx.lastCommitVersion, tx.version, curval, begval, txval)

			if !bok && !cok {
				continue // new key solely by this tx
			}
			if !bok && cok {
				return fmt.Errorf("precommit: %v key %q is also created by another tx", tx, key)
			}
			if bok && !cok {
				return fmt.Errorf("precommit: %v key %q is deleted by another tx", tx, key)
			}
			if curval.Version != begval.Version {
				return fmt.Errorf("precommit: %v key %q is updated by tx-%d after this tx-%d accessed version %d", tx, key, curval.Version, txval.Version, begval.Version)
			}
		}
	}

	newCommitVersion := db.maxCommitVersion + 1

	for key, value := range tx.accesses {
		if value.Version == tx.version {
			mv, ok := db.store.Load(key)
			if !ok {
				mv = new(multival.MultiValue)
				if _, loaded := db.store.LoadOrStore(key, mv); loaded {
					panic("unexpected: load-or-store failed")
				}
			}

			value.Version = newCommitVersion
			newmv := multival.Append(mv, value)
			newmv = multival.Compact(newmv, minVersion)

			if newmv.Empty() {
				// log.Printf("commit %v key %s oldmv %v newmv nil", tx, key, mv)
				if !db.store.CompareAndDelete(key, mv) {
					panic("compare-and-delete")
				}
			} else {
				// log.Printf("commit %v key %s oldmv %v newmv %v", tx, key, mv, newmv)
				if !db.store.CompareAndSwap(key, mv, newmv) {
					panic("compare-and-swap")
				}
			}
		}
	}

	db.maxCommitVersion = newCommitVersion
	return nil
}

// Compact removes unnecessary values from the database. Returns number of
// keys compacted.
func Compact(ctx context.Context, db *DB) int {
	db.mu.Lock()
	minVersion := db.maxCommitVersion
	for k := range db.pins {
		if k < minVersion {
			minVersion = k
		}
	}
	db.mu.Unlock()

	tx := db.NewTransaction()
	defer tx.Close()

	count := 0
	compact := func(key string, mval *multival.MultiValue) bool {
		curval, ok := mval.Fetch(math.MaxInt64)
		if !ok {
			return true
		}
		if curval.Deleted && curval.Version < minVersion {
			if err := tx.Delete(ctx, key); err != nil {
				return false
			}
			count++
		}
		return true
	}
	db.store.Range(compact)

	if err := tx.Commit(ctx); err != nil {
		return -1
	}
	return count
}

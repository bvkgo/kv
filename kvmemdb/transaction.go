// Copyright (c) 2023 BVK Chaitanya

package kvmemdb

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"slices"
	"sort"

	"github.com/bvkgo/kv"
	"github.com/bvkgo/kv/kvmemdb/internal/multival"
)

type Transaction struct {
	db *DB

	version           int64
	lastCommitVersion int64

	// accesses caches key-values that are read/written by this transaction.
	accesses map[string]*multival.Value
}

func (t *Transaction) String() string {
	return fmt.Sprintf("TX-%d (%d)", t.version, t.lastCommitVersion)
}

func (t *Transaction) Get(ctx context.Context, key string) (io.Reader, error) {
	if len(key) == 0 {
		return nil, os.ErrInvalid
	}

	if v, ok := t.accesses[key]; ok {
		if v.Deleted {
			return nil, os.ErrNotExist
		}
		return bytes.NewReader(v.Data), nil
	}

	if mv, ok := t.db.store.Load(key); ok {
		if v, ok := mv.Fetch(t.lastCommitVersion); ok {
			// Make a local copy of the already-committed value.
			t.accesses[key] = v
			if !v.Deleted {
				return bytes.NewReader(v.Data), nil
			}
			return nil, os.ErrNotExist
		}
	}

	return nil, os.ErrNotExist
}

func (t *Transaction) Set(ctx context.Context, key string, value io.Reader) error {
	if len(key) == 0 {
		return os.ErrInvalid
	}

	data, err := io.ReadAll(value)
	if err != nil {
		return err
	}

	if v, ok := t.accesses[key]; ok {
		// Do not modify the values that are not created by this transaction.
		if v.Version == t.version {
			v.Data = data
			v.Deleted = false
			return nil
		}
	}

	t.accesses[key] = &multival.Value{
		Version: t.version,
		Data:    data,
	}
	return nil
}

func (t *Transaction) Delete(ctx context.Context, key string) error {
	if len(key) == 0 {
		return os.ErrInvalid
	}

	if v, ok := t.accesses[key]; ok {
		// Do not modify the values that are not created by this transaction.
		if v.Version == t.version {
			v.Data = nil
			v.Deleted = true
			return nil
		}
	}

	t.accesses[key] = &multival.Value{
		Version: t.version,
		Deleted: true,
	}
	return nil
}

func (t *Transaction) Ascend(ctx context.Context, begin, end string) (kv.Iterator, error) {
	if end != "" && begin > end {
		return nil, os.ErrInvalid
	}
	keys := t.db.keys(t.accesses)
	for k := range t.accesses {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	i, n := 0, len(keys)
	if begin != "" {
		i, _ = slices.BinarySearch(keys, begin)
	}
	if end != "" {
		n, _ = slices.BinarySearch(keys, end)
	}
	return newIterator(t.Get, keys[i:n], false /* descending */), nil
}

func (t *Transaction) Descend(ctx context.Context, begin, end string) (kv.Iterator, error) {
	if end != "" && begin > end {
		return nil, os.ErrInvalid
	}
	keys := t.db.keys(t.accesses)
	for k := range t.accesses {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	i, n := 0, len(keys)
	if begin != "" {
		i, _ = slices.BinarySearch(keys, begin)
	}
	if end != "" {
		n, _ = slices.BinarySearch(keys, end)
	}
	return newIterator(t.Get, keys[i:n], true /* descending */), nil
}

func (t *Transaction) Scan(ctx context.Context) (kv.Iterator, error) {
	keys := t.db.keys(t.accesses)
	for k := range t.accesses {
		keys = append(keys, k)
	}
	return newIterator(t.Get, keys, false /* descending */), nil
}

func (t *Transaction) Rollback(ctx context.Context) error {
	if t.db == nil {
		return os.ErrClosed
	}
	t.db.releasePin(t.lastCommitVersion)
	t.db = nil
	return nil
}

func (t *Transaction) Commit(ctx context.Context) error {
	if t.db == nil {
		return os.ErrClosed
	}
	defer func() {
		t.db.releasePin(t.lastCommitVersion)
		t.db = nil
	}()

	return t.db.commit(t)
}

// Copyright (c) 2023 BVK Chaitanya

package kvmemdb

import (
	"bytes"
	"context"
	"io"
	"os"
	"slices"
	"sort"

	"github.com/bvkgo/kv"
)

type Snapshot struct {
	db *DB

	lastCommitVersion int64
}

func (s *Snapshot) Discard(ctx context.Context) error {
	if s.db == nil {
		return os.ErrClosed
	}
	s.db.releasePin(s.lastCommitVersion)
	s.db = nil
	return nil
}

func (s *Snapshot) Get(ctx context.Context, key string) (io.Reader, error) {
	if len(key) == 0 {
		return nil, os.ErrInvalid
	}

	if mv, ok := s.db.store.Load(key); ok {
		if value, ok := mv.Fetch(s.lastCommitVersion); ok {
			if !value.Deleted {
				return bytes.NewReader(value.Data), nil
			}
		}
	}
	return nil, os.ErrNotExist
}

func (s *Snapshot) Ascend(ctx context.Context, begin, end string) (kv.Iterator, error) {
	keys := s.db.keys(nil)
	sort.Strings(keys)
	i, n := 0, len(keys)
	if begin != "" {
		i, _ = slices.BinarySearch(keys, begin)
	}
	if end != "" {
		n, _ = slices.BinarySearch(keys, end)
	}
	return newIterator(s.Get, keys[i:n], false /* descending */), nil
}

func (s *Snapshot) Descend(ctx context.Context, begin, end string) (kv.Iterator, error) {
	keys := s.db.keys(nil)
	sort.Strings(keys)
	i, n := 0, len(keys)
	if begin != "" {
		i, _ = slices.BinarySearch(keys, begin)
	}
	if end != "" {
		n, _ = slices.BinarySearch(keys, end)
	}
	return newIterator(s.Get, keys[i:n], true /* descending */), nil
}

func (s *Snapshot) Scan(ctx context.Context) (kv.Iterator, error) {
	keys := s.db.keys(nil)
	return newIterator(s.Get, keys, false /* descending */), nil
}

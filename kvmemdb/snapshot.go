// Copyright (c) 2023 BVK Chaitanya

package kvmemdb

import (
	"bytes"
	"context"
	"io"
	"os"
	"sort"

	"github.com/bvkgo/kv"
)

type Snapshot struct {
	db *DB

	lastCommitVersion int64
}

func (s *Snapshot) Close() error {
	if s.db == nil {
		return os.ErrClosed
	}
	s.db.releasePin(s.lastCommitVersion)
	s.db = nil
	return nil
}

func (s *Snapshot) Get(ctx context.Context, key string) (io.Reader, error) {
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
	return newIterator(s.Get, keys, begin, end), nil
}

func (s *Snapshot) Descend(ctx context.Context, begin, end string) (kv.Iterator, error) {
	keys := s.db.keys(nil)
	sort.Slice(keys, func(i, j int) bool { return keys[i] > keys[j] })
	return newIterator(s.Get, keys, begin, end), nil
}

func (s *Snapshot) Scan(ctx context.Context) (kv.Iterator, error) {
	keys := s.db.keys(nil)
	return newIterator(s.Get, keys, "", ""), nil
}

// Copyright (c) 2023 BVK Chaitanya

package kv

import (
	"context"
	"io"
)

type Getter interface {
	// Get reads a key-value pair. Returns nil on success.
	//
	// It may return a nil io.Reader on success if backend supports it.
	Get(ctx context.Context, key string) (io.Reader, error)
}

type Setter interface {
	// Set creates or updates a key-value pair. Returns nil on success.
	//
	// Depending on the backend, a nil io.Reader value may also be valid.
	Set(ctx context.Context, key string, value io.Reader) error
}

type Deleter interface {
	// Delete removes a key-value pair. Returns nil on success.
	//
	// It may return nil or os.ErrNotExist if key doesn't exist.
	Delete(ctx context.Context, key string) error
}

// Iterator represents a position in a range of key-value pairs visited by
// Ascend, Descend and Scan operations. If there is any error in reading a
// key-value pair, it is retained in the iterator and the iteration is stopped.
//
//	it, err := db.Ascend(ctx, "aaa", "zzz")
//	if err != nil {
//	  return err
//	}
//	defer kv.Close(it)
//
//	for k, v, err := it.Fetch(ctx, false); err == nil; k, v, err = it.Fetch(ctx, true) {
//	  ...
//	  if ... {
//	    break
//	  }
//	  ...
//	}
//
//	if _, _, err := it.Fetch(ctx, false); err != nil && !errors.Is(err, io.EOF) {
//	  return err
//	}
type Iterator interface {
	// Fetch returns the key-value pair at the current iterator position or the
	// next position. If next parameter is true, iterator position is
	// pre-incremented (or pre-decremented) before fetching the key-value pair.
	//
	// Returns io.EOF when the iterator reaches to the end. Returns a non-nil
	// error if iterator encounters any failures and the same error is repeated
	// for all further calls.
	Fetch(ctx context.Context, next bool) (string, io.Reader, error)
}

type Ranger interface {
	// Ascend returns key-value pairs of a range in ascending order through an
	// iterator. Range is determined by the `begin` and `end` parameters.
	//
	// The `begin` parameter identifies the smaller side key and the `end`
	// parameter identifies the larger side key. When they are both non-empty
	// `begin` must be lesser than the `end` or os.ErrInvalid is returned.
	//
	// When both `begin` and `end` are non-empty, then range starts at the
	// `begin` key (included) and stops before the `end` key (excluded).
	//
	// When both `begin` and `end` are empty, then all key-value pairs are part
	// of the range. They are returned in ascending order for `Ascend` and in
	// descending order for `Descend`.
	//
	// When `begin` is empty then it represents the smallest key and when `end`
	// is empty it represents the key-after the largest key (so that the largest
	// key is included in the range).
	Ascend(ctx context.Context, begin, end string) (Iterator, error)

	// Descend is same as `Ascend`, but returns the determined range in
	// descending order of the keys.
	Descend(ctx context.Context, begin, end string) (Iterator, error)
}

type Scanner interface {
	// Scanner returns all key-value pairs through an iterator. No specific order
	// is guaranteed, but each key is visited exactly once.
	Scan(ctx context.Context) (Iterator, error)
}

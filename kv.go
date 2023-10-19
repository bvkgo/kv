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

// Iterator represents a position in a range of key-value pairs visited by the
// scan, ascend, descend, etc. operations. If there is any error encountered
// when reading a key-value pair, then error is recorded in the iterator and
// the iteration is stopped.
//
// In a trasactional context, only the key-value pairs that are accessed
// through the iterator *may* be considered as accessed.
//
//	it, err := db.Ascend(ctx, "aaa", "zzz")
//	if err != nil {
//	  return err
//	}
//	defer kv.Close(it)
//
//	for k, v, ok := it.Current(ctx); ok; k, v, ok = it.Next(ctx) {
//	  ...
//	  if ... {
//	    break
//	  }
//	  ...
//	}
//
//	if err := it.Err(); err != nil {
//	  return err
//	}
type Iterator interface {
	// Err returns any error recorded in the iterator. Returns nil when
	// iteration has reached the end of range with no errors.
	Err() error

	// Current returns the key-value pair at the current position of the
	// iterator. Iterator could encounter IO errors at any position, including
	// the first object.
	//
	// Returns the key-value pair and true on reading a key-value pair
	// successfully. Returns a false when end of range is reached or on
	// encountering any failure.
	Current(context.Context) (string, io.Reader, bool)

	// Next advances the iterator to next position and reads the key-value pair
	// at that position similar to Current.
	Next(context.Context) (string, io.Reader, bool)
}

type Ranger interface {
	// Ascend returns key-value pairs in a given range as an iterator. Values are
	// iterated in lexical ascending order of the keys.
	//
	// When both `begin` and `end` are non-empty, then range starts at the
	// `begin` key included and stops before `end` key which is excluded.
	//
	// When both `begin` and `end` are empty, then all key-value entries are part
	// of the range in the ascending order.
	//
	// When `begin` is empty then it represents the smallest key and when `end`
	// is empty it represents the key *beyond* the largest key (so that the
	// largest key is included).
	Ascend(ctx context.Context, begin, end string) (Iterator, error)

	// Descend is semantically similar to Ascend, but iterates in lexical
	// descending order of the keys.
	Descend(ctx context.Context, begin, end string) (Iterator, error)
}

type Scanner interface {
	// Scanner returns all key-value pairs through an iterator. No specific order
	// is guaranteed, but each key is visited exactly once.
	Scan(ctx context.Context) (Iterator, error)
}

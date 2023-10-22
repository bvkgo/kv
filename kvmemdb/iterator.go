// Copyright (c) 2023 BVK Chaitanya

package kvmemdb

import (
	"context"
	"io"
)

type itGetter = func(context.Context, string) (io.Reader, error)

type Iterator struct {
	getter itGetter

	keys []string

	i, n, incr int
}

func newIterator(getter itGetter, keys []string, descending bool) *Iterator {
	it := &Iterator{
		getter: getter,
		keys:   keys,
		i:      0,
		n:      len(keys),
		incr:   1,
	}
	if descending {
		it.i = len(keys) - 1
		it.n = -1
		it.incr = -1
	}
	return it
}

func (it *Iterator) Err() error {
	return nil
}

func (it *Iterator) Current(ctx context.Context) (string, io.Reader, bool) {
	stop := func() bool {
		if it.incr > 0 {
			return it.i < it.n
		}
		return it.i > it.n
	}
	// Skip over the keys returning os.ErrNotExist error.
	for ; stop(); it.i += it.incr {
		key := it.keys[it.i]
		if value, err := it.getter(ctx, key); err == nil {
			return key, value, true
		}
	}
	return "", nil, false
}

func (it *Iterator) Next(ctx context.Context) (string, io.Reader, bool) {
	it.i += it.incr
	return it.Current(ctx)
}

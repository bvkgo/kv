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

	i, incr int
}

func newIterator(getter itGetter, keys []string, descending bool) *Iterator {
	it := &Iterator{
		getter: getter,
		keys:   keys,
		i:      0,
		incr:   1,
	}
	if descending {
		it.i = len(keys) - 1
		it.incr = -1
	}
	return it
}

func (it *Iterator) Fetch(ctx context.Context, advance bool) (string, io.Reader, error) {
	stop := func() bool {
		if it.incr > 0 {
			return it.i < len(it.keys)
		}
		return it.i >= 0
	}

	if advance {
		it.i += it.incr
	}

	for ; stop(); it.i += it.incr {
		key := it.keys[it.i]
		if value, err := it.getter(ctx, key); err == nil {
			return key, value, nil
		}
	}

	return "", nil, io.EOF
}

// Copyright (c) 2023 BVK Chaitanya

package kvmemdb

import (
	"context"
	"io"
	"slices"
	"sort"
)

type itGetter = func(context.Context, string) (io.Reader, error)

type Iterator struct {
	getter itGetter

	keys []string

	index, end int
}

func newIterator(getter itGetter, keys []string, begin, end string) *Iterator {
	var i, n int
	switch {
	case begin == "" && end == "":
		i = 0
		n = len(keys)
	case begin == "" && end != "":
		i = 0
		n, _ = slices.BinarySearch(keys, end)
	case begin != "" && end == "":
		i, _ = slices.BinarySearch(keys, begin)
		n = len(keys)
	default:
		i, _ = slices.BinarySearch(keys, begin)
		n, _ = slices.BinarySearch(keys, end)
	}

	it := &Iterator{
		getter: getter,
		keys:   keys,
		index:  i,
		end:    n,
	}
	return it
}

func (it *Iterator) SetRange(begin, end string, ascend bool) {
	if ascend {
		sort.Strings(it.keys)
	} else {
		sort.Slice(it.keys, func(i, j int) bool { return it.keys[i] > it.keys[j] })
	}

	var i, n int
	switch {
	case begin == "" && end == "":
		i = 0
		n = len(it.keys)
	case begin == "" && end != "":
		i = 0
		n, _ = slices.BinarySearch(it.keys, end)
	case begin != "" && end == "":
		i, _ = slices.BinarySearch(it.keys, begin)
		n = len(it.keys)
	default:
		i, _ = slices.BinarySearch(it.keys, begin)
		n, _ = slices.BinarySearch(it.keys, end)
	}
	it.index, it.end = i, n
}

func (it *Iterator) Err() error {
	return nil
}

func (it *Iterator) Current(ctx context.Context) (string, io.Reader, bool) {
	// Skip over the keys returning os.ErrNotExist error.
	for ; it.index < it.end; it.index++ {
		key := it.keys[it.index]
		if value, err := it.getter(ctx, key); err == nil {
			return key, value, true
		}
	}
	return "", nil, false
}

func (it *Iterator) Next(ctx context.Context) (string, io.Reader, bool) {
	it.index++
	return it.Current(ctx)
}

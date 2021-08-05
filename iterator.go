package kv

import "context"

type Iterator interface {
	// Ascend runs the user callback with key-value pairs from a selected range
	// in ascending order. Every key-value pair passed to the callback is
	// considered as READ by the transaction.
	//
	// When both i and j are non-empty, the range begins with min(i,j) which is
	// included and ends at max(i,j) which is excluded.
	//
	// When one of i or j is an empty string, then range extends to the largest
	// key (inclusive). When both i and j are empty strings, then range is all
	// keys (inclusive) in the database.
	Ascend(ctx context.Context, i, j string, cb func(context.Context, string, string) error) error

	// Descend is similar to Ascend, but works in the opposite direction. Every
	// key-value pair passed to the callback is considered as READ by the
	// transaction.
	//
	// When both i and j are non-empty, the range begins with max(i,j) which is
	// included and ends at min(i,j) which is excluded.
	//
	// When one of i or j is an empty string, then range extends to the smallest
	// key (inclusive) in the database. When both i and j are empty strings, then
	// range is all keys (inclusive) in the database.
	Descend(ctx context.Context, i, j string, cb func(context.Context, string, string) error) error
}
